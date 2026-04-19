package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"fmt"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleLogin handles user authentication
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-FunctionFly-Auth", "1")
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).WithField("stack", string(debug.Stack())).Error("Login handler panic")
			writeJSONError(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.")
		}
	}()

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	identifier := req.Email
	logrus.WithField("identifier", identifier).Info("Login attempt")
	if identifier == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Username/email and password are required")
		return
	}

	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	user, _ := h.authSvc.Repo().GetUserByEmail(identifier)
	if user == nil {
		user, _ = h.authSvc.Repo().GetUserByUsername(identifier)
	}
	if user != nil {
		lockoutUntil, err := h.authSvc.Repo().GetUserLockoutStatus(user.ID)
		if err != nil {
			logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to check lockout status")
		} else if lockoutUntil != nil && time.Now().Before(*lockoutUntil) {
			remaining := time.Until(*lockoutUntil)
			minutes := int(remaining.Minutes()) + 1
			message := fmt.Sprintf("Account is temporarily locked due to too many failed login attempts. Try again in %d minutes.", minutes)

			failureReason := "account_locked"
			authEvent := &storage.AuthEvent{
				UserID:        &user.ID,
				TenantID:      &user.TenantID,
				EventType:     "login_failed",
				Success:       false,
				FailureReason: &failureReason,
				IPAddress:     ipAddress,
				UserAgent:     userAgent,
			}
			if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
				logrus.WithError(logErr).WithField("userID", user.ID).Warn("Failed to log lockout auth event")
			}

			writeJSONError(w, http.StatusTooManyRequests, message)
			return
		}
	}

	response, err := h.authSvc.Login(identifier, req.Password, ipAddress, userAgent)

	if user != nil {
		_, recordErr := h.authSvc.Repo().CreateLoginAttempt(user.ID, ipAddress, userAgent, err == nil, nil)
		if recordErr != nil {
			logrus.WithError(recordErr).WithField("userID", user.ID).Warn("Failed to record login attempt")
		}

		eventType := "login"
		failureReason := ""
		if err != nil {
			eventType = "login_failed"
			failureReason = "invalid_credentials"
		}

		authEvent := &storage.AuthEvent{
			UserID:        &user.ID,
			TenantID:      &user.TenantID,
			EventType:     eventType,
			Success:       err == nil,
			FailureReason: &failureReason,
			IPAddress:     ipAddress,
			UserAgent:     userAgent,
		}

		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", user.ID).Warn("Failed to log auth event")
		}

		if err != nil {
			h.handleFailedLoginAttempt(user.ID, ipAddress, userAgent)
		}
	}

	if err != nil {
		logrus.WithError(err).WithField("identifier", identifier).Warn("Login failed")
		if isLoginInternalError(err) {
			root := err
			for u := errors.Unwrap(root); u != nil; u = errors.Unwrap(root) {
				root = u
			}
			rootMsg := root.Error()
			logrus.WithError(err).WithField("identifier", identifier).WithField("root_cause", rootMsg).Error("Login internal error (500)")
			devDetail := ""
			if os.Getenv("DEVELOPMENT") == "true" {
				devDetail = rootMsg
			}
			if strings.Contains(rootMsg, "does not exist") || strings.Contains(rootMsg, "connection refused") ||
				strings.Contains(rootMsg, "no such table") || strings.Contains(rootMsg, "JWT secret not configured") {
				msg := "Service temporarily unavailable. Ensure the database is running and migrations have been applied."
				writeJSONErrorDetail(w, http.StatusServiceUnavailable, msg, devDetail)
				return
			}
			writeJSONErrorDetail(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.", devDetail)
			return
		}
		msg := err.Error()
		if msg == "invalid credentials" || strings.Contains(msg, "invalid credentials") ||
			strings.Contains(msg, "sql: no rows in result set") {
			msg = "Invalid credentials"
		}
		writeJSONError(w, http.StatusUnauthorized, msg)
		return
	}

	if user.Role == "super_admin" || user.Role == "admin" || user.Role == "support" || user.Role == "billing_admin" || user.Role == "developer_admin" {
		deviceFingerprint := r.Header.Get("X-Device-Fingerprint")
		expiresAt := time.Now().Add(24 * time.Hour)

		if postgresDB, ok := h.authSvc.Repo().(*storage.PostgresDB); ok {
			_, sessionErr := postgresDB.CreateAdminSession(user.ID, response.Token, ipAddress, userAgent, deviceFingerprint, expiresAt)
			if sessionErr != nil {
				logrus.WithError(sessionErr).WithField("user_id", user.ID).Warn("Failed to create admin session")
			} else {
				logrus.WithField("user_id", user.ID).Info("Admin session created")
			}
		} else {
			logrus.Warn("Repository is not PostgresDB, skipping admin session creation")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleFailedLoginAttempt checks if a user should be locked out after failed attempts
func (h *Handler) handleFailedLoginAttempt(userID uuid.UUID, ipAddress, userAgent string) {
	const maxFailedAttempts = 5
	const lockoutDuration = 15 * time.Minute
	const failureWindow = 15 * time.Minute

	failedCount, err := h.authSvc.Repo().GetRecentFailedLoginAttempts(userID, time.Now().Add(-failureWindow))
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("Failed to count recent failed login attempts")
		return
	}

	if failedCount >= maxFailedAttempts {
		lockoutUntil := time.Now().Add(lockoutDuration)

		_, err = h.authSvc.Repo().CreateLoginAttempt(userID, ipAddress, userAgent, false, &lockoutUntil)
		if err != nil {
			logrus.WithError(err).WithField("userID", userID).Warn("Failed to record lockout")
			return
		}

		eventType := "account_locked"
		authEvent := &storage.AuthEvent{
			UserID:    &userID,
			EventType: eventType,
			Success:   false,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata: map[string]interface{}{
				"failed_attempts":          failedCount,
				"lockout_duration_minutes": 15,
			},
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", userID).Warn("Failed to log account lockout event")
		}

		logrus.WithFields(logrus.Fields{
			"userID":         userID,
			"failedAttempts": failedCount,
			"lockoutUntil":   lockoutUntil,
		}).Warn("Account locked due to too many failed login attempts")
	}
}
