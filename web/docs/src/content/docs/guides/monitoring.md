---
title: Monitoring & Observability
description: Monitor your functions, agents, and infrastructure with FunctionFly's built-in observability tools.
sidebar:
  order: 15
---



This guide covers how to monitor your FunctionFly applications using built-in analytics, logs, and alerting.

## Overview

FunctionFly provides built-in observability across multiple dimensions:

| Capability | What It Tracks |
|------------|---------------|
| **Analytics** | Requests, latency, errors, geographic distribution |
| **Logs** | Function execution logs, system events |
| **Traces** | Request flow through functions and services |
| **Alerts** | Notifications when metrics exceed thresholds |
| **Status Page** | Real-time and historical uptime |

---

## Analytics Dashboard

### Accessing Analytics

1. Go to **Analytics** in the dashboard sidebar
2. Or click **Analytics** on any function detail page

### Available Metrics

#### Request Metrics

| Metric | Description |
|--------|-------------|
| **Total Requests** | Number of function invocations |
| **Request Rate** | Requests per minute/hour/day |
| **Unique Invokers** | Distinct sources calling your functions |
| **Error Rate** | Percentage of requests returning errors |

#### Performance Metrics

| Metric | Description |
|--------|-------------|
| **Latency p50** | Median response time |
| **Latency p95** | 95th percentile response time |
| **Latency p99** | 99th percentile response time |
| **Cold Start Rate** | Percentage of requests experiencing cold starts |

#### Resource Metrics

| Metric | Description |
|--------|-------------|
| **Memory Usage** | Average and peak memory consumption |
| **CPU Usage** | Average CPU time per execution |
| **Network I/O** | Data transferred in/out |

### Filtering Analytics

Filter by:
- **Time range** — Last hour, 24 hours, 7 days, 30 days, custom
- **Function** — Specific function or all functions
- **Region** — Specific edge location or global
- **Runtime** — Python, Node.js, Go, etc.

---

## Logs

### Accessing Logs

**Dashboard:**
1. Go to **Logs** in the sidebar
2. Select function and time range
3. Browse and search logs

**CLI:**
```bash
# Tail logs for a function
ffly logs my-function

# View last 100 lines
ffly logs my-function --lines 100

# Filter logs
ffly logs my-function --filter "ERROR"
```

### Log Structure

Each log entry includes:

```json
{
    "timestamp": "2026-05-08T10:30:00.000Z",
    "level": "INFO",
    "request_id": "req_abc123xyz",
    "function": "my-function",
    "version": "1.0.0",
    "region": "us-east-1",
    "message": "Function executed successfully",
    "duration_ms": 45,
    "memory_mb": 128,
    "metadata": {}
}
```

### Log Levels

| Level | Use Case |
|-------|----------|
| `DEBUG` | Detailed debugging information |
| `INFO` | Normal operation events |
| `WARN` | Potential issues, degraded performance |
| `ERROR` | Failed operations, exceptions |
| `FATAL` | Critical failures, system crashes |

### Structured Logging

Emit structured logs from your functions:

```python
import json
import time

def handler(request):
    start = time.time()
    
    # Your logic here
    result = process(request)
    
    # Structured log
    print(json.dumps({
        "event": "request_processed",
        "duration_ms": int((time.time() - start) * 1000),
        "status": "success",
        "request_id": request.get("request_id")
    }))
    
    return result
```

```javascript
export default async function handler(request) {
    const start = Date.now();
    
    // Your logic here
    const result = await process(request);
    
    // Structured log
    console.log(JSON.stringify({
        event: "request_processed",
        duration_ms: Date.now() - start,
        status: "success",
        request_id: request.request_id
    }));
    
    return result;
}
```

---

## Distributed Tracing

### How Tracing Works

When a request flows through multiple functions:

```
User Request
    │
    ▼
api-gateway
    │
    ├──▶ auth-service (span: 12ms)
    │
    ├──▶ user-service (span: 45ms)
    │       │
    │       ├──▶ database (span: 8ms)
    │       │
    │       └──▶ cache (span: 2ms)
    │
    └──▶ response (total: 78ms)
```

### Trace Context Propagation

Traces automatically propagate via headers:

| Header | Description |
|--------|-------------|
| `X-Trace-ID` | Unique trace identifier |
| `X-Span-ID` | Current span identifier |
| `X-Parent-Span-ID` | Parent span identifier |

### Enabling Tracing

1. Go to **Settings → Observability**
2. Enable **Distributed Tracing**
3. Choose sampling rate:
   - 100% (all requests) — Enterprise
   - 10% (sample) — Professional
   - 1% (sample) — Starter

### Viewing Traces

1. Go to **Analytics → Traces**
2. Click on a trace to see:
   - Waterfall chart of spans
   - Each service involved
   - Duration breakdown
   - Error locations

---

## Alerts

### Creating Alerts

1. Go to **Settings → Alerts**
2. Click **Create Alert**
3. Configure:

| Setting | Description |
|---------|-------------|
| **Name** | Descriptive alert name |
| **Metric** | What to monitor (error rate, latency, etc.) |
| **Condition** | Threshold (e.g., > 5%) |
| **Duration** | How long condition must persist |
| **Severity** | Info, Warning, Critical |

### Alert Conditions

| Metric | Operators | Example |
|--------|-----------|---------|
| Error Rate | `>`, `<`, `>=`, `<=` | Error rate > 5% |
| Latency p99 | `>`, `<`, `>=`, `<=` | Latency > 1000ms |
| Request Count | `>`, `<`, `>=`, `<=` | Requests < 10/min |
| Memory Usage | `>`, `<`, `>=`, `<=` | Memory > 512MB |
| Custom | Any metric combination | Custom formula |

### Alert Notifications

Configure how to receive alerts:

| Channel | Setup Required |
|---------|----------------|
| **In-App** | None (always enabled) |
| **Email** | Email address |
| **Slack** | Slack webhook URL |
| **Discord** | Discord webhook URL |
| **PagerDuty** | PagerDuty integration key |
| **Webhook** | HTTP webhook URL |

### Alert Examples

**High Error Rate:**
```
Alert: High Error Rate
Function: payment-processor
Condition: Error Rate > 5% for 5 minutes
Severity: Critical
Notify: Slack #alerts, Email ops@yourcompany.com
```

**Slow Response:**
```
Alert: Slow Response Time
Function: image-resizer
Condition: Latency p99 > 3000ms for 10 minutes
Severity: Warning
Notify: Slack #alerts
```

**Memory Leak:**
```
Alert: Memory Usage High
Function: background-worker
Condition: Memory > 800MB for 15 minutes
Severity: Critical
Notify: PagerDuty
```

---

## Uptime Monitoring

### Public Status Page

FunctionFly maintains a public status page at **status.functionfly.com**

This shows:
- Current system status
- Active incidents
- Historical uptime (last 90 days)
- Scheduled maintenance

### Your Functions' Status

Monitor your specific functions:

1. Go to **Settings → Monitoring → Uptime**
2. Add functions to monitor
3. Set check frequency (every 1, 5, 15, 30 minutes)
4. Configure failure response (notify, auto-restart)

### Uptime Checks

Each check verifies:
- Function is reachable
- Response time under threshold
- Returns expected response (optional)

---

## Metrics API

Access metrics programmatically:

```bash
# Get function metrics
curl -H "Authorization: Bearer $FFLY_TOKEN" \
  "https://api.functionfly.com/v1/analytics/functions/my-function/metrics?period=24h"

# Response
{
    "function": "my-function",
    "period": {
        "start": "2026-05-07T10:00:00Z",
        "end": "2026-05-08T10:00:00Z"
    },
    "metrics": {
        "requests": 125000,
        "errors": 125,
        "error_rate": 0.001,
        "latency_p50_ms": 23,
        "latency_p95_ms": 67,
        "latency_p99_ms": 145,
        "avg_memory_mb": 64,
        "region_distribution": {
            "us-east-1": 45000,
            "eu-west-1": 38000,
            "ap-southeast-1": 42000
        }
    }
}
```

### Exporting Metrics

Export to external monitoring tools:

**Datadog:**
```bash
ffly metrics export --format datadog --api-key your-datadog-key
```

**Prometheus:**
```bash
ffly metrics export --format prometheus --port 9090
```

**Grafana:**
Configure Prometheus as datasource, then use FunctionFly metrics endpoint.

---

## Health Checks

### Function Health

Check if a specific function is healthy:

```bash
curl https://api.functionfly.com/v1/functions/my-function/health
# Response: { "status": "healthy", "latency_ms": 12 }
```

### System Health

Check overall platform health:

```bash
curl https://api.functionfly.com/health
# Response: { "status": "operational", "services": {...} }
```

---

## Retention & Storage

| Data Type | Free | Starter | Professional | Enterprise |
|-----------|------|---------|--------------|------------|
| Metrics | 24 hours | 7 days | 30 days | 1 year |
| Logs | 24 hours | 7 days | 90 days | 1 year+ |
| Traces | — | 1 day | 7 days | 30 days |
| Alert History | 7 days | 30 days | 90 days | 1 year |

---

## Best Practices

1. **Set baseline alerts** — Error rate, latency p99, memory usage
2. **Use structured logging** — Include request IDs for correlation
3. **Sample traces in production** — 100% tracing can be expensive
4. **Review metrics daily** — Catch issues before they become outages
5. **Correlate logs and traces** — Use request ID to connect logs to traces
6. **Test alerting** — Regularly verify alert notifications work
7. **Document alert runbooks** — So team members know how to respond
