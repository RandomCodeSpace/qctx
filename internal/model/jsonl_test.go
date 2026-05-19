package model_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

func TestJSONLWriterEmitsTypedLines(t *testing.T) {
	var buf bytes.Buffer
	w := model.NewJSONLWriter(&buf)

	require.NoError(t, w.WriteMeta(model.Meta{
		Tool: "qctx", Version: "0.1.0",
		ScannedAt:    time.Date(2026, 5, 19, 1, 15, 0, 0, time.UTC),
		SourceStatus: map[string]string{"sonar": "ok"},
	}))
	require.NoError(t, w.WriteIssue(model.Issue{Key: "i1", Rule: "r", Severity: "MAJOR", Type: "BUG", File: "a.java", Message: "m", Status: "OPEN"}))
	require.NoError(t, w.WriteHotspot(model.Hotspot{Key: "h1", Rule: "r2", VulnerabilityProbability: "HIGH", Status: "TO_REVIEW", File: "b.java", Message: "m"}))
	require.NoError(t, w.WriteMeasure(model.Measure{Metric: "coverage", Value: 78.4}))
	require.NoError(t, w.WriteQualityGate(model.QualityGate{Status: "FAILED"}))
	require.NoError(t, w.WriteNexusViolation(model.Violation{Component: "x:y:1", Policy: "P", ThreatLevel: 5, Status: "open"}))
	require.NoError(t, w.WriteMR(model.MR{IID: 42, Title: "t", SourceBranch: "s", TargetBranch: "main"}))
	require.NoError(t, w.WriteDiffSummary(model.DiffSummary{FilesChanged: []string{"a.java"}, Additions: 1}))
	require.NoError(t, w.WriteDiscussion(model.Discussion{ID: "d1", Author: "a", Body: "b"}))
	require.NoError(t, w.WritePipeline(model.Pipeline{ID: 1001, Status: "failed", Ref: "feat", SHA: "abc"}))
	require.NoError(t, w.WriteJob(model.Job{Name: "test", Status: "failed", Stage: "test"}))
	require.NoError(t, w.WriteError(model.SourceError{Source: "sonar", Message: "401"}))
	require.NoError(t, w.Flush())

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 12)

	wantTypes := []string{
		"meta", "sonar.issue", "sonar.hotspot", "sonar.measure", "sonar.quality_gate",
		"nexus.violation", "gitlab.mr", "gitlab.mr.diff_summary", "gitlab.mr.discussion",
		"gitlab.pipeline", "gitlab.job", "error",
	}
	for i, want := range wantTypes {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(lines[i]), &m))
		require.Equal(t, want, m["type"], "line %d", i)
	}
}

func TestJSONLEachLineIsJSON(t *testing.T) {
	var buf bytes.Buffer
	w := model.NewJSONLWriter(&buf)
	require.NoError(t, w.WriteIssue(model.Issue{Key: "x"}))
	require.NoError(t, w.Flush())
	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		var v map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &v))
	}
}
