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

func TestClientSendsPrivateToken(t *testing.T) {
	var gotTok string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTok = r.Header.Get("PRIVATE-TOKEN")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, err := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "glpat-abc", HTTP: hc})
	require.NoError(t, err)
	require.NoError(t, c.GetJSON("/api/v4/x", nil, new(map[string]any)))
	require.Equal(t, "glpat-abc", gotTok)
}

func TestClientErrorsOnEmptyBaseURL(t *testing.T) {
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	_, err := gitlab.New(gitlab.Options{BaseURL: "", Token: "x", HTTP: hc})
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "base url")
}

func TestClientErrorsOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	err := c.GetJSON("/api/v4/x", nil, new(map[string]any))
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")
}
