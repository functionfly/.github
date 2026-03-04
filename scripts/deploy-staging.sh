#!/bin/bash

# ============================================================================
# FunctionFly Staging Deployment Script
# ============================================================================
# This script handles the complete deployment of the FunctionFly staging
# environment including validation, migrations, service startup, and health checks.
#
# Usage: ./scripts/deploy-staging.sh [--skip-build] [--skip-migrations] [--logs]
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script flags
SKIP_BUILD=false
SKIP_MIGRATIONS=false
FOLLOW_LOGS=false
HEALTH_CHECK_RETRIES=30
HEALTH_CHECK_INTERVAL=5

# ============================================================================
# Helper Functions
# ============================================================================

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_usage() {
    cat << EOF
Usage: ./scripts/deploy-staging.sh [OPTIONS]

OPTIONS:
    --skip-build        Skip Docker image builds
    --skip-migrations   Skip database migrations
    --logs, -l          Follow logs after deployment
    --help, -h          Show this help message

EXAMPLES:
    ./scripts/deploy-staging.sh                    # Full deployment
    ./scripts/deploy-staging.sh --skip-build       # Skip building images
    ./scripts/deploy-staging.sh --logs             # Follow logs after deploy

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-migrations)
                SKIP_MIGRATIONS=true
                shift
                ;;
            --logs|-l)
                FOLLOW_LOGS=true
                shift
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# ============================================================================
# Validation Functions
# ============================================================================

validate_environment() {
    log_info "Validating environment configuration..."

    # Check if .env.staging exists
    if [ ! -f ".env.staging" ]; then
        log_error ".env.staging file not found!"
        log_info "Please create .env.staging from .env.staging.example:"
        log_info "  cp .env.staging.example .env.staging"
        log_info "Then edit .env.staging with your actual values."
        exit 1
    fi

    # Check if docker-compose.staging.yml exists
    if [ ! -f "docker-compose.staging.yml" ]; then
        log_error "docker-compose.staging.yml not found!"
        exit 1
    fi

    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed!"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running!"
        exit 1
    fi

    # Check if docker-compose is available
    if ! command -v docker-compose &> /dev/null; then
        if ! docker compose version &> /dev/null; then
            log_error "Docker Compose is not installed!"
            exit 1
        fi
    fi

    # Load and validate environment variables
    set -a
    source .env.staging
    set +a

    # Check required variables
    local required_vars=(
        "DB_HOST"
        "DB_PASSWORD"
        "JWT_SECRET"
        "API_SHARED_SECRET"
    )

    local missing_vars=()
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ] || [ "${!var}" == "your-*-here" ] || [[ "${!var}" == *"change-this"* ]]; then
            missing_vars+=("$var")
        fi
    done

    if [ ${#missing_vars[@]} -ne 0 ]; then
        log_error "Missing or placeholder values for required environment variables:"
        printf '  - %s\n' "${missing_vars[@]}"
        log_info "Please update these values in .env.staging before deploying."
        exit 1
    fi

    log_success "Environment validation passed"
}

# ============================================================================
# Deployment Functions
# ============================================================================

stop_existing_services() {
    log_info "Stopping existing staging services..."
    docker-compose -f docker-compose.staging.yml down --remove-orphans 2>/dev/null || true
    log_success "Existing services stopped"
}

run_migrations() {
    if [ "$SKIP_MIGRATIONS" = true ]; then
        log_warning "Skipping database migrations"
        return 0
    fi

    log_info "Running database migrations..."

    # Build the migration runner if needed
    if ! command -v migrate &> /dev/null; then
        log_info "Installing golang-migrate..."
        if command -v go &> /dev/null; then
            go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        else
            log_warning "Go not found. Attempting to use Docker for migrations..."
            # Alternative: run migrations via Docker container
            docker run --rm \
                --network functionfly-staging \
                -v "$(pwd)/migrations:/migrations" \
                migrate/migrate \
                -path=/migrations \
                -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
                up
            return 0
        fi
    fi

    # Run migrations using golang-migrate
    if command -v migrate &> /dev/null; then
        migrate -path migrations \
            -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
            up
    fi

    log_success "Database migrations completed"
}

build_and_start_services() {
    log_info "Building and starting staging services..."

    if [ "$SKIP_BUILD" = true ]; then
        log_warning "Skipping Docker build (using existing images)"
        docker-compose -f docker-compose.staging.yml up -d
    else
        docker-compose -f docker-compose.staging.yml up --build -d
    fi

    log_success "Services started successfully"
}

# ============================================================================
# Health Check Functions
# ============================================================================

wait_for_service() {
    local service_name=$1
    local url=$2
    local max_retries=${3:-$HEALTH_CHECK_RETRIES}
    local retry=0

    log_info "Waiting for $service_name to be healthy..."

    while [ $retry -lt $max_retries ]; do
        if curl -sf "$url" &> /dev/null; then
            log_success "$service_name is healthy"
            return 0
        fi

        retry=$((retry + 1))
        if [ $retry -lt $max_retries ]; then
            echo -n "."
            sleep $HEALTH_CHECK_INTERVAL
        fi
    done

    echo
    log_error "$service_name failed to become healthy after $((max_retries * HEALTH_CHECK_INTERVAL)) seconds"
    return 1
}

perform_health_checks() {
    log_info "Performing health checks..."

    # Wait for services to initialize
    sleep 5

    # Check Orchestrator API
    if ! wait_for_service "Orchestrator API" "http://localhost:8082/health" 30; then
        log_error "Orchestrator API health check failed"
        show_service_logs
        return 1
    fi

    # Check Caddy Proxy
    if ! wait_for_service "Caddy Proxy" "http://localhost:8083/health" 20; then
        log_error "Caddy Proxy health check failed"
        show_service_logs
        return 1
    fi

    # Check Redis
    log_info "Checking Redis connection..."
    if docker-compose -f docker-compose.staging.yml exec -T redis redis-cli ping | grep -q "PONG"; then
        log_success "Redis is responding"
    else
        log_warning "Redis check inconclusive (may still be starting)"
    fi

    log_success "All health checks passed"
}

show_service_logs() {
    echo
    log_info "Recent service logs:"
    docker-compose -f docker-compose.staging.yml logs --tail=50
}

# ============================================================================
# Output Functions
# ============================================================================

print_deployment_summary() {
    echo
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║           FunctionFly Staging Deployment Complete              ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo
    log_success "All services are running successfully!"
    echo
    echo "📊 Service Endpoints:"
    echo "   • Orchestrator API:   http://localhost:8082"
    echo "   • Caddy Proxy:        http://localhost:8083"
    echo "   • Redis:              localhost:6380"
    echo
    echo "🌐 Staging Subdomains:"
    echo "   • Main Staging:       https://staging.functionfly.com"
    echo "   • API Staging:        https://api.staging.functionfly.com"
    echo "   • Edge Staging:       https://edge.staging.functionfly.com"
    echo "   • CDN Staging:        https://cdn.staging.functionfly.com"
    echo
    echo "🔍 Health Check Commands:"
    echo "   curl http://localhost:8082/health"
    echo "   curl http://localhost:8083/health"
    echo
    echo "📋 Useful Commands:"
    echo "   View logs:     docker-compose -f docker-compose.staging.yml logs -f"
    echo "   Stop services: docker-compose -f docker-compose.staging.yml down"
    echo "   Restart:       docker-compose -f docker-compose.staging.yml restart"
    echo "   Shell access:  docker-compose -f docker-compose.staging.yml exec orchestrator-api sh"
    echo

    if [ "$FOLLOW_LOGS" = true ]; then
        log_info "Following logs... (Press Ctrl+C to exit)"
        docker-compose -f docker-compose.staging.yml logs -f
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    echo "🚀 FunctionFly Staging Deployment"
    echo "=================================="
    echo

    parse_args "$@"

    # Change to project root
    cd "$(dirname "$0")/.."

    # Run deployment steps
    validate_environment
    stop_existing_services
    run_migrations
    build_and_start_services
    perform_health_checks
    print_deployment_summary
}

# Handle script interruption
trap 'log_error "Deployment interrupted"; exit 130' INT TERM

# Run main function
main "$@"
