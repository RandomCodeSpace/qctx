package sonar_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

func TestGetRuleHTMLCachesPerKey(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		require.Equal(t, "java:S1234", r.URL.Query().Get("key"))
		_, _ = w.Write([]byte(`{"rule":{"key":"java:S1234","htmlDesc":"<h2>Why</h2>"}}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	for i := 0; i < 5; i++ {
		got, err := c.GetRuleHTML("java:S1234")
		require.NoError(t, err)
		require.Contains(t, got, "Why")
	}
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}
