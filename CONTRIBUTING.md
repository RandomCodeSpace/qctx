# Contributing to qctx

Thanks for considering a contribution. This document explains how to build, test, and submit changes.

## Prerequisites

- Go ≥ 1.23
- `golangci-lint` v2.x (https://golangci-lint.run/usage/install/)
- `make`
- (optional) `docker` for the release image
- (optional) `goreleaser` for cutting a release locally

Check your toolchain:

```bash
go version && golangci-lint --version && make --version | head -1
```

## Project layout

```
cmd/qctx/                # CLI entrypoint
internal/cli/            # cobra commands (fetch, snapshot, version, root)
internal/bundle/         # concurrent source gather + adapters
internal/sonar/          # SonarQube REST client
internal/gitlab/         # GitLab REST client + MR URL parser + auto-discovery
internal/nexus/          # Nexus IQ JSON reader
internal/model/          # normalized output types + JSON / JSONL emitters
internal/config/         # config loader (flag > env > file)
internal/httpclient/     # shared retrying HTTP with custom CA + proxy + headers
internal/logging/        # zerolog bootstrap
internal/version/        # build-time version vars
test/e2e/                # end-to-end test (build-tag e2e) against mock servers
test/fixtures/           # canned API payloads + expected JSONL types
docs/                    # design spec, plan, enterprise/examples
```

## Workflow

```bash
make test            # fast inner loop
make lint            # lint everything
make ci              # full CI run (tidy + fmt + lint + cover + cover-check + build)
make e2e             # end-to-end test against mock servers
```

## Tests

- **TDD encouraged.** Each feature has a RED test before implementation.
- **Table-driven** for parsing/filtering.
- **Mock servers** (`net/http/httptest`) for API client tests — never hit a real Sonar / GitLab.
- **No flaky timing.** No `time.Sleep` in tests; use channels or `t.Cleanup`.
- **Coverage gate.** `make cover-check` enforces ≥80% per logic package. Push that bar up when you can.

## Commits

- [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `chore:`, `test:`, `docs:`, `refactor:`, `style:`, `ci:`.
- Atomic — one logical change per commit.
- Imperative mood (`add foo`, not `added foo`).
- First line ≤ 72 chars.

## Custom URLs / enterprise hosts

qctx must remain **enterprise-first**. Do **not** introduce hardcoded defaults for `sonarcloud.io`, `gitlab.com`, or `cloud.sonatype.com`. Every API client constructor takes a `BaseURL` and errors when it is empty. A CI grep enforces this:

```bash
! grep -rIE 'sonarcloud\.io|(^|[^.])gitlab\.com|cloud\.sonatype\.com' --include='*.go' .
```

## JSONL schema stability

The `type`-discriminated JSONL schema in `internal/model/jsonl.go` is part of the project's contract with downstream agents. Within a 0.x minor series:

- Adding **new fields** to a record type is **allowed**.
- Adding **new record types** is **allowed**.
- Renaming, removing, or changing the type of an existing field is a **breaking change** and requires a minor-version bump plus a CHANGELOG entry under `### Changed` or `### Removed`.

## Security

- Tokens live inside the `Secret` type (`internal/config`). `Secret.String()` returns `***redacted***`; do not log via `Reveal()`.
- `gosec` runs in CI. New `// #nosec` annotations must explain why (`// #nosec G304 -- reason`).
- The `--insecure` flag emits a stderr warning every run — keep that warning.

## Releasing

Tags drive GoReleaser. To cut a release locally for testing:

```bash
goreleaser release --clean --snapshot
```

A real tag (`vX.Y.Z`) on the default branch triggers the `release` stage in `.gitlab-ci.yml`.

## Reviewing checklist

Before requesting review:

- [ ] `make ci` is green locally (tidy, fmt, lint, test, cover, cover-check, build).
- [ ] If you touched the JSONL schema, the change is additive or you have CHANGELOG notes.
- [ ] If you touched a Sonar / GitLab endpoint, you added or updated an httptest-based test.
- [ ] No public-cloud URL literals were introduced.
- [ ] Any new flag is documented in `README.md`, the relevant `bindCommonFlags` block, and (if user-facing) `docs/enterprise.md`.

## Design context

Before changing core behavior, check the relevant ADR in [`docs/adr/`](docs/adr/):

- [0001](docs/adr/0001-jsonl-schema-for-ai-agents.md) — JSONL schema with `type` discriminator.
- [0002](docs/adr/0002-enterprise-first-networking.md) — no public-cloud defaults.
- [0003](docs/adr/0003-partial-success-policy.md) — partial-success and `--strict`.
- [0004](docs/adr/0004-secret-type-for-token-redaction.md) — `Secret` type for tokens.
- [0005](docs/adr/0005-cobra-cli-framework.md) — cobra as CLI framework.
- [0006](docs/adr/0006-jsonl-byte-stitching.md) — JSONL writer's byte-stitching approach.
- [0007](docs/adr/0007-retryablehttp.md) — HTTP retry library choice.

If your change supersedes an ADR, write a new ADR that references the old one.

## License

By contributing you agree your contributions are licensed under the project's Apache-2.0 license.
