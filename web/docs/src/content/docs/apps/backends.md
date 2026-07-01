---
title: Backends
description: Configure multi-cloud deploy targets for your app
sidebar:
  order: 2
---

# Backends

A **backend** is a deploy target attached to an app. Each backend points to a
specific cloud provider and region. An app can have multiple backends for
multi-cloud redundancy, geographic distribution, or blue/green deployments.

## Supported Providers

| Provider | Slug | Description |
|----------|------|-------------|
| FunctionFly Edge | `functionfly-edge` | FunctionFly's managed edge network |
| Cloudflare Workers | `workers` | Cloudflare Workers |
| Vercel | `vercel` | Vercel Serverless Functions |
| Fly.io | `fly` | Fly.io Machines |
| Deno Deploy | `deno-deploy` | Deno Deploy |

## Adding a Backend

### Dashboard

1. Go to **Apps → your app → Backends**
2. Click **Add Backend**
3. Select provider, region, and configure the connection

### API

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/backends \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "functionfly-edge",
    "region": "us-east-1",
    "url": "https://edge.functionfly.com/deploy/my-saas",
    "shared_secret": "sk_...",
    "priority": 1
  }'
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | Yes | Provider slug (see table above) |
| `region` | string | Yes | Deployment region (e.g. `us-east-1`, `eu-west-1`) |
| `url` | string | Yes | Deploy endpoint URL |
| `shared_secret` | string | Yes | HMAC shared secret for signed deploys |
| `priority` | int | No | Routing priority (lower = preferred) |
| `enabled` | bool | No | Enable/disable (default: true) |

## Routing

When a request arrives, the app routes to the highest-priority healthy backend.
If the primary backend fails, traffic automatically fails to the next backend.

Get the current routing decision:

```bash
curl https://api.functionfly.com/v1/apps/{appId}/route \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

```json
{
  "backend_id": "be_primary",
  "provider": "functionfly-edge",
  "region": "us-east-1",
  "latency_ms": 12,
  "failovers": [
    { "backend_id": "be_backup", "provider": "workers", "region": "eu-west-1" }
  ]
}
```

## Health Checks

Each backend runs periodic health checks. Results are stored and used for
routing decisions.

| State | Meaning |
|-------|---------|
| `closed` | Healthy — all traffic routed here |
| `open` | Unhealthy — traffic fails over |
| `half-open` | Testing recovery — limited traffic |

## Circuit Breaker

The circuit breaker tracks failure patterns per backend:

- **Closed → Open**: After consecutive failures exceed threshold
- **Open → Half-Open**: After cooldown period
- **Half-Open → Closed**: After successful test requests

## Listing Backends

```bash
curl https://api.functionfly.com/v1/apps/{appId}/backends \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Deleting a Backend

```bash
curl -X DELETE https://api.functionfly.com/v1/apps/{appId}/backends/{backendId} \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Next Steps

- [Deployments](/apps/deployments/) — Deploy and rollback
- [API Reference](/apps/api/) — Full endpoint docs
- [Providers](/providers/) — Provider setup guides
