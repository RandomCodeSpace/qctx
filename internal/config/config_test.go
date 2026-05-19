package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/config"
)

func TestPrecedence_FlagBeatsEnvBeatsFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "qctx.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(
		"sonar_url: https://from-file.example\n"+
			"gitlab_url: https://gl-from-file.example\n",
	), 0o600))

	t.Setenv("SONAR_HOST_URL", "https://from-env.example")
	c, err := config.Load(config.LoadOptions{
		ConfigFile: cfgPath,
		Flags:      config.Flags{SonarURL: "https://from-flag.example"},
	})
	require.NoError(t, err)
	require.Equal(t, "https://from-flag.example", c.SonarURL)
	require.Equal(t, "https://gl-from-file.example", c.GitLabURL)
}

func TestErrorWhenRequiredMissing(t *testing.T) {
	t.Setenv("SONAR_HOST_URL", "")
	_, err := config.Load(config.LoadOptions{Require: config.Required{Sonar: true}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "sonar")
}

func TestSecretRedaction(t *testing.T) {
	s := config.Secret("supersecret")
	require.Equal(t, "***redacted***", s.String())
	require.Equal(t, "supersecret", s.Reveal())
}

func TestExtraHeadersParse(t *testing.T) {
	c, err := config.Load(config.LoadOptions{
		Flags: config.Flags{
			SonarURL:     "https://s.example",
			GitLabURL:    "https://g.example",
			ExtraHeaders: []string{"X-A: 1", "X-B: 2 with spaces"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "1", c.ExtraHeaders["X-A"])
	require.Equal(t, "2 with spaces", c.ExtraHeaders["X-B"])
}

func TestMalformedHeaderIsError(t *testing.T) {
	_, err := config.Load(config.LoadOptions{Flags: config.Flags{ExtraHeaders: []string{"not-a-header"}}})
	require.Error(t, err)
}
