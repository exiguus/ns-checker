# Build directories
BUILD_DIR := build
AMD64_DIR := $(BUILD_DIR)/amd64
ARM7_DIR := $(BUILD_DIR)/armv7
BINARY_NAME := ns-checker

# Go build flags
GOOS := linux
GO_BUILD := go build -v

# Default target
.PHONY: all
all: clean build-all

# Build all architectures
.PHONY: build-all
build-all: build-amd64 build-arm7

# Build AMD64
.PHONY: build-amd64
build-amd64:
	@echo "Building for AMD64..."
	@mkdir -p $(AMD64_DIR)
	GOOS=$(GOOS) GOARCH=amd64 $(GO_BUILD) -o $(AMD64_DIR)/$(BINARY_NAME)

# Build ARM7
.PHONY: build-arm7
build-arm7:
	@echo "Building for ARMv7..."
	@mkdir -p $(ARM7_DIR)
	GOOS=$(GOOS) GOARCH=arm GOARM=7 $(GO_BUILD) -o $(ARM7_DIR)/$(BINARY_NAME)

# Clean build directory
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Environment variables for testing
ENV_VARS := \
	DNS_PORT=45353 \
	WORKER_COUNT=8 \
	RATE_LIMIT=200 \
	RATE_BURST=400 \
	CACHE_TTL=10m \
	CACHE_CLEANUP=20m \
	HEALTH_CHECK_PORT=9090 \
	LOGS_DIR=./logs \
	LOG_FILE=dns_listener.log

# Run tests with proper flags
.PHONY: test
test:
	$(ENV_VARS) GODEBUG=netdns=go go test ./... \
		-v \
		-failfast \
		-timeout=1m \
		-parallel=4 \
		-count=1

# Run DNS listener tests
.PHONY: test-listener
test-listener:
	$(ENV_VARS) go test -v ./dns_listener/... \
		-failfast \
		-timeout=1m \
		-parallel=4 \
		-count=1

.PHONY: test-typo
test-typo:
	go test -v ./dns_typo_checker/... \
		-failfast \
		-timeout=1m \
		-parallel=4 \
		-count=1

# Build
.PHONY: build
build:
	go build -v ./...

# Run AMD64 version
.PHONY: run
run: build-amd64
	@echo "Running AMD64 version..."
	@$(AMD64_DIR)/$(BINARY_NAME) listen

# Run with specific port
.PHONY: run-port
run-port: build-amd64
	@echo "Running AMD64 version on port $(port)..."
	@$(ENV_VARS) $(AMD64_DIR)/$(BINARY_NAME) listen $(port)

.PHONY: start-docker
start-docker:
	@echo "Running Docker container..."
	@docker compose up --build --force-recreate --detach --remove-orphans

.PHONY: stop-docker
stop-docker:
	@echo "Stopping Docker container..."
	@docker compose down

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make all          - Clean and build all architectures"
	@echo "  make build-all    - Build all architectures"
	@echo "  make build-amd64  - Build AMD64 version"
	@echo "  make build-arm7   - Build ARM7 version"
	@echo "  make clean        - Clean build directory"
	@echo "  make test         - Run tests"
	@echo "  make run          - Build and run AMD64 version"
	@echo "  make run-port port=53  - Run with specific port"
	@echo "  make start-docker   - Start Docker container"
	@echo "  make stop-docker  - Stop Docker container"
	@echo "  make help         - Show this help"
