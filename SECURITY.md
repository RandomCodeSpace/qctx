# Security Policy

## Supported versions

| Version | Status |
|---|---|
| 0.1.x | Supported |
| < 0.1 | Not released |

Only the latest minor in the active 0.x line receives security fixes.

## Reporting a vulnerability

**Please do not file a public issue.** Email the maintainer with details:

- Affected version (output of `qctx version`)
- Reproduction steps
- Impact assessment (what an attacker can read/write/cause)
- Suggested fix, if any

You will receive an acknowledgement within five business days. We will keep you informed throughout triage and remediation. If the issue is confirmed, we aim to ship a fix within 30 days and to credit you in the release notes unless you prefer otherwise.

## Threat model

qctx is a CLI run in trusted developer/CI environments. It is **not** a security boundary itself; it consumes credentials its operator supplies and forwards them to operator-configured hosts. The threat model assumes:

- The binary and its host are trusted.
- The operator's `SONAR_TOKEN` / `GITLAB_TOKEN` are scoped to the minimum needed (read-only on the relevant project).
- Network egress is constrained by the operator (corporate proxy, NO_PROXY, etc.).

Within that model the project takes care to:

- Wrap tokens in a `Secret` type that renders as `***redacted***` in logs and `%v` formatting (see [ADR-0004](docs/adr/0004-secret-type-for-token-redaction.md)).
- Require explicit configuration of every endpoint; **no defaults to public-cloud hosts** (see [ADR-0002](docs/adr/0002-enterprise-first-networking.md)).
- Run `gosec` in CI and treat `// #nosec` annotations as code review surface.
- Emit a stderr warning every run when `--insecure` is set.

## What is *not* in scope

- Sandboxing untrusted Sonar / GitLab / Nexus servers. A malicious server can return arbitrary JSON; qctx parses it with the standard library but does not assume the payload is benign for downstream consumers.
- Securing the JSONL artifact at rest. That is GitLab's responsibility (job artifacts inherit project permissions).
- Hardening against the operator's own machine being compromised.

## Vulnerability scope examples

We consider in-scope:

- Token leakage in logs, error messages, or panics.
- TLS misconfiguration (e.g. accepting expired/wrong-CN certs without `--insecure`).
- Path traversal via the `--out` flag.
- RCE / arbitrary process spawn via untrusted server responses.
- Goroutine leaks reachable from a server response.

We consider out-of-scope:

- Denial of service from a malicious server (we do not commit to throughput SLAs).
- Information disclosure from data the operator's tokens already have access to.
