package bundle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/bundle"
	"github.com/RandomCodeSpace/qctx/internal/model"
)

type fakeSonar struct {
	issues   []model.Issue
	hotspots []model.Hotspot
	measures []model.Measure
	gate     model.QualityGate
	failAll  bool
}

func (f *fakeSonar) Issues() ([]model.Issue, error) {
	if f.failAll {
		return nil, errors.New("boom")
	}
	return f.issues, nil
}
func (f *fakeSonar) Hotspots() ([]model.Hotspot, error)      { return f.hotspots, nil }
func (f *fakeSonar) Measures() ([]model.Measure, error)      { return f.measures, nil }
func (f *fakeSonar) QualityGate() (model.QualityGate, error) { return f.gate, nil }

type fakeGitLab struct {
	mr       model.MR
	diff     model.DiffSummary
	disc     []model.Discussion
	pipeline model.Pipeline
	jobs     []model.Job
	failMR   bool
}

func (f *fakeGitLab) MR() (model.MR, error) {
	if f.failMR {
		return model.MR{}, errors.New("mr-fail")
	}
	return f.mr, nil
}
func (f *fakeGitLab) DiffSummary() (model.DiffSummary, error)  { return f.diff, nil }
func (f *fakeGitLab) Discussions() ([]model.Discussion, error) { return f.disc, nil }
func (f *fakeGitLab) Pipeline() (model.Pipeline, error)        { return f.pipeline, nil }
func (f *fakeGitLab) Jobs() ([]model.Job, error)               { return f.jobs, nil }

type fakeNexus struct{ violations []model.Violation }

func (f *fakeNexus) Violations() ([]model.Violation, error) { return f.violations, nil }

func TestBundleGathersAllSources(t *testing.T) {
	res := bundle.Gather(context.Background(), bundle.Inputs{
		Sonar:  &fakeSonar{issues: []model.Issue{{Key: "i1", Status: "OPEN"}}},
		GitLab: &fakeGitLab{mr: model.MR{IID: 42}, jobs: []model.Job{{Name: "t", Status: "failed"}}},
		Nexus:  &fakeNexus{violations: []model.Violation{{Component: "c", Status: "open"}}},
	})
	require.Equal(t, "ok", res.Status["sonar"])
	require.Equal(t, "ok", res.Status["gitlab"])
	require.Equal(t, "ok", res.Status["nexus"])
	require.Len(t, res.SonarBundle.Issues, 1)
	require.NotNil(t, res.GitLabBundle.MR)
	require.Equal(t, 42, res.GitLabBundle.MR.IID)
	require.Len(t, res.GitLabBundle.JobsFailed, 1)
	require.Empty(t, res.Errors)
}

func TestBundlePartialFailure(t *testing.T) {
	res := bundle.Gather(context.Background(), bundle.Inputs{
		Sonar:  &fakeSonar{failAll: true},
		GitLab: &fakeGitLab{mr: model.MR{IID: 42}},
	})
	require.Equal(t, "error", res.Status["sonar"])
	require.Equal(t, "ok", res.Status["gitlab"])
	require.NotEmpty(t, res.Errors)
}

func TestBundleRespectsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res := bundle.Gather(ctx, bundle.Inputs{Sonar: &fakeSonar{}})
	require.NotEmpty(t, res.Errors)
}
