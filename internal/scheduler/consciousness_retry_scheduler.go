package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/consciousness"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

const (
	DefaultRetryMaxAttempts = 3
	DefaultRetryBaseDelay   = 5 * time.Minute
	DefaultRetryMaxDelay    = 1 * time.Hour
)

// RetrySchedulerConfig holds configuration for the retry scheduler.
type RetrySchedulerConfig struct {
	Enabled       bool
	Cron          string
	BatchSize     int
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
}

// DefaultRetrySchedulerConfig returns default retry scheduler configuration.
func DefaultRetrySchedulerConfig() *RetrySchedulerConfig {
	return &RetrySchedulerConfig{
		Enabled:     true,
		Cron:        "*/5 * * * *", // Every 5 minutes
		BatchSize:   100,
		MaxAttempts: DefaultRetryMaxAttempts,
		BaseDelay:   DefaultRetryBaseDelay,
		MaxDelay:    DefaultRetryMaxDelay,
	}
}

// LoadRetrySchedulerConfig loads retry configuration from environment.
func LoadRetrySchedulerConfig() *RetrySchedulerConfig {
	config := DefaultRetrySchedulerConfig()

	if v := os.Getenv("CONSCIOUSNESS_RETRY_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_RETRY_CRON"); v != "" {
		config.Cron = v
	}
	if v := os.Getenv("CONSCIOUSNESS_RETRY_BATCH_SIZE"); v != "" {
		if batch, err := strconv.Atoi(v); err == nil && batch > 0 {
			config.BatchSize = batch
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_RETRY_MAX_ATTEMPTS"); v != "" {
		if max, err := strconv.Atoi(v); err == nil && max > 0 {
			config.MaxAttempts = max
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_RETRY_BASE_DELAY"); v != "" {
		if delay, err := time.ParseDuration(v); err == nil {
			config.BaseDelay = delay
		}
	}
	if v := os.Getenv("CONSCIOUSNESS_RETRY_MAX_DELAY"); v != "" {
		if delay, err := time.ParseDuration(v); err == nil {
			config.MaxDelay = delay
		}
	}

	return config
}

// RetryScheduler handles retry processing for failed consciousness deliveries.
type RetryScheduler struct {
	cron       *cron.Cron
	dispatcher *consciousness.NotificationDispatcher
	repo       *consciousness.Repository
	logger     *logrus.Logger
	config     *RetrySchedulerConfig
	stopOnce   sync.Once
	cancel     context.CancelFunc
	httpClient *HTTPClient
}

// NewRetryScheduler creates a new consciousness retry scheduler.
func NewRetryScheduler(db *sql.DB, logger *logrus.Logger) *RetryScheduler {
	repo := consciousness.NewRepository(db, logger)
	dispatcher := consciousness.NewNotificationDispatcher(db, repo, logger)
	return &RetryScheduler{
		cron:       cron.New(),
		dispatcher: dispatcher,
		repo:       repo,
		logger:     logger,
		config:     LoadRetrySchedulerConfig(),
	}
}

// NewRetrySchedulerWithConfig creates a retry scheduler with custom configuration.
func NewRetrySchedulerWithConfig(db *sql.DB, logger *logrus.Logger, config *RetrySchedulerConfig) *RetryScheduler {
	repo := consciousness.NewRepository(db, logger)
	dispatcher := consciousness.NewNotificationDispatcher(db, repo, logger)
	return &RetryScheduler{
		cron:       cron.New(),
		dispatcher: dispatcher,
		repo:       repo,
		logger:     logger,
		config:     config,
	}
}

// Start begins the retry scheduler.
func (s *RetryScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Consciousness retry scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid retry cron expression: %w", err)
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.processRetries(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add retry cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":     s.config.Cron,
		"batch_size": s.config.BatchSize,
		"max_attempts": s.config.MaxAttempts,
	}).Info("Consciousness retry scheduler started")

	return nil
}

// Stop stops the retry scheduler.
func (s *RetryScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Consciousness retry scheduler stopped")
	})
	return nil
}

// processRetries processes due retries from the dead letter queue.
func (s *RetryScheduler) processRetries(ctx context.Context) {
	retries, err := s.repo.GetDueRetries(ctx, s.config.BatchSize)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch due retries")
		return
	}

	if len(retries) == 0 {
		return
	}

	s.logger.WithField("count", len(retries)).Info("Processing consciousness delivery retries")

	successCount := 0
	failCount := 0

	for _, retry := range retries {
		// Check if we've exceeded max attempts
		if retry.AttemptCount >= s.config.MaxAttempts {
			if err := s.repo.MarkRetryExhausted(ctx, retry.ID); err != nil {
				s.logger.WithError(err).WithField("retry_id", retry.ID).Error("Failed to mark retry as exhausted")
			}
			failCount++
			continue
		}

		// Attempt delivery
		err := s.dispatcher.RetryDelivery(ctx, retry)
		if err != nil {
			// Schedule next retry with exponential backoff
			delay := s.calculateBackoff(retry.AttemptCount)
			if err := s.repo.UpdateRetryNextAttempt(ctx, retry.ID, retry.AttemptCount+1, delay, err.Error()); err != nil {
				s.logger.WithError(err).WithField("retry_id", retry.ID).Error("Failed to update retry")
			}
			failCount++
		} else {
			// Success - mark as completed
			if err := s.repo.MarkRetryCompleted(ctx, retry.ID); err != nil {
				s.logger.WithError(err).WithField("retry_id", retry.ID).Error("Failed to mark retry as completed")
			}
			successCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"processed": len(retries),
		"success":   successCount,
		"failed":    failCount,
	}).Info("Consciousness delivery retry processing completed")
}

// calculateBackoff calculates the next retry delay using exponential backoff.
func (s *RetryScheduler) calculateBackoff(attemptCount int) time.Duration {
	delay := s.config.BaseDelay * time.Duration(1<<uint(attemptCount))
	if delay > s.config.MaxDelay {
		return s.config.MaxDelay
	}
	return delay
}

// GetStatus returns the current retry scheduler status.
func (s *RetryScheduler) GetStatus() map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":      s.config.Enabled,
		"cron":         s.config.Cron,
		"next_run":     nextRun,
		"batch_size":   s.config.BatchSize,
		"max_attempts": s.config.MaxAttempts,
	}
}
