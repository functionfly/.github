# Refresh Token Rotation with Device Families — Design

## Status
Approved

## Overview

Implement refresh token rotation with **per-device token families** to detect and prevent replay attacks. When a refresh token is used, the system detects if it's a previously rotated token (potential theft) and invalidates the entire device family if so.

## Current State

Refresh tokens are stored with:
- `token_hash` (SHA-256 of the token)
- `ip_address`, `user_agent` (already tracked but not used for families)
- `revoked` flag

On refresh, the old token is revoked and a new one issued — but there's no reuse detection.

## Data Model

### Schema Change

Add `family_id` (UUID) to `refresh_tokens` table:

```sql
ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS family_id UUID DEFAULT gen_random_uuid();
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens(family_id) WHERE family_id IS NOT NULL;
```

### Family Concept

- A **family** = all refresh tokens issued to the same device (identified by user_agent + ip_address hash)
- First login: new family created, first token issued with this family
- Subsequent refreshes: new token added to same family, old token revoked
- **Reuse detection**: if a revoked token (not the latest) is used, family is compromised → all tokens in family revoked

### Token States

| State | Description |
|-------|-------------|
| `active` | Latest token in family, can be used |
| `rotated` | Previous token, revoked, cannot be used again |
| `compromised` | Used after rotation → entire family revoked |

## Flow

### Normal Rotation

```
1. User logs in → Token A (family=F1, state=active)
2. Client refreshes → Token B issued, Token A revoked (state=rotated)
3. Client refreshes → Token C issued, Token B revoked
```

### Attack Detection

```
1. Attacker steals Token B
2. User uses Token C legitimately
3. Attacker tries Token B → REUSE DETECTED
4. Family F1 marked compromised → Token C revoked, user notified
5. User must re-authenticate
```

## API Changes

### `/auth/refresh` (POST)

**Request:** Unchanged (refresh_token in body or cookie)

**Response on reuse detected:**
```json
{
  "error": "token_reuse_detected",
  "code": "REFRESH_TOKEN_REUSE",
  "message": "Suspicious token reuse detected. All sessions on this device have been revoked for security. Please log in again.",
  "device_family_compromised": true
}
```

**New headers:**
- `X-Token-Family`: family ID (stable device identifier, can be stored by client)
- `X-Token-State`: `active` | `rotated` (client shouldn't see rotated, but for debugging)

### `/auth/sessions` (GET) — New

Returns user's active sessions (token families):

```json
{
  "sessions": [
    {
      "family_id": "uuid",
      "device": "Chrome on Windows",
      "ip_address": "1.2.3.4",
      "last_used": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-10T08:00:00Z"
    }
  ]
}
```

### `/auth/sessions/:family_id/revoke` (DELETE) — New

Revoke a specific device family (logout from one device).

### `/auth/sessions/revoke-all` (POST) — Existing

Already exists, will work with families.

## Implementation Plan

### Phase 1: Schema & Core Logic

1. Migration: add `family_id` column
2. Update `CreateRefreshToken` to accept and set family_id
3. Update `GetRefreshTokenByHash` to return family_id
4. Add `GetActiveTokenInFamily` — get latest active token for a family
5. Add `RevokeEntireFamily` — revoke all tokens in family
6. Add `MarkFamilyCompromised` — mark family and notify user

### Phase 2: Rotation Logic

1. Modify `HandleRefreshToken`:
   - If token is active → rotate normally
   - If token is rotated → call `MarkFamilyCompromised`, return reuse error
   - If token is compromised/family revoked → return error
2. Generate family_id on first login (stored client-side for subsequent requests)
3. Pass family_id via `X-Token-Family` header

### Phase 3: Session Management

1. Add `ListUserSessions` endpoint (group by family)
2. Add `RevokeFamily` endpoint
3. Update logout to work with families

### Phase 4: Notifications

1. On family compromise: log security event
2. Future: email notification to user (out of scope for initial impl)

## Security Considerations

- Family detection happens BEFORE issuing new token (prevent attacker's token from being valid)
- Compromise is per-family, not per-user (other devices unaffected)
- Logs include family_id for forensic investigation
- Rate limiting remains active on refresh endpoint

## Compatibility

- Existing refresh tokens without family_id: treated as single-token family (backwards compatible)
- Migration is additive (existing tokens work until next rotation)
- Clients that don't send family_id: new family created on first refresh

## Env Var

`REFRESH_TOKEN_ROTATION_ENABLED=true` (default: true)

Set to `false` to disable rotation (not recommended except for migration periods).
