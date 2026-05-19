package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/config"
)

func TestSecretGoStringAndIsEmpty(t *testing.T) {
	s := config.Secret("x")
	require.Equal(t, "***redacted***", s.GoString())
	require.False(t, s.IsEmpty())
	require.True(t, config.Secret("").IsEmpty())
}

func TestResolveTokenFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tok")
	require.NoError(t, os.WriteFile(path, []byte("file-token\n"), 0o600))

	c, err := config.Load(config.LoadOptions{
		Flags: config.Flags{
			SonarURL:       "https://s.example",
			GitLabURL:      "https://g.example",
			SonarTokenFile: path,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "file-token", c.SonarToken.Reveal())
}

func TestResolveTokenFileMissingErrors(t *testing.T) {
	_, err := config.Load(config.LoadOptions{
		Flags: config.Flags{
			SonarURL:       "https://s.example",
			GitLabURL:      "https://g.example",
			SonarTokenFile: "/no/such/token/file",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "sonar token")
}

func TestResolveTokenFallbackToEnv(t *testing.T) {
	t.Setenv("SONAR_TOKEN", "env-token")
	c, err := config.Load(config.LoadOptions{
		Flags: config.Flags{SonarURL: "https://s.example", GitLabURL: "https://g.example"},
	})
	require.NoError(t, err)
	require.Equal(t, "env-token", c.SonarToken.Reveal())
}

func TestLoadFileParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: [valid: yaml"), 0o600))

	_, err := config.Load(config.LoadOptions{ConfigFile: path})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse config")
}

func TestLoadFileMissingIsOK(t *testing.T) {
	_, err := config.Load(config.LoadOptions{
		ConfigFile: "/no/such/qctx.yaml",
		Flags:      config.Flags{SonarURL: "https://s.example", GitLabURL: "https://g.example"},
	})
	require.NoError(t, err)
}

func TestRequiredGitLabMissing(t *testing.T) {
	t.Setenv("GITLAB_HOST_URL", "")
	_, err := config.Load(config.LoadOptions{
		Flags:   config.Flags{SonarURL: "https://s.example"},
		Require: config.Required{GitLab: true},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "gitlab")
}
