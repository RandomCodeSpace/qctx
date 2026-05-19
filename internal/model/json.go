package model

import (
	"encoding/json"
	"io"
)

// SonarBundle groups Sonar data for live JSON output.
type SonarBundle struct {
	Issues      []Issue      `json:"issues,omitempty"`
	Hotspots    []Hotspot    `json:"hotspots,omitempty"`
	Measures    []Measure    `json:"measures,omitempty"`
	QualityGate *QualityGate `json:"quality_gate,omitempty"`
}

// NexusBundle groups Nexus data for live JSON output.
type NexusBundle struct {
	Violations []Violation `json:"violations,omitempty"`
}

// GitLabBundle groups GitLab data for live JSON output.
type GitLabBundle struct {
	MR          *MR          `json:"mr,omitempty"`
	DiffSummary *DiffSummary `json:"diff_summary,omitempty"`
	Discussions []Discussion `json:"discussions,omitempty"`
	Pipeline    *Pipeline    `json:"pipeline,omitempty"`
	JobsFailed  []Job        `json:"jobs_failed,omitempty"`
}

// Bundle is the live-mode root document.
type Bundle struct {
	Meta   Meta          `json:"meta"`
	Sonar  SonarBundle   `json:"sonar,omitempty"`
	Nexus  NexusBundle   `json:"nexus,omitempty"`
	GitLab GitLabBundle  `json:"gitlab,omitempty"`
	Errors []SourceError `json:"errors,omitempty"`
}

// WriteJSON encodes the bundle as pretty (indent=2) JSON.
func WriteJSON(w io.Writer, b Bundle) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}
