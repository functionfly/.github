package meter

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	m := New(nil)
	
	if m == nil {
		t.Fatal("New() returned nil")
	}
	
	if m.InstructionCount() != 0 {
		t.Errorf("Expected instruction count 0, got %d", m.InstructionCount())
	}
}

func TestNewWithLimits(t *testing.T) {
	limits := &Limits{
		MaxInstructions: 1000,
		MaxMemory:       1024,
		MaxSyscalls:     100,
	}
	
	m := New(limits)
	
	if m.InstructionCount() != 0 {
		t.Error("Expected 0 instructions")
	}
}

func TestAddInstructions(t *testing.T) {
	m := New(nil)
	
	m.AddInstructions(100)
	
	if m.InstructionCount() != 100 {
		t.Errorf("Expected 100 instructions, got %d", m.InstructionCount())
	}
	
	m.AddInstructions(50)
	
	if m.InstructionCount() != 150 {
		t.Errorf("Expected 150 instructions, got %d", m.InstructionCount())
	}
}

func TestIncInstructions(t *testing.T) {
	m := New(nil)
	
	for i := 0; i < 100; i++ {
		m.IncInstructions()
	}
	
	if m.InstructionCount() != 100 {
		t.Errorf("Expected 100 instructions, got %d", m.InstructionCount())
	}
}

func TestSetMemory(t *testing.T) {
	m := New(nil)
	
	m.SetMemory(1024)
	
	if m.CurrentMemory() != 1024 {
		t.Errorf("Expected current memory 1024, got %d", m.CurrentMemory())
	}
	
	if m.PeakMemory() != 1024 {
		t.Errorf("Expected peak memory 1024, got %d", m.PeakMemory())
	}
	
	// Set lower memory - peak should stay the same
	m.SetMemory(512)
	
	if m.CurrentMemory() != 512 {
		t.Errorf("Expected current memory 512, got %d", m.CurrentMemory())
	}
	
	if m.PeakMemory() != 1024 {
		t.Errorf("Expected peak memory 1024, got %d", m.PeakMemory())
	}
	
	// Set higher memory - peak should update
	m.SetMemory(2048)
	
	if m.PeakMemory() != 2048 {
		t.Errorf("Expected peak memory 2048, got %d", m.PeakMemory())
	}
}

func TestSyscallCount(t *testing.T) {
	m := New(nil)
	
	m.IncSyscallCount()
	
	if m.SyscallCount() != 1 {
		t.Errorf("Expected syscall count 1, got %d", m.SyscallCount())
	}
	
	m.AddSyscallCount(10)
	
	if m.SyscallCount() != 11 {
		t.Errorf("Expected syscall count 11, got %d", m.SyscallCount())
	}
}

func TestExceeded(t *testing.T) {
	limits := &Limits{
		MaxInstructions: 100,
		MaxMemory:       1024,
		MaxSyscalls:     10,
	}
	
	m := New(limits)
	
	// No limits exceeded
	exceeded := m.Exceeded()
	if len(exceeded) != 0 {
		t.Errorf("Expected no exceeded limits, got %v", exceeded)
	}
	
	// Exceed instructions
	m.AddInstructions(101)
	
	exceeded = m.Exceeded()
	if len(exceeded) != 1 || exceeded[0] != "instructions" {
		t.Errorf("Expected 'instructions' exceeded, got %v", exceeded)
	}
	
	// Exceed memory
	m.SetMemory(2048)
	
	exceeded = m.Exceeded()
	found := false
	for _, e := range exceeded {
		if e == "instructions" || e == "memory" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected 'instructions' and 'memory' exceeded, got %v", exceeded)
	}
}

func TestIsWithinLimits(t *testing.T) {
	limits := &Limits{
		MaxInstructions: 100,
		MaxMemory:       1024,
	}
	
	m := New(limits)
	
	if !m.IsWithinLimits() {
		t.Error("Should be within limits initially")
	}
	
	m.AddInstructions(101)
	
	if m.IsWithinLimits() {
		t.Error("Should not be within limits after exceeding")
	}
}

func TestResourceHash(t *testing.T) {
	m := New(nil)
	
	m.AddInstructions(100)
	m.SetMemory(1024)
	m.AddSyscallCount(10)
	
	hash1 := m.ResourceHash()
	
	// Hash should be consistent
	hash2 := m.ResourceHash()
	
	if hash1 != hash2 {
		t.Errorf("ResourceHash() should be consistent: %s vs %s", hash1, hash2)
	}
	
	if hash1 == "" {
		t.Error("ResourceHash() should not be empty")
	}
}

func TestDeterministicResourceHash(t *testing.T) {
	m1 := New(nil)
	m1.AddInstructions(100)
	m1.SetMemory(1024)
	m1.AddSyscallCount(10)
	hash1 := m1.ResourceHash()
	
	m2 := New(nil)
	m2.AddInstructions(100)
	m2.SetMemory(1024)
	m2.AddSyscallCount(10)
	hash2 := m2.ResourceHash()
	
	// Hashes should be identical
	if hash1 != hash2 {
		t.Errorf("ResourceHash() should be deterministic: %s vs %s", hash1, hash2)
	}
}

func TestDifferentMetricsDifferentHash(t *testing.T) {
	m1 := New(nil)
	m1.AddInstructions(100)
	hash1 := m1.ResourceHash()
	
	m2 := New(nil)
	m2.AddInstructions(101)
	hash2 := m2.ResourceHash()
	
	// Hashes should be different
	if hash1 == hash2 {
		t.Error("ResourceHash() should differ with different metrics")
	}
}

func TestReset(t *testing.T) {
	m := New(nil)
	
	m.AddInstructions(100)
	m.SetMemory(1024)
	m.AddSyscallCount(10)
	m.ResourceHash() // Pre-compute hash
	
	m.Reset()
	
	if m.InstructionCount() != 0 {
		t.Errorf("Expected instruction count 0 after reset, got %d", m.InstructionCount())
	}
	
	if m.PeakMemory() != 0 {
		t.Errorf("Expected peak memory 0 after reset, got %d", m.PeakMemory())
	}
	
	if m.SyscallCount() != 0 {
		t.Errorf("Expected syscall count 0 after reset, got %d", m.SyscallCount())
	}
}

func TestStats(t *testing.T) {
	m := New(nil)
	
	m.AddInstructions(100)
	m.SetMemory(1024)
	m.AddSyscallCount(10)
	
	stats := m.Stats()
	
	if stats.InstructionCount != 100 {
		t.Errorf("Expected instruction count 100, got %d", stats.InstructionCount)
	}
	
	if stats.CurrentMemory != 1024 {
		t.Errorf("Expected current memory 1024, got %d", stats.CurrentMemory)
	}
	
	if stats.SyscallCount != 10 {
		t.Errorf("Expected syscall count 10, got %d", stats.SyscallCount)
	}
}

func TestStartTime(t *testing.T) {
	st := NewStartTime()
	
	time.Sleep(10 * time.Millisecond)
	
	d := st.Duration()
	
	if d < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", d)
	}
}

func TestStartTimeIsExpired(t *testing.T) {
	st := NewStartTime()
	
	// Should not be expired for 1 second limit
	if st.IsExpired(1000) {
		t.Error("Should not be expired")
	}
	
	// Should be expired for 0ms limit (immediate)
	if !st.IsExpired(0) {
		t.Error("Should be expired for 0ms limit")
	}
}
