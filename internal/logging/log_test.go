// internal/logging/log_test.go
package logging_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/logging"
)

func TestInitWritesToWriterAtLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.Init(logging.Options{Level: "debug", Writer: &buf})
	logger.Debug().Msg("hello-debug")
	logger.Info().Msg("hello-info")
	out := buf.String()
	require.Contains(t, out, "hello-debug")
	require.Contains(t, out, "hello-info")
	require.True(t, strings.Contains(out, `"level":"debug"`) || strings.Contains(out, "DBG"))
}

func TestSilentLevelSuppressesNonErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.Init(logging.Options{Level: "error", Writer: &buf})
	logger.Info().Msg("should-not-appear")
	logger.Error().Msg("should-appear")
	require.NotContains(t, buf.String(), "should-not-appear")
	require.Contains(t, buf.String(), "should-appear")
}

func TestInvalidLevelDefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	_ = logging.Init(logging.Options{Level: "garbage", Writer: &buf})
	require.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}
