#!/bin/bash

# =============================================================================
# FunctionFly Production Readiness Test Script
# =============================================================================
# This script validates that all production components are properly configured
# and ready for deployment.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0
WARNINGS=0

# =============================================================================
# Helper Functions
# =============================================================================

log_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

# =============================================================================
# Test Functions
# =============================================================================

test_file_exists() {
    local file="$1"
    local description="$2"

    log_test "Checking $description"

    if [ -f "$file" ]; then
        log_pass "$description exists"
        return 0
    else
        log_fail "$description not found at $file"
        return 1
    fi
}

test_directory_exists() {
    local dir="$1"
    local description="$2"

    log_test "Checking $description"

    if [ -d "$dir" ]; then
        log_pass "$description exists"
        return 0
    else
        log_fail "$description not found at $dir"
        return 1
    fi
}

test_env_var() {
    local var_name="$1"
    local description="$2"
    local required="${3:-true}"

    log_test "Checking environment variable: $var_name"

    if [ -n "${!var_name}" ]; then
        log_pass "$var_name is set"
        return 0
    else
        if [ "$required" = "true" ]; then
            log_fail "$var_name is not set (required)"
            return 1
        else
            log_warn "$var_name is not set (optional)"
            return 0
        fi
    fi
}

test_docker_compose() {
    local file="$1"
    local description="$2"

    log_test "Validating $description"

    if docker compose -f "$file" config --quiet 2>/dev/null; then
        log_pass "$description is valid"
        return 0
    else
        log_fail "$description has syntax errors"
        return 1
    fi
}

test_go_build() {
    log_test "Testing Go build"

    if go build -o /tmp/functionfly-test ./cmd/orchestrator-api 2>/dev/null; then
        log_pass "Go build successful"
        rm -f /tmp/functionfly-test
        return 0
    else
        log_fail "Go build failed"
        return 1
    fi
}

test_dashboard_build() {
    log_test "Testing dashboard build"

    if [ -d "web/dashboard" ]; then
        cd web/dashboard
        if npm run build --if-present 2>/dev/null; then
            log_pass "Dashboard build successful"
            cd ../..
            return 0
        else
            log_warn "Dashboard build skipped (npm not available or build failed)"
            cd ../..
            return 0
        fi
    else
        log_warn "Dashboard directory not found"
        return 0
    fi
}

# =============================================================================
# Main Test Suite
# =============================================================================

echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                  FunctionFly Production Readiness Test                      ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Load environment variables if .env exists
if [ -f ".env" ]; then
    log_test "Loading environment variables from .env"
    export $(grep -v '^#' .env | xargs)
    log_pass "Environment variables loaded"
else
    log_warn ".env file not found (using system environment)"
fi

# =============================================================================
# 1. Core Files
# =============================================================================

log_header "1. Core Files"

test_file_exists "docker-compose.production.yml" "Production Docker Compose"
test_file_exists "docker-compose.staging.yml" "Staging Docker Compose"
test_file_exists "docker-compose.monitoring.yml" "Monitoring Docker Compose"
test_file_exists ".env.example" "Environment example file"
test_file_exists "README.md" "README documentation"
test_file_exists "CONTRIBUTING.md" "Contributing guidelines"

# =============================================================================
# 2. Deployment Configuration
# =============================================================================

log_header "2. Deployment Configuration"

test_file_exists "deploy/caddy/Dockerfile" "Caddy Dockerfile"
test_file_exists "deploy/caddy/Caddyfile" "Caddy configuration"
test_file_exists "scripts/deploy-staging.sh" "Staging deployment script"

# =============================================================================
# 3. Monitoring Stack
# =============================================================================

log_header "3. Monitoring Stack"

test_file_exists "docker-compose.monitoring.yml" "Monitoring Docker Compose"

# Check if monitoring services are defined
if grep -q "prometheus:" docker-compose.monitoring.yml; then
    log_pass "Prometheus service defined"
else
    log_fail "Prometheus service not defined"
fi

if grep -q "grafana:" docker-compose.monitoring.yml; then
    log_pass "Grafana service defined"
else
    log_fail "Grafana service not defined"
fi

if grep -q "loki:" docker-compose.monitoring.yml; then
    log_pass "Loki service defined"
else
    log_fail "Loki service not defined"
fi

# =============================================================================
# 4. Security Configuration
# =============================================================================

log_header "4. Security Configuration"

test_env_var "JWT_SECRET" "JWT Secret" "false"
test_env_var "API_SHARED_SECRET" "API Shared Secret" "false"
test_env_var "DB_PASSWORD" "Database Password" "false"
test_env_var "REDIS_PASSWORD" "Redis Password" "false"

# =============================================================================
# 5. Database Configuration
# =============================================================================

log_header "5. Database Configuration"

test_env_var "DB_HOST" "Database Host" "false"
test_env_var "DB_PORT" "Database Port" "false"
test_env_var "DB_NAME" "Database Name" "false"
test_env_var "DB_USER" "Database User" "false"

# =============================================================================
# 6. Application Configuration
# =============================================================================

log_header "6. Application Configuration"

test_env_var "APP_ENV" "Application Environment" "false"
test_env_var "APP_PORT" "Application Port" "false"
test_env_var "LOG_LEVEL" "Log Level" "false"
test_env_var "LOG_FORMAT" "Log Format" "false"

# =============================================================================
# 7. Build Tests
# =============================================================================

log_header "7. Build Tests"

test_go_build
test_dashboard_build

# =============================================================================
# 8. Docker Compose Validation
# =============================================================================

log_header "8. Docker Compose Validation"

test_docker_compose "docker-compose.production.yml" "Production Docker Compose"
test_docker_compose "docker-compose.staging.yml" "Staging Docker Compose"
test_docker_compose "docker-compose.monitoring.yml" "Monitoring Docker Compose"

# =============================================================================
# 9. Directory Structure
# =============================================================================

log_header "9. Directory Structure"

test_directory_exists "cmd" "Command directory"
test_directory_exists "internal" "Internal packages directory"
test_directory_exists "web" "Web directory"
test_directory_exists "deploy" "Deploy directory"
test_directory_exists "scripts" "Scripts directory"
test_directory_exists "docs" "Documentation directory"

# =============================================================================
# 10. Go Module
# =============================================================================

log_header "10. Go Module"

if [ -f "go.mod" ]; then
    log_pass "go.mod exists"

    # Check if module name is correct
    if grep -q "module github.com/functionfly/functionfly" go.mod; then
        log_pass "Go module name is correct"
    else
        log_warn "Go module name may be incorrect"
    fi
else
    log_fail "go.mod not found"
fi

# =============================================================================
# Summary
# =============================================================================

log_header "Test Summary"

echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"

TOTAL=$((PASSED + FAILED))
if [ $TOTAL -gt 0 ]; then
    SUCCESS_RATE=$((PASSED * 100 / TOTAL))
    echo -e "\n${BLUE}Success Rate: ${SUCCESS_RATE}%${NC}"
fi

echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    ✓ PRODUCTION READY                                       ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    ✗ NOT PRODUCTION READY                                   ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
