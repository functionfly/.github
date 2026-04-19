package scheduler

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// ExpiredEvaluationScheduler cleans up old cached trust policy evaluations
type ExpiredEvaluationScheduler struct {
	cron            *cron.Cron
	revocationRepo  *trustapi.RevocationRepository
	logger          *logrus.Logger

	// Configuration
	CronExpression string
	Enabled        bool
	MaxAge         time.Duration // Evaluations older than this are deleted
}

// ExpiredEvaluationMetrics tracks scheduler performance
type ExpiredEvaluationMetrics struct {
	TotalRuns      int64
	SuccessfulRuns int64
	FailedRuns     int64
	LastRunTime    time.Time
	LastRunErr     string
}

// NewExpiredEvaluationScheduler creates a new expired evaluation cleanup scheduler
func NewExpiredEvaluationScheduler(revocationRepo *trustapi.RevocationRepository) *ExpiredEvaluationScheduler {
	return &ExpiredEvaluationScheduler{
		cron:           cron.New(),
		revocationRepo: revocationRepo,
		logger:         logrus.WithField("scheduler", "expired_evaluation").Logger,
		CronExpression: "0 */6 * * *", // Default: every 6 hours
		Enabled:        true,
		MaxAge:         24 * time.Hour, // Default: evaluations older than 24 hours
	}
}

// Start begins the scheduled cleanup job
func (s *ExpiredEvaluationScheduler) Start(ctx context.Context) error {
	if !s.Enabled {
		s.logger.Info("Expired evaluation scheduler is disabled")
		return nil
	}

	_, err := s.cron.AddFunc(s.CronExpression, func() {
		s.runScheduledCleanup(ctx)
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.WithFields(logrus.Fields{
		"cron":    s.CronExpression,
		"max_age": s.MaxAge,
	}).Info("Expired evaluation scheduler started")

	return nil
}

// Stop halts the scheduler
func (s *ExpiredEvaluationScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Expired evaluation scheduler stopped")
}

// RunOnce executes a single cleanup run manually
func (s *ExpiredEvaluationScheduler) RunOnce(ctx context.Context) error {
	return s.runScheduledCleanup(ctx)
}

// runScheduledCleanup performs the expired evaluation cleanup
func (s *ExpiredEvaluationScheduler) runScheduledCleanup(ctx context.Context) error {
	cutoff := time.Now().Add(-s.MaxAge)
	s.logger.WithField("cutoff", cutoff).Info("Running expired evaluation cleanup")

	if err := s.revocationRepo.CleanExpiredEvaluations(cutoff); err != nil {
		s.logger.WithError(err).Error("Failed to clean expired evaluations")
		return err
	}

	s.logger.Info("Expired evaluation cleanup completed")
	return nil
}
