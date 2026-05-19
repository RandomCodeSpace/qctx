package gitlab

import (
	"fmt"
	"strings"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

type rawMR struct {
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	WebURL       string `json:"web_url"`
	Draft        bool   `json:"draft"`
	ChangesCount string `json:"changes_count"`
}

// GetMR returns the MR meta for project + iid.
func (c *Client) GetMR(projectPath string, iid int) (model.MR, error) {
	enc := encode(projectPath)
	var r rawMR
	if err := c.GetJSON(fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d", enc, iid), nil, &r); err != nil {
		return model.MR{}, err
	}
	return model.MR{
		IID: r.IID, Title: r.Title, Description: r.Description, Author: r.Author.Username,
		SourceBranch: r.SourceBranch, TargetBranch: r.TargetBranch, WebURL: r.WebURL,
		Draft: r.Draft, ChangesCount: r.ChangesCount,
	}, nil
}

type rawChanges struct {
	Changes []struct {
		NewPath string `json:"new_path"`
		Diff    string `json:"diff"`
	} `json:"changes"`
}

// GetMRDiffSummary returns a summarized diff (files + add/delete counts).
func (c *Client) GetMRDiffSummary(projectPath string, iid int) (model.DiffSummary, error) {
	enc := encode(projectPath)
	var r rawChanges
	if err := c.GetJSON(fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d/changes", enc, iid), nil, &r); err != nil {
		return model.DiffSummary{}, err
	}
	out := model.DiffSummary{}
	for _, ch := range r.Changes {
		out.FilesChanged = append(out.FilesChanged, ch.NewPath)
		add, del := diffStat(ch.Diff)
		out.Additions += add
		out.Deletions += del
	}
	return out, nil
}

func diffStat(diff string) (add, del int) {
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "@@"):
			continue
		case strings.HasPrefix(line, "+"):
			add++
		case strings.HasPrefix(line, "-"):
			del++
		}
	}
	return
}

type rawDiscussion struct {
	ID    string `json:"id"`
	Notes []struct {
		Author struct {
			Username string `json:"username"`
		} `json:"author"`
		Body     string `json:"body"`
		Resolved bool   `json:"resolved"`
		System   bool   `json:"system"`
		Position *struct {
			NewPath string `json:"new_path"`
			NewLine int    `json:"new_line"`
		} `json:"position"`
	} `json:"notes"`
}

// ListMRDiscussions returns non-system discussions on the MR.
func (c *Client) ListMRDiscussions(projectPath string, iid int) ([]model.Discussion, error) {
	enc := encode(projectPath)
	path := fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d/discussions", enc, iid)
	var out []model.Discussion
	err := c.ListJSON(path, nil,
		func() any { return new([]rawDiscussion) },
		func(v any) error {
			page := v.(*[]rawDiscussion)
			for _, d := range *page {
				if len(d.Notes) == 0 || d.Notes[0].System {
					continue
				}
				n := d.Notes[0]
				disc := model.Discussion{
					ID: d.ID, Author: n.Author.Username, Body: n.Body, Resolved: n.Resolved,
				}
				if n.Position != nil {
					disc.File = n.Position.NewPath
					disc.Line = n.Position.NewLine
				}
				out = append(out, disc)
			}
			return nil
		},
	)
	return out, err
}

// encode returns the URL-encoded project path for /api/v4/projects/:id/... usage.
func encode(p string) string {
	return (MRURL{ProjectPath: p}).EncodedProjectPath()
}
