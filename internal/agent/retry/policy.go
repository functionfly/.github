package retry

import (
	"context"
	"math"
	"time"
)

// Policy defines retry behavior
type Policy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// BackoffFactor is the multiplier for exponential backoff
	BackoffFactor float64
	// RetryableErrors is a list of errors that should trigger a retry
	RetryableErrors []error
	// OnRetry is called before each retry attempt
	OnRetry func(attempt int, err error, delay time.Duration)
}

// DefaultPolicy returns a sensible default retry policy
func DefaultPolicy() Policy {
	return Policy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Execute runs the given function with retry logic
func (p *Policy) Execute(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// Don't delay on first attempt
		if attempt > 0 {
			delay := p.calculateDelay(attempt)
			if p.OnRetry != nil {
				p.OnRetry(attempt, lastErr, delay)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !p.isRetryable(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

// ExecuteWithResult runs the given function with retry logic and returns a result
func ExecuteWithResult[T any](ctx context.Context, policy Policy, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// Don't delay on first attempt
		if attempt > 0 {
			delay := policy.calculateDelay(attempt)
			if policy.OnRetry != nil {
				policy.OnRetry(attempt, lastErr, delay)
			}

			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !policy.isRetryable(err) {
			return zero, err
		}
	}

	return zero, lastErr
}

func (p *Policy) calculateDelay(attempt int) time.Duration {
	delay := float64(p.InitialDelay) * math.Pow(p.BackoffFactor, float64(attempt-1))
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}
	return time.Duration(delay)
}

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

// WithMaxRetries sets the maximum number of retries
func (p Policy) WithMaxRetries(max int) Policy {
	p.MaxRetries = max
	return p
}

// WithInitialDelay sets the initial delay
func (p Policy) WithInitialDelay(delay time.Duration) Policy {
	p.InitialDelay = delay
	return p
}

// WithMaxDelay sets the maximum delay
func (p Policy) WithMaxDelay(delay time.Duration) Policy {
	p.MaxDelay = delay
	return p
}

// WithBackoffFactor sets the backoff factor
func (p Policy) WithBackoffFactor(factor float64) Policy {
	p.BackoffFactor = factor
	return p
}

// WithRetryableErrors sets the retryable errors
func (p Policy) WithRetryableErrors(errs ...error) Policy {
	p.RetryableErrors = errs
	return p
}

// WithOnRetry sets the retry callback
func (p Policy) WithOnRetry(fn func(attempt int, err error, delay time.Duration)) Policy {
	p.OnRetry = fn
	return p
}
