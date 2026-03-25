#!/bin/bash

# FunctionFly Production Deployment Script
# This script deploys the Flywheel social network to production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="${PROJECT_ROOT}/.env"
DOCKER_COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.production.yml"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    # Check if .env file exists
    if [ ! -f "$ENV_FILE" ]; then
        log_error ".env file not found. Please create it from .env.production"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

validate_environment() {
    log_info "Validating environment variables..."

    # Source the environment file
    source "$ENV_FILE"

    # Check required variables
    required_vars=(
        "DB_PASSWORD"
        "REDIS_PASSWORD"
        "JWT_SECRET"
        "API_SHARED_SECRET"
        "SSL_DOMAIN"
        "SSL_EMAIL"
    )

    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            log_error "Required environment variable $var is not set"
            exit 1
        fi
    done

    # Validate JWT_SECRET length
    if [ ${#JWT_SECRET} -lt 32 ]; then
        log_error "JWT_SECRET must be at least 32 characters long"
        exit 1
    fi

    # Validate API_SHARED_SECRET length
    if [ ${#API_SHARED_SECRET} -lt 32 ]; then
        log_error "API_SHARED_SECRET must be at least 32 characters long"
        exit 1
    fi

    log_info "Environment validation passed"
}

build_images() {
    log_info "Building Docker images..."

    cd "$PROJECT_ROOT"

    # Build orchestrator API
    log_info "Building orchestrator-api..."
    docker build -f deploy/production/Dockerfile.orchestrator -t functionfly/orchestrator-api:latest .

    # Build dashboard
    log_info "Building dashboard..."
    docker build -f web/dashboard/Dockerfile -t functionfly/dashboard:latest web/dashboard/

    log_info "Docker images built successfully"
}

deploy_services() {
    log_info "Deploying services..."

    cd "$PROJECT_ROOT"

    # Stop existing services
    log_info "Stopping existing services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down --remove-orphans

    # Pull latest images
    log_info "Pulling latest images..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" pull

    # Start services
    log_info "Starting services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

    log_info "Services deployed successfully"
}

wait_for_services() {
    log_info "Waiting for services to be healthy..."

    # Wait for PostgreSQL
    log_info "Waiting for PostgreSQL..."
    timeout 60 bash -c 'until docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T postgres pg_isready -U functionfly; do sleep 2; done'

    # Wait for Redis
    log_info "Waiting for Redis..."
    timeout 60 bash -c 'until docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T redis redis-cli -a "$REDIS_PASSWORD" ping | grep -q PONG; do sleep 2; done'

    # Wait for orchestrator API
    log_info "Waiting for orchestrator API..."
    timeout 120 bash -c 'until curl -f http://localhost:8080/health > /dev/null 2>&1; do sleep 5; done'

    # Wait for dashboard
    log_info "Waiting for dashboard..."
    timeout 60 bash -c 'until curl -f http://localhost:3000/health > /dev/null 2>&1; do sleep 5; done'

    log_info "All services are healthy"
}

run_migrations() {
    log_info "Running database migrations..."

    cd "$PROJECT_ROOT"

    # Run migrations
    docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T orchestrator-api ./migrate up

    log_info "Database migrations completed"
}

setup_monitoring() {
    log_info "Setting up monitoring..."

    cd "$PROJECT_ROOT"

    # Wait for Prometheus to be ready
    log_info "Waiting for Prometheus..."
    timeout 60 bash -c 'until curl -f http://localhost:9091/-/healthy > /dev/null 2>&1; do sleep 5; done'

    # Wait for Grafana to be ready
    log_info "Waiting for Grafana..."
    timeout 60 bash -c 'until curl -f http://localhost:3001/api/health > /dev/null 2>&1; do sleep 5; done'

    log_info "Monitoring setup completed"
}

print_deployment_info() {
    log_info "Deployment completed successfully!"
    echo ""
    echo "=========================================="
    echo "FunctionFly Production Deployment"
    echo "=========================================="
    echo ""
    echo "Services:"
    echo "  - Dashboard: https://${SSL_DOMAIN}"
    echo "  - API: https://${SSL_DOMAIN}/api"
    echo "  - Health: https://${SSL_DOMAIN}/health"
    echo ""
    echo "Monitoring:"
    echo "  - Grafana: https://monitoring.${SSL_DOMAIN}/grafana"
    echo "  - Prometheus: https://monitoring.${SSL_DOMAIN}/prometheus"
    echo "  - Loki: https://monitoring.${SSL_DOMAIN}/loki"
    echo ""
    echo "Useful commands:"
    echo "  - View logs: docker-compose -f ${DOCKER_COMPOSE_FILE} logs -f"
    echo "  - Stop services: docker-compose -f ${DOCKER_COMPOSE_FILE} down"
    echo "  - Restart services: docker-compose -f ${DOCKER_COMPOSE_FILE} restart"
    echo ""
    echo "=========================================="
}

# Main deployment flow
main() {
    log_info "Starting FunctionFly production deployment..."

    check_prerequisites
    validate_environment
    build_images
    deploy_services
    wait_for_services
    run_migrations
    setup_monitoring
    print_deployment_info

    log_info "Deployment completed successfully!"
}

# Run main function
main "$@"
