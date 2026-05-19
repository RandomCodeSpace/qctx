# 0004 — `Secret` type for token redaction

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx handles two long-lived secrets per run: a SonarQube token and a GitLab access token. Past incidents in similar CLI tools have leaked tokens through:

- Structured log fields (`logger.Info().Interface("cfg", cfg)`)
- Error messages constructed via `%v` / `%+v` on a struct containing the token
- Panics that print the full `Config` value
- Verbose `--help` or `--debug` dumps
- Crash dumps and core files
- Test failure output that prints whole structs

The standard advice — "be careful what you print" — fails in practice. Reviewers can not catch every `fmt.Sprintf("%+v", cfg)` across the codebase, especially in third-party logging shims.

## Decision

Tokens are stored in a `Secret` type (`internal/config.Secret`), defined as `type Secret string`. It overrides:

- `String() string` → `"***redacted***"`
- `GoString() string` → `"***redacted***"`
- `Reveal() string` → returns the underlying string (only the HTTP client calls this)
- `IsEmpty() bool` → checks empty without leaking

The `Config` struct holds `SonarToken Secret` and `GitLabToken Secret`. Any code that prints a `Config` — directly, via `%v`/`%+v`, via `json.Marshal` on a wrapper, or via a structured logger that calls `Stringer` — gets the redacted form by default. The only path that yields the raw token is an explicit `tok.Reveal()` call, which is grep-able and code-reviewable.

## Consequences

**Upside.**
- Default behavior is safe. Leaks require active misuse, not a missed safety call.
- New contributors do not need to learn a redaction policy; the type system enforces it.
- Test output, panics, and error messages stay clean by construction.

**Downside.**
- `Secret` is not directly comparable to `string`; tests must use `.Reveal()` or `.IsEmpty()`. Mild friction.
- `Reveal()` is grep-able, but a determined caller can still escape the type. The safety is conventional, not cryptographic. Mitigated by gosec rules and code review.
- JSON marshalling: `Secret` marshals as a JSON string of its underlying value, which can leak if a `Config` is ever serialized to JSON. We do not serialize `Config`; we serialize `model.Bundle` which contains no `Secret`. A `MarshalJSON` override that emits `"***redacted***"` would be safer but would also break round-trip; we accept the residual risk.

## Alternatives considered

- **Plain `string` + discipline.** Rejected: this is exactly what every past leak relied on.
- **Opaque struct with no accessor (token retrieved via callback).** Heavier; the HTTP client setup becomes more convoluted. The `Reveal()` accessor pattern is enough friction without becoming bureaucratic.
- **OS keychain integration.** Out of scope for a CI-targeted tool; CI runners do not have keychains.
- **`go.uber.org/zap`-style `Field` redaction.** Couples redaction to a specific logger; does not help with `fmt`/error messages.
