// Package logging configures zerolog for qctx.
package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Options struct {
	Level  string
	Writer io.Writer
	Pretty bool
}

func Init(o Options) zerolog.Logger {
	if o.Writer == nil {
		o.Writer = os.Stderr
	}
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(parseLevel(o.Level))

	var out io.Writer
	out = o.Writer
	if o.Pretty {
		out = zerolog.ConsoleWriter{Out: o.Writer, TimeFormat: time.RFC3339}
	}
	logger := zerolog.New(out).With().Timestamp().Logger()
	log.Logger = logger
	return logger
}

func parseLevel(s string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
