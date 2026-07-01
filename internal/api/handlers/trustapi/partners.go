package trustapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	slugRegex  = regexp.MustCompile(`^[a-z0-9]+([a-z0-9-]*[a-z0-9]+)?$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	validTiers = map[string]bool{
		string(trustapi.PartnerTierDeveloper):   true,
		string(trustapi.PartnerTierPayAsYouGo):  true,
		string(trustapi.PartnerTierStartup):     true,
		string(trustapi.PartnerTierBusiness):    true,
		string(trustapi.PartnerTierEnterprise):  true,
	}
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

	// --- Input validation ---
	if err := validatePartnerCreateRequest(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error(), "validation_error")
		return
	}

	// Normalize
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.ContactEmail = strings.ToLower(strings.TrimSpace(req.ContactEmail))
	req.Description = strings.TrimSpace(req.Description)
	req.ContactName = strings.TrimSpace(req.ContactName)
	req.WebsiteURL = strings.TrimSpace(req.WebsiteURL)

	// --- Turnstile/CAPTCHA verification ---
	if h.turnstileVerifier != nil && h.turnstileVerifier.IsEnabled() {
		token := r.Header.Get("X-Turnstile-Token")
		if token == "" {
			h.writeError(w, http.StatusBadRequest, "Security verification token is required", "captcha_required")
			return
		}
		result, err := h.turnstileVerifier.VerifyToken(token, getClientIP(r))
		if err != nil || !result.Success {
			h.logger.WithError(err).Warn("Turnstile verification failed for partner registration")
			h.writeError(w, http.StatusForbidden, "Security verification failed. Please try again.", "captcha_failed")
			return
		}
	}

	// --- Tier defaulting and validation ---
	tier := req.Tier
	if tier == "" {
		tier = string(trustapi.PartnerTierDeveloper)
	}
	tierConfig := trustapi.GetRateLimitConfig(tier)

	// Auto-activate free-tier partners; paid tiers stay pending until payment
	status := string(trustapi.PartnerStatusPending)
	if tier == string(trustapi.PartnerTierDeveloper) {
		status = string(trustapi.PartnerStatusActive)
	}

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
		Status:             status,
	}

	// --- Transactional create with UNIQUE constraint handling ---
	if err := h.trustRepo.CreatePartnerInTransaction(partner); err != nil {
		if isUniqueViolation(err) {
			h.writeError(w, http.StatusConflict, "A partner with that slug or email already exists", "conflict")
			return
		}
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

// validatePartnerCreateRequest validates all fields of a partner creation request.
func validatePartnerCreateRequest(req *trustapi.PartnerCreateRequest) error {
	// Name: required, 2-255 chars
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fieldError("name", "is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 255 {
		return fieldError("name", "must be between 2 and 255 characters")
	}

	// Slug: required, 2-100 chars, lowercase alphanumeric + hyphens, must start and end with alnum
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if req.Slug == "" {
		return fieldError("slug", "is required")
	}
	if len(req.Slug) < 2 || len(req.Slug) > 100 {
		return fieldError("slug", "must be between 2 and 100 characters")
	}
	if !slugRegex.MatchString(req.Slug) {
		return fieldError("slug", "must contain only lowercase letters, numbers, and hyphens, and must start and end with a letter or number")
	}

	// ContactEmail: required, valid email format
	req.ContactEmail = strings.TrimSpace(req.ContactEmail)
	if req.ContactEmail == "" {
		return fieldError("contact_email", "is required")
	}
	if len(req.ContactEmail) > 255 {
		return fieldError("contact_email", "must not exceed 255 characters")
	}
	if !emailRegex.MatchString(req.ContactEmail) {
		return fieldError("contact_email", "must be a valid email address")
	}

	// Description: optional, max 2000 chars
	if len(req.Description) > 2000 {
		return fieldError("description", "must not exceed 2000 characters")
	}

	// WebsiteURL: optional, valid URL if provided
	if req.WebsiteURL != "" {
		if len(req.WebsiteURL) > 500 {
			return fieldError("website_url", "must not exceed 500 characters")
		}
		parsed, err := url.ParseRequestURI(req.WebsiteURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fieldError("website_url", "must be a valid HTTP or HTTPS URL")
		}
	}

	// Tier: optional, must be one of the allowed values
	if req.Tier != "" {
		if !validTiers[req.Tier] {
			return fieldError("tier", "must be one of: developer, payg, startup, business, enterprise")
		}
	}

	return nil
}

// validationError is a simple error for field validation failures.
type validationError struct {
	field   string
	message string
}

func (e validationError) Error() string {
	return e.field + " " + e.message
}

func fieldError(field, message string) error {
	return validationError{field: field, message: message}
}

// isUniqueViolation checks if a GORM error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "SQLSTATE 23505")
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

	partners, total, err := h.trustRepo.ListPartners(status, tier, pageSize, offset)
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

	partner, err := h.trustRepo.GetPartnerByID(partnerID)
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

	partner, err := h.trustRepo.GetPartnerByID(partnerID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Partner not found", "partner_not_found")
		return
	}

	// Verify ownership: JWT user must own this partner
	claims := middleware.GetUserFromContext(r)
	if claims != nil && claims.Email != partner.ContactEmail {
		h.writeError(w, http.StatusForbidden, "Not authorized to update this partner", "forbidden")
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
		name = strings.TrimSpace(name)
		if len(name) < 2 || len(name) > 255 {
			h.writeError(w, http.StatusBadRequest, "name must be between 2 and 255 characters", "validation_error")
			return
		}
		partner.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		if len(description) > 2000 {
			h.writeError(w, http.StatusBadRequest, "description must not exceed 2000 characters", "validation_error")
			return
		}
		partner.Description = description
	}
	if contactName, ok := updates["contact_name"].(string); ok {
		partner.ContactName = contactName
	}
	if websiteURL, ok := updates["website_url"].(string); ok {
		partner.WebsiteURL = websiteURL
	}
	// Tier changes are NOT allowed via this endpoint — must go through Stripe checkout
	// to prevent partners from self-upgrading without payment
	if _, ok := updates["tier"]; ok {
		h.writeError(w, http.StatusBadRequest, "Tier cannot be changed directly. Use the billing checkout endpoint to upgrade.", "tier_change_forbidden")
		return
	}
	if webhookURL, ok := updates["webhook_url"].(string); ok {
		webhookURL = strings.TrimSpace(webhookURL)
		if webhookURL != "" {
			if len(webhookURL) > 500 {
				h.writeError(w, http.StatusBadRequest, "webhook_url must not exceed 500 characters", "validation_error")
				return
			}
			parsed, err := url.ParseRequestURI(webhookURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				h.writeError(w, http.StatusBadRequest, "webhook_url must be a valid HTTP or HTTPS URL", "validation_error")
				return
			}
		}
		partner.WebhookURL = webhookURL
	}
	// Status changes are NOT allowed via this endpoint — admin-only operation
	if _, ok := updates["status"]; ok {
		h.writeError(w, http.StatusBadRequest, "Status cannot be changed directly. Contact support.", "status_change_forbidden")
		return
	}

	if err := h.trustRepo.UpdatePartner(partner); err != nil {
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

	// Verify ownership: JWT user must own this partner
	partner, err := h.trustRepo.GetPartnerByID(partnerID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Partner not found", "partner_not_found")
		return
	}
	claims := middleware.GetUserFromContext(r)
	if claims != nil && claims.Email != partner.ContactEmail {
		h.writeError(w, http.StatusForbidden, "Not authorized to view this partner's usage", "forbidden")
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

	usage, err := h.trustRepo.GetUsageForPartner(partnerID, startDate, endDate)
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

	// Verify partner exists using trustRepo
	partner, err := h.trustRepo.GetPartnerByID(partnerID)
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

	// Generate API key using apikeyRepo
	createdBy := "partner:" + partner.ID.String()
	apiKeyCreateReq := &apikey.CreateTrustAPIKeyRequest{
		PartnerID:   partnerID,
		Name:        req.Name,
		Description: req.Description,
		Scopes:      req.Scopes,
		AllowedIPs:  req.AllowedIPs,
		ExpiresAt:   req.ExpiresAt,
		CreatedBy:   createdBy,
	}
	apiKey, rawKey, err := h.apikeyRepo.CreateTrustAPIKey(r.Context(), apiKeyCreateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate API key")
		h.writeError(w, http.StatusInternalServerError, "Failed to generate API key", "internal_error")
		return
	}

	// Parse scopes for response
	var scopes []string
	for k := range apiKey.Scopes {
		scopes = append(scopes, k)
	}
	var allowedIPs []string
	for k := range apiKey.AllowedIPs {
		allowedIPs = append(allowedIPs, k)
	}

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
			IsRevoked:   apiKey.IsRevoked,
			UseCount:    apiKey.UseCount,
			CreatedAt:   apiKey.CreatedAt,
			CreatedBy:   apiKey.CreatedBy,
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

	keys, err := h.apikeyRepo.ListAPIKeysForPartner(r.Context(), partnerID, includeRevoked)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list API keys")
		h.writeError(w, http.StatusInternalServerError, "Failed to list API keys", "internal_error")
		return
	}

	// Convert to response format
	responses := make([]trustapi.APIKeyResponse, len(keys))
	for i, k := range keys {
		var scopes []string
		for scope := range k.Scopes {
			scopes = append(scopes, scope)
		}
		var allowedIPs []string
		for ip := range k.AllowedIPs {
			allowedIPs = append(allowedIPs, ip)
		}

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
			CreatedBy:    k.CreatedBy,
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
	partnerIDStr := vars["partner_id"]

	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid key ID", "invalid_key_id")
		return
	}

	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid partner ID", "invalid_partner_id")
		return
	}

	// Verify the key belongs to this partner (ownership check)
	keys, err := h.apikeyRepo.ListAPIKeysForPartner(r.Context(), partnerID, true)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list API keys for ownership check")
		h.writeError(w, http.StatusInternalServerError, "Failed to verify key ownership", "internal_error")
		return
	}
	keyFound := false
	for _, k := range keys {
		if k.ID == keyID {
			keyFound = true
			break
		}
	}
	if !keyFound {
		h.writeError(w, http.StatusForbidden, "API key does not belong to this partner", "forbidden")
		return
	}

	// Parse revocation reason from body (optional)
	var reqBody map[string]string
	json.NewDecoder(r.Body).Decode(&reqBody)
	reason := "User requested revocation"
	if reqBody != nil && reqBody["reason"] != "" {
		reason = reqBody["reason"]
	}

	if err := h.apikeyRepo.RevokeAPIKey(keyID, reason); err != nil {
		h.logger.WithError(err).Error("Failed to revoke API key")
		h.writeError(w, http.StatusInternalServerError, "Failed to revoke API key", "internal_error")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"message": "API key revoked successfully",
	})
}
