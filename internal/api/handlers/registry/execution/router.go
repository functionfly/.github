package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/storage"
)

// ExecutionRequest contains all inputs needed to route and execute a function.
type ExecutionRequest struct {
	FunctionVersion *storage.RegistryFunctionVersion
	Function        *storage.RegistryFunction
	Input           json.RawMessage
	MaxMemoryMB     int
	MaxCPUTimeMs    int
	TenantID        string
	Tier            string // free, pro, business, enterprise
	Runtime         string
}

// ExecutionResult is the normalized output of any runtime engine.
type ExecutionResult struct {
	Output       json.RawMessage
	DurationMs   int
	CacheHit     bool
	ColdStart    bool
	ResourceUsage *ResourceUsage
	Error        *ExecutionError
}

// RuntimeEngine is the interface that every execution backend must implement.
type RuntimeEngine interface {
	// Execute runs the function and returns a normalized result.
	Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error)
	// Healthy reports whether the engine is ready to accept work.
	Healthy(ctx context.Context) bool
	// Close releases any engine-scoped resources.
	Close() error
}

// RuntimeRouter selects the appropriate engine based on runtime and tier.
type RuntimeRouter struct {
	wasmEngine    RuntimeEngine // Wasmtime pool (Rust, Go, C, C++)
	pythonEngine  RuntimeEngine // Firecracker MicroVM for enterprise Python
	cpythonEngine RuntimeEngine // In-process CPython-WASM for business Python
	nodeEngine    RuntimeEngine // Deno/Node V8 isolates
	fallback      RuntimeEngine // Legacy executor (RustPython / per-request sandbox)
	cacheL1       *cache.CacheService // deterministic result cache
	bundleService *bundler.BundleService
}

// NewRuntimeRouter creates a router with the given engines.
func NewRuntimeRouter(
	wasm, python, cpython, node, fallback RuntimeEngine,
	cacheL1 *cache.CacheService,
	bundleSvc *bundler.BundleService,
) *RuntimeRouter {
	return &RuntimeRouter{
		wasmEngine:    wasm,
		pythonEngine:  python,
		cpythonEngine: cpython,
		nodeEngine:    node,
		fallback:      fallback,
		cacheL1:       cacheL1,
		bundleService: bundleSvc,
	}
}

// Execute routes an execution request to the correct engine.
func (r *RuntimeRouter) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	// 1. L1 cache check for deterministic + idempotent functions.
	if r.cacheL1 != nil && req.FunctionVersion != nil && req.FunctionVersion.Deterministic {
		if hit, ok := r.checkCache(ctx, req); ok {
			return hit, nil
		}
	}

	// 2. Select engine by runtime and tier.
	engine := r.selectEngine(req)

	// 3. Pre-warm / fetch bundle (eager bundling at publish time).
	if r.bundleService != nil {
		_ = r.bundleService.Warm(ctx, req.FunctionVersion)
	}

	// 4. Execute.
	result, err := engine.Execute(ctx, req)
	if err != nil && engine != r.fallback {
		// Try fallback once before giving up.
		result, err = r.fallback.Execute(ctx, req)
	}

	// 5. Cache successful deterministic results.
	if err == nil && r.cacheL1 != nil && req.FunctionVersion != nil && req.FunctionVersion.Deterministic {
		r.storeCache(ctx, req, result)
	}

	return result, err
}

// IsHealthy returns true if at least one engine is healthy.
func (r *RuntimeRouter) IsHealthy() bool {
	for _, e := range []RuntimeEngine{r.wasmEngine, r.pythonEngine, r.cpythonEngine, r.nodeEngine, r.fallback} {
		if e != nil && e.Healthy(context.Background()) {
			return true
		}
	}
	return false
}

func (r *RuntimeRouter) selectEngine(req ExecutionRequest) RuntimeEngine {
	switch req.Runtime {
	case "rust", "go", "c", "cpp", "zig":
		if r.wasmEngine != nil && r.wasmEngine.Healthy(context.Background()) {
			return r.wasmEngine
		}
	case "node18", "node20", "deno", "bun":
		if r.nodeEngine != nil && r.nodeEngine.Healthy(context.Background()) {
			return r.nodeEngine
		}
	case "python3.11", "python3.12", "python3.13":
		// Tier-aware Python routing (see architecture decision log).
		engine := r.selectPythonEngine(req.Tier)
		if engine != nil && engine.Healthy(context.Background()) {
			return engine
		}
	}
	return r.fallback
}

func (r *RuntimeRouter) selectPythonEngine(tier string) RuntimeEngine {
	// Tier-aware Python routing:
	//   - enterprise → Firecracker MicroVM (r.pythonEngine)
	//   - business/pro → in-process CPython-WASM (r.cpythonEngine)
	//   - free/starter → daemon / sandboxEngine (r.fallback)
	//
	// Note: In-process CPython-WASM uses the official CPython 3.13 WASI build
	// with full stdlib support. It requires python.wasm at ./runtimes/cpython.wasm
	// and stdlib at ./runtimes/cpython-wasi/lib.
	//
	// "business" is the product tier name; internally it maps to "pro" plan
	// in resolveTierFromRequest(). Both strings are accepted for safety.
	switch tier {
	case "enterprise":
		if r.pythonEngine != nil {
			return r.pythonEngine
		}
	case "pro", "business":
		if r.cpythonEngine != nil && r.cpythonEngine.Healthy(context.Background()) {
			return r.cpythonEngine
		}
	}
	return r.fallback
}

func (r *RuntimeRouter) checkCache(ctx context.Context, req ExecutionRequest) (ExecutionResult, bool) {
	if r.cacheL1 == nil || req.FunctionVersion == nil {
		return ExecutionResult{}, false
	}
	key := r.cacheKey(req)
	cached, ok := r.cacheL1.GetDeterministicResult(ctx, key)
	if !ok {
		return ExecutionResult{}, false
	}
	return ExecutionResult{
		Output:     cached,
		DurationMs: 0,
		CacheHit:   true,
	}, true
}

func (r *RuntimeRouter) storeCache(ctx context.Context, req ExecutionRequest, res ExecutionResult) {
	if r.cacheL1 == nil || req.FunctionVersion == nil {
		return
	}
	key := r.cacheKey(req)
	ttl := time.Duration(req.FunctionVersion.CacheTTL) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if ttl > 24*time.Hour {
		ttl = 24 * time.Hour // Cap at 24h
	}
	r.cacheL1.SetDeterministicResult(ctx, key, res.Output, ttl)
}

func (r *RuntimeRouter) cacheKey(req ExecutionRequest) string {
	// Deterministic cache key: det:{fn_version_id}:{sha256(input)}
	inputHash := sha256.Sum256(req.Input)
	return fmt.Sprintf("det:%s:%s", req.FunctionVersion.ID.String(), hex.EncodeToString(inputHash[:]))
}

// Close shuts down all engines.
func (r *RuntimeRouter) Close() error {
	var firstErr error
	for _, e := range []RuntimeEngine{r.wasmEngine, r.pythonEngine, r.cpythonEngine, r.nodeEngine, r.fallback} {
		if e != nil {
			if err := e.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
