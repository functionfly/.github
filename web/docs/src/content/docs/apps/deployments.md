---
title: Deployments
description: Deploy, rollback, and blue/green deployments
sidebar:
  order: 3
---

# Deployments

A **deployment** records a function being deployed to one of your app's
backends. Deployments track the full lifecycle from build to rollback.

## Deploying a Function

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/deploy \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123",
    "backend_id": "be_xyz789"
  }'
```

## Deployment Status

| Status | Description |
|--------|-------------|
| `pending` | Queued, waiting to start |
| `building` | Building the deployment artifact |
| `deploying` | Pushing to the backend |
| `success` | Deployed successfully |
| `failed` | Deployment failed |

## Listing Deployments

```bash
curl https://api.functionfly.com/v1/apps/{appId}/deployments \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

```json
{
  "deployments": [
    {
      "id": "dep_001",
      "function_id": "fx_abc123",
      "backend_id": "be_xyz789",
      "provider": "functionfly-edge",
      "region": "us-east-1",
      "status": "success",
      "created_at": "2026-06-20T10:00:00Z"
    }
  ]
}
```

## Getting a Deployment

```bash
curl https://api.functionfly.com/v1/deployments/{deploymentId} \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Rollback

Roll back to a previous deployment:

```bash
curl -X POST https://api.functionfly.com/v1/deployments/{deploymentId}/rollback \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Rollback creates a new deployment using the artifact from the specified
previous deployment. The rollback itself is a new deployment record, so
the full history is preserved.

## Blue/Green Deployments

For zero-downtime deployments on Cloudflare Workers:

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/deploy/blue-green \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123"
  }'
```

Blue/green deploys:
1. Deploy to the inactive slot (blue or green)
2. Run health checks on the new deployment
3. Switch traffic to the new slot
4. Keep the old slot as an instant rollback target

## Secrets Management

For providers that require secrets (e.g. Fly.io API tokens):

### Set Secrets

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/secrets \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "FLY_API_TOKEN": "fm1_...",
    "DATABASE_URL": "postgres://..."
  }'
```

### List Secrets

```bash
curl https://api.functionfly.com/v1/apps/{appId}/secrets \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

:::caution
Secret values are never returned in API responses. Only secret names and
metadata are shown.
:::

## Linking to External Providers

### Vercel

Link your app to a Vercel project for automatic deployments:

```bash
curl -X POST https://api.functionfly.com/v1/apps/{appId}/link \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "vercel",
    "project_id": "prj_..."
  }'
```

## Next Steps

- [Backends](/apps/backends/) — Configure deploy targets
- [API Reference](/apps/api/) — Full endpoint docs
- [CI/CD Integration](/guides/ci-cd/) — Automate deployments
