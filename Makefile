# ==============================================================================
# Variables
# ==============================================================================

# Project paths
PROJECT_DIR := v2k-mac
FRONTEND_DIR := $(PROJECT_DIR)/frontend
BIN_DIR := $(PROJECT_DIR)/backend/dependency/bin/darwin_arm64
SCRIPT_DIR := $(PROJECT_DIR)/scripts

# Tools
WAILS := wails
GO := go
NPM := npm
NODE := node

# App Details
APP_NAME := v2k-mac
# Try to extract version, default to 0.0.0 if not found
VERSION := $(shell grep "version": $(PROJECT_DIR)/wails.json 2>/dev/null | sed 's/.*: \(.*\)\'",/\1/' || echo "0.0.0")

# Shell settings
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# ==============================================================================
# Targets
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%%-20s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------------------
# Setup & Dependencies
# ------------------------------------------------------------------------------

.PHONY: setup
setup: check-env install-wails deps-backend deps-frontend deps-binaries ## Full project setup (tools, libs, binaries)

.PHONY: check-env
check-env: ## Check if required tools are installed
	@which $(GO) > /dev/null || (echo "Error: go is not installed" && exit 1)
	@which $(NPM) > /dev/null || (echo "Error: npm is not installed" && exit 1)
	@echo "✅ Environment checks passed."

.PHONY: install-wails
install-wails: ## Install Wails CLI
	@echo "Installing Wails..."
	$(GO) install github.com/wailsapp/wails/v2/cmd/wails@latest

.PHONY: deps-backend
deps-backend: ## Install Go dependencies
	@echo "Downloading Go dependencies..."
	cd $(PROJECT_DIR) && $(GO) mod download && $(GO) mod tidy

.PHONY: deps-frontend
deps-frontend: ## Install Frontend dependencies
	@echo "Installing Frontend dependencies..."
	cd $(FRONTEND_DIR) && $(NPM) install

.PHONY: deps-binaries
deps-binaries: ## Download external binaries (yt-dlp, ffmpeg) via script
	@echo "Downloading external binaries..."
	cd $(PROJECT_DIR) && chmod +x scripts/download_deps.sh && ./scripts/download_deps.sh

# ------------------------------------------------------------------------------
# Development
# ------------------------------------------------------------------------------

.PHONY: dev
dev: ## Run the app in development mode (hot reload)
	cd $(PROJECT_DIR) && $(WAILS) dev

.PHONY: lint
lint: ## Run linters (Go vet & NPM lint if configured)
	@echo "Linting Go..."
	cd $(PROJECT_DIR) && $(GO) vet ./...
	@echo "Linting Frontend..."
	# Assuming a lint script exists in package.json, otherwise skipping
	cd $(FRONTEND_DIR) && if [ -n "$$($(NPM) run | grep lint)" ]; then $(NPM) run lint; fi

# ------------------------------------------------------------------------------
# Testing
# ------------------------------------------------------------------------------

.PHONY: test
test: test-backend test-frontend ## Run all unit tests

.PHONY: test-backend
test-backend: ## Run Go unit tests
	@echo "Running Backend Tests..."
	cd $(PROJECT_DIR) && $(GO) test -v ./...

.PHONY: test-frontend
test-frontend: ## Run Frontend Unit Tests (Vitest)
	@echo "Running Frontend Tests..."
	cd $(FRONTEND_DIR) && $(NPM) test -- run

.PHONY: test-e2e
test-e2e: ## Run End-to-End Tests (Playwright)
	@echo "Running E2E Tests..."
	cd $(FRONTEND_DIR) && $(NPM) run test:e2e

# ------------------------------------------------------------------------------
# Build & Release
# ------------------------------------------------------------------------------

.PHONY: build
build: clean ## Build the production application (macOS .app)
	@echo "Building $(APP_NAME) v$(VERSION)..."
	cd $(PROJECT_DIR) && $(WAILS) build -clean -platform darwin/arm64

.PHONY: clean
clean: ## Remove build artifacts and temp files
	@echo "Cleaning build artifacts..."
	rm -rf $(PROJECT_DIR)/build/bin
	rm -rf $(PROJECT_DIR)/frontend/dist
	rm -rf debug/

.PHONY: release
release: test build ## Run tests and build for release
	@echo "Release build complete."
	@echo "App location: $(PROJECT_DIR)/build/bin/$(APP_NAME).app"
	@# Optional: Add logic here to zip the .app or create a DMG
	@# zip -r release/$(APP_NAME)-$(VERSION).zip $(PROJECT_DIR)/build/bin/$(APP_NAME).app
