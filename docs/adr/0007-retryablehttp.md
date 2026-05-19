# 0007 — `hashicorp/go-retryablehttp` for HTTP retries

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx makes synchronous HTTP calls to SonarQube and GitLab inside CI jobs. Both APIs return 503 during upgrade windows and 429 under load. A 30-second flake should not fail an MR; brief retry with backoff is the right behavior. We need:

- Idempotent retry on transient errors (network errors, 5xx, 429).
- No retry on 4xx (auth failures, missing resources).
- Configurable max retries and base wait.
- A response we can read after retries complete.
- Survival behind a corporate proxy with its own retry quirks.
- Plays well with our custom `*http.Transport` (custom CA + proxy from env).

## Decision

Use `github.com/hashicorp/go-retryablehttp` v0.7+ as the HTTP middleware. It wraps `*http.Client` and exposes the same `Do(*http.Request)` contract via `retryablehttp.FromRequest`. Our `httpclient.New` configures it with:

- `RetryMax`: from `Options.MaxRetries` (default 3).
- `RetryWaitMin`: from `Options.RetryWait` (default 500 ms).
- `RetryWaitMax`: `8 * RetryWaitMin`.
- `Backoff`: `retryablehttp.DefaultBackoff` (exponential with jitter).
- `CheckRetry`: a custom function that retries on network errors, nil response, 5xx, and 429 — but **not** on 4xx.
- `Logger: nil` to suppress its default stdout chatter (we route logs through zerolog).

## Consequences

**Upside.**
- A small, focused dependency (`hashicorp/go-cleanhttp` is the only transitive). No connection pool reinvention.
- Test simulation is straightforward — `httptest.Server` returning 503 a few times, then 200.
- The library is widely deployed (Terraform, Vault, Consul) and stable.

**Downside.**
- It owns request reading (`FromRequest`). Streaming-upload requests aren't natively supported, but qctx is GET-only so this never bites.
- `Logger: nil` is a workaround for noisy defaults; future versions may change the API. Verified at the version pinned in `go.mod`.
- One more module in the dependency graph.

## Alternatives considered

- **Hand-rolled retry loop.** Possible, but exponential backoff with jitter is famously easy to get subtly wrong (esp. the "thundering herd" mitigation).
- **`avast/retry-go`.** Generic retry, not HTTP-aware. Would still require us to wire response-status conditions ourselves.
- **`heimdalr/heimdall`.** Heavier — pulls hystrix-style circuit breakers we do not need.
- **In-cluster sidecar / `envoy` retries.** Out of scope; qctx must work as a standalone binary from a developer laptop too.
