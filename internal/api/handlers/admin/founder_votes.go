package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type FounderVotesHandler struct {
	repo storage.Repository
	log  *logrus.Logger
}

func NewFounderVotesHandler(repo storage.Repository, log *logrus.Logger) *FounderVotesHandler {
	return &FounderVotesHandler{repo: repo, log: log}
}

type voteOptionInput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type changeDiffInput struct {
	Summary  string                   `json:"summary,omitempty"`
	Changes  []map[string]string      `json:"changes,omitempty"`
	Impact   string                   `json:"impact,omitempty"`
	Rationale string                  `json:"rationale,omitempty"`
	Category string                   `json:"category,omitempty"`
}

type createVoteRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Options     []voteOptionInput `json:"options"`
	ChangeDiff  *changeDiffInput  `json:"change_diff,omitempty"`
	Quorum      int               `json:"quorum"`
}

type updateVoteRequest struct {
	Title       *string                   `json:"title,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Status      *string                   `json:"status,omitempty"`
	Quorum      *int                      `json:"quorum,omitempty"`
}

func (h *FounderVotesHandler) HandleListVotes(w http.ResponseWriter, r *http.Request) {
	votes, err := h.repo.ListFounderVotes(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to list founder votes")
		apierror.WriteError(w, apierror.NewInternal("Failed to list votes"))
		return
	}

	type voteItem struct {
		ID          uuid.UUID   `json:"id"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Options     interface{} `json:"options"`
		ChangeDiff  interface{} `json:"change_diff,omitempty"`
		Status      string      `json:"status"`
		Quorum      int         `json:"quorum"`
		CreatedBy   uuid.UUID   `json:"created_by"`
		CreatedAt   string      `json:"created_at"`
		UpdatedAt   string      `json:"updated_at"`
	}

	items := make([]voteItem, 0, len(votes))
	for _, v := range votes {
		var opts interface{}
		_ = json.Unmarshal([]byte(v.Options), &opts)

		var changeDiff interface{}
		if v.ChangeDiff != "" && v.ChangeDiff != "{}" {
			_ = json.Unmarshal([]byte(v.ChangeDiff), &changeDiff)
		}

		_, total, _ := h.repo.GetFounderVoteResults(r.Context(), v.ID)

		items = append(items, voteItem{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			Options:     opts,
			ChangeDiff:  changeDiff,
			Status:      v.Status,
			Quorum:      v.Quorum,
			CreatedBy:   v.CreatedBy,
			CreatedAt:   v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   v.UpdatedAt.Format(time.RFC3339),
		})

		_ = total
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"votes": items,
		"total": len(items),
	})
}

func (h *FounderVotesHandler) HandleGetVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voteID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid vote ID"))
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		apierror.WriteError(w, apierror.NewNotFound("Vote not found"))
		return
	}

	var opts interface{}
	_ = json.Unmarshal([]byte(vote.Options), &opts)

	var changeDiff interface{}
	if vote.ChangeDiff != "" && vote.ChangeDiff != "{}" {
		_ = json.Unmarshal([]byte(vote.ChangeDiff), &changeDiff)
	}

	results, total, _ := h.repo.GetFounderVoteResults(r.Context(), voteID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vote": map[string]interface{}{
			"id":          vote.ID,
			"title":       vote.Title,
			"description": vote.Description,
			"options":     opts,
			"change_diff": changeDiff,
			"status":      vote.Status,
			"quorum":      vote.Quorum,
			"results":     results,
			"total_votes": total,
			"created_by":  vote.CreatedBy,
			"created_at":  vote.CreatedAt.Format(time.RFC3339),
			"updated_at":  vote.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *FounderVotesHandler) HandleCreateVote(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Title is required"))
		return
	}
	if len(req.Options) < 2 {
		apierror.WriteError(w, apierror.NewBadRequest("At least 2 options are required"))
		return
	}

	for i, opt := range req.Options {
		if opt.ID == "" {
			req.Options[i].ID = uuid.New().String()[:8]
		}
	}

	optsJSON, err := json.Marshal(req.Options)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid options"))
		return
	}

	changeDiffJSON := "{}"
	if req.ChangeDiff != nil {
		cd, err := json.Marshal(req.ChangeDiff)
		if err == nil {
			changeDiffJSON = string(cd)
		}
	}

	vote := &storage.FounderVote{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Options:     string(optsJSON),
		ChangeDiff:  changeDiffJSON,
		Status:      "active",
		Quorum:      req.Quorum,
		CreatedBy:   claims.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.CreateFounderVote(r.Context(), vote); err != nil {
		h.log.WithError(err).Error("Failed to create founder vote")
		apierror.WriteError(w, apierror.NewInternal("Failed to create vote"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"vote": map[string]interface{}{
			"id":     vote.ID,
			"title":  vote.Title,
			"status": vote.Status,
		},
	})
}

func (h *FounderVotesHandler) HandleUpdateVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voteID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid vote ID"))
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		apierror.WriteError(w, apierror.NewNotFound("Vote not found"))
		return
	}

	var req updateVoteRequest
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
	if req.Status != nil {
		validStatuses := map[string]bool{"active": true, "closed": true, "passed": true, "rejected": true}
		if !validStatuses[*req.Status] {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid status. Must be: active, closed, passed, rejected"))
			return
		}
		updates["status"] = *req.Status
	}
	if req.Quorum != nil {
		updates["quorum"] = *req.Quorum
	}

	if len(updates) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("No fields to update"))
		return
	}

	updates["updated_at"] = time.Now()

	if err := h.repo.UpdateFounderVote(r.Context(), voteID, updates); err != nil {
		h.log.WithError(err).Error("Failed to update founder vote")
		apierror.WriteError(w, apierror.NewInternal("Failed to update vote"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *FounderVotesHandler) HandleDeleteVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voteID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid vote ID"))
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		apierror.WriteError(w, apierror.NewNotFound("Vote not found"))
		return
	}

	if err := h.repo.DeleteFounderVote(r.Context(), voteID); err != nil {
		h.log.WithError(err).Error("Failed to delete founder vote")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete vote"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
