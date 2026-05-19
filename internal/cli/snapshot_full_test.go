package cli_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

// Exercises all JSONL record types in one snapshot run: meta, sonar.issue, sonar.hotspot,
// sonar.measure, sonar.quality_gate, nexus.violation, gitlab.mr, gitlab.mr.diff_summary,
// gitlab.mr.discussion, gitlab.pipeline, gitlab.job.
func TestSnapshotEmitsAllRecordTypes(t *testing.T) {
	dir := t.TempDir()
	nexusPath := filepath.Join(dir, "nexus.json")
	require.NoError(t, os.WriteFile(nexusPath, []byte(`{
	  "applicationId":"app",
	  "policyEvaluationResult":{"components":[{
	    "componentIdentifier":{"format":"maven","coordinates":{"groupId":"g","artifactId":"a","version":"1.0"}},
	    "pathnames":["pom.xml"],
	    "violations":[{"policyId":"p","policyName":"Security-High","policyThreatLevel":7,"constraints":[{"reasons":[{"reason":"CVE-2024-12345"}]}]}]
	  }]}
	}`), 0o600))

	sonar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"issues":[{"key":"i1","rule":"r","severity":"MAJOR","type":"BUG","component":"p:a.java","line":1,"message":"m","status":"OPEN"}]}`))
		case "/api/hotspots/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"hotspots":[{"key":"h1","ruleKey":"r","vulnerabilityProbability":"HIGH","status":"TO_REVIEW","component":"p:b.java","line":2,"message":"m"}]}`))
		case "/api/measures/component":
			_, _ = w.Write([]byte(`{"component":{"measures":[{"metric":"coverage","value":"77.7"},{"metric":"bugs","value":"3"}]}}`))
		case "/api/qualitygates/project_status":
			_, _ = w.Write([]byte(`{"projectStatus":{"status":"ERROR","conditions":[{"metricKey":"new_coverage","comparator":"LT","errorThreshold":"80","actualValue":"60","status":"ERROR"}]}}`))
		case "/api/rules/show":
			_, _ = w.Write([]byte(`{"rule":{"key":"r","htmlDesc":"<p>d</p>"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer sonar.Close()

	gl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/merge_requests/9/pipelines"):
			_, _ = w.Write([]byte(`[{"id":1,"status":"failed","ref":"feat","sha":"abc","created_at":"2026-05-19T01:00:00Z","duration":50,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/9"):
			_, _ = w.Write([]byte(`{"iid":9,"title":"t","description":"d","source_branch":"feat","target_branch":"main","web_url":"u","draft":false,"changes_count":"2","author":{"username":"a"}}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/9/changes"):
			_, _ = w.Write([]byte(`{"changes":[{"new_path":"a.java","diff":"@@ -0,0 +1,2 @@\n+x\n+y\n"}]}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/9/discussions"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":"d1","notes":[{"author":{"username":"bob"},"body":"nit","resolved":false,"position":{"new_path":"a.java","new_line":1}}]}]`))
		case strings.Contains(r.URL.Path, "/pipelines/1/jobs"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":5,"web_url":"u"},{"id":12,"name":"test","status":"failed","stage":"t","duration":15,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/jobs/11/trace"):
			_, _ = w.Write([]byte("-Dsonar.projectKey=p"))
		case strings.HasSuffix(r.URL.Path, "/jobs/12/trace"):
			_, _ = w.Write([]byte("--- FAIL: TestX (0.01s)\nFAIL\n"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer gl.Close()

	out := filepath.Join(dir, "out.jsonl")
	var stdout, stderr bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"snapshot",
			"--sonar-url", sonar.URL, "--sonar-token", "t",
			"--gitlab-url", gl.URL, "--gitlab-token", "t",
			"--mr", "https://gl.example/p/x/-/merge_requests/9",
			"--nexus-report", nexusPath,
			"--out", out,
		},
		Stdout: &stdout, Stderr: &stderr,
	})
	require.Equal(t, 0, rc, "stderr=%s", stderr.String())

	f, err := os.Open(out)
	require.NoError(t, err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	types := map[string]int{}
	for scanner.Scan() {
		var v struct {
			Type string `json:"type"`
		}
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &v))
		types[v.Type]++
	}
	for _, want := range []string{
		"meta", "sonar.issue", "sonar.hotspot", "sonar.measure", "sonar.quality_gate",
		"nexus.violation", "gitlab.mr", "gitlab.mr.diff_summary", "gitlab.mr.discussion",
		"gitlab.pipeline", "gitlab.job",
	} {
		require.Greater(t, types[want], 0, "missing record type %q", want)
	}
}
