---
title: API Keys API
description: Full API reference for API key management
sidebar:
  order: 4
---


All management endpoints require session authentication (JWT cookie).
Authentication endpoints accept the API key directly.

---

## Key Management

### Create API Key

```
POST /v1/api-keys
```

```json
{
  "name": "ci-deploy",
  "key_type": "platform",
  "description": "CI/CD deployment key",
  "rate_limit_rpm": 500,
  "rate_limit_rph": 30000,
  "rate_limit_rpd": 500000,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique name per tenant (max 255 chars) |
| `key_type` | string | Yes | `platform`, `function`, `agent`, `environment`, `oauth`, `trust`, `micropython`, `runtime` |
| `description` | string | No | Human-readable description |
| `rate_limit_rpm` | int | No | Requests per minute (default: 1000) |
| `rate_limit_rph` | int | No | Requests per hour (default: 60000) |
| `rate_limit_rpd` | int | No | Requests per day (default: 1000000) |
| `expires_at` | ISO 8601 | No | Optional expiration date |

Response includes `key` (plaintext, shown once):

```json
{
  "id": "key_abc123",
  "name": "ci-deploy",
  "key_type": "platform",
  "key_prefix": "ffp_",
  "key": "ffp_v1_a1b2c3d4..._ab",
  "is_active": true,
  "created_at": "2026-06-30T00:00:00Z"
}
```

### List API Keys

```
GET /v1/api-keys
```

Query parameters:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `key_type` | string | — | Filter by type |
| `is_active` | bool | — | Filter by active status |
| `search` | string | — | Search by name |
| `limit` | int | 50 | Page size |
| `offset` | int | 0 | Pagination offset |

### Get API Key

```
GET /v1/api-keys/{id}
```

Returns key details with permissions and environments (no plaintext).

### Update API Key

```
PATCH /v1/api-keys/{id}
```

```json
{
  "name": "ci-deploy-prod",
  "description": "Updated description",
  "is_active": true,
  "rate_limit_rpm": 2000,
  "rotation_frequency_days": 60
}
```

### Delete API Key

```
DELETE /v1/api-keys/{id}
```

Soft delete — sets `is_active=false`. The key can be reactivated via `PATCH`.

---

## Rotation

### Rotate API Key

```
POST /v1/api-keys/{id}/rotate
```

```json
{
  "reason": "manual"
}
```

Returns the new plaintext key (shown once).

### Get Rotation History

```
GET /v1/api-keys/{id}/rotations
```

---

## Permissions

### List Permissions

```
GET /v1/api-keys/{id}/permissions
```

### Add Permission

```
POST /v1/api-keys/{id}/permissions
```

```json
{
  "permission": "execute",
  "resource_type": "function",
  "resource_id": "fx_abc123"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `permission` | string | Yes | `read`, `write`, `execute`, `admin` |
| `resource_type` | string | Yes | `function`, `app`, `tenant`, `registry`, `deployment`, `secret` |
| `resource_id` | string | Yes | UUID of the specific resource |

### Remove Permission

```
DELETE /v1/api-keys/{id}/permissions/{permId}
```

---

## Environments

### List Linked Environments

```
GET /v1/api-keys/{id}/environments
```

### Link Environment

```
POST /v1/api-keys/{id}/environments
```

```json
{
  "environment_id": "env_prod_001",
  "environment_name": "production"
}
```

### Unlink Environment

```
DELETE /v1/api-keys/{id}/environments/{envId}
```

### List Available Environments

```
GET /v1/api-keys/environments/available
```

---

## Authentication

### Exchange Key for JWT

```
POST /v1/auth/api-key
```

Accepts the API key via `X-API-Key` or `Authorization: ApiKey` header.
Returns a 24-hour JWT.

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-07-01T00:00:00Z"
}
```

### Validate Key

```
POST /v1/auth/validate-key
```

Validates a key and returns its metadata without generating a JWT.
Used by the AI service for key verification.

---

## Error Codes

| Status | Code | Meaning |
|--------|------|---------|
| 400 | `INVALID_KEY_TYPE` | Unknown key type |
| 400 | `INVALID_PERMISSION` | Invalid permission or resource type |
| 404 | `KEY_NOT_FOUND` | Key ID not found |
| 409 | `KEY_NAME_EXISTS` | Name already in use in this tenant |
| 401 | `INVALID_KEY` | Key failed authentication |
| 403 | `KEY_EXPIRED` | Key has passed its expiration date |
| 403 | `KEY_REVOKED` | Key has been revoked |
| 429 | `RATE_LIMITED` | Per-key rate limit exceeded |
