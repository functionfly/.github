#!/bin/bash
set -euo pipefail

# Cloudflare DNS Update Script
# Updates DNS records based on the multi-region deployment configuration

# Configuration
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN:-}
CLOUDFLARE_ZONE_ID=${CLOUDFLARE_ZONE_ID:-}
CONFIG_FILE="deploy/dns/cloudflare-geo-dns.json"
DRY_RUN=${DRY_RUN:-false}

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

    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Config file not found: $CONFIG_FILE"
        exit 1
    fi
}

# Get current public IP or hostname for a Fly.io region.
# If flyctl is available and FLY_APP_NAME is set (or default functionfly-control), returns the
# first allocated IP for that region; otherwise returns the app hostname for the region.
get_region_ip() {
    local region=$1
    local app="${FLY_APP_NAME:-functionfly-control}"
    local ip
    if command -v flyctl &>/dev/null; then
        ip=$(flyctl ips list -a "$app" --json 2>/dev/null | \
            jq -r --arg r "$region" '(.[]? | select(.region == $r) | .address // .Address) // empty' | head -1)
    fi
    if [ -n "$ip" ]; then
        echo "$ip"
    else
        echo "${app}.${region}.fly.dev"
    fi
}

# Create or update a DNS record
upsert_dns_record() {
    local type=$1
    local name=$2
    local content=$3
    local proxied=${4:-false}
    local ttl=${5:-60}
    local geo=${6:-}

    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would create/update record: $name ($type) -> $content"
        return 0
    fi

    # Check if record exists
    local record_id
    record_id=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records?name=${name}.functionfly.com&type=${type}" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq -r '.result[0].id // empty')

    local data
    if [ -n "$geo" ] && [ "$geo" != "{}" ]; then
        # Handle geo routing
        data=$(jq -n \
            --arg type "$type" \
            --arg name "$name" \
            --arg content "$content" \
            --arg proxied "$proxied" \
            --arg ttl "$ttl" \
            --argjson geo "$geo" \
            '{
                type: $type,
                name: $name,
                content: $content,
                proxied: ($proxied == "true"),
                ttl: ($ttl | tonumber),
                data: $geo
            }')
    else
        data=$(jq -n \
            --arg type "$type" \
            --arg name "$name" \
            --arg content "$content" \
            --arg proxied "$proxied" \
            --arg ttl "$ttl" \
            '{
                type: $type,
                name: $name,
                content: $content,
                proxied: ($proxied == "true"),
                ttl: ($ttl | tonumber)
            }')
    fi

    if [ -n "$record_id" ]; then
        log_info "Updating existing record: $name"
        curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records/${record_id}" \
            -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d "$data" | jq -r '.success'
    else
        log_info "Creating new record: $name"
        curl -s -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
            -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d "$data" | jq -r '.success'
    fi
}

# Delete a DNS record
delete_dns_record() {
    local name=$1
    local type=${2:-A}

    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would delete record: $name ($type)"
        return 0
    fi

    local record_id
    record_id=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records?name=${name}.functionfly.com&type=${type}" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq -r '.result[0].id // empty')

    if [ -n "$record_id" ]; then
        log_info "Deleting record: $name ($type)"
        curl -s -X DELETE "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records/${record_id}" \
            -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" | jq -r '.success'
    fi
}

# Get Fly.io region IPs
get_fly_ips() {
    log_info "Getting Fly.io region IPs..."

    # Primary region
    IAD_IP=$(flyctl ips list --region iad --app functionfly-control 2>/dev/null | \
        jq -r '.[0].Address' || echo "")

    # Secondary region
    LAX_IP=$(flyctl ips list --region lax --app functionfly-control 2>/dev/null | \
        jq -r '.[0].Address' || echo "")

    # Tertiary region
    FRA_IP=$(flyctl ips list --region fra --app functionfly-control 2>/dev/null | \
        jq -r '.[0].Address' || echo "")

    log_info "iad: ${IAD_IP:-not configured}"
    log_info "lax: ${LAX_IP:-not configured}"
    log_info "fra: ${FRA_IP:-not configured}"
}

# Apply GeoDNS configuration
apply_geodns() {
    log_info "Applying GeoDNS configuration..."

    # Get IPs from Fly.io
    get_fly_ips

    # Set defaults if not available
    IAD_IP=${IAD_IP:-"functionfly-control.iad.fly.dev"}
    LAX_IP=${LAX_IP:-"functionfly-control.lax.fly.dev"}
    FRA_IP=${FRA_IP:-"functionfly-control.fra.fly.dev"}

    # Update API record with geo routing
    # This uses Cloudflare's Traffic Steering feature

    # For AAAA records (IPv6)
    upsert_dns_record "AAAA" "api" "$IAD_IP" "true" "60"

    # Note: Cloudflare GeoDNS is configured via the dashboard or API
    # The JSON config file serves as documentation

    log_info "GeoDNS configuration updated"
}

# Apply basic DNS configuration (non-geo)
apply_basic_dns() {
    log_info "Applying basic DNS configuration..."

    # Main API record - points to primary region
    upsert_dns_record "A" "api" "functionfly-control.iad.fly.dev" "true" "60"

    # Website
    upsert_dns_record "CNAME" "www" "functionfly.pages.dev" "true" "300"
    upsert_dns_record "CNAME" "@" "functionfly.pages.dev" "true" "300"

    # Dashboard and docs
    upsert_dns_record "CNAME" "dashboard" "functionfly-dashboard.pages.dev" "true" "300"
    upsert_dns_record "CNAME" "docs" "functionfly-docs.pages.dev" "true" "300"

    log_info "Basic DNS configuration updated"
}

# Verify DNS configuration
verify_dns() {
    log_info "Verifying DNS configuration..."

    local records
    records=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq '.result | length')

    log_info "Found $records DNS records"

    # Show API record
    curl -s -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records?name=api.functionfly.com" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" | \
        jq '.result[0] | {name, type, content, proxied, ttl}'
}

# Main
main() {
    log_info "Cloudflare DNS Update Script"
    log_info "================================"

    check_dependencies

    case "${1:-apply}" in
        apply)
            apply_basic_dns
            ;;
        geodns)
            apply_geodns
            ;;
        verify)
            verify_dns
            ;;
        delete)
            delete_dns_record "$2" "${3:-A}"
            ;;
        *)
            echo "Usage: $0 {apply|geodns|verify|delete}"
            echo ""
            echo "Commands:"
            echo "  apply          Apply basic DNS configuration"
            echo "  geodns         Apply GeoDNS configuration"
            echo "  verify         Verify DNS configuration"
            echo "  delete <name>  Delete a DNS record"
            exit 1
            ;;
    esac

    log_info "Done!"
}

main "$@"
