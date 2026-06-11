package statefabric

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*RetryableError)
	return ok
}

func RetryWithBackoff(ctx context.Context, config RetryConfig, operation string, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err
			case <-time.After(delay):
			}

			delay = time.Duration(float64(delay) * config.BackoffMultiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !IsRetryable(lastErr) {
			return lastErr
		}

		logrus.WithFields(logrus.Fields{
			"operation": operation,
			"attempt":   attempt + 1,
			"maxAttempts": config.MaxAttempts,
			"error":    lastErr.Error(),
		}).Warn("Retryable operation failed, will retry")
	}

	return fmt.Errorf("operation %s failed after %d attempts: %w", operation, config.MaxAttempts, lastErr)
}