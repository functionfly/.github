#!/bin/bash
# Switch existing Prometheus to Grafana Cloud remote_write
# This updates your current fly.monitoring.toml deployment

set -e

export PATH="$HOME/.fly/bin:$PATH"

echo "=== Switching to Grafana Cloud ==="
echo ""
echo "Stack: https://functionfly.grafana.net"
echo ""

# Check credentials
if [ -z "$GRAFANA_CLOUD_URL" ] || [ -z "$GRAFANA_CLOUD_USER" ] || [ -z "$GRAFANA_CLOUD_API_KEY" ]; then
    echo "❌ Missing Grafana Cloud credentials"
    echo ""
    echo "Get these from: https://functionfly.grafana.net/org/details"
    echo ""
    echo "Then run:"
    echo "  export GRAFANA_CLOUD_URL='https://prometheus-prod-XX-prod-us-east-0.grafana.net/api/prom/push'"
    echo "  export GRAFANA_CLOUD_USER='1583449'"
    echo "  export GRAFANA_CLOUD_API_KEY='YOUR_API_KEY'"
    echo "  ./scripts/deploy-grafana-cloud.sh"
    exit 1
fi

echo "✓ Credentials found"
echo ""

# Update fly.monitoring.toml to use Grafana Cloud Dockerfile
sed -i 's|dockerfile = "deploy/monitoring/Dockerfile.prometheus"|dockerfile = "deploy/monitoring/Dockerfile.grafana-cloud"|' fly.monitoring.toml

# Add secrets to prometheus app
echo "Setting secrets for functionfly-prometheus..."
fly secrets set \
    GRAFANA_CLOUD_URL="$GRAFANA_CLOUD_URL" \
    GRAFANA_CLOUD_USER="$GRAFANA_CLOUD_USER" \
    GRAFANA_CLOUD_API_KEY="$GRAFANA_CLOUD_API_KEY" \
    --app functionfly-prometheus --yes

echo ""
echo "Re-deploying Prometheus with Grafana Cloud..."
fly deploy --config fly.monitoring.toml --yes

echo ""
echo "=== Done! ==="
echo ""
echo "Metrics are now flowing to: https://functionfly.grafana.net/explore"
echo ""
echo "Import dashboards:"
echo "  1. Go to https://functionfly.grafana.net/dashboard/import"
echo "  2. Use ID: 10826 (Go Runtime)"
echo "  3. Use ID: 3662 (Prometheus Overview)"
