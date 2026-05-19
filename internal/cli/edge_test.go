package cli_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

func TestFetchMissingMRandProjectErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"fetch",
			"--sonar-url", "https://s.example",
			"--sonar-token", "t",
			"--gitlab-url", "https://g.example",
			"--gitlab-token", "t",
			"--no-nexus",
		},
		Stdout: &out, Stderr: &errOut,
	})
	require.NotEqual(t, 0, rc)
	require.Contains(t, errOut.String(), "sonar project key unknown")
}

func TestSnapshotMissingOutErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"snapshot",
			"--sonar-url", "https://s.example",
			"--gitlab-url", "https://g.example",
			"--no-sonar", "--no-gitlab", "--no-nexus",
		},
		Stdout: &out, Stderr: &errOut,
	})
	require.NotEqual(t, 0, rc)
}

func TestFetchStrictAllFailures(t *testing.T) {
	sonar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer sonar.Close()
	gl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/merge_requests/42/pipelines"):
			_, _ = w.Write([]byte(`[{"id":1001,"status":"failed","ref":"feat","sha":"abc","created_at":"2026-05-19T01:00:00Z","duration":50,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42"):
			_, _ = w.Write([]byte(`{"iid":42,"title":"t","source_branch":"feat","target_branch":"main","web_url":"u","author":{"username":"a"}}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/changes"):
			_, _ = w.Write([]byte(`{"changes":[]}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/discussions"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/pipelines/1001/jobs"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":5,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/jobs/11/trace"):
			_, _ = w.Write([]byte("-Dsonar.projectKey=p"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer gl.Close()

	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"fetch",
			"--sonar-url", sonar.URL, "--sonar-token", "t",
			"--gitlab-url", gl.URL, "--gitlab-token", "t",
			"--mr", "https://gl.example/p/x/-/merge_requests/42",
			"--no-nexus",
			"--strict",
		},
		Stdout: &out, Stderr: &errOut,
	})
	require.NotEqual(t, 0, rc)
}

func TestFetchProjectFlagOverridesAutoDiscover(t *testing.T) {
	sonar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			require.Equal(t, "explicit-key", r.URL.Query().Get("componentKeys"))
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":0},"issues":[]}`))
		case "/api/hotspots/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":0},"hotspots":[]}`))
		case "/api/measures/component":
			_, _ = w.Write([]byte(`{"component":{"measures":[]}}`))
		case "/api/qualitygates/project_status":
			_, _ = w.Write([]byte(`{"projectStatus":{"status":"OK","conditions":[]}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer sonar.Close()

	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{
		Argv: []string{
			"fetch",
			"--sonar-url", sonar.URL, "--sonar-token", "t",
			"--project", "explicit-key",
			"--no-gitlab", "--no-nexus",
		},
		Stdout: &out, Stderr: &errOut,
	})
	require.Equal(t, 0, rc, "stderr=%s", errOut.String())
	require.Contains(t, out.String(), `"sonar_project_key": "explicit-key"`)
}
