# Tenant Isolation for Bundle Provisioning

**Date:** 2026-07-01  
**Status:** Draft  
**Scope:** Full implementation — core fixes, maintenance cache, provisioning status, health endpoint, graceful degradation

---

## Problem

The billing handler's `ProvisionBundleAppAndBackend` in `internal/api/handlers/billing/bundles_provisioning.go` creates app, backend, and function resources directly in the **shared platform database**, bypassing the existing `BundleProvisioner` which already provisions isolated tenant databases.

### Observed errors (from production logs)

1. **Backend region overflow**: `pq: value too long for type character varying(10)` — `backends.region` is `VARCHAR(10)` but "eu-central-1" is 12 chars
2. **Missing providers**: `pq: null value in column "providers" of relation "functions" violates not-null constraint` — bundle function templates don't set `Providers` on `FunctionConfig`
3. **Slow INSERT**: 253ms for `INSERT INTO apps` (audit trigger + RLS policy evaluation)
4. **No tenant isolation**: Bundle users share the platform DB, Redis, and State Fabric with all other tenants

### Existing infrastructure (already built, not wired)

- `TenantDBProvisioner` creates `functionfly_tenant_{uuid[:8]}` databases from templates
- `TenantPoolManager` manages per-tenant connection pools with health monitoring
- `BundleProvisioner` orchestrates: UserDB → Auth → Payments → Email → Analytics
- Tenant-specific migrations exist for SaaS, Marketplace, and AI bundles
- Redis keys and State Fabric cache keys already include `tenantID`

---

## Design

### 1. Schema fix: widen `backends.region`

**Migration:** `20260702000000_widen_backend_region.up.sql`

```sql
ALTER TABLE backends ALTER COLUMN region TYPE VARCHAR(100);
```

Defense-in-depth regardless of which provisioning path is used.

### 2. Wire billing handler to `BundleProvisioner`

**Current state (partially wired):** The `BundleProvisioner` is already injected via `SetBundleProvisioner` (`routes.go:1184`) and the per-bundle provisioning methods (`provisionSaaSStarter`, `provisionMarketplace`, `provisionAIApp`) already call `h.provisionBundleFn` asynchronously. However, `ProvisionBundleAppAndBackend` (called from `bundles.go:830` and `stripe.go:679`) ALSO creates app, backend, and function records in the shared platform DB unconditionally — causing the observed errors.

**Current flow (two parallel paths):**
```
Stripe webhook / Deploy handler
  → ProvisionBundleAppAndBackend (shared DB) ← ERRORS HERE
    → repo.CreateApp (platform DB)
    → repo.CreateBackend (platform DB, fails on region varchar(10))
    → repo.CreateFunction (platform DB, fails on providers NOT NULL)
  → provisionBundleFn (async goroutine) ← WORKS but runs alongside broken path
    → BundleProvisioner.ProvisionBundle()
    → UserDBProvisioner → Auth → Payments → Email → Analytics
  → provisionSaaSStarter/provisionMarketplace/provisionAIApp
    → ALSO creates function templates in shared DB (without providers)
    → ALSO calls provisionBundleFn (duplicate!)
```

**New flow (single path, graceful degradation):**
```
Stripe webhook / Deploy handler
  → ProvisionBundleAppAndBackend (platform DB: only creates app record for routing)
    → repo.CreateApp (platform DB — needed for edge routing)
    → Skip backend/functions in shared DB when BundleProvisioner is available
  → BundleProvisioner.ProvisionBundle() (dedicated DB)
    → UserDBProvisioner (creates dedicated tenant DB)
    → AuthProvisioner (JWT keys, OAuth in tenant DB)
    → PaymentsProvisioner (Stripe customer, products in tenant DB)
    → EmailWorkflowProvisioner (templates in tenant DB)
    → AnalyticsProvisioner (dashboards in tenant DB)
    → Writes deployment_steps for real-time status
  → If dedicated DB fails: fall back to shared-DB provisioning with degraded_mode flag
```

The billing handler already has access to `BundleProvisioner` via the `BillingAdapter` (wired in `routes.go:1183-1184`).

### 3. Fix function template `providers` field

Even though the old path becomes a fallback, fix it so it works:

- Add `providers` field to `bundleFunctionTemplate` struct
- Set `Providers: []string{"functionfly"}` on each template
- Set `Providers` on the `FunctionConfig` in the provisioning loop

### 4. Cache maintenance check

The `maintenance_repository.go:40` query runs every 2-3 seconds:
```sql
SELECT * FROM platform_maintenance WHERE enabled = true ORDER BY id LIMIT 1
```

Add an in-memory cache with 30-second TTL:

```go
type MaintenanceRepository struct {
    db          *PostgresDB
    cachedResult *PlatformMaintenance
    cacheExpiry  time.Time
    mu           sync.RWMutex
}

func (r *MaintenanceRepository) GetActiveMaintenance(ctx context.Context) (*PlatformMaintenance, error) {
    r.mu.RLock()
    if r.cachedResult != nil && time.Now().Before(r.cacheExpiry) {
        defer r.mu.RUnlock()
        return r.cachedResult, nil
    }
    r.mu.RUnlock()
    
    // DB query
    r.mu.Lock()
    defer r.mu.Unlock()
    // Double-check after acquiring write lock
    if r.cachedResult != nil && time.Now().Before(r.cacheExpiry) {
        return r.cachedResult, nil
    }
    // Query DB, update cache
    r.cachedResult = result
    r.cacheExpiry = time.Now().Add(30 * time.Second)
    return result, nil
}
```

### 5. Provisioning status visibility

Add a `deployment_steps` table to track real-time provisioning progress:

```sql
CREATE TABLE deployment_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    bundle_slug VARCHAR(100) NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

Each sub-provisioner writes a step record before starting and updates on completion. The existing `/v1/billing/deployments/{id}/status` endpoint reads from this table.

### 6. Graceful degradation

If `CreateTenantDB` fails:

1. Log the error with full context
2. Fall back to shared-DB provisioning (the old path, with bugs fixed)
3. Set `tenant.degraded_mode = true` flag on the tenant record
4. Queue a retry job (exponential backoff, max 3 attempts)
5. Notify the user in the dashboard: "Your workspace is using shared infrastructure. Dedicated setup will be retried."

Add column to tenants table:
```sql
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS degraded_mode BOOLEAN DEFAULT false;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS degradation_reason TEXT;
```

### 7. Tenant health endpoint

`GET /api/tenant/health` returns:

```json
{
  "database": {
    "type": "dedicated",
    "status": "healthy",
    "latency_ms": 2
  },
  "state_fabric": {
    "status": "healthy",
    "fabrics_count": 3
  },
  "redis": {
    "status": "healthy",
    "latency_ms": 1
  },
  "degraded_mode": false
}
```

Implementation in `internal/api/handlers/tenant/health.go`:
- Check dedicated DB connectivity via `TenantPoolManager.GetPoolHealth(tenantID)`
- Check Redis with `PING`
- Check State Fabric via `StateFabricRepository.ListFabrics(tenantID)`
- Return degraded_mode from tenant record

---

## Tenant resource isolation matrix

| Resource | Isolation method | Location |
|----------|-----------------|----------|
| **PostgreSQL database** | Dedicated `functionfly_tenant_{uuid[:8]}` | TenantDBProvisioner |
| **State Fabric data** | `tenant_state_entries` table in tenant DB | tenant_migrations/tenant_base_schema |
| **Redis keys** | Prefixed with `tenantID` | Shared Redis, logically isolated |
| **Auth credentials** | JWT keys + OAuth configs in tenant DB | AuthProvisioner |
| **Payment data** | Stripe customer + subscriptions in tenant DB | PaymentsProvisioner |
| **Email templates** | Templates + workflows in tenant DB | EmailWorkflowProvisioner |
| **Analytics** | Dashboards + events in tenant DB | AnalyticsProvisioner |
| **SAR runtime** | Shared, state scoped via tenant_id | NATS + edge adapters |

---

## Files to modify

| File | Change |
|------|--------|
| `migrations/20260702000000_widen_backend_region.up.sql` | New migration: widen region to VARCHAR(100) |
| `migrations/20260702000000_widen_backend_region.down.sql` | Rollback migration |
| `migrations/20260702000100_deployment_steps.up.sql` | New: deployment_steps table + tenant degraded_mode columns |
| `migrations/20260702000100_deployment_steps.down.sql` | Rollback |
| `internal/api/handlers/billing/bundles_provisioning.go` | Wire to BundleProvisioner, fix providers field, add graceful degradation |
| `internal/storage/maintenance_repository.go` | Add in-memory cache with 30s TTL |
| `internal/provisioning/bundle_provisioner.go` | Write deployment_steps records per stage |
| `internal/provisioning/userdb_provisioner.go` | Add graceful degradation fallback |
| `internal/api/handlers/tenant/health.go` | New: tenant health endpoint |
| `internal/api/routes.go` | Register health endpoint |
| `internal/storage/tenant_repository.go` | Add degraded_mode fields |

---

## Non-goals

- Dedicated Redis instances per tenant (shared Redis with key prefixing is sufficient)
- Dedicated SAR runtime per tenant (shared with tenant-scoped state)
- Network-level isolation (VPC segmentation) — future enterprise feature
- Per-tenant backup scheduling — future feature

---

## Risks and mitigations

| Risk | Mitigation |
|------|-----------|
| Template DB doesn't exist on first provision | `CreateTenantDB` falls back to empty DB + runs migrations |
| Tenant pool exhaustion under load | Pool manager enforces min/max per tenant + health monitoring |
| Provisioning takes too long | Each step has timeout; status endpoint shows real-time progress |
| Shared Redis becomes noisy-neighbor | Monitor per-tenant key counts; future: dedicated Redis SELECT per tenant |
