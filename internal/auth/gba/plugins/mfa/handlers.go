package mfa

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	gba "github.com/functionfly/functionfly/internal/auth/gba"
)

// Handler provides HTTP handlers for MFA endpoints
type Handler struct {
	plugin *MFAPlugin
	logger *logrus.Logger
}

// NewHandler creates a new MFA handler
func NewHandler(plugin *MFAPlugin, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		plugin: plugin,
		logger: logger,
	}
}

// SetupRoutes registers all MFA routes with the provided mux
// Base path should be /v1/auth/mfa
func (h *Handler) SetupRoutes(mux *http.ServeMux, basePath string) {
	// MFA Setup
	mux.HandleFunc("POST "+basePath+"/setup", h.HandleSetup)

	// Verify and enable MFA
	mux.HandleFunc("POST "+basePath+"/verify", h.HandleVerify)

	// MFA Challenge (during login)
	mux.HandleFunc("POST "+basePath+"/challenge", h.HandleChallenge)

	// Disable MFA
	mux.HandleFunc("POST "+basePath+"/disable", h.HandleDisable)

	// Regenerate backup codes
	mux.HandleFunc("POST "+basePath+"/backup-codes/regenerate", h.HandleRegenerateBackupCodes)

	// Get MFA status
	mux.HandleFunc("GET "+basePath+"/status", h.HandleStatus)

	h.logger.WithField("path", basePath).Info("MFA routes registered")
}

// HandleSetup handles POST /v1/auth/mfa/setup
// Generates a new TOTP secret and returns QR code URL
func (h *Handler) HandleSetup(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "MFA is not enabled")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user's email for account name in authenticator app
	accountName := h.getUserEmailFromContext(r)
	if accountName == "" {
		accountName = userID.String()
	}

	// Generate TOTP setup
	setup, err := h.plugin.GenerateTOTP(r.Context(), userID, accountName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate TOTP")
		h.respondError(w, http.StatusInternalServerError, "Failed to setup MFA")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"secret":       setup.Secret,
		"qr_code_url":  setup.QRCodeURL,
		"backup_codes": setup.BackupCodes,
		"message":      "Scan the QR code with your authenticator app and verify a code to enable MFA",
	})
}

// HandleVerify handles POST /v1/auth/mfa/verify
// Verifies a TOTP code and enables MFA
func (h *Handler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "MFA is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		h.respondError(w, http.StatusBadRequest, "Code is required")
		return
	}

	// Verify and enable TOTP
	if err := h.plugin.VerifyAndEnableTOTP(r.Context(), userID, req.Code); err != nil {
		h.logger.WithError(err).Warn("TOTP verification failed")
		h.respondError(w, http.StatusBadRequest, "Invalid TOTP code")
		return
	}

	h.respondJSON(w, http.StatusOK, TOTPVerifyResponse{
		Enabled: true,
		Message: "MFA enabled successfully",
	})
}

// HandleChallenge handles POST /v1/auth/mfa/challenge
// Verifies MFA code during login flow
func (h *Handler) HandleChallenge(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "MFA is not enabled")
		return
	}

	// Get user ID from context (this might be set during partial login)
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req MFAChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get code from either field
	code := req.Code
	if code == "" {
		code = req.BackupCode
	}
	if code == "" {
		h.respondError(w, http.StatusBadRequest, "Code or backup_code is required")
		return
	}

	// Verify code
	if err := h.plugin.VerifyCode(r.Context(), userID, code); err != nil {
		h.logger.WithError(err).Warn("MFA challenge failed")
		h.respondJSON(w, http.StatusOK, MFAChallengeResponse{
			Valid:   false,
			Message: "Invalid MFA code",
		})
		return
	}

	h.respondJSON(w, http.StatusOK, MFAChallengeResponse{
		Valid:   true,
		Message: "MFA verified successfully",
	})
}

// HandleDisable handles POST /v1/auth/mfa/disable
// Disables MFA after verification
func (h *Handler) HandleDisable(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "MFA is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req MFADisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get code from either field
	code := req.Code
	if code == "" {
		code = req.BackupCode
	}
	if code == "" {
		h.respondError(w, http.StatusBadRequest, "Code or backup_code is required")
		return
	}

	// Disable MFA
	if err := h.plugin.Disable(r.Context(), userID, code); err != nil {
		h.logger.WithError(err).Warn("MFA disable failed")
		h.respondError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	h.respondJSON(w, http.StatusOK, MFADisableResponse{
		Disabled: true,
		Message:  "MFA disabled successfully",
	})
}

// HandleRegenerateBackupCodes handles POST /v1/auth/mfa/backup-codes/regenerate
// Regenerates backup codes after TOTP verification
func (h *Handler) HandleRegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "MFA is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req BackupCodesRegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		h.respondError(w, http.StatusBadRequest, "Code is required")
		return
	}

	// Regenerate backup codes
	codes, err := h.plugin.RegenerateBackupCodes(r.Context(), userID, req.Code)
	if err != nil {
		h.logger.WithError(err).Warn("Backup code regeneration failed")
		h.respondError(w, http.StatusBadRequest, "Invalid TOTP code")
		return
	}

	h.respondJSON(w, http.StatusOK, BackupCodesRegenerateResponse{
		BackupCodes: codes,
	})
}

// HandleStatus handles GET /v1/auth/mfa/status
// Returns MFA status for the authenticated user
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get MFA status
	status, err := h.plugin.GetStatus(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get MFA status")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve MFA status")
		return
	}

	h.respondJSON(w, http.StatusOK, MFAStatusResponse{
		Enabled:         status.Enabled,
		Verified:        status.Verified,
		HasBackupCodes:  status.HasBackupCodes,
		BackupCodeCount: status.BackupCodeCount,
	})
}

// Helper methods

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// getUserIDFromContext extracts the user ID from the request context.
// GBA auth middleware (OptionalAuth / RequirePermission) sets gba.ContextKeyUserID.
func (h *Handler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	if userID, ok := r.Context().Value(gba.ContextKeyUserID).(uuid.UUID); ok {
		return userID, nil
	}
	if session, ok := r.Context().Value(gba.ContextKeySession).(*gba.Session); ok && session != nil {
		return session.UserID, nil
	}
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		return uuid.Parse(userIDStr)
	}
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		return userID, nil
	}
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return uuid.Parse(userID)
	}
	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

// getUserEmailFromContext extracts the user email from the request context
func (h *Handler) getUserEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}
