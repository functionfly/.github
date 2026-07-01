---
title: API Keys
description: Create and manage API keys for programmatic access to FunctionFly
---

# API Keys

API keys are long-lived bearer tokens for programmatic access to the
FunctionFly platform. Use them with the CLI, SDKs, CI/CD pipelines,
and direct API calls.

## Key Format

```
{prefix}v1_{32 hex chars}_{2 hex checksum}
```

Example: `ffp_v1_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6_a1b2`

### Key Types

| Prefix | Type | Purpose |
|--------|------|---------|
| `ffp_` | Platform | Full platform API access |
| `fff_` | Function | Function execution only |
| `aep_` | Agent | AI agent access |
| `ffx_` | Edge | Edge API access |
| `ffe_` | Environment | Environment-specific access |
| `ffo_` | OAuth | OAuth-based access |
| `fft_` | Trust | Trust API partner access |
| `ffmp_` | MicroPython | MicroPython runtime (Enterprise) |
| `ffr_` | Runtime | Runtime execution (bun, deno, kotlin, etc.) |

## Security

- Keys are **hashed with bcrypt** (cost 12) — the platform never stores plaintext
- The plaintext key is shown **exactly once** at creation/rotation
- Keys support optional **IP allowlisting** (exact IPs and CIDR ranges)
- Rate limiting is enforced per key via Redis sliding-window algorithm

## Quick Start

### Create a Key

**Dashboard:** Settings → API Keys → Create API Key

**API:**

```bash
curl -X POST https://api.functionfly.com/v1/api-keys \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-deploy",
    "key_type": "platform",
    "description": "CI/CD deployment key"
  }'
```

Response (save the `key` — it is shown only once):

```json
{
  "id": "key_abc123",
  "name": "ci-deploy",
  "key_type": "platform",
  "key_prefix": "ffp_",
  "key": "ffp_v1_a1b2c3d4e5f6..._a1b2",
  "created_at": "2026-06-30T00:00:00Z"
}
```

### Use a Key

Include it in the `X-API-Key` or `Authorization` header:

```bash
# X-API-Key header
curl https://api.functionfly.com/v1/functions \
  -H "X-API-Key: ffp_v1_a1b2c3d4e5f6..._a1b2"

# Authorization header
curl https://api.functionfly.com/v1/functions \
  -H "Authorization: ApiKey ffp_v1_a1b2c3d4e5f6..._a1b2"
```

### Exchange for a JWT

For session-based workflows, exchange a key for a 24-hour JWT:

```bash
curl -X POST https://api.functionfly.com/v1/auth/api-key \
  -H "X-API-Key: ffp_v1_a1b2c3d4e5f6..._a1b2"
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-07-01T00:00:00Z"
}
```

## Rate Limits

Limits are enforced per key using a Redis sliding-window algorithm:

| Metric | Default |
|--------|---------|
| Requests per minute | 1,000 |
| Requests per hour | 60,000 |
| Requests per day | 1,000,000 |

Custom limits can be set per key. Response headers report usage:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 847
X-RateLimit-Reset: 1719705660
```

If Redis is unavailable, requests are **denied** (503) — the system fails closed.

## Permissions

Platform keys support fine-grained resource-based permissions:

| Permission | Description |
|------------|-------------|
| `read` | Read-only access to resources |
| `write` | Create and modify resources |
| `execute` | Invoke functions |
| `admin` | Full access (implies all others) |

Permissions are scoped to resource types:

| Resource Type | Description |
|---------------|-------------|
| `function` | Individual function |
| `app` | App and its backends |
| `tenant` | Tenant-wide operations |
| `registry` | Function registry |
| `deployment` | Deployment operations |
| `secret` | Secrets vault |

## Key Rotation

Rotate a key to generate new credentials while preserving the key ID:

```bash
curl -X POST https://api.functionfly.com/v1/api-keys/{id}/rotate \
  -H "Authorization: Bearer $SESSION_TOKEN"
```

- Old hash is archived in the rotation history
- New plaintext is returned once
- All existing integrations must update to the new key

## Next Steps

- [Permissions](/api-keys/permissions/) — Fine-grained access control
- [Rotation](/api-keys/rotation/) — Key rotation and lifecycle
- [API Reference](/api-keys/api/) — Full endpoint documentation
- [Authentication](/guides/authentication/) — Auth overview
