package mfa

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MFAHandler handles MFA-related API endpoints
type MFAHandler struct {
	authSvc *auth.AuthService
	logger  *logrus.Logger
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(authSvc *auth.AuthService) *MFAHandler {
	return &MFAHandler{
		authSvc: authSvc,
		logger:  logrus.New(),
	}
}

// SetupMFA handles MFA setup requests
func (h *MFAHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := auth.MFASetupRequest{
		UserID: claims.UserID,
	}

	response, err := h.authSvc.SetupMFA(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to setup MFA")
		http.Error(w, "Failed to setup MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyMFA handles MFA verification requests
func (h *MFAHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req auth.MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the request is for the authenticated user
	req.UserID = claims.UserID

	response, err := h.authSvc.VerifyMFA(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify MFA")
		http.Error(w, "Failed to verify MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EnableMFA handles MFA enable requests (after successful verification)
func (h *MFAHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.authSvc.EnableMFA(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to enable MFA")
		http.Error(w, "Failed to enable MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "MFA enabled successfully"})
}

// DisableMFA handles MFA disable requests
func (h *MFAHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req auth.MFADisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the request is for the authenticated user
	req.UserID = claims.UserID

	err := h.authSvc.DisableMFA(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to disable MFA")
		http.Error(w, "Failed to disable MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "MFA disabled successfully"})
}

// GetMFAStatus returns the current MFA status for the user
func (h *MFAHandler) GetMFAStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.authSvc.GetMFAStatus(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get MFA status")
		http.Error(w, "Failed to get MFA status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// AdminForceDisableMFA allows admins to force disable MFA for users (emergency access)
func (h *MFAHandler) AdminForceDisableMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is admin
	if claims.Role != "admin" && claims.Role != "super_admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Force disable MFA without verification
	err := h.authSvc.Repo().UpdateUserMFA(req.UserID, nil, false, nil, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to force disable MFA")
		http.Error(w, "Failed to force disable MFA", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"admin_user_id": claims.UserID,
		"target_user_id": req.UserID,
	}).Info("Admin force disabled MFA for user")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "MFA force disabled successfully"})
}