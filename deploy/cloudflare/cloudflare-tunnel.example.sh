#!/usr/bin/env bash
# Cloudflare Tunnel (cloudflared) – example for exposing the orchestrator API
# without opening server ports. See docs/CLOUDFLARE.md.
#
# 1. Create a Tunnel in Cloudflare Zero Trust → Tunnels (or via API).
# 2. Add a public hostname, e.g. api.functionfly.com → http://localhost:8080.
# 3. Copy the tunnel token and set CLOUDFLARE_TUNNEL_TOKEN (or pass below).
# 4. Run: ./deploy/cloudflare/cloudflare-tunnel.example.sh

set -euo pipefail

TUNNEL_TOKEN="${CLOUDFLARE_TUNNEL_TOKEN:-}"

if [ -z "$TUNNEL_TOKEN" ]; then
  echo "Set CLOUDFLARE_TUNNEL_TOKEN to the token from Cloudflare dashboard (Tunnel → Configure)."
  exit 1
fi

# Ensure cloudflared is installed (e.g. https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/)
if ! command -v cloudflared &>/dev/null; then
  echo "Install cloudflared first: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/"
  exit 1
fi

exec cloudflared tunnel run --token "$TUNNEL_TOKEN"
