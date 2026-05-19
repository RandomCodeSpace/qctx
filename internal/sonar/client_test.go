package sonar_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

func TestClientUsesBearerByDefault(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, err := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "abc123", HTTP: hc})
	require.NoError(t, err)

	var v map[string]any
	require.NoError(t, c.GetJSON("/api/ping", nil, &v))
	require.Equal(t, "Bearer abc123", gotAuth)
}

func TestClientFallsBackToBasic(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "tok", BasicAuth: true, HTTP: hc})
	require.NoError(t, c.GetJSON("/api/ping", nil, new(map[string]any)))

	expect := "Basic " + base64.StdEncoding.EncodeToString([]byte("tok:"))
	require.Equal(t, expect, gotAuth)
}

func TestClientErrorsOnEmptyBaseURL(t *testing.T) {
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	_, err := sonar.New(sonar.Options{BaseURL: "", Token: "t", HTTP: hc})
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "base url")
}

func TestGetJSONPassesQueryString(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	require.NoError(t, c.GetJSON("/api/x", map[string]string{"a": "1", "b": "two words"}, new(map[string]any)))
	require.Contains(t, gotQuery, "a=1")
	require.Contains(t, gotQuery, "b=two+words")
}

func TestGetJSON4xxReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"errors":[{"msg":"bad token"}]}`))
	}))
	defer srv.Close()

	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	err := c.GetJSON("/api/ping", nil, new(map[string]any))
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
}
