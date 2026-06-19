package statefabric

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/apierror"
	repo "github.com/functionfly/functionfly/internal/storage/statefabric"
)

type createFabricTriggerRequest struct {
	TriggerType             string                 `json:"triggerType"`
	KeyPattern              string                 `json:"keyPattern,omitempty"`
	Condition               map[string]interface{} `json:"condition,omitempty"`
	TargetFunctionID        *uuid.UUID             `json:"targetFunctionId,omitempty"`
	TargetFunction          string                 `json:"targetFunction"`
	IncludePrevious         bool                   `json:"includePrevious"`
	IncludeNew              bool                   `json:"includeNew"`
	MaxInvocationsPerMinute int                    `json:"maxInvocationsPerMinute"`
	IsActive                bool                   `json:"isActive"`
}

func (h *Handler) HandleListTriggers(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	if !h.requireFabricPermission(w, r, tenantID, userID, fabricID, fabricPermRead) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	triggers, err := h.repo.ListFabricTriggers(r.Context(), tenantID, fabricID)
	if err != nil {
		logrus.WithError(err).Error("failed to list fabric triggers")
		apierror.WriteError(w, apierror.NewInternal("failed to list triggers"))
		return
	}
	total := len(triggers)
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	page := triggers
	if offset > 0 || (limit > 0 && end < total) {
		page = triggers[offset:end]
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"triggers": page,
		"total":    total,
	})
}

func (h *Handler) HandleCreateTrigger(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	fabricID, parsed := parseID(w, mux.Vars(r)["id"], "state fabric id")
	if !parsed {
		return
	}
	if !h.requireFabricPermission(w, r, tenantID, userID, fabricID, fabricPermTrigger) {
		return
	}
	var req createFabricTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	if req.TargetFunction == "" && req.TargetFunctionID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("targetFunction is required"))
		return
	}
	created, err := h.repo.CreateFabricTrigger(r.Context(), tenantID, fabricID, repo.FabricTriggerInput{
		TriggerType:             req.TriggerType,
		KeyPattern:              req.KeyPattern,
		Condition:               req.Condition,
		TargetFunctionID:        req.TargetFunctionID,
		TargetFunction:          req.TargetFunction,
		IncludePrevious:         req.IncludePrevious,
		IncludeNew:              req.IncludeNew,
		MaxInvocationsPerMinute: req.MaxInvocationsPerMinute,
		IsActive:                req.IsActive,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to create fabric trigger")
		apierror.WriteError(w, apierror.NewInternal("failed to create trigger"))
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) HandleDeleteTrigger(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	if !h.requireFabricPermission(w, r, tenantID, userID, fabricID, fabricPermTrigger) {
		return
	}
	triggerID, parsed := parseID(w, vars["triggerId"], "trigger id")
	if !parsed {
		return
	}
	if err := h.repo.DeleteFabricTrigger(r.Context(), tenantID, fabricID, triggerID); err != nil {
		logrus.WithError(err).Error("failed to delete fabric trigger")
		apierror.WriteError(w, apierror.NewNotFound("trigger not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleUpdateTrigger(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := tenantAndUser(r, w)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	fabricID, parsed := parseID(w, vars["id"], "state fabric id")
	if !parsed {
		return
	}
	if !h.requireFabricPermission(w, r, tenantID, userID, fabricID, fabricPermTrigger) {
		return
	}
	triggerID, parsed := parseID(w, vars["triggerId"], "trigger id")
	if !parsed {
		return
	}
	var req createFabricTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}
	if req.TargetFunction == "" && req.TargetFunctionID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("targetFunction is required"))
		return
	}
	updated, err := h.repo.UpdateFabricTrigger(r.Context(), tenantID, fabricID, triggerID, repo.FabricTriggerInput{
		TriggerType:             req.TriggerType,
		KeyPattern:              req.KeyPattern,
		Condition:               req.Condition,
		TargetFunctionID:        req.TargetFunctionID,
		TargetFunction:          req.TargetFunction,
		IncludePrevious:         req.IncludePrevious,
		IncludeNew:              req.IncludeNew,
		MaxInvocationsPerMinute: req.MaxInvocationsPerMinute,
		IsActive:                req.IsActive,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to update fabric trigger")
		apierror.WriteError(w, apierror.NewNotFound("trigger not found"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
