# Makefile — Sadqa Ledger
#
# Two commands for a new contributor: `make setup` then `make dev`.
# Run `make help` to see everything else.
#
# Design notes (see CONTRIBUTING.md for the human-facing version of this):
#   - Tool versions are pinned in the variables below — bump one line to upgrade.
#   - Every target is safe to re-run: `make setup` twice never overwrites .env
#     or your local database, and skips downloads/installs that already exist.
#   - air (used by `make dev`) shells out to `make templ` and `make css` as its
#     pre-build steps, so the Makefile stays the single place that knows how
#     templ/Tailwind are actually invoked — nothing is duplicated in .air.toml.

# ---------------------------------------------------------------------------
# Pinned tool versions — change here to upgrade
# ---------------------------------------------------------------------------

TEMPL_VERSION   := v0.3.833
AIR_VERSION     := v1.61.7
TAILWIND_VERSION := v3.4.17

# Minimum Go version is read from go.mod's `go` directive so it can never drift
# out of sync with what the module actually requires. Falls back to 1.23 if
# go.mod is missing or unreadable (e.g. before Phase 0 code exists yet).
GO_MIN_VERSION := $(shell awk '/^go [0-9]/{print $$2; exit}' go.mod 2>/dev/null)
ifeq ($(GO_MIN_VERSION),)
GO_MIN_VERSION := 1.23
endif

# ---------------------------------------------------------------------------
# OS / architecture detection (for the Tailwind standalone CLI download)
# ---------------------------------------------------------------------------

UNAME_S := $(shell uname -s 2>/dev/null)
UNAME_M := $(shell uname -m 2>/dev/null)

ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
DETECTED_OS := windows
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
DETECTED_OS := windows
else ifeq ($(UNAME_S),Darwin)
DETECTED_OS := darwin
else ifeq ($(UNAME_S),Linux)
DETECTED_OS := linux
else
DETECTED_OS := windows
endif

ifeq ($(UNAME_M),x86_64)
DETECTED_ARCH := amd64
else ifeq ($(UNAME_M),amd64)
DETECTED_ARCH := amd64
else ifeq ($(UNAME_M),aarch64)
DETECTED_ARCH := arm64
else ifeq ($(UNAME_M),arm64)
DETECTED_ARCH := arm64
else
DETECTED_ARCH := amd64
endif

ifeq ($(DETECTED_OS),windows)
EXE := .exe
TAILWIND_ASSET := tailwindcss-windows-x64.exe
else ifeq ($(DETECTED_OS),darwin)
EXE :=
ifeq ($(DETECTED_ARCH),arm64)
TAILWIND_ASSET := tailwindcss-macos-arm64
else
TAILWIND_ASSET := tailwindcss-macos-x64
endif
else
EXE :=
ifeq ($(DETECTED_ARCH),arm64)
TAILWIND_ASSET := tailwindcss-linux-arm64
else
TAILWIND_ASSET := tailwindcss-linux-x64
endif
endif

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------

# Where `go install` puts templ/air. Falls back to $HOME/go/bin if `go env`
# can't be read yet (e.g. Go isn't installed — `make help` should still work).
GOPATH_DIR := $(shell go env GOPATH 2>/dev/null)
ifeq ($(GOPATH_DIR),)
GOPATH_DIR := $(HOME)/go
endif
GOBIN := $(GOPATH_DIR)/bin

TEMPL := $(GOBIN)/templ$(EXE)
AIR   := $(GOBIN)/air$(EXE)

TOOLS_DIR    := .tools
TAILWIND_BIN := $(TOOLS_DIR)/tailwindcss$(EXE)

BUILD_DIR   := bin
BINARY_NAME := sadqa-ledger$(EXE)

CSS_INPUT  := web/static/css/input.css
CSS_OUTPUT := web/static/css/output.css

.DEFAULT_GOAL := help

.PHONY: help setup dev build test lint fmt templ css migrate clean \
        check-go install-tools tailwind-cli env-file deps

# ---------------------------------------------------------------------------
# Primary targets
# ---------------------------------------------------------------------------

help: ## Show this list of commands
	@echo "Sadqa Ledger — available commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  make %-14s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "New here? Run: make setup   then:   make dev"

setup: check-go install-tools tailwind-cli env-file deps templ css ## Prepare a fresh clone: check Go, install tools, fetch Tailwind CLI, create .env, fetch deps, build once
	@echo ""
	@echo "Setup complete — run 'make dev' to start the app with hot reload."
	@echo ""

dev: check-go install-tools tailwind-cli env-file deps ## Run the app with hot reload (templ + Tailwind + Go), one terminal, via air
	$(AIR)

build: templ tailwind-cli ## Build a production binary (bin/sadqa-ledger) with minified CSS
	@mkdir -p $(BUILD_DIR)
	$(TAILWIND_BIN) -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run all Go tests
	go test ./...

lint: ## Check formatting (gofmt) and static analysis (go vet)
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Not gofmt-formatted:"; \
		echo "$$UNFORMATTED"; \
		echo "Run 'make fmt' to fix, then try again."; \
		exit 1; \
	fi
	go vet ./...

fmt: ## Format all Go source files with gofmt
	gofmt -w .

templ: install-tools ## Generate Go code from .templ files
	$(TEMPL) generate

css: tailwind-cli ## Build the Tailwind CSS once (non-minified, for local dev)
	$(TAILWIND_BIN) -i $(CSS_INPUT) -o $(CSS_OUTPUT)

migrate: ## Run pending database migrations manually (the app also applies them automatically on startup)
	go run ./cmd/migrate

clean: ## Remove build artifacts, generated CSS, and air's tmp/ dir (never touches .env or *.db)
	rm -rf $(BUILD_DIR) tmp $(CSS_OUTPUT)
	@echo "Cleaned build artifacts. .env and your database were left untouched."
	@echo "(Downloaded tools in .tools/ and Go tools in $(GOBIN) are kept — delete those by hand if you ever want a full reset.)"

# ---------------------------------------------------------------------------
# Internal helpers (used as prerequisites above; safe to run directly too)
# ---------------------------------------------------------------------------

check-go: ## Verify Go is installed and meets the minimum version declared in go.mod
	@command -v go >/dev/null 2>&1 || { \
		echo "Go is not installed. Install Go $(GO_MIN_VERSION) or newer from https://go.dev/dl/ and re-run 'make setup'."; \
		exit 1; \
	}
	@GOV=$$(go env GOVERSION | sed 's/^go//'); \
	ver_ge() { \
		A_MAJ=$$(echo "$$1" | cut -d. -f1); A_MIN=$$(echo "$$1" | cut -d. -f2); A_PAT=$$(echo "$$1" | cut -d. -f3); \
		B_MAJ=$$(echo "$$2" | cut -d. -f1); B_MIN=$$(echo "$$2" | cut -d. -f2); B_PAT=$$(echo "$$2" | cut -d. -f3); \
		[ -z "$$A_PAT" ] && A_PAT=0; [ -z "$$B_PAT" ] && B_PAT=0; \
		[ -z "$$A_MIN" ] && A_MIN=0; [ -z "$$B_MIN" ] && B_MIN=0; \
		if [ "$$A_MAJ" -gt "$$B_MAJ" ]; then return 0; fi; \
		if [ "$$A_MAJ" -lt "$$B_MAJ" ]; then return 1; fi; \
		if [ "$$A_MIN" -gt "$$B_MIN" ]; then return 0; fi; \
		if [ "$$A_MIN" -lt "$$B_MIN" ]; then return 1; fi; \
		[ "$$A_PAT" -ge "$$B_PAT" ]; \
	}; \
	if ver_ge "$$GOV" "$(GO_MIN_VERSION)"; then \
		echo "Go $$GOV OK (minimum $(GO_MIN_VERSION), from go.mod)"; \
	else \
		echo "Go $$GOV found, but this project requires Go $(GO_MIN_VERSION) or newer (see go.mod)."; \
		echo "Install the latest Go from https://go.dev/dl/ and re-run 'make setup'."; \
		exit 1; \
	fi

install-tools: check-go ## Install pinned templ and air versions as Go tools (skips if already installed)
	@if [ -x "$(TEMPL)" ]; then \
		echo "templ already installed at $(TEMPL), skipping."; \
	else \
		echo "Installing templ $(TEMPL_VERSION)..."; \
		GOBIN="$(GOBIN)" go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION); \
	fi
	@if [ -x "$(AIR)" ]; then \
		echo "air already installed at $(AIR), skipping."; \
	else \
		echo "Installing air $(AIR_VERSION)..."; \
		GOBIN="$(GOBIN)" go install github.com/air-verse/air@$(AIR_VERSION); \
	fi

tailwind-cli: ## Download the Tailwind standalone CLI for this OS/architecture (skips if already downloaded)
	@mkdir -p $(TOOLS_DIR)
	@if [ -x "$(TAILWIND_BIN)" ]; then \
		echo "Tailwind CLI already present at $(TAILWIND_BIN), skipping download."; \
	else \
		command -v curl >/dev/null 2>&1 || { echo "curl is required to download the Tailwind CLI. Install curl and re-run 'make setup'."; exit 1; }; \
		echo "Downloading Tailwind CLI $(TAILWIND_VERSION) for $(DETECTED_OS)/$(DETECTED_ARCH) ($(TAILWIND_ASSET))..."; \
		curl -sL -o "$(TAILWIND_BIN)" "https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/$(TAILWIND_ASSET)"; \
		chmod +x "$(TAILWIND_BIN)"; \
		echo "Tailwind CLI ready at $(TAILWIND_BIN)"; \
	fi

env-file: ## Copy .env.example to .env if .env does not already exist (never overwrites)
	@if [ -f .env ]; then \
		echo ".env already exists, leaving it untouched."; \
	else \
		cp .env.example .env; \
		echo "Created .env from .env.example — edit it with your local values."; \
	fi

deps: ## Download Go module dependencies
	go mod download
