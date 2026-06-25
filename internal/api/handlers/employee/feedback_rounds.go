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

func (h *Handler) HandleListFeedbackRounds(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListFeedbackRoundsOpts{
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

	rounds, total, err := h.repo.ListFeedbackRounds(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list feedback rounds")
		apierror.WriteError(w, apierror.NewInternal("Failed to list feedback rounds"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback_rounds": rounds,
		"total":           total,
		"limit":           opts.Limit,
		"offset":          opts.Offset,
	})
}

type createFeedbackRoundRequest struct {
	Name         string                   `json:"name"`
	Description  *string                  `json:"description,omitempty"`
	ReviewPeriod string                   `json:"review_period"`
	RoundType    string                   `json:"round_type,omitempty"`
	StartDate    string                   `json:"start_date"`
	EndDate      string                   `json:"end_date"`
	Questions    []map[string]interface{} `json:"questions,omitempty"`
}

func (h *Handler) HandleCreateFeedbackRound(w http.ResponseWriter, r *http.Request) {
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

	var req createFeedbackRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.ReviewPeriod == "" || req.StartDate == "" || req.EndDate == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name, review_period, start_date, and end_date are required"))
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid start_date format (use YYYY-MM-DD)"))
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid end_date format (use YYYY-MM-DD)"))
		return
	}

	roundType := "360"
	if req.RoundType != "" {
		roundType = req.RoundType
	}

	fr := &storage.FeedbackRound{
		TenantID:     claims.TenantID,
		Name:         req.Name,
		Description:  req.Description,
		ReviewPeriod: req.ReviewPeriod,
		RoundType:    roundType,
		Status:       "draft",
		StartDate:    startDate,
		EndDate:      endDate,
		CreatedBy:    emp.ID,
	}
	if req.Questions != nil {
		fr.Questions = storage.JSONMap{"questions": req.Questions}
	}

	created, err := h.repo.CreateFeedbackRound(r.Context(), fr)
	if err != nil {
		h.log.WithError(err).Error("Failed to create feedback round")
		apierror.WriteError(w, apierror.NewInternal("Failed to create feedback round"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback_round": created,
	})
}

func (h *Handler) HandleStartFeedbackRound(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid feedback round ID"))
		return
	}

	round, err := h.repo.GetFeedbackRoundByID(r.Context(), id)
	if err != nil || round == nil {
		apierror.WriteError(w, apierror.NewNotFound("Feedback round not found"))
		return
	}

	if round.Status != "draft" {
		apierror.WriteError(w, apierror.NewBadRequest("Feedback round is not in draft status"))
		return
	}

	if err := h.repo.UpdateFeedbackRound(r.Context(), id, map[string]interface{}{
		"status": "active",
	}); err != nil {
		h.log.WithError(err).Error("Failed to start feedback round")
		apierror.WriteError(w, apierror.NewInternal("Failed to start feedback round"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type submitFeedbackResponseRequest struct {
	AssignmentID   int64  `json:"assignment_id"`
	QuestionIndex  int    `json:"question_index"`
	ResponseText   string `json:"response_text,omitempty"`
	ResponseRating *int   `json:"response_rating,omitempty"`
}

func (h *Handler) HandleSubmitFeedbackResponse(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req submitFeedbackResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.AssignmentID == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("assignment_id is required"))
		return
	}

	resp := &storage.FeedbackRoundResponse{
		AssignmentID:   req.AssignmentID,
		QuestionIndex:  req.QuestionIndex,
		ResponseRating: req.ResponseRating,
	}
	if req.ResponseText != "" {
		resp.ResponseText = &req.ResponseText
	}

	created, err := h.repo.CreateFeedbackRoundResponse(r.Context(), resp)
	if err != nil {
		h.log.WithError(err).Error("Failed to submit feedback response")
		apierror.WriteError(w, apierror.NewInternal("Failed to submit feedback response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": created,
	})
}

func (h *Handler) HandleGetFeedbackResults(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid feedback round ID"))
		return
	}

	results, err := h.repo.GetFeedbackRoundResults(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get feedback results")
		apierror.WriteError(w, apierror.NewInternal("Failed to get feedback results"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}
