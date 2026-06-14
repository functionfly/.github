package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AuthEventCleanupService handles periodic cleanup of old authentication events
type AuthEventCleanupService struct {
	repo   Repository
	logger *logrus.Logger
	stop   chan struct{}
}

// NewAuthEventCleanupService creates a new auth event cleanup service
func NewAuthEventCleanupService(repo Repository) *AuthEventCleanupService {
	return &AuthEventCleanupService{
		repo:   repo,
		logger: logrus.New(),
		stop:   make(chan struct{}),
	}
}

// CleanupOldAuthEvents removes authentication events older than the specified retention period
func (s *AuthEventCleanupService) CleanupOldAuthEvents(ctx context.Context, retentionPeriod time.Duration) error {
	cutoff := time.Now().Add(-retentionPeriod)

	start := time.Now()
	deletedCount, err := s.repo.DeleteOldAuthEvents(ctx, cutoff)
	if err != nil {
		s.logger.WithError(err).Error("Failed to cleanup old auth events")
		return err
	}

	duration := time.Since(start)
	if deletedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_count":    deletedCount,
			"retention_period": retentionPeriod.String(),
			"duration_ms":      duration.Milliseconds(),
		}).Info("Cleaned up old auth events")
	} else {
		s.logger.WithField("duration_ms", duration.Milliseconds()).Debug("Auth event cleanup completed (no old events to clean)")
	}

	return nil
}

// StartCleanupRoutine starts a periodic cleanup routine
func (s *AuthEventCleanupService) StartCleanupRoutine(ctx context.Context, interval time.Duration, retentionPeriod time.Duration) {
	if interval <= 0 {
		interval = 24 * time.Hour // Default to daily cleanup
	}
	if retentionPeriod <= 0 {
		retentionPeriod = 90 * 24 * time.Hour // Default to 90 days retention for auth events
	}

	s.logger.WithFields(logrus.Fields{
		"interval":         interval.String(),
		"retention_period": retentionPeriod.String(),
	}).Info("Starting auth event cleanup routine")

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.WithFields(logrus.Fields{
					"panic": rec,
					"stack": fmt.Sprintf("%v", rec),
				}).Error("Auth event cleanup goroutine panicked")
			}
		}()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run cleanup immediately on startup
		if err := s.CleanupOldAuthEvents(ctx, retentionPeriod); err != nil {
			s.logger.WithError(err).Error("Initial auth event cleanup failed")
		}

		for {
			select {
			case <-ticker.C:
				if err := s.CleanupOldAuthEvents(ctx, retentionPeriod); err != nil {
					s.logger.WithError(err).Error("Periodic auth event cleanup failed")
				}
			case <-s.stop:
				s.logger.Info("Auth event cleanup routine stopping")
				return
			}
		}
	}()
}

// Stop stops the cleanup routine for graceful shutdown
func (s *AuthEventCleanupService) Stop() {
	close(s.stop)
}