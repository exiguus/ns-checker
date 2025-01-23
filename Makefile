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

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./... -v -cover -race

# Run AMD64 version
.PHONY: run
run: build-amd64
	@echo "Running AMD64 version..."
	@$(AMD64_DIR)/$(BINARY_NAME) listen

# Run with specific port
.PHONY: run-port
run-port: build-amd64
	@echo "Running AMD64 version on port $(port)..."
	@$(AMD64_DIR)/$(BINARY_NAME) listen $(port)

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
