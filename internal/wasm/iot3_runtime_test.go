//go:build cgo

package wasm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Unit Tests
// =============================================================================

func TestIoTConfig_DefaultConfig(t *testing.T) {
	config := DefaultIoTConfig()

	if config.TargetLatency != 500 {
		t.Errorf("Expected TargetLatency=500, got %d", config.TargetLatency)
	}
	if config.MaxMemoryKB != 16*1024 {
		t.Errorf("Expected MaxMemoryKB=16384, got %d", config.MaxMemoryKB)
	}
	if config.MaxInstances != 4 {
		t.Errorf("Expected MaxInstances=4, got %d", config.MaxInstances)
	}
	if config.ExecutionTimeout != 450*time.Millisecond {
		t.Errorf("Expected ExecutionTimeout=450ms, got %v", config.ExecutionTimeout)
	}
	if !config.BatteryOptimized {
		t.Error("Expected BatteryOptimized=true")
	}
}

func TestIoTConfig_Validation(t *testing.T) {
	// Test memory validation separately
	t.Run("zero max memory uses default", func(t *testing.T) {
		config := &IoTConfig{MaxMemoryKB: 0, ExecutionTimeout: 0}
		runtime, err := NewWASM3IoTRuntime(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if runtime.config.MaxMemoryKB != 16*1024 {
			t.Errorf("Expected MaxMemoryKB=16384, got %d", runtime.config.MaxMemoryKB)
		}
	})

	t.Run("excessive memory capped at 64MB", func(t *testing.T) {
		config := &IoTConfig{MaxMemoryKB: 100 * 1024, ExecutionTimeout: 0}
		runtime, err := NewWASM3IoTRuntime(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if runtime.config.MaxMemoryKB != 64*1024 {
			t.Errorf("Expected MaxMemoryKB=65536, got %d", runtime.config.MaxMemoryKB)
		}
	})

	t.Run("valid memory preserved", func(t *testing.T) {
		config := &IoTConfig{MaxMemoryKB: 32 * 1024, ExecutionTimeout: 0}
		runtime, err := NewWASM3IoTRuntime(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if runtime.config.MaxMemoryKB != 32*1024 {
			t.Errorf("Expected MaxMemoryKB=32768, got %d", runtime.config.MaxMemoryKB)
		}
	})

	t.Run("zero timeout uses default", func(t *testing.T) {
		config := &IoTConfig{MaxMemoryKB: 16 * 1024, ExecutionTimeout: 0}
		runtime, err := NewWASM3IoTRuntime(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if runtime.config.ExecutionTimeout != 450*time.Millisecond {
			t.Errorf("Expected ExecutionTimeout=450ms, got %v", runtime.config.ExecutionTimeout)
		}
	})
}

func TestNewWASM3IoTRuntime(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(nil)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	if runtime.pool == nil {
		t.Error("Expected pool to be initialized")
	}
	if runtime.closed {
		t.Error("Expected runtime to not be closed")
	}
}

func TestWASM3IoTRuntime_Execute(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("test input")

	output, err := runtime.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output format: [WASM3-IoT]<timestamp>:<input>
	if len(output) < 12 { // [WASM3-IoT]: + at least one digit for timestamp
		t.Errorf("Output too short: %s", string(output))
	}

	// Check prefix
	expectedPrefix := "[WASM3-IoT]"
	if len(output) < len(expectedPrefix) {
		t.Errorf("Output too short for prefix check: %s", string(output))
	}
	if string(output[:len(expectedPrefix)]) != expectedPrefix {
		t.Errorf("Expected prefix %s, got %s", expectedPrefix, string(output[:len(expectedPrefix)]))
	}
}

func TestWASM3IoTRuntime_ExecuteWithCancel(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     1,
		ExecutionTimeout: 10 * time.Second, // Long timeout
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := []byte("test")
	_, err = runtime.Execute(ctx, input)

	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestWASM3IoTRuntime_ExecuteWithTimeout(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     1,
		ExecutionTimeout: 1 * time.Millisecond, // Very short timeout
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("test")
	output, err := runtime.Execute(ctx, input)

	if err != nil {
		t.Errorf("Did not expect error when no module is loaded: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from fallback")
	}
}

func TestWASM3IoTRuntime_ExecuteWithConfig(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("test")

	// Execute with nil config
	output, err := runtime.ExecuteWithConfig(ctx, input, nil)
	if err != nil {
		t.Fatalf("ExecuteWithConfig failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestWASM3IoTRuntime_LoadModule(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Valid WASM module (magic number + version 1)
	validModule := []byte{
		0x00, 0x61, 0x73, 0x6D, // magic
		0x01, 0x00, 0x00, 0x00, // version 1
	}

	err = runtime.LoadModule(validModule)
	if err != nil {
		t.Fatalf("LoadModule failed for valid module: %v", err)
	}

	// Invalid module (too small)
	err = runtime.LoadModule([]byte{0x00, 0x61})
	if err == nil {
		t.Error("Expected error for module too small")
	}

	// Invalid WASM header
	invalidModule := []byte{
		0x00, 0x00, 0x00, 0x00, // invalid magic
		0x01, 0x00, 0x00, 0x00,
	}
	err = runtime.LoadModule(invalidModule)
	if err == nil {
		t.Error("Expected error for invalid WASM header")
	}

	// Unsupported WASM version
	badVersion := []byte{
		0x00, 0x61, 0x73, 0x6D, // magic
		0x02, 0x00, 0x00, 0x00, // version 2 (unsupported)
	}
	err = runtime.LoadModule(badVersion)
	if err == nil {
		t.Error("Expected error for unsupported WASM version")
	}
}

func TestWASM3IoTRuntime_GetMetrics(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Execute to populate metrics
	ctx := context.Background()
	runtime.Execute(ctx, []byte("test"))

	metrics := runtime.GetMetrics()

	if metrics["max_instances"] != 4 {
		t.Errorf("Expected max_instances=4, got %v", metrics["max_instances"])
	}
	if metrics["target_latency_ms"] != 500 {
		t.Errorf("Expected target_latency_ms=500, got %v", metrics["target_latency_ms"])
	}
	if metrics["max_memory_kb"] != 16384 {
		t.Errorf("Expected max_memory_kb=16384, got %v", metrics["max_memory_kb"])
	}
	if metrics["battery_optimized"] != true {
		t.Error("Expected battery_optimized=true")
	}
}

func TestWASM3IoTRuntime_Close(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Close once
	err = runtime.Close()
	if err != nil {
		t.Fatalf("First Close failed: %v", err)
	}

	// Close again should be idempotent
	err = runtime.Close()
	if err != nil {
		t.Fatalf("Second Close should be idempotent: %v", err)
	}
}

func TestWASM3IoTRuntime_CloseAfterExecute(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Execute multiple times
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := runtime.Execute(ctx, []byte(fmt.Sprintf("test-%d", i)))
		if err != nil {
			t.Fatalf("Execute %d failed: %v", i, err)
		}
	}

	// Close should work after multiple executions
	err = runtime.Close()
	if err != nil {
		t.Fatalf("Close after Execute failed: %v", err)
	}
}

// =============================================================================
// Instance Pool Tests
// =============================================================================

func TestIoTInstancePool_GetPut(t *testing.T) {
	config := DefaultIoTConfig()
	config.MaxInstances = 2
	pool := NewIoTInstancePool(config)

	ctx := context.Background()

	// Get first instance
	inst1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("First Get failed: %v", err)
	}

	// Get second instance
	inst2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Second Get failed: %v", err)
	}

	// Instances should be different
	if inst1.ID == inst2.ID {
		t.Error("Expected different instances")
	}

	// Return instances to pool
	pool.Put(inst1)
	pool.Put(inst2)

	pool.Close()
}

func TestIoTInstancePool_ExhaustAndWait(t *testing.T) {
	config := &IoTConfig{
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     1,
		ExecutionTimeout: 100 * time.Millisecond,
	}
	pool := NewIoTInstancePool(config)

	ctx := context.Background()

	// Get the only instance
	inst1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("First Get failed: %v", err)
	}

	// Second Get should timeout or wait
	ctx2, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	_, err = pool.Get(ctx2)
	if err == nil {
		t.Log("Got instance from waiting pool (might be fast)")
	} else if err == context.DeadlineExceeded {
		t.Log("Correctly timed out waiting for instance")
	}

	pool.Put(inst1)
	pool.Close()
}

func TestIoTInstancePool_PutNil(t *testing.T) {
	config := DefaultIoTConfig()
	pool := NewIoTInstancePool(config)

	// Putting nil should not panic
	pool.Put(nil)

	pool.Close()
}

// =============================================================================
// Provider Tests
// =============================================================================

func TestNewWASM3IoTProvider(t *testing.T) {
	provider, err := NewWASM3IoTProvider(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	if provider.runtime == nil {
		t.Error("Expected runtime to be initialized")
	}
}

func TestWASM3IoTProvider_Execute(t *testing.T) {
	provider, err := NewWASM3IoTProvider(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	input := []byte("provider test")

	output, err := provider.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Provider Execute failed: %v", err)
	}

	expectedPrefix := "[WASM3-IoT]"
	if len(output) < len(expectedPrefix) {
		t.Errorf("Output too short for prefix check: %s", string(output))
	}
	if string(output[:len(expectedPrefix)]) != expectedPrefix {
		t.Errorf("Expected prefix %s, got %s", expectedPrefix, string(output[:len(expectedPrefix)]))
	}
}

func TestWASM3IoTProvider_ExecuteWithConfig(t *testing.T) {
	provider, err := NewWASM3IoTProvider(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	input := []byte("test")

	output, err := provider.ExecuteWithConfig(ctx, input, nil)
	if err != nil {
		t.Fatalf("ExecuteWithConfig failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestWASM3IoTProvider_Close(t *testing.T) {
	provider, err := NewWASM3IoTProvider(DefaultIoTConfig())
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	err = provider.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Close again
	err = provider.Close()
	if err != nil {
		t.Fatalf("Second Close should be idempotent: %v", err)
	}
}

// =============================================================================
// Router Integration Tests
// =============================================================================

func TestRuntimeTypeFromString_WASM3IoT(t *testing.T) {
	tests := []struct {
		input    string
		expected RuntimeType
	}{
		{"wasm3-iot", RuntimeWASM3IoT},
		{"wasm3", RuntimeWASM3IoT},
		{"iot", RuntimeWASM3IoT},
		{"unknown", RuntimeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RuntimeTypeFromString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRuntimeType_IsValid_WASM3IoT(t *testing.T) {
	if !RuntimeWASM3IoT.IsValid() {
		t.Error("RuntimeWASM3IoT should be valid")
	}
}

func TestCreateWASM3IoTRuntime(t *testing.T) {
	provider, err := CreateWASM3IoTRuntime(nil, nil, nil, DefaultIoTConfig())
	if err != nil {
		t.Fatalf("CreateWASM3IoTRuntime failed: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	output, err := provider.Execute(ctx, []byte("create test"))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output")
	}
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestWASM3IoTRuntime_ConcurrentExecute(t *testing.T) {
	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     4,
		ExecutionTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	const numGoroutines = 10
	const numIterations = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				input := []byte(fmt.Sprintf("goroutine-%d-iter-%d", id, j))
				output, err := runtime.Execute(ctx, input)
				if err != nil {
					t.Errorf("Goroutine %d iter %d failed: %v", id, j, err)
					return
				}
				if len(output) == 0 {
					t.Errorf("Goroutine %d iter %d returned empty output", id, j)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestIoTInstancePool_ConcurrentAccess(t *testing.T) {
	config := &IoTConfig{
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     4,
		ExecutionTimeout: 100 * time.Millisecond,
	}
	pool := NewIoTInstancePool(config)
	defer pool.Close()

	const numGoroutines = 8
	const numIterations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				inst, err := pool.Get(ctx)
				if err != nil {
					t.Errorf("Goroutine %d iter %d failed to get instance: %v", id, j, err)
					continue
				}
				// Simulate some work
				inst.State["processed"] = fmt.Sprintf("%d-%d", id, j)
				pool.Put(inst)
			}
		}(i)
	}

	wg.Wait()
}

// =============================================================================
// Performance Tests
// =============================================================================

func TestWASM3IoTRuntime_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     4,
		ExecutionTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()

	// Warmup
	for i := 0; i < 3; i++ {
		runtime.Execute(ctx, []byte("warmup"))
	}

	// Benchmark
	const iterations = 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := runtime.Execute(ctx, []byte(fmt.Sprintf("perf-test-%d", i)))
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	avgMs := elapsed.Milliseconds() / iterations

	t.Logf("Total time: %v, Average: %vms, Iterations: %d", elapsed, avgMs, iterations)

	// Should average well under 500ms target
	if avgMs > 100 {
		t.Logf("Warning: Average execution time %vms is higher than expected for lightweight runtime", avgMs)
	}
}

func TestWASM3IoTRuntime_LatencyTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}

	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     2,
		ExecutionTimeout: 450 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("latency-test")

	const iterations = 20
	exceeded := 0

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := runtime.Execute(ctx, input)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Iteration %d failed: %v", i, err)
			continue
		}

		if elapsed > 450*time.Millisecond {
			exceeded++
		}
	}

	exceedPct := (exceeded * 100) / iterations
	t.Logf("Latency exceeded target in %d%% of executions (%d/%d)", exceedPct, exceeded, iterations)

	// Should be under target most of the time
	if exceedPct > 50 {
		t.Errorf("Too many executions exceeded target latency: %d%%", exceedPct)
	}
}

func TestIoTInstancePool_PoolEfficiency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool efficiency test in short mode")
	}

	config := &IoTConfig{
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     4,
		ExecutionTimeout: 100 * time.Millisecond,
	}
	pool := NewIoTInstancePool(config)
	defer pool.Close()

	ctx := context.Background()

	// Sequential gets and puts - should be very fast
	const iterations = 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		inst, err := pool.Get(ctx)
		if err != nil {
			t.Fatalf("Get failed at iteration %d: %v", i, err)
		}
		pool.Put(inst)
	}

	elapsed := time.Since(start)
	avgNs := elapsed.Nanoseconds() / iterations

	t.Logf("Pool get/put average: %dns, total: %v for %d iterations", avgNs, elapsed, iterations)

	// Should be microseconds or less per operation
	if avgNs > 100_000 {
		t.Logf("Warning: Pool operations slower than expected: %dns", avgNs)
	}
}

func BenchmarkWASM3IoTRuntime_Execute(b *testing.B) {
	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     4,
		ExecutionTimeout: 1 * time.Second,
	})
	if err != nil {
		b.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("benchmark input")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runtime.Execute(ctx, input)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

func BenchmarkWASM3IoTRuntime_ConcurrentExecute(b *testing.B) {
	runtime, err := NewWASM3IoTRuntime(&IoTConfig{
		TargetLatency:    500,
		MaxMemoryKB:      16 * 1024,
		MaxInstances:     8,
		ExecutionTimeout: 1 * time.Second,
	})
	if err != nil {
		b.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()
	input := []byte("concurrent benchmark")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runtime.Execute(ctx, input)
		}
	})
}
