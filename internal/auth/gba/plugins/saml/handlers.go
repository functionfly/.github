// Package saml provides SAML 2.0 SSO authentication support for GoBetterAuth
package saml

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"

	"github.com/functionfly/functionfly/internal/auth/gba"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler provides HTTP handlers for SAML endpoints
type Handler struct {
	plugin *SAMLPlugin
	logger *logrus.Logger
}

// NewHandler creates a new SAML handler
func NewHandler(plugin *SAMLPlugin, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		plugin: plugin,
		logger: logger,
	}
}

// SetupRoutes registers all SAML routes with the provided mux
// Base path should be /v1/auth/saml
func (h *Handler) SetupRoutes(mux *http.ServeMux, basePath string) {
	// SP metadata endpoint
	mux.HandleFunc("GET "+basePath+"/metadata/{tenant_id}", h.HandleMetadata)

	// SAML login/initiate endpoint
	mux.HandleFunc("GET "+basePath+"/login/{tenant_id}", h.HandleLogin)
	mux.HandleFunc("POST "+basePath+"/login/{tenant_id}", h.HandleLogin)

	// Assertion Consumer Service (ACS)
	mux.HandleFunc("POST "+basePath+"/acs/{tenant_id}", h.HandleACS)

	// Single Logout Service (SLO)
	mux.HandleFunc("POST "+basePath+"/slo/{tenant_id}", h.HandleSLO)
	mux.HandleFunc("GET "+basePath+"/slo/{tenant_id}", h.HandleSLO)

	// Admin endpoints for SAML configuration
	mux.HandleFunc("GET "+basePath+"/admin/config/{tenant_id}", h.HandleAdminGetConfig)
	mux.HandleFunc("PUT "+basePath+"/admin/config/{tenant_id}", h.HandleAdminUpdateConfig)
	mux.HandleFunc("DELETE "+basePath+"/admin/config/{tenant_id}", h.HandleAdminDeleteConfig)

	// Status endpoint
	mux.HandleFunc("GET "+basePath+"/status/{tenant_id}", h.HandleStatus)

	h.logger.WithField("path", basePath).Info("SAML routes registered")
}

// HandleMetadata handles GET /v1/auth/saml/metadata/{tenant_id}
// Returns the SP metadata XML for the tenant
func (h *Handler) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Warn("SAML config not found")
		h.respondError(w, http.StatusNotFound, "SAML not configured for this tenant")
		return
	}

	// Generate metadata
	metadata := h.plugin.service.GenerateMetadata(tenantID, config)

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metadata))
}

// HandleLogin handles GET/POST /v1/auth/saml/login/{tenant_id}
// Initiates SAML SSO by redirecting to the IdP
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Warn("SAML config not found")
		h.respondError(w, http.StatusNotFound, "SAML not configured for this tenant")
		return
	}

	// Build AuthnRequest
	authnRequest, requestID, err := h.plugin.service.BuildAuthnRequest(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build AuthnRequest")
		h.respondError(w, http.StatusInternalServerError, "Failed to initiate SAML login")
		return
	}

	// Store the request ID in session/cache for validation later
	relayState := r.URL.Query().Get("relay_state")
	if relayState == "" {
		relayState = generateRequestID()
	}

	// Build redirect URL
	redirectURL, err := h.plugin.service.BuildRedirectURL(authnRequest, config.IDPSSOURL)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build redirect URL")
		h.respondError(w, http.StatusInternalServerError, "Failed to initiate SAML login")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"request_id":  requestID,
		"relay_state": relayState,
	}).Info("SAML login initiated")

	// Return JSON response with redirect URL or perform redirect
	if r.Header.Get("Accept") == "application/json" {
		h.respondJSON(w, http.StatusOK, SAMLLoginResponse{
			AuthURL:    redirectURL,
			RequestID:  requestID,
			Binding:    "redirect",
			RelayState: relayState,
		})
		return
	}

	// Redirect to IdP
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleACS handles POST /v1/auth/saml/acs/{tenant_id}
// Assertion Consumer Service - processes SAML response from IdP
func (h *Handler) HandleACS(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	if samlResponse == "" {
		h.respondError(w, http.StatusBadRequest, "Missing SAMLResponse")
		return
	}

	relayState := r.FormValue("RelayState")

	h.logger.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"relay_state": relayState,
	}).Debug("Received SAML response")

	// Parse and validate the SAML response
	assertion, err := h.plugin.service.ParseSAMLResponse(r.Context(), tenantID, samlResponse)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse SAML response")
		h.respondACSResult(w, false, "", "", "", "", "/login?error=saml_failed", "Invalid SAML response", "")
		return
	}

	// Extract user information from assertion
	email := assertion.GetEmail(h.plugin.config.AttributeMapping)
	if email == "" {
		h.logger.Error("No email found in SAML assertion")
		h.respondACSResult(w, false, "", "", "", "", "/login?error=no_email", "No email found in SAML assertion", "")
		return
	}

	// Find or create user
	user, err := h.plugin.service.GetUserByEmail(r.Context(), tenantID, email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to lookup user")
		h.respondACSResult(w, false, "", "", "", "", "/login?error=user_lookup", "Failed to lookup user", "")
		return
	}

	if user == nil {
		// Auto-provision user if enabled
		if h.plugin.config.AutoProvision {
			user, err = h.plugin.service.CreateUser(r.Context(), tenantID, assertion)
			if err != nil {
				h.logger.WithError(err).Error("Failed to auto-provision user")
				h.respondACSResult(w, false, "", "", "", "", "/login?error=provision_failed", "Failed to create user account", "")
				return
			}
		} else {
			h.logger.WithField("email", email).Warn("User not found and auto-provisioning disabled")
			h.respondACSResult(w, false, "", "", "", "", "/login?error=user_not_found", "User account not found", "")
			return
		}
	} else {
		// Update user attributes from SAML
		if h.plugin.config.SyncAttributes {
			if err := h.plugin.service.UpdateUser(r.Context(), user, assertion); err != nil {
				h.logger.WithError(err).Warn("Failed to update user attributes")
			}
		}
	}

	// Create SAML session
	_, err = h.plugin.service.CreateSession(r.Context(), tenantID, user.ID, assertion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create SAML session")
	}

	// Generate JWT token
	token, err := h.plugin.GenerateToken(user.ID, tenantID, "saml")
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		h.respondACSResult(w, false, "", "", "", "", "/login?error=token_failed", "Failed to generate authentication token", "")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"tenant_id": tenantID,
		"email":     email,
	}).Info("SAML authentication successful")

	// Audit log
	h.plugin.AuditLog(r.Context(), user.ID, tenantID, "saml_login", "success", map[string]interface{}{
		"email":         email,
		"idp_entity_id": assertion.Issuer,
	})

	// Redirect or return JSON
	redirectURL := "/dashboard"
	if relayState != "" {
		redirectURL = relayState
	}

	h.respondACSResult(w, true, user.ID.String(), email, user.FirstName, user.LastName, redirectURL, "", token)
}

// HandleSLO handles GET/POST /v1/auth/saml/slo/{tenant_id}
// Single Logout Service - processes logout requests/responses
func (h *Handler) HandleSLO(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Handle logout request from IdP
	if r.Method == http.MethodPost {
		samlRequest := r.FormValue("SAMLRequest")
		if samlRequest != "" {
			h.logger.WithField("tenant_id", tenantID).Info("Received SAML logout request")
			nameID, sessionIndices, err := h.plugin.service.ParseLogoutRequest(samlRequest)
			if err != nil {
				h.logger.WithError(err).Warn("Failed to parse SAML logout request")
				h.respondError(w, http.StatusBadRequest, "Invalid SAML logout request")
				return
			}
			sessions, err := h.plugin.service.GetSessionsForLogout(r.Context(), tenantID, nameID, sessionIndices)
			if err != nil {
				h.logger.WithError(err).Warn("Failed to find sessions for logout")
				h.respondError(w, http.StatusInternalServerError, "Failed to process logout")
				return
			}
			userIDSet := make(map[uuid.UUID]struct{})
			for _, sess := range sessions {
				userIDSet[sess.UserID] = struct{}{}
				if err := h.plugin.service.DeleteSession(r.Context(), sess.SessionIndex); err != nil {
					h.logger.WithError(err).WithField("session_index", sess.SessionIndex).Warn("Failed to delete SAML session")
				}
			}
			userIDs := make([]uuid.UUID, 0, len(userIDSet))
			for id := range userIDSet {
				userIDs = append(userIDs, id)
			}
			if err := h.plugin.service.InvalidateGBASessionsForUsers(r.Context(), userIDs); err != nil {
				h.logger.WithError(err).Warn("Failed to invalidate GBA sessions on SLO")
			}
			h.respondJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
			return
		}
	}

	// Handle logout response from IdP
	samlResponse := r.FormValue("SAMLResponse")
	if samlResponse != "" {
		h.logger.WithField("tenant_id", tenantID).Info("Received SAML logout response")
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "logout_complete"})
		return
	}

	// Initiate logout - clear local session and redirect to IdP or login
	h.logger.WithField("tenant_id", tenantID).Info("Initiating SAML logout")

	sessionCookieName := os.Getenv("SESSION_COOKIE_NAME")
	if sessionCookieName == "" {
		sessionCookieName = "ff_session"
	}
	var sessionToken string
	if c, _ := r.Cookie(sessionCookieName); c != nil {
		sessionToken = c.Value
	}
	if sessionToken == "" && r.Header.Get("Authorization") != "" {
		const prefix = "Bearer "
		if len(r.Header.Get("Authorization")) > len(prefix) {
			sessionToken = r.Header.Get("Authorization")[len(prefix):]
		}
	}

	redirectURL := "/login?logged_out=1"
	if sessionToken != "" {
		userID, resolvedTenantID, err := h.plugin.service.GetUserFromSessionToken(r.Context(), sessionToken)
		if err == nil && resolvedTenantID == tenantID {
			if err := h.plugin.service.InvalidateSessionByToken(r.Context(), sessionToken); err != nil {
				h.logger.WithError(err).Warn("Failed to invalidate session on SLO initiate")
			}
			config, _ := h.plugin.service.GetConfig(r.Context(), tenantID)
			if config != nil && config.IDPSLOURL != "" {
				samlSess, err := h.plugin.service.GetSAMLSessionForUser(r.Context(), tenantID, userID)
				if err == nil && samlSess != nil {
					if url, err := h.plugin.service.BuildLogoutRequestRedirectURL(r.Context(), tenantID, samlSess.NameID, samlSess.SessionIndex); err == nil {
						redirectURL = url
					}
				}
			}
		} else if err == nil {
			_ = h.plugin.service.InvalidateSessionByToken(r.Context(), sessionToken)
		}
	}

	// Clear session cookie so the browser drops it
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleStatus handles GET /v1/auth/saml/status/{tenant_id}
// Returns SAML configuration status for a tenant
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondJSON(w, http.StatusOK, SAMLStatusResponse{
			Enabled:    false,
			Configured: false,
		})
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.respondJSON(w, http.StatusOK, SAMLStatusResponse{
			Enabled:    h.plugin.IsEnabled(),
			Configured: false,
		})
		return
	}

	h.respondJSON(w, http.StatusOK, SAMLStatusResponse{
		Enabled:     h.plugin.IsEnabled(),
		Configured:  true,
		IDPEntityID: config.IDPEntityID,
		SPEntityID:  config.SPEntityID,
	})
}

// requireAdminTenantScope ensures the request has an authenticated user and that the path tenant_id
// matches the context tenant (set by RequirePermission middleware). Call after parsing tenant_id from path.
// Returns false and sends 401/403 if not authorized.
func (h *Handler) requireAdminTenantScope(w http.ResponseWriter, r *http.Request, pathTenantID uuid.UUID) bool {
	userID, ok := r.Context().Value(gba.ContextKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		h.respondError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	ctxTenantID, ok := r.Context().Value(gba.ContextKeyTenantID).(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusForbidden, "Access denied")
		return false
	}
	if ctxTenantID != pathTenantID {
		h.logger.WithFields(logrus.Fields{
			"user_id":       userID,
			"path_tenant":   pathTenantID,
			"context_tenant": ctxTenantID,
		}).Warn("SAML admin access denied: tenant scope mismatch")
		h.respondError(w, http.StatusForbidden, "Access denied to this tenant")
		return false
	}
	return true
}

// HandleAdminGetConfig handles GET /v1/auth/saml/admin/config/{tenant_id}
// Returns SAML configuration (admin only)
func (h *Handler) HandleAdminGetConfig(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	if !h.requireAdminTenantScope(w, r, tenantID) {
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.respondError(w, http.StatusNotFound, "SAML configuration not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to get SAML configuration")
		return
	}

	h.respondJSON(w, http.StatusOK, SAMLConfigResponse{
		ID:           config.ID.String(),
		TenantID:     config.TenantID.String(),
		Enabled:      config.Enabled,
		IDPEntityID:  config.IDPEntityID,
		IDPSSOURL:    config.IDPSSOURL,
		IDPSLOURL:    config.IDPSLOURL,
		SPEntityID:   config.SPEntityID,
		ACSURL:       config.ACSURL,
		NameIDFormat: config.NameIDFormat,
		CreatedAt:    config.CreatedAt,
		UpdatedAt:    config.UpdatedAt,
	})
}

// HandleAdminUpdateConfig handles PUT /v1/auth/saml/admin/config/{tenant_id}
// Creates or updates SAML configuration (admin only)
func (h *Handler) HandleAdminUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	if !h.requireAdminTenantScope(w, r, tenantID) {
		return
	}

	var req SAMLConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.IDPEntityID == "" || req.IDPSSOURL == "" || req.IDPCertificate == "" {
		h.respondError(w, http.StatusBadRequest, "Missing required fields: idp_entity_id, idp_sso_url, idp_certificate")
		return
	}

	// Check if config already exists
	existingConfig, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		h.respondError(w, http.StatusInternalServerError, "Failed to check existing configuration")
		return
	}

	var config *SAMLConfig
	if existingConfig != nil {
		// Update existing config
		config, err = h.plugin.service.UpdateConfig(r.Context(), existingConfig.ID, &req)
		if err != nil {
			h.logger.WithError(err).Error("Failed to update SAML configuration")
			h.respondError(w, http.StatusInternalServerError, "Failed to update SAML configuration")
			return
		}
	} else {
		// Create new config
		config, err = h.plugin.service.CreateConfig(r.Context(), tenantID, &req)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create SAML configuration")
			h.respondError(w, http.StatusInternalServerError, "Failed to create SAML configuration")
			return
		}
	}

	h.respondJSON(w, http.StatusOK, SAMLConfigResponse{
		ID:           config.ID.String(),
		TenantID:     config.TenantID.String(),
		Enabled:      config.Enabled,
		IDPEntityID:  config.IDPEntityID,
		IDPSSOURL:    config.IDPSSOURL,
		IDPSLOURL:    config.IDPSLOURL,
		SPEntityID:   config.SPEntityID,
		ACSURL:       config.ACSURL,
		NameIDFormat: config.NameIDFormat,
		CreatedAt:    config.CreatedAt,
		UpdatedAt:    config.UpdatedAt,
	})
}

// HandleAdminDeleteConfig handles DELETE /v1/auth/saml/admin/config/{tenant_id}
// Deletes SAML configuration (admin only)
func (h *Handler) HandleAdminDeleteConfig(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SAML is not enabled")
		return
	}

	tenantIDStr := r.PathValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	if !h.requireAdminTenantScope(w, r, tenantID) {
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "SAML configuration not found")
		return
	}

	if err := h.plugin.service.DeleteConfig(r.Context(), config.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete SAML configuration")
		h.respondError(w, http.StatusInternalServerError, "Failed to delete SAML configuration")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// respondACSResult sends an ACS result (either JSON or HTML redirect)
func (h *Handler) respondACSResult(w http.ResponseWriter, success bool, userID, email, firstName, lastName, redirectURL, message, token string) {
	if success {
		h.respondJSON(w, http.StatusOK, SAMLACSResponse{
			Success:     true,
			UserID:      userID,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			RedirectURL: redirectURL,
			Token:       token,
		})
	} else {
		h.respondJSON(w, http.StatusUnauthorized, SAMLACSResponse{
			Success:     false,
			Message:     message,
			RedirectURL: redirectURL,
		})
	}
}

// XMLSchema represents the SAML XML schema for validation
type XMLSchema struct {
	XMLName xml.Name `xml:"schema"`
}
