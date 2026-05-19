// Package cli wires the cobra command tree.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/RandomCodeSpace/qctx/internal/version"
)

type Args struct {
	Argv   []string
	Stdout io.Writer
	Stderr io.Writer
}

func Execute(a Args) int {
	if a.Stdout == nil {
		a.Stdout = os.Stdout
	}
	if a.Stderr == nil {
		a.Stderr = os.Stderr
	}
	root := newRoot()
	root.SetArgs(a.Argv)
	root.SetOut(a.Stdout)
	root.SetErr(a.Stderr)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "qctx",
		Short:         "Quality context for AI agents (Sonar + Nexus IQ + GitLab)",
		Long:          "qctx gathers SonarQube + GitLab + Nexus IQ context for AI-driven fix workflows.",
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return cmd.Help()
		},
	}
	registerSubcommands(root)
	return root
}

// registerSubcommands is replaced by Phase 7 init() in fetch/snapshot/version files.
var registerSubcommands = func(_ *cobra.Command) {}
