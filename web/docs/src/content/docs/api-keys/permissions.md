---
title: Permissions
description: Fine-grained access control for API keys
sidebar:
  order: 2
---

# API Key Permissions

API keys support fine-grained permissions that control which resources a key
can access and what operations it can perform.

## Permission Model

A permission is a tuple of **(action, resource type, resource ID)**:

```
(read, function, fx_abc123)
```

This grants read-only access to the specific function `fx_abc123`.

## Actions

| Action | Description |
|--------|-------------|
| `read` | View resource metadata, list resources |
| `write` | Create, update, delete resources |
| `execute` | Invoke functions, run deployments |
| `admin` | Full access — implies read, write, and execute |

## Resource Types

| Resource Type | Scope |
|---------------|-------|
| `function` | A specific function |
| `app` | An app and its backends |
| `tenant` | Tenant-wide operations (billing, team, settings) |
| `registry` | Function registry (publish, search) |
| `deployment` | Deployment operations |
| `secret` | Secrets vault access |

## Managing Permissions

### Add a Permission

```bash
curl -X POST https://api.functionfly.com/v1/api-keys/{keyId}/permissions \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permission": "execute",
    "resource_type": "function",
    "resource_id": "fx_abc123"
  }'
```

### List Permissions

```bash
curl https://api.functionfly.com/v1/api-keys/{keyId}/permissions \
  -H "Authorization: Bearer $SESSION_TOKEN"
```

```json
{
  "permissions": [
    {
      "id": "perm_001",
      "permission": "execute",
      "resource_type": "function",
      "resource_id": "fx_abc123",
      "created_at": "2026-06-30T00:00:00Z"
    }
  ]
}
```

### Remove a Permission

```bash
curl -X DELETE https://api.functionfly.com/v1/api-keys/{keyId}/permissions/{permId} \
  -H "Authorization: Bearer $SESSION_TOKEN"
```

## Trust API Keys

Trust API keys (`fft_` prefix) use a different model — **scope-based
permissions** stored as a JSONB map on the key:

```json
{
  "trust:read": true,
  "trust:write": true
}
```

Scopes support wildcards (`*` grants all scopes). Trust keys also support
**IP allowlisting** with exact IPs and CIDR ranges.

## Environments

Keys can be linked to specific environments (e.g. `production`, `staging`).
When linked, the key can only access resources in those environments.

### Link an Environment

```bash
curl -X POST https://api.functionfly.com/v1/api-keys/{keyId}/environments \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "environment_id": "env_prod_001",
    "environment_name": "production"
  }'
```

### List Linked Environments

```bash
curl https://api.functionfly.com/v1/api-keys/{keyId}/environments \
  -H "Authorization: Bearer $SESSION_TOKEN"
```

## Best Practices

- **Least privilege** — Grant only the permissions needed
- **Scope to resources** — Avoid tenant-wide permissions for narrow use cases
- **Use environment links** — Restrict CI keys to `staging`, production keys to `production`
- **Rotate regularly** — Default rotation frequency is 90 days
- **Audit permissions** — Review permission grants periodically

## Next Steps

- [Rotation](/api-keys/rotation/) — Key lifecycle and rotation
- [API Reference](/api-keys/api/) — Full endpoint docs
