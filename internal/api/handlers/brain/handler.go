package brain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	brainEngine "github.com/functionfly/functionfly/internal/agent/brain"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type embeddingRequest struct {
	Text string `json:"text"`
}

type embeddingResponse struct {
	Embedding  []float64 `json:"embedding"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
}

type EmbeddingServiceClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewEmbeddingServiceClient() *EmbeddingServiceClient {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		baseURL = ""
	}
	return &EmbeddingServiceClient{
		baseURL: baseURL,
		apiKey:  os.Getenv("AI_SERVICE_API_KEY"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *EmbeddingServiceClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("AI_SERVICE_URL not configured")
	}

	url := fmt.Sprintf("%s/api/embed", c.baseURL)

	reqBody := embeddingRequest{Text: text}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var embedResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embedding := make([]float32, len(embedResp.Embedding))
	for i, v := range embedResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

type Handler struct {
	repo            *storage.BrainRepository
	contextBuilder  *brainEngine.ContextBuilder
	embedClient    *EmbeddingServiceClient
	logger          *logrus.Logger
}

func NewHandler(repo *storage.BrainRepository, logger *logrus.Logger) *Handler {
	return &Handler{
		repo:           repo,
		contextBuilder: brainEngine.NewContextBuilder(repo),
		embedClient:    NewEmbeddingServiceClient(),
		logger:         logger,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"code":    code,
		"message": message,
	})
}

func (h *Handler) getTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var tier string
	err := h.repo.GetDB().QueryRowContext(ctx, `
		SELECT COALESCE(pt.name, 'free')
		FROM subscriptions s
		JOIN pricing_tiers pt ON pt.id = s.pricing_tier_id
		WHERE s.tenant_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC LIMIT 1`, tenantID).Scan(&tier)
	if err != nil {
		return plans.PlanFree, err
	}
	return tier, nil
}

// HandleListSignals returns brain signals for the tenant
func (h *Handler) HandleListSignals(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	params := storage.SignalListParams{
		TenantID:      claims.TenantID,
		ConnectorSlug: r.URL.Query().Get("connector"),
		SignalType:    r.URL.Query().Get("type"),
		Limit:         limit,
		Offset:        offset,
		SortBy:        r.URL.Query().Get("sort"),
	}

	signals, total, err := h.repo.ListSignals(r.Context(), params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list signals")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list signals")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"signals": signals,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleSearchSignals performs semantic search on signals (Pro+)
func (h *Handler) HandleSearchSignals(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	tier, _ := h.getTenantPlan(r.Context(), claims.TenantID)
	limits := plans.GetBrainConnectorLimits(tier)
	if !limits.SemanticSearch {
		h.respondError(w, 403, "TIER_RESTRICTED", "Semantic search requires Pro or Enterprise plan")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondError(w, 400, "MISSING_FIELD", "Query parameter 'q' is required")
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	embedding, err := h.embedClient.GenerateEmbedding(r.Context(), query)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to generate embedding, falling back to recent signals")
		signals, err := h.repo.GetRecentSignals(r.Context(), claims.TenantID, 7, limit)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get recent signals")
			h.respondError(w, 500, "INTERNAL_ERROR", "Failed to search signals")
			return
		}
		results := make([]map[string]interface{}, len(signals))
		for i, s := range signals {
			results[i] = map[string]interface{}{
				"signal": s,
				"score":  1.0,
			}
		}
		h.respondJSON(w, 200, map[string]interface{}{
			"results":       results,
			"query":         query,
			"search_type":   "recent",
		})
		return
	}

	searchResults, err := h.repo.SemanticSearch(r.Context(), claims.TenantID, embedding, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search signals")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to search signals")
		return
	}

	results := make([]map[string]interface{}, len(searchResults))
	for i, sr := range searchResults {
		results[i] = map[string]interface{}{
			"signal": sr.Signal,
			"score":  sr.Score,
		}
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"results":     results,
		"query":       query,
		"search_type": "semantic",
	})
}

// HandleDeleteSignal deletes a specific signal
func (h *Handler) HandleDeleteSignal(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	signalID, err := uuid.Parse(vars["signal_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid signal ID")
		return
	}

	if err := h.repo.DeleteSignal(r.Context(), claims.TenantID, signalID); err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Signal not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "Signal deleted",
	})
}

// HandlePurgeSignals clears all brain signals for the tenant
func (h *Handler) HandlePurgeSignals(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	if err := h.repo.PurgeSignals(r.Context(), claims.TenantID); err != nil {
		h.logger.WithError(err).Error("Failed to purge signals")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to purge signals")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "All brain signals purged",
	})
}

// HandleGetStats returns brain usage stats
func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	stats, err := h.repo.GetBrainStats(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get brain stats")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to get stats")
		return
	}

	tier, _ := h.getTenantPlan(r.Context(), claims.TenantID)
	limits := plans.GetBrainConnectorLimits(tier)
	stats.MemoryMax = limits.MaxSignals
	stats.RetentionDays = limits.SignalRetentionDays

	h.respondJSON(w, 200, stats)
}

// HandleSubmitFeedback submits feedback on signal usefulness
func (h *Handler) HandleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req storage.BrainFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.SignalID == uuid.Nil {
		h.respondError(w, 400, "MISSING_FIELD", "signal_id is required")
		return
	}

	if err := h.repo.SaveFeedback(r.Context(), claims.TenantID, &req); err != nil {
		h.logger.WithError(err).Error("Failed to save feedback")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to save feedback")
		return
	}

	h.respondJSON(w, 201, map[string]interface{}{
		"message": "Feedback submitted",
	})
}

// HandleListComposers returns brain composers for the tenant
func (h *Handler) HandleListComposers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	composers, err := h.repo.ListComposers(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list composers")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list composers")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"composers": composers,
	})
}

// HandleCreateComposer creates a new brain composer
func (h *Handler) HandleCreateComposer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	tier, _ := h.getTenantPlan(r.Context(), claims.TenantID)
	limits := plans.GetBrainConnectorLimits(tier)
	if !limits.BrainComposer {
		h.respondError(w, 403, "TIER_RESTRICTED", "Brain Composer requires Pro or Enterprise plan")
		return
	}

	var composer storage.BrainComposer
	if err := json.NewDecoder(r.Body).Decode(&composer); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	composer.TenantID = claims.TenantID
	if composer.Schedule == "" {
		composer.Schedule = "0 8 * * *"
	}
	if composer.OutputFormat == "" {
		composer.OutputFormat = "briefing"
	}

	created, err := h.repo.CreateComposer(r.Context(), &composer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create composer")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to create composer")
		return
	}

	h.respondJSON(w, 201, map[string]interface{}{
		"composer": created,
	})
}

// HandleDeleteComposer deletes a brain composer
func (h *Handler) HandleDeleteComposer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	composerID, err := uuid.Parse(vars["composer_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid composer ID")
		return
	}

	if err := h.repo.DeleteComposer(r.Context(), claims.TenantID, composerID); err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Composer not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "Composer deleted",
	})
}

// HandleListTriggers returns brain triggers for the tenant
func (h *Handler) HandleListTriggers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	triggers, err := h.repo.ListTriggers(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list triggers")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list triggers")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"triggers": triggers,
	})
}

// HandleCreateTrigger creates a new brain trigger
func (h *Handler) HandleCreateTrigger(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	tier, _ := h.getTenantPlan(r.Context(), claims.TenantID)
	limits := plans.GetBrainConnectorLimits(tier)
	if !limits.BrainTriggerDaemon {
		h.respondError(w, 403, "TIER_RESTRICTED", "Brain triggers require Pro or Enterprise plan")
		return
	}

	var trigger storage.BrainTrigger
	if err := json.NewDecoder(r.Body).Decode(&trigger); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	trigger.TenantID = claims.TenantID
	if trigger.Schedule == "" {
		trigger.Schedule = "immediate"
	}
	if trigger.Action == "" {
		trigger.Action = "run_agent"
	}

	created, err := h.repo.CreateTrigger(r.Context(), &trigger)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create trigger")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to create trigger")
		return
	}

	h.respondJSON(w, 201, map[string]interface{}{
		"trigger": created,
	})
}

// HandleDeleteTrigger deletes a brain trigger
func (h *Handler) HandleDeleteTrigger(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	triggerID, err := uuid.Parse(vars["trigger_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid trigger ID")
		return
	}

	if err := h.repo.DeleteTrigger(r.Context(), claims.TenantID, triggerID); err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Trigger not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "Trigger deleted",
	})
}
