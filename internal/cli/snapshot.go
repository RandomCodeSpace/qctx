package cli

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/RandomCodeSpace/qctx/internal/bundle"
	"github.com/RandomCodeSpace/qctx/internal/config"
	"github.com/RandomCodeSpace/qctx/internal/model"
	"github.com/RandomCodeSpace/qctx/internal/version"
)

var snapshotFlags config.Flags

func init() {
	prev := registerSubcommands
	registerSubcommands = func(root *cobra.Command) {
		prev(root)
		cmd := &cobra.Command{
			Use:   "snapshot",
			Short: "Write JSONL artifact for downstream agent consumption",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, _ []string) error {
				return runSnapshot(c)
			},
		}
		bindCommonFlags(cmd, &snapshotFlags)
		_ = cmd.MarkFlagRequired("out")
		root.AddCommand(cmd)
	}
}

func runSnapshot(c *cobra.Command) error {
	deps, err := preparePipeline(snapshotFlags)
	if err != nil {
		return err
	}
	if deps.Config.OutPath == "" {
		return fmt.Errorf("--out PATH is required")
	}

	res := bundle.Gather(c.Context(), bundle.Inputs{
		Sonar:  maybeSonar(deps.Config, deps.Sonar, deps.SonarProject, deps.MR),
		GitLab: maybeGitLab(deps.Config, deps.GitLab, deps.MR),
		Nexus:  maybeNexus(deps.Config, deps.NexusVi),
	})

	f, err := os.OpenFile(deps.Config.OutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- user-supplied artifact path by design
	if err != nil {
		return fmt.Errorf("open out: %w", err)
	}
	defer func() { _ = f.Close() }()
	bw := bufio.NewWriter(f)
	w := model.NewJSONLWriter(bw)

	meta := model.Meta{
		Tool: "qctx", Version: version.Version, ScannedAt: time.Now().UTC(),
		SonarProjectKey: deps.SonarProject, GitLabProject: deps.MR.ProjectPath,
		Branch: deps.MR.Branch, MRIID: deps.MR.IID, SourceStatus: res.Status,
	}
	if err := w.WriteMeta(meta); err != nil {
		return err
	}
	for _, i := range res.SonarBundle.Issues {
		if err := w.WriteIssue(i); err != nil {
			return err
		}
	}
	for _, h := range res.SonarBundle.Hotspots {
		if err := w.WriteHotspot(h); err != nil {
			return err
		}
	}
	for _, m := range res.SonarBundle.Measures {
		if err := w.WriteMeasure(m); err != nil {
			return err
		}
	}
	if res.SonarBundle.QualityGate != nil {
		if err := w.WriteQualityGate(*res.SonarBundle.QualityGate); err != nil {
			return err
		}
	}
	for _, v := range res.NexusBundle.Violations {
		if err := w.WriteNexusViolation(v); err != nil {
			return err
		}
	}
	if res.GitLabBundle.MR != nil {
		if err := w.WriteMR(*res.GitLabBundle.MR); err != nil {
			return err
		}
	}
	if res.GitLabBundle.DiffSummary != nil {
		if err := w.WriteDiffSummary(*res.GitLabBundle.DiffSummary); err != nil {
			return err
		}
	}
	for _, d := range res.GitLabBundle.Discussions {
		if err := w.WriteDiscussion(d); err != nil {
			return err
		}
	}
	if res.GitLabBundle.Pipeline != nil {
		if err := w.WritePipeline(*res.GitLabBundle.Pipeline); err != nil {
			return err
		}
	}
	for _, j := range res.GitLabBundle.JobsFailed {
		if err := w.WriteJob(j); err != nil {
			return err
		}
	}
	for _, e := range res.Errors {
		if err := w.WriteError(e); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	return finalExit(deps.Config.Strict, res)
}
