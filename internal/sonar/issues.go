package sonar

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

type IssueQuery struct {
	ProjectKey  string
	Branch      string
	PullRequest string
	Severities  []string
	Types       []string
	Statuses    []string
	Resolved    bool // include resolved when true
	IssueKeys   []string
	PageSize    int
}

type issuesPage struct {
	Paging struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
		Total     int `json:"total"`
	} `json:"paging"`
	Issues []struct {
		Key       string `json:"key"`
		Rule      string `json:"rule"`
		Severity  string `json:"severity"`
		Type      string `json:"type"`
		Component string `json:"component"`
		Line      int    `json:"line"`
		TextRange *struct {
			EndLine int `json:"endLine"`
		} `json:"textRange,omitempty"`
		Message string   `json:"message"`
		Author  string   `json:"author"`
		Effort  string   `json:"effort"`
		Tags    []string `json:"tags"`
		Status  string   `json:"status"`
	} `json:"issues"`
}

// SearchIssues lists issues, paginating through all results.
func (c *Client) SearchIssues(q IssueQuery) ([]model.Issue, error) {
	page := 1
	ps := q.PageSize
	if ps <= 0 || ps > 500 {
		ps = 500
	}

	var out []model.Issue
	for {
		params := baseIssueParams(q)
		params["p"] = strconv.Itoa(page)
		params["ps"] = strconv.Itoa(ps)

		var resp issuesPage
		if err := c.GetJSON("/api/issues/search", params, &resp); err != nil {
			return nil, fmt.Errorf("issues page %d: %w", page, err)
		}
		for _, it := range resp.Issues {
			endLine := 0
			if it.TextRange != nil {
				endLine = it.TextRange.EndLine
			}
			out = append(out, model.Issue{
				Key: it.Key, Rule: it.Rule, Severity: it.Severity, Type: it.Type,
				File: stripProjectPrefix(it.Component), Line: it.Line, EndLine: endLine,
				Message: it.Message, Author: it.Author, Effort: it.Effort, Tags: it.Tags, Status: it.Status,
			})
		}
		if len(out) >= resp.Paging.Total || len(resp.Issues) == 0 || page*ps >= 10000 {
			break
		}
		page++
	}
	return out, nil
}

func baseIssueParams(q IssueQuery) map[string]string {
	p := map[string]string{}
	if q.ProjectKey != "" {
		p["componentKeys"] = q.ProjectKey
	}
	if q.Branch != "" {
		p["branch"] = q.Branch
	}
	if q.PullRequest != "" {
		p["pullRequest"] = q.PullRequest
	}
	if len(q.Severities) > 0 {
		p["severities"] = strings.Join(q.Severities, ",")
	}
	if len(q.Types) > 0 {
		p["types"] = strings.Join(q.Types, ",")
	}
	if len(q.Statuses) > 0 {
		p["statuses"] = strings.Join(q.Statuses, ",")
	}
	if len(q.IssueKeys) > 0 {
		p["issues"] = strings.Join(q.IssueKeys, ",")
	}
	if q.Resolved {
		p["resolved"] = "true"
	} else {
		p["resolved"] = "false"
	}
	return p
}

// stripProjectPrefix turns "myproj:src/A.java" into "src/A.java".
func stripProjectPrefix(component string) string {
	if i := strings.IndexByte(component, ':'); i >= 0 {
		return component[i+1:]
	}
	return component
}
