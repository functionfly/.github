---
title: Vercel
description: Deploy FunctionFly functions to Vercel Edge Functions
---

# Vercel

Deploy your FunctionFly functions to [Vercel Edge Functions](https://vercel.com/docs/concepts/functions/edge-functions) for ultra-low latency global distribution.

## Features

- **Global edge network** - 100+ PoPs worldwide
- **<50ms cold start** - Near-instant execution
- **Automatic scaling** - Handles any traffic load
- **Edge middleware** - Modify requests/responses

## Prerequisites

- Vercel account with Edge Functions enabled
- Vercel CLI installed: `npm i -g vercel`

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "vercel",
  "provider_config": {
    "framework": "nextjs",  // or "svelte", "remix", "other"
    "regions": ["iad1", "sfo1", "cdg1"]
  }
}
```

## Deployment

```bash
# Deploy to Vercel
ffly deploy --provider vercel

# Deploy to specific region
ffly deploy --provider vercel --region iad1
```

## Environment Variables

```bash
# Set Vercel tokens
ffly env set VERCEL_TOKEN=your_token --provider vercel
ffly env set VERCEL_ORG_ID=your_org --provider vercel
ffly env set VERCEL_PROJECT_ID=your_project --provider vercel
```

## Routes

Vercel Edge Functions use a `/api/edge/*` route pattern. FunctionFly functions are automatically mapped to edge functions.

## Limitations

- Maximum 3008MB memory
- Maximum 30 second timeout
- No WebSocket support
- Limited filesystem access