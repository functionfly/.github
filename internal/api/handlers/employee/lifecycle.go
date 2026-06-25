package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListLifecycleEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeID, err := uuid.Parse(mux.Vars(r)["employeeId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListLifecycleEventsOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if et := q.Get("event_type"); et != "" {
		opts.EventType = &et
	}

	events, total, err := h.repo.ListLifecycleEvents(r.Context(), employeeID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list lifecycle events")
		apierror.WriteError(w, apierror.NewInternal("Failed to list lifecycle events"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

type triggerLifecycleRequest struct {
	EmployeeID  string  `json:"employee_id"`
	EventType   string  `json:"event_type"`
	TriggeredBy *string `json:"triggered_by,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

func (h *Handler) HandleTriggerLifecycle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req triggerLifecycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.EmployeeID == "" || req.EventType == "" {
		apierror.WriteError(w, apierror.NewBadRequest("employee_id and event_type are required"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	validTypes := map[string]bool{
		"hired": true, "onboarded": true, "promoted": true, "transferred": true,
		"leave_start": true, "leave_end": true, "offboarding_started": true,
		"terminated": true, "reactivated": true,
	}
	if !validTypes[req.EventType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid event_type"))
		return
	}

	ev := &storage.LifecycleEvent{
		EmployeeID: employeeID,
		TenantID:   claims.TenantID,
		EventType:  req.EventType,
		Notes:      req.Notes,
	}
	if req.TriggeredBy != nil {
		if tid, err := uuid.Parse(*req.TriggeredBy); err == nil {
			ev.TriggeredBy = &tid
		}
	}
	if req.Payload != nil {
		ev.Payload = storage.JSONMap(req.Payload)
	}

	created, err := h.repo.CreateLifecycleEvent(r.Context(), ev)
	if err != nil {
		h.log.WithError(err).Error("Failed to trigger lifecycle event")
		apierror.WriteError(w, apierror.NewInternal("Failed to trigger lifecycle event"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": created,
	})
}

func (h *Handler) HandleListWorkflows(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	workflows, total, err := h.repo.ListLifecycleWorkflows(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list workflows")
		apierror.WriteError(w, apierror.NewInternal("Failed to list workflows"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflows": workflows,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

type createWorkflowRequest struct {
	Name         string   `json:"name"`
	Description  *string  `json:"description,omitempty"`
	TriggerEvent string   `json:"trigger_event"`
	Steps        []map[string]interface{} `json:"steps,omitempty"`
}

func (h *Handler) HandleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.TriggerEvent == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name and trigger_event are required"))
		return
	}

	wf := &storage.LifecycleWorkflow{
		TenantID:     claims.TenantID,
		Name:         req.Name,
		Description:  req.Description,
		TriggerEvent: req.TriggerEvent,
		IsActive:     true,
	}
	if req.Steps != nil {
		wf.Steps = storage.JSONMap{"steps": req.Steps}
	}

	created, err := h.repo.CreateLifecycleWorkflow(r.Context(), wf)
	if err != nil {
		h.log.WithError(err).Error("Failed to create workflow")
		apierror.WriteError(w, apierror.NewInternal("Failed to create workflow"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow": created,
	})
}

func (h *Handler) HandleGetWorkflowInstance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid instance ID"))
		return
	}

	inst, err := h.repo.GetLifecycleWorkflowInstance(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get workflow instance")
		apierror.WriteError(w, apierror.NewInternal("Failed to get workflow instance"))
		return
	}
	if inst == nil {
		apierror.WriteError(w, apierror.NewNotFound("Workflow instance not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"instance": inst,
	})
}

type completeStepRequest struct {
	StepIdx int     `json:"step_idx"`
	Notes   *string `json:"notes,omitempty"`
}

func (h *Handler) HandleCompleteWorkflowStep(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid instance ID"))
		return
	}

	inst, err := h.repo.GetLifecycleWorkflowInstance(r.Context(), id)
	if err != nil || inst == nil {
		apierror.WriteError(w, apierror.NewNotFound("Workflow instance not found"))
		return
	}

	var req completeStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := map[string]interface{}{
		"current_step": req.StepIdx + 1,
	}

	wf, err := h.repo.GetLifecycleWorkflowByID(r.Context(), inst.WorkflowID)
	if err == nil && wf != nil {
		if stepsRaw, ok := wf.Steps["steps"].([]interface{}); ok && req.StepIdx+1 >= len(stepsRaw) {
			updates["status"] = "completed"
		}
	}

	if err := h.repo.UpdateLifecycleWorkflowInstance(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to complete workflow step")
		apierror.WriteError(w, apierror.NewInternal("Failed to complete workflow step"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
