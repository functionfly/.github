#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ITERATIONS=3
POLL_INTERVAL=100
MAX_WAIT=60000
SERVICES=()

source "$PROJECT_ROOT/.env" 2>/dev/null || true
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-functionfly}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
export DEVELOPMENT=${DEVELOPMENT:-true}
export SKIP_MIGRATION_VALIDATION=${SKIP_MIGRATION_VALIDATION:-true}
export VERIFICATION_ENABLED=${VERIFICATION_ENABLED:-false}

usage() {
  cat << EOF
Usage: $0 [OPTIONS]

Measure cold start/restart times for FunctionFly services and runtimes.

SERVICES:
  orchestrator     - Go API server (port 8080)
  ai-service       - Python/FastAPI AI service (port 18081)
  runtime-local    - Rust local runtime (dynamic port)
  runtime-prism    - Rust Prism runtime (dynamic port)
  runtime-kotlin   - Kotlin JVM runtime (dynamic port)

OPTIONS:
  -i, --iterations N     Number of iterations per service (default: 3)
  -p, --poll-ms N        Poll interval in milliseconds (default: 100)
  -m, --max-wait-ms N    Max wait time in milliseconds (default: 60000)
  -s, --services LIST    Comma-separated list (default: all available)
  -h, --help             Show this help

EXAMPLES:
  $0                                    # Auto-detect available services
  $0 -s orchestrator                   # Test orchestrator only
  $0 -s orchestrator,ai-service        # Test specific services
  $0 -i 5                              # More iterations

EOF
}

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')]${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING:${NC} $1"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR:${NC} $1"; }

parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      -i|--iterations) ITERATIONS="$2"; shift 2 ;;
      -p|--poll-ms) POLL_INTERVAL="$2"; shift 2 ;;
      -m|--max-wait-ms) MAX_WAIT="$2"; shift 2 ;;
      -s|--services) IFS=',' read -ra SERVICES <<< "$2"; shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) error "Unknown option: $1"; usage; exit 1 ;;
    esac
  done
}

check_service() {
  local name=$1
  local url=$2
  
  if curl -sf --max-time 2 "$url" > /dev/null 2>&1; then
    return 0
  fi
  return 1
}

wait_for_health() {
  local url=$1
  local timeout=${2:-$MAX_WAIT}
  local start_time=$(date +%s%3N)
  
  while true; do
    if curl -sf --max-time 2 "$url" > /dev/null 2>&1; then
      echo $(($(date +%s%3N) - start_time))
      return 0
    fi
    local elapsed=$(($(date +%s%3N) - start_time))
    if [[ $elapsed -ge $timeout ]]; then
      echo $timeout
      return 1
    fi
    sleep $(echo "scale=3; $POLL_INTERVAL/1000" | bc)
  done
}

stop_service() {
  local name=$1
  
  case $name in
    orchestrator) pkill -f "orchestrator-api" 2>/dev/null || true ;;
    ai-service) pkill -f "uvicorn" 2>/dev/null || true ;;
    runtime-local) pkill -f "functionfly-local" 2>/dev/null || true ;;
    runtime-prism) pkill -f "prism" 2>/dev/null || true ;;
    runtime-kotlin) pkill -f "kotlin" 2>/dev/null || true ;;
  esac
  
  sleep 2
}

start_orchestrator() {
  cd "$PROJECT_ROOT"
  if [[ ! -f "./bin/orchestrator-api" ]]; then
    warn "Orchestrator binary not found, building..."
    go build -o ./bin/orchestrator-api ./cmd/orchestrator-api
  fi
  ./bin/orchestrator-api --skip-migrations > /dev/null 2>&1 &
  echo $! > "$SCRIPT_DIR/.coldstart_orchestrator.pid"
}

start_ai_service() {
  cd "$PROJECT_ROOT/ai-service"
  if [[ ! -f "pyproject.toml" ]]; then
    warn "AI service not found at $PROJECT_ROOT/ai-service"
    return 1
  fi
  uv run uvicorn src.main:app --host 127.0.0.1 --port 18081 > /dev/null 2>&1 &
  echo $! > "$SCRIPT_DIR/.coldstart_ai-service.pid"
}

start_runtime_local() {
  cd "$PROJECT_ROOT/runtimes/local"
  if [[ ! -f "Cargo.toml" ]]; then
    warn "Runtime local not found"
    return 1
  fi
  if [[ ! -f "target/release/functionfly-local" ]]; then
    warn "Runtime not built, building (this may take a while)..."
    cargo build --release 2>/dev/null || true
  fi
  if [[ -f "target/release/functionfly-local" ]]; then
    ./target/release/functionfly-local > /tmp/runtime-local.log 2>&1 &
    echo $! > "$SCRIPT_DIR/.coldstart_runtime-local.pid"
  fi
}

start_runtime_prism() {
  cd "$PROJECT_ROOT/runtimes/prism"
  if [[ ! -f "Cargo.toml" ]]; then
    warn "Runtime prism not found"
    return 1
  fi
  if [[ ! -f "target/release/prism" ]]; then
    warn "Runtime not built, building (this may take a while)..."
    cargo build --release 2>/dev/null || true
  fi
  if [[ -f "target/release/prism" ]]; then
    ./target/release/prism > /tmp/runtime-prism.log 2>&1 &
    echo $! > "$SCRIPT_DIR/.coldstart_runtime-prism.pid"
  fi
}

start_runtime_kotlin() {
  cd "$PROJECT_ROOT/runtimes/kotlin"
  if [[ ! -f "build.gradle.kts" ]]; then
    warn "Runtime kotlin not found"
    return 1
  fi
  ./gradlew run > /tmp/runtime-kotlin.log 2>&1 &
  echo $! > "$SCRIPT_DIR/.coldstart_runtime-kotlin.pid"
}

get_runtime_port() {
  local name=$1
  local default=$2
  
  case $name in
    runtime-local)
      grep -oP 'listening on \K\d+' /tmp/runtime-local.log 2>/dev/null || echo "$default"
      ;;
    runtime-prism)
      grep -oP 'listening on \K\d+' /tmp/runtime-prism.log 2>/dev/null || echo "$default"
      ;;
    runtime-kotlin)
      grep -oP 'Server running on \K\d+' /tmp/runtime-kotlin.log 2>/dev/null || echo "$default"
      ;;
    *)
      echo "$default"
      ;;
  esac
}

measure_cold_start() {
  local service=$1
  local health_url=$2
  local times=()
  local failures=0
  
  log "Measuring cold start for $service (URL: $health_url)"
  
  for i in $(seq 1 $ITERATIONS); do
    echo -n "  Iteration $i/$ITERATIONS... "
    
    stop_service "$service"
    sleep 1
    
    local start_time=$(date +%s%3N)
    
    case $service in
      orchestrator) start_orchestrator ;;
      ai-service) start_ai_service ;;
      runtime-local) start_runtime_local ;;
      runtime-prism) start_runtime_prism ;;
      runtime-kotlin) start_runtime_kotlin ;;
    esac
    
    sleep 2
    
    local elapsed=$(wait_for_health "$health_url")
    
    if [[ $elapsed -ge $MAX_WAIT ]]; then
      failures=$((failures + 1))
      echo -e "${RED}TIMEOUT${NC}"
    else
      echo -e "${GREEN}${elapsed}ms${NC}"
      times+=($elapsed)
    fi
  done
  
  echo ""
  
  if [[ ${#times[@]} -eq 0 ]]; then
    warn "All iterations failed for $service"
    echo "$service,failed,failed,failed,0,$failures" >> "$SCRIPT_DIR/coldstart_results.csv"
    return
  fi
  
  local sum=0
  local min=${times[0]}
  local max=${times[0]}
  
  for t in "${times[@]}"; do
    sum=$(echo "$sum + $t" | bc)
    [[ $(echo "$t < $min" | bc -l) ]] && min=$t
    [[ $(echo "$t > $max" | bc -l) ]] && max=$t
  done
  
  local avg=$(echo "scale=0; $sum / ${#times[@]}" | bc)
  
  echo "Results for $service (${#times[@]} successful, $failures failed):"
  echo "  Average: ${avg}ms | Min: ${min}ms | Max: ${max}ms"
  echo ""
  
  echo "$service,$avg,$min,$max,${#times[@]},$failures" >> "$SCRIPT_DIR/coldstart_results.csv"
}

parse_args "$@"
cd "$SCRIPT_DIR"

echo ""
log "Cold Start Performance Test"
log "==========================="
echo ""
echo "Configuration:"
echo "  Iterations: $ITERATIONS"
echo "  Poll interval: ${POLL_INTERVAL}ms"
echo "  Max wait: ${MAX_WAIT}ms"
echo ""

if [[ ${#SERVICES[@]} -eq 0 ]]; then
  log "Auto-detecting available services..."
  
  check_service "orchestrator" "http://localhost:8080/health" && SERVICES+=("orchestrator")
  check_service "ai-service" "http://localhost:18081/health" && SERVICES+=("ai-service")
  
  if [[ -f "$PROJECT_ROOT/runtimes/local/Cargo.toml" ]]; then
    if [[ -f "$PROJECT_ROOT/runtimes/local/target/release/functionfly-local" ]]; then
      SERVICES+=("runtime-local")
    else
      warn "runtime-local binary not built, skipping. Run: cd runtimes/local && cargo build --release"
    fi
  fi
  
  if [[ -f "$PROJECT_ROOT/runtimes/prism/Cargo.toml" ]]; then
    if [[ -f "$PROJECT_ROOT/runtimes/prism/target/release/prism" ]]; then
      SERVICES+=("runtime-prism")
    else
      warn "runtime-prism binary not built, skipping. Run: cd runtimes/prism && cargo build --release"
    fi
  fi
fi

if [[ ${#SERVICES[@]} -eq 0 ]]; then
  error "No services available to test"
  exit 1
fi

echo "Services to test: ${SERVICES[*]}"
echo ""

echo "service,avg_ms,min_ms,max_ms,successful,failed" > coldstart_results.csv

for service in "${SERVICES[@]}"; do
  case $service in
    orchestrator) measure_cold_start "orchestrator" "http://localhost:8080/health" ;;
    ai-service) measure_cold_start "ai-service" "http://localhost:18081/health" ;;
    runtime-local)
      port=$(get_runtime_port "runtime-local" "8083")
      measure_cold_start "runtime-local" "http://localhost:$port/health"
      ;;
    runtime-prism)
      port=$(get_runtime_port "runtime-prism" "8084")
      measure_cold_start "runtime-prism" "http://localhost:$port/health"
      ;;
    runtime-kotlin)
      port=$(get_runtime_port "runtime-kotlin" "8085")
      measure_cold_start "runtime-kotlin" "http://localhost:$port/health"
      ;;
    *) warn "Unknown service: $service" ;;
  esac
done

log "Test complete. Results saved to coldstart_results.csv"
cat coldstart_results.csv