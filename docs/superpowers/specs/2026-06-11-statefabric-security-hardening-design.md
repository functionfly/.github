# Security Hardening: State Fabric & API - Implementation Design

**Date:** 2026-06-11
**Status:** Draft
**Scope:** 10 security issues across `internal/storage/statefabric/` and `internal/api/`

---

## Overview

This design covers fixes for 10 security issues identified in State Fabric storage and admin API handlers. Issues range from simple validation fixes to architectural changes requiring Vault integration.

---

## Issue Summary & Fix Strategy

| # | Issue | Severity | Fix Approach | Complexity |
|---|-------|----------|--------------|------------|
| 1 | R2 Credentials in Env | High | Vault integration + short-lived creds | High |
| 2 | Admin No Rate Limiting | High | Add rate limiter middleware | Medium |
| 3 | No TLS Verification | High | Configure HTTP client TLS | Low |
| 4 | R2 Deletion Silent Failure | Medium | Transaction + retry queue | Medium |
| 5 | GitHub Vault Key Hardcoded | Medium | Fail startup if missing | Low |
| 6 | Snapshot No Size Limit | Medium | Add max size validation | Low |
| 7 | Replay Async No Access Control | Medium | Capture perms before spawn | Medium |
| 8 | No Encryption at Rest | Medium | Field-level encryption | High |
| 9 | Event List No Limit Cap | Medium | Cap limit at 1000 | Low |
| 10 | Store Type Not Validated | Low | Add allowlist validation | Low |

---

## Detailed Designs

### 1. R2 Storage Backend - Vault Integration

**Current State:** Credentials loaded from `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY` env vars.

**Proposed State:**
```
┌─────────────────────────────────────────────────────────────┐
│  R2 Storage Config                                          │
│  ┌─────────────┐    ┌──────────────┐    ┌───────────────┐ │
│  │   Vault     │───▶│  KVV2 Path   │───▶│  R2 Config    │ │
│  │  (primary)  │    │  /secrets/   │    │  struct       │ │
│  └─────────────┘    │  r2-creds    │    └───────────────┘ │
│         │           └──────────────┘                       │
│         │               │ Fallback                          │
│         ▼               ▼                                   │
│  ┌─────────────┐    ┌──────────────┐                     │
│  │   Env Var   │───▶│  Validation  │                     │
│  │  (dev-only) │    │  + Warning    │                     │
│  └─────────────┘    └──────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Add `R2StorageConfigLoader` interface with `Load(ctx) (*R2StorageConfig, error)`
- Create `VaultR2ConfigLoader` that reads from Vault KVV2 at path `secret/data/statefabric/r2`
- Create `EnvR2ConfigLoader` (dev-only) that reads from env vars with validation
- Add `USE_VAULT_FOR_R2` env var to toggle (default: `true` in production)
- For short-lived creds: Vault can issue temporary R2 tokens; loader refreshes on expiry
- Log all credential access for audit trail

**Files Changed:**
- `internal/storage/statefabric/r2_storage.go` - Add loader interface, refactor `NewR2StorageBackend`
- `internal/storage/statefabric/vault.go` (new) - Vault config loader

---

### 2. Admin Endpoints - Rate Limiting

**Current State:** Admin endpoints check permissions but have no rate limits.

**Proposed State:**
```
┌─────────────────────────────────────────────────────────────┐
│  Admin Handler Middleware Chain                              │
│  ┌──────────┐  ┌────────────┐  ┌────────────────────────┐  │
│  │  Auth    │─▶│ Rate Limit │─▶│  Admin Handler         │  │
│  │ Middleware│  │ (per-IP)   │  │  (hasAdminPermission)   │  │
│  └──────────┘  └────────────┘  └────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Add `RateLimiter` interface to `internal/api/middleware/`
- Implement token bucket rate limiter (existing infrastructure may have one)
- Create `AdminRateLimitConfig`:
  - `HandleRunTTLCleanup`: 1 req/min (expensive operation)
  - Other admin endpoints: 60 req/min
- Add `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
- Return `429 Too Many Requests` when exceeded
- Log rate limit violations for abuse detection

**Files Changed:**
- `internal/api/middleware/ratelimit.go` - Add rate limiter (if not exists)
- `internal/api/handlers/statefabric/handler_admin.go` - Wrap handlers with rate limiter
- `internal/api/routes.go` - Register middleware

---

### 3. TLS Verification for Pipeline Execution

**Current State:** HTTP client for pipeline execution uses default transport (no explicit TLS config).

**Proposed State:**
```go
// Repository initialization
transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        // In production, consider certificate pinning
        // For internal services, system CA pool is acceptable
    },
}
client := &http.Client{
    Transport: transport,
    Timeout: 30 * time.Second,
}
```

**Implementation:**
- Add `tls.Config` with `MinVersion: tls.VersionTLS12`
- For production with internal services: use system CA pool
- For external services: add certificate pinning via env var
- Add `PIPELINE_TLS_SKIP_VERIFY` env var (default: `false`, only for dev)

**Files Changed:**
- `internal/storage/statefabric/repository.go` - Configure TLS in `NewRepository`

---

### 4. Cleanup Service - R2 Deletion Failure Handling

**Current State:** R2 deletion failure logs warning but proceeds to delete DB record.

**Proposed State:**
```
┌─────────────────────────────────────────────────────────────┐
│  Cleanup Transaction Flow                                    │
│                                                              │
│  1. Mark snapshot as PENDING_DELETION in DB                 │
│  2. Delete R2 object                                       │
│     ├─ Success: Delete DB record (commit)                   │
│     └─ Failure: Mark FAILED, schedule retry                │
│                  (do NOT delete DB record)                  │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Add `Status` field to `StateFabricSnapshot`: `pending_deletion`, `deleting`, `failed_deletion`
- Use database transaction:
  ```sql
  BEGIN;
  UPDATE snapshots SET status = 'deleting' WHERE id = ?;
  -- R2 delete happens here
  -- If R2 fails: ROLLBACK (status stays 'deleting' for retry)
  -- If R2 succeeds: DELETE FROM snapshots WHERE id = ?;
  COMMIT;
  ```
- Add retry mechanism: failed deletions retried up to 3 times with exponential backoff
- Add `cleanup_retry_queue` table to track failed deletions
- Add metrics: `cleanup_r2_deletion_failures_total`

**Files Changed:**
- `internal/storage/statefabric/cleanup.go` - Transaction + retry logic
- `migrations/` - Add status column + retry queue table

---

### 5. GitHub Vault Key - Remove Hardcoded Fallback

**Current State:** Falls back to `default-dev-key-must-be-32-bytes!` when `GITHUB_VAULT_KEY` is empty in dev mode.

**Proposed State:**
- Fail startup if `GITHUB_VAULT_KEY` is not set (regardless of DEVELOPMENT mode)
- Remove hardcoded fallback entirely
- Add validation: key must be 32 bytes AND not a known weak key pattern

**Files Changed:**
- `internal/api/routes.go` - Remove fallback, add strict validation

---

### 6. Snapshot Creation - Size Limit

**Current State:** No absolute size limit on snapshot data.

**Proposed State:**
```go
const MaxSnapshotSize = 1 * 1024 * 1024 * 1024 // 1GB

func (r *Repository) CreateSnapshot(...) {
    if stateDataSize > MaxSnapshotSize {
        return nil, fmt.Errorf("snapshot size exceeds maximum allowed size of %d bytes", MaxSnapshotSize)
    }
}
```

**Files Changed:**
- `internal/storage/statefabric/repository.go` - Add size check in `CreateSnapshot`

---

### 7. Replay Execution - Permission Capture

**Current State:** `go r.executeReplay()` runs asynchronously; permissions verified at spawn time.

**Proposed State:**
```
┌─────────────────────────────────────────────────────────────┐
│  Replay Request Flow                                         │
│                                                              │
│  1. API receives replay request                             │
│  2. Capture caller permissions into replay context           │
│  3. Spawn async replay with captured context                 │
│  4. Replay handler verifies captured permissions             │
│     └─ If permissions insufficient: reject immediately      │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Create `ReplayContext` struct containing:
  - `TenantID`, `UserID`, `Permissions`, `Role`, `SpawnedAt`
- Pass `ReplayContext` to `executeReplay()` instead of relying on context
- Verify permissions at replay execution time against captured context
- Add timeout for replay execution (prevent stuck replays)

**Files Changed:**
- `internal/storage/statefabric/repository.go` - Capture context, verify at execution

---

### 8. No Encryption at Rest

**Current State:** State data stored in PostgreSQL without application-level encryption.

**Proposed State:**
- Implement field-level encryption for sensitive state data
- Use AES-256-GCM (same as Vault encryption)
- Encrypt fields: `state_data`, `snapshot_data` in application layer
- Key from Vault (aligned with R2 credential management)

**Note:** This is a large architectural change. For MVP:
- Encrypt `state_data` column using envelope encryption
- Snapshot data already in R2 (which uses R2's encryption at rest)

**Files Changed:**
- `internal/storage/statefabric/` - Add encryption service
- `internal/storage/statefabric/crypto.go` (new) - Encryption utilities

---

### 9. Event List - Limit Cap

**Current State:** `ListEvents` accepts arbitrary limit without maximum cap.

**Proposed State:**
```go
const MaxEventListLimit = 1000

func (r *Repository) ListEvents(...) {
    if limit > MaxEventListLimit {
        limit = MaxEventListLimit
    }
    // Also add query timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
}
```

**Files Changed:**
- `internal/storage/statefabric/repository.go` - Cap limit, add timeout

---

### 10. Store Type Validation

**Current State:** Unknown fabric types silently default to `keyvalue`.

**Proposed State:**
```go
var allowedFabricTypes = map[string]string{
    "catalog":   "document",
    "workflow":  "timeseries",
    "custom":    "graph",
}

func validateFabricType(fabricType string) error {
    if _, ok := allowedFabricTypes[fabricType]; !ok {
        return fmt.Errorf("unknown fabric type: %s", fabricType)
    }
    return nil
}
```

**Files Changed:**
- `internal/storage/statefabric/repository.go` - Add validation in `CreateFabric`

---

## Implementation Order

```
Phase 1 (Low Complexity, High Impact)
├── 5. GitHub Vault Key Fallback Removal
├── 3. TLS Configuration
├── 6. Snapshot Size Limit
├── 9. Event List Limit Cap
└── 10. Store Type Validation

Phase 2 (Medium Complexity)
├── 2. Admin Rate Limiting
├── 4. R2 Deletion Retry Mechanism
└── 7. Replay Permission Capture

Phase 3 (High Complexity - Architectural)
├── 1. Vault Integration for R2
└── 8. Encryption at Rest
```

---

## Testing Strategy

| Issue | Test Approach |
|-------|---------------|
| 1. Vault R2 | Unit test loader interface, integration test with Vault |
| 2. Rate Limiting | Integration test with multiple rapid requests |
| 3. TLS | Unit test TLS config generation |
| 4. R2 Deletion | Integration test: fail R2 delete, verify DB record preserved |
| 5. Vault Key | Unit test: missing env var fails startup |
| 6. Snapshot Size | Unit test: large snapshot rejected |
| 7. Replay Perms | Unit test: verify captured perms used at execution |
| 8. Encryption | Unit test: encrypt/decrypt roundtrip |
| 9. Event Limit | Unit test: limit > 1000 capped |
| 10. Store Type | Unit test: unknown type rejected |

---

## Rollback Plan

- Issues 3, 5, 6, 9, 10: Simple validation/addition - rollback by reverting code changes
- Issues 1, 2, 4, 7, 8: Feature flags to disable in case of issues
  - `FEATURE_VAULT_R2=true/false`
  - `FEATURE_ADMIN_RATE_LIMIT=true/false`
  - `FEATURE_R2_DELETION_RETRY=true/false`
  - `FEATURE_REPLAY_PERM_CAPTURE=true/false`
  - `FEATURE_ENCRYPTION_AT_REST=true/false`

---

## Files Summary

| File | Changes |
|------|---------|
| `internal/storage/statefabric/r2_storage.go` | Vault loader interface, refactor |
| `internal/storage/statefabric/vault.go` | New - Vault config loader |
| `internal/storage/statefabric/repository.go` | TLS, size limit, event cap, store type, replay context |
| `internal/storage/statefabric/cleanup.go` | Transaction + retry |
| `internal/storage/statefabric/crypto.go` | New - Encryption utilities |
| `internal/api/middleware/ratelimit.go` | New or existing rate limiter |
| `internal/api/handlers/statefabric/handler_admin.go` | Rate limit wrapper |
| `internal/api/routes.go` | Vault key validation |
| `migrations/` | Snapshot status, retry queue table |
