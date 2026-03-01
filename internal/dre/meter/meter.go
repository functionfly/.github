// Package meter implements instruction metering for the DCC Protocol.
// It tracks instruction count, memory usage, and syscall count for resource limiting.
package meter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Meter tracks resource usage during capsule execution.
type Meter struct {
	// Instruction count
	instructionCount uint64
	
	// Peak memory usage (bytes)
	peakMemory uint64
	
	// Current memory usage (bytes)
	currentMemory uint64
	
	// Syscall count
	syscallCount uint64
	
	// Limit settings
	limit *Limits
	
	// Resource tracking data
	resourceHash string
	
	mu sync.RWMutex
}

// Limits defines resource limits for the capsule.
type Limits struct {
	MaxInstructions uint64 `json:"max_instructions"`
	MaxMemory       uint64 `json:"max_memory"`
	MaxSyscalls     uint64 `json:"max_syscalls"`
	MaxDuration     int64  `json:"max_duration_ms"`
}

// DefaultLimits returns default resource limits.
func DefaultLimits() *Limits {
	return &Limits{
		MaxInstructions: 1_000_000_000, // 1B instructions
		MaxMemory:       128 * 1024 * 1024, // 128 MB
		MaxSyscalls:     1_000_000, // 1M syscalls
		MaxDuration:     300_000, // 5 minutes
	}
}

// New creates a new Meter with the given limits.
func New(limits *Limits) *Meter {
	if limits == nil {
		limits = DefaultLimits()
	}
	
	return &Meter{
		limit:          limits,
		instructionCount: 0,
		peakMemory:       0,
		currentMemory:    0,
		syscallCount:     0,
	}
}

// AddInstructions adds to the instruction count.
func (m *Meter) AddInstructions(count uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.instructionCount += count
}

// IncInstructions increments the instruction count by 1.
func (m *Meter) IncInstructions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.instructionCount++
}

// SetMemory sets the current memory usage.
func (m *Meter) SetMemory(bytes uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.currentMemory = bytes
	
	if bytes > m.peakMemory {
		m.peakMemory = bytes
	}
}

// IncSyscallCount increments the syscall count.
func (m *Meter) IncSyscallCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.syscallCount++
}

// AddSyscallCount adds to the syscall count.
func (m *Meter) AddSyscallCount(count uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.syscallCount += count
}

// InstructionCount returns the current instruction count.
func (m *Meter) InstructionCount() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instructionCount
}

// PeakMemory returns the peak memory usage.
func (m *Meter) PeakMemory() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peakMemory
}

// CurrentMemory returns the current memory usage.
func (m *Meter) CurrentMemory() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMemory
}

// SyscallCount returns the syscall count.
func (m *Meter) SyscallCount() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.syscallCount
}

// Exceeded returns which limits have been exceeded.
func (m *Meter) Exceeded() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var exceeded []string
	
	if m.limit != nil {
		if m.instructionCount > m.limit.MaxInstructions {
			exceeded = append(exceeded, "instructions")
		}
		if m.peakMemory > m.limit.MaxMemory {
			exceeded = append(exceeded, "memory")
		}
		if m.syscallCount > m.limit.MaxSyscalls {
			exceeded = append(exceeded, "syscalls")
		}
	}
	
	return exceeded
}

// IsWithinLimits returns true if all limits are satisfied.
func (m *Meter) IsWithinLimits() bool {
	return len(m.Exceeded()) == 0
}

// ResourceHash returns the hash of all resource metrics.
// This is included in the ResourceHash of the MEG.
func (m *Meter) ResourceHash() string {
	m.mu.RLock()
	
	if m.resourceHash != "" {
		m.mu.RUnlock()
		return m.resourceHash
	}
	
	instCount := m.instructionCount
	peakMem := m.peakMemory
	sysCount := m.syscallCount
	m.mu.RUnlock()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("instructions:%d", instCount)))
	h.Write([]byte(fmt.Sprintf("peak_memory:%d", peakMem)))
	h.Write([]byte(fmt.Sprintf("syscalls:%d", sysCount)))
	
	m.resourceHash = hex.EncodeToString(h.Sum(nil))
	return m.resourceHash
}

// Reset resets all counters.
func (m *Meter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.instructionCount = 0
	m.peakMemory = 0
	m.currentMemory = 0
	m.syscallCount = 0
	m.resourceHash = ""
}

// Stats returns the current stats.
func (m *Meter) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return Stats{
		InstructionCount: m.instructionCount,
		PeakMemory:       m.peakMemory,
		CurrentMemory:    m.currentMemory,
		SyscallCount:    m.syscallCount,
	}
}

// Stats represents resource usage statistics.
type Stats struct {
	InstructionCount uint64 `json:"instruction_count"`
	PeakMemory       uint64 `json:"peak_memory_bytes"`
	CurrentMemory    uint64 `json:"current_memory_bytes"`
	SyscallCount     uint64 `json:"syscall_count"`
}

// JSON returns the stats as JSON.
func (m *Meter) JSON() (string, error) {
	stats := m.Stats()
	
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal stats: %w", err)
	}
	
	return string(data), nil
}

// StartTime tracks when execution started (for duration limits).
type StartTime struct {
	start time.Time
	mu    sync.RWMutex
}

// NewStartTime creates a new StartTime.
func NewStartTime() *StartTime {
	return &StartTime{
		start: time.Now(),
	}
}

// Set sets the start time.
func (s *StartTime) Set(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.start = t
}

// Get returns the start time.
func (s *StartTime) Get() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.start
}

// Duration returns the elapsed duration.
func (s *StartTime) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.start)
}

// IsExpired checks if the duration exceeds the limit.
func (s *StartTime) IsExpired(limitMs int64) bool {
	d := s.Duration()
	return d.Milliseconds() > limitMs
}
