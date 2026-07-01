package auth

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleMagicLinkRequest handles POST /auth/magic-link - sends a magic link to the user's email
func (h *Handler) HandleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-FunctionFly-Auth", "1")

	var req auth.MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Request magic link - returns same message whether email exists or not
	response, err := h.authSvc.RequestMagicLink(req, ipAddress, userAgent)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Magic link request failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to process magic link request")
		return
	}

	// Log the event (non-sensitive)
	logrus.WithField("email", req.Email).Info("Magic link requested")

	writeJSON(w, http.StatusOK, response)
}

// HandleMagicLinkVerify handles POST /auth/magic-link/verify - verifies a magic link token
func (h *Handler) HandleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-FunctionFly-Auth", "1")

	var req auth.MagicLinkVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" {
		writeJSONError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Debug logging (truncate token for safety)
	tokenPreview := req.Token
	if len(tokenPreview) > 8 {
		tokenPreview = tokenPreview[:8] + "..."
	}
	logrus.WithField("token_preview", tokenPreview).WithField("token_len", len(req.Token)).Debug("Magic link verification request")

	// Set IP and user agent from request
	req.IPAddress = getClientIP(r)
	req.UserAgent = r.Header.Get("User-Agent")

	// Verify magic link
	response, err := h.authSvc.VerifyMagicLink(req)
	if err != nil {
		errMsg := err.Error()
		logrus.WithError(err).Warn("Magic link verification failed")

		// Log the auth event
		if authEventErr := h.authSvc.Repo().LogAuthEvent(r.Context(), &storage.AuthEvent{
			EventType:     "magic_link_verify_failed",
			Success:       false,
			FailureReason: &errMsg,
			IPAddress:     req.IPAddress,
			UserAgent:     req.UserAgent,
		}); authEventErr != nil {
			logrus.WithError(authEventErr).Warn("Failed to log magic link verify failure")
		}

		status := http.StatusBadRequest
		if errMsg == "invalid magic link token" || errMsg == "magic link has already been used" || errMsg == "magic link has expired" {
			status = http.StatusGone // 410 - the resource is permanently gone
		}

		writeJSONError(w, status, errMsg)
		return
	}

	if response.User != nil {
		if userID, parseErr := uuid.Parse(response.User.ID); parseErr == nil {
			if _, historyErr := h.authSvc.Repo().CreateLoginHistory(r.Context(), userID, "login", req.IPAddress, req.UserAgent, "", "magic_link", false, nil); historyErr != nil {
				logrus.WithError(historyErr).WithField("userID", userID).Warn("Failed to record magic link login history")
			}
		}
	}

	writeJSON(w, http.StatusOK, response)
}
