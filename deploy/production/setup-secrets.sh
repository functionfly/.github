#!/bin/bash
# FunctionFly Docker Production Secrets Setup
# Run this script before starting services in production
# Usage: ./setup-secrets.sh [--generate-certs] [--skip-certs]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SECRETS_DIR="${SCRIPT_DIR}/secrets"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"; }

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --generate-certs    Generate self-signed SSL certificates (dev/test only)"
    echo "  --skip-certs       Skip SSL certificate setup"
    echo "  --help             Show this help message"
    echo ""
    echo "This script sets up the secrets directory for Docker production deployment."
    echo "It creates:"
    echo "  - secrets/ directory"
    echo "  - PostgreSQL SSL certificates (optional)"
    echo "  - DataDog API key placeholder (optional)"
}

# Parse arguments
GENERATE_CERTS=false
SKIP_CERTS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --generate-certs)
            GENERATE_CERTS=true
            shift
            ;;
        --skip-certs)
            SKIP_CERTS=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check if running from correct directory
if [ ! -f "docker-compose.production.yml" ] && [ ! -f "deploy/production/docker-compose.yml" ]; then
    log_error "This script must be run from the repository root"
    exit 1
fi

# Create secrets directory
setup_secrets_dir() {
    log_info "Creating secrets directory..."

    if [ -d "$SECRETS_DIR" ]; then
        log_warn "Secrets directory already exists at $SECRETS_DIR"
        read -p "Continue and potentially overwrite existing secrets? (y/N): " confirm
        if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
            log_info "Aborted"
            exit 0
        fi
    else
        mkdir -p "$SECRETS_DIR"
        chmod 750 "$SECRETS_DIR"
        log_success "Created secrets directory: $SECRETS_DIR"
    fi
}

# Generate self-signed SSL certificates
generate_ssl_certs() {
    log_info "Generating self-signed SSL certificates for PostgreSQL..."

    local CERT_FILE="${SECRETS_DIR}/postgresql.crt"
    local KEY_FILE="${SECRETS_DIR}/postgresql.key"

    # Generate certificate
    openssl req -new -x509 -days 365 -nodes \
        -text \
        -out "$CERT_FILE" \
        -keyout "$KEY_FILE" \
        -subj "/C=US/ST=State/L=City/O=FunctionFly/CN=postgres" 2>/dev/null

    chmod 644 "$CERT_FILE"
    chmod 600 "$KEY_FILE"

    log_success "Generated SSL certificates:"
    log_success "  Certificate: $CERT_FILE"
    log_success "  Key: $KEY_FILE"
    log_warn "  NOTE: Use valid certificates from Let's Encrypt or your CA in production!"
}

# Setup DataDog API key placeholder
setup_datadog_placeholder() {
    local DD_FILE="${SECRETS_DIR}/datadog_api_key"
    if [ ! -f "$DD_FILE" ]; then
        touch "$DD_FILE"
        chmod 600 "$DD_FILE"
        echo "# Place your DataDog API key here" > "$DD_FILE"
        log_info "Created DataDog placeholder: $DD_FILE"
        log_info "  Edit this file and add your DataDog API key to enable remote_write"
    fi
}

# Generate FRG webhook secret for HMAC-SHA256 signature verification
generate_frg_webhook_secret() {
    local FRG_FILE="${SECRETS_DIR}/frg_webhook_secret"
    if [ ! -f "$FRG_FILE" ] || [ ! -s "$FRG_FILE" ]; then
        openssl rand -hex 32 > "$FRG_FILE"
        chmod 600 "$FRG_FILE"
        log_success "Generated FRG webhook secret: $FRG_FILE"
    else
        log_info "FRG webhook secret already exists: $FRG_FILE"
    fi
}

# Generate monitoring basic auth hash for Grafana/Prometheus endpoints
generate_monitoring_auth_hash() {
    local AUTH_FILE="${SECRETS_DIR}/monitoring_basic_auth_hash"
    if [ ! -f "$AUTH_FILE" ] || [ ! -s "$AUTH_FILE" ]; then
        if command -v docker &> /dev/null; then
            docker run -it --rm caddy:2-alpine caddy hash-password --algorithm bcrypt --password "$(openssl rand -base64 32)" > "$AUTH_FILE" 2>/dev/null || \
                openssl rand -hex 32 > "$AUTH_FILE"
        else
            openssl rand -hex 32 > "$AUTH_FILE"
        fi
        chmod 600 "$AUTH_FILE"
        log_success "Generated monitoring basic auth hash: $AUTH_FILE"
    else
        log_info "Monitoring basic auth hash already exists: $AUTH_FILE"
    fi
}

# Generate JWT secret
generate_jwt_secret() {
    local JWT_FILE="${SECRETS_DIR}/jwt_secret"
    if [ ! -f "$JWT_FILE" ] || [ ! -s "$JWT_FILE" ]; then
        openssl rand -hex 64 > "$JWT_FILE"
        chmod 600 "$JWT_FILE"
        log_success "Generated JWT secret: $JWT_FILE"
    else
        log_info "JWT secret already exists: $JWT_FILE"
    fi
}

# Generate API shared secret
generate_api_shared_secret() {
    local API_FILE="${SECRETS_DIR}/api_shared_secret"
    if [ ! -f "$API_FILE" ] || [ ! -s "$API_FILE" ]; then
        openssl rand -hex 64 > "$API_FILE"
        chmod 600 "$API_FILE"
        log_success "Generated API shared secret: $API_FILE"
    else
        log_info "API shared secret already exists: $API_FILE"
    fi
}

# Export secrets to environment file for docker-compose
export_secrets_to_env() {
    local ENV_FILE="${SECRETS_DIR}/secrets.env"
    log_info "Exporting secrets to $ENV_FILE for docker-compose..."

    {
        echo "# Auto-generated secrets - do not commit to version control"
        echo "# Generated at $(date -u +%Y-%m-%dT%H:%M:%SZ)"
        echo ""
        [ -f "${SECRETS_DIR}/jwt_secret" ] && echo "JWT_SECRET=$(cat "${SECRETS_DIR}/jwt_secret")"
        [ -f "${SECRETS_DIR}/api_shared_secret" ] && echo "API_SHARED_SECRET=$(cat "${SECRETS_DIR}/api_shared_secret")"
        [ -f "${SECRETS_DIR}/frg_webhook_secret" ] && echo "FRG_WEBHOOK_SECRET=$(cat "${SECRETS_DIR}/frg_webhook_secret")"
        [ -f "${SECRETS_DIR}/monitoring_basic_auth_hash" ] && echo "MONITORING_BASIC_AUTH_HASH=$(cat "${SECRETS_DIR}/monitoring_basic_auth_hash")"
    } > "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    log_success "Secrets exported to $ENV_FILE"
    log_info "Use: docker-compose --env-file deploy/production/secrets/secrets.env up -d"
}

# Verify required files exist
verify_secrets() {
    log_info "Verifying secrets..."

    local missing=0

    # Check PostgreSQL certificates
    if [ -f "${SECRETS_DIR}/postgresql.crt" ] && [ -f "${SECRETS_DIR}/postgresql.key" ]; then
        log_success "PostgreSQL SSL certificates: OK"
    else
        log_warn "PostgreSQL SSL certificates: MISSING"
        log_warn "  Required for database SSL/TLS connections"
        missing=1
    fi

    # Check FRG webhook secret (required)
    if [ -f "${SECRETS_DIR}/frg_webhook_secret" ] && [ -s "${SECRETS_DIR}/frg_webhook_secret" ]; then
        log_success "FRG webhook secret: OK"
    else
        log_error "FRG webhook secret: MISSING"
        log_error "  Required for webhook HMAC verification"
        missing=1
    fi

    # Check monitoring basic auth hash (required)
    if [ -f "${SECRETS_DIR}/monitoring_basic_auth_hash" ] && [ -s "${SECRETS_DIR}/monitoring_basic_auth_hash" ]; then
        log_success "Monitoring basic auth hash: OK"
    else
        log_error "Monitoring basic auth hash: MISSING"
        log_error "  Required for protecting monitoring endpoints"
        missing=1
    fi

    # Check JWT secret (required)
    if [ -f "${SECRETS_DIR}/jwt_secret" ] && [ -s "${SECRETS_DIR}/jwt_secret" ]; then
        log_success "JWT secret: OK"
    else
        log_error "JWT secret: MISSING"
        log_error "  Required for JWT token signing"
        missing=1
    fi

    # Check API shared secret (required)
    if [ -f "${SECRETS_DIR}/api_shared_secret" ] && [ -s "${SECRETS_DIR}/api_shared_secret" ]; then
        log_success "API shared secret: OK"
    else
        log_error "API shared secret: MISSING"
        log_error "  Required for internal API authentication"
        missing=1
    fi

    # Check DataDog API key (optional)
    if [ -f "${SECRETS_DIR}/datadog_api_key" ] && [ -s "${SECRETS_DIR}/datadog_api_key" ] && ! grep -q "^#" "${SECRETS_DIR}/datadog_api_key"; then
        log_success "DataDog API key: OK"
    else
        log_warn "DataDog API key: Not configured (optional)"
    fi

    return $missing
}

# Print environment variables needed
print_env_vars() {
    log_info "=========================================="
    log_info "Required Environment Variables"
    log_info "=========================================="
    echo ""
    echo "Set these in your .env.production file or environment:"
    echo ""
    echo "  DB_PASSWORD=<your-postgres-password>"
    echo "  REDIS_PASSWORD=<your-redis-password>"
    echo "  JWT_SECRET=<your-jwt-secret-min-32-chars>"
    echo "  API_SHARED_SECRET=<your-api-shared-secret>"
    echo "  SSL_EMAIL=ops@yourdomain.com"
    echo "  SSL_DOMAIN=yourdomain.com"
    echo ""
    echo "Optional:"
    echo "  STRIPE_SECRET_KEY=<your-stripe-key>"
    echo "  RESEND_API_KEY=<your-resend-key>"
    echo "  DATADOG_ENABLED=true  # If DataDog API key is set"
    echo ""
}

# Main
main() {
    echo "=========================================="
    echo "FunctionFly Production Secrets Setup"
    echo "=========================================="
    echo ""

    setup_secrets_dir

    if [ "$SKIP_CERTS" = "false" ]; then
        if [ "$GENERATE_CERTS" = "true" ]; then
            generate_ssl_certs
        else
            log_info "Skipping SSL certificate generation (use --generate-certs for self-signed)"
            log_info "To set up SSL certificates manually:"
            echo "  1. For Let's Encrypt certificates:"
            echo "     ln -s /etc/letsencrypt/live/\$(hostname)/fullchain.pem secrets/postgresql.crt"
            echo "     ln -s /etc/letsencrypt/live/\$(hostname)/privkey.pem secrets/postgresql.key"
            echo ""
            echo "  2. For production certificates from your CA:"
            echo "     cp /path/to/your/cert.pem secrets/postgresql.crt"
            echo "     cp /path/to/your/key.pem secrets/postgresql.key"
            echo ""
        fi
    else
        log_info "Skipping SSL certificate setup"
    fi

    setup_datadog_placeholder

    # Generate required application secrets
    echo ""
    log_info "Generating application secrets..."
    generate_jwt_secret
    generate_api_shared_secret
    generate_frg_webhook_secret
    generate_monitoring_auth_hash

    # Export secrets to env file for docker-compose
    export_secrets_to_env

    echo ""
    verify_secrets
    print_env_vars

    echo ""
    log_success "Secrets setup complete!"
    log_info "Next steps:"
    echo "  1. Set required environment variables (or use secrets.env)"
    echo "  2. Run: cd deploy/production && docker-compose --env-file secrets/secrets.env up -d"
    echo "  3. Verify health: curl http://localhost:8080/health/ready"
    echo ""
    log_warn "IMPORTANT: Add deploy/production/secrets/ to your .gitignore!"
}

main
