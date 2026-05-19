//go:build cgo

// Package cpython provides CPython WASI execution.
// Uses subprocess invocation of wasmtime CLI - simpler than embedding wasmtime-go
// and avoids the API mismatch between MicroPython and CPython WASM.
package cpython

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Config holds CPython WASI configuration.
type Config struct {
	MaxMemoryMB     int
	MaxExecutionSec int
	PoolSize        int
	WasmtimePath    string
	PythonWasmPath  string
	StdlibPath      string // path to CPython lib directory
	MapDir          string // directory mapping for wasmtime (e.g., "/home/user:/path/to/cpython-wasi")
}

// DefaultConfig returns default CPython WASI configuration.
func DefaultConfig() *Config {
	return &Config{
		MaxMemoryMB:     512,
		MaxExecutionSec: 30,
		PoolSize:        4,
		WasmtimePath:   "wasmtime",
		PythonWasmPath: "./runtimes/cpython-wasi/python.wasm",
		StdlibPath:     "./runtimes/cpython-wasi/lib",
		// Map cpython-wasi dir to /cpython so CPython can find its stdlib
		MapDir:         "./runtimes:/cpython",
	}
}

// ExecuteRequest contains the code and input for execution.
type ExecuteRequest struct {
	Code  string          `json:"code"`
	Input json.RawMessage `json:"input,omitempty"`
}

// ExecuteResult contains the execution result.
type ExecuteResult struct {
	Output      []byte `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	LatencyMs   int64  `json:"latency_ms"`
	MemoryBytes uint64 `json:"memory_bytes,omitempty"`
}

// Executor runs Python code via wasmtime subprocess.
type Executor struct {
	config *Config
}

// NewExecutor creates a new CPython executor.
func NewExecutor(config *Config) *Executor {
	if config == nil {
		config = DefaultConfig()
	}
	return &Executor{config: config}
}

// Execute runs Python code and returns the result.
func (e *Executor) Execute(code string, input []byte, timeoutSec int) (*ExecuteResult, error) {
	startTime := time.Now()

	if timeoutSec <= 0 {
		timeoutSec = e.config.MaxExecutionSec
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Create Python handler that reads input and executes code
	// The input is passed as a JSON string that gets parsed by Python
	var inputJSON string
	if len(input) == 0 || string(input) == "null" {
		inputJSON = "null"
	} else {
		inputJSON = string(input)
	}

	pythonHandler := fmt.Sprintf(`
import json, sys
input_json = %q
if input_json.strip() in ('null', 'None', ''):
    input_data = {}
elif input_json:
    input_data = json.loads(input_json)
else:
    input_data = {}
result = {}
try:
    exec(%q)
    print(json.dumps(result))
except Exception as ex:
    print(json.dumps({'status': 'error', 'error': str(ex)}))
`, string(inputJSON), code)

	// Build wasmtime command with directory mapping for CPython stdlib
	// Map runtimes dir to /cpython so CPython can find lib at /cpython/cpython-wasi/lib
	cmd := exec.CommandContext(ctx, e.config.WasmtimePath,
		"run",
		"--dir", "./runtimes::/cpython",
		"--env", "PYTHONPATH=/cpython/cpython-wasi/lib",
		"--env", "PYTHONHOME=/cpython/cpython-wasi",
		e.config.PythonWasmPath,
		"-c", pythonHandler,
	)

	// Capture stdout
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run
	err := cmd.Run()
	latencyMs := time.Since(startTime).Milliseconds()

	result := &ExecuteResult{
		Output:    stdout.Bytes(),
		LatencyMs: latencyMs,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("execution timeout after %d seconds", timeoutSec)
		} else {
			result.Error = fmt.Sprintf("execution failed: %v, stderr: %s", err, stderr.String())
		}
		result.LatencyMs = time.Since(startTime).Milliseconds()
	}

	return result, nil
}

// Pool manages a pool of CPython executors.
// Since CPython runs as a subprocess, we can have multiple ready executors.
type Pool struct {
	executors chan *Executor
	config    *Config
	maxSize   int
}

// NewPool creates a new CPython executor pool.
func NewPool(config *Config, maxSize int) *Pool {
	if maxSize <= 0 {
		maxSize = 4
	}
	if config == nil {
		config = DefaultConfig()
	}

	pool := &Pool{
		executors: make(chan *Executor, maxSize),
		config:    config,
		maxSize:   maxSize,
	}

	// Pre-populate the pool
	for i := 0; i < maxSize; i++ {
		pool.executors <- NewExecutor(config)
	}

	return pool
}

// Get retrieves an executor from the pool.
func (p *Pool) Get(ctx context.Context) (*Executor, error) {
	select {
	case exec := <-p.executors:
		return exec, nil
	default:
		// Pool exhausted, create a new one
		return NewExecutor(p.config), nil
	}
}

// Put returns an executor to the pool.
func (p *Pool) Put(exec *Executor) {
	if exec == nil {
		return
	}
	select {
	case p.executors <- exec:
	default:
		// Pool full, discard
	}
}

// Close closes the pool.
func (p *Pool) Close() error {
	close(p.executors)
	return nil
}