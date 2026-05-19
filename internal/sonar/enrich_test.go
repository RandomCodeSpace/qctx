package sonar_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/model"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

func TestEnrichIssueDescriptionsCachesPerRule(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		key := r.URL.Query().Get("key")
		_, _ = w.Write([]byte(`{"rule":{"key":"` + key + `","htmlDesc":"<p>desc for ` + key + `</p>"}}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	issues := []model.Issue{
		{Key: "i1", Rule: "java:S1"},
		{Key: "i2", Rule: "java:S1"}, // same rule, expect cache hit
		{Key: "i3", Rule: "java:S2"},
		{Key: "i4", Rule: ""}, // empty rule, skip
	}
	c.EnrichIssueDescriptions(issues)

	require.Equal(t, "<p>desc for java:S1</p>", issues[0].RuleDescHTML)
	require.Equal(t, "<p>desc for java:S1</p>", issues[1].RuleDescHTML)
	require.Equal(t, "<p>desc for java:S2</p>", issues[2].RuleDescHTML)
	require.Equal(t, "", issues[3].RuleDescHTML)
	require.Equal(t, int32(2), atomic.LoadInt32(&hits)) // S1 + S2 only
}

func TestEnrichIssueDescriptionsRuleFetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second, MaxRetries: 1, RetryWait: 5 * time.Millisecond})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	issues := []model.Issue{{Key: "i1", Rule: "java:S1"}}
	c.EnrichIssueDescriptions(issues)
	require.Equal(t, "", issues[0].RuleDescHTML) // failed fetch leaves it empty
}

func TestEnrichHotspotDescriptionsCachesPerRule(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		key := r.URL.Query().Get("key")
		_, _ = w.Write([]byte(`{"rule":{"key":"` + key + `","htmlDesc":"H:` + key + `"}}`))
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	hs := []model.Hotspot{
		{Key: "h1", Rule: "java:S2076"},
		{Key: "h2", Rule: "java:S2076"}, // dup
		{Key: "h3", Rule: ""},           // skip
	}
	c.EnrichHotspotDescriptions(hs)
	require.Equal(t, "H:java:S2076", hs[0].RuleDescHTML)
	require.Equal(t, "H:java:S2076", hs[1].RuleDescHTML)
	require.Equal(t, "", hs[2].RuleDescHTML)
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

func TestEnrichHotspotDescriptionsErrorSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second, MaxRetries: 1, RetryWait: 5 * time.Millisecond})
	c, _ := sonar.New(sonar.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	hs := []model.Hotspot{{Key: "h1", Rule: "java:S2076"}}
	c.EnrichHotspotDescriptions(hs)
	require.Equal(t, "", hs[0].RuleDescHTML)
}
