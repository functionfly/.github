package billing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/billing"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ExternalBillingHandler handles external billing system integration API endpoints
type ExternalBillingHandler struct {
	exportRepo *storage.ExportRepository
	repo       storage.Repository
	logger     *logrus.Logger
	syncJob    *billing.BillingSyncJob
}

// NewExternalBillingHandler creates a new external billing handler
func NewExternalBillingHandler(exportRepo *storage.ExportRepository, repo storage.Repository, syncJob *billing.BillingSyncJob) *ExternalBillingHandler {
	return &ExternalBillingHandler{
		exportRepo: exportRepo,
		repo:       repo,
		logger:     logrus.New(),
		syncJob:    syncJob,
	}
}

// ==================== External Billing System Endpoints ====================

// CreateExternalBillingSystem creates a new external billing system configuration
//
// POST /api/v1/exports/external-systems
//
// Request body:
//
//	{
//	  "name": "Stripe Production",
//	  "description": "Production Stripe billing system",
//	  "system_type": "stripe",
//	  "auth_type": "api_key",
//	  "api_key": "sk_live_...",
//	  "api_endpoint": "https://api.stripe.com/v1",
//	  "sync_frequency": "hourly",
//	  "sync_direction": "push"
//	}
func (h *ExternalBillingHandler) CreateExternalBillingSystem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	var req struct {
		Name           string                  `json:"name"`
		Description    string                  `json:"description"`
		SystemType     string                  `json:"system_type"`
		AuthType       string                  `json:"auth_type"`
		APIKey         string                  `json:"api_key"`
		APIEndpoint    string                  `json:"api_endpoint"`
		SyncEnabled    bool                    `json:"sync_enabled"`
		SyncFrequency  string                  `json:"sync_frequency"`
		SyncDirection  string                  `json:"sync_direction"`
		FieldMappings  map[string]string       `json:"field_mappings"`
		TransformRules []storage.TransformRule `json:"transform_rules"`
		IsActive       *bool                   `json:"is_active"`
		WebhookURL     string                  `json:"webhook_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "name is required")
		return
	}
	if req.SystemType == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "system_type is required")
		return
	}

	userID := h.extractUserID(r)

	system := &storage.ExternalBillingSystem{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           req.Name,
		Description:    req.Description,
		SystemType:     req.SystemType,
		AuthType:       req.AuthType,
		APIEndpoint:    req.APIEndpoint,
		IsActive:       true,
		SyncEnabled:    req.SyncEnabled,
		SyncFrequency:  req.SyncFrequency,
		SyncDirection:  req.SyncDirection,
		FieldMappings:  req.FieldMappings,
		TransformRules: req.TransformRules,
		WebhookURL:     req.WebhookURL,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if req.IsActive != nil {
		system.IsActive = *req.IsActive
	}

	// Store credentials securely using AES-256-GCM encryption
	if req.APIKey != "" {
		encryptedKey, err := h.repo.EncryptField(req.APIKey)
		if err != nil {
			h.logger.WithError(err).Error("Failed to encrypt API credential")
			h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to secure credentials")
			return
		}
		system.APICredentialKey = encryptedKey
	}

	if err := h.exportRepo.CreateExternalBillingSystem(r.Context(), system); err != nil {
		h.logger.WithError(err).Error("Failed to create external billing system")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to create external billing system")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.formatExternalBillingSystem(system, true))
}

// ListExternalBillingSystems lists all external billing systems for the tenant
//
// GET /api/v1/exports/external-systems
//
// Query params:
//   - limit: Page size (default: 20, max: 100)
//   - offset: Pagination offset
//   - active_only: Only return active systems (default: false)
func (h *ExternalBillingHandler) ListExternalBillingSystems(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	activeOnly := r.URL.Query().Get("active_only") == "true"

	systems, err := h.exportRepo.ListExternalBillingSystems(r.Context(), tenantID, limit, offset, activeOnly)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list external billing systems")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to list external billing systems")
		return
	}

	formatted := make([]map[string]interface{}, len(systems))
	for i, system := range systems {
		formatted[i] = h.formatExternalBillingSystem(system, false)
	}

	response := map[string]interface{}{
		"systems": formatted,
		"limit":   limit,
		"offset":  offset,
		"total":   len(formatted),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExternalBillingSystem gets a specific external billing system
//
// GET /api/v1/exports/external-systems/{id}
func (h *ExternalBillingHandler) GetExternalBillingSystem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	systemID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid system ID")
		return
	}

	system, err := h.exportRepo.GetExternalBillingSystem(r.Context(), systemID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "External billing system not found")
		return
	}

	if system.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExternalBillingSystem(system, true))
}

// UpdateExternalBillingSystem updates an external billing system configuration
//
// PUT /api/v1/exports/external-systems/{id}
func (h *ExternalBillingHandler) UpdateExternalBillingSystem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	systemID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid system ID")
		return
	}

	existing, err := h.exportRepo.GetExternalBillingSystem(r.Context(), systemID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "External billing system not found")
		return
	}

	if existing.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	var updates struct {
		Name           string                  `json:"name"`
		Description    string                  `json:"description"`
		AuthType       string                  `json:"auth_type"`
		APIKey         string                  `json:"api_key"`
		APIEndpoint    string                  `json:"api_endpoint"`
		SyncEnabled    bool                    `json:"sync_enabled"`
		SyncFrequency  string                  `json:"sync_frequency"`
		SyncDirection  string                  `json:"sync_direction"`
		FieldMappings  map[string]string       `json:"field_mappings"`
		TransformRules []storage.TransformRule `json:"transform_rules"`
		IsActive       *bool                   `json:"is_active"`
		WebhookURL     string                  `json:"webhook_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	// Update fields
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.AuthType != "" {
		existing.AuthType = updates.AuthType
	}
	if updates.APIEndpoint != "" {
		existing.APIEndpoint = updates.APIEndpoint
	}
	existing.SyncEnabled = updates.SyncEnabled
	if updates.SyncFrequency != "" {
		existing.SyncFrequency = updates.SyncFrequency
	}
	if updates.SyncDirection != "" {
		existing.SyncDirection = updates.SyncDirection
	}
	if updates.FieldMappings != nil {
		existing.FieldMappings = updates.FieldMappings
	}
	if updates.TransformRules != nil {
		existing.TransformRules = updates.TransformRules
	}
	if updates.IsActive != nil {
		existing.IsActive = *updates.IsActive
	}
	if updates.WebhookURL != "" {
		existing.WebhookURL = updates.WebhookURL
	}

	// Update credentials if provided - encrypt before storing
	if updates.APIKey != "" {
		encryptedKey, err := h.repo.EncryptField(updates.APIKey)
		if err != nil {
			h.logger.WithError(err).Error("Failed to encrypt API credential")
			h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to secure credentials")
			return
		}
		existing.APICredentialKey = encryptedKey
	}

	existing.UpdatedAt = time.Now()

	if err := h.exportRepo.UpdateExternalBillingSystem(r.Context(), existing); err != nil {
		h.logger.WithError(err).Error("Failed to update external billing system")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to update external billing system")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExternalBillingSystem(existing, true))
}

// DeleteExternalBillingSystem deletes an external billing system configuration
//
// DELETE /api/v1/exports/external-systems/{id}
func (h *ExternalBillingHandler) DeleteExternalBillingSystem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	systemID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid system ID")
		return
	}

	existing, err := h.exportRepo.GetExternalBillingSystem(r.Context(), systemID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "External billing system not found")
		return
	}

	if existing.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	if err := h.exportRepo.DeleteExternalBillingSystem(r.Context(), systemID); err != nil {
		h.logger.WithError(err).Error("Failed to delete external billing system")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to delete external billing system")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestExternalBillingSystem tests the connection to an external billing system
//
// POST /api/v1/exports/external-systems/{id}/test
func (h *ExternalBillingHandler) TestExternalBillingSystem(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	systemID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid system ID")
		return
	}

	system, err := h.exportRepo.GetExternalBillingSystem(r.Context(), systemID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "External billing system not found")
		return
	}

	if system.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	// Test the connection
	testResult := h.testConnection(system)

	response := map[string]interface{}{
		"success":   testResult.Success,
		"message":   testResult.Message,
		"tested_at": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ==================== Billing Integration Sync Endpoints ====================

// ListBillingSyncs lists billing sync records
//
// GET /api/v1/exports/syncs
//
// Query params:
//   - limit: Page size (default: 20, max: 100)
//   - offset: Pagination offset
//   - system_id: Filter by external system ID
//   - status: Filter by status (pending, running, completed, failed)
func (h *ExternalBillingHandler) ListBillingSyncs(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse optional system ID filter
	var systemID *uuid.UUID
	if sysIDStr := r.URL.Query().Get("system_id"); sysIDStr != "" {
		if parsed, err := uuid.Parse(sysIDStr); err == nil {
			systemID = &parsed
		}
	}

	status := r.URL.Query().Get("status")

	syncs, err := h.exportRepo.ListBillingIntegrationSyncs(r.Context(), tenantID, systemID, status, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list billing syncs")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to list billing syncs")
		return
	}

	formatted := make([]map[string]interface{}, len(syncs))
	for i, sync := range syncs {
		formatted[i] = h.formatBillingSync(sync)
	}

	response := map[string]interface{}{
		"syncs":  formatted,
		"limit":  limit,
		"offset": offset,
		"total":  len(formatted),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TriggerBillingSync manually triggers a billing sync
//
// POST /api/v1/exports/external-systems/{id}/sync
//
// Request body:
//
//	{
//	  "sync_type": "usage",
//	  "direction": "push"
//	}
func (h *ExternalBillingHandler) TriggerBillingSync(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	systemID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid system ID")
		return
	}

	system, err := h.exportRepo.GetExternalBillingSystem(r.Context(), systemID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "External billing system not found")
		return
	}

	if system.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	var req struct {
		SyncType  string `json:"sync_type"`
		Direction string `json:"direction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	if req.SyncType == "" {
		req.SyncType = "usage"
	}
	if req.Direction == "" {
		req.Direction = "push"
	}

	now := time.Now()
	userID := h.extractUserID(r)
	triggeredBy := "manual"
	if userID != uuid.Nil {
		triggeredBy = "manual"
	}

	sync := &storage.BillingIntegrationSync{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ExternalSystemID: systemID,
		SyncType:         req.SyncType,
		Direction:        req.Direction,
		Status:           "pending",
		StartedAt:        &now,
		TriggeredBy:      triggeredBy,
		CreatedAt:        time.Now(),
	}

	if err := h.exportRepo.CreateBillingIntegrationSync(r.Context(), sync); err != nil {
		h.logger.WithError(err).Error("Failed to create billing sync")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to create billing sync")
		return
	}

	// Trigger the actual sync job asynchronously
	if h.syncJob != nil {
		if err := h.syncJob.TriggerSync(r.Context(), sync.ID, tenantID, systemID); err != nil {
			h.logger.WithError(err).Warn("Failed to trigger sync job, sync will be processed by background worker")
		}
	} else {
		h.logger.Debug("Billing sync job not initialized, sync will be processed by background worker")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(h.formatBillingSync(sync))
}

// GetBillingSync gets a specific billing sync record
//
// GET /api/v1/exports/syncs/{id}
func (h *ExternalBillingHandler) GetBillingSync(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	syncID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid sync ID")
		return
	}

	sync, err := h.exportRepo.GetBillingIntegrationSync(r.Context(), syncID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Billing sync not found")
		return
	}

	if sync.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatBillingSync(sync))
}

// ==================== Billing System Types Endpoints ====================

// GetBillingSystemTypes returns supported billing system types
//
// GET /api/v1/exports/billing-system-types
func (h *ExternalBillingHandler) GetBillingSystemTypes(w http.ResponseWriter, r *http.Request) {
	systemTypes := []map[string]interface{}{
		{
			"id":          "stripe",
			"name":        "Stripe",
			"description": "Stripe is a comprehensive payment processing and billing platform",
			"auth_types":  []string{"api_key", "oauth2"},
			"features":    []string{"invoicing", "subscriptions", "usage_billing", "metered_billing"},
		},
		{
			"id":          "chargebee",
			"name":        "Chargebee",
			"description": "Chargebee is a subscription management and recurring billing platform",
			"auth_types":  []string{"api_key"},
			"features":    []string{"invoicing", "subscriptions", "usage_billing"},
		},
		{
			"id":          "recurly",
			"name":        "Recurly",
			"description": "Recurly is a subscription management platform",
			"auth_types":  []string{"api_key"},
			"features":    []string{"invoicing", "subscriptions"},
		},
		{
			"id":          "zuora",
			"name":        "Zuora",
			"description": "Zuora is an enterprise subscription billing platform",
			"auth_types":  []string{"api_key", "oauth2", "basic_auth"},
			"features":    []string{"invoicing", "subscriptions", "usage_billing", "metered_billing"},
		},
		{
			"id":          "netsuite",
			"name":        "NetSuite",
			"description": "NetSuite is an enterprise resource planning system",
			"auth_types":  []string{"token_based", "basic_auth"},
			"features":    []string{"invoicing"},
		},
		{
			"id":          "salesforce",
			"name":        "Salesforce Billing",
			"description": "Salesforce Billing for CPQ and billing management",
			"auth_types":  []string{"oauth2"},
			"features":    []string{"invoicing", "subscriptions"},
		},
		{
			"id":          "custom",
			"name":        "Custom API",
			"description": "Custom REST API integration",
			"auth_types":  []string{"api_key", "oauth2", "basic_auth"},
			"features":    []string{"invoicing", "usage_billing"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"types": systemTypes,
	})
}

// ==================== Helper Methods ====================

type ConnectionTestResult struct {
	Success bool
	Message string
}

func (h *ExternalBillingHandler) testConnection(system *storage.ExternalBillingSystem) ConnectionTestResult {
	// Decrypt credentials for validation (if encrypted)
	apiKey := system.APICredentialKey
	oauthToken := system.OAuthToken
	if apiKey != "" {
		decrypted, err := h.repo.DecryptField(apiKey)
		if err == nil && decrypted != apiKey {
			apiKey = decrypted
		}
	}
	if system.OAuthToken != "" {
		decrypted, err := h.repo.DecryptField(system.OAuthToken)
		if err == nil && decrypted != system.OAuthToken {
			oauthToken = decrypted
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch system.SystemType {
	case "stripe":
		return h.testStripeConnection(ctx, apiKey)
	case "chargebee":
		return h.testChargebeeConnection(ctx, system.APIEndpoint, apiKey)
	case "recurly":
		return h.testRecurlyConnection(ctx, system.APIEndpoint, apiKey)
	case "zuora":
		return h.testZuoraConnection(ctx, system.APIEndpoint, apiKey, oauthToken)
	case "netsuite":
		return h.testNetSuiteConnection(ctx, system.APIEndpoint, apiKey, system.APICredentialSecret)
	case "salesforce":
		return h.testSalesforceConnection(ctx, system.APIEndpoint, oauthToken)
	case "quickbooks":
		return h.testQuickBooksConnection(ctx, system, oauthToken)
	case "xero":
		return h.testXeroConnection(ctx, system, oauthToken)
	case "custom":
		return h.testCustomConnection(ctx, system.APIEndpoint, apiKey, oauthToken)
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection test not implemented for system type: %s", system.SystemType)}
	}
}

// testStripeConnection tests Stripe API connectivity using the balance endpoint
func (h *ExternalBillingHandler) testStripeConnection(ctx context.Context, apiKey string) ConnectionTestResult {
	if apiKey == "" {
		return ConnectionTestResult{Success: false, Message: "API key is required for Stripe"}
	}
	if len(apiKey) < 10 {
		return ConnectionTestResult{Success: false, Message: "API key appears to be invalid (too short)"}
	}

	endpoint := "https://api.stripe.com/v1/balance"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	switch resp.StatusCode {
	case 200:
		return ConnectionTestResult{Success: true, Message: "Successfully connected to Stripe API"}
	case 401:
		return ConnectionTestResult{Success: false, Message: "Authentication failed: Invalid API key"}
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Stripe API returned %d: %s", resp.StatusCode, string(body))}
	}
}

// testChargebeeConnection tests Chargebee API connectivity
func (h *ExternalBillingHandler) testChargebeeConnection(ctx context.Context, endpoint, apiKey string) ConnectionTestResult {
	if apiKey == "" {
		return ConnectionTestResult{Success: false, Message: "API key is required for Chargebee"}
	}
	if endpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for Chargebee (e.g., https://your-site.chargebee.com)"}
	}

	testURL := endpoint + "/api/v2/customers?limit=1"
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	// Chargebee uses basic auth with API key as username, empty password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(apiKey+":")))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	switch resp.StatusCode {
	case 200:
		return ConnectionTestResult{Success: true, Message: "Successfully connected to Chargebee API"}
	case 401:
		return ConnectionTestResult{Success: false, Message: "Authentication failed: Invalid API key"}
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Chargebee API returned %d: %s", resp.StatusCode, string(body))}
	}
}

// testRecurlyConnection tests Recurly API connectivity
func (h *ExternalBillingHandler) testRecurlyConnection(ctx context.Context, endpoint, apiKey string) ConnectionTestResult {
	if apiKey == "" {
		return ConnectionTestResult{Success: false, Message: "API key is required for Recurly"}
	}
	if endpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for Recurly (e.g., https://your-subdomain.recurly.com)"}
	}

	testURL := endpoint + "/v2/accounts"
	req, err := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	// Recurly uses basic auth with API key
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(apiKey+":")))
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("X-Api-Version", "2.29")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return ConnectionTestResult{Success: true, Message: "Successfully connected to Recurly API"}
	case 401:
		return ConnectionTestResult{Success: false, Message: "Authentication failed: Invalid API key"}
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Recurly API returned HTTP %d", resp.StatusCode)}
	}
}

// testZuoraConnection tests Zuora API connectivity
func (h *ExternalBillingHandler) testZuoraConnection(ctx context.Context, endpoint, apiKey, oauthToken string) ConnectionTestResult {
	if oauthToken == "" && apiKey == "" {
		return ConnectionTestResult{Success: false, Message: "Either OAuth token or API key is required for Zuora"}
	}
	if endpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for Zuora (e.g., https://rest.zuora.com)"}
	}

	testURL := endpoint + "/v1/accounts"
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	if oauthToken != "" {
		req.Header.Set("Authorization", "Bearer "+oauthToken)
	} else {
		req.Header.Set("apiKey", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return ConnectionTestResult{Success: true, Message: "Successfully connected to Zuora API"}
	case 401:
		return ConnectionTestResult{Success: false, Message: "Authentication failed: Invalid credentials"}
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Zuora API returned HTTP %d", resp.StatusCode)}
	}
}

// testNetSuiteConnection tests NetSuite connectivity via RESTlet or SOAP
func (h *ExternalBillingHandler) testNetSuiteConnection(ctx context.Context, endpoint, token, tokenSecret string) ConnectionTestResult {
	if token == "" {
		return ConnectionTestResult{Success: false, Message: "Token is required for NetSuite"}
	}
	if endpoint == "" {
		return ConnectionTestResult{Success: false, Message: "Account ID is required for NetSuite"}
	}

	// NetSuite REST API requires OAuth1 or Token-based authentication
	// This is a simplified check - full implementation would require OAuth1 signature
	return ConnectionTestResult{Success: true, Message: "NetSuite configuration validated (full connection test requires OAuth1 implementation)"}
}

// testSalesforceConnection tests Salesforce Billing API connectivity
func (h *ExternalBillingHandler) testSalesforceConnection(ctx context.Context, endpoint, oauthToken string) ConnectionTestResult {
	if oauthToken == "" {
		return ConnectionTestResult{Success: false, Message: "OAuth token is required for Salesforce"}
	}

	testURL := "https://yourInstance.salesforce.com/services/data/v58.0/"
	if endpoint != "" {
		testURL = endpoint + "/services/data/v58.0/"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+oauthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return ConnectionTestResult{Success: true, Message: "Successfully connected to Salesforce API"}
	case 401:
		return ConnectionTestResult{Success: false, Message: "Authentication failed: Invalid or expired OAuth token"}
	default:
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Salesforce API returned HTTP %d", resp.StatusCode)}
	}
}

// testQuickBooksConnection tests QuickBooks Online API using the existing exporter
func (h *ExternalBillingHandler) testQuickBooksConnection(ctx context.Context, system *storage.ExternalBillingSystem, oauthToken string) ConnectionTestResult {
	if oauthToken == "" {
		return ConnectionTestResult{Success: false, Message: "OAuth token is required for QuickBooks"}
	}
	if system.APIEndpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for QuickBooks (realm ID in URL)"}
	}

	// Use the existing exporter's TestConnection method
	exporter := billing.NewQuickBooksExporter()
	system.OAuthToken = oauthToken
	err := exporter.TestConnection(ctx, system)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("QuickBooks connection failed: %v", err)}
	}
	return ConnectionTestResult{Success: true, Message: "Successfully connected to QuickBooks API"}
}

// testXeroConnection tests Xero API using the existing exporter
func (h *ExternalBillingHandler) testXeroConnection(ctx context.Context, system *storage.ExternalBillingSystem, oauthToken string) ConnectionTestResult {
	if oauthToken == "" {
		return ConnectionTestResult{Success: false, Message: "OAuth token is required for Xero"}
	}
	if system.APIEndpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for Xero"}
	}

	// Use the existing exporter's TestConnection method
	exporter := billing.NewXeroExporter()
	system.OAuthToken = oauthToken
	err := exporter.TestConnection(ctx, system)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Xero connection failed: %v", err)}
	}
	return ConnectionTestResult{Success: true, Message: "Successfully connected to Xero API"}
}

// testCustomConnection tests a custom API endpoint
func (h *ExternalBillingHandler) testCustomConnection(ctx context.Context, endpoint, apiKey, oauthToken string) ConnectionTestResult {
	if endpoint == "" {
		return ConnectionTestResult{Success: false, Message: "API endpoint is required for custom integrations"}
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", endpoint, nil)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	} else if oauthToken != "" {
		req.Header.Set("Authorization", "Bearer "+oauthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Connection failed: %v", err)}
	}
	defer resp.Body.Close()

	// For custom endpoints, accept 2xx status codes
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ConnectionTestResult{Success: true, Message: fmt.Sprintf("Successfully connected to custom endpoint (HTTP %d)", resp.StatusCode)}
	}

	return ConnectionTestResult{Success: false, Message: fmt.Sprintf("Custom endpoint returned HTTP %d", resp.StatusCode)}
}

func (h *ExternalBillingHandler) extractTenantID(r *http.Request) uuid.UUID {
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}
	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID
	}
	return uuid.Nil
}

func (h *ExternalBillingHandler) extractUserID(r *http.Request) uuid.UUID {
	if user, ok := r.Context().Value("user").(*storage.User); ok {
		return user.ID
	}
	return uuid.Nil
}

func (h *ExternalBillingHandler) formatExternalBillingSystem(system *storage.ExternalBillingSystem, includeSecrets bool) map[string]interface{} {
	result := map[string]interface{}{
		"id":               system.ID.String(),
		"tenant_id":        system.TenantID.String(),
		"name":             system.Name,
		"description":      system.Description,
		"system_type":      system.SystemType,
		"auth_type":        system.AuthType,
		"api_endpoint":     system.APIEndpoint,
		"is_active":        system.IsActive,
		"sync_enabled":     system.SyncEnabled,
		"sync_frequency":   system.SyncFrequency,
		"sync_direction":   system.SyncDirection,
		"field_mappings":   system.FieldMappings,
		"transform_rules":  system.TransformRules,
		"webhook_url":      system.WebhookURL,
		"last_sync_at":     system.LastSyncAt,
		"last_sync_status": system.LastSyncStatus,
		"created_by":       system.CreatedBy.String(),
		"created_at":       system.CreatedAt.Format(time.RFC3339),
		"updated_at":       system.UpdatedAt.Format(time.RFC3339),
	}

	// Only include secrets in detailed view (not in list views)
	if includeSecrets {
		result["has_credentials"] = system.APICredentialKey != ""
		result["has_oauth"] = system.OAuthToken != ""
	}

	return result
}

func (h *ExternalBillingHandler) formatBillingSync(sync *storage.BillingIntegrationSync) map[string]interface{} {
	result := map[string]interface{}{
		"id":                 sync.ID.String(),
		"tenant_id":          sync.TenantID.String(),
		"external_system_id": sync.ExternalSystemID.String(),
		"sync_type":          sync.SyncType,
		"direction":          sync.Direction,
		"status":             sync.Status,
		"records_processed":  sync.RecordsProcessed,
		"records_created":    sync.RecordsCreated,
		"records_updated":    sync.RecordsUpdated,
		"records_failed":     sync.RecordsFailed,
		"records_skipped":    sync.RecordsSkipped,
		"triggered_by":       sync.TriggeredBy,
	}

	if sync.StartedAt != nil {
		result["started_at"] = sync.StartedAt.Format(time.RFC3339)
	}
	if sync.CompletedAt != nil {
		result["completed_at"] = sync.CompletedAt.Format(time.RFC3339)
	}
	if sync.ErrorMessage != "" {
		result["error_message"] = sync.ErrorMessage
	}

	return result
}

func (h *ExternalBillingHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
