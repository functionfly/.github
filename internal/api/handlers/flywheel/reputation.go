package flywheel

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/mux"
)

// GetMyReputation handles GET /api/v1/flywheel/reputation/me
func (h *Handler) GetMyReputation(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	scores, err := h.service.GetUserReputation(r.Context(), user.UserID)
	if err != nil {
		h.logger.WithError(err).Debug("Flywheel reputation table may not exist, returning default profile")
		scores = &flywheel.ReputationScores{
			UserID:             user.UserID,
			BuilderTier:        flywheel.TierBronze,
			OptimizerTier:      flywheel.TierBronze,
			MentorTier:         flywheel.TierBronze,
			AgentWhispererTier: flywheel.TierBronze,
			ReliabilityIndex:   100,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"profile": scores})
}

// GetUserReputation handles GET /api/v1/flywheel/reputation/:user_id
func (h *Handler) GetUserReputation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, ok := h.parseUUID(w, r, vars["user_id"], "user ID")
	if !ok {
		return
	}

	scores, err := h.service.GetUserReputation(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Debug("Flywheel reputation table may not exist, returning default profile")
		scores = &flywheel.ReputationScores{
			UserID:             userID,
			BuilderTier:        flywheel.TierBronze,
			OptimizerTier:      flywheel.TierBronze,
			MentorTier:         flywheel.TierBronze,
			AgentWhispererTier: flywheel.TierBronze,
			ReliabilityIndex:   100,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"profile": scores})
}

// GetLeaderboardQuery handles GET /api/v1/flywheel/leaderboards?type=overall&timeframe=...
func (h *Handler) GetLeaderboardQuery(w http.ResponseWriter, r *http.Request) {
	scoreTypeStr := r.URL.Query().Get("type")
	if scoreTypeStr == "" {
		scoreTypeStr = "overall"
	}
	scoreType := flywheel.ReputationScoreType(scoreTypeStr)
	validTypes := map[flywheel.ReputationScoreType]bool{
		flywheel.ReputationScoreTypeOverall:        true,
		flywheel.ReputationScoreTypeBuilder:        true,
		flywheel.ReputationScoreTypeOptimizer:      true,
		flywheel.ReputationScoreTypeMentor:         true,
		flywheel.ReputationScoreTypeAgentWhisperer: true,
		flywheel.ReputationScoreTypeReliability:    true,
	}
	if !validTypes[scoreType] {
		http.Error(w, `{"error":"Invalid score type"}`, http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	scores, count, err := h.service.GetLeaderboard(r.Context(), scoreType, limit, offset)
	if err != nil {
		h.logger.WithError(err).Debug("Flywheel leaderboard table may not exist, returning empty")
		scores = nil
		count = 0
	}

	response := map[string]interface{}{
		"scores": scores,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeaderboard handles GET /api/v1/flywheel/leaderboards/:score_type
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scoreType := flywheel.ReputationScoreType(vars["score_type"])

	validTypes := map[flywheel.ReputationScoreType]bool{
		flywheel.ReputationScoreTypeOverall:        true,
		flywheel.ReputationScoreTypeBuilder:        true,
		flywheel.ReputationScoreTypeOptimizer:      true,
		flywheel.ReputationScoreTypeMentor:         true,
		flywheel.ReputationScoreTypeAgentWhisperer: true,
		flywheel.ReputationScoreTypeReliability:    true,
	}
	if !validTypes[scoreType] {
		http.Error(w, `{"error":"Invalid score type"}`, http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	scores, count, err := h.service.GetLeaderboard(r.Context(), scoreType, limit, offset)
	if err != nil {
		h.logger.WithError(err).Debug("Flywheel leaderboard table may not exist, returning empty")
		scores = nil
		count = 0
	}

	response := map[string]interface{}{
		"scores": scores,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
