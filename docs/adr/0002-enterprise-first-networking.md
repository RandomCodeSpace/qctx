# 0002 — Enterprise-first networking, no public-cloud defaults

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx is designed to run inside corporate networks against self-hosted SonarQube, GitLab, and (read-only) Nexus IQ. In that environment:

- Hosts are private DNS names, often behind a corporate proxy and TLS-terminated by an internal CA.
- "Sensible defaults" pointing at `sonarcloud.io` / `gitlab.com` / `cloud.sonatype.com` would be **wrong by default** — and silent misdirection is worse than a hard error.
- SSO front-ends may require extra request headers.
- Some sites still run TLS with internal CAs that are not in the system trust pool.

## Decision

1. Every API client constructor takes `BaseURL string` and returns an error if it is empty. There is no default.
2. Resolution priority is **CLI flag > env var > config file > error**. No default fallback.
3. TLS extras:
   - `--ca-cert path/to/bundle.pem` appends additional roots to the system trust pool.
   - `SSL_CERT_FILE` / `SSL_CERT_DIR` environment variables are honored.
   - `--insecure` disables verification, emits an unconditional stderr warning every run.
4. Proxy: `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` are honored via Go's `http.ProxyFromEnvironment`. No alternative knob.
5. Extra headers: `--header 'Name: value'` is repeatable for SSO proxies that demand identity headers.
6. CI enforcement: a grep gate in CI rejects any commit that introduces the strings `sonarcloud.io`, `gitlab.com`, or `cloud.sonatype.com` in source.

## Consequences

**Upside.**
- An enterprise operator can deploy qctx without reading the source to find a hardcoded URL to override.
- Failure modes are loud and early (missing URL → fail at startup, not at first request).
- A leaked binary cannot accidentally exfiltrate data to a public-cloud control plane.

**Downside.**
- Slightly more boilerplate to get started for someone on SonarCloud / GitLab.com — they must export two env vars. Mitigated by a `docs/enterprise.md` example.
- The grep gate has minor false-positive risk in test fixtures referencing those domains as examples. We accept the false positive cost and use `gl.example` / `sonar.example.com` in synthetic data.

## Alternatives considered

- **Default to public clouds, override for enterprise.** Rejected. The cost of a wrong default in our target environment (sending tokens or scan data to a host the operator did not configure) is too high.
- **Auto-detect from environment variables popular in CI** (e.g. `CI_SERVER_URL`). Tempting and partially adopted via the `docs/examples/gitlab-ci.yml`'s `GITLAB_HOST_URL: "${CI_SERVER_URL}"` pattern — but pushed into the user's YAML, not the binary's runtime defaults, so the operator stays in control.
- **Hard error on `--insecure`.** Considered, but real CA-rotation incidents make a temporary opt-out useful. The stderr-on-every-run warning preserves loudness without blocking emergencies.
