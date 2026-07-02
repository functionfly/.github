---
title: Rotation
description: Key rotation, lifecycle management, and audit trail
sidebar:
  order: 3
---


Key rotation replaces a key's credentials while preserving its identity
(key ID, permissions, environments). This is essential for security hygiene
and incident response.

## When to Rotate

- **Scheduled** — Default rotation frequency is 90 days
- **Compromised** — Key exposed in logs, client-side code, or breach
- **Employee departure** — Team member with key access leaves
- **Compliance** — SOC 2 / ISO 27001 key rotation requirements

## Rotating a Key

### Dashboard

1. Go to **Settings → API Keys**
2. Find the key and click **Rotate**
3. Confirm the rotation
4. Copy the new key (shown once)

### API

```bash
curl -X POST https://api.functionfly.com/v1/api-keys/{id}/rotate \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "manual"
  }'
```

Response (save the new `key` — shown only once):

```json
{
  "id": "key_abc123",
  "name": "ci-deploy",
  "key_version": 2,
  "key": "ffp_v1_new_hex_string_checksum",
  "last_rotated_at": "2026-06-30T00:00:00Z"
}
```

## What Happens on Rotation

1. A new random key is generated
2. The new key's bcrypt hash replaces the old hash
3. The old hash is archived in `api_key_rotations`
4. `key_version` is incremented
5. The new plaintext is returned once
6. All existing integrations must update to the new key

:::caution
Rotation is immediate. The old key stops working as soon as rotation completes.
Update all integrations before rotating.
:::

## Rotation Reasons

| Reason | Description |
|--------|-------------|
| `manual` | User-initiated rotation |
| `automatic` | Scheduled rotation (platform-initiated) |
| `compromised` | Emergency rotation due to suspected compromise |

## Rotation History

View the full audit trail of rotations for a key:

```bash
curl https://api.functionfly.com/v1/api-keys/{id}/rotations \
  -H "Authorization: Bearer $SESSION_TOKEN"
```

```json
{
  "rotations": [
    {
      "id": "rot_002",
      "rotated_at": "2026-06-30T00:00:00Z",
      "rotation_reason": "manual",
      "created_by": "user_xyz"
    },
    {
      "id": "rot_001",
      "rotated_at": "2026-03-31T00:00:00Z",
      "rotation_reason": "automatic",
      "created_by": "system"
    }
  ]
}
```

## Key Lifecycle

```
  CREATE           USE           ROTATE          REVOKE
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Generate │   │ Authenticate│ │ Generate │   │ Soft     │
│ key      │──►│ + rate     │──►│ new key  │──►│ delete   │
│ Show once│   │ limit      │   │ Archive  │   │ is_active│
└──────────┘   └──────────┘   │ old hash │   │ = false  │
                              └──────────┘   └──────────┘
```

## Deactivation vs. Rotation

| Action | Effect | Reversible |
|--------|--------|------------|
| **Rotate** | New key, old key stops working | No (new key issued) |
| **Deactivate** | Key stops working, can be reactivated | Yes |
| **Delete** | Soft delete — key marked inactive | Yes (reactivate) |

Deactivation is useful for temporarily disabling a key without rotating:

```bash
curl -X PATCH https://api.functionfly.com/v1/api-keys/{id} \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "is_active": false }'
```

## Automatic Rotation

Set a rotation frequency when creating or updating a key:

```bash
curl -X PATCH https://api.functionfly.com/v1/api-keys/{id} \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "rotation_frequency_days": 30 }'
```

The platform will initiate automatic rotation when the key reaches its
rotation age. The new key is available via the dashboard (automatic rotation
does not expose plaintext via API for security).

## Best Practices

- **Rotate on compromise** — Immediately rotate if a key may have been exposed
- **Automate rotation** — Use CI/CD to update secrets on rotation
- **Monitor rotation history** — Audit for unexpected rotations
- **Use short-lived keys for CI** — 30-day rotation for pipeline keys
- **Keep production keys longer** — 90-day rotation with monitoring

## Next Steps

- [API Reference](/api-keys/api/) — Full endpoint docs
- [Permissions](/api-keys/permissions/) — Access control
- [Secrets Vault](/secrets-vault/) — Store keys securely
