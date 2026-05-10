# Function Time Machine — Architecture Plan

**Status:** Architecture Draft  
**Date:** 2026-05-02  
**Author:** System Architecture (FunctionFly)

---

## 1. Executive Summary

**Function Time Machine** allows developers to retroactively fix bugs in production as if they never happened. A developer fixes the function code, selects a time window, and the platform re-executes every real request through the corrected function, compares old vs new outputs, automatically reconciles state, and generates a compliance-grade audit certificate.

### Why This Is Possible on FunctionFly (and Nowhere Else)

FunctionFly already has:
- **Cryptographic execution certificates** (FXCERT + MEG records) for every execution
- **Full execution history** with inputs, outputs, timestamps, and state changes in `registry_function_executions` + `registry_executions_public`
- **Deterministic Replay Execution (DRE)** infrastructure that re-executes functions and compares root hashes
- **State Fabric replay system** with snapshot-based and event-range-based replay
- **FRG event sourcing** with checkpoints and state restoration
- **Webhook replay infrastructure** (30-day stored payloads with replay tracking)

Time Machine unifies these into a single product experience.

---

## 2. System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Dashboard (React)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Time      │  │ Diff     │  │ Replay   │  │ Audit Certificate │  │
│  │ Machine   │  │ Viewer   │  │ Progress │  │ Viewer            │  │
│  │ Page      │  │          │  │          │  │                   │  │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └────────┬──────────┘  │
│        │             │             │                 │              │
└────────┼─────────────┼─────────────┼─────────────────┼──────────────┘
         │             │             │                 │
    ─────┼─────────────┼─────────────┼─────────────────┼──── REST API
         │             │             │                 │
┌────────┼─────────────┼─────────────┼─────────────────┼──────────────┐
│  Go API Layer                                                  │
│  ┌─────┴────┐  ┌─────┴────┐  ┌─────┴────┐  ┌────────┴──────────┐  │
│  │ /time-   │  │ /time-   │  │ /time-   │  │ /time-machine/    │  │
│  │ machine/ │  │ machine/ │  │ machine/ │  │ audit-certificates│  │
│  │ replays  │  │ diff     │  │ status   │  │                   │  │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └────────┬──────────┘  │
│        │             │             │                 │              │
│  ┌─────┴─────────────┴─────────────┴─────────────────┴──────────┐  │
│  │              Time Machine Service Layer                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐    │  │
│  │  │ Replay   │ │ Diff     │ │ Reconcil │ │ Audit Cert   │    │  │
│  │  │ Engine   │ │ Engine   │ │ iation   │ │ Generator    │    │  │
│  │  │          │ │          │ │ Engine   │ │              │    │  │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────┬───────┘    │  │
│  └───────┼─────────────┼───────────┼───────────────┼────────────┘  │
│          │             │           │               │               │
│  ┌───────┴─────────────┴───────────┴───────────────┴────────────┐  │
│  │                    Storage Layer                              │  │
│  │  ┌────────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │  │
│  │  │ time_      │ │ time_    │ │ time_    │ │ time_        │  │  │
│  │  │ machine_   │ │ machine_ │ │ machine_ │ │ machine_     │  │  │
│  │  │ replays    │ │ diffs    │ │ reconcil │ │ audit_       │  │  │
│  │  │            │ │          │ │ ations   │ │ certificates │  │  │
│  │  └────────────┘ └──────────┘ └──────────┘ └──────────────┘  │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                 Replay Worker Pool                           │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │  │
│  │  │ Worker 1 │ │ Worker 2 │ │ Worker 3 │ │ Worker N │       │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │  │
│  │         In-Memory Priority Queue (existing infrastructure)  │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
         │
    ─────┼──── Existing Infrastructure
         │
┌────────┼───────────────────────────────────────────────────────────┐
│  ┌─────┴──────────┐  ┌──────────────┐  ┌────────────────────┐     │
│  │ PostgreSQL     │  │ Redis        │  │ Execution Runtime  │     │
│  │ (execution     │  │ (progress    │  │ (sandbox executor  │     │
│  │  history +     │  │  tracking,   │  │  re-uses existing  │     │
│  │  TM tables)    │  │  rate limits)│  │  pipeline)         │     │
│  └────────────────┘  └──────────────┘  └────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
```

---

## 3. Plan Integration & Feature Gating

### 3.1 Time Machine Tiers

| Capability | Free | Starter ($24) | Professional ($79) | Enterprise ($299) | Agent Enterprise ($499) |
|---|---|---|---|---|---|
| **Replay Window** | 24 hours | 72 hours | 30 days | 90 days | Unlimited |
| **Max Executions per Replay** | 100 | 1,000 | 10,000 | 100,000 | Unlimited |
| **Concurrent Replay Jobs** | 1 | 1 | 3 | 10 | Unlimited |
| **Diff Reports** | Basic (text) | Basic (text) | Full (structured JSON + visual) | Full + side-by-side | Full + side-by-side + export |
| **Auto-Reconciliation** | No | No | Yes (dry-run only) | Yes (live) | Yes (live + custom rules) |
| **Audit Certificates** | No | No | No | Yes (SOC2/HIPAA grade) | Yes (legal-grade + custom) |
| **Replay Scheduling** | No | No | No | Yes (cron-based) | Yes (cron + event-triggered) |
| **Incident Insurance** | No | No | No | No | Yes (dedicated engineer) |
| **Retention of Replay Data** | 7 days | 30 days | 90 days | 365 days | Unlimited |

### 3.2 Backend Feature Constants

**File: `internal/plans/features.go`** — Add new constants:

```go
// Time Machine features
const (
    FeatureTimeMachineBasic      = "time_machine_basic"      // Free+: 24h window, basic diff
    FeatureTimeMachineExtended   = "time_machine_extended"   // Starter+: 72h window
    FeatureTimeMachinePro        = "time_machine_pro"        // Pro+: 30d window, full diffs, dry-run reconciliation
    FeatureTimeMachineEnterprise = "time_machine_enterprise"  // Enterprise+: 90d window, live reconciliation, audit certs
    FeatureTimeMachineUnlimited  = "time_machine_unlimited"   // Agent Enterprise: unlimited everything
    FeatureTimeMachineInsurance  = "time_machine_insurance"   // Agent Enterprise: dedicated incident engineer
)
```

**Feature-to-plan mapping:**

```go
freeFeatures = append(freeFeatures, FeatureTimeMachineBasic)
starterFeatures = append(starterFeatures, FeatureTimeMachineBasic, FeatureTimeMachineExtended)
proFeatures = append(proFeatures, FeatureTimeMachineBasic, FeatureTimeMachineExtended, FeatureTimeMachinePro)
enterpriseFeatures = append(enterpriseFeatures, FeatureTimeMachineBasic, FeatureTimeMachineExtended, FeatureTimeMachinePro, FeatureTimeMachineEnterprise)
agentEnterpriseFeatures = append(agentEnterpriseFeatures, FeatureTimeMachineBasic, FeatureTimeMachineExtended, FeatureTimeMachinePro, FeatureTimeMachineEnterprise, FeatureTimeMachineUnlimited, FeatureTimeMachineInsurance)
```

### 3.3 Backend Numeric Limits

**File: `internal/plans/limits.go`** — Add new limit functions:

```go
// Time Machine limits
const (
    FreeReplayWindowHours          = 24
    StarterReplayWindowHours       = 72
    ProReplayWindowDays            = 30
    EnterpriseReplayWindowDays     = 90
    AgentEnterpriseReplayWindowDays = -1 // Unlimited

    FreeMaxExecutionsPerReplay          = 100
    StarterMaxExecutionsPerReplay       = 1_000
    ProMaxExecutionsPerReplay           = 10_000
    EnterpriseMaxExecutionsPerReplay    = 100_000
    AgentEnterpriseMaxExecutionsPerReplay = -1 // Unlimited

    FreeMaxConcurrentReplays          = 1
    StarterMaxConcurrentReplays       = 1
    ProMaxConcurrentReplays           = 3
    EnterpriseMaxConcurrentReplays    = 10
    AgentEnterpriseMaxConcurrentReplays = -1 // Unlimited
)

func GetReplayWindowHours(plan string) int { ... }
func GetMaxExecutionsPerReplay(plan string) int { ... }
func GetMaxConcurrentReplays(plan string) int { ... }
func SupportsAutoReconciliation(plan string) bool { ... } // Pro+ (dry-run), Enterprise+ (live)
func SupportsAuditCertificates(plan string) bool { ... }   // Enterprise+
func SupportsReplayScheduling(plan string) bool { ... }    // Enterprise+
```

### 3.4 Frontend Feature Gating

**File: `web/dashboard/src/lib/plan-utils.ts`** — Add to FEATURES:

```typescript
export const FEATURES: Record<string, readonly PlanTier[]> = {
    // ... existing features ...
    TIME_MACHINE_BASIC: ['free', 'starter', 'professional', 'enterprise', 'agent_enterprise'],
    TIME_MACHINE_EXTENDED: ['starter', 'professional', 'enterprise', 'agent_enterprise'],
    TIME_MACHINE_PRO: ['professional', 'enterprise', 'agent_enterprise'],
    TIME_MACHINE_ENTERPRISE: ['enterprise', 'agent_enterprise'],
    TIME_MACHINE_UNLIMITED: ['agent_enterprise'],
    TIME_MACHINE_INSURANCE: ['agent_enterprise'],
};
```

**File: `web/dashboard/src/lib/constants.ts`** — Add limits to each plan:

```typescript
// In each PLANS entry:
limits: {
    // ... existing ...
    replayWindowHours: 24 | 72 | 720 | 2160 | Infinity,  // Free/Starter/Pro/Enterprise/AE
    maxExecutionsPerReplay: 100 | 1000 | 10000 | 100000 | Infinity,
    maxConcurrentReplays: 1 | 1 | 3 | 10 | Infinity,
    autoReconciliation: false | false | 'dry-run' | 'live' | 'live',
    auditCertificates: false | false | false | true | true,
}
```

---

## 4. Database Schema

### 4.1 Migration: `20260503120000_create_time_machine_tables.up.sql`

```sql
-- ============================================================
-- Function Time Machine: Core Tables
-- ============================================================

-- A replay job: the top-level container for a time-travel operation
CREATE TABLE IF NOT EXISTS time_machine_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Time window selection
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,

    -- The corrected function version to replay through
    target_version_id UUID NOT NULL REFERENCES registry_function_versions(id),
    target_version VARCHAR(20) NOT NULL,

    -- Configuration
    max_executions INT NOT NULL DEFAULT 1000,
    reconciliation_mode TEXT NOT NULL DEFAULT 'dry_run', -- 'dry_run', 'live', 'preview_only'
    auto_reconcile BOOLEAN NOT NULL DEFAULT false,

    -- Status tracking
    status TEXT NOT NULL DEFAULT 'pending', -- pending, scanning, scanning_complete,
                                            -- replaying, replay_complete,
                                            -- diffing, diff_complete,
                                            -- reconciling, reconciled,
                                            -- completed, failed, cancelled
    progress_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_phase TEXT, -- 'scan', 'replay', 'diff', 'reconcile', 'audit'
    error_message TEXT,

    -- Execution counts
    total_executions_found INT NOT NULL DEFAULT 0,
    total_executions_replayed INT NOT NULL DEFAULT 0,
    total_executions_changed INT NOT NULL DEFAULT 0,
    total_executions_failed INT NOT NULL DEFAULT 0,

    -- Audit metadata
    reason TEXT NOT NULL, -- Developer-provided reason for the replay
    incident_url TEXT,    -- Optional link to incident ticket

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_replays_tenant ON time_machine_replays(tenant_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_function ON time_machine_replays(function_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_status ON time_machine_replays(status);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_created ON time_machine_replays(created_at DESC);

-- Individual execution replay result: one row per original execution
CREATE TABLE IF NOT EXISTS time_machine_replay_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    original_execution_id UUID NOT NULL, -- FK to registry_function_executions or registry_executions_public

    -- Original execution snapshot (frozen at scan time)
    original_input JSONB NOT NULL,
    original_output JSONB NOT NULL,
    original_version VARCHAR(20) NOT NULL,
    original_duration_ms INT NOT NULL,
    original_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    original_meg_root_hash TEXT,
    original_certificate_id TEXT,

    -- Replay result
    new_output JSONB,
    new_duration_ms INT,
    new_meg_root_hash TEXT,
    new_status_code INT,

    -- Diff analysis
    output_changed BOOLEAN, -- NULL until diff is computed
    diff_type TEXT,         -- 'identical', 'minor', 'major', 'breaking', 'error'
    diff_summary TEXT,      -- Human-readable summary
    diff_detail JSONB,      -- Structured diff (field-level changes)

    -- Reconciliation
    reconciliation_status TEXT DEFAULT 'pending', -- pending, reconciled, skipped, failed
    reconciliation_actions JSONB, -- Array of actions taken/required
    reconciled_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    replay_error TEXT,
    replay_error_code TEXT,

    -- Status
    status TEXT NOT NULL DEFAULT 'pending', -- pending, replaying, completed, failed, skipped
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_replay ON time_machine_replay_items(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_status ON time_machine_replay_items(status);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_diff ON time_machine_replay_items(diff_type);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_changed ON time_machine_replay_items(output_changed) WHERE output_changed = true;

-- Reconciliation actions log: tracks every side-effect modification
CREATE TABLE IF NOT EXISTS time_machine_reconciliations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    replay_item_id UUID NOT NULL REFERENCES time_machine_replay_items(id) ON DELETE CASCADE,

    -- What was reconciled
    action_type TEXT NOT NULL, -- 'database_update', 'api_call', 'email_correction',
                               -- 'state_rollback', 'webhook_replay', 'custom'
    target_resource TEXT NOT NULL, -- e.g., 'users.balance', 'orders.total', 'email:user@example.com'

    -- Before/after
    old_value JSONB,
    new_value JSONB,

    -- Execution
    status TEXT NOT NULL DEFAULT 'pending', -- pending, applied, failed, rolled_back
    applied_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,

    -- Metadata
    dry_run BOOLEAN NOT NULL DEFAULT false,
    reversible BOOLEAN NOT NULL DEFAULT true,
    reversal_data JSONB, -- Data needed to undo this action

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_reconciliations_replay ON time_machine_reconciliations(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_reconciliations_status ON time_machine_reconciliations(status);

-- Audit certificates: compliance-grade proof of what changed
CREATE TABLE IF NOT EXISTS time_machine_audit_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    certificate_id TEXT NOT NULL UNIQUE, -- Human-readable ID: TM-{replay_id_short}-{seq}

    -- Certificate content
    cert_json JSONB NOT NULL, -- Full certificate document
    cert_hash TEXT NOT NULL,   -- SHA-256 of canonicalized cert_json

    -- Cryptographic chain
    previous_cert_hash TEXT,   -- Links to previous audit cert for chain integrity
    merkle_root TEXT,          -- Merkle root of all replay item hashes
    signature TEXT,            -- Platform signature over cert_hash

    -- Compliance metadata
    compliance_frameworks TEXT[] DEFAULT '{}', -- ['SOC2', 'HIPAA', 'ISO27001', 'GDPR']
    legal_hold_ref TEXT,       -- Reference to legal hold if applicable
    retention_policy TEXT NOT NULL DEFAULT '7_years', -- Matches financial data retention

    -- Anchoring (optional, for enterprise)
    anchored BOOLEAN NOT NULL DEFAULT false,
    anchor_chain TEXT,
    anchor_tx_hash TEXT,
    anchored_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_audit_certs_replay ON time_machine_audit_certificates(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_audit_certs_cert_id ON time_machine_audit_certificates(certificate_id);

-- Scheduled replays: for enterprise replay-on-schedule
CREATE TABLE IF NOT EXISTS time_machine_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Schedule config
    cron_expression TEXT NOT NULL, -- e.g., '0 3 * * 1' (weekly Monday 3am)
    timezone TEXT NOT NULL DEFAULT 'UTC',

    -- Replay config (template for creating replay jobs)
    replay_window_hours INT NOT NULL DEFAULT 24,
    target_version_strategy TEXT NOT NULL DEFAULT 'latest', -- 'latest', 'pinned', 'previous'
    pinned_version_id UUID,
    reconciliation_mode TEXT NOT NULL DEFAULT 'dry_run',
    auto_reconcile BOOLEAN NOT NULL DEFAULT false,
    reason_template TEXT NOT NULL DEFAULT 'Scheduled replay',

    -- Status
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    total_runs INT NOT NULL DEFAULT 0,
    last_replay_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_tenant ON time_machine_schedules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_next_run ON time_machine_schedules(next_run_at) WHERE enabled = true;
```

### 4.2 Down Migration: `20260503120000_create_time_machine_tables.down.sql`

```sql
DROP TABLE IF EXISTS time_machine_audit_certificates CASCADE;
DROP TABLE IF EXISTS time_machine_reconciliations CASCADE;
DROP TABLE IF EXISTS time_machine_replay_items CASCADE;
DROP TABLE IF EXISTS time_machine_schedules CASCADE;
DROP TABLE IF EXISTS time_machine_replays CASCADE;
```

---

## 5. API Contracts

### 5.1 Route Registration

**File: `internal/api/routes.go`** — Add after existing registry routes:

```go
// ── Time Machine ──────────────────────────────────────────────────────
timemachineRepo := storage.NewTimeMachineRepository(s.postgresDB.GORM)
timemachineHandler := timemachine.NewHandler(
    timemachineRepo,
    s.repo,                   // Main repository (for function/execution lookups)
    s.redisClient,            // Progress tracking
    realtimeUsageTracker,     // Quota integration
    s.notificationSvc,        // Completion notifications
)
registerTimeMachineRoutes(api, timemachineHandler, fm, authMiddleware)
```

**New file: `internal/api/routes_timemachine.go`**

### 5.2 API Endpoints

#### Replay Management

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `POST` | `/v1/time-machine/replays` | Create a new replay job | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays` | List replay jobs (filtered) | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays/:id` | Get replay job details | User | `time_machine_basic` |
| `DELETE` | `/v1/time-machine/replays/:id` | Cancel a running replay | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays/:id/progress` | Real-time progress (SSE/WebSocket) | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays/:id/items` | List replay items (paginated, filterable) | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays/:id/items/:itemId` | Get single replay item with full diff | User | `time_machine_basic` |

#### Diff & Analysis

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `GET` | `/v1/time-machine/replays/:id/diff-summary` | Aggregated diff statistics | User | `time_machine_basic` |
| `GET` | `/v1/time-machine/replays/:id/diff-report` | Full structured diff report (JSON) | User | `time_machine_pro` |
| `GET` | `/v1/time-machine/replays/:id/diff-report/export` | Export diff as CSV/PDF | User | `time_machine_pro` |

#### Reconciliation

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `POST` | `/v1/time-machine/replays/:id/reconcile` | Execute reconciliation (dry-run or live) | User | `time_machine_pro` (dry-run) / `time_machine_enterprise` (live) |
| `GET` | `/v1/time-machine/replays/:id/reconciliations` | List reconciliation actions | User | `time_machine_pro` |
| `POST` | `/v1/time-machine/replays/:id/reconciliations/:reconId/rollback` | Rollback a specific reconciliation | User | `time_machine_enterprise` |

#### Audit Certificates

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `GET` | `/v1/time-machine/replays/:id/audit-certificate` | Get audit certificate | User | `time_machine_enterprise` |
| `GET` | `/v1/time-machine/audit-certificates` | List all audit certificates | User | `time_machine_enterprise` |
| `GET` | `/v1/time-machine/audit-certificates/:certId` | Get specific certificate | User | `time_machine_enterprise` |
| `GET` | `/v1/time-machine/audit-certificates/:certId/verify` | Verify certificate chain integrity | User | `time_machine_enterprise` |
| `POST` | `/v1/time-machine/audit-certificates/:certId/anchor` | Anchor to blockchain (enterprise) | User | `time_machine_enterprise` |

#### Scheduling (Enterprise)

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `POST` | `/v1/time-machine/schedules` | Create a scheduled replay | User | `time_machine_enterprise` |
| `GET` | `/v1/time-machine/schedules` | List scheduled replays | User | `time_machine_enterprise` |
| `PUT` | `/v1/time-machine/schedules/:id` | Update schedule | User | `time_machine_enterprise` |
| `DELETE` | `/v1/time-machine/schedules/:id` | Delete schedule | User | `time_machine_enterprise` |

#### Incident Insurance (Agent Enterprise)

| Method | Path | Description | Auth | Feature Gate |
|--------|------|-------------|------|-------------|
| `POST` | `/v1/time-machine/insurance/request` | Request dedicated engineer for incident | User | `time_machine_insurance` |
| `GET` | `/v1/time-machine/insurance/status` | Check insurance request status | User | `time_machine_insurance` |

### 5.3 Key Request/Response Schemas

#### Create Replay (`POST /v1/time-machine/replays`)

```json
{
    "function_id": "uuid",
    "window_start": "2026-05-01T00:00:00Z",
    "window_end": "2026-05-02T00:00:00Z",
    "target_version_id": "uuid",
    "reason": "Payment calculation was using stale tax rates. Fix deployed as v2.3.1.",
    "incident_url": "https://jira.company.com/PROD-1234",
    "reconciliation_mode": "dry_run",
    "max_executions": 5000,
    "options": {
        "include_cached": false,
        "include_failed": false,
        "parallelism": 4,
        "skip_identical": true
    }
}
```

**Response (201):**

```json
{
    "id": "uuid",
    "status": "pending",
    "function_id": "uuid",
    "window_start": "2026-05-01T00:00:00Z",
    "window_end": "2026-05-02T00:00:00Z",
    "target_version": "2.3.1",
    "reason": "Payment calculation was using stale tax rates.",
    "reconciliation_mode": "dry_run",
    "created_at": "2026-05-02T18:00:00Z",
    "plan_limits": {
        "replay_window_hours": 72,
        "max_executions": 10000,
        "concurrent_replays": 3,
        "auto_reconciliation": "dry-run",
        "audit_certificates": false
    }
}
```

#### Replay Progress (`GET /v1/time-machine/replays/:id/progress`)

SSE stream:

```
event: progress
data: {"phase":"scanning","percent":15,"executions_found":342,"message":"Scanning execution history..."}

event: progress
data: {"phase":"replaying","percent":45,"executions_replayed":154,"executions_changed":12,"executions_failed":1}

event: progress
data: {"phase":"diffing","percent":80,"message":"Computing output diffs..."}

event: complete
data: {"phase":"complete","percent":100,"total_executions":342,"changed":47,"identical":294,"failed":1}
```

#### Diff Summary (`GET /v1/time-machine/replays/:id/diff-summary`)

```json
{
    "replay_id": "uuid",
    "total_executions": 342,
    "identical": 294,
    "changed": 47,
    "failed": 1,
    "breakdown": {
        "identical": 294,
        "minor": 31,
        "major": 14,
        "breaking": 2,
        "error": 1
    },
    "top_changes": [
        {
            "field": "tax_amount",
            "old_pattern": "0.08 (8% rate)",
            "new_pattern": "0.0825 (8.25% rate)",
            "affected_executions": 45
        }
    ],
    "reconciliation_preview": {
        "total_actions": 47,
        "by_type": {
            "database_update": 42,
            "email_correction": 3,
            "webhook_replay": 2
        }
    }
}
```

#### Audit Certificate (`GET /v1/time-machine/replays/:id/audit-certificate`)

```json
{
    "certificate_id": "TM-a3f8c2-001",
    "replay_id": "uuid",
    "issued_at": "2026-05-02T19:30:00Z",
    "issuer": "FunctionFly Time Machine v1.0",
    "compliance_frameworks": ["SOC2", "HIPAA"],

    "incident": {
        "description": "Payment calculation was using stale tax rates",
        "incident_url": "https://jira.company.com/PROD-1234",
        "detected_at": "2026-05-02T14:00:00Z",
        "resolved_at": "2026-05-02T19:30:00Z"
    },

    "function": {
        "id": "uuid",
        "name": "calculate-payment",
        "author": "acme-corp",
        "original_version": "2.3.0",
        "corrected_version": "2.3.1",
        "original_hash": "sha256:abc123...",
        "corrected_hash": "sha256:def456..."
    },

    "replay_window": {
        "start": "2026-04-29T00:00:00Z",
        "end": "2026-05-02T00:00:00Z",
        "duration_hours": 72
    },

    "execution_summary": {
        "total_executions": 342,
        "replayed": 342,
        "identical": 294,
        "changed": 47,
        "failed": 1,
        "skipped": 0
    },

    "reconciliation_summary": {
        "mode": "live",
        "total_actions": 47,
        "applied": 45,
        "failed": 2,
        "rolled_back": 0,
        "actions": [
            {
                "type": "database_update",
                "target": "orders.tax_amount",
                "count": 42,
                "total_delta_cents": 15230
            },
            {
                "type": "email_correction",
                "target": "customer_receipt",
                "count": 3
            }
        ]
    },

    "cryptographic_proof": {
        "merkle_root": "sha256:7a8b9c...",
        "item_hashes": ["sha256:...", "..."],
        "cert_hash": "sha256:f1e2d3...",
        "previous_cert_hash": null,
        "platform_signature": "ecdsa-p256:..."
    },

    "legal_statement": "This certificate documents a controlled retroactive correction of function execution outputs performed via FunctionFly Time Machine. All original execution data is preserved. The correction was applied with full auditability and cryptographic integrity verification."
}
```

---

## 6. Go Implementation Structure

### 6.1 New Package: `internal/timemachine/`

```
internal/timemachine/
├── handler.go              # HTTP handlers (TimeMachineHandler struct)
├── handler_replays.go      # Replay CRUD + SSE progress
├── handler_diffs.go        # Diff endpoints
├── handler_reconcile.go    # Reconciliation endpoints
├── handler_audit.go        # Audit certificate endpoints
├── handler_schedules.go    # Scheduled replay endpoints
├── service.go              # TimeMachineService (business logic orchestrator)
├── replay_engine.go        # Core replay engine (scan → replay → diff)
├── diff_engine.go          # Output comparison engine
├── reconciliation_engine.go # State reconciliation engine
├── audit_generator.go      # Audit certificate generator
├── scheduler.go            # Cron-based replay scheduler
├── worker_pool.go          # Replay worker pool
├── types.go                # Request/response types
└── limits.go               # Plan-aware limit enforcement
```

### 6.2 New Package: `internal/storage/timemachine/`

```
internal/storage/timemachine/
├── repository.go           # TimeMachineRepository (GORM-based)
├── replay_repo.go          # CRUD for time_machine_replays
├── item_repo.go            # CRUD for time_machine_replay_items
├── reconciliation_repo.go  # CRUD for time_machine_reconciliations
├── audit_repo.go           # CRUD for time_machine_audit_certificates
├── schedule_repo.go        # CRUD for time_machine_schedules
└── types.go                # GORM models
```

### 6.3 Service Layer Design

```go
// internal/timemachine/service.go
type TimeMachineService struct {
    repo        *timemachinerepo.Repository
    mainRepo    storage.Repository       // For function/execution lookups
    executor    FunctionExecutor         // Re-uses existing sandbox executor
    redis       *redis.Client            // Progress tracking
    notifier    *notification.Service    // Completion notifications
    planLimits  PlanLimitChecker         // Plan-aware enforcement
}

type PlanLimitChecker interface {
    GetReplayWindowHours(plan string) int
    GetMaxExecutionsPerReplay(plan string) int
    GetMaxConcurrentReplays(plan string) int
    CanAutoReconcile(plan string) bool
    CanGenerateAuditCert(plan string) bool
    SupportsLiveReconciliation(plan string) bool
}
```

### 6.4 Replay Engine Pipeline

```
┌───────────┐    ┌───────────┐    ┌───────────┐    ┌───────────┐    ┌───────────┐
│  1. SCAN  │───▶│ 2. REPLAY │───▶│  3. DIFF  │───▶│4. RECONCILE│───▶│5. AUDIT   │
│           │    │           │    │           │    │           │    │           │
│ Query     │    │ Execute   │    │ Compare   │    │ Apply     │    │ Generate  │
│ execution │    │ each      │    │ old vs    │    │ corrective│    │ certificate│
│ history   │    │ through   │    │ new       │    │ actions   │    │           │
│ for time  │    │ corrected │    │ outputs   │    │           │    │           │
│ window    │    │ function  │    │           │    │           │    │           │
└───────────┘    └───────────┘    └───────────┘    └───────────┘    └───────────┘
```

**Phase 1 — SCAN:**
- Query `registry_function_executions` + `registry_executions_public` for the function within the time window
- Respect plan limits (max executions, window boundaries)
- Snapshot original inputs/outputs into `time_machine_replay_items`
- Count total and report progress

**Phase 2 — REPLAY:**
- For each replay item, execute the corrected function with the original input
- Use the existing `SandboxExecutor` / runtime pipeline (same path as normal execution)
- Capture new output, duration, status code
- Build MEG record for the replayed execution (for cryptographic verification)
- Report progress via Redis pub/sub → SSE endpoint

**Phase 3 — DIFF:**
- Compare original output vs new output
- Classification: `identical` (byte-equal), `minor` (cosmetic/formatting), `major` (semantic change), `breaking` (contract violation), `error` (replay failed)
- Generate structured diff (JSON Patch format for JSON outputs, semantic diff for structured data)
- Aggregate statistics for the diff summary

**Phase 4 — RECONCILE (if enabled):**
- Parse reconciliation actions from diff results
- For `dry_run`: generate action list with previews, do NOT apply
- For `live`: apply actions with transactional safety, record before/after values
- Support rollback of individual actions
- Generate reconciliation log

**Phase 5 — AUDIT (if enabled):**
- Build Merkle tree from all replay item hashes
- Generate certificate JSON with full provenance
- Sign with platform key
- Optionally anchor to blockchain
- Store in `time_machine_audit_certificates`

### 6.5 Progress Tracking via Redis

```go
// Redis key pattern for live progress
// time_machine:progress:{replay_id} → JSON progress object
// TTL: 24 hours after completion

type ReplayProgress struct {
    Phase               string  `json:"phase"`
    Percent             float64 `json:"percent"`
    ExecutionsFound     int     `json:"executions_found"`
    ExecutionsReplayed  int     `json:"executions_replayed"`
    ExecutionsChanged   int     `json:"executions_changed"`
    ExecutionsFailed    int     `json:"executions_failed"`
    Message             string  `json:"message"`
    UpdatedAt           time.Time `json:"updated_at"`
}
```

---

## 7. Worker Architecture

### 7.1 Replay Worker Pool

Extends the existing `internal/queue/execution_queue.go` pattern:

```go
// internal/timemachine/worker_pool.go
type ReplayWorkerPool struct {
    queue        *queue.ExecutionQueue   // Reuse existing priority queue
    concurrency  int                     // Plan-aware worker count
    service      *TimeMachineService
    redis        *redis.Client
}

// Each worker processes one replay item at a time
// Items are batched per replay job for efficiency
func (p *ReplayWorkerPool) processReplayJob(replayID uuid.UUID) error {
    // 1. Fetch pending items from DB
    // 2. Execute each through the corrected function
    // 3. Update progress in Redis
    // 4. On completion, trigger diff phase
    // 5. On diff completion, trigger reconciliation (if enabled)
    // 6. On reconciliation completion, generate audit cert (if enabled)
    // 7. Send notification to user
}
```

### 7.2 Scheduler (Enterprise)

Extends `internal/scheduler/scheduler.go` pattern:

```go
// internal/timemachine/scheduler.go
type TimeMachineScheduler struct {
    cron         *cron.Schedule
    repo         *timemachinerepo.Repository
    service      *TimeMachineService
}

// Runs on a cron schedule, checks time_machine_schedules for due jobs
func (s *TimeMachineScheduler) tick() {
    // 1. Query schedules WHERE enabled=true AND next_run_at <= NOW()
    // 2. For each, create a replay job automatically
    // 3. Update next_run_at
}
```

---

## 8. Frontend Dashboard

### 8.1 New Pages

| Route | Component | Description |
|-------|-----------|-------------|
| `/time-machine` | `TimeMachinePage` | Main Time Machine dashboard — list of replay jobs, quick-start wizard |
| `/time-machine/new` | `NewReplayPage` | Create replay wizard (function select, time window, options) |
| `/time-machine/:id` | `ReplayDetailPage` | Replay detail — progress, diff viewer, reconciliation, audit cert |
| `/time-machine/:id/diff` | `DiffExplorerPage` | Full diff explorer with filtering and export |
| `/time-machine/schedules` | `ReplaySchedulesPage` | Manage scheduled replays (enterprise) |
| `/time-machine/audit` | `AuditCertificatesPage` | List and verify audit certificates (enterprise) |

### 8.2 New Components

```
web/dashboard/src/components/time-machine/
├── TimeMachineLayout.tsx          # Layout wrapper with sidebar nav
├── ReplayWizard.tsx               # Multi-step replay creation wizard
│   ├── FunctionSelector.tsx       # Step 1: Select function
│   ├── TimeWindowSelector.tsx     # Step 2: Date/time range picker
│   ├── VersionSelector.tsx        # Step 3: Select corrected version
│   ├── ReplayOptions.tsx          # Step 4: Configure options
│   └── ReplayConfirmation.tsx     # Step 5: Review and confirm
├── ReplayProgressCard.tsx         # Real-time progress with SSE
├── DiffViewer.tsx                 # Side-by-side diff viewer
│   ├── DiffSummary.tsx            # Aggregated statistics
│   ├── DiffItemRow.tsx            # Individual execution diff row
│   ├── DiffDetailModal.tsx        # Full diff for single execution
│   └── DiffExportButton.tsx       # Export as CSV/PDF
├── ReconciliationPanel.tsx        # Reconciliation actions list
│   ├── ReconciliationAction.tsx   # Single action card
│   ├── ReconciliationDryRun.tsx   # Dry-run preview
│   └── ReconciliationRollback.tsx # Rollback confirmation
├── AuditCertificateViewer.tsx     # Audit certificate display
│   ├── CertificateChain.tsx       # Visual chain of custody
│   ├── ComplianceBadges.tsx       # SOC2/HIPAA badges
│   └── CertificateExport.tsx      # PDF export
├── ReplayHistoryTable.tsx         # Paginated replay history
├── ScheduleManager.tsx            # Schedule CRUD (enterprise)
└── IncidentInsurancePanel.tsx     # Request dedicated engineer (AE)
```

### 8.3 New API Module

**File: `web/dashboard/src/api/timeMachine.ts`**

```typescript
import { apiClient } from './client';

export interface CreateReplayRequest { ... }
export interface ReplayJob { ... }
export interface ReplayItem { ... }
export interface DiffSummary { ... }
export interface AuditCertificate { ... }
export interface ReplaySchedule { ... }

export async function createReplay(req: CreateReplayRequest): Promise<ReplayJob> { ... }
export async function listReplays(params?: ListParams): Promise<Paginated<ReplayJob>> { ... }
export async function getReplay(id: string): Promise<ReplayJob> { ... }
export async function cancelReplay(id: string): Promise<void> { ... }
export async function getReplayProgress(id: string): EventSource { ... } // SSE
export async function getReplayItems(id: string, params?: ListParams): Promise<Paginated<ReplayItem>> { ... }
export async function getDiffSummary(id: string): Promise<DiffSummary> { ... }
export async function getDiffReport(id: string): Promise<DiffReport> { ... }
export async function exportDiffReport(id: string, format: 'csv' | 'pdf'): Promise<Blob> { ... }
export async function startReconciliation(id: string, mode: 'dry_run' | 'live'): Promise<void> { ... }
export async function getReconciliations(id: string): Promise<Reconciliation[]> { ... }
export async function rollbackReconciliation(id: string, reconId: string): Promise<void> { ... }
export async function getAuditCertificate(id: string): Promise<AuditCertificate> { ... }
export async function listAuditCertificates(params?: ListParams): Promise<Paginated<AuditCertificate>> { ... }
export async function verifyCertificate(certId: string): Promise<VerificationResult> { ... }
export async function createSchedule(req: CreateScheduleRequest): Promise<ReplaySchedule> { ... }
export async function listSchedules(): Promise<ReplaySchedule[]> { ... }
export async function updateSchedule(id: string, req: UpdateScheduleRequest): Promise<ReplaySchedule> { ... }
export async function deleteSchedule(id: string): Promise<void> { ... }
```

### 8.4 New Hooks

**File: `web/dashboard/src/hooks/useTimeMachine.ts`**

```typescript
export function useReplays(params?: ListParams) { ... }
export function useReplay(id: string) { ... }
export function useCreateReplay() { ... }
export function useCancelReplay() { ... }
export function useReplayProgress(id: string) { ... } // SSE-based
export function useReplayItems(id: string, params?: ListParams) { ... }
export function useDiffSummary(id: string) { ... }
export function useDiffReport(id: string) { ... }
export function useStartReconciliation() { ... }
export function useReconciliations(id: string) { ... }
export function useRollbackReconciliation() { ... }
export function useAuditCertificate(id: string) { ... }
export function useAuditCertificates(params?: ListParams) { ... }
export function useVerifyCertificate() { ... }
export function useSchedules() { ... }
export function useCreateSchedule() { ... }
export function useUpdateSchedule() { ... }
export function useDeleteSchedule() { ... }
```

### 8.5 Sidebar Navigation

Add to `web/dashboard/src/components/layout/Sidebar.tsx`:

```typescript
{
    label: 'Time Machine',
    path: '/time-machine',
    icon: ClockRewind, // from lucide-react
    feature: 'TIME_MACHINE_BASIC', // Always visible on free+
    children: [
        { label: 'Replays', path: '/time-machine' },
        { label: 'Schedules', path: '/time-machine/schedules', feature: 'TIME_MACHINE_ENTERPRISE' },
        { label: 'Audit Certs', path: '/time-machine/audit', feature: 'TIME_MACHINE_ENTERPRISE' },
    ],
}
```

---

## 9. Billing Integration

### 9.1 Usage-Based Metering

Time Machine replays are metered as a distinct usage event type, integrated with the existing billing system:

**File: `internal/api/handlers/billing/pricing_v2.go`** — Add:

```go
// Time Machine usage events
"timemachine_replay_execution": {
    Name:         "Time Machine Replay Execution",
    FreeAllowance: map[string]int{
        "free":          100,    // 100 replay executions/month free
        "starter":       500,
        "professional":  5000,
        "enterprise":    50000,
        "agent_enterprise": -1,  // Unlimited
    },
    OveragePriceCents: map[string]int{
        "starter":       50,     // $0.50 per 1K replay executions
        "professional":  30,     // $0.30 per 1K
        "enterprise":    15,     // $0.15 per 1K
    },
},
"timemachine_reconciliation_action": {
    Name:         "Time Machine Reconciliation Action",
    FreeAllowance: map[string]int{
        "free":          0,
        "starter":       0,
        "professional":  0,      // Dry-run only, not billed
        "enterprise":    1000,   // 1K live reconciliation actions/month
        "agent_enterprise": -1,
    },
    OveragePriceCents: map[string]int{
        "enterprise":    200,    // $2.00 per 1K reconciliation actions
    },
},
```

### 9.2 Stripe Product/Price IDs

Add to Stripe configuration:

| Product | Price | Billing |
|---------|-------|---------|
| TM Replay Executions (Starter Overage) | $0.50/1K | Metered |
| TM Replay Executions (Pro Overage) | $0.30/1K | Metered |
| TM Replay Executions (Enterprise Overage) | $0.15/1K | Metered |
| TM Reconciliation Actions (Enterprise Overage) | $2.00/1K | Metered |
| TM Incident Insurance (Agent Enterprise) | Included | Included in $499/mo |

### 9.3 Upgrade Nudge Strategy

When a user hits their Time Machine limit, the API returns a structured error with upgrade CTA:

```json
{
    "error": {
        "code": "TIME_MACHINE_LIMIT_EXCEEDED",
        "message": "Your plan allows a 24-hour replay window. Upgrade to replay up to 90 days of history.",
        "current_plan": "free",
        "limit_type": "replay_window",
        "current_value": 24,
        "upgrade_options": [
            { "plan": "starter", "window": "72 hours", "price": "$24/mo" },
            { "plan": "professional", "window": "30 days", "price": "$79/mo" },
            { "plan": "enterprise", "window": "90 days", "price": "$299/mo" }
        ]
    }
}
```

---

## 10. Data Flow: End-to-End Replay

```
Developer discovers bug
         │
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 1. Developer fixes function code, deploys as v2.3.1                │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 2. Developer opens Time Machine → "New Replay"                     │
│    - Selects function: "calculate-payment"                          │
│    - Time window: "Last 72 hours"                                   │
│    - Target version: v2.3.1 (auto-detected as latest)              │
│    - Reason: "Tax rate was 8% instead of 8.25%"                    │
│    - Mode: "Dry Run" (preview first)                               │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 3. Platform SCANs execution history                                 │
│    - Queries registry_function_executions WHERE                     │
│      function_id=X AND timestamp BETWEEN window_start AND window_end│
│    - Found: 342 executions                                          │
│    - Snapshots original inputs/outputs into replay_items            │
│    - Progress: 15% → "Found 342 executions in window"              │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 4. Platform REPLAYs each execution through v2.3.1                   │
│    - Worker pool processes items (4 parallel workers)               │
│    - Each item: execute v2.3.1 with original input → capture output│
│    - Builds MEG record for each replayed execution                  │
│    - Progress: 45% → "Replayed 154/342, 12 changed, 1 failed"     │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 5. Platform DIFFs old vs new outputs                                │
│    - Byte comparison → semantic diff → field-level diff             │
│    - Classifies: 294 identical, 31 minor, 14 major, 2 breaking    │
│    - "tax_amount" field: old=0.08 → new=0.0825 (45 executions)    │
│    - Progress: 80% → "Diff complete"                               │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 6. Developer reviews diff report                                    │
│    - Sees side-by-side comparison for each changed execution        │
│    - Aggregated: "Total tax delta: +$152.30 across 45 orders"     │
│    - Decides: "This looks correct. Apply reconciliation."           │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 7. Platform RECONCILEs (enterprise: live mode)                      │
│    - 42 database updates: orders.tax_amount adjusted               │
│    - 3 email corrections: resend receipt with correct amount       │
│    - 2 webhook replays: notify downstream with corrected data      │
│    - Each action logged with before/after values                    │
│    - Progress: 95% → "45/47 actions applied, 2 failed"            │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 8. AUDIT CERTIFICATE generated (enterprise)                         │
│    - Merkle tree of all 342 replay item hashes                     │
│    - Signed by platform key                                         │
│    - Includes: incident description, function versions,            │
│      execution summary, reconciliation summary, timestamps         │
│    - SOC2/HIPAA compliance metadata                                │
│    - Chain-linked to previous audit certificates                    │
│    - Progress: 100% → "Replay complete. Certificate issued."       │
└────────────────────────┬────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 9. Developer receives notification                                  │
│    - In-app notification: "Time Machine replay complete"            │
│    - Email summary with diff report                                 │
│    - Audit certificate available for download                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 11. Security Considerations

### 11.1 Authorization

- All Time Machine endpoints require authenticated user + tenant context
- Replays are scoped to the tenant's own functions only
- Reconciliation actions require explicit user confirmation (no auto-apply without consent)
- Audit certificates are immutable once generated

### 11.2 Rate Limiting

- Max concurrent replay jobs per tenant (plan-aware: 1/1/3/10/unlimited)
- Replay execution rate limited to avoid impacting live traffic (max 10% of normal execution rate)
- Dedicated replay worker pool with separate resource allocation

### 11.3 Data Preservation

- Original execution data is NEVER modified — replays create new records
- Reconciliation actions are fully reversible (reversal data stored with each action)
- Audit certificates are append-only and chain-linked
- Deleted functions: replay items preserve original data even if function is later deleted

### 11.4 Compliance

- Audit certificates include compliance framework references (SOC2, HIPAA, ISO27001, GDPR)
- Retention policy for replay data matches financial data retention (7 years for enterprise)
- Legal hold integration: `time_machine_audit_certificates` references `legal_holds` table
- All reconciliation actions are logged in `retention_audit_log` (existing table)

---

## 12. Scaling Strategy

### 12.1 Horizontal Scaling

- Replay worker pool scales with instance count (each instance runs N workers)
- Redis-based distributed lock prevents duplicate processing
- Postgres advisory locks for replay job claiming

### 12.2 Performance Targets

| Metric | Target |
|--------|--------|
| Scan phase (10K executions) | < 5 seconds |
| Replay throughput | 100 executions/second/worker |
| Diff computation | < 1ms per execution |
| Full replay (1K executions) | < 2 minutes |
| Full replay (10K executions) | < 15 minutes |
| Audit certificate generation | < 10 seconds |

### 12.3 Resource Isolation

- Replay workers run in separate goroutine pool (don't compete with live execution)
- Replay executions bypass quota middleware (tracked separately for billing)
- Database queries use read replicas where available for scan phase

---

## 13. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Database migration (5 tables)
- [ ] Storage repository (`internal/storage/timemachine/`)
- [ ] Core replay engine (scan → replay → diff)
- [ ] Basic API handlers (create, list, get, cancel)
- [ ] Worker pool (in-memory queue integration)
- [ ] Progress tracking (Redis + SSE)
- [ ] Feature constants (backend + frontend)
- [ ] Plan limits integration

### Phase 2: Dashboard (Week 2-3)
- [ ] Time Machine page + replay list
- [ ] New Replay wizard
- [ ] Replay detail page with progress
- [ ] Diff viewer component
- [ ] API module + hooks
- [ ] Sidebar navigation integration

### Phase 3: Reconciliation (Week 3-4)
- [ ] Reconciliation engine (dry-run + live)
- [ ] Reconciliation API endpoints
- [ ] Reconciliation UI (panel, actions, rollback)
- [ ] Diff export (CSV/PDF)

### Phase 4: Audit & Enterprise (Week 4-5)
- [ ] Audit certificate generator
- [ ] Audit certificate API + viewer
- [ ] Certificate verification + chain integrity
- [ ] Blockchain anchoring integration
- [ ] Scheduled replay engine
- [ ] Schedule management API + UI

### Phase 5: Billing & Polish (Week 5-6)
- [ ] Stripe metered billing integration
- [ ] Usage tracking for replay executions
- [ ] Upgrade nudges and limit enforcement
- [ ] Incident insurance request flow (AE)
- [ ] Notification integration (email + in-app)
- [ ] E2E testing
- [ ] Documentation

---

## 14. MRR Projections

### Conservative Estimate (6 months post-launch)

| Tier | New Conversions/mo | ARPU | Monthly Revenue |
|------|-------------------|------|----------------|
| Starter (TM-driven) | 50 | $24 | $1,200 |
| Professional (TM-driven) | 30 | $79 | $2,370 |
| Enterprise (TM-driven) | 10 | $299 | $2,990 |
| Agent Enterprise (TM-driven) | 3 | $499 | $1,497 |
| Overage revenue | — | — | ~$2,000 |
| **Total TM-driven MRR** | | | **~$10,057** |

### Growth Loop Validation

```
Developer has production incident on another platform
    → Spends 3 days manually fixing database records
    → Hears about Time Machine
    → Signs up for FunctionFly (free tier)
    → Experiences first 24-hour replay
    → Upgrades to Professional for 30-day window
    → Company mandates FunctionFly for compliance
    → Enterprise contract signed
```

---

## 15. Existing Infrastructure Reused

| Component | Existing Code | Reuse |
|-----------|--------------|-------|
| Execution history | `registry_function_executions`, `registry_executions_public` | Primary data source for scan phase |
| Sandbox executor | `internal/api/handlers/registry/execution/handlers.go` | Replay engine re-uses same execution pipeline |
| MEG/certificates | `internal/storage/registry/dre_repository.go` | Builds MEG for replayed executions |
| Feature gating | `internal/plans/features.go`, `feature_checker.go`, `middleware/features.go` | New feature constants added to existing system |
| Plan limits | `internal/plans/limits.go` | New limit functions follow existing pattern |
| Worker queue | `internal/queue/execution_queue.go` | Replay worker pool extends existing pattern |
| Progress tracking | Redis + existing patterns | Same Redis instance, new key namespace |
| Notifications | `internal/notification/service.go` | Replay completion notifications |
| Billing metering | `internal/api/handlers/billing/pricing_v2.go` | New usage event types |
| Frontend gating | `plan-utils.ts`, `usePlan.ts`, `EnterpriseFeature.tsx` | New features added to FEATURES map |
| Route middleware | `middleware/features.go` | `RequireFeature()` for Time Machine endpoints |
| Batch operations | `internal/storage/postgres_execution_retention.go` | `deleteOldRecordsInBatches` pattern for scan phase |
| Audit log | `retention_audit_log` table | Reconciliation actions logged here |
| Legal holds | `legal_holds` table | Referenced by audit certificates |
