package gitlab_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/gitlab"
	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

func TestDiscoverSonarProjectKey_FromMavenInvocation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/p%2Fx/pipelines/1001/jobs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Next-Page", "")
		_, _ = w.Write([]byte(`[
		  {"id":11,"name":"sonar","status":"success","stage":"analysis","duration":30,"web_url":"u11"},
		  {"id":12,"name":"test","status":"success","stage":"test","duration":127,"web_url":"u12"}
		]`))
	})
	mux.HandleFunc("/api/v4/projects/p%2Fx/jobs/11/trace", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, `mvn sonar:sonar -Dsonar.host.url=https://sonar.example.com -Dsonar.projectKey=team_my-svc -Dsonar.login=$T`)
	})
	mux.HandleFunc("/api/v4/projects/p%2Fx/jobs/12/trace", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	key, err := c.DiscoverSonarProjectKey("p/x", 1001)
	require.NoError(t, err)
	require.Equal(t, "team_my-svc", key)
}

func TestDiscoverSonarProjectKey_FromPropertiesEcho(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/p%2Fx/pipelines/1001/jobs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Next-Page", "")
		_, _ = w.Write([]byte(`[{"id":11,"name":"sonar","status":"success","stage":"a","duration":10,"web_url":"u"}]`))
	})
	mux.HandleFunc("/api/v4/projects/p%2Fx/jobs/11/trace", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, strings.Join([]string{
			"+ cat sonar-project.properties",
			"sonar.projectKey=my-svc",
			"sonar.host.url=https://sonar.example.com",
		}, "\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	key, err := c.DiscoverSonarProjectKey("p/x", 1001)
	require.NoError(t, err)
	require.Equal(t, "my-svc", key)
}

func TestDiscoverSonarProjectKey_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/p%2Fx/pipelines/1001/jobs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Next-Page", "")
		_, _ = w.Write([]byte(`[{"id":11,"name":"test","status":"success","stage":"t","duration":10,"web_url":"u"}]`))
	})
	mux.HandleFunc("/api/v4/projects/p%2Fx/jobs/11/trace", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "no sonar here, just test output")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	hc, _ := httpclient.New(httpclient.Options{Timeout: 2 * time.Second})
	c, _ := gitlab.New(gitlab.Options{BaseURL: srv.URL, Token: "t", HTTP: hc})

	_, err := c.DiscoverSonarProjectKey("p/x", 1001)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not discover")
}
