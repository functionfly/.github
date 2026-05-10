# Playground Execution Flow — Production Recovery Plan

**Status:** Draft  
**Target:** Launch-ready by 2026-05-15  
**Owner:** Lead System Architect (FunctionFly)  
**Scope:** Backend execution pipeline, SAR runtime, frontend playground, database schema

---

## 1. Executive Summary

The playground "Execution Failed" error and `data: null` responses stem from **six interconnected failures** across the Go backend, Rust SAR runtime, and frontend response mapping. This plan provides a **phased fix strategy** with immediate hotfixes (P0), hardening (P1), and a 2026-scalable architecture (P2).

---

## 2. Root Cause Analysis

### Issue A — Input Corruption in Daemon Mode (P0)
**File:** `internal/api/handlers/registry/execution/sandbox.go:759`  
**Bug:** `json.Marshal(input)` where `input` is `json.RawMessage` (`[]byte`) produces a JSON array of byte values instead of the original JSON string.  
**Impact:** All daemon-mode executions receive garbage input, causing silent handler failures and `null` output.  
**Fix:** `inputBytes := input` (raw bytes, no re-marshal). *(Already applied — verify in build).*

### Issue B — Lazy Bundling Sends Unusable WASM for Python (P0)
**File:** `internal/api/handlers/registry/execution/sandbox.go:934-1020`  
**Bug:** For `python3.12` (non-microvm), `executeWithLazyBundling` calls `bundler.BundleForWasmRuntimeWithWorkingDirectory()`, which returns raw `micropython.wasm` **without embedded user code**. The engine treats it as generic WASM, but micropython.wasm has no `handler` export.  
**Impact:** Python functions execute the empty MicroPython runtime and return `null`.  
**Fix:** Route all `python*` runtimes through the direct source-code path (same as `python-microvm`), bypassing the broken WASM bundler for daemon mode. *(Already applied — verify).*

### Issue C — RustPython Wrapper Never Calls `handler()` (P0)
**File:** `runtimes/local/src/python/runtime.rs` (all 3 wrapper sites)  
**Bug:** The wrapper defines the user's `handler` function but never invokes it. The last expression is a `def` statement, which evaluates to `None`.  
**Impact:** Every Python execution returns `None` → `data: null`.  
**Fix:** Append `handler(input_data)` as the final expression in all three wrapper templates. *(Already applied — verify build).*

### Issue D — RustPython `import json` Fails in `without_stdlib()` Mode (P0)
**File:** `runtimes/local/src/python/runtime.rs:172`  
**Bug:** `vm::Interpreter::without_stdlib()` starts RustPython with **no standard library modules**. The wrapper hardcodes `import json`, which raises `ModuleNotFoundError` before any user code runs.  
**Impact:** **All Python executions fail** with a swallowed `PyBaseException`, explaining the persistent "Execution Failed" message.  
**Fix Options:**
1. **Hotfix:** Remove `import json` from the wrapper; pass raw string to `handler()` and let user functions parse JSON themselves.
2. **Proper:** Switch to `Interpreter::default()` (with stdlib) or pre-load `json` into the VM scope.
3. **2026:** Replace RustPython with a WASI-compiled CPython or Deno runtime that has full stdlib support.

### Issue E — Playground Proxy Uses Wrong Port (P1)
**File:** `internal/api/handlers/registry/playground.go:106-110`  
**Bug:** `serverPort := os.Getenv("PORT")` with fallback to `"8090"`, but the orchestrator listens on `8080`. The playground makes internal HTTP calls to the wrong port, causing connection refused / timeout.  
**Impact:** Playground UI shows "unknown error" even when direct API calls work.  
**Fix:** Read the actual listen address from the server config or default to `8080`.

### Issue F — Frontend Response Format Mismatch (P1)
**File:** `web/dashboard/src/pages/PlaygroundPage/store/playgroundStore.ts`  
**Bug:** Frontend expects `PlaygroundExecuteResponse` (`{success, output, latency_ms}`) but the execution endpoint returns `ExecutionResponse` (`{ok, data, duration_ms}`).  
**Impact:** Successful executions are rendered as failures.  
**Fix:** Implement a mapper function that normalizes both formats (partially done — needs completion and unit tests).

### Issue G — Database Schema Drift (P0)
**File:** `internal/storage/registry/types.go:418-419`  
**Bug:** `RegistryFunctionVerificationStatus` model expects `BlockedAt`/`BlockReason` columns that don't exist in the table.  
**Impact:** Queries crash with `pq: column "blocked_at" does not exist`.  
**Fix:** Migration `20260509172354_add_blocking_columns_to_verification_status.up.sql` *(already created — apply and validate).*

---

## 3. Phased Fix Plan

### Phase 0 — Hotfixes (Today, 2026-05-09)
Goal: Stop the bleeding. Get a single successful playground execution.

| # | Fix | File | Status |
|---|-----|------|--------|
| 0.1 | `inputBytes := input` (no `json.Marshal`) | `sandbox.go:759` | ✅ Applied |
| 0.2 | Route `python*` → source-code path | `sandbox.go:948` | ✅ Applied |
| 0.3 | Append `handler(input_data)` to wrappers | `runtime.rs` (3 sites) | ✅ Applied |
| 0.4 | Remove `import json` from RustPython wrapper | `runtime.rs` | 🔲 **NEXT** |
| 0.5 | Apply DB migration | `migrations/20260509172354_*` | 🔲 **NEXT** |
| 0.6 | Fix playground proxy port (`8080` not `8090`) | `playground.go:108` | 🔲 **NEXT** |

**Validation Gate:** `curl -X POST /v1/fx/functionfly/xml-to-json -d '{"data": "<root><item>test</item></root>"}'` must return `"data": {"ok": true, "result": {...}}`.

### Phase 1 — Hardening (This Week, 2026-05-10 → 05-12)
Goal: Make execution reliable for all supported runtimes, fix frontend, add observability.

| # | Fix | Approach |
|---|-----|----------|
| 1.1 | Frontend response mapper | Complete `mapExecutionResponseToExecutionResult()` with unit tests |
| 1.2 | RustPython stdlib support | Switch to `Interpreter::default()` or vendor a minimal `json` module |
| 1.3 | Input parsing in user functions | Update all Python templates to parse `input_data` string via `json.loads()` inside `handler()` |
| 1.4 | Execution observability | Add structured logging to `executeLocallyWithLimits`: runtime type, bundler path, daemon vs legacy, result size |
| 1.5 | Fallback chain | Ensure `client.Execute` → `client.IsRunning()` → legacy executor → error response propagates correctly |
| 1.6 | Add `data` to response even when empty | Change `json.RawMessage` `omitempty` to explicit `"data": null` so frontend can distinguish empty vs missing |

### Phase 2 — 2026 Production Architecture (Next Sprint)
Goal: Replace the fragile RustPython fallback with a robust, scalable execution engine.

---

## 4. 2026 Architecture: Execution Engine Redesign

### 4.1 Current State (Problems)

```
┌─────────────────────────────────────────────────────────────────┐
│  Go Orchestrator (8080)                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Registry   │───▶│  Execution  │───▶│  SandboxClient      │  │
│  │  Handler    │    │  Handler    │    │  (HTTP to daemon)   │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                              │                                    │
│                              ▼                                    │
│                    ┌─────────────────┐                            │
│                    │  Lazy Bundler   │  ←── Returns micropython.wasm│
│                    │  (Python→WASM)  │     WITHOUT user code        │
│                    └─────────────────┘                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  SAR Daemon (Rust, dynamic port)                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Axum Router  │───▶│  detect_runtime│───▶│  RustPython VM      │  │
│  │  /execute/... │    │  (magic bytes) │    │  without stdlib     │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                                                ▲                │
│                                                │                │
│                                         Broken wrapper:         │
│                                         never calls handler()   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Target State (2026)

```
┌─────────────────────────────────────────────────────────────────┐
│  Go Orchestrator (8080) — Control Plane                           │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Registry    │───▶│  Execution   │───▶│  Runtime Router    │  │
│  │  Handler     │    │  Handler     │    │  (runtime-aware)     │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                              │                                    │
│         ┌────────────────────┼────────────────────┐               │
│         ▼                    ▼                    ▼               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │ WASM Engine │    │ Python Micro│    │ Deno/Node Runtime   │  │
│  │ (Wasmtime)  │    │ VM (Firecracker)│  │ (V8 Isolate Pool)   │  │
│  │ Pool-per-fn │    │ Enterprise  │    │ sandbox_worker      │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Worker Pool — Execution Plane (Redis + BullMQ)                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Bull Queue  │───▶│  Worker      │───▶│  Runtime Sandbox    │  │
│  │  "execution" │    │  (Go/ Rust)   │    │  (pre-warmed pool)  │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                              │                                    │
│                              ▼                                    │
│                    ┌─────────────────┐                            │
│                    │  Redis Cache     │                            │
│                    │  (L1: result,    │                            │
│                    │   L2: WASM bytes)│                            │
│                    └─────────────────┘                            │
└─────────────────────────────────────────────────────────────────┘
```

### 4.3 Component Design

#### A. Runtime Router (`internal/api/handlers/registry/execution/router.go`)
Replace `executeLocallyWithLimits` with a runtime-aware dispatcher:

```go
type RuntimeRouter struct {
    wasmEngine    *WasmEngine          // Wasmtime pool
    pythonEngine  *PythonMicroVMEngine // Firecracker/ gVisor
    denoEngine    *DenoEngine          // V8 isolates
    fallback      *LegacyExecutor      // Phase-out
}

func (r *RuntimeRouter) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
    switch req.Runtime {
    case "node18", "node20", "deno", "bun":
        return r.denoEngine.Execute(ctx, req)
    case "python3.11", "python3.12":
        // Tier-aware: MicroVM for Enterprise, Deno Python polyfill for Pro,
        // CPython-WASM (full stdlib) for Free tier
        return r.selectPythonEngine(req.Tier).Execute(ctx, req)
    case "rust", "go", "c", "cpp":
        return r.wasmEngine.Execute(ctx, req)
    default:
        return r.fallback.Execute(ctx, req)
    }
}
```

#### B. Python Execution Strategy (3 Tiers)

| Tier | Engine | Rationale | Fallback |
|------|--------|-----------|----------|
| **Free / Pro** | CPython-WASM (WASI) | Full stdlib, `xml`, `json`, `re` work | Deno `pyodide` |
| **Business** | MicroPython WASM | Fast cold start, small binary, no stdlib limits | CPython-WASM |
| **Enterprise** | Firecracker MicroVM | Real Linux kernel, pip install, native extensions | MicroPython |

**Key Decision:** Phase out RustPython entirely. It lacks stdlib support, has poor error messages, and its WASM compilation path is unmaintained. Replace with:
1. **CPython compiled to WASM** (WASI target) — already partially supported via `cpython.wasm`
2. **MicroPython** for constrained environments
3. **Firecracker MicroVM** for enterprise workloads

#### C. Bundling Service (`internal/bundler/service.go`)
Move bundling from lazy (at-execution) to eager (at-publish):

```go
type BundleService struct {
    cache      *redis.Client
    compilers  map[string]RuntimeCompiler
}

func (s *BundleService) Bundle(ctx context.Context, fn *RegistryFunction) (*Bundle, error) {
    // Check Redis cache first (key: sha256(source_code))
    if cached := s.cache.Get(ctx, cacheKey); cached != nil {
        return cached, nil
    }
    
    compiler := s.compilers[fn.Runtime]
    bundle, err := compiler.Compile(ctx, fn.SourceCode, fn.Manifest)
    if err != nil {
        return nil, fmt.Errorf("compilation failed: %w", err)
    }
    
    s.cache.Set(ctx, cacheKey, bundle, 24*time.Hour)
    return bundle, nil
}
```

**Benefits:**
- Eliminates 5-second cold starts in lazy bundling
- Allows compilation errors to surface at publish time, not runtime
- Enables AOT caching in Redis (WASM bytes rarely change)

#### D. Worker Pool (Redis + BullMQ)
For high-load scenarios, queue executions to a worker pool:

```go
type ExecutionWorker struct {
    queue   *bull.Queue
    pool    *sandbox.Pool
}

func (w *ExecutionWorker) Process(job *bull.Job) error {
    req := job.Data.(ExecutionRequest)
    result, err := w.pool.Execute(req)
    if err != nil {
        return err // Bull auto-retry with backoff
    }
    return job.MoveToCompleted(result, true)
}
```

#### E. Frontend Playground — Streaming Response
The 2026 frontend should use Server-Sent Events (SSE) for real-time execution feedback:

```typescript
// playgroundStore.ts
export class PlaygroundStore {
    async executeStream(input: unknown): AsyncGenerator<ExecutionEvent> {
        const response = await fetch('/v1/playground/execute/stream', {
            method: 'POST',
            body: JSON.stringify({ input }),
        });
        
        const reader = response.body!.getReader();
        const decoder = new TextDecoder();
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            const chunk = decoder.decode(value);
            for (const line of chunk.split('\n')) {
                if (line.startsWith('data: ')) {
                    yield JSON.parse(line.slice(6)) as ExecutionEvent;
                }
            }
        }
    }
}
```

Event types:
- `compilation:start` — bundling begins
- `compilation:complete` — WASM ready
- `execution:start` — sandbox invoked
- `execution:progress` — stdout/stderr streams (for long-running)
- `execution:complete` — result + metrics
- `execution:error` — structured error with recovery suggestions

---

## 5. Data Flow: Execution Request Lifecycle

```
1. User clicks "Run" in Playground
        │
        ▼
2. POST /v1/playground/execute
   Body: { input: {...} }
        │
        ▼
3. PlaygroundHandler.HandlePlaygroundExecute()
   a. Resolve function version from DB (Prisma)
   b. Check verification status (blocked_at?)
   c. Build ExecutionRequest
        │
        ▼
4. RuntimeRouter.Execute()
   a. Resolve runtime from fnVersion.Runtime
   b. Check Redis L1 cache (deterministic + idempotent only)
   c. If miss:
      - WASM runtimes: fetch from Redis L2 (bundled WASM cache)
      - Python: fetch or compile CPython-WASM
      - Node: fetch or compile Deno bundle
        │
        ▼
5. Engine.Execute()
   a. Acquire warm instance from pool (or cold start)
   b. Inject input + environment
   c. Run with fuel metering + memory limits
   d. Capture stdout + structured result
        │
        ▼
6. Result Normalization
   a. Validate JSON schema against fnVersion.OutputSchema
   b. If valid: cache in Redis L1, return { ok: true, data, ... }
   c. If invalid: return { ok: false, error: { code, message } }
        │
        ▼
7. PlaygroundHandler Response
   a. Map ExecutionResponse → PlaygroundExecuteResponse
   b. Return 200 with SSE stream (or 202 queued if worker pool)
        │
        ▼
8. Frontend renders:
   a. Monaco diff (input vs expected)
   b. Syntax-highlighted output (JSON tree)
   c. Latency waterfall (compile / cold start / execute)
   d. Cache hit indicator
```

---

## 6. API Contract Changes

### 6.1 Execution Endpoint (Stable)
```http
POST /v1/fx/{author}/{name}[@{version}]
Content-Type: application/json
Authorization: Bearer <token>

{ "key": "value" }
```

Response (unchanged format, guaranteed `data` presence):
```json
{
  "ok": true,
  "data": { "result": "hello" },
  "cached": false,
  "duration_ms": 45,
  "version": "1.0.0",
  "execution_id": "uuid"
}
```

### 6.2 Playground Endpoint (New Streaming Support)
```http
POST /v1/playground/execute
Content-Type: application/json

{ "function_id": "...", "input": {...} }
```

Response:
```json
{
  "success": true,
  "output": { "result": "hello" },
  "latency_ms": 45,
  "compilation_ms": 120,
  "version": "1.0.0",
  "execution_id": "uuid",
  "cached": false
}
```

### 6.3 SSE Stream Endpoint (2026)
```http
POST /v1/playground/execute/stream
Accept: text/event-stream
```

---

## 7. Database Schema

### 7.1 Current Drift Fix
Apply migration `20260509172354_add_blocking_columns_to_verification_status`:

```sql
ALTER TABLE registry_function_verification_status
    ADD COLUMN IF NOT EXISTS blocked_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS block_reason TEXT;
```

### 7.2 New Table: `function_bundles` (P2)
Store pre-compiled WASM / JS bundles for fast retrieval:

```sql
CREATE TABLE function_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    runtime VARCHAR(50) NOT NULL,
    bundle_hash VARCHAR(64) NOT NULL, -- SHA-256 of source_code
    wasm_binary BYTEA,               -- NULL for JS/Deno (stored in S3/R2)
    compiled_size_bytes INT,
    compilation_duration_ms INT,
    compiled_at TIMESTAMPTZ DEFAULT now(),
    is_valid BOOLEAN DEFAULT true,
    UNIQUE(function_version_id, runtime, bundle_hash)
);

CREATE INDEX idx_function_bundles_lookup ON function_bundles(function_version_id, runtime, bundle_hash);
```

### 7.3 New Table: `execution_pool_metrics` (P2)
Track pool health for autoscaling:

```sql
CREATE TABLE execution_pool_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime VARCHAR(50) NOT NULL,
    pool_size INT NOT NULL,
    warm_instances INT NOT NULL,
    cold_starts_1m INT NOT NULL,
    avg_latency_ms INT,
    recorded_at TIMESTAMPTZ DEFAULT now()
);
```

---

## 8. Testing & Validation Plan

### 8.1 Integration Test Matrix

| Runtime | Function | Input | Expected Output | Tier |
|---------|----------|-------|-----------------|------|
| `python3.12` | xml-to-json | `{"data": "<root><item>x</item></root>"}` | `{"ok": true, "result": {"item": {"_text": "x"}}}` | Free |
| `node20` | json-prettify | `{"input": "{\"a\":1}"}` | `{"ok": true, "result": "{\n  \"a\": 1\n}"}` | Free |
| `deno` | hello-world | `{}` | `{"ok": true, "result": "hello"}` | Free |
| `python3.12` | numpy-sum | `{"arr": [1,2,3]}` | Error: `numpy not available` | Free |
| `python3.12` | pandas-read | `{"csv": "a,b\n1,2"}` | Success (Enterprise MicroVM) | Enterprise |
| `rust` | fibonacci | `{"n": 10}` | `{"ok": true, "result": 55}` | Free |

### 8.2 Load Test
```bash
# 100 RPS, 60 seconds
k6 run --vus 100 --duration 60s tests/load/execution.js
```

Success criteria:
- p99 latency < 500ms (warm pool)
- p99 latency < 3s (cold start)
- Error rate < 0.1%
- Cache hit rate > 60% (deterministic functions)

### 8.3 Chaos Test
- Kill SAR daemon mid-execution → verify fallback to legacy executor
- Corrupt WASM binary in Redis → verify re-compilation
- Saturate memory limit → verify `ResourceExhausted` error

---

## 9. Observability

### 9.1 Metrics (Prometheus)
- `functionfly_execution_total{runtime, status, tier}` — counter
- `functionfly_execution_duration_ms{runtime, tier, cache_hit}` — histogram
- `functionfly_cold_start_total{runtime}` — counter
- `functionfly_pool_size{runtime}` — gauge
- `functionfly_bundle_cache_hit_ratio` — gauge

### 9.2 Tracing (OpenTelemetry)
Span structure:
```
playground.execute
├── db.resolve_function       (Prisma)
├── cache.check               (Redis)
├── bundler.compile           (optional)
│   └── bundler.cache_store   (Redis)
├── sandbox.execute
│   ├── pool.acquire
│   ├── wasm.compile          (cold start only)
│   ├── wasm.instantiate
│   └── wasm.run
└── response.normalize
```

### 9.3 Structured Logging
Every execution emits:
```json
{
  "level": "info",
  "msg": "function execution completed",
  "execution_id": "uuid",
  "function_id": "uuid",
  "runtime": "python3.12",
  "tier": "free",
  "duration_ms": 45,
  "compilation_ms": 120,
  "cache_hit": false,
  "result_size_bytes": 256,
  "status": "success"
}
```

---

## 10. Rollback Strategy

| Scenario | Rollback Action |
|----------|----------------|
| New SAR binary crashes | Pin to previous release in `findLocalRuntime()`; restart API |
| Bundler produces invalid WASM | Disable eager bundling; fall back to lazy bundler |
| CPython-WASM too slow | Route to MicroPython for free tier; gate CPython to paid tiers |
| Database migration fails | Migration is idempotent (`IF NOT EXISTS`); no rollback needed |
| Frontend mapper broken | Feature-flag `playgroundV2` off; return v1 response format |

---

## 11. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-05-09 | Replace RustPython with CPython-WASM | RustPython lacks stdlib (`json`, `xml`) and has poor error diagnostics |
| 2026-05-09 | Eager bundling at publish time | Eliminates runtime compilation latency; surfaces errors early |
| 2026-05-09 | Runtime-aware router in Go | Keeps orchestrator as control plane; runtimes as pluggable workers |
| 2026-05-09 | Keep BullMQ worker pool optional | Default to inline execution; queue only under load (>80% pool utilization) |
| 2026-05-09 | SSE streaming for playground | Real-time UX for long compilations; aligns with 2026 UX standards |

---

## 12. Immediate Action Items (Next 4 Hours)

1. **Fix RustPython `import json`** — Remove import from wrapper; pass raw string to handler
2. **Apply DB migration** — `psql -f migrations/20260509172354_add_blocking_columns_to_verification_status.up.sql`
3. **Fix playground proxy port** — Change fallback from `8090` to `8080` (or read from server config)
4. **Rebuild SAR runtime** — `cargo build --release` in `runtimes/local/`
5. **Restart API** — Verify health; test `xml-to-json` execution
6. **Validate end-to-end** — Playground UI → execute → see structured JSON output

---

*End of Plan*
