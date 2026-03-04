//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
package wasm

import (
	"fmt"
	"io"
	"log"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// PythonRuntime manages a WASI-compatible Python runtime
type PythonRuntime struct {
	engine   *wasmtime.Engine
	module   *wasmtime.Module
	store    *wasmtime.Store
	instance *wasmtime.Instance
	memory   *wasmtime.Memory
	handler  HostFunctionHandler
	debug    bool
}

// NewPythonRuntime creates a new Python runtime instance
func NewPythonRuntime(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error) {
	return NewPythonRuntimeWithDebug(wasmPath, stdout, stderr, handler, false)
}

// debugf logs debug information if debug mode is enabled
func (r *PythonRuntime) debugf(format string, args ...interface{}) {
	if r.debug {
		log.Printf("[WASM Debug] "+format, args...)
	}
}

// NewPythonRuntimeWithDebug creates a new Python runtime instance with debug mode
func NewPythonRuntimeWithDebug(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, debug bool) (*PythonRuntime, error) {
	// Create engine with WASI support
	engine := wasmtime.NewEngine()

	// Load precompiled WASM module (WAT should be pre-compiled to WASM)
	module, err := wasmtime.NewModuleFromFile(engine, wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load WASM module: %w", err)
	}

	// Create store with WASI configuration
	store := wasmtime.NewStore(engine)

	// Configure WASI
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.InheritStdout()
	wasiConfig.InheritStderr()
	wasiConfig.SetEnv([]string{"PYTHONPATH"}, []string{"/lib"})
	store.SetWasi(wasiConfig)

	// Create linker with WASI and host functions
	linker := wasmtime.NewLinker(engine)
	if err := linker.DefineWasi(); err != nil {
		return nil, fmt.Errorf("failed to define WASI: %w", err)
	}

	// Define FunctionFly host functions
	if err := defineHostFunctions(linker, store, handler); err != nil {
		return nil, fmt.Errorf("failed to define host functions: %w", err)
	}

	// Define MicroPython env.* functions for micropython.wasm
	if err := DefineMicropythonHostFunctions(linker, store); err != nil {
		return nil, fmt.Errorf("failed to define micropython host functions: %w", err)
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

	return &PythonRuntime{
		engine:   engine,
		module:   module,
		store:    store,
		instance: instance,
		memory:   memory,
		handler:  handler,
		debug:    debug,
	}, nil
}

// Init initializes the Python runtime
func (r *PythonRuntime) Init() error {
	r.debugf("Initializing Python runtime")

	initFunc := r.instance.GetExport(r.store, "init").Func()
	if initFunc == nil {
		return fmt.Errorf("module does not export init function")
	}

	result, err := initFunc.Call(r.store)
	if err != nil {
		r.debugf("Init call failed: %v", err)
		return fmt.Errorf("init call failed: %w", err)
	}

	// Check result (0 = success, non-zero = error)
	if resultI32, ok := result.(int32); ok && resultI32 != 0 {
		r.debugf("Init returned error code: %d", resultI32)
		return fmt.Errorf("init returned error code: %d", resultI32)
	}

	r.debugf("Python runtime initialized successfully")
	return nil
}

// LoadCode loads Python source code into the runtime
func (r *PythonRuntime) LoadCode(code string) error {
	loadCodeFunc := r.instance.GetExport(r.store, "load_code").Func()
	if loadCodeFunc == nil {
		return fmt.Errorf("module does not export load_code function")
	}

	// Allocate memory for the code
	codePtr, err := r.allocate(len(code))
	if err != nil {
		return fmt.Errorf("failed to allocate memory for code: %w", err)
	}

	// Write code to memory
	if err := r.writeMemory(codePtr, []byte(code)); err != nil {
		return fmt.Errorf("failed to write code to memory: %w", err)
	}

	// Call load_code
	result, err := loadCodeFunc.Call(r.store, codePtr, len(code))
	if err != nil {
		return fmt.Errorf("load_code call failed: %w", err)
	}

	// Check result (0 = success, non-zero = error)
	if resultI32, ok := result.(int32); ok && resultI32 != 0 {
		return fmt.Errorf("load_code returned error code: %d", resultI32)
	}

	return nil
}

// Execute runs the loaded code with the given input and returns output
func (r *PythonRuntime) Execute(input []byte) ([]byte, error) {
	r.debugf("Executing code with input: %s", string(input))

	executeFunc := r.instance.GetExport(r.store, "execute").Func()
	if executeFunc == nil {
		return nil, fmt.Errorf("module does not export execute function")
	}

	// Allocate memory for input
	inputPtr, err := r.allocate(len(input))
	if err != nil {
		r.debugf("Failed to allocate memory for input: %v", err)
		return nil, fmt.Errorf("failed to allocate memory for input: %w", err)
	}
	r.debugf("Allocated input memory at: %d", inputPtr)

	// Write input to memory
	if err := r.writeMemory(inputPtr, input); err != nil {
		r.debugf("Failed to write input to memory: %v", err)
		return nil, fmt.Errorf("failed to write input to memory: %w", err)
	}

	// Call execute - WAT signature: (param i32 i32) -> i32
	result, err := executeFunc.Call(r.store, inputPtr, len(input))
	if err != nil {
		r.debugf("Execute call failed: %v", err)
		return nil, fmt.Errorf("execute call failed: %w", err)
	}

	// Result is a pointer: either to a null-terminated string, or to the embedder's
	// result structure { status i32, input_ref i32, result_data i32 } where result_data
	// at offset 8 is the pointer to the actual output string.
	resultPtr, ok := result.(int32)
	if !ok {
		r.debugf("Execute returned invalid result type: %T", result)
		return nil, fmt.Errorf("execute returned invalid result type")
	}

	r.debugf("Execute returned result pointer: %d", resultPtr)
	if resultPtr == 0 {
		return nil, fmt.Errorf("execute returned null pointer")
	}

	// Try to interpret as embedder result structure (12 bytes: status, input_ref, result_data)
	const maxOutputLen = 65536
	output, err := r.extractOutputFromResult(resultPtr, maxOutputLen)
	if err != nil {
		return nil, err
	}
	r.debugf("Execute completed successfully. Output length: %d", len(output))
	return output, nil
}

// extractOutputFromResult extracts the output string from the execute() return value.
// The embedded Python WASM embedder returns a pointer to a 12-byte structure:
//
//	offset 0: status (i32), offset 4: input_ref (i32), offset 8: result_data (i32)
//
// where result_data is a pointer to the actual JSON/output string.
// If the pointer does not look like that structure, it is treated as a direct string pointer.
func (r *PythonRuntime) extractOutputFromResult(resultPtr int32, maxLen int) ([]byte, error) {
	// Read 12 bytes to check for embedder result structure
	header, err := r.readMemory(resultPtr, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to read result header: %w", err)
	}
	// result_data at offset 8 (little-endian i32)
	if len(header) >= 12 {
		resultDataPtr := int32(header[8]) | int32(header[9])<<8 | int32(header[10])<<16 | int32(header[11])<<24
		status := int32(header[0]) | int32(header[1])<<8 | int32(header[2])<<16 | int32(header[3])<<24
		// Embedder uses status=1 for success and result_data as pointer to output
		if status == 1 && resultDataPtr != 0 && resultDataPtr != -1 {
			// Read string from result_data pointer
			out, err := r.readNullTerminatedString(resultDataPtr, maxLen)
			if err == nil && len(out) > 0 {
				return out, nil
			}
			// result_data might be a small value (e.g. 0, 1) used as result; format as JSON
			if resultDataPtr == 1 {
				return []byte("true"), nil
			}
			if resultDataPtr == 0 {
				return []byte("0"), nil
			}
		}
		// -1 is used for None
		if resultDataPtr == -1 {
			return []byte("null"), nil
		}
	}

	// Treat resultPtr as direct pointer to null-terminated string
	return r.readNullTerminatedString(resultPtr, maxLen)
}

// readNullTerminatedString reads a null-terminated string from WASM memory.
func (r *PythonRuntime) readNullTerminatedString(ptr int32, maxLen int) ([]byte, error) {
	output, err := r.readMemory(ptr, maxLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}
	for i, b := range output {
		if b == 0 {
			return output[:i], nil
		}
	}
	return output, nil
}

// allocate calls the WASM alloc function to allocate memory
func (r *PythonRuntime) allocate(size int) (int32, error) {
	allocFunc := r.instance.GetExport(r.store, "alloc").Func()
	if allocFunc == nil {
		return 0, fmt.Errorf("module does not export alloc function")
	}

	result, err := allocFunc.Call(r.store, int32(size))
	if err != nil {
		return 0, fmt.Errorf("alloc call failed: %w", err)
	}

	if ptr, ok := result.(int32); ok {
		return ptr, nil
	}

	return 0, fmt.Errorf("alloc returned invalid pointer")
}

// deallocate calls the WASM dealloc function to free memory
func (r *PythonRuntime) deallocate(ptr int32, size int) error {
	deallocFunc := r.instance.GetExport(r.store, "dealloc").Func()
	if deallocFunc == nil {
		return fmt.Errorf("module does not export dealloc function")
	}

	_, err := deallocFunc.Call(r.store, ptr)
	return err
}

// writeMemory writes data to the WASM memory at the specified pointer
func (r *PythonRuntime) writeMemory(ptr int32, data []byte) error {
	dataLen := len(data)
	memoryData := r.memory.UnsafeData(r.store)

	if ptr < 0 || int(ptr)+dataLen > len(memoryData) {
		return fmt.Errorf("memory write out of bounds")
	}

	copy(memoryData[ptr:], data)
	return nil
}

// readMemory reads data from WASM memory at the specified pointer
func (r *PythonRuntime) readMemory(ptr int32, size int) ([]byte, error) {
	memoryData := r.memory.UnsafeData(r.store)

	if ptr < 0 || int(ptr)+size > len(memoryData) {
		return nil, fmt.Errorf("memory read out of bounds")
	}

	data := make([]byte, size)
	copy(data, memoryData[ptr:ptr+int32(size)])
	return data, nil
}

// GetMemoryUsage returns the current memory usage in bytes
func (r *PythonRuntime) GetMemoryUsage() uint64 {
	if r.memory == nil {
		return 0
	}
	// Each page is 64KB (65536 bytes)
	return uint64(r.memory.Size(r.store)) * 65536
}

// Close cleans up the runtime resources
func (r *PythonRuntime) Close() error {
	// The store and engine will be garbage collected
	// Additional cleanup can be added here if needed
	return nil
}
