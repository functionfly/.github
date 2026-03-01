package clock

import (
	"testing"
)

func TestNew(t *testing.T) {
	execID := "test-execution-id-123"
	c := New(execID)
	
	if c == nil {
		t.Fatal("New() returned nil")
	}
	
	if c.Now() == 0 {
		t.Error("Initial virtual time should not be zero (derived from execution ID)")
	}
}

func TestDeterministic(t *testing.T) {
	execID := "deterministic-execution-id-123"
	
	// First instance
	c1 := New(execID)
	time1 := c1.Now()
	c1.Tick()
	time1AfterTick := c1.Now()
	
	// Second instance with same execution ID
	c2 := New(execID)
	time2 := c2.Now()
	c2.Tick()
	time2AfterTick := c2.Now()
	
	// Both should produce identical results
	if time1 != time2 {
		t.Errorf("Initial time differs: %d vs %d", time1, time2)
	}
	
	if time1AfterTick != time2AfterTick {
		t.Errorf("Time after tick differs: %d vs %d", time1AfterTick, time2AfterTick)
	}
}

func TestDifferentExecutionIDs(t *testing.T) {
	execID1 := "execution-id-one-1234567890"
	execID2 := "execution-id-two-1234567890"
	
	c1 := New(execID1)
	c2 := New(execID2)
	
	// Different execution IDs should produce different base times
	if c1.Now() == c2.Now() {
		t.Error("Different execution IDs should produce different base times")
	}
}

func TestTick(t *testing.T) {
	c := New("test-execution-tick")
	initialTime := c.Now()
	
	c.Tick()
	afterTick := c.Now()
	
	if afterTick <= initialTime {
		t.Errorf("Tick should advance time: initial=%d, after=%d", initialTime, afterTick)
	}
	
	if c.TickCount() != 1 {
		t.Errorf("TickCount should be 1, got %d", c.TickCount())
	}
}

func TestAdvance(t *testing.T) {
	c := New("test-execution-advance")
	initialTime := c.Now()
	
	c.Advance(1000) // Advance 1000 nanoseconds
	afterAdvance := c.Now()
	
	expected := initialTime + 1000
	if afterAdvance != expected {
		t.Errorf("Expected time %d, got %d", expected, afterAdvance)
	}
}

func TestSleep(t *testing.T) {
	c := New("test-execution-sleep")
	initialTime := c.Now()
	
	c.Sleep(5000000000) // Sleep 5 seconds (5 billion nanoseconds)
	afterSleep := c.Now()
	
	expected := initialTime + 5000000000
	if afterSleep != expected {
		t.Errorf("Expected time %d, got %d", expected, afterSleep)
	}
}

func TestSetBase(t *testing.T) {
	c := New("test-execution-setbase")
	
	newBase := uint64(1000000)
	c.SetBase(newBase)
	
	if c.Now() != newBase {
		t.Errorf("Expected base time %d, got %d", newBase, c.Now())
	}
}

func TestSetTickSize(t *testing.T) {
	c := New("test-execution-ticksize")
	
	// Set tick size to 100 nanoseconds
	c.SetTickSize(100)
	
	initialTime := c.Now()
	c.Tick()
	afterTick := c.Now()
	
	expected := initialTime + 100
	if afterTick != expected {
		t.Errorf("Expected time %d, got %d", expected, afterTick)
	}
}

func TestReset(t *testing.T) {
	execID1 := "original-execution-id"
	execID2 := "new-execution-id"
	
	c := New(execID1)
	originalTime := c.Now()
	c.Tick()
	c.Tick()
	
	// Reset with new execution ID
	c.Reset(execID2)
	newTime := c.Now()
	
	// Reset should restore to base time derived from new ID
	if c.TickCount() != 0 {
		t.Errorf("TickCount should be 0 after reset, got %d", c.TickCount())
	}
	
	// Base time should be different
	if originalTime == newTime {
		t.Error("Reset should change base time")
	}
}

func TestSince(t *testing.T) {
	c := New("test-execution-since")
	
	start := c.Now()
	c.Tick()
	c.Tick()
	
	since := c.Since(start)
	
	if since == 0 {
		t.Error("Since should return positive duration")
	}
}

func TestUntil(t *testing.T) {
	c := New("test-execution-until")
	
	current := c.Now()
	end := current + 1000
	
	until := c.Until(end)
	
	if until != 1000 {
		t.Errorf("Expected until to be 1000, got %d", until)
	}
}

func TestMultipleTicks(t *testing.T) {
	c := New("test-multiple-ticks")
	initialTime := c.Now()
	
	for i := 0; i < 1000; i++ {
		c.Tick()
	}
	
	expected := initialTime + 1000
	if c.Now() != expected {
		t.Errorf("Expected time %d after 1000 ticks, got %d", expected, c.Now())
	}
	
	if c.TickCount() != 1000 {
		t.Errorf("TickCount should be 1000, got %d", c.TickCount())
	}
}
