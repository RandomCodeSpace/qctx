GO      ?= go
PKG     := github.com/RandomCodeSpace/qctx
BINARY  := qctx
BIN_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
           -X $(PKG)/internal/version.Version=$(VERSION) \
           -X $(PKG)/internal/version.Commit=$(COMMIT) \
           -X $(PKG)/internal/version.Date=$(DATE)

# COVER_THRESHOLD: minimum per-package coverage for logic packages.
# Wiring/entrypoint packages (cmd/qctx, internal/bundle adapters) are excluded.
COVER_THRESHOLD ?= 80.0
COVER_EXCLUDE   := cmd/qctx

.PHONY: all build test lint cover cover-check e2e clean fmt tidy ci doctor help

.DEFAULT_GOAL := help

help:  ## show this help
	@awk 'BEGIN { FS=":.*##"; printf "qctx Makefile targets:\n\n" } \
		/^[a-zA-Z0-9_-]+:.*##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build & run

all: lint test build  ## lint + test + build

build:  ## build the qctx binary to bin/qctx
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(BINARY) ./cmd/qctx

##@ Test

test:  ## run unit tests with the race detector
	$(GO) test -race -count=1 ./...

cover:  ## run tests with coverage, print total
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -func=coverage.out | tail -n 1

# cover-check enforces COVER_THRESHOLD per logic package using statement-weighted
# coverage parsed from `go test -cover` (not per-function averages, which mis-weight).
# Logic packages = all packages minus those in COVER_EXCLUDE.
cover-check:  ## fail if any logic package is below COVER_THRESHOLD (default 80%)
	@set -e; fail=0; \
	for pkg in $$($(GO) list ./... | sed 's|$(PKG)/||'); do \
		skip=0; \
		for ex in $(COVER_EXCLUDE); do \
			case "$$pkg" in "$$ex"|"$$ex"/*) skip=1 ;; esac; \
		done; \
		line=$$($(GO) test -cover ./$$pkg/... 2>&1 | grep -E 'coverage:|no test files' | tail -n 1); \
		pct=$$(echo "$$line" | sed -nE 's/.*coverage: ([0-9.]+)%.*/\1/p'); \
		if [ -z "$$pct" ]; then \
			pct=0.0; \
		fi; \
		if [ "$$skip" = "1" ]; then \
			printf "[skip] %-40s %s%%\n" "$$pkg" "$$pct"; \
			continue; \
		fi; \
		below=$$(awk -v p="$$pct" -v t="$(COVER_THRESHOLD)" 'BEGIN { print (p+0 < t+0) ? 1 : 0 }'); \
		if [ "$$below" = "1" ]; then \
			printf "[FAIL] %-40s %s%%\n" "$$pkg" "$$pct"; \
			fail=1; \
		else \
			printf "[OK  ] %-40s %s%%\n" "$$pkg" "$$pct"; \
		fi; \
	done; \
	if [ "$$fail" = "1" ]; then \
		echo "cover-check: at least one package below $(COVER_THRESHOLD)%"; exit 1; \
	fi

##@ Quality

lint:  ## run golangci-lint
	golangci-lint run ./...

fmt:  ## gofmt + goimports on the whole tree
	gofmt -s -w .
	goimports -w -local $(PKG) .

tidy:  ## go mod tidy
	$(GO) mod tidy

e2e: build  ## run end-to-end test against mock servers
	$(GO) test -tags=e2e ./test/e2e/...

clean:  ## remove build artifacts and coverage output
	rm -rf $(BIN_DIR) dist coverage.out coverage.html

ci: tidy fmt lint cover cover-check build  ## full local CI pipeline

# doctor checks that all required (and recommended) tools are installed.
# Run this after a fresh clone or before opening a PR.
doctor:  ## verify required + optional tools are installed
	@ok=1; \
	check() { name=$$1; minver=$$2; cmd=$$3; \
		if ! command -v "$$name" >/dev/null 2>&1; then \
			printf "[MISS] %-18s (required: %s)\n" "$$name" "$$minver"; ok=0; return; \
		fi; \
		v=$$($$cmd 2>&1 | head -n 1); \
		printf "[ OK ] %-18s %s\n" "$$name" "$$v"; \
	}; \
	check go "1.23+" "go version"; \
	check golangci-lint "v2.x" "golangci-lint --version"; \
	check goimports "any" "goimports -h"; \
	check make "any" "make --version"; \
	check git "any" "git --version"; \
	printf "\n[optional]\n"; \
	if command -v docker >/dev/null 2>&1; then printf "[ OK ] docker             %s\n" "$$(docker --version)"; else printf "[MISS] docker             (optional, needed for release image)\n"; fi; \
	if command -v goreleaser >/dev/null 2>&1; then printf "[ OK ] goreleaser         %s\n" "$$(goreleaser --version | head -n 1)"; else printf "[MISS] goreleaser         (optional, needed for local release)\n"; fi; \
	if [ "$$ok" = "0" ]; then echo; echo "doctor: required tool(s) missing"; exit 1; fi
