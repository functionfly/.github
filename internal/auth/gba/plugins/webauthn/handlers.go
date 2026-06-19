// Package webauthn provides WebAuthn/Passkeys authentication support for GoBetterAuth
package webauthn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/auth/gba"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler provides HTTP handlers for WebAuthn endpoints
type Handler struct {
	plugin *WebAuthnPlugin
	logger *logrus.Logger
}

// NewHandler creates a new WebAuthn handler
func NewHandler(plugin *WebAuthnPlugin, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		plugin: plugin,
		logger: logger,
	}
}

// SetupRoutes registers all WebAuthn routes with the provided mux
// Base path should be /v1/auth/webauthn
func (h *Handler) SetupRoutes(mux *http.ServeMux, basePath string) {
	// Registration endpoints
	mux.HandleFunc("POST "+basePath+"/register/begin", h.HandleBeginRegistration)
	mux.HandleFunc("POST "+basePath+"/register/complete", h.HandleFinishRegistration)

	// Authentication endpoints
	mux.HandleFunc("POST "+basePath+"/authenticate/begin", h.HandleBeginAuthentication)
	mux.HandleFunc("POST "+basePath+"/authenticate/complete", h.HandleFinishAuthentication)

	// Discoverable authentication (for passkeys without user ID)
	mux.HandleFunc("POST "+basePath+"/authenticate/discoverable/begin", h.HandleBeginDiscoverableAuthentication)
	mux.HandleFunc("POST "+basePath+"/authenticate/discoverable/complete", h.HandleFinishDiscoverableAuthentication)

	// Credential management endpoints
	mux.HandleFunc("GET "+basePath+"/credentials", h.HandleListCredentials)
	mux.HandleFunc("DELETE "+basePath+"/credentials/{id}", h.HandleDeleteCredential)
	mux.HandleFunc("PATCH "+basePath+"/credentials/{id}", h.HandleUpdateCredential)

	// Status endpoint
	mux.HandleFunc("GET "+basePath+"/status", h.HandleStatus)

	h.logger.WithField("path", basePath).Info("WebAuthn routes registered")
}

// HandleBeginRegistration handles POST /v1/auth/webauthn/register/begin
// Starts the WebAuthn registration ceremony
func (h *Handler) HandleBeginRegistration(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req BeginRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Begin registration
	resp, err := h.plugin.BeginRegistration(r.Context(), userID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin registration")
		h.respondError(w, http.StatusInternalServerError, "Failed to begin registration")
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleFinishRegistration handles POST /v1/auth/webauthn/register/complete
// Completes the WebAuthn registration ceremony
func (h *Handler) HandleFinishRegistration(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request
	var req FinishRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Finish registration
	resp, err := h.plugin.FinishRegistration(r.Context(), userID, req)
	if err != nil {
		h.logger.WithError(err).Warn("Registration completion failed")
		h.respondErrorFromErr(r, w, http.StatusBadRequest, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleBeginAuthentication handles POST /v1/auth/webauthn/authenticate/begin
// Starts the WebAuthn authentication ceremony
func (h *Handler) HandleBeginAuthentication(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Begin authentication
	resp, err := h.plugin.BeginAuthentication(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to begin authentication")
		h.respondErrorFromErr(r, w, http.StatusBadRequest, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleFinishAuthentication handles POST /v1/auth/webauthn/authenticate/complete
// Completes the WebAuthn authentication ceremony
func (h *Handler) HandleFinishAuthentication(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Parse request
	var req FinishAuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Finish authentication
	resp, err := h.plugin.FinishAuthentication(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Warn("Authentication completion failed")
		h.respondErrorFromErr(r, w, http.StatusUnauthorized, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleBeginDiscoverableAuthentication handles POST /v1/auth/webauthn/authenticate/discoverable/begin
// Starts a discoverable credential authentication (for passkeys)
func (h *Handler) HandleBeginDiscoverableAuthentication(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Begin discoverable authentication (no user ID required)
	resp, err := h.plugin.BeginAuthenticationDiscoverable(r.Context())
	if err != nil {
		h.logger.WithError(err).Warn("Failed to begin discoverable authentication")
		h.respondErrorFromErr(r, w, http.StatusBadRequest, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleFinishDiscoverableAuthentication handles POST /v1/auth/webauthn/authenticate/discoverable/complete
// Completes a discoverable credential authentication
func (h *Handler) HandleFinishDiscoverableAuthentication(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "WebAuthn is not enabled")
		return
	}

	// Parse request
	var req FinishAuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Finish discoverable authentication
	resp, err := h.plugin.FinishAuthenticationDiscoverable(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Warn("Discoverable authentication completion failed")
		h.respondErrorFromErr(r, w, http.StatusUnauthorized, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleListCredentials handles GET /v1/auth/webauthn/credentials
// Returns a list of the user's WebAuthn credentials
func (h *Handler) HandleListCredentials(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get credentials
	credentials, err := h.plugin.GetUserCredentials(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list credentials")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve credentials")
		return
	}

	h.respondJSON(w, http.StatusOK, CredentialListResponse{
		Credentials: credentials,
	})
}

// HandleDeleteCredential handles DELETE /v1/auth/webauthn/credentials/{id}
// Deletes a WebAuthn credential
func (h *Handler) HandleDeleteCredential(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get credential ID from path
	credentialIDStr := r.PathValue("id")
	if credentialIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "Credential ID is required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid credential ID")
		return
	}

	// Delete credential
	if err := h.plugin.DeleteCredential(r.Context(), credentialID, userID); err != nil {
		h.logger.WithError(err).Warn("Failed to delete credential")
		h.respondErrorFromErr(r, w, http.StatusNotFound, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, DeleteCredentialResponse{
		Deleted: true,
		Message: "Credential deleted successfully",
	})
}

// HandleUpdateCredential handles PATCH /v1/auth/webauthn/credentials/{id}
// Updates a WebAuthn credential (e.g., renaming)
func (h *Handler) HandleUpdateCredential(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get credential ID from path
	credentialIDStr := r.PathValue("id")
	if credentialIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "Credential ID is required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid credential ID")
		return
	}

	// Parse request
	var req UpdateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Update credential
	if err := h.plugin.UpdateCredentialName(r.Context(), credentialID, userID, req.Name); err != nil {
		h.logger.WithError(err).Warn("Failed to update credential")
		h.respondErrorFromErr(r, w, http.StatusNotFound, "webauthn handler", err)
		return
	}

	h.respondJSON(w, http.StatusOK, UpdateCredentialResponse{
		ID:      credentialID.String(),
		Name:    req.Name,
		Message: "Credential updated successfully",
	})
}

// HandleStatus handles GET /v1/auth/webauthn/status
// Returns WebAuthn status for the authenticated user
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get status
	status, err := h.plugin.GetStatus(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get WebAuthn status")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve status")
		return
	}

	h.respondJSON(w, http.StatusOK, status)
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

// respondErrorFromErr logs err server-side with context and writes a generic
// client-visible message. Use in place of h.respondError(w, status, err.Error()).
func (h *Handler) respondErrorFromErr(r *http.Request, w http.ResponseWriter, status int, contextMsg string, err error) {
	if err != nil {
		fields := logrus.Fields{
			"status":  status,
			"context": contextMsg,
			"method":  "",
			"path":    "",
		}
		if r != nil {
			fields["method"] = r.Method
			if r.URL != nil {
				fields["path"] = r.URL.Path
			}
		}
		entry := logrus.WithError(err).WithFields(fields)
		if status >= 500 {
			entry.Error("webauthn handler error")
		} else {
			entry.Info("webauthn handler client error")
		}
	}
	h.respondError(w, status, sanitizedWebAuthnMessage(status))
}

func sanitizedWebAuthnMessage(status int) string {
	if status >= 500 {
		return "Internal server error"
	}
	switch status {
	case http.StatusBadRequest:
		return "Invalid request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusConflict:
		return "Conflict"
	}
	return http.StatusText(status)
}

// getUserIDFromContext extracts the user ID from the request context.
// GBA auth middleware (RequireAuth / RequirePermission / OptionalAuth) sets gba.ContextKeyUserID or gba.ContextKeySession.
func (h *Handler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	if userID, ok := r.Context().Value(gba.ContextKeyUserID).(uuid.UUID); ok && userID != uuid.Nil {
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
