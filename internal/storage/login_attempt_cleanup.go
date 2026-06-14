package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// LoginAttemptCleanupService handles periodic cleanup of old login attempts
type LoginAttemptCleanupService struct {
	repo   Repository
	logger *logrus.Logger
	stop   chan struct{}
}

// NewLoginAttemptCleanupService creates a new login attempt cleanup service
func NewLoginAttemptCleanupService(repo Repository) *LoginAttemptCleanupService {
	return &LoginAttemptCleanupService{
		repo:   repo,
		logger: logrus.New(),
		stop:   make(chan struct{}),
	}
}

// CleanupOldLoginAttempts removes login attempts older than the specified retention period
func (s *LoginAttemptCleanupService) CleanupOldLoginAttempts(ctx context.Context, retentionPeriod time.Duration) error {
	cutoff := time.Now().Add(-retentionPeriod)

	start := time.Now()
	deletedCount, err := s.repo.DeleteOldLoginAttempts(ctx, cutoff)
	if err != nil {
		s.logger.WithError(err).Error("Failed to cleanup old login attempts")
		return err
	}

	duration := time.Since(start)
	if deletedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_count":    deletedCount,
			"retention_period": retentionPeriod.String(),
			"duration_ms":      duration.Milliseconds(),
		}).Info("Cleaned up old login attempts")
	} else {
		s.logger.WithField("duration_ms", duration.Milliseconds()).Debug("Login attempt cleanup completed (no old attempts to clean)")
	}

	return nil
}

// StartCleanupRoutine starts a periodic cleanup routine
func (s *LoginAttemptCleanupService) StartCleanupRoutine(ctx context.Context, interval time.Duration, retentionPeriod time.Duration) {
	if interval <= 0 {
		interval = 24 * time.Hour // Default to daily cleanup
	}
	if retentionPeriod <= 0 {
		retentionPeriod = 30 * 24 * time.Hour // Default to 30 days retention
	}

	s.logger.WithFields(logrus.Fields{
		"interval":         interval.String(),
		"retention_period": retentionPeriod.String(),
	}).Info("Starting login attempt cleanup routine")

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.WithFields(logrus.Fields{
					"panic": rec,
					"stack": fmt.Sprintf("%v", rec),
				}).Error("Login attempt cleanup goroutine panicked")
			}
		}()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run cleanup immediately on startup
		if err := s.CleanupOldLoginAttempts(ctx, retentionPeriod); err != nil {
			s.logger.WithError(err).Error("Initial login attempt cleanup failed")
		}

		for {
			select {
			case <-ticker.C:
				if err := s.CleanupOldLoginAttempts(ctx, retentionPeriod); err != nil {
					s.logger.WithError(err).Error("Periodic login attempt cleanup failed")
				}
			case <-s.stop:
				s.logger.Info("Login attempt cleanup routine stopping")
				return
			}
		}
	}()
}

// Stop stops the cleanup routine for graceful shutdown
func (s *LoginAttemptCleanupService) Stop() {
	close(s.stop)
}