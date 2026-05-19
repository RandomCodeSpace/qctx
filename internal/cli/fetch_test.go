package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

func TestFetchProducesJSONBundle(t *testing.T) {
	sonar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"issues":[{"key":"i1","rule":"java:S1","severity":"MAJOR","type":"BUG","component":"p:a.java","line":1,"message":"m","status":"OPEN"}]}`))
		case "/api/hotspots/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":0},"hotspots":[]}`))
		case "/api/measures/component":
			_, _ = w.Write([]byte(`{"component":{"measures":[{"metric":"coverage","value":"80.0"}]}}`))
		case "/api/qualitygates/project_status":
			_, _ = w.Write([]byte(`{"projectStatus":{"status":"OK","conditions":[]}}`))
		case "/api/rules/show":
			_, _ = w.Write([]byte(`{"rule":{"key":"java:S1","htmlDesc":"d"}}`))
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
			_, _ = w.Write([]byte(`{"changes":[{"new_path":"a.java","diff":"+a\n"}]}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/discussions"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/pipelines/1001/jobs"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":5,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/jobs/11/trace"):
			_, _ = w.Write([]byte("mvn sonar:sonar -Dsonar.projectKey=p"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer gl.Close()

	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"fetch",
			"--sonar-url", sonar.URL,
			"--sonar-token", "t",
			"--gitlab-url", gl.URL,
			"--gitlab-token", "t",
			"--mr", "https://gl.example/p/x/-/merge_requests/42",
			"--no-nexus",
		},
		Stdout: &out, Stderr: &errOut,
	})
	require.Equal(t, 0, rc, "stderr=%s", errOut.String())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.NotNil(t, got["meta"])
	require.NotNil(t, got["sonar"])
	require.Contains(t, got["meta"].(map[string]any), "sonar_project_key")
}
