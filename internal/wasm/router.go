// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains runtime router implementation
package wasm

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/sirupsen/logrus"
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

	// RuntimeGoWASM represents Go WASM (wasi) runtime
	RuntimeGoWASM RuntimeType = "go-wasm"

	// RuntimeC represents C WASM runtime (via Emscripten/WASI-SDK)
	RuntimeC RuntimeType = "c"

	// RuntimeCpp represents C++ WASM runtime (via Emscripten/WASI-SDK)
	RuntimeCpp RuntimeType = "cpp"

	// RuntimeRubyWASM represents Ruby WASM runtime (mruby interpreter)
	RuntimeRubyWASM RuntimeType = "ruby-wasm"

	// RuntimeKotlinWASM represents Kotlin WASM runtime
	RuntimeKotlinWASM RuntimeType = "kotlin-wasm"

	// RuntimeSwiftWASM represents Swift WASM runtime (SwiftWasm)
	RuntimeSwiftWASM RuntimeType = "swift-wasm"

	// RuntimeUnknown represents an unknown runtime
	RuntimeUnknown RuntimeType = "unknown"
)

// IsValid checks if the runtime type is valid
func (r RuntimeType) IsValid() bool {
	switch r {
	case RuntimePython, RuntimePythonWASM, RuntimeJavaScript, RuntimeTypeScriptWASM,
		RuntimePythonWasi, RuntimeJavaScriptQuickJS, RuntimeRust, RuntimeBrowserNativeWASM,
		RuntimeWASM3IoT, RuntimeGoWASM, RuntimeC, RuntimeCpp, RuntimeRubyWASM,
		RuntimeKotlinWASM, RuntimeSwiftWASM:
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

	logger := logrus.WithField("runtime", runtimeType)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runtimes[runtimeType] = provider
	logger.Info("Registered runtime")

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
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < 0 {
			return nil, fmt.Errorf("context deadline already exceeded")
		}
	}

	ctx, _ = tracing.StartSpan(ctx, fmt.Sprintf("wasm.execute.%s", runtimeType))
	defer tracing.Finish(ctx)

	tracing.SetAttribute(ctx, "runtime_type", string(runtimeType))
	tracing.SetAttribute(ctx, "input_size", len(input))

	provider, err := r.GetRuntime(runtimeType)
	if err != nil {
		tracing.RecordError(ctx, err)
		return nil, err
	}

	output, err := provider.Execute(ctx, input)

	if err != nil {
		tracing.RecordError(ctx, err)
		tracing.SetAttribute(ctx, "error", true)
	} else {
		tracing.SetAttribute(ctx, "output_size", len(output))
	}

	return output, err
}

// ExecuteWithTimeout executes with a timeout wrapper
func (r *RuntimeRouter) ExecuteWithTimeout(ctx context.Context, runtimeType RuntimeType, input []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})
	var output []byte
	var execErr error

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic":   rec,
					"runtime": runtimeType,
				}).Error("RuntimeRouter ExecuteWithTimeout goroutine panicked")
				execErr = fmt.Errorf("execution panicked: %v", rec)
				close(done)
			}
		}()
		output, execErr = r.Execute(ctx, runtimeType, input)
		close(done)
	}()

	select {
	case <-done:
		return output, execErr
	case <-ctx.Done():
		logrus.WithFields(logrus.Fields{
			"runtime":    runtimeType,
			"timeout_ms": timeout.Milliseconds(),
		}).Warn("Execution timed out")
		return nil, fmt.Errorf("execution timed out after %v", timeout)
	}
}

// ExecuteWithPool executes using the instance pool
func (r *RuntimeRouter) ExecuteWithPool(ctx context.Context, tenantID string, runtimeType RuntimeType, input []byte) ([]byte, error) {
	ctx, _ = tracing.StartSpan(ctx, fmt.Sprintf("wasm.execute_pool.%s", runtimeType))
	defer tracing.Finish(ctx)

	tracing.SetAttribute(ctx, "runtime_type", string(runtimeType))
	tracing.SetAttribute(ctx, "tenant_id", tenantID)
	tracing.SetAttribute(ctx, "input_size", len(input))

	r.mu.RLock()
	pool := r.pool
	r.mu.RUnlock()

	if pool == nil {
		err := fmt.Errorf("instance pool not configured")
		tracing.RecordError(ctx, err)
		return nil, err
	}

	// Get instance from pool
	inst, err := pool.Get(ctx, tenantID, string(runtimeType))
	if err != nil {
		tracing.RecordError(ctx, err)
		return nil, fmt.Errorf("failed to get instance from pool: %w", err)
	}

	// Ensure instance is always returned to pool
	var output []byte
	var execErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("runtime", runtimeType).Error("Recovered from panic during execution")
				execErr = fmt.Errorf("execution panicked: %v", r)
				tracing.RecordError(ctx, execErr)
			}
			pool.Put(inst)
		}()
		output, execErr = inst.Instance.ExecuteWithContext(ctx, input)
		if execErr != nil {
			tracing.RecordError(ctx, execErr)
			tracing.SetAttribute(ctx, "error", true)
		} else {
			tracing.SetAttribute(ctx, "output_size", len(output))
		}
	}()

	return output, execErr
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
	case "go", "go1.21", "go-wasm":
		return RuntimeGoWASM
	case "c", "c11":
		return RuntimeC
	case "cpp", "cpp17", "c++":
		return RuntimeCpp
	case "ruby", "ruby3.3", "ruby-wasm", "mruby":
		return RuntimeRubyWASM
	case "kotlin", "kotlin1.9", "kotlin-wasm":
		return RuntimeKotlinWASM
	case "swift", "swift5.9", "swift-wasm":
		return RuntimeSwiftWASM
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

// Close closes all registered runtimes and releases resources
func (r *RuntimeRouter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for rt, provider := range r.runtimes {
		if err := provider.Close(); err != nil {
			logrus.WithError(err).WithField("runtime", rt).Error("Failed to close runtime")
			errs = append(errs, fmt.Errorf("failed to close runtime %s: %w", rt, err))
		}
	}

	r.runtimes = make(map[RuntimeType]RuntimeProvider)

	if r.pool != nil {
		if err := r.pool.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close instance pool")
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// Shutdown gracefully shuts down the router with a timeout
func (r *RuntimeRouter) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- r.Close()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		logrus.Warn("Router shutdown timed out")
		return ctx.Err()
	}
}

// HealthStatus represents the health status of the router
type HealthStatus struct {
	Status    string            `json:"status"`
	Runtimes  map[string]bool   `json:"runtimes"`
	PoolReady bool              `json:"pool_ready"`
	Version   string            `json:"version"`
}

// HealthCheck returns the health status of the router
func (r *RuntimeRouter) HealthCheck() *HealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	runtimes := make(map[string]bool)
	for rt := range r.runtimes {
		runtimes[string(rt)] = true
	}

	return &HealthStatus{
		Status:    "healthy",
		Runtimes:  runtimes,
		PoolReady: r.pool != nil,
		Version:   "1.0.0",
	}
}

// IsHealthy returns true if the router is healthy
func (r *RuntimeRouter) IsHealthy() bool {
	health := r.HealthCheck()
	return health.Status == "healthy"
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
	logrus.Info("Initialized default runtime router")
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

// ExecutionConfig defines per-execution configuration options
type ExecutionConfig struct {
	// Timeout overrides the default execution timeout (uses runtime default if 0)
	Timeout time.Duration

	// MaxInputSize overrides the default max input size (uses runtime default if 0)
	MaxInputSize uint32

	// EnvironmentVars allows setting per-execution environment variables
	EnvironmentVars map[string]string
}

// ValidateInputSize checks if the input size is within configured limits
func (c *ExecutionConfig) ValidateInputSize(size uint32, defaultMax uint32) bool {
	if c == nil || c.MaxInputSize == 0 {
		return size <= defaultMax
	}
	return size <= c.MaxInputSize
}

// GetTimeout returns the timeout to use, falling back to default
func (c *ExecutionConfig) GetTimeout(defaultTimeout time.Duration) time.Duration {
	if c == nil || c.Timeout == 0 {
		return defaultTimeout
	}
	return c.Timeout
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

// ExecuteWithConfig executes with custom configuration
func (p *WASMRuntimeProvider) ExecuteWithConfig(ctx context.Context, input []byte, config interface{}) ([]byte, error) {
	var execConfig *ExecutionConfig
	if cfg, ok := config.(*ExecutionConfig); ok {
		execConfig = cfg
	}

	// Validate input size if config specifies a limit
	if execConfig != nil && execConfig.MaxInputSize > 0 {
		if uint32(len(input)) > execConfig.MaxInputSize {
			return nil, fmt.Errorf("input size exceeds configured limit: %d > %d bytes", len(input), execConfig.MaxInputSize)
		}
	} else if p.runtime.config != nil {
		if !p.runtime.config.ValidateInputSize(uint32(len(input))) {
			return nil, fmt.Errorf("input size exceeds maximum allowed: %d > %d bytes", len(input), p.runtime.config.MaxInputSize)
		}
	}

	// Create context with custom timeout if specified
	timeout := 30 * time.Second
	if p.runtime.config != nil {
		timeout = p.runtime.config.MaxExecutionTime
	}
	if execConfig != nil && execConfig.Timeout > 0 {
		timeout = execConfig.Timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return p.runtime.ExecuteWithContext(execCtx, input)
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

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for runtime failures
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failureThreshold int
	resetTimeout     time.Duration
	failureCount     int
	lastFailure      time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

// Allow checks if a request should be allowed through
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			logrus.Info("Circuit breaker entering half-open state")
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		logrus.Info("Circuit breaker closed after successful execution")
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		logrus.Warn("Circuit breaker opened after failure in half-open state")
		return
	}

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
		logrus.WithField("failures", cb.failureCount).Warn("Circuit breaker opened")
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// RuntimeRouterWithCircuitBreaker wraps RuntimeRouter with circuit breaker protection
type RuntimeRouterWithCircuitBreaker struct {
	router    *RuntimeRouter
	breakers  map[RuntimeType]*CircuitBreaker
	mu        sync.RWMutex
	threshold int
	timeout   time.Duration
}

// NewRuntimeRouterWithCircuitBreaker creates a new router with circuit breakers
func NewRuntimeRouterWithCircuitBreaker(router *RuntimeRouter, failureThreshold int, resetTimeout time.Duration) *RuntimeRouterWithCircuitBreaker {
	return &RuntimeRouterWithCircuitBreaker{
		router:    router,
		breakers:  make(map[RuntimeType]*CircuitBreaker),
		threshold: failureThreshold,
		timeout:   resetTimeout,
	}
}

// getBreaker returns the circuit breaker for a runtime, creating one if needed
func (r *RuntimeRouterWithCircuitBreaker) getBreaker(runtimeType RuntimeType) *CircuitBreaker {
	r.mu.RLock()
	breaker, exists := r.breakers[runtimeType]
	r.mu.RUnlock()

	if exists {
		return breaker
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if breaker, exists = r.breakers[runtimeType]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(r.threshold, r.timeout)
	r.breakers[runtimeType] = breaker
	return breaker
}

// ExecuteWithCircuitBreaker executes with circuit breaker protection
func (r *RuntimeRouterWithCircuitBreaker) ExecuteWithCircuitBreaker(ctx context.Context, tenantID string, runtimeType RuntimeType, input []byte) ([]byte, error) {
	breaker := r.getBreaker(runtimeType)

	if !breaker.Allow() {
		logrus.WithField("runtime", runtimeType).Warn("Circuit breaker rejected request")
		return nil, fmt.Errorf("circuit breaker open for runtime %s", runtimeType)
	}

	output, err := r.router.ExecuteWithPool(ctx, tenantID, runtimeType, input)

	if err != nil {
		breaker.RecordFailure()
		logrus.WithError(err).WithField("runtime", runtimeType).Warn("Execution failed")
	} else {
		breaker.RecordSuccess()
	}

	return output, err
}

// GetBreakerState returns the state of a circuit breaker for a runtime
func (r *RuntimeRouterWithCircuitBreaker) GetBreakerState(runtimeType RuntimeType) CircuitBreakerState {
	breaker := r.getBreaker(runtimeType)
	return breaker.State()
}
