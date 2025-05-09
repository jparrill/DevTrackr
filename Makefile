.PHONY: build clean test release release-local run lint deps dev help verify

# Build variables
BINARY_NAME=devtrackr
VERSION=$(shell git describe --tags --always --dirty)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Default target
all: deps build

# Build the application
build:
	@echo "Building DevTrackr..."
	@mkdir -p $(GOBIN)
	@go build -o $(GOBIN)/$(BINARY_NAME) ./cmd/devtrackr

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@rm -rf dist/
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application
run: build
	@echo "Running DevTrackr..."
	@$(GOBIN)/$(BINARY_NAME) serve

# Run in development mode with hot reload
dev:
	@echo "Running in development mode..."
	@go run ./cmd/devtrackr serve

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/goreleaser/goreleaser@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Verify code quality and tests
verify: deps lint test
	@echo "Verification completed successfully"

# Create a release
release:
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN is not set. Please export it first with:"; \
		echo "export GITHUB_TOKEN=<your_token_here> make release"; \
		exit 1; \
	fi
	@echo "Cleaning previous release artifacts..."
	@rm -rf dist/
	@echo "Creating release..."
	@goreleaser release

# Create a local release
release-local:
	@echo "Cleaning previous release artifacts..."
	@rm -rf dist/
	@echo "Creating local release snapshot..."
	@goreleaser release --snapshot --clean

# Create a release snapshot
release-snapshot:
	@echo "Creating release snapshot..."
	@goreleaser release --snapshot --rm-dist

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Install dependencies and build the application"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode with hot reload"
	@echo "  deps           - Install dependencies"
	@echo "  lint           - Run linter"
	@echo "  verify         - Run deps, lint and test in sequence"
	@echo "  release        - Create a release"
	@echo "  release-local  - Create a local release"
	@echo "  release-snapshot - Create a release snapshot"