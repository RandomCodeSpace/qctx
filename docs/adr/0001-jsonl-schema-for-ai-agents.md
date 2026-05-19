# 0001 — Type-discriminated JSONL schema for AI agent consumption

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx must hand a multi-source, multi-record dataset (issues, hotspots, measures, quality-gate verdict, dependency violations, MR meta, pipeline meta, jobs, errors) to an AI agent. Two transports exist:

- **Live** (`qctx fetch`) — synchronous, interactive: the agent shells out and pipes the output into its prompt.
- **Snapshot** (`qctx snapshot --out report.jsonl`) — pipeline artifact: a GitLab job writes the file, later jobs / agents download and read it.

Both transports must:

1. Be easy for a stream-oriented consumer to read line-by-line without loading everything into memory.
2. Be obvious to humans inspecting the artifact during debugging.
3. Allow new record kinds to be added without breaking older readers.
4. Make filtering (e.g. "show me only `sonar.issue` lines") trivial.

## Decision

The snapshot format is **JSON Lines (JSONL)** — one self-contained JSON object per line — with a mandatory `type` field on every object acting as a discriminator. Live mode produces the equivalent grouped JSON object (`meta` + `sonar` + `nexus` + `gitlab` + `errors`) for backward compatibility with `jq`-style pipelines.

Type values use a `<source>.<kind>` namespace:

```
meta
sonar.issue
sonar.hotspot
sonar.measure
sonar.quality_gate
nexus.violation
gitlab.mr
gitlab.mr.diff_summary
gitlab.mr.discussion
gitlab.pipeline
gitlab.job
error
```

## Consequences

**Upside.**
- Stream consumers — `jq -c '. | select(.type=="sonar.issue")'`, `grep -c '"type":"error"'`, `awk` — work directly.
- Adding a new record type is purely additive; readers that do not know the type can ignore it.
- The line-per-record shape sidesteps the worst-case memory profile of large bundles.
- Tail-readability: a tail of the file shows the most recent records of any type.

**Downside.**
- The discriminator field is duplicated on every line; for very large reports the overhead is real (~15 bytes per record). Mitigated by gzip on GitLab's artifact transport.
- `type` is reserved at the top level of every record schema. Consumers cannot use `type` for record-internal meaning.

## Alternatives considered

- **Single nested JSON object.** The live mode does this. For snapshot mode it loses the stream-friendly property and forces consumers to load the whole file before doing anything useful.
- **One file per record type** (`issues.json`, `hotspots.json`, …). Multiplies the artifact count, complicates the GitLab `artifacts:` block, and forces consumers to join records that arrived from a single scan.
- **Protocol Buffers / Avro.** Higher consumer friction. AI agents are happiest reading text; binary formats require a schema file to make sense of the bytes.
- **NDJSON without a discriminator** (rely on shape inference). Brittle once two record types share a field name; loses the "ignore unknown types" robustness property.
