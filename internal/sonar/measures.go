package sonar

import (
	"strconv"
	"strings"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

// DefaultMetrics returned when none specified.
var DefaultMetrics = []string{
	"coverage", "new_coverage",
	"bugs", "new_bugs",
	"vulnerabilities", "new_vulnerabilities",
	"code_smells", "new_code_smells",
	"security_hotspots", "new_security_hotspots",
	"duplicated_lines_density",
}

type measuresResp struct {
	Component struct {
		Measures []struct {
			Metric string `json:"metric"`
			Value  string `json:"value"`
		} `json:"measures"`
	} `json:"component"`
}

// GetMeasures fetches the requested metrics. If metrics is empty, uses DefaultMetrics.
func (c *Client) GetMeasures(projectKey, branch, pullRequest string, metrics []string) ([]model.Measure, error) {
	if len(metrics) == 0 {
		metrics = DefaultMetrics
	}
	params := map[string]string{
		"component":  projectKey,
		"metricKeys": strings.Join(metrics, ","),
	}
	if branch != "" {
		params["branch"] = branch
	}
	if pullRequest != "" {
		params["pullRequest"] = pullRequest
	}
	var resp measuresResp
	if err := c.GetJSON("/api/measures/component", params, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Measure, 0, len(resp.Component.Measures))
	for _, m := range resp.Component.Measures {
		v, _ := strconv.ParseFloat(m.Value, 64)
		out = append(out, model.Measure{Metric: m.Metric, Value: v})
	}
	return out, nil
}
