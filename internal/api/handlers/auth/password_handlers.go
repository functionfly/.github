package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// VerifyPasswordRequest is the body for POST /auth/verify-password (re-auth for sensitive actions).
type VerifyPasswordRequest struct {
	Password string `json:"password"`
}

// HandlePasswordResetRequest initiates a password reset by sending an email with a reset token.
// Always returns 200 to avoid leaking whether the email exists.
func (h *Handler) HandlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	err := h.authSvc.RequestPasswordReset(req.Email)

	authEvent := &storage.AuthEvent{
		EventType: "password_reset_request",
		Success:   err == nil,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Metadata: map[string]interface{}{
			"email": req.Email,
		},
	}
	if err != nil {
		authEvent.FailureReason = stringPtr("reset_request_failed")
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log password reset request")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "If that email is registered, a password reset link has been sent.",
	})
}

// HandlePasswordResetConfirm completes a password reset using the token from the email.
func (h *Handler) HandlePasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Token == "" || req.NewPassword == "" {
		writeJSONError(w, http.StatusBadRequest, "token and newPassword are required")
		return
	}

	err := h.authSvc.ConfirmPasswordReset(req.Token, req.NewPassword)
	if err != nil {
		logrus.WithError(err).Debug("Password reset confirm failed")

		authEvent := &storage.AuthEvent{
			EventType:     "password_reset_confirm",
			Success:       false,
			FailureReason: stringPtr("invalid_token"),
			IPAddress:     getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log password reset confirmation failure")
		}

		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	authEvent := &storage.AuthEvent{
		EventType: "password_reset_confirm",
		Success:   true,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log password reset confirmation success")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

// HandleVerifyPassword verifies the current user's password. Used for re-authentication (e.g. reveal secret).
func (h *Handler) HandleVerifyPassword(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header")
		return
	}
	claims, err := h.authSvc.ValidateToken(parts[1])
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	var req VerifyPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Password is required")
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}
	if user.PasswordHash == "" {
		writeJSONError(w, http.StatusBadRequest, "This account uses sign-in with a provider; password verification is not available")
		return
	}

	valid, err := h.authSvc.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Debug("Verify password error")
		writeJSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}
	if !valid {
		writeJSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Password verified"})
}
