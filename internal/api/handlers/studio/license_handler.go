package studio

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func licenseErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidLicenseKey),
		errors.Is(err, ErrLicenseRevoked),
		errors.Is(err, ErrLicenseExpired),
		errors.Is(err, ErrActivationLimit):
		return http.StatusForbidden
	case errors.Is(err, ErrLicenseNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidLicenseType),
		errors.Is(err, ErrInvalidSPDXLicense):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// HandleGetLicensePolicy handles GET /v1/marketplace/functions/{id}/license
func (h *MarketplaceHandler) HandleGetLicensePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	policy, err := h.repo.GetLicensePolicy(r.Context(), tenantID, functionID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get license policy")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get license policy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"policy": policy})
}

// HandleUpdateLicense handles PUT /v1/marketplace/functions/{id}/license
func (h *MarketplaceHandler) HandleUpdateLicense(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		License           string  `json:"license"`
		SPDXLicense       string  `json:"spdx_license"`
		CustomLicenseText *string `json:"custom_license_text"`
		CommercialType    string  `json:"commercial_type"`
		MaxActivations    *int    `json:"max_activations_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	spdx := spdxLicenseFromRequest(req.License, req.SPDXLicense)
	if spdx == "" {
		writeJSONError(w, http.StatusBadRequest, "license or spdx_license is required")
		return
	}

	policy, err := h.repo.UpsertLicensePolicy(
		r.Context(),
		tenantID,
		functionID,
		spdx,
		req.CustomLicenseText,
		req.CommercialType,
		req.MaxActivations,
	)
	if err != nil {
		status := licenseErrorStatus(err)
		msg := "Failed to update license"
		if status == http.StatusBadRequest {
			msg = err.Error()
		}
		logrus.WithError(err).Error("marketplace: failed to update license policy")
		writeJSONError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "License updated",
		"policy":  policy,
	})
}

func spdxLicenseFromRequest(legacy, spdx string) string {
	if spdx != "" {
		return spdx
	}
	return legacy
}

// HandleListLicenses handles GET /v1/marketplace/licenses
func (h *MarketplaceHandler) HandleListLicenses(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	params := ListLicenseGrantsParams{TenantID: tenantID, Limit: 50}
	if fn := r.URL.Query().Get("function_id"); fn != "" {
		params.FunctionID = &fn
	}
	if revoked := r.URL.Query().Get("revoked"); revoked != "" {
		v := revoked == "true"
		params.Revoked = &v
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			params.Limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			params.Offset = n
		}
	}

	licenses, active, revoked, err := h.repo.ListLicenseGrants(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to list licenses")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"licenses":     []LicenseGrant{},
			"totalActive":  0,
			"totalRevoked": 0,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"licenses":     licenses,
		"totalActive":  active,
		"totalRevoked": revoked,
	})
}

// HandleCreateLicense handles POST /v1/marketplace/licenses
func (h *MarketplaceHandler) HandleCreateLicense(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		FunctionID     string  `json:"function_id"`
		FunctionName   string  `json:"function_name"`
		LicenseType    string  `json:"license_type"`
		PurchaserName  string  `json:"purchaser_name"`
		PurchaserID    string  `json:"purchaser_id"`
		MaxActivations *int    `json:"max_activations"`
		ExpiresAt      *string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.FunctionID == "" {
		writeJSONError(w, http.StatusBadRequest, "function_id is required")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpiresAt))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "expires_at must be RFC3339")
			return
		}
		expiresAt = &parsed
	}

	grant, err := h.repo.CreateLicenseGrant(r.Context(), CreateLicenseGrantParams{
		TenantID:       tenantID,
		UserID:         userID,
		FunctionID:     req.FunctionID,
		FunctionName:   req.FunctionName,
		LicenseType:    req.LicenseType,
		PurchaserName:  req.PurchaserName,
		PurchaserID:    req.PurchaserID,
		MaxActivations: req.MaxActivations,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		status := licenseErrorStatus(err)
		msg := "Failed to create license"
		if status == http.StatusBadRequest {
			msg = err.Error()
		}
		logrus.WithError(err).Error("marketplace: failed to create license")
		writeJSONError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"license": grant})
}

// HandleRevokeLicense handles POST /v1/marketplace/licenses/{id}/revoke
func (h *MarketplaceHandler) HandleRevokeLicense(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	grantID := mux.Vars(r)["id"]
	if grantID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.repo.RevokeLicenseGrant(r.Context(), tenantID, userID, grantID); err != nil {
		status := licenseErrorStatus(err)
		msg := "Failed to revoke license"
		if status == http.StatusNotFound {
			msg = "License not found"
		}
		logrus.WithError(err).Error("marketplace: failed to revoke license")
		writeJSONError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "License revoked"})
}

// HandleActivateLicense handles POST /v1/marketplace/licenses/activate
func (h *MarketplaceHandler) HandleActivateLicense(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		LicenseKey      string `json:"license_key"`
		ActivationLabel string `json:"activation_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if strings.TrimSpace(req.LicenseKey) == "" {
		writeJSONError(w, http.StatusBadRequest, "license_key is required")
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	grant, err := h.repo.ActivateLicenseGrant(r.Context(), ActivateLicenseParams{
		LicenseKey:      req.LicenseKey,
		TenantID:        tenantID,
		UserID:          getUserID(r),
		ActivationLabel: req.ActivationLabel,
		IPAddress:       ip,
	})
	if err != nil {
		status := licenseErrorStatus(err)
		msg := err.Error()
		logrus.WithError(err).Warn("marketplace: license activation failed")
		writeJSONError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "License activated",
		"license": grant,
	})
}
