package billing

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaxSettingsResponse represents tax settings in API responses
type TaxSettingsResponse struct {
	BillingCountry    *string `json:"billing_country,omitempty"`
	BillingState      *string `json:"billing_state,omitempty"`
	BillingPostalCode *string `json:"billing_postal_code,omitempty"`
	TaxID             *string `json:"tax_id,omitempty"`
	TaxIDType         *string `json:"tax_id_type,omitempty"`
	TaxStatus         string  `json:"tax_status"`
	TaxExempt         bool    `json:"tax_exempt"`
	ApplicableTaxType *string `json:"applicable_tax_type,omitempty"`
}

// UpdateTaxSettingsRequest represents the request to update tax settings
type UpdateTaxSettingsRequest struct {
	BillingCountry    string `json:"billing_country,omitempty"`
	BillingState      string `json:"billing_state,omitempty"`
	BillingPostalCode string `json:"billing_postal_code,omitempty"`
	TaxID             string `json:"tax_id,omitempty"`
	TaxIDType         string `json:"tax_id_type,omitempty"`
	TaxExempt         *bool  `json:"tax_exempt,omitempty"`
}

// Validate implements the ValidatedRequest interface
func (r UpdateTaxSettingsRequest) Validate() error {
	// Basic validation - if tax ID is provided, type must also be provided
	if r.TaxID != "" && r.TaxIDType == "" {
		return &ValidationError{Field: "tax_id_type", Message: "tax_id_type is required when tax_id is provided"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// TaxTypesResponse represents available tax types for a country
type TaxTypesResponse struct {
	Country  string              `json:"country"`
	TaxTypes []map[string]string `json:"tax_types"`
	Required bool                `json:"required"`
}

// TaxCalculationRequest represents a request to calculate tax
type TaxCalculationRequest struct {
	AmountCents     int    `json:"amount_cents"`
	Currency        string `json:"currency"`
	TransactionType string `json:"transaction_type"` // 'subscription', 'one_time', 'usage'
}

// TaxCalculationResponse represents the result of a tax calculation
type TaxCalculationResponse struct {
	TaxAmountCents      int64                      `json:"tax_amount_cents"`
	SubtotalCents       int64                      `json:"subtotal_cents"`
	TotalCents          int64                      `json:"total_cents"`
	Currency            string                     `json:"currency"`
	TaxRatePercentage   float64                    `json:"tax_rate_percentage"`
	TaxName             string                     `json:"tax_name"`
	Jurisdiction        string                     `json:"jurisdiction"`
	TaxBreakdown        []payment.TaxBreakdownItem `json:"tax_breakdown,omitempty"`
	StripeCalculationID string                     `json:"stripe_tax_calculation_id,omitempty"`
}

// HandleGetTaxSettings returns the current tenant's tax settings
// GET /v1/billing/tax/settings
func (h *Handler) HandleGetTaxSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("tax settings: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tax settings")
		return
	}
	if tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	// Determine applicable tax type based on country
	var applicableTaxType *string
	if tenant.BillingCountry != nil && *tenant.BillingCountry != "" {
		types := payment.GetApplicableTaxTypes(*tenant.BillingCountry)
		if len(types) > 0 {
			taxType := types[0]["type"]
			applicableTaxType = &taxType
		}
	}

	response := TaxSettingsResponse{
		BillingCountry:    tenant.BillingCountry,
		BillingState:      tenant.BillingState,
		BillingPostalCode: tenant.BillingPostalCode,
		TaxID:             tenant.TaxID,
		TaxIDType:         tenant.TaxIDType,
		TaxStatus:         tenant.TaxStatus,
		TaxExempt:         tenant.TaxExempt,
		ApplicableTaxType: applicableTaxType,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateTaxSettings updates the tenant's tax settings
// POST /v1/billing/tax/settings
func (h *Handler) HandleUpdateTaxSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateTaxSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			writeJSONError(w, http.StatusBadRequest, ve.Message)
			return
		}
		apierror.LogAndBadRequest(w, r, err, "validate tax request")
		return
	}

	taxService := payment.NewTaxService()

	// Convert request to storage.TaxSettings
	taxSettings := &storage.TaxSettings{
		BillingCountry:    nilIfEmpty(req.BillingCountry),
		BillingState:      nilIfEmpty(req.BillingState),
		BillingPostalCode: nilIfEmpty(req.BillingPostalCode),
		TaxID:             nilIfEmpty(req.TaxID),
		TaxIDType:         nilIfEmpty(req.TaxIDType),
	}

	if req.TaxExempt != nil {
		taxSettings.TaxExempt = *req.TaxExempt
	}

	// Validate tax ID if provided
	if taxSettings.TaxID != nil && *taxSettings.TaxID != "" && taxSettings.TaxIDType != nil {
		valid, msg := taxService.ValidateTaxID(*taxSettings.TaxID, *taxSettings.TaxIDType)
		if !valid {
			logrus.WithFields(logrus.Fields{
				"tenant_id":   claims.TenantID,
				"tax_id":      *taxSettings.TaxID,
				"tax_id_type": *taxSettings.TaxIDType,
				"error":       msg,
			}).Warn("tax settings: invalid tax ID format")
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Invalid tax ID: %s", msg))
			return
		}
	}

	// Update tax settings
	if err := taxService.UpdateTenantTaxSettings(r.Context(), h.repo, claims.TenantID, taxSettings); err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("tax settings: failed to update")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update tax settings")
		return
	}

	// Get updated settings
	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve updated settings")
		return
	}

	var applicableTaxType *string
	if tenant.BillingCountry != nil && *tenant.BillingCountry != "" {
		types := payment.GetApplicableTaxTypes(*tenant.BillingCountry)
		if len(types) > 0 {
			taxType := types[0]["type"]
			applicableTaxType = &taxType
		}
	}

	response := TaxSettingsResponse{
		BillingCountry:    tenant.BillingCountry,
		BillingState:      tenant.BillingState,
		BillingPostalCode: tenant.BillingPostalCode,
		TaxID:             tenant.TaxID,
		TaxIDType:         tenant.TaxIDType,
		TaxStatus:         tenant.TaxStatus,
		TaxExempt:         tenant.TaxExempt,
		ApplicableTaxType: applicableTaxType,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetTaxTypes returns the applicable tax types for a country
// GET /v1/billing/tax/types?country=XX
func (h *Handler) HandleGetTaxTypes(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		writeJSONError(w, http.StatusBadRequest, "country parameter is required")
		return
	}

	taxTypes := payment.GetApplicableTaxTypes(country)

	// Check if tax ID is required (has specific type, not "other")
	required := false
	if len(taxTypes) > 0 && taxTypes[0]["type"] != "other" {
		required = true
	}

	response := TaxTypesResponse{
		Country:  country,
		TaxTypes: taxTypes,
		Required: required,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleCalculateTax calculates tax for a hypothetical transaction
// POST /v1/billing/tax/calculate
func (h *Handler) HandleCalculateTax(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req TaxCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.AmountCents <= 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_cents must be greater than 0")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tenant")
		return
	}
	if tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	// Check if we have enough information for tax calculation
	if tenant.BillingCountry == nil || *tenant.BillingCountry == "" {
		writeJSONError(w, http.StatusBadRequest, "Billing country is required for tax calculation. Please update your tax settings first.")
		return
	}

	taxService := payment.NewTaxService()
	if !taxService.IsEnabled() {
		writeJSONError(w, http.StatusServiceUnavailable, "Tax calculation is not available")
		return
	}

	// Get customer ID if available
	var customerID string
	if tenant.StripeCustomerID != nil {
		customerID = *tenant.StripeCustomerID
	}

	customerCountry := ""
	if tenant.BillingCountry != nil {
		customerCountry = *tenant.BillingCountry
	}

	customerState := ""
	if tenant.BillingState != nil {
		customerState = *tenant.BillingState
	}

	customerPostalCode := ""
	if tenant.BillingPostalCode != nil {
		customerPostalCode = *tenant.BillingPostalCode
	}

	calcParams := payment.TaxCalculationParams{
		AmountCents:         int64(req.AmountCents),
		Currency:            req.Currency,
		CustomerID:          customerID,
		CustomerCountry:     customerCountry,
		CustomerState:       customerState,
		CustomerPostalCode:  customerPostalCode,
		TaxID:               "",
		TaxIDType:           "",
		LineItemDescription: req.TransactionType,
	}

	result, err := taxService.CalculateTax(r.Context(), calcParams)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": claims.TenantID,
			"amount":    req.AmountCents,
		}).Warn("tax calculation failed")

		// Return a response without tax if calculation fails
		response := TaxCalculationResponse{
			TaxAmountCents:    0,
			SubtotalCents:     int64(req.AmountCents),
			TotalCents:        int64(req.AmountCents),
			Currency:          req.Currency,
			TaxRatePercentage: 0,
			TaxName:           "Tax calculation unavailable",
			Jurisdiction:      "",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TaxCalculationResponse{
		TaxAmountCents:      result.TaxAmountCents,
		SubtotalCents:       result.SubtotalCents,
		TotalCents:          result.TotalCents,
		Currency:            req.Currency,
		TaxRatePercentage:   result.TaxRatePercentage,
		TaxName:             result.TaxName,
		Jurisdiction:        result.Jurisdiction,
		TaxBreakdown:        result.TaxBreakdown,
		StripeCalculationID: result.StripeCalculationID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleValidateTaxID validates a tax ID format without saving it
// POST /v1/billing/tax/validate
func (h *Handler) HandleValidateTaxID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaxID     string `json:"tax_id"`
		TaxIDType string `json:"tax_id_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.TaxID == "" || req.TaxIDType == "" {
		writeJSONError(w, http.StatusBadRequest, "tax_id and tax_id_type are required")
		return
	}

	taxService := payment.NewTaxService()
	valid, message := taxService.ValidateTaxID(req.TaxID, req.TaxIDType)

	response := struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message,omitempty"`
	}{
		Valid:   valid,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	if valid {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}

// nilIfEmpty returns nil if the string is empty, otherwise returns a pointer to the string
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
