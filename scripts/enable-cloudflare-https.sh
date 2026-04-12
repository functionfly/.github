#!/bin/bash
set -euo pipefail

# Enable "Always Use HTTPS" for Cloudflare zones
# This script configures Cloudflare settings to enforce HTTPS connections

# Configuration
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN:-}
CLOUDFLARE_ZONE_ID=${CLOUDFLARE_ZONE_ID:-}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check dependencies
check_dependencies() {
    if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
        log_error "CLOUDFLARE_API_TOKEN not set"
        exit 1
    fi

    if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
        log_error "CLOUDFLARE_ZONE_ID not set"
        exit 1
    fi
}

# Enable "Always Use HTTPS" setting
enable_always_use_https() {
    log_info "Enabling 'Always Use HTTPS' for zone..."

    local response
    response=$(curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/always_use_https" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{"value":"on"}')

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        log_info "Successfully enabled 'Always Use HTTPS'"
    else
        log_error "Failed to enable 'Always Use HTTPS'"
        echo "$response" | jq -r '.errors // .messages // "Unknown error"'
        return 1
    fi
}

# Set SSL/TLS mode to Full (strict)
set_ssl_mode() {
    log_info "Setting SSL/TLS mode to Full (strict)..."

    local response
    response=$(curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/ssl" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{"value":"full_strict"}')

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        log_info "Successfully set SSL/TLS mode to Full (strict)"
    else
        log_warn "Failed to set SSL/TLS mode"
        echo "$response" | jq -r '.errors // .messages // "Unknown error"'
    fi
}

# Enable Automatic HTTPS Rewrites
enable_https_rewrites() {
    log_info "Enabling Automatic HTTPS Rewrites..."

    local response
    response=$(curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/automatic_https_rewrites" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{"value":"on"}')

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        log_info "Successfully enabled Automatic HTTPS Rewrites"
    else
        log_warn "Failed to enable Automatic HTTPS Rewrites"
        echo "$response" | jq -r '.errors // .messages // "Unknown error"'
    fi
}

# Verify settings
verify_settings() {
    log_info "Verifying HTTPS settings..."

    local settings
    settings=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json")

    echo "$settings" | jq -r '.result[] | select(.id | test("always_use_https|ssl|automatic_https_rewrites")) | "\(.id): \(.value)"'
}

# Main
main() {
    log_info "Cloudflare HTTPS Security Configuration"
    log_info "======================================="

    check_dependencies

    case "${1:-apply}" in
        apply|enable)
            enable_always_use_https
            set_ssl_mode
            enable_https_rewrites
            log_info "Done! HTTPS security settings applied."
            ;;
        verify)
            verify_settings
            ;;
        *)
            echo "Usage: $0 {apply|verify}"
            echo ""
            echo "Commands:"
            echo "  apply   Enable HTTPS security settings"
            echo "  verify  Verify current HTTPS settings"
            exit 1
            ;;
    esac
}

main "$@"
