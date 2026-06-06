package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/policy"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

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

	var req struct {
		MaxCallsPerMinute  *int     `json:"max_calls_per_minute"`
		MaxCallsPerDay     *int     `json:"max_calls_per_day"`
		MaxDailySpendUSD   *float64 `json:"max_daily_spend_usd"`
		AllowedFunctions   *[]string `json:"allowed_functions"`
		ForbiddenFunctions *[]string `json:"forbidden_functions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	payload := make(map[string]interface{})
	if req.MaxCallsPerMinute != nil {
		payload["max_calls_per_minute"] = *req.MaxCallsPerMinute
	}
	if req.MaxCallsPerDay != nil {
		payload["max_calls_per_day"] = *req.MaxCallsPerDay
	}
	if req.MaxDailySpendUSD != nil {
		payload["max_daily_spend_usd"] = *req.MaxDailySpendUSD
	}
	if req.AllowedFunctions != nil {
		payload["allowed_functions"] = *req.AllowedFunctions
	}
	if req.ForbiddenFunctions != nil {
		payload["forbidden_functions"] = *req.ForbiddenFunctions
	}

	if err := h.identityRepo.UpdateQuotaConfig(r.Context(), agentID, payload); err != nil {
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