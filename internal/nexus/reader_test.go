package nexus_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/nexus"
)

func TestReadReportNormalizes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "iq.json")
	require.NoError(t, os.WriteFile(path, []byte(sampleReport), 0o600))

	got, app, err := nexus.ReadReport(path)
	require.NoError(t, err)
	require.Equal(t, "team-my-svc", app)
	require.Len(t, got, 1)
	v := got[0]
	require.Equal(t, "org.apache.commons:commons-text:1.9", v.Component)
	require.Equal(t, "pom.xml", v.Manifest)
	require.Equal(t, "Security-High", v.Policy)
	require.Equal(t, 8, v.ThreatLevel)
	require.Equal(t, []string{"CVE-2022-42889"}, v.CVEs)
	require.Equal(t, "1.10.0", v.FixVersion)
	require.Equal(t, "open", v.Status)
}

func TestReadReportWaivedMarkedClosed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "iq.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
	  "applicationId":"a",
	  "policyEvaluationResult":{"components":[{
	    "componentIdentifier":{"format":"maven","coordinates":{"groupId":"g","artifactId":"a","version":"1.0"}},
	    "pathnames":["pom.xml"],
	    "violations":[{"policyId":"p","policyName":"X","policyThreatLevel":3,"waived":true}]
	  }]}
	}`), 0o600))
	got, _, err := nexus.ReadReport(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "waived", got[0].Status)
}

func TestReadReportMissingFile(t *testing.T) {
	_, _, err := nexus.ReadReport("/no/such/file.json")
	require.Error(t, err)
}
