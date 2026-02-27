package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetFunctionStats handles getting function statistics
func (h *Handler) HandleGetFunctionStats(w http.ResponseWriter, r *http.Request) {
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

	// Get ratings
	rating, err := h.repo.GetOrCreateRating(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get rating")
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	// Get recent stats (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	totalCalls, successRate, avgLatency, p95Latency, err := h.repo.GetFunctionStats(fn.ID, since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function stats")
	}

	response := map[string]interface{}{
		"function_id":       fn.ID,
		"author":            author,
		"name":              name,
		"total_calls":       totalCalls,
		"success_rate":      successRate,
		"avg_latency_ms":    avgLatency,
		"p95_latency_ms":    p95Latency,
		"reliability_score": rating.ReliabilityScore,
		"latency_score":     rating.LatencyScore,
		"overall_score":     rating.OverallScore,
		"total_ratings":     rating.TotalRatings,
		"popularity_score":  fn.PopularityScore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSubmitRating handles submitting a rating for a function
func (h *Handler) HandleSubmitRating(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var ratingReq struct {
		OverallScore       float64 `json:"overall_score"`
		ReliabilityScore   float64 `json:"reliability_score"`
		LatencyScore       float64 `json:"latency_score"`
		DocumentationScore float64 `json:"documentation_score"`
	}

	if err := json.Unmarshal(body, &ratingReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	rating, err := h.repo.GetOrCreateRating(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get rating")
		http.Error(w, "Failed to submit rating", http.StatusInternalServerError)
		return
	}

	// Update ratings (weighted average with existing)
	if ratingReq.OverallScore > 0 {
		if rating.TotalRatings == 0 {
			rating.OverallScore = ratingReq.OverallScore
		} else {
			rating.OverallScore = (rating.OverallScore*float64(rating.TotalRatings) + ratingReq.OverallScore) / float64(rating.TotalRatings+1)
		}
	}
	if ratingReq.ReliabilityScore > 0 {
		if rating.TotalRatings == 0 {
			rating.ReliabilityScore = ratingReq.ReliabilityScore
		} else {
			rating.ReliabilityScore = (rating.ReliabilityScore*float64(rating.TotalRatings) + ratingReq.ReliabilityScore) / float64(rating.TotalRatings+1)
		}
	}
	if ratingReq.LatencyScore > 0 {
		if rating.TotalRatings == 0 {
			rating.LatencyScore = ratingReq.LatencyScore
		} else {
			rating.LatencyScore = (rating.LatencyScore*float64(rating.TotalRatings) + ratingReq.LatencyScore) / float64(rating.TotalRatings+1)
		}
	}
	if ratingReq.DocumentationScore > 0 {
		if rating.TotalRatings == 0 {
			rating.DocumentationScore = ratingReq.DocumentationScore
		} else {
			rating.DocumentationScore = (rating.DocumentationScore*float64(rating.TotalRatings) + ratingReq.DocumentationScore) / float64(rating.TotalRatings+1)
		}
	}

	rating.TotalRatings++

	if err := h.repo.UpdateRating(rating); err != nil {
		logrus.WithError(err).Error("Failed to update rating")
		http.Error(w, "Failed to submit rating", http.StatusInternalServerError)
		return
	}

	// Broadcast rating update to all connected clients
	h.realtimeMonitor.BroadcastRegistryUpdate(fn.ID, "rating", map[string]interface{}{
		"function_name":       fmt.Sprintf("%s/%s", author, name),
		"overall_score":       rating.OverallScore,
		"reliability_score":   rating.ReliabilityScore,
		"latency_score":       rating.LatencyScore,
		"documentation_score": rating.DocumentationScore,
		"total_ratings":       rating.TotalRatings,
		"user_id":             user.UserID,
	})

	response := map[string]interface{}{
		"ok":      true,
		"message": "Rating submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAggregateStats handles forcing stats aggregation
func (h *Handler) HandleAggregateStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if err := h.updateStats(fn.ID); err != nil {
		logrus.WithError(err).Error("Failed to update stats")
		http.Error(w, "Failed to update stats", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ok":      true,
		"message": "Stats aggregated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateStats recalculates and updates function statistics
func (h *Handler) updateStats(functionID uuid.UUID) error {
	// Get stats for last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	totalCalls, successRate, avgLatency, p95Latency, err := h.repo.GetFunctionStats(functionID, since)
	if err != nil {
		return fmt.Errorf("failed to get function stats: %w", err)
	}

	// Calculate popularity score based on recent activity and success rate
	// Base score is total calls, weighted by success rate (as percentage)
	popularityScore := int(float64(totalCalls) * (successRate / 100.0))

	// Update the function's popularity score
	if err := h.repo.UpdateFunctionPopularity(functionID, popularityScore); err != nil {
		return fmt.Errorf("failed to update popularity score: %w", err)
	}

	// Update ratings as well (similar to StatsAggregator)
	rating, err := h.repo.GetOrCreateRating(functionID)
	if err != nil {
		return fmt.Errorf("failed to get rating: %w", err)
	}

	rating.SuccessRate = successRate
	rating.AvgLatencyMs = avgLatency
	rating.P95LatencyMs = p95Latency
	rating.TotalRatings = totalCalls

	// Calculate reliability score (based on success rate)
	rating.ReliabilityScore = successRate

	// Calculate latency score (inverse - lower latency = higher score)
	if p95Latency < 50 {
		rating.LatencyScore = 100
	} else if p95Latency < 200 {
		rating.LatencyScore = 80
	} else if p95Latency < 500 {
		rating.LatencyScore = 60
	} else if p95Latency < 1000 {
		rating.LatencyScore = 40
	} else {
		rating.LatencyScore = 20
	}

	// Calculate overall score (weighted average)
	rating.OverallScore = (rating.ReliabilityScore + rating.LatencyScore + rating.DocumentationScore) / 3

	if err := h.repo.UpdateRating(rating); err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	return nil
}
