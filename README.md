# qctx

[![CI](https://img.shields.io/github/actions/workflow/status/RandomCodeSpace/qctx/ci.yml?branch=main&style=for-the-badge&logo=github&label=CI)](https://github.com/RandomCodeSpace/qctx/actions/workflows/ci.yml)
[![Security](https://img.shields.io/github/actions/workflow/status/RandomCodeSpace/qctx/security.yml?branch=main&style=for-the-badge&logo=github&label=Security)](https://github.com/RandomCodeSpace/qctx/actions/workflows/security.yml)
[![Sonar](https://img.shields.io/sonar/quality_gate/RandomCodeSpace_qctx?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarcloud)](https://sonarcloud.io/project/overview?id=RandomCodeSpace_qctx)
[![Coverage](https://img.shields.io/sonar/coverage/RandomCodeSpace_qctx?server=https%3A%2F%2Fsonarcloud.io&style=for-the-badge&logo=sonarcloud)](https://sonarcloud.io/component_measures?id=RandomCodeSpace_qctx&metric=coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/RandomCodeSpace/qctx?style=for-the-badge)](https://goreportcard.com/report/github.com/RandomCodeSpace/qctx)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/RandomCodeSpace/qctx)
[![Go Version](https://img.shields.io/github/go-mod/go-version/RandomCodeSpace/qctx?style=for-the-badge&logo=go)](go.mod)
[![Release](https://img.shields.io/github/v/release/RandomCodeSpace/qctx?style=for-the-badge&logo=github)](https://github.com/RandomCodeSpace/qctx/releases)
[![License](https://img.shields.io/github/license/RandomCodeSpace/qctx?style=for-the-badge)](LICENSE)
[![CodeQL](https://img.shields.io/badge/CodeQL-enabled-2ea44f?style=for-the-badge&logo=github)](https://github.com/RandomCodeSpace/qctx/security/code-scanning)

Quality context for AI agents. Gathers SonarQube + Nexus IQ + GitLab MR/pipeline state into one normalized payload so agents can fix issues without round-tripping multiple APIs.

## Status

Pre-1.0. CLI flags, env vars, and JSONL schema are stable across patch releases; minor releases may add fields but not remove them. Breaking changes only at 0.x → 0.(x+1) where flagged in CHANGELOG.

## Install

```bash
go install github.com/RandomCodeSpace/qctx/cmd/qctx@latest
# or download a release binary from your GitLab/GitHub releases page
```

Docker:
```bash
docker pull registry.example.com/qctx:latest
```

## Modes

| Mode | Command | Output | Use case |
|---|---|---|---|
| Live | `qctx fetch` | JSON to stdout | Interactive AI agent shells out |
| Snapshot | `qctx snapshot --out report.jsonl` | JSONL file | GitLab pipeline writes an artifact |

## Live mode quick start

```bash
export SONAR_HOST_URL=https://sonar.example.com
export SONAR_TOKEN=...
export GITLAB_HOST_URL=https://gitlab.example.com
export GITLAB_TOKEN=...

qctx fetch \
  --mr "https://gitlab.example.com/team/my-svc/-/merge_requests/42" \
  --nexus-report nexus-iq-report.json | jq '.sonar.issues[].severity'
```

## Pipeline snapshot quick start

```yaml
# .gitlab-ci.yml in your project
include: docs/examples/gitlab-ci.yml
```

See `docs/examples/gitlab-ci.yml` for the full job spec.

## Filters

| Flag | Effect |
|---|---|
| `--severity BLOCKER --severity CRITICAL` | repeatable severity filter |
| `--type BUG --type VULNERABILITY` | repeatable type filter |
| `--branch feat-x` | override branch |
| `--all` | include all open issues, not just MR-touched |
| `--include-resolved` | include resolved/closed |
| `--no-sonar` / `--no-gitlab` / `--no-nexus` | disable a source |
| `--strict` | non-zero exit on any source failure |
| `--config PATH` | path to YAML config (default: `$QCTX_CONFIG` or `~/.qctx.yaml`) |
| `--log-level LEVEL` | `debug` / `info` / `warn` / `error` (env: `QCTX_LOG_LEVEL`) |

## Enterprise

Custom URLs, CA bundles, SSO proxies, NO_PROXY, header injection: see `docs/enterprise.md`.

## JSONL schema

Each line is one record with a `type` discriminator. Types: `meta`, `sonar.issue`, `sonar.hotspot`, `sonar.measure`, `sonar.quality_gate`, `nexus.violation`, `gitlab.mr`, `gitlab.mr.diff_summary`, `gitlab.mr.discussion`, `gitlab.pipeline`, `gitlab.job`, `error`. Full spec: `docs/superpowers/specs/2026-05-19-qctx-design.md`.

## Auto-discovery

Given just a GitLab MR URL, qctx infers the Sonar project key from the pipeline's job traces (looks for `-Dsonar.projectKey=…` or `sonar.projectKey=…`). Override with `--project KEY`.

## Development

```bash
make tidy
make ci         # tidy + fmt + lint + cover + cover-check + build
make e2e        # e2e against mock servers
make doctor     # verify required + optional tools are installed
make help       # list all targets
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the workflow, commit conventions, and the enterprise-host policy.

## Project files

- [`CHANGELOG.md`](CHANGELOG.md) — release notes and pending changes
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how to build, test, and submit
- [`SECURITY.md`](SECURITY.md) — vulnerability reporting and threat model
- [`docs/adr/`](docs/adr/) — architecture decision records (why, not just what)
- [`docs/enterprise.md`](docs/enterprise.md) — self-hosted / CA / proxy / SSO guidance

## License

Apache-2.0
