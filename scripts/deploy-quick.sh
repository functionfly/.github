#!/bin/bash
#
# Quick Deploy FunctionFly Registry to Fly.io
#
# COST-OPTIMIZED: Uses minimal resources for budget-conscious deployment
# Usage: ./scripts/deploy-quick.sh [staging|production]
#

set -e

APP_NAME="functionfly-orchestrator"
STAGING_APP="functionfly-orchestrator-staging"
ENVIRONMENT="${1:-staging}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
    echo "Usage: $0 [ENVIRONMENT]"
    echo ""
    echo "Environments:"
    echo "  staging     Deploy to staging (default)"
    echo "  production  Deploy to production"
    echo ""
    echo "Examples:"
    echo "  $0                    # Deploy to staging"
    echo "  $0 staging           # Deploy to staging"
    echo "  $0 production        # Deploy to production"
    echo ""
    echo "Cost-optimized settings:"
    echo "  - shared-cpu-1x (1 CPU, 1GB RAM) = ~\$15/month"
    echo "  - Auto-stop enabled (saves costs during idle)"
    echo "  - No persistent volume (uses ephemeral storage)"
    echo "  - 14-day backup retention"
    echo "  - Disabled: read replicas, detailed tracing"
}

check_fly() {
    if ! command -v fly &> /dev/null; then
        log_error "fly CLI not found. Install with:"
        echo "  curl -L https://fly.io/install.sh | sh"
        exit 1
    fi

    if ! fly auth whoami &> /dev/null; then
        log_error "Not authenticated. Run: fly auth login"
        exit 1
    fi
}

quick_deploy_staging() {
    log_info "🚀 Quick deploy to STAGING"
    log_info "App: $STAGING_APP"

    # Create app if doesn't exist
    if ! fly apps list 2>/dev/null | grep -q "^${STAGING_APP}$"; then
        log_info "Creating staging app..."
        fly apps create "$STAGING_APP"
    fi

    # Deploy with fly CLI directly
    log_info "Deploying..."
    fly deploy \
        --config fly.staging.toml \
        --app "$STAGING_APP" \
        --yes \
        --wait-timeout=300s

    # Quick health check
    log_info "Health check..."
    sleep 10
    if curl -sf "https://${STAGING_APP}.fly.dev/health" &>/dev/null; then
        log_success "✅ Staging deployed and healthy!"
        echo ""
        echo "URL: https://${STAGING_APP}.fly.dev"
        echo "Logs: fly logs --app ${STAGING_APP}"
    else
        log_warn "⚠️  App deployed but health check pending (may need more time)"
    fi
}

quick_deploy_production() {
    log_info "🚀 Quick deploy to PRODUCTION"
    log_info "App: $APP_NAME"
    log_warn "⚠️  This will deploy to production!"

    # Safety check
    read -p "Continue? [y/N] " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log_info "Cancelled"
        exit 0
    fi

    # Create app if doesn't exist
    if ! fly apps list 2>/dev/null | grep -q "^${APP_NAME}$"; then
        log_info "Creating production app..."
        fly apps create "$APP_NAME"
        log_warn "⚠️  New app created. Set secrets before deploying:"
        echo "  fly secrets set --app $APP_NAME DATABASE_URL='...'"
        echo "  fly secrets set --app $APP_NAME REDIS_URL='...'"
        exit 0
    fi

    # Deploy with fly CLI directly
    log_info "Deploying..."
    fly deploy \
        --config fly.toml \
        --app "$APP_NAME" \
        --yes \
        --wait-timeout=300s

    # Quick health check
    log_info "Health check..."
    sleep 15
    if curl -sf "https://${APP_NAME}.fly.dev/health" &>/dev/null; then
        log_success "✅ Production deployed and healthy!"
        echo ""
        echo "URL: https://${APP_NAME}.fly.dev"
        echo "Status: fly status --app ${APP_NAME}"
        echo "Logs: fly logs --app ${APP_NAME}"
    else
        log_warn "⚠️  App deployed but health check pending"
        echo "Check status: fly status --app ${APP_NAME}"
    fi
}

show_costs() {
    echo ""
    echo "💰 COST-OPTIMIZED DEPLOYMENT"
    echo "============================="
    echo ""
    echo "Configuration:"
    echo "  • VM: shared-cpu-1x (1 CPU, 1GB RAM)"
    echo "  • Auto-stop: Enabled (saves ~50% during idle)"
    echo "  • Storage: Ephemeral (no volume cost)"
    echo "  • Redis: Use Fly.io Redis or skip (saves ~\$80)"
    echo "  • Backups: 14-day retention"
    echo ""
    echo "Estimated monthly costs:"
    echo "  ├─ Fly.io VM (auto-stop):        ~\$5-15"
    echo "  ├─ Neon Postgres (free tier):    ~\$0"
    echo "  ├─ R2 Storage (minimal):         ~\$0-5"
    echo "  └─ Bandwidth:                    ~\$5-10"
    echo ""
    echo "  TOTAL: ~\$10-30/month (vs ~\$188 originally)"
    echo ""
    echo "To scale up later:"
    echo "  fly scale vm shared-cpu-2x --app ${APP_NAME}"
    echo "  fly scale count 2 --app ${APP_NAME}"
    echo ""
}

# Main
case "${1:-}" in
    help|--help|-h)
        show_help
        exit 0
        ;;
    staging|s)
        check_fly
        quick_deploy_staging
        ;;
    production|prod|p)
        check_fly
        quick_deploy_production
        ;;
    costs|--costs)
        show_costs
        exit 0
        ;;
    "")
        # Default to staging
        check_fly
        quick_deploy_staging
        ;;
    *)
        log_error "Unknown environment: $1"
        show_help
        exit 1
        ;;
esac
