package httpclient_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

func TestDoSendsAuthAndCustomHeaders(t *testing.T) {
	var gotAuth, gotXA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotXA = r.Header.Get("X-A")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c, err := httpclient.New(httpclient.Options{
		ExtraHeaders: map[string]string{"X-A": "1"},
		Timeout:      2 * time.Second,
	})
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.Header.Set("Authorization", "Bearer t")
	resp, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "Bearer t", gotAuth)
	require.Equal(t, "1", gotXA)
}

func TestRetriesOn5xx(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c, err := httpclient.New(httpclient.Options{MaxRetries: 4, RetryWait: 10 * time.Millisecond, Timeout: 2 * time.Second})
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, int32(3), atomic.LoadInt32(&hits))
}

func TestDoesNotRetryOn4xx(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(404)
	}))
	defer srv.Close()

	c, err := httpclient.New(httpclient.Options{MaxRetries: 3, RetryWait: 10 * time.Millisecond, Timeout: 2 * time.Second})
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

func TestInsecureTLSOption(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()

	c, err := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	_, err = c.Do(req)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "x509") || strings.Contains(err.Error(), "certificate"))

	c2, err := httpclient.New(httpclient.Options{Insecure: true, Timeout: 2 * time.Second})
	require.NoError(t, err)
	resp, err := c2.Do(req.Clone(req.Context()))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}
