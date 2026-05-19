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

func TestGetQualityGate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "p", r.URL.Query().Get("projectKey"))
		_, _ = w.Write([]byte(`{"projectStatus":{"status":"ERROR","conditions":[
		  {"metricKey":"new_coverage","comparator":"LT","errorThreshold":"80","actualValue":"72.1","status":"ERROR"}
		]}}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.GetQualityGate("p", "", "")
	require.NoError(t, err)
	require.Equal(t, "FAILED", got.Status)
	require.Len(t, got.Conditions, 1)
	require.Equal(t, "new_coverage", got.Conditions[0].Metric)
	require.Equal(t, "LT", got.Conditions[0].Op)
	require.Equal(t, "72.1", got.Conditions[0].Actual)
	require.Equal(t, "ERROR", got.Conditions[0].Status)
}
