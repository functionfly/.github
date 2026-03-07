// Package scim provides SCIM 2.0 provisioning support for GoBetterAuth
// This is Phase 4 of the Better Auth migration plan - Enterprise user lifecycle management
package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SCIMPlugin provides SCIM 2.0 provisioning functionality
type SCIMPlugin struct {
	db      *gorm.DB
	service *SCIMService
	config  *SCIMPluginConfig
	logger  *logrus.Logger
}

// SCIMPluginConfig holds SCIM plugin configuration
type SCIMPluginConfig struct {
	Enabled     bool          // Master switch for SCIM
	BaseURL     string        // Base URL for SCIM endpoints
	TokenExpiry time.Duration // Token expiry duration
}

// DefaultSCIMPluginConfig returns default SCIM plugin configuration
func DefaultSCIMPluginConfig() *SCIMPluginConfig {
	return &SCIMPluginConfig{
		Enabled:     getEnvBool("GBA_SCIM_ENABLED", false),
		BaseURL:     getEnvOrDefault("SCIM_BASE_URL", "https://app.functionfly.com/v1/scim"),
		TokenExpiry: 365 * 24 * time.Hour,
	}
}

// New creates a new SCIM plugin instance
func New(db *gorm.DB, config *SCIMPluginConfig, logger *logrus.Logger) (*SCIMPlugin, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if config == nil {
		config = DefaultSCIMPluginConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Create SCIM service
	service, err := NewSCIMService(db, logger, config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCIM service: %w", err)
	}

	plugin := &SCIMPlugin{
		db:      db,
		service: service,
		config:  config,
		logger:  logger,
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&SCIMConfig{}, &SCIMUser{}, &SCIMGroup{}); err != nil {
		return nil, fmt.Errorf("failed to migrate SCIM models: %w", err)
	}

	logger.Info("SCIM plugin initialized")
	return plugin, nil
}

// IsEnabled returns true if SCIM is enabled
func (p *SCIMPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// IsEnabledForTenant checks if SCIM is enabled for a specific tenant
func (p *SCIMPlugin) IsEnabledForTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	if !p.IsEnabled() {
		return false, nil
	}

	config, err := p.service.GetConfig(ctx, tenantID)
	if err != nil {
		return false, nil
	}

	return config.Enabled, nil
}

// GetStatus returns the SCIM status for a tenant
func (p *SCIMPlugin) GetStatus(ctx context.Context, tenantID uuid.UUID) (*SCIMStatusResponse, error) {
	if !p.IsEnabled() {
		return &SCIMStatusResponse{
			Enabled:    false,
			Configured: false,
		}, nil
	}

	_, err := p.service.GetConfig(ctx, tenantID)
	if err != nil {
		return &SCIMStatusResponse{
			Enabled:    true,
			Configured: false,
		}, nil
	}

	return &SCIMStatusResponse{
		Enabled:    true,
		Configured: true,
	}, nil
}

// GetService returns the SCIM service
func (p *SCIMPlugin) GetService() *SCIMService {
	return p.service
}

// GetHandler returns an HTTP handler for SCIM endpoints
func (p *SCIMPlugin) GetHandler() *Handler {
	return NewHandler(p, p.logger)
}

// SetupRoutes registers SCIM routes with the provided mux
func (p *SCIMPlugin) SetupRoutes(mux *http.ServeMux, basePath string) {
	handler := p.GetHandler()
	handler.SetupRoutes(mux, basePath)
}

// AuditLog logs a SCIM-related audit event
func (p *SCIMPlugin) AuditLog(ctx context.Context, userID, tenantID uuid.UUID, action, result string, metadata map[string]interface{}) {
	auditEntry := map[string]interface{}{
		"id":         uuid.Must(uuid.NewRandom()),
		"user_id":    userID,
		"tenant_id":  tenantID,
		"action":     "scim_" + action,
		"result":     result,
		"metadata":   metadata,
		"created_at": time.Now(),
	}

	dbResult := p.db.WithContext(ctx).Table("gba_audit_logs").Create(auditEntry)
	if dbResult.Error != nil {
		p.logger.WithError(dbResult.Error).Warn("Failed to create audit log entry")
	}

	p.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
		"action":    action,
		"result":    result,
	}).Info("SCIM audit event")
}

// Admin Handlers

// AdminHandler provides admin HTTP handlers for SCIM configuration
type AdminHandler struct {
	plugin *SCIMPlugin
	logger *logrus.Logger
}

// NewAdminHandler creates a new SCIM admin handler
func NewAdminHandler(plugin *SCIMPlugin, logger *logrus.Logger) *AdminHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &AdminHandler{
		plugin: plugin,
		logger: logger,
	}
}

// SetupAdminRoutes registers SCIM admin routes
func (h *AdminHandler) SetupAdminRoutes(mux *http.ServeMux, basePath string) {
	// Token management
	mux.HandleFunc("POST "+basePath+"/admin/tenants/{id}/scim/token", h.HandleGenerateToken)
	mux.HandleFunc("DELETE "+basePath+"/admin/tenants/{id}/scim/token", h.HandleRevokeToken)

	// Configuration management
	mux.HandleFunc("GET "+basePath+"/admin/tenants/{id}/scim/config", h.HandleGetConfig)
	mux.HandleFunc("PUT "+basePath+"/admin/tenants/{id}/scim/config", h.HandleUpdateConfig)

	h.logger.WithField("path", basePath).Info("SCIM admin routes registered")
}

// HandleGenerateToken handles POST /admin/tenants/{id}/scim/token
func (h *AdminHandler) HandleGenerateToken(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SCIM is not enabled")
		return
	}

	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	token, config, err := h.plugin.service.RegenerateToken(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate SCIM token")
		h.respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.respondJSON(w, http.StatusOK, SCIMTokenResponse{
		Token:     token,
		TokenHash: config.TokenHash[:16] + "...",
		CreatedAt: config.CreatedAt,
	})
}

// HandleRevokeToken handles DELETE /admin/tenants/{id}/scim/token
func (h *AdminHandler) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SCIM is not enabled")
		return
	}

	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "SCIM not configured for this tenant")
		return
	}

	// Disable SCIM by updating the config
	config.Enabled = false
	if err := h.plugin.db.WithContext(r.Context()).Save(config).Error; err != nil {
		h.logger.WithError(err).Error("Failed to disable SCIM")
		h.respondError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "token_revoked"})
}

// HandleGetConfig handles GET /admin/tenants/{id}/scim/config
func (h *AdminHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondJSON(w, http.StatusOK, SCIMStatusResponse{
			Enabled:    false,
			Configured: false,
		})
		return
	}

	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	config, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		h.respondJSON(w, http.StatusOK, SCIMStatusResponse{
			Enabled:    true,
			Configured: false,
		})
		return
	}

	h.respondJSON(w, http.StatusOK, SCIMConfigResponse{
		ID:         config.ID.String(),
		TenantID:   config.TenantID.String(),
		Enabled:    config.Enabled,
		SyncGroups: config.SyncGroups,
		SyncUsers:  config.SyncUsers,
		LastSyncAt: config.LastSyncAt,
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  config.UpdatedAt,
	})
}

// HandleUpdateConfig handles PUT /admin/tenants/{id}/scim/config
func (h *AdminHandler) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if !h.plugin.IsEnabled() {
		h.respondError(w, http.StatusServiceUnavailable, "SCIM is not enabled")
		return
	}

	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req SCIMConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if config already exists
	existingConfig, err := h.plugin.service.GetConfig(r.Context(), tenantID)
	if err != nil && err.Error() != "SCIM not configured for tenant" {
		h.respondError(w, http.StatusInternalServerError, "Failed to check existing configuration")
		return
	}

	var config *SCIMConfig
	if existingConfig != nil {
		// Update existing config
		config, err = h.plugin.service.UpdateConfig(r.Context(), existingConfig.ID, &req)
		if err != nil {
			h.logger.WithError(err).Error("Failed to update SCIM configuration")
			h.respondError(w, http.StatusInternalServerError, "Failed to update configuration")
			return
		}
	} else {
		// Create new config
		config, err = h.plugin.service.CreateConfig(r.Context(), tenantID, &req)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create SCIM configuration")
			h.respondError(w, http.StatusInternalServerError, "Failed to create configuration")
			return
		}
	}

	h.respondJSON(w, http.StatusOK, SCIMConfigResponse{
		ID:         config.ID.String(),
		TenantID:   config.TenantID.String(),
		Enabled:    config.Enabled,
		SyncGroups: config.SyncGroups,
		SyncUsers:  config.SyncUsers,
		LastSyncAt: config.LastSyncAt,
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  config.UpdatedAt,
	})
}

// respondJSON sends a JSON response
func (h *AdminHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *AdminHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the boolean value of an environment variable or a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
