package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ============================================
// Trust Score Calculation Methods
// ============================================

// CalculateTrustScore computes the trust score for a function based on execution metrics
func (r *RegistryRepository) CalculateTrustScore(functionID uuid.UUID, windowStart, windowEnd time.Time) (*TrustHistory, error) {
	// Get execution metrics for the window
	metrics, err := r.GetExecutionMetrics(functionID, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution metrics: %w", err)
	}

	// Get consumer diversity
	uniqueIPs, uniqueTenants, uniqueUsers, err := r.GetConsumerDiversity(functionID, windowStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer diversity: %w", err)
	}

	// Get verification status
	isVerified, verificationLevel, err := r.GetFunctionVerificationStatus(functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification status: %w", err)
	}

	// Get user rating
	rating, err := r.GetRatingByFunctionID(functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	// Get weights
	weights := DefaultTrustScoreWeights()

	// Calculate component scores
	reliabilityScore := calculateReliabilityScore(metrics)
	latencyScore := calculateLatencyScore(metrics)
	errorRateScore := calculateErrorRateScore(metrics)
	userRatingScore := calculateUserRatingScore(rating)
	verificationBonus := calculateVerificationBonus(isVerified, verificationLevel)

	// Calculate overall trust score
	trustScore := (reliabilityScore * weights.Reliability) +
		(latencyScore * weights.Latency) +
		(errorRateScore * weights.ErrorRate) +
		(userRatingScore * weights.UserRating) +
		(verificationBonus * weights.Verification)

	// Determine trust tier
	trustTier := determineTrustTier(trustScore, isVerified)

	// Create trust history entry
	history := &TrustHistory{
		ID:                 uuid.New(),
		FunctionID:         functionID,
		TrustScore:         trustScore,
		ReliabilityScore:   reliabilityScore,
		LatencyScore:       latencyScore,
		ErrorRateScore:     errorRateScore,
		UserRatingScore:    userRatingScore,
		VerificationBonus:  verificationBonus,
		TotalCalls:         metrics.TotalCalls,
		SuccessRate:        metrics.SuccessRate,
		P50LatencyMs:       metrics.LatencyP50,
		P95LatencyMs:       metrics.LatencyP95,
		P99LatencyMs:       metrics.LatencyP99,
		ErrorRate:          metrics.ErrorRate,
		TimeoutRate:        metrics.TimeoutRate,
		ConsumerDiversity:   uniqueIPs,
		TenantDiversity:    uniqueTenants,
		UserDiversity:      uniqueUsers,
		IsVerified:         isVerified,
		VerificationLevel:  verificationLevel,
		TrustTier:          trustTier,
		CalculatedAt:       time.Now(),
		WindowStart:         windowStart,
		WindowEnd:           windowEnd,
		CalculationVersion:  1,
	}

	return history, nil
}

// calculateReliabilityScore calculates the reliability component score (0-100)
func calculateReliabilityScore(metrics *ExecutionMetrics) float64 {
	if metrics == nil || metrics.TotalCalls == 0 {
		return 50.0 // Default neutral score
	}
	// Success rate is already calculated as a percentage (0-100)
	return metrics.SuccessRate
}

// calculateLatencyScore calculates the latency component score (0-100)
// Lower latency = higher score
func calculateLatencyScore(metrics *ExecutionMetrics) float64 {
	if metrics == nil || metrics.TotalCalls == 0 {
		return 50.0 // Default neutral score
	}

	p95 := float64(metrics.LatencyP95)

	// Scoring based on p95 latency
	if p95 < 50 {
		return 100.0
	} else if p95 < 100 {
		return 90.0
	} else if p95 < 200 {
		return 80.0
	} else if p95 < 500 {
		return 70.0
	} else if p95 < 1000 {
		return 60.0
	} else if p95 < 2000 {
		return 50.0
	} else if p95 < 5000 {
		return 40.0
	} else {
		return 30.0
	}
}

// calculateErrorRateScore calculates the error rate component score (0-100)
// Lower error rate = higher score
func calculateErrorRateScore(metrics *ExecutionMetrics) float64 {
	if metrics == nil || metrics.TotalCalls == 0 {
		return 50.0 // Default neutral score
	}

	errorRate := metrics.ErrorRate + metrics.TimeoutRate

	// Scoring based on combined error + timeout rate
	if errorRate == 0 {
		return 100.0
	} else if errorRate < 0.1 {
		return 95.0
	} else if errorRate < 0.5 {
		return 85.0
	} else if errorRate < 1.0 {
		return 75.0
	} else if errorRate < 2.0 {
		return 60.0
	} else if errorRate < 5.0 {
		return 40.0
	} else {
		return 20.0
	}
}

// calculateUserRatingScore calculates the user rating component score (0-100)
func calculateUserRatingScore(rating *RegistryFunctionRating) float64 {
	if rating == nil || rating.TotalRatings == 0 {
		return 50.0 // Default neutral score when no ratings
	}
	return rating.OverallScore
}

// calculateVerificationBonus calculates the verification bonus (0-15)
func calculateVerificationBonus(isVerified bool, level string) float64 {
	if !isVerified {
		return 0.0
	}

	switch level {
	case "basic":
		return 5.0
	case "standard":
		return 10.0
	case "enterprise":
		return 15.0
	default:
		return 5.0
	}
}

// determineTrustTier determines the trust tier based on trust score and verification
func determineTrustTier(trustScore float64, isVerified bool) TrustTier {
	if trustScore >= 90 {
		if isVerified {
			return TrustTierHighlyTrusted
		}
		return TrustTierVerified
	} else if trustScore >= 70 {
		return TrustTierVerified
	} else if trustScore >= 50 {
		return TrustTierTrusted
	}
	return TrustTierUntrusted
}

// ============================================
// Trust History Methods
// ============================================

// CreateTrustHistory creates a new trust history entry
func (r *RegistryRepository) CreateTrustHistory(history *TrustHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create trust history: %w", err)
	}
	return nil
}

// GetTrustHistory retrieves trust history for a function
func (r *RegistryRepository) GetTrustHistory(functionID uuid.UUID, limit, offset int) ([]TrustHistory, int, error) {
	var history []TrustHistory
	var total int64

	// Count total records
	if err := r.db.Model(&TrustHistory{}).Where("function_id = ?", functionID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count trust history: %w", err)
	}

	// Get paginated results
	if err := r.db.Where("function_id = ?", functionID).
		Order("calculated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get trust history: %w", err)
	}

	return history, int(total), nil
}

// GetLatestTrustHistory retrieves the most recent trust history for a function
func (r *RegistryRepository) GetLatestTrustHistory(functionID uuid.UUID) (*TrustHistory, error) {
	var history TrustHistory
	if err := r.db.Where("function_id = ?", functionID).
		Order("calculated_at DESC").
		First(&history).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest trust history: %w", err)
	}
	return &history, nil
}

// ============================================
// Execution Metrics Methods
// ============================================

// CreateOrUpdateExecutionMetrics creates or updates execution metrics
func (r *RegistryRepository) CreateOrUpdateExecutionMetrics(metrics *ExecutionMetrics) error {
	metrics.UpdatedAt = time.Now()

	// Try to find existing metrics for this function and window
	var existing ExecutionMetrics
	err := r.db.Where("function_id = ? AND window_start = ? AND window_type = ?",
		metrics.FunctionID, metrics.WindowStart, metrics.WindowType).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		metrics.ID = uuid.New()
		metrics.CreatedAt = time.Now()
		if err := r.db.Create(metrics).Error; err != nil {
			return fmt.Errorf("failed to create execution metrics: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check existing metrics: %w", err)
	}

	// Update existing
	metrics.ID = existing.ID
	metrics.CreatedAt = existing.CreatedAt
	if err := r.db.Save(metrics).Error; err != nil {
		return fmt.Errorf("failed to update execution metrics: %w", err)
	}

	return nil
}

// GetExecutionMetrics retrieves execution metrics for a function and time window
func (r *RegistryRepository) GetExecutionMetrics(functionID uuid.UUID, windowStart, windowEnd time.Time) (*ExecutionMetrics, error) {
	// First try to get aggregated metrics from the execution_metrics table
	var metrics ExecutionMetrics
	err := r.db.Where("function_id = ? AND window_start >= ? AND window_end <= ?",
		functionID, windowStart, windowEnd).
		Order("window_start DESC").
		First(&metrics).Error

	if err == gorm.ErrRecordNotFound {
		// Fall back to calculating from raw executions
		return r.calculateMetricsFromExecutions(functionID, windowStart, windowEnd)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get execution metrics: %w", err)
	}

	return &metrics, nil
}

// calculateMetricsFromExecutions calculates metrics from raw execution data
func (r *RegistryRepository) calculateMetricsFromExecutions(functionID uuid.UUID, windowStart, windowEnd time.Time) (*ExecutionMetrics, error) {
	var result struct {
		TotalCalls    int     `json:"total_calls"`
		SuccessRate   float64 `json:"success_rate"`
		AvgLatency    float64 `json:"avg_latency"`
		P50Latency    int     `json:"p50_latency"`
		P95Latency    int     `json:"p95_latency"`
		P99Latency    int     `json:"p99_latency"`
		TimeoutRate   float64 `json:"timeout_rate"`
		ErrorRate     float64 `json:"error_rate"`
		UniqueIPs     int     `json:"unique_ips"`
		UniqueTenants int     `json:"unique_tenants"`
		UniqueUsers   int     `json:"unique_users"`
	}

	query := `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p50_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p95_latency,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p99_latency,
			COALESCE(SUM(CASE WHEN outcome = 'timeout' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as timeout_rate,
			COALESCE(SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as error_rate,
			COUNT(DISTINCT caller_ip) as unique_ips,
			COUNT(DISTINCT tenant_id) as unique_tenants,
			COUNT(DISTINCT user_id) as unique_users
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp >= ? AND timestamp <= ?
	`

	if err := r.db.Raw(query, functionID, windowStart, windowEnd).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate metrics from executions: %w", err)
	}

	return &ExecutionMetrics{
		ID:              uuid.New(),
		FunctionID:      functionID,
		WindowStart:     windowStart,
		WindowEnd:       windowEnd,
		WindowType:      "hourly",
		TotalCalls:      result.TotalCalls,
		SuccessfulCalls: int(float64(result.TotalCalls) * result.SuccessRate / 100),
		SuccessRate:     result.SuccessRate,
		LatencyAvg:      result.AvgLatency,
		LatencyP50:     result.P50Latency,
		LatencyP95:     result.P95Latency,
		LatencyP99:     result.P99Latency,
		TimeoutRate:     result.TimeoutRate,
		ErrorRate:       result.ErrorRate,
		UniqueIPs:       result.UniqueIPs,
		UniqueTenants:   result.UniqueTenants,
		UniqueUsers:     result.UniqueUsers,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

// AggregateHourlyMetrics aggregates execution data into hourly metrics
func (r *RegistryRepository) AggregateHourlyMetrics(hour time.Time) error {
	windowStart := hour.Truncate(time.Hour)
	windowEnd := windowStart.Add(time.Hour)

	// Get all functions with executions in this window
	var functionIDs []uuid.UUID
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("timestamp >= ? AND timestamp < ?", windowStart, windowEnd).
		Distinct("function_id").
		Pluck("function_id", &functionIDs).Error; err != nil {
		return fmt.Errorf("failed to get functions with executions: %w", err)
	}

	for _, functionID := range functionIDs {
		metrics, err := r.calculateMetricsFromExecutions(functionID, windowStart, windowEnd)
		if err != nil {
			logrus.Errorf("Failed to calculate metrics for function %s: %v", functionID, err)
			continue
		}

		if err := r.CreateOrUpdateExecutionMetrics(metrics); err != nil {
			logrus.Errorf("Failed to save metrics for function %s: %v", functionID, err)
		}
	}

	return nil
}

// ============================================
// Trust Score Update Methods
// ============================================

// UpdateFunctionTrustScore updates the trust score on the function record
func (r *RegistryRepository) UpdateFunctionTrustScore(functionID uuid.UUID, trustScore float64, trustTier TrustTier) error {
	now := time.Now()
	updates := map[string]interface{}{
		"trust_score":                trustScore,
		"trust_tier":                 trustTier,
		"trust_updated_at":           now,
		"trust_calculation_version":   gorm.Expr("trust_calculation_version + 1"),
	}

	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", functionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update function trust score: %w", err)
	}

	return nil
}

// RecalculateTrustScore recalculates and updates trust score for a single function
func (r *RegistryRepository) RecalculateTrustScore(functionID uuid.UUID) error {
	// Get the time window (last 100 executions or last 24 hours, whichever is longer)
	since := time.Now().Add(-24 * time.Hour)

	history, err := r.CalculateTrustScore(functionID, since, time.Now())
	if err != nil {
		return fmt.Errorf("failed to calculate trust score: %w", err)
	}

	// Save to history
	if err := r.CreateTrustHistory(history); err != nil {
		return fmt.Errorf("failed to create trust history: %w", err)
	}

	// Update function record
	if err := r.UpdateFunctionTrustScore(functionID, history.TrustScore, history.TrustTier); err != nil {
		return fmt.Errorf("failed to update function trust score: %w", err)
	}

	return nil
}

// ============================================
// Verification Status Methods
// ============================================

// GetFunctionVerificationStatus retrieves verification status for a function
func (r *RegistryRepository) GetFunctionVerificationStatus(functionID uuid.UUID) (bool, string, error) {
	// Get the latest version's verification status
	var version RegistryFunctionVersion
	if err := r.db.Where("function_id = ?", functionID).
		Order("published_at DESC").
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "none", nil
		}
		return false, "none", fmt.Errorf("failed to get latest version: %w", err)
	}

	var status RegistryFunctionVerificationStatus
	if err := r.db.Where("function_version_id = ?", version.ID).First(&status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "none", nil
		}
		return false, "none", fmt.Errorf("failed to get verification status: %w", err)
	}

	isVerified := status.OverallStatus == "verified"
	level := "basic"
	if status.ApprovalStatus == "approved" {
		level = "standard"
	}
	if status.MalwareScanned && status.MalwareStatus == "clean" {
		level = "enterprise"
	}

	return isVerified, level, nil
}

// ============================================
// Trust Score Job Methods
// ============================================

// CreateTrustScoreJob creates a new trust score job record
func (r *RegistryRepository) CreateTrustScoreJob(job *TrustScoreJob) error {
	job.ID = uuid.New()
	job.CreatedAt = time.Now()
	job.Errors, _ = json.Marshal([]string{})

	if err := r.db.Create(job).Error; err != nil {
		return fmt.Errorf("failed to create trust score job: %w", err)
	}
	return nil
}

// UpdateTrustScoreJob updates a trust score job
func (r *RegistryRepository) UpdateTrustScoreJob(jobID uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.Model(&TrustScoreJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update trust score job: %w", err)
	}
	return nil
}

// GetTrustScoreJob retrieves a trust score job by ID
func (r *RegistryRepository) GetTrustScoreJob(jobID uuid.UUID) (*TrustScoreJob, error) {
	var job TrustScoreJob
	if err := r.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		return nil, fmt.Errorf("failed to get trust score job: %w", err)
	}
	return &job, nil
}

// ============================================
// Trust Score Refresh Methods
// ============================================

// RefreshAllTrustScores recalculates trust scores for all functions
func (r *RegistryRepository) RefreshAllTrustScores() (*TrustScoreJob, error) {
	// Create job record
	job := &TrustScoreJob{
		JobType:        "full_recalculation",
		Status:         "running",
		StartedAt:      timePtr(time.Now()),
		FunctionsTotal: 0,
	}
	if err := r.CreateTrustScoreJob(job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Get all functions
	var functions []RegistryFunction
	if err := r.db.Find(&functions).Error; err != nil {
		r.UpdateTrustScoreJob(job.ID, map[string]interface{}{
			"status": "failed",
		})
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}

	job.FunctionsTotal = len(functions)
	r.UpdateTrustScoreJob(job.ID, map[string]interface{}{
		"functions_total": len(functions),
	})

	var errors []string
	processed := 0

	for _, fn := range functions {
		if err := r.RecalculateTrustScore(fn.ID); err != nil {
			errors = append(errors, fmt.Sprintf("function %s: %v", fn.ID, err))
			logrus.Errorf("Failed to recalculate trust score for function %s: %v", fn.ID, err)
		}
		processed++

		// Update progress every 100 functions
		if processed%100 == 0 {
			r.UpdateTrustScoreJob(job.ID, map[string]interface{}{
				"functions_processed": processed,
			})
		}
	}

	// Complete the job
	status := "completed"
	if len(errors) > 0 {
		status = "failed"
	}
	now := time.Now()
	r.UpdateTrustScoreJob(job.ID, map[string]interface{}{
		"status":               status,
		"functions_processed":  processed,
		"errors":               errors,
		"completed_at":         &now,
	})

	job.Status = status
	job.FunctionsProcessed = processed
	return job, nil
}

// Helper function to get pointer to time
func timePtr(t time.Time) *time.Time {
	return &t
}

// ============================================
// Cache Helper Methods
// ============================================

// InvalidateTrustScoreCache invalidates the trust score cache for a function
// Uses the rating cache key as a proxy since trust scores are derived from ratings
func (r *RegistryRepository) InvalidateTrustScoreCache(functionID uuid.UUID) error {
	if r.cache != nil && r.keyGen != nil {
		// Trust scores are derived from ratings, so invalidate the rating cache
		cacheKey := r.keyGen.FunctionRating(functionID.String())
		go func() {
			if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
				logrus.Errorf("Failed to invalidate trust score cache: %v", err)
			}
		}()
	}
	return nil
}
