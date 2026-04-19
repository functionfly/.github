//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains JavaScript/TypeScript-specific security validation and Javy runtime support
package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// JavaScriptSecurityConfig extends WASMSecurityConfig with JS-specific security
type JavaScriptSecurityConfig struct {
	*WASMSecurityConfig

	// MaxSourceSize is the maximum source code size in bytes (default: 1MB)
	MaxSourceSize uint32 `json:"max_source_size"`

	// MaxCompilationTime is the maximum time for Javy compilation (default: 60s)
	MaxCompilationTime time.Duration `json:"max_compilation_time"`

	// EnableEvalBlock blocks use of eval(), Function(), etc. (default: true)
	EnableEvalBlock bool `json:"enable_eval_block"`

	// BlockedPatterns are regex patterns to block in source code
	BlockedPatterns []string `json:"blocked_patterns"`

	// RequireFunctionExport enforces that code must export a function (default: true)
	RequireFunctionExport bool `json:"require_function_export"`
}

// DefaultJavaScriptSecurityConfig returns a default JS security config
func DefaultJavaScriptSecurityConfig() *JavaScriptSecurityConfig {
	return &JavaScriptSecurityConfig{
		WASMSecurityConfig:    NewDefaultSecurityConfig(),
		MaxSourceSize:         1024 * 1024, // 1MB
		MaxCompilationTime:    60 * time.Second,
		EnableEvalBlock:       true,
		BlockedPatterns:       []string{},
		RequireFunctionExport: true,
	}
}

// ValidateSourceCode validates JavaScript/TypeScript source code for security issues
func ValidateSourceCode(sourceCode []byte, config *JavaScriptSecurityConfig) error {
	if config == nil {
		config = DefaultJavaScriptSecurityConfig()
	}

	// Check source size
	if len(sourceCode) > int(config.MaxSourceSize) {
		return fmt.Errorf("source code size %d exceeds maximum %d bytes", len(sourceCode), config.MaxSourceSize)
	}

	source := string(sourceCode)

	// Check for blocked patterns
	for _, pattern := range config.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}
		if re.MatchString(source) {
			return fmt.Errorf("source code contains blocked pattern: %s", pattern)
		}
	}

	// Block dangerous patterns if eval blocking is enabled
	if config.EnableEvalBlock {
		dangerousPatterns := []string{
			`\beval\s*\(`,                    // eval()
			`\bFunction\s*\(`,                 // Function()
			`\bsetTimeout\s*\(\s*["']`,        // setTimeout with string (indirect eval)
			`\bsetInterval\s*\(\s*["']`,       // setInterval with string (indirect eval)
			`\bexecScript\s*\(`,               // IE-specific
			`\bnew\s+Function\s*\(`,           // new Function()
			`\bimport\s*\(\s*["']`,            // dynamic import (potential code injection)
			`__proto__`,                       // prototype pollution
			`constructor`,                     // constructor access
			`__defineGetter__`,                // defineGetter
			`__defineSetter__`,                // defineSetter
		}

		for _, pattern := range dangerousPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(source) {
				return fmt.Errorf("source code contains potentially dangerous pattern: %s", pattern)
			}
		}
	}

	// Check for function export if required
	if config.RequireFunctionExport {
		hasExport := strings.Contains(source, "export") ||
			strings.Contains(source, "module.exports") ||
			strings.Contains(source, "exports.") ||
			strings.Contains(source, "export default") ||
			strings.Contains(source, "export function") ||
			strings.Contains(source, "export const") ||
			strings.Contains(source, "export var") ||
			strings.Contains(source, "export let")

		if !hasExport {
			return fmt.Errorf("source code must export a function (use 'export default' or 'module.exports')")
		}
	}

	return nil
}

// SourceCodeHash computes a SHA-256 hash of the source code for integrity verification
func SourceCodeHash(sourceCode []byte) string {
	hash := sha256.Sum256(sourceCode)
	return hex.EncodeToString(hash[:])
}

// JavyRuntimeStats holds statistics for the Javy runtime
type JavyRuntimeStats struct {
	MemoryUsed    uint64
	MemoryTotal   uint64
	InstanceCount int
	CompileCount  int64
}

// JavyRuntime manages a Javy (QuickJS-based) WASM runtime for JavaScript execution
type JavyRuntime struct {
	engine   *wasmtime.Engine
	module   *wasmtime.Module
	store    *wasmtime.Store
	instance *wasmtime.Instance
	memory   *wasmtime.Memory
	handler  HostFunctionHandler
	debug    bool
	config   *JavaScriptSecurityConfig
	// Metadata
	sourceHash string
	stats     *JavyRuntimeStats
}

// NewJavyRuntime creates a new Javy runtime instance
func NewJavyRuntime(wasmBinary []byte, stdout, stderr io.Writer, handler HostFunctionHandler) (*JavyRuntime, error) {
	return NewJavyRuntimeWithConfig(wasmBinary, stdout, stderr, handler, nil)
}

// NewJavyRuntimeWithConfig creates a new Javy runtime instance with security config
func NewJavyRuntimeWithConfig(wasmBinary []byte, stdout, stderr io.Writer, handler HostFunctionHandler, config *JavaScriptSecurityConfig) (*JavyRuntime, error) {
	if config == nil {
		config = DefaultJavaScriptSecurityConfig()
	}

	// Extract WASM binary (in case it's bundled with metadata)
	actualWasmBinary, err := extractWasmBinary(wasmBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to extract WASM binary: %w", err)
	}

	// Validate WASM binary size
	if len(actualWasmBinary) > int(config.MaxSourceSize) {
		return nil, fmt.Errorf("WASM binary size %d exceeds maximum %d bytes", len(actualWasmBinary), config.MaxSourceSize)
	}

	// Create engine with WASI support
	engineConfig := wasmtime.NewConfig()

	// Enable fuel/energy for instruction counting if deterministic mode
	if config.EnableDeterministic {
		engineConfig.SetConsumeFuel(true)
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
	if err := defineJavyHostFunctions(linker, store, handler); err != nil {
		return nil, fmt.Errorf("failed to define Javy host functions: %w", err)
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

	return &JavyRuntime{
		engine:   engine,
		module:   module,
		store:    store,
		instance: instance,
		memory:   memory,
		handler:  handler,
		config:   config,
		stats: &JavyRuntimeStats{
			InstanceCount: 1,
		},
	}, nil
}

// SetSourceHash sets the source code hash for integrity verification
func (r *JavyRuntime) SetSourceHash(hash string) {
	r.sourceHash = hash
}

// GetSourceHash returns the source code hash
func (r *JavyRuntime) GetSourceHash() string {
	return r.sourceHash
}

// Init initializes the Javy runtime
func (r *JavyRuntime) Init() error {
	// Try to call init function if exported
	if initFunc := r.instance.GetExport(r.store, "init"); initFunc.Func() != nil {
		result, err := initFunc.Func().Call(r.store)
		if err != nil {
			return fmt.Errorf("init call failed: %w", err)
		}
		// Check result (0 = success)
		if resultI32, ok := result.(int32); ok && resultI32 != 0 {
			return fmt.Errorf("init returned error code: %d", resultI32)
		}
	}

	return nil
}

// Execute runs the JavaScript function with the given input
func (r *JavyRuntime) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return r.ExecuteWithTimeout(ctx, input, r.config.MaxExecutionTime)
}

// ExecuteWithTimeout runs the JavaScript function with a custom timeout
func (r *JavyRuntime) ExecuteWithTimeout(ctx context.Context, input []byte, timeout time.Duration) ([]byte, error) {
	// Validate input size
	if !r.config.ValidateInputSize(uint32(len(input))) {
		return nil, fmt.Errorf("input size exceeds maximum allowed: %d > %d bytes", len(input), r.config.MaxInputSize)
	}

	// Set up execution timeout
	if timeout <= 0 {
		timeout = r.config.MaxExecutionTime
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
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
		return nil, fmt.Errorf("execution timed out after %v", timeout)
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("execution failed: %w", result.err)
		}
		return result.output, nil
	}
}

// executeInternal performs the actual execution
func (r *JavyRuntime) executeInternal(input []byte) ([]byte, error) {
	// Try to find execute function
	executeFunc := r.instance.GetExport(r.store, "execute")
	if executeFunc.Func() == nil {
		// Fall back to _start for non-WASI modules
		executeFunc = r.instance.GetExport(r.store, "_start")
		if executeFunc.Func() == nil {
			return nil, fmt.Errorf("module does not export execute or _start function")
		}
	}

	// Allocate memory for input
	inputPtr, err := r.allocate(len(input))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate input memory: %w", err)
	}

	// Copy input to WASM memory
	memoryData := r.memory.UnsafeData(r.store)
	inputPtrSafe := int(inputPtr)
	if inputPtrSafe+len(input) > len(memoryData) {
		return nil, fmt.Errorf("input write would exceed memory bounds")
	}
	copy(memoryData[inputPtrSafe:inputPtrSafe+len(input)], input)

	// Call execute function
	result, err := executeFunc.Func().Call(r.store, inputPtr, int32(len(input)))
	if err != nil {
		return nil, fmt.Errorf("execute call failed: %w", err)
	}

	// Extract result pointer
	resultPtr, ok := result.(int32)
	if !ok {
		return nil, fmt.Errorf("execute returned invalid result type")
	}

	if resultPtr < 0 {
		return nil, fmt.Errorf("execution returned error: %d", resultPtr)
	}

	// Read output from memory
	maxOut := int(r.config.MaxOutputSize)
	if maxOut <= 0 {
		maxOut = 1024 * 1024 // 1MB default
	}

	memLen := len(memoryData)
	if int(resultPtr) >= memLen {
		return nil, fmt.Errorf("execute returned invalid pointer: %d (memory size %d)", resultPtr, memLen)
	}

	// Try length-prefixed format first (Javy/QuickJS format)
	payloadStart := resultPtr
	var outputLen int
	if resultPtr+4 <= int32(memLen) {
		prefixLen := readUint32LE(memoryData[resultPtr : resultPtr+4])
		if prefixLen <= uint32(maxOut) && int(resultPtr)+4+int(prefixLen) <= memLen {
			outputLen = int(prefixLen)
			payloadStart = resultPtr + 4
		}
	}

	// Fallback: null-terminated string
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

	return output, nil
}

// allocate allocates memory in the WASM module with security checks
func (r *JavyRuntime) allocate(size int) (int32, error) {
	if size < 0 {
		return 0, fmt.Errorf("invalid allocation size: negative")
	}

	// Security: Check against max memory limit
	currentMem := r.GetMemoryUsage()
	if uint64(size)+currentMem > uint64(r.config.MaxMemory) {
		return 0, fmt.Errorf("allocation would exceed memory limit: current=%d requested=%d max=%d",
			currentMem, size, r.config.MaxMemory)
	}

	// Try alloc function first
	allocFunc := r.instance.GetExport(r.store, "alloc")
	if allocFunc.Func() != nil {
		result, err := allocFunc.Func().Call(r.store, int32(size))
		if err != nil {
			return 0, fmt.Errorf("alloc call failed: %w", err)
		}
		return result.(int32), nil
	}

	// Fallback: return 0 (module should export alloc)
	return 0, nil
}

// GetMemoryUsage returns the current memory usage in bytes
func (r *JavyRuntime) GetMemoryUsage() uint64 {
	if r.memory == nil {
		return 0
	}
	return uint64(r.memory.Size(r.store)) * 65536
}

// Stats returns runtime statistics
func (r *JavyRuntime) Stats() (*JavyRuntimeStats, error) {
	return &JavyRuntimeStats{
		MemoryUsed:    r.GetMemoryUsage(),
		MemoryTotal:   uint64(r.config.MaxMemory),
		InstanceCount: r.stats.InstanceCount,
		CompileCount:  r.stats.CompileCount,
	}, nil
}

// Close releases resources
func (r *JavyRuntime) Close() error {
	r.instance = nil
	r.store = nil
	r.module = nil
	r.engine = nil
	return nil
}

// defineJavyHostFunctions registers Javy-specific host functions
func defineJavyHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store, handler HostFunctionHandler) error {
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

	// time_ms: (result f64) - milliseconds since Unix epoch
	if err := linker.DefineFunc(store, "functionfly", "time_ms",
		func(caller *wasmtime.Caller) float64 {
			return float64(time.Now().UnixMilli())
		}); err != nil {
		return fmt.Errorf("failed to define time_ms function: %w", err)
	}

	return nil
}

// readUint32LE reads a little-endian uint32 from a byte slice
func readUint32LE(data []byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

// extractWasmBinary extracts the WASM binary from potentially bundled data
func extractWasmBinary(wasmBinary []byte) ([]byte, error) {
	// Simple extraction - if it starts with \0asm magic bytes, it's a raw WASM module
	if len(wasmBinary) >= 4 && string(wasmBinary[0:4]) == "\x00asm" {
		return wasmBinary, nil
	}

	// Handle metadata-wrapped FunctionFly bundles:
	// "FFWB" + 4-byte big-endian metadata length + metadata JSON + raw WASM
	const header = "FFWB"
	const metadataLengthSize = 4

	if len(wasmBinary) >= len(header)+metadataLengthSize && string(wasmBinary[:len(header)]) == header {
		metadataLenOffset := len(header)
		metadataLen := int(wasmBinary[metadataLenOffset])<<24 |
			int(wasmBinary[metadataLenOffset+1])<<16 |
			int(wasmBinary[metadataLenOffset+2])<<8 |
			int(wasmBinary[metadataLenOffset+3])

		if metadataLen < 0 {
			return nil, fmt.Errorf("invalid metadata length: %d", metadataLen)
		}

		wasmStart := len(header) + metadataLengthSize + metadataLen
		if wasmStart > len(wasmBinary) {
			return nil, fmt.Errorf("invalid bundled binary: metadata length exceeds bounds")
		}

		actualWasm := wasmBinary[wasmStart:]
		if len(actualWasm) < 4 || string(actualWasm[:4]) != "\x00asm" {
			return nil, fmt.Errorf("invalid bundled binary: extracted payload is not a valid WASM module")
		}

		return actualWasm, nil
	}

	return wasmBinary, nil
}
