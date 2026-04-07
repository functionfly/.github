# Datadog Evaluation Setup for FunctionFly

This directory contains the configuration to evaluate **Datadog** as a drop-in addition to the existing Prometheus/Grafana/Loki stack. No existing monitoring infrastructure is removed.

---

## What this sets up

1. **Prometheus `remote_write`** — Every metric Prometheus scrapes is simultaneously forwarded to Datadog via the `remote_write` config in `prometheus-remote-write.yml`.
2. **Fly.io Log Drains** — All logs from `functionfly-api` and `functionfly-ai-service` are streamed to Datadog Logs via Fly.io's native log drain feature.

---

## Prerequisites

- A Datadog account (free 14-day trial at [datadoghq.com](https://www.datadoghq.com))
- Your Datadog API key: **Settings → Organization Settings → API Keys**
- Fly.io CLI authenticated: `fly auth login`

---

## Step 1: Add `remote_write` to Prometheus

### 1a. Copy the config file to your Prometheus host

```bash
# On the host running Prometheus:
sudo cp deploy/datadog/prometheus-remote-write.yml /etc/prometheus/remote_write.yml
sudo chmod 644 /etc/prometheus/remote_write.yml
```

### 1b. Create the API key secret file

```bash
# Create the secrets directory if it doesn't exist
sudo mkdir -p /etc/secrets

# Write your API key (no trailing newline — echo -n preserves this)
echo -n "your-dd-api-key" | sudo tee /etc/secrets/datadog_api_key
sudo chmod 600 /etc/secrets/datadog_api_key
```

### 1c. Add the include to `prometheus.yml`

Add the following line to the `rule_files` section of your Prometheus config (e.g. `deploy/production/monitoring/prometheus.yml`):

```yaml
rule_files:
  - "alert_rules.yml"
  - "/etc/prometheus/alerts/*.yml"
  - "/etc/prometheus/remote_write.yml"   # <-- add this line
```

Alternatively, you can copy the `remote_write` block directly into your `prometheus.yml` — see the inline instructions in the file comments.

### 1d. Reload Prometheus

```bash
# If Prometheus has the --web.enable-lifecycle flag (recommended):
curl -X POST http://localhost:9090/-/reload

# Or restart the service:
sudo systemctl restart prometheus
```

Verify Prometheus started without errors:

```bash
curl -s http://localhost:9090/status/runtimeinfo | jq .version
curl -s http://localhost:9090/api/v1/status/tsdb | jq .data
```

---

## Step 2: Add Fly.io Log Drains

Log drains stream **all** output from a Fly.io app to Datadog. No application code changes are needed.

### 2a. Add the log drain to `functionfly-api`

```bash
fly log drain add https://http-intake.logs.datadoghq.com/api/v2/logs \
  --header "DD-API-KEY: ${DD_API_KEY}" \
  --app functionfly-api
```

### 2b. Add the log drain to `functionfly-ai-service`

```bash
fly log drain add https://http-intake.logs.datadoghq.com/api/v2/logs \
  --header "DD-API-KEY: ${DD_API_KEY}" \
  --app functionfly-ai-service
```

### 2c. Verify the drains are attached

```bash
fly log drain list --app functionfly-api
fly log drain list --app functionfly-ai-service
```

Expected output shows a drain with `url: https://http-intake.logs.datadoghq.com/api/v2/logs` and `status: active`.

---

## Step 3: Verify in Datadog

### Metrics

1. Open **Metrics → Explorer** in the Datadog sidebar.
2. In the "Enter a metric name" box, type `functionfly`. Datadog normalizes Prometheus metric names, so `functionfly_http_requests_total` becomes `functionfly.http.requests.total`.
3. Select any metric — it should appear within **30–60 seconds** of the next Prometheus scrape cycle.

### Logs

1. Open **Logs → Explorer** in the Datadog sidebar.
2. Filter by `app:functionfly-api`. Log lines from `fly logs --app functionfly-api` should appear within **30 seconds**.
3. Logs are tagged with:
   - `app:functionfly-api` (or `functionfly-ai-service`)
   - `platform:fly-io`
   - `fly_region:<region>` (e.g. `fly_region:ord`)

### Dashboards (optional)

Create a new Datadog dashboard to mirror your Grafana panels:

1. **New Dashboard → Add Widget**
2. Choose **Timeseries** or **Hostmap**
3. In the metric picker, search `functionfly`
4. Group by labels that match your Prometheus labels (e.g. `job`, `endpoint`, `method`)

---

## Cleanup (after eval)

To remove Datadog and return to the pure Prometheus stack:

### Prometheus

```bash
# 1. Remove the rule_files entry from prometheus.yml
#    (remove the line: - "/etc/prometheus/remote_write.yml")

# 2. Reload Prometheus
curl -X POST http://localhost:9090/-/reload

# 3. Remove the config and secret files
sudo rm /etc/prometheus/remote_write.yml
sudo rm /etc/secrets/datadog_api_key
```

### Fly.io Log Drains

```bash
# List drains to get the drain ID
fly log drain list --app functionfly-api

# Remove each drain by ID
fly log drain remove <drain-id> --app functionfly-api
fly log drain remove <drain-id> --app functionfly-ai-service
```

---

## Troubleshooting

### Metrics not appearing in Datadog

```bash
# Check Prometheus is writing to remote_write successfully
curl -s http://localhost:9090/api/v1/query?query=up | jq .

# Check the remote_write queue (Prometheus 2.39+):
curl -s http://localhost:9090/api/v1/query?query=prometheus_remote_write_queue_length | jq .
```

If `prometheus_remote_write_queue_length` is non-zero and growing, the remote_write is failing to send. Check:

- The API key file exists at `/etc/secrets/datadog_api_key` and has no trailing newline
- The API key is valid (regenerate at [Datadog API Settings](https://app.datadoghq.com/organization-settings/api-keys))
- Network connectivity from the Prometheus host to `api.datadoghq.com:443`

### Logs not appearing in Datadog

```bash
# Verify the drain is active
fly log drain list --app functionfly-api

# Check Fly.io status (log drains may be temporarily disabled during incidents)
fly doctor
```

Logs can take up to 30 seconds to appear after the drain is attached. If no logs appear after 2 minutes, verify the `DD_API_KEY` header value is correct (the key must be passed as a header value, not in the URL).

### Datadog metric naming

Prometheus metric names like `functionfly_http_requests_total` are **normalized** by Datadog's Prometheus remote_write integration:

- Dots replace underscores: `functionfly.http.requests.total`
- Labels are preserved exactly as-is

Use the Datadog Metrics Explorer to discover the normalized names — searching `functionfly` will surface all forwarded metrics.
