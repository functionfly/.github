# Function DNA — Living Code That Evolves Itself In Production

## Overview

Function DNA gives every deployed function a **genetic fingerprint** derived from real production traffic. Instead of treating code as static, the platform continuously collects execution micro-data, uses AI to identify optimization opportunities, generates improved code variants, and presents developers with an accept/reject diff — all without the developer touching the function.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Execution Pipeline                           │
│                                                                     │
│  Function Execute → DNA Collector → Execution Metrics Queue         │
│                                                                     │
│  ┌──────────────┐    ┌───────────────┐    ┌──────────────────────┐  │
│  │ SAR Runtime   │───▶│ DNA Collector │───▶│ PostgreSQL           │  │
│  │ Node.js RT    │    │ (middleware)   │    │ function_dna_*       │  │
│  │ WASM Runtime  │    └───────────────┘    └──────────────────────┘  │
│  └──────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────┘
         │                                        │
         ▼                                        ▼
┌─────────────────────────┐         ┌──────────────────────────────┐
│  DNA Analysis Worker    │         │  DNA API (Go handler)        │
│  (Go goroutine pool)    │         │  /v1/functions/{id}/dna/*    │
│                         │         │                              │
│  1. Aggregate metrics   │         │  GET /profile                │
│  2. Detect patterns     │         │  GET /mutations              │
│  3. Call AI Service     │         │  GET /mutations/{id}         │
│  4. Generate variants   │         │  GET /variants               │
│  5. Store proposals     │         │  GET /variants/{id}          │
│                         │         │  POST /variants/{id}/accept  │
└────────┬────────────────┘         │  POST /variants/{id}/reject  │
         │                          │  GET /insights               │
         ▼                          │  POST /analyze               │
┌─────────────────────────┐         └──────────────────────────────┘
│  AI Service (Python)    │
│  /api/dna/analyze       │
│  /api/dna/generate      │
│                         │
│  - Pattern detection    │
│  - Code optimization    │
│  - Variant generation   │
│  - Impact estimation    │
└─────────────────────────┘
```

---

## Data Model

### `function_dna_profiles`

The master record for a function's genetic identity.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | PK |
| `function_id` | TEXT | FK → `registry_functions.id` or managed function ID |
| `function_type` | TEXT | `registry` or `managed` |
| `tenant_id` | TEXT | Owner tenant |
| `generation` | INT | Current evolution generation (starts at 1) |
| `fitness_score` | FLOAT | 0-100 composite health score |
| `total_executions` | BIGINT | Lifetime execution count analyzed |
| `total_mutations` | INT | Number of accepted evolutions |
| `avg_latency_ms` | FLOAT | Rolling average latency |
| `p99_latency_ms` | FLOAT | 99th percentile latency |
| `success_rate` | FLOAT | 0-1 success rate |
| `error_distribution` | JSONB | Error category → count map |
| `input_patterns` | JSONB | Detected input shape patterns |
| `bottleneck_signature` | JSONB | Identified bottleneck fingerprints |
| `dna_hash` | TEXT | SHA-256 of current genetic fingerprint |
| `last_analyzed_at` | TIMESTAMPTZ | Last analysis run |
| `evolution_enabled` | BOOLEAN | Owner toggle |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

### `function_dna_execution_metrics`

Rolling window of per-execution micro-data. Partitioned by month, auto-cleaned after 90 days.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | PK |
| `function_id` | TEXT | |
| `function_type` | TEXT | `registry` or `managed` |
| `execution_id` | TEXT | FK to execution record |
| `duration_ms` | INT | |
| `memory_peak_mb` | FLOAT | |
| `cpu_time_ms` | INT | |
| `input_size_bytes` | INT | |
| `output_size_bytes` | INT | |
| `input_shape_hash` | TEXT | Hash of input schema shape |
| `status_code` | INT | |
| `error_category` | TEXT | `timeout`, `oom`, `runtime`, `logic`, `network`, `none` |
| `cold_start` | BOOLEAN | |
| `cache_hit` | BOOLEAN | |
| `region` | TEXT | |
| `timestamp` | TIMESTAMPTZ | |

Indexes: `(function_id, timestamp DESC)`, `(function_id, error_category)`, `(function_id, input_shape_hash)`

### `function_dna_mutations`

Immutable log of every evolution event (accepted or rejected).

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | PK |
| `function_id` | TEXT | |
| `function_type` | TEXT | |
| `tenant_id` | TEXT | |
| `generation` | INT | Generation this mutation targets |
| `mutation_type` | TEXT | `optimize_latency`, `reduce_memory`, `fix_error_pattern`, `improve_reliability`, `refactor_hotpath` |
| `status` | TEXT | `proposed`, `accepted`, `rejected`, `deployed`, `rolled_back` |
| `trigger_reason` | TEXT | Human-readable why this was proposed |
| `original_code` | TEXT | Code before mutation |
| `mutated_code` | TEXT | Code after mutation |
| `original_hash` | TEXT | SHA-256 of original |
| `mutated_hash` | TEXT | SHA-256 of mutated |
| `diff` | TEXT | Unified diff |
| `estimated_impact` | JSONB | `{ latency_improvement_pct, memory_reduction_pct, reliability_improvement_pct }` |
| `actual_impact` | JSONB | Measured after deployment (null until deployed) |
| `confidence` | FLOAT | AI confidence 0-1 |
| `model_used` | TEXT | Which AI model generated this |
| `analysis_window_hours` | INT | Hours of data analyzed |
| `executions_analyzed` | INT | Number of executions in analysis window |
| `accepted_by` | TEXT | User ID who accepted |
| `accepted_at` | TIMESTAMPTZ | |
| `deployed_at` | TIMESTAMPTZ | |
| `rolled_back_at` | TIMESTAMPTZ | |
| `created_at` | TIMESTAMPTZ | |

Indexes: `(function_id, generation)`, `(function_id, status)`, `(tenant_id, created_at DESC)`

### `function_dna_insights`

Aggregated analytics for the enterprise insights dashboard.

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | PK |
| `tenant_id` | TEXT | |
| `period_start` | TIMESTAMPTZ | |
| `period_end` | TIMESTAMPTZ | |
| `total_functions_analyzed` | INT | |
| `total_mutations_proposed` | INT | |
| `total_mutations_accepted` | INT | |
| `avg_fitness_score` | FLOAT | |
| `avg_latency_improvement_pct` | FLOAT | |
| `total_cost_savings_usd` | FLOAT | Estimated compute savings |
| `top_bottleneck_categories` | JSONB | Top 5 bottleneck types |
| `evolution_leaderboard` | JSONB | Top 10 most-evolved functions |
| `created_at` | TIMESTAMPTZ | |

---

## API Contract

All endpoints require authentication. Tenant isolation enforced on every request.

### `GET /v1/functions/{id}/dna`

Returns the DNA profile for a function.

```json
{
  "id": "uuid",
  "function_id": "func-123",
  "function_type": "registry",
  "generation": 7,
  "fitness_score": 82.5,
  "total_executions": 145000,
  "total_mutations": 6,
  "avg_latency_ms": 45.2,
  "p99_latency_ms": 180.0,
  "success_rate": 0.997,
  "error_distribution": { "timeout": 12, "runtime": 3 },
  "input_patterns": [
    { "shape": "object:{id:string, data:object}", "frequency": 0.85 },
    { "shape": "object:{id:string}", "frequency": 0.15 }
  ],
  "bottleneck_signature": [
    { "type": "cold_start", "severity": "medium", "frequency": 0.12 },
    { "type": "db_query", "severity": "low", "frequency": 0.03 }
  ],
  "dna_hash": "sha256:abc123...",
  "evolution_enabled": true,
  "last_analyzed_at": "2026-05-02T10:00:00Z",
  "created_at": "2026-04-01T00:00:00Z"
}
```

### `GET /v1/functions/{id}/dna/mutations`

List mutation history with filtering.

Query params: `status`, `mutation_type`, `limit` (default 20), `offset` (default 0)

```json
{
  "mutations": [
    {
      "id": "uuid",
      "generation": 6,
      "mutation_type": "optimize_latency",
      "status": "accepted",
      "trigger_reason": "P99 latency increased 40% over 48h due to N+1 query pattern in hot path",
      "estimated_impact": {
        "latency_improvement_pct": 38.5,
        "memory_reduction_pct": 12.0,
        "reliability_improvement_pct": 2.1
      },
      "actual_impact": {
        "latency_improvement_pct": 41.2,
        "memory_reduction_pct": 8.5,
        "reliability_improvement_pct": 0.5
      },
      "confidence": 0.87,
      "executions_analyzed": 50000,
      "accepted_by": "user-123",
      "accepted_at": "2026-04-28T14:00:00Z",
      "created_at": "2026-04-28T10:00:00Z"
    }
  ],
  "total": 6,
  "limit": 20,
  "offset": 0
}
```

### `GET /v1/functions/{id}/dna/mutations/{mutation_id}`

Full mutation detail including code diff.

```json
{
  "id": "uuid",
  "generation": 6,
  "mutation_type": "optimize_latency",
  "status": "accepted",
  "trigger_reason": "...",
  "original_code": "async function handler(input) { ... }",
  "mutated_code": "async function handler(input) { ... }",
  "diff": "--- a/original\n+++ b/mutated\n...",
  "estimated_impact": { ... },
  "actual_impact": { ... },
  "confidence": 0.87,
  "model_used": "deepseek-v3",
  "analysis_window_hours": 48,
  "executions_analyzed": 50000,
  "created_at": "2026-04-28T10:00:00Z"
}
```

### `POST /v1/functions/{id}/dna/variants/{mutation_id}/accept`

Accept a proposed variant. Triggers canary deployment.

```json
// Request
{ "canary_percentage": 10 }

// Response
{
  "mutation_id": "uuid",
  "status": "deploying",
  "canary_percentage": 10,
  "deployment_id": "deploy-456"
}
```

### `POST /v1/functions/{id}/dna/variants/{mutation_id}/reject`

Reject a proposed variant.

```json
// Request
{ "reason": "We're refactoring this manually" }

// Response
{
  "mutation_id": "uuid",
  "status": "rejected"
}
```

### `GET /v1/functions/{id}/dna/insights`

Time-series performance data for the DNA dashboard.

Query params: `period` (7d, 30d, 90d), `granularity` (hour, day)

```json
{
  "function_id": "func-123",
  "period": "30d",
  "timeline": [
    {
      "timestamp": "2026-04-01T00:00:00Z",
      "fitness_score": 72.0,
      "avg_latency_ms": 85.0,
      "success_rate": 0.99,
      "executions": 5000,
      "generation": 3
    }
  ],
  "mutation_outcomes": {
    "accepted": 4,
    "rejected": 2,
    "proposed": 1,
    "deployed": 3,
    "rolled_back": 1
  },
  "cumulative_improvement": {
    "latency_reduction_pct": 45.0,
    "reliability_gain_pct": 2.3,
    "estimated_monthly_savings_usd": 127.50
  }
}
```

### `POST /v1/functions/{id}/dna/analyze`

Manually trigger DNA analysis for a function.

```json
// Response
{
  "analysis_id": "uuid",
  "status": "queued",
  "estimated_duration_seconds": 120
}
```

### `GET /v1/dna/enterprise/insights`

Enterprise-wide DNA insights (requires enterprise plan).

Query params: `period` (7d, 30d, 90d)

```json
{
  "period": "30d",
  "total_functions_analyzed": 142,
  "total_mutations_proposed": 38,
  "total_mutations_accepted": 29,
  "avg_fitness_score": 78.3,
  "avg_latency_improvement_pct": 32.1,
  "total_cost_savings_usd": 2340.00,
  "top_bottleneck_categories": [
    { "category": "cold_start", "count": 15 },
    { "category": "db_query", "count": 12 },
    { "category": "memory_allocation", "count": 8 }
  ],
  "evolution_leaderboard": [
    { "function_id": "func-123", "name": "processPayment", "generation": 12, "fitness_score": 95.2 },
    { "function_id": "func-456", "name": "generateReport", "generation": 8, "fitness_score": 88.7 }
  ]
}
```

---

## Background Workers

### DNA Collector (Middleware)

Lives in the execution pipeline. After every function execution:

1. Extract micro-metrics (duration, memory, input shape, error category)
2. Insert into `function_dna_execution_metrics`
3. Increment `total_executions` on the DNA profile (debounced, every 100th execution)
4. If `total_executions % 10000 == 0` and evolution is enabled → queue analysis

### DNA Analysis Worker

Go goroutine pool, polls for queued analysis tasks:

1. Load last 48h of execution metrics for the function
2. Aggregate: percentile latencies, error distributions, input patterns, cold start rates
3. Detect regressions: latency trending up, error rate increasing, new bottleneck patterns
4. Call AI Service `/api/dna/analyze` with aggregated metrics + current code
5. If AI returns a variant proposal → store as `function_dna_mutations` with status `proposed`
6. Notify the developer via WebSocket/notification

### AI Service Endpoints (Python)

New routes in `ai-service/src/routes/routes_dna.py`:

- `POST /api/dna/analyze` — Receives aggregated metrics + code, returns optimization analysis
- `POST /api/dna/generate` — Receives analysis + code, returns mutated code variant

Uses existing `ProviderRouter` with preference for DeepInfra (background tasks) or Fireworks (structured output).

---

## Billing / Monetization

| Feature | Tier | Cost |
|---------|------|------|
| DNA profile visibility | All | Free |
| View mutation proposals | All | Free |
| Accept/reject mutations | DNA Pro | Credits per acceptance |
| Full evolution history | DNA Pro | Included |
| DNA Insights dashboard | Enterprise | Included in enterprise plan |
| Marketplace DNA badge | All | Free (trust signal) |

Credit cost per acceptance: **50 credits** (~$0.50)

---

## Integration Points

### Existing Systems Used

| System | Integration |
|--------|-------------|
| `registry_functions` | DNA profiles link to registry functions |
| `function_deployments` | Accepted mutations create new deployments |
| Canary system | Accepted mutations deploy via existing canary flow |
| AI Service | New routes for code analysis and variant generation |
| Notification system | New mutation proposals trigger notifications |
| Billing wallet | Accepting mutations debits credits |
| WebSocket | Real-time analysis progress streaming |
| Marketplace | DNA badge and generation count displayed on function cards |

### New Code Required

| Component | Location | Lines (est.) |
|-----------|----------|-------------|
| DB migration | `migrations/20260502*_function_dna.up.sql` | ~120 |
| DNA repository | `internal/storage/dna/repository.go` | ~400 |
| DNA service | `internal/dna/service.go` | ~500 |
| DNA handler | `internal/api/handlers/dna/handler.go` | ~350 |
| DNA routes | `internal/api/routes_dna.go` | ~60 |
| AI routes | `ai-service/src/routes/routes_dna.py` | ~200 |
| Frontend types | `web/dashboard/src/types/dna.ts` | ~120 |
| Frontend API | `web/dashboard/src/api/dna.ts` | ~100 |
| Frontend hooks | `web/dashboard/src/hooks/useFunctionDNA.ts` | ~150 |
| DNA Helix UI | `web/dashboard/src/components/dna/DNAHelix.tsx` | ~250 |
| Evolution Timeline | `web/dashboard/src/components/dna/EvolutionTimeline.tsx` | ~300 |
| Variant Diff | `web/dashboard/src/components/dna/DNAVariantDiff.tsx` | ~250 |
| Trust Badge | `web/dashboard/src/components/dna/DNATrustBadge.tsx` | ~120 |
| Insights Dashboard | `web/dashboard/src/components/dna/DNAInsightsDashboard.tsx` | ~400 |
| **Total** | | **~3,300** |

---

## Scaling Strategy

- **Execution metrics table**: Partitioned by month, 90-day retention with automatic cleanup
- **Analysis worker**: Configurable concurrency (default 4), backpressure via Redis queue
- **AI calls**: Rate-limited per tenant, cached analysis results for 1 hour
- **DNA profiles**: Read-heavy (cached in Redis with 5-min TTL), write-light (only on analysis)
- **Enterprise insights**: Pre-computed daily via background job, cached aggressively

---

## Security

- All endpoints require JWT auth
- Tenant isolation: `WHERE tenant_id = $claims.TenantID` on every query
- Original/mutated code encrypted at rest (AES-256-GCM, same as vault)
- AI Service calls authenticated via `AI_SERVICE_API_KEY`
- Mutation acceptance requires wallet balance check before proceeding
