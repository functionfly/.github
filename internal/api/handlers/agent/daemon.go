package agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DaemonConfig represents the daemon configuration for an agent
type DaemonConfig struct {
	IsEnabled        bool                    `json:"is_enabled"`
	Mode             string                  `json:"mode"` // "on_demand" | "always_on"
	EventSources     []EventSourceConfig     `json:"event_sources"`
	IdleTimeout      int                     `json:"idle_timeout_minutes"`
	MaxExecutionsDay int                     `json:"max_executions_per_day"`
	WebhookEndpoints []WebhookEndpointConfig `json:"webhook_endpoints,omitempty"`
	Schedules        []ScheduleConfig        `json:"schedules,omitempty"`
	DatabaseTriggers []DatabaseTriggerConfig `json:"database_triggers,omitempty"`
}

// EventSourceConfig represents an event source configuration
type EventSourceConfig struct {
	Type   string                 `json:"type"` // webhook | database | scheduled
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// WebhookEndpointConfig represents a webhook endpoint
type WebhookEndpointConfig struct {
	Service      string `json:"service"` // stripe | shopify | resend | custom
	Path         string `json:"path"`
	Method       string `json:"method"`
	SecretHeader string `json:"secret_header,omitempty"`
}

// ScheduleConfig represents a scheduled trigger
type ScheduleConfig struct {
	ID       string `json:"id"`
	Cron     string `json:"cron"`
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
}

// DatabaseTriggerConfig represents a database trigger
type DatabaseTriggerConfig struct {
	Table      string   `json:"table"`
	Operations []string `json:"operations"` // INSERT | UPDATE | DELETE
	Enabled    bool     `json:"enabled"`
}

// DaemonStatus represents the current daemon status
type DaemonStatus struct {
	IsRunning           bool       `json:"is_running"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	LastActivity        *time.Time `json:"last_activity,omitempty"`
	ExecutionCount      int64      `json:"execution_count"`
	ExecutionCountToday int64      `json:"execution_count_today"`
	IsIdle              bool       `json:"is_idle"`
	IdleMinutes         int        `json:"idle_minutes"`
	EventSourcesActive  []string   `json:"event_sources_active"`
}

// DaemonHandler handles always-on daemon control endpoints
type DaemonHandler struct {
	db *gorm.DB
}

// NewDaemonHandler creates a new daemon handler
func NewDaemonHandler(db *gorm.DB) *DaemonHandler {
	return &DaemonHandler{db: db}
}

// requireAgentTenant verifies the request is authenticated and the agent belongs to the caller's tenant
func (h *DaemonHandler) requireAgentTenant(w http.ResponseWriter, r *http.Request, agentID string) bool {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return false
	}
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return false
	}
	agent, err := identity.NewRepository(h.db).GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return false
	}
	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return false
	}
	return true
}

// GetDaemonConfig handles GET /api/agents/{id}/daemon/config
func (h *DaemonHandler) GetDaemonConfig(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Get agent with daemon config
	var agent identity.AgentIdentity
	if err := h.db.WithContext(r.Context()).
		Select("agent_id", "daemon_config").
		Where("agent_id = ?", agentID).
		First(&agent).Error; err != nil {
		// Return default config if not set
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
			"config": DaemonConfig{
				IsEnabled:        false,
				Mode:             "on_demand",
				EventSources:     []EventSourceConfig{},
				IdleTimeout:      15,
				MaxExecutionsDay: 100,
			},
		})
		return
	}

	var config DaemonConfig
	if agent.DaemonConfig != nil {
		if err := json.Unmarshal(agent.DaemonConfig, &config); err != nil {
			logrus.WithError(err).Warn("Failed to parse daemon config")
			config = DaemonConfig{
				IsEnabled: false,
				Mode:      "on_demand",
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"config": config,
	})
}

// UpdateDaemonConfig handles PUT /api/agents/{id}/daemon/config
func (h *DaemonHandler) UpdateDaemonConfig(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var config DaemonConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Validate config
	if config.Mode != "on_demand" && config.Mode != "always_on" {
		writeError(w, http.StatusBadRequest, "INVALID_MODE", "mode must be 'on_demand' or 'always_on'")
		return
	}

	// Store config
	configJSON, err := json.Marshal(config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to serialize config")
		return
	}

	result := h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Update("daemon_config", datatypes.JSON(configJSON))

	if result.Error != nil {
		logrus.WithError(result.Error).Error("Failed to update daemon config")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save config")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"config": config,
	})
}

// StartDaemon handles POST /api/agents/{id}/daemon/start
func (h *DaemonHandler) StartDaemon(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Get agent config
	var agent identity.AgentIdentity
	if err := h.db.WithContext(r.Context()).
		Where("agent_id = ?", agentID).
		First(&agent).Error; err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	// Check always-on allowance (billing check)
	// In production, this would check the tenant's plan limits
	tier := getTenantTier(agent.TenantID.String())
	if tier == "free" && agent.AlwaysOnCount >= 1 {
		writeError(w, http.StatusPaymentRequired, "LIMIT_REACHED",
			"Free tier allows only 1 always-on agent. Upgrade to enable more.")
		return
	}

	// In production, this would call the Rust runtime to start the daemon
	// For now, we update the status
	now := time.Now().UTC()
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"daemon_started_at": now,
			"is_daemon_running": true,
		})

	logrus.WithField("agent_id", agentID).Info("Daemon started")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"message":    "Daemon started successfully",
		"started_at": now,
	})
}

// StopDaemon handles POST /api/agents/{id}/daemon/stop
func (h *DaemonHandler) StopDaemon(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// In production, this would call the Rust runtime to stop the daemon
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"is_daemon_running": false,
		})

	logrus.WithField("agent_id", agentID).Info("Daemon stopped")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Daemon stopped successfully",
	})
}

// GetDaemonStatus handles GET /api/agents/{id}/daemon/status
func (h *DaemonHandler) GetDaemonStatus(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Get agent with daemon status
	var agent identity.AgentIdentity
	if err := h.db.WithContext(r.Context()).
		Where("agent_id = ?", agentID).
		First(&agent).Error; err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	// Build status response
	status := DaemonStatus{
		IsRunning:           agent.IsDaemonRunning,
		StartedAt:           agent.DaemonStartedAt,
		ExecutionCount:      agent.DaemonExecutionCount,
		ExecutionCountToday: 0, // Would be calculated from execution records
		IsIdle:              false,
		IdleMinutes:         0,
		EventSourcesActive:  []string{},
	}

	if status.IsRunning && status.StartedAt != nil {
		status.LastActivity = status.StartedAt
	}

	// Calculate idle time
	if status.IsRunning && status.LastActivity != nil {
		idleMinutes := int(time.Since(*status.LastActivity).Minutes())
		status.IdleMinutes = idleMinutes
		status.IsIdle = idleMinutes > 15 // Default idle threshold
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"status": status,
	})
}

// GetAlwaysOnAllowance handles GET /api/agents/{id}/daemon/allowance
// Returns the always-on agent allowance for the tenant
func (h *DaemonHandler) GetAlwaysOnAllowance(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var agent identity.AgentIdentity
	if err := h.db.WithContext(r.Context()).
		Where("agent_id = ?", agentID).
		First(&agent).Error; err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	tier := getTenantTier(agent.TenantID.String())

	var allowance int
	switch tier {
	case "free":
		allowance = 1
	case "builder":
		allowance = 3
	case "pro":
		allowance = 10
	default:
		allowance = 1
	}

	// Count always-on agents for this tenant
	var used int64
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("tenant_id = ? AND is_daemon_running = ?", agent.TenantID, true).
		Count(&used)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"tier":      tier,
		"allowance": allowance,
		"used":      int(used),
		"remaining": allowance - int(used),
	})
}

// RegisterDaemonRoutes registers the daemon control routes
func (h *DaemonHandler) RegisterDaemonRoutes(router *mux.Router, basePath string, authMiddleware *middleware.AuthMiddleware) {
	auth := authMiddleware.RequireAuth
	agent := router.PathPrefix(basePath + "/agents").Subrouter()

	agent.HandleFunc("/{id}/daemon/config", auth(h.GetDaemonConfig)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/daemon/config", auth(h.UpdateDaemonConfig)).Methods(http.MethodPut)
	agent.HandleFunc("/{id}/daemon/start", auth(h.StartDaemon)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/daemon/stop", auth(h.StopDaemon)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/daemon/status", auth(h.GetDaemonStatus)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/daemon/allowance", auth(h.GetAlwaysOnAllowance)).Methods(http.MethodGet)
}

// getTenantTier returns the tier for a tenant
// In production, this would query the billing/subscription system
func getTenantTier(tenantID string) string {
	// For now, return based on a simple lookup
	// In production, this would check the subscription table
	if tenantID == "free" {
		return "free"
	}
	return "pro" // Default to pro for development
}
