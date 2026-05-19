package sonar

import (
	"fmt"
	"strconv"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

// HotspotQuery configures a SearchHotspots call.
type HotspotQuery struct {
	ProjectKey  string
	Branch      string
	PullRequest string
	Status      string // TO_REVIEW | REVIEWED
	PageSize    int
}

type hotspotsPage struct {
	Paging struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
		Total     int `json:"total"`
	} `json:"paging"`
	Hotspots []struct {
		Key                      string `json:"key"`
		RuleKey                  string `json:"ruleKey"`
		VulnerabilityProbability string `json:"vulnerabilityProbability"`
		Status                   string `json:"status"`
		Component                string `json:"component"`
		Line                     int    `json:"line"`
		Message                  string `json:"message"`
	} `json:"hotspots"`
}

// SearchHotspots paginates Sonar's /api/hotspots/search and returns normalized model.Hotspot slice.
func (c *Client) SearchHotspots(q HotspotQuery) ([]model.Hotspot, error) {
	page := 1
	ps := q.PageSize
	if ps <= 0 || ps > 500 {
		ps = 500
	}
	var out []model.Hotspot
	for {
		params := map[string]string{"projectKey": q.ProjectKey}
		if q.Branch != "" {
			params["branch"] = q.Branch
		}
		if q.PullRequest != "" {
			params["pullRequest"] = q.PullRequest
		}
		if q.Status != "" {
			params["status"] = q.Status
		}
		params["p"] = strconv.Itoa(page)
		params["ps"] = strconv.Itoa(ps)

		var resp hotspotsPage
		if err := c.GetJSON("/api/hotspots/search", params, &resp); err != nil {
			return nil, fmt.Errorf("hotspots page %d: %w", page, err)
		}
		for _, h := range resp.Hotspots {
			out = append(out, model.Hotspot{
				Key:                      h.Key,
				Rule:                     h.RuleKey,
				VulnerabilityProbability: h.VulnerabilityProbability,
				Status:                   h.Status,
				File:                     stripProjectPrefix(h.Component),
				Line:                     h.Line,
				Message:                  h.Message,
			})
		}
		if len(out) >= resp.Paging.Total || len(resp.Hotspots) == 0 {
			break
		}
		page++
	}
	return out, nil
}
