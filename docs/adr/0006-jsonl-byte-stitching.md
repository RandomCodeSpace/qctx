# 0006 — JSONL writer uses raw byte-stitching for the type discriminator

**Status:** Accepted
**Date:** 2026-05-19

## Context

Every JSONL record needs a leading `"type":"<discriminator>"` field that the consumer relies on (see [ADR-0001](0001-jsonl-schema-for-ai-agents.md)). We must inject this into structs whose JSON tag layout is defined elsewhere (`internal/model/types.go`) without:

1. Modifying every record type to know about its own discriminator.
2. Forcing the caller to assemble a wrapper struct per record kind.
3. Forcing a `map[string]any` round-trip that loses key order and field omission semantics.

## Decision

`internal/model/jsonl.go`'s `writeTyped` helper marshals the payload struct into raw JSON bytes, then splices `{"type":"<name>"` into the front of the existing object:

```go
raw, _ := json.Marshal(payload)       // -> `{"key":"v",...}`
out := append([]byte(`{"type":`), tb...)
out = append(out, ',')
out = append(out, raw[1:]...)          // payload minus its leading '{'
```

The empty-object case (`{}`) is handled explicitly to avoid emitting `{"type":"x",}` (invalid JSON).

## Consequences

**Upside.**
- Zero impact on `model.Issue`, `model.Hotspot`, etc. — they keep their pristine struct definitions and JSON tags.
- One pass over payload bytes; no reflection beyond what `json.Marshal` already does.
- `type` is always the **first** key on every line, which makes `grep -m1 '"type":"sonar.issue"'` work without parsing the whole line.

**Downside.**
- The stitching is fragile to non-object payloads (`null`, arrays, primitives). Caller contract: only structs that marshal to a JSON object. Enforced with a runtime check + `fmt.Errorf`.
- We assume `json.Marshal` produces an object starting with `'{'` and ending with `'}'`. This is guaranteed by the encoding/json spec for non-nil struct values, but a future Go change could in principle violate it. Mitigated by a `JSONLEachLineIsJSON` smoke test that re-parses every line.

## Alternatives considered

- **Per-type wrapper struct** (`type wrappedIssue struct { Type string; ... Issue }`). Multiplies the type surface; embeds make field-ordering surprises in marshal output.
- **`json.RawMessage` field on a common wrapper.** Two marshals per record (inner payload, then wrapper) — measurable overhead on large reports.
- **`map[string]any` re-key.** Loses `omitempty`, loses field order, slower.
- **A custom encoder.** Over-engineered for one well-bounded concern.
