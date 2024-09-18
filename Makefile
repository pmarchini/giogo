# Makefile for building giogo for Linux architectures

# Default target
.PHONY: all
all: build

# The module path (replace with your actual module path)
MODULE_PATH := github.com/yourusername/giogo

# Output directory for binaries
BIN_DIR := bin

# List of supported Linux architectures
ARCHS := \
    amd64 \
    arm64 \
    arm \
    386

# Build for the local architecture
.PHONY: build
build:
	@echo "Building giogo for local Linux architecture"
	go build -o giogo ./cmd/giogo

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts"
	rm -rf $(BIN_DIR)
	rm -f giogo

# Build for all supported Linux architectures
.PHONY: build-all
build-all: clean
	@echo "Building giogo for all Linux architectures"
	@mkdir -p $(BIN_DIR)
	@for arch in $(ARCHS); do \
		output=$(BIN_DIR)/giogo-linux-$$arch; \
		echo "Building $$output"; \
		GOOS=linux GOARCH=$$arch go build -o $$output ./cmd/giogo; \
	done

# Build for a specific Linux architecture
.PHONY: build-%
build-%:
	@arch=$*; \
	output=$(BIN_DIR)/giogo-linux-$$arch; \
	echo "Building $$output"; \
	GOOS=linux GOARCH=$$arch go build -o $$output ./cmd/giogo

.PHONY: test
test:
	@echo "Running all tests with coverage"
	go test -cover ./...
