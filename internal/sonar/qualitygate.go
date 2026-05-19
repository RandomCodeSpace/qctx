package sonar

import "github.com/RandomCodeSpace/qctx/internal/model"

type qgResp struct {
	ProjectStatus struct {
		Status     string `json:"status"`
		Conditions []struct {
			MetricKey      string `json:"metricKey"`
			Comparator     string `json:"comparator"`
			ErrorThreshold string `json:"errorThreshold"`
			ActualValue    string `json:"actualValue"`
			Status         string `json:"status"`
		} `json:"conditions"`
	} `json:"projectStatus"`
}

// GetQualityGate returns the gate verdict for the project/branch/PR.
func (c *Client) GetQualityGate(projectKey, branch, pullRequest string) (model.QualityGate, error) {
	params := map[string]string{"projectKey": projectKey}
	if branch != "" {
		params["branch"] = branch
	}
	if pullRequest != "" {
		params["pullRequest"] = pullRequest
	}
	var r qgResp
	if err := c.GetJSON("/api/qualitygates/project_status", params, &r); err != nil {
		return model.QualityGate{}, err
	}
	status := r.ProjectStatus.Status
	switch status {
	case "OK":
		status = "PASSED"
	case "ERROR":
		status = "FAILED"
	case "WARN":
		status = "WARNING"
	}
	out := model.QualityGate{Status: status}
	for _, cnd := range r.ProjectStatus.Conditions {
		out.Conditions = append(out.Conditions, model.GateCondition{
			Metric: cnd.MetricKey, Op: cnd.Comparator,
			Threshold: cnd.ErrorThreshold, Actual: cnd.ActualValue, Status: cnd.Status,
		})
	}
	return out, nil
}
