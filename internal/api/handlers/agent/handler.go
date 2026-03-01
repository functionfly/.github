// Package agent provides HTTP handlers for the Agent Execution Plan (AEP) API.
// All routes are prefixed with /v1/agent/ and provide agent-native execution,
// discovery, quota management, policy enforcement, and economic controls.
package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/agent/concurrency"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/policy"
	"github.com/functionfly/functionfly/internal/agent/quota"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage/registry"
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
}

// NewHandler creates a new AEP handler
func NewHandler(
	db *gorm.DB,
	redisClient *redis.Client,
	registryRepo *registry.RegistryRepository,
) *Handler {
	return &Handler{
		identityRepo:    identity.NewRepository(db),
		quotaEnforcer:   quota.NewEnforcer(db, redisClient),
		policyEngine:    policy.NewEngine(db, redisClient),
		attributionRepo: attribution.NewRepository(db),
		billingCtrl:     billing.NewController(db, redisClient),
		scheduler:       concurrency.NewPriorityScheduler(),
		registryRepo:    registryRepo,
	}
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

	agent, apiKey, err := h.identityRepo.CreateAgent(r.Context(), claims.TenantID, &req)
	if err != nil {
		logrus.WithError(err).Error("failed to register agent")
		writeError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, identity.RegisterAgentResponse{
		OK:     true,
		Agent:  agent,
		APIKey: apiKey,
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
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

// HandleGetBillingSummary returns the billing summary for an agent
// GET /v1/agent/{agent_id}/billing/summary
func (h *Handler) HandleGetBillingSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	period := r.URL.Query().Get("period")
	summary, err := h.billingCtrl.GetAgentSpend(r.Context(), agentID, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to get billing summary")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"summary": summary,
	})
}

// HandleGetCreditBalance returns the credit balance for an agent
// GET /v1/agent/{agent_id}/credits/balance
func (h *Handler) HandleGetCreditBalance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	controls, err := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to get credit balance")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                 true,
		"agent_id":           agentID,
		"credit_balance_usd": controls.CreditBalanceUSD,
	})
}

// HandleUpdateSpendCap updates the spend cap for an agent
// PUT /v1/agent/{agent_id}/billing/spend-cap
func (h *Handler) HandleUpdateSpendCap(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		SpendCapDailyUSD   *float64 `json:"spend_cap_daily_usd"`
		SpendCapMonthlyUSD *float64 `json:"spend_cap_monthly_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if err := h.billingCtrl.UpdateSpendCap(r.Context(), agentID, req.SpendCapDailyUSD, req.SpendCapMonthlyUSD); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update spend cap")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "spend cap updated",
	})
}

// HandleGetCostBreakdown returns cost breakdown by function for an agent
// GET /v1/agent/{agent_id}/cost-breakdown
func (h *Handler) HandleGetCostBreakdown(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	// Get cost breakdown from attribution repository
	breakdown, err := h.attributionRepo.GetCostBreakdown(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "COST_BREAKDOWN_FAILED", "failed to get cost breakdown")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"agent_id":  agentID,
		"breakdown": breakdown,
	})
}

// HandlePurchaseCredits purchases execution credits for an agent
// POST /v1/agent/{agent_id}/credits/purchase
func (h *Handler) HandlePurchaseCredits(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		AmountUSD       float64 `json:"amount_usd"`
		PaymentMethodID string  `json:"payment_method_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.AmountUSD <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive")
		return
	}

	if payment.IsConfigured() {
		if req.PaymentMethodID == "" {
			writeError(w, http.StatusBadRequest, "PAYMENT_METHOD_REQUIRED", "payment_method_id is required when Stripe is configured")
			return
		}
		_, err := payment.Charge(r.Context(), req.PaymentMethodID, req.AmountUSD, map[string]string{
			"agent_id":  agentID,
			"tenant_id": agent.TenantID.String(),
		})
		if err != nil {
			logrus.WithError(err).Warn("credit purchase payment failed")
			writeError(w, http.StatusPaymentRequired, "PAYMENT_FAILED", err.Error())
			return
		}
	} else if req.PaymentMethodID != "" {
		writeError(w, http.StatusBadRequest, "PAYMENTS_NOT_CONFIGURED", "Stripe is not configured; omit payment_method_id for simulated credit purchase")
		return
	}

	if err := h.billingCtrl.AddCredits(r.Context(), agentID, req.AmountUSD); err != nil {
		writeError(w, http.StatusInternalServerError, "PURCHASE_FAILED", "failed to add credits")
		return
	}

	controls, _ := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"message":           "credits purchased successfully",
		"agent_id":          agentID,
		"credits_added_usd": req.AmountUSD,
		"new_balance_usd":   controls.CreditBalanceUSD,
	})
}

// ============================================================
// Concurrency Stats
// ============================================================

// HandleGetConcurrencyStats returns concurrency pool statistics
// GET /v1/agent/concurrency/stats
func (h *Handler) HandleGetConcurrencyStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	stats := h.scheduler.GetAllStats()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                      true,
		"pools":                   stats,
		"total_active_executions": h.scheduler.TotalActiveExecutions(),
	})
}

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

func generateSessionID() string {
	return "sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
