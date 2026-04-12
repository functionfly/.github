package flywheel

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/mux"
)

// ListChallenges handles GET /api/v1/flywheel/challenges
func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	filter := flywheel.ChallengeFilter{
		Status: flywheel.ChallengeStatus(r.URL.Query().Get("status")),
	}

	if r.URL.Query().Get("active_only") == "true" {
		filter.ActiveOnly = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	challenges, count, err := h.service.ListChallenges(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Debug("Flywheel challenges table may not exist, returning empty list")
		challenges = nil
		count = 0
	}

	response := map[string]interface{}{
		"challenges": challenges,
		"total":      count,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetChallenge handles GET /api/v1/flywheel/challenges/:id
func (h *Handler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := h.parseUUID(w, r, vars["id"], "challenge ID")
	if !ok {
		return
	}

	challenge, err := h.service.GetChallenge(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Challenge not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(challenge)
}

// SubmitChallenge handles POST /api/v1/flywheel/challenges/:id/submit
func (h *Handler) SubmitChallenge(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	challengeID, ok := h.parseUUID(w, r, vars["id"], "challenge ID")
	if !ok {
		return
	}

	var req ChallengeSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	submission := &flywheel.ChallengeSubmission{
		ChallengeID:    challengeID,
		ParticipantID:  user.UserID,
		SubmissionType: req.SubmissionType,
		CodeSubmission: req.CodeSubmission,
		Notes:          SanitizeContent(req.Notes),
	}

	if req.SubmittedCapsuleID != nil {
		id, err := parseUUIDStr(*req.SubmittedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid submitted_capsule_id"}`, http.StatusBadRequest)
			return
		}
		submission.SubmittedCapsuleID = &id
	}

	if err := h.service.SubmitChallengeEntry(r.Context(), submission); err != nil {
		h.logger.WithError(err).Error("Failed to submit challenge entry")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submission)
}

// GetChallengeLeaderboard handles GET /api/v1/flywheel/challenges/:id/leaderboard
func (h *Handler) GetChallengeLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	challengeID, ok := h.parseUUID(w, r, vars["id"], "challenge ID")
	if !ok {
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	submissions, err := h.service.GetChallengeLeaderboard(r.Context(), challengeID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get challenge leaderboard")
		http.Error(w, `{"error":"Failed to get leaderboard"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submissions": submissions,
	})
}
