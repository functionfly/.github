package reputation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// Handler handles reputation API requests
type Handler struct {
	repo   *storage.PostgresDB
	logger *logrus.Logger
}

// NewHandler creates a new reputation handler
func NewHandler(repo *storage.PostgresDB, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		repo:   repo,
		logger: logger,
	}
}

// ReputationProfileResponse represents the API response for a reputation profile
type ReputationProfileResponse struct {
	UserID              uuid.UUID                      `json:"user_id"`
	TenantID            uuid.UUID                      `json:"tenant_id"`
	Username            string                         `json:"username,omitempty"`
	DisplayName         string                         `json:"display_name,omitempty"`
	BuilderScore        int                            `json:"builder_score"`
	OptimizerScore      int                            `json:"optimizer_score"`
	MentorScore         int                            `json:"mentor_score"`
	AgentWhispererScore int                            `json:"agent_whisperer_score"`
	ReliabilityIndex    float64                        `json:"reliability_index"`
	ConsistencyScore    float64                        `json:"consistency_score"`
	OverallScore        int                            `json:"overall_score"`
	Tier                string                         `json:"tier"`
	Badges              []storage.ReputationBadge      `json:"badges"`
	Stats               storage.ReputationStats        `json:"stats"`
	Rank                int                            `json:"rank,omitempty"`
	FunctionCount       int                            `json:"function_count,omitempty"`
	UpdatedAt           time.Time                      `json:"updated_at"`
}

// ReputationLeaderboardResponse represents the leaderboard API response
type ReputationLeaderboardResponse struct {
	Entries []ReputationLeaderboardEntry `json:"entries"`
	Total   int64                        `json:"total"`
	Page    int                          `json:"page"`
	PageSize int                         `json:"page_size"`
}

// ReputationLeaderboardEntry represents a single leaderboard entry
type ReputationLeaderboardEntry struct {
	Rank                int       `json:"rank"`
	UserID              uuid.UUID `json:"user_id"`
	Username            string    `json:"username"`
	DisplayName         string    `json:"display_name"`
	OverallScore        int       `json:"overall_score"`
	Tier                string    `json:"tier"`
	BuilderScore        int       `json:"builder_score"`
	OptimizerScore      int       `json:"optimizer_score"`
	MentorScore         int       `json:"mentor_score"`
	AgentWhispererScore int       `json:"agent_whisperer_score"`
	FunctionCount       int       `json:"function_count"`
}

// ReputationEventResponse represents a reputation event API response
type ReputationEventResponse struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	EventType    string                 `json:"event_type"`
	ScoreChange  int                    `json:"score_change"`
	Component    string                 `json:"component"`
	ReferenceID  *uuid.UUID             `json:"reference_id,omitempty"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ReputationFarmingAlertResponse represents an alert API response
type ReputationFarmingAlertResponse struct {
	ID                uuid.UUID              `json:"id"`
	Type              string                `json:"type"`
	Description       string                `json:"description"`
	AffectedFunctions []uuid.UUID           `json:"affected_functions"`
	AffectedUsers     []uuid.UUID           `json:"affected_users"`
	Severity          string                `json:"severity"`
	Status            string                `json:"status"`
	DetectedAt        time.Time             `json:"detected_at"`
	ResolvedAt        *time.Time            `json:"resolved_at,omitempty"`
	Notes             string                `json:"notes,omitempty"`
	Details           map[string]interface{} `json:"details"`
}

// TrustScoreWeightsConfigResponse represents trust weights API response
type TrustScoreWeightsConfigResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Reliability  float64   `json:"reliability"`
	Latency      float64   `json:"latency"`
	ErrorRate    float64   `json:"error_rate"`
	UserRating   float64   `json:"user_rating"`
	Verification float64   `json:"verification"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReputationStatsResponse represents aggregated reputation stats
type ReputationStatsResponse struct {
	TotalProfiles    int64            `json:"total_profiles"`
	AvgScore         float64          `json:"avg_score"`
	TierDistribution map[string]int64 `json:"tier_distribution"`
}

// HandleGetProfile handles GET /v1/reputation/profile
// Returns the current user's reputation profile
func (h *Handler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	profile, err := repo.GetProfile(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation profile")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation profile"))
		return
	}

	if profile == nil {
		// Create profile if it doesn't exist
		profile, err = repo.GetOrCreateProfile(claims.UserID, claims.TenantID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create reputation profile")
			apierror.WriteError(w, apierror.NewInternal("Failed to create reputation profile"))
			return
		}
	}

	// Get user's rank
	rank, _ := repo.GetUserRank(claims.UserID)

	// Get function count
	var functionCount int64
	h.repo.GORM.Raw(`
		SELECT COUNT(*) FROM registry_functions
		WHERE owner_user_id = ? AND tenant_id = ?
	`, claims.UserID, claims.TenantID).Scan(&functionCount)

	response := ReputationProfileResponse{
		UserID:              profile.UserID,
		TenantID:            profile.TenantID,
		BuilderScore:        profile.BuilderScore,
		OptimizerScore:      profile.OptimizerScore,
		MentorScore:         profile.MentorScore,
		AgentWhispererScore: profile.AgentWhispererScore,
		ReliabilityIndex:    profile.ReliabilityIndex,
		ConsistencyScore:    profile.ConsistencyScore,
		OverallScore:        profile.OverallScore,
		Tier:                string(profile.Tier),
		Badges:              profile.GetBadges(),
		Stats:               profile.GetStats(),
		Rank:                rank,
		FunctionCount:       int(functionCount),
		UpdatedAt:           profile.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetProfileByUserID handles GET /v1/reputation/profile/{userId}
// Returns a specific user's reputation profile (public)
func (h *Handler) HandleGetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	if userIDStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("user_id required"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user_id"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	profile, err := repo.GetProfile(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation profile")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation profile"))
		return
	}

	if profile == nil {
		apierror.WriteError(w, apierror.NewNotFound("Profile not found"))
		return
	}


	response := ReputationProfileResponse{
		UserID:              profile.UserID,
		TenantID:            profile.TenantID,
		BuilderScore:        profile.BuilderScore,
		OptimizerScore:      profile.OptimizerScore,
		MentorScore:         profile.MentorScore,
		AgentWhispererScore: profile.AgentWhispererScore,
		ReliabilityIndex:    profile.ReliabilityIndex,
		ConsistencyScore:    profile.ConsistencyScore,
		OverallScore:        profile.OverallScore,
		Tier:                string(profile.Tier),
		Badges:              profile.GetBadges(),
		Stats:               profile.GetStats(),
		FunctionCount:       profile.FunctionCount,
		UpdatedAt:           profile.UpdatedAt,
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetLeaderboard handles GET /v1/reputation/leaderboard
// Returns the reputation leaderboard
func (h *Handler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	entries, err := repo.GetLeaderboard(pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get leaderboard")
		apierror.WriteError(w, apierror.NewInternal("Failed to get leaderboard"))
		return
	}

	// Get total count for pagination
	var total int64
	h.repo.GORM.Model(&storage.ReputationProfile{}).Count(&total)

	response := ReputationLeaderboardResponse{
		Entries:  make([]ReputationLeaderboardEntry, 0, len(entries)),
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}

	for _, entry := range entries {
		response.Entries = append(response.Entries, ReputationLeaderboardEntry{
			Rank:                entry.Rank,
			UserID:              entry.UserID,
			Username:            entry.Username,
			DisplayName:         entry.DisplayName,
			OverallScore:        entry.OverallScore,
			Tier:                entry.Tier,
			BuilderScore:        entry.BuilderScore,
			OptimizerScore:      entry.OptimizerScore,
			MentorScore:         entry.MentorScore,
			AgentWhispererScore: entry.AgentWhispererScore,
			FunctionCount:       entry.FunctionCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetReputationEvents handles GET /v1/reputation/events
// Returns reputation events for the current user
func (h *Handler) HandleGetReputationEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	events, err := repo.GetReputationEvents(claims.UserID, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation events")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation events"))
		return
	}

	response := struct {
		Events   []ReputationEventResponse `json:"events"`
		Page     int                      `json:"page"`
		PageSize int                      `json:"page_size"`
	}{
		Events:   make([]ReputationEventResponse, 0, len(events)),
		Page:     page,
		PageSize: pageSize,
	}

	for _, event := range events {
		response.Events = append(response.Events, ReputationEventResponse{
			ID:          event.ID,
			UserID:      event.UserID,
			EventType:   event.EventType,
			ScoreChange: event.ScoreChange,
			Component:   event.Component,
			ReferenceID: event.ReferenceID,
			Description: event.Description,
			Metadata:    event.Metadata,
			CreatedAt:   event.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAddScore handles POST /v1/reputation/score
// Adds score to a user's reputation (internal API for other services)
func (h *Handler) HandleAddScore(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var request struct {
		UserID      uuid.UUID `json:"user_id"`
		TenantID    uuid.UUID `json:"tenant_id"`
		Component   string    `json:"component"` // builder, optimizer, mentor, agent_whisperer
		ScoreChange int       `json:"score_change"`
		Description string    `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if request.UserID == uuid.Nil {
		request.UserID = claims.UserID
	}
	if request.TenantID == uuid.Nil {
		request.TenantID = claims.TenantID
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)

	// Record the score change
	if err := repo.RecordScoreChange(request.UserID, request.TenantID, request.Component, request.ScoreChange); err != nil {
		h.logger.WithError(err).Error("Failed to record score change")
		apierror.WriteError(w, apierror.NewInternal("Failed to record score change"))
		return
	}

	// Record the event
	event := &storage.ReputationEvent{
		UserID:       request.UserID,
		EventType:    "score_change",
		ScoreChange:  request.ScoreChange,
		Component:    request.Component,
		Description: request.Description,
	}
	if err := repo.RecordReputationEvent(event); err != nil {
		h.logger.WithError(err).Warn("Failed to record reputation event")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleGetStats handles GET /v1/reputation/stats
// Returns aggregated reputation statistics
func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	stats, err := repo.GetReputationStats(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get reputation stats"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetTrustWeights handles GET /v1/reputation/trust-weights
// Returns the active trust score weights configuration
func (h *Handler) HandleGetTrustWeights(w http.ResponseWriter, r *http.Request) {
	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	config, err := repo.GetActiveTrustScoreWeights()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trust score weights")
		apierror.WriteError(w, apierror.NewInternal("Failed to get trust score weights"))
		return
	}

	response := TrustScoreWeightsConfigResponse{
		ID:           config.ID,
		Name:         config.Name,
		Description:  config.Description,
		Reliability:  config.Reliability,
		Latency:      config.Latency,
		ErrorRate:    config.ErrorRate,
		UserRating:   config.UserRating,
		Verification: config.Verification,
		IsActive:     config.IsActive,
		CreatedAt:    config.CreatedAt,
		UpdatedAt:    config.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateTrustWeights handles PUT /v1/reputation/trust-weights
// Updates the trust score weights configuration (admin only)
func (h *Handler) HandleUpdateTrustWeights(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemWrite) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req TrustScoreWeightsConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	config := &storage.TrustScoreWeightsConfigV2{
		ID:           req.ID,
		Name:         req.Name,
		Description:  req.Description,
		Reliability:  req.Reliability,
		Latency:      req.Latency,
		ErrorRate:    req.ErrorRate,
		UserRating:   req.UserRating,
		Verification: req.Verification,
		IsActive:     req.IsActive,
	}

	if err := repo.UpdateTrustScoreWeights(config); err != nil {
		h.logger.WithError(err).Error("Failed to update trust score weights")
		apierror.WriteError(w, apierror.NewInternal("Failed to update trust score weights"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetTrustWeightsHistory handles GET /v1/reputation/trust-weights/history
// Returns the history of trust score weights configurations
func (h *Handler) HandleGetTrustWeightsHistory(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	configs, total, err := repo.GetTrustScoreWeightsHistory(pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trust score weights history")
		apierror.WriteError(w, apierror.NewInternal("Failed to get trust score weights history"))
		return
	}

	response := struct {
		Configs  []TrustScoreWeightsConfigResponse `json:"configs"`
		Total    int64                             `json:"total"`
		Page     int                               `json:"page"`
		PageSize int                               `json:"page_size"`
	}{
		Configs:  make([]TrustScoreWeightsConfigResponse, 0, len(configs)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	for _, config := range configs {
		response.Configs = append(response.Configs, TrustScoreWeightsConfigResponse{
			ID:           config.ID,
			Name:         config.Name,
			Description:  config.Description,
			Reliability:  config.Reliability,
			Latency:      config.Latency,
			ErrorRate:    config.ErrorRate,
			UserRating:   config.UserRating,
			Verification: config.Verification,
			IsActive:     config.IsActive,
			CreatedAt:    config.CreatedAt,
			UpdatedAt:    config.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetReputationFarmingAlerts handles GET /v1/reputation/alerts
// Returns reputation farming alerts (admin only)
func (h *Handler) HandleGetReputationFarmingAlerts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemRead) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status := r.URL.Query().Get("status")

	offset := (page - 1) * pageSize

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	alerts, total, err := repo.GetReputationFarmingAlerts(status, pageSize, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation farming alerts")
		apierror.WriteError(w, apierror.NewInternal("Failed to get alerts"))
		return
	}

	response := struct {
		Alerts   []ReputationFarmingAlertResponse `json:"alerts"`
		Total    int64                            `json:"total"`
		Page     int                              `json:"page"`
		PageSize int                              `json:"page_size"`
	}{
		Alerts:   make([]ReputationFarmingAlertResponse, 0, len(alerts)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	for _, alert := range alerts {
		response.Alerts = append(response.Alerts, ReputationFarmingAlertResponse{
			ID:                alert.ID,
			Type:              alert.Type,
			Description:       alert.Description,
			AffectedFunctions: alert.AffectedFunctions,
			AffectedUsers:     alert.AffectedUsers,
			Severity:          alert.Severity,
			Status:            alert.Status,
			DetectedAt:        alert.DetectedAt,
			ResolvedAt:        alert.ResolvedAt,
			Notes:             alert.Notes,
			Details:           alert.Details,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleResolveReputationFarmingAlert handles POST /v1/reputation/alerts/{alertId}/resolve
// Resolves a reputation farming alert (admin only)
func (h *Handler) HandleResolveReputationFarmingAlert(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemWrite) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	vars := mux.Vars(r)
	alertIDStr := vars["alertId"]
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert_id"))
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	if err := repo.ResolveReputationFarmingAlert(alertID, claims.UserID, req.Notes); err != nil {
		h.logger.WithError(err).Error("Failed to resolve alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to resolve alert"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDismissReputationFarmingAlert handles POST /v1/reputation/alerts/{alertId}/dismiss
// Dismisses a reputation farming alert (admin only)
func (h *Handler) HandleDismissReputationFarmingAlert(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemWrite) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	vars := mux.Vars(r)
	alertIDStr := vars["alertId"]
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert_id"))
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	if err := repo.DismissReputationFarmingAlert(alertID, claims.UserID, req.Notes); err != nil {
		h.logger.WithError(err).Error("Failed to dismiss alert")
		apierror.WriteError(w, apierror.NewInternal("Failed to dismiss alert"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDetectReputationFarming handles POST /v1/reputation/detect-farming
// Triggers reputation farming detection (admin only)
func (h *Handler) HandleDetectReputationFarming(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemWrite) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	alerts, err := repo.DetectReputationFarming(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to detect reputation farming")
		apierror.WriteError(w, apierror.NewInternal("Failed to detect reputation farming"))
		return
	}

	response := struct {
		AlertsDetected int                            `json:"alerts_detected"`
		Alerts        []ReputationFarmingAlertResponse `json:"alerts"`
	}{
		AlertsDetected: len(alerts),
		Alerts:        make([]ReputationFarmingAlertResponse, 0, len(alerts)),
	}

	for _, alert := range alerts {
		response.Alerts = append(response.Alerts, ReputationFarmingAlertResponse{
			ID:                alert.ID,
			Type:              alert.Type,
			Description:       alert.Description,
			AffectedFunctions: alert.AffectedFunctions,
			AffectedUsers:     alert.AffectedUsers,
			Severity:          alert.Severity,
			Status:            alert.Status,
			DetectedAt:        alert.DetectedAt,
			ResolvedAt:        alert.ResolvedAt,
			Notes:             alert.Notes,
			Details:           alert.Details,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCleanupTrustHistory handles POST /v1/reputation/cleanup-trust-history
// Triggers cleanup of old trust history entries (admin only)
func (h *Handler) HandleCleanupTrustHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || !claims.HasPermission(auth.PermSystemWrite) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req struct {
		RetentionDays int `json:"retention_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.RetentionDays = 90 // default to 90 days
	}

	repo := storage.NewReputationRepository(h.repo.GORM, h.logger)
	deleted, err := repo.CleanupOldTrustHistory(r.Context(), req.RetentionDays)
	if err != nil {
		h.logger.WithError(err).Error("Failed to cleanup trust history")
		apierror.WriteError(w, apierror.NewInternal("Failed to cleanup trust history"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted_entries": deleted,
	})
}
