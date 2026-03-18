#!/bin/bash
#
# ============================================================================
# FunctionFly Production DNS Setup Script
# ============================================================================
# This script configures DNS records in Cloudflare for production deployment.
#
# Prerequisites:
#   - Cloudflare account with domain configured
#   - API token with Zone:DNS:Edit permissions
#   - Fly.io app deployed (functionfly-api)
#   - Cloudflare Pages project deployed (functionfly-dashboard)
#
# Usage:
#   ./setup-dns.sh [OPTIONS]
#
# Examples:
#   # Dry run - show what would be created
#   ./setup-dns.sh
#
#   # Apply DNS configuration
#   ./setup-dns.sh --apply --zone ZONE_ID --token API_TOKEN
#
#   # Custom Fly.io and Pages targets
#   ./setup-dns.sh --apply --fly-target my-api.fly.dev --pages-target my-dash.pages.dev
# ============================================================================

set -euo pipefail

# Configuration - can be overridden via environment or arguments
ZONE_NAME="${ZONE_NAME:-functionfly.com}"
FLY_API_TARGET="${FLY_API_TARGET:-functionfly-api.fly.dev}"
PAGES_DASHBOARD_TARGET="${PAGES_DASHBOARD_TARGET:-functionfly-dashboard.pages.dev}"
PAGES_DOCS_TARGET="${PAGES_DOCS_TARGET:-functionfly-docs.pages.dev}"
PAGES_MAIN_TARGET="${PAGES_MAIN_TARGET:-functionfly.pages.dev}"

# Cloudflare credentials (can be set via arguments)
CLOUDFLARE_ZONE_ID="${CLOUDFLARE_ZONE_ID:-}"
CLOUDFLARE_API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"

# Environment: production or staging
ENVIRONMENT="${ENVIRONMENT:-production}"

# Apply changes flag
APPLY_CHANGES=false

# ============================================================================
# Colors
# ============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# ============================================================================
# Helper Functions
# ============================================================================

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_cmd() { echo -e "${CYAN}$1${NC}"; }

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

OPTIONS:
    --apply                  Apply DNS changes to Cloudflare (default: dry-run)
    --zone ZONE_ID           Cloudflare Zone ID (or set CLOUDFLARE_ZONE_ID)
    --token TOKEN            Cloudflare API Token (or set CLOUDFLARE_API_TOKEN)
    --fly-target HOST        Fly.io API target (default: $FLY_API_TARGET)
    --pages-target HOST      Cloudflare Pages dashboard target (default: $PAGES_DASHBOARD_TARGET)
    --env ENV                Environment: production or staging (default: production)
    --help, -h               Show this help message

ENVIRONMENT VARIABLES:
    CLOUDFLARE_ZONE_ID       Cloudflare Zone ID
    CLOUDFLARE_API_TOKEN     Cloudflare API Token
    FLY_API_TARGET           Fly.io API hostname
    PAGES_DASHBOARD_TARGET   Cloudflare Pages dashboard hostname
    ZONE_NAME                Domain name (default: functionfly.com)

EXAMPLES:
    # Dry run (preview changes)
    CLOUDFLARE_ZONE_ID=abc123 CLOUDFLARE_API_TOKEN=xyz $0

    # Apply production DNS
    $0 --apply --zone abc123 --token xyz

    # Apply with custom targets
    $0 --apply --fly-target my-api.fly.dev --pages-target my-dash.pages.dev

EOF
}

# ============================================================================
# Argument Parsing
# ============================================================================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --apply)
                APPLY_CHANGES=true
                shift
                ;;
            --zone)
                CLOUDFLARE_ZONE_ID="$2"
                shift 2
                ;;
            --token)
                CLOUDFLARE_API_TOKEN="$2"
                shift 2
                ;;
            --fly-target)
                FLY_API_TARGET="$2"
                shift 2
                ;;
            --pages-target)
                PAGES_DASHBOARD_TARGET="$2"
                shift 2
                ;;
            --env)
                ENVIRONMENT="$2"
                shift 2
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
# Validation
# ============================================================================

validate() {
    if [ "$APPLY_CHANGES" = true ]; then
        if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
            log_error "CLOUDFLARE_ZONE_ID is required (use --zone or set environment variable)"
            exit 1
        fi

        if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
            log_error "CLOUDFLARE_API_TOKEN is required (use --token or set environment variable)"
            exit 1
        fi
    fi
}

# ============================================================================
# DNS Records Configuration
# ============================================================================

# Production DNS records
declare -a PROD_DNS_RECORDS=(
    # Format: TYPE|NAME|CONTENT|TTL|PROXIED|COMMENT
    "CNAME|api|$FLY_API_TARGET|60|true|Production API - Fly.io"
    "CNAME|app|$PAGES_DASHBOARD_TARGET|300|true|Production Dashboard - Cloudflare Pages"
    "CNAME|dashboard|$PAGES_DASHBOARD_TARGET|300|true|Legacy dashboard redirect"
    "CNAME|admin|$PAGES_DASHBOARD_TARGET|300|true|Admin UI - Cloudflare Pages"
    "CNAME|www|$PAGES_MAIN_TARGET|300|true|WWW redirect"
    "CNAME|@|$PAGES_MAIN_TARGET|300|true|Main website - Cloudflare Pages"
    "CNAME|docs|$PAGES_DOCS_TARGET|300|true|Documentation site"
)

# Staging DNS records
declare -a STAGING_DNS_RECORDS=(
    "CNAME|staging|$FLY_API_TARGET|60|true|Staging - Fly.io"
    "CNAME|api.staging|$FLY_API_TARGET|60|true|Staging API - Fly.io"
    "CNAME|app.staging|$PAGES_DASHBOARD_TARGET|300|true|Staging Dashboard - Cloudflare Pages"
    "CNAME|admin.staging|$PAGES_DASHBOARD_TARGET|300|true|Staging Admin UI"
    "CNAME|cdn.staging|cdn.staging.r2.cloudflarestorage.com|300|true|Staging CDN - R2"
)

# ============================================================================
# DNS Operations
# ============================================================================

get_record_id() {
    local name="$1"
    local type="$2"

    curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records?name=${name}.${ZONE_NAME}&type=${type}" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq -r '.result[0].id // empty'
}

upsert_record() {
    local type="$1"
    local name="$2"
    local content="$3"
    local ttl="$4"
    local proxied="$5"
    local comment="$6"

    local full_name="${name}.${ZONE_NAME}"
    if [ "$name" = "@" ]; then
        full_name="$ZONE_NAME"
    fi

    local proxied_json="true"
    [ "$proxied" = "false" ] && proxied_json="false"

    # Check if record exists
    local record_id
    record_id=$(get_record_id "$name" "$type")

    local data
    data=$(jq -n \
        --arg type "$type" \
        --arg name "$full_name" \
        --arg content "$content" \
        --arg proxied "$proxied_json" \
        --arg ttl "$ttl" \
        --arg comment "$comment" \
        '{
            type: $type,
            name: $name,
            content: $content,
            proxied: ($proxied == "true"),
            ttl: ($ttl | tonumber),
            comment: $comment
        }')

    if [ -n "$record_id" ]; then
        log_info "Updating record: $full_name ($type) -> $content"

        if [ "$APPLY_CHANGES" = true ]; then
            local response
            response=$(curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records/${record_id}" \
                -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
                -H "Content-Type: application/json" \
                -d "$data")

            if echo "$response" | grep -q '"success":true'; then
                log_success "Updated: $full_name"
            else
                log_error "Failed to update: $full_name"
                echo "$response" | jq -r '.errors[].message // "Unknown error"'
                return 1
            fi
        fi
    else
        log_info "Creating record: $full_name ($type) -> $content"

        if [ "$APPLY_CHANGES" = true ]; then
            local response
            response=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
                -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
                -H "Content-Type: application/json" \
                -d "$data")

            if echo "$response" | grep -q '"success":true'; then
                log_success "Created: $full_name"
            else
                log_error "Failed to create: $full_name"
                echo "$response" | jq -r '.errors[].message // "Unknown error"'
                return 1
            fi
        fi
    fi

    return 0
}

apply_dns_records() {
    local -n records="$1"

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "DNS Records: ${ENVIRONMENT^}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    printf "${CYAN}%-8s ${CYAN}%-20s ${CYAN}%-35s ${CYAN}%-6s ${CYAN}%-8s${NC}\n" "TYPE" "NAME" "TARGET" "TTL" "PROXIED"
    echo "────────────────────────────────────────────────────────────────────────────────"

    for record in "${records[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"
        printf "%-8s %-20s %-35s %-6s %-8s\n" "$type" "$name.${ZONE_NAME}" "$content" "$ttl" "$proxied"
    done

    echo

    if [ "$APPLY_CHANGES" = true ]; then
        log_info "Applying DNS changes..."
        echo

        for record in "${records[@]}"; do
            IFS='|' read -r type name content ttl proxied comment <<< "$record"
            upsert_record "$type" "$name" "$content" "$ttl" "$proxied" "$comment" || true
        done

        log_success "DNS configuration applied!"
    else
        log_warn "DRY RUN - No changes applied"
        log_info "Use --apply to apply changes"
    fi
}

verify_dns() {
    if [ "$APPLY_CHANGES" != true ]; then
        return 0
    fi

    echo
    log_info "Verifying DNS configuration..."

    local records
    records=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq '.result | length')

    log_info "Found $records DNS records in zone"
}

print_summary() {
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    log_info "Zone: ${ZONE_NAME}"
    log_info "Environment: ${ENVIRONMENT}"
    log_info "Fly.io API: ${FLY_API_TARGET}"
    log_info "Dashboard: ${PAGES_DASHBOARD_TARGET}"
    log_info "Docs: ${PAGES_DOCS_TARGET}"
    echo

    if [ "$APPLY_CHANGES" = true ]; then
        log_success "DNS records have been configured!"
        echo
        log_info "Next steps:"
        echo "  1. Wait for DNS propagation (up to 5 minutes)"
        echo "  2. Verify SSL/TLS in Cloudflare Dashboard"
        echo "  3. Test endpoints:"
        echo "     - https://api.${ZONE_NAME}/healthz"
        echo "     - https://app.${ZONE_NAME}"
    else
        log_info "Run with --apply to apply these DNS changes"
    fi
}

# ============================================================================
# Main
# ============================================================================

main() {
    echo "🌐 FunctionFly Production DNS Setup"
    echo "==================================="
    echo

    parse_args "$@"
    validate

    # Select records based on environment
    if [ "$ENVIRONMENT" = "staging" ]; then
        apply_dns_records STAGING_DNS_RECORDS
    else
        apply_dns_records PROD_DNS_RECORDS
    fi

    verify_dns
    print_summary
}

main "$@"
