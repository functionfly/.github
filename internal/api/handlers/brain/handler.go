package brain

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	brainEngine "github.com/functionfly/functionfly/internal/agent/brain"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo            *storage.BrainRepository
	contextBuilder  *brainEngine.ContextBuilder
	logger          *logrus.Logger
}

func NewHandler(repo *storage.BrainRepository, logger *logrus.Logger) *Handler {
	return &Handler{
		repo:           repo,
		contextBuilder: brainEngine.NewContextBuilder(repo),
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

	// In production, generate embedding from query text using an embedding model
	// For now, return recent signals as a fallback
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	signals, err := h.repo.GetRecentSignals(r.Context(), claims.TenantID, 7, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search signals")
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
		"results": results,
		"query":   query,
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
