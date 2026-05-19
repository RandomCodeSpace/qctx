package bundle_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/bundle"
	"github.com/RandomCodeSpace/qctx/internal/gitlab"
	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

func newHTTP(t *testing.T) *httpclient.Client {
	t.Helper()
	c, err := httpclient.New(httpclient.Options{Timeout: 2 * time.Second, MaxRetries: 1, RetryWait: 5 * time.Millisecond})
	require.NoError(t, err)
	return c
}

func TestSonarAdapterDelegatesAllMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"issues":[{"key":"i1","rule":"r","severity":"MAJOR","type":"BUG","component":"p:a.java","line":1,"message":"m","status":"OPEN"}]}`))
		case "/api/hotspots/search":
			_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":1},"hotspots":[{"key":"h1","ruleKey":"r","vulnerabilityProbability":"HIGH","status":"TO_REVIEW","component":"p:b.java","line":2,"message":"m"}]}`))
		case "/api/measures/component":
			_, _ = w.Write([]byte(`{"component":{"measures":[{"metric":"coverage","value":"90.0"}]}}`))
		case "/api/qualitygates/project_status":
			_, _ = w.Write([]byte(`{"projectStatus":{"status":"OK","conditions":[]}}`))
		case "/api/rules/show":
			_, _ = w.Write([]byte(`{"rule":{"key":"r","htmlDesc":"<p>d</p>"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	c, err := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: newHTTP(t)})
	require.NoError(t, err)

	a := bundle.NewSonarAdapter(bundle.SonarAdapterDeps{Client: c, ProjectKey: "p", Branch: "feat"})

	iss, err := a.Issues()
	require.NoError(t, err)
	require.Len(t, iss, 1)
	require.Equal(t, "<p>d</p>", iss[0].RuleDescHTML)

	hs, err := a.Hotspots()
	require.NoError(t, err)
	require.Len(t, hs, 1)
	require.Equal(t, "<p>d</p>", hs[0].RuleDescHTML)

	ms, err := a.Measures()
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.InDelta(t, 90.0, ms[0].Value, 1e-6)

	qg, err := a.QualityGate()
	require.NoError(t, err)
	require.Equal(t, "PASSED", qg.Status)
}

func TestSonarAdapterNilClientYieldsZero(t *testing.T) {
	a := bundle.NewSonarAdapter(bundle.SonarAdapterDeps{})
	iss, err := a.Issues()
	require.NoError(t, err)
	require.Nil(t, iss)
	hs, err := a.Hotspots()
	require.NoError(t, err)
	require.Nil(t, hs)
	ms, err := a.Measures()
	require.NoError(t, err)
	require.Nil(t, ms)
	qg, err := a.QualityGate()
	require.NoError(t, err)
	require.Equal(t, "", qg.Status)
}

func TestGitLabAdapterDelegatesAllMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/merge_requests/42/pipelines"):
			_, _ = w.Write([]byte(`[{"id":1001,"status":"success","ref":"feat","sha":"a1","web_url":"u","created_at":"2026-05-19T00:00:00Z","duration":10}]`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42"):
			_, _ = w.Write([]byte(`{"iid":42,"title":"t","source_branch":"feat","target_branch":"main","web_url":"u","author":{"username":"alice"}}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/changes"):
			_, _ = w.Write([]byte(`{"changes":[{"new_path":"a.java","diff":"+a\n"}]}`))
		case strings.HasSuffix(r.URL.Path, "/merge_requests/42/discussions"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/pipelines/1001/jobs"):
			w.Header().Set("X-Next-Page", "")
			_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":5,"web_url":"u"},{"id":12,"name":"test","status":"failed","stage":"t","duration":15,"web_url":"u"}]`))
		case strings.HasSuffix(r.URL.Path, "/jobs/12/trace"):
			_, _ = w.Write([]byte("--- FAIL: TestX\nFAIL\n"))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	c, err := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: newHTTP(t)})
	require.NoError(t, err)

	a := bundle.NewGitLabAdapter(bundle.GitLabAdapterDeps{Client: c, ProjectPath: "p/x", MRIID: 42})

	mr, err := a.MR()
	require.NoError(t, err)
	require.Equal(t, 42, mr.IID)

	diff, err := a.DiffSummary()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a.java"}, diff.FilesChanged)

	disc, err := a.Discussions()
	require.NoError(t, err)
	require.Empty(t, disc)

	pl, err := a.Pipeline()
	require.NoError(t, err)
	require.Equal(t, 1001, pl.ID)

	jobs, err := a.Jobs()
	require.NoError(t, err)
	require.Len(t, jobs, 2)
	for _, j := range jobs {
		if j.Status == "failed" {
			require.Contains(t, j.FailureExcerpt, "FAIL")
		}
	}
}

func TestGitLabAdapterPipelineByID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/merge_requests/9/pipelines") {
			_, _ = w.Write([]byte(`[
			  {"id":1,"status":"success","ref":"feat","sha":"a","web_url":"u","created_at":"2026-05-19T00:00:00Z","duration":1},
			  {"id":2,"status":"failed","ref":"feat","sha":"b","web_url":"u","created_at":"2026-05-19T01:00:00Z","duration":2}
			]`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: newHTTP(t)})

	a := bundle.NewGitLabAdapter(bundle.GitLabAdapterDeps{Client: c, ProjectPath: "p/x", MRIID: 9, PipelineID: 1})
	pl, err := a.Pipeline()
	require.NoError(t, err)
	require.Equal(t, 1, pl.ID, "should pick specified PipelineID, not latest")
}

func TestGitLabAdapterNilClientYieldsZero(t *testing.T) {
	a := bundle.NewGitLabAdapter(bundle.GitLabAdapterDeps{})
	_, err := a.MR()
	require.NoError(t, err)
	_, err = a.DiffSummary()
	require.NoError(t, err)
	disc, err := a.Discussions()
	require.NoError(t, err)
	require.Nil(t, disc)
	_, err = a.Pipeline()
	require.NoError(t, err)
	jobs, err := a.Jobs()
	require.NoError(t, err)
	require.Nil(t, jobs)
}

func TestNexusAdapterReturnsStoredViolations(t *testing.T) {
	a := bundle.NewNexusAdapter(bundle.NexusAdapterDeps{})
	v, err := a.Violations()
	require.NoError(t, err)
	require.Empty(t, v)
}
