.PHONY: build test run clean deps help install-local fmt lint generate clean-generated

BINARY_NAME=invoice-management
OPENAPI_SPEC=internal/assets/openapi.yaml
BUILD_DIR=./bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: deps build test

# Download and tidy dependencies
deps:
	go mod download
	go mod tidy

# Build the Go binary
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/main.go

# Run tests with 30s timeout
test:
	go test -v -timeout 30s ./...

# Run tests with coverage output
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic -timeout 30s ./...

# Run the server directly (no build)
run:
	go run ./cmd/server/main.go

# Run the built binary
run-bin: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run E2E API tests
test-e2e:
	go test -v -timeout 30s ./e2e/api/...

# Install locally to /usr/local/bin
install-local: clean build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out
	go clean

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Security scan
sec:
	gosec ./...

# Generate code from OpenAPI spec
generate:
	@echo "Generating code from OpenAPI spec..."
	@mkdir -p internal/api/generated
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		--config oapi-codegen/types.cfg.yaml \
		$(OPENAPI_SPEC)
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		--config oapi-codegen/server.cfg.yaml \
		$(OPENAPI_SPEC)
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		--config oapi-codegen/client.cfg.yaml \
		$(OPENAPI_SPEC)
	@echo "Code generation complete"

# Clean generated files
clean-generated:
	rm -f internal/api/generated/*.gen.go

# Show help
help:
	@echo "Invoice Management System - Makefile Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  build        - Build the project"
	@echo "  run          - Run the server directly"
	@echo "  run-bin      - Build and run the binary"
	@echo "  install-local - Install to /usr/local/bin (requires sudo)"
	@echo ""
	@echo "Testing:"
	@echo "  test         - Run all tests (30s timeout)"
	@echo "  test-e2e     - Run E2E API tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  sec          - Run security scan"
	@echo ""
	@echo "Code Generation:"
	@echo "  generate     - Generate code from OpenAPI spec"
	@echo "  clean-generated - Remove generated files"
	@echo ""
	@echo "Other:"
	@echo "  deps         - Download dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"
