# Grafana Cloud Setup for FunctionFly
# This file helps configure remote_write to Grafana Cloud
# Sign up at https://grafana.com/products/cloud/

# After signup, you'll get:
# - PROMETHEUS_URL: https://prometheus-prod-XX-prod-us-east-0.grafana.net/api/prom/push
# - PROMETHEUS_USERNAME: Your Grafana Cloud instance ID (e.g., 123456)
# - PROMETHEUS_PASSWORD: A Grafana Cloud API key with 'MetricsPublisher' role

## Setup Steps:

1. **Sign up for Grafana Cloud** (free tier):
   https://grafana.com/auth/sign-up/create-org

2. **Get your Prometheus URL and credentials** from:
   https://<your-org>.grafana.net/org/details

3. **Create an API Key**:
   - Go to https://<your-org>.grafana.net/org/api-keys
   - Create key with role: "MetricsPublisher"

4. **Set Fly.io secrets**:
   ```bash
   export PATH="$HOME/.fly/bin:$PATH"
   
   # For orchestrator app
   fly secrets set GRAFANA_CLOUD_URL="https://prometheus-prod-XX-prod-us-east-0.grafana.net/api/prom/push" \
                   GRAFANA_CLOUD_USER="YOUR_INSTANCE_ID" \
                   GRAFANA_CLOUD_API_KEY="YOUR_API_KEY" \
                   --app functionfly-orchestrator
   
   # For control app
   fly secrets set GRAFANA_CLOUD_URL="https://prometheus-prod-XX-prod-us-east-0.grafana.net/api/prom/push" \
                   GRAFANA_CLOUD_USER="YOUR_INSTANCE_ID" \
                   GRAFANA_CLOUD_API_KEY="YOUR_API_KEY" \
                   --app functionfly-control
   ```

5. **Enable remote_write** in your apps (see prometheus.fly.yml)

## Alternative: Grafana Agent (lightweight)

Instead of running full Prometheus, you can run Grafana Agent on Fly.io:

```yaml
# fly.agent.toml
app = "functionfly-grafana-agent"
[build]
  dockerfile = "deploy/monitoring/Dockerfile.grafana-agent"
```

## Cost Comparison

| Plan | Active Series | Retention | Cost |
|------|---------------|-----------|------|
| **Free** | 10,000 | 14 days | $0 |
| **Pro** | 100,000 | 13 months | ~$8/10k series |

Your current setup: ~500-1000 series (well within free tier)

## Pre-built Dashboards

After setup, import these dashboards in Grafana Cloud:
1. Go Explore → Import Dashboard
2. Use ID: `3662` (Prometheus 2.0 Overview)
3. Use ID: `10826` (Go Runtime Metrics)
4. Use ID: `9628` (PostgreSQL Database)

## URLs

- Grafana Cloud Portal: https://<your-org>.grafana.net
- Metrics explore: https://<your-org>.grafana.net/explore
- Alerting: https://<your-org>.grafana.net/alerting
