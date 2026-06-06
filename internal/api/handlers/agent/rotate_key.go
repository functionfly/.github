package agent

import (
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleRotateAPIKey rotates the API key for an agent, returning the new plaintext key once.
// POST /v1/agent/{agent_id}/keys/rotate
func (h *Handler) HandleRotateAPIKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id required")
		return
	}

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	plaintextKey, err := h.identityRepo.RotateAPIKey(r.Context(), agentID)
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Error("failed to rotate agent API key")
		writeError(w, http.StatusInternalServerError, "ROTATION_FAILED", "failed to rotate API key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"api_key":    plaintextKey,
		"agent_id":   agentID,
		"rotated_at": time.Now().Format(time.RFC3339),
	})
}
