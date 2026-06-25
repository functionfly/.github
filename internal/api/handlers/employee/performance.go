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

func (h *Handler) HandleListGoals(w http.ResponseWriter, r *http.Request) {
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
	opts := storage.ListPerformanceGoalsOpts{
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
	if c := q.Get("category"); c != "" {
		opts.Category = &c
	}
	if p := q.Get("priority"); p != "" {
		opts.Priority = &p
	}

	goals, total, err := h.repo.ListPerformanceGoals(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list goals")
		apierror.WriteError(w, apierror.NewInternal("Failed to list goals"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"goals":  goals,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

type createGoalRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Category    string  `json:"category,omitempty"`
	Priority    string  `json:"priority,omitempty"`
	TargetDate  string  `json:"target_date,omitempty"`
}

func (h *Handler) HandleCreateGoal(w http.ResponseWriter, r *http.Request) {
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

	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Title is required"))
		return
	}

	goal := &storage.PerformanceGoal{
		EmployeeID:  emp.ID,
		TenantID:    claims.TenantID,
		Title:       req.Title,
		Status:      "not_started",
		ProgressPct: 0,
	}
	if req.Description != "" {
		goal.Description = &req.Description
	}
	if req.Category != "" {
		goal.Category = req.Category
	} else {
		goal.Category = "professional"
	}
	if req.Priority != "" {
		goal.Priority = req.Priority
	} else {
		goal.Priority = "medium"
	}
	if req.TargetDate != "" {
		if t, err := time.Parse("2006-01-02", req.TargetDate); err == nil {
			goal.TargetDate = &t
		}
	}

	created, err := h.repo.CreatePerformanceGoal(r.Context(), goal)
	if err != nil {
		h.log.WithError(err).Error("Failed to create goal")
		apierror.WriteError(w, apierror.NewInternal("Failed to create goal"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"goal": created,
	})
}

type updateGoalRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
	ProgressPct *int    `json:"progress_pct,omitempty"`
}

func (h *Handler) HandleUpdateGoal(w http.ResponseWriter, r *http.Request) {
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

	goal, err := h.repo.GetPerformanceGoalByID(r.Context(), id)
	if err != nil || goal == nil {
		apierror.WriteError(w, apierror.NewNotFound("Goal not found"))
		return
	}

	var req updateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == "completed" {
			now := time.Now()
			updates["completed_at"] = &now
			updates["progress_pct"] = 100
		}
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.TargetDate != nil {
		if t, err := time.Parse("2006-01-02", *req.TargetDate); err == nil {
			updates["target_date"] = &t
		}
	}
	if req.ProgressPct != nil {
		updates["progress_pct"] = *req.ProgressPct
	}

	if err := h.repo.UpdatePerformanceGoal(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update goal")
		apierror.WriteError(w, apierror.NewInternal("Failed to update goal"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleListReviews(w http.ResponseWriter, r *http.Request) {
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
	opts := storage.ListPerformanceReviewsOpts{
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
	if t := q.Get("review_type"); t != "" {
		opts.ReviewType = &t
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if p := q.Get("period"); p != "" {
		opts.Period = &p
	}

	reviews, total, err := h.repo.ListPerformanceReviews(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list reviews")
		apierror.WriteError(w, apierror.NewInternal("Failed to list reviews"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reviews": reviews,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

type createReviewRequest struct {
	ReviewerID   string `json:"reviewer_id"`
	ReviewPeriod string `json:"review_period"`
	ReviewType   string `json:"review_type,omitempty"`
}

func (h *Handler) HandleCreateReview(w http.ResponseWriter, r *http.Request) {
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

	var req createReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.ReviewerID == "" || req.ReviewPeriod == "" {
		apierror.WriteError(w, apierror.NewBadRequest("reviewer_id and review_period are required"))
		return
	}

	reviewerID, err := uuid.Parse(req.ReviewerID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid reviewer ID"))
		return
	}

	revType := req.ReviewType
	if revType == "" {
		revType = "self"
	}

	rev := &storage.PerformanceReview{
		EmployeeID:   emp.ID,
		ReviewerID:   reviewerID,
		TenantID:     claims.TenantID,
		ReviewPeriod: req.ReviewPeriod,
		ReviewType:   revType,
		Status:       "draft",
	}

	created, err := h.repo.CreatePerformanceReview(r.Context(), rev)
	if err != nil {
		h.log.WithError(err).Error("Failed to create review")
		apierror.WriteError(w, apierror.NewInternal("Failed to create review"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"review": created,
	})
}

type submitReviewRequest struct {
	Strengths           *string `json:"strengths,omitempty"`
	AreasForImprovement *string `json:"areas_for_improvement,omitempty"`
	OverallRating       *int    `json:"overall_rating,omitempty"`
	Comments            *string `json:"comments,omitempty"`
}

func (h *Handler) HandleSubmitReview(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid review ID"))
		return
	}

	rev, err := h.repo.GetPerformanceReviewByID(r.Context(), id)
	if err != nil || rev == nil {
		apierror.WriteError(w, apierror.NewNotFound("Review not found"))
		return
	}

	var req submitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "submitted",
		"submitted_at": &now,
	}
	if req.Strengths != nil {
		updates["strengths"] = *req.Strengths
	}
	if req.AreasForImprovement != nil {
		updates["areas_for_improvement"] = *req.AreasForImprovement
	}
	if req.OverallRating != nil {
		updates["overall_rating"] = *req.OverallRating
	}
	if req.Comments != nil {
		updates["comments"] = *req.Comments
	}

	if err := h.repo.UpdatePerformanceReview(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to submit review")
		apierror.WriteError(w, apierror.NewInternal("Failed to submit review"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleListFeedback(w http.ResponseWriter, r *http.Request) {
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

	feedbacks, err := h.repo.ListPeerFeedback(r.Context(), emp.ID, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list feedback")
		apierror.WriteError(w, apierror.NewInternal("Failed to list feedback"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback": feedbacks,
	})
}

type giveFeedbackRequest struct {
	ReviewID     *string `json:"review_id,omitempty"`
	ToEmployeeID string  `json:"to_employee_id"`
	FeedbackText string  `json:"feedback_text"`
	Rating       *int    `json:"rating,omitempty"`
	IsAnonymous  bool    `json:"is_anonymous,omitempty"`
}

func (h *Handler) HandleGiveFeedback(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	fromEmp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || fromEmp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req giveFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.ToEmployeeID == "" || req.FeedbackText == "" {
		apierror.WriteError(w, apierror.NewBadRequest("to_employee_id and feedback_text are required"))
		return
	}

	toID, err := uuid.Parse(req.ToEmployeeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid to_employee_id"))
		return
	}

	fb := &storage.PeerFeedback{
		FromEmployeeID: fromEmp.ID,
		ToEmployeeID:   toID,
		TenantID:       claims.TenantID,
		FeedbackText:   req.FeedbackText,
		Rating:         req.Rating,
		IsAnonymous:    req.IsAnonymous,
	}
	if req.ReviewID != nil {
		rid, err := uuid.Parse(*req.ReviewID)
		if err == nil {
			fb.ReviewID = &rid
		}
	}

	created, err := h.repo.CreatePeerFeedback(r.Context(), fb)
	if err != nil {
		h.log.WithError(err).Error("Failed to create feedback")
		apierror.WriteError(w, apierror.NewInternal("Failed to create feedback"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback": created,
	})
}
