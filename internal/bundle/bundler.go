// Package bundle gathers data from each source in parallel and assembles model.Bundle.
package bundle

import (
	"context"
	"fmt"
	"sync"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

// SonarSource is the minimum surface bundler needs from a Sonar client.
type SonarSource interface {
	Issues() ([]model.Issue, error)
	Hotspots() ([]model.Hotspot, error)
	Measures() ([]model.Measure, error)
	QualityGate() (model.QualityGate, error)
}

// GitLabSource is the minimum surface bundler needs from a GitLab client.
type GitLabSource interface {
	MR() (model.MR, error)
	DiffSummary() (model.DiffSummary, error)
	Discussions() ([]model.Discussion, error)
	Pipeline() (model.Pipeline, error)
	Jobs() ([]model.Job, error)
}

// NexusSource is the minimum surface bundler needs from the Nexus reader.
type NexusSource interface {
	Violations() ([]model.Violation, error)
}

// Inputs are pluggable sources. Any may be nil → that source is skipped.
type Inputs struct {
	Sonar  SonarSource
	GitLab GitLabSource
	Nexus  NexusSource
}

// Result is the bundler's product, ready to render as Bundle (live) or JSONL (snapshot).
type Result struct {
	SonarBundle  model.SonarBundle
	GitLabBundle model.GitLabBundle
	NexusBundle  model.NexusBundle
	Status       map[string]string
	Errors       []model.SourceError
}

// Gather runs all source fetches concurrently. It never panics; per-source failures
// are captured in Errors and Status.
func Gather(ctx context.Context, in Inputs) Result {
	r := Result{Status: map[string]string{}}
	var mu sync.Mutex
	addErr := func(src string, err error) {
		mu.Lock()
		defer mu.Unlock()
		r.Status[src] = "error"
		r.Errors = append(r.Errors, model.SourceError{Source: src, Message: err.Error()})
	}
	setOK := func(src string) {
		mu.Lock()
		defer mu.Unlock()
		if _, ok := r.Status[src]; !ok {
			r.Status[src] = "ok"
		}
	}

	var wg sync.WaitGroup

	if in.Sonar != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ctx.Err(); err != nil {
				addErr("sonar", fmt.Errorf("context: %w", err))
				return
			}
			iss, err := in.Sonar.Issues()
			if err != nil {
				addErr("sonar", err)
				return
			}
			hs, err := in.Sonar.Hotspots()
			if err != nil {
				addErr("sonar", err)
				return
			}
			ms, err := in.Sonar.Measures()
			if err != nil {
				addErr("sonar", err)
				return
			}
			qg, err := in.Sonar.QualityGate()
			if err != nil {
				addErr("sonar", err)
				return
			}
			mu.Lock()
			r.SonarBundle = model.SonarBundle{Issues: iss, Hotspots: hs, Measures: ms, QualityGate: &qg}
			mu.Unlock()
			setOK("sonar")
		}()
	}

	if in.GitLab != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ctx.Err(); err != nil {
				addErr("gitlab", fmt.Errorf("context: %w", err))
				return
			}
			mr, err := in.GitLab.MR()
			if err != nil {
				addErr("gitlab", err)
				return
			}
			diff, err := in.GitLab.DiffSummary()
			if err != nil {
				addErr("gitlab", err)
				return
			}
			disc, err := in.GitLab.Discussions()
			if err != nil {
				addErr("gitlab", err)
				return
			}
			pl, err := in.GitLab.Pipeline()
			if err != nil {
				addErr("gitlab", err)
				return
			}
			jobs, err := in.GitLab.Jobs()
			if err != nil {
				addErr("gitlab", err)
				return
			}
			failed := make([]model.Job, 0, len(jobs))
			for _, j := range jobs {
				if j.Status == "failed" {
					failed = append(failed, j)
				}
			}
			mu.Lock()
			r.GitLabBundle = model.GitLabBundle{
				MR: &mr, DiffSummary: &diff, Discussions: disc,
				Pipeline: &pl, JobsFailed: failed,
			}
			mu.Unlock()
			setOK("gitlab")
		}()
	}

	if in.Nexus != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ctx.Err(); err != nil {
				addErr("nexus", fmt.Errorf("context: %w", err))
				return
			}
			vs, err := in.Nexus.Violations()
			if err != nil {
				addErr("nexus", err)
				return
			}
			mu.Lock()
			r.NexusBundle = model.NexusBundle{Violations: vs}
			mu.Unlock()
			setOK("nexus")
		}()
	}

	wg.Wait()
	return r
}
