---
title: Apps API
description: Full API reference for apps, backends, and deployments
sidebar:
  order: 4
---

# Apps API Reference

All endpoints require authentication. Apps are scoped to the authenticated
user's tenant.

---

## Apps

### List Apps

```
GET /v1/apps
```

Returns all apps for the current tenant.

```json
{
  "apps": [
    {
      "id": "app_abc",
      "tenant_id": "ten_xyz",
      "name": "My SaaS",
      "slug": "my-saas",
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ]
}
```

### Create App

```
POST /v1/apps
```

```json
{
  "name": "My SaaS",
  "slug": "my-saas"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable name (max 50 chars) |
| `slug` | string | Yes | URL-safe identifier (max 63 chars, lowercase, no spaces, globally unique) |

**Errors:**

| Status | Code | Meaning |
|--------|------|---------|
| 400 | `INVALID_SLUG` | Slug contains invalid characters |
| 409 | `SLUG_TAKEN` | Slug is already in use |
| 403 | `APP_LIMIT_REACHED` | Plan limit exceeded |

### Get App

```
GET /v1/apps/{appId}
```

`appId` can be a UUID or a slug.

### Get App Status

```
GET /v1/apps/{appId}/status
```

Returns the app with backend statuses (health checks, circuit breaker state).

```json
{
  "app": { ... },
  "backends": [
    {
      "id": "be_001",
      "provider": "functionfly-edge",
      "region": "us-east-1",
      "enabled": true,
      "priority": 1,
      "health": { "ok": true, "status_code": 200, "latency_ms": 12 },
      "circuit_state": "closed"
    }
  ]
}
```

---

## Backends

### List Backends

```
GET /v1/apps/{appId}/backends
```

### Create Backend

```
POST /v1/apps/{appId}/backends
```

```json
{
  "provider": "functionfly-edge",
  "region": "us-east-1",
  "url": "https://edge.functionfly.com/deploy/my-saas",
  "shared_secret": "sk_...",
  "priority": 1,
  "enabled": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | Yes | Provider slug |
| `region` | string | Yes | Region identifier |
| `url` | string | Yes | Deploy endpoint URL |
| `shared_secret` | string | Yes | HMAC shared secret |
| `priority` | int | No | Routing priority (lower = preferred) |
| `enabled` | bool | No | Default: true |

### Delete Backend

```
DELETE /v1/apps/{appId}/backends/{backendId}
```

### Get Route

```
GET /v1/apps/{appId}/route
```

Returns the routing decision — which backend would handle a request right now,
with failover order.

---

## Deployments

### Deploy Function

```
POST /v1/apps/{appId}/deploy
```

```json
{
  "function_id": "fx_abc123",
  "backend_id": "be_xyz789"
}
```

### List Deployments

```
GET /v1/apps/{appId}/deployments
```

### Get Deployment

```
GET /v1/deployments/{deploymentId}
```

### Rollback

```
POST /v1/deployments/{deploymentId}/rollback
```

Creates a new deployment using the artifact from the specified deployment.

### Blue/Green Deploy

```
POST /v1/apps/{appId}/deploy/blue-green
```

```json
{
  "function_id": "fx_abc123"
}
```

---

## Secrets

### Set Secrets

```
POST /v1/apps/{appId}/secrets
```

```json
{
  "FLY_API_TOKEN": "fm1_...",
  "DATABASE_URL": "postgres://..."
}
```

### List Secrets

```
GET /v1/apps/{appId}/secrets
```

Returns secret names and metadata only — values are never returned.

---

## Link

### Link to External Provider

```
POST /v1/apps/{appId}/link
```

```json
{
  "provider": "vercel",
  "project_id": "prj_..."
}
```
