package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

type SAMlHandler struct {
	db *storage.PostgresDB
}

func NewSAMlHandler(db *storage.PostgresDB) *SAMlHandler {
	return &SAMlHandler{
		db: db,
	}
}

type AdminSAMLConfig struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id,omitempty"`
	EntityID    string `json:"entity_id"`
	SSOURL      string `json:"sso_url"`
	SLOURL      string `json:"slo_url,omitempty"`
	Certificate string `json:"certificate,omitempty"`
	Enabled     bool   `json:"enabled"`
}

type AdminSAMLMetadata struct {
	EntityID    string `json:"entity_id"`
	ACSURL      string `json:"acs_url"`
	SLOURL      string `json:"slo_url,omitempty"`
	MetadataXML string `json:"metadata_xml,omitempty"`
}

type UpdateSAMLConfigRequest struct {
	Enabled     bool   `json:"enabled"`
	EntityID   string `json:"entity_id"`
	SSOURL     string `json:"sso_url"`
	SLOURL     string `json:"slo_url,omitempty"`
	Certificate string `json:"certificate,omitempty"`
}

func (h *SAMlHandler) HandleGetSAMLConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	samlConfigRepo := storage.NewSAMLConfigRepository(h.db)
	samlConfig, err := samlConfigRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		if err.Error() == "SAML config not found for tenant" {
			w.WriteHeader(http.StatusOK)
			writeJSON(w, AdminSAMLConfig{
				TenantID: tenantIDStr,
				Enabled:  false,
			})
			return
		}
		logrus.WithError(err).Error("Failed to get SAML config")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get SAML configuration")
		return
	}

	entityID := ""
	if samlConfig.IDPEntityID != nil {
		entityID = *samlConfig.IDPEntityID
	}
	ssoURL := ""
	if samlConfig.IDPSSOURL != nil {
		ssoURL = *samlConfig.IDPSSOURL
	}
	cert := ""
	if samlConfig.IDPCertificate != nil {
		cert = *samlConfig.IDPCertificate
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, AdminSAMLConfig{
		ID:          samlConfig.ID.String(),
		TenantID:    tenantIDStr,
		EntityID:    entityID,
		SSOURL:      ssoURL,
		Certificate: cert,
		Enabled:     samlConfig.Enabled,
	})
}

func (h *SAMlHandler) HandleUpdateSAMLConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req UpdateSAMLConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EntityID == "" || req.SSOURL == "" {
		writeJSONError(w, http.StatusBadRequest, "entity_id and sso_url are required")
		return
	}

	samlConfigRepo := storage.NewSAMLConfigRepository(h.db)
	samlConfig := &storage.SAMLConfig{
		TenantID:       tenantID,
		Enabled:        req.Enabled,
		IDPEntityID:    &req.EntityID,
		IDPSSOURL:      &req.SSOURL,
		IDPCertificate: &req.Certificate,
	}

	ctx := context.Background()
	existing, err := samlConfigRepo.GetByTenantID(ctx, tenantID)
	if err != nil && err.Error() != "SAML config not found for tenant" {
		logrus.WithError(err).Error("Failed to check existing SAML config")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save SAML configuration")
		return
	}

	if existing != nil {
		samlConfig.ID = existing.ID
		samlConfig.CreatedAt = existing.CreatedAt
		if err := samlConfigRepo.Update(ctx, samlConfig); err != nil {
			logrus.WithError(err).Error("Failed to update SAML config")
			writeJSONError(w, http.StatusInternalServerError, "Failed to save SAML configuration")
			return
		}
	} else {
		if err := samlConfigRepo.Create(ctx, samlConfig); err != nil {
			logrus.WithError(err).Error("Failed to create SAML config")
			writeJSONError(w, http.StatusInternalServerError, "Failed to save SAML configuration")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]string{"message": "SAML configuration saved"})
}

func (h *SAMlHandler) HandleGetSAMLMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	baseURL := "https://api.functionfly.com"
	if envURL := r.Header.Get("X-Forwarded-Host"); envURL != "" {
		proto := "https"
		if r.Header.Get("X-Forwarded-Proto") == "http" {
			proto = "http"
		}
		baseURL = proto + "://" + envURL
	}

	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="functionfly">
  <md:SPSSODescriptor AuthnRequestsSigned="false" WantAssertionsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>
    <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s/v1/auth/saml/sso" index="0" isDefault="true"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`, baseURL)

	_ = tenantID

	w.WriteHeader(http.StatusOK)
	writeJSON(w, AdminSAMLMetadata{
		EntityID:    "functionfly",
		ACSURL:      baseURL + "/v1/auth/saml/sso",
		SLOURL:      "",
		MetadataXML: metadata,
	})
}
