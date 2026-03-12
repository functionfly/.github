#!/bin/bash
# Upload STAR_functionfly_com certs to edge VPS nodes.
# Prerequisites:
#   1. Run prepare-certs.sh so ./certs-out/ has fullchain.pem (and privkey.pem unless keys on server)
#   2. SSH access to both nodes (key-based or password)
# Usage:
#   ./upload-certs.sh [SSH user, default: root]
#   KEYS_ON_SERVER=1 ./upload-certs.sh [SSH user]   # Only upload fullchain; privkey already on each VPS
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_OUT="${SCRIPT_DIR}/certs-out"
NODES=( "217.160.124.206" "209.46.125.113" )
SSH_USER="${1:-root}"
REMOTE_DIR="/etc/ssl/functionfly"
KEYS_ON_SERVER="${KEYS_ON_SERVER:-0}"

if [ ! -f "${CERTS_OUT}/fullchain.pem" ]; then
  echo "Missing ${CERTS_OUT}/fullchain.pem. Run prepare-certs.sh first."
  exit 1
fi
if [ "$KEYS_ON_SERVER" != "1" ] && [ ! -f "${CERTS_OUT}/privkey.pem" ]; then
  echo "Missing ${CERTS_OUT}/privkey.pem. Run prepare-certs.sh, or set KEYS_ON_SERVER=1 if keys are already on each VPS."
  exit 1
fi

echo "Uploading certs to edge VPS (user=${SSH_USER}, dir=${REMOTE_DIR})..."
for node in "${NODES[@]}"; do
  echo "  → ${node}"
  ssh "${SSH_USER}@${node}" "mkdir -p ${REMOTE_DIR}"
  scp "${CERTS_OUT}/fullchain.pem" "${SSH_USER}@${node}:${REMOTE_DIR}/"
  ssh "${SSH_USER}@${node}" "chmod 644 ${REMOTE_DIR}/fullchain.pem"
  if [ "$KEYS_ON_SERVER" != "1" ]; then
    scp "${CERTS_OUT}/privkey.pem" "${SSH_USER}@${node}:${REMOTE_DIR}/"
    ssh "${SSH_USER}@${node}" "chmod 600 ${REMOTE_DIR}/privkey.pem"
  fi
done

echo "Done. Restart Caddy on each node to pick up certs: sudo systemctl restart caddy"
