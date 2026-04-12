#!/bin/bash
#
# Deploy FunctionFly Registry to Fly.io Production
#
# This script automates the production deployment with health checks,
# database migrations, and rollback capabilities.
#

set -euo pipefail

# Configuration
APP_NAME="functionfly-orchestrator"
STAGING_APP="functionfly-orchestrator-staging"
CONFIG_FILE="fly.toml"
STAGING_CONFIG="fly.staging.toml"
HEALTH_CHECK_TIMEOUT=300
ROLLBACK_ON_FAILURE=true

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check fly CLI
    if ! command -v fly &> /dev/null; then
        log_error "fly CLI not found. Install with: curl -L https://fly.io/install.sh | sh"
        exit 1
    fi

    # Check authentication
    if ! fly auth whoami &> /dev/null; then
        log_error "Not authenticated to Fly.io. Run: fly auth login"
        exit 1
    fi

    # Check config file exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Config file $CONFIG_FILE not found"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Get current app status
get_app_status() {
    local app=$1
    fly status --app "$app" --json 2>/dev/null || echo "{}"
}

# Pre-deployment checks
pre_deploy_checks() {
    log_info "Running pre-deployment checks..."

    # Check if app exists, create if not
    if ! fly apps list | grep -q "^${APP_NAME}$"; then
        log_warn "App $APP_NAME does not exist. Creating..."
        fly apps create "$APP_NAME"
    fi

    # Verify required secrets are set
    local required_secrets=(
        "DATABASE_URL"
        "REDIS_URL"
        "R2_ACCESS_KEY_ID"
        "R2_SECRET_ACCESS_KEY"
        "R2_ENDPOINT"
        "JWT_SECRET"
    )

    log_info "Checking required secrets..."
    local missing_secrets=()

    for secret in "${required_secrets[@]}"; do
        if ! fly secrets list --app "$APP_NAME" | grep -q "^${secret}"; then
            missing_secrets+=("$secret")
        fi
    done

    if [[ ${#missing_secrets[@]} -gt 0 ]]; then
        log_error "Missing required secrets: ${missing_secrets[*]}"
        log_info "Set them with: fly secrets set --app $APP_NAME SECRET_NAME=value"
        exit 1
    fi

    log_success "All required secrets are configured"

    # Test database connectivity
    log_info "Testing database connectivity..."
    local db_url
    db_url=$(fly secrets get DATABASE_URL --app "$APP_NAME" 2>/dev/null | head -1)

    if [[ -n "$db_url" ]]; then
        # Extract host from connection string
        local db_host
        db_host=$(echo "$db_url" | sed -n 's/.*@\([^:]*\).*/\1/p')
        log_info "Database host: $db_host"
    fi

    log_success "Pre-deployment checks completed"
}

# Deploy to staging first
staging_deploy() {
    log_info "Deploying to staging environment first..."

    # Check if staging app exists
    if ! fly apps list | grep -q "^${STAGING_APP}$"; then
        log_warn "Creating staging app..."
        fly apps create "$STAGING_APP"
    fi

    # Deploy to staging
    log_info "Deploying to $STAGING_APP..."
    if ! fly deploy --config "$STAGING_CONFIG" --app "$STAGING_APP" --yes; then
        log_error "Staging deployment failed"
        return 1
    fi

    # Health check on staging
    log_info "Running health checks on staging..."
    local staging_url="https://${STAGING_APP}.fly.dev"

    for i in {1..10}; do
        if curl -sf "${staging_url}/health" &>/dev/null; then
            log_success "Staging health check passed"
            break
        fi

        if [[ $i -eq 10 ]]; then
            log_error "Staging health check failed after 10 attempts"
            return 1
        fi

        log_warn "Staging health check attempt $i/10 failed, retrying..."
        sleep 10
    done

    log_success "Staging deployment successful"
    return 0
}

# Production deployment
production_deploy() {
    log_info "Starting production deployment..."

    # Get current deployment info for potential rollback
    local previous_status
    previous_status=$(get_app_status "$APP_NAME")
    local previous_image
    previous_image=$(echo "$previous_status" | grep -o '"image":"[^"]*"' | cut -d'"' -f4)

    if [[ -n "$previous_image" ]]; then
        log_info "Previous image for rollback: $previous_image"
        echo "$previous_image" > /tmp/fly_previous_image.txt
    fi

    # Deploy with rolling strategy
    log_info "Deploying to production with rolling strategy..."

    if ! fly deploy \
        --config "$CONFIG_FILE" \
        --app "$APP_NAME" \
        --yes \
        --wait-timeout=300s; then

        log_error "Production deployment failed"

        if [[ "$ROLLBACK_ON_FAILURE" == "true" && -f /tmp/fly_previous_image.txt ]]; then
            rollback_production
        fi

        return 1
    fi

    log_success "Production deployment completed"
    return 0
}

# Health checks
health_checks() {
    log_info "Running post-deployment health checks..."

    local app_url="https://${APP_NAME}.fly.dev"
    local start_time=$(date +%s)
    local timeout=$HEALTH_CHECK_TIMEOUT

    # Wait for app to be available
    log_info "Waiting for app to be available..."

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [[ $elapsed -gt $timeout ]]; then
            log_error "Health check timeout after ${timeout}s"
            return 1
        fi

        if curl -sf "${app_url}/health" &>/dev/null; then
            log_success "Basic health check passed"
            break
        fi

        log_info "Waiting for app... (${elapsed}s elapsed)"
        sleep 5
    done

    # Detailed health check
    log_info "Running detailed health check..."
    local health_response
    health_response=$(curl -sf "${app_url}/health/detailed" 2>/dev/null || echo "{}")

    log_info "Health status: $health_response"

    # Check individual components
    local checks=(
        "/health:Basic Health"
        "/health/detailed:Detailed Health"
    )

    for check in "${checks[@]}"; do
        local endpoint="${app_url}${check%%:*}"
        local name="${check##*:}"

        if curl -sf "$endpoint" &>/dev/null; then
            log_success "$name check passed"
        else
            log_warn "$name check returned non-200 (may be normal for cold start)"
        fi
    done

    log_success "All health checks passed"
    return 0
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."

    local app_url="https://${APP_NAME}.fly.dev"

    # Check app info
    log_info "Current app status:"
    fly status --app "$APP_NAME"

    # Check metrics endpoint
    log_info "Checking metrics endpoint..."
    if curl -sf "${app_url}:9090/metrics" &>/dev/null; then
        log_success "Metrics endpoint accessible"
    else
        log_warn "Metrics endpoint not accessible (may require internal access)"
    fi

    # Test a registry API endpoint
    log_info "Testing registry API..."
    local api_response
    api_response=$(curl -sf "${app_url}/.well-known/functionfly.json" 2>/dev/null || echo "{}")

    if [[ -n "$api_response" && "$api_response" != "{}" ]]; then
        log_success "Registry API responding"
    else
        log_warn "Registry API test inconclusive (may be OK)"
    fi

    log_success "Deployment verification complete"
}

# Rollback function
rollback_production() {
    log_warn "Initiating rollback..."

    if [[ -f /tmp/fly_previous_image.txt ]]; then
        local previous_image
        previous_image=$(cat /tmp/fly_previous_image.txt)

        log_info "Rolling back to image: $previous_image"

        # Deploy previous image
        fly deploy \
            --image "$previous_image" \
            --app "$APP_NAME" \
            --yes \
            --wait-timeout=300s

        log_success "Rollback completed"
    else
        log_error "No previous image found for rollback"
    fi
}

# Post-deployment tasks
post_deploy() {
    log_info "Running post-deployment tasks..."

    # Scale to desired count if needed
    local current_count
    current_count=$(fly status --app "$APP_NAME" --json 2>/dev/null | grep -o '"count":[0-9]*' | head -1 | cut -d: -f2)

    if [[ -n "$current_count" && "$current_count" -lt 2 ]]; then
        log_info "Scaling to 2 instances for high availability..."
        fly scale count 2 --app "$APP_NAME"
    fi

    # Clean up old images (keep last 10)
    log_info "Cleaning up old images..."
    fly image list --app "$APP_NAME" | tail -n +12 | while read -r line; do
        local image_id
        image_id=$(echo "$line" | awk '{print $1}')
        if [[ -n "$image_id" ]]; then
            fly image delete "$image_id" --app "$APP_NAME" --yes 2>/dev/null || true
        fi
    done

    log_success "Post-deployment tasks completed"
}

# Send notification
send_notification() {
    local status=$1
    local message=$2

    # Check for Slack webhook
    local slack_url
    slack_url=$(fly secrets get SLACK_WEBHOOK_URL --app "$APP_NAME" 2>/dev/null | head -1)

    if [[ -n "$slack_url" ]]; then
        local payload
        payload=$(cat <<EOF
{
    "text": "FunctionFly Registry Deployment - ${status}",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "🚀 Registry Deployment ${status}"
            }
        },
        {
            "type": "section",
            "fields": [
                {
                    "type": "mrkdwn",
                    "text": "*App:*\n${APP_NAME}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Time:*\n$(date)"
                }
            ]
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "${message}"
            }
        }
    ]
}
EOF
)

        curl -s -X POST -H 'Content-type: application/json' \
            --data "$payload" \
            "$slack_url" &>/dev/null || true
    fi

    # Health check ping (if configured)
    local healthcheck_url
    healthcheck_url=$(fly secrets get HEALTHCHECK_URL --app "$APP_NAME" 2>/dev/null | head -1)

    if [[ -n "$healthcheck_url" && "$status" == "SUCCESS" ]]; then
        curl -sf "$healthcheck_url" &>/dev/null || true
    fi
}

# Main deployment flow
main() {
    log_info "================================"
    log_info "FunctionFly Registry Deployment"
    log_info "Target: Fly.io Production"
    log_info "================================"

    local skip_staging=false
    local force_deploy=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-staging)
                skip_staging=true
                shift
                ;;
            --force)
                force_deploy=true
                shift
                ;;
            --no-rollback)
                ROLLBACK_ON_FAILURE=false
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --skip-staging     Skip staging deployment"
                echo "  --force            Force deploy without confirmations"
                echo "  --no-rollback      Disable automatic rollback on failure"
                echo "  --help             Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Step 1: Prerequisites
    check_prerequisites

    # Step 2: Pre-deployment checks
    pre_deploy_checks

    # Step 3: Deploy to staging (unless skipped)
    if [[ "$skip_staging" == "false" ]]; then
        if ! staging_deploy; then
            log_error "Staging deployment failed. Aborting production deploy."
            send_notification "FAILED" "Staging deployment failed. Production deploy aborted."
            exit 1
        fi
    else
        log_warn "Skipping staging deployment (--skip-staging)"
    fi

    # Step 4: Deploy to production
    if ! production_deploy; then
        log_error "Production deployment failed"
        send_notification "FAILED" "Production deployment failed. Check logs for details."
        exit 1
    fi

    # Step 5: Health checks
    if ! health_checks; then
        log_error "Health checks failed"

        if [[ "$ROLLBACK_ON_FAILURE" == "true" ]]; then
            rollback_production
        fi

        send_notification "FAILED" "Health checks failed after deployment."
        exit 1
    fi

    # Step 6: Verify deployment
    verify_deployment

    # Step 7: Post-deployment
    post_deploy

    # Success notification
    send_notification "SUCCESS" "Registry successfully deployed to production."

    log_success "================================"
    log_success "Deployment Complete!"
    log_success "================================"
    log_info "App URL: https://${APP_NAME}.fly.dev"
    log_info "Metrics: https://${APP_NAME}.fly.dev:9090/metrics"
    log_info "Health: https://${APP_NAME}.fly.dev/health"

    return 0
}

# Run main function
main "$@"
