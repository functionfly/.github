//go:build cgo

// go_runtime.go implements the production Go runtime on top of wasmtime.
//
// File name does NOT end in `_wasm.go` because the Go toolchain treats
// that suffix as GOARCH=wasm-only sources.
package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// GoRuntime executes a single Go WASM module via wasmtime.
type GoRuntime struct {
	mu        sync.Mutex
	wasmBytes []byte
	cfg       GoRuntimeConfig
	handler   HostFunctionHandler
	scanner   *RuntimeScanner
	workDir   string

	engine *wasmtime.Engine
	module *wasmtime.Module
	store  *wasmtime.Store

	createdAt time.Time
	execCount int64
}

// NewGoRuntime compiles, links, and instantiates the Go module.
func NewGoRuntime(wasmBytes []byte, handler HostFunctionHandler, cfg GoRuntimeConfig) (*GoRuntime, error) {
	if len(wasmBytes) == 0 {
		return nil, fmt.Errorf("go runtime: empty wasm bytes")
	}
	if cfg.MaxMemoryMB <= 0 {
		cfg = NewDefaultGoRuntimeConfig()
	}
	if cfg.BaseWorkDir == "" {
		cfg.BaseWorkDir = "/var/lib/functionfly/go-instances"
	}
	if err := os.MkdirAll(cfg.BaseWorkDir, 0o755); err != nil {
		return nil, fmt.Errorf("go runtime: mkdir base workdir: %w", err)
	}
	workDir, err := os.MkdirTemp(cfg.BaseWorkDir, "instance-")
	if err != nil {
		return nil, fmt.Errorf("go runtime: mkdir workdir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "functionfly"), 0o755); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, fmt.Errorf("go runtime: mkdir functionfly: %w", err)
	}

	if handler == nil {
		handler = NewDefaultHostHandler(nil)
	}
	scanner := NewRuntimeScanner()

	engineConfig := wasmtime.NewConfig()
	if cfg.EnableFuel {
		engineConfig.SetConsumeFuel(true)
	}
	engine := wasmtime.NewEngineWithConfig(engineConfig)

	module, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		_ = os.RemoveAll(workDir)
		return nil, fmt.Errorf("go runtime: compile wasm: %w", err)
	}

	return &GoRuntime{
		wasmBytes: wasmBytes,
		cfg:       cfg,
		handler:   handler,
		scanner:   scanner,
		workDir:   workDir,
		engine:    engine,
		module:    module,
		createdAt: time.Now(),
	}, nil
}

// WorkDir returns the per-instance working directory.
func (r *GoRuntime) WorkDir() string { return r.workDir }

// CreatedAt returns the runtime's instantiation time.
func (r *GoRuntime) CreatedAt() time.Time { return r.createdAt }

// Healthy reports whether the runtime can still execute.
func (r *GoRuntime) Healthy(_ context.Context) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.module != nil
}

// LoadCode is a no-op; provided for parity with PythonRuntime.
func (r *GoRuntime) LoadCode(_ []byte) error { return nil }

// Init is a no-op; instantiation replaces Init.
func (r *GoRuntime) Init() error { return nil }

const goInputPath = "/functionfly/input.json"

// ExecuteWithContext runs the user function with input, honoring
// cfg.Timeout via context cancellation and cfg.MaxInstructions via
// wasmtime fuel.
func (r *GoRuntime) ExecuteWithContext(ctx context.Context, input []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.module == nil {
		return nil, ErrGoRuntimeClosed
	}
	if len(input) > r.cfg.MaxInputBytes {
		return nil, fmt.Errorf("%w: %d > %d", ErrGoInputTooLarge, len(input), r.cfg.MaxInputBytes)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	inputPath := filepath.Join(r.workDir, "input.json")
	outputPath := filepath.Join(r.workDir, "output.json")
	_ = os.Remove(outputPath)

	if err := os.WriteFile(inputPath, input, 0o600); err != nil {
		return nil, fmt.Errorf("go runtime: write input: %w", err)
	}

	// Fresh store + linker for each call (wasmtime host functions
	// capture the Store they were defined against).
	store := wasmtime.NewStore(r.engine)
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.InheritStdout()
	wasiConfig.InheritStderr()
	if err := wasiConfig.PreopenDir(r.workDir, "/functionfly"); err != nil {
		return nil, fmt.Errorf("go runtime: preopen: %w", err)
	}
	store.SetWasi(wasiConfig)

	if r.cfg.EnableFuel {
		if err := store.SetFuel(r.cfg.MaxInstructions); err != nil {
			return nil, fmt.Errorf("go runtime: set fuel: %w", err)
		}
	}

	linker := wasmtime.NewLinker(r.engine)
	if err := linker.DefineWasi(); err != nil {
		return nil, fmt.Errorf("go runtime: define wasi: %w", err)
	}
	if err := defineHostFunctions(linker, store, r.handler); err != nil {
		return nil, fmt.Errorf("go runtime: define host functions: %w", err)
	}

	instance, err := linker.Instantiate(store, r.module)
	if err != nil {
		return nil, fmt.Errorf("go runtime: instantiate: %w", err)
	}
	r.store = store

	start := instance.GetExport(store, "_start").Func()
	if start == nil {
		return nil, fmt.Errorf("go runtime: module does not export _start (not a WASI module?)")
	}

	resultCh := make(chan error, 1)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				resultCh <- fmt.Errorf("%w: %v", ErrGoPanic, rec)
			}
		}()
		_, err := start.Call(store)
		resultCh <- err
	}()

	select {
	case err := <-resultCh:
		// Go-WASI programs trap with "Exited with i32 exit status N"
		// after proc_exit. Status 0 is normal completion; non-zero is
		// a runtime error.
		if err != nil && !isProcExit(err) {
			if isFuelExhaustion(err) {
				return nil, fmt.Errorf("%w: %v", ErrGoFuelExhausted, err)
			}
			return nil, fmt.Errorf("%w: %v", ErrGoExecutionFailed, err)
		}
		if isNonZeroExit(err) {
			return nil, fmt.Errorf("%w: %v", ErrGoExecutionFailed, err)
		}
	case <-execCtx.Done():
		return nil, fmt.Errorf("%w: %v", ErrGoTimeout, execCtx.Err())
	}

	r.execCount++

	data, err := os.ReadFile(outputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrGoNoOutput
		}
		return nil, fmt.Errorf("go runtime: read output: %w", err)
	}
	if len(data) > r.cfg.MaxOutputBytes {
		return nil, fmt.Errorf("%w: %d > %d", ErrGoOutputTooLarge, len(data), r.cfg.MaxOutputBytes)
	}

	envelope := struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
		Error   string          `json:"error"`
	}{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoBadEnvelope, err)
	}
	if !envelope.Success {
		return data, fmt.Errorf("%w: %s", ErrGoHandlerError, envelope.Error)
	}
	return data, nil
}

// Execute is the no-context convenience wrapper.
func (r *GoRuntime) Execute(input []byte) ([]byte, error) {
	return r.ExecuteWithContext(context.Background(), input)
}

// Stats returns runtime statistics.
func (r *GoRuntime) Stats() GoRuntimeStats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return GoRuntimeStats{
		WorkDir:   r.workDir,
		ExecCount: r.execCount,
		CreatedAt: r.createdAt,
		Uptime:    time.Since(r.createdAt),
	}
}

// Close releases the per-instance working directory.
func (r *GoRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.module = nil
	r.store = nil
	if r.workDir != "" {
		_ = os.RemoveAll(r.workDir)
		r.workDir = ""
	}
	return nil
}

func isFuelExhaustion(err error) bool {
	if err == nil {
		return false
	}
	return hasAnySubstring(err.Error(), "all fuel consumed", "fuel exhausted")
}

func isProcExit(err error) bool {
	if err == nil {
		return false
	}
	return hasAnySubstring(err.Error(), "Exited with i32 exit status")
}

func isNonZeroExit(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !hasAnySubstring(msg, "Exited with i32 exit status") {
		return false
	}
	last := msg[len(msg)-1]
	if last < '0' || last > '9' {
		return true
	}
	return last != '0'
}

var (
	_ GoRuntimeIfc = (*GoRuntime)(nil)
	_ io.Closer    = (*GoRuntime)(nil)
)
