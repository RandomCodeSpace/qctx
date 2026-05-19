package nexus

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

var cveRE = regexp.MustCompile(`CVE-\d{4}-\d{4,7}`)

// ReadReport parses a Nexus IQ scan results JSON and returns normalized violations
// plus the application id from the report.
func ReadReport(path string) ([]model.Violation, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("nexus: read %q: %w", path, err)
	}
	var r Report
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, "", fmt.Errorf("nexus: parse %q: %w", path, err)
	}

	var out []model.Violation
	for _, comp := range r.PolicyEvaluationResult.Components {
		coord := componentLabel(comp.ComponentIdentifier)
		manifest := ""
		if len(comp.Pathnames) > 0 {
			manifest = comp.Pathnames[0]
		}
		fixVersion := ""
		if len(comp.Remediation.VersionChanges) > 0 {
			fixVersion = comp.Remediation.VersionChanges[0].ToVersion
		}
		for _, v := range comp.Violations {
			cves := extractCVEs(v)
			summary := extractSummary(v)
			status := "open"
			if v.Waived {
				status = "waived"
			}
			out = append(out, model.Violation{
				Component:   coord,
				Manifest:    manifest,
				Policy:      v.PolicyName,
				ThreatLevel: v.PolicyThreatLevel,
				CVEs:        cves,
				Summary:     summary,
				FixVersion:  fixVersion,
				Status:      status,
			})
		}
	}
	return out, r.ApplicationID, nil
}

func componentLabel(c ComponentIdentifier) string {
	switch c.Format {
	case "maven":
		return fmt.Sprintf("%s:%s:%s", c.Coordinates.GroupID, c.Coordinates.ArtifactID, c.Coordinates.Version)
	case "npm":
		if c.Coordinates.PackageID != "" {
			return fmt.Sprintf("%s@%s", c.Coordinates.PackageID, c.Coordinates.Version)
		}
		return fmt.Sprintf("%s@%s", c.Coordinates.Name, c.Coordinates.Version)
	default:
		if c.Coordinates.Name != "" {
			return fmt.Sprintf("%s@%s", c.Coordinates.Name, c.Coordinates.Version)
		}
		return fmt.Sprintf("%s:%s:%s", c.Coordinates.GroupID, c.Coordinates.ArtifactID, c.Coordinates.Version)
	}
}

func extractCVEs(v Violation) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, c := range v.Constraints {
		for _, r := range c.Reasons {
			for _, m := range cveRE.FindAllString(r.Reason, -1) {
				if _, ok := seen[m]; ok {
					continue
				}
				seen[m] = struct{}{}
				out = append(out, m)
			}
		}
	}
	return out
}

func extractSummary(v Violation) string {
	for _, c := range v.Constraints {
		for _, r := range c.Reasons {
			s := strings.TrimSpace(r.Reason)
			if s != "" {
				return s
			}
		}
	}
	return ""
}
