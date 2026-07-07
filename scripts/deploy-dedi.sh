#!/bin/bash
#
# FunctionFly Backend API Deployment to Dedicated Server
# Usage: ./scripts/deploy-dedi.sh [server] [options]
#
# Options:
#   --build        Build the binary before deploying (default: use existing bin/orchestrator-api)
#   --fresh        Create initial deployment (systemd service, env file)
#   --update       Only update the binary and restart
#   --full         Full deployment: build, upload, configure, restart
#
# Examples:
#   ./scripts/deploy-dedi.sh function@194.107.163.34 --full
#   ./scripts/deploy-dedi.sh function@194.107.163.34 --update
#

set -e

# Configuration
SERVER="${1:-}"
DEPLOY_MODE="${2:-full}"
BINARY_NAME="orchestrator-api"
BINARY_LOCAL="./bin/${BINARY_NAME}"
REMOTE_DIR="/opt/functionfly"
REMOTE_BINARY="${REMOTE_DIR}/${BINARY_NAME}"
REMOTE_ENV="${REMOTE_DIR}/.env"
REMOTE_SERVICE="functionfly-api"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
    echo "Usage: $0 <server> [mode]"
    echo ""
    echo "Server: user@host format, e.g., function@194.107.163.34"
    echo ""
    echo "Modes:"
    echo "  --fresh   Initial deployment (creates systemd service, env file, enables service)"
    echo "  --update  Update binary and restart service only"
    echo "  --full    Full deployment: build, upload, configure, restart (default)"
    echo ""
    echo "Examples:"
    echo "  $0 function@194.107.163.34 --full   # Full deployment"
    echo "  $0 function@194.107.163.34 --update # Quick update"
    echo "  $0 function@194.107.163.34 --fresh  # First time setup"
}

# Validate server argument
if [[ -z "$SERVER" ]]; then
    log_error "Server argument required"
    show_help
    exit 1
fi

if [[ ! "$SERVER" == *@* ]]; then
    log_error "Server must be in user@host format"
    exit 1
fi

# Parse deploy mode
case "$DEPLOY_MODE" in
    --fresh|--update|--full)
        ;;
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        log_warn "Unknown mode '$DEPLOY_MODE', using --full"
        DEPLOY_MODE="--full"
        ;;
esac

# Build binary if requested
build_binary() {
    log_info "Building ${BINARY_NAME}..."
    cd /home/micro/projects/functionfly

    # Ensure binary exists
    if [[ ! -f "$BINARY_LOCAL" ]] || [[ "$DEPLOY_MODE" == "--fresh" ]]; then
        log_info "Building fresh binary..."
        make build-fast
    fi

    if [[ ! -f "$BINARY_LOCAL" ]]; then
        log_error "Binary not found at $BINARY_LOCAL"
        exit 1
    fi

    log_success "Binary ready: $(ls -lh "$BINARY_LOCAL" | awk '{print $5}')"
}

# Create environment file template
create_env_template() {
    cat << 'ENVEOF'
# FunctionFly API Environment Configuration
# Copy this file to .env and update values for your environment

# Server
PORT=8080
DEVELOPMENT=false

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=YOUR_DB_PASSWORD
DB_NAME=functionfly
DB_SSLMODE=disable

# Tenant Databases (if using multi-tenant)
TENANT_DB_ENABLED=false
TENANT_DB_HOST=localhost
TENANT_DB_PORT=5432
TENANT_DB_USER=postgres
TENANT_DB_PASSWORD=YOUR_DB_PASSWORD

# Dedicated Database (for /health/dedicated endpoint - status page)
DEDI_DB_HOST=194.107.163.34
DEDI_DB_PORT=5432
DEDI_DB_USER=postgres
DEDI_DB_PASSWORD=YOUR_DB_PASSWORD
DEDI_DB_NAME=functionfly
DEDI_DB_SSLMODE=require

# Cloudflare (CDN health check for status page)
CF_API_TOKEN=YOUR_CF_API_TOKEN
CF_ZONE_ID=f751c7d4ae9cc200cf3085ddfdf93e77

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Security
JWT_SECRET=CHANGE_THIS_TO_A_SECURE_SECRET
JWT_EXPIRATION=24h
PRIVACY_SALT=CHANGE_THIS_TO_A_SECURE_SALT

# Feature Flags
SKIP_MIGRATION_VALIDATION=false
VERIFICATION_ENABLED=true
USE_SERVICE_REGISTRY=false

# Optional: Set to true for production
# DATA_RETENTION_ENABLED=true
# DATA_RETENTION_CRON="0 3 * * *"
ENVEOF
}

# Create systemd service file
create_systemd_service() {
    cat << 'SYSTEMDEOF'
[Unit]
Description=FunctionFly Orchestrator API
After=network.target postgres.service redis-server.service
Wants=postgres.service redis-server.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/functionfly
EnvironmentFile=/opt/functionfly/.env
ExecStart=/opt/functionfly/orchestrator-api --skip-migrations
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/functionfly
PrivateTmp=true

[Install]
WantedBy=multi-user.target
SYSTEMDEOF
}

# Upload binary
upload_binary() {
    log_info "Uploading binary to ${SERVER}..."
    scp -o StrictHostKeyChecking=no "$BINARY_LOCAL" "${SERVER}:${REMOTE_BINARY}.new"
    ssh -o StrictHostKeyChecking=no "$SERVER" "mv ${REMOTE_BINARY}.new ${REMOTE_BINARY} && chmod +x ${REMOTE_BINARY}"
    log_success "Binary uploaded"
}

# Deploy fresh installation
deploy_fresh() {
    log_info "Starting fresh deployment to ${SERVER}..."

    # Create remote directory
    ssh -o StrictHostKeyChecking=no "$SERVER" "mkdir -p ${REMOTE_DIR}"

    # Upload binary
    upload_binary

    # Create environment file
    log_info "Creating environment file on server..."
    ssh -o StrictHostKeyChecking=no "$SERVER" "cat > ${REMOTE_ENV}.template << 'REMOTEENV'
# FunctionFly API Environment Configuration
# Fill in your actual values below

PORT=8080
DEVELOPMENT=false

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=CHANGE_ME
DB_NAME=functionfly
DB_SSLMODE=disable

TENANT_DB_ENABLED=false
TENANT_DB_HOST=localhost
TENANT_DB_PORT=5432
TENANT_DB_USER=postgres
TENANT_DB_PASSWORD=CHANGE_ME

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=CHANGE_ME_TO_32_BYTE_HEX
JWT_EXPIRATION=24h
PRIVACY_SALT=CHANGE_ME

SKIP_MIGRATION_VALIDATION=false
VERIFICATION_ENABLED=true
USE_SERVICE_REGISTRY=false
REMOTEENV"

    # Check if .env exists, if not create from template
    ssh -o StrictHostKeyChecking=no "$SERVER" "if [ ! -f ${REMOTE_ENV} ]; then cp ${REMOTE_ENV}.template ${REMOTE_ENV}; fi"

    # Create systemd service (requires sudo)
    log_info "Creating systemd service..."
    ssh -o StrictHostKeyChecking=no "$SERVER" "sudo tee /etc/systemd/system/${REMOTE_SERVICE}.service > /dev/null << 'REMOTESVC'
[Unit]
Description=FunctionFly Orchestrator API
After=network.target

[Service]
Type=simple
User=function
WorkingDirectory=/opt/functionfly
EnvironmentFile=/opt/functionfly/.env
ExecStart=/opt/functionfly/orchestrator-api --skip-migrations
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
REMOTESVC"

    # Reload systemd, enable and start service
    log_info "Enabling and starting service..."
    ssh -o StrictHostKeyChecking=no "$SERVER" "sudo systemctl daemon-reload && sudo systemctl enable ${REMOTE_SERVICE} && sudo systemctl start ${REMOTE_SERVICE}"
    ssh -o StrictHostKeyChecking=no "$SERVER" "sudo systemctl status ${REMOTE_SERVICE} --no-pager || true"

    log_success "Deployment complete!"
    log_info "Next steps:"
    echo "  1. Edit ${REMOTE_ENV} on the server with your actual values"
    echo "  2. Restart the service: systemctl restart ${REMOTE_SERVICE}"
    echo "  3. Check logs: journalctl -u ${REMOTE_SERVICE} -f"
}

# Update existing deployment
deploy_update() {
    log_info "Updating ${BINARY_NAME} on ${SERVER}..."

    # Upload binary
    upload_binary

    # Restart service
    log_info "Restarting service..."
    ssh -o StrictHostKeyChecking=no "$SERVER" "systemctl restart ${REMOTE_SERVICE} && systemctl status ${REMOTE_SERVICE} --no-pager"

    log_success "Update complete!"
}

# Full deployment (build + upload + config + restart)
deploy_full() {
    build_binary
    deploy_fresh
}

# Main
main() {
    log_info "Deploying ${BINARY_NAME} to ${SERVER} (mode: ${DEPLOY_MODE#--})"

    case "$DEPLOY_MODE" in
        --fresh)
            deploy_fresh
            ;;
        --update)
            deploy_update
            ;;
        --full)
            deploy_full
            ;;
    esac

    log_success "Done!"
}

main "$@"
