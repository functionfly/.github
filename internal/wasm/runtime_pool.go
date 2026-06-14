//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
package wasm

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PythonRuntimePool manages warm PythonRuntime instances keyed by wasmPath.
// Each wasmPath gets its own pool of pre-initialized instances, eliminating
// the ~50ms Python interpreter cold-start overhead on hot paths.
type PythonRuntimePool struct {
	mu       sync.RWMutex
	pools    map[string]*pathPool
	factory  PythonRuntimeFactory
	maxSize  int

	// Global metrics (aggregated across all path pools)
	metrics *PoolMetrics
}

// pathPool is a pool of PythonRuntime instances for a single wasmPath.
type pathPool struct {
	mu          sync.Mutex
	instances   chan *PythonRuntime
	maxSize     int
	activeCount int
}

// PythonRuntimeFactory creates a new PythonRuntime instance.
type PythonRuntimeFactory func(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error)

// NewPythonRuntimePool creates a new Python runtime pool.
func NewPythonRuntimePool(factory PythonRuntimeFactory, maxSize int) *PythonRuntimePool {
	if maxSize <= 0 {
		maxSize = DefaultPoolSize
	}
	return &PythonRuntimePool{
		pools:   make(map[string]*pathPool),
		factory: factory,
		maxSize: maxSize,
		metrics: NewPoolMetrics(),
	}
}

// getOrCreatePool returns the pathPool for the given wasmPath, creating it if needed.
func (p *PythonRuntimePool) getOrCreatePool(wasmPath string) *pathPool {
	p.mu.RLock()
	pp, exists := p.pools[wasmPath]
	p.mu.RUnlock()
	if exists {
		return pp
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	// Double-check after acquiring write lock
	if pp, exists = p.pools[wasmPath]; exists {
		return pp
	}

	pp = &pathPool{
		instances: make(chan *PythonRuntime, p.maxSize),
		maxSize:   p.maxSize,
	}
	p.pools[wasmPath] = pp
	return pp
}

// Prewarm creates warm instances for the given wasmPath.
func (p *PythonRuntimePool) Prewarm(ctx context.Context, wasmPath string, count int) error {
	if count <= 0 {
		count = p.maxSize
	}

	pp := p.getOrCreatePool(wasmPath)

	var wg sync.WaitGroup
	errChan := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rt, err := p.factory(wasmPath, io.Discard, io.Discard, NewDefaultHostHandler(nil))
			if err != nil {
				errChan <- err
				return
			}
			if err := rt.Init(); err != nil {
				rt.Close()
				errChan <- fmt.Errorf("runtime init failed: %w", err)
				return
			}
			pp.Put(rt)
		}()
	}

	wg.Wait()
	close(errChan)

	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	logrus.WithFields(logrus.Fields{
		"wasm_path": wasmPath,
		"count":     count,
	}).Info("PythonRuntimePool prewarmed")
	return firstErr
}

// Put adds an instance back to its path pool.
func (pp *pathPool) Put(rt *PythonRuntime) {
	if rt == nil {
		return
	}
	select {
	case pp.instances <- rt:
	default:
		pp.mu.Lock()
		pp.activeCount--
		pp.mu.Unlock()
		rt.Close()
	}
}

// Get retrieves a warm PythonRuntime from the pool, or creates a new one.
func (p *PythonRuntimePool) Get(ctx context.Context, wasmPath string) (*PythonRuntime, error) {
	pp := p.getOrCreatePool(wasmPath)

	// Try to get from pool
	select {
	case rt := <-pp.instances:
		p.metrics.RecordHit()
		pp.mu.Lock()
		pp.mu.Unlock()
		return rt, nil
	default:
		p.metrics.RecordMiss()
	}

	// Pool empty — create new instance
	pp.mu.Lock()
	if pp.activeCount >= pp.maxSize {
		pp.mu.Unlock()
		select {
		case rt := <-pp.instances:
			return rt, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("timeout waiting for available PythonRuntime for %s", wasmPath)
		}
	}
	pp.activeCount++
	pp.mu.Unlock()

	p.metrics.RecordColdStart()

	rt, err := p.factory(wasmPath, io.Discard, io.Discard, NewDefaultHostHandler(nil))
	if err != nil {
		pp.mu.Lock()
		pp.activeCount--
		pp.mu.Unlock()
		return nil, fmt.Errorf("failed to create PythonRuntime: %w", err)
	}

	if err := rt.Init(); err != nil {
		rt.Close()
		pp.mu.Lock()
		pp.activeCount--
		pp.mu.Unlock()
		return nil, fmt.Errorf("failed to init PythonRuntime: %w", err)
	}

	return rt, nil
}

// Put returns a PythonRuntime to the pool for reuse.
func (p *PythonRuntimePool) Put(rt *PythonRuntime, wasmPath string) error {
	if rt == nil {
		return nil
	}
	pp := p.getOrCreatePool(wasmPath)
	pp.Put(rt)
	return nil
}

// Close shuts down all pools and closes all runtimes.
func (p *PythonRuntimePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for path, pp := range p.pools {
		close(pp.instances)
		for rt := range pp.instances {
			rt.Close()
		}
		delete(p.pools, path)
	}
	return nil
}

// Stats returns pool statistics for each path and globally.
func (p *PythonRuntimePool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	paths := make(map[string]interface{})
	totalActive := 0
	totalIdle := 0

	for path, pp := range p.pools {
		pp.mu.Lock()
		active := pp.activeCount
		idle := len(pp.instances)
		pp.mu.Unlock()

		paths[path] = map[string]interface{}{
			"active_count": active,
			"idle_count":   idle,
			"max_size":     pp.maxSize,
		}
		totalActive += active
		totalIdle += idle
	}

	return map[string]interface{}{
		"path_pools": paths,
		"total_active": totalActive,
		"total_idle":   totalIdle,
		"total_paths":  len(p.pools),
		"metrics": map[string]interface{}{
			"hits":        p.metrics.GetHits(),
			"misses":      p.metrics.GetMisses(),
			"cold_starts": p.metrics.GetColdStarts(),
			"hit_rate":    p.metrics.HitRate(),
		},
	}
}

// ---------------------------------------------------------------------------
// Global PythonRuntimePool for sandbox execution
// ---------------------------------------------------------------------------

var (
	// GlobalPythonRuntimePool is the shared pool for Python runtime reuse.
	// Initialized via InitPythonRuntimePool().
	GlobalPythonRuntimePool *PythonRuntimePool
	pythonPoolInit          sync.Once
)

// InitPythonRuntimePool initializes the global Python runtime pool.
func InitPythonRuntimePool(factory PythonRuntimeFactory, poolSize int) {
	pythonPoolInit.Do(func() {
		GlobalPythonRuntimePool = NewPythonRuntimePool(factory, poolSize)
		logrus.Printf("[WASM] PythonRuntimePool initialized with size %d", poolSize)
	})
}

// InitPythonRuntimePoolWithPrewarm initializes the pool and prewarms it.
func InitPythonRuntimePoolWithPrewarm(ctx context.Context, factory PythonRuntimeFactory, poolSize int, wasmPath string, prewarmCount int) error {
	InitPythonRuntimePool(factory, poolSize)
	if GlobalPythonRuntimePool == nil {
		return fmt.Errorf("failed to initialize PythonRuntimePool")
	}
	return GlobalPythonRuntimePool.Prewarm(ctx, wasmPath, prewarmCount)
}
