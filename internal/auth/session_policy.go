package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Custom error types for session policy violations
var (
	ErrTooManySessions = errors.New("user has reached maximum concurrent sessions limit")
	ErrSessionExpired  = errors.New("session has expired")
	ErrSessionIdle     = errors.New("session has been idle too long")
)

// SessionPolicy represents the session policy configuration for a tenant
type SessionPolicy struct {
	MaxDuration       time.Duration `json:"max_duration"`       // maximum session duration
	IdleTimeout       time.Duration `json:"idle_timeout"`       // idle timeout duration
	MaxConcurrent     int           `json:"max_concurrent"`     // maximum concurrent sessions
	RequireMFA        bool          `json:"require_mfa"`        // whether MFA is required
	DevicePersistence bool          `json:"device_persistence"` // whether device persistence is enabled
}

// SessionPolicyService handles session policy operations
type SessionPolicyService struct {
	repo storage.Repository
}

// NewSessionPolicyService creates a new session policy service
func NewSessionPolicyService(repo storage.Repository) *SessionPolicyService {
	return &SessionPolicyService{
		repo: repo,
	}
}

// DefaultSessionPolicy returns the default session policy
func DefaultSessionPolicy() *SessionPolicy {
	return &SessionPolicy{
		MaxDuration:       24 * time.Hour, // 24 hours
		IdleTimeout:       8 * time.Hour,  // 8 hours
		MaxConcurrent:     5,              // 5 concurrent sessions
		RequireMFA:        false,          // MFA not required by default
		DevicePersistence: true,           // device persistence enabled by default
	}
}

// GetSessionPolicy retrieves the session policy for a tenant
func (s *SessionPolicyService) GetSessionPolicy(ctx context.Context, tenantID uuid.UUID) (*SessionPolicy, error) {
	tenant, err := s.repo.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, errors.New("tenant not found")
	}

	policy := DefaultSessionPolicy()
	if tenant.SessionMaxDuration != nil {
		policy.MaxDuration = *tenant.SessionMaxDuration
	}
	if tenant.SessionIdleTimeout != nil {
		policy.IdleTimeout = *tenant.SessionIdleTimeout
	}
	if tenant.ConcurrentSessions != nil {
		policy.MaxConcurrent = *tenant.ConcurrentSessions
	}
	policy.RequireMFA = tenant.MFAPolicy == "required"
	policy.DevicePersistence = tenant.SessionPersistence

	return policy, nil
}

// UpdateSessionPolicy updates the session policy for a tenant
func (s *SessionPolicyService) UpdateSessionPolicy(ctx context.Context, tenantID uuid.UUID, policy *SessionPolicy) error {
	// Validate the policy
	if policy.MaxDuration < 1*time.Minute || policy.MaxDuration > 30*24*time.Hour {
		return errors.New("max duration must be between 1 minute and 30 days")
	}
	if policy.IdleTimeout < 1*time.Minute || policy.IdleTimeout > 30*24*time.Hour {
		return errors.New("idle timeout must be between 1 minute and 30 days")
	}
	if policy.MaxConcurrent < 1 || policy.MaxConcurrent > 100 {
		return errors.New("max concurrent sessions must be between 1 and 100")
	}

	// Convert from time.Duration to minutes
	updates := map[string]interface{}{
		"session_max_duration": int(policy.MaxDuration.Minutes()),
		"session_idle_timeout": int(policy.IdleTimeout.Minutes()),
		"concurrent_sessions":  policy.MaxConcurrent,
		"session_persistence":  "device",
	}

	if policy.DevicePersistence {
		updates["session_persistence"] = "device"
	} else {
		updates["session_persistence"] = "browser"
	}

	// Update MFA policy based on require_mfa
	if policy.RequireMFA {
		updates["mfa_policy"] = "required"
	} else {
		updates["mfa_policy"] = "optional"
	}

	_, err := s.repo.UpdateTenant(ctx, tenantID, updates)
	if err != nil {
		return fmt.Errorf("failed to update tenant session policy: %w", err)
	}

	logrus.WithField("tenant_id", tenantID).Info("Session policy updated")

	return nil
}

// EnforcePolicy checks if a new session can be created based on the tenant's concurrent session limit
func (s *SessionPolicyService) EnforcePolicy(ctx context.Context, tenantID, userID uuid.UUID) error {
	policy, err := s.GetSessionPolicy(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get session policy: %w", err)
	}

	// Check concurrent sessions limit
	activeSessions, err := s.repo.CountActiveUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to count active sessions: %w", err)
	}

	if activeSessions >= policy.MaxConcurrent {
		logrus.WithFields(logrus.Fields{
			"tenant_id":       tenantID,
			"user_id":         userID,
			"active_sessions": activeSessions,
			"max_concurrent":  policy.MaxConcurrent,
		}).Warn("User has reached maximum concurrent sessions limit")
		return ErrTooManySessions
	}

	return nil
}

// IsSessionExpired checks if a session has exceeded its maximum duration
func (s *SessionPolicyService) IsSessionExpired(session *storage.Session, tenantID uuid.UUID) bool {
	policy, err := s.GetSessionPolicy(context.Background(), tenantID)
	if err != nil {
		// If we can't get the policy, use default
		policy = DefaultSessionPolicy()
	}

	maxDuration := policy.MaxDuration
	if session.CreatedAt.Add(maxDuration).Before(time.Now()) {
		return true
	}

	return false
}

// IsSessionIdle checks if a session has been idle beyond the idle timeout
func (s *SessionPolicyService) IsSessionIdle(session *storage.Session, tenantID uuid.UUID) bool {
	policy, err := s.GetSessionPolicy(context.Background(), tenantID)
	if err != nil {
		// If we can't get the policy, use default
		policy = DefaultSessionPolicy()
	}

	idleTimeout := policy.IdleTimeout
	if session.LastActivity.Add(idleTimeout).Before(time.Now()) {
		return true
	}

	return false
}

// GetActiveSessions returns all active sessions for a tenant
func (s *SessionPolicyService) GetActiveSessions(ctx context.Context, tenantID uuid.UUID) ([]*storage.Session, error) {
	sessions, err := s.repo.ListTenantSessions(ctx, tenantID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant sessions: %w", err)
	}

	return sessions, nil
}

// RevokeSession revokes a specific session
func (s *SessionPolicyService) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	err := s.repo.DeleteSessionByID(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	}).Info("Session revoked")

	return nil
}

// RevokeSessionByID revokes a specific session by ID only (for admin operations)
func (s *SessionPolicyService) RevokeSessionByID(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	if session == nil {
		return errors.New("session not found")
	}
	if err := s.repo.DeleteSessionByIDOnly(ctx, sessionID, session.UserID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	logrus.WithField("session_id", sessionID).Info("Session revoked")

	return nil
}

// RevokeAllSessions revokes all sessions for a tenant
func (s *SessionPolicyService) RevokeAllSessions(ctx context.Context, tenantID uuid.UUID) error {
	sessions, err := s.repo.ListTenantSessions(ctx, tenantID, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to list tenant sessions: %w", err)
	}

	for _, session := range sessions {
		err := s.repo.DeleteSessionByID(ctx, session.ID, session.UserID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"session_id": session.ID,
				"user_id":    session.UserID,
				"error":      err,
			}).Warn("Failed to revoke session")
		}
	}

	logrus.WithField("tenant_id", tenantID).Info("All tenant sessions revoked")

	return nil
}

// ValidateSession validates a session against the tenant's session policy
// Returns nil if valid, error if invalid or expired
func (s *SessionPolicyService) ValidateSession(ctx context.Context, session *storage.Session, tenantID uuid.UUID) error {
	// Check if session is expired based on max duration
	if s.IsSessionExpired(session, tenantID) {
		return ErrSessionExpired
	}

	// Check if session is idle
	if s.IsSessionIdle(session, tenantID) {
		return ErrSessionIdle
	}

	// Check if session has expired (based on expires_at field)
	if session.ExpiresAt.Before(time.Now()) {
		return ErrSessionExpired
	}

	return nil
}
