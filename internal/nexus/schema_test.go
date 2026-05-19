package nexus_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/nexus"
)

// sampleReport is shared by Task 23 (schema) and Task 24 (reader) tests.
const sampleReport = `{
  "scanId":"abc",
  "applicationId":"team-my-svc",
  "policyEvaluationResult":{
    "components":[
      {
        "componentIdentifier":{"format":"maven","coordinates":{"groupId":"org.apache.commons","artifactId":"commons-text","version":"1.9"}},
        "pathnames":["pom.xml"],
        "violations":[
          {
            "policyId":"p1","policyName":"Security-High","policyThreatCategory":"SECURITY","policyThreatLevel":8,
            "constraints":[
              {"reasons":[{"reason":"Found security vulnerability CVE-2022-42889 with severity 9.8"}]}
            ],
            "waived":false
          }
        ],
        "remediation":{"versionChanges":[{"toVersion":"1.10.0","type":"next-no-violations"}]}
      }
    ]
  }
}`

func TestParseReport(t *testing.T) {
	var r nexus.Report
	require.NoError(t, json.Unmarshal([]byte(sampleReport), &r))
	require.Equal(t, "team-my-svc", r.ApplicationID)
	require.Len(t, r.PolicyEvaluationResult.Components, 1)
	comp := r.PolicyEvaluationResult.Components[0]
	require.Equal(t, "commons-text", comp.ComponentIdentifier.Coordinates.ArtifactID)
	require.Equal(t, "1.9", comp.ComponentIdentifier.Coordinates.Version)
	require.Len(t, comp.Violations, 1)
	v := comp.Violations[0]
	require.Equal(t, "Security-High", v.PolicyName)
	require.Equal(t, 8, v.PolicyThreatLevel)
	require.False(t, v.Waived)
	require.Equal(t, "1.10.0", comp.Remediation.VersionChanges[0].ToVersion)
}
