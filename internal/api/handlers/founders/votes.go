package founders

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type VoteOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type VotesHandler struct {
	repo  storage.Repository
	log   *logrus.Logger
}

func NewVotesHandler(repo storage.Repository, log *logrus.Logger) *VotesHandler {
	return &VotesHandler{
		repo:  repo,
		log:   log,
	}
}

func (h *VotesHandler) HandleListVotes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	votes, err := h.repo.ListActiveFounderVotes(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed to list founder votes")
		http.Error(w, "failed to list votes", http.StatusInternalServerError)
		return
	}

	type VoteResponse struct {
		ID          uuid.UUID      `json:"id"`
		Title       string         `json:"title"`
		Description string         `json:"description"`
		VoteType    string         `json:"vote_type"`
		Status      string         `json:"status"`
		Options     []VoteOption   `json:"options"`
		HasVoted    bool          `json:"has_voted"`
		MyVote      string        `json:"my_vote,omitempty"`
		StartsAt    *interface{}   `json:"starts_at,omitempty"`
		EndsAt      *interface{}   `json:"ends_at,omitempty"`
	}

	responses := make([]VoteResponse, 0, len(votes))
	for _, vote := range votes {
		var opts []VoteOption
		if err := json.Unmarshal(vote.Options, &opts); err != nil {
			h.log.WithError(err).Warn("Failed to unmarshal vote options")
			continue
		}

		existingResponse, _ := h.repo.GetFounderVoteResponse(r.Context(), vote.ID, claims.UserID)
		hasVoted := existingResponse != nil

		resp := VoteResponse{
			ID:          vote.ID,
			Title:       vote.Title,
			Description: vote.Description,
			VoteType:    vote.VoteType,
			Status:      vote.Status,
			Options:     opts,
			HasVoted:    hasVoted,
		}
		if hasVoted {
			resp.MyVote = existingResponse.OptionID
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"votes": responses,
	})
}

func (h *VotesHandler) HandleGetVote(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	voteIDStr := r.URL.Query().Get(":id")
	voteID, err := uuid.Parse(voteIDStr)
	if err != nil {
		http.Error(w, "invalid vote id", http.StatusBadRequest)
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		http.Error(w, "vote not found", http.StatusNotFound)
		return
	}

	var opts []VoteOption
	if err := json.Unmarshal(vote.Options, &opts); err != nil {
		h.log.WithError(err).Warn("Failed to unmarshal vote options")
	}

	results, total, _ := h.repo.GetFounderVoteResults(r.Context(), voteID)

	existingResponse, _ := h.repo.GetFounderVoteResponse(r.Context(), voteID, claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vote": map[string]interface{}{
			"id":          vote.ID,
			"title":       vote.Title,
			"description": vote.Description,
			"vote_type":   vote.VoteType,
			"status":      vote.Status,
			"options":     opts,
			"results":     results,
			"total_votes": total,
			"has_voted":   existingResponse != nil,
			"my_vote":     existingResponse.OptionID,
		},
	})
}

func (h *VotesHandler) HandleCastVote(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	voteIDStr := r.URL.Query().Get(":id")
	voteID, err := uuid.Parse(voteIDStr)
	if err != nil {
		http.Error(w, "invalid vote id", http.StatusBadRequest)
		return
	}

	var req struct {
		OptionID string `json:"option_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.OptionID == "" {
		http.Error(w, "option_id is required", http.StatusBadRequest)
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		http.Error(w, "vote not found", http.StatusNotFound)
		return
	}

	if vote.Status != "active" {
		http.Error(w, "vote is not active", http.StatusBadRequest)
		return
	}

	existingResponse, _ := h.repo.GetFounderVoteResponse(r.Context(), voteID, claims.UserID)
	if existingResponse != nil {
		http.Error(w, "already voted", http.StatusConflict)
		return
	}

	if err := h.repo.CastFounderVote(r.Context(), voteID, claims.UserID, req.OptionID); err != nil {
		h.log.WithError(err).Error("Failed to cast vote")
		http.Error(w, "failed to cast vote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Vote cast successfully",
	})
}

func (h *VotesHandler) HandleGetResults(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder {
		http.Error(w, "founder access required", http.StatusForbidden)
		return
	}

	voteIDStr := r.URL.Query().Get(":id")
	voteID, err := uuid.Parse(voteIDStr)
	if err != nil {
		http.Error(w, "invalid vote id", http.StatusBadRequest)
		return
	}

	vote, err := h.repo.GetFounderVote(r.Context(), voteID)
	if err != nil || vote == nil {
		http.Error(w, "vote not found", http.StatusNotFound)
		return
	}

	var opts []VoteOption
	if err := json.Unmarshal(vote.Options, &opts); err != nil {
		h.log.WithError(err).Warn("Failed to unmarshal vote options")
	}

	results, total, _ := h.repo.GetFounderVoteResults(r.Context(), voteID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vote_id":     voteID,
		"title":      vote.Title,
		"status":     vote.Status,
		"options":    opts,
		"results":    results,
		"total_votes": total,
	})
}
