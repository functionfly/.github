//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly.
// This file implements Swift WASM execution via WASI stdin/stdout.
package wasm

import (
	"context"
	"fmt"
	"os"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// SwiftWASIRuntime executes Swift WASM modules via WASI.
// Swift functions read input from stdin and write output to stdout,
// unlike Python/JS runtimes which use exported execute() functions.
type SwiftWASIRuntime struct {
	engine  *wasmtime.Engine
	module  *wasmtime.Module
	config  *WASMSecurityConfig
	handler HostFunctionHandler
	debug   bool
}

// NewSwiftWASIRuntime creates a runtime from pre-compiled WASM bytes.
func NewSwiftWASIRuntime(wasmBinary []byte, handler HostFunctionHandler) (*SwiftWASIRuntime, error) {
	return NewSwiftWASIRuntimeWithConfig(wasmBinary, handler, NewDefaultSecurityConfig())
}

// NewSwiftWASIRuntimeWithConfig creates a runtime with custom security config.
func NewSwiftWASIRuntimeWithConfig(wasmBinary []byte, handler HostFunctionHandler, config *WASMSecurityConfig) (*SwiftWASIRuntime, error) {
	if config == nil {
		config = NewDefaultSecurityConfig()
	}

	engineConfig := wasmtime.NewConfig()
	if config.EnableDeterministic {
		engineConfig.SetConsumeFuel(true)
	}

	engine := wasmtime.NewEngineWithConfig(engineConfig)

	module, err := wasmtime.NewModule(engine, wasmBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to compile Swift WASM module: %w", err)
	}

	return &SwiftWASIRuntime{
		engine:  engine,
		module:  module,
		config:  config,
		handler: handler,
	}, nil
}

// Execute runs the Swift function: WASI _start → reads stdin → writes stdout.
func (r *SwiftWASIRuntime) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if r.config.MaxInputSize > 0 && uint32(len(input)) > r.config.MaxInputSize {
		return nil, fmt.Errorf("input size exceeds maximum: %d > %d", len(input), r.config.MaxInputSize)
	}

	execTimeout := r.config.MaxExecutionTime
	if execTimeout <= 0 {
		execTimeout = DefaultMaxExecutionTime
	}

	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	type result struct {
		output []byte
		err    error
	}
	ch := make(chan result, 1)

	go func() {
		out, err := r.executeWASI(ctx, input)
		ch <- result{out, err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("Swift execution timed out after %v", execTimeout)
	case res := <-ch:
		return res.output, res.err
	}
}

// executeWASI performs the actual WASI execution with temp files for stdin/stdout.
func (r *SwiftWASIRuntime) executeWASI(ctx context.Context, input []byte) ([]byte, error) {
	// Write input to temp file for WASI stdin
	stdinFile, err := os.CreateTemp("", "swift-stdin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin temp: %w", err)
	}
	defer os.Remove(stdinFile.Name())

	if _, err := stdinFile.Write(input); err != nil {
		stdinFile.Close()
		return nil, fmt.Errorf("failed to write stdin: %w", err)
	}
	stdinFile.Close()

	// Create temp file for WASI stdout
	stdoutFile, err := os.CreateTemp("", "swift-stdout-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout temp: %w", err)
	}
	defer os.Remove(stdoutFile.Name())
	stdoutFile.Close()

	// Create temp file for WASI stderr
	stderrFile, err := os.CreateTemp("", "swift-stderr-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr temp: %w", err)
	}
	defer os.Remove(stderrFile.Name())
	stderrFile.Close()

	wasiConfig := wasmtime.NewWasiConfig()
	if err := wasiConfig.SetStdinFile(stdinFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to set stdin: %w", err)
	}
	if err := wasiConfig.SetStdoutFile(stdoutFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to set stdout: %w", err)
	}
	if err := wasiConfig.SetStderrFile(stderrFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to set stderr: %w", err)
	}

	store := wasmtime.NewStore(r.engine)
	store.SetWasi(wasiConfig)

	if r.config.EnableDeterministic {
		store.SetFuel(r.config.MaxInstructions)
	}

	linker := wasmtime.NewLinker(r.engine)
	if err := linker.DefineWasi(); err != nil {
		return nil, fmt.Errorf("failed to define WASI: %w", err)
	}

	if r.handler != nil {
		if err := defineHostFunctions(linker, store, r.handler); err != nil {
			return nil, fmt.Errorf("failed to define host functions: %w", err)
		}
	}

	instance, err := linker.Instantiate(store, r.module)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Swift WASM module: %w", err)
	}

	startFunc := instance.GetExport(store, "_start")
	if startFunc == nil || startFunc.Func() == nil {
		return nil, fmt.Errorf("Swift WASM module does not export _start")
	}

	_, err = startFunc.Func().Call(store)
	if err != nil {
		// Read partial stdout before returning error
		stderrBytes, _ := os.ReadFile(stderrFile.Name())
		stdoutBytes, _ := os.ReadFile(stdoutFile.Name())
		if len(stdoutBytes) > 0 {
			return stdoutBytes, nil
		}
		return nil, fmt.Errorf("Swift _start failed: %w (stderr: %s)", err, string(stderrBytes))
	}

	output, err := os.ReadFile(stdoutFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	if r.config.MaxOutputSize > 0 && uint32(len(output)) > r.config.MaxOutputSize {
		return nil, fmt.Errorf("output size exceeds maximum: %d > %d", len(output), r.config.MaxOutputSize)
	}

	return output, nil
}

// Close releases resources.
func (r *SwiftWASIRuntime) Close() error {
	return nil
}

// SetDebug enables debug logging.
func (r *SwiftWASIRuntime) SetDebug(debug bool) {
	r.debug = debug
}
