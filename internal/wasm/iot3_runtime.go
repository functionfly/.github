//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains the WASM3 IoT runtime implementation for constrained environments
package wasm

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
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
		TargetLatency:    500, // 500ms target
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
	mu       sync.RWMutex
	instances chan *IoTInstance
	maxSize  int
	active   int
	config   *IoTConfig
}

// NewIoTInstancePool creates a new IoT instance pool
func NewIoTInstancePool(config *IoTConfig) *IoTInstancePool {
	if config == nil {
		config = DefaultIoTConfig()
	}
	pool := &IoTInstancePool{
		instances: make(chan *IoTInstance, config.MaxInstances),
		maxSize:  config.MaxInstances,
		config:   config,
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

// executeInstance executes the WASM code on a single instance
func (r *WASM3IoTRuntime) executeInstance(ctx context.Context, inst *IoTInstance, input []byte) ([]byte, error) {
	// For IoT WASM3, we simulate lightweight execution
	// In production, this would use the actual WASM3 C library via CGO

	// Simple processing: echo with IoT prefix and timestamp
	// Real implementation would execute actual WASM bytecode

	output := make([]byte, 0, len(input)+32)
	output = append(output, []byte("[WASM3-IoT]")...)

	// Add timestamp
	ts := time.Now().UnixMilli()
	output = append(output, []byte(fmt.Sprintf("%d:", ts))...)

	// Echo input (in real impl, this would be WASM execution)
	output = append(output, input...)

	// Simulate minimal processing delay for lightweight execution
	// Real WASM3 has ~50-100ms cold start vs wasmtime's ~500ms
	time.Sleep(5 * time.Millisecond) // Simulated minimal overhead

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
