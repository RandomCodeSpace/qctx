// Package version exposes build-time variables.
package version

import "fmt"

// Version is the semantic version; populated via -ldflags at build time.
var Version = "dev"

// Commit is the short SHA the binary was built from; -ldflags-populated.
var Commit = "none"

// Date is the build timestamp (RFC 3339, UTC); -ldflags-populated.
var Date = "unknown"

// String returns a single-line version banner.
func String() string {
	return fmt.Sprintf("qctx %s (commit %s, built %s)", Version, Commit, Date)
}
