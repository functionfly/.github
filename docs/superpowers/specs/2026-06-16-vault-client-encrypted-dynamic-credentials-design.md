# Client-Side Encrypted Dynamic Credentials - Implementation Design

**Date:** 2026-06-16
**Status:** Draft
**Scope:** Add client-side encryption at rest for the FunctionFly vault's dynamic credentials subsystem, while preserving issuance to customer databases. Reuse existing `VaultCrypto` primitives. Support interactive (dashboard) and headless (CI/CD agent) flows. Work for every existing account type / plan tier.

---

## Overview

The FunctionFly vault is already zero-knowledge for **static secrets** — `Secret.EncryptedValue` is AES-256-GCM ciphertext produced in the browser, derived from the user's vault passphrase via Argon2id, and the server has no key to decrypt it. See `web/dashboard/src/utils/vault-crypto.ts:55-319` and `internal/storage/vault/models.go:12-67`.

**Dynamic credentials are not zero-knowledge today.** The target's admin password is wrapped with `Argon2id(SERVER_MASTER_KEY || tenantID, salt)` — the server holds the wrapping key (`SERVER_MASTER_KEY` env var + tenant ID) and can decrypt at any time. See `internal/crypto/server_encryption.go:28-150` and `internal/storage/vault/dynamic_encryption.go:15-45`. The minted lease password is generated server-side, returned as plaintext over TLS, and never persisted (per `internal/storage/vault/dynamic_service.go:258-269` and `internal/api/handlers/vault/types.go:493-505`).

This design closes the "server can decrypt the target admin password" gap. The architectural change is significant because the server must continue to run `CREATE USER` / `DROP USER` against the customer's database, which inherently requires the admin password in plaintext at issuance time. The solution is the well-established **"client-wraps, server-proxies"** pattern (cf. HashiCorp Vault's database secrets engine):

1. The decryption key (a per-tenant **DEK**) lives only on the client (browser or agent binary).
2. The server stores the admin password as opaque ciphertext.
3. At issuance, the client decrypts the admin password locally, sends it in a single HTTPS request, the server uses it once to run SQL against the customer DB, **explicitly zeroizes the in-memory copy**, and returns the lease material.

The user's "vault passphrase" (already used to wrap static secrets) is reused. A new per-tenant DEK is generated client-side at first vault unlock, wrapped under the user's passphrase-derived KEK, and POSTed to the server as opaque bytes.

The minted lease password becomes **client-generated** so the server never sees it in memory at all. The lease row in the database is unchanged (no password column).

---

## Goals & Non-Goals

### Goals
- Make target admin passwords **unrecoverable by the server** at rest, even with full DB read access and `SERVER_MASTER_KEY`.
- Work for **every existing account type and plan tier** (free → agent enterprise; personal → team).
- Support **interactive** (dashboard tab) and **headless** (CI/CD agent binary) issuance.
- **Reuse** the existing `VaultCrypto` primitives and the static-secret zero-knowledge pattern wherever possible.
- **Close** the pre-existing RBAC-enforcement gap on dynamic-cred handlers as part of this change.
- **Add** MFA gating, audit logging, and SIEM delivery to dynamic-cred handlers.
- **Add** CI/CD access via a new `dynamic_wrapped_access_tokens` model.
- **Migrate** escrow to wrap the new DEK (one escrow blob per tenant, recovers both static secrets and dynamic target admin passwords).
- **Backwards-compatible**: existing server-mode targets continue to work unchanged.

### Non-Goals (deferred to v2)
- Asymmetric key wrapping (Ed25519 / X25519 / YubiKey / post-quantum). The data model is reserved (`vault_user_keys` table added but unused in v1) so v2 is purely additive.
- Per-lease sharing (today: target-level shares only).
- Direct client-to-DB connectivity (the HashiCorp Vault Agent pattern, Approach B from the brainstorming). The architecture supports it as a v2 addition; v1 stays client-wraps, server-proxies.
- Removal of the legacy `server` encryption mode. v1 ships both side-by-side; deprecation timeline is in §11.

---

## Threat Model

### What this change closes

| Threat | Today | After v1 |
|---|---|---|
| Attacker reads DB dump + `SERVER_MASTER_KEY` → recovers target admin password | ❌ Possible | ✅ Admin password is encrypted to a per-tenant **client-derived key** the server does not hold |
| Operator with `SELECT *` on `dynamic_secret_targets` recovers admin password | ❌ Possible | ✅ Same as above — ciphertext is opaque |
| Attacker compromises a single API replica's process memory → recovers target admin password | ❌ Possible (persisted) | ⚠️ Only during an active issuance request (~10s window), and only in the issuing handler's stack frame |
| Attacker compromises API and replays old HTTP traffic | ❌ Possible (no replay protection on `generate`) | ⚠️ Same residual — JWT lifetime + audit anomaly detection. No new vector opened. |
| Attacker compromises browser sessionStorage | ⚠️ Same as today for static secrets | ⚠️ Same — DEK only exists in memory while a tab is open. Agent binary uses OS keychain. |

### Residual accepted risk

- **The server sees the target admin password in memory for ~10s during each issuance.** The client decrypts locally, submits over TLS, server uses once, zeroizes, returns the lease. This is the well-bounded residual that every comparable system (HashiCorp Vault database engine, AWS RDS IAM, Akeyless dynamic secrets) accepts. Mitigations:
  - Short request timeout (existing dynamic-cred handlers use 30s; tighten to 15s for client-mode generate).
  - Explicit `zeroize` of the in-memory password bytes after use.
  - Audit row recorded with `client_unwrap_target_id` and IP, but **no** password material.
  - No disk persistence (request body is not written to any log; the gin/chi default access log is configured to omit request bodies for these routes).
  - Rate limiting per `(user, target, IP)` (10/minute, see §10).
- **The user loses their vault passphrase AND has no escrow → they lose all client-mode dynamic targets AND all client-mode static secrets.** Same risk as today for static secrets. Escrow (`SupportsVaultEscrow`, Enterprise/Agent Enterprise) recovers both.
- **The minted lease password lives only on the client** (browser/agent memory) and the customer DB. If the client loses it, the lease is effectively dead — the DB user exists but nobody has the password. This is the same shape as today; we don't change the lifecycle.

---

## Architecture

### "Client-wraps, server-proxies"

```
┌─────────────────────┐                  ┌─────────────────────┐                 ┌──────────────────┐
│  Dashboard / Agent  │                  │  FunctionFly API    │                 │ Customer DB      │
│                     │                  │                     │                 │ (Postgres/MySQL) │
│ Holds:              │   ① POST target  │ Holds:              │                 │                  │
│  - vault_passphrase │ ────────────────►│  - encrypted blob   │                 │                  │
│  - derived KEK      │   (ciphertext)   │    (cannot decrypt) │                 │                  │
│  - in-memory DEK    │                  │                     │                 │                  │
│                     │                  │                     │                 │                  │
│                     │   ② POST generate │                    │                 │                  │
│  - unwraps admin pw │   (one-shot:     │  - opens DB conn    │   ③ CREATE USER │                  │
│  - generates new pw │    admin_pw +    │  - runs CREATE USER │ ───────────────►│                  │
│  - signs request    │    new_user_pw)  │  - closes conn      │   ④ OK          │                  │
│                     │ ────────────────►│  - zeros memory     │ ◄───────────────│                  │
│  - stores lease pw  │   ⑤ lease info   │                     │                 │                  │
│  in memory          │ ◄────────────────│                     │                 │                  │
└─────────────────────┘                  └─────────────────────┘                 └──────────────────┘
```

**Trust boundary:** the server never holds a key that can unwrap the target admin password. The client is the only entity that can. The "proof of possession" of the DEK is the unwrapped admin password itself — only a client that holds the DEK can produce it.

### Replay protection

No HMAC-based `client_unwrap_token` is used. The natural proof-of-possession model — the unwrapped admin password being correct — is sufficient. The server's checks are:

1. JWT is valid.
2. The admin password, when used to connect to the customer DB, succeeds (implicit verification: the `CREATE USER` statement will fail if the password is wrong).

Replayed requests cannot succeed because:
- The minted lease username is unique per request (`vault_p_<random>`, see `internal/storage/vault/dynamic_service.go:241-256`).
- Even if an attacker replays the same `(target_id, admin_password, ttl)` triple, they would re-`CREATE USER` with a fresh random username, minting a new lease — the previous lease is unaffected.

We add **per-`(user, target, IP)` rate limiting** (10 requests/minute) to bound the issuance rate and detect anomalies. See §10.

---

## Crypto Envelope

### Per-tenant DEK

A new **per-tenant DEK** (256-bit AES key) is the unit of encryption. The DEK is:
- **Generated client-side** at first vault unlock (`crypto.getRandomValues(32)`).
- **Wrapped under the user's KEK** (which is `Argon2id(vault_passphrase, dek_salt)`) and POSTed to the server as opaque bytes.
- **Cached in browser/agent memory** for the session's lifetime. Never persisted on the server in plaintext.
- **Multi-user**: each user has their own `vault_tenant_keys` row with their own wrapped DEK. Sharing DEK access between users is an admin operation that re-wraps under the target user's KEK.

**Why a per-tenant DEK (not pure passphrase-derived per row)?**
- Argon2id at 64 MiB × 3 iters × p=4 = ~250ms per derivation. We don't want that on every dashboard load or on every CI/CD issuance.
- The static-secrets model accepts this cost *once per page load* because secrets are read infrequently. Dynamic credentials can be issued many times per minute from a CI/CD pipeline.
- A cached DEK enables ~1ms per unwrap in JS (AES-GCM with a held key).

### Per-target admin password encryption

```text
encrypted_admin_password = AES-256-GCM(
    plaintext = admin_password,
    key       = DEK,
    iv        = wrap_iv,                  -- 12 random bytes
    aad       = "client-wrap:" ‖ tenantID ‖ ":" ‖ targetID
)
```

AAD binds the ciphertext to its row — a ciphertext created for `target A` cannot be replayed as a ciphertext for `target B`, even by a server that somehow held the DEK.

### Per-mint request signature (the new password)

The minted DB user password is generated **client-side**:
```text
password = generatePassword(24)  -- 24 chars from 56-char alphabet, ~140 bits entropy
        -- (existing generatePassword() in internal/storage/vault/dynamic_service.go:258-269, but moved to client)
```

The password is submitted to the server in the one-shot `/generate` request body. The server passes it directly to `CREATE USER` and never persists it. The lease row continues to store only the username (no password column).

### Comparison: today vs. v1

| Layer | Today | v1 |
|---|---|---|
| **Target admin password at rest** | `Argon2id(SERVER_MASTER_KEY ‖ tenantID, salt)` + AES-256-GCM, AAD = `server-encrypt:‹tenantID›` | `Argon2id(vault_passphrase ‖ tenant_salt, dek_salt)` + AES-256-GCM, AAD = `client-wrap:‹tenantID›:‹targetID›` |
| **Per-tenant key** | `SERVER_MASTER_KEY` env var | **Per-tenant DEK** generated client-side, wrapped per-user under each user's KEK |
| **Minted lease password** | Generated server-side, returned as plaintext over TLS, never persisted | Generated **client-side**, submitted in one-shot request, never persisted |
| **Generate response** | `{lease_id, username, password (plaintext), host, port, db, expires_at}` | Unchanged in v1 — plaintext over TLS. **v2 enhancement:** wrap with `response_key` for at-rest safety in HTTP logs. |
| **Static secrets** | Client-side encrypted with passphrase-derived key | **Unchanged** — same model. Escrow blob updated to wrap the DEK instead. |
| **KDF** | Argon2id (64 MiB, t=3, p=4) for static-secret v3 | **Same** — `key_version=3` (Argon2id) for the DEK wrap. Reuse `VaultCrypto.encryptWithPassphrase` (lines 267-293). |

### Rejected alternatives

- **Per-row passphrase-derived key (no DEK):** ~250ms per unwrap. Too slow for CI/CD hot paths.
- **Asymmetric key wrapping (v2):** Deferred. Data model reserved.
- **Pure server-held DEK (YubiCloud / HSM):** Doesn't close the threat model. Deferred.
- **MPC / threshold:** Out of scope. v2 if customer demand emerges.

---

## Data Model

### New: `vault_tenant_keys`

```sql
CREATE TABLE vault_tenant_keys (
  tenant_id      UUID         NOT NULL,
  user_id        UUID         NOT NULL,
  wrapped_dek    BYTEA        NOT NULL,        -- AES-256-GCM(DEK, KEK)
  dek_iv         BYTEA        NOT NULL,
  dek_auth_tag   BYTEA        NOT NULL,
  dek_salt       BYTEA        NOT NULL,        -- salt for KEK Argon2id
  key_version    INT          NOT NULL DEFAULT 3,
  kdf_params     JSONB        NOT NULL DEFAULT '{"t":3,"m":65536,"p":4}',
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  rotated_at     TIMESTAMPTZ,
  PRIMARY KEY (tenant_id, user_id)
);
CREATE INDEX idx_vault_tenant_keys_tenant ON vault_tenant_keys(tenant_id);
```

### New (v2-reserved, unused in v1): `vault_user_keys`

```sql
CREATE TABLE vault_user_keys (
  tenant_id    UUID         NOT NULL,
  user_id      UUID         NOT NULL,
  public_key   BYTEA        NOT NULL,         -- X25519 pubkey (v2)
  key_type     TEXT         NOT NULL DEFAULT 'x25519',
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id)
);
```

### Modified: `dynamic_secret_targets`

```sql
ALTER TABLE dynamic_secret_targets
  ADD COLUMN encryption_mode      TEXT  NOT NULL DEFAULT 'server',  -- 'server' | 'client'
  ADD COLUMN key_version          INT   NOT NULL DEFAULT 1,        -- 1 for server, 3 for client
  ADD COLUMN wrap_iv              BYTEA,                            -- IV for AES-GCM in client mode
  ADD COLUMN wrap_auth_tag        BYTEA,                            -- GCM tag in client mode
  ADD COLUMN namespace            TEXT NOT NULL DEFAULT '/',        -- for RBAC scoping
  ADD CONSTRAINT chk_encryption_mode CHECK (encryption_mode IN ('server', 'client'));
```

For `encryption_mode='server'` (legacy): `password_nonce` continues to carry the IV; `wrap_iv` and `wrap_auth_tag` are NULL. Existing rows are unchanged.

For `encryption_mode='client'` (new): `wrap_iv` + `wrap_auth_tag` + `encrypted_admin_password` carry the new envelope. `password_nonce` is NULL.

### New: `dynamic_wrapped_access_tokens`

```sql
CREATE TABLE dynamic_wrapped_access_tokens (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID         NOT NULL,
  credential_id   UUID         NOT NULL,                  -- DynamicCredential.ID
  token_hash      TEXT         NOT NULL UNIQUE,           -- SHA-256(raw), shown once
  name            TEXT         NOT NULL,
  scopes          JSONB        NOT NULL DEFAULT '[]',
  expires_at      TIMESTAMPTZ  NOT NULL,
  is_revoked      BOOLEAN      NOT NULL DEFAULT false,
  revoked_at      TIMESTAMPTZ,
  revoked_reason  TEXT,
  allowed_ips     JSONB        NOT NULL DEFAULT '[]',
  denied_ips      JSONB        NOT NULL DEFAULT '[]',
  ip_restriction_enabled BOOLEAN NOT NULL DEFAULT false,
  last_used_at    TIMESTAMPTZ,
  use_count       INT          NOT NULL DEFAULT 0,
  created_by      UUID         NOT NULL,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_dwat_tenant_credential ON dynamic_wrapped_access_tokens(tenant_id, credential_id);
```

Mirrors the static-secret `AccessToken` model (`internal/storage/vault/models.go:137-198`) but scoped to a dynamic credential template. The raw token is returned exactly once at mint time.

### New: `dynamic_target_shares`

```sql
CREATE TABLE dynamic_target_shares (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  target_id            UUID         NOT NULL,
  source_tenant_id     UUID         NOT NULL,
  granted_to_tenant_id UUID         NOT NULL,
  granted_by_user      UUID         NOT NULL,
  permissions          TEXT         NOT NULL DEFAULT 'read',  -- 'read' | 'admin'
  expires_at           TIMESTAMPTZ,
  revoked_at           TIMESTAMPTZ,
  revoked_by           UUID,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_dts_grantee_active ON dynamic_target_shares(granted_to_tenant_id)
  WHERE revoked_at IS NULL;
CREATE INDEX idx_dts_target ON dynamic_target_shares(target_id);
```

### New audit actions (extend `internal/storage/vault/types.go:37-74`)

```go
AuditActionClientUnwrapFailed   = "client_unwrap_failed"
AuditActionClientWrapCreate     = "client_wrap_create"
AuditActionClientWrapIssue      = "client_wrap_issue"
AuditActionClientWrapRotate     = "client_wrap_rotate"
AuditActionClientDekInit        = "client_dek_init"
AuditActionClientDekShare       = "client_dek_share"
AuditActionDYNTokenCreate       = "dynamic_token_create"
AuditActionDYNTokenRevoke       = "dynamic_token_revoke"
AuditActionDYNTokenUse          = "dynamic_token_use"
```

### Migration: timestamp-format SQL file

A new migration `migrations/20260616220000_vault_client_encrypted_dynamic_creds.up.sql` is created (using the existing helper `./scripts/create-migration.sh` per AGENTS.md). Idempotent SQL with `IF NOT EXISTS` clauses. The migration is additive — no destructive changes.

---

## API Surface

### New endpoints (all under `/v1/vault/...`)

| Method | Path | Purpose | Auth |
|---|---|---|---|
| `GET` | `/keys/wrapped-dek` | Return the current user's `(wrapped_dek, iv, tag, salt, key_version, kdf_params)` or 404 if not yet initialized. | JWT |
| `POST` | `/keys/wrapped-dek` | Create or rotate the current user's wrapped DEK. Body: `{wrapped_dek, iv, tag, salt, key_version, kdf_params}`. Server stores opaquely. | JWT + MFA if `EnforceForAPI` |
| `POST` | `/keys/wrapped-dek/rotate` | Rotate the DEK (generate a new DEK, re-wrap all `encryption_mode='client'` rows for this user). Admin can also call this on behalf of a target user. | JWT + MFA + `dynamic_credentials:client_manage_keys` |
| `POST` | `/keys/share` | Admin re-wraps the per-tenant DEK for a target user. Body: `{target_user_id, wrapped_dek, dek_iv, dek_auth_tag, dek_salt, key_version, kdf_params}`. | JWT + MFA + `rbac:manage` (or built-in admin) |
| `POST` | `/keys/cross-tenant-wrap` | Admin re-wraps a target's admin password under a different tenant's DEK. Body: `{target_id, target_tenant_wrapped_admin_password, dest_tenant_dek_id}`. Returns the re-wrapped admin password blob. | JWT + MFA + `shares:manage` |
| `GET` | `/dynamic-secret-targets/{id}/wrapped` | Return the wrapped admin password ciphertext (no plaintext). Client uses this to unwrap locally. | JWT + `dynamic_credentials:read` + MFA |
| `POST` | `/dynamic-tokens` | Mint a `dynamic_wrapped_access_tokens` row. Body: `{credential_id, name, scopes, expires_in_hours, ip_policy?}`. Response: `{id, token (raw, shown once), expires_at, ...}`. | JWT + `dynamic_credentials:token_mint` + MFA |
| `GET` | `/dynamic-tokens` | List tokens (hash-only, like static-secret `AccessToken`). | JWT + `audit:read` |
| `DELETE` | `/dynamic-tokens/{id}` | Revoke. | JWT + `dynamic_credentials:token_mint` + MFA |
| `GET` | `/dynamic-tokens/{id}/usage` | Usage stats (last_used_at, use_count). | JWT + `audit:read` |

### Modified endpoints

| Method | Path | Change |
|---|---|---|
| `POST` | `/dynamic-secret-targets` | Body accepts `encryption_mode: 'server' \| 'client'`. For `client` mode: body must include `wrapped_admin_password, wrap_iv, wrap_auth_tag, key_version`. Plaintext `admin_password` accepted only for `server` mode. Returns the same shape as today. |
| `POST` | `/dynamic-credentials/{id}/generate` | For `encryption_mode='client'` targets: body must include `target_admin_password` (single-use plaintext, server zeroizes) + `new_db_password` (client-generated). For `server` mode: behavior unchanged. |
| `POST` | `/dynamic-credentials/{id}/leases/{lease_id}/renew` | For `client`-mode targets: body must include `target_admin_password` (single-use). Server uses it for `ALTER USER ... VALID UNTIL ...`. |
| `POST` | `/dynamic-credentials/{id}/leases/{lease_id}/revoke` | For `client`-mode targets: body must include `target_admin_password` (single-use). Server uses it for `DROP USER`. |
| `POST` | `/dynamic-credentials/{id}/revoke` (revoke all leases) | Same as above. |
| `POST` | `/vault/escrow` (existing) | Body now includes the DEK-wrapped escrow blob. Server stores opaquely. Backwards-compat: legacy escrow blobs are migrated to the new format on next successful vault unlock. |
| `POST` | `/dynamic-credentials/{id}/generate` | New authentication path: accepts `Authorization: Bearer ff_dyn_<token>` (a `dynamic_wrapped_access_tokens` raw value) in addition to JWT. Server hashes, looks up, validates scopes + IP policy + expiry. |

### New auth middleware: `RequireVaultMFA`

A new helper `internal/api/middleware/vault_mfa.go`:

```go
// RequireVaultMFA returns 401 with X-Vault-MFA-Required if the tenant
// has VaultMFAConfig.EnforceForAPI=true and the user has not performed
// MFA within the configured MFASessionTTLSeconds. Active break-glass
// requests bypass this check.
func RequireVaultMFA(repo *vault.Repository) func(http.Handler) http.Handler
```

Wrapped around every new and modified dynamic-credential route, mirroring the existing `authMiddleware.RequireAuth` pattern.

### RBAC enforcement wiring

`Handler.RBAC.Check(...)` is called in every new and modified dynamic-credential handler. This **closes** the pre-existing gap (current dynamic-cred handlers do not check RBAC; the engine is built but unused). Failures return 403 with `apierror.ErrorCode("VAULT_PERMISSION_DENIED")`.

---

## Issuance Flow

### Setup (one time per user per tenant)

1. User logs in, opens dashboard, accepts `VaultSetupDialog` (`'unlock'` mode — same component used for static secrets today, `web/dashboard/src/components/api-keys/VaultSetupDialog.tsx:29-222`).
2. Dashboard calls `GET /v1/vault/keys/wrapped-dek`.
   - **404** (no DEK yet): dashboard generates a fresh DEK (`crypto.getRandomValues(32)`), wraps it under the user's KEK (`Argon2id(passphrase, dek_salt)`), and POSTs to `/keys/wrapped-dek`. The DEK plaintext is held in browser memory only.
   - **200** (DEK exists): dashboard unwraps the DEK using the user's passphrase and holds it in memory.
3. Dashboard caches the DEK in a `Map<tenantID, CryptoKey>` in memory for the tab's lifetime. Mirrors the existing `setVaultPassphrase` session cache pattern at `web/dashboard/src/services/vault-api-key-storage.ts:114-140`.

### Create a target (client mode)

1. User fills in target form (name, host, port, db, admin user, admin password).
2. Client generates a random `target_id` (UUID v4) for the AAD computation.
3. Client encrypts `admin_password` with the DEK, AAD = `client-wrap:‹tenantID›:‹target_id›`.
4. Client POSTs:
   ```json
   {
     "name": "production-pg",
     "host": "db.example.com",
     "port": 5432,
     "database_name": "appdb",
     "admin_username": "admin",
     "encryption_mode": "client",
     "wrapped_admin_password": "<base64>",
     "wrap_iv": "<base64>",
     "wrap_auth_tag": "<base64>",
     "key_version": 3,
     "default_ttl_seconds": 3600,
     "max_ttl_seconds": 86400
   }
   ```
5. Server validates, persists, generates a real `target_id`, and returns the row. **Server cannot unwrap.**
6. Audit row: `AuditActionClientWrapCreate`.

### Generate a credential (the moment of truth)

1. User clicks "Generate" on a dynamic credential template.
2. Client calls `GET /v1/vault/dynamic-secret-targets/{id}/wrapped` to fetch the wrapped admin password.
3. Client decrypts admin password locally: `admin_pw = AES-GCM-decrypt(wrapped, DEK, aad="client-wrap:‹tenantID›:‹targetID›")`. AEAD tag mismatch → throws, audit row `AuditActionClientUnwrapFailed`, returns "Invalid target: corruption detected" to the user.
4. Client generates new DB user password: 24 chars from the 56-char alphabet (`crypto.getRandomValues`).
5. Client POSTs to `/generate`:
   ```json
   {
     "target_admin_password": "<plaintext, single-use>",
     "new_db_username": "vault_p_<random>",
     "new_db_password": "<plaintext, single-use>",
     "ttl_seconds": 3600
   }
   ```
6. Server authenticates via JWT, validates inputs, opens a `*sql.DB` connection to the customer DB using the provided admin password, runs `CREATE USER ... WITH PASSWORD ...` and `GRANT ... TO ...` in a transaction, closes the connection, **explicitly zeroizes the in-memory copy** of `target_admin_password` and `new_db_password`, and creates a `DynamicCredentialLease` row.
7. Server returns the lease:
   ```json
   {
     "lease_id": "lease_xxxx",
     "username": "vault_p_xxxx",
     "password": "<plaintext, single-use, for the new DB user>",
     "host": "...",
     "port": 5432,
     "database": "appdb",
     "expires_at": "2026-06-16T18:00:00Z"
   }
   ```
8. Client stores the new password in agent memory / browser memory for the application's use. The server never persists it. (v1 mirrors today's plaintext-over-TLS response; v2 will wrap with `response_key`.)
9. Audit row: `AuditActionClientWrapIssue` with `target_id`, `credential_id`, `lease_id`, `encryption_mode`, `aad_context_hash`, `kdf_key_version`. **No password material.**

### Renew / Revoke

Same flow: client must provide the unwrapped admin password to the server (because the server must run `ALTER USER ... VALID UNTIL ...` or `DROP USER`). The client unwraps the target's admin password, sends it in a one-shot request, server executes, zeros memory.

### Test target (`HandleTestTarget`)

Client unwraps admin password, sends it, server opens a 60s test connection, immediately revokes, returns success. Same pattern as today.

### CI/CD agent (headless)

The agent binary `functionfly-vault-agent` is a small Go program (~5 MB) shipped at `cmd/vault-agent/`. Workflow:

1. **Enrollment** (one time per agent):
   ```
   functionfly-vault-agent enroll \
     --server=https://api.functionfly.com \
     --enrollment-token=ff_enr_...
   ```
   The enrollment token is a one-time, single-use token generated from the dashboard (Settings → Agents). The agent POSTs the token to `/v1/vault/agents/enroll`, receives the wrapped DEK, prompts for the vault passphrase (or reads it from a `VAULT_PASSPHRASE` env var — discouraged), unwraps the DEK locally, and stores it in the OS keychain (`keyring.Set(tenantID, "vault-dek", hex.EncodeToString(dek))`).
2. **Usage** (per request from the application):
   - Application: `curl http://localhost:8090/v1/creds/<credential_id>` (agent exposes a local HTTP proxy on port 8090 by default).
   - Agent: fetches the wrapped target admin password, unwraps it using the DEK from keychain, calls `/generate`, returns the new lease material.
   - Application: uses the lease, then optionally calls `POST /v1/creds/<credential_id>/renew` or `/revoke`.
3. **Optional env-var injection mode**: `--inject-env -- cmd...` — the agent spawns the child process with the lease material as env vars (`VAULT_DB_USER`, `VAULT_DB_PASSWORD`, etc.) and never exposes them via HTTP.

The agent uses the same `dynamic_wrapped_access_tokens` mint flow for authentication: a token is created in the dashboard, copied to the agent's config, and used as `Authorization: Bearer ff_dyn_<token>`.

---

## Client Implementation

### Dashboard (browser)

New code in `web/dashboard/src/`:

- `utils/vault-crypto.ts` (existing) — extend with `wrapDEK(dek, passphrase, salt)`, `unwrapDEK(wrapped, passphrase, salt)`, `wrapAdminPassword(admin_pw, dek, targetID, tenantID)`, `unwrapAdminPassword(wrapped, dek, targetID, tenantID)`. The existing `encryptWithPassphrase` / `decryptWithPassphrase` (lines 267-293) are reused for the DEK wrap.
- `hooks/useVaultTenantKey.ts` (new) — React hook managing the per-tenant DEK in browser memory, including setup, rotation, share, error paths. Backed by a new `Map<tenantID, CryptoKey>` cache.
- `hooks/useDynamicCredentialActions.ts` (new) — wraps the unwrap-encrypt-decrypt flow for the dynamic creds UI.
- `components/VaultEnterprise/tabs/DynamicCredsTab.tsx` (existing) — extend:
  - Per-target badge: "🔒 Client-encrypted" or "🔓 Server-encrypted (legacy)".
  - Replace the `alert(password)` (lines 304-306) with a copy-to-clipboard + "show password" toggle (with auto-hide after 30s) + warning.
  - Block creation of `client`-mode targets if no DEK is set up (redirect to vault setup).
- `components/VaultEnterprise/tabs/DynamicTargetsTab.tsx` (existing) — extend with a "Migrate to client-encrypted" wizard for server-mode targets.
- `components/api-keys/VaultSetupDialog.tsx` (existing) — extend the unlock flow to fetch and cache the DEK after passphrase verification.

### Headless agent (Go binary: `functionfly-vault-agent`)

New code:

- `cmd/vault-agent/main.go` (new) — Cobra CLI with subcommands `enroll`, `serve`, `version`. Flags: `--server`, `--enrollment-token`, `--bind=127.0.0.1:8090`, `--cache-ttl=15m`, `--log-level=info`, `--inject-env`.
- `sdk/go-vault-sdk/agent/` (new package) — small library wrapping the existing `go-vault-sdk` (`sdk/go-vault-sdk/`) with:
  - `agent.UnwrapDEK(ctx, wrapped, passphrase) (dek, error)`
  - `agent.IssueCredential(ctx, dek, credentialID, ttl) (lease, error)`
  - `agent.RenewLease(ctx, dek, leaseID, ttl) error`
  - `agent.RevokeLease(ctx, dek, leaseID) error`
- OS keychain integration: `github.com/zalando/go-keyring` (Linux/macOS/Windows).

The agent binary is built via the existing `make build-all-modules` and shipped as a standalone download on the marketing site (`web/site/`).

### Build & release

- `cmd/vault-agent/` is a separate Go module added to `go.work` (mirroring `cmd/delete-functions/`).
- CI: `make build-all-modules` produces `bin/functionfly-vault-agent-{linux,darwin,windows}-{amd64,arm64}`.
- Release: GitHub Releases with a checksums file, mirrored on the FunctionFly downloads page.

---

## Account-Type Coverage

| Account type | Works? | Notes |
|---|---|---|
| Personal account (free) | ✅ | Free tier: `maxDynamicCreds=100`, postgres-only. Uses the same flow. |
| Starter ($24/mo) | ✅ | Inherits free vault tier. |
| Professional ($79/mo) | ✅ | Vault Pro: postgres, 5,000 leases. |
| Enterprise ($299/mo) | ✅ | Vault Team: postgres + mysql, 50,000 leases. |
| Agent Enterprise ($499/mo) | ✅ | Vault Enterprise: all features, 1,000,000 leases. |
| Multi-user team (any plan) | ✅ | Admin re-wraps the per-tenant DEK for new members via `/keys/share`. |
| SSO-enforced tenant (Agent Enterprise) | ✅ | Same flow, audit row includes `saml_asserted_user`. |
| MFA-enforced tenant (Pro+) | ✅ | `EnforceForAPI` gates the new endpoints. |
| CI/CD agent user | ✅ | Agent holds DEK in OS keychain. |
| Break-glass emergency access | ✅ | Active `BreakGlassRequest` allows issuance during the window. |
| Existing tenant with server-mode targets | ✅ | Server-mode targets continue to work. Migration is opt-in. |

---

## RBAC Integration

### New permission strings

Add to `internal/storage/vault/rbac.go` `BuiltinRolePermissions`:

| Permission | Description |
|---|---|
| `dynamic_credentials:client_unwrap` | Can unwrap a target's admin password client-side and submit it in a one-shot request. Required for `generate`/`renew`/`revoke` on `client`-mode targets. |
| `dynamic_credentials:client_manage_keys` | Can rotate the tenant DEK, re-wrap for new users, manage the DEK lifecycle. Admin-only. |
| `dynamic_credentials:token_mint` | Can mint `dynamic_wrapped_access_tokens` (CI/CD agent tokens). |
| `dynamic_credentials:read` | Can read target metadata + the wrapped admin password (for client-side unwrap). |

### Updated built-in roles

```go
// Admin — gets all new permissions
admin: { ... existing ..., dynamic_credentials:client_unwrap,
         dynamic_credentials:client_manage_keys,
         dynamic_credentials:token_mint,
         dynamic_credentials:read }

// Operator — gets client_unwrap and token_mint, NOT client_manage_keys
operator: { ... existing ..., dynamic_credentials:client_unwrap,
            dynamic_credentials:token_mint,
            dynamic_credentials:read }

// Reader — unchanged; readers still get dynamic_credentials:generate per the existing model
reader: { ..., dynamic_credentials:generate, dynamic_credentials:read }
```

### Enforcement wiring (closes the existing gap)

Every new and modified dynamic-credential handler calls `h.RBAC.Check(ctx, claims.TenantID, claims.UserID, "<perm>", namespacePath)`. Failures return 403 with `apierror.ErrorCode("VAULT_PERMISSION_DENIED")`. The pre-existing gap (current dynamic-cred handlers do not call `RBAC.Check`) is closed as part of this change.

### Namespace support

`DynamicSecretTarget` gets an optional `namespace` column (default `"/"`). New requests accept `namespace` in the body; handlers pass it to `RBAC.Check(..., namespacePath)`. The existing `vault_namespaces` table is reused. Existing static-secret `Secret.namespace` gap (per the integration report) is **not** closed by this design — that's a separate cleanup.

---

## MFA / Break-Glass Integration

### MFA gating

All new and modified dynamic-credential endpoints honor `vault_mfa_config.EnforceForAPI`. The new `RequireVaultMFA` middleware:

1. Reads the tenant's `VaultMFAConfig`. If `EnforceForAPI=false`, pass through.
2. Queries `secrets_audit_log` for the most recent `Action=AuditActionMFAVerify` row for this `(user_id, tenant_id)` within `MFASessionTTLSeconds`. If absent, return 401 with `MFA_SESSION_EXPIRED` and a header `X-Vault-MFA-Required: true`.
3. If the user has an active `BreakGlassRequest`, bypass and log `break_glass_id` in the audit metadata for the operation.

The existing `POST /v1/vault/mfa/verify` flow is reused — no new endpoint.

### Break-glass

Active `BreakGlassRequest` is checked in the same middleware path. The `AuditActionBreakGlass` row is queried for `(tenant_id, user_id, status='approved', expires_at > now())`. If present, MFA is bypassed for the window. All operations performed under break-glass include `break_glass_id` in the audit metadata.

### Plan gating

MFA is gated by `SupportsVaultMFA(plan)` (Pro/Enterprise/Agent Enterprise). For tenants on Free/Starter, `VaultMFAConfig` is never created (lazy-create returns nil), and the middleware passes through.

---

## Escrow Integration

### Updated escrow blob structure

The escrow blob in `vault_escrow_config` (`internal/storage/vault/security_models.go:120-151`) stores `EncryptedBlob`, `BlobIV`, `BlobAuthTag`, `BlobKeyVersion`. The blob contents evolve:

| Field | v1 (static secrets only) | v2 (this change) |
|---|---|---|
| Blob contents | Optional wrapped KEK (opaque to server) | Wrapped **DEK** (the new per-tenant DEK) |
| KDF | Security-question-derived KEK, Argon2id | Same — security-question-derived KEK, Argon2id |
| Recovery flow | Unwrap KEK → derive static-secret key → decrypt secrets | Unwrap DEK → re-wrap under new passphrase → decrypt secrets AND dynamic target admin passwords |

**Why wrap the DEK (not a separate dynamic-cred key)?**
- One escrow blob per tenant. Simpler ops, simpler UX.
- Recovery is a single, well-known flow.
- The DEK is the unit of recovery: anything the DEK encrypts is recovered together.

### Escrow enable flow

1. User enables escrow via the dashboard.
2. Dashboard derives a KEK from the security questions: `KEK = Argon2id(concat(answers...), escrow_salt)` using the existing Argon2id params (`security.go:518-522`).
3. Dashboard unwraps the in-memory DEK.
4. Dashboard re-wraps the DEK under the security-question KEK: `escrow_blob = AES-256-GCM(DEK, KEK, iv, aad="escrow:v2")`.
5. Dashboard POSTs `{escrow_blob, blob_iv, blob_auth_tag, blob_key_version: 2, kdf_salt, kdf_params, security_question_hashes}`.
6. Server stores opaquely.

### Escrow recovery flow

1. User answers security questions on the recovery page.
2. Server looks up the blob and returns it (it cannot decrypt).
3. Browser derives the KEK from the answers, decrypts the DEK.
4. User enters a new vault passphrase. Browser re-wraps the DEK under the new passphrase's KEK and POSTs to `/keys/wrapped-dek`.
5. The new wrapped DEK replaces the old row. The old row is left in place for audit (or hard-deleted, per security policy).

### Backwards compat

If the existing blob is detected (legacy format with KEK instead of DEK), the dashboard's vault-unlock flow attempts to:
1. Unwrap the legacy KEK from the blob.
2. Unwrap the user's existing static-secret encryption key from the legacy KEK.
3. Decrypt all `Secret.EncryptedValue` rows with the unwrapped key.
4. Re-encrypt all rows under the new DEK.
5. Write the new escrow blob (DEK under security-question KEK).
6. Delete the legacy blob.

This migration is best-effort: if any step fails, the user is informed and the legacy blob is preserved.

### Free/Starter plan customers

Escrow is gated by `SupportsVaultEscrow(plan)` (Enterprise/Agent Enterprise). Free/Starter/Pro customers who lose their passphrase lose their static secrets AND their client-mode dynamic targets. **This is the same risk as today for static secrets (no escrow = no recovery).** Documentation must be explicit about this. The dashboard's vault setup wizard shows a "Recovery options" section that surfaces this limitation.

---

## Audit + SIEM Integration

### New audit actions

```go
AuditActionClientUnwrapFailed   = "client_unwrap_failed"
AuditActionClientWrapCreate     = "client_wrap_create"
AuditActionClientWrapIssue      = "client_wrap_issue"
AuditActionClientWrapRotate     = "client_wrap_rotate"
AuditActionClientDekInit        = "client_dek_init"
AuditActionClientDekShare       = "client_dek_share"
AuditActionDYNTokenCreate       = "dynamic_token_create"
AuditActionDYNTokenRevoke       = "dynamic_token_revoke"
AuditActionDYNTokenUse          = "dynamic_token_use"
```

### Audit metadata conventions

For `client_wrap_issue` (a successful issuance):
```json
{
  "operation": "client_wrap_issue",
  "target_id": "<uuid>",
  "credential_id": "<uuid>",
  "lease_id": "<lease_xxxx>",
  "encryption_mode": "client",
  "aad_context_hash": "<sha256 of AAD string>",
  "kdf_key_version": 3,
  "ip": "<ip>",
  "auth_method": "jwt" | "dyn_token"
}
```

For `client_unwrap_failed`:
```json
{
  "operation": "client_unwrap_failed",
  "target_id": "<uuid>",
  "reason": "aead_tag_mismatch" | "iv_length_invalid" | "tag_length_invalid",
  "kdf_key_version": 3
}
```

**Never** log the plaintext password, the DEK, the KEK, the wrapped blob, the salt, the IV, or any unwrapped material. The `aad_context_hash` is a SHA-256 of the AAD string (which is non-secret but useful for correlating events).

### SIEM wiring

The existing `SIEMDispatcher` is built (`internal/api/handlers/vault/handler.go:44`) but never called from any handler. The new `logAudit` helper in `internal/api/handlers/vault/security.go:578-604` is extended to call `s.Dispatch(ctx, audit)` for every row, gated by `s.SIEM != nil` and an env var `VAULT_SIEM_ENABLED=true` (default: true). This single change ships SIEM delivery for both new and existing audit actions. Webhook signing keys continue to be per-tenant.

---

## Sharing / Collaboration

### Cross-tenant target shares

Use the `dynamic_target_shares` table. The share flow:

1. Source tenant admin: dashboard → target → "Share with tenant" → enter destination tenant ID + permissions.
2. Source admin's client fetches the destination tenant's per-user wrapped DEK (admin must be a member of both tenants, or use a server-side `cross-tenant-wrap` endpoint that re-wraps without the server seeing the DEK).
3. Source admin's client re-wraps the target's admin password under the destination tenant's DEK. **Server proxies the re-wrap and never sees the plaintext.**
4. Server creates a `dynamic_target_shares` row + a "shadow" target row in the destination tenant (or a `shared_target_id` pointer).
5. Destination tenant users can now issue leases on the shared target by unwrapping the re-wrapped admin password with their DEK.

**Key property:** shares are *target-level*, not *lease-level*. The grantee can generate new leases on the target; the grantee cannot see existing leases of the source tenant. Matches the principle of least privilege.

### Same-tenant DEK sharing (multi-user)

1. Source user: dashboard → settings → "Share vault access with user" → enter target user.
2. Source user's client re-wraps the per-tenant DEK under the target user's KEK (uses `/v1/vault/keys/share`).
3. Server creates a new `vault_tenant_keys` row for the target user.
4. Target user can now unwrap the DEK and use all client-mode targets + static secrets in the tenant.

### Plan gating

Both flows are gated by `SupportsVaultShares(plan)` (Enterprise/Agent Enterprise). Free/Pro tenants can still have multiple users — but they all derive the DEK from their own passphrase, and the admin must re-wrap for new members. The "share" UX is Enterprise-only for clarity.

---

## Rate Limiting

Add a per-`(user, target, IP)` rate limiter for the issuance endpoints (`generate`, `renew`, `revoke`):

- **10 requests/minute** per `(user, target)`.
- **100 requests/minute** per `IP` (across all users).
- **1000 requests/hour** per `(user, target)` (catches longer-running abuse).

Backed by Redis (existing `REDIS_ADDR` env var). Implemented as a new middleware `internal/api/middleware/vault_ratelimit.go`. Returns 429 with `Retry-After` header.

The static-secret `/secrets` endpoints already have `vaultRateLimiter` (`internal/api/handlers/vault/handler.go` references); we extend the same pattern.

---

## Migration & Backwards Compatibility

### New tenants

- Default: `client` mode for new dynamic credential targets.
- Free-tier tenants can opt out per-target (a "Use server encryption" checkbox in the target form) — useful for legacy integrations that don't want to manage a vault passphrase.

### Existing tenants

- Existing `dynamic_secret_targets` rows are **unchanged** (`encryption_mode='server'`).
- New targets default to `client` mode.
- A migration wizard (`DynamicTargetsTab.tsx`) allows admins to bulk-migrate targets to `client` mode by re-entering the admin password once per target. The wizard is opt-in.

### What breaks and what doesn't

| Surface | Before | After | Backwards compat |
|---|---|---|---|
| `POST /v1/vault/dynamic-secret-targets` body `admin_password: string` | Plaintext | **Modified**: accepts `admin_password` (server mode) OR `wrapped_admin_password + wrap_iv + wrap_auth_tag` (client mode) + `encryption_mode` | ✅ Existing clients still work for server mode |
| `POST /v1/vault/dynamic-credentials/{id}/generate` body | Empty (server generates everything) | **Modified**: for client-mode targets, requires `target_admin_password` + `new_db_password` in body. For server mode, unchanged. | ✅ |
| All other dynamic-cred endpoints | n/a | **Modified**: pass-through for server-mode targets; client-mode requires `target_admin_password` in body. | ✅ |
| Static secret endpoints | n/a | **Unchanged** | ✅ |
| New DEK / token endpoints | n/a | New | n/a (additive) |
| New audit actions | n/a | New | ✅ Existing audit consumers ignore unknown actions |

### Deprecation timeline

- **v1 (this PR):** `server` mode is documented as legacy. No deprecation warnings.
- **v2 (Q1 2027):** add a "rotate to client mode" wizard; deprecate `server` mode for new tenants on Free/Starter; encourage migration.
- **v3 (Q3 2027):** remove `server` mode entirely. Existing server-mode targets are force-migrated using break-glass. Migration tool re-wraps the admin password using the existing target secret + the user's DEK (requires the user to re-enter the admin password).

---

## Testing Strategy

### Unit tests (Go)

- `internal/crypto/server_encryption_test.go` (existing) — kept; extended with new helpers.
- `internal/storage/vault/dynamic_encryption_test.go` (new) — client-wrap/unwrap, AAD binding, error paths (wrong DEK, wrong AAD, tampered ciphertext, reused nonce detection).
- `internal/storage/vault/dynamic_service_test.go` (new) — client-mode issuance (success, failure, AAD mismatch, expiry).
- `internal/api/handlers/vault/dynamic_test.go` (new) — handler-level integration: server-mode fallback, client-mode flow, MFA gating, RBAC enforcement, audit logging, rate limiting, zeroization.
- `internal/api/middleware/vault_mfa_test.go` (new) — MFA gating, break-glass bypass, MFASessionTTL enforcement.
- `internal/api/middleware/vault_ratelimit_test.go` (new) — rate limit enforcement, Redis-backed counters.
- `cmd/vault-agent/agent_test.go` (new) — agent enrollment, DEK unwrap, keychain integration (mocked), issuance, renew, revoke.

### Unit tests (TS / Vitest)

- `web/dashboard/src/utils/vault-crypto.test.ts` (existing) — extended with DEK wrap/unwrap helpers, AAD binding, error paths.
- `web/dashboard/src/hooks/useVaultTenantKey.test.ts` (new) — setup, rotate, share, error paths.
- `web/dashboard/src/components/VaultEnterprise/tabs/DynamicCredsTab.test.tsx` (new) — target creation (both modes), generate, revoke, renew, copy-to-clipboard UX.
- `web/dashboard/src/components/VaultEnterprise/tabs/DynamicTargetsTab.test.tsx` (new) — migration wizard flows.

### Integration tests (Go)

- `internal/storage/vault/dynamic_integration_test.go` (new) — uses testcontainers-go to spin up Postgres + MySQL, exercises the full issuance → renew → revoke → sweeper cycle for both `server` and `client` modes.
- `internal/storage/vault/agent_e2e_test.go` (new) — spins up the agent binary, enrolls, generates a credential, verifies the local HTTP proxy works, verifies env-var injection mode.

### End-to-end tests (Playwright)

- `web/dashboard/e2e/dynamic-credentials.spec.ts` (new) — user logs in, creates a target, generates a credential, copies the password, verifies the new DB user can connect.
- `web/dashboard/e2e/dynamic-credentials-mfa.spec.ts` (new) — same flow with MFA enforced; verifies the MFA challenge flow.
- `web/dashboard/e2e/dynamic-credentials-rbac.spec.ts` (new) — operator role can issue but not manage keys; reader role is denied.
- `web/dashboard/e2e/dynamic-credentials-share.spec.ts` (new) — admin shares DEK with another user; second user can issue.

### Security tests

A dedicated suite that verifies:

- Server cannot unwrap a client-encrypted target admin password even with DB read access + `SERVER_MASTER_KEY`.
- Wrong AAD fails decryption.
- Tampered ciphertext fails decryption.
- Reused nonce is detected (we generate a fresh IV per wrap; this test enforces the invariant).
- DEK rotation invalidates old wrapped rows.
- Passphrase rotation re-wraps all rows.
- The server-side test simulates a compromised server (DB + `SERVER_MASTER_KEY` leaked) and asserts that target admin passwords are not recoverable.

### Test data

- Reuses existing test fixture for tenants, users, plans.
- Adds `createTestTargetWithMode(t, encryptionMode)` helper.
- Adds `unwrapTestTarget(t, targetID, dek)` helper.
- Adds `createTestAgent(t, tenantID, userID) *agent.Client` helper.

---

## Rollout Plan

### Phase 1 — Internal (this PR)

- Ship all code, migrations, dashboard UI, agent binary.
- Internal tenants only (FunctionFly staff).
- Both `server` and `client` modes supported; new tenants default to `client`.
- Monitor: wrap/unwrap failures, AAD mismatches, MFA gating bypasses, RBAC denials, rate-limit hits.

### Phase 2 — Limited beta (Pro + Enterprise)

- Enable for Pro and Enterprise tenants via feature flag `vault.client_encryption.enabled`.
- Collect metrics: wrap/unwrap latency, error rates, agent binary feedback.
- Iterate on UX (copy-to-clipboard, auto-hide, agent DX).

### Phase 3 — GA (all plans)

- Enable for all tenants.
- Update documentation.
- Announce in changelog and on the marketing site.

### Phase 4 — Migration tooling

- Build the "rotate to client mode" wizard.
- Begin encouraging migration of legacy server-mode targets.
- Set a deprecation date for `server` mode (v2 timeframe).

### Phase 5 — Deprecation (v2, ~6-9 months out)

- Default new tenants on Free/Starter to `client` only (no opt-out).
- Document the deprecation timeline.
- Provide bulk-migration tooling for Enterprise customers.

### Rollback

If v1 ships and a critical issue is discovered, the rollback is straightforward:

1. Feature flag `vault.client_encryption.enabled=false` disables new `client`-mode target creation.
2. Existing `client`-mode targets continue to work (the new flow is self-contained).
3. If the new flow is itself broken: feature flag `vault.client_encryption.issuance_enabled=false` reverts issuance to server-side (requires the user to re-enter the admin password, but the wrapper DEK can be unwrapped by the server using a **newly-introduced** `vault_server_recovery_key` env var that is generated only as a rollback mechanism — see `internal/crypto/server_recovery.go`).

The `vault_server_recovery_key` is **not** generated by default. It is generated by an explicit operator command (`make gen-recovery-key`) and stored in a separate secrets manager. It is never read by the application code unless the feature flag `vault.client_encryption.issuance_enabled=false` is set AND an emergency override is granted. This is the explicit, auditable emergency escape hatch.

---

## Open Questions / Future Work

1. **Asymmetric key wrapping (v2):** per-user Ed25519 + X25519 keys for YubiKey / hardware token support. The `vault_user_keys` table is reserved for this.
2. **Response wrapping (v2):** wrap the `/generate` response with a `response_key` so the lease material is never in plaintext over TLS. Useful for paranoid customers.
3. **Direct agent-to-DB connectivity (v2):** the agent makes a direct outbound connection to the customer DB, removing the server from the trust path entirely. Matches HashiCorp Vault Agent.
4. **Per-lease sharing (v2):** today, shares are target-level. v2 could allow sharing individual leases with a third party (e.g. a contractor) for a bounded time.
5. **M-of-N escrow (v2):** current escrow is 1-of-1 (single user, single security-question set). v2 could add M-of-N (multiple users, multi-sig recovery) for high-security tenants.
6. **Per-tenant DEK rotation policy (v2):** today, DEK rotation is manual. v2 could add automatic rotation (e.g. every 90 days) with a grace period for re-wrapping.
7. **Server-side anomaly detection (v2):** today, the audit log is the only signal. v2 could add ML-based anomaly detection (e.g. unusual issuance rates, unusual times, unusual IPs).

---

## Appendix: File-Line Index

| Topic | File | Lines |
|---|---|---|
| Target / template / lease GORM models | `internal/storage/vault/dynamic_models.go` | 30-131 |
| Server-side envelope (Argon2id + AES-GCM) | `internal/crypto/server_encryption.go` | 28-150 |
| Encrypt / decrypt of admin password | `internal/storage/vault/dynamic_encryption.go` | 15-45 |
| Issuance / renewal / revocation service | `internal/storage/vault/dynamic_service.go` | 106-236 |
| Postgres + MySQL DDL | `internal/storage/vault/dynamic_backends.go` | 25-251 |
| Expired-lease sweeper | `internal/storage/vault/dynamic_sweeper.go` | 50-114 |
| Sweeper wiring in API server | `internal/api/server.go` | 720-769 |
| HTTP handlers | `internal/api/handlers/vault/dynamic.go` | 22-537 |
| Route registration | `internal/api/routes_platform.go` | 486-499 |
| Request / response DTOs | `internal/api/handlers/vault/types.go` | 427-505 |
| Frontend TanStack hooks | `web/dashboard/src/api/vault.ts` | 237-346 |
| Frontend types | `web/dashboard/src/types/vault-enterprise.ts` | 165-254 |
| Frontend UI (incl. `alert(password)`) | `web/dashboard/src/components/VaultEnterprise/tabs/DynamicCredsTab.tsx` | 65-389 |
| Plan limits (server) | `internal/plans/limits.go` | 48-52, 1125-1139, 1157-1193 |
| Plan limits (dashboard) | `web/dashboard/src/lib/vaultPlans.ts` | 24-155 |
| Quota enforcement | `internal/storage/vault/quota/quota.go` | 152-158 |
| Quota store | `internal/storage/vault/quota_store.go` | 69-125 |
| RBAC permission map | `internal/storage/vault/rbac.go` | 22-67 |
| Client crypto class (zero-knowledge) | `web/dashboard/src/utils/vault-crypto.ts` | 55-319 |
| Static-secret model (zero-knowledge) | `internal/storage/vault/models.go` | 12-67 |
| Static-secret create handler | `internal/api/handlers/vault/secrets.go` | 31-159 |
| Escrow (zero-knowledge recovery blob) | `internal/api/handlers/vault/security.go` | 471-535 |
| Vault setup dialog (client) | `web/dashboard/src/components/api-keys/VaultSetupDialog.tsx` | 29-222 |
| Vault service layer | `web/dashboard/src/services/vault-api-key-storage.ts` | 114-140 |
| Audit log table | `internal/storage/vault/models.go` | 203-241 |
| Audit actions enum | `internal/storage/vault/types.go` | 37-74 |
| SIEM dispatcher (built, not wired) | `internal/storage/vault/audit_export.go` | 250-309 |
| Namespace model | `internal/storage/vault/enterprise_models.go` | 14-46 |
| Share model | `internal/storage/vault/enterprise_models.go` | 111-145 |
| SSO config model | `internal/storage/vault/enterprise_models.go` | 148-174 |
| MFA config model | `internal/storage/vault/security_models.go` | 39-57 |
| Break-glass models | `internal/storage/vault/security_models.go` | 59-118 |
| Escrow config model | `internal/storage/vault/security_models.go` | 120-151 |
| AccessToken model | `internal/storage/vault/models.go` | 137-198 |
| Token minting flow | `internal/api/handlers/vault/tokens.go` | 18-120 |
