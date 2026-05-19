// Package nexus parses Sonatype Nexus IQ policy evaluation reports.
package nexus

// Report is the top-level shape of a Nexus IQ scan results JSON.
type Report struct {
	ScanID                 string                 `json:"scanId"`
	ApplicationID          string                 `json:"applicationId"`
	PolicyEvaluationResult PolicyEvaluationResult `json:"policyEvaluationResult"`
}

// PolicyEvaluationResult wraps the components scanned.
type PolicyEvaluationResult struct {
	Components []Component `json:"components"`
}

// Component is one scanned artifact + its violations.
type Component struct {
	ComponentIdentifier ComponentIdentifier `json:"componentIdentifier"`
	Pathnames           []string            `json:"pathnames"`
	Violations          []Violation         `json:"violations"`
	Remediation         Remediation         `json:"remediation"`
}

// ComponentIdentifier names the component (maven, npm, generic).
type ComponentIdentifier struct {
	Format      string      `json:"format"`
	Coordinates Coordinates `json:"coordinates"`
}

// Coordinates is the format-specific identity (groupId/artifactId/version, or name/version).
type Coordinates struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version"`
	PackageID  string `json:"packageId"`
	Name       string `json:"name"`
}

// Violation is one policy rule the component breaks.
type Violation struct {
	PolicyID             string       `json:"policyId"`
	PolicyName           string       `json:"policyName"`
	PolicyThreatCategory string       `json:"policyThreatCategory"`
	PolicyThreatLevel    int          `json:"policyThreatLevel"`
	Constraints          []Constraint `json:"constraints"`
	Waived               bool         `json:"waived"`
}

// Constraint groups Reasons that triggered the Violation.
type Constraint struct {
	Reasons []Reason `json:"reasons"`
}

// Reason is a single human-readable reason string.
type Reason struct {
	Reason string `json:"reason"`
}

// Remediation suggests upgrade paths.
type Remediation struct {
	VersionChanges []VersionChange `json:"versionChanges"`
}

// VersionChange is one suggested upgrade target.
type VersionChange struct {
	ToVersion string `json:"toVersion"`
	Type      string `json:"type"`
}
