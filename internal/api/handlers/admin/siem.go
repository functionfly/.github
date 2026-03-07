package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SIEMHandler contains SIEM configuration handlers
type SIEMHandler struct {
	siemService *auth.SIEMService
	siemRepo    *storage.SIEMRepository
}

// NewSIEMHandler creates a new SIEM handler
func NewSIEMHandler(siemService *auth.SIEMService, siemRepo *storage.SIEMRepository) *SIEMHandler {
	return &SIEMHandler{
		siemService: siemService,
		siemRepo:    siemRepo,
	}
}

// RegisterRoutes registers SIEM routes
func (h *SIEMHandler) RegisterRoutes(router *mux.Router) {
	// Tenant-scoped routes
	router.HandleFunc("/tenants/{tenantId}/siem-configs", h.HandleListSIEMConfigs).Methods(http.MethodGet)
	router.HandleFunc("/tenants/{tenantId}/siem-configs", h.HandleCreateSIEMConfig).Methods(http.MethodPost)

	// Config-scoped routes
	router.HandleFunc("/siem-configs/{id}", h.HandleGetSIEMConfig).Methods(http.MethodGet)
	router.HandleFunc("/siem-configs/{id}", h.HandleUpdateSIEMConfig).Methods(http.MethodPut)
	router.HandleFunc("/siem-configs/{id}", h.HandleDeleteSIEMConfig).Methods(http.MethodDelete)
	router.HandleFunc("/siem-configs/{id}/test", h.HandleTestSIEMConfig).Methods(http.MethodPost)
	router.HandleFunc("/siem-configs/{id}/logs", h.HandleGetSIEMExportLogs).Methods(http.MethodGet)
	router.HandleFunc("/siem-configs/{id}/trigger", h.HandleTriggerSIEMExport).Methods(http.MethodPost)
}

// HandleListSIEMConfigs lists all SIEM configs for a tenant
// GET /tenants/{tenantId}/siem-configs
func (h *SIEMHandler) HandleListSIEMConfigs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	configs, err := h.siemRepo.GetByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list SIEM configs")
		http.Error(w, "Failed to list SIEM configs", http.StatusInternalServerError)
		return
	}

	if configs == nil {
		configs = []*storage.SIEMConfig{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"siem_configs": configs,
		"tenant_id":    tenantID,
	})
}

// HandleCreateSIEMConfig creates a new SIEM config
// POST /tenants/{tenantId}/siem-configs
func (h *SIEMHandler) HandleCreateSIEMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name            string                 `json:"name"`
		Enabled         bool                   `json:"enabled"`
		ExportFormat    string                 `json:"export_format"`
		DestinationType string                 `json:"destination_type"`
		Config          map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.DestinationType == "" {
		http.Error(w, "Destination type is required", http.StatusBadRequest)
		return
	}

	if req.Config == nil {
		req.Config = make(map[string]interface{})
	}

	// Validate destination type
	validDestinations := map[string]bool{
		auth.DestinationWebhook:       true,
		auth.DestinationCloudWatch:    true,
		auth.DestinationAzureSentinel: true,
		auth.DestinationGCPChronicle:  true,
		auth.DestinationSplunk:        true,
	}

	if !validDestinations[req.DestinationType] {
		http.Error(w, "Invalid destination type", http.StatusBadRequest)
		return
	}

	// Validate export format
	if req.ExportFormat == "" {
		req.ExportFormat = auth.FormatJSON
	}
	validFormats := map[string]bool{
		auth.FormatJSON: true,
		auth.FormatCEF:  true,
		auth.FormatLEEF: true,
	}
	if !validFormats[req.ExportFormat] {
		http.Error(w, "Invalid export format", http.StatusBadRequest)
		return
	}

	config := &storage.SIEMConfig{
		TenantID:        tenantID,
		Name:            req.Name,
		Enabled:         req.Enabled,
		ExportFormat:    req.ExportFormat,
		DestinationType: req.DestinationType,
		Config:          req.Config,
	}

	if err := h.siemRepo.Create(r.Context(), config); err != nil {
		logrus.WithError(err).Error("Failed to create SIEM config")
		http.Error(w, "Failed to create SIEM config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// HandleGetSIEMConfig gets a SIEM config by ID
// GET /siem-configs/{id}
func (h *SIEMHandler) HandleGetSIEMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	config, err := h.siemRepo.GetByID(r.Context(), configID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SIEM config")
		http.Error(w, "Failed to get SIEM config", http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "SIEM config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateSIEMConfig updates a SIEM config
// PUT /siem-configs/{id}
func (h *SIEMHandler) HandleUpdateSIEMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Get existing config
	existing, err := h.siemRepo.GetByID(r.Context(), configID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SIEM config")
		http.Error(w, "Failed to get SIEM config", http.StatusInternalServerError)
		return
	}

	if existing == nil {
		http.Error(w, "SIEM config not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name            string                 `json:"name"`
		Enabled         bool                   `json:"enabled"`
		ExportFormat    string                 `json:"export_format"`
		DestinationType string                 `json:"destination_type"`
		Config          map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.ExportFormat != "" {
		existing.ExportFormat = req.ExportFormat
	}
	if req.DestinationType != "" {
		existing.DestinationType = req.DestinationType
	}
	if req.Config != nil {
		existing.Config = req.Config
	}

	// Handle enabled separately (it's a bool, so we need to check if it was explicitly set)
	// We use a separate field in the JSON to allow disabling
	var enabledReq struct {
		Enabled *bool `json:"enabled"`
	}
	json.NewDecoder(r.Body).Decode(&enabledReq)
	if enabledReq.Enabled != nil {
		existing.Enabled = *enabledReq.Enabled
	}

	if err := h.siemRepo.Update(r.Context(), existing); err != nil {
		logrus.WithError(err).Error("Failed to update SIEM config")
		http.Error(w, "Failed to update SIEM config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// HandleDeleteSIEMConfig deletes a SIEM config
// DELETE /siem-configs/{id}
func (h *SIEMHandler) HandleDeleteSIEMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	if err := h.siemRepo.Delete(r.Context(), configID); err != nil {
		logrus.WithError(err).Error("Failed to delete SIEM config")
		http.Error(w, "Failed to delete SIEM config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleTestSIEMConfig tests the connection to a SIEM destination
// POST /siem-configs/{id}/test
func (h *SIEMHandler) HandleTestSIEMConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Verify config exists
	config, err := h.siemRepo.GetByID(r.Context(), configID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SIEM config")
		http.Error(w, "Failed to get SIEM config", http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "SIEM config not found", http.StatusNotFound)
		return
	}

	// Test connection
	err = h.siemService.TestConnection(r.Context(), configID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Connection test failed: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connection test successful",
	})
}

// HandleGetSIEMExportLogs gets export logs for a SIEM config
// GET /siem-configs/{id}/logs
func (h *SIEMHandler) HandleGetSIEMExportLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Get limit from query params
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parseLimit(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.siemRepo.GetExportLogs(r.Context(), configID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SIEM export logs")
		http.Error(w, "Failed to get SIEM export logs", http.StatusInternalServerError)
		return
	}

	if logs == nil {
		logs = []*storage.SIEMExportLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":      logs,
		"config_id": configID,
	})
}

// HandleTriggerSIEMExport manually triggers an export
// POST /siem-configs/{id}/trigger
func (h *SIEMHandler) HandleTriggerSIEMExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configIDStr := vars["id"]
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		http.Error(w, "Invalid config ID", http.StatusBadRequest)
		return
	}

	// Verify config exists
	config, err := h.siemRepo.GetByID(r.Context(), configID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get SIEM config")
		http.Error(w, "Failed to get SIEM config", http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "SIEM config not found", http.StatusNotFound)
		return
	}

	if !config.Enabled {
		http.Error(w, "SIEM config is disabled", http.StatusBadRequest)
		return
	}

	// Trigger export in background (don't wait)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		h.siemService.Export(ctx, configID)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Export triggered",
	})
}

// parseLimit parses a limit string to an integer
func parseLimit(s string) (int, error) {
	var limit int
	_, err := parseLimitParams(s, &limit)
	return limit, err
}

func parseLimitParams(s string, limit *int) (bool, error) {
	if s == "" {
		return false, nil
	}
	n, err := parsePositiveInt(s)
	if err != nil || n > 1000 {
		return false, err
	}
	*limit = n
	return true, nil
}

func parsePositiveInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
