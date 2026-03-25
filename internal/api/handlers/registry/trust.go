package registry

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ============================================
// Trust Score API Handlers
// ============================================

// HandleGetTrustScore handles GET /v1/registry/functions/{id}/trust
// Returns the current trust score and components for a function
func (h *Handler) HandleGetTrustScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Get the function first to verify it exists
	fn, err := h.repo.GetFunctionByID(functionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Get the latest trust history
	history, err := h.repo.GetLatestTrustHistory(functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get trust history")
		http.Error(w, "Failed to get trust score", http.StatusInternalServerError)
		return
	}

	// If no trust history exists, calculate it on-demand
	if history == nil {
		windowStart := time.Now().Add(-24 * time.Hour)
		history, err = h.repo.CalculateTrustScore(functionID, windowStart, time.Now())
		if err != nil {
			logrus.WithError(err).Error("Failed to calculate trust score")
			http.Error(w, "Failed to calculate trust score", http.StatusInternalServerError)
			return
		}
	}

	// Build response
	response := registry.TrustScoreResponse{
		FunctionID:        functionID,
		TrustScore:       history.TrustScore,
		TrustTier:        history.TrustTier,
		IsVerified:       history.IsVerified,
		VerificationLevel: history.VerificationLevel,
		LastUpdated:      history.CalculatedAt,
		WindowStart:      history.WindowStart,
		WindowEnd:        history.WindowEnd,
	}

	// Set component scores
	response.Components.Reliability = history.ReliabilityScore
	response.Components.Latency = history.LatencyScore
	response.Components.ErrorRate = history.ErrorRateScore
	response.Components.UserRating = history.UserRatingScore
	response.Components.Verification = history.VerificationBonus

	// Set metrics
	response.Metrics.TotalCalls = history.TotalCalls
	response.Metrics.SuccessRate = history.SuccessRate
	response.Metrics.P50LatencyMs = history.P50LatencyMs
	response.Metrics.P95LatencyMs = history.P95LatencyMs
	response.Metrics.P99LatencyMs = history.P99LatencyMs
	response.Metrics.ErrorRate = history.ErrorRate
	response.Metrics.TimeoutRate = history.TimeoutRate

	// Set diversity
	response.Diversity.Consumers = history.ConsumerDiversity
	response.Diversity.Tenants = history.TenantDiversity
	response.Diversity.Users = history.UserDiversity

	// Add function info for context
	_ = fn // fn is used to verify function exists

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTrustHistory handles GET /v1/registry/functions/{id}/trust/history
// Returns the trust score history for a function
func (h *Handler) HandleGetTrustHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Get trust history
	history, total, err := h.repo.GetTrustHistory(functionID, pageSize, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get trust history")
		http.Error(w, "Failed to get trust history", http.StatusInternalServerError)
		return
	}

	response := registry.TrustHistoryResponse{
		FunctionID:  functionID,
		History:    history,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRefreshTrustScore handles POST /v1/registry/functions/{id}/trust/refresh
// Forces a recalculation of the trust score
func (h *Handler) HandleRefreshTrustScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Verify function exists
	fn, err := h.repo.GetFunctionByID(functionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Recalculate trust score
	if err := h.repo.RecalculateTrustScore(functionID); err != nil {
		logrus.WithError(err).Error("Failed to recalculate trust score")
		http.Error(w, "Failed to recalculate trust score", http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	h.repo.InvalidateTrustScoreCache(functionID)

	// Get updated trust score
	history, err := h.repo.GetLatestTrustHistory(functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get updated trust history")
		http.Error(w, "Trust score recalculated but failed to fetch result", http.StatusInternalServerError)
		return
	}

	_ = fn // fn is used to verify function exists

	response := map[string]interface{}{
		"ok":             true,
		"message":        "Trust score refreshed successfully",
		"function_id":    functionID,
		"trust_score":    history.TrustScore,
		"trust_tier":     history.TrustTier,
		"calculated_at":  history.CalculatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunctionTrustByAuthorName handles GET /v1/registry/functions/{author}/{name}/trust
// Alternative endpoint using author/name instead of function ID
func (h *Handler) HandleGetFunctionTrustByAuthorName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Get the latest trust history
	history, err := h.repo.GetLatestTrustHistory(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get trust history")
		http.Error(w, "Failed to get trust score", http.StatusInternalServerError)
		return
	}

	// If no trust history exists, calculate it on-demand
	if history == nil {
		windowStart := time.Now().Add(-24 * time.Hour)
		history, err = h.repo.CalculateTrustScore(fn.ID, windowStart, time.Now())
		if err != nil {
			logrus.WithError(err).Error("Failed to calculate trust score")
			http.Error(w, "Failed to calculate trust score", http.StatusInternalServerError)
			return
		}
	}

	// Build response
	response := registry.TrustScoreResponse{
		FunctionID:        fn.ID,
		TrustScore:       history.TrustScore,
		TrustTier:        history.TrustTier,
		IsVerified:       history.IsVerified,
		VerificationLevel: history.VerificationLevel,
		LastUpdated:      history.CalculatedAt,
		WindowStart:      history.WindowStart,
		WindowEnd:        history.WindowEnd,
	}

	// Set component scores
	response.Components.Reliability = history.ReliabilityScore
	response.Components.Latency = history.LatencyScore
	response.Components.ErrorRate = history.ErrorRateScore
	response.Components.UserRating = history.UserRatingScore
	response.Components.Verification = history.VerificationBonus

	// Set metrics
	response.Metrics.TotalCalls = history.TotalCalls
	response.Metrics.SuccessRate = history.SuccessRate
	response.Metrics.P50LatencyMs = history.P50LatencyMs
	response.Metrics.P95LatencyMs = history.P95LatencyMs
	response.Metrics.P99LatencyMs = history.P99LatencyMs
	response.Metrics.ErrorRate = history.ErrorRate
	response.Metrics.TimeoutRate = history.TimeoutRate

	// Set diversity
	response.Diversity.Consumers = history.ConsumerDiversity
	response.Diversity.Tenants = history.TenantDiversity
	response.Diversity.Users = history.UserDiversity

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRefreshAllTrustScores handles POST /v1/admin/trust/refresh-all
// Forces a full recalculation of all trust scores (admin only)
func (h *Handler) HandleRefreshAllTrustScores(w http.ResponseWriter, r *http.Request) {
	// This should be called by an admin or system process
	// In production, you would add admin authentication check here

	job, err := h.repo.RefreshAllTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh all trust scores")
		http.Error(w, "Failed to refresh trust scores", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ok":                  true,
		"message":             "Trust score refresh job started",
		"job_id":             job.ID,
		"job_type":           job.JobType,
		"status":             job.Status,
		"functions_total":    job.FunctionsTotal,
		"functions_processed": job.FunctionsProcessed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================
// Trust Score Calculation Helper
// ============================================

// GetTrustScoreComponents calculates individual trust score components
func (h *Handler) GetTrustScoreComponents(functionID uuid.UUID) (*registry.TrustScoreResponse, error) {
	windowStart := time.Now().Add(-24 * time.Hour)
	history, err := h.repo.CalculateTrustScore(functionID, windowStart, time.Now())
	if err != nil {
		return nil, err
	}

	response := &registry.TrustScoreResponse{
		FunctionID:        functionID,
		TrustScore:       history.TrustScore,
		TrustTier:        history.TrustTier,
		IsVerified:       history.IsVerified,
		VerificationLevel: history.VerificationLevel,
		LastUpdated:      history.CalculatedAt,
		WindowStart:      history.WindowStart,
		WindowEnd:        history.WindowEnd,
	}

	response.Components.Reliability = history.ReliabilityScore
	response.Components.Latency = history.LatencyScore
	response.Components.ErrorRate = history.ErrorRateScore
	response.Components.UserRating = history.UserRatingScore
	response.Components.Verification = history.VerificationBonus

	response.Metrics.TotalCalls = history.TotalCalls
	response.Metrics.SuccessRate = history.SuccessRate
	response.Metrics.P50LatencyMs = history.P50LatencyMs
	response.Metrics.P95LatencyMs = history.P95LatencyMs
	response.Metrics.P99LatencyMs = history.P99LatencyMs
	response.Metrics.ErrorRate = history.ErrorRate
	response.Metrics.TimeoutRate = history.TimeoutRate

	response.Diversity.Consumers = history.ConsumerDiversity
	response.Diversity.Tenants = history.TenantDiversity
	response.Diversity.Users = history.UserDiversity

	return response, nil
}

// ============================================
// Internal Helper Methods
// ============================================

// getTrustTierFromScore converts a trust score to a trust tier
func getTrustTierFromScore(score float64, isVerified bool) registry.TrustTier {
	if score >= 90 {
		if isVerified {
			return registry.TrustTierHighlyTrusted
		}
		return registry.TrustTierVerified
	} else if score >= 70 {
		return registry.TrustTierVerified
	} else if score >= 50 {
		return registry.TrustTierTrusted
	}
	return registry.TrustTierUntrusted
}
