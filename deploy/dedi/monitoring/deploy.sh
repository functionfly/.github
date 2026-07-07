#!/bin/bash
# Deploy monitoring stack to dedi server
# Usage: ./deploy.sh user@host

set -e
SERVER="${1:-}"

if [[ -z "$SERVER" ]]; then
    echo "Usage: $0 user@host"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Deploying monitoring stack to $SERVER..."

scp -r "$SCRIPT_DIR"/*.yml "$SCRIPT_DIR"/*.ini "$SCRIPT_DIR"/*.service "$SERVER:/tmp/monitoring/"

ssh -o StrictHostKeyChecking=no "$SERVER" "sudo bash -c '
    cp /tmp/monitoring/*.yml /etc/prometheus/ 2>/dev/null || true
    cp /tmp/monitoring/loki.yml /etc/loki/ 2>/dev/null || true
    cp /tmp/monitoring/promtail.yml /etc/promtail/ 2>/dev/null || true
    cp /tmp/monitoring/*.ini /etc/grafana/ 2>/dev/null || true
    cp /tmp/monitoring/*.service /etc/systemd/system/

    mkdir -p /var/lib/prometheus /var/lib/grafana /var/lib/loki /var/lib/promtail /var/log/grafana
    chown -R prometheus:prometheus /var/lib/prometheus /etc/prometheus
    chown -R grafana:grafana /var/lib/grafana /etc/grafana /var/log/grafana
    chown -R loki:loki /var/lib/loki /etc/loki /var/lib/promtail

    systemctl daemon-reload
    systemctl enable prometheus loki promtail grafana node_exporter 2>/dev/null || true
    systemctl start prometheus loki promtail node_exporter
'"

echo "Done! Access Grafana at http://\$SERVER:3000"
