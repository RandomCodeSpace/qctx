# qctx — Quality Context for AI Agents — Design Spec

**Status:** Approved
**Date:** 2026-05-19
**Owner:** akshay
**Working name:** `qctx` (override at module rename task if a different name wins)

---

## 1. Problem

When an AI agent is asked to fix code-quality, security-hotspot, or vulnerable-dependency findings, it needs **structured, ranked, actionable context** from the tools that found them. SonarQube, Nexus IQ, and GitLab CI each expose this data through different APIs and reports. Today the agent must either be hand-fed a curated bundle by a human, or shell out to multiple bespoke clients and reconcile shapes itself.

**qctx** is a single static Go binary that gathers Sonar + Nexus IQ + GitLab MR/pipeline context for one project and emits a single normalized JSON (live) or JSONL artifact (pipeline snapshot) the agent can consume directly.

## 2. Goals

1. One invocation, everything the agent needs for a focused fix session.
2. Enterprise-first: every endpoint configurable; self-hosted Sonar / GitLab / behind-corporate-CA Just Works.
3. Two modes, one binary: live (stdout JSON) for interactive agents; snapshot (JSONL artifact) for pipelines.
4. Auto-discover Sonar project key from GitLab pipeline logs.
5. No code-fixing. The CLI only gathers context; the agent fixes.

## 3. Non-Goals

- Replacing `sonar-scanner`. We don't analyze source.
- Live Nexus IQ API integration. The user already produces a Nexus IQ JSON report in their pipeline; qctx reads the file.
- Multi-project orchestration. One project per invocation.
- VCS providers other than GitLab. GitLab first; the internal VCS interface keeps the door open for GitHub/Bitbucket later.

## 4. Architecture

```
   +--------------------------------------------------------+
   |                    qctx (single binary)                |
   |                                                        |
   |  +---------+ +----------+ +----------+  +-----------+  |
   |  |   CLI   |->|  Config |->|  Bundle |->|  Output   |  |
   |  | (cobra) | | (env +   | | (concur- | |  JSON /   |  |
   |  |         | |  flag +  | |  rent    | |  JSONL    |  |
   |  |         | |  file)   | |  gather) | |  streamed |  |
   |  +---------+ +----------+ +-----+----+ +-----------+  |
   |                                 |                      |
   |           +----------+----------+----------+           |
   |           v          v                     v           |
   |       +--------+ +--------+   +--------+ +---------+   |
   |       | Sonar  | | GitLab |   | Nexus  | | Auto-   |   |
   |       | client | | client |   | reader | | discover|   |
   |       +---+----+ +---+----+   +--------+ +---------+   |
   |           |          |                                  |
   |       HTTP + TLS + custom CA + proxy (per source)       |
   +--------------------------------------------------------+
```

## 5. Modes

### 5.1 Live (`qctx fetch`)
- Talks to all sources at invocation time.
- Emits one JSON object to stdout.
- Exit 0 if any source succeeded; non-zero only if every source failed (or in `--strict`).
- Used by an AI agent in an interactive session: `qctx fetch --mr <url> | claude-code feed`.

### 5.2 Snapshot (`qctx snapshot`)
- Same data, written as JSONL to `--out path/to/file.jsonl`.
- Designed to run as a GitLab job. The pipeline's `artifacts:` block keeps the file as a downloadable artifact.
- Later jobs, or external agents, pull the artifact via the GitLab artifacts API and read line-by-line.

## 6. Data Sources

### 6.1 SonarQube (REST API)

| Endpoint | Purpose |
|---|---|
| `GET /api/issues/search` | Issues (paginated, `componentKeys`, `branch`, `pullRequest`, `severities`, `types`, `statuses`) |
| `GET /api/hotspots/search` | Security hotspots (paginated) |
| `GET /api/measures/component` | Metrics: coverage, bugs, vulnerabilities, code_smells, security_hotspots, duplicated_lines_density, new_coverage, new_bugs, etc. |
| `GET /api/qualitygates/project_status` | Gate verdict + per-condition results |
| `GET /api/rules/show` | Rule HTML description (per unique rule key, cached) |

**Auth:** `Authorization: Bearer ${SONAR_TOKEN}`. Falls back to Basic (token-as-username) for older Sonar (<10.0).

### 6.2 GitLab (REST API)

MR URL parsing: `https://<host>/<namespace>/<...>/<project>/-/merge_requests/<iid>` → host, URL-encoded project path, IID.

| Endpoint | Purpose |
|---|---|
| `GET /api/v4/projects/{encpath}/merge_requests/{iid}` | MR meta |
| `GET /api/v4/projects/{encpath}/merge_requests/{iid}/changes` | Diff summary |
| `GET /api/v4/projects/{encpath}/merge_requests/{iid}/discussions` | Review threads (paginated) |
| `GET /api/v4/projects/{encpath}/merge_requests/{iid}/pipelines` | Pipelines belonging to the MR |
| `GET /api/v4/projects/{encpath}/pipelines/{pid}/jobs` | Jobs in the pipeline |
| `GET /api/v4/projects/{encpath}/jobs/{jid}/trace` | Plain-text job trace, used for both auto-discovery and failure excerpts |

**Auth:** `PRIVATE-TOKEN: ${GITLAB_TOKEN}` header (works with personal, project, or group access tokens).

### 6.3 Nexus IQ (file)

The user already runs Nexus IQ in their pipeline (`nexus-iq-cli -r results.json …`). qctx reads the report; no IQ API call.

Schema reference: Sonatype Policy Evaluation REST API v2 — `policyEvaluationResult.components[].violations[]`, with `componentIdentifier`, `pathnames`, `policyId`, `policyName`, `policyThreatCategory`, `policyThreatLevel`, `constraints`, `reasons`, and optionally `remediation.versionChanges[].toVersion`.

## 7. Custom URL & Enterprise Support

Every client constructor takes a `baseURL string`. **No defaults** to `sonarcloud.io`, `gitlab.com`, or `cloud.sonatype.com`. Resolution priority (highest wins):

1. CLI flag (`--sonar-url`, `--gitlab-url`)
2. Env var (`SONAR_HOST_URL`, `GITLAB_HOST_URL`)
3. Config file (`~/.qctx.yaml`)
4. No default — config error if unset.

### 7.1 TLS

- `--ca-cert path/to/ca-bundle.pem` — additional trust roots, concatenated with the system pool.
- `--insecure` — skip verify (dev/debug; emits a stderr warning every run).
- Honors `SSL_CERT_FILE` / `SSL_CERT_DIR`.

### 7.2 Proxy

Honors `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` automatically (Go's `http.ProxyFromEnvironment`).

### 7.3 Auth

- Env vars (CI-friendly): `SONAR_TOKEN`, `GITLAB_TOKEN`.
- Flags override: `--sonar-token`, `--gitlab-token`.
- File-based: `--sonar-token-file`, `--gitlab-token-file`.

### 7.4 Per-host headers

- `--header 'X-Forwarded-User: ci-bot'` — repeatable, applied to all requests. Covers enterprise SSO proxies that demand extra headers.

## 8. Auto-Discovery

Given `--mr <gitlab-mr-url>`:

1. Resolve project + MR IID.
2. Fetch the MR's most recent non-skipped pipeline.
3. For each job in the pipeline (parallel), fetch the **last 64 KiB** of the trace (range header; falls back to full trace on 416).
4. Regex-scan the trace for `-Dsonar.projectKey=<key>` or `sonar.projectKey=<key>` or `sonar.projectKey: <key>`. First match wins; results cached in-process.
5. If discovery fails and `--project KEY` was not provided, error with: `could not discover sonar.projectKey from pipeline {pid}; pass --project KEY explicitly`.

## 9. JSONL Schema

One record per line. Each line has a `type` discriminator.

```jsonl
{"type":"meta","tool":"qctx","version":"0.1.0","scanned_at":"2026-05-19T01:15:00Z","sonar_project_key":"my-svc","gitlab_project":"team/my-svc","branch":"feat-x","commit_sha":"abc123","mr_iid":42,"source_status":{"sonar":"ok","gitlab":"ok","nexus":"ok"}}
{"type":"sonar.issue","key":"AYn-aB...","rule":"java:S2095","severity":"MAJOR","issue_type":"BUG","file":"src/main/java/Foo.java","line":42,"end_line":42,"message":"Use try-with-resources.","author":"alice@example.com","effort":"5min","tags":["leak"],"status":"OPEN","rule_desc_html":"<h2>Why is this an issue?</h2>..."}
{"type":"sonar.hotspot","key":"AYn-h...","rule":"java:S2076","vulnerability_probability":"HIGH","status":"TO_REVIEW","file":"src/main/java/Cmd.java","line":15,"message":"Make sure that executing this OS command is safe here.","rule_desc_html":"..."}
{"type":"sonar.measure","metric":"coverage","value":78.4}
{"type":"sonar.quality_gate","status":"FAILED","conditions":[{"metric":"new_coverage","op":"LT","threshold":"80","actual":"72.1","status":"ERROR"}]}
{"type":"nexus.violation","component":"org.apache.commons:commons-text:1.9","manifest":"pom.xml","line":48,"policy":"Security-High","threat_level":8,"cves":["CVE-2022-42889"],"summary":"Apache Commons Text RCE","fix_version":"1.10.0","status":"open"}
{"type":"gitlab.mr","iid":42,"title":"feat: add export","description":"...","author":"alice","source_branch":"feat-x","target_branch":"main","web_url":"...","draft":false,"changes_count":"12"}
{"type":"gitlab.mr.diff_summary","files_changed":["src/main/java/Foo.java","pom.xml"],"additions":120,"deletions":34}
{"type":"gitlab.mr.discussion","id":"d1","author":"bob","body":"nit: rename","resolved":false,"file":"src/main/java/Foo.java","line":42}
{"type":"gitlab.pipeline","id":123456,"status":"failed","ref":"feat-x","sha":"abc123","web_url":"...","created_at":"...","duration":482}
{"type":"gitlab.job","name":"test","status":"failed","stage":"test","duration":127,"web_url":"...","failure_excerpt":"FAIL: TestFoo (0.01s)\n  expected 5, got 6"}
{"type":"error","source":"sonar","message":"401 unauthorized at /api/issues/search"}
```

Live mode emits the same records inside a single JSON object:

```json
{
  "meta": { ... },
  "sonar": { "issues": [...], "hotspots": [...], "measures": [...], "quality_gate": {...} },
  "nexus": { "violations": [...] },
  "gitlab": { "mr": {...}, "pipeline": {...}, "jobs_failed": [...] },
  "errors": [ { "source": "...", "message": "..." } ]
}
```

## 10. Filtering

| Flag | Effect | Default |
|---|---|---|
| `--severity` (repeat) | Sonar severity filter | none (all) |
| `--type` (repeat) | Sonar issue type filter | none (all) |
| `--branch <name>` | Override branch | MR source branch |
| `--all` | Include all open issues, not just MR-touched | off (MR-scoped) |
| `--include-resolved` | Include resolved/closed | off |
| `--no-sonar` `--no-gitlab` `--no-nexus` `--no-mr` `--no-pipeline` | Disable a source/feature | off |
| `--nexus-report PATH` | Nexus IQ JSON report path | unset (skips Nexus) |

## 11. Error Handling

- **Partial success:** each source failure goes into `meta.source_status` and as `{"type":"error", ...}` lines. Exit 0 if ≥1 source produced data.
- **All sources fail:** non-zero exit, error to stderr. Exit code mapped to category: auth=10, network=11, parse=12, config=13.
- **Strict mode** (`--strict`): any failure → non-zero.

## 12. Performance Targets

| Metric | Target |
|---|---|
| Cold start | < 100 ms |
| MR with ≤50 issues / ≤200 hotspots / ≤10 dep violations | < 5 s wall (RTT-dominated) |
| Memory | O(records) — JSONL streamed, not buffered |
| Concurrent fetch | Sonar + GitLab + Nexus run in parallel |
| Binary size | < 15 MB stripped |

## 13. Tech Stack

| Concern | Choice | Why |
|---|---|---|
| Language | Go 1.23+ | Single static binary, easy CI deploy |
| CLI | `github.com/spf13/cobra` | Industry standard |
| HTTP retries | `github.com/hashicorp/go-retryablehttp` | Tiny, battle-tested |
| Logging | `github.com/rs/zerolog` | Fast, structured, easy redaction |
| Tests | `github.com/stretchr/testify` + stdlib `httptest` | Familiar |
| Distribution | GoReleaser | Multi-platform binaries + Docker |

Total runtime deps: 4. No transitive bloat.

## 14. Repo Layout

```
qctx/
├── cmd/qctx/main.go
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── fetch.go
│   │   ├── snapshot.go
│   │   └── version.go
│   ├── config/config.go
│   ├── httpclient/client.go
│   ├── sonar/
│   │   ├── client.go
│   │   ├── issues.go
│   │   ├── hotspots.go
│   │   ├── measures.go
│   │   ├── qualitygate.go
│   │   └── rules.go
│   ├── gitlab/
│   │   ├── client.go
│   │   ├── mrurl.go
│   │   ├── mr.go
│   │   ├── pipeline.go
│   │   ├── jobtrace.go
│   │   └── projectkey.go
│   ├── nexus/
│   │   ├── schema.go
│   │   └── reader.go
│   ├── model/
│   │   ├── types.go
│   │   ├── json.go
│   │   └── jsonl.go
│   ├── bundle/
│   │   └── bundler.go
│   └── version/version.go
├── test/
│   ├── fixtures/
│   └── e2e/main_test.go
├── docs/
│   ├── enterprise.md
│   └── examples/
│       ├── gitlab-ci.yml
│       └── docker-run.sh
├── .gitlab-ci.yml
├── .goreleaser.yaml
├── .golangci.yaml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## 15. Testing Strategy

- **Unit:** every public function. Table-driven for parsing/filtering. ≥80% line coverage gate enforced in CI.
- **Mock-server:** `httptest.Server` for every API client.
- **Fixture:** anonymized real payloads in `test/fixtures/` for Sonar / GitLab / Nexus.
- **E2E:** `test/e2e/main_test.go` builds the binary, runs it against mock servers, parses stdout/JSONL.
- **Lint:** `golangci-lint` strict — `gosec`, `errcheck`, `govet`, `staticcheck`, `ineffassign`, `unused`, `goimports`.

## 16. Security

- Tokens are wrapped in a `Secret` type whose `String()` returns `***redacted***`.
- `--insecure` warns to stderr on every run.
- No telemetry, no phone-home.
- `gosec` on every CI run.

## 17. Out of Scope (explicit)

- Live Nexus IQ API client
- Code modification / fix application
- Issue triage UI or web dashboard
- Multi-project rollups
- SARIF or any other Sonar variants
- Non-GitLab VCS
- Webhook / push-driven invocation

---

## Appendix A — Example Pipeline Use

```yaml
quality-context:
  stage: report
  image: registry.example.com/qctx:0.1.0
  needs: ["sonar-scan", "nexus-iq-scan"]
  variables:
    SONAR_HOST_URL: "https://sonar.example.com"
    GITLAB_HOST_URL: "https://gitlab.example.com"
  script:
    - qctx snapshot
        --mr "$CI_MERGE_REQUEST_PROJECT_URL/-/merge_requests/$CI_MERGE_REQUEST_IID"
        --nexus-report nexus-iq-report.json
        --out qctx-report.jsonl
  artifacts:
    paths: [qctx-report.jsonl]
    expire_in: 30 days
```

## Appendix B — Example Agent Use (live)

```bash
qctx fetch --mr "https://gitlab.example.com/team/my-svc/-/merge_requests/42" \
  | jq '.sonar.issues[] | select(.severity == "CRITICAL")' \
  | claude-code fix-from-stdin
```
