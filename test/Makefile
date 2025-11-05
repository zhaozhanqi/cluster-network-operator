# Makefile for CNO E2E Test

# Binary name
BINARY_NAME := cno-e2e-test

# Build directory
BUILD_DIR := build

# Go module and package
GO_MODULE := github.com/openshift/cluster-network-operator
GO_PKG := $(GO_MODULE)/test/cmd

# Version information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE_STATE := $(shell if git diff --quiet 2>/dev/null; then echo "clean"; else echo "dirty"; fi)

# Go build flags
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
LDFLAGS := -X '$(GO_MODULE)/pkg/version.commitFromGit=$(GIT_COMMIT)' \
           -X '$(GO_MODULE)/pkg/version.buildDate=$(BUILD_DATE)' \
           -X '$(GO_MODULE)/pkg/version.versionFromGit=$(GIT_COMMIT)'

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	cd cmd && go build -o ../$(BUILD_DIR)/$(BINARY_NAME) -ldflags="$(LDFLAGS)" .

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	cd cmd && GO_COMPLIANCE_POLICY=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags="$(LDFLAGS)" .

# Build for macOS
.PHONY: build-darwin
build-darwin:
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	cd cmd && GOOS=darwin GOARCH=amd64 \
		go build -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags="$(LDFLAGS)" .

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin

# Run e2e tests
.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	go test -v ./e2e/...

# Run all tests
.PHONY: test
test:
	@echo "Running all tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	cd .. && go mod download
	cd .. && go mod tidy

# Vendor dependencies
.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	cd .. && go mod vendor

# Run the binary
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Verify build
.PHONY: verify
verify: fmt lint test

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Build the binary (default)"
	@echo "  build         - Build the binary for current platform"
	@echo "  build-linux   - Build for Linux AMD64"
	@echo "  build-darwin  - Build for macOS AMD64"
	@echo "  build-all     - Build for all platforms"
	@echo "  test          - Run all tests"
	@echo "  test-e2e      - Run E2E tests only"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  vendor        - Vendor dependencies"
	@echo "  run           - Build and run the binary"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  verify        - Format, lint, and test"
	@echo "  help          - Show this help"
