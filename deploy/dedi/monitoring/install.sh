#!/bin/bash
# Install monitoring binaries on dedi server
# Run this ONCE on the server as the first step

set -e

echo "Installing Prometheus, Grafana, Loki, Promtail..."

# Create users
sudo useradd --system --no-create-home --shell=/bin/false prometheus || true
sudo useradd --system --no-create-home --shell=/bin/false grafana || true
sudo useradd --system --no-create-home --shell=/bin/false loki || true

# Create directories
sudo mkdir -p /etc/prometheus /etc/grafana /etc/loki /etc/promtail
sudo mkdir -p /var/lib/prometheus /var/lib/grafana /var/lib/loki /var/lib/promtail /var/log/grafana
sudo chmod -R 755 /etc/prometheus /etc/grafana /etc/loki /etc/promtail
sudo chmod -R 755 /var/lib/prometheus /var/lib/grafana /var/lib/loki /var/lib/promtail /var/log/grafana

# Download and install Prometheus
cd /tmp
curl -sLO https://github.com/prometheus/prometheus/releases/download/v2.52.0/prometheus-2.52.0.linux-amd64.tar.gz
tar -xzf prometheus-2.52.0.linux-amd64.tar.gz
sudo cp prometheus-2.52.0.linux-amd64/prometheus /usr/local/bin/
sudo cp prometheus-2.52.0.linux-amd64/promtool /usr/local/bin/
sudo cp -r prometheus-2.52.0.linux-amd64/console_libraries /etc/prometheus/
sudo cp -r prometheus-2.52.0.linux-amd64/consoles /etc/prometheus/
sudo chown prometheus:prometheus /usr/local/bin/prometheus /usr/local/bin/promtool
sudo chown -R prometheus:prometheus /etc/prometheus/console_libraries /etc/prometheus/consoles
rm -rf prometheus-2.52.0.linux-amd64*

# Download and install node_exporter
curl -sLO https://github.com/prometheus/node_exporter/releases/download/v1.8.0/node_exporter-1.8.0.linux-amd64.tar.gz
tar -xzf node_exporter-1.8.0.linux-amd64.tar.gz
sudo cp node_exporter-1.8.0.linux-amd64/node_exporter /usr/local/bin/
sudo chown prometheus:prometheus /usr/local/bin/node_exporter
rm -rf node_exporter-1.8.0.linux-amd64*

# Download and install Grafana
curl -sLO https://dl.grafana.com/enterprise/release/grafana-10.4.2.linux-amd64.tar.gz
tar -xzf grafana-10.4.2.linux-amd64.tar.gz
sudo cp grafana-10.4.2/bin/grafana /usr/local/bin/
sudo mkdir -p /usr/local/share/grafana
sudo cp -r grafana-10.4.2/public /usr/local/share/grafana/
sudo chown grafana:grafana /usr/local/bin/grafana
sudo chown -R grafana:grafana /usr/local/share/grafana
rm -rf grafana-10.4.2*

# Download and install Loki
curl -sLO https://github.com/grafana/loki/releases/download/v2.9.6/loki-linux-amd64.zip
sudo unzip -o loki-linux-amd64.zip -d /usr/local/bin/
sudo chmod +x /usr/local/bin/loki
sudo chown loki:loki /usr/local/bin/loki
rm loki-linux-amd64.zip

# Download and install Promtail
curl -sLO https://github.com/grafana/loki/releases/download/v2.9.6/promtail-linux-amd64.zip
sudo unzip -o promtail-linux-amd64.zip -d /usr/local/bin/
sudo chmod +x /usr/local/bin/promtail
sudo chown loki:loki /usr/local/bin/promtail
rm promtail-linux-amd64.zip

echo "Installation complete!"
