//go:build !cgo

// Package wasm: stub types for InstancePool, SimpleInstancePool, DeterministicConfig,
// DeterministicResult, and DeterministicExecutor when CGO is disabled (e.g. Docker Alpine build).
// Full implementations are in pool.go and deterministic.go (build tag: cgo).
package wasm

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var errWasmNotAvailable = errors.New("WASM runtime not available: CGO disabled")

// PooledInstance stub (returned by pool.Get in cgo build; stub Get returns error).
type PooledInstance struct {
	Instance    *PythonRuntime
	TenantID    string
	Runtime     string
	CreatedAt   time.Time
	LastUsed    time.Time
	ExecuteCount int64
}

// InstancePool stub when built without CGO (no-op pool).
type InstancePool struct {
	mu      sync.RWMutex
	metrics *PoolMetrics
}

// PoolMetrics stub: methods return zero so metrics.Collect is safe when pool is nil.
type PoolMetrics struct{}

func (m *PoolMetrics) GetHits() int64        { return 0 }
func (m *PoolMetrics) GetMisses() int64      { return 0 }
func (m *PoolMetrics) GetColdStarts() int64  { return 0 }
func (m *PoolMetrics) GetEvictions() int64   { return 0 }

// InstanceFactory is a function that creates a new WASM instance (stub type for !cgo).
type InstanceFactory func() (*PythonRuntime, error)

// NewInstancePool returns a stub pool when CGO is disabled.
func NewInstancePool(factory InstanceFactory, defaultSize, maxSize int) *InstancePool {
	return &InstancePool{metrics: &PoolMetrics{}}
}

// Get returns an error when CGO is disabled (no pooled instances).
func (p *InstancePool) Get(ctx context.Context, tenantID, runtime string) (*PooledInstance, error) {
	return nil, errWasmNotAvailable
}

// Put is a no-op for the stub pool.
func (p *InstancePool) Put(pi *PooledInstance) error {
	return nil
}

// Close is a no-op for the stub pool.
func (p *InstancePool) Close() error {
	return nil
}

// SimpleInstancePool stub when built without CGO.
type SimpleInstancePool struct {
	mu sync.Mutex
}

// NewSimpleInstancePool returns a stub simple pool when CGO is disabled.
func NewSimpleInstancePool(factory InstanceFactory, maxSize int) *SimpleInstancePool {
	return &SimpleInstancePool{}
}

// Close is a no-op for the stub simple pool.
func (p *SimpleInstancePool) Close() error {
	return nil
}

// DeterministicConfig stub (same shape as cgo version).
type DeterministicConfig struct {
	FixedTimeWindow        time.Duration
	MaxInstructions        uint64
	DeterministicRandom    bool
	RandomSeed             uint64
	NormalizeMemoryAccess  bool
	ConstantTimeExecution bool
}

// DefaultDeterministicConfig returns default config when CGO is disabled.
func DefaultDeterministicConfig() *DeterministicConfig {
	return &DeterministicConfig{
		FixedTimeWindow:        100 * time.Millisecond,
		MaxInstructions:        DefaultMaxInstructions,
		DeterministicRandom:    false,
		RandomSeed:             0,
		NormalizeMemoryAccess:  true,
		ConstantTimeExecution: true,
	}
}

// DeterministicResult stub (same shape as cgo version).
type DeterministicResult struct {
	Output           []byte
	ExecutionTime    time.Duration
	InstructionsUsed uint64
	DeterministicID  string
	Status           string
}

// DeterministicExecutor stub: Execute returns errWasmNotAvailable when CGO is disabled.
type DeterministicExecutor struct {
	pool   *InstancePool
	config *DeterministicConfig
}

// NewDeterministicExecutor returns a stub executor when CGO is disabled.
func NewDeterministicExecutor(pool *InstancePool, config *DeterministicConfig) *DeterministicExecutor {
	if config == nil {
		config = DefaultDeterministicConfig()
	}
	return &DeterministicExecutor{pool: pool, config: config}
}

// Execute returns an error when CGO is disabled (WASM execution not available).
func (e *DeterministicExecutor) Execute(ctx context.Context, tenantID, functionID, executionID string, runtimeType string, input []byte) (*DeterministicResult, error) {
	return nil, errWasmNotAvailable
}

// PerTenantPools stub - not available without CGO
var PerTenantPools *InstancePool

// InitPoolsWithConfig is a no-op stub for non-CGO builds.
func InitPoolsWithConfig(factory InstanceFactory, poolSize int, maxInstanceAge time.Duration) {
	// No-op: WASM pools not available without CGO
	logrus.Warn("InitPoolsWithConfig: WASM pools not available without CGO")
}
