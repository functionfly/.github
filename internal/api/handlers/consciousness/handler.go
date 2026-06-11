package consciousness

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/consciousness"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler provides HTTP handlers for the consciousness API.
type Handler struct {
	repo         *consciousness.Repository
	engine       *consciousness.Engine
	logger       *logrus.Logger
	rateLimiter  *middleware.ConsciousnessRateLimiter
	authMiddleware *middleware.AuthMiddleware
}

// NewHandler creates a new consciousness handler.
func NewHandler(db *sql.DB, logger *logrus.Logger, rateLimiter *middleware.ConsciousnessRateLimiter, authMiddleware *middleware.AuthMiddleware) *Handler {
	repo := consciousness.NewRepository(db, logger)
	engine := consciousness.NewEngine(db, logger)
	return &Handler{repo: repo, engine: engine, logger: logger, rateLimiter: rateLimiter, authMiddleware: authMiddleware}
}

// RegisterRoutes registers consciousness API routes on the given router.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/consciousness").Subrouter()

	requireAuth := func(next http.HandlerFunc) http.HandlerFunc {
		if h.authMiddleware != nil {
			return h.authMiddleware.RequireAuth(next)
		}
		return next
	}

	scoreHandler := h.GetAwarenessScore
	insightsHandler := h.ListInsights
	runHandler := h.RunAnalysis

	if h.rateLimiter != nil {
		scoreHandler = h.rateLimiter.LimitScore(h.GetAwarenessScore)
		insightsHandler = h.rateLimiter.LimitListInsights(h.ListInsights)
		runHandler = h.rateLimiter.LimitRunAnalysis(h.RunAnalysis)
	}

	s.HandleFunc("/score", requireAuth(scoreHandler)).Methods("GET")
	s.HandleFunc("/insights", requireAuth(insightsHandler)).Methods("GET")
	s.HandleFunc("/insights/{id}", requireAuth(h.GetInsight)).Methods("GET")
	s.HandleFunc("/insights/{id}/dismiss", requireAuth(h.DismissInsight)).Methods("POST")
	s.HandleFunc("/insights/{id}/apply", requireAuth(h.ApplyAction)).Methods("POST")
	s.HandleFunc("/preferences", requireAuth(h.GetPreferences)).Methods("GET")
	s.HandleFunc("/preferences", requireAuth(h.UpdatePreferences)).Methods("PUT")
	s.HandleFunc("/run", requireAuth(runHandler)).Methods("POST")
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
		if !cat.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid category")
			return
		}
		params.Category = &cat
	}
	if v := r.URL.Query().Get("severity"); v != "" {
		sev := consciousness.InsightSeverity(v)
		if !sev.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid severity")
			return
		}
		params.Severity = &sev
	}
	if v := r.URL.Query().Get("status"); v != "" {
		st := consciousness.InsightStatus(v)
		if !st.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid status")
			return
		}
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
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}
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
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid insight ID")
		return
	}

	if err := h.repo.DismissInsight(r.Context(), id, tenantID); err != nil {
		writeError(w, http.StatusNotFound, "insight not found or already dismissed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}

// ApplyAction applies the suggested action for an insight.
func (h *Handler) ApplyAction(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}
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
	if r.Header.Get("Content-Type") != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}
	if r.Body != nil {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if req.DryRun {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"dry_run":     true,
			"action_type": insight.ActionType,
			"preview":     insight.ActionPreview,
			"message":     "This is a preview. No changes were made.",
		})
		return
	}

	if err := h.repo.ApplyInsight(r.Context(), id, tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to apply action")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "applied",
		"action_type": insight.ActionType,
		"message":     "Action has been applied. Changes will take effect shortly.",
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

	if r.Header.Get("Content-Type") != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var input struct {
		EmailEnabled       bool     `json:"email_enabled"`
		SlackEnabled       bool     `json:"slack_enabled"`
		SlackWebhookURL    *string  `json:"slack_webhook_url,omitempty"`
		InAppEnabled       bool     `json:"inapp_enabled"`
		WebhookEnabled     bool     `json:"webhook_enabled"`
		WebhookURL         *string  `json:"webhook_url,omitempty"`
		DigestFrequency    string   `json:"digest_frequency"`
		QuietHoursStart    *string   `json:"quiet_hours_start,omitempty"`
		QuietHoursEnd      *string   `json:"quiet_hours_end,omitempty"`
		Timezone           string   `json:"timezone"`
		EnabledCategories  []string `json:"enabled_categories"`
		MinNotifySeverity   string   `json:"min_notify_severity"`
		AutoApplyEnabled   bool     `json:"auto_apply_enabled"`
		AutoApplyCategories []string `json:"auto_apply_categories"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536)).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validatePreferencesInput(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	prefs := consciousness.Preferences{
		TenantID:           tenantID,
		EmailEnabled:       input.EmailEnabled,
		SlackEnabled:       input.SlackEnabled,
		SlackWebhookURL:    input.SlackWebhookURL,
		InAppEnabled:       input.InAppEnabled,
		WebhookEnabled:     input.WebhookEnabled,
		WebhookURL:         input.WebhookURL,
		DigestFrequency:    input.DigestFrequency,
		QuietHoursStart:    input.QuietHoursStart,
		QuietHoursEnd:      input.QuietHoursEnd,
		Timezone:           input.Timezone,
		EnabledCategories:  input.EnabledCategories,
		MinNotifySeverity:  input.MinNotifySeverity,
		AutoApplyEnabled:   input.AutoApplyEnabled,
		AutoApplyCategories: input.AutoApplyCategories,
	}

	if err := h.repo.UpsertPreferences(r.Context(), &prefs); err != nil {
		h.logger.WithError(err).Error("Failed to update preferences")
		writeError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

func validatePreferencesInput(input *struct {
	EmailEnabled       bool     `json:"email_enabled"`
	SlackEnabled       bool     `json:"slack_enabled"`
	SlackWebhookURL    *string  `json:"slack_webhook_url,omitempty"`
	InAppEnabled       bool     `json:"inapp_enabled"`
	WebhookEnabled     bool     `json:"webhook_enabled"`
	WebhookURL         *string  `json:"webhook_url,omitempty"`
	DigestFrequency    string   `json:"digest_frequency"`
	QuietHoursStart    *string   `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd      *string   `json:"quiet_hours_end,omitempty"`
	Timezone           string   `json:"timezone"`
	EnabledCategories  []string `json:"enabled_categories"`
	MinNotifySeverity   string   `json:"min_notify_severity"`
	AutoApplyEnabled   bool     `json:"auto_apply_enabled"`
	AutoApplyCategories []string `json:"auto_apply_categories"`
}) error {
	if input.WebhookEnabled && input.WebhookURL != nil && *input.WebhookURL != "" {
		if err := validateWebhookURL(*input.WebhookURL); err != nil {
			return fmt.Errorf("webhook_url: %w", err)
		}
	}
	if input.SlackEnabled && input.SlackWebhookURL != nil && *input.SlackWebhookURL != "" {
		if err := validateWebhookURL(*input.SlackWebhookURL); err != nil {
			return fmt.Errorf("slack_webhook_url: %w", err)
		}
	}
	if input.QuietHoursStart != nil && *input.QuietHoursStart != "" {
		if err := validateTimeFormat(*input.QuietHoursStart); err != nil {
			return fmt.Errorf("quiet_hours_start: %w", err)
		}
	}
	if input.QuietHoursEnd != nil && *input.QuietHoursEnd != "" {
		if err := validateTimeFormat(*input.QuietHoursEnd); err != nil {
			return fmt.Errorf("quiet_hours_end: %w", err)
		}
	}
	if input.Timezone != "" {
		if err := validateTimezone(input.Timezone); err != nil {
			return fmt.Errorf("timezone: %w", err)
		}
	}
	return nil
}

func validateTimeFormat(t string) error {
	_, err := time.Parse("15:04", t)
	if err != nil {
		return fmt.Errorf("must be in HH:MM format (24-hour)")
	}
	return nil
}

func validateTimezone(tz string) error {
	// Check if the timezone is valid by loading the timezone
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		return fmt.Errorf("invalid timezone")
	}
	return nil
}

func validateWebhookURL(webhookURL string) error {
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("must use HTTPS")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("must have a valid host")
	}

	host := parsedURL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("cannot point to localhost")
	}

	if isPrivateIP(host) {
		return fmt.Errorf("cannot point to private IP addresses")
	}

	return nil
}

func isPrivateIP(host string) bool {
	privatePrefixes := []string{
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"127.",
		"0.",
		"::1",
		"fc00:",
		"fe80:",
	}

	for _, prefix := range privatePrefixes {
		if len(host) >= len(prefix) && host[:len(prefix)] == prefix {
			return true
		}
	}

	return false
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
