package recommendations

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/recommendations"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles recommendation API requests
type Handler struct {
	service *recommendations.Service
}

// NewHandler creates a new recommendations handler
func NewHandler(service *recommendations.Service) *Handler {
	return &Handler{service: service}
}

// HandleGetRecommendations handles getting recommendations
// GET /v1/recommendations
// Query params:
//   - function_id: UUID of the function to get recommendations for
//   - category: Category to get recommendations for
//   - q: Search query for use case-based recommendations
//   - user_id: UUID of the user for personalized recommendations
//   - types: Comma-separated list of recommendation types (similar,frequently_used_together,same_category,trending,personalized)
//   - limit: Number of recommendations (default 10, max 50)
//   - offset: Pagination offset
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	var functionID *uuid.UUID
	if fid := r.URL.Query().Get("function_id"); fid != "" {
		id, err := uuid.Parse(fid)
		if err != nil {
			http.Error(w, "Invalid function_id format", http.StatusBadRequest)
			return
		}
		functionID = &id
	}

	var category *string
	if cat := r.URL.Query().Get("category"); cat != "" {
		category = &cat
	}

	var query *string
	if q := r.URL.Query().Get("q"); q != "" {
		query = &q
	}

	var userID *uuid.UUID
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err != nil {
			http.Error(w, "Invalid user_id format", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var types []recommendations.RecommendationType
	if typesParam := r.URL.Query().Get("types"); typesParam != "" {
		for _, t := range strings.Split(typesParam, ",") {
			t = strings.TrimSpace(t)
			switch t {
			case "similar":
				types = append(types, recommendations.RecommendationTypeSimilar)
			case "frequently_used_together":
				types = append(types, recommendations.RecommendationTypeFrequentlyUsedTogether)
			case "same_category":
				types = append(types, recommendations.RecommendationTypeSameCategory)
			case "trending":
				types = append(types, recommendations.RecommendationTypeTrending)
			case "personalized":
				types = append(types, recommendations.RecommendationTypePersonalized)
			}
		}
	}

	includePersonalized := r.URL.Query().Get("include_personalized") == "true"

	req := &recommendations.RecommendationRequest{
		FunctionID:          functionID,
		UserID:              userID,
		Category:            category,
		Query:               query,
		Limit:               limit,
		Offset:              offset,
		Types:               types,
		IncludePersonalized: includePersonalized,
	}

	result, err := h.service.GetRecommendations(r.Context(), req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get recommendations")
		http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleGetRelatedFunctions handles getting related functions for a specific function
// GET /v1/registry/functions/{author}/{name}/related
// Query params:
//   - limit: Number of related functions (default 10, max 20)
func (h *Handler) HandleGetRelatedFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	if author == "" || name == "" {
		http.Error(w, "Author and name are required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	fn, err := h.service.GetRegistryRepository().GetFunctionByAuthorName(author, name)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"author": author, "name": name}).Warn("Function not found for related functions lookup")
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	results, err := h.service.GetRelatedFunctions(r.Context(), fn.ID, limit)
	if err != nil {
		logrus.WithError(err).WithField("function_id", fn.ID).Error("Failed to get related functions")
		http.Error(w, "Failed to get related functions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": results,
		"total":           len(results),
		"function": map[string]string{
			"author": author,
			"name":   name,
		},
	})
}

// HandleRecordInteraction records a user interaction
// POST /v1/recommendations/interactions
func (h *Handler) HandleRecordInteraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		UserID              *string `json:"user_id"`
		FunctionID          string  `json:"function_id"`
		InteractionType    string  `json:"interaction_type"`
		SessionID           *string `json:"session_id"`
		ReferrerFunctionID *string `json:"referrer_function_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		http.Error(w, "Invalid function_id format", http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		id, err := uuid.Parse(*req.UserID)
		if err == nil {
			userID = &id
		}
	}

	var sessionID *string
	if req.SessionID != nil && *req.SessionID != "" {
		sessionID = req.SessionID
	}

	var referrerFunctionID *uuid.UUID
	if req.ReferrerFunctionID != nil && *req.ReferrerFunctionID != "" {
		id, err := uuid.Parse(*req.ReferrerFunctionID)
		if err == nil {
			referrerFunctionID = &id
		}
	}

	interactionType := recommendations.InteractionType(req.InteractionType)
	switch interactionType {
	case recommendations.InteractionTypeView,
		recommendations.InteractionTypeExecute,
		recommendations.InteractionTypeSave,
		recommendations.InteractionTypeFollow,
		recommendations.InteractionTypeRate,
		recommendations.InteractionTypeCopyCode,
		recommendations.InteractionTypeShare:
		// Valid types
	default:
		http.Error(w, "Invalid interaction_type", http.StatusBadRequest)
		return
	}

	err = h.service.RecordInteraction(r.Context(), userID, functionID, interactionType, sessionID, referrerFunctionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to record interaction")
		http.Error(w, "Failed to record interaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// HandleRecordExecution records a function execution
// POST /v1/recommendations/executions
func (h *Handler) HandleRecordExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		UserID    *string `json:"user_id"`
		FunctionID string `json:"function_id"`
		SessionID string  `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		http.Error(w, "Invalid function_id format", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		id, err := uuid.Parse(*req.UserID)
		if err == nil {
			userID = &id
		}
	}

	err = h.service.RecordExecution(r.Context(), userID, functionID, req.SessionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to record execution")
		http.Error(w, "Failed to record execution", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// HandleRecordFeedback records recommendation feedback
// POST /v1/recommendations/feedback
func (h *Handler) HandleRecordFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		UserID                *string `json:"user_id"`
		FunctionID            string  `json:"function_id"`
		RecommendedFunctionID string  `json:"recommended_function_id"`
		FeedbackType         string  `json:"feedback_type"` // clicked, executed, dismissed, not_relevant, helpful
		RecommendationType   *string `json:"recommendation_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	functionID, err := uuid.Parse(req.FunctionID)
	if err != nil {
		http.Error(w, "Invalid function_id format", http.StatusBadRequest)
		return
	}

	recommendedFunctionID, err := uuid.Parse(req.RecommendedFunctionID)
	if err != nil {
		http.Error(w, "Invalid recommended_function_id format", http.StatusBadRequest)
		return
	}

	// Validate feedback type
	switch req.FeedbackType {
	case "clicked", "executed", "dismissed", "not_relevant", "helpful":
		// Valid types
	default:
		http.Error(w, "Invalid feedback_type", http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != "" {
		id, err := uuid.Parse(*req.UserID)
		if err == nil {
			userID = &id
		}
	}

	var recommendationType *string
	if req.RecommendationType != nil {
		recommendationType = req.RecommendationType
	}

	err = h.service.RecordFeedback(r.Context(), userID, functionID, recommendedFunctionID, req.FeedbackType, recommendationType)
	if err != nil {
		logrus.WithError(err).Error("Failed to record feedback")
		http.Error(w, "Failed to record feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// HandleRefreshRecommendations refreshes all recommendations
// POST /v1/recommendations/refresh (admin only - for now just a simple endpoint)
func (h *Handler) HandleRefreshRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Protected by RequirePermission(auth.PermSystemWrite) in routes — admin only.

	err := h.service.RefreshAllRecommendations(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh recommendations")
		http.Error(w, "Failed to refresh recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "refreshed"})
}

// HandleTripleSearch handles triple-vector search requests
// POST /v1/recommendations/triple-search
func (h *Handler) HandleTripleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
		Weights *struct {
			Contract float64 `json:"contract"`
			Semantic float64 `json:"semantic"`
			Code     float64 `json:"code"`
		} `json:"weights,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var weights *recommendations.TripleSearchWeights
	if req.Weights != nil {
		weights = &recommendations.TripleSearchWeights{
			Contract: req.Weights.Contract,
			Semantic: req.Weights.Semantic,
			Code:     req.Weights.Code,
		}
	}

	results, err := h.service.SearchByTripleEmbedding(r.Context(), req.Query, weights, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed triple search")
		http.Error(w, "Failed to perform triple search", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": results,
		"total":          len(results),
		"query":          req.Query,
	})
}

// HandleFindComposable finds functions composable with the target
// GET /v1/recommendations/composable/{function_id}
func (h *Handler) HandleFindComposable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["function_id"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function_id format", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	results, err := h.service.FindComposableFunctions(r.Context(), functionID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to find composable functions")
		http.Error(w, "Failed to find composable functions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": results,
		"total":          len(results),
		"function_id":    functionIDStr,
	})
}
