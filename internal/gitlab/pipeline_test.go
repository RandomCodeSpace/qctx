package gitlab_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/gitlab"
	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

func TestListMRPipelines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/merge_requests/42/pipelines") {
			_, _ = w.Write([]byte(`[
			  {"id":1001,"status":"success","ref":"feat","sha":"a1","web_url":"u1","created_at":"2026-05-19T00:00:00Z"},
			  {"id":1002,"status":"failed","ref":"feat","sha":"a2","web_url":"u2","created_at":"2026-05-19T01:00:00Z"}
			]`))
		}
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.ListMRPipelines("p/x", 42)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, 1001, got[0].ID)
	require.Equal(t, "failed", got[1].Status)
}

func TestLatestPipelinePicksMostRecentNonSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[
		  {"id":1,"status":"skipped","ref":"feat","sha":"a1","created_at":"2026-05-19T00:00:00Z"},
		  {"id":2,"status":"success","ref":"feat","sha":"a2","created_at":"2026-05-19T02:00:00Z"},
		  {"id":3,"status":"failed","ref":"feat","sha":"a3","created_at":"2026-05-19T01:00:00Z"}
		]`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.LatestMRPipeline("p/x", 42)
	require.NoError(t, err)
	require.Equal(t, 2, got.ID) // most recent created_at, non-skipped
}

func TestListPipelineJobs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Next-Page", "")
		_, _ = w.Write([]byte(`[
		  {"id":11,"name":"sonar","status":"success","stage":"analysis","duration":30,"web_url":"u11"},
		  {"id":12,"name":"test","status":"failed","stage":"test","duration":127,"web_url":"u12"}
		]`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.ListPipelineJobs("p/x", 1001)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "failed", got[1].Status)
}
