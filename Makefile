# Default target
.PHONY: all
all: build

MODULE_PATH := github.com/pmarchini/giogo

# Output directory for binaries
BIN_DIR := bin

# List of supported Linux architectures
ARCHS := \
    amd64 \
    arm64 \
    arm \
    386

# Feature toggles file
FT_FILE := .feature.toggles

# Read feature toggles from file and create -ldflags string
FT_FLAGS := $(shell if [ -f $(FT_FILE) ]; then \
	awk -F '=' '{print "-X $(MODULE_PATH)/ft." $$1 "=" $$2}' $(FT_FILE) | paste -sd " "; \
	fi)


# Build for the local architecture
.PHONY: build
build:
	@echo "Building giogo for local Linux architecture with feature toggles"
	go build -ldflags "$(FT_FLAGS)" -o giogo ./cmd/giogo

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts"
	rm -rf $(BIN_DIR)
	rm -f giogo

# Build for all supported Linux architectures
.PHONY: build-all
build-all: clean
	@echo "Building giogo for all Linux architectures with feature toggles"
	@mkdir -p $(BIN_DIR)
	@for arch in $(ARCHS); do \
		output=$(BIN_DIR)/giogo-linux-$$arch; \
		echo "Building $$output"; \
		GOOS=linux GOARCH=$$arch go build -ldflags "$(FT_FLAGS)" -o $$output ./cmd/giogo; \
	done

# Build for a specific Linux architecture
.PHONY: build-%
build-%:
	@arch=$*; \
	output=$(BIN_DIR)/giogo-linux-$$arch; \
	echo "Building $$output"; \
	GOOS=linux GOARCH=$$arch go build -ldflags "$(FT_FLAGS)" -o $$output ./cmd/giogo

.PHONY: test
test:
	@echo "Running all tests with coverage"
	go test -cover ./...
