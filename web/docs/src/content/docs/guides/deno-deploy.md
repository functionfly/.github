---
title: Deno Deploy
description: Deploy FunctionFly functions to Deno Deploy
---

# Deno Deploy

Deploy your FunctionFly functions to [Deno Deploy](https://deno.com/deploy) for a modern JavaScript/TypeScript edge runtime with native Deno APIs.

## Features

- **Native Deno APIs** - Full Deno runtime at the edge
- **<10ms cold start** - Lightning fast execution
- **TypeScript native** - No build step required
- **Web standard APIs** - Use browser-compatible APIs

## Prerequisites

- Deno account
- Deno Deploy project created at [dash.deno.com](https://dash.deno.com)

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "deno-deploy",
  "provider_config": {
    "project": "my-project",
    "kv_namespace": "my-kv",
    "regions": ["us-east", "eu-west", "ap-northeast"]
  }
}
```

## Deployment

```bash
# Deploy to Deno Deploy
ffly deploy --provider deno-deploy

# Deploy specific region
ffly deploy --provider deno-deploy --region us-east
```

## Environment Variables

```bash
# Set Deno Deploy credentials
ffly env set DENO_DEPLOY_TOKEN=your_token --provider deno-deploy
ffly env set DENO_DEPLOY_PROJECT=my-project --provider deno-deploy
```

## Deno KV Integration

FunctionFly integrates with Deno's built-in KV:

```javascript
// Use Deno KV at the edge
const kv = await Deno.openKv();
const result = await kv.get(["users", "123"]);
await kv.set(["users", "123"], { name: "Alice" });
```

## Supported APIs

- `fetch` - Standard Web Fetch API
- `WebSocket` - Real-time connections
- `Crypto` - Web Crypto API
- `Streams` - Web Streams API
- `Deno KV` - Key-value store
- `Deno Queues` - Message queues

## Limitations

- Maximum 512MB memory
- Maximum 30 second timeout
- JavaScript/TypeScript only
- Deno-specific APIs (not Node.js)