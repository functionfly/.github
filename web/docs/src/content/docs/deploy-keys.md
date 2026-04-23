---
title: "DEPLOY KEYS"
---

# SSH Deploy Keys

SSH Deploy Keys provide secure authentication for function deployments using SSH public key infrastructure.

## Overview

Deploy keys allow CI/CD systems, automation tools, and external services to authenticate when deploying functions without requiring personal user credentials.

## Key Features

- **SSH Public Key Support**: Supports `ssh-ed25519`, `ssh-rsa`, and `ecdsa-sha2-nistp*` key types
- **SHA256 Fingerprint**: Keys are indexed by SHA256:base64 fingerprint (same format as GitHub/GitLab deploy keys)
- **Expiration**: Optional expiration dates for temporary deploy keys
- **Last Used Tracking**: Audit trail of when deploy keys were last used

## Model

```go
type DeployKey struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Name        string
    PublicKey   string  // OpenSSH format
    Fingerprint string  // SHA256:base64
    CreatedAt   time.Time
    ExpiresAt   *time.Time
    LastUsedAt  *time.Time
    CreatedBy   uuid.UUID
}
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/deploy-keys` | Create a new deploy key |
| GET | `/v1/deploy-keys` | List all deploy keys |
| GET | `/v1/deploy-keys/{id}` | Get a specific deploy key |
| DELETE | `/v1/deploy-keys/{id}` | Delete a deploy key |
| POST | `/v1/deploy-keys/{id}/verify` | Verify a deploy key is valid |

## Creating a Deploy Key

```bash
curl -X POST https://api.functionfly.com/v1/deploy-keys \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-server-deploy-key",
    "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "ci-server-deploy-key",
  "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...",
  "fingerprint": "SHA256:abc123...",
  "created_at": "2026-04-21T18:57:00Z"
}
```

## Supported Key Formats

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB...
ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY...
```

## Security

- Public keys are validated on creation (correct format, valid base64)
- Fingerprint is computed as SHA256 of the raw key data
- Failed verification attempts are logged
- Expired keys automatically fail verification

## Database

Table: `deploy_keys`

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant ownership |
| name | VARCHAR(255) | Human-readable name |
| public_key | TEXT | OpenSSH public key |
| fingerprint | VARCHAR(64) | SHA256:base64, unique |
| created_at | TIMESTAMPTZ | Creation timestamp |
| expires_at | TIMESTAMPTZ | Optional expiration |
| last_used_at | TIMESTAMPTZ | Last verification |
| created_by | UUID | User who created |

Indexes: `idx_deploy_keys_tenant`, `idx_deploy_keys_fingerprint`