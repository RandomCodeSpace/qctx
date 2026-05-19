package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

func TestIssueRoundTrip(t *testing.T) {
	in := model.Issue{Key: "k1", Rule: "java:S1", Severity: "MAJOR", Type: "BUG", File: "a.java", Line: 10, Message: "x", Status: "OPEN"}
	b, err := json.Marshal(in)
	require.NoError(t, err)
	var out model.Issue
	require.NoError(t, json.Unmarshal(b, &out))
	require.Equal(t, in, out)
}

func TestQualityGateZero(t *testing.T) {
	var qg model.QualityGate
	require.Equal(t, "", qg.Status)
	require.Empty(t, qg.Conditions)
}
