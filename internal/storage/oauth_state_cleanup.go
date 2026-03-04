package storage

import (
	"time"

	"github.com/sirupsen/logrus"
)

// OAuthStateCleanupService handles periodic cleanup of expired OAuth states
type OAuthStateCleanupService struct {
	repo   Repository
	logger *logrus.Logger
}

// NewOAuthStateCleanupService creates a new OAuth state cleanup service
func NewOAuthStateCleanupService(repo Repository) *OAuthStateCleanupService {
	return &OAuthStateCleanupService{
		repo:   repo,
		logger: logrus.New(),
	}
}

// CleanupExpiredOAuthStates removes all expired OAuth states from the database
func (s *OAuthStateCleanupService) CleanupExpiredOAuthStates() error {
	start := time.Now()
	deletedCount, err := s.repo.DeleteExpiredOAuthStates()
	if err != nil {
		s.logger.WithError(err).Error("Failed to cleanup expired OAuth states")
		return err
	}

	duration := time.Since(start)
	if deletedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_count": deletedCount,
			"duration_ms":   duration.Milliseconds(),
		}).Info("Cleaned up expired OAuth states")
	} else {
		s.logger.WithField("duration_ms", duration.Milliseconds()).Debug("OAuth state cleanup completed (no states to clean)")
	}

	return nil
}

// StartCleanupRoutine starts a periodic cleanup routine that runs at the specified interval
func (s *OAuthStateCleanupService) StartCleanupRoutine(interval time.Duration) {
	if interval <= 0 {
		interval = 6 * time.Hour // Default to 6 hours for OAuth states
	}

	s.logger.WithField("interval", interval).Info("Starting OAuth state cleanup routine")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run cleanup immediately on startup
		if err := s.CleanupExpiredOAuthStates(); err != nil {
			s.logger.WithError(err).Error("Initial OAuth state cleanup failed")
		}

		for range ticker.C {
			if err := s.CleanupExpiredOAuthStates(); err != nil {
				s.logger.WithError(err).Error("Periodic OAuth state cleanup failed")
			}
		}
	}()
}

// StopCleanupRoutine stops the cleanup routine (for testing or graceful shutdown)
func (s *OAuthStateCleanupService) StopCleanupRoutine() {
	s.logger.Info("OAuth state cleanup routine stopped")
}