---
title: Cloudflare Workers
description: Deploy FunctionFly functions to Cloudflare Workers
---

# Cloudflare Workers

Deploy your FunctionFly functions to [Cloudflare Workers](https://workers.cloudflare.com/) for ultra-low latency edge compute with global anycast.

## Features

- **<5ms cold start** - Fastest edge runtime
- **Global anycast** - 300+ cities worldwide
- ** Durable Objects** - Stateful edge compute
- **Workers KV** - Global key-value store

## Prerequisites

- Cloudflare account with Workers enabled
- Wrangler CLI installed: `npm i -g wrangler`

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "workers",
  "provider_config": {
    "account_id": "your-cloudflare-account-id",
    "workers_subdomain": "your-subdomain",
    "routes": ["*.yourdomain.com/*"]
  }
}
```

## Deployment

```bash
# Deploy to Cloudflare Workers
ffly deploy --provider workers

# Deploy to specific zone
ffly deploy --provider workers --zone yourdomain.com
```

## Environment Variables

```bash
# Set Cloudflare credentials
ffly env set CF_ACCOUNT_ID=your_account --provider workers
ffly env set CF_API_TOKEN=your_api_token --provider workers
ffly env set CF_ZONE_ID=your_zone_id --provider workers
```

## Workers KV Integration

FunctionFly automatically integrates with Workers KV:

```javascript
// Access KV from your function
const value = await KV.get('my-key');
await KV.put('my-key', 'my-value');
```

## Limitations

- Maximum 512MB memory
- Maximum 30 second timeout
- No Node.js APIs (use WinterCG)
- CPU time limited (not wall time)