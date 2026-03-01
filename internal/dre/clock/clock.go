// Package clock implements the virtual clock for the DCC Protocol.
// It provides deterministic time progression independent of wall-clock time,
// ensuring reproducible execution across replays.
package clock

import (
	"encoding/binary"
	"sync"
)

// VirtualClock provides deterministic virtual time for capsule execution.
// Time advances only through explicit operations (sleep, syscall, tick),
// never from wall-clock.
type VirtualClock struct {
	baseTime   uint64    // Base time derived from execution_id
	tickCount  uint64    // Number of ticks since start
	tickSize   uint64    // Nanoseconds per tick (default: 1)
	mu         sync.Mutex
}

// New creates a new VirtualClock with base time derived from execution ID.
// virtual_time = base_time + (tick_count * tick_size)
func New(executionID string) *VirtualClock {
	baseTime := deriveBaseTime(executionID)
	
	return &VirtualClock{
		baseTime: baseTime,
		tickCount: 0,
		tickSize: 1, // 1 nanosecond per tick by default
	}
}

// Now returns the current virtual time in nanoseconds.
func (c *VirtualClock) Now() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.baseTime + (c.tickCount * c.tickSize)
}

// Advance advances the virtual time by the given delta in nanoseconds.
func (c *VirtualClock) Advance(delta uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if delta > 0 {
		c.tickCount += delta / c.tickSize
	}
}

// Tick advances the virtual time by one tick (one instruction batch).
func (c *VirtualClock) Tick() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.tickCount++
}

// SetBase sets the base time (from execution_id).
func (c *VirtualClock) SetBase(base uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.baseTime = base
}

// SetTickSize sets the tick size in nanoseconds.
func (c *VirtualClock) SetTickSize(size uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if size > 0 {
		c.tickSize = size
	}
}

// TickCount returns the current tick count.
func (c *VirtualClock) TickCount() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.tickCount
}

// Reset resets the clock to initial state with new execution ID.
func (c *VirtualClock) Reset(executionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.baseTime = deriveBaseTime(executionID)
	c.tickCount = 0
}

// Sleep advances the virtual time by the given duration in nanoseconds.
// This should be called when the WASM runtime executes a sleep operation.
func (c *VirtualClock) Sleep(duration uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Advance by the sleep duration (converted to ticks)
	c.tickCount += duration / c.tickSize
}

// UnixNow returns the virtual time as Unix nanoseconds.
// This is for compatibility with code expecting time.Time.
func (c *VirtualClock) UnixNow() int64 {
	return int64(c.Now())
}

// Since returns the duration since the given virtual time.
func (c *VirtualClock) Since(start uint64) uint64 {
	now := c.Now()
	if now > start {
		return now - start
	}
	return 0
}

// Until returns the duration until the given virtual time.
func (c *VirtualClock) Until(end uint64) uint64 {
	now := c.Now()
	if end > now {
		return end - now
	}
	return 0
}

// deriveBaseTime derives a deterministic base time from execution ID.
// base_time = H(execution_id) mod 2^64
func deriveBaseTime(executionID string) uint64 {
	// Use a simple hash for deterministic base time derivation
	// In production, this should use a proper cryptographic hash
	h := hashString(executionID)
	return binary.LittleEndian.Uint64(h[:8])
}

// hashString is a simple non-cryptographic hash for base time derivation.
func hashString(s string) [8]byte {
	var h [8]byte
	seed := uint64(0x1234567890abcdef)
	
	for i := 0; i < len(s); i++ {
		seed = seed*31 + uint64(s[i])
		h[i%8] = byte(seed)
	}
	
	// Fill remaining bytes
	for i := len(s); i < 8; i++ {
		seed = seed*31 + uint64(i)
		h[i] = byte(seed)
	}
	
	return h
}
