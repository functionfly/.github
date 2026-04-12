//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains the WASM3 IoT runtime implementation for constrained environments
package wasm

/*
#cgo CFLAGS: -I/home/micro/.local/include
#cgo LDFLAGS: -L/home/micro/.local/lib -lwasm3
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <wasm3.h>

// Helper bridge functions to simplify Go/C interface
static inline void* wasm3_new_env() {
    return (void*)m3_NewEnvironment();
}
static inline void wasm3_free_env(void* env) {
    m3_FreeEnvironment((IM3Environment)env);
}
static inline void* wasm3_new_runtime(void* env, uint32_t stackSize) {
    // Note: m3_NewRuntime takes 3 args: env, stack size, and optional userdata
    // We're using 64KB stack for IoT
    return (void*)m3_NewRuntime((IM3Environment)env, stackSize, NULL);
}
static inline void wasm3_free_runtime(void* rt) {
    m3_FreeRuntime((IM3Runtime)rt);
}
static inline int wasm3_parse_module(void* env, void** module, const uint8_t* wasm, uint32_t size) {
    M3Result r = m3_ParseModule((IM3Environment)env, (IM3Module*)module, wasm, size);
    return r != NULL; // Return non-zero on error
}
static inline void wasm3_free_module(void* mod) {
    m3_FreeModule((IM3Module)mod);
}
static inline int wasm3_load_module(void* rt, void* mod) {
    M3Result r = m3_LoadModule((IM3Runtime)rt, (IM3Module)mod);
    return r != NULL;
}
static inline int wasm3_find_function(void** func, void* rt, const char* name) {
    M3Result r = m3_FindFunction((IM3Function*)func, (IM3Runtime)rt, name);
    return r != NULL;
}
static inline int wasm3_call_argv(void* func, int argc, const char* argv[]) {
    M3Result r = m3_CallArgv((IM3Function)func, argc, argv);
    return r != NULL;
}
static inline uint8_t* wasm3_get_memory(void* rt, uint32_t* size, uint32_t idx) {
    return m3_GetMemory((IM3Runtime)rt, size, idx);
}
static inline const char* wasm3_get_error(void* rt) {
    M3ErrorInfo info;
    m3_GetErrorInfo((IM3Runtime)rt, &info);
    return info.message ? info.message : "unknown error";
}
*/
import "C"

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"unsafe"
)

// IoTConfig contains WASM3-specific IoT configuration
type IoTConfig struct {
	// TargetLatency is the target execution latency in milliseconds
	TargetLatency int `json:"target_latency"`

	// MaxMemoryKB is the maximum memory in KB for IoT devices (default: 16MB)
	MaxMemoryKB int `json:"max_memory_kb"`

	// EnableOTAUpdate enables OTA update capability
	EnableOTAUpdate bool `json:"enable_ota_update"`

	// BatteryOptimized enables battery optimization mode
	BatteryOptimized bool `json:"battery_optimized"`

	// MaxInstances is the maximum number of concurrent instances
	MaxInstances int `json:"max_instances"`

	// Default timeout for IoT execution
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// DefaultIoTConfig returns the default IoT configuration optimized for ~500ms latency
func DefaultIoTConfig() *IoTConfig {
	return &IoTConfig{
		TargetLatency:    500,       // 500ms target
		MaxMemoryKB:      16 * 1024, // 16MB
		EnableOTAUpdate:  false,
		BatteryOptimized: true,
		MaxInstances:     4,
		ExecutionTimeout: 450 * time.Millisecond, // Slightly less than target latency
	}
}

// WASM3IoTRuntime implements a lightweight WASM3 runtime for IoT devices
type WASM3IoTRuntime struct {
	mu       sync.RWMutex
	config   *IoTConfig
	pool     *IoTInstancePool
	closed   bool
	execTime time.Duration
}

// IoTInstance represents a single IoT WASM instance
type IoTInstance struct {
	ID       string
	Runtime  *WASM3IoTRuntime
	Memory   []byte
	Code     []byte
	State    map[string]interface{}
	LastUsed time.Time
}

// IoTInstancePool manages a pool of IoT instances
type IoTInstancePool struct {
	mu        sync.RWMutex
	instances chan *IoTInstance
	maxSize   int
	active    int
	config    *IoTConfig
}

// NewIoTInstancePool creates a new IoT instance pool
func NewIoTInstancePool(config *IoTConfig) *IoTInstancePool {
	if config == nil {
		config = DefaultIoTConfig()
	}
	pool := &IoTInstancePool{
		instances: make(chan *IoTInstance, config.MaxInstances),
		maxSize:   config.MaxInstances,
		config:    config,
	}

	// Pre-populate with idle instances
	for i := 0; i < config.MaxInstances; i++ {
		inst := &IoTInstance{
			ID:       fmt.Sprintf("iot-%d", i),
			Runtime:  nil,
			Memory:   make([]byte, config.MaxMemoryKB*1024),
			State:    make(map[string]interface{}),
			LastUsed: time.Now(),
		}
		pool.instances <- inst
	}

	return pool
}

// Get retrieves an instance from the pool
func (p *IoTInstancePool) Get(ctx context.Context) (*IoTInstance, error) {
	select {
	case inst := <-p.instances:
		inst.LastUsed = time.Now()
		return inst, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// If no instances available, wait with timeout
		select {
		case inst := <-p.instances:
			inst.LastUsed = time.Now()
			return inst, nil
		case <-time.After(100 * time.Millisecond):
			return nil, fmt.Errorf("timeout waiting for IoT instance")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Put returns an instance to the pool
func (p *IoTInstancePool) Put(inst *IoTInstance) {
	if inst == nil {
		return
	}
	inst.LastUsed = time.Now()
	select {
	case p.instances <- inst:
	default:
		// Pool is full, let it be garbage collected
	}
}

// Close closes the instance pool
func (p *IoTInstancePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.instances)
	return nil
}

// NewWASM3IoTRuntime creates a new WASM3 IoT runtime
func NewWASM3IoTRuntime(config *IoTConfig) (*WASM3IoTRuntime, error) {
	if config == nil {
		config = DefaultIoTConfig()
	}

	// Validate IoT config
	if config.MaxMemoryKB <= 0 {
		config.MaxMemoryKB = 16 * 1024
	}
	if config.MaxMemoryKB > 64*1024 { // Max 64MB
		config.MaxMemoryKB = 64 * 1024
	}
	if config.ExecutionTimeout <= 0 {
		config.ExecutionTimeout = 450 * time.Millisecond
	}

	runtime := &WASM3IoTRuntime{
		config: config,
		pool:   NewIoTInstancePool(config),
	}

	log.Printf("[WASM3 IoT] Created runtime with target latency: %dms, max memory: %dKB",
		config.TargetLatency, config.MaxMemoryKB)

	return runtime, nil
}

// Execute runs the IoT WASM code with the given input
func (r *WASM3IoTRuntime) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return r.ExecuteWithConfig(ctx, input, nil)
}

// ExecuteWithConfig runs the IoT WASM code with custom configuration
func (r *WASM3IoTRuntime) ExecuteWithConfig(ctx context.Context, input []byte, config interface{}) ([]byte, error) {
	startTime := time.Now()

	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.ExecutionTimeout)
		defer cancel()
	}

	// Get instance from pool
	inst, err := r.pool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IoT instance: %w", err)
	}
	defer r.pool.Put(inst)

	// Execute in a goroutine to allow cancellation
	resultChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	go func() {
		output, err := r.executeInstance(ctx, inst, input)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- output
		}
	}()

	select {
	case output := <-resultChan:
		r.execTime = time.Since(startTime)
		if r.execTime > time.Duration(r.config.TargetLatency)*time.Millisecond {
			log.Printf("[WASM3 IoT] Warning: execution time %v exceeds target %dms",
				r.execTime, r.config.TargetLatency)
		}
		return output, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("execution cancelled or timed out: %w", ctx.Err())
	}
}

// executeInstance executes the WASM code on a single instance using WASM3 C library
func (r *WASM3IoTRuntime) executeInstance(ctx context.Context, inst *IoTInstance, input []byte) ([]byte, error) {
	return r.executeWASM3CGO(inst, input)
}

// executeWASM3CGO executes WASM using the actual WASM3 C library via CGO
// This implementation uses C helper functions to bridge the Go/C interface
func (r *WASM3IoTRuntime) executeWASM3CGO(inst *IoTInstance, input []byte) ([]byte, error) {
	// Create WASM3 environment
	env := C.wasm3_new_env()
	if env == nil {
		return nil, fmt.Errorf("failed to create WASM3 environment")
	}
	defer C.wasm3_free_env(env)

	// Create runtime with configured memory (64KB stack for IoT)
	stackSize := C.uint32_t(64 * 1024)
	wasmRuntime := C.wasm3_new_runtime(env, stackSize)
	if wasmRuntime == nil {
		return nil, fmt.Errorf("failed to create WASM3 runtime")
	}
	defer C.wasm3_free_runtime(wasmRuntime)

	// Check if we have pre-loaded code
	if inst.Code == nil || len(inst.Code) < 8 {
		return nil, fmt.Errorf("no WASM module loaded for instance")
	}

	// Parse the WASM module
	var module unsafe.Pointer
	codePtr := (*C.uint8_t)(unsafe.Pointer(&inst.Code[0]))
	codeSize := C.uint32_t(len(inst.Code))

	if ret := C.wasm3_parse_module(env, &module, codePtr, codeSize); ret != 0 {
		errMsg := C.GoString(C.wasm3_get_error(wasmRuntime))
		return nil, fmt.Errorf("failed to parse WASM module: %s", errMsg)
	}
	defer C.wasm3_free_module(module)

	// Load module into runtime
	if ret := C.wasm3_load_module(wasmRuntime, module); ret != 0 {
		errMsg := C.GoString(C.wasm3_get_error(wasmRuntime))
		return nil, fmt.Errorf("failed to load WASM module: %s", errMsg)
	}

	// Find the main/handler function
	var funcPtr unsafe.Pointer
	funcName := C.CString("handler")
	defer C.free(unsafe.Pointer(funcName))

	if ret := C.wasm3_find_function(&funcPtr, wasmRuntime, funcName); ret != 0 {
		// Try "main" as fallback
		funcName = C.CString("main")
		defer C.free(unsafe.Pointer(funcName))
		if ret := C.wasm3_find_function(&funcPtr, wasmRuntime, funcName); ret != 0 {
			errMsg := C.GoString(C.wasm3_get_error(wasmRuntime))
			return nil, fmt.Errorf("failed to find handler function: %s", errMsg)
		}
	}

	// Copy input to WASM memory
	var memSize C.uint32_t
	memPtr := C.wasm3_get_memory(wasmRuntime, &memSize, 0)
	if memPtr == nil {
		return nil, fmt.Errorf("failed to get WASM memory")
	}

	// Write input length at offset 0, input data at offset 4
	if len(input)+4 > int(memSize) {
		return nil, fmt.Errorf("input too large for WASM memory: %d > %d", len(input), memSize-4)
	}

	// Write input size as little-endian 32-bit
	memBytes := (*[1 << 30]byte)(unsafe.Pointer(memPtr))
	memBytes[0] = byte(len(input))
	memBytes[1] = byte(len(input) >> 8)
	memBytes[2] = byte(len(input) >> 16)
	memBytes[3] = byte(len(input) >> 24)

	// Copy input
	copy(memBytes[4:], input)

	// Call the handler function with input offset and length
	inputOffset := C.CString("0")
	inputLen := C.CString(fmt.Sprintf("%d", len(input)))
	defer C.free(unsafe.Pointer(inputOffset))
	defer C.free(unsafe.Pointer(inputLen))

	// Build argv array for C call
	argv := make([]*C.char, 2)
	argv[0] = inputOffset
	argv[1] = inputLen
	if ret := C.wasm3_call_argv(funcPtr, 2, (**C.char)(unsafe.Pointer(&argv[0]))); ret != 0 {
		errMsg := C.GoString(C.wasm3_get_error(wasmRuntime))
		return nil, fmt.Errorf("WASM execution failed: %s", errMsg)
	}

	// Read output length from memory (written by handler at offset 0)
	outputLen := int(memBytes[0]) | int(memBytes[1])<<8 | int(memBytes[2])<<16 | int(memBytes[3])<<24
	if outputLen < 0 || outputLen > int(memSize)-4 {
		return nil, fmt.Errorf("invalid output length from WASM: %d", outputLen)
	}

	// Read output data
	output := make([]byte, outputLen)
	copy(output, memBytes[4:4+outputLen])

	return output, nil
}

// LoadModule loads a WASM module for IoT execution
func (r *WASM3IoTRuntime) LoadModule(moduleData []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("runtime is closed")
	}

	// Validate module header (WASM magic number)
	if len(moduleData) < 8 {
		return fmt.Errorf("module data too small")
	}
	if moduleData[0] != 0x00 || moduleData[1] != 0x61 || moduleData[2] != 0x73 || moduleData[3] != 0x6D {
		return fmt.Errorf("invalid WASM module header")
	}

	// Validate version
	version := uint32(moduleData[4]) | uint32(moduleData[5])<<8 | uint32(moduleData[6])<<16 | uint32(moduleData[7])<<24
	if version != 1 && version != 13 { // Version 1 (standard) or 13 (dynamic)
		return fmt.Errorf("unsupported WASM version: %d", version)
	}

	log.Printf("[WASM3 IoT] Loaded module: %d bytes", len(moduleData))
	return nil
}

// GetMetrics returns IoT-specific metrics
func (r *WASM3IoTRuntime) GetMetrics() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	if r.pool != nil {
		stats["max_instances"] = r.pool.maxSize
		stats["available_instances"] = len(r.pool.instances)
	}
	stats["avg_execution_time_ms"] = r.execTime.Milliseconds()
	stats["target_latency_ms"] = r.config.TargetLatency
	stats["max_memory_kb"] = r.config.MaxMemoryKB
	stats["battery_optimized"] = r.config.BatteryOptimized

	return stats
}

// Close closes the WASM3 IoT runtime
func (r *WASM3IoTRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if r.pool != nil {
		return r.pool.Close()
	}

	log.Printf("[WASM3 IoT] Runtime closed")
	return nil
}

// WASM3IoTProvider wraps WASM3IoTRuntime for RuntimeRouter
type WASM3IoTProvider struct {
	runtime *WASM3IoTRuntime
}

// NewWASM3IoTProvider creates a new WASM3 IoT provider
func NewWASM3IoTProvider(config *IoTConfig) (*WASM3IoTProvider, error) {
	runtime, err := NewWASM3IoTRuntime(config)
	if err != nil {
		return nil, err
	}
	return &WASM3IoTProvider{runtime: runtime}, nil
}

// Execute executes the function
func (p *WASM3IoTProvider) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return p.runtime.Execute(ctx, input)
}

// ExecuteWithConfig executes with custom configuration
func (p *WASM3IoTProvider) ExecuteWithConfig(ctx context.Context, input []byte, config interface{}) ([]byte, error) {
	return p.runtime.ExecuteWithConfig(ctx, input, config)
}

// Close closes the runtime
func (p *WASM3IoTProvider) Close() error {
	if p.runtime != nil {
		return p.runtime.Close()
	}
	return nil
}

// CreateWASM3IoTRuntime creates a WASM3 IoT runtime provider for the router
func CreateWASM3IoTRuntime(stdout, stderr io.Writer, handler HostFunctionHandler, config *IoTConfig) (RuntimeProvider, error) {
	return NewWASM3IoTProvider(config)
}
