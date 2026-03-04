SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help build build-local-runtime test clean docker-up docker-down dev api health-monitor migrate migrate-down migrate-status migrate-version wasm-bundle staging-up staging-down staging-logs staging-migrate staging-api staging-health-monitor test-db-setup test-db-up test-db-migrate test-db-status test-api-cmds load-test-init load-test-tpcb load-test-mixed load-test-custom load-test-stress bench bench-db bench-db-profile db-maintenance venv

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	go build -o bin/orchestrator-api ./cmd/orchestrator-api
	go build -o bin/health-monitor ./cmd/health-monitor
	go build -o bin/ffly ./cmd/ffly

build-local-runtime: ## Build the local Rust runtime
	cd runtimes/local && cargo build --release
	cp runtimes/local/target/release/functionfly-local bin/

build-fly: ## Build the fly CLI (bin/fly)
	go build -o bin/fly ./cmd/fly

venv: ## Create .venv for local dev (Python from .python-version) and install functions/functionfly/requirements.txt. Run: source .venv/bin/activate
	@python3 --version 2>/dev/null || { echo "Python 3 required (pyenv recommended: pyenv install 3.12)"; exit 1; }; \
	rm -rf .venv && python3 -m venv .venv && .venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r functions/functionfly/requirements.txt && \
	echo "Done. Activate with: source .venv/bin/activate"

test-functions: ## Run unit tests for functions/functionfly (stdlib handlers)
	@cd functions/functionfly && (test -d .venv || python3 -m venv .venv) && .venv/bin/pip install -q -r requirements-test.txt && .venv/bin/pytest tests/unit -v -m "not e2e" --tb=short

test-functions-e2e: ## Run unit + e2e tests for functions/functionfly
	@cd functions/functionfly && (test -d .venv || python3 -m venv .venv) && .venv/bin/pip install -q -r requirements-test.txt && .venv/bin/pytest tests/ -v --tb=short

publish-stdlib: build-fly ## Dev login then publish all functions in functions/functionfly. Requires FFLY_API_URL, FFLY_DEV_EMAIL, FFLY_DEV_PASSWORD.
	@test -n "$$FFLY_API_URL" || (echo "FFLY_API_URL is required (e.g. http://localhost:8080)"; exit 1)
	@test -n "$$FFLY_DEV_EMAIL" || (echo "FFLY_DEV_EMAIL is required (e.g. admin@functionfly.local)"; exit 1)
	@test -n "$$FFLY_DEV_PASSWORD" || (echo "FFLY_DEV_PASSWORD is required"; exit 1)
	FFLY_API_URL=$$FFLY_API_URL FFLY_DEV_EMAIL=$$FFLY_DEV_EMAIL FFLY_DEV_PASSWORD=$$FFLY_DEV_PASSWORD ./bin/fly login --dev
	FFLY_API_URL=$$FFLY_API_URL ./bin/fly publish-batch functions/functionfly --conflict-strategy overwrite --concurrency 5

wasm-bundle: ## Bundle function to Wasm for testing
	go run ./cmd/ffly bundle --wasm

test: ## Run tests
	go test ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-coverage-report: ## Generate detailed coverage report
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tee coverage.txt
	@echo "Coverage report generated: coverage.html and coverage.txt"

test-coverage-analysis: ## Run comprehensive coverage analysis
	./scripts/test-coverage.sh

test-integration-coverage: ## Run integration tests with coverage
	go test -v -race -tags=integration -coverprofile=coverage-integration.out ./...
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	go tool cover -func=coverage-integration.out

test-all: ## Run all tests (unit + integration) with coverage
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage-unit.out ./...
	@echo "Running integration tests..."
	go test -v -race -tags=integration -coverprofile=coverage-integration.out ./...
	@echo "Generating coverage reports..."
	go tool cover -html=coverage-unit.out -o coverage-unit.html
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	go tool cover -func=coverage-unit.out | grep total | tee coverage-unit.txt
	go tool cover -func=coverage-integration.out | grep total | tee coverage-integration.txt
	@echo "Coverage reports generated"

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./...

test-api-cmds: ## Run API smoke tests (health + login). Set API_URL for base (default http://localhost:8080)
	./scripts/test-cmds.sh

bench: ## Run all benchmarks
	go test -bench=. -benchmem ./...

bench-db: ## Run database benchmarks only
	go test -bench=. -benchmem ./internal/storage/... -benchtime=5s

bench-db-profile: ## Run database benchmarks with CPU profiling
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/storage/...
	go tool pprof -web cpu.prof

lint: ## Run linter
	golangci-lint run

security-scan: ## Run security vulnerability scan
	gosec ./...
	govulncheck ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start docker services
	docker compose up -d

docker-down: ## Stop docker services
	docker compose down

docker-logs: ## Show docker logs
	docker compose logs -f

dev: ## Start development environment (local Postgres + Redis, no Docker). Set DB_PORT=5434 for Docker Postgres. Start Prometheus with: docker compose up -d prometheus (then status page will show component health).
	@echo "Using local services: DB_PORT=$${DB_PORT:-5432}, REDIS_ADDR=$${REDIS_ADDR:-localhost:6379}, PROMETHEUS_URL=$${PROMETHEUS_URL:-http://127.0.0.1:9091}"
	@if command -v infisical >/dev/null 2>&1; then \
		DEVELOPMENT=true infisical run --env=dev -- env DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} PROMETHEUS_URL=$${PROMETHEUS_URL:-http://127.0.0.1:9091} \
		SKIP_MIGRATION_VALIDATION=true \
		go run ./cmd/orchestrator-api; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} PROMETHEUS_URL=$${PROMETHEUS_URL:-http://127.0.0.1:9091} DEVELOPMENT=true \
		SKIP_MIGRATION_VALIDATION=true \
		go run ./cmd/orchestrator-api; \
	fi

api: ## Run orchestrator API (local services; use infisical if available). Set DB_PORT=5434 for Docker Postgres.
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- env DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} \
		SKIP_MIGRATION_VALIDATION=true \
		go run ./cmd/orchestrator-api; \
	else \
		./scripts/run-api-local.sh; \
	fi

health-monitor: ## Run health monitor service (local DB/Redis). Set DB_PORT=5434 for Docker Postgres.
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- env DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} \
		SKIP_MIGRATION_VALIDATION=true \
		go run ./cmd/health-monitor; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} \
		SKIP_MIGRATION_VALIDATION=true \
		go run ./cmd/health-monitor; \
	fi

migrate: ## Run database migrations (up). Uses local DB by default (DB_PORT=5432). Set DB_PORT=5434 for Docker Postgres.
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- env DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/migrate up; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/migrate up; \
	fi

migrate-down: ## Rollback database migrations
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- go run ./cmd/migrate down; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/migrate down; \
	fi

migrate-status: ## Show migration status
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- go run ./cmd/migrate status; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/migrate status; \
	fi

migrate-version: ## Show current migration version
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- go run ./cmd/migrate version; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/migrate version; \
	fi

setup: ## Setup initial data (tenant, user)
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- go run ./cmd/setup; \
	else \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
		go run ./cmd/setup; \
	fi

seed-blog: ## Seed default blog post (State Fabric) into DB. Uses DB_* from env or defaults (load .env first if needed).
	@test -f scripts/seed-blog-post-state-fabric.sql || (echo "Missing scripts/seed-blog-post-state-fabric.sql"; exit 1)
	@PGPASSWORD=$${DB_PASSWORD:-postgres} psql -h $${DB_HOST:-localhost} -p $${DB_PORT:-5432} -U $${DB_USER:-postgres} -d $${DB_NAME:-functionfly} -f scripts/seed-blog-post-state-fabric.sql

seed-blog-docker: ## Seed blog post using Docker Postgres (port 5434). Use if your app runs with docker compose and DB_PORT=5434.
	@test -f scripts/seed-blog-post-state-fabric.sql || (echo "Missing scripts/seed-blog-post-state-fabric.sql"; exit 1)
	@PGPASSWORD=$${DB_PASSWORD:-postgres} psql -h $${DB_HOST:-localhost} -p 5434 -U $${DB_USER:-postgres} -d $${DB_NAME:-functionfly} -f scripts/seed-blog-post-state-fabric.sql

update-blog-from-md: ## Update State Fabric blog post content from content/blog/introducing-state-fabric.md. Run after seed-blog. Uses DB_* from env.
	@./scripts/update-blog-post-from-markdown.sh

update-blog-from-md-docker: ## Same as update-blog-from-md but for Docker Postgres (port 5434).
	@DB_PORT=5434 ./scripts/update-blog-post-from-markdown.sh

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

deps: ## Download dependencies
	go mod download
	go mod tidy

monitoring-up: ## Start monitoring stack
	docker compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d prometheus grafana node-exporter postgres-exporter redis-exporter

monitoring-down: ## Stop monitoring stack
	docker compose -f docker-compose.monitoring.yml down

monitoring-logs: ## Show monitoring logs
	docker compose -f docker-compose.monitoring.yml logs -f

monitoring-restart: ## Restart monitoring stack
	docker compose -f docker-compose.monitoring.yml restart

grafana-init: ## Initialize Grafana with default dashboard
	@echo "Grafana will be available at http://localhost:3000"
	@echo "Default credentials: admin/admin"
	@echo "Dashboard will be automatically provisioned"

prometheus-init: ## Show Prometheus configuration
	@echo "Prometheus configuration:"
	@cat deploy/monitoring/prometheus.yml

# Database commands
db-prod-up: ## Start production database stack (primary + replica + pgbouncer)
	docker compose -f docker-compose.production.yml up -d postgres postgres-replica pgbouncer

db-prod-down: ## Stop production database stack
	docker compose -f docker-compose.production.yml down

db-replica-setup: ## Setup and configure read replica
	@echo "Setting up PostgreSQL read replica..."
	./scripts/setup-read-replica.sh

db-replica-status: ## Check read replica status and lag
	@echo "=== Primary Replication Status ==="
	@docker compose -f docker-compose.production.yml exec postgres psql -U functionfly_prod -d functionfly_prod -c "SELECT * FROM pg_stat_replication;" 2>/dev/null || echo "Primary not accessible"
	@echo ""
	@echo "=== Replica Lag Status ==="
	@docker compose -f docker-compose.production.yml exec postgres-replica psql -U functionfly_prod -d functionfly_prod -c "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;" 2>/dev/null || echo "Replica not accessible"

db-prod-logs: ## Show production database logs
	docker compose -f docker-compose.production.yml logs -f

db-prod-backup: ## Create database backup
	./scripts/backup-database.sh production

db-prod-restore: ## Restore database from backup (requires BACKUP_FILE env var)
	@echo "Restoring from: $$BACKUP_FILE"
	@PGPASSWORD=$$DB_PASSWORD pg_restore \
		--host=localhost \
		--port=6432 \
		--username=functionfly_prod \
		--dbname=functionfly_prod \
		--verbose \
		$$BACKUP_FILE

db-restore: ## Restore database using the enhanced script (BACKUP_SPEC defaults to 'latest')
	./scripts/restore-database.sh production $(BACKUP_SPEC)

db-backup-list: ## List available backups
	@echo "Local backups:"
	@ls -la /var/backups/functionfly/ 2>/dev/null || echo "No local backups found"
	@echo ""
	@echo "Remote backups (configure DB_BACKUP_STORAGE_BACKEND):"
	@case "$$DB_BACKUP_STORAGE_BACKEND" in \
		"s3") aws s3 ls s3://$$DB_BACKUP_S3_BUCKET/$$DB_BACKUP_S3_PREFIX --recursive | grep functionfly || echo "No S3 backups found" ;; \
		"b2") b2 ls $$DB_BACKUP_B2_BUCKET $$DB_BACKUP_S3_PREFIX | grep functionfly || echo "No B2 backups found" ;; \
		*) echo "Configure DB_BACKUP_STORAGE_BACKEND to list remote backups" ;; \
	esac

db-maintenance: ## Run database maintenance (analyze/vacuum) to optimize performance
	@echo "Running database maintenance..."
	@if command -v infisical >/dev/null 2>&1; then \
		infisical run --env=dev -- ./scripts/db-maintenance.sh; \
	else \
		./scripts/db-maintenance.sh; \
	fi

db-migrate-prod: ## Run migrations on production database
	@echo "Running migrations on production database..."
	infisical run --env=prod -- go run ./cmd/migrate up

# Logging commands
logging-up: ## Start logging stack (Loki, Elasticsearch, Kibana, Fluent Bit)
	docker compose -f docker-compose.monitoring.yml up -d loki promtail elasticsearch kibana fluent-bit

logging-down: ## Stop logging stack
	docker compose -f docker-compose.monitoring.yml down

logging-logs: ## Show logging stack logs
	docker compose -f docker-compose.monitoring.yml logs -f loki promtail elasticsearch kibana fluent-bit

logs-view: ## View application logs
	tail -f logs/*.log

logs-clean: ## Clean old log files (older than 7 days)
	find logs/ -name "*.log" -mtime +7 -delete

# Staging environment commands
staging-up: ## Start staging environment
	docker compose -f docker-compose.staging.yml --env-file .env.staging up -d

staging-down: ## Stop staging environment
	docker compose -f docker-compose.staging.yml --env-file .env.staging down

staging-logs: ## Show staging environment logs
	docker compose -f docker-compose.staging.yml --env-file .env.staging logs -f

staging-restart: ## Restart staging environment
	docker compose -f docker-compose.staging.yml --env-file .env.staging restart

staging-build: ## Build staging containers
	docker compose -f docker-compose.staging.yml --env-file .env.staging build

staging-migrate: ## Run migrations on staging database
	@echo "Running migrations on staging database..."
	DB_HOST=$$(grep DB_HOST .env.staging | cut -d '=' -f2) \
	DB_PORT=$$(grep DB_PORT .env.staging | cut -d '=' -f2) \
	DB_USER=$$(grep DB_USER .env.staging | cut -d '=' -f2) \
	DB_PASSWORD=$$(grep DB_PASSWORD .env.staging | cut -d '=' -f2) \
	DB_NAME=$$(grep DB_NAME .env.staging | cut -d '=' -f2) \
	DB_SSLMODE=$$(grep DB_SSLMODE .env.staging | cut -d '=' -f2) \
	go run ./cmd/migrate run

staging-api: ## Run staging API server locally (not in Docker)
	@echo "Starting staging API server..."
	DB_HOST=$$(grep DB_HOST .env.staging | cut -d '=' -f2) \
	DB_PORT=$$(grep DB_PORT .env.staging | cut -d '=' -f2) \
	DB_USER=$$(grep DB_USER .env.staging | cut -d '=' -f2) \
	DB_PASSWORD=$$(grep DB_PASSWORD .env.staging | cut -d '=' -f2) \
	DB_NAME=$$(grep DB_NAME .env.staging | cut -d '=' -f2) \
	DB_SSLMODE=$$(grep DB_SSLMODE .env.staging | cut -d '=' -f2) \
	USE_SUPABASE=false \
	REDIS_ADDR=$$(grep REDIS_ADDR .env.staging | cut -d '=' -f2) \
	go run ./cmd/orchestrator-api

staging-health-monitor: ## Run staging health monitor locally
	@echo "Starting staging health monitor..."
	DB_HOST=$$(grep DB_HOST .env.staging | cut -d '=' -f2) \
	DB_PORT=$$(grep DB_PORT .env.staging | cut -d '=' -f2) \
	DB_USER=$$(grep DB_USER .env.staging | cut -d '=' -f2) \
	DB_PASSWORD=$$(grep DB_PASSWORD .env.staging | cut -d '=' -f2) \
	DB_NAME=$$(grep DB_NAME .env.staging | cut -d '=' -f2) \
	DB_SSLMODE=$$(grep DB_SSLMODE .env.staging | cut -d '=' -f2) \
	USE_SUPABASE=false \
	go run ./cmd/health-monitor

staging-psql: ## Connect to staging database with psql
	@echo "Connecting to staging database..."
	DB_HOST=$$(grep DB_HOST .env.staging | cut -d '=' -f2) \
	DB_PORT=$$(grep DB_PORT .env.staging | cut -d '=' -f2) \
	DB_USER=$$(grep DB_USER .env.staging | cut -d '=' -f2) \
	DB_PASSWORD=$$(grep DB_PASSWORD .env.staging | cut -d '=' -f2) \
	DB_NAME=$$(grep DB_NAME .env.staging | cut -d '=' -f2) \
	psql "postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=require"

staging-status: ## Show staging environment status
	@echo "=== Staging Environment Status ==="
	@echo "Docker containers:"
	@docker compose -f docker-compose.staging.yml --env-file .env.staging ps
	@echo ""
	@echo "Database connection:"
	@DB_HOST=$$(grep DB_HOST .env.staging | cut -d '=' -f2) \
	DB_PORT=$$(grep DB_PORT .env.staging | cut -d '=' -f2) \
	DB_USER=$$(grep DB_USER .env.staging | cut -d '=' -f2) \
	DB_PASSWORD=$$(grep DB_PASSWORD .env.staging | cut -d '=' -f2) \
	DB_NAME=$$(grep DB_NAME .env.staging | cut -d '=' -f2) \
	psql "postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=require" -c "SELECT version();" 2>/dev/null | head -3 || echo "Database connection failed"

# Test Database commands
test-db-up: ## Start test database container
	docker compose up -d postgres
	@echo "Waiting for test database to be ready..."
	@sleep 5

test-db-setup: ## Create test database
	docker compose exec postgres psql -U postgres -c "CREATE DATABASE functionfly_test;" 2>/dev/null || echo "Database functionfly_test may already exist"

test-db-migrate: ## Run migrations on test database
	@echo "Running migrations on test database..."
	DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly_test DB_SSLMODE=disable go run ./cmd/migrate run

test-db-status: ## Show test database status
	@echo "Test database tables:"
	@docker compose exec postgres psql -U postgres -d functionfly_test -c "\dt" 2>/dev/null || echo "Test database not accessible"

test-db-setup-all: test-db-up test-db-setup test-db-migrate test-db-status ## Set up complete test database environment
	@echo "Test database setup complete!"

# Load Testing Commands
load-test-init: ## Initialize load testing database
	@echo "Initializing load testing database..."
	./scripts/pgbench-load-test.sh production 50 10 2 60 init

load-test-tpcb: ## Run TPC-B benchmark (standard OLTP test)
	@echo "Running TPC-B load test..."
	./scripts/pgbench-load-test.sh production 50 10 2 60 tpcb

load-test-mixed: ## Run mixed read/write load test
	@echo "Running mixed load test..."
	./scripts/pgbench-load-test.sh production 50 20 4 120 mixed

load-test-custom: ## Run custom load test based on app queries
	@echo "Running custom application load test..."
	./scripts/pgbench-load-test.sh production 50 15 3 180 custom

load-test-stress: ## Run high-concurrency stress test
	@echo "Running high-concurrency stress test..."
	./scripts/pgbench-load-test.sh production 100 50 8 300 tpcb

# Performance Testing Commands
perf-routing-bench: ## Run routing engine benchmarks
	@echo "Running routing performance benchmarks..."
	go test -bench=. -benchmem ./internal/routing/...

perf-statefabric-bench: ## Run StateFabric Rust benchmarks
	@echo "Running StateFabric performance benchmarks..."
	cd statefabric && cargo bench

perf-api-load-test: ## Run API load tests with Artillery
	@echo "Running API load tests..."
	@if command -v artillery >/dev/null 2>&1; then \
		artillery run load-tests/api-load-test.yml --output artillery-report.json; \
	else \
		echo "Artillery not installed. Install with: npm install -g artillery"; \
		exit 1; \
	fi

perf-k6-load-test: ## Run advanced load tests with k6
	@echo "Running advanced load tests with k6..."
	@if command -v k6 >/dev/null 2>&1; then \
		k6 run load-tests/k6-load-test.js; \
	else \
		echo "k6 not installed. Install from: https://k6.io/docs/get-started/installation/"; \
		exit 1; \
	fi

perf-distributed-test: ## Run distributed load tests
	@echo "Running distributed load tests..."
	@if command -v python3 >/dev/null 2>&1; then \
		python3 load-tests/distributed-load-test.py; \
	else \
		echo "Python 3 not available"; \
		exit 1; \
	fi

perf-all-tests: perf-routing-bench perf-statefabric-bench perf-api-load-test perf-k6-load-test ## Run all performance tests
	@echo "All performance tests completed!"
	@echo "Check the following for results:"
	@echo "- artillery-report.json (API load test results)"
	@echo "- statefabric/target/criterion/ (Rust benchmark results)"
	@echo "- load-test-results/ (distributed test results)"
