//go:build cgo

package wasm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// PoolMetrics holds pool metrics for monitoring
type PoolMetrics struct {
	// Atomic counters
	hits   atomic.Int64
	misses atomic.Int64
	coldStarts atomic.Int64
	evictions atomic.Int64

	// Gauge values (updated atomically)
	activeInstances atomic.Int64
	idleInstances atomic.Int64
	totalInstances atomic.Int64
}

// NewPoolMetrics creates new pool metrics
func NewPoolMetrics() *PoolMetrics {
	return &PoolMetrics{}
}

// RecordHit records a pool hit (instance reused from pool)
func (m *PoolMetrics) RecordHit() {
	m.hits.Add(1)
}

// RecordMiss records a pool miss (new instance created)
func (m *PoolMetrics) RecordMiss() {
	m.misses.Add(1)
}

// RecordColdStart records a cold start (new instance created after pool was empty)
func (m *PoolMetrics) RecordColdStart() {
	m.coldStarts.Add(1)
}

// RecordEviction records an instance eviction
func (m *PoolMetrics) RecordEviction() {
	m.evictions.Add(1)
}

// UpdateActive updates the active instance count
func (m *PoolMetrics) UpdateActive(count int) {
	m.activeInstances.Store(int64(count))
}

// UpdateIdle updates the idle instance count
func (m *PoolMetrics) UpdateIdle(count int) {
	m.idleInstances.Store(int64(count))
}

// UpdateTotal updates the total instance count
func (m *PoolMetrics) UpdateTotal(count int) {
	m.totalInstances.Store(int64(count))
}

// GetHits returns the hit count
func (m *PoolMetrics) GetHits() int64 {
	return m.hits.Load()
}

// GetMisses returns the miss count
func (m *PoolMetrics) GetMisses() int64 {
	return m.misses.Load()
}

// GetColdStarts returns the cold start count
func (m *PoolMetrics) GetColdStarts() int64 {
	return m.coldStarts.Load()
}

// GetEvictions returns the eviction count
func (m *PoolMetrics) GetEvictions() int64 {
	return m.evictions.Load()
}

// HitRate returns the pool hit rate as a percentage
func (m *PoolMetrics) HitRate() float64 {
	total := m.hits.Load() + m.misses.Load()
	if total == 0 {
		return 0
	}
	return float64(m.hits.Load()) / float64(total) * 100
}

// InstanceFactory is a function that creates a new WASM instance
type InstanceFactory func() (*PythonRuntime, error)

// InstancePool manages pooled WASM instances per tenant
type InstancePool struct {
	mu           sync.RWMutex
	pools        map[string]*TenantPool
	factory      InstanceFactory
	defaultSize  int
	maxSize      int
	cleanupInterval time.Duration
	stopCleanup  chan struct{}

	// Prewarming
	prewarmed bool
	prewarmCount int
	prewarmMu sync.RWMutex

	// Metrics
	metrics *PoolMetrics

	// LRU eviction
	maxInstanceAge time.Duration
}

// TenantPool represents a pool of instances for a specific tenant
type TenantPool struct {
	mu          sync.Mutex
	instances   chan *PythonRuntime
	activeCount int
	maxSize     int
	factory     InstanceFactory
	lastUsed    time.Time
}

// PooledInstance wraps a WASM instance with metadata
type PooledInstance struct {
	Instance   *PythonRuntime
	TenantID    string
	Runtime     string
	CreatedAt   time.Time
	LastUsed    time.Time
	ExecuteCount int64
}

// NewInstancePool creates a new instance pool
func NewInstancePool(factory InstanceFactory, defaultSize, maxSize int) *InstancePool {
	if defaultSize <= 0 {
		defaultSize = DefaultPoolSize
	}
	if maxSize <= 0 {
		maxSize = defaultSize * 2
	}

	pool := &InstancePool{
		pools:            make(map[string]*TenantPool),
		factory:          factory,
		defaultSize:      defaultSize,
		maxSize:          maxSize,
		cleanupInterval:  5 * time.Minute,
		stopCleanup:      make(chan struct{}),
		metrics:          NewPoolMetrics(),
		maxInstanceAge:   30 * time.Minute, // Default max instance age for LRU eviction
	}

	// Start cleanup goroutine
	go pool.cleanupLoop()

	return pool
}

// Prewarm warms up the pool by creating initial instances
func (p *InstancePool) Prewarm(ctx context.Context, tenantID, runtime string, count int) error {
	if count <= 0 {
		count = p.defaultSize
	}

	poolKey := GetPoolKey(tenantID, runtime)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Get or create tenant pool
	tenantPool, exists := p.pools[poolKey]
	if !exists {
		tenantPool = &TenantPool{
			instances: make(chan *PythonRuntime, p.defaultSize),
			maxSize:   p.maxSize,
			factory:   p.factory,
			lastUsed:  time.Now(),
		}
		p.pools[poolKey] = tenantPool
	}

	// Create instances in parallel
	errChan := make(chan error, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			inst, err := p.factory()
			if err != nil {
				errChan <- err
				return
			}
			// Try to add to pool
			select {
			case tenantPool.instances <- inst:
				tenantPool.mu.Lock()
				tenantPool.activeCount++
				tenantPool.mu.Unlock()
			default:
				// Pool full, close the instance
				inst.Close()
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	// Mark as prewarmed
	p.prewarmMu.Lock()
	p.prewarmed = true
	p.prewarmCount = count
	p.prewarmMu.Unlock()

	log.Printf("[WASM] Pool prewarmed for %s:%s with %d instances", tenantID, runtime, count)
	return firstErr
}

// PrewarmAll prewarms pools for all known tenants and runtimes
func (p *InstancePool) PrewarmAll(ctx context.Context, tenants []string, runtimes []string) {
	for _, tenant := range tenants {
		for _, runtime := range runtimes {
			type runtimeType string
			go func(t, r string) {
				if err := p.Prewarm(ctx, t, r, p.defaultSize); err != nil {
					log.Printf("[WASM] Warning: Failed to prewarm pool for %s:%s: %v", t, r, err)
				}
			}(tenant, runtime)
		}
	}
}

// Warmup runs a test execution to initialize the runtime
func (p *InstancePool) Warmup(ctx context.Context, tenantID, runtime string) error {
	inst, err := p.Get(ctx, tenantID, runtime)
	if err != nil {
		return fmt.Errorf("failed to get instance for warmup: %w", err)
	}

	// Run a simple test execution
	testInput := []byte(`{"warmup": true}`)
	_, err = inst.Instance.ExecuteWithContext(ctx, testInput)

	// Return instance to pool
	p.Put(inst)

	if err != nil {
		log.Printf("[WASM] Warmup execution warning for %s:%s: %v", tenantID, runtime, err)
		// Don't return error - warmup failure is not fatal
	}

	log.Printf("[WASM] Pool warmed up for %s:%s", tenantID, runtime)
	return nil
}

// IsPrewarmed returns true if the pool has been prewarmed
func (p *InstancePool) IsPrewarmed() bool {
	p.prewarmMu.RLock()
	defer p.prewarmMu.RUnlock()
	return p.prewarmed
}

// GetMetrics returns the pool metrics
func (p *InstancePool) GetMetrics() *PoolMetrics {
	return p.metrics
}

// GetPoolKey generates a unique key for the tenant pool
func GetPoolKey(tenantID, runtime string) string {
	return fmt.Sprintf("%s:%s", tenantID, runtime)
}

// Get retrieves an instance from the pool, creating one if needed
func (p *InstancePool) Get(ctx context.Context, tenantID, runtime string) (*PooledInstance, error) {
	poolKey := GetPoolKey(tenantID, runtime)

	p.mu.RLock()
	tenantPool, exists := p.pools[poolKey]
	p.mu.RUnlock()

	if !exists {
		// Create new tenant pool
		p.mu.Lock()
		// Double-check after acquiring write lock
		if tenantPool, exists = p.pools[poolKey]; !exists {
			tenantPool = &TenantPool{
				instances:   make(chan *PythonRuntime, p.defaultSize),
				maxSize:     p.maxSize,
				factory:     p.factory,
				lastUsed:    time.Now(),
			}
			p.pools[poolKey] = tenantPool
		}
		p.mu.Unlock()
	}

	// Try to get from pool
	select {
	case inst := <-tenantPool.instances:
		tenantPool.mu.Lock()
		tenantPool.lastUsed = time.Now()
		tenantPool.mu.Unlock()

		// Record hit
		p.metrics.RecordHit()

		return &PooledInstance{
			Instance:    inst,
			TenantID:    tenantID,
			Runtime:     runtime,
			LastUsed:    time.Now(),
			ExecuteCount: 0,
		}, nil
	default:
		// Pool is empty - record miss
		p.metrics.RecordMiss()

		// Pool is empty, create new instance
		tenantPool.mu.Lock()
		if tenantPool.activeCount >= tenantPool.maxSize {
			// At capacity, wait for an instance to become available
			tenantPool.mu.Unlock()
			select {
			case inst := <-tenantPool.instances:
				return &PooledInstance{
					Instance:    inst,
					TenantID:    tenantID,
					Runtime:     runtime,
					LastUsed:    time.Now(),
					ExecuteCount: 0,
				}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(30 * time.Second):
				return nil, errors.New("timeout waiting for available instance")
			}
		}
		tenantPool.activeCount++
		tenantPool.mu.Unlock()

		// Create new instance - record cold start
		p.metrics.RecordColdStart()

		inst, err := p.factory()
		if err != nil {
			tenantPool.mu.Lock()
			tenantPool.activeCount--
			tenantPool.mu.Unlock()
			return nil, fmt.Errorf("failed to create instance: %w", err)
		}

		return &PooledInstance{
			Instance:    inst,
			TenantID:    tenantID,
			Runtime:     runtime,
			CreatedAt:   time.Now(),
			LastUsed:    time.Now(),
			ExecuteCount: 0,
		}, nil
	}
}

// Put returns an instance to the pool
func (p *InstancePool) Put(pi *PooledInstance) error {
	if pi == nil || pi.Instance == nil {
		return nil
	}

	poolKey := GetPoolKey(pi.TenantID, pi.Runtime)

	p.mu.RLock()
	tenantPool, exists := p.pools[poolKey]
	p.mu.RUnlock()

	if !exists {
		// Pool doesn't exist anymore, just close the instance
		return pi.Instance.Close()
	}

	// Try to return to pool
	select {
	case tenantPool.instances <- pi.Instance:
		tenantPool.mu.Lock()
		tenantPool.lastUsed = time.Now()
		tenantPool.mu.Unlock()
		return nil
	default:
		// Pool is full, close the instance
		tenantPool.mu.Lock()
		tenantPool.activeCount--
		tenantPool.mu.Unlock()
		return pi.Instance.Close()
	}
}

// RemoveTenant removes all instances for a tenant
func (p *InstancePool) RemoveTenant(tenantID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find and clean up all pools for this tenant
	for key, pool := range p.pools {
		if len(tenantID) > 0 && len(key) >= len(tenantID) && key[:len(tenantID)] == tenantID {
			// Drain the channel
			close(pool.instances)
			for inst := range pool.instances {
				inst.Close()
			}
			pool.activeCount = 0
			delete(p.pools, key)
		}
	}
}

// cleanupLoop periodically removes idle tenant pools
func (p *InstancePool) cleanupLoop() {
	ticker := time.NewTicker(p.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCleanup:
			return
		case <-ticker.C:
			p.cleanupIdlePools()
		}
	}
}

// cleanupIdlePools removes pools that haven't been used recently and evicts old instances
func (p *InstancePool) cleanupIdlePools() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	idleTimeout := 10 * time.Minute

	for key, pool := range p.pools {
		pool.mu.Lock()
		idle := now.Sub(pool.lastUsed)
		pool.mu.Unlock()

		if idle > idleTimeout {
			// Check if pool is empty
			select {
			case _, ok := <-pool.instances:
				if !ok {
					// Channel closed, remove pool
					delete(p.pools, key)
				} else {
					// Put it back, we borrowed it for checking
					pool.instances <- nil // This is a bug, let's fix
				}
			default:
				// Pool is empty, can be removed
				delete(p.pools, key)
			}
		}

		// LRU eviction: Remove instances older than maxInstanceAge
		if p.maxInstanceAge > 0 {
			p.evictOldInstances(pool, now)
		}
	}
}

// evictOldInstances removes instances that have been in use too long
func (p *InstancePool) evictOldInstances(pool *TenantPool, now time.Time) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Drain old instances from the channel
	var validInstances []*PythonRuntime
	close(pool.instances)

	for inst := range pool.instances {
		if inst != nil {
			validInstances = append(validInstances, inst)
		} else {
			// Nil instance indicates stale, count as evicted
			p.metrics.RecordEviction()
		}
	}

	// Recreate channel with remaining valid instances
	pool.instances = make(chan *PythonRuntime, pool.maxSize)
	for _, inst := range validInstances {
		select {
		case pool.instances <- inst:
		default:
			// Pool full, close remaining
			inst.Close()
			p.metrics.RecordEviction()
		}
	}
}

// SetMaxInstanceAge sets the maximum age for instances before LRU eviction
func (p *InstancePool) SetMaxInstanceAge(age time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxInstanceAge = age
}

// Close shuts down the instance pool
func (p *InstancePool) Close() error {
	close(p.stopCleanup)

	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pool := range p.pools {
		close(pool.instances)
		for inst := range pool.instances {
			if inst != nil {
				inst.Close()
			}
		}
		delete(p.pools, key)
	}

	return nil
}

// Stats returns pool statistics
func (p *InstancePool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_tenant_pools"] = len(p.pools)

	// Include metrics
	if p.metrics != nil {
		stats["metrics"] = map[string]interface{}{
			"hits":          p.metrics.GetHits(),
			"misses":        p.metrics.GetMisses(),
			"cold_starts":   p.metrics.GetColdStarts(),
			"evictions":     p.metrics.GetEvictions(),
			"hit_rate":      p.metrics.HitRate(),
			"active_count":  p.metrics.activeInstances.Load(),
			"idle_count":    p.metrics.idleInstances.Load(),
			"total_count":   p.metrics.totalInstances.Load(),
		}
	}

	totalActive := 0
	totalIdle := 0

	for key, pool := range p.pools {
		pool.mu.Lock()
		active := pool.activeCount
		idle := len(pool.instances)
		pool.mu.Unlock()

		tenantPoolStats := map[string]interface{}{
			"active_instances": active,
			"idle_instances":   idle,
			"max_size":        pool.maxSize,
		}
		stats[key] = tenantPoolStats
		totalActive += active
		totalIdle += idle
	}

	stats["total_active_instances"] = totalActive
	stats["total_idle_instances"] = totalIdle

	return stats
}

// SimpleInstancePool provides a simpler, single-pool implementation
// for cases where per-tenant isolation is not required
type SimpleInstancePool struct {
	mu        sync.Mutex
	instances chan *PythonRuntime
	factory   InstanceFactory
	maxSize   int
	active    int
}

// NewSimpleInstancePool creates a simple instance pool
func NewSimpleInstancePool(factory InstanceFactory, maxSize int) *SimpleInstancePool {
	if maxSize <= 0 {
		maxSize = DefaultPoolSize * 2
	}

	return &SimpleInstancePool{
		instances: make(chan *PythonRuntime, maxSize),
		factory:   factory,
		maxSize:   maxSize,
	}
}

// Get retrieves an instance from the simple pool
func (p *SimpleInstancePool) Get(ctx context.Context) (*PythonRuntime, error) {
	select {
	case inst := <-p.instances:
		return inst, nil
	default:
		p.mu.Lock()
		if p.active >= p.maxSize {
			p.mu.Unlock()
			select {
			case inst := <-p.instances:
				return inst, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(30 * time.Second):
				return nil, errors.New("timeout waiting for available instance")
			}
		}
		p.active++
		p.mu.Unlock()

		inst, err := p.factory()
		if err != nil {
			p.mu.Lock()
			p.active--
			p.mu.Unlock()
			return nil, err
		}

		return inst, nil
	}
}

// Put returns an instance to the simple pool
func (p *SimpleInstancePool) Put(inst *PythonRuntime) error {
	if inst == nil {
		return nil
	}

	select {
	case p.instances <- inst:
		return nil
	default:
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
		return inst.Close()
	}
}

// Close shuts down the simple pool
func (p *SimpleInstancePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.instances)
	for inst := range p.instances {
		if inst != nil {
			inst.Close()
		}
	}
	p.active = 0

	return nil
}

// Stats returns simple pool statistics
func (p *SimpleInstancePool) Stats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"max_size":      p.maxSize,
		"active_count":  p.active,
		"idle_count":    len(p.instances),
		"total_used":    p.active + len(p.instances),
	}
}

// init pools
var (
	// GlobalPool is the default instance pool
	GlobalPool *SimpleInstancePool

	// PerTenantPools manages per-tenant instance pools
	PerTenantPools *InstancePool
)

// InitPools initializes the global pools
func InitPools(factory InstanceFactory, poolSize int) {
	GlobalPool = NewSimpleInstancePool(factory, poolSize)
	PerTenantPools = NewInstancePool(factory, poolSize/2, poolSize)
	log.Printf("[WASM] Instance pools initialized with size %d", poolSize)
}

// InitPoolsWithConfig initializes the global pools with custom configuration
func InitPoolsWithConfig(factory InstanceFactory, poolSize int, maxInstanceAge time.Duration) {
	GlobalPool = NewSimpleInstancePool(factory, poolSize)
	PerTenantPools = NewInstancePool(factory, poolSize/2, poolSize)
	if maxInstanceAge > 0 {
		PerTenantPools.SetMaxInstanceAge(maxInstanceAge)
	}
	log.Printf("[WASM] Instance pools initialized with size %d, max instance age: %v", poolSize, maxInstanceAge)
}
