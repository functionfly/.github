// Package agent provides HTTP handlers for the Agent Execution Plan (AEP) API.
// All routes are prefixed with /v1/agent/ and provide agent-native execution,
// discovery, quota management, policy enforcement, and economic controls.
package agent

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/agent/concurrency"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/policy"
	"github.com/functionfly/functionfly/internal/agent/quota"
	"github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler contains all AEP API handlers
type Handler struct {
	identityRepo    *identity.Repository
	quotaEnforcer   *quota.Enforcer
	policyEngine    *policy.Engine
	attributionRepo *attribution.Repository
	billingCtrl     *billing.Controller
	scheduler       *concurrency.PriorityScheduler
	registryRepo    *registry.RegistryRepository
	userRepo        storage.Repository
	financialTxRepo *storage.FinancialTransactionRepository
	notificationSvc *notification.Service
	toolRegistry    tools.Registry
	prefsRepo       *storage.AIModelPreferencesRepository
}

// NewHandler creates a new AEP handler
func NewHandler(
	db *gorm.DB,
	redisClient *redis.Client,
	registryRepo *registry.RegistryRepository,
	userRepo storage.Repository,
	notificationSvc *notification.Service,
) *Handler {
	sqlDB, _ := db.DB()
	return &Handler{
		identityRepo:    identity.NewRepository(db),
		quotaEnforcer:   quota.NewEnforcer(db, redisClient),
		policyEngine:    policy.NewEngine(db, redisClient),
		attributionRepo: attribution.NewRepository(db),
		billingCtrl:     billing.NewController(db, redisClient),
		scheduler:       concurrency.NewPriorityScheduler(),
		registryRepo:    registryRepo,
		userRepo:        userRepo,
		financialTxRepo: storage.NewFinancialTransactionRepository(db),
		notificationSvc: notificationSvc,
		toolRegistry:    tools.NewRegistry(),
		prefsRepo:       storage.NewAIModelPreferencesRepository(sqlDB),
	}
}

// SetToolRegistry sets the tool registry (allows injection of custom registry)
func (h *Handler) SetToolRegistry(r tools.Registry) {
	h.toolRegistry = r
}

// getAgentByIDOrUUID looks up an agent by agent_id string first, then falls back to UUID lookup.
// This matches the behavior of HandleGetAgent so all handlers resolve agents consistently.
func (h *Handler) getAgentByIDOrUUID(r *http.Request, agentID string) (*identity.AgentIdentity, error) {
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		if parsedUUID, parseErr := uuid.Parse(agentID); parseErr == nil {
			agent, err = h.identityRepo.GetAgentByUUID(r.Context(), parsedUUID)
		}
	}
	return agent, err
}

// ============================================================
// Agent Registration & Management
// ============================================================

// HandleRegisterAgent registers a new agent identity
// POST /v1/agent/register
func (h *Handler) HandleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req identity.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.AgentID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "agent_id and name are required")
		return
	}

	agent, apiKey, signingKey, err := h.identityRepo.CreateAgent(r.Context(), claims.TenantID, &req)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": claims.TenantID,
			"agent_id":  req.AgentID,
		}).Error("failed to register agent")
		errStr := err.Error()
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
			writeError(w, http.StatusConflict, "AGENT_ID_TAKEN", "This agent ID is already in use.")
			return
		}
		writeError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", "Failed to register agent. Check server logs for details.")
		return
	}

	writeJSON(w, http.StatusCreated, identity.RegisterAgentResponse{
		OK:        true,
		Agent:     agent,
		APIKey:    apiKey,
		SigningKey: signingKey,
	})
}

// HandleGetAgent retrieves an agent by ID
// GET /v1/agent/{agent_id}
func (h *Handler) HandleGetAgent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	// Ensure the agent belongs to the requesting tenant
	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"agent": agent,
	})
}

// HandleUpdateAgent updates an agent's mutable fields (name, description, capabilities, etc.)
// PUT /v1/agent/{agent_id}
func (h *Handler) HandleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	existing, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if existing.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req identity.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	updated, err := h.identityRepo.UpdateAgent(r.Context(), agentID, &req)
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Error("failed to update agent")
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update agent")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"agent": updated,
	})
}

// HandleListAgents lists all agents for the authenticated tenant
// GET /v1/agent
func (h *Handler) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	agents, total, err := h.identityRepo.ListAgents(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to list agents")
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list agents")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"agents": agents,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleDeleteAgent deregisters an agent
// DELETE /v1/agent/{agent_id}
func (h *Handler) HandleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	if err := h.identityRepo.UpdateAgentStatus(r.Context(), agentID, identity.AgentStatusDeleted); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete agent")
		return
	}

	// Remove concurrency pool
	h.scheduler.RemovePool(agentID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "agent deregistered",
	})
}

// ============================================================
// Quota Management
// ============================================================

// HandleUpdateQuota updates the quota config for an agent
// PUT /v1/agent/{agent_id}/quota
func (h *Handler) HandleUpdateQuota(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := h.identityRepo.UpdateQuotaConfig(r.Context(), agentID, updates); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update quota config")
		return
	}

	quota, _ := h.identityRepo.GetQuotaConfig(r.Context(), agentID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"quota": quota,
	})
}

// HandleGetUsage returns current usage counters for an agent
// GET /v1/agent/{agent_id}/usage
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	usage, err := h.quotaEnforcer.GetCurrentUsage(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "USAGE_FAILED", "failed to get usage")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"usage": usage,
	})
}

// ============================================================
// Policy Management
// ============================================================

// HandleUpdatePolicy updates the behavioral policy for an agent
// PUT /v1/agent/{agent_id}/policy
func (h *Handler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var p policy.BehavioralPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	p.AgentID = agentID

	if err := h.policyEngine.UpsertPolicy(r.Context(), &p); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update policy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"policy": p,
	})
}

// HandleGetPolicy retrieves the behavioral policy for an agent
// GET /v1/agent/{agent_id}/policy
func (h *Handler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	p, err := h.policyEngine.GetPolicy(r.Context(), agentID)
	if err != nil {
		// Return default policy if none configured
		p = &policy.BehavioralPolicy{
			AgentID:           agentID,
			MaxExecutionDepth: 10,
			MaxRecursionDepth: 3,
			MaxWallTimeMs:     300000,
			MaxMemoryGrowthMB: 512,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"policy": p,
	})
}

// ============================================================
// Attribution & Observability
// ============================================================

// HandleListExecutions lists execution records for an agent
// GET /v1/agent/{agent_id}/executions
func (h *Handler) HandleListExecutions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	records, total, err := h.attributionRepo.ListExecutions(r.Context(), agentID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list executions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"executions": records,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// HandleGetExecution retrieves a specific execution record
// GET /v1/agent/{agent_id}/executions/{exec_id}
func (h *Handler) HandleGetExecution(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	execID := mux.Vars(r)["exec_id"]

	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	record, err := h.attributionRepo.GetExecution(r.Context(), execID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "execution record not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"execution": record,
	})
}

// HandleGetAnalytics returns aggregated analytics for an agent
// GET /v1/agent/{agent_id}/analytics
func (h *Handler) HandleGetAnalytics(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	// Default: last 7 days
	since := time.Now().UTC().AddDate(0, 0, -7)
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}

	analytics, err := h.attributionRepo.GetAnalytics(r.Context(), agentID, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ANALYTICS_FAILED", "failed to get analytics")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"analytics": analytics,
	})
}

// ============================================================
// Session Management
// ============================================================

// HandleStartSession starts a new agent session
// POST /v1/agent/{agent_id}/session/start
func (h *Handler) HandleStartSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	session, err := h.attributionRepo.StartSession(r.Context(), agentID, claims.TenantID, sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_FAILED", "failed to start session")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":      true,
		"session": session,
	})
}

// HandleEndSession ends an agent session
// POST /v1/agent/{agent_id}/session/{session_id}/end
func (h *Handler) HandleEndSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	sessionID := mux.Vars(r)["session_id"]

	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	if err := h.attributionRepo.EndSession(r.Context(), sessionID, attribution.SessionStatusCompleted); err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_FAILED", "failed to end session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "session ended",
	})
}

// HandleGetSession retrieves session details
// GET /v1/agent/{agent_id}/session/{session_id}
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	sessionID := mux.Vars(r)["session_id"]

	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	session, err := h.attributionRepo.GetSession(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"session": session,
	})
}

// ============================================================
// Billing & Economic Controls
// ============================================================
// Helpers
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}

func writeErrorFromErr(r *http.Request, w http.ResponseWriter, status int, code, contextMsg string, err error) {
	if err != nil {
		fields := logrus.Fields{
			"status":  status,
			"code":    code,
			"context": contextMsg,
			"method":  "",
			"path":    "",
		}
		if r != nil {
			fields["method"] = r.Method
			if r.URL != nil {
				fields["path"] = r.URL.Path
			}
		}
		entry := logrus.WithError(err).WithFields(fields)
		if status >= 500 {
			entry.Error("agent handler error")
		} else {
			entry.Info("agent handler client error")
		}
	}
	message := sanitizedAgentErrorMessage(status, code, contextMsg)
	writeError(w, status, code, message)
}

func sanitizedAgentErrorMessage(status int, code, contextMsg string) string {
	if status >= 500 {
		switch code {
		case "INTERNAL_ERROR":
			return "Internal server error"
		case "REGISTRATION_FAILED":
			return "Failed to register agent. Check server logs for details."
		case "SESSION_FAILED":
			return "Failed to start session. Check server logs for details."
		case "EXECUTION_FAILED":
			return "Failed to execute agent. Check server logs for details."
		case "STATS_FAILED":
			return "Failed to retrieve statistics."
		case "DELETE_FAILED":
			return "Failed to delete agent."
		case "UPDATE_FAILED":
			return "Failed to update agent."
		case "LIST_FAILED":
			return "Failed to list agents."
		}
		return "Internal server error"
	}
	switch code {
	case "INVALID_REQUEST":
		return "Invalid request body"
	case "INVALID_PARAMETERS":
		return "Invalid parameters"
	}
	return contextMsg
}

func generateSessionID() string {
	return "sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// HandleListModels returns the curated list of AI models available for agents.
// GET /v1/ai/models
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	models := []map[string]string{
		// OpenRouter — frontier
		{"id": "anthropic/claude-opus-4.7", "name": "Claude Opus 4.7", "provider": "openrouter", "tier": "frontier", "cost": "$$$"},
		{"id": "anthropic/claude-sonnet-4.6", "name": "Claude Sonnet 4.6", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "openai/gpt-5.5", "name": "GPT-5.5", "provider": "openrouter", "tier": "frontier", "cost": "$$$"},
		{"id": "openai/gpt-5.4", "name": "GPT-5.4", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "google/gemini-3.1-pro", "name": "Gemini 3.1 Pro", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "deepseek/deepseek-v4-pro", "name": "DeepSeek V4 Pro", "provider": "openrouter", "tier": "frontier", "cost": "$"},
		{"id": "z-ai/glm-5.2", "name": "GLM 5.2", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "x-ai/grok-4", "name": "Grok 4", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "qwen/qwen3.7-plus", "name": "Qwen 3.7 Plus", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		{"id": "nex-agi/nex-n2-pro", "name": "Nex N2 Pro", "provider": "openrouter", "tier": "frontier", "cost": "$$"},
		// OpenRouter — fast
		{"id": "anthropic/claude-haiku-4", "name": "Claude Haiku 4", "provider": "openrouter", "tier": "fast", "cost": "$"},
		{"id": "openai/gpt-5-mini", "name": "GPT-5 Mini", "provider": "openrouter", "tier": "fast", "cost": "$"},
		{"id": "google/gemini-2.5-flash", "name": "Gemini 2.5 Flash", "provider": "openrouter", "tier": "fast", "cost": "$"},
		{"id": "deepseek/deepseek-v4-flash", "name": "DeepSeek V4 Flash", "provider": "openrouter", "tier": "fast", "cost": "$"},
		{"id": "meta-llama/llama-3.3-70b-instruct", "name": "Llama 3.3 70B", "provider": "openrouter", "tier": "fast", "cost": "$"},
		{"id": "mistralai/mistral-large", "name": "Mistral Large", "provider": "openrouter", "tier": "fast", "cost": "$$"},
		// OpenRouter — reasoning
		{"id": "openai/o3", "name": "OpenAI o3", "provider": "openrouter", "tier": "reasoning", "cost": "$$$"},
		{"id": "deepseek/deepseek-r1", "name": "DeepSeek R1", "provider": "openrouter", "tier": "reasoning", "cost": "$"},
		{"id": "openrouter/free", "name": "Free Models Router", "provider": "openrouter", "tier": "fast", "cost": "free"},
		{"id": "openai/gpt-oss-120b", "name": "GPT-OSS 120B", "provider": "openrouter", "tier": "reasoning", "cost": "free"},
		{"id": "nvidia/nemotron-3-super-120b-a12b:free", "name": "Nemotron 3 Super 120B", "provider": "openrouter", "tier": "reasoning", "cost": "free"},
		{"id": "nvidia/nemotron-3-ultra-550b-a55b:free", "name": "Nemotron 3 Ultra 550B", "provider": "openrouter", "tier": "reasoning", "cost": "free"},
		// OpenRouter — code
		{"id": "qwen/qwen3-coder", "name": "Qwen3 Coder", "provider": "openrouter", "tier": "code", "cost": "$$"},
		{"id": "openai/gpt-5-codex", "name": "GPT-5 Codex", "provider": "openrouter", "tier": "code", "cost": "$$$"},
		{"id": "cohere/north-mini-code:free", "name": "North Mini Code", "provider": "openrouter", "tier": "code", "cost": "free"},
		{"id": "moonshotai/kimi-k2.7-code", "name": "Kimi K2.7 Code", "provider": "openrouter", "tier": "code", "cost": "$$"},
		// Direct — OpenAI
		{"id": "gpt-4o", "name": "GPT-4o", "provider": "openai", "tier": "fast", "cost": "$$"},
		{"id": "gpt-4o-mini", "name": "GPT-4o Mini", "provider": "openai", "tier": "fast", "cost": "$"},
		// Direct — Anthropic
		{"id": "claude-sonnet-4-6", "name": "Claude Sonnet 4.6", "provider": "anthropic", "tier": "frontier", "cost": "$$"},
		{"id": "claude-3-5-sonnet-20241022", "name": "Claude 3.5 Sonnet", "provider": "anthropic", "tier": "fast", "cost": "$$"},
		{"id": "claude-3-5-haiku-20241022", "name": "Claude 3.5 Haiku", "provider": "anthropic", "tier": "fast", "cost": "$"},
		// Groq — low latency
		{"id": "llama-3.3-70b-versatile", "name": "Llama 3.3 70B (Groq)", "provider": "groq", "tier": "fast", "cost": "$"},
		{"id": "llama-3.1-8b-instant", "name": "Llama 3.1 8B Instant", "provider": "groq", "tier": "fast", "cost": "$"},
	}

	if h.prefsRepo != nil {
		if claims := middleware.GetUserFromContext(r); claims != nil {
			prefs, err := h.prefsRepo.GetTenantAIPreferences(r.Context(), claims.TenantID)
			if err == nil {
				if len(prefs.EnabledProviders) > 0 {
					allowProv := make(map[string]struct{}, len(prefs.EnabledProviders))
					for _, p := range prefs.EnabledProviders {
						allowProv[p] = struct{}{}
					}
					filtered := make([]map[string]string, 0, len(models))
					for _, m := range models {
						if _, ok := allowProv[m["provider"]]; ok {
							filtered = append(filtered, m)
						}
					}
					models = filtered
				}
				if len(prefs.EnabledModels) > 0 {
					allow := make(map[string]struct{}, len(prefs.EnabledModels))
					for _, item := range prefs.EnabledModels {
						allow[item.Provider+":"+item.ModelID] = struct{}{}
					}
					filtered := make([]map[string]string, 0, len(models))
					for _, m := range models {
						if _, ok := allow[m["provider"]+":"+m["id"]]; ok {
							filtered = append(filtered, m)
						}
					}
					models = filtered
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"models": models,
	})
}
