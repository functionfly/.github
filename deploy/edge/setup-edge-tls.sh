#!/bin/bash
# Run this ON each edge VPS (after certs are in /etc/ssl/functionfly/).
# Installs Caddy and configures TLS termination in front of functionfly-edge on 8080.
# Usage: copy this file and Caddyfile.edge to the server, then run as root:
#   scp deploy/edge/setup-edge-tls.sh deploy/edge/Caddyfile.edge root@VPS_IP:/tmp/
#   ssh root@VPS_IP 'bash /tmp/setup-edge-tls.sh'
set -e

if [ ! -f /etc/ssl/functionfly/fullchain.pem ] || [ ! -f /etc/ssl/functionfly/privkey.pem ]; then
  echo "Upload certs first: run ./upload-certs.sh from your machine"
  exit 1
fi

# Install Caddy (official script, no plugins - we only need TLS + reverse_proxy)
if ! command -v caddy &>/dev/null; then
  echo "Installing Caddy..."
  apt-get update -qq && apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
  apt-get update -qq && apt-get install -y caddy
fi

# Deploy Caddyfile (assumes Caddyfile.edge is in same dir as this script or in /tmp)
CADDYFILE_SRC="/tmp/Caddyfile.edge"
if [ ! -f "$CADDYFILE_SRC" ]; then
  echo "Copy Caddyfile.edge to this server (e.g. /tmp/Caddyfile.edge) and run again"
  exit 1
fi
cp "$CADDYFILE_SRC" /etc/caddy/Caddyfile

# Caddy listens on 80/443; ensure functionfly-edge is on 8080
systemctl enable caddy
systemctl restart caddy
echo "Caddy restarted. Check: systemctl status caddy && curl -sI https://edge.functionfly.com/healthz"
