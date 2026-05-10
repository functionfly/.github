package scheduler

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CertCredentialExpiryScheduler marks credentials as expired past their expiry date
type CertCredentialExpiryScheduler struct {
	cron           *cron.Cron
	certRepo       *storage.CertificationRepository
	logger         *logrus.Logger
	CronExpression string
	Enabled        bool
}

// NewCertCredentialExpiryScheduler creates a new credential expiry scheduler
func NewCertCredentialExpiryScheduler(certRepo *storage.CertificationRepository) *CertCredentialExpiryScheduler {
	return &CertCredentialExpiryScheduler{
		cron:           cron.New(),
		certRepo:       certRepo,
		logger:         logrus.WithField("scheduler", "cert_credential_expiry").Logger,
		CronExpression: "0 2 * * *", // daily at 2 AM
		Enabled:        true,
	}
}

// Start begins the scheduled cleanup job
func (s *CertCredentialExpiryScheduler) Start(ctx context.Context) error {
	if !s.Enabled {
		s.logger.Info("Cert credential expiry scheduler is disabled")
		return nil
	}

	_, err := s.cron.AddFunc(s.CronExpression, func() {
		s.run(ctx)
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.WithField("cron", s.CronExpression).Info("Cert credential expiry scheduler started")
	return nil
}

// Stop halts the scheduler
func (s *CertCredentialExpiryScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Cert credential expiry scheduler stopped")
}

func (s *CertCredentialExpiryScheduler) run(ctx context.Context) {
	count, err := s.certRepo.ExpireCredentials(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to expire credentials")
		return
	}
	if count > 0 {
		s.logger.WithField("count", count).Info("Expired credentials")
	}
}
