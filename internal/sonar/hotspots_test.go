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

func TestSearchHotspots(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "p", r.URL.Query().Get("projectKey"))
		_, _ = w.Write([]byte(`{
		  "paging":{"pageIndex":1,"pageSize":500,"total":1},
		  "hotspots":[{"key":"h1","ruleKey":"java:S2076","vulnerabilityProbability":"HIGH","status":"TO_REVIEW","component":"p:src/A.java","line":15,"message":"m"}]
		}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})
	got, err := c.SearchHotspots(sonar.HotspotQuery{ProjectKey: "p"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "h1", got[0].Key)
	require.Equal(t, "src/A.java", got[0].File)
}
