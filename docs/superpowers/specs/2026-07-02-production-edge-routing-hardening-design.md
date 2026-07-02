# Production Edge Routing Hardening — Design Spec

**Date:** 2026-07-02
**Status:** Draft
**Scope:** Full production overhaul of the edge routing, health checking, circuit breaking, and observability systems across all 6 backend providers.

---

## Problem Statement

The FunctionFly edge routing system (Caddy → Go Orchestrator → Backend) has architectural gaps that cause production failures:

1. **No `/healthz` fallback** — user functions without `/healthz` become unreachable (circuit opens, backend excluded)
2. **Stale circuit breaker** — `ProxyToBackend` doesn't update circuit state; only the 5-second monitor loop does
3. **DB query per request** — circuit state queried from DB on every proxied request (no in-memory cache)
4. **4 separate circuit breaker implementations** — inconsistent behavior across health monitor, WASM router, DNA service, and agent pipeline
5. **EWMA uses probe latency** — scoring reflects health check probes, not real user request latency
6. **Adapter instances recreated every 5s** — `http.Client` not reused, wasting connections
7. **No health check data retention** — `health_checks` table grows unboundedly
8. **Hardcoded timeouts** — proxy 30s, health checks 25-30s, not configurable

The `saas-starter` app error (-330) is a direct manifestation of gap #1: the Cloudflare Worker at `test-app.microog.workers.dev` returns 404 on `/healthz`, the circuit opens, and the app's only backend is excluded from routing.

---

## Section 1: Shared Circuit Breaker

### Goal
Consolidate 4 separate circuit breaker implementations into one shared package with exponential backoff, DB persistence, Prometheus metrics, and environment-based configuration.

### Package: `internal/circuitbreaker/`

```
internal/circuitbreaker/
  breaker.go       — core Breaker with exponential backoff
  manager.go       — BreakerManager (sync.Map keyed by string/UUID)
  config.go        — Config + env var loading
  metrics.go       — Prometheus gauge/counter integration
  persistence.go   — optional DB sync interface
```

### Core Breaker

Extends the existing `internal/agent/circuitbreaker/breaker.go`:

```go
type Config struct {
    FailureThreshold    int           // CIRCUIT_BREAKER_FAILURE_THRESHOLD (default: 3)
    SuccessThreshold    int           // CIRCUIT_BREAKER_SUCCESS_THRESHOLD (default: 2)
    BaseCooldown        time.Duration // CIRCUIT_BREAKER_BASE_COOLDOWN (default: 30s)
    MaxCooldown         time.Duration // CIRCUIT_BREAKER_MAX_COOLDOWN (default: 5m)
    BackoffMultiplier   float64       // CIRCUIT_BREAKER_BACKOFF_MULTIPLIER (default: 2.0)
    HalfOpenMaxRequests int           // CIRCUIT_BREAKER_HALF_OPEN_MAX (default: 3)
    OnStateChange       func(from, to State)
    Persistence         Persistence   // optional
}
```

**Key differences from existing `agent/circuitbreaker`:**
- `CooldownDuration` → `BaseCooldown` with exponential backoff: `base * (multiplier ^ reopenCount)`
- Track `reopenCount int` — increments each time circuit reopens from half-open after failure
- `ProbeAllow() bool` — always returns true (for health monitoring), doesn't count against half-open limit
- `Allow() bool` — respects open state (for routing/gating)
- `Snapshot() StateInfo` — returns current state for health endpoints

**DB Persistence:**

```go
type Persistence interface {
    Load(ctx context.Context, key string) (*StoredState, error)
    Save(ctx context.Context, key string, state *StoredState) error
}
```

- On state transition: async-write to DB with 1-second coalescing (buffer writes, deduplicate)
- On `New()`: load initial state from DB if available
- Keeps in-memory fast path while maintaining cross-instance consistency

### BreakerManager

```go
type Manager struct {
    breakers sync.Map // map[string]*Breaker
    config   Config
    persist  Persistence
}

func (m *Manager) For(key string) *Breaker
func (m *Manager) ForBackend(id uuid.UUID) *Breaker
func (m *Manager) ForProvider(provider string) *Breaker
```

Eliminates each caller managing their own map. Health monitor, WASM router, and DNA service all use `manager.ForBackend(backendID)`.

### Migration Path

| Current location | Migrates to |
|---|---|
| `internal/health/monitor.go` (inline circuit logic) | `circuitbreaker.Manager.ForBackend()` |
| `internal/agent/circuitbreaker/breaker.go` | Thin wrapper or import alias to `internal/circuitbreaker` |
| `internal/wasm/router.go` (custom breaker) | `circuitbreaker.Manager.For("wasm:" + runtimeType)` |
| `internal/dna/service.go` (inline int-based) | `circuitbreaker.Manager.For("dna")` |

---

## Section 2: Health Check Hardening

### Goal
Ensure health checks work for all backend types, even when user functions don't implement `/healthz`. Reduce wasted resources. Add data retention.

### 2a. `/healthz` Fallback Chain

Each adapter defines its own fallback via the `common.ProviderAdapter` interface:

```go
type ProviderAdapter interface {
    HealthCheck(ctx, backend) (*HealthCheckResult, error)           // existing
    HealthCheckFallback(ctx, backend) (*HealthCheckResult, error)    // new — optional
}
```

Provider-specific fallback strategies:

| Provider | Primary (`/healthz`) | Fallback |
|---|---|---|
| Cloudflare Workers | `GET /healthz` | `GET /` on worker URL (workers always respond) |
| Fly.io | `GET /healthz` | `GET /` on app URL |
| Vercel | `GET /healthz` | `GET /` on deployment URL |
| Deno Deploy | `GET /healthz` | `GET /` on deployment URL |
| AWS Lambda | SDK `ListFunctions` | No fallback (SDK-based already) |
| FunctionFly Edge | `GET /healthz` | TCP connect to runtime port |

Fallback is triggered when primary returns 404. 5xx responses do NOT trigger fallback (treated as genuinely unhealthy).

`HealthCheckResult` gains a new field: `Degraded bool` — set when fallback succeeds but `/healthz` didn't. The router penalizes degraded backends in scoring (1.5x EWMA multiplier) without excluding them.

Config: `HEALTH_CHECK_FALLBACK=true` (default: true)

### 2b. Adaptive Probe Interval

If a backend has failed N consecutive health checks, reduce probe frequency:

| Consecutive failures | Probe interval |
|---|---|
| 0-2 | 5s (default) |
| 3-10 | 15s |
| 11-30 | 60s |
| 30+ | 5min |

Reset to 5s on first success. Tracked per-backend in the Monitor struct.

### 2c. Adapter Instance Pooling

```go
var adapterPool sync.Map // map[string]common.ProviderAdapter

func (m *Monitor) getAdapterForProvider(provider string) common.ProviderAdapter {
    if v, ok := m.adapterPool.Load(provider); ok {
        return v.(common.ProviderAdapter)
    }
    adapter := createAdapter(provider)
    m.adapterPool.Store(provider, adapter)
    return adapter
}
```

Each adapter's `http.Client` is reused across probe cycles, preserving connection pools.

### 2d. Health Check Result Caching

Cache the last health check result per backend with a 3-second TTL:

```go
type cachedResult struct {
    result    *common.HealthCheckResult
    cachedAt  time.Time
}
```

The router checks the cache first. If the result is fresh (< 3s), use it. If stale, fall back to circuit breaker state. Eliminates DB queries for routing decisions under high traffic.

### 2e. Health Response Body Parsing

Parse structured responses when available:

```json
{"status": "healthy", "version": "1.2.3", "uptime": 3600}
```

Store `version` and `uptime` in the health check record. Gives richer observability without provider-specific logic in the router.

### 2f. Data Retention

Cleanup goroutine in the monitor, runs once per hour:

```go
func (m *Monitor) cleanupLoop() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        m.repo.DeleteHealthChecksBefore(ctx, time.Now().AddDate(0, 0, -retentionDays))
    }
}
```

Config: `HEALTH_CHECK_RETENTION_DAYS=7` (default: 7)

---

## Section 3: Routing Accuracy

### Goal
Route based on real user request latency, not synthetic probes. Update circuit breaker immediately on proxy failures. Cache routing decisions.

### 3a. Wire `RecordRoutingResult()` into ProxyToBackend

`ProxyToBackend` gets two new interface parameters:

```go
type RoutingRecorder interface {
    RecordRoutingResult(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
}

type CircuitRecorder interface {
    RecordFailure(backendID uuid.UUID, err error)
}
```

After each proxy attempt (success or failure):

```go
if recorder != nil {
    recorder.RecordRoutingResult(appID, backend.ID, int(latency.Milliseconds()), outcome, requestID)
}
```

### 3b. EWMA from Routing Events

Add `GetRecentRoutingEvents()` to the `RouterRepository` interface in `internal/routing/router.go`:

```go
type RouterRepository interface {
    // existing methods...
    GetRecentRoutingEvents(ctx context.Context, backendID uuid.UUID, limit int) ([]*RoutingEvent, error)
}
```

Change `Router.calculateEWMAScore()` to read from `routing_events` table:

```go
func (r *Router) calculateEWMAScore(backendID uuid.UUID) float64 {
    events, err := r.repo.GetRecentRoutingEvents(ctx, backendID, 20)
    // ... EWMA on real latencies
    // Fall back to health_checks if no routing events exist
}
```

**Shadow routing mode** for safe migration:

- `EWMA_SOURCE=health` (default) — use health check latency
- `EWMA_SOURCE=real` — use routing event latency
- When `EWMA_SOURCE=health`, still compute and log real EWMA for validation
- After 48 hours of stable data, flip to `real`

### 3c. Proxy Failure → Immediate Circuit Breaker Update

When `ProxyToBackend` gets a connection error or 5xx, immediately record the failure:

```go
if circuitRecorder != nil {
    circuitRecorder.RecordFailure(backend.ID, fmt.Errorf("proxy failure: %d", resp.StatusCode))
}
```

The health monitor's 5-second probe loop continues as a secondary signal. The proxy provides the primary signal.

### 3d. Routing Decision Cache

Cache routing decisions per app ID with a 1-second TTL:

```go
type decisionCache struct {
    cache sync.Map // map[uuid.UUID]*cachedDecision
    ttl   time.Duration
}
```

Cache invalidates on circuit breaker state transitions (via `OnStateChange` callback). Eliminates repeated DB queries for circuit state and EWMA scoring under load.

### 3e. Latency Percentile Tracking

In-memory `LatencyTracker` per backend — rolling window of last 100 latencies:

```go
type LatencyTracker struct {
    window []int64
    pos    int
    mu     sync.RWMutex
}
func (t *LatencyTracker) Record(latencyMs int64)
func (t *LatencyTracker) Percentiles() (p50, p95, p99 int64)
```

Exposed via Prometheus:

```
functionfly_backend_latency_p50_ms{backend_id="..."} 12
functionfly_backend_latency_p95_ms{backend_id="..."} 45
functionfly_backend_latency_p99_ms{backend_id="..."} 120
```

### 3f. Configurable Proxy Timeout

```go
timeout := env.Duration("PROXY_TIMEOUT_MS", 30*time.Second)
client := &http.Client{Timeout: timeout}
```

Reuse `http.Client` per backend (connection pooling) instead of creating a new one per request.

---

## Section 4: Edge Caddy Hardening + Observability

### Goal
Harden the Caddy edge proxy for production. Add Prometheus alerts. Implement distributed tracing.

### 4a. Caddy Edge Hardening

**`deploy/edge/Caddyfile.edge` additions:**

```caddyfile
http://*.localhost:8082 {
    request_body {
        max_size 10MB
    }

    reverse_proxy 127.0.0.1:8080 {
        header_up X-FF-Slug {labels.1}
        transport http {
            dial_timeout 5s
            response_header_timeout 30s
            read_timeout 60s
            write_timeout 60s
        }
        health_uri /health
        health_interval 10s
        health_timeout 5s
    }

    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        Referrer-Policy strict-origin-when-cross-origin
        -Server
    }
}
```

**Production `Caddyfile.edge` additions:**
- `grace_period 30s` for connection draining on shutdown
- TLS: `protocols tls1.2 tls1.3`, modern cipher suite
- Structured JSON access logs

### 4b. Rate Limiting in Go Middleware

Implement per-tenant rate limiting in Go (not Caddy) using Redis:

```go
// Per-plan limits:
// Starter: 100 req/s
// Professional: 1000 req/s
// Enterprise: 10000 req/s
```

Integrates with the existing plan system and tenant request counting in `handlePublicRoute`.

### 4c. Prometheus Alerts

```yaml
- alert: CircuitBreakerOpen
  expr: functionfly_circuit_breaker_state{circuit_breaker_state="open"} > 0
  for: 1m
  severity: critical

- alert: AllBackendsDown
  expr: min by (app_id) (functionfly_circuit_breaker_state{circuit_breaker_state="closed"}) == 0
  for: 30s
  severity: critical

- alert: HighBackendLatency
  expr: functionfly_backend_latency_p99_ms > 5000
  for: 5m
  severity: warning

- alert: HealthCheckFailureRate
  expr: rate(functionfly_health_check_failures_total[5m]) > 0.5
  for: 2m
  severity: warning
```

### 4d. OpenTelemetry Distributed Tracing

Leverage existing `X-Functionfly-Trace-Id` and `Traceparent` headers. Add OTel spans:

| Span name | Attributes |
|---|---|
| `edge-middleware` | app_slug, rewrite_from, rewrite_to |
| `handle-public-route` | app_id, tenant_id, tenant_plan |
| `select-backend` | selected_backend_id, ewma_score, circuit_state, reason |
| `proxy-to-backend` | backend_id, provider, region, status_code, latency_ms, outcome |

Export via OTLP to Jaeger/Tempo/Grafana (configurable via `OTEL_EXPORTER_OTLP_ENDPOINT`).

**Trace propagation in ProxyToBackend:**
- Copy `Traceparent` from original request
- Generate new `X-FunctionFly-Span-ID` for the proxy hop
- Set `X-FunctionFly-Request-ID` (already done)

### 4e. Access Log Sampling

In production, don't log every routing decision to stdout:
- 100% of failures, errors, and failovers
- 10% of successful requests (configurable via `ROUTING_LOG_SAMPLE_RATE=0.1`)

OTel handles this naturally with its own sampling configuration.

---

## Configuration Reference

| Variable | Default | Description |
|---|---|---|
| `CIRCUIT_BREAKER_FAILURE_THRESHOLD` | 3 | Failures before opening circuit |
| `CIRCUIT_BREAKER_SUCCESS_THRESHOLD` | 2 | Successes to close from half-open |
| `CIRCUIT_BREAKER_BASE_COOLDOWN` | 30s | Base cooldown before half-open |
| `CIRCUIT_BREAKER_MAX_COOLDOWN` | 5m | Max cooldown with backoff |
| `CIRCUIT_BREAKER_BACKOFF_MULTIPLIER` | 2.0 | Exponential backoff multiplier |
| `CIRCUIT_BREAKER_HALF_OPEN_MAX` | 3 | Max requests in half-open |
| `HEALTH_CHECK_INTERVAL` | 5s | Base probe interval |
| `HEALTH_CHECK_TIMEOUT` | 30s | Per-probe timeout |
| `HEALTH_CHECK_FALLBACK` | true | Enable `/healthz` fallback chain |
| `HEALTH_CHECK_RETENTION_DAYS` | 7 | Days to keep health check records |
| `PROXY_TIMEOUT_MS` | 30s | Backend proxy timeout |
| `EWMA_SOURCE` | health | `health` or `real` latency for EWMA |
| `ROUTING_LOG_SAMPLE_RATE` | 0.1 | Fraction of successful requests to log |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | (unset) | OTel collector endpoint |
| `REQUEST_BODY_MAX_SIZE` | 10MB | Max request body size |

---

## Migration Strategy

1. **Phase 1 — Shared circuit breaker** (no behavior change): Extract `internal/circuitbreaker/`, migrate all 4 implementations to use it. DB schema unchanged. All existing tests pass.

2. **Phase 2 — Health check hardening**: Add fallback chain, adapter pooling, adaptive intervals, data retention. Existing health checks continue to work; fallback adds resilience.

3. **Phase 3 — Routing accuracy**: Wire `RecordRoutingResult()` into proxy, add shadow EWMA mode. Default `EWMA_SOURCE=health` preserves existing behavior. Flip to `real` after validation.

4. **Phase 4 — Observability**: Add OTel spans, Prometheus alerts, Caddy hardening. No behavior change to routing logic.

Each phase is independently deployable and rollback-safe.

---

## Affected Files

| File | Change |
|---|---|
| `internal/circuitbreaker/` (new) | Shared circuit breaker package |
| `internal/agent/circuitbreaker/breaker.go` | Redirect to shared package |
| `internal/health/monitor.go` | Use shared breaker, adapter pooling, adaptive intervals, data retention |
| `internal/wasm/router.go` | Use shared breaker |
| `internal/dna/service.go` | Use shared breaker |
| `internal/routing/router.go` | EWMA from routing events, decision cache, percentile tracking |
| `internal/api/utils/proxy.go` | Circuit recorder interface, routing recorder, configurable timeout, http.Client reuse |
| `internal/api/routing.go` | Pass recorder interfaces to ProxyToBackend |
| `internal/api/server.go` | Wire OTel, rate limiter middleware |
| `internal/adapters/common/adapter.go` | Add `HealthCheckFallback` to interface |
| `internal/adapters/cloudflare/adapter.go` | Implement fallback |
| `internal/adapters/fly/adapter.go` | Implement fallback |
| `internal/adapters/vercel/adapter.go` | Implement fallback |
| `internal/adapters/deno/adapter.go` | Implement fallback |
| `internal/adapters/functionfly/adapter.go` | Implement fallback |
| `internal/adapters/aws/adapter.go` | No fallback (SDK-based) |
| `deploy/edge/Caddyfile.edge` | Hardening |
| `deploy/edge/Caddyfile.local` | Hardening |
| `deploy/edge/Caddyfile.lb` | Hardening |
| `migrations/` | `routing_logs` table, index on `routing_events` |
