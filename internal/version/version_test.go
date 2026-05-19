package version_test

import (
	"strings"
	"testing"

	"github.com/RandomCodeSpace/qctx/internal/version"
)

func TestStringFormat(t *testing.T) {
	if !strings.Contains(version.String(), "qctx ") {
		t.Fatalf("expected 'qctx ' prefix, got %q", version.String())
	}
}

func TestDefaultsArePlaceholders(t *testing.T) {
	if version.Version == "" {
		t.Fatal("Version must not be empty")
	}
}
