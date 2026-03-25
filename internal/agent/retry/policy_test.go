package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPolicy_RetryOnTransientError(t *testing.T) {
	attempts := 0
	policy := Policy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	err := policy.Execute(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.New("transient")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPolicy_NoRetryOnNonRetryableError(t *testing.T) {
	attempts := 0
	nonRetryableErr := errors.New("non-retryable")

	policy := Policy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		RetryableErrors: []error{
			errors.New("retryable"),
		},
	}

	err := policy.Execute(context.Background(), func() error {
		attempts++
		return nonRetryableErr
	})

	if err != nonRetryableErr {
		t.Errorf("expected non-retryable error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestPolicy_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	policy := Policy{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	err := policy.Execute(context.Background(), func() error {
		attempts++
		return errors.New("always fail")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPolicy_ContextCancellation(t *testing.T) {
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())

	policy := Policy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
	}

	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := policy.Execute(ctx, func() error {
		attempts++
		return errors.New("fail")
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPolicy_CalculateDelay(t *testing.T) {
	policy := Policy{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1600 * time.Millisecond},
		{6, 3200 * time.Millisecond},
		{7, 5 * time.Second}, // capped at max
	}

	for _, tt := range tests {
		delay := policy.calculateDelay(tt.attempt)
		if delay != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
		}
	}
}

func TestPolicy_OnRetry(t *testing.T) {
	var retryAttempts []int
	var retryErrors []error
	var retryDelays []time.Duration

	policy := Policy{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		OnRetry: func(attempt int, err error, delay time.Duration) {
			retryAttempts = append(retryAttempts, attempt)
			retryErrors = append(retryErrors, err)
			retryDelays = append(retryDelays, delay)
		},
	}

	_ = policy.Execute(context.Background(), func() error {
		return errors.New("fail")
	})

	if len(retryAttempts) != 2 {
		t.Errorf("expected 2 retries, got %d", len(retryAttempts))
	}
	if retryAttempts[0] != 1 {
		t.Errorf("expected first retry attempt 1, got %d", retryAttempts[0])
	}
	if retryAttempts[1] != 2 {
		t.Errorf("expected second retry attempt 2, got %d", retryAttempts[1])
	}
}

func TestExecuteWithResult(t *testing.T) {
	attempts := 0
	policy := Policy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	result, err := ExecuteWithResult(context.Background(), policy, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("transient")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected success, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPolicy_WithMethods(t *testing.T) {
	policy := DefaultPolicy()

	policy = policy.
		WithMaxRetries(5).
		WithInitialDelay(200 * time.Millisecond).
		WithMaxDelay(10 * time.Second).
		WithBackoffFactor(3.0)

	if policy.MaxRetries != 5 {
		t.Errorf("expected MaxRetries=5, got %d", policy.MaxRetries)
	}
	if policy.InitialDelay != 200*time.Millisecond {
		t.Errorf("expected InitialDelay=200ms, got %v", policy.InitialDelay)
	}
	if policy.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay=10s, got %v", policy.MaxDelay)
	}
	if policy.BackoffFactor != 3.0 {
		t.Errorf("expected BackoffFactor=3.0, got %f", policy.BackoffFactor)
	}
}
