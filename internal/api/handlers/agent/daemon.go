package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	daemonStartSubject  = "orchestrator.agent.daemon.start"
	daemonStopSubject   = "orchestrator.agent.daemon.stop"
	daemonResponseSubj  = "orchestrator.agent.daemon.response"
	daemonTimeout       = 10 * time.Second
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

// DaemonCommand represents a command to send to the SAR runtime
type DaemonCommand struct {
	AgentID      string                 `json:"agent_id"`
	TenantID     string                 `json:"tenant_id"`
	Command      string                 `json:"command"` // "start" | "stop" | "status"
	DaemonConfig map[string]interface{}  `json:"daemon_config,omitempty"`
}

// DaemonResponse represents the response from the SAR runtime
type DaemonResponse struct {
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
	CellID string `json:"cell_id,omitempty"`
}

// sendDaemonCommand sends a command to the SAR runtime via NATS and waits for response
func (h *DaemonHandler) sendDaemonCommand(ctx context.Context, cmd *DaemonCommand) (*DaemonResponse, error) {
	if !h.natsEnabled || h.natsConn == nil {
		logrus.WithField("command", cmd.Command).Debug("NATS not enabled, skipping runtime call")
		return &DaemonResponse{OK: true}, nil
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}

	// Determine subject based on command
	subject := daemonStartSubject
	if cmd.Command == "stop" {
		subject = daemonStopSubject
	}

	// Create unique inbox for response
	inbox := h.natsConn.NewRespInbox()

	// Subscribe to response
	sub, err := h.natsConn.SubscribeSync(inbox)
	if err != nil {
		return nil, fmt.Errorf("subscribe to response: %w", err)
	}
	defer sub.Unsubscribe()

	// Publish command
	if err := h.natsConn.PublishRequest(subject, inbox, cmdBytes); err != nil {
		return nil, fmt.Errorf("publish command: %w", err)
	}

	// Wait for response with timeout
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for response: %w", err)
	}

	var resp DaemonResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// DaemonHandler handles always-on daemon control endpoints
type DaemonHandler struct {
	db          *gorm.DB
	natsConn    *nats.Conn
	natsEnabled bool
}

// NewDaemonHandler creates a new daemon handler
func NewDaemonHandler(db *gorm.DB) *DaemonHandler {
	return &DaemonHandler{db: db, natsEnabled: false}
}

// SetNATSConnection sets the NATS connection for runtime communication
func (h *DaemonHandler) SetNATSConnection(nc *nats.Conn) {
	h.natsConn = nc
	h.natsEnabled = nc != nil
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
		// Convert JSONBMap to JSON bytes then unmarshal to struct
		configBytes, err := json.Marshal(agent.DaemonConfig)
		if err != nil {
			logrus.WithError(err).Warn("Failed to marshal daemon config")
			config = DaemonConfig{IsEnabled: false, Mode: "on_demand"}
		} else if err := json.Unmarshal(configBytes, &config); err != nil {
			logrus.WithError(err).Warn("Failed to parse daemon config")
			config = DaemonConfig{IsEnabled: false, Mode: "on_demand"}
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
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode daemon request", err)
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

	// Get tenant plan from auth context
	tier := middleware.GetTenantPlan(r)
	if tier == "" {
		tier = plans.PlanStarter
		logrus.WithField("tenant_id", agent.TenantID).Warn("tenant plan not in context, defaulting to starter")
	}

	// Get max always-on agents for this plan
	maxAgents, _, _, _ := plans.GetAgentLimitsForPlan(tier)

	// Count currently running always-on agents for this tenant
	var runningCount int64
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("tenant_id = ? AND is_daemon_running = ?", agent.TenantID, true).
		Count(&runningCount)

	if int(runningCount) >= maxAgents {
		writeError(w, http.StatusPaymentRequired, "LIMIT_REACHED",
			fmt.Sprintf("Your plan (%s) allows %d always-on agent(s). Upgrade to enable more.", tier, maxAgents))
		return
	}

	// Call the SAR runtime via NATS to start the daemon
	ctx, cancel := context.WithTimeout(r.Context(), daemonTimeout)
	defer cancel()

	now := time.Now().UTC()
	startResp, err := h.sendDaemonCommand(ctx, &DaemonCommand{
		AgentID:      agentID,
		TenantID:     agent.TenantID.String(),
		Command:      "start",
		DaemonConfig: agent.DaemonConfig,
	})
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Warn("Failed to start daemon via runtime, updating DB only")
		// Fallback: update DB only (daemon may start later via heartbeat registration)
	}

	if startResp != nil && !startResp.OK {
		writeError(w, http.StatusInternalServerError, "DAEMON_START_FAILED", startResp.Error)
		return
	}

	// Update daemon status in database
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"daemon_started_at": now,
			"is_daemon_running": true,
		})

	logrus.WithField("agent_id", agentID).Info("Daemon started via SAR runtime")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"message":    "Daemon started successfully",
		"started_at": now,
		"runtime":    "sar",
	})
}

// StopDaemon handles POST /api/agents/{id}/daemon/stop
func (h *DaemonHandler) StopDaemon(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Get agent to find tenant
	var agent identity.AgentIdentity
	if err := h.db.WithContext(r.Context()).
		Where("agent_id = ?", agentID).
		First(&agent).Error; err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	// Call the SAR runtime via NATS to stop the daemon
	ctx, cancel := context.WithTimeout(r.Context(), daemonTimeout)
	defer cancel()

	stopResp, err := h.sendDaemonCommand(ctx, &DaemonCommand{
		AgentID:  agentID,
		TenantID: agent.TenantID.String(),
		Command:  "stop",
	})
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Warn("Failed to stop daemon via runtime, updating DB only")
	}

	if stopResp != nil && !stopResp.OK {
		writeError(w, http.StatusInternalServerError, "DAEMON_STOP_FAILED", stopResp.Error)
		return
	}

	// Update daemon status in database
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"is_daemon_running": false,
		})

	logrus.WithField("agent_id", agentID).Info("Daemon stopped via SAR runtime")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Daemon stopped successfully",
		"runtime": "sar",
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

	tier := middleware.GetTenantPlan(r)
	if tier == "" {
		tier = plans.PlanStarter
	}

	maxAgents, _, _, _ := plans.GetAgentLimitsForPlan(tier)

	// Count always-on agents for this tenant
	var used int64
	h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("tenant_id = ? AND is_daemon_running = ?", agent.TenantID, true).
		Count(&used)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"tier":      tier,
		"allowance": maxAgents,
		"used":      int(used),
		"remaining": maxAgents - int(used),
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
