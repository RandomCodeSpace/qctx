package model_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

func TestBundleMarshal(t *testing.T) {
	b := model.Bundle{
		Meta: model.Meta{
			Tool: "qctx", Version: "0.1.0",
			ScannedAt:       time.Date(2026, 5, 19, 1, 15, 0, 0, time.UTC),
			SonarProjectKey: "p",
			SourceStatus:    map[string]string{"sonar": "ok"},
		},
		Sonar: model.SonarBundle{Issues: []model.Issue{{Key: "i1", Rule: "r", Severity: "MAJOR", Type: "BUG", File: "a.java", Message: "m", Status: "OPEN"}}},
	}
	var buf bytes.Buffer
	require.NoError(t, model.WriteJSON(&buf, b))
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Equal(t, "qctx", got["meta"].(map[string]any)["tool"])
	require.Len(t, got["sonar"].(map[string]any)["issues"], 1)
}
