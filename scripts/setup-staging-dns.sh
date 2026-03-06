#!/bin/bash

# ============================================================================
# FunctionFly Staging DNS Configuration Script
# ============================================================================
# This script generates Cloudflare CLI commands and API calls needed to create
# the required DNS records for FunctionFly staging subdomains.
#
# Usage: ./scripts/setup-staging-dns.sh [--apply] [--zone ZONE_ID] [--token TOKEN]
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
ZONE_NAME="functionfly.com"
APPLY_CHANGES=false
CF_API_TOKEN=""
CF_ZONE_ID=""

# Staging targets (update these with your actual staging infrastructure)
STAGING_MAIN_TARGET="functionfly-staging.iad.fly.dev"
STAGING_API_TARGET="functionfly-staging.iad.fly.dev"
STAGING_EDGE_TARGET="functionfly-staging-edge.iad.fly.dev"
STAGING_CDN_TARGET="functionfly-staging-cdn.r2.cloudflarestorage.com"
STAGING_DASHBOARD_TARGET="functionfly-staging-dashboard.pages.dev"

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

log_cmd() {
    echo -e "${CYAN}$1${NC}"
}

show_usage() {
    cat << EOF
Usage: ./scripts/setup-staging-dns.sh [OPTIONS]

OPTIONS:
    --apply                  Actually apply the DNS changes (requires --zone and --token)
    --zone ZONE_ID           Cloudflare Zone ID for functionfly.com
    --token TOKEN            Cloudflare API Token with DNS edit permissions
    --target-main HOST       Target for staging.functionfly.com (default: $STAGING_MAIN_TARGET)
    --target-api HOST        Target for api.staging.functionfly.com (default: $STAGING_API_TARGET)
    --target-edge HOST       Target for edge.staging.functionfly.com (default: $STAGING_EDGE_TARGET)
    --target-cdn HOST        Target for cdn.staging.functionfly.com (default: $STAGING_CDN_TARGET)
    --target-dashboard HOST  Target for app.staging.functionfly.com (default: $STAGING_DASHBOARD_TARGET)
    --help, -h               Show this help message

EXAMPLES:
    # Generate commands only (dry run)
    ./scripts/setup-staging-dns.sh

    # Apply changes with explicit credentials
    ./scripts/setup-staging-dns.sh --apply --zone YOUR_ZONE_ID --token YOUR_API_TOKEN

    # Custom targets
    ./scripts/setup-staging-dns.sh --target-main my-staging.fly.dev --target-api my-staging.fly.dev

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --apply)
                APPLY_CHANGES=true
                shift
                ;;
            --zone)
                CF_ZONE_ID="$2"
                shift 2
                ;;
            --token)
                CF_API_TOKEN="$2"
                shift 2
                ;;
            --target-main)
                STAGING_MAIN_TARGET="$2"
                shift 2
                ;;
            --target-api)
                STAGING_API_TARGET="$2"
                shift 2
                ;;
            --target-edge)
                STAGING_EDGE_TARGET="$2"
                shift 2
                ;;
            --target-cdn)
                STAGING_CDN_TARGET="$2"
                shift 2
                ;;
            --target-dashboard)
                STAGING_DASHBOARD_TARGET="$2"
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
# DNS Record Definitions
# ============================================================================

declare -a DNS_RECORDS=(
    # Format: TYPE|NAME|CONTENT|TTL|PROXIED|COMMENT

    # Main Staging Domain
    "CNAME|staging|$STAGING_MAIN_TARGET|300|true|Staging Environment - Main Entry"

    # API Staging Subdomain
    "CNAME|api.staging|$STAGING_API_TARGET|60|true|Staging API - Single region, no GeoDNS"

    # Edge Staging Subdomain
    "CNAME|edge.staging|$STAGING_EDGE_TARGET|300|true|Staging Edge Functions"

    # CDN Staging Subdomain
    "CNAME|cdn.staging|$STAGING_CDN_TARGET|300|true|Staging CDN - R2 Storage"

    # Dashboard/App Staging Subdomain
    "CNAME|app.staging|$STAGING_DASHBOARD_TARGET|300|true|Staging Dashboard - Cloudflare Pages"

    # Admin UI Staging Subdomain (separate origin)
    "CNAME|admin.staging|$STAGING_DASHBOARD_TARGET|300|true|Staging Admin UI - Cloudflare Pages"
)

# ============================================================================
# Cloudflare CLI Functions
# ============================================================================

check_cloudflare_cli() {
    if ! command -v cloudflare &> /dev/null; then
        log_warning "Cloudflare CLI not found. Install with:"
        log_info "  npm install -g cloudflared"
        log_info "  or"
        log_info "  brew install cloudflared"
        log_info ""
        log_info "Alternatively, use the curl API commands shown below."
        return 1
    fi
    return 0
}

validate_credentials() {
    if [ "$APPLY_CHANGES" = true ]; then
        if [ -z "$CF_ZONE_ID" ]; then
            log_error "Zone ID is required when using --apply"
            log_info "Get your Zone ID from Cloudflare dashboard > Overview > API section"
            exit 1
        fi

        if [ -z "$CF_API_TOKEN" ]; then
            log_error "API Token is required when using --apply"
            log_info "Create a token at: https://dash.cloudflare.com/profile/api-tokens"
            log_info "Required permissions: Zone > DNS > Edit"
            exit 1
        fi
    fi
}

# ============================================================================
# Output Generation Functions
# ============================================================================

print_table_header() {
    echo
    echo "╔════════╦══════════════════════════╦══════════════════════════════════════╦══════╦══════════╗"
    echo "║ Type   ║ Name                     ║ Target                               ║ TTL  ║ Proxied  ║"
    echo "╠════════╬══════════════════════════╬══════════════════════════════════════╬══════╬══════════╣"
}

print_table_row() {
    local type="$1"
    local name="$2"
    local content="$3"
    local ttl="$4"
    local proxied="$5"

    printf "║ %-6s ║ %-24s ║ %-36s ║ %-4s ║ %-8s ║\n" "$type" "$name" "$content" "$ttl" "$proxied"
}

print_table_footer() {
    echo "╚════════╩══════════════════════════╩══════════════════════════════════════╩══════╩══════════╝"
    echo
}

print_cloudflare_cli_commands() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Cloudflare CLI Commands"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    for record in "${DNS_RECORDS[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"

        local proxied_flag="--proxied"
        [ "$proxied" = "false" ] && proxied_flag=""

        log_cmd "cloudflared tunnel route dns $name $content"
        echo "# $comment"
        echo
    done
}

print_curl_api_commands() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "cURL API Commands (requires CF_ZONE_ID and CF_API_TOKEN)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    for record in "${DNS_RECORDS[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"

        local proxied_json="true"
        [ "$proxied" = "false" ] && proxied_json="false"

        log_cmd "curl -X POST \"https://api.cloudflare.com/client/v4/zones/\${CF_ZONE_ID}/dns_records\" \\"
        log_cmd "  -H \"Authorization: Bearer \${CF_API_TOKEN}\" \\"
        log_cmd "  -H \"Content-Type: application/json\" \\"
        log_cmd "  -d '{\"type\":\"$type\",\"name\":\"$name\",\"content\":\"$content\",\"ttl\":$ttl,\"proxied\":$proxied_json}'"
        log_cmd "# $comment"
        echo
    done
}

print_terraform_config() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Terraform Configuration"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    cat << 'EOF'
# Add to your Terraform configuration (deploy/dns/staging.tf)

terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for functionfly.com"
  type        = string
}

EOF

    local idx=0
    for record in "${DNS_RECORDS[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"

        local proxied_tf="true"
        [ "$proxied" = "false" ] && proxied_tf="false"

        cat << EOF
# $comment
resource "cloudflare_record" "staging_${name//./_}" {
  zone_id = var.cloudflare_zone_id
  name    = "$name"
  type    = "$type"
  value   = "$content"
  ttl     = $ttl
  proxied = $proxied_tf
}

EOF
        idx=$((idx + 1))
    done
}

# ============================================================================
# Application Functions
# ============================================================================

apply_dns_records() {
    if [ "$APPLY_CHANGES" = false ]; then
        return 0
    fi

    log_info "Applying DNS records to Cloudflare..."

    for record in "${DNS_RECORDS[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"

        local proxied_json="true"
        [ "$proxied" = "false" ] && proxied_json="false"

        log_info "Creating $name -> $content"

        local response
        response=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records" \
            -H "Authorization: Bearer ${CF_API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"type\":\"$type\",\"name\":\"$name\",\"content\":\"$content\",\"ttl\":$ttl,\"proxied\":$proxied_json,\"comment\":\"$comment\"}")

        if echo "$response" | grep -q '"success":true'; then
            log_success "Created $name"
        else
            log_error "Failed to create $name"
            echo "$response" | grep -o '"message":"[^"]*"' | head -1
        fi
    done

    log_success "DNS records applied successfully!"
}

print_next_steps() {
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Next Steps"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    log_info "1. Verify DNS propagation:"
    log_cmd "   dig staging.functionfly.com"
    log_cmd "   dig api.staging.functionfly.com"
    echo
    log_info "2. Test SSL/TLS configuration in Cloudflare Dashboard:"
    log_cmd "   https://dash.cloudflare.com > SSL/TLS > Overview"
    log_info "   Recommended: Full (strict) mode"
    echo
    log_info "3. Verify staging environment is accessible:"
    log_cmd "   curl https://staging.functionfly.com/health"
    log_cmd "   curl https://api.staging.functionfly.com/healthz"
    echo
    log_info "4. Configure Page Rules (optional):"
    log_info "   - Cache Level: Cache Everything for cdn.staging.functionfly.com/*"
    log_info "   - Security Level: High for edge.staging.functionfly.com/*"
    echo
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    echo "🌐 FunctionFly Staging DNS Configuration"
    echo "========================================"
    echo

    parse_args "$@"
    validate_credentials

    # Print summary of records to be created
    log_info "Staging DNS Records to be created:"
    print_table_header

    for record in "${DNS_RECORDS[@]}"; do
        IFS='|' read -r type name content ttl proxied comment <<< "$record"
        print_table_row "$type" "$name.$ZONE_NAME" "$content" "$ttl" "$proxied"
    done

    print_table_footer

    # Show command options
    if [ "$APPLY_CHANGES" = false ]; then
        log_info "DRY RUN MODE - No changes will be applied"
        log_info "Use --apply flag to actually create the DNS records"
        echo
    fi

    # Show various command options
    print_cloudflare_cli_commands
    print_curl_api_commands
    print_terraform_config

    # Apply if requested
    apply_dns_records

    # Show next steps
    print_next_steps
}

# Run main function
main "$@"
