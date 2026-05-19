package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/RandomCodeSpace/qctx/internal/bundle"
	"github.com/RandomCodeSpace/qctx/internal/config"
	"github.com/RandomCodeSpace/qctx/internal/gitlab"
	"github.com/RandomCodeSpace/qctx/internal/httpclient"
	"github.com/RandomCodeSpace/qctx/internal/logging"
	"github.com/RandomCodeSpace/qctx/internal/model"
	"github.com/RandomCodeSpace/qctx/internal/sonar"
	"github.com/RandomCodeSpace/qctx/internal/version"
)

var fetchFlags config.Flags

func init() {
	prev := registerSubcommands
	registerSubcommands = func(root *cobra.Command) {
		prev(root)
		cmd := &cobra.Command{
			Use:   "fetch",
			Short: "Fetch live JSON bundle for an AI agent",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, _ []string) error {
				return runFetch(c)
			},
		}
		bindCommonFlags(cmd, &fetchFlags)
		root.AddCommand(cmd)
	}
}

// bindCommonFlags wires the flag set used by both fetch and snapshot.
func bindCommonFlags(cmd *cobra.Command, f *config.Flags) {
	cmd.Flags().StringVar(&f.SonarURL, "sonar-url", "", "SonarQube base URL (env: SONAR_HOST_URL)")
	cmd.Flags().StringVar(&f.SonarToken, "sonar-token", "", "SonarQube token (env: SONAR_TOKEN)")
	cmd.Flags().StringVar(&f.SonarTokenFile, "sonar-token-file", "", "Path to file containing the Sonar token")
	cmd.Flags().StringVar(&f.GitLabURL, "gitlab-url", "", "GitLab base URL (env: GITLAB_HOST_URL)")
	cmd.Flags().StringVar(&f.GitLabToken, "gitlab-token", "", "GitLab token (env: GITLAB_TOKEN)")
	cmd.Flags().StringVar(&f.GitLabTokenFile, "gitlab-token-file", "", "Path to file containing the GitLab token")
	cmd.Flags().StringVar(&f.CACertPath, "ca-cert", "", "PEM bundle of additional trusted CAs")
	cmd.Flags().BoolVar(&f.Insecure, "insecure", false, "Skip TLS verification (dev only)")
	cmd.Flags().StringArrayVar(&f.ExtraHeaders, "header", nil, "Additional header to send (repeatable; 'Name: value')")
	cmd.Flags().StringVar(&f.MRURL, "mr", "", "GitLab MR URL (enables auto-discovery)")
	cmd.Flags().StringVar(&f.ProjectKey, "project", "", "Sonar project key (overrides auto-discovery)")
	cmd.Flags().StringVar(&f.Branch, "branch", "", "Branch override")
	cmd.Flags().StringVar(&f.NexusReport, "nexus-report", "", "Path to Nexus IQ JSON report")
	cmd.Flags().StringArrayVar(&f.Severities, "severity", nil, "Sonar severity filter (repeatable)")
	cmd.Flags().StringArrayVar(&f.Types, "type", nil, "Sonar issue type filter (repeatable)")
	cmd.Flags().BoolVar(&f.AllIssues, "all", false, "Include all open issues, not just MR-touched")
	cmd.Flags().BoolVar(&f.IncludeResolved, "include-resolved", false, "Include resolved/closed issues")
	cmd.Flags().BoolVar(&f.DisableSonar, "no-sonar", false, "Skip Sonar")
	cmd.Flags().BoolVar(&f.DisableGitLab, "no-gitlab", false, "Skip GitLab")
	cmd.Flags().BoolVar(&f.DisableNexus, "no-nexus", false, "Skip Nexus")
	cmd.Flags().BoolVar(&f.DisableMR, "no-mr", false, "Skip MR fetch")
	cmd.Flags().BoolVar(&f.DisablePipeline, "no-pipeline", false, "Skip pipeline fetch")
	cmd.Flags().BoolVar(&f.Strict, "strict", false, "Exit non-zero on any source error")
	cmd.Flags().StringVar(&f.OutPath, "out", "", "Output file (snapshot only)")
	cmd.Flags().StringVar(&f.LogLevel, "log-level", "", "Log level: debug, info, warn, error (env: QCTX_LOG_LEVEL)")
	cmd.Flags().StringVar(&f.ConfigFile, "config", "", "Path to config file (default: $QCTX_CONFIG or ~/.qctx.yaml)")
}

func runFetch(c *cobra.Command) error {
	deps, err := preparePipeline(fetchFlags)
	if err != nil {
		return err
	}
	res := bundle.Gather(c.Context(), bundle.Inputs{
		Sonar:  maybeSonar(deps.Config, deps.Sonar, deps.SonarProject, deps.MR),
		GitLab: maybeGitLab(deps.Config, deps.GitLab, deps.MR),
		Nexus:  maybeNexus(deps.Config, deps.NexusVi),
	})
	meta := model.Meta{
		Tool: "qctx", Version: version.Version, ScannedAt: time.Now().UTC(),
		SonarProjectKey: deps.SonarProject,
		GitLabProject:   deps.MR.ProjectPath,
		Branch:          deps.MR.Branch,
		MRIID:           deps.MR.IID,
		SourceStatus:    res.Status,
	}
	b := model.Bundle{
		Meta: meta, Sonar: res.SonarBundle, Nexus: res.NexusBundle, GitLab: res.GitLabBundle, Errors: res.Errors,
	}
	if err := model.WriteJSON(c.OutOrStdout(), b); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return finalExit(deps.Config.Strict, res)
}

// mrCoords holds resolved MR identity for downstream helpers.
type mrCoords struct {
	ProjectPath string
	IID         int
	PipelineID  int
	Branch      string
}

// pipelineDeps is the resolved set of clients, config, and MR coordinates that
// runFetch / runSnapshot need. preparePipeline returns one of these instead of
// a 7-tuple.
type pipelineDeps struct {
	Config       config.Config
	Sonar        *sonar.Client
	GitLab       *gitlab.Client
	NexusVi      []model.Violation
	SonarProject string
	MR           mrCoords
}

// preparePipeline builds clients + resolves MR coordinates from flags.
func preparePipeline(f config.Flags) (pipelineDeps, error) {
	configPath := f.ConfigFile
	if configPath == "" {
		configPath = defaultConfigFile()
	}
	cfg, err := config.Load(config.LoadOptions{
		Flags:      f,
		ConfigFile: configPath,
		Require: config.Required{
			Sonar:  !f.DisableSonar,
			GitLab: !f.DisableGitLab,
		},
	})
	if err != nil {
		return pipelineDeps{}, err
	}
	logging.Init(logging.Options{Level: cfg.LogLevel})

	deps := pipelineDeps{Config: cfg, MR: mrCoords{Branch: cfg.Branch}}

	hc, err := httpclient.New(httpclient.Options{
		CACertPath: cfg.CACertPath, Insecure: cfg.Insecure, ExtraHeaders: cfg.ExtraHeaders,
	})
	if err != nil {
		return deps, err
	}

	if !cfg.DisableSonar && cfg.SonarURL != "" {
		deps.Sonar, err = sonar.New(sonar.Options{BaseURL: cfg.SonarURL, Token: cfg.SonarToken.Reveal(), HTTP: hc})
		if err != nil {
			return deps, err
		}
	}

	if !cfg.DisableGitLab && cfg.GitLabURL != "" {
		deps.GitLab, err = gitlab.New(gitlab.Options{BaseURL: cfg.GitLabURL, Token: cfg.GitLabToken.Reveal(), HTTP: hc})
		if err != nil {
			return deps, err
		}
	}

	if cfg.MRURL != "" && deps.GitLab != nil {
		parsed, err := gitlab.ParseMRURL(cfg.MRURL)
		if err != nil {
			return deps, err
		}
		deps.MR.ProjectPath = parsed.ProjectPath
		deps.MR.IID = parsed.IID
		if pl, err := deps.GitLab.LatestMRPipeline(deps.MR.ProjectPath, deps.MR.IID); err == nil {
			deps.MR.PipelineID = pl.ID
			if deps.MR.Branch == "" {
				deps.MR.Branch = pl.Ref
			}
		}
	}

	deps.SonarProject = cfg.ProjectKey
	if deps.SonarProject == "" && !cfg.DisableSonar && deps.GitLab != nil && deps.MR.PipelineID != 0 {
		if k, dErr := deps.GitLab.DiscoverSonarProjectKey(deps.MR.ProjectPath, deps.MR.PipelineID); dErr == nil {
			deps.SonarProject = k
		}
	}
	if !cfg.DisableSonar && deps.SonarProject == "" {
		return deps, fmt.Errorf("sonar project key unknown: pass --project or enable auto-discovery via --mr")
	}

	if !cfg.DisableNexus && cfg.NexusReport != "" {
		v, _, rErr := bundle.ReadNexusViolations(cfg.NexusReport)
		if rErr != nil {
			return deps, rErr
		}
		deps.NexusVi = v
	}

	return deps, nil
}

func maybeSonar(cfg config.Config, c *sonar.Client, project string, mr mrCoords) bundle.SonarSource {
	if cfg.DisableSonar || c == nil {
		return nil
	}
	pr := ""
	if mr.IID > 0 && !cfg.AllIssues {
		pr = strconv.Itoa(mr.IID)
	}
	branch := cfg.Branch
	if branch == "" {
		branch = mr.Branch
	}
	return bundle.NewSonarAdapter(bundle.SonarAdapterDeps{
		Client: c, ProjectKey: project, Branch: branch, PullRequest: pr,
		Severities: cfg.Severities, Types: cfg.Types, Resolved: cfg.IncludeResolved,
	})
}

func maybeGitLab(cfg config.Config, c *gitlab.Client, mr mrCoords) bundle.GitLabSource {
	if cfg.DisableGitLab || c == nil || mr.ProjectPath == "" || mr.IID == 0 {
		return nil
	}
	return bundle.NewGitLabAdapter(bundle.GitLabAdapterDeps{Client: c, ProjectPath: mr.ProjectPath, MRIID: mr.IID, PipelineID: mr.PipelineID})
}

func maybeNexus(cfg config.Config, vs []model.Violation) bundle.NexusSource {
	if cfg.DisableNexus || cfg.NexusReport == "" {
		return nil
	}
	return bundle.NewNexusAdapter(bundle.NexusAdapterDeps{Violations: vs})
}

func finalExit(strict bool, r bundle.Result) error {
	if !strict {
		ok := 0
		for _, s := range r.Status {
			if s == "ok" {
				ok++
			}
		}
		if ok == 0 && len(r.Status) > 0 {
			return fmt.Errorf("all sources failed: %d error(s)", len(r.Errors))
		}
		return nil
	}
	if len(r.Errors) > 0 {
		return fmt.Errorf("strict mode: %d source error(s)", len(r.Errors))
	}
	return nil
}

func defaultConfigFile() string {
	if v := os.Getenv("QCTX_CONFIG"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	return home + "/.qctx.yaml"
}
