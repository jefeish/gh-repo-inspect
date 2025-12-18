# Makefile for gh repo-inspect

.PHONY: build install clean test lint

# Build the extension
build:
	go build -o gh-repo-inspect .

# Install the extension
install: build
	gh extension install .

# Install for development (creates symlink)
install-dev:
	gh extension install . --pin

# Clean build artifacts
clean:
	rm -f gh-repo-inspect

# Run tests with timeout (only utils package to avoid GitHub CLI hanging)
test:
	go test -timeout 10s -v ./utils

# Run all tests (may hang due to GitHub CLI library)
test-all:
	go test -timeout 30s -v ./...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt lint test

# Build for multiple platforms
build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/gh-repo-inspect-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/gh-repo-inspect-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/gh-repo-inspect-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/gh-repo-inspect-windows-amd64.exe .

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the extension"
	@echo "  install    - Install the extension"
	@echo "  install-dev - Install for development"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  fmt        - Format code"
	@echo "  check      - Run all checks"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  help       - Show this help"