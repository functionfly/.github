package trustapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ============================================
// Partner Management Endpoints
// ============================================

// HandleCreatePartner handles POST /v1/partners
// Registers a new partner (self-service onboarding)
func (h *Handler) HandleCreatePartner(w http.ResponseWriter, r *http.Request) {
	var req trustapi.PartnerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Name is required", "validation_error")
		return
	}
	if req.Slug == "" {
		h.writeError(w, http.StatusBadRequest, "Slug is required", "validation_error")
		return
	}
	if req.ContactEmail == "" {
		h.writeError(w, http.StatusBadRequest, "Contact email is required", "validation_error")
		return
	}

	// Check if slug is already taken
	existingPartner, _ := h.repo.GetPartnerBySlug(req.Slug)
	if existingPartner != nil {
		h.writeError(w, http.StatusConflict, "Slug already in use", "slug_conflict")
		return
	}

	// Check if email is already registered
	existingPartner, _ = h.repo.GetPartnerByContactEmail(req.ContactEmail)
	if existingPartner != nil {
		h.writeError(w, http.StatusConflict, "Email already registered", "email_conflict")
		return
	}

	// Get tier config for rate limits
	tier := req.Tier
	if tier == "" {
		tier = string(trustapi.PartnerTierDeveloper)
	}
	tierConfig := trustapi.GetRateLimitConfig(tier)

	// Create partner
	partner := &trustapi.TrustAPIPartner{
		Name:                req.Name,
		Slug:                req.Slug,
		Description:        req.Description,
		ContactEmail:       req.ContactEmail,
		ContactName:        req.ContactName,
		WebsiteURL:         req.WebsiteURL,
		Tier:               tier,
		RateLimitPerMinute: tierConfig.PerMinute,
		RateLimitPerDay:    tierConfig.PerDay,
		MonthlyRequestLimit: tierConfig.MonthlyRequestLimit,
		Status:             string(trustapi.PartnerStatusPending),
	}

	if err := h.repo.CreatePartner(partner); err != nil {
		h.logger.WithError(err).Error("Failed to create partner")
		h.writeError(w, http.StatusInternalServerError, "Failed to create partner", "internal_error")
		return
	}

	response := trustapi.PartnerResponse{
		ID:                  partner.ID,
		Name:                partner.Name,
		Slug:                partner.Slug,
		Description:        partner.Description,
		ContactEmail:       partner.ContactEmail,
		ContactName:        partner.ContactName,
		WebsiteURL:         partner.WebsiteURL,
		Tier:                partner.Tier,
		RateLimitPerMinute:  partner.RateLimitPerMinute,
		RateLimitPerDay:     partner.RateLimitPerDay,
		MonthlyRequestLimit: partner.MonthlyRequestLimit,
		CurrentMonthUsage:   partner.CurrentMonthUsage,
		Status:              partner.Status,
		CreatedAt:           partner.CreatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleListPartners handles GET /v1/partners
// Lists all partners (admin only)
func (h *Handler) HandleListPartners(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	tier := r.URL.Query().Get("tier")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	partners, total, err := h.repo.ListPartners(status, tier, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list partners")
		h.writeError(w, http.StatusInternalServerError, "Failed to list partners", "internal_error")
		return
	}

	// Convert to response format
	responses := make([]trustapi.PartnerResponse, len(partners))
	for i, p := range partners {
		responses[i] = trustapi.PartnerResponse{
			ID:                  p.ID,
			Name:                p.Name,
			Slug:                p.Slug,
			Description:        p.Description,
			ContactEmail:       p.ContactEmail,
			ContactName:        p.ContactName,
			WebsiteURL:         p.WebsiteURL,
			Tier:                p.Tier,
			RateLimitPerMinute: p.RateLimitPerMinute,
			RateLimitPerDay:     p.RateLimitPerDay,
			MonthlyRequestLimit: p.MonthlyRequestLimit,
			CurrentMonthUsage:   p.CurrentMonthUsage,
			Status:              p.Status,
			CreatedAt:           p.CreatedAt,
			ActivatedAt:         p.ActivatedAt,
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"partners":    responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// HandleGetPartner handles GET /v1/partners/{partner_id}
// Gets a specific partner's details
func (h *Handler) HandleGetPartner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerIDStr := vars["partner_id"]

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	partner, err := h.repo.GetPartnerByID(partnerID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Partner not found", "partner_not_found")
		return
	}

	response := trustapi.PartnerResponse{
		ID:                  partner.ID,
		Name:                partner.Name,
		Slug:                partner.Slug,
		Description:        partner.Description,
		ContactEmail:       partner.ContactEmail,
		ContactName:        partner.ContactName,
		WebsiteURL:         partner.WebsiteURL,
		Tier:                partner.Tier,
		RateLimitPerMinute: partner.RateLimitPerMinute,
		RateLimitPerDay:     partner.RateLimitPerDay,
		MonthlyRequestLimit: partner.MonthlyRequestLimit,
		CurrentMonthUsage:   partner.CurrentMonthUsage,
		Status:              partner.Status,
		CreatedAt:           partner.CreatedAt,
		ActivatedAt:         partner.ActivatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleUpdatePartner handles PATCH /v1/partners/{partner_id}
// Updates a partner's details
func (h *Handler) HandleUpdatePartner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerIDStr := vars["partner_id"]

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	partner, err := h.repo.GetPartnerByID(partnerID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Partner not found", "partner_not_found")
		return
	}

	// Parse update fields
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		partner.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		partner.Description = description
	}
	if contactName, ok := updates["contact_name"].(string); ok {
		partner.ContactName = contactName
	}
	if websiteURL, ok := updates["website_url"].(string); ok {
		partner.WebsiteURL = websiteURL
	}
	if tier, ok := updates["tier"].(string); ok {
		partner.Tier = tier
		// Update rate limits based on new tier
		tierConfig := trustapi.GetRateLimitConfig(tier)
		partner.RateLimitPerMinute = tierConfig.PerMinute
		partner.RateLimitPerDay = tierConfig.PerDay
		partner.MonthlyRequestLimit = tierConfig.MonthlyRequestLimit
	}
	if webhookURL, ok := updates["webhook_url"].(string); ok {
		partner.WebhookURL = webhookURL
	}
	if status, ok := updates["status"].(string); ok {
		if err := h.repo.UpdatePartnerStatus(partnerID, status); err != nil {
			h.logger.WithError(err).Error("Failed to update partner status")
			h.writeError(w, http.StatusInternalServerError, "Failed to update partner status", "internal_error")
			return
		}
		partner.Status = status
	}

	if err := h.repo.UpdatePartner(partner); err != nil {
		h.logger.WithError(err).Error("Failed to update partner")
		h.writeError(w, http.StatusInternalServerError, "Failed to update partner", "internal_error")
		return
	}

	response := trustapi.PartnerResponse{
		ID:                  partner.ID,
		Name:                partner.Name,
		Slug:                partner.Slug,
		Description:        partner.Description,
		ContactEmail:       partner.ContactEmail,
		ContactName:        partner.ContactName,
		WebsiteURL:         partner.WebsiteURL,
		Tier:                partner.Tier,
		RateLimitPerMinute: partner.RateLimitPerMinute,
		RateLimitPerDay:     partner.RateLimitPerDay,
		MonthlyRequestLimit: partner.MonthlyRequestLimit,
		CurrentMonthUsage:   partner.CurrentMonthUsage,
		Status:              partner.Status,
		CreatedAt:           partner.CreatedAt,
		ActivatedAt:         partner.ActivatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetPartnerUsage handles GET /v1/partners/{partner_id}/usage
// Gets usage statistics for a partner
func (h *Handler) HandleGetPartnerUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerIDStr := vars["partner_id"]

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	// Parse date range (default: last 30 days)
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid start_date format (use RFC3339)", "invalid_date_format")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid end_date format (use RFC3339)", "invalid_date_format")
			return
		}
	} else {
		endDate = time.Now()
	}

	usage, err := h.repo.GetUsageForPartner(partnerID, startDate, endDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get partner usage")
		h.writeError(w, http.StatusInternalServerError, "Failed to get partner usage", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, usage)
}

// HandleCreateAPIKey handles POST /v1/partners/{partner_id}/api-keys
// Creates a new API key for a partner
func (h *Handler) HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerIDStr := vars["partner_id"]

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	// Verify partner exists
	partner, err := h.repo.GetPartnerByID(partnerID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Partner not found", "partner_not_found")
		return
	}

	var req trustapi.APIKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request")
		return
	}

	// Validate scopes
	if len(req.Scopes) == 0 {
		req.Scopes = trustapi.DefaultAPIKeyScopes()
	}

	// Validate scopes are valid
	for _, scope := range req.Scopes {
		valid := false
		for _, s := range trustapi.AllAPIKeyScopes {
			if scope == s {
				valid = true
				break
			}
		}
		if !valid {
			h.writeError(w, http.StatusBadRequest, "Invalid scope: "+scope, "invalid_scope")
			return
		}
	}

	// Generate API key
	createdBy := "partner:" + partner.ID.String()
	apiKey, rawKey, err := h.repo.GenerateAPIKey(partnerID, &req, createdBy)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate API key")
		h.writeError(w, http.StatusInternalServerError, "Failed to generate API key", "internal_error")
		return
	}

	// Parse scopes for response
	var scopes []string
	json.Unmarshal(apiKey.Scopes, &scopes)
	var allowedIPs []string
	json.Unmarshal(apiKey.AllowedIPs, &allowedIPs)

	// Return response with the actual key (only shown once)
	response := trustapi.APIKeyCreatedResponse{
		APIKeyResponse: trustapi.APIKeyResponse{
			ID:          apiKey.ID,
			KeyID:       apiKey.KeyID,
			KeyPrefix:   apiKey.KeyPrefix,
			Name:        apiKey.Name,
			Description: apiKey.Description,
			Scopes:      scopes,
			AllowedIPs:  allowedIPs,
			ExpiresAt:  apiKey.ExpiresAt,
			IsRevoked:  apiKey.IsRevoked,
			UseCount:   apiKey.UseCount,
			CreatedAt:  apiKey.CreatedAt,
			CreatedBy:  apiKey.CreatedBy,
		},
		Key:       rawKey,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	}

	// Also include partner info for convenience
	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"api_key": response,
		"partner": trustapi.PartnerResponse{
			ID:                  partner.ID,
			Name:                partner.Name,
			Slug:                partner.Slug,
			Tier:                partner.Tier,
			Status:              partner.Status,
			RateLimitPerMinute:  partner.RateLimitPerMinute,
			RateLimitPerDay:     partner.RateLimitPerDay,
		},
		"message": "Save the API key securely. It will not be shown again.",
	})
}

// HandleListAPIKeys handles GET /v1/partners/{partner_id}/api-keys
// Lists all API keys for a partner
func (h *Handler) HandleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerIDStr := vars["partner_id"]

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	includeRevoked := r.URL.Query().Get("include_revoked") == "true"

	keys, err := h.repo.ListAPIKeysForPartner(partnerID, includeRevoked)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list API keys")
		h.writeError(w, http.StatusInternalServerError, "Failed to list API keys", "internal_error")
		return
	}

	// Convert to response format
	responses := make([]trustapi.APIKeyResponse, len(keys))
	for i, k := range keys {
		var scopes []string
		json.Unmarshal(k.Scopes, &scopes)
		var allowedIPs []string
		json.Unmarshal(k.AllowedIPs, &allowedIPs)

		responses[i] = trustapi.APIKeyResponse{
			ID:          k.ID,
			KeyID:       k.KeyID,
			KeyPrefix:   k.KeyPrefix,
			Name:        k.Name,
			Description: k.Description,
			Scopes:      scopes,
			AllowedIPs:  allowedIPs,
			ExpiresAt:   k.ExpiresAt,
			IsRevoked:   k.IsRevoked,
			LastUsedAt:  k.LastUsedAt,
			UseCount:    k.UseCount,
			CreatedAt:   k.CreatedAt,
			CreatedBy:   k.CreatedBy,
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"api_keys": responses,
		"total":    len(responses),
	})
}

// HandleRevokeAPIKey handles DELETE /v1/partners/{partner_id}/api-keys/{key_id}
// Revokes an API key
func (h *Handler) HandleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyIDStr := vars["key_id"]

	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid key ID", "invalid_key_id")
		return
	}

	// Parse revocation reason from body (optional)
	var reqBody map[string]string
	json.NewDecoder(r.Body).Decode(&reqBody)
	reason := "User requested revocation"
	if reqBody != nil && reqBody["reason"] != "" {
		reason = reqBody["reason"]
	}

	if err := h.repo.RevokeAPIKey(keyID, reason); err != nil {
		h.logger.WithError(err).Error("Failed to revoke API key")
		h.writeError(w, http.StatusInternalServerError, "Failed to revoke API key", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"message": "API key revoked successfully",
	})
}
