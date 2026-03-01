package syscall

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New("strict-v1")
	
	if g == nil {
		t.Fatal("New() returned nil")
	}
	
	if g.profile != "strict-v1" {
		t.Errorf("Expected profile 'strict-v1', got '%s'", g.profile)
	}
}

func TestAllowed(t *testing.T) {
	g := New("strict-v1")
	
	// Test allowed syscalls
	if !g.Allowed(SyscallMemoryAlloc) {
		t.Error("MemoryAlloc should be allowed")
	}
	
	if !g.Allowed(SyscallClockRead) {
		t.Error("ClockRead should be allowed")
	}
	
	if !g.Allowed(SyscallRNG) {
		t.Error("RNG should be allowed")
	}
	
	// Test blocked syscalls
	if g.Allowed(SyscallSocketCreate) {
		t.Error("SocketCreate should be blocked")
	}
	
	if g.Allowed(SyscallSystemClock) {
		t.Error("SystemClock should be blocked")
	}
	
	if g.Allowed(SyscallDiskAccess) {
		t.Error("DiskAccess should be blocked")
	}
}

func TestRecord(t *testing.T) {
	g := New("strict-v1")
	
	args := []byte(`{"size": 1024}`)
	result := []byte(`{"ptr": 0x1000}`)
	
	g.Record(SyscallMemoryAlloc, args, result, 1000)
	
	trace := g.Trace()
	
	if len(trace) != 1 {
		t.Errorf("Expected trace length 1, got %d", len(trace))
	}
	
	if trace[0].ID != SyscallMemoryAlloc {
		t.Errorf("Expected syscall ID %d, got %d", SyscallMemoryAlloc, trace[0].ID)
	}
	
	if trace[0].Timestamp != 1000 {
		t.Errorf("Expected timestamp 1000, got %d", trace[0].Timestamp)
	}
}

func TestDeterministicHash(t *testing.T) {
	// First execution
	g1 := New("strict-v1")
	args1 := []byte(`{"size": 1024}`)
	result1 := []byte(`{"ptr": 0x1000}`)
	g1.Record(SyscallMemoryAlloc, args1, result1, 1000)
	g1.Record(SyscallRNG, nil, []byte(`{"value": 42}`), 1001)
	hash1 := g1.Hash()
	
	// Second execution (replay) with same inputs
	g2 := New("strict-v1")
	args2 := []byte(`{"size": 1024}`)
	result2 := []byte(`{"ptr": 0x1000}`)
	g2.Record(SyscallMemoryAlloc, args2, result2, 1000)
	g2.Record(SyscallRNG, nil, []byte(`{"value": 42}`), 1001)
	hash2 := g2.Hash()
	
	// Hashes should be identical
	if hash1 != hash2 {
		t.Errorf("Hashes should be identical: %s vs %s", hash1, hash2)
	}
}

func TestDifferentResultsDifferentHash(t *testing.T) {
	// First execution
	g1 := New("strict-v1")
	g1.Record(SyscallMemoryAlloc, []byte(`{"size": 1024}`), []byte(`{"ptr": 0x1000}`), 1000)
	hash1 := g1.Hash()
	
	// Second execution with different result
	g2 := New("strict-v1")
	g2.Record(SyscallMemoryAlloc, []byte(`{"size": 1024}`), []byte(`{"ptr": 0x2000}`), 1000)
	hash2 := g2.Hash()
	
	// Hashes should be different
	if hash1 == hash2 {
		t.Error("Hashes should be different with different results")
	}
}

func TestClearTrace(t *testing.T) {
	g := New("strict-v1")
	
	g.Record(SyscallMemoryAlloc, nil, nil, 1000)
	g.Record(SyscallRNG, nil, nil, 1001)
	
	if len(g.Trace()) != 2 {
		t.Errorf("Expected trace length 2, got %d", len(g.Trace()))
	}
	
	g.ClearTrace()
	
	if len(g.Trace()) != 0 {
		t.Errorf("Expected trace length 0 after clear, got %d", len(g.Trace()))
	}
	
	// Hash should be recalculated
	hash := g.Hash()
	if hash == "" {
		t.Error("Hash should not be empty after clear")
	}
}

func TestDisableTrace(t *testing.T) {
	g := New("strict-v1")
	
	g.Record(SyscallMemoryAlloc, nil, nil, 1000)
	
	if len(g.Trace()) != 1 {
		t.Errorf("Expected trace length 1, got %d", len(g.Trace()))
	}
	
	g.DisableTrace()
	g.Record(SyscallMemoryAlloc, nil, nil, 1001)
	
	// Should not record when disabled
	if len(g.Trace()) != 1 {
		t.Errorf("Expected trace length 1 (disabled), got %d", len(g.Trace()))
	}
	
	g.EnableTrace()
	g.Record(SyscallMemoryAlloc, nil, nil, 1002)
	
	// Should record when enabled
	if len(g.Trace()) != 2 {
		t.Errorf("Expected trace length 2 (re-enabled), got %d", len(g.Trace()))
	}
}

func TestSyscallName(t *testing.T) {
	tests := []struct {
		id     SyscallID
		expect string
	}{
		{SyscallMemoryAlloc, "memory_alloc"},
		{SyscallClockRead, "clock_read"},
		{SyscallRNG, "rng"},
		{SyscallSocketCreate, "socket_create"},
		{SyscallSystemClock, "system_clock"},
	}
	
	for _, tc := range tests {
		name := SyscallName(tc.id)
		if name != tc.expect {
			t.Errorf("SyscallName(%d) = %s, expected %s", tc.id, name, tc.expect)
		}
	}
}

func TestJSON(t *testing.T) {
	g := New("strict-v1")
	
	g.Record(SyscallMemoryAlloc, []byte(`{"size": 1024}`), []byte(`{"ptr": 0x1000}`), 1000)
	
	jsonStr, err := g.JSON()
	if err != nil {
		t.Errorf("JSON() returned error: %v", err)
	}
	
	if jsonStr == "" {
		t.Error("JSON() should not return empty string")
	}
}
