//go:build e2e

package e2e

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQctxSnapshotE2E(t *testing.T) {
	sonar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"issues":[{"key":"i1","rule":"java:S2095","severity":"MAJOR","type":"BUG","component":"my-svc:src/Foo.java","line":42,"message":"Use try-with-resources.","author":"a","tags":["leak"],"status":"OPEN"}]}`))
		case "/api/hotspots/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":0},"hotspots":[]}`))
		case "/api/measures/component":
			_, _ = w.Write([]byte(`{"component":{"measures":[{"metric":"coverage","value":"78.4"}]}}`))
		case "/api/qualitygates/project_status":
			_, _ = w.Write([]byte(`{"projectStatus":{"status":"ERROR","conditions":[{"metricKey":"new_coverage","comparator":"LT","errorThreshold":"80","actualValue":"72.1","status":"ERROR"}]}}`))
		case "/api/rules/show":
			_, _ = w.Write([]byte(`{"rule":{"key":"java:S2095","htmlDesc":"<h2>Why</h2>"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer sonar.Close()

	gl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/merge_requests/42/pipelines"):
			_, _ = w.Write([]byte(`[{"id":1001,"status":"failed","ref":"feat","sha":"abc","created_at":"2026-05-19T01:00:00Z","duration":50,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42"):
			_, _ = w.Write([]byte(`{"iid":42,"title":"feat","source_branch":"feat","target_branch":"main","web_url":"u","draft":false,"changes_count":"1","author":{"username":"alice"}}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/changes"):
			_, _ = w.Write([]byte(`{"changes":[{"new_path":"src/Foo.java","diff":"@@ -0,0 +1,3 @@\n+a\n+b\n+c\n"}]}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/discussions"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/pipelines/1001/jobs"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":5,"web_url":"u"},{"id":12,"name":"test","status":"failed","stage":"t","duration":30,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/jobs/11/trace"):
			_, _ = w.Write([]byte("mvn sonar:sonar -Dsonar.projectKey=my-svc"))
		case strings.HasSuffix(r.URL.Path, "/jobs/12/trace"):
			_, _ = w.Write([]byte("--- FAIL: TestFoo (0.01s)\n  foo_test.go:12: expected 5 got 6\nFAIL\n"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer gl.Close()

	tmp := t.TempDir()
	out := filepath.Join(tmp, "report.jsonl")
	bin := os.Getenv("QCTX_BIN")
	if bin == "" {
		bin = "../../bin/qctx"
	}
	cmd := exec.Command(bin,
		"snapshot",
		"--sonar-url", sonar.URL, "--sonar-token", "t",
		"--gitlab-url", gl.URL, "--gitlab-token", "t",
		"--mr", "https://gl.example/team/my-svc/-/merge_requests/42",
		"--nexus-report", "../fixtures/nexus_sample.json",
		"--out", out,
	)
	combined, err := cmd.CombinedOutput()
	require.NoError(t, err, "stdout/err: %s", string(combined))

	expectedB, err := os.ReadFile("../fixtures/expected_jsonl_types.txt")
	require.NoError(t, err)
	expected := map[string]bool{}
	for _, line := range strings.Split(strings.TrimRight(string(expectedB), "\n"), "\n") {
		expected[line] = false
	}

	f, err := os.Open(out)
	require.NoError(t, err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256<<10), 1<<20)
	for scanner.Scan() {
		var v struct {
			Type string `json:"type"`
		}
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &v))
		if _, ok := expected[v.Type]; ok {
			expected[v.Type] = true
		}
	}
	require.NoError(t, scanner.Err())
	for typ, seen := range expected {
		require.True(t, seen, "expected JSONL type %q in artifact", typ)
	}
}
