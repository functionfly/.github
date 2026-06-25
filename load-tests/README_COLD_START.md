# Cold Start Performance Testing

This directory contains tools for measuring cold start times across all FunctionFly services and runtimes.

## Quick Start

### 1. Start Dependencies First

```bash
# PostgreSQL
sudo pg_ctlcluster 17 main start

# Redis
redis-server --daemonize yes
```

### 2. Run Cold Start Test

```bash
cd load-tests

# All services (orchestrator, ai-service, and all runtimes)
./cold-start.sh

# Test specific runtimes only
./cold-start.sh -s runtime-local,runtime-prism

# Quick test (1 iteration)
./cold-start.sh -i 1
```

## Services & Runtimes Tested

| Service | Type | Port | Language | Notes |
|---------|------|------|----------|-------|
| orchestrator | API Server | 8080 | Go | Main backend API |
| ai-service | AI Service | 18081 | Python | FlyMind AI integration |
| runtime-local | Runtime | 8083 | Rust | Primary local execution |
| runtime-prism | Runtime | 8084 | Rust | Prism WASM runtime |
| runtime-kotlin | Runtime | 8085 | Kotlin | JVM-based execution |

## Test Tools

### 1. Shell Script (`cold-start.sh`)
Measures true cold starts by killing and restarting services.

**Options:**
- `-i, --iterations N` - Iterations per service (default: 5)
- `-p, --poll-ms N` - Poll interval ms (default: 100)
- `-m, --max-wait-ms N` - Max wait ms (default: 60000)
- `-s, --services LIST` - Comma-separated list

**Service Names:**
- `orchestrator`, `ai-service`
- `runtime-local`, `runtime-prism`, `runtime-kotlin`

**Examples:**
```bash
./cold-start.sh -i 3 -s orchestrator
./cold-start.sh -s runtime-local,runtime-prism
./cold-start.sh -s ai-service -m 120000
```

### 2. k6 Test (`cold-start-test.js`)
Load test with cold start scenarios using k6.

```bash
# Install k6
brew install k6  # macOS
# or: sudo apt-get install k6  # Ubuntu

# Run all services
k6 run cold-start-test.js

# Run specific service
SERVICE=runtime-local k6 run cold-start-test.js

# Custom URLs
BASE_URL=http://localhost:8080 \
RUNTIME_LOCAL_URL=http://localhost:8083 \
k6 run cold-start-test.js
```

### 3. Go Benchmarks (`internal/benchmark/startup_test.go`)
Unit-level benchmarks for Go component initialization.

```bash
# Run all benchmarks
go test -bench=. -benchmem ./internal/benchmark/

# Database connection benchmark only
go test -bench=ColdStart -benchmem ./internal/benchmark/

# CPU/memory profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/benchmark/
```

## What Gets Measured

| Component | Metric | Typical Range |
|-----------|--------|---------------|
| orchestrator | Time to `/health` 200 | 500-2000ms |
| ai-service | Time to `/health` 200 | 2000-5000ms |
| runtime-local | Time to `/health` 200 | 500-3000ms |
| runtime-prism | Time to `/health` 200 | 500-3000ms |
| runtime-kotlin | Time to `/health` 200 | 3000-10000ms |
| Database connection | `NewPostgresDB()` | 100-500ms |
| Database ping | `PingContext()` | 1-10ms |

## Cold Start Categories

| Category | Time | Action |
|----------|------|--------|
| Excellent | < 500ms | Optimized |
| Good | 500ms - 2s | Acceptable |
| Acceptable | 2s - 5s | May need review |
| Slow | 5s - 10s | Investigate |
| Critical | > 10s | Optimize immediately |

## Typical Cold Start Breakdown

```
API Server = Database Connection + Migrations + Service Init + HTTP Listen
           = ~100-500ms + ~0-2000ms + ~50-100ms + ~10-50ms
           = ~160-2650ms typical

Runtime = Language Init + Sandbox Creation + Handler Setup + Port Listen
        = ~200-2000ms + ~100-500ms + ~50-100ms + ~10-50ms
        = ~360-2650ms typical (varies by language)
```

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Cold Start Test
  run: |
    cd load-tests
    chmod +x cold-start.sh
    ./cold-start.sh -i 3

- name: Go Benchmarks
  run: |
    go test -bench=. -benchmem -count=3 ./internal/benchmark/
```

## Results

Results are saved to `coldstart_results.csv`:

```csv
service,avg_ms,min_ms,max_ms,iterations
orchestrator,1234,1100,1400,5
ai-service,2500,2200,2800,5
runtime-local,890,750,1050,5
runtime-prism,920,800,1100,5
runtime-kotlin,4500,4000,5200,5
```