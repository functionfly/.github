package employee

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListCareerTimeline(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeIDStr := mux.Vars(r)["employeeId"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	events, err := h.repo.GetCareerTimeline(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get career timeline")
		apierror.WriteError(w, apierror.NewInternal("Failed to get career timeline"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline": events,
	})
}

type createTimelineEventRequest struct {
	EventType   string  `json:"event_type"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	EventDate   string  `json:"event_date"`
}

func (h *Handler) HandleCreateTimelineEvent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeIDStr := mux.Vars(r)["employeeId"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	var req createTimelineEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	validTypes := map[string]bool{
		"joined": true, "promoted": true, "transferred": true,
		"project_completed": true, "certification_earned": true,
		"achievement_unlocked": true, "award_received": true,
	}
	if !validTypes[req.EventType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid event_type"))
		return
	}

	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title is required"))
		return
	}

	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid event_date format (use YYYY-MM-DD)"))
		return
	}

	ev := &types.CareerTimelineEvent{
		EmployeeID:  employeeID,
		TenantID:    claims.TenantID,
		EventType:   req.EventType,
		Title:       req.Title,
		Description: req.Description,
		Metadata:    types.JSONMap(req.Metadata),
		EventDate:   eventDate,
	}

	created, err := h.repo.CreateCareerTimelineEvent(r.Context(), ev)
	if err != nil {
		h.log.WithError(err).Error("Failed to create timeline event")
		apierror.WriteError(w, apierror.NewInternal("Failed to create timeline event"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"event":   created,
	})
}
