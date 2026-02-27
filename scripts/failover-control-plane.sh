#!/bin/bash
set -euo pipefail

# Automated Failover Script for Multi-Region Control Plane
# Promotes a secondary region to primary when the primary fails

# Configuration
PRIMARY_REGION=${PRIMARY_REGION:-iad}
SECONDARY_REGION=${SECONDARY_REGION:-lax}
TERTIARY_REGION=${TERTIARY_REGION:-fra}
HEALTH_CHECK_URL=${HEALTH_CHECK_URL:-http://localhost:8080/healthz}
FAILOVER_THRESHOLD=${FAILOVER_THRESHOLD:-3}
CHECK_INTERVAL=${CHECK_INTERVAL:-2}
TIMEOUT=${TIMEOUT:-5}
DRY_RUN=${DRY_RUN:-false}

# Cloudflare configuration
CLOUDFLARE_ZONE_ID=${CLOUDFLARE_ZONE_ID:-}
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN:-}

# Fly.io configuration
FLY_APP=${FLY_APP:-functionfly-control}

# Alert configuration
ALERT_WEBHOOK_URL=${ALERT_WEBHOOK_URL:-}
STATUS_PAGE_URL=${STATUS_PAGE_URL:-}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_debug() { echo -e "${BLUE}[DEBUG]${NC} $1"; }

# Check if running in dry-run mode
if [ "$DRY_RUN" = "true" ]; then
    log_warn "DRY RUN MODE - No actual changes will be made"
fi

# Get region endpoint
get_region_endpoint() {
    local region=$1
    # Fly.io automatically assigns IPs based on region
    # Format: region.app.internal for internal DNS
    echo "${FLY_APP}.${region}.internal"
}

# Check primary region health
check_region_health() {
    local region=$1
    local health=0

    log_debug "Checking health of region: $region"

    for i in $(seq 1 $FAILOVER_THRESHOLD); do
        # Try to connect to the health endpoint
        if curl -sf --max-time "$TIMEOUT" "http://${region}.${FLY_APP}.internal:8080/healthz" > /dev/null 2>&1; then
            ((health++))
        fi
        sleep "$CHECK_INTERVAL"
    done

    # Return 0 if at least 2 out of 3 checks passed
    if [ $health -ge 2 ]; then
        log_debug "Region $region is healthy ($health/$FAILOVER_THRESHOLD checks passed)"
        return 0
    else
        log_debug "Region $region is unhealthy ($health/$FAILOVER_THRESHOLD checks passed)"
        return 1
    fi
}

# Check region health via Fly.io API
check_fly_region_health() {
    local region=$1

    if ! command -v flyctl &> /dev/null; then
        log_warn "flyctl not installed, using basic health check"
        check_region_health "$region"
        return $?
    fi

    # Get region status from Fly.io
    local status
    status=$(flyctl regions list --app "$FLY_APP" --json 2>/dev/null | \
        jq -r ".[] | select(.Region == \"$region\") | .Status" || echo "unknown")

    [ "$status" = "available" ]
}

# Update DNS to point to new primary
update_dns() {
    local new_primary=$1

    log_info "Updating DNS to point to region: $new_primary"

    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would update DNS to: $new_primary"
        return 0
    fi

    if [ -n "$CLOUDFLARE_ZONE_ID" ] && [ -n "$CLOUDFLARE_API_TOKEN" ]; then
        # Update Cloudflare DNS
        local record_name="api.functionfly.com"

        # Get the new IP or CNAME
        local new_value
        new_value=$(flyctl ips list --region "$new_primary" --app "$FLY_APP" 2>/dev/null | \
            jq -r '.[0].Address' || echo "${FLY_APP}.fly.dev")

        # Update the DNS record
        curl -s -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
            -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
            -H "Content-Type: application/json" \
            --data '{"type":"A","name":"'"$record_name"'","content":"'"$new_value"'","ttl":60}' || {
                log_error "Failed to update Cloudflare DNS"
                return 1
            }

        log_info "DNS updated successfully"
    else
        log_warn "Cloudflare not configured, skipping DNS update"
    fi
}

# Update Fly.io region weights
update_fly_regions() {
    local current_primary=$1
    local new_primary=$2

    log_info "Updating Fly.io region configuration"

    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would update Fly.io regions: $current_primary -> $new_primary"
        return 0
    fi

    if ! command -v flyctl &> /dev/null; then
        log_error "flyctl not installed, cannot update Fly.io regions"
        return 1
    fi

    # Set new primary region
    flyctl regions set "$new_primary" --app "$FLY_APP" || {
        log_error "Failed to set region: $new_primary"
        return 1
    }

    # Remove failed region from allowed list
    flyctl regions remove "$current_primary" --app "$FLY_APP" || {
        log_warn "Failed to remove region: $current_primary"
    }

    log_info "Fly.io regions updated successfully"
}

# Send alert notification
send_alert() {
    local alert_type=$1
    local from_region=$2
    local to_region=$3

    local payload
    payload=$(cat <<EOF
{
    "alert": "$alert_type",
    "from": "$from_region",
    "to": "$to_region",
    "timestamp": "$(date -Iseconds)",
    "app": "$FLY_APP"
}
EOF
)

    if [ -n "$ALERT_WEBHOOK_URL" ]; then
        log_info "Sending alert to webhook"
        curl -s -X POST "$ALERT_WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "$payload" || log_warn "Failed to send alert"
    fi

    if [ -n "$STATUS_PAGE_URL" ]; then
        log_info "Updating status page"
        curl -s -X POST "$STATUS_PAGE_URL" \
            -H "Content-Type: application/json" \
            -d "$payload" || log_warn "Failed to update status page"
    fi
}

# Verify failover
verify_failover() {
    local new_primary=$1

    log_info "Verifying failover to region: $new_primary"

    # Wait for the region to come up
    sleep 10

    # Check health of new primary
    if check_region_health "$new_primary"; then
        log_info "Failover verification successful"
        return 0
    else
        log_error "Failover verification failed"
        return 1
    fi
}

# Promote secondary region to primary
promote_region() {
    local current_primary=$1
    local new_primary=$2

    log_info "========================================="
    log_info "PROMOTING REGION: $new_primary -> PRIMARY"
    log_info "========================================="

    # Send pre-failover alert
    send_alert "failover_start" "$current_primary" "$new_primary"

    # Update DNS
    if ! update_dns "$new_primary"; then
        log_error "DNS update failed"
        send_alert "failover_dns_failed" "$current_primary" "$new_primary"
        return 1
    fi

    # Update Fly.io regions
    if ! update_fly_regions "$current_primary" "$new_primary"; then
        log_error "Fly.io region update failed"
        send_alert "failover_fly_failed" "$current_primary" "$new_primary"
        return 1
    fi

    # Verify failover
    if ! verify_failover "$new_primary"; then
        log_error "Failover verification failed"
        send_alert "failover_verify_failed" "$current_primary" "$new_primary"
        return 1
    fi

    # Send success alert
    send_alert "failover_complete" "$current_primary" "$new_primary"

    log_info "========================================="
    log_info "FAILOVER COMPLETED SUCCESSFULLY"
    log_info "New primary region: $new_primary"
    log_info "========================================="

    return 0
}

# Main failover logic
main() {
    local current_primary=$PRIMARY_REGION

    log_info "Starting failover check"
    log_info "Current primary region: $current_primary"
    log_info "Secondary region: $SECONDARY_REGION"
    log_info "Tertiary region: $TERTIARY_REGION"

    # Check primary health
    log_info "Checking health of primary region: $current_primary"

    if check_region_health "$current_primary"; then
        log_info "Primary region $current_primary is healthy - no failover needed"
        exit 0
    fi

    log_error "Primary region $current_primary is unhealthy, initiating failover"

    # Try secondary region
    log_info "Checking secondary region: $SECONDARY_REGION"
    if check_region_health "$SECONDARY_REGION"; then
        log_info "Secondary region $SECONDARY_REGION is healthy"
        promote_region "$current_primary" "$SECONDARY_REGION"
        exit $?
    fi

    # Try tertiary region
    log_info "Checking tertiary region: $TERTIARY_REGION"
    if check_region_health "$TERTIARY_REGION"; then
        log_info "Tertiary region $TERTIARY_REGION is healthy"
        promote_region "$current_primary" "$TERTIARY_REGION"
        exit $?
    fi

    # No healthy regions available
    log_error "No healthy regions available for failover"
    send_alert "failover_no_regions" "$current_primary" "none"
    exit 1
}

# Auto-failover mode (continuous monitoring)
auto_failover() {
    log_info "Starting auto-failover monitoring mode"
    log_info "Press Ctrl+C to stop"

    while true; do
        main
        sleep 30
# Rollback to previous primary
rollback() {
    local from_region=$1
    local to_region    done
}

=$2

    log_info "Rolling back from $from_region to $to_region"

    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would rollback: $from_region -> $to_region"
        return 0
    fi

    send_alert "rollback_start" "$from_region" "$to_region"

    update_dns "$to_region"
    update_fly_regions "$from_region" "$to_region"

    send_alert "rollback_complete" "$from_region" "$to_region"

    log_info "Rollback completed"
}

# Handle script arguments
case "${1:-check}" in
    check)
        main
        ;;
    auto)
        auto_failover
        ;;
    promote)
        promote_region "$2" "$3"
        ;;
    rollback)
        rollback "$2" "$3"
        ;;
    health)
        check_region_health "$2"
        ;;
    *)
        echo "Usage: $0 {check|auto|promote|rollback|health}"
        echo ""
        echo "Commands:"
        echo "  check                          Check primary health and failover if needed"
        echo "  auto                           Run continuous auto-failover monitoring"
        echo "  promote <from> <to>            Manually promote a region"
        echo "  rollback <from> <to>          Rollback to a previous primary"
        echo "  health <region>                Check health of a specific region"
        echo ""
        echo "Environment Variables:"
        echo "  PRIMARY_REGION                Primary region (default: iad)"
        echo "  SECONDARY_REGION              Secondary region (default: lax)"
        echo "  TERTIARY_REGION               Tertiary region (default: fra)"
        echo "  FAILOVER_THRESHOLD            Checks before failover (default: 3)"
        echo "  DRY_RUN                       Run without making changes (default: false)"
        exit 1
        ;;
esac
