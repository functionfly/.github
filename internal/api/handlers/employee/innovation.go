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

func (h *Handler) HandleListGrants(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListInnovationGrantsOpts{
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

	grants, total, err := h.repo.ListInnovationGrants(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list grants")
		apierror.WriteError(w, apierror.NewInternal("Failed to list grants"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grants": grants,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

type createGrantRequest struct {
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	Category             string  `json:"category,omitempty"`
	RequestedAmountCents *int64  `json:"requested_amount_cents,omitempty"`
}

func (h *Handler) HandleCreateGrant(w http.ResponseWriter, r *http.Request) {
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

	var req createGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" || req.Description == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title and description are required"))
		return
	}

	grant := &storage.InnovationGrant{
		TenantID:             claims.TenantID,
		ProposerID:           emp.ID,
		Title:                req.Title,
		Description:          req.Description,
		Category:             "technical",
		RequestedAmountCents: req.RequestedAmountCents,
		Status:               "draft",
	}
	if req.Category != "" {
		grant.Category = req.Category
	}

	created, err := h.repo.CreateInnovationGrant(r.Context(), grant)
	if err != nil {
		h.log.WithError(err).Error("Failed to create grant")
		apierror.WriteError(w, apierror.NewInternal("Failed to create grant"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grant": created,
	})
}

func (h *Handler) HandleSubmitGrant(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid grant ID"))
		return
	}

	grant, err := h.repo.GetInnovationGrantByID(r.Context(), id)
	if err != nil || grant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Grant not found"))
		return
	}

	if grant.Status != "draft" {
		apierror.WriteError(w, apierror.NewBadRequest("Grant can only be submitted from draft status"))
		return
	}

	if err := h.repo.UpdateInnovationGrant(r.Context(), id, map[string]interface{}{
		"status": "submitted",
	}); err != nil {
		h.log.WithError(err).Error("Failed to submit grant")
		apierror.WriteError(w, apierror.NewInternal("Failed to submit grant"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type voteGrantRequest struct {
	Vote    bool   `json:"vote"`
	Comment string `json:"comment,omitempty"`
}

func (h *Handler) HandleVoteGrant(w http.ResponseWriter, r *http.Request) {
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

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid grant ID"))
		return
	}

	grant, err := h.repo.GetInnovationGrantByID(r.Context(), id)
	if err != nil || grant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Grant not found"))
		return
	}

	existing, err := h.repo.GetInnovationGrantVoteByVoter(r.Context(), id, emp.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to check existing vote")
		apierror.WriteError(w, apierror.NewInternal("Failed to check vote"))
		return
	}
	if existing != nil {
		apierror.WriteError(w, apierror.NewBadRequest("You have already voted on this grant"))
		return
	}

	var req voteGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	vote := &storage.InnovationGrantVote{
		GrantID: id,
		VoterID: emp.ID,
		Vote:    req.Vote,
	}
	if req.Comment != "" {
		vote.Comment = &req.Comment
	}

	if _, err := h.repo.CreateInnovationGrantVote(r.Context(), vote); err != nil {
		h.log.WithError(err).Error("Failed to create vote")
		apierror.WriteError(w, apierror.NewInternal("Failed to create vote"))
		return
	}

	updates := map[string]interface{}{}
	if req.Vote {
		updates["votes_for"] = grant.VotesFor + 1
	} else {
		updates["votes_against"] = grant.VotesAgainst + 1
	}
	if err := h.repo.UpdateInnovationGrant(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update vote count")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type reviewGrantRequest struct {
	Status          string  `json:"status"`
	FeasibilityScore *float64 `json:"feasibility_score,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

func (h *Handler) HandleReviewGrant(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	reviewer, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || reviewer == nil {
		apierror.WriteError(w, apierror.NewNotFound("Reviewer profile not found"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid grant ID"))
		return
	}

	grant, err := h.repo.GetInnovationGrantByID(r.Context(), id)
	if err != nil || grant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Grant not found"))
		return
	}

	if grant.Status != "submitted" && grant.Status != "under_review" {
		apierror.WriteError(w, apierror.NewBadRequest("Grant must be submitted or under review"))
		return
	}

	var req reviewGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      req.Status,
		"reviewed_by": reviewer.ID,
		"reviewed_at": &now,
	}
	if req.FeasibilityScore != nil {
		updates["feasibility_score"] = *req.FeasibilityScore
	}
	if req.RejectionReason != nil {
		updates["rejection_reason"] = *req.RejectionReason
	}
	if req.Status == "funded" {
		updates["funded_at"] = &now
	}

	if err := h.repo.UpdateInnovationGrant(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to review grant")
		apierror.WriteError(w, apierror.NewInternal("Failed to review grant"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
