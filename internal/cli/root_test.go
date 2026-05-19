// internal/cli/root_test.go
package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

func TestRootShowsHelpWithoutArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{Argv: []string{"--help"}, Stdout: &out, Stderr: &errOut})
	require.Equal(t, 0, rc)
	require.Contains(t, out.String()+errOut.String(), "qctx")
}

func TestRootUnknownCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	rc := cli.Execute(cli.Args{Argv: []string{"nope"}, Stdout: &out, Stderr: &errOut})
	require.NotEqual(t, 0, rc)
}
