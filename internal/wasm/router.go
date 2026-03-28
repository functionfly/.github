// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains runtime router implementation
package wasm

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// RuntimeType represents the type of WASM runtime
type RuntimeType string

const (
	// RuntimePython represents Python WASM runtime
	RuntimePython RuntimeType = "python"

	// RuntimePythonWASM represents Python WASM (Micropython) runtime
	RuntimePythonWASM RuntimeType = "python-wasm"

	// RuntimeJavaScript represents JavaScript runtime
	RuntimeJavaScript RuntimeType = "javascript"

	// RuntimeTypeScriptWASM represents TypeScript WASM runtime
	RuntimeTypeScriptWASM RuntimeType = "typescript-wasm"

	// RuntimePythonWasi represents Python with WASI support
	RuntimePythonWasi RuntimeType = "python-wasi"

	// RuntimeJavaScriptQuickJS represents QuickJS JavaScript runtime
	RuntimeJavaScriptQuickJS RuntimeType = "javascript-quickjs"

	// RuntimeRust represents Rust WASM (wasm32-wasi) runtime
	RuntimeRust RuntimeType = "rust"

	// RuntimeBrowserNativeWASM represents Browser Native WebAssembly (0ms cold start)
	RuntimeBrowserNativeWASM RuntimeType = "browser-wasm"

	// RuntimeWASM3IoT represents WASM3 IoT runtime (~500ms cold start, lightweight)
	RuntimeWASM3IoT RuntimeType = "wasm3-iot"

	// RuntimeUnknown represents an unknown runtime
	RuntimeUnknown RuntimeType = "unknown"
)

// IsValid checks if the runtime type is valid
func (r RuntimeType) IsValid() bool {
	switch r {
	case RuntimePython, RuntimePythonWASM, RuntimeJavaScript, RuntimeTypeScriptWASM,
		RuntimePythonWasi, RuntimeJavaScriptQuickJS, RuntimeRust, RuntimeBrowserNativeWASM,
		RuntimeWASM3IoT:
		return true
	default:
		return false
	}
}

// String returns the string representation of the runtime type
func (r RuntimeType) String() string {
	return string(r)
}

// RuntimeRouter routes execution to the correct runtime based on runtime type
type RuntimeRouter struct {
	mu           sync.RWMutex
	runtimes     map[RuntimeType]RuntimeProvider
	pool         *InstancePool
	defaultPool  *SimpleInstancePool
	config       *WASMSecurityConfig
}

// RuntimeProvider defines the interface for runtime providers
type RuntimeProvider interface {
	// Execute executes the function with the given input
	Execute(ctx context.Context, input []byte) ([]byte, error)

	// ExecuteWithConfig executes with custom configuration
	ExecuteWithConfig(ctx context.Context, input []byte, config interface{}) ([]byte, error)

	// Close closes the runtime
	Close() error
}

// RuntimeProviderFactory creates runtime providers
type RuntimeProviderFactory func() (RuntimeProvider, error)

// NewRuntimeRouter creates a new runtime router
func NewRuntimeRouter(config *WASMSecurityConfig) *RuntimeRouter {
	if config == nil {
		config = NewDefaultSecurityConfig()
	}

	return &RuntimeRouter{
		runtimes: make(map[RuntimeType]RuntimeProvider),
		config:    config,
	}
}

// RegisterRuntime registers a runtime provider for a runtime type
func (r *RuntimeRouter) RegisterRuntime(runtimeType RuntimeType, provider RuntimeProvider) error {
	if !runtimeType.IsValid() {
		return fmt.Errorf("invalid runtime type: %s", runtimeType)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.runtimes[runtimeType] = provider
	log.Printf("[WASM Router] Registered runtime: %s", runtimeType)

	return nil
}

// RegisterRuntimeFactory registers a runtime factory for a runtime type
func (r *RuntimeRouter) RegisterRuntimeFactory(runtimeType RuntimeType, factory RuntimeProviderFactory) error {
	provider, err := factory()
	if err != nil {
		return fmt.Errorf("failed to create runtime provider: %w", err)
	}

	return r.RegisterRuntime(runtimeType, provider)
}

// GetRuntime returns the runtime provider for a runtime type
func (r *RuntimeRouter) GetRuntime(runtimeType RuntimeType) (RuntimeProvider, error) {
	r.mu.RLock()
	provider, exists := r.runtimes[runtimeType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("runtime not registered: %s", runtimeType)
	}

	return provider, nil
}

// HasRuntime checks if a runtime is registered
func (r *RuntimeRouter) HasRuntime(runtimeType RuntimeType) bool {
	r.mu.RLock()
	_, exists := r.runtimes[runtimeType]
	r.mu.RUnlock()

	return exists
}

// SetPool sets the instance pool for the router
func (r *RuntimeRouter) SetPool(pool *InstancePool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pool = pool
}

// SetDefaultPool sets the default instance pool
func (r *RuntimeRouter) SetDefaultPool(pool *SimpleInstancePool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultPool = pool
}

// Execute executes a function using the appropriate runtime
func (r *RuntimeRouter) Execute(ctx context.Context, runtimeType RuntimeType, input []byte) ([]byte, error) {
	provider, err := r.GetRuntime(runtimeType)
	if err != nil {
		return nil, err
	}

	return provider.Execute(ctx, input)
}

// ExecuteWithPool executes using the instance pool
func (r *RuntimeRouter) ExecuteWithPool(ctx context.Context, tenantID string, runtimeType RuntimeType, input []byte) ([]byte, error) {
	r.mu.RLock()
	pool := r.pool
	r.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("instance pool not configured")
	}

	// Get instance from pool
	inst, err := pool.Get(ctx, tenantID, string(runtimeType))
	if err != nil {
		return nil, fmt.Errorf("failed to get instance from pool: %w", err)
	}

	// Execute
	output, err := inst.Instance.ExecuteWithContext(ctx, input)

	// Return instance to pool
	pool.Put(inst)

	return output, err
}

// ExecuteDeterministic executes with deterministic mode using DeterministicExecutor.
func (r *RuntimeRouter) ExecuteDeterministic(ctx context.Context, tenantID string, runtimeType RuntimeType, input []byte, detConfig *DeterministicConfig) (*DeterministicResult, error) {
	r.mu.RLock()
	pool := r.pool
	r.mu.RUnlock()
	if pool == nil {
		return nil, fmt.Errorf("instance pool not configured")
	}
	if detConfig == nil {
		detConfig = DefaultDeterministicConfig()
	}
	execID := "det-"
	if len(input) >= 8 {
		execID += hex.EncodeToString(input[:8])
	} else if len(input) > 0 {
		execID += hex.EncodeToString(input)
	} else {
		execID += "0"
	}
	det := NewDeterministicExecutor(pool, detConfig)
	return det.Execute(ctx, tenantID, "", execID, string(runtimeType), input)
}

// RuntimeTypeFromString parses a runtime type from string
func RuntimeTypeFromString(s string) RuntimeType {
	switch s {
	case "python", "py":
		return RuntimePython
	case "python-wasm", "py-wasm", "micropython":
		return RuntimePythonWASM
	case "javascript", "js":
		return RuntimeJavaScript
	case "typescript-wasm", "ts-wasm", "ts":
		return RuntimeTypeScriptWASM
	case "python-wasi", "py-wasi":
		return RuntimePythonWasi
	case "javascript-quickjs", "quickjs":
		return RuntimeJavaScriptQuickJS
	case "rust", "rs":
		return RuntimeRust
	case "browser-wasm", "browser", "browser-native", "native-wasm":
		return RuntimeBrowserNativeWASM
	case "wasm3-iot", "wasm3", "iot":
		return RuntimeWASM3IoT
	default:
		return RuntimeUnknown
	}
}

// GetSupportedRuntimes returns the list of supported runtime types
func (r *RuntimeRouter) GetSupportedRuntimes() []RuntimeType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	runtimes := make([]RuntimeType, 0, len(r.runtimes))
	for rt := range r.runtimes {
		runtimes = append(runtimes, rt)
	}

	return runtimes
}

// Close closes all registered runtimes
func (r *RuntimeRouter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for rt, provider := range r.runtimes {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close runtime %s: %w", rt, err))
		}
	}

	r.runtimes = make(map[RuntimeType]RuntimeProvider)

	if len(errs) > 0 {
		// Return first error
		return errs[0]
	}

	return nil
}

// DefaultRuntimeRouter is the global runtime router
var defaultRouter *RuntimeRouter
var routerOnce sync.Once

// GetDefaultRouter returns the default runtime router
func GetDefaultRouter() *RuntimeRouter {
	routerOnce.Do(func() {
		defaultRouter = NewRuntimeRouter(nil)
	})

	return defaultRouter
}

// InitDefaultRouter creates a runtime router with the given security config.
// Callers must call SetPool or RegisterRuntime to enable execution.
func InitDefaultRouter(config *WASMSecurityConfig) *RuntimeRouter {
	router := NewRuntimeRouter(config)
	log.Printf("[WASM Router] Initialized default runtime router")
	return router
}

// RouterExecutor wraps a runtime router for use with instance pools
type RouterExecutor struct {
	router      *RuntimeRouter
	runtimeType RuntimeType
	tenantID    string
}

// NewRouterExecutor creates a new router executor
func NewRouterExecutor(router *RuntimeRouter, runtimeType RuntimeType, tenantID string) *RouterExecutor {
	return &RouterExecutor{
		router:      router,
		runtimeType: runtimeType,
		tenantID:    tenantID,
	}
}

// Execute executes a function using the router
func (e *RouterExecutor) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if e.router == nil {
		return nil, fmt.Errorf("router not initialized")
	}

	return e.router.ExecuteWithPool(ctx, e.tenantID, e.runtimeType, input)
}

// RouterWithMetrics wraps RuntimeRouter with metrics collection
type RouterWithMetrics struct {
	router   *RuntimeRouter
	recorder *MetricsRecorder
}

// NewRouterWithMetrics creates a new router with metrics
func NewRouterWithMetrics(router *RuntimeRouter, tenantID string) *RouterWithMetrics {
	return &RouterWithMetrics{
		router:   router,
		recorder: NewMetricsRecorder("router", tenantID),
	}
}

// ExecuteWithMetrics executes with metrics collection
func (m *RouterWithMetrics) ExecuteWithMetrics(ctx context.Context, runtimeType RuntimeType, input []byte) ([]byte, error) {
	startTime := time.Now()

	output, err := m.router.Execute(ctx, runtimeType, input)

	duration := time.Since(startTime)
	status := "success"
	if err != nil {
		status = "error"
		m.recorder.RecordError(ParseErrorType(err))
	}

	m.recorder.RecordExecutionWithSizes(duration, status, len(input), len(output))

	return output, err
}

// WASMRuntimeProvider wraps PythonRuntime for use with RuntimeRouter
type WASMRuntimeProvider struct {
	runtime *PythonRuntime
}

// NewWASMRuntimeProvider creates a new WASM runtime provider
func NewWASMRuntimeProvider(runtime *PythonRuntime) *WASMRuntimeProvider {
	return &WASMRuntimeProvider{runtime: runtime}
}

// Execute executes the function
func (p *WASMRuntimeProvider) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return p.runtime.ExecuteWithContext(ctx, input)
}

// ExecuteWithConfig executes with custom configuration (placeholder)
func (p *WASMRuntimeProvider) ExecuteWithConfig(ctx context.Context, input []byte, config interface{}) ([]byte, error) {
	return p.runtime.ExecuteWithContext(ctx, input)
}

// Close closes the runtime
func (p *WASMRuntimeProvider) Close() error {
	if p.runtime != nil {
		return p.runtime.Close()
	}
	return nil
}

// CreateWASMProviderFromFactory creates a WASM runtime provider from a factory function
func CreateWASMProviderFromFactory(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (RuntimeProvider, error) {
	runtime, err := NewPythonRuntimeWithConfig(wasmPath, stdout, stderr, handler, NewDefaultSecurityConfig())
	if err != nil {
		return nil, err
	}

	return &WASMRuntimeProvider{runtime: runtime}, nil
}
