// Package executor provides Python WASM execution using wasmtime subprocess.
// This package requires CGO and is designed to run in the python-runtime service.
package executor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/functionfly/functionfly/cmd/python-runtime/wasmrunner"
)

// PythonExecutor wraps wasmtime-based Python execution with pooling.
type PythonExecutor struct {
	cpythonPath string
	pool        *cpython.Pool
	config      *cpython.Config
	metrics     *ExecutorMetrics
	mu          sync.RWMutex
	closed      bool
}

// ExecutorMetrics tracks execution metrics.
type ExecutorMetrics struct {
	Executions    int64
	Errors       int64
	TotalLatency time.Duration
	Active       int32
	Pooled       int32
}

// NewPythonExecutor creates a new Python executor with the given configuration.
func NewPythonExecutor(cpythonPath string, poolSize, maxMemoryMB int) (*PythonExecutor, error) {
	config := &cpython.Config{
		MaxMemoryMB:     maxMemoryMB,
		MaxExecutionSec: 30,
		PoolSize:        poolSize,
		PythonWasmPath:  cpythonPath,
		WasmtimePath:    "wasmtime", // Use system wasmtime
	}

	pool := cpython.NewPool(config, poolSize)

	return &PythonExecutor{
		cpythonPath: cpythonPath,
		pool:        pool,
		config:      config,
		metrics: &ExecutorMetrics{
			Pooled: int32(poolSize),
		},
	}, nil
}

// Execute runs the given Python code with the provided input.
func (e *PythonExecutor) Execute(ctx context.Context, code string, input []byte, timeoutMs int) (*ExecutionResult, error) {
	e.mu.RLock()
	if e.closed {
		e.mu.RUnlock()
		return nil, fmt.Errorf("executor is closed")
	}
	e.mu.RUnlock()

	// Get executor from pool
	exec, err := e.pool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor: %w", err)
	}
	defer e.pool.Put(exec)

	// Update active count
	atomic.AddInt32(&e.metrics.Active, 1)
	defer atomic.AddInt32(&e.metrics.Active, -1)

	start := time.Now()

	// Execute
	result, err := exec.Execute(code, input, timeoutMs/1000)
	latency := time.Since(start)

	// Update metrics
	atomic.AddInt64(&e.metrics.Executions, 1)
	if result != nil && result.Error != "" {
		atomic.AddInt64(&e.metrics.Errors, 1)
	}
	atomic.AddInt64((*int64)(&e.metrics.TotalLatency), int64(latency))

	if err != nil {
		return &ExecutionResult{
			Error:       err.Error(),
			LatencyMs:   latency.Milliseconds(),
		}, nil
	}

	if result != nil && result.Error != "" {
		return &ExecutionResult{
			Error:       result.Error,
			LatencyMs:   result.LatencyMs,
		}, nil
	}

	return &ExecutionResult{
		Output:      result.Output,
		LatencyMs:   result.LatencyMs,
		MemoryBytes: result.MemoryBytes,
	}, nil
}

// Prewarm initializes the executor pool.
func (e *PythonExecutor) Prewarm() error {
	// Prewarm is handled by pool initialization
	return nil
}

// Close shuts down the executor and releases resources.
func (e *PythonExecutor) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	return e.pool.Close()
}

// Stats returns current executor metrics.
func (e *PythonExecutor) Stats() ExecutorMetrics {
	return ExecutorMetrics{
		Executions: atomic.LoadInt64(&e.metrics.Executions),
		Errors:     atomic.LoadInt64(&e.metrics.Errors),
		Pooled:     atomic.LoadInt32(&e.metrics.Pooled),
		Active:     atomic.LoadInt32(&e.metrics.Active),
	}
}

// Healthy returns true if the executor can accept requests.
func (e *PythonExecutor) Healthy(ctx context.Context) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
		return false
	}
	return true // Pool will create new executors on demand
}

// ExecutionResult contains the result of a Python execution.
type ExecutionResult struct {
	Output      []byte `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	LatencyMs   int64  `json:"latency_ms"`
	MemoryBytes uint64 `json:"memory_bytes"`
}
