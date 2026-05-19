package gitlab

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

type rawPipeline struct {
	ID        int         `json:"id"`
	Status    string      `json:"status"`
	Ref       string      `json:"ref"`
	SHA       string      `json:"sha"`
	WebURL    string      `json:"web_url"`
	CreatedAt string      `json:"created_at"`
	Duration  json.Number `json:"duration"`
}

// ListMRPipelines lists all pipelines belonging to an MR.
func (c *Client) ListMRPipelines(projectPath string, iid int) ([]model.Pipeline, error) {
	enc := encode(projectPath)
	var raw []rawPipeline
	if err := c.GetJSON(fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d/pipelines", enc, iid), nil, &raw); err != nil {
		return nil, err
	}
	out := make([]model.Pipeline, 0, len(raw))
	for _, p := range raw {
		dur, _ := p.Duration.Int64()
		out = append(out, model.Pipeline{
			ID: p.ID, Status: p.Status, Ref: p.Ref, SHA: p.SHA, WebURL: p.WebURL,
			CreatedAt: p.CreatedAt, Duration: int(dur),
		})
	}
	return out, nil
}

// LatestMRPipeline returns the most recent non-skipped pipeline.
func (c *Client) LatestMRPipeline(projectPath string, iid int) (model.Pipeline, error) {
	all, err := c.ListMRPipelines(projectPath, iid)
	if err != nil {
		return model.Pipeline{}, err
	}
	filtered := all[:0]
	for _, p := range all {
		if p.Status != "skipped" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return model.Pipeline{}, fmt.Errorf("no non-skipped pipelines for MR !%d", iid)
	}
	sort.Slice(filtered, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, filtered[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, filtered[j].CreatedAt)
		return ti.After(tj)
	})
	return filtered[0], nil
}

type rawJob struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Status   string      `json:"status"`
	Stage    string      `json:"stage"`
	Duration json.Number `json:"duration"`
	WebURL   string      `json:"web_url"`
}

// ListPipelineJobs returns all jobs in the pipeline, paginated.
func (c *Client) ListPipelineJobs(projectPath string, pipelineID int) ([]model.Job, error) {
	enc := encode(projectPath)
	path := fmt.Sprintf("/api/v4/projects/%s/pipelines/%d/jobs", enc, pipelineID)
	var out []model.Job
	err := c.ListJSON(path, map[string]string{"include_retried": "false"},
		func() any { return new([]rawJob) },
		func(v any) error {
			page := v.(*[]rawJob)
			for _, j := range *page {
				dur, _ := j.Duration.Int64()
				out = append(out, model.Job{
					ID: j.ID, Name: j.Name, Status: j.Status, Stage: j.Stage,
					Duration: int(dur), WebURL: j.WebURL,
				})
			}
			return nil
		},
	)
	return out, err
}
