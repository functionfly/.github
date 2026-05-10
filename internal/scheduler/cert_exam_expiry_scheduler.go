package scheduler

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CertExamExpiryScheduler marks in-progress exams as expired when their time limit is up
type CertExamExpiryScheduler struct {
	cron            *cron.Cron
	certRepo        *storage.CertificationRepository
	logger          *logrus.Logger
	CronExpression  string
	Enabled         bool
}

// NewCertExamExpiryScheduler creates a new exam expiry scheduler
func NewCertExamExpiryScheduler(certRepo *storage.CertificationRepository) *CertExamExpiryScheduler {
	return &CertExamExpiryScheduler{
		cron:           cron.New(),
		certRepo:       certRepo,
		logger:         logrus.WithField("scheduler", "cert_exam_expiry").Logger,
		CronExpression: "*/5 * * * *", // every 5 minutes
		Enabled:        true,
	}
}

// Start begins the scheduled cleanup job
func (s *CertExamExpiryScheduler) Start(ctx context.Context) error {
	if !s.Enabled {
		s.logger.Info("Cert exam expiry scheduler is disabled")
		return nil
	}

	_, err := s.cron.AddFunc(s.CronExpression, func() {
		s.run(ctx)
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.WithField("cron", s.CronExpression).Info("Cert exam expiry scheduler started")
	return nil
}

// Stop halts the scheduler
func (s *CertExamExpiryScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Cert exam expiry scheduler stopped")
}

func (s *CertExamExpiryScheduler) run(ctx context.Context) {
	count, err := s.certRepo.ExpireStaleExams(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to expire stale exams")
		return
	}
	if count > 0 {
		s.logger.WithField("count", count).Info("Expired stale exam sessions")
	}
}
