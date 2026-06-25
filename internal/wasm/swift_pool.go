//go:build cgo

// Swift runtime pool — caches compiled SwiftWASIRuntime instances by binary hash.
// The expensive operation (wasmtime engine + module compilation) happens once;
// each Execute() creates only a lightweight store + instance.
package wasm

import (
	"crypto/sha256"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// SwiftPoolEntry caches a compiled SwiftWASIRuntime keyed by binary hash.
type SwiftPoolEntry struct {
	runtime   *SwiftWASIRuntime
	hash      [32]byte
	createdAt time.Time
	useCount  atomic.Int64
}

// SwiftRuntimePool manages a cache of compiled SwiftWASIRuntime instances.
// Instances are keyed by SHA-256 of the WASM binary to avoid recompilation.
type SwiftRuntimePool struct {
	mu       sync.RWMutex
	entries  map[[32]byte]*SwiftPoolEntry
	maxSize  int
	maxAge   time.Duration
	hits     atomic.Int64
	misses   atomic.Int64
	evictions atomic.Int64
}

// NewSwiftRuntimePool creates a pool with the given limits.
func NewSwiftRuntimePool(maxSize int, maxAge time.Duration) *SwiftRuntimePool {
	if maxSize <= 0 {
		maxSize = 64
	}
	if maxAge <= 0 {
		maxAge = 30 * time.Minute
	}

	p := &SwiftRuntimePool{
		entries: make(map[[32]byte]*SwiftPoolEntry),
		maxSize: maxSize,
		maxAge:  maxAge,
	}

	go p.cleanupLoop()
	return p
}

// Get returns a cached SwiftWASIRuntime for the given binary, or creates one.
func (p *SwiftRuntimePool) Get(wasmBinary []byte, handler HostFunctionHandler, config *WASMSecurityConfig) (*SwiftWASIRuntime, error) {
	hash := sha256.Sum256(wasmBinary)

	p.mu.RLock()
	entry, exists := p.entries[hash]
	p.mu.RUnlock()

	if exists && entry.runtime != nil {
		entry.useCount.Add(1)
		p.hits.Add(1)
		return entry.runtime, nil
	}

	p.misses.Add(1)

	runtime, err := NewSwiftWASIRuntimeWithConfig(wasmBinary, handler, config)
	if err != nil {
		return nil, err
	}

	entry = &SwiftPoolEntry{
		runtime:   runtime,
		hash:      hash,
		createdAt: time.Now(),
	}
	entry.useCount.Add(1)

	p.mu.Lock()
	// Evict oldest if at capacity
	if len(p.entries) >= p.maxSize {
		p.evictOldest()
	}
	p.entries[hash] = entry
	p.mu.Unlock()

	logrus.WithField("hash", fmtHash(hash)).Debug("SwiftRuntimePool: compiled new entry")
	return runtime, nil
}

// evictOldest removes the oldest entry. Caller must hold p.mu write lock.
func (p *SwiftRuntimePool) evictOldest() {
	var oldest *SwiftPoolEntry
	for _, e := range p.entries {
		if oldest == nil || e.createdAt.Before(oldest.createdAt) {
			oldest = e
		}
	}
	if oldest != nil {
		oldest.runtime.Close()
		delete(p.entries, oldest.hash)
		p.evictions.Add(1)
	}
}

// cleanupLoop periodically removes expired entries.
func (p *SwiftRuntimePool) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanup()
	}
}

// cleanup removes expired entries.
func (p *SwiftRuntimePool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for hash, entry := range p.entries {
		if now.Sub(entry.createdAt) > p.maxAge {
			entry.runtime.Close()
			delete(p.entries, hash)
			p.evictions.Add(1)
		}
	}
}

// Stats returns pool statistics.
func (p *SwiftRuntimePool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"entries":    len(p.entries),
		"max_size":   p.maxSize,
		"hits":       p.hits.Load(),
		"misses":     p.misses.Load(),
		"evictions":  p.evictions.Load(),
		"hit_rate":   p.hitRate(),
	}
}

func (p *SwiftRuntimePool) hitRate() float64 {
	total := p.hits.Load() + p.misses.Load()
	if total == 0 {
		return 0
	}
	return float64(p.hits.Load()) / float64(total) * 100
}

// Close shuts down the pool and releases all cached runtimes.
func (p *SwiftRuntimePool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, entry := range p.entries {
		entry.runtime.Close()
	}
	p.entries = make(map[[32]byte]*SwiftPoolEntry)
}

// fmtHash returns a short hex representation of a hash.
func fmtHash(h [32]byte) string {
	const hex = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 0; i < 8; i++ {
		buf[i*2] = hex[h[i]>>4]
		buf[i*2+1] = hex[h[i]&0x0f]
	}
	return string(buf)
}

// GlobalSwiftPool is the process-wide cache of compiled Swift WASM runtimes.
var GlobalSwiftPool = NewSwiftRuntimePool(64, 30*time.Minute)
