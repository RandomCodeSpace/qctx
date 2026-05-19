# 0003 — Partial-success policy and `--strict` opt-in

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx gathers data from three independent sources (SonarQube, GitLab, Nexus IQ JSON). Real-world failures are common and uncorrelated:

- SonarQube is down for an upgrade window while GitLab is fine.
- The MR exists but its pipeline was pruned (404 on jobs).
- The Nexus IQ JSON path is wrong because the upstream job was skipped.
- Auth token rotated for one source but not the others.

An AI agent consuming qctx output usually wants *as much context as we can give it*, not *all-or-nothing*. Returning nothing because one source failed forces the operator to manually fan out and re-collect.

On the other hand, CI gates sometimes need the opposite: any source failure must block.

## Decision

By default qctx applies a **partial-success policy**:

1. Each source fetch runs concurrently in its own goroutine. Failures are captured per-source, not propagated.
2. The `meta.source_status` map records `"ok"` or `"error"` per source.
3. Each source error becomes an `{"type":"error","source":"…","message":"…"}` record (snapshot mode) or an entry in the `errors` array (live mode).
4. Exit code is `0` if *at least one source produced data*; non-zero only when every source failed.

A `--strict` flag opts into all-or-nothing: any source failure is fatal, exit code is non-zero.

## Consequences

**Upside.**
- The common, useful-context-now case is the default. AI agents always get something to work with when something is available.
- Operators can debug a single broken integration without losing visibility into the others.
- `--strict` keeps the all-or-nothing option available for CI gates that demand it.

**Downside.**
- Consumers must look at `source_status` (and the `error` records) to know that part of the data is missing. A naive consumer might fix only the Sonar issues and miss that Nexus was unreachable.
- Mitigation: the per-source error records are loud and easy to grep; the bundler emits them in both modes.

## Alternatives considered

- **All-or-nothing by default.** Rejected for the reasons above — too brittle for the day-to-day enterprise environment.
- **Per-source exit codes** (`exit_sonar`, `exit_gitlab`, …). Cannot express via process exit code; would require parsing the output, which defeats the point of a simple CLI contract.
- **Best-effort with no error records.** Silent partial failure is the worst of both worlds. A downstream agent has no way to know that a "no issues" result might mean "Sonar was unreachable" rather than "the code is clean".
