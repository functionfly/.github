---
title: Migration from Cloudflare Workers
description: Step-by-step guide to migrating your Cloudflare Workers to FunctionFly.
---

# Migration from Cloudflare Workers

This guide walks you through migrating Workers from Cloudflare to FunctionFly.

## Overview

| Aspect | Cloudflare Workers | FunctionFly |
|--------|-------------------|-------------|
| Runtimes | V8 (JavaScript/WASM) | 10+ languages |
| Max Timeout | 30s (CPU) / 300s (Wall) | 300s |
| Storage | KV, Durable Objects | StateFabric |
| Cold Start | ~5ms | ~50ms |

## Prerequisites

```bash
# Install FunctionFly CLI
npm install -g @functionfly/cli

# Login
ffly login
```

## Step 1: Export Workers

```bash
# Verify wrangler authentication
wrangler whoami

# Download Worker code from Cloudflare Dashboard
# Or use wrangler for WASM modules
```

## Step 2: Create FunctionFly Function

```bash
# Create function with Deno runtime (closest to Workers)
ffly init my-worker --runtime deno

# Or use Node.js/JavaScript
ffly init my-worker --runtime node20
```

## Step 3: Update Handler Code

### Basic Worker

```javascript
// Cloudflare Worker
export default {
  async fetch(request) {
    return new Response('Hello', { status: 200 });
  }
};

// FunctionFly
export default async function handler(context) {
  return {
    statusCode: 200,
    body: 'Hello'
  };
}
```

### With KV Access

```javascript
// Cloudflare Worker with KV
const value = await NAMESPACE.get('key');
await NAMESPACE.put('key', 'value');

// FunctionFly with StateFabric
const state = context.state('my-namespace');
const value = await state.get('key');
await state.set('key', 'value');
```

### Fetch API

```javascript
// Cloudflare Worker
const response = await fetch('https://api.example.com');
const data = await response.json();

// FunctionFly
const response = await fetch('https://api.example.com');
const data = await response.json();
```

## Step 4: Migrate Environment Variables

```bash
# Export from Cloudflare
# Dashboard: Workers > Settings > Environment Variables

# Import to FunctionFly
ffly env set API_URL=https://api.example.com
```

## Step 5: Deploy

```bash
# Deploy
ffly deploy --prod

# Set environment variables per stage
ffly env set --stage production --env-file prod.env
```

## Key Differences

| Workers Concept | FunctionFly Equivalent |
|-----------------|------------------------|
| KV Namespace | StateFabric |
| Durable Objects | StateFabric + Functions |
| Service Workers | Functions |
| `addEventListener('fetch')` | `export default` |
| `waitUntil()` | Background tasks |
| `caches.default` | Built-in CDN |

## Workers Specific Migrations

### Service Bindings

```javascript
// Cloudflare Worker with service binding
const response = await env.MY_SERVICE.fetch(request);

// FunctionFly
const response = await context.call('my-service', request);
```

### WebSockets

```javascript
// Cloudflare (limited support)
// Use Durable Objects for stateful connections

// FunctionFly
// WebSocket support via extensions
export default async function handler(context) {
  if (context.isWebSocket) {
    const ws = context.websocket;
    ws.on('message', (msg) => ws.send(msg));
    return;
  }
  return { statusCode: 200 };
}
```

## Need Help?

- **Discord**: [community server](https://discord.gg/functionfly)
- **Support**: support@functionfly.com
