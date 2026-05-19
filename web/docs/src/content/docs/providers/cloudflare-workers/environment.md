---
title: Cloudflare Workers Environment
description: Environment variables specific to Cloudflare Workers deployment.
---

Deploy FunctionFly functions to Cloudflare's global network.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `CF_ID` | Cloudflare account ID |
| `CF_REGION` | Region (auto-populated by Cloudflare) |
| `CF_COLO` | Colocation code |
| `CF_DATACENTER` | Data center identifier |
| `CLOUDFLARE_WORKER` | Set to `1` when running in Workers |
| `CLOUDFLARE_PAGES` | Set to `1` when running on Pages |

## Request Context

| Variable | Description |
|----------|-------------|
| `CF_RAY` | Cloudflare ray ID |
| `CF_REQUEST_ID` | Unique request ID |
| `CF_WEBHOOK_SOURCE_IP` | Original visitor IP |

## Workers KV Integration

| Variable | Description |
|----------|-------------|
| `KV_CACHE_TTL` | Cache TTL for KV reads (seconds) |
| `KV_NAMESPACE_ID` | Workers KV namespace ID |
| `KV_DO_PUBLISH` | Enable automatic KV publishing |

## Durable Objects

| Variable | Description |
|----------|-------------|
| `DO_CLASS` | Durable Object class name |
| `DO_ID` | Durable Object instance ID |
| `DO_PERSISTENCE` | Enable durable object persistence |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `WORKER_MEMORY_LIMIT_MB` | `128` | Memory limit in MB |
| `WORKER_CPU_TIME_MS` | `50` | CPU time limit in ms |
| `WORKER_MAX_TICK` | `10` | Maximum ticks per request |

## Secrets

Store secrets in Workers:

```javascript
// wrangler.toml
[vars]
API_KEY = "your-secret-value"
```

Or use Cloudflare's secret management:

```bash
wrangler secret put API_KEY
```

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "workers",
  "environment": {
    "WORKER_MEMORY_LIMIT_MB": "128",
    "KV_CACHE_TTL": "60"
  }
}
```

## Cold Start

Cloudflare Workers have ultra-fast cold starts:

| Metric | Value |
|--------|-------|
| Cold start | < 5ms |
| Memory | Up to 128MB |
| CPU | 50ms compute per request |