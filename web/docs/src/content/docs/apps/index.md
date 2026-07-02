---
title: Apps
description: Organize functions and manage multi-cloud deployments with Apps
---


An **App** is the top-level organizational unit in FunctionFly. Apps organize
your functions and manage multi-cloud deployments under a single project.

## What Is an App?

- A **named project** (e.g. `my-saas`) with a unique slug
- Gets a **public deploy URL**: `https://{slug}.functionfly.com`
- Acts as a **container for backends** — deploy targets on different cloud providers
- Is **scoped to a tenant** — all team members share access

Apps do not directly own functions. Instead, functions are **deployed to**
an app's backends. This separation lets you deploy the same function to
multiple providers (Cloudflare Workers, Vercel, Fly.io, etc.) under one app.

## Plan Limits

| Plan | Max Apps |
|------|----------|
| Free | 1 |
| Starter | 3 |
| Professional | 10 |
| Enterprise | Unlimited |

## Quick Start

### Create an App

```bash
curl -X POST https://api.functionfly.com/v1/apps \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My SaaS",
    "slug": "my-saas"
  }'
```

Or from the dashboard: **Apps → Create App**.

### Add a Backend

Add a deploy target (e.g. FunctionFly Edge):

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/backends \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "functionfly-edge",
    "region": "us-east-1",
    "url": "https://edge.functionfly.com/deploy/my-saas",
    "shared_secret": "sk_..."
  }'
```

### Deploy a Function

Deploy a function to one of your app's backends:

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/deploy \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123",
    "backend_id": "be_xyz789"
  }'
```

## How Apps Relate to Other Entities

| Entity | Relationship |
|--------|-------------|
| **Tenant** | Each app belongs to one tenant. Team members share access. |
| **Backends** | One-to-many. Each backend is a deploy target (provider + region). |
| **Functions** | Functions are deployed *to* app backends. They exist independently. |
| **Deployments** | One-to-many. Tracks deploy history with rollback support. |
| **Providers** | Cloud integrations (Vercel, Fly.io, etc.) referenced by backends. |
| **Bundles** | Bundle subscriptions can auto-create a default app. |

## Next Steps

- [Backends](/apps/backends/) — Configure multi-cloud deploy targets
- [Deployments](/apps/deployments/) — Deploy, rollback, and blue/green
- [API Reference](/apps/api/) — Full endpoint documentation
- [Providers](/providers/) — Supported cloud providers
