package notification

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultReadRetentionDays   = 90
	defaultPendingStaleMinutes = 15
)

// RetentionConfig controls notification cleanup and stale processing recovery.
type RetentionConfig struct {
	ReadRetentionDays   int
	PendingStaleMinutes int
	CleanupInterval     time.Duration
}

// DefaultRetentionConfig returns production-safe defaults.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		ReadRetentionDays:   defaultReadRetentionDays,
		PendingStaleMinutes: defaultPendingStaleMinutes,
		CleanupInterval:     24 * time.Hour,
	}
}

// StartRetentionRoutine periodically deletes old read/archived notifications and resets stale processing rows.
func (s *Service) StartRetentionRoutine(ctx context.Context, cfg RetentionConfig) {
	if cfg.ReadRetentionDays <= 0 {
		cfg.ReadRetentionDays = defaultReadRetentionDays
	}
	if cfg.PendingStaleMinutes <= 0 {
		cfg.PendingStaleMinutes = defaultPendingStaleMinutes
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 24 * time.Hour
	}

	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()

		run := func() {
			cutoff := time.Now().AddDate(0, 0, -cfg.ReadRetentionDays)
			deleted, err := s.repo.CleanupOldNotifications(ctx, cutoff)
			if err != nil {
				s.logger.WithError(err).Warn("notification retention cleanup failed")
			} else if deleted > 0 {
				s.logger.WithField("deleted", deleted).Info("notification retention cleanup completed")
			}

			stale := time.Duration(cfg.PendingStaleMinutes) * time.Minute
			reset, err := s.repo.ResetStaleProcessing(ctx, stale)
			if err != nil {
				s.logger.WithError(err).Warn("notification stale processing reset failed")
			} else if reset > 0 {
				s.logger.WithField("reset", reset).Info("reset stale processing notifications to pending")
				s.RequeuePending(ctx, 500)
			}
		}

		run()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"read_retention_days": cfg.ReadRetentionDays,
		"cleanup_interval":    cfg.CleanupInterval.String(),
	}).Info("Notification retention routine started")
}
