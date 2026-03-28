package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TransactionManager provides enhanced transaction management with timeouts and saga patterns
type TransactionManager struct {
	db *PostgresDB
}

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	Timeout       time.Duration
	IsolationLevel sql.IsolationLevel
	ReadOnly      bool
}

// DefaultTransactionOptions returns sensible defaults
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		Timeout:       30 * time.Second,
		IsolationLevel: sql.LevelReadCommitted,
		ReadOnly:      false,
	}
}

// TransactionContext holds transaction state and metadata
type TransactionContext struct {
	ID        uuid.UUID
	StartTime time.Time
	Timeout   time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *PostgresDB) *TransactionManager {
	return &TransactionManager{db: db}
}

// ExecuteInTransaction executes a function within a transaction with timeout
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, opts *TransactionOptions, fn func(*gorm.DB) error) error {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	// Create transaction context with timeout
	txCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	txID := uuid.New()
	startTime := time.Now()

	logrus.WithFields(logrus.Fields{
		"tx_id":          txID,
		"timeout":        opts.Timeout,
		"isolation_level": opts.IsolationLevel,
		"read_only":      opts.ReadOnly,
	}).Debug("Starting transaction")

	// Begin transaction with options
	tx := tm.db.GORM.Begin(&sql.TxOptions{
		Isolation: opts.IsolationLevel,
		ReadOnly:  opts.ReadOnly,
	})

	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logrus.WithFields(logrus.Fields{
				"tx_id": txID,
				"panic": r,
			}).Error("Transaction panicked, rolled back")
			panic(r) // Re-panic after cleanup
		}
	}()

	// Execute the transaction function
	err := fn(tx)
	duration := time.Since(startTime)

	if err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			logrus.WithFields(logrus.Fields{
				"tx_id":    txID,
				"duration": duration,
				"error":    err,
				"rollback_error": rollbackErr,
			}).Error("Transaction failed and rollback also failed")
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}

		logrus.WithFields(logrus.Fields{
			"tx_id":    txID,
			"duration": duration,
			"error":    err,
		}).Warn("Transaction rolled back due to error")

		return err
	}

	// Check for context timeout
	if txCtx.Err() == context.DeadlineExceeded {
		tx.Rollback()
		logrus.WithFields(logrus.Fields{
			"tx_id":    txID,
			"duration": duration,
		}).Warn("Transaction timed out, rolled back")
		return fmt.Errorf("transaction timeout after %v", duration)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"tx_id":    txID,
			"duration": duration,
			"error":    err,
		}).Error("Transaction commit failed")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"tx_id":    txID,
		"duration": duration,
	}).Debug("Transaction committed successfully")

	return nil
}

// SagaStep represents a step in a saga transaction
type SagaStep struct {
	Name      string
	Execute   func(*gorm.DB) error
	Compensate func(*gorm.DB) error // Function to undo the step if later steps fail
}

// ExecuteSaga executes a saga pattern transaction
func (tm *TransactionManager) ExecuteSaga(ctx context.Context, opts *TransactionOptions, steps []SagaStep) error {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	sagaID := uuid.New()
	startTime := time.Now()

	logrus.WithFields(logrus.Fields{
		"saga_id": sagaID,
		"steps":   len(steps),
	}).Info("Starting saga transaction")

	// Track completed steps for compensation
	completedSteps := make([]SagaStep, 0, len(steps))

	for i, step := range steps {
		stepStart := time.Now()

		logrus.WithFields(logrus.Fields{
			"saga_id":  sagaID,
			"step":     i + 1,
			"step_name": step.Name,
		}).Debug("Executing saga step")

		// Execute step in its own transaction
		err := tm.ExecuteInTransaction(ctx, opts, step.Execute)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"saga_id":  sagaID,
				"step":     i + 1,
				"step_name": step.Name,
				"error":     err,
				"duration":  time.Since(stepStart),
			}).Error("Saga step failed, starting compensation")

			// Compensate completed steps in reverse order
			for j := len(completedSteps) - 1; j >= 0; j-- {
				compStep := completedSteps[j]
				logrus.WithFields(logrus.Fields{
					"saga_id":      sagaID,
					"compensating": compStep.Name,
				}).Info("Compensating saga step")

				if compErr := tm.ExecuteInTransaction(ctx, opts, compStep.Compensate); compErr != nil {
					logrus.WithFields(logrus.Fields{
						"saga_id":      sagaID,
						"step_name":     compStep.Name,
						"compensation_error": compErr,
					}).Error("Compensation failed")
				}
			}

			return fmt.Errorf("saga failed at step %d (%s): %w", i+1, step.Name, err)
		}

		// Step completed successfully
		completedSteps = append(completedSteps, step)

		logrus.WithFields(logrus.Fields{
			"saga_id":  sagaID,
			"step":     i + 1,
			"step_name": step.Name,
			"duration":  time.Since(stepStart),
		}).Debug("Saga step completed")
	}

	totalDuration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"saga_id":  sagaID,
		"steps":    len(steps),
		"duration": totalDuration,
	}).Info("Saga transaction completed successfully")

	return nil
}

// ExecuteInReadTransaction executes read-only operations with snapshot isolation
func (tm *TransactionManager) ExecuteInReadTransaction(ctx context.Context, timeout time.Duration, fn func(*gorm.DB) error) error {
	opts := &TransactionOptions{
		Timeout:       timeout,
		IsolationLevel: sql.LevelRepeatableRead, // Higher isolation for reads
		ReadOnly:      true,
	}

	return tm.ExecuteInTransaction(ctx, opts, fn)
}

// ExecuteWithRetry executes a transaction with retry logic for transient failures
func (tm *TransactionManager) ExecuteWithRetry(ctx context.Context, opts *TransactionOptions, maxRetries int, fn func(*gorm.DB) error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := tm.ExecuteInTransaction(ctx, opts, fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable (serialization failures, deadlocks, etc.)
		if !isRetryableError(err) {
			break // Don't retry non-retryable errors
		}

		if attempt < maxRetries {
			// Exponential backoff
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			logrus.WithFields(logrus.Fields{
				"attempt": attempt + 1,
				"backoff": backoff,
				"error":   err,
			}).Warn("Transaction failed, retrying")

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError checks if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Common retryable PostgreSQL errors
	retryableErrors := []string{
		"serialization_failure",
		"deadlock_detected",
		"lock_not_available",
		"connection_reset",
		"connection_refused",
	}

	for _, retryable := range retryableErrors {
		if containsString(errStr, retryable) {
			return true
		}
	}

	return false
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Iterative search to avoid stack overflow on large inputs
	// Convert to lowercase once for case-insensitive comparison
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// TransactionScope provides a fluent interface for transaction configuration
type TransactionScope struct {
	manager *TransactionManager
	ctx     context.Context
	opts    *TransactionOptions
}

// NewTransactionScope creates a new transaction scope builder
func (tm *TransactionManager) NewTransactionScope(ctx context.Context) *TransactionScope {
	return &TransactionScope{
		manager: tm,
		ctx:     ctx,
		opts:    DefaultTransactionOptions(),
	}
}

// WithTimeout sets the transaction timeout
func (ts *TransactionScope) WithTimeout(timeout time.Duration) *TransactionScope {
	ts.opts.Timeout = timeout
	return ts
}

// WithIsolationLevel sets the isolation level
func (ts *TransactionScope) WithIsolationLevel(level sql.IsolationLevel) *TransactionScope {
	ts.opts.IsolationLevel = level
	return ts
}

// ReadOnly sets the transaction to read-only
func (ts *TransactionScope) ReadOnly() *TransactionScope {
	ts.opts.ReadOnly = true
	return ts
}

// Execute runs the transaction with the configured options
func (ts *TransactionScope) Execute(fn func(*gorm.DB) error) error {
	return ts.manager.ExecuteInTransaction(ts.ctx, ts.opts, fn)
}