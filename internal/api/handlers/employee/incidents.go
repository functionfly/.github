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

func (h *Handler) HandleListIncidents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListIncidentsOpts{
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
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if s := q.Get("severity"); s != "" {
		opts.Severity = &s
	}

	incidents, total, err := h.repo.ListFWOSIncidents(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list incidents")
		apierror.WriteError(w, apierror.NewInternal("Failed to list incidents"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incidents": incidents,
		"total":     total,
		"limit":     opts.Limit,
		"offset":    opts.Offset,
	})
}

func (h *Handler) HandleGetIncident(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	inc, err := h.repo.GetFWOSIncidentByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to get incident"))
		return
	}
	if inc == nil {
		apierror.WriteError(w, apierror.NewNotFound("Incident not found"))
		return
	}

	responders, _ := h.repo.ListIncidentResponders(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incident":   inc,
		"responders": responders,
	})
}

type createIncidentRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Severity    string  `json:"severity,omitempty"`
	CommanderID *string `json:"commander_id,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
}

func (h *Handler) HandleCreateIncident(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title is required"))
		return
	}

	inc := &storage.FWOSIncident{
		TenantID:    claims.TenantID,
		Title:       req.Title,
		Description: req.Description,
		Severity:    "medium",
		Status:      "open",
	}
	if req.Severity != "" {
		inc.Severity = req.Severity
	}
	if req.CommanderID != nil {
		if cid, err := uuid.Parse(*req.CommanderID); err == nil {
			inc.CommanderID = &cid
		}
	}
	if req.ProjectID != nil {
		if pid, err := uuid.Parse(*req.ProjectID); err == nil {
			inc.ProjectID = &pid
		}
	}

	created, err := h.repo.CreateFWOSIncident(r.Context(), inc)
	if err != nil {
		h.log.WithError(err).Error("Failed to create incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to create incident"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incident": created,
	})
}

func (h *Handler) HandleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if err := h.repo.UpdateFWOSIncident(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update incident")
		apierror.WriteError(w, apierror.NewInternal("Failed to update incident"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type addIncidentEventRequest struct {
	AuthorID  string                 `json:"author_id"`
	EventType string                 `json:"event_type,omitempty"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func (h *Handler) HandleAddIncidentEvent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	incidentID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	var req addIncidentEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("body is required"))
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid author_id"))
		return
	}

	ev := &storage.IncidentEvent{
		IncidentID: incidentID,
		AuthorID:   authorID,
		EventType:  "update",
		Body:       req.Body,
	}
	if req.EventType != "" {
		ev.EventType = req.EventType
	}
	if req.Metadata != nil {
		ev.Metadata = storage.JSONMap(req.Metadata)
	}

	created, err := h.repo.CreateIncidentEvent(r.Context(), ev)
	if err != nil {
		h.log.WithError(err).Error("Failed to create incident event")
		apierror.WriteError(w, apierror.NewInternal("Failed to create incident event"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": created,
	})
}

func (h *Handler) HandleListIncidentEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	incidentID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	limit := 50
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

	events, err := h.repo.ListIncidentEvents(r.Context(), incidentID, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list incident events")
		apierror.WriteError(w, apierror.NewInternal("Failed to list incident events"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

type addResponderRequest struct {
	EmployeeID string `json:"employee_id"`
	Role       string `json:"role,omitempty"`
}

func (h *Handler) HandleAddResponder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	incidentID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	var req addResponderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee_id"))
		return
	}

	resp := &storage.IncidentResponder{
		IncidentID: incidentID,
		EmployeeID: employeeID,
		Role:       "responder",
	}
	if req.Role != "" {
		resp.Role = req.Role
	}

	created, err := h.repo.AddIncidentResponder(r.Context(), resp)
	if err != nil {
		h.log.WithError(err).Error("Failed to add responder")
		apierror.WriteError(w, apierror.NewInternal("Failed to add responder"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"responder": created,
	})
}

type createPostmortemRequest struct {
	IncidentID         string   `json:"incident_id"`
	Summary            string   `json:"summary"`
	RootCause          string   `json:"root_cause"`
	ContributingFactors *string `json:"contributing_factors,omitempty"`
	WhatWentWell       *string `json:"what_went_well,omitempty"`
	WhatWentWrong      *string `json:"what_went_wrong,omitempty"`
	ActionItems        []map[string]interface{} `json:"action_items,omitempty"`
	LessonsLearned     *string `json:"lessons_learned,omitempty"`
}

func (h *Handler) HandleCreatePostmortem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req createPostmortemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.IncidentID == "" || req.Summary == "" || req.RootCause == "" {
		apierror.WriteError(w, apierror.NewBadRequest("incident_id, summary, and root_cause are required"))
		return
	}

	incidentID, err := uuid.Parse(req.IncidentID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident_id"))
		return
	}

	pm := &storage.Postmortem{
		IncidentID:         incidentID,
		TenantID:           claims.TenantID,
		AuthorID:           emp.ID,
		Summary:            req.Summary,
		RootCause:          req.RootCause,
		ContributingFactors: req.ContributingFactors,
		WhatWentWell:       req.WhatWentWell,
		WhatWentWrong:      req.WhatWentWrong,
		LessonsLearned:     req.LessonsLearned,
		Status:             "draft",
	}
	if req.ActionItems != nil {
		pm.ActionItems = storage.JSONMap{"items": req.ActionItems}
	}

	created, err := h.repo.CreatePostmortem(r.Context(), pm)
	if err != nil {
		h.log.WithError(err).Error("Failed to create postmortem")
		apierror.WriteError(w, apierror.NewInternal("Failed to create postmortem"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"postmortem": created,
	})
}

func (h *Handler) HandleGetPostmortem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	incidentID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid incident ID"))
		return
	}

	pm, err := h.repo.GetPostmortemByIncident(r.Context(), incidentID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get postmortem")
		apierror.WriteError(w, apierror.NewInternal("Failed to get postmortem"))
		return
	}
	if pm == nil {
		apierror.WriteError(w, apierror.NewNotFound("Postmortem not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"postmortem": pm,
	})
}
