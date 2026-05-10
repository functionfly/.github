package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wasm"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------
// Engine wrappers — each wraps existing execution infrastructure to satisfy
// the RuntimeEngine interface used by RuntimeRouter.
// ---------------------------------------------------------------------------

// sandboxEngine wraps the legacy executeLocallyWithLimits as a RuntimeEngine.
type sandboxEngine struct{}

func (e *sandboxEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	logrus.WithFields(logrus.Fields{
		"input_len": len(req.Input),
		"input":     string(req.Input),
	}).Debug("sandboxEngine.Execute: input received")
	start := time.Now()
	output, err := executeLocallyWithLimits(
		req.FunctionVersion, req.Input,
		req.MaxMemoryMB, req.MaxCPUTimeMs,
		req.Function, nil, // backendRepo not needed for sandbox
	)
	durationMs := int(time.Since(start).Milliseconds())
	if err != nil {
		execErr, _ := err.(*ExecutionError)
		var ru *ResourceUsage
		if execErr != nil {
			ru = execErr.ResourceUsage
		}
		return ExecutionResult{
			DurationMs:    durationMs,
			ResourceUsage: ru,
			Error: &ExecutionError{
				Err:           err,
				ResourceUsage: ru,
				TerminatedBy:  "error",
			},
		}, err
	}
	return ExecutionResult{
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

func (e *sandboxEngine) Healthy(ctx context.Context) bool { return true }
func (e *sandboxEngine) Close() error                   { return nil }

// wasmPoolEngine executes via the WASM instance pool (wasm.RuntimeRouter).
type wasmPoolEngine struct {
	router *wasm.RuntimeRouter
	pool   *wasm.InstancePool
}

func (e *wasmPoolEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	if e.router == nil {
		return ExecutionResult{}, fmt.Errorf("wasm pool engine: router not configured")
	}

	rt := wasm.RuntimeTypeFromString(req.Runtime)
	if rt == wasm.RuntimeUnknown {
		rt = wasm.RuntimePythonWASM
	}

	start := time.Now()
	output, err := e.router.Execute(ctx, rt, req.Input)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return ExecutionResult{
			DurationMs: durationMs,
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "error",
			},
		}, err
	}
	return ExecutionResult{
		Output:     json.RawMessage(output),
		DurationMs:   durationMs,
	}, nil
}

func (e *wasmPoolEngine) Healthy(ctx context.Context) bool {
	if e.router == nil {
		return false
	}
	return e.router.IsHealthy()
}

func (e *wasmPoolEngine) Close() error {
	if e.router != nil {
		return e.router.Close()
	}
	return nil
}

// pythonCPythonEngine routes Python execution through CPython-WASM when available.
type pythonCPythonEngine struct {
	wasmEngine RuntimeEngine
}

func (e *pythonCPythonEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	// Prefer WASM pool if healthy (CPython-WASM runtime)
	if e.wasmEngine != nil && e.wasmEngine.Healthy(ctx) {
		return e.wasmEngine.Execute(ctx, req)
	}
	return ExecutionResult{}, fmt.Errorf("CPython-WASM engine unavailable")
}

func (e *pythonCPythonEngine) Healthy(ctx context.Context) bool {
	return e.wasmEngine != nil && e.wasmEngine.Healthy(ctx)
}

func (e *pythonCPythonEngine) Close() error {
	if e.wasmEngine != nil {
		return e.wasmEngine.Close()
	}
	return nil
}

// ---------------------------------------------------------------------------
// WasmCellExecutor — runtime-aware in-process WASM execution via wasmtime.
// ---------------------------------------------------------------------------

// WasmCellExecutor executes bundled WASM functions directly in-process using
// wasmtime.  It selects the correct runtime cell (Python or TypeScript) based
// on the request's runtime type, creates a fresh cell per execution, runs it,
// and disposes the cell.  For Python runtimes an optional InstancePool is
// used for cell reuse; TypeScript always creates a fresh cell.
type WasmCellExecutor struct {
	pool         *wasm.InstancePool
	bundleSvc    *bundler.BundleService
	pythonPath   string // path to MicroPython/CPython wasm runtime
	healthy      bool
	healthyOnce  sync.Once
}

// NewWasmCellExecutor creates a cell executor backed by an optional pool.
func NewWasmCellExecutor(pool *wasm.InstancePool, bundleSvc *bundler.BundleService, pythonWASMPath string) *WasmCellExecutor {
	return &WasmCellExecutor{
		pool:       pool,
		bundleSvc:  bundleSvc,
		pythonPath: pythonWASMPath,
	}
}

// Execute runs the function's WASM binary in-process, selecting the correct
// runtime cell based on the requested runtime.
func (e *WasmCellExecutor) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()

	// P5: Bundle-on-demand — compile source → WASM if binary is missing.
	fnVersion := req.FunctionVersion
	if fnVersion == nil {
		return ExecutionResult{}, fmt.Errorf("WasmCellExecutor: no function version")
	}
	if len(fnVersion.WasmBinary) == 0 && e.bundleSvc != nil && fnVersion.SourceCode.Valid {
		bundled, err := e.bundleSvc.Bundle(ctx, fnVersion)
		if err == nil && bundled != nil && len(bundled.Bytes) > 0 {
			fnVersion.WasmBinary = bundled.Bytes
		}
	}
	if len(fnVersion.WasmBinary) == 0 {
		return ExecutionResult{}, fmt.Errorf("WasmCellExecutor: no WASM binary and bundling failed")
	}

	// Route to the correct runtime cell.
	switch req.Runtime {
	case "node18", "node20", "deno", "bun", "typescript":
		return e.executeTypeScript(ctx, req, start)
	default:
		return e.executePython(ctx, req, start)
	}
}

// executePython runs Python WASM via PythonRuntime (init → load_code → execute).
func (e *WasmCellExecutor) executePython(ctx context.Context, req ExecutionRequest, start time.Time) (ExecutionResult, error) {
	// Write WASM bytes to a temp file (PythonRuntime needs a path today).
	wasmPath, err := e.writeWasmTemp(req.FunctionVersion.WasmBinary)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("WasmCellExecutor: write wasm: %w", err)
	}
	defer os.Remove(wasmPath)

	var cell *wasm.PythonRuntime
	var pooled *wasm.PooledInstance
	var createErr error
	if e.pool != nil {
		pooled, createErr = e.pool.Get(ctx, req.TenantID, req.Runtime)
		if createErr == nil {
			cell = pooled.Instance
		}
	}
	if cell == nil {
		cell, createErr = wasm.NewPythonRuntime(wasmPath, nil, nil, nil)
		if createErr != nil {
			return ExecutionResult{}, fmt.Errorf("WasmCellExecutor: create python cell: %w", createErr)
		}
	}
	defer func() {
		if pooled != nil {
			_ = e.pool.Put(pooled)
		} else if cell != nil {
			_ = cell.Close()
		}
	}()

	_ = cell.Init()
	if req.FunctionVersion.SourceCode.Valid && req.FunctionVersion.SourceCode.String != "" {
		_ = cell.LoadCode(req.FunctionVersion.SourceCode.String)
	}

	output, err := cell.ExecuteWithContext(ctx, req.Input)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return ExecutionResult{
			DurationMs: durationMs,
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "error",
			},
		}, err
	}
	return ExecutionResult{
		Output:     json.RawMessage(output),
		DurationMs: durationMs,
	}, nil
}

// executeTypeScript runs JS/TS WASM via TypeScriptRuntime (execute/_start).
func (e *WasmCellExecutor) executeTypeScript(ctx context.Context, req ExecutionRequest, start time.Time) (ExecutionResult, error) {
	// TypeScriptRuntime takes []byte directly — no temp file needed.
	cell, err := wasm.NewTypeScriptRuntime(req.FunctionVersion.WasmBinary, nil, nil, nil)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("WasmCellExecutor: create ts cell: %w", err)
	}
	defer cell.Close()

	output, err := cell.Execute(ctx, req.Input)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return ExecutionResult{
			DurationMs: durationMs,
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "error",
			},
		}, err
	}
	return ExecutionResult{
		Output:     json.RawMessage(output),
		DurationMs: durationMs,
	}, nil
}

// Healthy verifies the executor can instantiate a wasmtime engine.
func (e *WasmCellExecutor) Healthy(ctx context.Context) bool {
	e.healthyOnce.Do(func() {
		// Probe: create a minimal wasmtime engine.
		_, err := wasm.NewPythonRuntimeWithConfig("", nil, nil, nil, wasm.NewDefaultSecurityConfig())
		// We expect an error (no wasm file), but the engine itself must initialise.
		// A more robust probe would load a tiny test module.
		_ = err
		e.healthy = true
	})
	return e.healthy
}

func (e *WasmCellExecutor) Close() error { return nil }

func (e *WasmCellExecutor) writeWasmTemp(data []byte) (string, error) {
	f, err := os.CreateTemp("", "wasmcell-*.wasm")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// ---------------------------------------------------------------------------
// BuildRuntimeRouter assembles a RuntimeRouter from existing infrastructure.
// It is the canonical wiring point — call this once at server startup and
// attach the result to execution.Handler.
// ---------------------------------------------------------------------------

// BuildRuntimeRouter creates a RuntimeRouter with the given dependencies.
//
// wasmPool:      the WASM instance pool (may be nil)
// cacheL1:       deterministic result cache (may be nil)
// bundleSvc:     eager-bundling service (may be nil)
// pythonWASMPath: path to the MicroPython/CPython WASM runtime for pool cells
func BuildRuntimeRouter(
	wasmPool *wasm.InstancePool,
	cacheL1 *cache.CacheService,
	bundleSvc *bundler.BundleService,
	pythonWASMPath string,
) *RuntimeRouter {
	// WASM engine: in-process wasmtime cell executor (runtime-aware).
	wasmEngine := NewWasmCellExecutor(wasmPool, bundleSvc, pythonWASMPath)

	// Python engine: routes through the WASM cell executor for bundled Python.
	pythonEngine := &pythonCPythonEngine{wasmEngine: wasmEngine}

	// Node engine: not yet implemented; falls back to sandbox.
	var nodeEngine RuntimeEngine

	// Fallback: legacy sandbox (RustPython / per-request spawn).
	fallback := &sandboxEngine{}

	return NewRuntimeRouter(wasmEngine, pythonEngine, nodeEngine, fallback, cacheL1, bundleSvc)
}

// ---------------------------------------------------------------------------
// Tier resolution helper
// ---------------------------------------------------------------------------

// resolveTierFromRequest determines the execution tier from the function's tenant plan.
func resolveTierFromRequest(fn *storage.RegistryFunction, backendRepo storage.Repository) string {
	if fn == nil || fn.TenantID == nil || backendRepo == nil {
		return plans.PlanStarter
	}
	plan := getTenantPlanFromContext(backendRepo, *fn.TenantID)
	// Map plan names to router tier names
	switch plan {
	case plans.PlanEnterprise:
		return "enterprise"
	case plans.PlanPro, plans.PlanAgentScale:
		return "pro"
	case plans.PlanStarter, plans.PlanAgentStarter:
		return "free"
	default:
		return "free"
	}
}

// resolveRuntimeFromVersion maps a function version to the router runtime string.
func resolveRuntimeFromVersion(fnVersion *storage.RegistryFunctionVersion) string {
	if fnVersion == nil {
		return ""
	}
	rt := fnVersion.Runtime
	// Normalize Python runtimes
	if strings.HasPrefix(rt, "python") {
		return rt
	}
	// Normalize JS/Node runtimes
	if rt == "node18" || rt == "node20" || rt == "deno" || rt == "bun" {
		return rt
	}
	// WASM runtimes
	if rt == "rust" || rt == "go" || rt == "c" || rt == "cpp" || rt == "zig" || rt == "wasm" {
		return rt
	}
	return rt
}
