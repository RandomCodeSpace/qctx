package sonar_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

func TestGetMeasures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "p", r.URL.Query().Get("component"))
		require.Contains(t, r.URL.Query().Get("metricKeys"), "coverage")
		_, _ = w.Write([]byte(`{"component":{"measures":[{"metric":"coverage","value":"78.4"},{"metric":"bugs","value":"3"}]}}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.GetMeasures("p", "", "", []string{"coverage", "bugs"})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "coverage", got[0].Metric)
	require.InDelta(t, 78.4, got[0].Value, 1e-6)
}
