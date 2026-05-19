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

func TestGetMR(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/merge_requests/42") {
			_, _ = w.Write([]byte(`{
			  "iid":42,"title":"feat","description":"d","author":{"username":"alice"},
			  "source_branch":"feat-x","target_branch":"main","web_url":"http://x/mr/42",
			  "draft":false,"changes_count":"12"
			}`))
		}
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	mr, err := c.GetMR("p/x", 42)
	require.NoError(t, err)
	require.Equal(t, 42, mr.IID)
	require.Equal(t, "feat", mr.Title)
	require.Equal(t, "feat-x", mr.SourceBranch)
	require.Equal(t, "alice", mr.Author)
	require.Equal(t, "12", mr.ChangesCount)
}

func TestGetMRDiffSummary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
		  "changes":[
		    {"new_path":"a.java","diff":"@@ -0,0 +1,3 @@\n+a\n+b\n+c\n"},
		    {"new_path":"b.java","diff":"@@ -1,3 +1,1 @@\n-x\n-y\n+z\n"}
		  ]
		}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.GetMRDiffSummary("p/x", 42)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a.java", "b.java"}, got.FilesChanged)
	require.Equal(t, 4, got.Additions) // 3 + 1
	require.Equal(t, 2, got.Deletions) // 0 + 2
}

func TestListDiscussions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Next-Page", "")
		_, _ = w.Write([]byte(`[
		  {"id":"d1","notes":[{"author":{"username":"bob"},"body":"nit","resolved":false,"position":{"new_path":"a.java","new_line":42}}]},
		  {"id":"d2","notes":[{"author":{"username":"sys"},"body":"sys event","system":true}]}
		]`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.ListMRDiscussions("p/x", 42)
	require.NoError(t, err)
	require.Len(t, got, 1) // system note excluded
	require.Equal(t, "d1", got[0].ID)
	require.Equal(t, "bob", got[0].Author)
	require.Equal(t, "a.java", got[0].File)
	require.Equal(t, 42, got[0].Line)
}
