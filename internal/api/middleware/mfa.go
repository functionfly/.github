package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// MFARequiredMiddleware enforces MFA for protected routes
type MFARequiredMiddleware struct {
	authSvc *auth.AuthService
	repo    storage.Repository
	logger  *logrus.Logger
}

// NewMFARequiredMiddleware creates a new MFA required middleware
func NewMFARequiredMiddleware(authSvc *auth.AuthService, repo storage.Repository) *MFARequiredMiddleware {
	return &MFARequiredMiddleware{
		authSvc: authSvc,
		repo:    repo,
		logger:  logrus.New(),
	}
}

// extractSessionToken extracts the JWT token from the Authorization header
func (m *MFARequiredMiddleware) extractSessionToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// ensureSessionExists creates or updates a session for the current request
func (m *MFARequiredMiddleware) ensureSessionExists(ctx context.Context, r *http.Request, claims *auth.Claims) error {
	sessionToken, err := m.extractSessionToken(r)
	if err != nil {
		return fmt.Errorf("failed to extract session token: %w", err)
	}

	// Parse IP address from RemoteAddr (which may include port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If parsing fails, use RemoteAddr as-is (fallback)
		host = r.RemoteAddr
	}
	ipAddress := host
	userAgent := r.Header.Get("User-Agent")

	// Try to get existing session
	_, err = m.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		// Session doesn't exist or is expired, create a new one
		expiresAt := time.Now().Add(24 * time.Hour) // 24 hours from now
		_, err = m.repo.CreateSession(ctx, claims.UserID, sessionToken, ipAddress, userAgent, expiresAt)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	} else {
		// Update session activity
		err = m.repo.UpdateSessionActivity(ctx, sessionToken)
		if err != nil {
			m.logger.WithError(err).Warn("Failed to update session activity")
		}
	}

	return nil
}

// RequireMFA enforces MFA verification for admin users
func (m *MFARequiredMiddleware) RequireMFA(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			m.logger.Warn("RequireMFA: No claims in context - authentication failed before MFA check")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// MFA_BYPASS_ENABLED is only for emergency access when MFA system is down
		// SECURITY: Only allowed in non-production environments to prevent accidental bypass in production
		if os.Getenv("MFA_BYPASS_ENABLED") == "true" && os.Getenv("MFA_BYPASS_SECRET") != "" {
			isProduction := os.Getenv("PRODUCTION_ENV") == "true" || os.Getenv("NODE_ENV") == "production"
			if isProduction {
				m.logger.Error("MFA_BYPASS_ENABLED is set in production - rejecting request for security")
				http.Error(w, "MFA bypass not allowed in production", http.StatusForbidden)
				return
			}
			m.logger.WithFields(logrus.Fields{"user_id": claims.UserID, "email": claims.Email}).Warn("MFA bypassed via MFA_BYPASS_ENABLED - emergency access only (development mode)")
			next(w, r)
			return
		}

		m.logger.WithFields(logrus.Fields{
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
		}).Info("RequireMFA: Checking MFA for user")

		// Ensure session exists and is up to date
		if err := m.ensureSessionExists(r.Context(), r, claims); err != nil {
			m.logger.WithError(err).Error("Failed to ensure session exists")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Check if MFA is required for this user
		required, err := m.authSvc.IsMFARequired(claims.UserID)
		if err != nil {
			m.logger.WithError(err).Error("Failed to check MFA requirement")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !required {
			// MFA not required for this user
			next.ServeHTTP(w, r)
			return
		}

		// Check if user has MFA enabled
		status, err := m.authSvc.GetMFAStatus(claims.UserID)
		if err != nil {
			m.logger.WithError(err).Error("Failed to get MFA status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !status.Enabled {
			// MFA is required but not yet enabled - allow access so admin can reach the UI and enable MFA
			m.logger.WithFields(logrus.Fields{
				"user_id": claims.UserID,
				"email":   claims.Email,
			}).Info("MFA required but not enabled - allowing access so user can set up MFA")
			next.ServeHTTP(w, r)
			return
		}

		// Check if MFA verification has been completed for this session
		if !m.isMFASessionVerified(r.Context(), r) {
			m.logger.WithFields(logrus.Fields{
				"user_id": claims.UserID,
				"email":   claims.Email,
			}).Warn("MFA is enabled but session not verified - blocking access")
			// MFA enabled but not verified for this session - require verification
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":           "MFA verification required",
				"message":         "Please verify your identity with MFA to access this resource.",
				"action_required": "verify_mfa",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequireMFAAndVerify combines authentication and MFA verification in one step
func (m *MFARequiredMiddleware) RequireMFAAndVerify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if MFA is required
		required, err := m.authSvc.IsMFARequired(claims.UserID)
		if err != nil {
			m.logger.WithError(err).Error("Failed to check MFA requirement")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !required {
			next.ServeHTTP(w, r)
			return
		}

		// Check MFA status
		status, err := m.authSvc.GetMFAStatus(claims.UserID)
		if err != nil {
			m.logger.WithError(err).Error("Failed to get MFA status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !status.Enabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":           "MFA required",
				"message":         "Multi-factor authentication is required for administrative access.",
				"action_required": "setup_mfa",
			})
			return
		}

		// Extract MFA code from request
		mfaCode := m.extractMFACode(r)
		if mfaCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "MFA code required",
				"message": "Please provide an MFA verification code.",
			})
			return
		}

		// Verify MFA code
		verifyReq := auth.MFAVerifyRequest{
			UserID: claims.UserID,
			Code:   mfaCode,
		}

		verifyResp, err := m.authSvc.VerifyMFA(verifyReq)
		if err != nil {
			m.logger.WithError(err).Error("Failed to verify MFA")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !verifyResp.Verified {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Invalid MFA code",
				"message": "The provided MFA code is invalid or expired.",
			})
			return
		}

		// Mark session as MFA verified
		m.markMFASessionVerified(r.Context(), r)

		next.ServeHTTP(w, r)
	}
}

// extractMFACode extracts MFA code from various sources
// SECURITY: Query parameters are NOT accepted because they are logged in proxies and browser history
func (m *MFARequiredMiddleware) extractMFACode(r *http.Request) string {
	// Try header first (X-MFA-Code)
	if code := r.Header.Get("X-MFA-Code"); code != "" {
		return code
	}

	// Try form data
	if err := r.ParseForm(); err == nil {
		if code := r.Form.Get("mfa_code"); code != "" {
			return code
		}
	}

	// Try JSON body for POST requests (handler will parse)
	// Note: Query parameters are explicitly NOT accepted for security reasons

	return ""
}

// isMFASessionVerified checks if the current session has been MFA verified
func (m *MFARequiredMiddleware) isMFASessionVerified(ctx context.Context, r *http.Request) bool {
	sessionToken, err := m.extractSessionToken(r)
	if err != nil {
		m.logger.WithError(err).Debug("Failed to extract session token for MFA verification")
		return false
	}

	session, err := m.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		m.logger.WithError(err).Debug("Failed to get session for MFA verification")
		return false
	}

	return session.MFAVerified
}

// markMFASessionVerified marks the current session as MFA verified
func (m *MFARequiredMiddleware) markMFASessionVerified(ctx context.Context, r *http.Request) {
	sessionToken, err := m.extractSessionToken(r)
	if err != nil {
		m.logger.WithError(err).Error("Failed to extract session token for MFA verification")
		return
	}

	err = m.repo.UpdateSessionMFAStatus(ctx, sessionToken, true)
	if err != nil {
		m.logger.WithError(err).Error("Failed to update session MFA status")
		return
	}
}
