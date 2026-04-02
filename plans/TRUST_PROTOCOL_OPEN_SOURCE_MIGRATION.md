# Trust Protocol — Migration to Separate Open-Source Repository

**Date**: 2026-03-29
**Status**: Postponed (Year 2)
**Target repo**: `github.com/functionfly/trust-protocol` (Go module: `github.com/functionfly/trust-protocol`)

**Note (2026-03-31):** Open-sourcing the Trust Protocol is deferred until year 2. Trust Protocol code remains **internal** to this repository (`internal/functionregistry/`, `internal/api/handlers/trustapi/`, `internal/storage/trustapi/`, `internal/scheduler/`, etc.). The separate `trust-protocol` repo is on hold; FunctionFly does not depend on it as an external Go module.

---

## Executive Summary

The Trust Protocol is a standalone, standards-grade system for evaluating, verifying, and communicating function trustworthiness. It has strong value as an independent open-source project that other platforms can adopt. This plan describes extracting all Trust Protocol code from `functionfly/functionfly` into a new `functionfly/trust-protocol` repository, with minimal coupling back to the FunctionFly platform. **Execution of that extraction is postponed until year 2**; until then, keep the implementation in this monorepo only.

The new repo will contain:
1. **Core trust scoring engine** (v1 + v2 with DRE sub-scores)
2. **Trust API** (external B2B2B HTTP API with partner management)
3. **Verification pipeline types and models**
4. **Trust data models** (GORM models, DTOs, response types)
5. **Trust score scheduler**
6. **Protocol specification docs**
7. **SDK examples** (Python LangChain/CrewAI/AutoGen adapters)

---

## 1. Scope: What Moves to `trust-protocol`

### 1.1 Core Go Packages (extract as-is, change import paths)

| Source (in `functionfly/functionfly`) | Target (in `functionfly/trust-protocol`) | Notes |
|---|---|---|
| `internal/functionregistry/trust_score.go` | `pkg/scoring/calculator.go` | Pure computation, no DB deps |
| `internal/functionregistry/trust_score_test.go` | `pkg/scoring/calculator_test.go` | Unit tests move with it |
| `internal/storage/trustapi/models.go` | `pkg/models/trustapi.go` | All GORM models + DTOs |
| `internal/storage/trustapi/repository.go` | `pkg/repository/trustapi.go` | DB operations for Trust API |
| `internal/api/handlers/trustapi/handler.go` | `pkg/httpapi/handler.go` | HTTP handler + route registration |
| `internal/api/handlers/trustapi/trust.go` | `pkg/httpapi/trust_endpoints.go` | Trust score/history/verify/report handlers |
| `internal/api/handlers/trustapi/middleware.go` | `pkg/httpapi/middleware.go` | API key auth, rate limiting, usage tracking |
| `internal/api/handlers/trustapi/partners.go` | `pkg/httpapi/partners.go` | Partner CRUD + API key management |
| `internal/api/routes_trustapi.go` | `pkg/httpapi/routes.go` | Route wiring (adapted to standalone) |
| `internal/scheduler/trust_score_scheduler.go` | `pkg/scheduler/scheduler.go` | Cron-based trust score recalculation |

### 1.2 Shared Types (must be extracted or redefined)

These types currently live in `internal/storage/registry/types.go` and are used by the trust code. They need to be either **extracted** into the new repo or **redefined** with a thin interface.

| Type | Lines in `types.go` | Used By | Strategy |
|---|---|---|---|
| `TrustTier` (string enum) | 551-559 | calculator, repository, handlers, scheduler | **Extract** — canonical in new repo |
| `TrustScoreWeights` | 561-578 | calculator, repository | **Extract** — canonical in new repo |
| `TrustHistory` | 581-613 | repository, handlers, scheduler | **Extract** — canonical in new repo |
| `ExecutionMetrics` | 616-651 | repository, scheduler | **Extract** — canonical in new repo |
| `TrustScoreWeightsConfig` | 653-667 | repository | **Extract** |
| `TrustScoreJob` | 669-685 | repository, scheduler | **Extract** |
| `TrustScoreResponse` | 687-718 | handlers | **Extract** |
| `TrustHistoryResponse` | 720-727 | handlers | **Extract** |
| `DREScores` | 459-465 | calculator (v2), repository | **Extract** |
| `RegistryFunctionRating` | 192-225 | calculator (v1+2) | **Interface** — define `TrustMetricsProvider` interface |
| `RegistryFunction` | 22-50 | calculator (deterministic check) | **Interface** — define `FunctionInfo` interface |
| `RegistryFunctionVersion` | 90-100 | repository (verification status) | **Interface** — define `FunctionVersionInfo` interface |
| `RegistryFunctionVerificationStatus` | — | repository | **Interface** |
| Verification pipeline types (Job, Result, LevelConfig, AuditLog, Schedule, ManualReviewQueue) | 733-890 | verification pipeline | **Extract** — these are trust-protocol specific |

### 1.3 Database Migrations (move to new repo)

| Migration File | Tables Created |
|---|---|
| `migrations/20260331000000_trust_scoring_system.up.sql` | `trust_history`, `execution_metrics`, `trust_score_weights`, `trust_score_jobs`, adds `trust_score`/`trust_tier` to `registry_functions` |
| `migrations/000036_trust_score_fields.up.sql` | Trust score columns on `registry_function_ratings` |
| `migrations/000112_trust_score_fields.up.sql` | Additional trust score fields |
| `internal/storage/sql/migrations/20260321000000_create_trust_api_tables.up.sql` | `trust_api_partners`, `trust_api_keys`, `trust_api_usage`, `trust_api_rate_limits`, `trust_api_reports`, `trust_api_verifications` |

Also move the corresponding `.down.sql` files.

### 1.4 Documentation (move to new repo)

| Source | Target |
|---|---|
| `web/docs/src/content/docs/trust-protocol-spec.md` | `docs/spec/trust-protocol-spec.md` |
| `web/docs/src/content/docs/trust-api.md` | `docs/api/trust-api.md` |
| `web/docs/src/content/docs/trust-protocol-open-source.md` | `README.md` (adapted as main README) |
| `plans/TRUST_SCORING_SYSTEM.md` | `docs/implementation/trust-scoring-system.md` |
| `plans/fxcert_implementation_plan.md` | `docs/implementation/fxcert.md` |
| `plans/DRE_2_PROTOCOL.md` | `docs/implementation/dre-2-protocol.md` |
| `plans/DRE_2_GAP_ANALYSIS.md` | `docs/implementation/dre-2-gap-analysis.md` |

### 1.5 SDK Examples (move to new repo)

| Source | Target |
|---|---|
| `sdk/python/examples/langchain_trusted_tools.py` | `examples/python/langchain_trusted_tools.py` |
| `sdk/python/examples/crewai_trusted_tools.py` | `examples/python/crewai_trusted_tools.py` |
| `sdk/python/examples/autogen_trusted_tools.py` | `examples/python/autogen_trusted_tools.py` |

### 1.6 Marketing / Blog (stays in functionfly, not extracted)

| File | Action |
|---|---|
| `web/site/src/pages/trust.astro` | **Keep** — marketing site is FunctionFly-specific |
| `cmd/blog-api/src/data/default-posts/trust-layer-for-ai-agents.ts` | **Keep** — blog is FunctionFly-specific |

---

## 2. New Repository Structure

```
trust-protocol/
├── go.mod                          # module github.com/functionfly/trust-protocol
├── go.sum
├── LICENSE                         # Apache-2.0 or MIT (for open-source adoption)
├── README.md                       # From trust-protocol-open-source.md
├── CONTRIBUTING.md
├── docs/
│   ├── spec/
│   │   └── trust-protocol-spec.md  # Formal protocol specification
│   ├── api/
│   │   └── trust-api.md            # HTTP API reference
│   └── implementation/
│       ├── trust-scoring-system.md
│       ├── fxcert.md
│       ├── dre-2-protocol.md
│       └── dre-2-gap-analysis.md
├── pkg/
│   ├── scoring/                    # Trust score calculation engine
│   │   ├── calculator.go           # TrustScoreCalculator (v1 + v2)
│   │   ├── calculator_test.go
│   │   └── types.go                # TrustMetrics, TrustScoreResult, TrustMetricsV2, etc.
│   ├── models/                     # Data models (GORM + DTOs)
│   │   ├── trustapi.go             # Partner, APIKey, Usage, RateLimit, Report, Verification
│   │   ├── trust.go                # TrustHistory, ExecutionMetrics, TrustScoreJob, TrustTier
│   │   ├── verification.go         # VerificationJob, VerificationResult, ManualReviewQueue, etc.
│   │   └── interfaces.go           # Interfaces for cross-repo integration
│   ├── repository/                 # Database operations
│   │   └── trustapi.go             # Trust API repository (partners, keys, usage, etc.)
│   ├── httpapi/                    # HTTP handlers + middleware
│   │   ├── handler.go              # Handler struct + constructor
│   │   ├── routes.go               # Route registration
│   │   ├── trust_endpoints.go      # Trust score/history/verify/report handlers
│   │   ├── partners.go             # Partner + API key CRUD
│   │   └── middleware.go           # Auth, rate limiting, usage tracking
│   └── scheduler/                  # Trust score recalculation
│       └── scheduler.go            # Cron-based scheduler
├── migrations/
│   ├── 0001_trust_scoring_system.up.sql
│   ├── 0001_trust_scoring_system.down.sql
│   ├── 0002_trust_score_fields.up.sql
│   ├── 0002_trust_score_fields.down.sql
│   ├── 0003_trust_api_tables.up.sql
│   └── 0003_trust_api_tables.down.sql
├── examples/
│   └── python/
│       ├── langchain_trusted_tools.py
│       ├── crewai_trusted_tools.py
│       └── autogen_trusted_tools.py
└── cmd/                            # Optional: standalone Trust API server
    └── trust-api/
        └── main.go
```

---

## 3. Interfaces for Cross-Repository Integration

The new `trust-protocol` repo must NOT import `functionfly/functionfly`. Instead, define interfaces that the FunctionFly platform implements:

```go
// pkg/models/interfaces.go

package models

import "github.com/google/uuid"

// FunctionInfo provides the minimal function metadata needed for trust scoring.
// FunctionFly's RegistryFunction implements this.
type FunctionInfo interface {
    GetID() uuid.UUID
    GetCreatedAt() time.Time
    IsDeterministic() bool
    GetDeterministicScore() float64
}

// FunctionRatingProvider provides execution metrics and ratings for trust calculation.
// FunctionFly's RegistryFunctionRating implements this.
type FunctionRatingProvider interface {
    GetSuccessRate() float64
    GetP50LatencyMs() int
    GetP95LatencyMs() int
    GetAvgLatencyMs() int
    GetTimeoutRate() float64
    GetErrorRate() float64
    GetTotalRatings() int
    GetTenantDiversity() int
    GetUserDiversity() int
    GetConsumerDiversity() float64
}

// FunctionVersionInfo provides verification status for trust scoring.
type FunctionVersionInfo interface {
    GetID() uuid.UUID
    GetFunctionID() uuid.UUID
    IsVerified() bool
    GetVerificationLevel() string
}
```

FunctionFly implements these interfaces via thin adapter methods (or struct embedding) — no import of `trust-protocol` types into the registry package's own types needed.

---

## 4. What StAYS in `functionfly/functionfly`

### 4.1 Internal Trust Integration Points

These files in `functionfly/functionfly` call into trust scoring and need to be updated to import from the new repo:

| File | Change Required |
|---|---|
| `internal/api/routes.go:311-327` | Import `trust-protocol/pkg/scheduler`, instantiate with functionfly's registry repo adapter |
| `internal/api/routes.go:527` | Import `trust-protocol/pkg/httpapi`, register trust routes |
| `internal/api/routes_registry.go:217-221` | Keep registry trust endpoints, but they call through to the shared trust engine |
| `internal/api/handlers/registry/trust.go` | Refactor to call `trust-protocol/pkg/scoring` calculator |
| `internal/api/handlers/registry/query.go` | Uses `DREScores` type — import from `trust-protocol/pkg/models` |
| `internal/api/handlers/registry/execution/handlers.go` | Uses `DREScores` — import from `trust-protocol/pkg/models` |
| `internal/api/handlers/registry/publish.go` | Uses `DREScores` — import from `trust-protocol/pkg/models` |
| `internal/storage/registry/trust_repository.go` | Refactor to delegate to `trust-protocol/pkg/repository` or keep as adapter layer |
| `internal/storage/registry/dre_repository.go` | Uses `DREScores` — import from `trust-protocol/pkg/models` |
| `internal/storage/registry_models.go:37` | Type alias `DREScores = trustmodels.DREScores` — update import |

### 4.2 Registry Types That Stay (with thin adapters)

`RegistryFunction`, `RegistryFunctionRating`, `RegistryFunctionVersion` — these are core FunctionFly platform types. They stay in `internal/storage/registry/types.go` but gain interface-satisfying methods:

```go
// In functionfly — adapter methods on RegistryFunction
func (f *RegistryFunction) GetID() uuid.UUID       { return f.ID }
func (f *RegistryFunction) GetCreatedAt() time.Time { return f.CreatedAt }
func (f *RegistryFunction) IsDeterministic() bool   { return f.DeterministicScore > 0 }
func (f *RegistryFunction) GetDeterministicScore() float64 { return f.DeterministicScore }

// In functionfly — adapter methods on RegistryFunctionRating
func (r *RegistryFunctionRating) GetSuccessRate() float64    { return r.SuccessRate }
func (r *RegistryFunctionRating) GetP50LatencyMs() int       { return r.P50LatencyMs }
func (r *RegistryFunctionRating) GetP95LatencyMs() int       { return r.P95LatencyMs }
// ... etc
```

### 4.3 Things That Explicitly Do NOT Move

| Component | Reason |
|---|---|
| `internal/dre/` | DRE (Deterministic Replay Engine) is a separate protocol/component, not Trust-specific. Trust *consumes* DRE scores but DRE has its own lifecycle. |
| `internal/storage/registry/function_crud.go` | Core registry CRUD — platform-specific |
| `internal/api/handlers/registry/` (non-trust) | Platform-specific handlers |
| `web/dashboard/` | Frontend stays with FunctionFly |
| `web/site/` | Marketing site stays with FunctionFly |
| `cmd/blog-api/` | Blog stays with FunctionFly |

---

## 5. Migration Steps (Execution Order)

### Phase 1: Scaffold the new repo
1. Create `github.com/functionfly/trust-protocol` repo
2. Initialize Go module: `github.com/functionfly/trust-protocol`
3. Copy all files listed in Section 1
4. Rename packages per the new structure
5. Resolve all internal imports — trust-protocol code should only import itself + external deps
6. Create `cmd/trust-api/main.go` as a standalone server entry point
7. Add `go.mod` with external dependencies:
   - `github.com/google/uuid`
   - `github.com/gorilla/mux`
   - `github.com/sirupsen/logrus`
   - `gorm.io/gorm`
   - `gorm.io/driver/postgres`
   - `github.com/robfig/cron/v3`
8. Ensure `go build ./...` and `go test ./...` pass

### Phase 2: Define interfaces
1. Create `pkg/models/interfaces.go` with `FunctionInfo`, `FunctionRatingProvider`, `FunctionVersionInfo`
2. Update `pkg/scoring/calculator.go` to accept interfaces instead of concrete `storage.*` types
3. Update `pkg/httpapi/handler.go` to accept interface-based registry dependency
4. Update tests

### Phase 3: Update `functionfly/functionfly`
1. Add `go.mod` dependency: `require github.com/functionfly/trust-protocol v0.x.x`
2. Create adapter methods on `RegistryFunction`, `RegistryFunctionRating` to satisfy trust-protocol interfaces
3. Replace inline trust types with imports from `trust-protocol/pkg/models`:
   - `TrustTier`, `TrustHistory`, `ExecutionMetrics`, `DREScores`, `TrustScoreJob`, etc.
   - Use type aliases where needed for backward compatibility: `type TrustTier = trustmodels.TrustTier`
4. Refactor `internal/api/routes.go` to import and use `trust-protocol/pkg/scheduler` and `trust-protocol/pkg/httpapi`
5. Refactor `internal/storage/registry/trust_repository.go` to delegate to trust-protocol where possible, or keep as thin adapter
6. Remove migrated files:
   - `internal/functionregistry/trust_score.go` (replaced by `pkg/scoring/`)
   - `internal/functionregistry/trust_score_test.go`
   - `internal/storage/trustapi/` (entire package — replaced by `trust-protocol/pkg/`)
   - `internal/api/handlers/trustapi/` (entire package — replaced by `trust-protocol/pkg/httpapi/`)
   - `internal/api/routes_trustapi.go`
   - `internal/scheduler/trust_score_scheduler.go`
7. Remove migrated migrations (or mark as handled by trust-protocol)
8. Run `go build ./...` and `go test ./...`

### Phase 4: Publish & document
1. Tag `trust-protocol` v0.1.0
2. Update FunctionFly docs to link to `trust-protocol` repo
3. Update `web/docs/src/content/docs/trust-protocol-open-source.md` to point to the new repo
4. Update SDK examples to import from `trust-protocol`

---

## 6. Dependency Graph

```
trust-protocol (standalone)
  ├── external: uuid, mux, logrus, gorm, cron/v3
  └── NO dependency on functionfly/functionfly

functionfly/functionfly (platform)
  ├── depends on: trust-protocol/pkg/scoring
  ├── depends on: trust-protocol/pkg/models
  ├── depends on: trust-protocol/pkg/httpapi
  ├── depends on: trust-protocol/pkg/scheduler
  └── implements: trust-protocol FunctionInfo, FunctionRatingProvider interfaces
```

---

## 7. Risk Assessment

| Risk | Mitigation |
|---|---|
| **Circular dependency** | trust-protocol has zero imports from functionfly. Interfaces break the cycle. |
| **Type drift** | Use type aliases in functionfly: `type TrustTier = trustmodels.TrustTier`. Single source of truth in trust-protocol. |
| **Breaking API contract** | Trust API endpoints remain at `/v1/trust/*` and `/v1/partners/*`. No URL changes. |
| **Migration ordering** | Trust-protocol migrations run first (lower sequence numbers). FunctionFly's `--skip-migrations` flag handles conflicts. |
| **Version pinning** | Use Go module version pinning. FunctionFly pins to specific trust-protocol release. |
| **DRE scores coupling** | `DREScores` type moves to trust-protocol. DRE engine (`internal/dre/`) stays in functionfly. DRE produces scores, trust-protocol consumes them. |

---

## 8. Open Questions

1. **License**: Apache-2.0 (permissive, standard for Go libraries) or MIT? Recommendation: Apache-2.0 for patent protection.
2. **Standalone server**: Should `cmd/trust-api/` ship as a standalone binary, or is trust-protocol purely a library? Recommendation: Ship both — library for integration, binary for standalone deployment.
3. **Versioning**: Start at v0.1.0 (pre-1.0 allows breaking changes) or v1.0.0 (matches the spec version)? Recommendation: v0.1.0 until the protocol spec is finalized.
4. **Registry trust endpoints**: The `/v1/registry/functions/{id}/trust` endpoints are FunctionFly-specific (use function author/name slugs). These should stay in functionfly as thin wrappers around the trust-protocol scoring engine.
