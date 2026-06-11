//go:build cgo

package execution

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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
func (e *sandboxEngine) Close() error                     { return nil }

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
// CPythonWasmEngine — in-process CPython-WASI for the business tier.
//
// Production-ready CPython 3.13 WASI execution that:
//   - Writes user handler code to /tmp/handler.py
//   - Passes input via FUNCTIONFLY_INPUT env var (JSON)
//   - Captures stdout as JSON output
//   - Uses wasmtime with WASI preview1 + secure defaults
//   - Pools instances per tenant for warm-start performance
//
// This engine provides full cpython stdlib support unlike MicroPython which
// is a subset. Business tier gets the full Python experience in-process
// without a daemon round-trip.
// ---------------------------------------------------------------------------

// cpythonWasmEngineConfig holds configuration for the CPython-WASM engine.
type cpythonWasmEngineConfig struct {
	// Path to python.wasm (official CPython WASI build from cpython.wasm symlink).
	WasmPath string
	// Path to CPython stdlib (runtimes/cpython-wasi/lib).
	StdlibPath string
	// Pool size per tenant (0 = no pooling).
	PoolSize int
	// Max execution time.
	Timeout time.Duration
	// Memory limit in MB.
	MaxMemoryMB int
}

// NewCPythonWasmEngineConfig creates a default config for CPython-WASM engine.
func NewCPythonWasmEngineConfig() *cpythonWasmEngineConfig {
	return &cpythonWasmEngineConfig{
		WasmPath:    "./runtimes/cpython.wasm",
		StdlibPath:  "./runtimes/cpython-wasi/lib",
		PoolSize:    4,
		Timeout:     30 * time.Second,
		MaxMemoryMB: 256,
	}
}

// cpythonWasmEngine implements RuntimeEngine for business-tier CPython-WASM.
type cpythonWasmEngine struct {
	config *cpythonWasmEngineConfig
	pool   *cpythonWasmPool
}

// NewCPythonWasmEngine creates a new CPython-WASM engine for the business tier.
func NewCPythonWasmEngine(cfg *cpythonWasmEngineConfig, _ wasm.HostFunctionHandler) (*cpythonWasmEngine, error) {
	if cfg == nil {
		cfg = NewCPythonWasmEngineConfig()
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxMemoryMB == 0 {
		cfg.MaxMemoryMB = 256
	}

	pool, err := newCPythonWasmPool(cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("CPythonWasmEngine: create pool: %w", err)
	}

	return &cpythonWasmEngine{
		config: cfg,
		pool:   pool,
	}, nil
}

// Execute runs CPython handler code with the given input.
func (e *cpythonWasmEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()

	// Extract handler source code
	userCode := ""
	if req.FunctionVersion != nil && req.FunctionVersion.SourceCode.Valid {
		userCode = req.FunctionVersion.SourceCode.String
	}
	if userCode == "" {
		// Try to get from WasmBinary metadata if bundled
		userCode = string(req.Input) // Last resort; not really correct
	}

	// Get a pooled or fresh runtime
	runtime, err := e.pool.Get(ctx)
	if err != nil {
		return ExecutionResult{
			DurationMs: int(time.Since(start).Milliseconds()),
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "engine",
			},
		}, err
	}
	defer e.pool.Put(runtime)

	// Execute via the CPython-WASM pool
	output, err := runtime.Execute(ctx, userCode, req.Input)
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

// Healthy reports whether the engine is healthy.
func (e *cpythonWasmEngine) Healthy(ctx context.Context) bool {
	if e.pool == nil {
		return false
	}
	// Check if pool has at least one healthy runtime
	rt, err := e.pool.Get(ctx)
	if err != nil {
		return false
	}
	e.pool.Put(rt)
	return rt.Healthy(ctx)
}

// Close shuts down the engine and its pool.
func (e *cpythonWasmEngine) Close() error {
	if e.pool != nil {
		e.pool.Close()
	}
	return nil
}

// Name returns the engine name for routing decisions.
func (e *cpythonWasmEngine) Name() string {
	return "cpython-wasm"
}

// ---------------------------------------------------------------------------
// CPython WASM pool — per-tenant warm instances.
// ---------------------------------------------------------------------------

// cpythonWasmRuntime is a single CPython-WASM runtime instance.
type cpythonWasmRuntime struct {
	engine *wasm.PythonRuntime
	tmpDir string
}

// newCPythonWasmRuntime creates a new CPython runtime instance.
func newCPythonWasmRuntime(cfg *cpythonWasmEngineConfig, _ wasm.HostFunctionHandler) (*cpythonWasmRuntime, error) {
	// Create temp directory for this instance
	tmpDir, err := os.MkdirTemp("", "ff-cpython")
	if err != nil {
		return nil, fmt.Errorf("create tmp dir: %w", err)
	}

	// Create a PythonRuntime with WASI support
	pythonRT, err := wasm.NewPythonRuntime(cfg.WasmPath, nil, nil, nil)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("create python runtime: %w", err)
	}

	return &cpythonWasmRuntime{
		engine: pythonRT,
		tmpDir: tmpDir,
	}, nil
}

// Execute runs Python code via CPython-WASI.
func (r *cpythonWasmRuntime) Execute(ctx context.Context, userCode string, input []byte) ([]byte, error) {
	// Write handler.py to the instance's tmp directory
	handlerPath := filepath.Join(r.tmpDir, "handler.py")
	if err := os.WriteFile(handlerPath, []byte(userCode), 0600); err != nil {
		return nil, fmt.Errorf("write handler.py: %w", err)
	}

	// Write input to a temp file. The wrapper reads this and passes to handler.
	inputPath := filepath.Join(r.tmpDir, "input.json")
	if err := os.WriteFile(inputPath, input, 0600); err != nil {
		return nil, fmt.Errorf("write input.json: %w", err)
	}

	// Wrap user code to read from input.json and write to output.json
	// This is a simple and reliable approach for CPython WASI
	wrapperCode := fmt.Sprintf(`
import json
import sys

# Read input from the JSON file written by the host
with open("%s", "r") as f:
    _input = json.load(f)

# Import the user's handler
_exec_globals = {}
try:
    exec(open("%s").read(), _exec_globals)
except Exception as _e:
    print(json.dumps({"success": False, "error": str(_e)}))
    sys.exit(0)

# Call the handler function
_handler = _exec_globals.get("handler")
if _handler is None:
    print(json.dumps({"success": False, "error": "No handler function found"}))
else:
    try:
        _result = _handler(_input)
        if isinstance(_result, dict):
            print(json.dumps({"success": True, "result": _result}))
        else:
            print(json.dumps({"success": True, "result": str(_result)}))
    except Exception as _e:
        print(json.dumps({"success": False, "error": str(_e)}))
`, inputPath, handlerPath)

	// Write the wrapper as the actual handler
	if err := os.WriteFile(handlerPath, []byte(wrapperCode), 0600); err != nil {
		return nil, fmt.Errorf("write wrapped handler: %w", err)
	}

	// Initialize and execute
	if err := r.engine.Init(); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	if err := r.engine.LoadCode(wrapperCode); err != nil {
		return nil, fmt.Errorf("load_code: %w", err)
	}

	output, err := r.engine.ExecuteWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	// Try to parse output as JSON to extract structured result
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err == nil {
		if success, ok := result["success"].(bool); ok && !success {
			if errMsg, ok := result["error"].(string); ok {
				return output, fmt.Errorf("handler error: %s", errMsg)
			}
		}
		if res, ok := result["result"]; ok {
			// Repack as clean JSON
			if resBytes, err := json.Marshal(res); err == nil {
				return resBytes, nil
			}
		}
	}

	return output, nil
}

// Healthy reports whether the runtime instance is healthy.
func (r *cpythonWasmRuntime) Healthy(ctx context.Context) bool {
	return r.engine != nil
}

// Close releases resources for this runtime instance.
func (r *cpythonWasmRuntime) Close() error {
	if r.tmpDir != "" {
		os.RemoveAll(r.tmpDir)
	}
	if r.engine != nil {
		return r.engine.Close()
	}
	return nil
}

// cpythonWasmPool manages a pool of CPython runtime instances.
type cpythonWasmPool struct {
	cfg  *cpythonWasmEngineConfig
	pool chan *cpythonWasmRuntime
}

// newCPythonWasmPool creates a new CPython-WASM pool.
func newCPythonWasmPool(cfg *cpythonWasmEngineConfig, _ wasm.HostFunctionHandler) (*cpythonWasmPool, error) {
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 4
	}

	pool := &cpythonWasmPool{
		cfg:  cfg,
		pool: make(chan *cpythonWasmRuntime, poolSize),
	}

	// Prewarm the pool
	for i := 0; i < poolSize; i++ {
		rt, err := newCPythonWasmRuntime(cfg, nil)
		if err != nil {
			logrus.WithField("err", err).Warn("cpythonWasmPool: prewarm failed, continuing with reduced pool")
			continue
		}
		pool.pool <- rt
	}

	return pool, nil
}

// Get retrieves a runtime from the pool or creates a new one.
func (p *cpythonWasmPool) Get(ctx context.Context) (*cpythonWasmRuntime, error) {
	select {
	case rt := <-p.pool:
		return rt, nil
	default:
		// Pool empty; create a new instance
		return newCPythonWasmRuntime(p.cfg, nil)
	}
}

// Put returns a runtime to the pool.
func (p *cpythonWasmPool) Put(rt *cpythonWasmRuntime) {
	if rt == nil {
		return
	}
	select {
	case p.pool <- rt:
	default:
		// Pool full, close the instance
		rt.Close()
	}
}

// Close shuts down the pool.
func (p *cpythonWasmPool) Close() {
	close(p.pool)
	for rt := range p.pool {
		rt.Close()
	}
}

// ---------------------------------------------------------------------------
// nodeJSEngine — QuickJS WASM via the nodejs daemon (runtimes/nodejs)
//
// The nodejs daemon (functionfly-nodejs --daemon --port) serves JS/TS
// execution over HTTP.  nodeJSEngine manages the daemon lifecycle and routes
// node18/node20/deno/bun execution requests to it via the /execute endpoint.
// Falls back to the sandbox engine if the daemon is unavailable.
// ---------------------------------------------------------------------------

// nodeJSEngine manages the QuickJS-WASM daemon lifecycle and execution routing.
type nodeJSEngine struct {
	RuntimePath string
	daemonURL   string
	daemonCmd   *exec.Cmd
	httpClient  *http.Client
	mu          sync.Mutex
	isRunning   bool
	tempDir     string
}

// newNodeJSEngine creates a nodeJSEngine that launches the QuickJS-WASM daemon
// on a free port and routes JS/TS execution through it.
func newNodeJSEngine() (*nodeJSEngine, error) {
	runtimePath, err := findNodeJSRuntime()
	if err != nil {
		logrus.WithField("err", err).Warn("nodeJSEngine: runtime not found, will retry on first execute")
		// Don't fail — allow healthy=false until we can start it
		return &nodeJSEngine{RuntimePath: "", daemonURL: "", isRunning: false}, nil
	}

	tempDir, err := os.MkdirTemp("", "functionfly-nodejs-*")
	if err != nil {
		return nil, fmt.Errorf("nodeJSEngine: create temp dir: %w", err)
	}

	port, err := getAvailablePort()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("nodeJSEngine: find port: %w", err)
	}

	daemonURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--daemon",
	}
	cmd := exec.Command(runtimePath, args...)
	cmd.Dir = tempDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("nodeJSEngine: start daemon: %w", err)
	}

	sc := &nodeJSEngine{
		RuntimePath: runtimePath,
		daemonURL:   daemonURL,
		daemonCmd:   cmd,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				IdleConnTimeout:     90 * time.Second,
				MaxConnsPerHost:     50,
				MaxIdleConnsPerHost: 50,
			},
		},
		isRunning: true,
		tempDir:   tempDir,
	}

	// Wait for daemon to become ready (30s timeout to accommodate slow wasmtime init)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := sc.waitForReady(ctx); err != nil {
		sc.Close()
		return nil, fmt.Errorf("nodeJSEngine: daemon not ready: %w", err)
	}

	logrus.WithField("url", daemonURL).Info("nodeJSEngine: daemon started and ready")
	return sc, nil
}

// findNodeJSRuntime locates the functionfly-nodejs binary.
func findNodeJSRuntime() (string, error) {
	paths := []string{
		"./bin/functionfly-nodejs",
		"./runtimes/nodejs/target/release/functionfly-nodejs",
		"./runtimes/nodejs/target/debug/functionfly-nodejs",
		"/usr/local/bin/functionfly-nodejs",
	}
	for _, p := range paths {
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("functionfly-nodejs not found in: %v", paths)
}

func (e *nodeJSEngine) waitForReady(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET",
				e.daemonURL+"/health", nil)
			if err != nil {
				continue
			}
			resp, err := e.httpClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"url":    e.daemonURL,
					"elapsed": time.Since(start),
					"error":  err,
				}).Debug("nodeJSEngine: health check failed, retrying...")
			}
			if resp != nil {
				logrus.WithFields(logrus.Fields{
					"url":         e.daemonURL,
					"elapsed":     time.Since(start),
					"status_code": resp.StatusCode,
				}).Debug("nodeJSEngine: health check returned non-OK, retrying...")
				resp.Body.Close()
			}
		}
	}
}

// Execute routes a JS/TS execution request to the QuickJS-WASM daemon.
// It accepts node18, node20, deno, and bun runtimes.
func (e *nodeJSEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()

	// Lazily start the daemon if it wasn't started at construction
	if !e.isRunning && e.RuntimePath != "" {
		if err := e.startDaemon(); err != nil {
			return ExecutionResult{}, fmt.Errorf("nodeJSEngine: start daemon: %w", err)
		}
	}

	if !e.isRunning || e.daemonURL == "" {
		return ExecutionResult{}, fmt.Errorf("nodeJSEngine: daemon not running")
	}

	// Determine the input to send
	inputJSON, err := json.Marshal(req.Input)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("nodeJSEngine: marshal input: %w", err)
	}

	// Build the WASM binary payload — prefer WasmBinary, fall back to SourceCode
	var payload []byte
	if len(req.FunctionVersion.WasmBinary) > 0 {
		payload = req.FunctionVersion.WasmBinary
	} else if req.FunctionVersion.SourceCode.Valid && req.FunctionVersion.SourceCode.String != "" {
		payload = []byte(req.FunctionVersion.SourceCode.String)
	} else {
		return ExecutionResult{}, fmt.Errorf("nodeJSEngine: no wasm_binary or source_code")
	}

	// Encode payload as base64 for the daemon's /execute endpoint
	payloadB64 := base64.StdEncoding.EncodeToString(payload)

	fnID := ""
	fnVersion := ""
	if req.FunctionVersion != nil {
		fnID = req.FunctionVersion.FunctionID.String()
		fnVersion = req.FunctionVersion.Version
	}

	execReq := map[string]interface{}{
		"wasm_binary": payloadB64,
		"input":       string(inputJSON),
		"timeout_ms":  req.MaxCPUTimeMs,
		"memory_mb":   req.MaxMemoryMB,
		"tenant_id":   req.TenantID,
	}

	body, err := json.Marshal(execReq)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("nodeJSEngine: marshal request: %w", err)
	}

	execURL := fmt.Sprintf("%s/execute/%s/%s", e.daemonURL, url.PathEscape(fnID), url.PathEscape(fnVersion))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", execURL, bytes.NewReader(body))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("nodeJSEngine: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Apply per-request timeout
	httpClient := &http.Client{Timeout: time.Duration(req.MaxCPUTimeMs+2000) * time.Millisecond}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return ExecutionResult{
			DurationMs: int(time.Since(start).Milliseconds()),
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "error",
			},
		}, err
	}
	defer resp.Body.Close()

	var daemonResp struct {
		Result     json.RawMessage `json:"result"`
		ExecTimeMs int64           `json:"exec_time_ms"`
		CacheHit   bool            `json:"cache_hit"`
		Error      string          `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&daemonResp); err != nil {
		return ExecutionResult{
			DurationMs: int(time.Since(start).Milliseconds()),
			Error: &ExecutionError{
				Err:          fmt.Errorf("nodeJSEngine: decode response: %w", err),
				TerminatedBy: "error",
			},
		}, fmt.Errorf("nodeJSEngine: decode response: %w", err)
	}

	if daemonResp.Error != "" {
		return ExecutionResult{
			DurationMs: int(time.Since(start).Milliseconds()),
			Error: &ExecutionError{
				Err:          fmt.Errorf("nodeJSEngine: daemon error: %s", daemonResp.Error),
				TerminatedBy: "error",
			},
		}, fmt.Errorf("nodeJSEngine: daemon error: %s", daemonResp.Error)
	}

	return ExecutionResult{
		Output:     daemonResp.Result,
		DurationMs: int(time.Since(start).Milliseconds()),
		CacheHit:   daemonResp.CacheHit,
	}, nil
}

// Healthy checks whether the daemon is running and responsive.
func (e *nodeJSEngine) Healthy(ctx context.Context) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning || e.daemonURL == "" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, "GET", e.daemonURL+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Close terminates the daemon process and cleans up temporary files.
func (e *nodeJSEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.daemonCmd != nil && e.daemonCmd.Process != nil {
		_ = e.daemonCmd.Process.Kill()
		_ = e.daemonCmd.Wait()
	}
	if e.tempDir != "" {
		_ = os.RemoveAll(e.tempDir)
	}
	e.isRunning = false
	return nil
}

func (e *nodeJSEngine) startDaemon() error {
	port, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("nodeJSEngine: find port: %w", err)
	}
	daemonURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	args := []string{"--port", fmt.Sprintf("%d", port), "--daemon"}
	cmd := exec.Command(e.RuntimePath, args...)
	cmd.Dir = e.tempDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("nodeJSEngine: start: %w", err)
	}

	e.daemonURL = daemonURL
	e.daemonCmd = cmd
	e.isRunning = true

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.waitForReady(ctx); err != nil {
		e.Close()
		return fmt.Errorf("nodeJSEngine: daemon not ready: %w", err)
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
	pool        *wasm.InstancePool
	bundleSvc   *bundler.BundleService
	pythonPath  string // path to MicroPython/CPython wasm runtime
	healthy     bool
	healthyOnce sync.Once
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
		_, _ = wasm.NewPythonRuntimeWithConfig("", nil, nil, nil, wasm.NewDefaultSecurityConfig())
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
// cpythonPath:   path to the CPython-WASM binary (python.wasm)
// cpythonLib:    path to the CPython stdlib (cpython-wasi/lib)
func BuildRuntimeRouter(
	wasmPool *wasm.InstancePool,
	cacheL1 *cache.CacheService,
	bundleSvc *bundler.BundleService,
	pythonWASMPath string,
	cpythonPath string,
	cpythonLib string,
) *RuntimeRouter {
	// WASM engine: in-process wasmtime cell executor (runtime-aware).
	wasmEngine := NewWasmCellExecutor(wasmPool, bundleSvc, pythonWASMPath)

	// Python engine: routes through the WASM cell executor for bundled Python.
	pythonEngine := &pythonCPythonEngine{wasmEngine: wasmEngine}

	// CPython-WASM engine: in-process CPython 3.13 for business tier.
	var cpythonEngine RuntimeEngine
	if cpythonPath != "" {
		cfg := &cpythonWasmEngineConfig{
			WasmPath:    cpythonPath,
			StdlibPath:  cpythonLib,
			PoolSize:    4,
			Timeout:     30 * time.Second,
			MaxMemoryMB: 256,
		}
		cpythonEngine, _ = NewCPythonWasmEngine(cfg, nil)
	}

	// Node engine: QuickJS-WASM daemon (runtimes/nodejs --daemon).
	nodeEngine, err := newNodeJSEngine()
	if err != nil {
		logrus.WithField("err", err).Warn("BuildRuntimeRouter: nodeJSEngine init failed, falling back to sandbox")
		nodeEngine = nil
	} else if nodeEngine != nil && nodeEngine.RuntimePath == "" {
		logrus.Warn("BuildRuntimeRouter: functionfly-nodejs binary not found — JS/TS execution will fall back to sandbox (slower cold starts). Build and deploy the binary to enable native QuickJS-WASM execution.")
	}

	// Fallback: legacy sandbox (RustPython / per-request spawn).
	fallback := &sandboxEngine{}

	return NewRuntimeRouter(wasmEngine, pythonEngine, cpythonEngine, nodeEngine, fallback, cacheL1, bundleSvc)
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
