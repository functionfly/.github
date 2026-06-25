package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListMentorships(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	opts := storage.ListMentorshipMatchesOpts{
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

	matches, total, err := h.repo.ListMentorshipMatches(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list mentorships")
		apierror.WriteError(w, apierror.NewInternal("Failed to list mentorships"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mentorships": matches,
		"total":       total,
		"limit":       opts.Limit,
		"offset":      opts.Offset,
	})
}

type requestMentorshipRequest struct {
	MentorID  string `json:"mentor_id"`
	FocusArea string `json:"focus_area,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

func (h *Handler) HandleRequestMentorship(w http.ResponseWriter, r *http.Request) {
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

	var req requestMentorshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.MentorID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("mentor_id is required"))
		return
	}

	mentorID, err := uuid.Parse(req.MentorID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid mentor ID"))
		return
	}

	match := &storage.MentorshipMatch{
		TenantID:  claims.TenantID,
		MentorID:  mentorID,
		MenteeID:  emp.ID,
		Status:    "active",
		MeetingFrequency: strPtr("biweekly"),
	}
	if req.FocusArea != "" {
		match.FocusArea = &req.FocusArea
	}
	if req.Notes != "" {
		match.Notes = &req.Notes
	}

	created, err := h.repo.CreateMentorshipMatch(r.Context(), match)
	if err != nil {
		h.log.WithError(err).Error("Failed to create mentorship")
		apierror.WriteError(w, apierror.NewInternal("Failed to create mentorship"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mentorship": created,
	})
}

type updateMentorshipRequest struct {
	Status           *string `json:"status,omitempty"`
	FocusArea        *string `json:"focus_area,omitempty"`
	MeetingFrequency *string `json:"meeting_frequency,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

func (h *Handler) HandleUpdateMentorship(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid mentorship ID"))
		return
	}

	match, err := h.repo.GetMentorshipMatchByID(r.Context(), id)
	if err != nil || match == nil {
		apierror.WriteError(w, apierror.NewNotFound("Mentorship not found"))
		return
	}

	var req updateMentorshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := map[string]interface{}{}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == "completed" || *req.Status == "cancelled" {
			now := time.Now()
			updates["ended_at"] = &now
		}
	}
	if req.FocusArea != nil {
		updates["focus_area"] = *req.FocusArea
	}
	if req.MeetingFrequency != nil {
		updates["meeting_frequency"] = *req.MeetingFrequency
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := h.repo.UpdateMentorshipMatch(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update mentorship")
		apierror.WriteError(w, apierror.NewInternal("Failed to update mentorship"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func strPtr(s string) *string {
	return &s
}
