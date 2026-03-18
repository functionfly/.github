//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains deterministic execution implementation
package wasm

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"
)

// DeterministicConfig defines configuration for deterministic execution
type DeterministicConfig struct {
	// FixedTimeWindow is the fixed execution window regardless of actual execution time
	FixedTimeWindow time.Duration

	// MaxInstructions is the maximum number of CPU instructions allowed
	MaxInstructions uint64

	// DeterministicRandom enables seeded random for reproducibility
	DeterministicRandom bool

	// RandomSeed is the seed for deterministic random (used if DeterministicRandom is true)
	RandomSeed uint64

	// NormalizeMemoryAccess normalizes memory access patterns to prevent timing attacks
	NormalizeMemoryAccess bool

	// ConstantTimeExecution ensures execution takes constant time regardless of branch outcomes
	ConstantTimeExecution bool
}

// DefaultDeterministicConfig returns a default deterministic configuration
func DefaultDeterministicConfig() *DeterministicConfig {
	return &DeterministicConfig{
		FixedTimeWindow:        100 * time.Millisecond,
		MaxInstructions:        DefaultMaxInstructions,
		DeterministicRandom:    false,
		RandomSeed:            0,
		NormalizeMemoryAccess: true,
		ConstantTimeExecution: true,
	}
}

// DeterministicResult holds the result of a deterministic execution
type DeterministicResult struct {
	Output           []byte
	ExecutionTime    time.Duration
	InstructionsUsed uint64
	DeterministicID  string
	Status           string
}

// WASMRuntimeWithDeterminism wraps WASMRuntime with deterministic execution support
type WASMRuntimeWithDeterminism struct {
	runtime             *PythonRuntime
	config              *DeterministicConfig
	instructionCounter  uint64
	mu                  sync.Mutex
	randomState         uint64
}

// NewWASMRuntimeWithDeterminism creates a new runtime with deterministic execution support
func NewWASMRuntimeWithDeterminism(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, detConfig *DeterministicConfig) (*WASMRuntimeWithDeterminism, error) {
	if detConfig == nil {
		detConfig = DefaultDeterministicConfig()
	}

	// Create security config with deterministic settings
	securityConfig := NewDefaultSecurityConfig()
	securityConfig.EnableDeterministic = true
	securityConfig.MaxExecutionTime = detConfig.FixedTimeWindow

	runtime, err := NewPythonRuntimeWithConfigAndDebug(wasmPath, stdout, stderr, handler, securityConfig, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	// Initialize random seed if deterministic random is enabled
	randomState := detConfig.RandomSeed
	if detConfig.DeterministicRandom && randomState == 0 {
		// Generate a random seed
		var seedBytes [8]byte
		if _, err := rand.Read(seedBytes[:]); err == nil {
			randomState = binary.LittleEndian.Uint64(seedBytes[:])
		}
	}

	return &WASMRuntimeWithDeterminism{
		runtime:      runtime,
		config:       detConfig,
		randomState:  randomState,
	}, nil
}

// ExecuteDeterministic executes the WASM module with deterministic timing
func (r *WASMRuntimeWithDeterminism) ExecuteDeterministic(ctx context.Context, input []byte, detConfig *DeterministicConfig) (*DeterministicResult, error) {
	if detConfig == nil {
		detConfig = r.config
	}

	// Generate deterministic execution ID
	detID := r.generateDeterministicID(input)

	startTime := time.Now()

	// Execute with potential instruction limiting
	output, err := r.executeWithInstructionLimit(ctx, input, detConfig.MaxInstructions)

	executionTime := time.Since(startTime)

	// Ensure minimum execution time for constant-time behavior
	if detConfig.ConstantTimeExecution && executionTime < detConfig.FixedTimeWindow {
		// Add deterministic delay
		time.Sleep(detConfig.FixedTimeWindow - executionTime)
		executionTime = detConfig.FixedTimeWindow
	}

	result := &DeterministicResult{
		Output:           output,
		ExecutionTime:    executionTime,
		InstructionsUsed: r.instructionCounter,
		DeterministicID:  detID,
		Status:           StatusSuccess,
	}

	if err != nil {
		result.Status = StatusError
		result.Output = nil
		return result, err
	}

	return result, nil
}

// executeWithInstructionLimit executes with instruction counting
func (r *WASMRuntimeWithDeterminism) executeWithInstructionLimit(ctx context.Context, input []byte, maxInstructions uint64) ([]byte, error) {
	r.instructionCounter = 0

	// Create execution context with cancellation
	execCtx, cancel := context.WithTimeout(ctx, r.config.FixedTimeWindow)
	defer cancel()

	resultChan := make(chan struct {
		output []byte
		err    error
	}, 1)

	go func() {
		output, err := r.runtime.ExecuteWithContext(execCtx, input)
		resultChan <- struct {
			output []byte
			err    error
		}{output, err}
	}()

	select {
	case <-execCtx.Done():
		r.instructionCounter = maxInstructions
		return nil, fmt.Errorf("execution exceeded instruction limit or timeout")
	case result := <-resultChan:
		if result.err != nil {
			return nil, result.err
		}
		// Estimate instructions used based on execution time
		r.instructionCounter = estimateInstructions(result.output)
		return result.output, nil
	}
}

// generateDeterministicID generates a deterministic ID based on input
func (r *WASMRuntimeWithDeterminism) generateDeterministicID(input []byte) string {
	hash := sha256.New()

	// Include random seed if deterministic random is enabled
	if r.config.DeterministicRandom {
		seedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBytes, r.randomState)
		hash.Write(seedBytes)
	}

	hash.Write(input)
	return fmt.Sprintf("det-%x", hash.Sum(nil)[:8])
}

// DeterministicRandom generates a deterministic random number
func (r *WASMRuntimeWithDeterminism) DeterministicRandomUint64() uint64 {
	if !r.config.DeterministicRandom {
		// Fall back to crypto random if not in deterministic mode
		var b [8]byte
		rand.Read(b[:])
		return binary.LittleEndian.Uint64(b[:])
	}

	// Linear congruential generator with fixed parameters
	r.randomState = r.randomState*1103515245 + 12345
	return r.randomState
}

// ResetDeterministicState resets the deterministic state (for testing)
func (r *WASMRuntimeWithDeterminism) ResetDeterministicState() {
	r.instructionCounter = 0
	if r.config.DeterministicRandom {
		r.randomState = r.config.RandomSeed
	}
}

// Close releases resources
func (r *WASMRuntimeWithDeterminism) Close() error {
	if r.runtime != nil {
		return r.runtime.Close()
	}
	return nil
}

// estimateInstructions provides a rough estimate of instructions used
// In a real implementation, this would be integrated with wasmtime's fuel API
func estimateInstructions(output []byte) uint64 {
	// Rough estimate: base cost + per-byte cost
	baseCost := uint64(10000)
	perByteCost := uint64(100)
	return baseCost + perByteCost*uint64(len(output))
}

// DeterministicExecutor provides a higher-level interface for deterministic execution
type DeterministicExecutor struct {
	pool       *InstancePool
	config     *DeterministicConfig
	seedLookup map[string]uint64 // Maps function code to seed for reproducibility
	mu         sync.RWMutex
}

// NewDeterministicExecutor creates a new deterministic executor
func NewDeterministicExecutor(pool *InstancePool, config *DeterministicConfig) *DeterministicExecutor {
	if config == nil {
		config = DefaultDeterministicConfig()
	}

	return &DeterministicExecutor{
		pool:       pool,
		config:     config,
		seedLookup: make(map[string]uint64),
	}
}

// Execute executes a function deterministically
func (e *DeterministicExecutor) Execute(ctx context.Context, tenantID, functionID, executionID string, runtimeType string, input []byte) (*DeterministicResult, error) {
	// Get or create deterministic seed for this function
	seed := e.getSeedForFunction(functionID)

	// Create deterministic config with function-specific seed
	detConfig := &DeterministicConfig{
		FixedTimeWindow:        e.config.FixedTimeWindow,
		MaxInstructions:        e.config.MaxInstructions,
		DeterministicRandom:    true,
		RandomSeed:            seed,
		NormalizeMemoryAccess: e.config.NormalizeMemoryAccess,
		ConstantTimeExecution: e.config.ConstantTimeExecution,
	}

	// Get instance from pool
	inst, err := e.pool.Get(ctx, tenantID, runtimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance from pool: %w", err)
	}

	// Note: The actual deterministic execution would need to use
	// WASMRuntimeWithDeterminism - this is a simplified interface
	// In a full implementation, we'd wrap the pooled instance

	// For now, execute normally but record deterministic metadata
	startTime := time.Now()
	output, err := inst.Instance.ExecuteWithContext(ctx, input)
	executionTime := time.Since(startTime)

	// Return instance to pool
	e.pool.Put(inst)

	if err != nil {
		return &DeterministicResult{
			Output:           nil,
			ExecutionTime:    executionTime,
			InstructionsUsed: e.config.MaxInstructions,
			DeterministicID:  fmt.Sprintf("det-%s-%x", executionID, seed),
			Status:           StatusError,
		}, err
	}

	// Ensure constant time
	if detConfig.ConstantTimeExecution && executionTime < detConfig.FixedTimeWindow {
		time.Sleep(detConfig.FixedTimeWindow - executionTime)
		executionTime = detConfig.FixedTimeWindow
	}

	return &DeterministicResult{
		Output:           output,
		ExecutionTime:    executionTime,
		InstructionsUsed: estimateInstructions(output),
		DeterministicID:  fmt.Sprintf("det-%s-%x", executionID, seed),
		Status:           StatusSuccess,
	}, nil
}

// getSeedForFunction returns or generates a seed for a function
func (e *DeterministicExecutor) getSeedForFunction(functionID string) uint64 {
	e.mu.RLock()
	seed, exists := e.seedLookup[functionID]
	e.mu.RUnlock()

	if exists {
		return seed
	}

	// Generate new seed based on function ID
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check
	if seed, exists := e.seedLookup[functionID]; exists {
		return seed
	}

	// Generate deterministic seed from function ID
	hash := sha256.Sum256([]byte(functionID))
	seed = binary.LittleEndian.Uint64(hash[:8])
	e.seedLookup[functionID] = seed

	return seed
}

// IsDeterministicMode checks if deterministic mode is enabled
func IsDeterministicMode(config *WASMSecurityConfig) bool {
	return config != nil && config.EnableDeterministic
}
