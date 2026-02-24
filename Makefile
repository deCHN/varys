# ==============================================================================
# Variables
# ==============================================================================

# Project paths
PROJECT_DIR := .
FRONTEND_DIR := frontend

# Tools
WAILS := wails
GO := go
NPM := npm

# App Details
APP_NAME := Varys
# Try to extract version, default to 0.0.0 if not found
VERSION := $(shell grep '"version":' $(PROJECT_DIR)/wails.json 2>/dev/null | sed -E 's/.*"version": "(.*)".*/\1/' || echo "0.0.0")

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
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------------------
# Setup & Dependencies
# ------------------------------------------------------------------------------

.PHONY: setup
setup: check-env install-wails deps-backend deps-frontend ## Full project setup (tools, libs)

.PHONY: check-env
check-env: ## Check if required tools are installed
	@which $(GO) > /dev/null || (echo "Error: go is not installed" && exit 1)
	@which $(NPM) > /dev/null || (echo "Error: npm is not installed" && exit 1)
	@echo "Environment checks passed."

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
	cd $(FRONTEND_DIR) && if [ -n "$$($(NPM) run | grep lint)" ]; then $(NPM) run lint; fi

# ------------------------------------------------------------------------------
# Testing
# ------------------------------------------------------------------------------

TIMEOUT ?= 20s
TIMEOUT_INT ?= 5m

.PHONY: test
test: test-unit test-integration test-frontend ## Run all tests (unit, integration, frontend)

.PHONY: test-backend
test-backend: test-unit test-integration ## Alias for all backend tests

.PHONY: test-unit
test-unit: ## Run Go unit tests (excludes benchmarks and integration tests)
	@echo "Running Unit Tests (Timeout: $(TIMEOUT))..."
	cd $(PROJECT_DIR) && $(GO) test -v -timeout $(TIMEOUT) $$(go list ./... | grep -vE "benchmarks|tests/integration")

.PHONY: test-integration
test-integration: ## Run Go integration tests
	@echo "Running Integration Tests (Timeout: $(TIMEOUT_INT))..."
	cd $(PROJECT_DIR) && $(GO) test -v -timeout $(TIMEOUT_INT) -tags integration ./...

.PHONY: test-benchmark
test-benchmark: ## Run Go performance benchmarks
	@echo "Running Performance Benchmarks (this may take time)..."
	cd $(PROJECT_DIR) && $(GO) test -v ./backend/benchmarks/...

.PHONY: test-coverage
test-coverage: ## Run unit tests with coverage report
	@echo "Running Test Coverage..."
	cd $(PROJECT_DIR) && $(GO) test -coverprofile=coverage.out $$(go list ./backend/... | grep -v benchmarks)
	@$(GO) tool cover -func=coverage.out

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

BUILD_LOG := .build.log

.PHONY: build
build: clean ## Build the production application (macOS .app)
	@echo -n "  • Building $(APP_NAME) v$(VERSION) (GUI)... "
	@cd $(PROJECT_DIR) && $(WAILS) build -clean -platform darwin/arm64 > $(BUILD_LOG) 2>&1 || (echo "FAILED (see $(BUILD_LOG))" && exit 1)
	@echo "DONE"

.PHONY: build-cli
build-cli: ## Build the standalone CLI binary
	@echo -n "  • Building $(APP_NAME) v$(VERSION) (CLI)... "
	@$(GO) build -o build/bin/varys-cli ./cmd/cli/main.go > $(BUILD_LOG) 2>&1 || (echo "FAILED (see $(BUILD_LOG))" && exit 1)
	@echo "DONE"

.PHONY: install-cli
install-cli: build-cli ## Install Varys CLI to GOPATH/bin
	@echo -n "  • Installing CLI to $(shell go env GOPATH)/bin... "
	@mkdir -p $(shell go env GOPATH)/bin
	@cp ./build/bin/varys-cli $(shell go env GOPATH)/bin/varys-cli
	@echo "DONE"

.PHONY: clean
clean: ## Remove build artifacts and temp files
	@echo -n "  • Cleaning build artifacts... "
	@rm -rf $(PROJECT_DIR)/build/bin > /dev/null 2>&1
	@rm -rf $(PROJECT_DIR)/frontend/dist > /dev/null 2>&1
	@rm -rf debug/ > /dev/null 2>&1
	@rm -f $(BUILD_LOG) > /dev/null 2>&1
	@echo "DONE"

.PHONY: release
release: test build build-cli ## Run tests and build for release
	@echo "Release build complete."
	@echo "App location: $(PROJECT_DIR)/build/bin/$(APP_NAME).app"
	@echo "CLI location: $(PROJECT_DIR)/build/bin/varys-cli"

.PHONY: install
install: build build-cli ## Build and Install the app and CLI
	@echo "  • Preparation: stopping $(APP_NAME) if running."
	@-pkill -x "$(APP_NAME)" > /dev/null 2>&1 || true
	@echo -n "  • Installing $(APP_NAME).app to /Applications... "
	@rm -rf "/Applications/$(APP_NAME).app"
	@cp -R "$(PROJECT_DIR)/build/bin/$(APP_NAME).app" /Applications/
	@echo "DONE"
	@$(MAKE) -s install-cli
	@echo ""
	@echo "🎉 $(APP_NAME) v$(VERSION) installed successfully."