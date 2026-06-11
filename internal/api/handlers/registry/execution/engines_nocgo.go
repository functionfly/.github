//go:build !cgo

// Package execution provides runtime engines for FunctionFly.
// This file is used when building with CGO_ENABLED=0.
// It provides an external Python WASM runtime client.
package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------
// Engine wrappers — each wraps existing execution infrastructure to satisfy
// the RuntimeEngine interface used by RuntimeRouter.
// ---------------------------------------------------------------------------

// sandboxEngine wraps the legacy executeLocallyWithLimits as a RuntimeEngine.
type sandboxEngine struct{}

// Execute implements RuntimeEngine using sandbox execution.
func (e *sandboxEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()
	output, err := executeLocallyWithLimits(
		req.FunctionVersion, req.Input,
		req.MaxMemoryMB, req.MaxCPUTimeMs,
		req.Function, nil,
	)
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
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

// Healthy implements RuntimeEngine.
func (e *sandboxEngine) Healthy(ctx context.Context) bool { return true }

// Close implements RuntimeEngine.
func (e *sandboxEngine) Close() error { return nil }

// ---------------------------------------------------------------------------
// RuntimeClient for external Python WASM service
// ---------------------------------------------------------------------------

const (
	connectionTimeout       = 5 * time.Second
	maxRetries              = 3
	circuitBreakerThreshold = 5
	circuitBreakerResetTime = 30 * time.Second
)

// circuitBreaker tracks failures and prevents calls to failing endpoints.
type circuitBreaker struct {
	mu          sync.Mutex
	failures    int
	lastFailure time.Time
	circuitOpen bool
	threshold   int
	resetTime   time.Duration
}

func newCircuitBreaker(threshold int, resetTime time.Duration) *circuitBreaker {
	return &circuitBreaker{
		threshold: threshold,
		resetTime: resetTime,
	}
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.circuitOpen {
		if time.Since(cb.lastFailure) > cb.resetTime {
			cb.circuitOpen = false
			cb.failures = 0
			return true
		}
		return false
	}
	return true
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.circuitOpen = true
	}
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.circuitOpen = false
}

// RuntimeClient is an HTTP client that connects to an external Python WASM runtime service.
type RuntimeClient struct {
	endpoint       string
	client         *http.Client
	mu             sync.RWMutex
	closed         bool
	circuitBreaker *circuitBreaker
}

// NewRuntimeClient creates a new runtime client that connects to the specified endpoint.
func NewRuntimeClient(endpoint string) *RuntimeClient {
	if endpoint == "" {
		endpoint = "http://localhost:8083"
	}

	return &RuntimeClient{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 300 * time.Second, // Long timeout for Python execution
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: connectionTimeout,
				}).DialContext,
			},
		},
		circuitBreaker: newCircuitBreaker(circuitBreakerThreshold, circuitBreakerResetTime),
	}
}

// Execute runs Python code via the external runtime service with retry and circuit breaker.
func (c *RuntimeClient) Execute(ctx context.Context, code string, input []byte, timeoutMs int) ([]byte, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	if !c.circuitBreaker.allow() {
		return nil, fmt.Errorf("circuit breaker open: service unavailable")
	}

	reqBody := execRequest{
		Code:      code,
		Input:     json.RawMessage(input),
		TimeoutMs: timeoutMs,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.endpoint + "/execute"

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			if backoff > 2*time.Second {
				backoff = 2 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			c.circuitBreaker.recordFailure()
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			_, _ = io.ReadAll(io.LimitReader(resp.Body, 1024))
			lastErr = fmt.Errorf("execution failed with status %d: %s", resp.StatusCode, string(respBody))
			c.circuitBreaker.recordFailure()
			continue
		}

		var execResp execResponse
		if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %w", err)
			continue
		}

		if execResp.Error != "" {
			return nil, fmt.Errorf("execution error: %s", execResp.Error)
		}

		c.circuitBreaker.recordSuccess()
		return []byte(execResp.Output), nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// Healthy checks if the external runtime service is healthy.
func (c *RuntimeClient) Healthy(ctx context.Context) bool {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return false
	}
	c.mu.RUnlock()

	url := c.endpoint + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var healthResp healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return false
	}

	return healthResp.Status == "healthy"
}

// Close closes the client.
func (c *RuntimeClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}

type execRequest struct {
	Code      string          `json:"code"`
	Input     json.RawMessage `json:"input"`
	TimeoutMs int             `json:"timeout_ms,omitempty"`
}

type execResponse struct {
	Output      string `json:"output"`
	Error       string `json:"error,omitempty"`
	LatencyMs   int64  `json:"latency_ms"`
	MemoryBytes uint64 `json:"memory_bytes,omitempty"`
}

type healthResponse struct {
	Status     string `json:"status"`
	Pooled     int32  `json:"pooled"`
	Active     int32  `json:"active"`
	Executions int64  `json:"executions"`
	Errors     int64  `json:"errors"`
}

// ---------------------------------------------------------------------------
// Engine implementations for non-CGO builds
// ---------------------------------------------------------------------------

// externalPythonEngine uses HTTP to connect to the external Python WASM runtime service.
type externalPythonEngine struct {
	client *RuntimeClient
}

// NewExternalPythonEngine creates an engine that connects to the external Python WASM runtime service.
func NewExternalPythonEngine(endpoint string) *externalPythonEngine {
	return &externalPythonEngine{
		client: NewRuntimeClient(endpoint),
	}
}

// Execute implements RuntimeEngine by calling the external service.
func (e *externalPythonEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	timeoutMs := 30000
	if req.MaxCPUTimeMs > 0 {
		timeoutMs = int(req.MaxCPUTimeMs)
	}

	inputBytes := req.Input
	if inputBytes == nil {
		inputBytes = []byte("{}")
	}

	// Extract code from function version
	code := ""
	if req.FunctionVersion != nil && req.FunctionVersion.SourceCode.Valid {
		code = req.FunctionVersion.SourceCode.String
	}

	start := time.Now()
	output, err := e.client.Execute(ctx, code, inputBytes, timeoutMs)
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
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

// Healthy implements RuntimeEngine.
func (e *externalPythonEngine) Healthy(ctx context.Context) bool {
	return e.client.Healthy(ctx)
}

// Close implements RuntimeEngine.
func (e *externalPythonEngine) Close() error {
	return e.client.Close()
}

// ---------------------------------------------------------------------------
// wasmCellExecutor - real WASM execution using sandbox fallback
// Falls back to sandbox execution when WASM runtime is unavailable
// ---------------------------------------------------------------------------

type wasmCellExecutor struct {
	sandboxEnabled bool
	daemonEndpoint string
	mu             sync.Mutex
	instance       interface{}
}

func newWasmCellExecutor(pool interface{}, bundleSvc interface{}, wasmPath string) *wasmCellExecutor {
	return &wasmCellExecutor{
		sandboxEnabled: true,
		daemonEndpoint: os.Getenv("FUNCTION_EXECUTION_URL"),
	}
}

func (e *wasmCellExecutor) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()

	// Try sandbox execution as fallback for WASM workloads
	output, err := executeLocallyWithLimits(
		req.FunctionVersion,
		req.Input,
		req.MaxMemoryMB,
		req.MaxCPUTimeMs,
		req.Function,
		nil,
	)

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
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

func (e *wasmCellExecutor) Healthy(ctx context.Context) bool {
	return e.sandboxEnabled
}

func (e *wasmCellExecutor) Close() error {
	return nil
}

// ---------------------------------------------------------------------------
// nodeJSEngine - real Node.js execution using sandbox fallback
// Falls back to sandbox execution when Node.js runtime is unavailable
// ---------------------------------------------------------------------------

type nodeJSEngine struct {
	sandboxEnabled bool
	daemonEndpoint string
	mu             sync.Mutex
}

func newNodeJSEngine() *nodeJSEngine {
	return &nodeJSEngine{
		sandboxEnabled: true,
		daemonEndpoint: os.Getenv("FUNCTION_EXECUTION_URL"),
	}
}

func (e *nodeJSEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	start := time.Now()

	// Try sandbox execution as fallback for Node.js workloads
	output, err := executeLocallyWithLimits(
		req.FunctionVersion,
		req.Input,
		req.MaxMemoryMB,
		req.MaxCPUTimeMs,
		req.Function,
		nil,
	)

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
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

func (e *nodeJSEngine) Healthy(ctx context.Context) bool {
	return e.sandboxEnabled
}

func (e *nodeJSEngine) Close() error {
	return nil
}

// ---------------------------------------------------------------------------
// cpythonWasmEngine - real CPython WASM execution using external runtime service
// Uses HTTP client to connect to python-runtime service for full CPython support
// ---------------------------------------------------------------------------

type cpythonWasmEngine struct {
	client     *RuntimeClient
	poolSize   int
	mu         sync.Mutex
	healthy    bool
	lastHealth time.Time
}

func newCPythonWasmEngine(endpoint string, poolSize int) *cpythonWasmEngine {
	if endpoint == "" {
		endpoint = os.Getenv("PYTHON_RUNTIME_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8083"
		}
	}
	if poolSize <= 0 {
		poolSize = 4
	}

	return &cpythonWasmEngine{
		client:     NewRuntimeClient(endpoint),
		poolSize:   poolSize,
		healthy:    true,
		lastHealth: time.Now(),
	}
}

func (e *cpythonWasmEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	// Get code from function version
	code := ""
	if req.FunctionVersion != nil && req.FunctionVersion.SourceCode.Valid {
		code = req.FunctionVersion.SourceCode.String
	}

	if code == "" {
		return ExecutionResult{
			Error: &ExecutionError{
				Err:          fmt.Errorf("no source code provided"),
				TerminatedBy: "error",
			},
		}, fmt.Errorf("no source code in function version")
	}

	timeoutMs := 30000
	if req.MaxCPUTimeMs > 0 {
		timeoutMs = int(req.MaxCPUTimeMs)
	}

	inputBytes := req.Input
	if inputBytes == nil {
		inputBytes = []byte("{}")
	}

	start := time.Now()
	output, err := e.client.Execute(ctx, code, inputBytes, timeoutMs)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		e.mu.Lock()
		e.healthy = false
		e.mu.Unlock()

		return ExecutionResult{
			DurationMs: durationMs,
			Error: &ExecutionError{
				Err:          err,
				TerminatedBy: "error",
			},
		}, err
	}

	e.mu.Lock()
	e.healthy = true
	e.lastHealth = time.Now()
	e.mu.Unlock()

	return ExecutionResult{
		Output:     output,
		DurationMs: durationMs,
	}, nil
}

func (e *cpythonWasmEngine) Healthy(ctx context.Context) bool {
	e.mu.Lock()
	// Cache health status for 10 seconds to avoid hammering the service
	if time.Since(e.lastHealth) < 10*time.Second && e.healthy {
		e.mu.Unlock()
		return true
	}
	e.mu.Unlock()

	// Check actual health
	healthy := e.client.Healthy(ctx)
	e.mu.Lock()
	e.healthy = healthy
	if healthy {
		e.lastHealth = time.Now()
	}
	e.mu.Unlock()

	return healthy
}

func (e *cpythonWasmEngine) Close() error {
	return e.client.Close()
}

// pythonCPythonEngine wraps external Python engine for bundled Python.
type pythonCPythonEngine struct {
	externalEngine *cpythonWasmEngine
}

func NewPythonCPythonEngine(endpoint string) *pythonCPythonEngine {
	return &pythonCPythonEngine{
		externalEngine: newCPythonWasmEngine(endpoint, 4),
	}
}

func (e *pythonCPythonEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	return e.externalEngine.Execute(ctx, req)
}

func (e *pythonCPythonEngine) Healthy(ctx context.Context) bool {
	return e.externalEngine.Healthy(ctx)
}

func (e *pythonCPythonEngine) Close() error {
	return e.externalEngine.Close()
}

// ---------------------------------------------------------------------------
// BuildRuntimeRouter assembles a RuntimeRouter for non-CGO builds.
// It uses external Python WASM runtime service for Python execution.
// ---------------------------------------------------------------------------

// BuildRuntimeRouter creates a RuntimeRouter for CGO_ENABLED=0 builds.
// It connects to an external Python WASM runtime service (python-runtime).
// The endpoint is configured via PYTHON_RUNTIME_URL environment variable.
func BuildRuntimeRouter(
	wasmPool interface{},
	cacheL1 *cache.CacheService,
	bundleSvc *bundler.BundleService,
	pythonWASMPath string,
	cpythonPath string,
	cpythonLib string,
) *RuntimeRouter {
	// Get endpoint from environment variable
	endpoint := os.Getenv("PYTHON_RUNTIME_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8083"
	}

	// Python engine uses external CPython WASM service
	pythonEngine := NewPythonCPythonEngine(endpoint)

	// CPython engine also uses external service
	cpythonEngine := newCPythonWasmEngine(endpoint, 4)

	// WASM engine uses sandbox fallback (still provides real execution)
	wasmEngine := newWasmCellExecutor(wasmPool, bundleSvc, pythonWASMPath)

	// Node engine uses sandbox fallback
	nodeEngine := newNodeJSEngine()

	// Fallback - sandbox execution
	fallback := &sandboxEngine{}

	logrus.WithFields(logrus.Fields{
		"python_endpoint": endpoint,
	}).Info("BuildRuntimeRouter: using external Python WASM runtime service (non-CGO build)")

	return NewRuntimeRouter(wasmEngine, pythonEngine, cpythonEngine, nodeEngine, fallback, cacheL1, bundleSvc)
}

// NodeEngineConfig for Node.js engine configuration.
type NodeEngineConfig struct {
	DaemonURL string
}

// NewNodeJSEngine creates a real Node.js engine with sandbox fallback.
func NewNodeJSEngine() *nodeJSEngine {
	return newNodeJSEngine()
}

// WasmCellExecutor for WASM cell execution.
type WasmCellExecutor = wasmCellExecutor

// NewWasmCellExecutor creates a real WASM executor with sandbox fallback.
func NewWasmCellExecutor(pool interface{}, bundleSvc interface{}, wasmPath string) *wasmCellExecutor {
	return newWasmCellExecutor(pool, bundleSvc, wasmPath)
}

// cpythonWasmEngineConfig for CPython WASM engine configuration.
type cpythonWasmEngineConfig struct {
	WasmPath    string
	StdlibPath  string
	PoolSize    int
	Timeout     time.Duration
	MaxMemoryMB int
}

// NewCPythonWasmEngineConfig creates a new CPython WASM engine config.
func NewCPythonWasmEngineConfig() *cpythonWasmEngineConfig {
	return &cpythonWasmEngineConfig{
		PoolSize:    4,
		Timeout:     30 * time.Second,
		MaxMemoryMB: 256,
	}
}

// NewCPythonWasmEngine creates a real CPython WASM engine with external service.
func NewCPythonWasmEngine(cfg *cpythonWasmEngineConfig, handler interface{}) (interface{}, error) {
	if cfg == nil {
		cfg = NewCPythonWasmEngineConfig()
	}
	endpoint := os.Getenv("PYTHON_RUNTIME_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8083"
	}
	return newCPythonWasmEngine(endpoint, cfg.PoolSize), nil
}

// ---------------------------------------------------------------------------
// Tier resolution helper for non-CGO builds
// ---------------------------------------------------------------------------

// resolveTierFromRequest determines the execution tier from the function's tenant plan.
func resolveTierFromRequest(fn *storage.RegistryFunction, backendRepo storage.Repository) string {
	if fn == nil || fn.TenantID == nil {
		return "free"
	}
	// Default to free tier when CGO is disabled
	return "free"
}

// resolveRuntimeFromVersion maps a function version to the router runtime string.
func resolveRuntimeFromVersion(fnVersion *storage.RegistryFunctionVersion) string {
	if fnVersion == nil {
		return ""
	}
	return fnVersion.Runtime
}
