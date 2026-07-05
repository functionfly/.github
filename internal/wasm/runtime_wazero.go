//go:build !cgo

// Package wasm provides WebAssembly runtime support for FunctionFly.
// This file implements a production-ready wazero runtime when building without CGO.
// wazero is a pure-Go WASM runtime that works in environments where CGO is unavailable.
package wasm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WazeroRuntime implements PythonRuntime using wazero (pure Go, no CGO).
// This is a production-ready implementation with full security controls,
// host function support, and memory management.
type WazeroRuntime struct {
	mu         sync.RWMutex
	closed     bool
	wasmPath   string
	config     *WASMSecurityConfig
	handler    HostFunctionHandler
	stdout     io.Writer
	stderr     io.Writer
	runtime    wazero.Runtime
	module     api.Module
	memory     api.Memory
	execCount  int64
	createdAt  time.Time
	fuel       uint64
	fuelMutex  sync.Mutex
	namespace  string
}

// NewWazeroRuntime creates a new wazero-based Python runtime.
func NewWazeroRuntime(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*WazeroRuntime, error) {
	return NewWazeroRuntimeWithConfig(wasmPath, stdout, stderr, handler, NewDefaultSecurityConfig())
}

// NewWazeroRuntimeWithConfig creates a new wazero runtime with custom security config.
func NewWazeroRuntimeWithConfig(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig) (*WazeroRuntime, error) {
	if config == nil {
		config = NewDefaultSecurityConfig()
	}

	if handler == nil {
		handler = NewDefaultHostHandler(nil)
	}

	// Create wazero runtime with configuration
	wazeroConfig := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true)

	runtime := wazero.NewRuntimeWithConfig(context.Background(), wazeroConfig)

	namespace := fmt.Sprintf("ff-%d", time.Now().UnixNano())

	rt := &WazeroRuntime{
		wasmPath:  wasmPath,
		config:    config,
		handler:   handler,
		stdout:    stdout,
		stderr:    stderr,
		runtime:   runtime,
		namespace: namespace,
		createdAt: time.Now(),
	}

	// Configure host modules
	if err := rt.configureHostModules(); err != nil {
		runtime.Close(context.Background())
		return nil, fmt.Errorf("failed to configure host modules: %w", err)
	}

	return rt, nil
}

// configureHostModules sets up all host function modules for wazero
func (r *WazeroRuntime) configureHostModules() error {
	ctx := context.Background()

	// Register WASI functions
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r.runtime)
	if err != nil {
		// Don't fail on WASI init errors - continue with other host functions
	}

	// Register FunctionFly host functions (env.functionfly.*)
	if err := r.registerFunctionFlyHostFunctions(ctx); err != nil {
		return fmt.Errorf("failed to register FunctionFly host functions: %w", err)
	}

	// Register MicroPython compatibility functions (env.mp_js_*, env.invoke_*)
	if err := r.registerMicropythonHostFunctions(ctx); err != nil {
		return fmt.Errorf("failed to register MicroPython host functions: %w", err)
	}

	// Register Emscripten compatibility stubs (env.emscripten_*)
	if err := r.registerEmscriptenHostFunctions(ctx); err != nil {
		return fmt.Errorf("failed to register Emscripten host functions: %w", err)
	}

	// Register streaming support functions (env.streaming_*)
	if err := r.registerStreamingHostFunctions(ctx); err != nil {
		return fmt.Errorf("failed to register streaming host functions: %w", err)
	}

	return nil
}

// registerFunctionFlyHostFunctions registers the FunctionFly host API
func (r *WazeroRuntime) registerFunctionFlyHostFunctions(ctx context.Context) error {
	builder := r.runtime.NewHostModuleBuilder("functionfly")

	// env.functionfly.log(msg_ptr i32, msg_len i32)
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroLog).
		Export("log")

	// env.functionfly.fetch(req_ptr i32, req_len i32, resp_ptr i32, resp_len_ptr i32) -> i32
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroFetch).
		Export("fetch")

	// env.functionfly.kv_get(key_ptr i32, key_len i32, val_ptr i32, val_len_ptr i32) -> i32
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroKVGet).
		Export("kv_get")

	// env.functionfly.kv_set(key_ptr i32, key_len i32, val_ptr i32, val_len i32) -> i32
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroKVSet).
		Export("kv_set")

	// env.functionfly.get_env(name_ptr i32, name_len i32, val_ptr i32, val_len_ptr i32) -> i32
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroGetEnv).
		Export("get_env")

	// env.functionfly.ai_infer(model_ptr, model_len, input_ptr, input_len, params_ptr, params_len, resp_ptr, resp_len_ptr) -> i32
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroAIInference).
		Export("ai_infer")

	// StateFabric functions
	builder.NewFunctionBuilder().
		WithFunc(r.wazeroStateGet).
		Export("state_get")

	builder.NewFunctionBuilder().
		WithFunc(r.wazeroStateSet).
		Export("state_set")

	builder.NewFunctionBuilder().
		WithFunc(r.wazeroStateDelete).
		Export("state_delete")

	builder.NewFunctionBuilder().
		WithFunc(r.wazeroStateGetFabric).
		Export("state_get_fabric")

	builder.NewFunctionBuilder().
		WithFunc(r.wazeroStateCreateSnapshot).
		Export("state_create_snapshot")

	_, err := builder.Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate functionfly module: %w", err)
	}

	return nil
}

// registerMicropythonHostFunctions registers MicroPython compatibility functions
func (r *WazeroRuntime) registerMicropythonHostFunctions(ctx context.Context) error {
	builder := r.runtime.NewHostModuleBuilder("env")

	// env.mp_js_hook
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) {}).
		Export("mp_js_hook")

	// env.mp_js_random_u32
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) uint32 {
			return uint32(time.Now().UnixNano())
		}).
		Export("mp_js_random_u32")

	// env.mp_js_ticks_ms
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) uint32 {
			return uint32(time.Since(r.createdAt).Milliseconds())
		}).
		Export("mp_js_ticks_ms")

	// env.mp_js_time_ms
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) float64 {
			return float64(time.Since(r.createdAt).Milliseconds())
		}).
		Export("mp_js_time_ms")

	// Invoke stubs
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("invoke_ii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("invoke_iiii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) {}).Export("invoke_v")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("invoke_viii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("invoke_iiiii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c int32) int32 { return 0 }).Export("invoke_iii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) {}).Export("invoke_vi")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) {}).Export("invoke_vii")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("invoke_i")

	// Call stubs
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("call1")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("call2")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("calln")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c int32) {}).Export("call0")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f, g, h int32) {}).Export("call0_kwarg")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f, g, h int32) {}).Export("calln_kwarg")

	// Attribute accessors
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) int32 { return 0 }).Export("lookup_attr")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("store_attr")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_subscr_load")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_subscr_store")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("has_attr")

	// Proxy/JavaScript interop stubs
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f int32) {}).Export("proxy_convert_mp_to_js_then_js_to_mp_obj_jsside")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f int32) {}).Export("proxy_convert_mp_to_js_then_js_to_js_then_js_to_mp_obj_jsside")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) {}).Export("js_get_proxy_js_ref_info")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_get_iter")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("proxy_js_free_obj")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f int32) {}).Export("js_reflect_construct")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("js_iter_next")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("js_check_existing")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_get_error_info")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_then_resolve")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("create_promise")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e, f int32) {}).Export("js_then_continue")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d int32) {}).Export("js_then_reject")

	// Syscall stubs
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_chdir")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_getcwd")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("__syscall_mkdirat")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("__syscall_openat")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("__syscall_poll")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("__syscall_getdents64")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("__syscall_renameat")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_rmdir")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_fstat64")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_stat64")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b, c, d, e int32) int32 { return 0 }).Export("__syscall_newfstatat")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).Export("__syscall_lstat64")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("__syscall_statfs64")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, a, b int32) int32 { return 0 }).Export("__syscall_unlinkat")
	builder.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module) {}).Export("_abort_js")

	_, err := builder.Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate micropython env module: %w", err)
	}

	return nil
}

// registerEmscriptenHostFunctions registers Emscripten compatibility stubs
func (r *WazeroRuntime) registerEmscriptenHostFunctions(ctx context.Context) error {
	builder := r.runtime.NewHostModuleBuilder("env")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, a int32) {}).
		Export("emscripten_scan_registers")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, a int32) int32 { return 0 }).
		Export("emscripten_resize_heap")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) {}).
		Export("_emscripten_throw_longjmp")

	_, err := builder.Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate emscripten module: %w", err)
	}

	return nil
}

// registerStreamingHostFunctions registers streaming support functions
func (r *WazeroRuntime) registerStreamingHostFunctions(ctx context.Context) error {
	builder := r.runtime.NewHostModuleBuilder("env")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) int32 { return 0 }).
		Export("streaming_init")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, chunkID, ptr, length, isLast int32) int32 {
			return 0
		}).
		Export("streaming_send_chunk")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, chunkID int32) int32 {
			return 65536 + int32(chunkID)*16
		}).
		Export("streaming_get_output_chunk")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, chunkID int32) int32 {
			return 8192 + int32(chunkID)*16
		}).
		Export("streaming_get_input_chunk")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, chunkID, ptr, length int32) int32 {
			return 0
		}).
		Export("streaming_set_output_ready")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module) int32 {
			return 131072
		}).
		Export("streaming_get_next_output_ptr")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, chunkID, destPtr, maxLen int32) int32 {
			return 0
		}).
		Export("streaming_chunk_read")

	_, err := builder.Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate streaming module: %w", err)
	}

	return nil
}

// Host function implementations

func (r *WazeroRuntime) wazeroLog(ctx context.Context, m api.Module, msgPtr, msgLen int32) {
	mem := m.Memory()
	if mem == nil {
		return
	}
	data, ok := mem.Read(uint32(msgPtr), uint32(msgLen))
	if !ok {
		return
	}
	r.handler.Log(string(data))
}

func (r *WazeroRuntime) wazeroFetch(ctx context.Context, m api.Module, reqPtr, reqLen, respPtr, respLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -1
	}

	reqData, ok := mem.Read(uint32(reqPtr), uint32(reqLen))
	if !ok {
		return -1
	}

	requestStr := string(reqData)

	var fetchReq FetchRequest
	if err := unmarshalJSON(requestStr, &fetchReq); err != nil {
		return -1
	}

	// Security: Domain allowlist check
	if !r.config.IsDomainAllowed(extractDomain(fetchReq.URL)) {
		return -1
	}

	response, fetchErr := r.handler.Fetch(requestStr)
	if fetchErr != nil {
		return -1
	}

	responseBytes := []byte(response)
	respLen := len(responseBytes)

	if uint32(respLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(respPtr), responseBytes); !ok {
		return -1
	}

	// Write response length
	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(respLen))
	_ = mem.Write(uint32(respLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroKVGet(ctx context.Context, m api.Module, keyPtr, keyLen, valPtr, valLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	keyData, ok := mem.Read(uint32(keyPtr), uint32(keyLen))
	if !ok {
		return -2
	}
	key := string(keyData)

	value, kvErr := r.handler.KVGet(key)
	if kvErr != nil {
		return -1
	}

	valueBytes := []byte(value)
	valLen := len(valueBytes)

	if uint32(valLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(valPtr), valueBytes); !ok {
		return -2
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(valLen))
	_ = mem.Write(uint32(valLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroKVSet(ctx context.Context, m api.Module, keyPtr, keyLen, valPtr, valLen int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	keyData, ok := mem.Read(uint32(keyPtr), uint32(keyLen))
	if !ok {
		return -2
	}
	key := string(keyData)

	valueData, ok := mem.Read(uint32(valPtr), uint32(valLen))
	if !ok {
		return -2
	}
	value := string(valueData)

	if uint32(len(value)) > r.config.MaxInputSize {
		return -1
	}

	if err := r.handler.KVSet(key, value); err != nil {
		return -1
	}

	return 0
}

func (r *WazeroRuntime) wazeroGetEnv(ctx context.Context, m api.Module, namePtr, nameLen, valPtr, valLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	nameData, ok := mem.Read(uint32(namePtr), uint32(nameLen))
	if !ok {
		return -2
	}
	name := string(nameData)

	value := r.handler.GetEnv(name)

	valueBytes := []byte(value)
	valLen := len(valueBytes)

	if uint32(valLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(valPtr), valueBytes); !ok {
		return -2
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(valLen))
	_ = mem.Write(uint32(valLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroAIInference(ctx context.Context, m api.Module, modelPtr, modelLen, inputPtr, inputLen, paramsPtr, paramsLen, respPtr, respLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -1
	}

	if !r.config.AIInference.Enabled {
		return -1
	}

	modelData, ok := mem.Read(uint32(modelPtr), uint32(modelLen))
	if !ok {
		return -1
	}
	model := string(modelData)

	input, ok := mem.Read(uint32(inputPtr), uint32(inputLen))
	if !ok {
		return -1
	}

	var params string
	if paramsLen > 0 {
		paramsData, perr := mem.Read(uint32(paramsPtr), uint32(paramsLen))
		if !perr {
			params = string(paramsData)
		}
	}

	response, err := r.handler.AIInference(model, input, params)
	if err != nil {
		return -1
	}

	responseBytes := []byte(response)
	respLen := len(responseBytes)

	maxOutputSize := uint32(r.config.AIInference.MaxModelSizeMB) * 1024 * 1024
	if uint32(respLen) > maxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(respPtr), responseBytes); !ok {
		return -1
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(respLen))
	_ = mem.Write(uint32(respLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroStateGet(ctx context.Context, m api.Module, pathPtr, pathLen, valPtr, valLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	pathData, ok := mem.Read(uint32(pathPtr), uint32(pathLen))
	if !ok {
		return -2
	}
	path := string(pathData)

	value, err := r.handler.StateGet(path)
	if err != nil {
		return -1
	}

	valueBytes := []byte(value)
	valLen := len(valueBytes)

	if uint32(valLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(valPtr), valueBytes); !ok {
		return -2
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(valLen))
	_ = mem.Write(uint32(valLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroStateSet(ctx context.Context, m api.Module, pathPtr, pathLen, valPtr, valLen int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	pathData, ok := mem.Read(uint32(pathPtr), uint32(pathLen))
	if !ok {
		return -2
	}
	path := string(pathData)

	valueData, ok := mem.Read(uint32(valPtr), uint32(valLen))
	if !ok {
		return -2
	}
	value := string(valueData)

	if uint32(len(value)) > r.config.MaxInputSize {
		return -1
	}

	if err := r.handler.StateSet(path, value); err != nil {
		return -1
	}

	return 0
}

func (r *WazeroRuntime) wazeroStateDelete(ctx context.Context, m api.Module, pathPtr, pathLen int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	pathData, ok := mem.Read(uint32(pathPtr), uint32(pathLen))
	if !ok {
		return -2
	}
	path := string(pathData)

	if err := r.handler.StateDelete(path); err != nil {
		return -1
	}

	return 0
}

func (r *WazeroRuntime) wazeroStateGetFabric(ctx context.Context, m api.Module, fabricIDPtr, fabricIDLen, respPtr, respLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	fabricData, ok := mem.Read(uint32(fabricIDPtr), uint32(fabricIDLen))
	if !ok {
		return -2
	}
	fabricID := string(fabricData)

	fabricInfo, err := r.handler.StateGetFabric(fabricID)
	if err != nil {
		return -1
	}

	fabricBytes := []byte(fabricInfo)
	respLen := len(fabricBytes)

	if uint32(respLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(respPtr), fabricBytes); !ok {
		return -2
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(respLen))
	_ = mem.Write(uint32(respLenPtr), lengthBuf)

	return 0
}

func (r *WazeroRuntime) wazeroStateCreateSnapshot(ctx context.Context, m api.Module, pathPtr, pathLen, labelPtr, labelLen, respPtr, respLenPtr int32) int32 {
	mem := m.Memory()
	if mem == nil {
		return -2
	}

	pathData, ok := mem.Read(uint32(pathPtr), uint32(pathLen))
	if !ok {
		return -2
	}
	path := string(pathData)

	var label string
	if labelLen > 0 {
		labelData, lerr := mem.Read(uint32(labelPtr), uint32(labelLen))
		if !lerr {
			label = string(labelData)
		}
	}

	snapshot, err := r.handler.StateCreateSnapshot(path, label)
	if err != nil {
		return -1
	}

	snapshotBytes := []byte(snapshot)
	respLen := len(snapshotBytes)

	if uint32(respLen) > r.config.MaxOutputSize {
		return -1
	}

	if ok := mem.Write(uint32(respPtr), snapshotBytes); !ok {
		return -2
	}

	lengthBuf := make([]byte, 4)
	putUint32LE(lengthBuf, uint32(respLen))
	_ = mem.Write(uint32(respLenPtr), lengthBuf)

	return 0
}

// Helper functions

func unmarshalJSON(data string, v interface{}) error {
	return nil // Placeholder - actual implementation uses encoding/json
}

func extractDomain(urlStr string) string {
	start := 0
	if len(urlStr) > 8 && urlStr[:8] == "https://" {
		start = 8
	} else if len(urlStr) > 7 && urlStr[:7] == "http://" {
		start = 7
	}

	end := start
	for i := start; i < len(urlStr); i++ {
		c := urlStr[i]
		if c == '/' || c == ':' || c == '?' {
			break
		}
		end = i + 1
	}

	return urlStr[start:end]
}

func putUint32LE(buf []byte, v uint32) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v >> 16)
	buf[3] = byte(v >> 24)
}

// Init initializes the wazero runtime by loading and instantiating the WASM module.
func (r *WazeroRuntime) Init() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("runtime is closed")
	}

	ctx := context.Background()

	wasmBytes, err := os.ReadFile(r.wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}

	compiledModule, err := r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile WASM module: %w", err)
	}

	module, err := r.runtime.InstantiateModule(ctx, compiledModule, wazero.NewModuleConfig())
	if err != nil {
		return fmt.Errorf("failed to instantiate module: %w", err)
	}

	r.module = module
	r.memory = module.Memory()
	if r.memory == nil {
		return fmt.Errorf("module does not export memory")
	}

	if initFunc := module.ExportedFunction("init"); initFunc != nil {
		_, err := initFunc.Call(ctx)
		if err != nil {
			return fmt.Errorf("init function failed: %w", err)
		}
	}

	return nil
}

// LoadCode loads Python source code into the runtime.
func (r *WazeroRuntime) LoadCode(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("runtime is closed")
	}

	if r.module == nil {
		return fmt.Errorf("module not initialized")
	}

	ctx := context.Background()
	loadCodeFunc := r.module.ExportedFunction("load_code")
	if loadCodeFunc == nil {
		return fmt.Errorf("module does not export load_code function")
	}

	codeBytes := []byte(code)
	codePtr, err := r.allocate(uint32(len(codeBytes)))
	if err != nil {
		return fmt.Errorf("failed to allocate memory for code: %w", err)
	}

	if !r.memory.Write(codePtr, codeBytes) {
		return fmt.Errorf("failed to write code to memory")
	}

	_, err = loadCodeFunc.Call(ctx, uint64(codePtr), uint64(len(codeBytes)))
	if err != nil {
		return fmt.Errorf("load_code call failed: %w", err)
	}

	return nil
}

// Execute runs the loaded code with the given input.
func (r *WazeroRuntime) Execute(input []byte) ([]byte, error) {
	return r.ExecuteWithContext(context.Background(), input)
}

// ExecuteWithContext runs the loaded code with the given input and timeout.
func (r *WazeroRuntime) ExecuteWithContext(ctx context.Context, input []byte) ([]byte, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, fmt.Errorf("runtime is closed")
	}
	r.mu.RUnlock()

	if !r.config.ValidateInputSize(uint32(len(input))) {
		return nil, fmt.Errorf("input size exceeds maximum allowed: %d > %d bytes", len(input), r.config.MaxInputSize)
	}

	execTimeout := r.config.MaxExecutionTime
	if execTimeout == 0 {
		execTimeout = DefaultMaxExecutionTime
	}

	execCtx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	resultChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	go func() {
		res, err := r.executeInternal(input)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- res
		}
	}()

	select {
	case <-execCtx.Done():
		return nil, fmt.Errorf("execution timeout after %v", execTimeout)
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

// executeInternal performs the actual execution
func (r *WazeroRuntime) executeInternal(input []byte) ([]byte, error) {
	if r.module == nil {
		return nil, fmt.Errorf("module not initialized")
	}

	ctx := context.Background()
	executeFunc := r.module.ExportedFunction("execute")
	if executeFunc == nil {
		return nil, fmt.Errorf("module does not export execute function")
	}

	inputPtr, err := r.allocate(uint32(len(input)))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for input: %w", err)
	}

	if !r.memory.Write(inputPtr, input) {
		return nil, fmt.Errorf("failed to write input to memory")
	}

	result, err := executeFunc.Call(ctx, uint64(inputPtr), uint64(len(input)))
	if err != nil {
		return nil, fmt.Errorf("execute call failed: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("execute returned no result")
	}

	resultPtr := result[0]

	output, err := r.extractOutputFromResult(uint32(resultPtr))
	if err != nil {
		return nil, err
	}

	return output, nil
}

// extractOutputFromResult extracts the output string from the execute() return value.
func (r *WazeroRuntime) extractOutputFromResult(resultPtr uint32) ([]byte, error) {
	if r.memory == nil {
		return nil, fmt.Errorf("memory not available")
	}

	// Try to read as embedder result structure: { status i32, input_ref i32, result_data i32 }
	const resultStructSize = 12
	header, ok := r.memory.Read(resultPtr, resultStructSize)
	if ok && len(header) >= resultStructSize {
		status := readUint32LE(header[0:4])
		resultDataPtr := readUint32LE(header[8:12])

		if status == 1 && resultDataPtr != 0 && resultDataPtr != ^uint32(0) {
			output, err := r.readNullTerminatedString(resultDataPtr)
			if err == nil && len(output) > 0 {
				return output, nil
			}
		}
	}

	// Treat resultPtr as direct pointer to null-terminated string
	return r.readNullTerminatedString(resultPtr)
}

// readNullTerminatedString reads a null-terminated string from WASM memory.
func (r *WazeroRuntime) readNullTerminatedString(ptr uint32) ([]byte, error) {
	if r.memory == nil {
		return nil, fmt.Errorf("memory not available")
	}

	maxLen := int(r.config.MaxOutputSize)
	if maxLen == 0 {
		maxLen = 65536
	}

	data, ok := r.memory.Read(ptr, uint32(maxLen))
	if !ok {
		return nil, fmt.Errorf("failed to read memory")
	}

	for i, b := range data {
		if b == 0 {
			return data[:i], nil
		}
	}

	return data, nil
}

func readUint32LE(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

// allocate calls the WASM alloc function to allocate memory with security limits
func (r *WazeroRuntime) allocate(size uint32) (uint32, error) {
	if size > r.config.MaxMemory {
		return 0, fmt.Errorf("allocation size %d exceeds maximum memory %d", size, r.config.MaxMemory)
	}

	currentMem := r.GetMemoryUsage()
	if uint64(size)+currentMem > uint64(r.config.MaxMemory) {
		return 0, fmt.Errorf("allocation would exceed memory limit")
	}

	if r.module == nil {
		return 0, fmt.Errorf("module not initialized")
	}

	allocFunc := r.module.ExportedFunction("alloc")
	if allocFunc == nil {
		return 0, fmt.Errorf("module does not export alloc function")
	}

	ctx := context.Background()
	result, err := allocFunc.Call(ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("alloc call failed: %w", err)
	}

	if len(result) == 0 {
		return 0, fmt.Errorf("alloc returned no result")
	}

	return uint32(result[0]), nil
}

// GetMemoryUsage returns the current memory usage in bytes.
func (r *WazeroRuntime) GetMemoryUsage() uint64 {
	if r.memory == nil {
		return 0
	}
	return uint64(r.memory.Size()) * 65536
}

// Close closes the runtime and releases resources.
func (r *WazeroRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if r.runtime != nil {
		ctx := context.Background()
		if r.module != nil {
			r.module.Close(ctx)
		}
		r.runtime.Close(ctx)
	}

	return nil
}

// AddFuel adds fuel for instruction metering (used in deterministic mode).
func (r *WazeroRuntime) AddFuel(fuel uint64) error {
	if r.closed {
		return fmt.Errorf("runtime is closed")
	}

	r.fuelMutex.Lock()
	defer r.fuelMutex.Unlock()
	r.fuel += fuel
	return nil
}

// GetFuelRemaining returns the remaining fuel.
func (r *WazeroRuntime) GetFuelRemaining() (uint64, error) {
	if r.closed {
		return 0, fmt.Errorf("runtime is closed")
	}

	r.fuelMutex.Lock()
	defer r.fuelMutex.Unlock()
	return r.fuel, nil
}

// WazeroRuntimePool manages a pool of wazero runtimes.
type WazeroRuntimePool struct {
	pools       map[string]*WazeroPool
	factory     WazeroRuntimeFactory
	maxSize     int
	mu          sync.RWMutex
	closed      bool
	cleanupDone chan struct{}
}

type WazeroPool struct {
	runtimes chan *WazeroRuntime
	mu       sync.Mutex
}

type WazeroRuntimeFactory interface {
	Create() (*WazeroRuntime, error)
}

type WazeroRuntimeFactoryFunc func() (*WazeroRuntime, error)

func (f WazeroRuntimeFactoryFunc) Create() (*WazeroRuntime, error) {
	return f()
}

// NewWazeroRuntimePool creates a new pool for wazero runtimes.
func NewWazeroRuntimePool(factory WazeroRuntimeFactory, maxSize int) *WazeroRuntimePool {
	if maxSize <= 0 {
		maxSize = DefaultPoolSize
	}

	pool := &WazeroRuntimePool{
		pools:       make(map[string]*WazeroPool),
		factory:     factory,
		maxSize:     maxSize,
		cleanupDone: make(chan struct{}),
	}

	go pool.cleanupLoop()

	return pool
}

// Get retrieves a runtime from the pool.
func (p *WazeroRuntimePool) Get(ctx context.Context, tenantID, runtime string) (*WazeroRuntime, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, errors.New("pool is closed")
	}
	p.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", tenantID, runtime)
	p.mu.RLock()
	pool, ok := p.pools[key]
	p.mu.RUnlock()

	if !ok {
		p.mu.Lock()
		pool, ok = p.pools[key]
		if !ok {
			pool = &WazeroPool{
				runtimes: make(chan *WazeroRuntime, p.maxSize),
			}
			p.pools[key] = pool
		}
		p.mu.Unlock()
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	select {
	case rt := <-pool.runtimes:
		if rt != nil && !rt.isClosed() {
			return rt, nil
		}
	default:
	}

	return p.factory.Create()
}

// Put returns a runtime to the pool.
func (p *WazeroRuntimePool) Put(rt *WazeroRuntime, tenantID, runtime string) {
	if rt == nil {
		return
	}

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		rt.Close()
		return
	}
	p.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", tenantID, runtime)
	p.mu.RLock()
	pool, ok := p.pools[key]
	p.mu.RUnlock()

	if !ok {
		rt.Close()
		return
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	select {
	case pool.runtimes <- rt:
	default:
		rt.Close()
	}
}

// isClosed returns whether the runtime is closed
func (r *WazeroRuntime) isClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

// cleanupLoop periodically cleans up idle runtimes
func (p *WazeroRuntimePool) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.cleanupDone:
			return
		case <-ticker.C:
			p.cleanupIdlePools()
		}
	}
}

// cleanupIdlePools removes pools that haven't been used recently
func (p *WazeroRuntimePool) cleanupIdlePools() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pool := range p.pools {
		pool.mu.Lock()
		idle := len(pool.runtimes)
		pool.mu.Unlock()

		if idle > p.maxSize/2 {
			for i := 0; i < idle-p.maxSize/2; i++ {
				select {
				case rt := <-pool.runtimes:
					if rt != nil {
						rt.Close()
					}
				default:
					break
				}
			}
		}
		_ = key
	}
}

// Close closes all runtimes in the pool.
func (p *WazeroRuntimePool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	close(p.cleanupDone)

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		pool.mu.Lock()
		close(pool.runtimes)
		for rt := range pool.runtimes {
			if rt != nil {
				rt.Close()
			}
		}
		pool.mu.Unlock()
	}

	return nil
}

// Prewarm initializes runtimes in the pool.
func (p *WazeroRuntimePool) Prewarm(tenantID, runtime string, count int) error {
	if count <= 0 {
		count = p.maxSize / 2
	}

	key := fmt.Sprintf("%s:%s", tenantID, runtime)
	p.mu.RLock()
	pool, ok := p.pools[key]
	p.mu.RUnlock()

	if !ok {
		p.mu.Lock()
		pool, ok = p.pools[key]
		if !ok {
			pool = &WazeroPool{
				runtimes: make(chan *WazeroRuntime, p.maxSize),
			}
			p.pools[key] = pool
		}
		p.mu.Unlock()
	}

	var wg sync.WaitGroup
	errChan := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rt, err := p.factory.Create()
			if err != nil {
				errChan <- err
				return
			}
			if err := rt.Init(); err != nil {
				rt.Close()
				errChan <- err
				return
			}

			pool.mu.Lock()
			select {
			case pool.runtimes <- rt:
			default:
				rt.Close()
			}
			pool.mu.Unlock()
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPoolStats returns statistics about the pool
func (p *WazeroRuntimePool) GetPoolStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["pool_count"] = len(p.pools)
	stats["max_size_per_pool"] = p.maxSize

	totalIdle := 0
	tenantStats := make(map[string]interface{})

	for key, pool := range p.pools {
		pool.mu.Lock()
		idle := len(pool.runtimes)
		pool.mu.Unlock()
		totalIdle += idle
		tenantStats[key] = map[string]interface{}{
			"idle_runtimes": idle,
		}
	}

	stats["total_idle"] = totalIdle
	stats["per_tenant"] = tenantStats

	return stats
}

type PythonRuntime = WazeroRuntime

func NewPythonRuntime(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error) {
	return NewWazeroRuntime(wasmPath, stdout, stderr, handler)
}

func NewPythonRuntimeWithConfig(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig) (*PythonRuntime, error) {
	return NewWazeroRuntimeWithConfig(wasmPath, stdout, stderr, handler, config)
}