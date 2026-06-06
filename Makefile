.PHONY: build run test clean migrate deps docker-build docker-up docker-down install-cli

APP_NAME=adaptive-cron
VERSION=1.0.0

build:
	@echo "Building all services..."
	go build -o bin/task-manager ./cmd/task-manager
	go build -o bin/scheduler ./cmd/scheduler
	go build -o bin/metric-agent ./cmd/metric-agent
	go build -o bin/worker ./cmd/worker
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/taskctl ./cmd/taskctl
	@echo "Build complete!"

run: build
	@echo "Starting services..."
	./bin/api-gateway &
	./bin/task-manager &
	./bin/scheduler &
	./bin/metric-agent &
	./bin/worker &
	@echo "All services started"

test:
	go test -v -race ./...

test-unit:
	go test -v -short ./...

test-integration:
	go test -v -tags=integration ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

deps:
	go mod download
	go mod tidy

migrate:
	@echo "Running migrations..."
	go run migrations/migrate.go up

lint:
	golangci-lint run

docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Services started. API Gateway available at http://localhost:8080"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

install-cli:
	@echo "Installing taskctl..."
	go build -o /usr/local/bin/taskctl ./cmd/taskctl
	@echo "TaskCtl installed successfully!"

dev: docker-up
	@echo "Development environment ready!"
	@echo "API Gateway: http://localhost:8080"
	@echo "Task Manager: http://localhost:8081"
	@echo "Scheduler: http://localhost:8082"
	@echo "Metric Agent: http://localhost:8083"
	@echo "Worker: http://localhost:8084"

status:
	@echo "Checking service status..."
	@curl -s http://localhost:8080/health | jq . || echo "API Gateway not running"
	@curl -s http://localhost:8081/health | jq . || echo "Task Manager not running"
	@curl -s http://localhost:8082/health | jq . || echo "Scheduler not running"
	@curl -s http://localhost:8083/api/v1/metrics/health | jq . || echo "Metric Agent not running"
	@curl -s http://localhost:8084/health | jq . || echo "Worker not running"

help:
	@echo "Available commands:"
	@echo "  make build          - Build all services"
	@echo "  make run            - Run all services locally"
	@echo "  make test           - Run all tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-up      - Start all services with Docker"
	@echo "  make docker-down    - Stop all Docker services"
	@echo "  make install-cli    - Install taskctl CLI"
	@echo "  make dev            - Start development environment"
	@echo "  make status         - Check service status"