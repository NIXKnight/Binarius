# Binarius - Universal Binary Version Manager
# Makefile for build automation

.PHONY: build test clean lint fmt help smoke-test

# Binary name
BINARY_NAME=binarius

# Build directory
BUILD_DIR=.

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags="-s -w"

## help: Display this help message
help:
	@echo "Binarius - Universal Binary Version Manager"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binarius binary"
	@echo "  test        - Run all tests"
	@echo "  test-race   - Run tests with race detector"
	@echo "  smoke-test  - Run end-to-end smoke tests on built binary"
	@echo "  clean       - Remove built binaries and artifacts"
	@echo "  lint        - Run golangci-lint (requires golangci-lint installed)"
	@echo "  fmt         - Format Go source code"
	@echo "  tidy        - Tidy Go module dependencies"
	@echo "  help        - Display this help message"

## build: Compile the binarius binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run all tests with coverage
test:
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race ./...

## smoke-test: Run end-to-end smoke tests on built binary
## Usage: make smoke-test BINARY=./binarius-linux-amd64
smoke-test:
	@echo "Running smoke tests..."
	@./scripts/smoke-test.sh $(if $(BINARY),$(BINARY),./$(BINARY_NAME))

## test-coverage: Generate test coverage report
test-coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove binaries and build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Error: golangci-lint is not installed"; \
		echo "Install it from: https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	fi

## fmt: Format all Go source files
fmt:
	@echo "Formatting Go source files..."
	$(GOFMT) ./...
	@echo "Format complete"

## tidy: Tidy Go module dependencies
tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy
	@echo "Tidy complete"

## install: Install binarius to /usr/local/bin (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

## install-user: Install binarius to ~/.local/bin
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	mkdir -p ~/.local/bin
	mv $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@echo "Installation complete"
	@echo "Make sure ~/.local/bin is in your PATH"

## cross-compile: Build for multiple Linux architectures
cross-compile:
	@echo "Cross-compiling for Linux platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "Cross-compilation complete:"
	@ls -lh $(BINARY_NAME)-linux-*

.DEFAULT_GOAL := help
