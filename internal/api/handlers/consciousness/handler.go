package consciousness

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/consciousness"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo        *consciousness.Repository
	auditRepo   storage.Repository
	engine      *consciousness.Engine
	logger      *logrus.Logger
	engineOwner bool
}

func NewHandler(db *sql.DB, auditRepo storage.Repository, logger *logrus.Logger) *Handler {
	return &Handler{
		repo:     consciousness.NewRepository(db, logger),
		auditRepo: auditRepo,
		engine:   consciousness.NewEngine(db, logger),
		logger:   logger,
	}
}

func NewHandlerWithEngine(db *sql.DB, logger *logrus.Logger, engine *consciousness.Engine) *Handler {
	return &Handler{
		repo:        consciousness.NewRepository(db, logger),
		engine:      engine,
		logger:      logger,
		engineOwner: false,
	}
}

// RegisterRoutes registers consciousness API routes on the given router.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/consciousness").Subrouter()
	s.HandleFunc("/score", h.GetAwarenessScore).Methods("GET")
	s.HandleFunc("/insights", h.ListInsights).Methods("GET")
	s.HandleFunc("/insights/{id}", h.GetInsight).Methods("GET")
	s.HandleFunc("/insights/{id}", h.DeleteInsight).Methods("DELETE")
	s.HandleFunc("/insights/{id}/dismiss", h.DismissInsight).Methods("POST")
	s.HandleFunc("/insights/{id}/apply", h.ApplyAction).Methods("POST")
	s.HandleFunc("/preferences", h.GetPreferences).Methods("GET")
	s.HandleFunc("/preferences", h.UpdatePreferences).Methods("PUT")
	s.HandleFunc("/run", h.RunAnalysis).Methods("POST")
	s.HandleFunc("/export", h.ExportData).Methods("GET")
}

// GetAwarenessScore returns the System Awareness Score for the tenant.
func (h *Handler) GetAwarenessScore(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	score, err := h.repo.GetScore(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get awareness score")
		writeError(w, http.StatusInternalServerError, "failed to get score")
		return
	}
	if score == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"score":   nil,
			"message": "No score computed yet. The next analysis run will generate your score.",
		})
		return
	}

	label := consciousness.ScoreLabel(score.OverallScore)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"score": score.OverallScore,
		"label": label,
		"components": map[string]float64{
			"health":       score.HealthScore,
			"efficiency":   score.EfficiencyScore,
			"scalability":  score.ScalabilityScore,
			"reliability":  score.ReliabilityScore,
			"optimization": score.OptimizationScore,
		},
		"trend":             score.Trend,
		"previousScore":     score.PreviousScore,
		"functionsAnalyzed": score.FunctionsAnalyzed,
		"activeInsights":    score.ActiveInsights,
		"criticalInsights":  score.CriticalInsights,
		"computedAt":        score.ComputedAt,
	})
}

// ListInsights returns filtered insights for the tenant.
func (h *Handler) ListInsights(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	params := consciousness.ListInsightsParams{
		TenantID: tenantID,
		Limit:    20,
	}

	if v := r.URL.Query().Get("category"); v != "" {
		cat := consciousness.InsightCategory(v)
		params.Category = &cat
	}
	if v := r.URL.Query().Get("severity"); v != "" {
		sev := consciousness.InsightSeverity(v)
		params.Severity = &sev
	}
	if v := r.URL.Query().Get("status"); v != "" {
		st := consciousness.InsightStatus(v)
		params.Status = &st
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			params.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			params.Offset = n
		}
	}

	insights, total, err := h.repo.ListInsights(r.Context(), params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list insights")
		writeError(w, http.StatusInternalServerError, "failed to list insights")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"insights": insights,
		"total":    total,
		"limit":    params.Limit,
		"offset":   params.Offset,
	})
}

// GetInsight returns a single insight by ID.
func (h *Handler) GetInsight(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r)
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid insight ID")
		return
	}

	insight, err := h.repo.GetInsight(r.Context(), id, tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get insight")
		writeError(w, http.StatusInternalServerError, "failed to get insight")
		return
	}
	if insight == nil {
		writeError(w, http.StatusNotFound, "insight not found")
		return
	}

	writeJSON(w, http.StatusOK, insight)
}

// DismissInsight marks an insight as dismissed.
func (h *Handler) DismissInsight(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r)
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid insight ID")
		return
	}

	if err := h.repo.DismissInsight(r.Context(), id, tenantID); err != nil {
		if errors.Is(err, consciousness.ErrInsightNotFound) {
			writeError(w, http.StatusNotFound, "insight not found or already dismissed")
			return
		}
		h.logger.WithError(err).Error("Failed to dismiss insight")
		writeError(w, http.StatusInternalServerError, "failed to dismiss insight")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}

// ApplyAction applies the suggested action for an insight.
func (h *Handler) ApplyAction(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r)
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid insight ID")
		return
	}

	insight, err := h.repo.GetInsight(r.Context(), id, tenantID)
	if err != nil || insight == nil {
		writeError(w, http.StatusNotFound, "insight not found")
		return
	}

	if insight.ActionType == consciousness.ActionNone {
		writeError(w, http.StatusBadRequest, "this insight has no actionable recommendation")
		return
	}

	var req struct {
		DryRun bool `json:"dry_run"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.DryRun {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"dry_run":    true,
			"action_type": insight.ActionType,
			"preview":    insight.ActionPreview,
			"message":    "This is a preview. No changes were made.",
		})
		return
	}

	if err := h.repo.ApplyInsight(r.Context(), id, tenantID); err != nil {
		if errors.Is(err, consciousness.ErrInsightNotFound) {
			writeError(w, http.StatusNotFound, "insight not found or already applied")
			return
		}
		h.logger.WithError(err).Error("Failed to apply insight")
		writeError(w, http.StatusInternalServerError, "failed to apply action")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "applied",
		"action_type": insight.ActionType,
		"message":    "Action has been applied. Changes will take effect shortly.",
	})
}

// GetPreferences returns the tenant's consciousness preferences.
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	prefs, err := h.repo.GetPreferences(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get preferences")
		writeError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

// UpdatePreferences updates the tenant's consciousness preferences.
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	var prefs consciousness.Preferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	prefs.TenantID = tenantID

	if err := h.repo.UpsertPreferences(r.Context(), &prefs); err != nil {
		h.logger.WithError(err).Error("Failed to update preferences")
		writeError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

// RunAnalysis triggers an immediate analysis for the tenant.
func (h *Handler) RunAnalysis(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	lookback := 7
	plan := middleware.GetTenantPlan(r)
	switch plan {
	case "enterprise":
		lookback = 30
	case "agent_enterprise":
		lookback = 90
	}

	params := consciousness.AnalysisParams{
		LookbackDays: lookback,
	}

	result, err := h.engine.AnalyzeTenant(r.Context(), tenantID, params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to run analysis")
		writeError(w, http.StatusInternalServerError, "analysis failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"insights_generated": len(result.Insights),
		"score":              result.Score,
		"duration_ms":        result.DurationMs,
	})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
// DeleteInsight permanently deletes an insight (GDPR Article 17 - Right to Erasure).
func (h *Handler) DeleteInsight(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r)
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid insight ID")
		return
	}

	// Get insight before deletion for audit logging
	insight, err := h.repo.GetInsight(r.Context(), id, tenantID)
	if err != nil || insight == nil {
		writeError(w, http.StatusNotFound, "insight not found")
		return
	}

	if err := h.repo.DeleteInsight(r.Context(), id, tenantID); err != nil {
		if errors.Is(err, consciousness.ErrInsightNotFound) {
			writeError(w, http.StatusNotFound, "insight not found")
			return
		}
		h.logger.WithError(err).Error("Failed to delete insight")
		writeError(w, http.StatusInternalServerError, "failed to delete insight")
		return
	}

	utils.LogAuditEvent(r.Context(), h.auditRepo, r, "consciousness.insight.deleted", "consciousness_insight", &id, insight, nil, true)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
// ExportData returns all consciousness data for the tenant (GDPR Article 20 - Data Portability).
func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	export, err := h.repo.ExportTenantConsciousnessData(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to export consciousness data")
		writeError(w, http.StatusInternalServerError, "failed to export data")
		return
	}

	writeJSON(w, http.StatusOK, export)
}

// HealthCheck returns the health status of the consciousness service.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database connectivity
	if err := h.repo.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	// Get scheduler status if available
	schedulerStatus := map[string]interface{}{}
	if s := h.engine.GetSchedulerStatus(); s != nil {
		schedulerStatus = s
	}

	// Get retry queue size
	retryQueueSize, _ := h.repo.GetRetryQueueSize(ctx)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "healthy",
		"retry_queue_size": retryQueueSize,
		"scheduler":        schedulerStatus,
	})
}
