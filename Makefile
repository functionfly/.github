.PHONY: help build test clean docker-up docker-down dev api health-monitor migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	go build -o bin/orchestrator-api ./cmd/orchestrator-api
	go build -o bin/health-monitor ./cmd/health-monitor

test: ## Run tests
	go test ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start docker services
	docker-compose up -d

docker-down: ## Stop docker services
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

dev: ## Start development environment
	docker-compose up postgres -d
	@echo "Waiting for postgres to be ready..."
	@sleep 5
	go run ./cmd/orchestrator-api

api: ## Run orchestrator API
	go run ./cmd/orchestrator-api

health-monitor: ## Run health monitor service
	go run ./cmd/health-monitor

migrate: ## Run database migrations
	go run -tags=dev ./cmd/orchestrator-api migrate

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

deps: ## Download dependencies
	go mod download
	go mod tidy