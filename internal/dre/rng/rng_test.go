package rng

import (
	"testing"
)

func TestNew(t *testing.T) {
	seed := []byte("test-seed-12345678901234567890")
	r := New(seed)
	
	if r == nil {
		t.Fatal("New() returned nil")
	}
	
	if r.counter != 0 {
		t.Errorf("Expected counter to be 0, got %d", r.counter)
	}
}

func TestDeterministic(t *testing.T) {
	seed := []byte("deterministic-test-seed-123456")
	
	// First run
	r1 := New(seed)
	result1 := make([]byte, 100)
	r1.Read(result1)
	
	// Second run with same seed should produce identical results
	r2 := New(seed)
	result2 := make([]byte, 100)
	r2.Read(result2)
	
	for i := 0; i < 100; i++ {
		if result1[i] != result2[i] {
			t.Errorf("Results differ at byte %d: %d vs %d", i, result1[i], result2[i])
		}
	}
}

func TestDifferentSeeds(t *testing.T) {
	seed1 := []byte("seed-one-12345678901234567890")
	seed2 := []byte("seed-two-12345678901234567890")
	
	r1 := New(seed1)
	r2 := New(seed2)
	
	result1 := make([]byte, 100)
	result2 := make([]byte, 100)
	
	r1.Read(result1)
	r2.Read(result2)
	
	// Results should be different with different seeds
	different := false
	for i := 0; i < 100; i++ {
		if result1[i] != result2[i] {
			different = true
			break
		}
	}
	
	if !different {
		t.Error("Expected different results with different seeds")
	}
}

func TestNext(t *testing.T) {
	seed := []byte("next-test-seed-123456789012")
	r := New(seed)
	
	val1 := r.Next()
	val2 := r.Next()
	
	if val1 == val2 {
		t.Errorf("Next() returned same value twice: %d", val1)
	}
	
	if r.counter != 2 {
		t.Errorf("Expected counter to be 2, got %d", r.counter)
	}
}

func TestNext64(t *testing.T) {
	seed := []byte("next64-test-seed-1234567890")
	r := New(seed)
	
	val1 := r.Next64()
	val2 := r.Next64()
	
	if val1 == val2 {
		t.Errorf("Next64() returned same value twice: %d", val1)
	}
}

func TestIntn(t *testing.T) {
	seed := []byte("intn-test-seed-123456789012")
	r := New(seed)
	
	// Test with various bounds
	for _, n := range []int{1, 2, 10, 100, 1000} {
		for i := 0; i < 1000; i++ {
			val := r.Intn(n)
			if val < 0 || val >= n {
				t.Errorf("Intn(%d) returned out of range value: %d", n, val)
			}
		}
	}
}

func TestIntnZero(t *testing.T) {
	seed := []byte("intn-zero-test-seed-1234567890")
	r := New(seed)
	
	val := r.Intn(0)
	if val != 0 {
		t.Errorf("Intn(0) should return 0, got %d", val)
	}
}

func TestFloat64(t *testing.T) {
	seed := []byte("float64-test-seed-1234567890")
	r := New(seed)
	
	for i := 0; i < 1000; i++ {
		val := r.Float64()
		if val < 0 || val >= 1 {
			t.Errorf("Float64() returned out of range value: %f", val)
		}
	}
}

func TestCounter(t *testing.T) {
	seed := []byte("counter-test-seed-1234567890")
	r := New(seed)
	
	if r.Counter() != 0 {
		t.Errorf("Initial counter should be 0, got %d", r.Counter())
	}
	
	r.Next()
	if r.Counter() != 1 {
		t.Errorf("Counter should be 1 after Next(), got %d", r.Counter())
	}
	
	r.Next64()
	if r.Counter() != 2 {
		t.Errorf("Counter should be 2 after Next64(), got %d", r.Counter())
	}
	
	buf := make([]byte, 10)
	r.Read(buf)
	if r.Counter() != 3 {
		t.Errorf("Counter should be 3 after Read(), got %d", r.Counter())
	}
}

func TestSeed(t *testing.T) {
	seed1 := []byte("seed-one-for-reseed-test-123456")
	seed2 := []byte("seed-two-for-reseed-test-123456")
	
	r := New(seed1)
	val1 := r.Next()
	
	// Reseed with same seed - should produce same sequence
	r.Seed(seed1)
	val1AfterReseed := r.Next()
	
	if val1 != val1AfterReseed {
		t.Errorf("Reseed with same seed should produce same sequence: %d vs %d", val1, val1AfterReseed)
	}
	
	// Reseed with different seed - should produce different sequence
	r.Seed(seed2)
	val2 := r.Next()
	
	if val1 == val2 {
		t.Errorf("Reseed with different seed should produce different sequence: %d vs %d", val1, val2)
	}
}
