#!/bin/bash
# Run this ON the new LB VPS (74.208.108.203).
# Installs Caddy and configures load balancing in front of the two edge VPS origin nodes.
# Usage:
#   scp deploy/edge/setup-lb-tls.sh deploy/edge/Caddyfile.lb root@74.208.108.203:/tmp/
#   ssh root@74.208.108.203 'bash /tmp/setup-lb-tls.sh'
set -e

if [ ! -f /etc/ssl/functionfly/fullchain.pem ] || [ ! -f /etc/ssl/functionfly/privkey.pem ]; then
  echo "Upload certs first: run ./upload-certs.sh from your machine"
  exit 1
fi

# Install Caddy (official script)
if ! command -v caddy &>/dev/null; then
  echo "Installing Caddy..."
  apt-get update -qq && apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
  apt-get update -qq && apt-get install -y caddy
fi

# Deploy Caddyfile
CADDYFILE_SRC="/tmp/Caddyfile.lb"
if [ ! -f "$CADDYFILE_SRC" ]; then
  echo "Copy Caddyfile.lb to this server (e.g. /tmp/Caddyfile.lb) and run again"
  exit 1
fi
cp "$CADDYFILE_SRC" /etc/caddy/Caddyfile

# Create log dir
mkdir -p /var/log/caddy

# Enable and restart
systemctl enable caddy
systemctl restart caddy
echo "Caddy restarted. Check: systemctl status caddy"
