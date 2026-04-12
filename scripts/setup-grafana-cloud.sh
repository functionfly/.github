#!/bin/bash
# Setup Grafana Cloud metrics forwarding for FunctionFly
# Your stack: https://grafana.com/orgs/functionfly/stacks/1583449

set -e

export PATH="$HOME/.fly/bin:$PATH"

echo "=== FunctionFly Grafana Cloud Setup ==="
echo ""
echo "Your Grafana Cloud stack: https://functionfly.grafana.net"
echo ""

# Check if user has the required credentials
if [ -z "$GRAFANA_CLOUD_URL" ] || [ -z "$GRAFANA_CLOUD_USER" ] || [ -z "$GRAFANA_CLOUD_API_KEY" ]; then
    echo "⚠️  Grafana Cloud credentials not found in environment"
    echo ""
    echo "Get your credentials from: https://functionfly.grafana.net/org/details"
    echo ""
    echo "Required environment variables:"
    echo "  - GRAFANA_CLOUD_URL: Prometheus remote_write URL"
    echo "    (e.g., https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom/push)"
    echo "  - GRAFANA_CLOUD_USER: Your instance ID (e.g., 1583449)"
    echo "  - GRAFANA_CLOUD_API_KEY: API key with 'MetricsPublisher' role"
    echo ""
    echo "Create API key at: https://functionfly.grafana.net/org/api-keys"
    echo ""
    echo "Then run:"
    echo "  export GRAFANA_CLOUD_URL='https://...'"
    echo "  export GRAFANA_CLOUD_USER='1583449'"
    echo "  export GRAFANA_CLOUD_API_KEY='your-key'"
    echo "  ./scripts/setup-grafana-cloud.sh"
    exit 1
fi

echo "✓ Credentials found"
echo ""

# Apps to configure
APPS=(
    "functionfly-orchestrator"
    "functionfly-control"
    "functionfly-ai-service"
)

# Set secrets for each app
echo "Setting Grafana Cloud secrets for FunctionFly apps..."
echo ""

for app in "${APPS[@]}"; do
    echo "→ Configuring $app..."

    # Check if app exists
    if ! fly apps list | grep -q "^${app}$"; then
        echo "  ⚠️  App $app not found, skipping..."
        continue
    fi

    # Set secrets
    fly secrets set \
        GRAFANA_CLOUD_URL="$GRAFANA_CLOUD_URL" \
        GRAFANA_CLOUD_USER="$GRAFANA_CLOUD_USER" \
        GRAFANA_CLOUD_API_KEY="$GRAFANA_CLOUD_API_KEY" \
        --app "$app" --yes 2>/dev/null || {
            echo "  ⚠️  Failed to set secrets for $app (may not be deployed yet)"
            continue
        }

    echo "  ✓ Secrets set for $app"
done

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Deploy the Grafana Cloud forwarder:"
echo "   fly deploy --config fly.grafana-cloud.toml"
echo ""
echo "2. Or update your existing Prometheus to use Grafana Cloud:"
echo "   fly deploy --config fly.monitoring.toml"
echo ""
echo "3. View your metrics at: https://functionfly.grafana.net/explore"
echo ""
echo "Pre-built dashboards to import:"
echo "  - Go Runtime: https://grafana.com/grafana/dashboards/10826"
echo "  - Node Exporter: https://grafana.com/grafana/dashboards/1860"
echo "  - PostgreSQL: https://grafana.com/grafana/dashboards/9628"
