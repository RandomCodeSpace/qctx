// Package config loads runtime configuration from flags, env, and an optional YAML file.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Secret string

func (s Secret) String() string   { return "***redacted***" }
func (s Secret) GoString() string { return "***redacted***" }
func (s Secret) Reveal() string   { return string(s) }
func (s Secret) IsEmpty() bool    { return string(s) == "" }

type Config struct {
	SonarURL    string
	SonarToken  Secret
	GitLabURL   string
	GitLabToken Secret

	CACertPath   string
	Insecure     bool
	ExtraHeaders map[string]string

	DisableSonar    bool
	DisableGitLab   bool
	DisableNexus    bool
	DisableMR       bool
	DisablePipeline bool

	Severities      []string
	Types           []string
	Branch          string
	AllIssues       bool
	IncludeResolved bool

	MRURL       string
	ProjectKey  string
	NexusReport string
	OutPath     string
	Strict      bool

	LogLevel string // "debug" | "info" | "warn" | "error"; default "info"
}

type Flags struct {
	SonarURL, SonarToken, SonarTokenFile    string
	GitLabURL, GitLabToken, GitLabTokenFile string
	CACertPath                              string
	Insecure                                bool
	ExtraHeaders                            []string
	DisableSonar, DisableGitLab, DisableNexus,
	DisableMR, DisablePipeline bool
	Severities, Types                       []string
	Branch                                  string
	AllIssues, IncludeResolved              bool
	MRURL, ProjectKey, NexusReport, OutPath string
	Strict                                  bool
	LogLevel                                string
	ConfigFile                              string // explicit path overrides default discovery
}

type Required struct {
	Sonar  bool
	GitLab bool
}

type LoadOptions struct {
	ConfigFile string
	Flags      Flags
	Require    Required
}

func Load(opts LoadOptions) (Config, error) {
	fc, err := loadFile(opts.ConfigFile)
	if err != nil {
		return Config{}, err
	}

	c := Config{
		SonarURL:        pick(opts.Flags.SonarURL, os.Getenv("SONAR_HOST_URL"), fc.SonarURL),
		GitLabURL:       pick(opts.Flags.GitLabURL, os.Getenv("GITLAB_HOST_URL"), fc.GitLabURL),
		CACertPath:      pick(opts.Flags.CACertPath, os.Getenv("QCTX_CA_CERT"), fc.CACertPath),
		Insecure:        opts.Flags.Insecure || envTrue("QCTX_INSECURE") || fc.Insecure,
		DisableSonar:    opts.Flags.DisableSonar,
		DisableGitLab:   opts.Flags.DisableGitLab,
		DisableNexus:    opts.Flags.DisableNexus,
		DisableMR:       opts.Flags.DisableMR,
		DisablePipeline: opts.Flags.DisablePipeline,
		Severities:      opts.Flags.Severities,
		Types:           opts.Flags.Types,
		Branch:          opts.Flags.Branch,
		AllIssues:       opts.Flags.AllIssues,
		IncludeResolved: opts.Flags.IncludeResolved,
		MRURL:           opts.Flags.MRURL,
		ProjectKey:      opts.Flags.ProjectKey,
		NexusReport:     opts.Flags.NexusReport,
		OutPath:         opts.Flags.OutPath,
		Strict:          opts.Flags.Strict,
		LogLevel:        pick(opts.Flags.LogLevel, os.Getenv("QCTX_LOG_LEVEL")),
	}

	sTok, err := resolveToken(opts.Flags.SonarToken, opts.Flags.SonarTokenFile, "SONAR_TOKEN", fc.SonarToken)
	if err != nil {
		return Config{}, fmt.Errorf("sonar token: %w", err)
	}
	c.SonarToken = Secret(sTok)

	gTok, err := resolveToken(opts.Flags.GitLabToken, opts.Flags.GitLabTokenFile, "GITLAB_TOKEN", fc.GitLabToken)
	if err != nil {
		return Config{}, fmt.Errorf("gitlab token: %w", err)
	}
	c.GitLabToken = Secret(gTok)

	h, err := parseHeaders(opts.Flags.ExtraHeaders)
	if err != nil {
		return Config{}, err
	}
	c.ExtraHeaders = h

	if opts.Require.Sonar && c.SonarURL == "" {
		return Config{}, errors.New("sonar URL is required: pass --sonar-url or set SONAR_HOST_URL")
	}
	if opts.Require.GitLab && c.GitLabURL == "" {
		return Config{}, errors.New("gitlab URL is required: pass --gitlab-url or set GITLAB_HOST_URL")
	}
	return c, nil
}

func pick(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func envTrue(k string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	return v == "1" || v == "true" || v == "yes"
}

func resolveToken(flagVal, fileFlag, envVar, fileCfgVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if fileFlag != "" {
		b, err := os.ReadFile(fileFlag)
		if err != nil {
			return "", fmt.Errorf("read token file %q: %w", fileFlag, err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	if v := os.Getenv(envVar); v != "" {
		return v, nil
	}
	return fileCfgVal, nil
}

func parseHeaders(in []string) (map[string]string, error) {
	out := map[string]string{}
	for _, h := range in {
		i := strings.Index(h, ":")
		if i <= 0 || i == len(h)-1 {
			return nil, fmt.Errorf("malformed --header %q (expected 'Name: value')", h)
		}
		name := strings.TrimSpace(h[:i])
		val := strings.TrimSpace(h[i+1:])
		if name == "" || val == "" {
			return nil, fmt.Errorf("malformed --header %q", h)
		}
		out[name] = val
	}
	return out, nil
}
