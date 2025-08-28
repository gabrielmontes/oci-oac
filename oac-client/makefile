# Makefile for oac-client

# Output binary
BINARY=dist/oac-client

# Go module
MODULE=oac-client

# Default target
.PHONY: all
all: build

# Build the CLI
.PHONY: build
build:
	@echo "Building $(BINARY)..."
	go mod tidy
	go build -o $(BINARY)

# Run linting with golangci-lint
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)

# Run CLI for testing
.PHONY: run
run:
	@echo "Running $(BINARY)..."
	$(BINARY) $(ARGS)