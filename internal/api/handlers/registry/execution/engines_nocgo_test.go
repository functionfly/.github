//go:build !cgo

package execution

import (
	"context"
	"testing"
	"time"
)

func TestCircuitBreaker_Allow(t *testing.T) {
	cb := newCircuitBreaker(3, 1*time.Second)

	if !cb.allow() {
		t.Error("Circuit breaker should allow requests when closed")
	}
}

func TestCircuitBreaker_Opens(t *testing.T) {
	cb := newCircuitBreaker(3, 1*time.Hour)

	for i := 0; i < 3; i++ {
		cb.recordFailure()
	}

	if cb.allow() {
		t.Error("Circuit breaker should open after threshold failures")
	}
}

func TestCircuitBreaker_Resets(t *testing.T) {
	cb := newCircuitBreaker(2, 50*time.Millisecond)

	cb.recordFailure()
	cb.recordFailure()

	if cb.allow() {
		t.Error("Circuit breaker should be open")
	}

	time.Sleep(60 * time.Millisecond)

	if !cb.allow() {
		t.Error("Circuit breaker should reset after timeout")
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	cb := newCircuitBreaker(2, 1*time.Hour)

	cb.recordFailure()
	cb.recordFailure()
	cb.recordSuccess()

	if cb.failures != 0 {
		t.Errorf("Expected failures to be reset, got %d", cb.failures)
	}
}

func TestRuntimeClient_CircuitBreaker(t *testing.T) {
	client := NewRuntimeClient("http://localhost:8083")

	if !client.circuitBreaker.allow() {
		t.Error("Client should allow requests initially")
	}

	client.circuitBreaker.recordFailure()
	client.circuitBreaker.recordFailure()
	client.circuitBreaker.recordFailure()
	client.circuitBreaker.recordFailure()
	client.circuitBreaker.recordFailure()

	if client.circuitBreaker.allow() {
		t.Error("Circuit breaker should be open after threshold failures")
	}
}

func TestRuntimeClient_Execute_CircuitOpen(t *testing.T) {
	client := NewRuntimeClient("http://localhost:8083")

	for i := 0; i < 5; i++ {
		client.circuitBreaker.recordFailure()
	}

	_, err := client.Execute(context.Background(), "print(1)", nil, 5000)
	if err == nil {
		t.Error("Expected error when circuit breaker is open")
	}
}

func TestRuntimeClient_Closed(t *testing.T) {
	client := NewRuntimeClient("http://localhost:8083")
	client.Close()

	_, err := client.Execute(context.Background(), "print(1)", nil, 5000)
	if err == nil {
		t.Error("Expected error when client is closed")
	}
}

func TestRuntimeClient_Healthy(t *testing.T) {
	client := NewRuntimeClient("http://localhost:8083")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if client.Healthy(ctx) {
		t.Error("Should not be healthy for non-existent endpoint")
	}
}
