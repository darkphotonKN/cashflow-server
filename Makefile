.PHONY: help dev build test lint clean deps air install-air

# Default target
help:
	@echo "Available commands:"
	@echo "  make dev          - Run with hot reload (air)"
	@echo "  make build        - Build binary"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make install-air  - Install air for hot reloading"

# Hot reload development server
dev: install-air
	@echo "Starting development server with hot reload..."
	air

# Build the binary
build:
	@echo "Building cashflow server..."
	go build -o bin/cashflow ./cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf tmp/

# Install air for hot reloading
install-air:
	@command -v air >/dev/null 2>&1 || { \
		echo "Installing air for hot reloading..."; \
		go install github.com/cosmtrek/air@latest; \
	}

# Install golangci-lint
install-lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}

# Setup development environment
setup: deps install-air install-lint
	@echo "Development environment setup complete!"

# Run the built binary
run: build
	@echo "Running cashflow server..."
	./bin/cashflow

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t cashflow:latest .

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Full check (format, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

# Release build (optimized)
release:
	@echo "Building optimized release binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/cashflow-linux-amd64 ./cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/cashflow-darwin-amd64 ./cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/cashflow-darwin-arm64 ./cmd/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/cashflow-windows-amd64.exe ./cmd/main.go

# Generate mocks (if using gomock)
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Security check
security:
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	}
	@echo "Running security check..."
	gosec ./...

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...