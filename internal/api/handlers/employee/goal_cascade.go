package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetGoalTree(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	goals, err := h.repo.GetGoalTree(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get goal tree")
		apierror.WriteError(w, apierror.NewInternal("Failed to get goal tree"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"goals": goals,
		"total": len(goals),
	})
}

type cascadeGoalRequest struct {
	ParentGoalID      *string `json:"parent_goal_id,omitempty"`
	GoalLevel         string  `json:"goal_level,omitempty"`
	CascadeVisibility string  `json:"cascade_visibility,omitempty"`
}

func (h *Handler) HandleCascadeGoal(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid goal ID"))
		return
	}

	var req cascadeGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	goalLevel := "personal"
	if req.GoalLevel != "" {
		goalLevel = req.GoalLevel
	}
	cascadeVisibility := "private"
	if req.CascadeVisibility != "" {
		cascadeVisibility = req.CascadeVisibility
	}

	var parentGoalID *uuid.UUID
	if req.ParentGoalID != nil {
		pid, err := uuid.Parse(*req.ParentGoalID)
		if err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid parent_goal_id"))
			return
		}
		parentGoalID = &pid
	}

	if err := h.repo.UpdateGoalCascade(r.Context(), id, parentGoalID, goalLevel, cascadeVisibility); err != nil {
		h.log.WithError(err).Error("Failed to update goal cascade")
		apierror.WriteError(w, apierror.NewInternal("Failed to update goal cascade"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
