// go_pool.go implements a per-tenant pool for GoRuntime instances.
//
// Key design decisions:
//
//   - Each unique compiled WASM artifact gets its own pool bucket keyed
//     by sha256(wasmBytes). This lets us dedupe identical code across
//     tenants; tenant isolation is enforced by the host handler + per-
//     instance WorkDir.
//
//   - Each bucket is a buffered channel of GoRuntimeIfc with size from
//     GO_RUNTIME_POOL_SIZE (default 4). Get blocks on context.Done; Put
//     returns the instance to the channel or drops it if the bucket is
//     full.
//
//   - A background cleanup goroutine evicts idle instances every minute.
//     Default idle TTL is 30m (GO_RUNTIME_POOL_MAX_AGE).
package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PooledGoInstance struct {
	Instance GoRuntimeIfc
	TenantID string
	Runtime  string
	Created  time.Time
	LastUsed time.Time

	pool *GoInstancePool
	key  string
}

type GoInstancePool struct {
	mu sync.RWMutex
	// buckets keyed by sha256(wasmBytes); each bucket is a per-tenant
	// queue. Inside the bucket we further key by tenantID so that
	// identical code from two tenants doesn't share an instance.
	buckets        map[string]*goPoolBucket
	cfg            GoRuntimeConfig
	handler        HostFunctionHandler
	defaultSize    int
	maxSize        int
	maxInstanceAge time.Duration
	maxTotalBytes  uint64
	stopCleanup    chan struct{}
	closed         atomic.Bool
	metrics        goPoolMetrics
}

type goPoolBucket struct {
	mu        sync.Mutex
	perTenant map[string]chan GoRuntimeIfc
	maxSize   int
	bytes     uint64
}

type goPoolMetrics struct {
	hits      atomic.Int64
	misses    atomic.Int64
	evictions atomic.Int64
	timeouts  atomic.Int64
}

type GoPoolStats struct {
	Buckets        int    `json:"buckets"`
	TotalInstances int    `json:"total_instances"`
	TotalBytes     uint64 `json:"total_bytes"`
	Hits           int64  `json:"hits"`
	Misses         int64  `json:"misses"`
	Evictions      int64  `json:"evictions"`
	Timeouts       int64  `json:"timeouts"`
}

func NewGoInstancePool(handler HostFunctionHandler, cfg GoRuntimeConfig) *GoInstancePool {
	if cfg.MaxMemoryMB <= 0 {
		cfg = NewDefaultGoRuntimeConfig()
	}
	defaultSize := 4
	if v := os.Getenv("GO_RUNTIME_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			defaultSize = n
		}
	}
	maxSize := defaultSize * 2
	if v := os.Getenv("GO_RUNTIME_POOL_MAX_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxSize = n
		}
	}
	maxInstanceAge := 30 * time.Minute
	if v := os.Getenv("GO_RUNTIME_POOL_MAX_AGE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			maxInstanceAge = d
		}
	}
	maxTotalBytes := uint64(256) * 1024 * 1024
	if v := os.Getenv("GO_RUNTIME_POOL_MAX_BYTES"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
			maxTotalBytes = n
		}
	}

	p := &GoInstancePool{
		buckets:        make(map[string]*goPoolBucket),
		cfg:            cfg,
		handler:        handler,
		defaultSize:    defaultSize,
		maxSize:        maxSize,
		maxInstanceAge: maxInstanceAge,
		maxTotalBytes:  maxTotalBytes,
		stopCleanup:    make(chan struct{}),
	}
	go p.cleanupLoop()
	return p
}

func HashWasmBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (p *GoInstancePool) Get(ctx context.Context, tenantID, runtime string, wasmBytes []byte) (*PooledGoInstance, error) {
	if p.closed.Load() {
		return nil, ErrGoPoolClosed
	}
	if tenantID == "" {
		tenantID = "_default"
	}
	key := HashWasmBytes(wasmBytes)

	p.mu.RLock()
	bucket, ok := p.buckets[key]
	p.mu.RUnlock()

	if !ok {
		p.mu.Lock()
		bucket, ok = p.buckets[key]
		if !ok {
			bucket = &goPoolBucket{
				perTenant: make(map[string]chan GoRuntimeIfc),
				maxSize:   p.maxSize,
				bytes:     uint64(p.cfg.MaxMemoryMB) * 1024 * 1024,
			}
			p.buckets[key] = bucket
		}
		p.mu.Unlock()
	}

	bucket.mu.Lock()
	q, ok := bucket.perTenant[tenantID]
	if !ok {
		q = make(chan GoRuntimeIfc, bucket.maxSize)
		bucket.perTenant[tenantID] = q
	}
	bucket.mu.Unlock()

	select {
	case inst := <-q:
		if inst == nil {
			return nil, ErrGoPoolClosed
		}
		p.metrics.hits.Add(1)
		return &PooledGoInstance{
			Instance: inst,
			TenantID: tenantID,
			Runtime:  runtime,
			Created:  inst.CreatedAt(),
			LastUsed: time.Now(),
			pool:     p,
			key:      key,
		}, nil
	default:
	}

	if p.totalBytes()+p.bucketBytes() > p.maxTotalBytes {
		p.evictOldest()
	}

	p.metrics.misses.Add(1)
	inst, err := NewGoRuntime(wasmBytes, p.handler, p.cfg)
	if err != nil {
		return nil, err
	}
	return &PooledGoInstance{
		Instance: inst,
		TenantID: tenantID,
		Runtime:  runtime,
		Created:  inst.CreatedAt(),
		LastUsed: time.Now(),
		pool:     p,
		key:      key,
	}, nil
}

func (p *GoInstancePool) Put(pi *PooledGoInstance) error {
	if pi == nil || pi.Instance == nil {
		return nil
	}
	if p.closed.Load() {
		_ = pi.Instance.Close()
		return nil
	}
	if time.Since(pi.LastUsed) > p.maxInstanceAge {
		_ = pi.Instance.Close()
		p.metrics.evictions.Add(1)
		return nil
	}

	p.mu.RLock()
	bucket, ok := p.buckets[pi.key]
	p.mu.RUnlock()
	if !ok {
		_ = pi.Instance.Close()
		return nil
	}

	bucket.mu.Lock()
	q, ok := bucket.perTenant[pi.TenantID]
	if !ok {
		q = make(chan GoRuntimeIfc, bucket.maxSize)
		bucket.perTenant[pi.TenantID] = q
	}
	bucket.mu.Unlock()

	select {
	case q <- pi.Instance:
		return nil
	default:
		_ = pi.Instance.Close()
		p.metrics.evictions.Add(1)
		return nil
	}
}

func (p *GoInstancePool) Prewarm(ctx context.Context, tenantID, runtime string, wasmBytes []byte, count int) error {
	if count <= 0 {
		count = p.defaultSize
	}
	for i := 0; i < count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		pi, err := p.Get(ctx, tenantID, runtime, wasmBytes)
		if err != nil {
			return err
		}
		if err := p.Put(pi); err != nil {
			return err
		}
	}
	return nil
}

func (p *GoInstancePool) RemoveTenant(tenantID string) {
	p.mu.RLock()
	buckets := make([]*goPoolBucket, 0, len(p.buckets))
	for _, b := range p.buckets {
		buckets = append(buckets, b)
	}
	p.mu.RUnlock()

	for _, b := range buckets {
		b.mu.Lock()
		q, ok := b.perTenant[tenantID]
		if ok {
			delete(b.perTenant, tenantID)
		}
		b.mu.Unlock()
		if !ok {
			continue
		}
		close(q)
		for inst := range q {
			_ = inst.Close()
			p.metrics.evictions.Add(1)
		}
	}
}

func (p *GoInstancePool) Stats() GoPoolStats {
	p.mu.RLock()
	totalInst := 0
	for _, b := range p.buckets {
		b.mu.Lock()
		for _, q := range b.perTenant {
			totalInst += len(q)
		}
		b.mu.Unlock()
	}
	p.mu.RUnlock()
	return GoPoolStats{
		Buckets:        len(p.buckets),
		TotalInstances: totalInst,
		TotalBytes:     p.totalBytes(),
		Hits:           p.metrics.hits.Load(),
		Misses:         p.metrics.misses.Load(),
		Evictions:      p.metrics.evictions.Load(),
		Timeouts:       p.metrics.timeouts.Load(),
	}
}

func (p *GoInstancePool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(p.stopCleanup)

	p.mu.Lock()
	buckets := p.buckets
	p.buckets = nil
	p.mu.Unlock()

	for _, b := range buckets {
		b.mu.Lock()
		perTenant := b.perTenant
		b.perTenant = nil
		b.mu.Unlock()
		for _, q := range perTenant {
			close(q)
			for inst := range q {
				_ = inst.Close()
			}
		}
	}
	return nil
}

var ErrGoPoolClosed = errors.New("go runtime: pool is closed")

func (p *GoInstancePool) totalBytes() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var total uint64
	for _, b := range p.buckets {
		b.mu.Lock()
		n := uint64(0)
		for _, q := range b.perTenant {
			n += uint64(len(q))
		}
		bytes := n * b.bytes
		b.mu.Unlock()
		total += bytes
	}
	return total
}

func (p *GoInstancePool) bucketBytes() uint64 {
	return uint64(p.cfg.MaxMemoryMB) * 1024 * 1024
}

func (p *GoInstancePool) evictOldest() {
	var oldest GoRuntimeIfc
	var oldestTime time.Time
	p.mu.RLock()
	for _, b := range p.buckets {
		b.mu.Lock()
		for _, q := range b.perTenant {
			for inst := range q {
				if oldest == nil || inst.CreatedAt().Before(oldestTime) {
					if oldest != nil {
						_ = oldest.Close()
					}
					oldest = inst
					oldestTime = inst.CreatedAt()
				} else {
					_ = inst.Close()
				}
			}
		}
		b.mu.Unlock()
	}
	p.mu.RUnlock()
	if oldest != nil {
		_ = oldest.Close()
		p.metrics.evictions.Add(1)
	}
}

func (p *GoInstancePool) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-p.stopCleanup:
			return
		case <-ticker.C:
			p.evictIdle()
		}
	}
}

func (p *GoInstancePool) evictIdle() {
	cutoff := time.Now().Add(-p.maxInstanceAge)
	p.mu.RLock()
	buckets := make([]*goPoolBucket, 0, len(p.buckets))
	for _, b := range p.buckets {
		buckets = append(buckets, b)
	}
	p.mu.RUnlock()

	for _, b := range buckets {
		b.mu.Lock()
		for tid, q := range b.perTenant {
			drained := make([]GoRuntimeIfc, 0, len(q))
			close(q)
			for inst := range q {
				if inst != nil && inst.CreatedAt().After(cutoff) {
					drained = append(drained, inst)
				} else if inst != nil {
					_ = inst.Close()
					p.metrics.evictions.Add(1)
				}
			}
			q2 := make(chan GoRuntimeIfc, b.maxSize)
			for _, inst := range drained {
				q2 <- inst
			}
			b.perTenant[tid] = q2
		}
		b.mu.Unlock()
	}
}

func (p *GoInstancePool) describe() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("buckets=%d hits=%d misses=%d evictions=%d",
		len(p.buckets), p.metrics.hits.Load(), p.metrics.misses.Load(), p.metrics.evictions.Load())
}
