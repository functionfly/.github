package storage

import (
	"time"

	"github.com/sirupsen/logrus"
)

// SessionCleanupService handles periodic cleanup of expired sessions
type SessionCleanupService struct {
	repo   Repository
	logger *logrus.Logger
}

// NewSessionCleanupService creates a new session cleanup service
func NewSessionCleanupService(repo Repository) *SessionCleanupService {
	return &SessionCleanupService{
		repo:   repo,
		logger: logrus.New(),
	}
}

// CleanupExpiredSessions removes all expired sessions from the database
func (s *SessionCleanupService) CleanupExpiredSessions() error {
	start := time.Now()
	deletedCount, err := s.repo.DeleteExpiredSessions()
	if err != nil {
		s.logger.WithError(err).Error("Failed to cleanup expired sessions")
		return err
	}

	duration := time.Since(start)
	if deletedCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"deleted_sessions": deletedCount,
			"duration_ms":      duration.Milliseconds(),
		}).Info("Cleaned up expired sessions")
	}

	return nil
}

// StartCleanupRoutine starts a background goroutine that periodically cleans up expired sessions
// The cleanup interval is configurable, defaulting to 1 hour
func (s *SessionCleanupService) StartCleanupRoutine(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour // Default to 1 hour
	}

	s.logger.WithField("interval", interval).Info("Starting session cleanup routine")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run cleanup immediately on startup
		if err := s.CleanupExpiredSessions(); err != nil {
			s.logger.WithError(err).Error("Initial session cleanup failed")
		}

		for range ticker.C {
			if err := s.CleanupExpiredSessions(); err != nil {
				s.logger.WithError(err).Error("Periodic session cleanup failed")
			}
		}
	}()
}

// CleanupUserSessions removes all sessions for a specific user (useful when user logs out or is deactivated)
func (s *SessionCleanupService) CleanupUserSessions(userID string) error {
	// Note: This would need to be added to the repository interface and implementation
	// For now, this is a placeholder for future implementation
	s.logger.WithField("user_id", userID).Info("User session cleanup requested (not yet implemented)")
	return nil
}