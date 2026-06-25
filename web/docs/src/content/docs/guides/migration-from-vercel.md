---
title: Migration from Vercel
description: Step-by-step guide to migrating your Vercel Functions to FunctionFly.
---

# Migration from Vercel

This guide walks you through migrating functions from Vercel to FunctionFly.

## Overview

| Aspect | Vercel Functions | FunctionFly |
|--------|------------------|-------------|
| Timeouts | 10s (Hobby) / 60s (Pro) | Up to 300s |
| Frameworks | Next.js required | Any HTTP handler |
| State | External services | Built-in StateFabric |
| Verification | None | Trust Protocol |

## Prerequisites

```bash
# Install FunctionFly CLI
npm install -g @functionfly/cli

# Login
ffly login
```

## Step 1: Export from Vercel

```bash
# Download your project
vercel download

# Export environment variables
vercel env pull .env.vercel
```

## Step 2: Create FunctionFly Project

```bash
# Create project
ffly init my-project

# Copy your files
cp -r my-vercel-project/* ./my-project/
```

## Step 3: Update Handler Code

### Next.js API Routes

```javascript
// Vercel (Next.js API route)
// pages/api/hello.js or app/api/hello/route.js
export async function handler(req, res) {
  res.status(200).json({ message: 'Hello' });
}

// FunctionFly
export default async function handler(context) {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello' })
  };
}
```

### Standard Serverless

```javascript
// Vercel
module.exports = (req, res) => {
  res.status(200).json({ message: 'Hello' });
};

// FunctionFly
export default async function handler(context) {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello' })
  };
}
```

## Step 4: Migrate Environment Variables

```bash
# Import from Vercel export
ffly env set --env-file .env.vercel
```

## Step 5: Deploy

```bash
# Deploy to production
ffly deploy --prod

# Or use a custom domain
ffly domain add mysite.com
```

## Vercel to FunctionFly Reference

| Vercel Concept | FunctionFly Equivalent |
|----------------|------------------------|
| API Routes | Functions |
| Serverless Functions | Functions |
| Edge Functions | Edge Runtime |
| Vercel KV | StateFabric |
| Environment Variables | `ffly env` |
| `vercel.json` | `ffly.yml` |
| ISR / Revalidation | Built-in caching |

## Key Differences

1. **No framework required**: FunctionFly works with plain HTTP handlers
2. **Longer timeouts**: Up to 300s vs Vercel's 60s
3. **Built-in state**: StateFabric provides persistent storage
4. **Trust certificates**: Automatic execution verification

## Need Help?

- **Discord**: [community server](https://discord.gg/functionfly)
- **Support**: support@functionfly.com
