# 0005 — Cobra as the CLI framework

**Status:** Accepted
**Date:** 2026-05-19

## Context

qctx is a CLI with three subcommands (`fetch`, `snapshot`, `version`) and a flag set of ~20 entries shared between two of them. We need:

- A clean subcommand surface — not all flags apply to all commands.
- Shared flag definitions (the `fetch` and `snapshot` configs differ only by an `--out` requirement and the output sink).
- Auto-generated help and shell completion.
- Testable invocation (Stdout/Stderr capturable; no hard-coded `os.Exit` inside command logic).

## Decision

Use `github.com/spf13/cobra` v1.x. Each subcommand registers itself in an `init()` via a `registerSubcommands` closure on the root, so the root command stays unaware of which subcommands exist at compile time. `cli.Execute(Args)` accepts injected `Stdout`/`Stderr`/`Argv`, making the whole CLI tree testable from Go without forking a child process.

Each command sets `Args: cobra.NoArgs` plus an explicit `RunE` that errors on unexpected positionals — cobra's default of "print help and return nil" was the source of a real test bug we hit in Phase 1 and is not what we want.

## Consequences

**Upside.**
- Help text, completion, and flag inheritance are free.
- Subcommand isolation: adding a new command is one file with an `init()`.
- `Args: cobra.NoArgs` + `RunE` gives the strictness we want.
- Wide ecosystem familiarity; new contributors usually know cobra already.

**Downside.**
- Cobra is a non-trivial dependency. It pulls `spf13/pflag` and `inconshreveable/mousetrap`. We accept this; the productivity gain dominates.
- Cobra's default behavior on unknown commands is surprising; we have to set `Args: cobra.NoArgs` and `RunE` explicitly. Documented in `internal/cli/root.go` and reinforced by `TestRootUnknownCommand`.

## Alternatives considered

- **stdlib `flag` + manual dispatch.** Works for a binary with two subcommands but the flag-inheritance and help-text overhead grows fast. We have ~20 flags shared across two subcommands; rolling that by hand is busywork.
- **`urfave/cli`.** Comparable feature set but less common in our reviewer base.
- **`kingpin`.** Project is stale.
- **`fang`** (a cobra wrapper). Pretty but adds a layer for marginal benefit.
