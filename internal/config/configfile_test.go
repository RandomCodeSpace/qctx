package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/config"
)

func TestConfigFileProvidesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "qctx.yaml")
	require.NoError(t, os.WriteFile(path, []byte(
		"sonar_url: https://from-file.example\n"+
			"gitlab_url: https://gl-from-file.example\n"+
			"sonar_token: file-tok\n",
	), 0o600))

	t.Setenv("SONAR_HOST_URL", "")
	t.Setenv("GITLAB_HOST_URL", "")
	t.Setenv("SONAR_TOKEN", "")

	c, err := config.Load(config.LoadOptions{ConfigFile: path})
	require.NoError(t, err)
	require.Equal(t, "https://from-file.example", c.SonarURL)
	require.Equal(t, "https://gl-from-file.example", c.GitLabURL)
	require.Equal(t, "file-tok", c.SonarToken.Reveal())
}

func TestLogLevelFromEnvWhenFlagUnset(t *testing.T) {
	t.Setenv("QCTX_LOG_LEVEL", "debug")
	c, err := config.Load(config.LoadOptions{
		Flags: config.Flags{SonarURL: "https://s.example", GitLabURL: "https://g.example"},
	})
	require.NoError(t, err)
	require.Equal(t, "debug", c.LogLevel)
}

func TestLogLevelFlagBeatsEnv(t *testing.T) {
	t.Setenv("QCTX_LOG_LEVEL", "warn")
	c, err := config.Load(config.LoadOptions{
		Flags: config.Flags{
			SonarURL:  "https://s.example",
			GitLabURL: "https://g.example",
			LogLevel:  "error",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "error", c.LogLevel)
}
