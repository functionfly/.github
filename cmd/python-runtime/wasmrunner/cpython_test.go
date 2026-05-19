//go:build cgo

package cpython

import (
	"context"
	"testing"
	"time"
)

func TestExecutor_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping executor test in short mode")
	}

	config := DefaultConfig()
	executor := NewExecutor(config)

	result, err := executor.Execute("print(1+1)", nil, 10)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}

	if len(result.Output) == 0 {
		t.Error("Expected output")
	}
}

func TestExecutor_ExecuteWithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping executor test in short mode")
	}

	config := DefaultConfig()
	executor := NewExecutor(config)

	result, err := executor.Execute("import time; time.sleep(0.1); print('done')", nil, 5)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Error != "" && result.Error != "execution timeout after 5 seconds" {
		t.Errorf("Expected timeout error, got: %s", result.Error)
	}
}

func TestExecutor_ExecuteWithInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping executor test in short mode")
	}

	config := DefaultConfig()
	executor := NewExecutor(config)

	input := []byte(`{"name": "test"}`)
	result, err := executor.Execute("import json, sys; data = json.loads(sys.stdin.read()); print(data['name'])", input, 10)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
}

func TestPool_GetPut(t *testing.T) {
	config := DefaultConfig()
	pool := NewPool(config, 2)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	exec1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get executor: %v", err)
	}

	pool.Put(exec1)

	exec2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get executor: %v", err)
	}
	pool.Put(exec2)
}

func TestPool_Exhausted(t *testing.T) {
	config := DefaultConfig()
	pool := NewPool(config, 1)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	exec1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get executor: %v", err)
	}

	exec2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Should create new executor when pool is exhausted: %v", err)
	}

	pool.Put(exec1)
	pool.Put(exec2)
}