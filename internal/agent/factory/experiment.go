package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Experiment represents an A/B test for prompt variants
type Experiment struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string     `json:"name" gorm:"not null;index"`
	Description   string     `json:"description" gorm:"type:text"`
	Status        string     `json:"status" gorm:"not null;default:'draft';index"` // draft, running, paused, completed
	AgentID       string     `json:"agent_id" gorm:"not null;index"`
	WinnerID      *uuid.UUID `json:"winner_id" gorm:"type:uuid"`
	WinnerVariant string     `json:"winner_variant"`
	AutoSelect    bool       `json:"auto_select" gorm:"not null;default:true"`
	// Statistical settings
	MinSamples      int     `json:"min_samples" gorm:"not null;default:10"`
	ConfidenceLevel float64 `json:"confidence_level" gorm:"not null;default:0.95"`
	// Scheduling
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	// Metadata
	Metadata map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at"`
	// Relations
	Variants []ExperimentVariant `json:"variants" gorm:"foreignKey:ExperimentID"`
}

func (Experiment) TableName() string { return "factory_experiments" }

// ExperimentVariant represents a prompt variant in an experiment
type ExperimentVariant struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExperimentID   uuid.UUID `json:"experiment_id" gorm:"type:uuid;not null;index"`
	Name           string    `json:"name" gorm:"not null"` // e.g., "control", "variant_a", "variant_b"
	Description    string    `json:"description" gorm:"type:text"`
	PromptTemplate string    `json:"prompt_template" gorm:"type:text;not null"`
	// Weight for traffic distribution (percentage 0-100)
	Weight int `json:"weight" gorm:"not null;default:50"`
	// Whether this is the control variant
	IsControl bool `json:"is_control" gorm:"not null;default:false"`
	// Whether this variant is currently active
	IsActive bool `json:"is_active" gorm:"not null;default:true"`
	// Metadata
	Metadata map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (ExperimentVariant) TableName() string { return "factory_experiment_variants" }

// ExperimentMetric records metrics for a specific variant
type ExperimentMetric struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExperimentID uuid.UUID  `json:"experiment_id" gorm:"type:uuid;not null;index"`
	VariantID    uuid.UUID  `json:"variant_id" gorm:"type:uuid;not null;index"`
	RunID        *uuid.UUID `json:"run_id" gorm:"type:uuid"`        // Optional link to factory run
	GenerationID *uuid.UUID `json:"generation_id" gorm:"type:uuid"` // Link to generation attempt
	// Core metrics
	Success        bool    `json:"success" gorm:"not null;default:false"`
	QualityScore   float64 `json:"quality_score" gorm:"type:decimal(5,2);default:0"`
	TestScore      float64 `json:"test_score" gorm:"type:decimal(5,2);default:0"`
	AllTestsPassed bool    `json:"all_tests_passed" gorm:"not null;default:false"`
	// Latency metrics (in milliseconds)
	LatencyMs float64 `json:"latency_ms" gorm:"type:decimal(10,2);default:0"`
	// Token usage (if available)
	PromptTokens     *int `json:"prompt_tokens" gorm:"type:int"`
	CompletionTokens *int `json:"completion_tokens" gorm:"type:int"`
	TotalTokens      *int `json:"total_tokens" gorm:"type:int"`
	// Error information
	ErrorMessage *string `json:"error_message" gorm:"type:text"`
	// Metadata
	Metadata map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ExperimentMetric) TableName() string { return "factory_experiment_metrics" }

// ExperimentService handles A/B testing logic
type ExperimentService struct {
	db *gorm.DB
}

// NewExperimentService creates a new experiment service
func NewExperimentService(db *gorm.DB) *ExperimentService {
	return &ExperimentService{db: db}
}

// AutoMigrate creates the experiment tables
func (s *ExperimentService) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&Experiment{}, &ExperimentVariant{}, &ExperimentMetric{})
}

// CreateExperiment creates a new experiment with variants
func (s *ExperimentService) CreateExperiment(ctx context.Context, exp *Experiment) error {
	if len(exp.Variants) < 2 {
		return fmt.Errorf("experiment must have at least 2 variants")
	}
	// Ensure weights sum to 100
	totalWeight := 0
	for i := range exp.Variants {
		totalWeight += exp.Variants[i].Weight
	}
	if totalWeight != 100 {
		// Normalize weights
		for i := range exp.Variants {
			exp.Variants[i].Weight = (exp.Variants[i].Weight * 100) / totalWeight
		}
	}
	return s.db.WithContext(ctx).Create(exp).Error
}

// GetExperiment retrieves an experiment by ID
func (s *ExperimentService) GetExperiment(ctx context.Context, id uuid.UUID) (*Experiment, error) {
	var exp Experiment
	err := s.db.WithContext(ctx).Preload("Variants").First(&exp, "id = ?", id).Error
	return &exp, err
}

// GetExperimentByName retrieves an experiment by name
func (s *ExperimentService) GetExperimentByName(ctx context.Context, name string) (*Experiment, error) {
	var exp Experiment
	err := s.db.WithContext(ctx).Preload("Variants").First(&exp, "name = ?", name).Error
	return &exp, err
}

// ListExperiments lists all experiments with optional filtering
func (s *ExperimentService) ListExperiments(ctx context.Context, agentID string, status string, limit, offset int) ([]Experiment, int64, error) {
	var experiments []Experiment
	var total int64

	query := s.db.WithContext(ctx).Model(&Experiment{}).Where("agent_id = ?", agentID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Variants").Order("created_at DESC").Limit(limit).Offset(offset).Find(&experiments).Error; err != nil {
		return nil, 0, err
	}

	return experiments, total, nil
}

// UpdateExperimentStatus updates the status of an experiment
func (s *ExperimentService) UpdateExperimentStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}

	if status == "running" {
		updates["start_date"] = now
	} else if status == "completed" {
		updates["end_date"] = now
	}

	return s.db.WithContext(ctx).Model(&Experiment{}).Where("id = ?", id).Updates(updates).Error
}

// AddVariant adds a new variant to an experiment
func (s *ExperimentService) AddVariant(ctx context.Context, experimentID uuid.UUID, variant *ExperimentVariant) error {
	variant.ExperimentID = experimentID
	return s.db.WithContext(ctx).Create(variant).Error
}

// UpdateVariant updates an existing variant
func (s *ExperimentService) UpdateVariant(ctx context.Context, variantID uuid.UUID, updates map[string]any) error {
	updates["updated_at"] = time.Now().UTC()
	return s.db.WithContext(ctx).Model(&ExperimentVariant{}).Where("id = ?", variantID).Updates(updates).Error
}

// DeleteVariant soft-deletes a variant
func (s *ExperimentService) DeleteVariant(ctx context.Context, variantID uuid.UUID) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Model(&ExperimentVariant{}).Where("id = ?", variantID).Update("deleted_at", now).Error
}

// AssignVariant assigns a random variant based on weights
func (s *ExperimentService) AssignVariant(ctx context.Context, experimentID uuid.UUID) (*ExperimentVariant, error) {
	var experiment Experiment
	if err := s.db.WithContext(ctx).Preload("Variants", "is_active = ? AND deleted_at IS NULL", true).
		First(&experiment, "id = ? AND status = ?", experimentID, "running").Error; err != nil {
		return nil, err
	}

	if len(experiment.Variants) == 0 {
		return nil, fmt.Errorf("no active variants for experiment")
	}

	// Random assignment based on weights
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalWeight := 0
	for _, v := range experiment.Variants {
		totalWeight += v.Weight
	}

	randomWeight := r.Intn(totalWeight)
	cumulative := 0
	for _, v := range experiment.Variants {
		cumulative += v.Weight
		if randomWeight < cumulative {
			return &v, nil
		}
	}

	// Fallback to last variant
	return &experiment.Variants[len(experiment.Variants)-1], nil
}

// RecordMetric records a metric for a variant
func (s *ExperimentService) RecordMetric(ctx context.Context, metric *ExperimentMetric) error {
	return s.db.WithContext(ctx).Create(metric).Error
}

// GetVariantMetrics retrieves aggregated metrics for a variant
func (s *ExperimentService) GetVariantMetrics(ctx context.Context, variantID uuid.UUID) (VariantMetrics, error) {
	var metrics VariantMetrics

	// Get raw metrics
	var rawMetrics []ExperimentMetric
	if err := s.db.WithContext(ctx).Where("variant_id = ?", variantID).Find(&rawMetrics).Error; err != nil {
		return metrics, err
	}

	if len(rawMetrics) == 0 {
		return metrics, nil
	}

	// Calculate aggregates
	var totalSuccess, totalTestsPassed int
	var totalQuality, totalLatency float64
	var totalPromptTokens, totalCompletionTokens, totalTokens int

	for _, m := range rawMetrics {
		metrics.TotalSamples++
		if m.Success {
			totalSuccess++
		}
		if m.AllTestsPassed {
			totalTestsPassed++
		}
		totalQuality += m.QualityScore
		totalLatency += m.LatencyMs

		if m.PromptTokens != nil {
			totalPromptTokens += *m.PromptTokens
		}
		if m.CompletionTokens != nil {
			totalCompletionTokens += *m.CompletionTokens
		}
		if m.TotalTokens != nil {
			totalTokens += *m.TotalTokens
		}
	}

	metrics.TotalSamples = len(rawMetrics)
	metrics.SuccessRate = float64(totalSuccess) / float64(metrics.TotalSamples) * 100
	metrics.TestPassRate = float64(totalTestsPassed) / float64(metrics.TotalSamples) * 100
	metrics.AverageQualityScore = totalQuality / float64(metrics.TotalSamples)
	metrics.AverageLatencyMs = totalLatency / float64(metrics.TotalSamples)
	metrics.TotalPromptTokens = totalPromptTokens
	metrics.TotalCompletionTokens = totalCompletionTokens
	metrics.TotalTokens = totalTokens

	return metrics, nil
}

// VariantMetrics holds aggregated metrics for a variant
type VariantMetrics struct {
	TotalSamples          int
	SuccessRate           float64
	TestPassRate          float64
	AverageQualityScore   float64
	AverageLatencyMs      float64
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalTokens           int
}

// GetExperimentStats returns statistical analysis for an experiment
func (s *ExperimentService) GetExperimentStats(ctx context.Context, experimentID uuid.UUID) (ExperimentStats, error) {
	var stats ExperimentStats

	// Get experiment with variants
	var experiment Experiment
	if err := s.db.WithContext(ctx).Preload("Variants", "deleted_at IS NULL").
		First(&experiment, "id = ?", experimentID).Error; err != nil {
		return stats, err
	}

	stats.ExperimentID = experimentID
	stats.ExperimentName = experiment.Name
	stats.Status = experiment.Status

	// Get metrics for each variant
	for _, variant := range experiment.Variants {
		metrics, err := s.GetVariantMetrics(ctx, variant.ID)
		if err != nil {
			logrus.WithError(err).Warnf("failed to get metrics for variant %s", variant.ID)
			continue
		}

		variantStat := VariantStat{
			VariantID:           variant.ID,
			VariantName:         variant.Name,
			IsControl:           variant.IsControl,
			TotalSamples:        metrics.TotalSamples,
			SuccessRate:         metrics.SuccessRate,
			TestPassRate:        metrics.TestPassRate,
			AverageQualityScore: metrics.AverageQualityScore,
			AverageLatencyMs:    metrics.AverageLatencyMs,
		}
		stats.Variants = append(stats.Variants, variantStat)
	}

	// Perform statistical analysis if we have enough data
	if len(stats.Variants) >= 2 {
		stats.Analysis = s.analyzeExperiment(ctx, stats.Variants, experiment.MinSamples, experiment.ConfidenceLevel)
	}

	return stats, nil
}

// ExperimentStats holds statistical analysis results
type ExperimentStats struct {
	ExperimentID   uuid.UUID           `json:"experiment_id"`
	ExperimentName string              `json:"experiment_name"`
	Status         string              `json:"status"`
	Variants       []VariantStat       `json:"variants"`
	Analysis       StatisticalAnalysis `json:"analysis"`
}

// VariantStat holds statistics for a single variant
type VariantStat struct {
	VariantID           uuid.UUID `json:"variant_id"`
	VariantName         string    `json:"variant_name"`
	IsControl           bool      `json:"is_control"`
	TotalSamples        int       `json:"total_samples"`
	SuccessRate         float64   `json:"success_rate"`
	TestPassRate        float64   `json:"test_pass_rate"`
	AverageQualityScore float64   `json:"average_quality_score"`
	AverageLatencyMs    float64   `json:"average_latency_ms"`
}

// StatisticalAnalysis holds the results of statistical tests
type StatisticalAnalysis struct {
	WinnerID        *uuid.UUID `json:"winner_id"`
	WinnerName      string     `json:"winner_name"`
	ConfidenceLevel float64    `json:"confidence_level"`
	IsSignificant   bool       `json:"is_significant"`
	PrimaryMetric   string     `json:"primary_metric"` // quality, success_rate, test_pass_rate
	// Detailed comparisons
	Comparisons []VariantComparison `json:"comparisons"`
	// Error message if analysis failed
	Error string `json:"error,omitempty"`
}

// VariantComparison compares two variants
type VariantComparison struct {
	VariantA      string  `json:"variant_a"`
	VariantB      string  `json:"variant_b"`
	Metric        string  `json:"metric"`
	Difference    float64 `json:"difference"`
	RelativeGain  float64 `json:"relative_gain"`
	PValue        float64 `json:"p_value"`
	IsSignificant bool    `json:"is_significant"`
}

// analyzeExperiment performs statistical analysis to determine the winner
func (s *ExperimentService) analyzeExperiment(ctx context.Context, variants []VariantStat, minSamples int, confidenceLevel float64) StatisticalAnalysis {
	analysis := StatisticalAnalysis{
		ConfidenceLevel: confidenceLevel,
		PrimaryMetric:   "quality_score",
	}

	// Find control variant
	var control *VariantStat
	for i := range variants {
		if variants[i].IsControl {
			control = &variants[i]
			break
		}
	}

	// If no control, use first variant as baseline
	if control == nil && len(variants) > 0 {
		control = &variants[0]
	}

	// Check minimum samples
	hasEnoughData := true
	for _, v := range variants {
		if v.TotalSamples < minSamples {
			hasEnoughData = false
			break
		}
	}

	if !hasEnoughData {
		analysis.Error = fmt.Sprintf("not enough samples (minimum %d required per variant)", minSamples)
		return analysis
	}

	// Compare each variant against control
	var bestVariant *VariantStat
	bestScore := -1.0

	for i := range variants {
		v := &variants[i]

		// Skip if same as control
		if control != nil && v.VariantID == control.VariantID {
			continue
		}

		// Calculate comparison
		comparison := VariantComparison{
			VariantA: control.VariantName,
			VariantB: v.VariantName,
			Metric:   "quality_score",
		}

		if control != nil {
			comparison.Difference = v.AverageQualityScore - control.AverageQualityScore
			if control.AverageQualityScore > 0 {
				comparison.RelativeGain = (comparison.Difference / control.AverageQualityScore) * 100
			}
			// Simplified p-value calculation (in production, use proper statistical test)
			comparison.PValue = s.calculatePValue(control.TotalSamples, v.TotalSamples, control.AverageQualityScore, v.AverageQualityScore)
			comparison.IsSignificant = comparison.PValue < (1 - confidenceLevel)
		}

		analysis.Comparisons = append(analysis.Comparisons, comparison)

		// Track best variant based on quality score
		score := v.AverageQualityScore
		if v.SuccessRate > 0 {
			score = score*0.5 + v.SuccessRate*0.3 + v.TestPassRate*0.2
		}

		if score > bestScore {
			bestScore = score
			bestVariant = v
		}
	}

	// Set winner if statistically significant
	if bestVariant != nil && len(analysis.Comparisons) > 0 {
		// Check if the best variant is significantly better than control
		for _, c := range analysis.Comparisons {
			if c.VariantB == bestVariant.VariantName && c.IsSignificant {
				analysis.WinnerID = &bestVariant.VariantID
				analysis.WinnerName = bestVariant.VariantName
				analysis.IsSignificant = true
				break
			}
		}

		// If no significant difference, still pick the best performing one
		if analysis.WinnerID == nil {
			analysis.WinnerID = &bestVariant.VariantID
			analysis.WinnerName = bestVariant.VariantName
			analysis.IsSignificant = false
		}
	}

	return analysis
}

// calculatePValue calculates a simplified p-value using z-test approximation
// In production, use a proper statistical library
func (s *ExperimentService) calculatePValue(n1, n2 int, mean1, mean2 float64) float64 {
	if n1 == 0 || n2 == 0 || mean1 == mean2 {
		return 1.0
	}

	// Calculate standard error (simplified)
	pooledSE := math.Sqrt((1.0/float64(n1) + 1.0/float64(n2)) * ((mean1*mean1 + mean2*mean2) / 2))
	if pooledSE == 0 {
		return 1.0
	}

	z := (mean2 - mean1) / pooledSE

	// Convert z-score to p-value (two-tailed)
	pValue := 2 * (1 - normalCDF(math.Abs(z)))

	return pValue
}

// normalCDF calculates the cumulative distribution function of standard normal
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

// GenerationExperimentAdapter provides integration between experiments and generation service
type GenerationExperimentAdapter struct {
	db            *gorm.DB
	experimentSvc *ExperimentService
}

// NewGenerationExperimentAdapter creates a new generation experiment adapter
func NewGenerationExperimentAdapter(db *gorm.DB, experimentSvc *ExperimentService) *GenerationExperimentAdapter {
	return &GenerationExperimentAdapter{
		db:            db,
		experimentSvc: experimentSvc,
	}
}

// GetPromptForGeneration returns the appropriate prompt for a generation request
// It checks if there's an active experiment for the agent and assigns a variant
func (a *GenerationExperimentAdapter) GetPromptForGeneration(ctx context.Context, agentID, basePrompt string) (prompt string, variantID *uuid.UUID, experimentID *uuid.UUID, err error) {
	// Try to get active experiment for this agent
	exp, err := a.experimentSvc.GetActiveExperiment(ctx, agentID)
	if err != nil || exp == nil {
		// No active experiment, use base prompt
		return basePrompt, nil, nil, nil
	}

	// Assign a variant
	variant, err := a.experimentSvc.AssignVariant(ctx, exp.ID)
	if err != nil || variant == nil {
		// Assignment failed, use base prompt
		return basePrompt, nil, nil, nil
	}

	// Replace placeholders in prompt template with base prompt
	prompt = strings.ReplaceAll(variant.PromptTemplate, "{{prompt}}", basePrompt)
	prompt = strings.ReplaceAll(prompt, "{{description}}", basePrompt)

	return prompt, &variant.ID, &exp.ID, nil
}

// RecordGenerationResult records the result of a generation for metric tracking
func (a *GenerationExperimentAdapter) RecordGenerationResult(ctx context.Context, experimentID, variantID uuid.UUID, result GenerationResult) error {
	if experimentID == uuid.Nil || variantID == uuid.Nil {
		return nil // Skip if no experiment tracking
	}

	metric := &ExperimentMetric{
		ExperimentID:   experimentID,
		VariantID:      variantID,
		GenerationID:   &result.FunctionID,
		Success:        result.Success,
		QualityScore:   result.QualityScore,
		TestScore:      result.TestScore,
		AllTestsPassed: result.AllTestsPassed,
		LatencyMs:      result.LatencyMs,
		ErrorMessage:   result.Error,
	}

	if result.PromptTokens != nil {
		metric.PromptTokens = result.PromptTokens
	}
	if result.CompletionTokens != nil {
		metric.CompletionTokens = result.CompletionTokens
	}
	if result.TotalTokens != nil {
		metric.TotalTokens = result.TotalTokens
	}

	return a.experimentSvc.RecordMetric(ctx, metric)
}

// GenerationResult represents the result of a generation for metric tracking
type GenerationResult struct {
	FunctionID       uuid.UUID
	Success          bool
	QualityScore     float64
	TestScore        float64
	AllTestsPassed   bool
	LatencyMs        float64
	Error            *string
	PromptTokens     *int
	CompletionTokens *int
	TotalTokens      *int
}

// DetermineWinner determines and sets the winner for an experiment
func (s *ExperimentService) DetermineWinner(ctx context.Context, experimentID uuid.UUID) (*ExperimentVariant, error) {
	stats, err := s.GetExperimentStats(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	if stats.Analysis.WinnerID == nil {
		return nil, fmt.Errorf("no winner could be determined")
	}

	// Update experiment with winner
	var winner ExperimentVariant
	if err := s.db.WithContext(ctx).First(&winner, "id = ?", *stats.Analysis.WinnerID).Error; err != nil {
		return nil, err
	}

	// Update experiment
	updates := map[string]any{
		"winner_id":      *stats.Analysis.WinnerID,
		"winner_variant": stats.Analysis.WinnerName,
		"status":         "completed",
		"updated_at":     time.Now().UTC(),
		"end_date":       time.Now().UTC(),
	}

	if err := s.db.WithContext(ctx).Model(&Experiment{}).Where("id = ?", experimentID).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Deactivate all variants except winner
	if err := s.db.WithContext(ctx).Model(&ExperimentVariant{}).
		Where("experiment_id = ? AND id != ?", experimentID, *stats.Analysis.WinnerID).
		Updates(map[string]any{"is_active": false, "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, err
	}

	return &winner, nil
}

// GetActiveExperiment returns the currently active experiment for an agent
func (s *ExperimentService) GetActiveExperiment(ctx context.Context, agentID string) (*Experiment, error) {
	var exp Experiment
	err := s.db.WithContext(ctx).Preload("Variants", "is_active = ? AND deleted_at IS NULL", true).
		First(&exp, "agent_id = ? AND status = ?", agentID, "running").
		Error
	return &exp, err
}

// MarshalJSON implements custom JSON marshaling
func (e Experiment) MarshalJSON() ([]byte, error) {
	type Alias Experiment
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias:     (*Alias)(&e),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
	return json.Marshal(aux)
}
