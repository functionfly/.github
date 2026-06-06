package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ============================================================
// Session Management
// ============================================================

var sessionLogger = logrus.WithField("component", "agent_sessions")

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
		sessionLogger.WithError(err).WithField("agent_id", agentID).Warn("start session: agent not found")
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

	// Check payment status for the agent's tenant
	billingStatus, err := payment.GetTenantPaymentStatus(r.Context(), h.userRepo, claims.TenantID)
	if err != nil {
		sessionLogger.WithError(err).WithField("tenant_id", claims.TenantID).Warn("start session: could not check billing status, allowing through")
	} else if billingStatus.PaymentMode == "suspended" {
		sessionLogger.WithFields(logrus.Fields{
			"agent_id":  agentID,
			"tenant_id": claims.TenantID,
		}).Warn("start session denied: tenant billing suspended")
		writeError(w, http.StatusPaymentRequired, "BILLING_SUSPENDED", "billing is suspended for this tenant")
		return
	}

	session, err := h.attributionRepo.StartSession(r.Context(), agentID, claims.TenantID, sessionID)
	if err != nil {
		sessionLogger.WithError(err).WithField("agent_id", agentID).Error("failed to start session")
		writeError(w, http.StatusInternalServerError, "SESSION_FAILED", "failed to start session")
		return
	}

	sessionLogger.WithFields(logrus.Fields{
		"agent_id":   agentID,
		"session_id": sessionID,
		"tenant_id":  claims.TenantID,
		"user_id":    claims.UserID,
	}).Info("agent session started")

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

	if err := h.attributionRepo.EndSession(r.Context(), sessionID, "completed"); err != nil {
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

func generateSessionID() string {
	return "sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}