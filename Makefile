.PHONY: build build-docker clean test lint run-lite run-standard run-enterprise \
        docker-lite docker-standard docker-enterprise docker-down install-cli \
        deps help

# Variables
APP_NAME=adaptive-cron
VERSION=1.0.0
BIN_DIR=bin
CONFIGS_DIR=configs

# Default target
.DEFAULT_GOAL := help

# Build targets
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/adaptive-cron
	@echo "Build complete!"

build-docker: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "Docker build complete!"

# Run targets
run-lite: build ## Run with lite profile
	./$(BIN_DIR)/$(APP_NAME) --profile lite

run-standard: build ## Run with standard profile
	./$(BIN_DIR)/$(APP_NAME) --profile standard --config $(CONFIGS_DIR)/standard.yaml

run-enterprise: build ## Run with enterprise profile
	./$(BIN_DIR)/$(APP_NAME) --profile enterprise --config $(CONFIGS_DIR)/enterprise.yaml

run-custom: build ## Run with custom config
	./$(BIN_DIR)/$(APP_NAME) --config $(CONFIGS_DIR)/custom.yaml

# Docker compose targets
docker-lite: build-docker ## Run lite profile with Docker
	docker-compose -f docker-compose.yml -f docker-compose.lite.yml up

docker-standard: build-docker ## Run standard profile with Docker
	docker-compose -f docker-compose.yml -f docker-compose.standard.yml up

docker-enterprise: build-docker ## Run enterprise profile with Docker
	docker-compose -f docker-compose.yml -f docker-compose.enterprise.yml up

docker-down: ## Stop all Docker containers
	docker-compose -f docker-compose.yml -f docker-compose.lite.yml down
	docker-compose -f docker-compose.yml -f docker-compose.standard.yml down
	docker-compose -f docker-compose.yml -f docker-compose.enterprise.yml down

# Testing
test: ## Run all tests
	go test -v -race ./...

test-unit: ## Run unit tests
	go test -v -short ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Linting
lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

# Clean
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)/
	rm -f coverage.out
	docker-compose -f docker-compose.yml -f docker-compose.lite.yml down -v
	docker-compose -f docker-compose.yml -f docker-compose.standard.yml down -v
	docker-compose -f docker-compose.yml -f docker-compose.enterprise.yml down -v
	docker rmi $(APP_NAME):$(VERSION) $(APP_NAME):latest 2>/dev/null || true

# CLI installation
install-cli: build ## Install CLI to /usr/local/bin
	@echo "Installing $(APP_NAME) CLI..."
	sudo cp $(BIN_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "Installation complete!"

# Dependencies
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: deps build ## Setup development environment
	@echo "Development environment ready!"
	@echo "Run 'make run-lite' to start with lite profile"
	@echo "Run 'make run-standard' to start with standard profile"
	@echo "Run 'make run-enterprise' to start with enterprise profile"

# Migration
migrate: ## Run database migrations
	go run ./migrations/main.go up

migrate-down: ## Rollback migrations
	go run ./migrations/main.go down

# Version
version: ## Show version
	@echo "$(APP_NAME) version $(VERSION)"

# Profile list
profiles: ## Show available profiles
	./$(BIN_DIR)/$(APP_NAME) --list-profiles || true