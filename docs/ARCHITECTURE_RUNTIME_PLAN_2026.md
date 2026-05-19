# FunctionFly 2026 Runtime Architecture Plan

**Status**: Planning
**Last Updated**: 2026-05-11
**Author**: System Architecture

---

## Executive Summary

This document outlines a production-grade architecture for Python/WASM runtime execution without CGO dependencies. The plan addresses the fundamental constraint that `wasmtime-go` requires CGO for its performance-critical JIT compilation, while delivering a practical path to non-CGO production builds.

**Key Insight**: The user's concern is valid—implementing a full Python/WASM runtime in pure Go from scratch is extremely complex (wasmtime uses CGO for the C/C++ JIT compiler). However, practical alternatives exist that can achieve production-quality Python execution without CGO.

---

## Current State Analysis

### Existing Architecture

| Component | Technology | CGO Required |
|-----------|------------|--------------|
| Go Orchestrator | Pure Go | No |
| Python WASM Runtime | wasmtime-go v19 | **Yes** |
| Node.js Runtime | QuickJS WASM | No |
| SAR (Stateful Agent) | Rust + NATS | No |
| MicroVM Runtime | Firecracker | No (isolated) |

### The CGO Problem

```go
// runtime.go (CGO build - works but requires C dependency)
github.com/bytecodealliance/wasmtime-go/v19  // Requires CGO for JIT

// runtime_stub.go (Non-CGO build - completely disabled)
errWasmNotAvailable  // All operations return error
```

**Impact**: `make build-ci` produces binaries without Python/WASM support.

---

## 2026 Production Architecture

### Recommended Strategy: Tiered Runtime Model

Instead of a single Python/WASM solution, we adopt a tiered approach that matches runtime capabilities to use cases:

```
┌─────────────────────────────────────────────────────────────┐
│                    Request Classification                     │
└─────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
   │   Simple    │     │  Business   │     │  Enterprise │
   │  (Budget)   │     │  (WASM)     │     │  (MicroVM)  │
   └─────────────┘     └─────────────┘     └─────────────┘
          │                   │                   │
          ▼                   ▼                   ▼
   Pure Go Handler     External Runtime      Firecracker
   (no CGO)           (see options)          (isolated VM)
```

---

## Production Implementation Options

### Option A: External WASM Runtime Service (Recommended)

**Architecture**: Keep the Go orchestrator pure Go, run Python WASM execution in a separate service.

```
┌─────────────────────┐         ┌─────────────────────────┐
│   Go Orchestrator   │         │   Python WASM Runtime   │
│   (CGO_ENABLED=0)   │   HTTP  │   (Rust + wasmtime)    │
│                     │────────▶│   or                     │
│  - Routing          │         │   CGO-enabled Go service │
│  - Auth             │         │                         │
│  - Billing          │         │   (Can use CGO)         │
└─────────────────────┘         └─────────────────────────┘
```

**Implementation**:

1. **New service**: `cmd/wasm-runtime/` - HTTP service wrapping wasmtime
2. **Protocol**: Simple HTTP/JSON or gRPC
3. **Benefits**:
   - Orchestrator stays pure Go
   - Runtime can use CGO for performance
   - Independent scaling
   - Failover capability

**Example API**:
```yaml
POST /execute
Content-Type: application/json

{
  "code": "def handler(event): return {'result': event['x'] * 2}",
  "input": {"x": 21},
  "runtime": "python-wasm",
  "timeout_ms": 5000
}

Response:
{
  "output": {"result": 42},
  "execution_time_ms": 12,
  "memory_mb": 8
}
```

### Option B: Pure Go WASM with Wazero

**Architecture**: Use `wazero` - a pure Go WASM runtime that doesn't need CGO.

```
┌─────────────────────────────────────┐
│         Go Orchestrator              │
│         (CGO_ENABLED=0)              │
│                                     │
│  ┌───────────────────────────────┐  │
│  │     wazero Runtime            │  │
│  │     (Pure Go, No CGO)        │  │
│  │                               │  │
│  │  - Compiles WASM to Go code  │  │
│  │  - AOT compilation support    │  │
│  │  - WASI compatibility        │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Trade-offs**:
| Aspect | wasmtime (CGO) | wazero (Pure Go) |
|--------|-----------------|-------------------|
| Performance | ~2-5x faster | Slower cold start |
| WASI Support | Full | Partial (preview1) |
| Python Compatibility | Excellent | Limited* |
| Memory Usage | Higher | Lower |

*Python support requires a WASM-compiled Python that works with wazero's limitations.

**Implementation Path**:
```go
// internal/wasm/runtime_wazero.go
//go:build !cgo

package wasm

import (
    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type PythonRuntime struct {
    runtime wazero.Runtime
    module  wazero.CompiledModule
    // ...
}
```

### Option C: Hybrid Build with Build Tags

**Architecture**: Maintain both CGO and non-CGO implementations, selected at build time.

```go
// runtime_cgo.go
//go:build cgo
package wasm

// Uses wasmtime-go (full performance)

// runtime_wazero.go
//go:build wazero && !cgo
package wasm

// Uses wazero (pure Go)

// runtime_stub.go
//go:build (!cgo) && !wazero
package wasm

// Returns errWasmNotAvailable
```

**Build Matrix**:
| Build Target | Tags | Python WASM |
|-------------|------|-------------|
| `make build` | Default (CGO) | Full |
| `make build-ci` | `CGO_ENABLED=0` | Stub (disabled) |
| `make build-wazero` | `wazero` | wazero-based |

---

## Recommended Implementation: Option A + C Hybrid

For maximum production flexibility:

### Phase 1: External Runtime Service (Q1 2026)

**Objective**: Decouple Python WASM from orchestrator build

```bash
# New service structure
cmd/
  wasm-runtime/          # New service
    main.go
    executor.go          # Wraps wasmtime
    pool.go              # Instance pooling

# Service responsibilities
- Load Python WASM modules (CPython 3.13 WASI, MicroPython)
- Manage per-tenant runtime pools
- Execute user code with isolation
- Return results via HTTP/gRPC
```

**API Design**:
```yaml
WASM Runtime Service (Port 8083):

POST /v1/execute
  → Execute Python code in WASM runtime

POST /v1/execute/stream
  → Streaming execution for long outputs

GET  /v1/health
  → Health check with runtime status

POST /v1/pool/maintain
  → Trigger pool maintenance
```

**Orchestrator Changes**:
```go
// internal/wasm/runtime_client.go
//go:build !cgo

package wasm

type RuntimeClient struct {
    endpoint string  // http://localhost:8083
    client   *http.Client
}

func (r *RuntimeClient) Execute(ctx context.Context, code string, input []byte) ([]byte, error) {
    // HTTP call to external runtime
}
```

### Phase 2: Wazero as Fallback (Q2 2026)

**Objective**: Enable basic Python execution without external service

- Implement wazero-based runtime for simple Python
- Focus on MicroPython-compatible code
- Add as fallback when external service unavailable

### Phase 3: Build System Updates (Q3 2026)

**Updated Makefile**:
```makefile
# Pure Go build (no Python WASM)
build-ci:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) $(LD_FLAGS) -o bin/orchestrator-api ./cmd/orchestrator-api

# Build with external runtime support
build-with-wasm:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -tags=wasm_runtime -o bin/orchestrator-api ./cmd/orchestrator-api
	# Note: External wasm-runtime service must be running

# Full build with CGO (development)
build-full:
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/orchestrator-api ./cmd/orchestrator-api
```

---

## Detailed Implementation: External WASM Runtime Service

### Service Structure

```
cmd/wasm-runtime/
├── main.go                 # Service entry, HTTP server
├── executor/
│   ├── executor.go         # Core execution logic
│   ├── pool.go             # Runtime instance pool
│   └── security.go         # Memory limits, timeouts
├── runtime/
│   ├── python.go           # Python WASM runtime
│   └── types.go            # Request/response types
├── api/
│   └── handlers.go         # HTTP handlers
└── Dockerfile
```

### Execution Flow

```
1. Request arrives at orchestrator
2. Orchestrator (pure Go) routes to appropriate engine
3. For Python WASM:
   a. If external runtime enabled: HTTP call to wasm-runtime
   b. If CGO build: local wasmtime execution
   c. If wazero build: wazero execution (limited)
4. Result returned to caller
```

### Security Model

```go
type SecurityConfig struct {
    MaxMemoryMB     int           // Per-execution memory limit
    MaxTimeoutMs    int           // Execution timeout
    MaxOutputKB     int           // Output size limit
    EnableWASI      bool          // WASI filesystem access
    AllowedModules  []string      // Pre-approved imports
}
```

### Pool Management

```go
type RuntimePool struct {
    // Per-tenant pools
    tenants map[string]*TenantPool

    // Global config
    maxInstancesPerTenant int
    maxMemoryMB           int
    instanceTTL           time.Duration
}

type TenantPool struct {
    mu        sync.Mutex
    instances []*RuntimeInstance
    lastUsed  time.Time
}
```

---

## CPython WASM Support (Business Tier)

### Current Assets

| File | Description | Status |
|------|-------------|--------|
| `runtimes/cpython-wasi/` | CPython 3.13 WASM + stdlib | Production-ready |
| `runtimes/cpython.wasm` | Compiled CPython binary | Tested |

### Architecture

```
┌────────────────────────────────────────────────────────┐
│                   Function Request                     │
└────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│              RuntimeRouter                            │
│  Tier: business → cpythonWasmEngine                   │
└────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│              CPython WASM Pool                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │Instance 1│  │Instance 2│  │Instance 3│  ...       │
│  │ (tenant) │  │ (tenant) │  │ (tenant) │           │
│  └──────────┘  └──────────┘  └──────────┘           │
└────────────────────────────────────────────────────────┘
```

### Execution Wrapper

The CPython WASM requires a wrapper for the FunctionFly API:

```python
# wrapper.py (embedded in WASM or loaded)
import json
import sys

def _ff_main():
    # Read input from stdin or memory
    input_data = json.loads(sys.stdin.read())

    # Execute user handler
    code = input_data.get('code', '')
    event = input_data.get('input', {})

    # User function is defined in the uploaded module
    result = handler(event)

    # Output JSON
    print(json.dumps({'result': result}))

if __name__ == '__main__':
    _ff_main()
```

---

## Testing Strategy

### Unit Tests

```go
// internal/wasm/runtime_test.go
func TestPythonRuntime_Execute(t *testing.T) {
    // Test with actual WASM module
    runtime, err := NewPythonRuntime("./testdata/micropython.wasm", nil, nil, handler)
    require.NoError(t, err)

    err = runtime.LoadCode("def handler(e): return e['x'] * 2")
    require.NoError(t, err)

    output, err := runtime.Execute([]byte(`{"x": 21}`))
    require.NoError(t, err)
    assert.Contains(t, string(output), "42")
}
```

### Integration Tests

```bash
# Test external runtime service
cargo test --manifest-path runtimes/local/Cargo.toml

# Test orchestrator integration
go test -tags=integration ./internal/api/...
```

---

## Migration Path

### For Existing Deployments

1. **No changes required** - Current CGO builds continue working
2. **New deployments** can use pure Go orchestrator + external runtime
3. **Gradual migration** - enable features incrementally

### For New Deployments

```bash
# Option 1: Pure Go orchestrator + external runtime service
make build-ci                    # Pure Go orchestrator
make build-wasm-runtime          # CGO-enabled runtime service
./bin/wasm-runtime &             # Start runtime service
./bin/orchestrator-api           # Start orchestrator

# Option 2: Full CGO build (simpler, single binary)
make build                       # Includes CGO
```

---

## Rollout Timeline

| Quarter | Milestone | Deliverables |
|---------|----------|--------------|
| Q1 2026 | External Runtime Service | `cmd/wasm-runtime/`, HTTP API, pool management |
| Q2 2026 | Wazero Fallback | Basic wazero implementation for simple Python |
| Q3 2026 | Build System Update | Updated Makefile, documentation |
| Q4 2026 | Production Hardening | Load testing, security audit, monitoring |

---

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| wazero Python compatibility gaps | Medium | High | External runtime primary, wazero fallback only |
| Performance regression | Low | Medium | Benchmark comparisons, gradual rollout |
| External service availability | Low | Medium | Graceful degradation, local fallback |
| CPython WASM memory issues | Medium | Low | Pool limits, instance recycling |

---

## Appendix: Technical Notes

### Why wasmtime Requires CGO

The WebAssembly JIT compilation is performance-critical code that benefits from:
- Native code generation
- SIMD optimizations
- Custom memory management

Pure Go alternatives like wazero use AOT compilation and interpreter approaches, which are portable but slower for complex workloads.

### Memory Overhead Comparison

| Runtime | Cold Start | Memory per Instance |
|---------|------------|---------------------|
| wasmtime (CGO) | ~50ms | ~5-20MB |
| wazero | ~5-10ms | ~2-10MB |
| CPython WASI | ~100ms | ~15-30MB |
| Firecracker MicroVM | ~150ms | ~5MB (shared) |

### WASM Module Compatibility

| Python Runtime | wasmtime | wazero | Notes |
|---------------|----------|--------|-------|
| MicroPython | ✓ | Partial | Limited stdlib |
| CPython 3.13 WASI | ✓ | ✗ | Requires WASI preview2 |
| Pyodide | ✓ | ✗ | Complex JS interop |

---

## Conclusion

The recommended path forward is **Option A + C hybrid**:

1. **External WASM Runtime Service** - Primary solution for production Python/WASM
2. **Wazero fallback** - For development and simple use cases
3. **CGO build** - Available for full performance when needed

This approach:
- ✅ Maintains orchestrator as pure Go (CGO_ENABLED=0 compatible)
- ✅ Delivers production-quality Python execution
- ✅ Provides flexibility for different deployment scenarios
- ✅ Avoids stub implementations
- ✅ Enables independent scaling of runtime components

The external runtime service is the key innovation - it separates the CGO dependency from the main orchestrator while maintaining full Python/WASM capability.
