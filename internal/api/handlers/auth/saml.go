package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/functionfly/functionfly/internal/auth"
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
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Redirect to IdP
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleSSO processes the SAML Response from the IdP (ACS endpoint)
func (h *SAMLHandler) HandleSSO(w http.ResponseWriter, r *http.Request) {
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

	if samlResponse == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing SAMLResponse parameter")
		return
	}

	// Get tenant ID from query or use default
	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID uuid.UUID
	var err error
	if tenantIDStr != "" {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
			return
		}
	}

	// Process the SAML response
	resp, err := h.samlSvc.ProcessResponse(tenantID, samlResponse)
	if err != nil {
		logrus.WithError(err).Error("Failed to process SAML response")
		writeJSONError(w, http.StatusUnauthorized, "SAML authentication failed: "+err.Error())
		return
	}

	// Return JWT token to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": resp.Token,
		"user": map[string]interface{}{
			"id":    resp.User.ID,
			"email": resp.User.Email,
			"name":  resp.User.Name,
		},
		"name_id": resp.NameID,
	})
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
	if err := h.samlSvc.SaveConfig(config); err != nil {
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
