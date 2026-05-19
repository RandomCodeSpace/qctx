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

func TestGetJobTraceTail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rh := r.Header.Get("Range"); strings.HasPrefix(rh, "bytes=-") {
			body := "tail-content"
			w.WriteHeader(206)
			_, _ = w.Write([]byte(body))
			return
		}
		_, _ = w.Write([]byte("full-trace-content"))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	got, err := c.GetJobTraceTail("p/x", 42, 100)
	require.NoError(t, err)
	require.Equal(t, "tail-content", got)
}

func TestFailureExcerpt(t *testing.T) {
	trace := strings.Join([]string{
		"go build ./...",
		"all good",
		"=== RUN TestFoo",
		"    foo_test.go:12: expected 5 got 6",
		"--- FAIL: TestFoo (0.01s)",
		"FAIL github.com/x/y/foo",
		"FAIL",
		"exit code: 1",
	}, "\n")
	excerpt := gitlab.FailureExcerpt(trace, 6)
	require.Contains(t, excerpt, "--- FAIL: TestFoo")
	require.Equal(t, 6, strings.Count(excerpt, "\n")+1)
}

func TestFailureExcerptEmpty(t *testing.T) {
	require.Equal(t, "", gitlab.FailureExcerpt("", 5))
}
