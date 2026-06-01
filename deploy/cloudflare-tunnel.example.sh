#!/bin/bash
# Cloudflare Tunnel – expose services via cloudflared
# Usage: ./deploy/cloudflare-tunnel.example.sh
#
# Prerequisites:
# 1. cloudflared installed (see deploy/cloudflare-tunnel.example.sh)
# 2. A Cloudflare Tunnel created in Cloudflare Zero Trust:
#    - Go to https://one.dash.cloudflare.com
#    - Networks > Tunnels > Create a Tunnel (select "Cloudflared")
#    - Copy the tunnel token
#
# This script exposes both the orchestrator API (port 8080) and dashboard (port 3000).
# Run in two terminals OR use --token flag: ./cloudflare-tunnel.example.sh --token <TOKEN>

set -e

API_PORT=${API_PORT:-8080}
DASHBOARD_PORT=${DASHBOARD_PORT:-3000}
TOKEN="${CLOUDFLARE_TUNNEL_TOKEN:-}"

usage() {
  echo "Usage: $0 --token <TUNNEL_TOKEN>"
  echo "  or:  CLOUDFLARE_TUNNEL_TOKEN=<TOKEN> $0"
  exit 1
}

if [ -z "$TOKEN" ]; then
  for arg in "$@"; do
    if [ "$arg" = "--token" ]; then
      TOKEN="${2:-}"
      break
    fi
  done
fi

if [ -z "$TOKEN" ]; then
  echo "Error: Tunnel token required. Set CLOUDFLARE_TUNNEL_TOKEN or use --token <TOKEN>"
  usage
fi

CLOUDFLARED="$(command -v cloudflared 2>/dev/null || command -v /home/micro/.local/bin/cloudflared 2>/dev/null)"

if [ -z "$CLOUDFLARED" ]; then
  echo "cloudflared not found. Install from: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/install-and-setup/installation/"
  exit 1
fi

echo "Starting Cloudflare Tunnel..."
echo "  API:       http://localhost:$API_PORT"
echo "  Dashboard: http://localhost:$DASHBOARD_PORT"
echo
echo "Configure in Cloudflare Zero Trust:"
echo "  - Networks > Tunnels > your tunnel > Public Hostname"
echo "  - Add: api.localhost → http://localhost:$API_PORT"
echo "  - Add: dashboard.localhost → http://localhost:$DASHBOARD_PORT"
echo

exec "$CLOUDFLARED" tunnel run --token "$TOKEN"
