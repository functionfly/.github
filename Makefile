SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Go build cache configuration (override via env)
GOCACHE ?= $(shell go env GOCACHE)
GOMODCACHE ?= $(shell go env GOMODCACHE)

# Build optimization flags
BUILD_FLAGS := -trimpath
GC_FLAGS := -gcflags="-e"
LD_FLAGS := -ldflags "-s -w"

# Test configuration
TEST_PARALLEL := -parallel=8
TEST_COUNT := -count=1

# Detect available CPUs for parallel builds
NCPU := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

.PHONY: help build build-fast build-ci build-local-runtime build-microvm-orchestrator build-python-runtime test test-short test-parallel test-fast test-changed test-watch clean docker-up docker-down dev dev-neon api api-local health-monitor migrate migrate-local migrate-down migrate-status migrate-version wasm-bundle staging-up staging-down staging-logs staging-migrate staging-api staging-health-monitor test-db-setup test-db-up test-db-migrate test-db-status test-api-cmds load-test-init load-test-tpcb load-test-mixed load-test-custom load-test-stress bench bench-db bench-db-profile db-maintenance venv build-fly build-fly-release release-dry-run release release-snapshot install-locally dist build-coming-soon deploy-coming-soon deploy-admin-dashboard

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services (optimized with trimpath)
	go build $(BUILD_FLAGS) -o bin/orchestrator-api ./cmd/orchestrator-api
	go build $(BUILD_FLAGS) -o bin/health-monitor ./cmd/health-monitor
	go build $(BUILD_FLAGS) -o bin/ff ./cmd/fly

build-fast: ## Fast build for development (allows multiple errors, smaller binaries)
	go build $(BUILD_FLAGS) $(GC_FLAGS) -o bin/orchestrator-api ./cmd/orchestrator-api
	go build $(BUILD_FLAGS) $(GC_FLAGS) -o bin/health-monitor ./cmd/health-monitor
	go build $(BUILD_FLAGS) $(GC_FLAGS) -o bin/ff ./cmd/fly

build-ci: ## CI-optimized build (cached, parallel, no cgo)
	CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LD_FLAGS) -o bin/orchestrator-api ./cmd/orchestrator-api
	CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LD_FLAGS) -o bin/health-monitor ./cmd/health-monitor
	CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LD_FLAGS) -o bin/ff ./cmd/fly

build-local-runtime: ## Build the local Rust runtime
	cd runtimes/local && cargo build --release
	cp runtimes/local/target/release/functionfly-local bin/

build-sar: ## Build the SAR (Stateful Agent Runtime)
	cd runtimes/sar && cargo build --release
	cp runtimes/sar/target/release/functionfly-sar bin/

build-microvm-orchestrator: ## Build MicroVM orchestrator (HTTP on :9091; set FUNCTIONFLY_MICROVM_DEV_MODE=1 for host Python without Firecracker)
	cd runtimes/microvm && cargo build --release
	cp runtimes/microvm/target/release/functionfly-microvm bin/

build-microvm-rootfs: ## Build CPython 3.11 ext4 rootfs for Firecracker (requires Docker + root). Set PYTHON_VER=3.12 for Python 3.12.
	@echo "Building Firecracker rootfs (Python $${PYTHON_VER:-3.11})..."
	@mkdir -p bin/vmimages
	cd runtimes/microvm/images && bash build-rootfs.sh $${PYTHON_VER:-3.11} $(CURDIR)/bin/vmimages
	@echo "Rootfs ready: bin/vmimages/python$${PYTHON_VER:-3.11}.ext4"
	@echo "Next: copy a Firecracker-compatible vmlinux kernel to bin/vmimages/"

build-python-runtime: ## Build the Python WASM runtime service (requires CGO for wasmtime)
	@echo "Building Python WASM runtime service..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/python-runtime ./cmd/python-runtime
	@echo "Python WASM runtime built: bin/python-runtime"
	@echo "Start with: ./bin/python-runtime [--port 8083]"

dev-microvm: build-microvm-orchestrator ## Run MicroVM orchestrator in dev mode (host CPython, no Firecracker)
	FUNCTIONFLY_MICROVM_DEV_MODE=1 ./bin/functionfly-microvm \
		--port 9091 \
		--image-path bin/vmimages \
		--max-vms 4 \
		--warm-idle-secs 120 \
		--debug

nats: ## Start NATS server on port 4222 (idempotent — skips if already running)
	@if pgrep -x nats-server > /dev/null 2>&1; then \
		echo "NATS already running (PID: $$(pgrep -x nats-server))"; \
	else \
		nohup nats-server -p 4222 > /tmp/nats.log 2>&1 & \
		sleep 1 && echo "NATS started on port 4222 (PID: $$(pgrep -x nats-server))"; \
	fi

nats-stop: ## Stop NATS server
	@if pgrep -x nats-server > /dev/null 2>&1; then \
		pkill nats-server && echo "NATS stopped"; \
	else \
		echo "NATS not running"; \
	fi

dev-sar: nats ## Start SAR (Stateful Agent Runtime) on port 8082
	@echo "Starting SAR on port $${SAR_PORT:-8082}..."
	@NATS_URL="$${NATS_URL:-nats://localhost:4222}" \
	 REDIS_URL="$${REDIS_URL:-}" \
	 DATABASE_URL="$${DATABASE_URL:-}" \
	 cargo run --manifest-path runtimes/sar/Cargo.toml

run-microvm: build-microvm-orchestrator ## Run MicroVM orchestrator in production mode (requires Firecracker + VM images in MICROVM_IMAGE_PATH)
	@[ -f "$${MICROVM_IMAGE_PATH:-bin/vmimages}/python311.ext4" ] || \
		{ echo "ERROR: rootfs not found. Run: make build-microvm-rootfs"; exit 1; }
	@[ -n "$${FUNCTIONFLY_MICROVM_API_TOKEN:-}" ] || \
		{ echo "ERROR: FUNCTIONFLY_MICROVM_API_TOKEN must be set for production"; exit 1; }
	./bin/functionfly-microvm \
		--image-path "$${MICROVM_IMAGE_PATH:-bin/vmimages}" \
		--port "$${MICROVM_PORT:-9091}" \
		--max-vms "$${MICROVM_MAX_VMS:-20}"

build-ff: ## Build the ff CLI (bin/ff)
	go build $(BUILD_FLAGS) -o bin/ff ./cmd/fly

build-ff-fast: ## Fast build of ff CLI for development
	go build $(BUILD_FLAGS) $(GC_FLAGS) -o bin/ff ./cmd/fly

build-ff-release: ## Build the ff CLI with version ldflags (for release-like local binary)
	@v=$$(git describe --tags --always 2>/dev/null || echo "0.0.0"); \
	c=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	d=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	go build $(BUILD_FLAGS) -ldflags "-s -w -X github.com/functionfly/functionfly/internal/version.Version=$$v -X github.com/functionfly/functionfly/internal/version.Commit=$$c -X github.com/functionfly/functionfly/internal/version.Date=$$d" ./cmd/fly

release-dry-run: ## Run GoReleaser in dry-run mode (no publish)
	goreleaser release --clean --dry-run

release: ## Create and publish a CLI release (requires GITHUB_TOKEN, tag e.g. v1.0.0)
	goreleaser release --clean

release-snapshot: ## Create a snapshot release (no tag required)
	goreleaser release --clean --snapshot

install-locally: ## Install ff CLI to GOPATH/bin (optimized)
	go install $(BUILD_FLAGS) ./cmd/fly

dist: ## Build distribution packages for current platform only (no publish)
	goreleaser build --clean --single-target

build-all-modules: ## Build all workspace modules
	go build $(BUILD_FLAGS) ./cmd/...
	go build $(BUILD_FLAGS) ./internal/...

build-coming-soon: ## Build static coming-soon page for deploy (output: web/dashboard/dist). Uses web/coming-soon/index.html. Set API_URL to override feedback API (default https://api.functionfly.com).
	@rm -rf web/dashboard/dist && mkdir -p web/dashboard/dist && cp -r web/coming-soon/* web/dashboard/dist/ && \
	([ -z "$${API_URL:-}" ] || sed -i 's|<html lang="en">|<html lang="en" data-api-url="'"$$API_URL"'">|' web/dashboard/dist/index.html) && \
	echo "Coming-soon built: web/dashboard/dist (API: $${API_URL:-https://api.functionfly.com})"

deploy-coming-soon: build-coming-soon ## Build and deploy coming-soon to Cloudflare Pages. Requires CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID in .env (see docs/DOMAIN_AND_COMING_SOON_SETUP.md).
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$${CLOUDFLARE_API_TOKEN:-}" ]; then echo "ERROR: CLOUDFLARE_API_TOKEN not set. Add to .env or export it."; exit 1; fi; \
	if [ -z "$${CLOUDFLARE_ACCOUNT_ID:-}" ]; then echo "ERROR: CLOUDFLARE_ACCOUNT_ID not set (avoids API 10001). Add to .env: CLOUDFLARE_ACCOUNT_ID=your_account_id_from_dashboard_url"; exit 1; fi; \
	npx wrangler pages project create functionfly-dashboard 2>/dev/null || true; \
	npx wrangler pages deploy web/dashboard/dist --project-name=functionfly-dashboard --branch=master

deploy-admin-dashboard: ## Build admin dashboard (vite) and deploy to Cloudflare Pages (project: functionfly-admin-dashboard). Requires CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID in .env (same as deploy-coming-soon).
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$${CLOUDFLARE_API_TOKEN:-}" ]; then echo "ERROR: CLOUDFLARE_API_TOKEN not set. Add to .env or export it."; exit 1; fi; \
	if [ -z "$${CLOUDFLARE_ACCOUNT_ID:-}" ]; then echo "ERROR: CLOUDFLARE_ACCOUNT_ID not set. Add to .env (see deploy-coming-soon)." ; exit 1; fi; \
	cd web/admin-dashboard && (bunx wrangler pages project create functionfly-admin-dashboard --production-branch=master 2>/dev/null || true) && \
	bun run pages:deploy

venv: ## Create .venv for local dev (Python from .python-version) and install functions/functionfly/requirements.txt. Run: source .venv/bin/activate
	@python3 --version 2>/dev/null || { echo "Python 3 required (pyenv recommended: pyenv install 3.12)"; exit 1; }; \
	rm -rf .venv && python3 -m venv .venv && .venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r functions/functionfly/requirements.txt && \
	echo "Done. Activate with: source .venv/bin/activate"

test-functions: ## Run unit tests for functions/functionfly (stdlib handlers)
	@cd functions/functionfly && (test -d .venv || python3 -m venv .venv) && .venv/bin/pip install -q -r requirements-test.txt && .venv/bin/pytest tests/unit -v -m "not e2e" --tb=short

test-functions-e2e: ## Run unit + e2e tests for functions/functionfly
	@cd functions/functionfly && (test -d .venv || python3 -m venv .venv) && .venv/bin/pip install -q -r requirements-test.txt && .venv/bin/pytest tests/ -v --tb=short

publish-stdlib: build-fly ## Dev login then publish all functions in functions/functionfly. Requires FFLY_API_URL, FFLY_DEV_EMAIL, FFLY_DEV_PASSWORD. API must be running with DEVELOPMENT=true (see AGENTS.md) or restarted after middleware changes.
	@test -n "$$FFLY_API_URL" || (echo "FFLY_API_URL is required (e.g. http://localhost:8080)"; exit 1)
	@test -n "$$FFLY_DEV_EMAIL" || (echo "FFLY_DEV_EMAIL is required (e.g. admin@functionfly.local)"; exit 1)
	@test -n "$$FFLY_DEV_PASSWORD" || (echo "FFLY_DEV_PASSWORD is required"; exit 1)
	FFLY_API_URL=$$FFLY_API_URL FFLY_DEV_EMAIL=$$FFLY_DEV_EMAIL FFLY_DEV_PASSWORD=$$FFLY_DEV_PASSWORD ./bin/fly login --dev
	FFLY_API_URL=$$FFLY_API_URL ./bin/fly publish-batch functions/functionfly --conflict-strategy overwrite --concurrency 5

wasm-bundle: ## Bundle function to Wasm for testing
	go run ./cmd/ffly bundle --wasm

test: ## Run tests (cached, parallel)
	go test $(TEST_PARALLEL) ./...

test-short: ## Run tests in short mode (skip heavy integration tests)
	go test -short $(TEST_PARALLEL) ./...

test-parallel: ## Run tests with high parallelism
	go test -parallel=$(NCPU) ./...

test-fast: ## Fast test for local dev (cached, parallel, short)
	go test -short $(TEST_PARALLEL) -count=1 ./...

test-changed: ## Run tests only for changed packages (requires gotestsum)
	@which gotestsum > /dev/null || (echo "Installing gotestsum..." && go install gotest.tools/gotestsum@latest)
	gotestsum --format=dots -- -short $(TEST_PARALLEL) ./...

test-watch: ## Run tests in watch mode (requires gotestsum)
	@which gotestsum > /dev/null || (echo "Installing gotestsum..." && go install gotest.tools/gotestsum@latest)
	gotestsum --watch --format=dots -- -short $(TEST_PARALLEL) ./...

test-ci: ## CI-optimized tests (rerun fails, dots format, parallel)
	@which gotestsum > /dev/null || (echo "Installing gotestsum..." && go install gotest.tools/gotestsum@latest)
	gotestsum --rerun-fails=3 --format=dots -- -race $(TEST_PARALLEL) -count=1 ./...

check-deps: ## Verify and tidy dependencies
	go mod verify
	go mod tidy
	git diff --exit-code go.mod go.sum || (echo "Error: go.mod or go.sum changed. Run 'go mod tidy' and commit." && exit 1)

test-coverage: ## Run tests with coverage (parallel)
	go test -v -race $(TEST_PARALLEL) -coverprofile=coverage.out ./...
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

test-integration: ## Run integration tests (parallel)
	go test -v -race $(TEST_PARALLEL) -tags=integration ./...

test-integration-short: ## Run integration tests in short mode (faster)
	go test -short -v -race $(TEST_PARALLEL) -tags=integration ./...

test-api-cmds: ## Run API smoke tests (health + login). Set API_URL for base (default http://localhost:8080)
	./scripts/test-cmds.sh

bench: ## Run all benchmarks
	go test -bench=. -benchmem $(TEST_PARALLEL) ./...

bench-db: ## Run database benchmarks only
	go test -bench=. -benchmem $(TEST_PARALLEL) ./internal/storage/... -benchtime=5s

bench-db-profile: ## Run database benchmarks with CPU profiling
	go test -bench=. -benchmem $(TEST_PARALLEL) -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/storage/...
	go tool pprof -web cpu.prof

lint: ## Run linter
	golangci-lint run

security-scan: ## Run security vulnerability scan
	gosec ./...
	govulncheck ./...

clean: ## Clean build artifacts (bin/ and any binaries/logs in project root)
	rm -rf bin/
	rm -f api-test create-admin ffly fly flypy-test functionfly functionfly-server health-monitor migrate orchestrator-api server setup-bin simple_server *.test
	rm -f api_startup.log api_test.log dev_test.log runtime.log server_output.log coverage.out

docker-up: ## Start docker services
	docker compose up -d

docker-down: ## Stop docker services
	docker compose down

docker-logs: ## Show docker logs
	docker compose logs -f

dev-neon: ## Start API with Neon DB (no local Postgres). Starts FlyMind (ai-service) on :8081 if needed. See AGENTS.md.
	@( set -a; [ -f .env ] && . ./.env; set +a; \
	export REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} DEVELOPMENT=true DEVELOPMENT_ALLOW_NONLOCAL_HOST=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false; \
	echo "Starting API (Neon DB, Redis $$REDIS_ADDR) + FlyMind if needed..."; \
	exec ./scripts/run-orchestrator-with-ai.sh ./bin/orchestrator-api --skip-migrations )

dev-local: ## Start API with local Postgres + Redis + FlyMind. No Neon/Upstash needed. Set DB_PORT=5434 for Docker Postgres.
	@echo "Using local services: DB_PORT=$${DB_PORT:-5432}, REDIS_ADDR=$${REDIS_ADDR:-localhost:6379}"
	@DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
	DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
	REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} DEVELOPMENT=true DEVELOPMENT_ALLOW_NONLOCAL_HOST=true \
	SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false \
	JWT_SECRET=$${JWT_SECRET:-dev-jwt-secret-not-for-production} \
	./scripts/run-orchestrator-with-ai.sh ./bin/orchestrator-api --skip-migrations

dev: ## Start development environment (Orchestrator + SAR + NATS). Set DB_PORT=5434 for Docker Postgres.
	@echo "Using local services: DB_PORT=$${DB_PORT:-5432}, REDIS_ADDR=$${REDIS_ADDR:-localhost:6379}"
	@# Start NATS if not already running
	@if ! pgrep -x nats-server > /dev/null 2>&1; then \
		echo "Starting NATS on port 4222..."; \
		nohup nats-server -p 4222 > /tmp/nats.log 2>&1 & \
		sleep 1; \
	fi
	@# Start SAR in background
	@echo "Starting SAR (Stateful Agent Runtime) on port $${SAR_PORT:-8082}..."
	@NATS_URL="$${NATS_URL:-nats://localhost:4222}" nohup cargo run --manifest-path runtimes/sar/Cargo.toml > /tmp/sar.log 2>&1 & echo "SAR PID: $$!"
	@# Start Orchestrator API (foreground)
	@DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
	DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
	REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} PROMETHEUS_URL=$${PROMETHEUS_URL:-http://127.0.0.1:9091} DEVELOPMENT=true \
	SKIP_MIGRATION_VALIDATION=true \
	go run ./cmd/orchestrator-api

api: ## Run orchestrator API (local services). Set DB_PORT=5434 for Docker Postgres.
	@./scripts/run-api-local.sh

api-local: ## Run orchestrator API without Infisical (uses run-api-local.sh). Use when Infisical returns 403 or is unavailable.
	./scripts/run-api-local.sh

health-monitor: ## Run health monitor service (local DB/Redis). Set DB_PORT=5434 for Docker Postgres.
	@DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
	DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
	REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} \
	SKIP_MIGRATION_VALIDATION=true \
	go run ./cmd/health-monitor

dev-python-runtime: build-python-runtime ## Start Python WASM runtime service (requires CGO for wasmtime)
	@echo "Starting Python WASM runtime on port $${PYTHON_RUNTIME_PORT:-8083}..."
	@CPYTHON_WASM_PATH=./runtimes/cpython-wasi/python.wasm \
	PYTHON_RUNTIME_PORT=$${PYTHON_RUNTIME_PORT:-8083} \
	POOL_SIZE=$${POOL_SIZE:-4} \
	MAX_MEMORY_MB=$${MAX_MEMORY_MB:-512} \
	./bin/python-runtime

dev: ## Start development environment (Orchestrator + SAR + NATS + Python Runtime). Set DB_PORT=5434 for Docker Postgres.
	@echo "Using local services: DB_PORT=$${DB_PORT:-5432}, REDIS_ADDR=$${REDIS_ADDR:-localhost:6379}"
	@# Start NATS if not already running
	@if ! pgrep -x nats-server > /dev/null 2>&1; then \
		echo "Starting NATS on port 4222..."; \
		nohup nats-server -p 4222 > /tmp/nats.log 2>&1 & \
		sleep 1; \
	fi
	@# Start SAR in background
	@echo "Starting SAR (Stateful Agent Runtime) on port $${SAR_PORT:-8082}..."
	@NATS_URL="$${NATS_URL:-nats://localhost:4222}" nohup cargo run --manifest-path runtimes/sar/Cargo.toml > /tmp/sar.log 2>&1 & echo "SAR PID: $$!"
	@# Start Python WASM runtime in background (requires CGO)
	@if [ "$${SKIP_PYTHON_RUNTIME:-0}" != "1" ]; then \
		echo "Starting Python WASM runtime on port $${PYTHON_RUNTIME_PORT:-8083}..."; \
		mkdir -p bin && \
		CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/python-runtime ./cmd/python-runtime 2>/dev/null && \
		CPYTHON_WASM_PATH=./runtimes/cpython-wasi/python.wasm \
		PYTHON_RUNTIME_PORT=$${PYTHON_RUNTIME_PORT:-8083} \
		POOL_SIZE=4 \
		MAX_MEMORY_MB=512 \
		nohup ./bin/python-runtime > /tmp/python-runtime.log 2>&1 & \
		echo "Python Runtime PID: $$!"; \
	fi
	@# Start Orchestrator API (foreground)
	@DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
	DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-functionfly} DB_SSLMODE=$${DB_SSLMODE:-disable} \
	REDIS_ADDR=$${REDIS_ADDR:-localhost:6379} PROMETHEUS_URL=$${PROMETHEUS_URL:-http://127.0.0.1:9091} DEVELOPMENT=true \
	PYTHON_RUNTIME_URL=$${PYTHON_RUNTIME_URL:-http://localhost:8083} \
	SKIP_MIGRATION_VALIDATION=true \
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
