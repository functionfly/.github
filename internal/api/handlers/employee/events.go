package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListFWOSEventsOpts{
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
	if src := q.Get("source"); src != "" {
		opts.Source = &src
	}
	if rt := q.Get("resource_type"); rt != "" {
		opts.ResourceType = &rt
	}

	events, total, err := h.repo.ListFWOSEvents(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list FWOS events")
		apierror.WriteError(w, apierror.NewInternal("Failed to list events"))
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

type createEventRequest struct {
	EventType    string                 `json:"event_type"`
	Source       string                 `json:"source"`
	ActorID      *string                `json:"actor_id,omitempty"`
	ResourceType *string                `json:"resource_type,omitempty"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

func (h *Handler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.EventType == "" || req.Source == "" {
		apierror.WriteError(w, apierror.NewBadRequest("event_type and source are required"))
		return
	}

	ev := &storage.FWOSEvent{
		TenantID:     claims.TenantID,
		EventType:    req.EventType,
		Source:       req.Source,
		ResourceType: req.ResourceType,
	}
	if req.ActorID != nil {
		if aid, err := uuid.Parse(*req.ActorID); err == nil {
			ev.ActorID = &aid
		}
	}
	if req.ResourceID != nil {
		if rid, err := uuid.Parse(*req.ResourceID); err == nil {
			ev.ResourceID = &rid
		}
	}
	if req.Payload != nil {
		ev.Payload = storage.JSONMap(req.Payload)
	}

	created, err := h.repo.CreateFWOSEvent(r.Context(), ev)
	if err != nil {
		h.log.WithError(err).Error("Failed to create FWOS event")
		apierror.WriteError(w, apierror.NewInternal("Failed to create event"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event": created,
	})
}
