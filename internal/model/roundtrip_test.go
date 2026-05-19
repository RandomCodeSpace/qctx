package model_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

func TestJSONLLinesAreParseable(t *testing.T) {
	var buf bytes.Buffer
	w := model.NewJSONLWriter(&buf)
	require.NoError(t, w.WriteIssue(model.Issue{Key: "k1", Rule: "r", Severity: "MAJOR", Type: "BUG", File: "a.java", Line: 10, Message: "m", Status: "OPEN"}))
	require.NoError(t, w.WriteHotspot(model.Hotspot{Key: "h1", Rule: "r", VulnerabilityProbability: "HIGH", Status: "TO_REVIEW", File: "b.java", Message: "m"}))
	require.NoError(t, w.Flush())

	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	var seen []string
	for scanner.Scan() {
		var v struct {
			Type string `json:"type"`
		}
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &v))
		seen = append(seen, v.Type)
	}
	require.Equal(t, []string{"sonar.issue", "sonar.hotspot"}, seen)
}
