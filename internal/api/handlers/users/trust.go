package users

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleGetUserTrust returns per-user trust score breakdown by averaging trust components
// across all of the user's published functions.
// GET /users/{username}/trust
func (h *Handler) HandleGetUserTrust(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := h.repo.GetUserByUsername(username)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	stats, err := h.repo.GetUserProfileStats(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load user stats")
		return
	}

	breakdown, err := h.repo.GetUserTrustBreakdown(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load trust breakdown")
		return
	}

	trustScore := 0
	if ts, ok := stats["trustScore"].(int); ok {
		trustScore = ts
	}

	writeJSON(w, http.StatusOK, UserTrustResponse{
		UserID:      user.ID,
		Username:    username,
		TrustScore:  trustScore,
		Components: UserTrustComponents{
			Reliability:  toFloat64(breakdown["reliability"]),
			Latency:      toFloat64(breakdown["latency"]),
			ErrorRate:    toFloat64(breakdown["error_rate"]),
			UserRating:   toFloat64(breakdown["user_rating"]),
			Verification: toFloat64(breakdown["verification"]),
		},
		Metrics: UserTrustMetrics{
			TotalCalls:        toInt(breakdown["total_calls"]),
			SuccessRate:       toFloat64(breakdown["success_rate"]),
			AvgP50LatencyMs:  toFloat64(breakdown["avg_p50_latency_ms"]),
			AvgP95LatencyMs:  toFloat64(breakdown["avg_p95_latency_ms"]),
		},
		FunctionsWithTrust: toInt(breakdown["functions_with_trust"]),
	})
}

// HandleGetMyTrust returns per-user trust breakdown for the authenticated user.
// GET /users/me/trust
func (h *Handler) HandleGetMyTrust(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	stats, err := h.repo.GetUserProfileStats(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load user stats")
		return
	}

	breakdown, err := h.repo.GetUserTrustBreakdown(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load trust breakdown")
		return
	}

	trustScore := 0
	if ts, ok := stats["trustScore"].(int); ok {
		trustScore = ts
	}

	writeJSON(w, http.StatusOK, UserTrustResponse{
		UserID:      user.ID,
		Username:    username,
		TrustScore:  trustScore,
		Components: UserTrustComponents{
			Reliability:  toFloat64(breakdown["reliability"]),
			Latency:      toFloat64(breakdown["latency"]),
			ErrorRate:    toFloat64(breakdown["error_rate"]),
			UserRating:   toFloat64(breakdown["user_rating"]),
			Verification: toFloat64(breakdown["verification"]),
		},
		Metrics: UserTrustMetrics{
			TotalCalls:        toInt(breakdown["total_calls"]),
			SuccessRate:       toFloat64(breakdown["success_rate"]),
			AvgP50LatencyMs:  toFloat64(breakdown["avg_p50_latency_ms"]),
			AvgP95LatencyMs:  toFloat64(breakdown["avg_p95_latency_ms"]),
		},
		FunctionsWithTrust: toInt(breakdown["functions_with_trust"]),
	})
}

// UserTrustResponse is the API response for user trust breakdown.
type UserTrustResponse struct {
	UserID               uuid.UUID           `json:"user_id"`
	Username             string             `json:"username"`
	TrustScore           int                `json:"trust_score"`
	Components           UserTrustComponents `json:"components"`
	Metrics             UserTrustMetrics    `json:"metrics"`
	FunctionsWithTrust   int                `json:"functions_with_trust"`
}

// UserTrustComponents are the individual trust score components (0-100).
type UserTrustComponents struct {
	Reliability  float64 `json:"reliability"`
	Latency      float64 `json:"latency"`
	ErrorRate    float64 `json:"error_rate"`
	UserRating  float64 `json:"user_rating"`
	Verification float64 `json:"verification"`
}

// UserTrustMetrics are execution metrics for the user's functions.
type UserTrustMetrics struct {
	TotalCalls       int     `json:"total_calls"`
	SuccessRate      float64 `json:"success_rate"`
	AvgP50LatencyMs float64 `json:"avg_p50_latency_ms"`
	AvgP95LatencyMs float64 `json:"avg_p95_latency_ms"`
}

// toFloat64 safely converts an interface{} to float64.
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
}

// toInt safely converts an interface{} to int.
func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
