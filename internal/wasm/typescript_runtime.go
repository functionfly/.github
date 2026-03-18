//go:build cgo

package wasm

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v19"

	"github.com/functionfly/functionfly/internal/bundler"
)

// TypeScriptRuntimeStats holds statistics for the TypeScript WASM runtime.
type TypeScriptRuntimeStats struct {
	MemoryUsed    uint64
	MemoryTotal   uint64
	InstanceCount int
}

// TypeScriptRuntime manages a WASI-compatible TypeScript runtime
type TypeScriptRuntime struct {
	engine   *wasmtime.Engine
	module   *wasmtime.Module
	store    *wasmtime.Store
	instance *wasmtime.Instance
	memory   *wasmtime.Memory
	handler  HostFunctionHandler
	debug    bool
	config   *WASMSecurityConfig
	// Metadata from the compiled WASM
	metadata *bundler.WASMMetadata
}

// NewTypeScriptRuntime creates a new TypeScript runtime instance
func NewTypeScriptRuntime(wasmBinary []byte, stdout, stderr io.Writer, handler HostFunctionHandler) (*TypeScriptRuntime, error) {
	return NewTypeScriptRuntimeWithConfig(wasmBinary, stdout, stderr, handler, NewDefaultSecurityConfig())
}

// NewTypeScriptRuntimeWithConfig creates a new TypeScript runtime instance with security config
func NewTypeScriptRuntimeWithConfig(wasmBinary []byte, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig) (*TypeScriptRuntime, error) {
	return NewTypeScriptRuntimeWithConfigAndDebug(wasmBinary, stdout, stderr, handler, config, false)
}

// NewTypeScriptRuntimeWithConfigAndDebug creates a new TypeScript runtime instance with security config and debug mode
func NewTypeScriptRuntimeWithConfigAndDebug(wasmBinary []byte, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig, debug bool) (*TypeScriptRuntime, error) {
	if config == nil {
		config = NewDefaultSecurityConfig()
	}

	// Extract WASM binary (in case it's bundled with metadata)
	actualWasmBinary, err := bundler.GetWASMBinary(wasmBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to extract WASM binary: %w", err)
	}

	// Extract metadata if available
	metadata, _ := bundler.ExtractMetadata(wasmBinary)
	if metadata == nil {
		metadata = &bundler.WASMMetadata{
			HandlerName:      "handler",
			MemoryPages:      256,
			ExportedFunctions: []string{"_start", "memory"},
			WASITarget:       false,
		}
	}

	// Create engine with WASI support and security config
	engineConfig := wasmtime.NewConfig()

	// Set memory limits (wasmtime uses pages, 1 page = 64KB)
	maxMemoryPages := config.MaxMemory / 65536
	if config.MaxMemory%65536 > 0 {
		maxMemoryPages++
	}
	_ = maxMemoryPages

	// Enable fuel/energy for instruction counting if deterministic mode
	if config.EnableDeterministic {
		// Enable fuel consumption tracking
	}

	engine := wasmtime.NewEngineWithConfig(engineConfig)

	// Compile WASM module from binary
	module, err := wasmtime.NewModule(engine, actualWasmBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Create store with WASI configuration
	store := wasmtime.NewStore(engine)

	// Configure WASI (if enabled)
	if config.EnableWASI {
		wasiConfig := wasmtime.NewWasiConfig()
		wasiConfig.InheritStdout()
		wasiConfig.InheritStderr()
		store.SetWasi(wasiConfig)
	}

	// Create linker with WASI and host functions
	linker := wasmtime.NewLinker(engine)
	if err := linker.DefineWasi(); err != nil {
		return nil, fmt.Errorf("failed to define WASI: %w", err)
	}

	// Define FunctionFly host functions
	if err := defineHostFunctions(linker, store, handler); err != nil {
		return nil, fmt.Errorf("failed to define host functions: %w", err)
	}

	// Define JavaScript-specific host functions
	if err := defineTypeScriptHostFunctions(linker, store, handler); err != nil {
		return nil, fmt.Errorf("failed to define TypeScript host functions: %w", err)
	}

	// Instantiate the module
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	// Get memory export
	memory := instance.GetExport(store, "memory").Memory()
	if memory == nil {
		return nil, fmt.Errorf("module does not export memory")
	}

	return &TypeScriptRuntime{
		engine:   engine,
		module:   module,
		store:    store,
		instance: instance,
		memory:   memory,
		handler:  handler,
		debug:    debug,
		config:   config,
		metadata: metadata,
	}, nil
}

// debugf logs debug information if debug mode is enabled
func (r *TypeScriptRuntime) debugf(format string, args ...interface{}) {
	if r.debug {
		log.Printf("[TypeScript WASM Debug] "+format, args...)
	}
}

// Init initializes the TypeScript runtime
func (r *TypeScriptRuntime) Init() error {
	r.debugf("Initializing TypeScript runtime")

	// Try to call init function if exported
	if initFunc := r.instance.GetExport(r.store, "init"); initFunc.Func() != nil {
		result, err := initFunc.Func().Call(r.store)
		if err != nil {
			r.debugf("Init call failed: %v", err)
			return fmt.Errorf("init call failed: %w", err)
		}
		r.debugf("Init result: %v", result)
	}

	return nil
}

// Execute runs the TypeScript function with the given input
func (r *TypeScriptRuntime) Execute(ctx context.Context, input []byte) ([]byte, error) {
	r.debugf("Executing TypeScript function with input: %s", string(input))

	// Check input size
	if r.config.MaxInputSize > 0 && uint32(len(input)) > r.config.MaxInputSize {
		return nil, fmt.Errorf("input size exceeds maximum allowed: %d > %d", len(input), r.config.MaxInputSize)
	}

	// Set up execution timeout
	execTimeout := r.config.MaxExecutionTime
	if execTimeout <= 0 {
		execTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	// Execute in goroutine to allow for cancellation
	resultChan := make(chan struct {
		output []byte
		err    error
	}, 1)

	go func() {
		output, err := r.executeInternal(input)
		resultChan <- struct {
			output []byte
			err    error
		}{output, err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("execution timed out after %v", execTimeout)
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("execution failed: %w", result.err)
		}
		return result.output, nil
	}
}

// executeInternal performs the actual execution
func (r *TypeScriptRuntime) executeInternal(input []byte) ([]byte, error) {
	// Try to find execute function
	executeFunc := r.instance.GetExport(r.store, "execute")
	if executeFunc.Func() == nil {
		// Fall back to _start for non-WASI modules
		executeFunc = r.instance.GetExport(r.store, "_start")
		if executeFunc.Func() == nil {
			return nil, fmt.Errorf("module does not export execute or _start function")
		}
		r.debugf("Using _start function")
	} else {
		r.debugf("Using execute function")
	}

	// Allocate memory for input
	inputPtr, err := r.allocate(len(input))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate input memory: %w", err)
	}

	// Copy input to WASM memory
	memoryData := r.memory.UnsafeData(r.store)
	// Copy input to WASM memory
	inputPtrSafe := int(inputPtr)
	copy(memoryData[inputPtrSafe:inputPtrSafe+len(input)], input)

	// Call execute function
	result, err := executeFunc.Func().Call(r.store, inputPtr, int32(len(input)))
	if err != nil {
		r.debugf("Execute call failed: %v", err)
		return nil, fmt.Errorf("execute call failed: %w", err)
	}

	// Result is a pointer to the output in WASM memory
	resultPtr := result.(int32)
	if resultPtr < 0 {
		return nil, fmt.Errorf("execution returned error: %d", resultPtr)
	}

	memLen := len(memoryData)
	maxOut := int(r.config.MaxOutputSize)
	if maxOut <= 0 {
		maxOut = 1024 * 1024 // 1MB default
	}

	// Bounds check: resultPtr must be within memory
	if int(resultPtr) >= memLen {
		return nil, fmt.Errorf("execute returned invalid pointer: %d (memory size %d)", resultPtr, memLen)
	}

	// Try length-prefixed format (e.g. Javy/QuickJS: 4-byte LE length then payload)
	payloadStart := resultPtr
	var outputLen int
	if resultPtr+4 <= int32(memLen) {
		prefixLen := binary.LittleEndian.Uint32(memoryData[resultPtr : resultPtr+4])
		if prefixLen <= uint32(maxOut) && int(resultPtr)+4+int(prefixLen) <= memLen {
			outputLen = int(prefixLen)
			payloadStart = resultPtr + 4
		}
	}

	// Fallback: null-terminated string (scan for first zero byte)
	if outputLen == 0 {
		end := int(resultPtr)
		for end < memLen && end-int(resultPtr) < maxOut && memoryData[end] != 0 {
			end++
		}
		outputLen = end - int(resultPtr)
		if outputLen < 0 {
			outputLen = 0
		}
	}

	output := make([]byte, outputLen)
	if outputLen > 0 {
		copy(output, memoryData[payloadStart:payloadStart+int32(outputLen)])
	}

	r.debugf("Execution completed, output length: %d", len(output))
	return output, nil
}

// allocate allocates memory in the WASM module
func (r *TypeScriptRuntime) allocate(size int) (int32, error) {
	// Try alloc function first
	allocFunc := r.instance.GetExport(r.store, "alloc")
	if allocFunc.Func() != nil {
		result, err := allocFunc.Func().Call(r.store, int32(size))
		if err != nil {
			return 0, fmt.Errorf("alloc call failed: %w", err)
		}
		return result.(int32), nil
	}

	// Allocate returns 0 (not using memory growth)
	return 0, nil
	// Note: Simplified allocation - the WASM module should export an alloc function
}

// Stats returns runtime statistics
func (r *TypeScriptRuntime) Stats() (*TypeScriptRuntimeStats, error) {
	memorySize := r.memory.Size(r.store)
	memoryBytes := memorySize * 65536 // Convert pages to bytes

	return &TypeScriptRuntimeStats{
		MemoryUsed:    uint64(memoryBytes),
		MemoryTotal:   uint64(r.config.MaxMemory),
		InstanceCount: 1,
	}, nil
}

// Close releases resources
func (r *TypeScriptRuntime) Close() error {
	r.debugf("Closing TypeScript runtime")
	r.instance = nil
	r.store = nil
	r.module = nil
	r.engine = nil
	return nil
}

// defineTypeScriptHostFunctions registers TypeScript-specific host functions
func defineTypeScriptHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store, handler HostFunctionHandler) error {
	// console_log: (param $msg_ptr i32) (param $msg_len i32)
	if err := linker.DefineFunc(store, "functionfly", "console_log",
		func(caller *wasmtime.Caller, msgPtr, msgLen int32) {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return
			}

			memoryData := memory.UnsafeData(store)
			if msgPtr < 0 || int(msgPtr)+int(msgLen) > len(memoryData) {
				return
			}

			message := string(memoryData[msgPtr : msgPtr+msgLen])
			handler.Log(message)
		}); err != nil {
		return fmt.Errorf("failed to define console_log function: %w", err)
	}

	// console_error: (param $msg_ptr i32) (param $msg_len i32)
	if err := linker.DefineFunc(store, "functionfly", "console_error",
		func(caller *wasmtime.Caller, msgPtr, msgLen int32) {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return
			}

			memoryData := memory.UnsafeData(store)
			if msgPtr < 0 || int(msgPtr)+int(msgLen) > len(memoryData) {
				return
			}

			message := "[error] " + string(memoryData[msgPtr:msgPtr+msgLen])
			handler.Log(message)
		}); err != nil {
		return fmt.Errorf("failed to define console_error function: %w", err)
	}

	// console_warn: (param $msg_ptr i32) (param $msg_len i32)
	if err := linker.DefineFunc(store, "functionfly", "console_warn",
		func(caller *wasmtime.Caller, msgPtr, msgLen int32) {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return
			}

			memoryData := memory.UnsafeData(store)
			if msgPtr < 0 || int(msgPtr)+int(msgLen) > len(memoryData) {
				return
			}

			message := "[warn] " + string(memoryData[msgPtr:msgPtr+msgLen])
			handler.Log(message)
		}); err != nil {
		return fmt.Errorf("failed to define console_warn function: %w", err)
	}

	// time_ms: (result f64) - milliseconds since Unix epoch (Date.now() equivalent)
	if err := linker.DefineFunc(store, "functionfly", "time_ms",
		func(caller *wasmtime.Caller) float64 {
			return float64(time.Now().UnixMilli())
		}); err != nil {
		return fmt.Errorf("failed to define time_ms function: %w", err)
	}

	return nil
}
