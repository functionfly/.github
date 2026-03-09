package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ExperimentHandler handles experiment API endpoints
type ExperimentHandler struct {
	db            *gorm.DB
	experimentSvc *factory.ExperimentService
	generationSvc *factory.GenerationExperimentAdapter
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(db *gorm.DB, experimentSvc *factory.ExperimentService, generationSvc *factory.GenerationExperimentAdapter) *ExperimentHandler {
	return &ExperimentHandler{
		db:            db,
		experimentSvc: experimentSvc,
		generationSvc: generationSvc,
	}
}

// HandleCreateExperiment creates a new experiment
func (h *ExperimentHandler) HandleCreateExperiment(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		Name            string           `json:"name"`
		Description     string           `json:"description"`
		AgentID         string           `json:"agent_id"`
		AutoSelect      bool             `json:"auto_select"`
		MinSamples      int              `json:"min_samples"`
		ConfidenceLevel float64          `json:"confidence_level"`
		Variants        []VariantRequest `json:"variants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.AgentID == "" {
		experimentWriteError(w, http.StatusBadRequest, "name and agent_id are required")
		return
	}

	if len(req.Variants) < 2 {
		experimentWriteError(w, http.StatusBadRequest, "at least 2 variants are required")
		return
	}

	// Check for duplicate name
	existing, err := h.experimentSvc.GetExperimentByName(r.Context(), req.Name)
	if err == nil && existing != nil {
		experimentWriteError(w, http.StatusConflict, "experiment with this name already exists")
		return
	}

	// Convert variants
	variants := make([]factory.ExperimentVariant, len(req.Variants))
	for i, v := range req.Variants {
		variants[i] = factory.ExperimentVariant{
			Name:            v.Name,
			Description:     v.Description,
			PromptTemplate:  v.PromptTemplate,
			Weight:          v.Weight,
			IsControl:       v.IsControl,
			IsActive:        true,
			Metadata:        v.Metadata,
		}
	}

	exp := &factory.Experiment{
		Name:            req.Name,
		Description:     req.Description,
		AgentID:         req.AgentID,
		Status:          "draft",
		AutoSelect:      req.AutoSelect,
		MinSamples:      req.MinSamples,
		ConfidenceLevel: req.ConfidenceLevel,
		Variants:        variants,
		Metadata:        map[string]any{},
	}

	if exp.MinSamples == 0 {
		exp.MinSamples = 10
	}
	if exp.ConfidenceLevel == 0 {
		exp.ConfidenceLevel = 0.95
	}

	if err := h.experimentSvc.CreateExperiment(r.Context(), exp); err != nil {
		logrus.WithError(err).Error("failed to create experiment")
		experimentWriteError(w, http.StatusInternalServerError, "failed to create experiment")
		return
	}

	experimentWriteJSON(w, http.StatusCreated, exp)
}

// HandleGetExperiment gets an experiment by ID
func (h *ExperimentHandler) HandleGetExperiment(w http.ResponseWriter, r *http.Request) {
	id := extractExperimentID(r)
	if id == "" {
		experimentWriteError(w, http.StatusBadRequest, "experiment ID required")
		return
	}

	expID, err := uuid.Parse(id)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid experiment ID")
		return
	}

	exp, err := h.experimentSvc.GetExperiment(r.Context(), expID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			experimentWriteError(w, http.StatusNotFound, "experiment not found")
			return
		}
		experimentWriteError(w, http.StatusInternalServerError, "failed to get experiment")
		return
	}

	experimentWriteJSON(w, http.StatusOK, exp)
}

// HandleListExperiments lists all experiments
func (h *ExperimentHandler) HandleListExperiments(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")
	limit, offset := parseExperimentLimitOffset(r, 20, 100)

	if agentID == "" {
		experimentWriteError(w, http.StatusBadRequest, "agent_id is required")
		return
	}

	experiments, total, err := h.experimentSvc.ListExperiments(r.Context(), agentID, status, limit, offset)
	if err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to list experiments")
		return
	}

	experimentWriteJSON(w, http.StatusOK, map[string]any{
		"experiments": experiments,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// HandleUpdateExperimentStatus updates experiment status
func (h *ExperimentHandler) HandleUpdateExperimentStatus(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := extractExperimentID(r)
	if id == "" {
		experimentWriteError(w, http.StatusBadRequest, "experiment ID required")
		return
	}

	expID, err := uuid.Parse(id)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid experiment ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validStatuses := map[string]bool{
		"draft":     true,
		"running":   true,
		"paused":    true,
		"completed": true,
	}

	if !validStatuses[req.Status] {
		experimentWriteError(w, http.StatusBadRequest, "invalid status")
		return
	}

	if err := h.experimentSvc.UpdateExperimentStatus(r.Context(), expID, req.Status); err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to update experiment status")
		return
	}

	experimentWriteJSON(w, http.StatusOK, map[string]any{"status": req.Status})
}

// HandleAddVariant adds a variant to an experiment
func (h *ExperimentHandler) HandleAddVariant(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := extractExperimentID(r)
	if id == "" {
		experimentWriteError(w, http.StatusBadRequest, "experiment ID required")
		return
	}

	expID, err := uuid.Parse(id)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid experiment ID")
		return
	}

	var req VariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.PromptTemplate == "" {
		experimentWriteError(w, http.StatusBadRequest, "name and prompt_template are required")
		return
	}

	variant := &factory.ExperimentVariant{
		Name:            req.Name,
		Description:     req.Description,
		PromptTemplate:  req.PromptTemplate,
		Weight:          req.Weight,
		IsControl:       req.IsControl,
		IsActive:        true,
		Metadata:        req.Metadata,
	}

	if variant.Weight == 0 {
		variant.Weight = 50
	}

	if err := h.experimentSvc.AddVariant(r.Context(), expID, variant); err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to add variant")
		return
	}

	experimentWriteJSON(w, http.StatusCreated, variant)
}

// HandleUpdateVariant updates a variant
func (h *ExperimentHandler) HandleUpdateVariant(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	variantID := extractVariantID(r)
	if variantID == "" {
		experimentWriteError(w, http.StatusBadRequest, "variant ID required")
		return
	}

	vid, err := uuid.Parse(variantID)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid variant ID")
		return
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate allowed fields
	allowedFields := map[string]bool{
		"name":            true,
		"description":     true,
		"prompt_template": true,
		"weight":          true,
		"is_control":      true,
		"is_active":       true,
		"metadata":        true,
	}

	for key := range req {
		if !allowedFields[key] {
			experimentWriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid field: %s", key))
			return
		}
	}

	if err := h.experimentSvc.UpdateVariant(r.Context(), vid, req); err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to update variant")
		return
	}

	experimentWriteJSON(w, http.StatusOK, map[string]any{"message": "variant updated"})
}

// HandleDeleteVariant soft-deletes a variant
func (h *ExperimentHandler) HandleDeleteVariant(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	variantID := extractVariantID(r)
	if variantID == "" {
		experimentWriteError(w, http.StatusBadRequest, "variant ID required")
		return
	}

	vid, err := uuid.Parse(variantID)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid variant ID")
		return
	}

	if err := h.experimentSvc.DeleteVariant(r.Context(), vid); err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to delete variant")
		return
	}

	experimentWriteJSON(w, http.StatusOK, map[string]any{"message": "variant deleted"})
}

// HandleGetExperimentStats gets experiment statistics
func (h *ExperimentHandler) HandleGetExperimentStats(w http.ResponseWriter, r *http.Request) {
	id := extractExperimentID(r)
	if id == "" {
		experimentWriteError(w, http.StatusBadRequest, "experiment ID required")
		return
	}

	expID, err := uuid.Parse(id)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid experiment ID")
		return
	}

	stats, err := h.experimentSvc.GetExperimentStats(r.Context(), expID)
	if err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to get experiment stats")
		return
	}

	experimentWriteJSON(w, http.StatusOK, stats)
}

// HandleDetermineWinner determines the winner of an experiment
func (h *ExperimentHandler) HandleDetermineWinner(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserFromContext(r) == nil {
		experimentWriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := extractExperimentID(r)
	if id == "" {
		experimentWriteError(w, http.StatusBadRequest, "experiment ID required")
		return
	}

	expID, err := uuid.Parse(id)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid experiment ID")
		return
	}

	winner, err := h.experimentSvc.DetermineWinner(r.Context(), expID)
	if err != nil {
		experimentWriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	experimentWriteJSON(w, http.StatusOK, winner)
}

// HandleRecordMetric records a metric for a variant
func (h *ExperimentHandler) HandleRecordMetric(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExperimentID    uuid.UUID      `json:"experiment_id"`
		VariantID       uuid.UUID      `json:"variant_id"`
		RunID           *uuid.UUID     `json:"run_id"`
		GenerationID    *uuid.UUID     `json:"generation_id"`
		Success         bool           `json:"success"`
		QualityScore    float64        `json:"quality_score"`
		TestScore       float64        `json:"test_score"`
		AllTestsPassed  bool           `json:"all_tests_passed"`
		LatencyMs       float64        `json:"latency_ms"`
		PromptTokens    *int           `json:"prompt_tokens"`
		CompletionTokens *int          `json:"completion_tokens"`
		TotalTokens     *int           `json:"total_tokens"`
		ErrorMessage    *string        `json:"error_message"`
		Metadata        map[string]any `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ExperimentID == uuid.Nil || req.VariantID == uuid.Nil {
		experimentWriteError(w, http.StatusBadRequest, "experiment_id and variant_id are required")
		return
	}

	metric := &factory.ExperimentMetric{
		ExperimentID:    req.ExperimentID,
		VariantID:      req.VariantID,
		RunID:          req.RunID,
		GenerationID:   req.GenerationID,
		Success:        req.Success,
		QualityScore:   req.QualityScore,
		TestScore:      req.TestScore,
		AllTestsPassed: req.AllTestsPassed,
		LatencyMs:      req.LatencyMs,
		PromptTokens:   req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		TotalTokens:    req.TotalTokens,
		ErrorMessage:   req.ErrorMessage,
		Metadata:       req.Metadata,
	}

	if err := h.experimentSvc.RecordMetric(r.Context(), metric); err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to record metric")
		return
	}

	experimentWriteJSON(w, http.StatusCreated, map[string]any{"message": "metric recorded"})
}

// HandleGetVariantMetrics gets metrics for a variant
func (h *ExperimentHandler) HandleGetVariantMetrics(w http.ResponseWriter, r *http.Request) {
	variantID := extractVariantID(r)
	if variantID == "" {
		experimentWriteError(w, http.StatusBadRequest, "variant ID required")
		return
	}

	vid, err := uuid.Parse(variantID)
	if err != nil {
		experimentWriteError(w, http.StatusBadRequest, "invalid variant ID")
		return
	}

	metrics, err := h.experimentSvc.GetVariantMetrics(r.Context(), vid)
	if err != nil {
		experimentWriteError(w, http.StatusInternalServerError, "failed to get metrics")
		return
	}

	experimentWriteJSON(w, http.StatusOK, metrics)
}

// VariantRequest represents a variant in API requests
type VariantRequest struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	PromptTemplate string         `json:"prompt_template"`
	Weight         int            `json:"weight"`
	IsControl      bool           `json:"is_control"`
	Metadata       map[string]any `json:"metadata"`
}

// Helper functions
func extractExperimentID(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/factory/experiments/")
	if idx := strings.Index(path, "/"); idx > 0 {
		path = path[:idx]
	}
	return path
}

func extractVariantID(r *http.Request) string {
	// URL format: /factory/experiments/{expID}/variants/{variantID}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/factory/experiments/"), "/")
	if len(parts) >= 3 && parts[1] == "variants" {
		return parts[2]
	}
	return ""
}

func parseExperimentLimitOffset(r *http.Request, defaultLimit, maxLimit int) (int, int) {
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

func experimentWriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func experimentWriteError(w http.ResponseWriter, status int, message string) {
	experimentWriteJSON(w, status, map[string]any{"error": message})
}
