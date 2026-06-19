package state

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleCreateTrigger handles POST /v1/triggers
func (h *Handler) HandleCreateTrigger(w http.ResponseWriter, r *http.Request) {
	var req CreateTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	if req.StatePath != "" {
		state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, req.StatePath)
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("state not found"))
			return
		}

		if !h.requirePermission(w, r, state.ID, claims.UserID, "can_trigger") {
			return
		}

		trigger := &staterepo.StateTrigger{
			TenantID:                tenantID,
			SourceStateID:           &state.ID,
			TriggerType:             req.TriggerType,
			KeyPattern:              strPtr(req.KeyPattern),
			Condition:               req.Condition,
			TargetFunctionID:        req.TargetFunctionID,
			TargetFunction:          req.TargetFunction,
			IncludePrevious:         req.IncludePrevious,
			IncludeNew:              req.IncludeNew,
			MaxInvocationsPerMinute: req.MaxInvocationsPerMinute,
			IsActive:                req.IsActive,
		}

		created, err := h.stateRepo.CreateTrigger(r.Context(), trigger)
		if err != nil {
			logrus.Errorf("failed to create trigger: %v", err)
			apierror.WriteError(w, apierror.NewInternal("failed to create trigger"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(created)
		return
	}

	trigger := &staterepo.StateTrigger{
		TenantID:                tenantID,
		TriggerType:             req.TriggerType,
		KeyPattern:              strPtr(req.KeyPattern),
		Condition:               req.Condition,
		TargetFunctionID:        req.TargetFunctionID,
		TargetFunction:          req.TargetFunction,
		IncludePrevious:         req.IncludePrevious,
		IncludeNew:              req.IncludeNew,
		MaxInvocationsPerMinute: req.MaxInvocationsPerMinute,
		IsActive:                req.IsActive,
	}

	created, err := h.stateRepo.CreateTrigger(r.Context(), trigger)
	if err != nil {
		logrus.Errorf("failed to create trigger: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to create trigger"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

// HandleGetTriggers handles GET /v1/triggers
func (h *Handler) HandleGetTriggers(w http.ResponseWriter, r *http.Request) {
	statePath := r.URL.Query().Get("state")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	if statePath != "" {
		state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, statePath)
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("state not found"))
			return
		}

		if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
			return
		}

		triggers, err := h.stateRepo.GetTriggers(r.Context(), state.ID)
		if err != nil {
			logrus.Errorf("failed to get triggers: %v", err)
			apierror.WriteError(w, apierror.NewInternal("failed to get triggers"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(triggers)
	} else {
		// List all triggers for the tenant with pagination
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = 20
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset == 0 {
			offset = 0
		}

		triggers, total, err := h.stateRepo.ListTriggersByTenant(r.Context(), tenantID, limit, offset)
		if err != nil {
			logrus.Errorf("failed to list triggers: %v", err)
			apierror.WriteError(w, apierror.NewInternal("failed to list triggers"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"triggers": triggers,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		})
	}
}

// HandleDeleteTrigger handles DELETE /v1/triggers/{id}
func (h *Handler) HandleDeleteTrigger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	triggerID := vars["id"]

	triggerUUID, err := uuid.Parse(triggerID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid trigger ID"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	trigger, err := h.stateRepo.GetTrigger(r.Context(), triggerUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("trigger not found"))
		return
	}

	if trigger.SourceStateID != nil {
		if !h.requirePermission(w, r, *trigger.SourceStateID, claims.UserID, "can_trigger") {
			return
		}
	}

	err = h.stateRepo.DeleteTrigger(r.Context(), triggerUUID)
	if err != nil {
		logrus.Errorf("failed to delete trigger: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to delete trigger"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
