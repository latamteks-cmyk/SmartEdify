# SmartEdify Auth Service Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop migrate-up migrate-down deps lint format

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  deps         - Download dependencies"
	@echo "  lint         - Run linter"
	@echo "  format       - Format code"

# Build the application
build:
	@echo "Building auth-service..."
	@go build -o bin/auth-service ./cmd/auth-service

# Run the application locally
run:
	@echo "Running auth-service..."
	@go run ./cmd/auth-service

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t smartedify/auth-service:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

# Stop Docker Compose
docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

# Run database migrations up
migrate-up:
	@echo "Running database migrations..."
	@migrate -path migrations -database "postgres://postgres:password@localhost:5432/smartedify_auth?sslmode=disable" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "postgres://postgres:password@localhost:5432/smartedify_auth?sslmode=disable" down

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@make deps
	@make docker-run
	@sleep 10
	@make migrate-up

# Generate mocks (if using mockery)
mocks:
	@echo "Generating mocks..."
	@mockery --all --output=mocks

# Security scan
security:
	@echo "Running security scan..."
	@gosec ./...

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Full CI pipeline
ci: deps lint test security

# Development workflow
dev: format lint test

# Production build
prod-build:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/auth-service ./cmd/auth-service