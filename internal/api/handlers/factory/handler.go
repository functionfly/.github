package factory

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/discovery"
	factorysvc "github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Handler struct {
	db             *gorm.DB
	runner         pipelineRunner
	discovery      *discovery.Service
	publisher      publishedFunctionsLister
	config         *factorysvc.Config
	configSvc      *factorysvc.Service
	scheduleRunner schedulerInterface
}

type pipelineRunner interface {
	Run(ctx context.Context) (*factorysvc.FactoryRun, error)
}

type schedulerInterface interface {
	UpdateConfig(ctx context.Context, config scheduler.FactoryScheduleConfig) error
	GetStatus() scheduler.FactoryScheduleStatus
}

type publishedFunctionsLister interface {
	GetPublishedFunctions(ctx context.Context, agentID string, limit, offset int) ([]deployment.PublishedFunction, int64, error)
}

func NewHandler(db *gorm.DB, factory *factorysvc.Service, discovery *discovery.Service, publisher *deployment.Publisher, config *factorysvc.Config, schedulerrunner ...schedulerInterface) *Handler {
	h := &Handler{db: db, runner: factory, discovery: discovery, publisher: publisher, config: config, configSvc: factory}
	if len(schedulerrunner) > 0 {
		h.scheduleRunner = schedulerrunner[0]
	}
	return h
}

func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	// Get latest config from database
	loadedConfig, err := h.configSvc.GetConfig(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("failed to load config from database for status")
	}

	var latest factorysvc.FactoryRun
	db := h.db.WithContext(r.Context()).Session(&gorm.Session{Logger: h.db.Logger.LogMode(logger.Silent)})
	latestErr := db.Order("created_at DESC").First(&latest).Error

	var totals struct {
		Runs          int64
		Versions      int64
		Published     int64
		Opportunities int64
	}
	h.db.WithContext(r.Context()).Model(&factorysvc.FactoryRun{}).Count(&totals.Runs)
	h.db.WithContext(r.Context()).Model(&factorysvc.FactoryVersion{}).Count(&totals.Versions)
	h.db.WithContext(r.Context()).Model(&deployment.PublishedFunction{}).Where("agent_id = ?", loadedConfig.AgentID).Count(&totals.Published)
	h.db.WithContext(r.Context()).Model(&discovery.Opportunity{}).Count(&totals.Opportunities)

	status := map[string]any{
		"agent_id": loadedConfig.AgentID,
		"config":   loadedConfig,
		"totals": map[string]any{
			"runs":            totals.Runs,
			"versions":        totals.Versions,
			"published":       totals.Published,
			"opportunities":   totals.Opportunities,
			"autopublish":     loadedConfig.AutoPublish,
			"quality_minimum": loadedConfig.MinimumQualityScore,
			"test_minimum":    loadedConfig.MinimumTestScore,
		},
	}
	if latestErr == nil {
		status["latest_run"] = map[string]any{
			"id":                    latest.ID,
			"status":                latest.Status,
			"started_at":             latest.CreatedAt,
			"completed_at":           latest.CompletedAt,
			"opportunities_found":    latest.OpportunitiesScanned,
			"opportunities_approved": 0,
			"opportunities_rejected": 0,
			"functions_created":     latest.FunctionsGenerated,
			"functions_failed":      0,
			"functions_published":   latest.FunctionsPublished,
			"quality_score":         latest.AverageQualityScore,
			"error":                latest.ErrorMessage,
			"metadata":              latest.Metadata,
		}
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) HandleListOpportunities(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r, 20, 100)
	query := h.db.WithContext(r.Context()).Model(&discovery.Opportunity{})
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if source := strings.TrimSpace(r.URL.Query().Get("source")); source != "" {
		query = query.Where("source = ?", source)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count opportunities")
		return
	}
	var items []discovery.Opportunity
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list opportunities")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"opportunities": items, "total": total, "limit": limit, "offset": offset})
}

func (h *Handler) HandleRunPipeline(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	run, err := h.runner.Run(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("factory pipeline run completed with error")
		writeJSON(w, http.StatusAccepted, map[string]any{"run": run, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"run": run})
}

func (h *Handler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r, 20, 100)
	includePublished := r.URL.Query().Get("include_published") != "false"

	// Get latest config from database for agent ID
	loadedConfig, err := h.configSvc.GetConfig(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("failed to load config from database for list functions")
	}

	var versions []factorysvc.FactoryVersion
	var versionsTotal int64
	versionQuery := h.db.WithContext(r.Context()).Model(&factorysvc.FactoryVersion{})
	if err := versionQuery.Count(&versionsTotal).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count factory versions")
		return
	}
	if err := versionQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&versions).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list factory versions")
		return
	}

	response := map[string]any{"versions": versions, "total_versions": versionsTotal, "limit": limit, "offset": offset}
	if includePublished {
		published, total, err := h.publisher.GetPublishedFunctions(r.Context(), loadedConfig.AgentID, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list published factory functions")
			return
		}
		response["published_functions"] = published
		response["total_published"] = total
	}
	writeJSON(w, http.StatusOK, response)
}

// HandleListPendingReviews returns opportunities that require manual review.
func (h *Handler) HandleListPendingReviews(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r, 20, 100)
	query := h.db.WithContext(r.Context()).Model(&discovery.Opportunity{}).
		Where("status = ? AND review_status = ?", discovery.OpportunityStatusNeedsReview, discovery.ReviewStatusPending)

	if source := strings.TrimSpace(r.URL.Query().Get("source")); source != "" {
		query = query.Where("source = ?", source)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count pending reviews")
		return
	}

	var items []discovery.Opportunity
	if err := query.Order("review_requested_at ASC, created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list pending reviews")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"reviews": items, "total": total, "limit": limit, "offset": offset})
}

// HandleGetOpportunity returns details of a specific opportunity for review.
func (h *Handler) HandleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/factory/opportunities/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "opportunity ID required")
		return
	}

	var opp discovery.Opportunity
	if err := h.db.WithContext(r.Context()).First(&opp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "opportunity not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get opportunity")
		return
	}

	writeJSON(w, http.StatusOK, opp)
}

// HandleApproveOpportunity approves an opportunity for publishing.
func (h *Handler) HandleApproveOpportunity(w http.ResponseWriter, r *http.Request) {
	// URL path is /v1/admin/factory/opportunities/{id}/approve
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/v1/admin/factory/opportunities/")
	id = strings.TrimSuffix(id, "/approve")
	if id == "" || id == path || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "opportunity ID required")
		return
	}

	userID := "anonymous"
	if user := middleware.GetUserFromContext(r); user != nil {
		userID = user.ID
	}

	err := h.discovery.ApplyReviewDecision(r.Context(), id, discovery.ReviewDecision{
		Approved: true,
		Reason:   "approved via manual review",
		Actor:    userID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "opportunity not found")
			return
		}
		logrus.WithError(err).Error("failed to approve opportunity")
		writeError(w, http.StatusInternalServerError, "failed to approve opportunity")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "opportunity approved", "id": id})
}

// HandleRejectOpportunity rejects an opportunity with a reason.
func (h *Handler) HandleRejectOpportunity(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/v1/admin/factory/opportunities/")
	id = strings.TrimSuffix(id, "/reject")
	if id == "" || id == path || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "opportunity ID required")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Reason) == "" {
		writeError(w, http.StatusBadRequest, "rejection reason required")
		return
	}

	userID := "anonymous"
	if user := middleware.GetUserFromContext(r); user != nil {
		userID = user.ID
	}

	err := h.discovery.ApplyReviewDecision(r.Context(), id, discovery.ReviewDecision{
		Approved: false,
		Reason:   req.Reason,
		Actor:    userID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "opportunity not found")
			return
		}
		logrus.WithError(err).Error("failed to reject opportunity")
		writeError(w, http.StatusInternalServerError, "failed to reject opportunity")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "opportunity rejected", "id": id})
}

func (h *Handler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	// Get the latest config from the database
	loadedConfig, err := h.configSvc.GetConfig(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("failed to load config from database, using in-memory")
		writeJSON(w, http.StatusOK, h.config)
		return
	}
	// Update in-memory config to match database
	*h.config = loadedConfig
	writeJSON(w, http.StatusOK, h.config)
}

func (h *Handler) HandleGetScheduleStatus(w http.ResponseWriter, r *http.Request) {
	if h.scheduleRunner == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_enabled": false,
			"is_running": false,
			"message":    "scheduler not available",
		})
		return
	}
	status := h.scheduleRunner.GetStatus()
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req configUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	h.applyConfigUpdate(req)
	// Persist the updated config to the database
	if err := h.configSvc.SaveConfig(r.Context(), *h.config); err != nil {
		logrus.WithError(err).Error("failed to persist factory config")
		writeError(w, http.StatusInternalServerError, "failed to persist configuration")
		return
	}
	// Update the scheduler if available
	if h.scheduleRunner != nil {
		scheduleConfig := scheduler.FactoryScheduleConfig{
			Enabled:  h.config.ScheduleEnabled,
			Cron:     h.config.ScheduleCron,
			Timezone: h.config.ScheduleTimezone,
		}
		if err := h.scheduleRunner.UpdateConfig(r.Context(), scheduleConfig); err != nil {
			logrus.WithError(err).Error("failed to update factory pipeline scheduler")
			writeError(w, http.StatusInternalServerError, "failed to update schedule")
			return
		}
	}
	// Reload from database to ensure consistency
	loadedConfig, err := h.configSvc.GetConfig(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("failed to reload config after save")
	} else {
		*h.config = loadedConfig
	}
	writeJSON(w, http.StatusOK, h.config)
}

type configUpdateRequest struct {
	DiscoveryBatchSize          *int            `json:"discovery_batch_size"`
	MinimumQualityScore         *float64        `json:"minimum_quality_score"`
	MinimumTestScore            *float64        `json:"minimum_test_score"`
	RequireAllTestsPass         *bool           `json:"require_all_tests_pass"`
	AutoPublish                 *bool           `json:"auto_publish"`
	MaxOpportunitiesPerRun      *int            `json:"max_opportunities_per_run"`
	RetryAttempts               *int            `json:"retry_attempts"`
	RetryBackoffMs              *int            `json:"retry_backoff_ms"`
	ScheduleEnabled             *bool           `json:"schedule_enabled"`
	ScheduleCron                *string         `json:"schedule_cron"`
	ScheduleTimezone            *string         `json:"schedule_timezone"`
	NotificationWebhookURL      *string         `json:"notification_webhook_url"`
	RateLimitPerHour            *int            `json:"rate_limit_per_hour"`
	MaxConcurrentRuns           *int            `json:"max_concurrent_runs"`
	DryRunMode                  *bool           `json:"dry_run_mode"`
	DiscoverySources            []string        `json:"discovery_sources"`
	FeatureFlags                map[string]bool `json:"feature_flags"`
	ApprovalRequiredAboveQuality *int           `json:"approval_required_above_quality"`
	ApprovalRequiredAboveTest   *int            `json:"approval_required_above_test"`
	LogLevel                    *string         `json:"log_level"`
	NotifyOnFailure             *bool           `json:"notify_on_failure"`
	NotifyOnReviewRequired      *bool           `json:"notify_on_review_required"`
	DiscoveryCooldownMinutes    *int            `json:"discovery_cooldown_minutes"`
	MaxVersionsPerFunction      *int            `json:"max_versions_per_function"`
}

func (h *Handler) applyConfigUpdate(req configUpdateRequest) {
	if req.DiscoveryBatchSize != nil && *req.DiscoveryBatchSize > 0 {
		h.config.DiscoveryBatchSize = *req.DiscoveryBatchSize
	}
	if req.MinimumQualityScore != nil {
		h.config.MinimumQualityScore = *req.MinimumQualityScore
	}
	if req.MinimumTestScore != nil {
		h.config.MinimumTestScore = *req.MinimumTestScore
	}
	if req.RequireAllTestsPass != nil {
		h.config.RequireAllTestsPass = *req.RequireAllTestsPass
	}
	if req.AutoPublish != nil {
		h.config.AutoPublish = *req.AutoPublish
	}
	if req.MaxOpportunitiesPerRun != nil && *req.MaxOpportunitiesPerRun > 0 {
		h.config.MaxOpportunitiesPerRun = *req.MaxOpportunitiesPerRun
	}
	if req.RetryAttempts != nil && *req.RetryAttempts >= 0 {
		h.config.RetryAttempts = *req.RetryAttempts
	}
	if req.RetryBackoffMs != nil && *req.RetryBackoffMs >= 0 {
		h.config.RetryBackoff = time.Duration(*req.RetryBackoffMs) * time.Millisecond
		h.config.RetryBackoffMs = *req.RetryBackoffMs
	}
	if req.ScheduleEnabled != nil {
		h.config.ScheduleEnabled = *req.ScheduleEnabled
	}
	if req.ScheduleCron != nil {
		h.config.ScheduleCron = *req.ScheduleCron
	}
	if req.ScheduleTimezone != nil {
		h.config.ScheduleTimezone = *req.ScheduleTimezone
	}
	if req.NotificationWebhookURL != nil {
		h.config.NotificationWebhookURL = *req.NotificationWebhookURL
	}
	if req.RateLimitPerHour != nil && *req.RateLimitPerHour >= 0 {
		h.config.RateLimitPerHour = *req.RateLimitPerHour
	}
	if req.MaxConcurrentRuns != nil && *req.MaxConcurrentRuns > 0 {
		h.config.MaxConcurrentRuns = *req.MaxConcurrentRuns
	}
	if req.DryRunMode != nil {
		h.config.DryRunMode = *req.DryRunMode
	}
	if req.DiscoverySources != nil {
		h.config.DiscoverySources = req.DiscoverySources
	}
	if req.FeatureFlags != nil {
		h.config.FeatureFlags = req.FeatureFlags
	}
	if req.ApprovalRequiredAboveQuality != nil && *req.ApprovalRequiredAboveQuality >= 0 {
		h.config.ApprovalRequiredAboveQuality = *req.ApprovalRequiredAboveQuality
	}
	if req.ApprovalRequiredAboveTest != nil && *req.ApprovalRequiredAboveTest >= 0 {
		h.config.ApprovalRequiredAboveTest = *req.ApprovalRequiredAboveTest
	}
	if req.LogLevel != nil && *req.LogLevel != "" {
		h.config.LogLevel = *req.LogLevel
	}
	if req.NotifyOnFailure != nil {
		h.config.NotifyOnFailure = *req.NotifyOnFailure
	}
	if req.NotifyOnReviewRequired != nil {
		h.config.NotifyOnReviewRequired = *req.NotifyOnReviewRequired
	}
	if req.DiscoveryCooldownMinutes != nil && *req.DiscoveryCooldownMinutes >= 0 {
		h.config.DiscoveryCooldownMinutes = *req.DiscoveryCooldownMinutes
	}
	if req.MaxVersionsPerFunction != nil && *req.MaxVersionsPerFunction > 0 {
		h.config.MaxVersionsPerFunction = *req.MaxVersionsPerFunction
	}
}

func parseLimitOffset(r *http.Request, defaultLimit, maxLimit int) (int, int) {
	limit := defaultLimit
	offset := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
