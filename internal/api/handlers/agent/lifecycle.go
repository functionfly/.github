package agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ============================================================
// Lifecycle Management
// ============================================================

func (h *Handler) HandleAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	agentIDFromPath := mux.Vars(r)["agent_id"]
	if agentIDFromPath == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id required")
		return
	}

	agentID, _, err := h.authenticateAgent(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	if agentID != agentIDFromPath {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "agent_id mismatch")
		return
	}

	var req struct {
		Status        string         `json:"status"`
		StateSnapshot map[string]any `json:"state_snapshot,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	logrus.WithFields(logrus.Fields{
		"agent_id": agentID,
		"status":   req.Status,
	}).Debug("Agent heartbeat received")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"next_heartbeat": time.Now().Add(30 * time.Second).Format(time.RFC3339),
	})
}

func (h *Handler) HandleAgentShutdown(w http.ResponseWriter, r *http.Request) {
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
		GracePeriodSeconds int `json:"grace_period_seconds,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	gracePeriod := 30
	if req.GracePeriodSeconds > 0 {
		gracePeriod = req.GracePeriodSeconds
	}

	const maxGracePeriodSeconds = 3600
	if gracePeriod > maxGracePeriodSeconds {
		gracePeriod = maxGracePeriodSeconds
	}

	logrus.WithFields(logrus.Fields{
		"agent_id":             agentID,
		"grace_period_seconds": gracePeriod,
	}).Info("Agent graceful shutdown initiated")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                  true,
		"message":              "graceful shutdown initiated",
		"grace_period_seconds": gracePeriod,
	})
}

func (h *Handler) HandleAgentLifecycleStatus(w http.ResponseWriter, r *http.Request) {
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

	lifecycleStatus := "active"
	if agent.Status == "suspended" {
		lifecycleStatus = "suspended"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"agent_id":       agentID,
		"status":         lifecycleStatus,
		"last_heartbeat": agent.UpdatedAt.Format(time.RFC3339),
	})
}

// HandleAgentPause puts an agent into a paused/suspended state
// PUT /v1/agent/{agent_id}/lifecycle/pause
func (h *Handler) HandleAgentPause(w http.ResponseWriter, r *http.Request) {
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

	if agent.Status == "suspended" {
		writeError(w, http.StatusConflict, "ALREADY_PAUSED", "agent is already paused")
		return
	}

	if err := h.identityRepo.UpdateAgentStatus(r.Context(), agentID, "suspended"); err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Error("failed to pause agent")
		writeError(w, http.StatusInternalServerError, "PAUSE_FAILED", "failed to pause agent")
		return
	}

	logrus.WithFields(logrus.Fields{
		"agent_id": agentID,
		"tenant":   claims.TenantID,
	}).Info("Agent paused")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"agent_id": agentID,
		"status":   "suspended",
		"message":  "agent paused successfully",
	})
}

// HandleAgentResume resumes a paused/suspended agent back to active state
// PUT /v1/agent/{agent_id}/lifecycle/resume
func (h *Handler) HandleAgentResume(w http.ResponseWriter, r *http.Request) {
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

	if agent.Status != "suspended" {
		writeError(w, http.StatusConflict, "NOT_PAUSED", "agent is not paused")
		return
	}

	if err := h.identityRepo.UpdateAgentStatus(r.Context(), agentID, "active"); err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Error("failed to resume agent")
		writeError(w, http.StatusInternalServerError, "RESUME_FAILED", "failed to resume agent")
		return
	}

	logrus.WithFields(logrus.Fields{
		"agent_id": agentID,
		"tenant":   claims.TenantID,
	}).Info("Agent resumed")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"agent_id": agentID,
		"status":   "active",
		"message":  "agent resumed successfully",
	})
}

// HandleAgentTerminate initiates graceful termination of an agent
// POST /v1/agent/{agent_id}/lifecycle/terminate
func (h *Handler) HandleAgentTerminate(w http.ResponseWriter, r *http.Request) {
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
		GracePeriodSeconds int    `json:"grace_period_seconds"`
		Reason             string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	gracePeriod := 30
	if req.GracePeriodSeconds > 0 {
		gracePeriod = req.GracePeriodSeconds
	}
	const maxGracePeriodSeconds = 3600
	if gracePeriod > maxGracePeriodSeconds {
		gracePeriod = maxGracePeriodSeconds
	}

	if err := h.identityRepo.UpdateAgentStatus(r.Context(), agentID, "terminating"); err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Error("failed to initiate agent termination")
		writeError(w, http.StatusInternalServerError, "TERMINATE_FAILED", "failed to initiate agent termination")
		return
	}

	logrus.WithFields(logrus.Fields{
		"agent_id":             agentID,
		"tenant":               claims.TenantID,
		"grace_period_seconds": gracePeriod,
		"reason":               req.Reason,
	}).Info("Agent graceful termination initiated")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                  true,
		"agent_id":            agentID,
		"status":              "terminating",
		"grace_period_seconds": gracePeriod,
		"message":             "agent termination initiated",
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
