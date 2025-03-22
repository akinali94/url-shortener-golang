# Makefile for URL Shortener with ID Generator

# Variables
GO=go
GOFLAGS=-v
BUILD_DIR=build
CMD_DIR=cmd
SRC_DIR=src

# Service names
URLSHORTENER=urlshortener
IDGENERATOR=idgenerator
COMMAND_RUNNER=runner

# Build all services
.PHONY: all
all: clean build-all

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Create build directory
.PHONY: init
init:
	@echo "Creating build directory..."
	mkdir -p $(BUILD_DIR)

# Build all services
.PHONY: build-all
build-all: init build-urlshortener build-idgenerator build-runner

# Build URL Shortener service
.PHONY: build-urlshortener
build-urlshortener: init
	@echo "Building URL Shortener service..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(URLSHORTENER) ./$(SRC_DIR)/$(URLSHORTENER)

# Build ID Generator service
.PHONY: build-idgenerator
build-idgenerator: init
	@echo "Building ID Generator service..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(IDGENERATOR) ./$(SRC_DIR)/$(IDGENERATOR)

# Build runner (cmd)
.PHONY: build-runner
build-runner: init
	@echo "Building command runner..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(COMMAND_RUNNER) ./$(CMD_DIR)

# Run URL Shortener directly
.PHONY: run-urlshortener
run-urlshortener: build-urlshortener
	@echo "Starting URL Shortener service..."
	./$(BUILD_DIR)/$(URLSHORTENER)

# Run ID Generator directly
.PHONY: run-idgenerator
run-idgenerator: build-idgenerator
	@echo "Starting ID Generator service..."
	./$(BUILD_DIR)/$(IDGENERATOR)

# Run both services using the command runner
.PHONY: run-all
run-all: build-runner
	@echo "Starting both services using runner..."
	./$(BUILD_DIR)/$(COMMAND_RUNNER) --all

# Run URL Shortener using the command runner
.PHONY: run-urlshortener-cmd
run-urlshortener-cmd: build-runner
	@echo "Starting URL Shortener using runner..."
	./$(BUILD_DIR)/$(COMMAND_RUNNER) --service=$(URLSHORTENER)

# Run ID Generator using the command runner
.PHONY: run-idgenerator-cmd
run-idgenerator-cmd: build-runner
	@echo "Starting ID Generator using runner..."
	./$(BUILD_DIR)/$(COMMAND_RUNNER) --service=$(IDGENERATOR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod tidy

# Help command
.PHONY: help
help:
	@echo "URL Shortener with ID Generator Makefile commands:"
	@echo "  make build-all             - Build all services"
	@echo "  make build-urlshortener    - Build URL Shortener service only"
	@echo "  make build-idgenerator     - Build ID Generator service only"
	@echo "  make build-runner          - Build command runner only"
	@echo "  make run-urlshortener      - Run URL Shortener service directly"
	@echo "  make run-idgenerator       - Run ID Generator service directly" 
	@echo "  make run-all               - Run both services using command runner"
	@echo "  make run-urlshortener-cmd  - Run URL Shortener via command runner"
	@echo "  make run-idgenerator-cmd   - Run ID Generator via command runner"
	@echo "  make clean                 - Remove build artifacts"
	@echo "  make test                  - Run tests"
	@echo "  make deps                  - Install dependencies"