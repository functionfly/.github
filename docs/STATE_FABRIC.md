# State Fabric

Zero-knowledge state management service built into the FunctionFly orchestrator. Provides multi-tenant state stores, event sourcing, snapshots, replays, and triggers for serverless functions.

## Overview

State Fabric lets tenants create isolated **fabrics** (state containers) that hold key-value state, emit ordered event logs, support point-in-time snapshots with replay, and trigger function execution on state changes.

### Core concepts

| Concept | Description |
|---------|-------------|
| **Fabric** | A tenant-owned state container. One per use-case. |
| **Store** | A storage backend attached to a fabric (`queue`, `persistent`, `memory`). |
| **Pipeline** | A trigger-driven function execution chain (stored as a `state_trigger` with `target_function`). |
| **Event** | An append-only mutation record (`set`, `delete`, `snapshot`, `restore`, `merge`). |
| **Snapshot** | Point-in-time capture of all keys in a fabric. |
| **Replay** | Re-applies events from a snapshot to reconstruct state. |
| **Trigger** | Fires a function on state change (`on_write`, with key pattern matching). |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  HTTP API (gorilla/mux, /v1/state-fabrics/*)       │
├─────────────────────────────────────────────────────┤
│  Handler layer  (internal/api/handlers/statefabric/)│
│   ├─ handler.go         CRUD + metrics + events     │
│   ├─ handler_triggers.go triggers + pipelines       │
│   ├─ handler_admin.go   admin endpoints             │
│   ├─ handler_snapshots.go                            │
│   ├─ permissions.go     quota + add-on + perms      │
├─────────────────────────────────────────────────────┤
│  Repository  (internal/storage/statefabric/)        │
│   ├─ repository.go      fabric CRUD, stores, etc.   │
│   ├─ cache.go           Redis-backed read cache     │
│   ├─ models.go          GORM models                 │
├─────────────────────────────────────────────────────┤
│  State store  (internal/storage/state/)             │
│   ├─ state.go           fabric (states) table       │
│   ├─ state_values.go    key-value rows              │
│   ├─ state_events.go    event log                   │
│   ├─ state_snapshots.go snapshot rows               │
│   ├─ state_triggers.go  trigger rows                │
│   ├─ state_permissions.go fabric-level RBAC         │
├─────────────────────────────────────────────────────┤
│  PostgreSQL (primary) + Redis (cache, optional)     │
└─────────────────────────────────────────────────────┘
```

All persistence goes through GORM on the `states`, `state_values`, `state_events`, `state_snapshots`, `state_triggers`, and `state_permissions` tables. Redis is used opportunistically for metrics/list caching.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/state-fabrics/health` | none | Liveness check |
| `GET` | `/v1/state-fabrics/ready` | none | Readiness (R2 backend must be configured) |
| `GET` | `/v1/state-fabrics/feature-flags` | none | Feature flag catalog |
| `GET` | `/v1/state-fabrics` | tenant | List fabrics |
| `POST` | `/v1/state-fabrics` | tenant | Create fabric |
| `GET` | `/v1/state-fabrics/{id}` | tenant | Get fabric |
| `PATCH` | `/v1/state-fabrics/{id}` | tenant | Update fabric |
| `DELETE` | `/v1/state-fabrics/{id}` | tenant | Delete fabric |
| `GET` | `/v1/state-fabrics/{id}/metrics` | tenant + `advanced_insights` add-on | Fabric metrics |
| `GET/POST` | `/v1/state-fabrics/{id}/stores` | tenant | Stores sub-resource |
| `DELETE` | `/v1/state-fabrics/{id}/stores/{storeId}` | tenant | Delete store |
| `GET/POST` | `/v1/state-fabrics/{id}/pipelines` | tenant | Pipelines (stored as triggers) |
| `PATCH/DELETE` | `/v1/state-fabrics/{id}/pipelines/{pipelineId}` | tenant | Update/delete pipeline |
| `POST` | `/v1/state-fabrics/{id}/pipelines/{pipelineId}/execute` | tenant | Execute pipeline |
| `GET/POST` | `/v1/state-fabrics/{id}/snapshots` | tenant | Snapshots |
| `DELETE` | `/v1/state-fabrics/{id}/snapshots/{snapshotId}` | tenant | Delete snapshot |
| `GET` | `/v1/state-fabrics/{id}/events` | tenant + `advanced_security_pack` add-on | Event log |
| `GET/POST` | `/v1/state-fabrics/{id}/replays` | tenant + `hot_cache_booster` add-on | Replays |
| `GET` | `/v1/state-fabrics/{id}/replays/{replayId}` | tenant | Get replay status |
| `GET` | `/v1/state-fabrics/{id}/replays/{replayId}/progress` | none | Replay progress (SSE) |
| `GET/POST` | `/v1/state-fabrics/{id}/triggers` | tenant + fabric RBAC | Triggers |
| `PATCH/DELETE` | `/v1/state-fabrics/{id}/triggers/{triggerId}` | tenant + fabric RBAC | Update/delete trigger |
| `GET` | `/v1/admin/state-fabrics` | admin (`tenants:read`) | List all fabrics (admin) |
| `GET` | `/v1/admin/state-fabrics/stats` | admin | Aggregate stats |
| `GET` | `/v1/admin/state-fabrics/settings` | admin | Global settings |
| `POST` | `/v1/admin/state-fabrics/cleanup` | admin | Run TTL cleanup |
| `GET` | `/v1/admin/state-fabrics/cleanup/stats` | admin | Cleanup stats |

## Plans & Quotas

Defined in `internal/plans/limits.go`:

| Plan | Max Fabrics | Feature |
|------|-------------|---------|
| Free | 0 | not available |
| Starter | 1 | included |
| Professional | 10 | included |
| Enterprise | unlimited (-1) | included |
| Agent Enterprise | unlimited | included |

Add-ons (required for some endpoints):

| Add-on | Grants access to |
|--------|-----------------|
| `advanced_insights` | `/metrics` |
| `advanced_security_pack` | `/events` |
| `hot_cache_booster` | `/replays` |
| `ai_memory_pack` | `vector`/`embedding`/`ai-memory` store types |

Add-on entitlements live in `state_fabric_addon_entitlements` (tenant_id, addon_id, status).

## Testing

### Unit tests

```bash
go test -count=1 -short ./internal/api/handlers/statefabric/...
go test -count=1 -short ./internal/plans/...
DB_SSLMODE=disable go test -count=1 -short ./internal/storage/statefabric/...
```

### Smoke test

End-to-end HTTP smoke test: **`scripts/statefabric_smoke_test.sh`**

```bash
bash scripts/statefabric_smoke_test.sh
```

Covers 35 checks across 14 sections:

1. Health & readiness (unauthenticated)
2. Authentication & token extraction
3. Create fabric (CRUD)
4. Get / list fabrics
5. Update fabric
6. Metrics (requires `advanced_insights` add-on)
7. Stores sub-resource (list/create/delete)
8. Pipelines sub-resource (list/create/update/delete)
9. Snapshots sub-resource (list/create/delete)
10. Events sub-resource (requires `advanced_security_pack` add-on)
11. Replays sub-resource (requires `hot_cache_booster` add-on)
12. Triggers sub-resource
13. Delete fabric (cleanup)
14. Admin endpoints

### Stress test

Concurrent load generator: **`scripts/statefabric_stress.go`**

```bash
DURATION=30s WORKERS=20 go run scripts/statefabric_stress.go
```

Mixed read/write workload with weighted phases (health, list, create, feature-flags, admin stats). Reports throughput, success rate, p50/p95/p99/max latency, and status code distribution. Fails if server error rate exceeds 5%.

### Pre-prod validation history (2026-06-16)

35/35 smoke tests passing, 0.1% server error rate under stress. Bugs found and fixed in this round:

| # | Bug | File | Severity |
|---|-----|------|----------|
| 1 | `PlanHasStateFabricFeature` returned `false` for Enterprise (unlimited = -1, not > 0) | `internal/plans/limits.go:1040` | **Critical** — blocks all Enterprise tenants |
| 2 | `CreateReplay` did not set `TenantID`; `StateFabricReplay` model missing `TenantID` | `internal/storage/statefabric/repository.go:1643`, `models.go:155` | **Critical** — replay creation fails |
| 3 | `recordEventTX` did not set `CorrelationID` (NOT NULL in schema) | `internal/storage/state/state_events.go:30` | **Critical** — all snapshot/event creation fails |
| 4 | `fabricToAPI` panicked with "assignment to entry in nil map" on unmarshal failure | `internal/api/handlers/statefabric/handler.go:151` | **Critical** — admin list all crashes server |
| 5 | `requireFabricPermission` had no admin role bypass | `internal/api/handlers/statefabric/permissions.go:13` | **High** — admin users get 403 on trigger list |
| 6 | Route `/state-fabrics/{id}` matched before `/state-fabrics/health`/`/ready`/`/feature-flags` | `internal/api/routes_platform.go:413-442` | **High** — health endpoints return 404 |
| 7 | Test schema `createTestTables` out of sync with GORM models (missing `state_values`, wrong `state_triggers`, wrong `state_snapshots` columns, duplicate `state_size_bytes`) | `internal/storage/statefabric/repository_test.go` | **High** — all storage tests were broken |
| 8 | Missing DB columns: `state_snapshots.r2_object_key/r2_bucket/r2_content_hash`, `state_events.r2_object_key/r2_bucket/batch_id/is_archived/archived_at` | DB schema | **Critical** — inserts fail in production |

#### Stress test results (dev environment, 20 workers, 30s)

| Metric | Value |
|--------|-------|
| Total requests | 8,728 |
| Throughput | 290.9 req/s |
| Success (2xx) | 51.1% (rest are 429 rate-limited — expected) |
| Server errors (5xx) | **0.1%** (10/8728) |
| p50 latency | 33ms |
| p95 latency | 246ms |
| p99 latency | 472ms |
| Max latency | 957ms |
| Health under load | All checks <1ms |

Rate limiting is working correctly — 48.8% of requests received 429 under sustained 20-worker load. This is the AdvancedRateLimit middleware doing its job.

## Known limitations / future work

- **R2 storage** must be configured via env (`R2_*`) for `/ready` to return 200 and for large snapshot offloading (>100KB).
- **Empty response bodies** on create/list/get — the `writeJSON` helper serializes the payload but some clients see `Content-Length: 0`. Investigation pending (likely a middleware interaction with `responseWriterTracker.Flush`).
- **Pipelines stored as triggers** — `CreatePipeline` inserts into `state_triggers` with `IsActive: false` and a `condition` JSONB containing steps. This is by design but means pipeline/trigger share a table.
- **Stores stored on fabric metadata** — `CreateStore` updates `states.storage_type`, `states.max_size_mb`, and `states.tags` rather than inserting into `state_fabric_stores`. The `state_fabric_stores` table exists for future per-store records.
- **Replay runs async** in a goroutine; clients must poll `GET /replays/{id}` for status.
- **Follow-up:** move `internal/wasm/state_fabric_handler.go` to the new module (see `docs/followups/move-state-fabric-handler.md`).

## Environment variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `R2_*` | R2/S3-compatible storage for snapshot offloading | unset (disabled) |
| `REDIS_ADDR` | Redis for metrics/list cache | `localhost:6379` |
| `DEVELOPMENT` | Dev mode (bypasses IP allowlist, CSRF, etc.) | `false` |
| `SKIP_MIGRATION_VALIDATION` | Required for `--skip-migrations` mode | `false` |
