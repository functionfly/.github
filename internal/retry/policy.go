package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// Policy defines retry behavior with exponential backoff and jitter
type Policy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the initial delay before first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier
	Multiplier float64
	// Jitter adds randomness to delay to prevent thundering herd
	Jitter float64
	// RetryableErrors defines which errors should trigger a retry
	RetryableErrors []error
	// RetryBudget is the maximum number of retries per time window
	RetryBudget *RetryBudget
}

// RetryBudget tracks retry budget to prevent excessive retries
type RetryBudget struct {
	// MaxRetriesPerWindow is the maximum retries allowed in the time window
	MaxRetriesPerWindow int
	// Window is the time window for budget tracking
	Window time.Duration
	// currentRetries tracks retries in the current window
	currentRetries int
	// windowStart tracks when the current window started
	windowStart time.Time
	// mutex protects concurrent access
	mutex chan struct{}
}

// DefaultPolicy returns a production-ready retry policy
func DefaultPolicy() *Policy {
	return &Policy{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryBudget: &RetryBudget{
			MaxRetriesPerWindow: 100,
			Window:              1 * time.Minute,
			currentRetries:      0,
			windowStart:         time.Now(),
			mutex:               make(chan struct{}, 1),
		},
	}
}

// AggressivePolicy returns a more aggressive retry policy for critical operations
func AggressivePolicy() *Policy {
	return &Policy{
		MaxRetries:   5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   1.5,
		Jitter:       0.2,
		RetryBudget: &RetryBudget{
			MaxRetriesPerWindow: 200,
			Window:              1 * time.Minute,
			currentRetries:      0,
			windowStart:         time.Now(),
			mutex:               make(chan struct{}, 1),
		},
	}
}

// ConservativePolicy returns a conservative retry policy for non-critical operations
func ConservativePolicy() *Policy {
	return &Policy{
		MaxRetries:   2,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryBudget: &RetryBudget{
			MaxRetriesPerWindow: 50,
			Window:              1 * time.Minute,
			currentRetries:      0,
			windowStart:         time.Now(),
			mutex:               make(chan struct{}, 1),
		},
	}
}

// Execute executes a function with retry logic
func (p *Policy) Execute(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// Check retry budget
		if p.RetryBudget != nil && attempt > 0 {
			if !p.RetryBudget.CanRetry() {
				logrus.WithFields(logrus.Fields{
					"operation": operation,
					"attempt":   attempt,
				}).Warn("Retry budget exhausted")
				return fmt.Errorf("retry budget exhausted for operation %s", operation)
			}
			p.RetryBudget.RecordRetry()
		}

		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 0 {
				logrus.WithFields(logrus.Fields{
					"operation": operation,
					"attempt":   attempt,
				}).Info("Operation succeeded after retry")
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			logrus.WithFields(logrus.Fields{
				"operation": operation,
				"attempt":   attempt,
				"error":     err,
			}).Debug("Error is not retryable")
			return err
		}

		// Don't sleep after last attempt
		if attempt == p.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := p.calculateDelay(attempt)

		logrus.WithFields(logrus.Fields{
			"operation": operation,
			"attempt":   attempt,
			"delay":     delay.String(),
			"error":     err,
		}).Debug("Retrying operation")

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	logrus.WithFields(logrus.Fields{
		"operation":  operation,
		"max_retries": p.MaxRetries,
		"error":      lastErr,
	}).Error("Operation failed after all retries")

	return fmt.Errorf("operation %s failed after %d retries: %w", operation, p.MaxRetries, lastErr)
}

// calculateDelay calculates the delay for a given attempt with exponential backoff and jitter
func (p *Policy) calculateDelay(attempt int) time.Duration {
	// Exponential backoff: delay = initial * multiplier^attempt
	delay := float64(p.InitialDelay) * math.Pow(p.Multiplier, float64(attempt))

	// Apply jitter
	if p.Jitter > 0 {
		jitter := delay * p.Jitter * (rand.Float64()*2 - 1) // Random between -jitter and +jitter
		delay += jitter
	}

	// Cap at max delay
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}

	return time.Duration(delay)
}

// isRetryable checks if an error should trigger a retry
func (p *Policy) isRetryable(err error) bool {
	if len(p.RetryableErrors) == 0 {
		// If no specific errors defined, retry all errors
		return true
	}

	for _, retryableErr := range p.RetryableErrors {
		if err == retryableErr {
			return true
		}
	}

	return false
}

// CanRetry checks if the retry budget allows another retry
func (rb *RetryBudget) CanRetry() bool {
	rb.mutex <- struct{}{}
	defer func() { <-rb.mutex }()

	now := time.Now()

	// Reset window if expired
	if now.Sub(rb.windowStart) > rb.Window {
		rb.currentRetries = 0
		rb.windowStart = now
	}

	return rb.currentRetries < rb.MaxRetriesPerWindow
}

// RecordRetry records a retry attempt
func (rb *RetryBudget) RecordRetry() {
	rb.mutex <- struct{}{}
	defer func() { <-rb.mutex }()

	now := time.Now()

	// Reset window if expired
	if now.Sub(rb.windowStart) > rb.Window {
		rb.currentRetries = 0
		rb.windowStart = now
	}

	rb.currentRetries++
}

// GetRetryStats returns current retry budget statistics
func (rb *RetryBudget) GetRetryStats() (current int, max int, windowStart time.Time) {
	rb.mutex <- struct{}{}
	defer func() { <-rb.mutex }()

	return rb.currentRetries, rb.MaxRetriesPerWindow, rb.windowStart
}
