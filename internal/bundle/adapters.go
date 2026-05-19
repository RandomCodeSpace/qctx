package bundle

import (
	"github.com/RandomCodeSpace/qctx/internal/gitlab"
	"github.com/RandomCodeSpace/qctx/internal/model"
	"github.com/RandomCodeSpace/qctx/internal/nexus"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
)

// ---- Sonar adapter ----

// SonarAdapterDeps configures a Sonar source adapter.
type SonarAdapterDeps struct {
	Client      *sonar.Client
	ProjectKey  string
	Branch      string
	PullRequest string
	Severities  []string
	Types       []string
	Resolved    bool
	IssueKeys   []string
}

type sonarAdapter struct {
	d SonarAdapterDeps
}

// NewSonarAdapter builds a SonarSource bound to deps.
func NewSonarAdapter(d SonarAdapterDeps) SonarSource { return &sonarAdapter{d: d} }

func (a *sonarAdapter) Issues() ([]model.Issue, error) {
	if a.d.Client == nil {
		return nil, nil
	}
	iss, err := a.d.Client.SearchIssues(sonar.IssueQuery{
		ProjectKey: a.d.ProjectKey, Branch: a.d.Branch, PullRequest: a.d.PullRequest,
		Severities: a.d.Severities, Types: a.d.Types,
		IssueKeys: a.d.IssueKeys, Resolved: a.d.Resolved,
	})
	if err != nil {
		return nil, err
	}
	a.d.Client.EnrichIssueDescriptions(iss)
	return iss, nil
}

func (a *sonarAdapter) Hotspots() ([]model.Hotspot, error) {
	if a.d.Client == nil {
		return nil, nil
	}
	hs, err := a.d.Client.SearchHotspots(sonar.HotspotQuery{
		ProjectKey: a.d.ProjectKey, Branch: a.d.Branch, PullRequest: a.d.PullRequest,
	})
	if err != nil {
		return nil, err
	}
	a.d.Client.EnrichHotspotDescriptions(hs)
	return hs, nil
}

func (a *sonarAdapter) Measures() ([]model.Measure, error) {
	if a.d.Client == nil {
		return nil, nil
	}
	return a.d.Client.GetMeasures(a.d.ProjectKey, a.d.Branch, a.d.PullRequest, nil)
}

func (a *sonarAdapter) QualityGate() (model.QualityGate, error) {
	if a.d.Client == nil {
		return model.QualityGate{}, nil
	}
	return a.d.Client.GetQualityGate(a.d.ProjectKey, a.d.Branch, a.d.PullRequest)
}

// ---- GitLab adapter ----

// GitLabAdapterDeps configures a GitLab source adapter.
type GitLabAdapterDeps struct {
	Client      *gitlab.Client
	ProjectPath string
	MRIID       int
	PipelineID  int
}

type gitlabAdapter struct {
	d GitLabAdapterDeps
}

// NewGitLabAdapter builds a GitLabSource bound to deps.
func NewGitLabAdapter(d GitLabAdapterDeps) GitLabSource { return &gitlabAdapter{d: d} }

func (a *gitlabAdapter) MR() (model.MR, error) {
	if a.d.Client == nil {
		return model.MR{}, nil
	}
	return a.d.Client.GetMR(a.d.ProjectPath, a.d.MRIID)
}

func (a *gitlabAdapter) DiffSummary() (model.DiffSummary, error) {
	if a.d.Client == nil {
		return model.DiffSummary{}, nil
	}
	return a.d.Client.GetMRDiffSummary(a.d.ProjectPath, a.d.MRIID)
}

func (a *gitlabAdapter) Discussions() ([]model.Discussion, error) {
	if a.d.Client == nil {
		return nil, nil
	}
	return a.d.Client.ListMRDiscussions(a.d.ProjectPath, a.d.MRIID)
}

func (a *gitlabAdapter) Pipeline() (model.Pipeline, error) {
	if a.d.Client == nil {
		return model.Pipeline{}, nil
	}
	if a.d.PipelineID != 0 {
		pls, err := a.d.Client.ListMRPipelines(a.d.ProjectPath, a.d.MRIID)
		if err != nil {
			return model.Pipeline{}, err
		}
		for _, p := range pls {
			if p.ID == a.d.PipelineID {
				return p, nil
			}
		}
	}
	return a.d.Client.LatestMRPipeline(a.d.ProjectPath, a.d.MRIID)
}

func (a *gitlabAdapter) Jobs() ([]model.Job, error) {
	if a.d.Client == nil {
		return nil, nil
	}
	p, err := a.Pipeline()
	if err != nil {
		return nil, err
	}
	jobs, err := a.d.Client.ListPipelineJobs(a.d.ProjectPath, p.ID)
	if err != nil {
		return nil, err
	}
	for i := range jobs {
		if jobs[i].Status != "failed" {
			continue
		}
		trace, err := a.d.Client.GetJobTraceTail(a.d.ProjectPath, jobs[i].ID, 16<<10)
		if err != nil {
			continue
		}
		jobs[i].FailureExcerpt = gitlab.FailureExcerpt(trace, 40)
	}
	return jobs, nil
}

// ---- Nexus adapter ----

// NexusAdapterDeps configures a Nexus source adapter.
type NexusAdapterDeps struct {
	Violations []model.Violation
}

type nexusAdapter struct{ d NexusAdapterDeps }

// NewNexusAdapter builds a NexusSource bound to deps.
func NewNexusAdapter(d NexusAdapterDeps) NexusSource { return &nexusAdapter{d: d} }

func (a *nexusAdapter) Violations() ([]model.Violation, error) { return a.d.Violations, nil }

// ReadNexusViolations is a thin helper so the CLI doesn't import internal/nexus directly.
func ReadNexusViolations(path string) ([]model.Violation, string, error) {
	return nexus.ReadReport(path)
}
