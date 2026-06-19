package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type SessionValidationMiddleware struct {
	sessionCache     *auth.SessionCache
	sessionPolicySvc *auth.SessionPolicyService
	logger           *logrus.Entry
}

func NewSessionValidationMiddlewareWithCache(sessionCache *auth.SessionCache, policySvc *auth.SessionPolicyService) *SessionValidationMiddleware {
	return &SessionValidationMiddleware{
		sessionCache:     sessionCache,
		sessionPolicySvc: policySvc,
		logger:           logrus.WithField("middleware", "session_validation"),
	}
}

func (m *SessionValidationMiddleware) validateSession(w http.ResponseWriter, r *http.Request) bool {
	if m.sessionCache == nil {
		return true
	}

	claims := GetUserFromContext(r)
	if claims == nil {
		return true
	}

	if claims.SessionID == uuid.Nil {
		return true
	}

	ctx := r.Context()
	status, err := m.sessionCache.GetSessionStatus(ctx, claims.SessionID)
	if err != nil {
		m.logger.WithError(err).WithField("session_id", claims.SessionID).Warn("Failed to get session status")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return false
	}

	if !status.Valid || status.Revoked {
		m.logger.WithFields(logrus.Fields{
			"session_id": claims.SessionID,
			"valid":      status.Valid,
			"revoked":    status.Revoked,
		}).Info("Session invalid or revoked")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return false
	}

	if status.ExpiresAt.Before(time.Now()) {
		m.logger.WithFields(logrus.Fields{
			"session_id": claims.SessionID,
			"expires_at": status.ExpiresAt,
		}).Info("Session expired")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return false
	}

	policy, err := m.sessionPolicySvc.GetSessionPolicy(ctx, claims.TenantID)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to get session policy, using defaults")
		policy = auth.DefaultSessionPolicy()
	}

	idleTimeout := policy.IdleTimeout
	if status.LastActivity.Add(idleTimeout).Before(time.Now()) {
		m.logger.WithFields(logrus.Fields{
			"session_id":    claims.SessionID,
			"last_activity": status.LastActivity,
			"idle_timeout":  idleTimeout,
		}).Info("Session idle timeout exceeded")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return false
	}

	go func() {
		if err := m.sessionCache.UpdateSessionActivity(context.Background(), claims.SessionID); err != nil {
			m.logger.WithError(err).WithField("session_id", claims.SessionID).Warn("Failed to update session activity")
		}
	}()

	return true
}

func (m *SessionValidationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.validateSession(w, r) {
			next.ServeHTTP(w, r)
		}
	})
}

func (m *SessionValidationMiddleware) RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.validateSession(w, r) {
			next.ServeHTTP(w, r)
		}
	}
}
