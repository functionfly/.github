#!/bin/bash
# FunctionFly Edge VPS Certificate Renewal Cron Script
# This script should be run daily via cron to check and renew certificates
# Usage: Add to crontab: 0 2 * * * /opt/functionfly/cert-renewal-cron.sh >> /var/log/cert-renewal.log 2>&1
#
# Prerequisites:
# 1. Certs must be available at /etc/ssl/functionfly/ (fullchain.pem, privkey.pem)
# 2. SSH access to both edge VPS nodes configured (217.160.124.206, 209.46.125.113)
# 3. Caddy must be installed on each node

set -e

LOG_FILE="/var/log/cert-renewal.log"
EDGE_NODES=("217.160.124.206" "209.46.125.113")
CERT_DIR="/etc/ssl/functionfly"
SSH_USER="${SSH_USER:-root}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_cert_expiry() {
    local cert_file="$1"
    if [ ! -f "$cert_file" ]; then
        return 1
    fi

    # Get expiry date (works with openssl)
    local expiry_date=$(openssl x509 -in "$cert_file" -enddate -noout 2>/dev/null | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry_date" +%s 2>/dev/null || echo 0)
    local now_epoch=$(date +%s)
    local days_until_expiry=$(( (expiry_epoch - now_epoch) / 86400 ))

    echo "$days_until_expiry"
}

renew_certs() {
    log "Starting certificate renewal process..."

    # Check local cert first
    if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
        log "ERROR: Local certificate not found at $CERT_DIR/fullchain.pem"
        return 1
    fi

    local days_left=$(check_cert_expiry "$CERT_DIR/fullchain.pem")
    log "Certificate expires in $days_left days"

    # Only renew if within 30 days of expiry
    if [ "$days_left" -gt 30 ]; then
        log "Certificate still valid (>30 days), no renewal needed"
        return 0
    fi

    log "Certificate needs renewal (within 30 days of expiry)"

    # In production, you would:
    # 1. Use certbot with DNS challenge for Let's Encrypt wildcard certs
    # 2. Or download from your CA's portal
    # 3. Then upload to edge nodes

    # For automated Let's Encrypt renewal, uncomment below:
    # certbot certonly --manual --preferred-challenges=dns -d "*.functionfly.com" --renew-by-default

    # Upload to each edge node and restart Caddy
    for node in "${EDGE_NODES[@]}"; do
        log "Uploading certificates to $node..."

        # Upload fullchain and privkey
        scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
            "$CERT_DIR/fullchain.pem" "$CERT_DIR/privkey.pem" \
            "${SSH_USER}@${node}:/etc/ssl/functionfly/"

        # Set correct permissions
        ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
            "${SSH_USER}@${node}" \
            "chmod 644 /etc/ssl/functionfly/fullchain.pem && chmod 600 /etc/ssl/functionfly/privkey.pem"

        # Restart Caddy to reload certificates
        log "Restarting Caddy on $node..."
        ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
            "${SSH_USER}@${node}" \
            "systemctl restart caddy"

        log "Certificate renewal completed on $node"
    done

    log "Certificate renewal process completed successfully"
}

# Main execution
log "=== Certificate Renewal Check Started ==="
renew_certs
log "=== Certificate Renewal Check Finished ==="
