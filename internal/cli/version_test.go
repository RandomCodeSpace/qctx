package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

func TestVersionSubcommand(t *testing.T) {
	var out, eout bytes.Buffer
	rc := cli.Execute(cli.Args{Argv: []string{"version"}, Stdout: &out, Stderr: &eout})
	require.Equal(t, 0, rc)
	require.True(t, strings.HasPrefix(out.String(), "qctx "))
}
