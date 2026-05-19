# Changelog

All notable changes to qctx are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `--log-level LEVEL` CLI flag (`debug` / `info` / `warn` / `error`); still honors `QCTX_LOG_LEVEL` env var.
- `--config PATH` CLI flag to point at a specific config YAML; overrides `$QCTX_CONFIG` and `~/.qctx.yaml`.
- `make doctor` target — verifies required (Go, golangci-lint, goimports, make, git) and optional (docker, goreleaser) tools.
- `make help` self-documenting target (auto-extracted from `##` comments).
- `make cover-check` per-package coverage gate (default ≥80%, configurable via `COVER_THRESHOLD`); now part of `make ci`.
- ADRs 0001–0007 covering JSONL schema, enterprise networking, partial-success, Secret type, cobra, byte-stitching, retryablehttp.
- `CONTRIBUTING.md` with build/test/commit guidance and enterprise-host policy.

### Documentation

- README documents `--config` and `--log-level`.

## [0.1.0] — 2026-05-19

Initial release. Single static Go binary that gathers SonarQube + Nexus IQ + GitLab MR/pipeline context for AI-driven fix workflows.

### Added

- **`qctx fetch`** — live JSON bundle to stdout for interactive AI agents.
- **`qctx snapshot --out PATH.jsonl`** — JSONL artifact for GitLab pipeline jobs (`type`-discriminated records: `meta`, `sonar.issue`, `sonar.hotspot`, `sonar.measure`, `sonar.quality_gate`, `nexus.violation`, `gitlab.mr`, `gitlab.mr.diff_summary`, `gitlab.mr.discussion`, `gitlab.pipeline`, `gitlab.job`, `error`).
- **`qctx version`** — print version banner with build-time `Version`/`Commit`/`Date` from ldflags.
- **SonarQube client**: paginated `/api/issues/search` and `/api/hotspots/search`, `/api/measures/component`, `/api/qualitygates/project_status`, in-memory rule-description cache.
- **GitLab client**: MR meta + diff summary + discussions, MR pipelines, jobs, job-trace tail (Range bytes=-N).
- **Auto-discovery**: scrapes pipeline job traces for `-Dsonar.projectKey=…` or `sonar.projectKey=…` so users do not have to re-configure the project key the pipeline already declares.
- **Nexus IQ JSON reader**: parses policy-evaluation results into normalized `model.Violation` (component coords, CVEs, threat level, remediation `toVersion`, waiver status).
- **Enterprise-first networking**: per-source `--{sonar,gitlab}-url` (and env vars) with no public-cloud defaults; `--ca-cert` adds to the system trust pool; `--insecure` for dev (stderr warning every run); `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` honored; repeatable `--header 'Name: value'` for SSO proxies; `--{sonar,gitlab}-token-file` for CI secret-file mounts.
- **Filters**: `--severity` (repeatable), `--type` (repeatable), `--branch`, `--all`, `--include-resolved`, `--no-sonar` / `--no-gitlab` / `--no-nexus` / `--no-mr` / `--no-pipeline`.
- **Partial-success policy**: one source's failure does not abort the whole run; per-source `source_status` and error records are emitted. `--strict` opts into all-or-nothing.
- **`Secret` type** redacts tokens in logs and `%v` formatting.
- **GoReleaser** config: linux/darwin/windows × amd64/arm64 tarballs/zip + checksums + distroless Docker image.
- **GitLab CI pipeline** (`.gitlab-ci.yml`): lint → test+cover → build → e2e → release-on-tag.
- **Coverage gate** (`make cover-check`): per-package threshold defaults to 80%; CI runs it via `make ci`.

### Security

- `gosec` clean; `--insecure` warns every run; tokens are wrapped in a redacting `Secret` type.

### Documentation

- `README.md` — install, both modes, filters table, JSONL schema overview, auto-discovery, links to enterprise docs.
- `docs/enterprise.md` — custom URLs, CA bundles, proxy, SSO header injection, `--insecure`.
- `docs/examples/gitlab-ci.yml` — drop-in pipeline job for downstream projects.
- `docs/examples/docker-run.sh` — Docker invocation against self-hosted Sonar/GitLab with corporate CA.
- `docs/superpowers/specs/2026-05-19-qctx-design.md` — full design spec.
- `docs/superpowers/plans/2026-05-19-qctx.md` — 36-task implementation plan (executed via Ralph Loop).

[Unreleased]: https://example.com/yourorg/qctx/compare/v0.1.0...HEAD
[0.1.0]: https://example.com/yourorg/qctx/releases/tag/v0.1.0
