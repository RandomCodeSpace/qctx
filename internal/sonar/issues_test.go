package sonar_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

// strconv-free helper to avoid extra import noise in tests.
func fmtSscanf(s, f string, a ...any) (int, error) { return fmt.Sscanf(s, f, a...) }

func TestSearchIssuesPaginates(t *testing.T) {
	var page int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ = strconvAtoiSafe(r.URL.Query().Get("p"))
		w.Header().Set("Content-Type", "application/json")
		var body any
		switch page {
		case 1:
			body = map[string]any{
				"paging": map[string]any{"pageIndex": 1, "pageSize": 1, "total": 2},
				"issues": []map[string]any{{"key": "i1", "rule": "java:S1", "severity": "MAJOR", "type": "BUG", "component": "p:src/A.java", "line": 1, "message": "m1", "status": "OPEN"}},
			}
		case 2:
			body = map[string]any{
				"paging": map[string]any{"pageIndex": 2, "pageSize": 1, "total": 2},
				"issues": []map[string]any{{"key": "i2", "rule": "java:S2", "severity": "MINOR", "type": "CODE_SMELL", "component": "p:src/B.java", "line": 2, "message": "m2", "status": "OPEN"}},
			}
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	got, err := c.SearchIssues(sonar.IssueQuery{ProjectKey: "p", PageSize: 1})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "i1", got[0].Key)
	require.Equal(t, "src/A.java", got[0].File)
	require.Equal(t, "i2", got[1].Key)
}

func TestSearchIssuesAppliesFilters(t *testing.T) {
	var gotQ url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":500,"total":0},"issues":[]}`))
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	_, err := c.SearchIssues(sonar.IssueQuery{
		ProjectKey: "p", Branch: "feat-x", PullRequest: "42",
		Severities: []string{"MAJOR", "CRITICAL"},
		Types:      []string{"BUG"},
		Resolved:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "p", gotQ.Get("componentKeys"))
	require.Equal(t, "feat-x", gotQ.Get("branch"))
	require.Equal(t, "42", gotQ.Get("pullRequest"))
	require.Equal(t, "MAJOR,CRITICAL", gotQ.Get("severities"))
	require.Equal(t, "BUG", gotQ.Get("types"))
	require.Equal(t, "false", gotQ.Get("resolved"))
}

func strconvAtoiSafe(s string) (int, error) {
	if s == "" {
		return 1, nil
	}
	var n int
	_, err := fmtSscanf(s, "%d", &n)
	return n, err
}
