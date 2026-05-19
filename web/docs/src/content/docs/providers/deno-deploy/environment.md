---
title: Deno Deploy Environment
description: Environment variables specific to Deno Deploy deployment.
---

Deploy FunctionFly functions to Deno's global edge network.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `DENO` | Set to `true` when running on Deno Deploy |
| `DENO_REGION` | Region where function is running |
| `DENO_DEPLOYMENT_ID` | Unique deployment ID |
| `DENO_ENVIRONMENT` | Environment (production, preview) |

## Regions

| Region | Code |
|--------|------|
| Asia Pacific | `ap-east-1` |
| Asia Pacific | `ap-northeast-1` |
| Asia Pacific | `ap-southeast-1` |
| Europe | `eu-west-1` |
| Europe | `eu-central-1` |
| US East | `us-east-1` |
| US West | `us-west-1` |
| US West | `us-west-2` |

## Deno KV

| Variable | Description |
|----------|-------------|
| `DENO_KV_PATH` | Path to local KV database (development) |
| `DENO_KV_REMOTE_URL` | Remote KV URL |
| `DENO_KV_REGION` | KV region |
| `DENO_KV_ORIGIN` | KV origin |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DENO_FUNCTION_MEMORY_MB` | `256` | Memory limit in MB |
| `DENO_FUNCTION_CPU_MS` | `50` | CPU time limit in ms |
| `DENO_FUNCTION_READINESS_TIMEOUT_MS` | `30000` | Readiness timeout |

## Secrets

Store secrets in Deno Deploy:

```bash
# Using Deno CLI
deno kv secret set API_KEY=your-secret

# Or via dashboard
```

Access in code:

```javascript
const apiKey = Deno.env.get('API_KEY');
```

## Cold Start

| Metric | Value |
|--------|-------|
| Cold start | < 5ms |
| Memory | Up to 512MB |
| CPU | 50ms compute per request |

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "deno-deploy",
  "environment": {
    "DENO_FUNCTION_MEMORY_MB": "256",
    "DENO_FUNCTION_CPU_MS": "50"
  },
  "runtime": "deno"
}
```

## Database Integration

Deno Deploy supports multiple databases:

```javascript
// PostgreSQL via Supabase
const { Client } = await import("postgres");

// PlanetScale
const mysql = await import("@planetscale/database");

// Upstash Redis
const redis = new Redis(process.env.UPSTASH_REDIS_URL);
```

## Web Assembly

Deploy WASM modules:

```javascript
const wasm = await WebAssembly.instantiateStreaming(
  fetch("https://your-cdn.com/module.wasm")
);
```