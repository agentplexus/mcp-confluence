.PHONY: build test lint install clean

# Build variables
BINARY_NAME=mcp-confluence
BUILD_DIR=./cmd/mcp-confluence
INSTALL_PATH=/usr/local/bin

# Default target
all: lint test build

# Build the binary
build:
	go build -o $(BINARY_NAME) $(BUILD_DIR)

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Install the binary
install: build
	cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

# Uninstall the binary
uninstall:
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run the server (requires env vars)
run:
	go run $(BUILD_DIR)

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build (default)"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  install       - Build and install to $(INSTALL_PATH)"
	@echo "  uninstall     - Remove installed binary"
	@echo "  clean         - Remove build artifacts"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy go.mod"
	@echo "  run           - Run the server"
