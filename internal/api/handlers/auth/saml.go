package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SAMLHandler handles SAML authentication endpoints
type SAMLHandler struct {
	samlSvc *auth.SAMLService
}

// NewSAMLHandler creates a new SAML handler
func NewSAMLHandler(samlSvc *auth.SAMLService) *SAMLHandler {
	return &SAMLHandler{
		samlSvc: samlSvc,
	}
}

// HandleGetMetadata returns the Service Provider metadata
func (h *SAMLHandler) HandleGetMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	metadata, err := h.samlSvc.GetSPMetadata(tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SAML metadata")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get SAML metadata")
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metadata))
}

// HandleLogin initiates SAML SSO login
func (h *SAMLHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get relay state from query params if present
	relayState := r.URL.Query().Get("relay_state")

	redirectURL, err := h.samlSvc.InitiateLogin(tenantID, relayState)
	if err != nil {
		logrus.WithError(err).Error("Failed to initiate SAML login")
		writeJSONErrorFromErr(r, w, http.StatusBadRequest, "saml handler", err)
		return
	}

	// Redirect to IdP
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// getSAMLFrontendURL returns the SAML callback URL for redirects
func getSAMLFrontendURL() string {
	return config.GetFrontendURL() + "/auth/saml/callback"
}

// buildSAMLRedirectURL builds the redirect URL for SAML callbacks with unified result structure
func buildSAMLRedirectURL(baseURL string, result *auth.AuthCallbackResult) string {
	u, _ := url.Parse(baseURL)
	q := u.Query()

	if result.Success {
		q.Set("token", result.Token)
		q.Set("refresh_token", result.RefreshToken)
		q.Set("new_user", fmt.Sprintf("%t", result.NewUser))
		if result.NameID != "" {
			q.Set("name_id", result.NameID)
		}
	} else {
		q.Set("error", result.Error)
		if result.ErrorDesc != "" {
			q.Set("error_description", result.ErrorDesc)
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// HandleSSO processes the SAML Response from the IdP (ACS endpoint)
// For browser flows: redirects to frontend with query params (?error=code&desc=message or ?token=...)
// For CLI flows: redirects to localhost with same params (consistent with OAuth behavior)
// If RelayState is provided, redirects to RelayState with token (consistent with OAuth)
// Otherwise returns HTML that posts back to parent window (for iframe flows)
func (h *SAMLHandler) HandleSSO(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get SAMLResponse from form
	samlResponse := r.FormValue("SAMLResponse")
	if samlResponse == "" {
		// Try reading from body as JSON
		body, err := io.ReadAll(r.Body)
		if err == nil {
			var req struct {
				SAMLResponse string `json:"saml_response"`
			}
			if json.Unmarshal(body, &req) == nil && req.SAMLResponse != "" {
				samlResponse = req.SAMLResponse
			}
		}
	}

	// Get relay state for redirect decisions
	relayState := r.FormValue("RelayState")
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	if samlResponse == "" {
		failureReason := "Missing SAMLResponse parameter"
		logrus.Warn("SAML SSO failed: missing SAMLResponse parameter")

		// Log authentication failure
		h.logSAMLAuthEvent(r.Context(), nil, nil, false, "saml_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"relay_state": relayState})

		result := &auth.AuthCallbackResult{
			Success:   false,
			Error:     string(auth.AuthErrMissingParameter),
			ErrorDesc: failureReason,
		}
		// If RelayState provided, redirect there with error
		if relayState != "" {
			u, _ := url.Parse(relayState)
			q := u.Query()
			q.Set("error", string(result.Error))
			q.Set("error_description", result.ErrorDesc)
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return
		}
		http.Redirect(w, r, buildSAMLRedirectURL(getSAMLFrontendURL(), result), http.StatusFound)
		return
	}

	// Get tenant ID from query or use default
	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID uuid.UUID
	var err error
	if tenantIDStr != "" {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			failureReason := "Invalid tenant ID format"
			logrus.WithError(err).Warn("SAML SSO failed: invalid tenant ID")

			// Log authentication failure
			h.logSAMLAuthEvent(r.Context(), nil, &tenantID, false, "saml_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"relay_state": relayState, "parse_error": err.Error()})

			result := &auth.AuthCallbackResult{
				Success:   false,
				Error:     string(auth.AuthErrSAMLInvalidTenant),
				ErrorDesc: failureReason,
			}
			// If RelayState provided, redirect there with error
			if relayState != "" {
				u, _ := url.Parse(relayState)
				q := u.Query()
				q.Set("error", string(result.Error))
				q.Set("error_description", result.ErrorDesc)
				u.RawQuery = q.Encode()
				http.Redirect(w, r, u.String(), http.StatusFound)
				return
			}
			http.Redirect(w, r, buildSAMLRedirectURL(getSAMLFrontendURL(), result), http.StatusFound)
			return
		}
	}

	// Process the SAML response
	resp, err := h.samlSvc.ProcessResponse(tenantID, samlResponse, relayState)
	if err != nil {
		logrus.WithError(err).Error("Failed to process SAML response")

		// Map error to standardized error code
		errorCode := auth.AuthErrSAMLInvalidResponse
		errorDesc := err.Error()

		// Classify specific SAML errors
		switch {
		case containsAny(err.Error(), []string{"not enabled", "SAML is not enabled"}):
			errorCode = auth.AuthErrSAMLEnabled
			errorDesc = "SAML authentication is not enabled for this tenant"
		case containsAny(err.Error(), []string{"signature verification failed", "could not find SignedInfo"}):
			errorCode = auth.AuthErrSAMLInvalidSignature
			errorDesc = "SAML response signature verification failed"
		case containsAny(err.Error(), []string{"failed to decode SAML response", "failed to parse"}):
			errorCode = auth.AuthErrSAMLInvalidResponse
			errorDesc = "Invalid SAML response format"
		case containsAny(err.Error(), []string{"no assertion", "no email found"}):
			errorCode = auth.AuthErrSAMLNoAssertion
			errorDesc = "Missing required information in SAML response"
		case containsAny(err.Error(), []string{"failed to get SAML config"}):
			errorCode = auth.AuthErrSAMLInvalidConfig
			errorDesc = "SAML configuration error"
		}

		// Log authentication failure
		h.logSAMLAuthEvent(r.Context(), nil, &tenantID, false, "saml_login", clientIP, userAgent, &errorDesc, time.Since(startTime), map[string]interface{}{
			"error_code":  string(errorCode),
			"relay_state": relayState,
			"raw_error":   err.Error(),
		})

		result := &auth.AuthCallbackResult{
			Success:   false,
			Error:     string(errorCode),
			ErrorDesc: errorDesc,
		}
		// If RelayState provided, redirect there with error
		if relayState != "" {
			u, _ := url.Parse(relayState)
			q := u.Query()
			q.Set("error", string(result.Error))
			q.Set("error_description", result.ErrorDesc)
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return
		}
		http.Redirect(w, r, buildSAMLRedirectURL(getSAMLFrontendURL(), result), http.StatusFound)
		return
	}

	// Generate refresh token for SAML login
	refreshToken, refreshTokenHash, err := h.samlSvc.GenerateRefreshToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate refresh token for SAML login")
		// Continue without refresh token - access token is still valid
		refreshToken = ""
	} else {
		// Store refresh token in database
		refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
		_, err = h.samlSvc.Repo().CreateRefreshToken(r.Context(), resp.User.ID, refreshTokenHash, "saml", "saml-callback", refreshExpiresAt)
		if err != nil {
			logrus.WithError(err).WithField("userID", resp.User.ID).Warn("Failed to store SAML refresh token")
			refreshToken = ""
		}
	}

	// Log successful SAML login with metadata including latency
	h.logSAMLAuthEvent(r.Context(), &resp.User.ID, &resp.User.TenantID, true, "saml_login", clientIP, userAgent, nil, time.Since(startTime), map[string]interface{}{
		"name_id":        resp.NameID,
		"relay_state":    relayState,
		"refresh_issued": refreshToken != "",
		"token_version":  "v1",
	})

	// Option A: If RelayState is provided and is an allowed redirect, redirect with token (consistent with OAuth)
	if relayState != "" && auth.IsAllowedRedirectURI(relayState) {
		u, _ := url.Parse(relayState)
		q := u.Query()
		q.Set("token", resp.Token)
		if refreshToken != "" {
			q.Set("refresh_token", refreshToken)
		}
		q.Set("new_user", "false")
		if resp.NameID != "" {
			q.Set("name_id", resp.NameID)
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	// Option B: If RelayState is a frontend URL, redirect there
	if relayState != "" {
		result := &auth.AuthCallbackResult{
			Success:      true,
			Token:        resp.Token,
			RefreshToken: refreshToken,
			NewUser:      false,
			NameID:       resp.NameID,
		}
		http.Redirect(w, r, buildSAMLRedirectURL(relayState, result), http.StatusFound)
		return
	}

	// Option C: Return HTML that posts back to parent window (for iframe flows)
	// This allows the parent window to receive the token via postMessage
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>SAML Authentication Complete</title>
</head>
<body>
  <script>
    (function() {
      var authResult = {
        success: true,
        token: %q,
        refresh_token: %q,
        new_user: false,
        name_id: %q,
        source: "saml"
      };
      if (window.parent && window.parent !== window) {
        window.parent.postMessage({
          type: "saml_auth_complete",
          payload: authResult
        }, "*");
      } else if (window.opener) {
        window.opener.postMessage({
          type: "saml_auth_complete",
          payload: authResult
        }, "*");
        window.close();
      } else {
        // No parent window, redirect to frontend
        var params = new URLSearchParams({
          token: authResult.token,
          refresh_token: authResult.refresh_token,
          new_user: String(authResult.new_user),
          name_id: authResult.name_id
        });
        window.location.href = %q + "?" + params.toString();
      }
    })();
  </script>
  <p>Authentication complete. You may close this window.</p>
</body>
</html>`, resp.Token, refreshToken, resp.NameID, getSAMLFrontendURL())
}

// HandleSLO processes Single Logout
func (h *SAMLHandler) HandleSLO(w http.ResponseWriter, r *http.Request) {
	// Get user from context (should be set by auth middleware)
	// For now, return a simple success response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// HandleGetConfig returns SAML configuration for a tenant
func (h *SAMLHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.samlSvc.GetConfig(tenantID)
	if err != nil {
		// Config doesn't exist - return empty config
		config = &storage.SAMLConfig{
			TenantID:     tenantID,
			Enabled:      false,
			SPEntityID:   "functionfly",
			NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateConfig updates SAML configuration for a tenant
func (h *SAMLHandler) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Parse request body
	var req struct {
		Enabled        bool     `json:"enabled"`
		IDPMetadata    string   `json:"idp_metadata"`
		IDPEntityID    string   `json:"idp_entity_id"`
		IDPSSOURL      string   `json:"idp_sso_url"`
		IDPCertificate string   `json:"idp_certificate"`
		SPEntityID     string   `json:"sp_entity_id"`
		NameIDFormat   string   `json:"name_id_format"`
		AuthnContexts  []string `json:"authn_contexts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build config
	config := &storage.SAMLConfig{
		TenantID:      tenantID,
		Enabled:       req.Enabled,
		SPEntityID:    req.SPEntityID,
		NameIDFormat:  req.NameIDFormat,
		AuthnContexts: req.AuthnContexts,
	}

	if req.IDPMetadata != "" {
		config.IDPMetadata = &req.IDPMetadata
	}
	if req.IDPEntityID != "" {
		config.IDPEntityID = &req.IDPEntityID
	}
	if req.IDPSSOURL != "" {
		config.IDPSSOURL = &req.IDPSSOURL
	}
	if req.IDPCertificate != "" {
		config.IDPCertificate = &req.IDPCertificate
	}

	// Save config
	if err := h.samlSvc.SaveConfig(r.Context(), config); err != nil {
		logrus.WithError(err).Error("Failed to save SAML config")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save SAML configuration")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "SAML configuration saved successfully",
	})
}

// containsAny checks if the error string contains any of the given substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// logSAMLAuthEvent logs a SAML authentication event for security auditing.
// Records success/failure, provider type, IP + user agent, user ID (on success),
// failure reason (on failure), and time taken (for latency monitoring).
func (h *SAMLHandler) logSAMLAuthEvent(ctx context.Context, userID, tenantID *uuid.UUID, success bool, eventType, clientIP, userAgent string, failureReason *string, duration time.Duration, metadata map[string]interface{}) {
	// Add latency information to metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["duration_ms"] = duration.Milliseconds()

	authEvent := &storage.AuthEvent{
		UserID:    userID,
		TenantID:  tenantID,
		EventType: eventType,
		Success:   success,
		IPAddress: clientIP,
		UserAgent: userAgent,
		Provider:  stringPtr("saml"),
		Metadata:  metadata,
	}

	if failureReason != nil {
		authEvent.FailureReason = failureReason
	}

	if logErr := h.samlSvc.Repo().LogAuthEvent(ctx, authEvent); logErr != nil {
		fields := logrus.Fields{
			"event_type":  eventType,
			"success":     success,
			"duration_ms": duration.Milliseconds(),
		}
		if userID != nil {
			fields["user_id"] = userID.String()
		}
		if failureReason != nil {
			fields["failure_reason"] = *failureReason
		}
		logrus.WithError(logErr).WithFields(fields).Warn("Failed to log SAML auth event")
	}
}
