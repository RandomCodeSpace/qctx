// Command qctx is the CLI entrypoint.
package main

import (
	"os"

	"github.com/RandomCodeSpace/qctx/internal/cli"
)

func main() {
	os.Exit(cli.Execute(cli.Args{Argv: os.Args[1:], Stdout: os.Stdout, Stderr: os.Stderr}))
}
