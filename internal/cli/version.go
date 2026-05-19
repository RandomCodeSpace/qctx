package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RandomCodeSpace/qctx/internal/version"
)

func init() {
	prev := registerSubcommands
	registerSubcommands = func(root *cobra.Command) {
		prev(root)
		root.AddCommand(&cobra.Command{
			Use:   "version",
			Short: "Print version",
			Args:  cobra.NoArgs,
			Run: func(c *cobra.Command, _ []string) {
				_, _ = fmt.Fprintln(c.OutOrStdout(), version.String())
			},
		})
	}
}
