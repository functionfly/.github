package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		ConsumerDiversity:  uniqueIPs,
		TenantDiversity:    uniqueTenants,
		UserDiversity:      uniqueUsers,
		IsVerified:         isVerified,
		VerificationLevel:  verificationLevel,
		TrustTier:          trustTier,
		CalculatedAt:       time.Now(),
		WindowStart:        windowStart,
		WindowEnd:          windowEnd,
		CalculationVersion: 1,
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
		LatencyP50:      result.P50Latency,
		LatencyP95:      result.P95Latency,
		LatencyP99:      result.P99Latency,
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

	if len(functionIDs) == 0 {
		return nil
	}

	// Batch calculate all metrics in a single query
	allMetrics, err := r.batchCalculateMetricsFromExecutions(functionIDs, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to batch calculate metrics: %w", err)
	}

	for _, metrics := range allMetrics {
		if err := r.CreateOrUpdateExecutionMetrics(metrics); err != nil {
			logrus.WithError(err).WithField("function_id", metrics.FunctionID).Error("Failed to save metrics")
		}
	}

	return nil
}

// batchCalculateMetricsFromExecutions calculates metrics for multiple functions in a single query
func (r *RegistryRepository) batchCalculateMetricsFromExecutions(functionIDs []uuid.UUID, windowStart, windowEnd time.Time) (map[uuid.UUID]*ExecutionMetrics, error) {
	results := make(map[uuid.UUID]*ExecutionMetrics, len(functionIDs))

	query := `
		SELECT
			function_id,
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
		WHERE function_id = ANY($1) AND timestamp >= $2 AND timestamp <= $3
		GROUP BY function_id
	`

	rows, err := r.db.Raw(query, pq.Array(functionIDs), windowStart, windowEnd).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to batch calculate metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var functionID uuid.UUID
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

		if err := rows.Scan(&functionID, &result.TotalCalls, &result.SuccessRate, &result.AvgLatency, &result.P50Latency, &result.P95Latency, &result.P99Latency, &result.TimeoutRate, &result.ErrorRate, &result.UniqueIPs, &result.UniqueTenants, &result.UniqueUsers); err != nil {
			logrus.WithError(err).Warn("Failed to scan batch metrics row")
			continue
		}

		results[functionID] = &ExecutionMetrics{
			ID:              uuid.New(),
			FunctionID:      functionID,
			WindowStart:     windowStart,
			WindowEnd:       windowEnd,
			WindowType:      "hourly",
			TotalCalls:      result.TotalCalls,
			SuccessfulCalls: int(float64(result.TotalCalls) * result.SuccessRate / 100),
			SuccessRate:     result.SuccessRate,
			LatencyAvg:      result.AvgLatency,
			LatencyP50:      result.P50Latency,
			LatencyP95:      result.P95Latency,
			LatencyP99:      result.P99Latency,
			TimeoutRate:     result.TimeoutRate,
			TimeoutCalls:    int(float64(result.TotalCalls) * result.TimeoutRate / 100),
			ErrorRate:       result.ErrorRate,
			ErrorCalls:      int(float64(result.TotalCalls) * result.ErrorRate / 100),
			UniqueIPs:       result.UniqueIPs,
			UniqueTenants:   result.UniqueTenants,
			UniqueUsers:     result.UniqueUsers,
			CreatedAt:       time.Now(),
		}
	}

	return results, nil
}

// ============================================
// Trust Score Update Methods
// ============================================

// UpdateFunctionTrustScore updates the trust score on the function record
func (r *RegistryRepository) UpdateFunctionTrustScore(functionID uuid.UUID, trustScore float64, trustTier TrustTier) error {
	now := time.Now()
	updates := map[string]interface{}{
		"trust_score":               trustScore,
		"trust_tier":                trustTier,
		"trust_updated_at":          now,
		"trust_calculation_version": gorm.Expr("trust_calculation_version + 1"),
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
// Sliding Window Calculation Methods
// ============================================

// CalculateTrustScoreSliding computes a trust score using sliding window with exponential smoothing
func (r *RegistryRepository) CalculateTrustScoreSliding(functionID uuid.UUID, config SlidingWindowConfig) (*TrustHistory, *TrustScoreDelta, error) {
	now := time.Now()
	windowStart := now.Add(-config.WindowDuration)

	// Get current sliding window state
	state, err := r.getOrCreateSlidingWindowState(functionID, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sliding window state: %w", err)
	}

	// Calculate metrics for the sliding window
	metrics, err := r.calculateMetricsFromExecutions(functionID, windowStart, now)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate window metrics: %w", err)
	}

	// Check minimum data points
	if metrics.TotalCalls < config.MinDataPoints {
		// Not enough data - use previous score or neutral default
		if state.CurrentScore > 0 {
			return nil, nil, fmt.Errorf("insufficient data points: %d < %d", metrics.TotalCalls, config.MinDataPoints)
		}
	}

	// Get other trust components
	isVerified, verificationLevel, err := r.GetFunctionVerificationStatus(functionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get verification status: %w", err)
	}

	rating, err := r.GetRatingByFunctionID(functionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rating: %w", err)
	}

	// Calculate component scores
	reliabilityScore := calculateReliabilityScore(metrics)
	latencyScore := calculateLatencyScore(metrics)
	errorRateScore := calculateErrorRateScore(metrics)
	userRatingScore := calculateUserRatingScore(rating)
	verificationBonus := calculateVerificationBonus(isVerified, verificationLevel)

	// Get diversity metrics
	uniqueIPs, uniqueTenants, uniqueUsers, err := r.GetConsumerDiversity(functionID, windowStart)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consumer diversity: %w", err)
	}

	// Calculate new score with weighted components
	weights := DefaultTrustScoreWeights()
	newScore := (reliabilityScore * weights.Reliability) +
		(latencyScore * weights.Latency) +
		(errorRateScore * weights.ErrorRate) +
		(userRatingScore * weights.UserRating) +
		(verificationBonus * weights.Verification)

	// Apply exponential smoothing if we have a previous score
	var finalScore float64
	if state.CurrentScore > 0 {
		// EMA: S_t = alpha * Y_t + (1-alpha) * S_{t-1}
		finalScore = config.SmoothingFactor*newScore + (1-config.SmoothingFactor)*state.CurrentScore
	} else {
		finalScore = newScore
	}

	// Determine trust tier
	newTier := determineTrustTier(finalScore, isVerified)
	oldTier := determineTrustTier(state.CurrentScore, isVerified)

	// Create delta
	delta := &TrustScoreDelta{
		FunctionID:         functionID,
		PreviousScore:      state.CurrentScore,
		CurrentScore:       finalScore,
		ScoreChange:        finalScore - state.CurrentScore,
		ScoreChangePercent: 0,
		PreviousTier:       oldTier,
		CurrentTier:        newTier,
		TierChanged:        oldTier != newTier,
		CalculatedAt:       now,
		WindowType:         WindowTypeSliding,
		ComponentChanges: map[string]float64{
			"reliability":  reliabilityScore - state.GetComponentScore("reliability"),
			"latency":      latencyScore - state.GetComponentScore("latency"),
			"error_rate":   errorRateScore - state.GetComponentScore("error_rate"),
			"user_rating":  userRatingScore - state.GetComponentScore("user_rating"),
			"verification": verificationBonus - state.GetComponentScore("verification"),
		},
	}

	if state.CurrentScore > 0 {
		delta.ScoreChangePercent = (delta.ScoreChange / state.CurrentScore) * 100
	}

	// Create history entry
	history := &TrustHistory{
		ID:                 uuid.New(),
		FunctionID:         functionID,
		TrustScore:         finalScore,
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
		ConsumerDiversity:  uniqueIPs,
		TenantDiversity:    uniqueTenants,
		UserDiversity:      uniqueUsers,
		IsVerified:         isVerified,
		VerificationLevel:  verificationLevel,
		TrustTier:          newTier,
		CalculatedAt:       now,
		WindowStart:        windowStart,
		WindowEnd:          now,
		CalculationVersion: 2, // Sliding window version
	}

	// Update sliding window state with component scores
	state.PreviousScore = state.CurrentScore
	state.CurrentScore = finalScore
	state.LastUpdated = now
	state.WindowStart = windowStart
	state.WindowEnd = now
	state.TotalCallsInWindow = metrics.TotalCalls
	state.LastCalculation = now

	// Store component scores for next delta calculation
	state.ReliabilityScore = reliabilityScore
	state.LatencyScore = latencyScore
	state.ErrorRateScore = errorRateScore
	state.UserRatingScore = userRatingScore
	state.VerificationBonus = verificationBonus

	if err := r.saveSlidingWindowState(state); err != nil {
		logrus.WithError(err).Warn("Failed to save sliding window state")
	}

	return history, delta, nil
}

// getOrCreateSlidingWindowState retrieves or creates sliding window state for a function
func (r *RegistryRepository) getOrCreateSlidingWindowState(functionID uuid.UUID, config SlidingWindowConfig) (*SlidingWindowState, error) {
	var state SlidingWindowState
	err := r.db.Where("function_id = ?", functionID).First(&state).Error

	if err == gorm.ErrRecordNotFound {
		// Create new state
		now := time.Now()
		state = SlidingWindowState{
			FunctionID:      functionID,
			LastUpdated:     now,
			WindowStart:     now.Add(-config.WindowDuration),
			WindowEnd:       now,
			SmoothingFactor: config.SmoothingFactor,
		}
		return &state, nil
	} else if err != nil {
		return nil, err
	}

	return &state, nil
}

// saveSlidingWindowState saves the sliding window state
func (r *RegistryRepository) saveSlidingWindowState(state *SlidingWindowState) error {
	return r.db.Save(state).Error
}

// UpdateSlidingWindowScores updates scores for all functions using sliding window
func (r *RegistryRepository) UpdateSlidingWindowScores(config SlidingWindowConfig) ([]TrustScoreDelta, error) {
	var functions []RegistryFunction
	if err := r.db.Find(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}

	if len(functions) == 0 {
		return nil, nil
	}

	var deltas []TrustScoreDelta

	for _, fn := range functions {
		history, delta, err := r.CalculateTrustScoreSliding(fn.ID, config)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to calculate sliding window score for %s", fn.ID)
			continue
		}

		// Save to history
		if err := r.CreateTrustHistory(history); err != nil {
			logrus.WithError(err).Warnf("Failed to save trust history for %s", fn.ID)
		}

		// Update function record
		if err := r.UpdateFunctionTrustScore(fn.ID, history.TrustScore, history.TrustTier); err != nil {
			logrus.WithError(err).Warnf("Failed to update function trust score for %s", fn.ID)
		}

		if delta.ScoreChange != 0 || delta.TierChanged {
			deltas = append(deltas, *delta)
		}
	}

	return deltas, nil
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
		"status":              status,
		"functions_processed": processed,
		"errors":              errors,
		"completed_at":        &now,
	})

	job.Status = status
	job.FunctionsProcessed = processed
	return job, nil
}

// Helper function to get pointer to time
func timePtr(t time.Time) *time.Time {
	return &t
}

// GetAllFunctionsWithTrustScores retrieves all functions with their current trust scores
// Used for delta tracking in streaming updates
func (r *RegistryRepository) GetAllFunctionsWithTrustScores() ([]RegistryFunction, error) {
	var functions []RegistryFunction
	// Select only the fields we need for delta tracking
	if err := r.db.Select("id, trust_score, trust_tier, trust_updated_at").Find(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get functions with trust scores: %w", err)
	}
	return functions, nil
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
			defer func() {
				if rec := recover(); rec != nil {
					logrus.WithFields(logrus.Fields{
						"panic": rec,
						"stack": fmt.Sprintf("%v", rec),
					}).Error("InvalidateTrustScoreCache goroutine panicked")
				}
			}()
			if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
				logrus.Errorf("Failed to invalidate trust score cache: %v", err)
			}
		}()
	}
	return nil
}

// ============================================
// Remix History Methods
// ============================================

// RecordRemix creates a remix history record linking source and target functions
func (r *RegistryRepository) RecordRemix(sourceFunctionID, targetFunctionID, remixedByUserID uuid.UUID, customization string, costUSD float64) error {
	remix := &RemixHistory{
		ID:               uuid.New(),
		SourceFunctionID: sourceFunctionID,
		TargetFunctionID: targetFunctionID,
		RemixedByUserID:  remixedByUserID,
		RemixedAt:        time.Now(),
		Customization:    customization,
		CostUSD:          costUSD,
	}
	if err := r.db.Create(remix).Error; err != nil {
		return fmt.Errorf("failed to record remix: %w", err)
	}
	return nil
}

// GetRemixHistoryForFunction returns all remixes created from this function
func (r *RegistryRepository) GetRemixHistoryForFunction(sourceFunctionID uuid.UUID) ([]RemixHistory, error) {
	var history []RemixHistory
	if err := r.db.Where("source_function_id = ?", sourceFunctionID).
		Order("remixed_at DESC").
		Preload("TargetFunction").
		Find(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to get remix history: %w", err)
	}
	return history, nil
}

// CountRemixesForFunction returns the number of remixes for a function
func (r *RegistryRepository) CountRemixesForFunction(sourceFunctionID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&RemixHistory{}).Where("source_function_id = ?", sourceFunctionID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count remixes: %w", err)
	}
	return count, nil
}

// IsFunctionRemix checks if a function was created by remixing another function
// Returns true if the function appears as a target in remix_history
func (r *RegistryRepository) IsFunctionRemix(targetFunctionID uuid.UUID) (bool, *RemixHistory, error) {
	var history RemixHistory
	err := r.db.Where("target_function_id = ?", targetFunctionID).
		Preload("SourceFunction").
		First(&history).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check if function is remix: %w", err)
	}
	return true, &history, nil
}

// ============================================
// Function Likes Methods
// ============================================

// ToggleLike toggles a user's like on a function. Returns (liked, totalLikes, error).
// If the user already liked it, removes the like (unlike). Otherwise adds it.
func (r *RegistryRepository) ToggleLike(functionID, userID uuid.UUID) (bool, int64, error) {
	var existing FunctionLike
	err := r.db.Where("function_id = ? AND user_id = ?", functionID, userID).First(&existing).Error

	if err == nil {
		// User already liked - remove it (unlike)
		if err := r.db.Delete(&existing).Error; err != nil {
			return false, 0, fmt.Errorf("failed to remove like: %w", err)
		}
		count, _ := r.CountLikesForFunction(functionID)
		return false, count, nil
	} else if err == gorm.ErrRecordNotFound {
		// User hasn't liked - add like
		like := FunctionLike{
			ID:         uuid.New(),
			FunctionID: functionID,
			UserID:     userID,
			LikedAt:    time.Now(),
		}
		if err := r.db.Create(&like).Error; err != nil {
			return false, 0, fmt.Errorf("failed to add like: %w", err)
		}
		count, _ := r.CountLikesForFunction(functionID)
		return true, count, nil
	}

	return false, 0, fmt.Errorf("failed to check existing like: %w", err)
}

// HasUserLiked checks if a specific user has liked a function
func (r *RegistryRepository) HasUserLiked(functionID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&FunctionLike{}).
		Where("function_id = ? AND user_id = ?", functionID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check like status: %w", err)
	}
	return count > 0, nil
}

// CountLikesForFunction returns the total number of likes for a function
func (r *RegistryRepository) CountLikesForFunction(functionID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&FunctionLike{}).Where("function_id = ?", functionID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count likes: %w", err)
	}
	return count, nil
}

// CountLikesForFunctions returns a map of function IDs to their like counts
func (r *RegistryRepository) CountLikesForFunctions(functionIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	type result struct {
		FunctionID uuid.UUID
		Count     int64
	}
	var results []result
	if err := r.db.Model(&FunctionLike{}).
		Select("function_id, COUNT(*) as count").
		Where("function_id IN ?", functionIDs).
		Group("function_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count likes for functions: %w", err)
	}
	m := make(map[uuid.UUID]int64, len(results))
	for _, r := range results {
		m[r.FunctionID] = r.Count
	}
	return m, nil
}

// CountRemixesForFunctions returns a map of function IDs to their remix counts
func (r *RegistryRepository) CountRemixesForFunctions(functionIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	type result struct {
		SourceFunctionID uuid.UUID
		Count            int64
	}
	var results []result
	if err := r.db.Model(&RemixHistory{}).
		Select("source_function_id, COUNT(*) as count").
		Where("source_function_id IN ?", functionIDs).
		Group("source_function_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count remixes for functions: %w", err)
	}
	m := make(map[uuid.UUID]int64, len(results))
	for _, r := range results {
		m[r.SourceFunctionID] = r.Count
	}
	return m, nil
}

// GetLikesForFunction returns the list of users who liked a function (with pagination)
func (r *RegistryRepository) GetLikesForFunction(functionID uuid.UUID, limit, offset int) ([]FunctionLike, int64, error) {
	var likes []FunctionLike
	var total int64

	if err := r.db.Model(&FunctionLike{}).Where("function_id = ?", functionID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count likes: %w", err)
	}

	if err := r.db.Where("function_id = ?", functionID).
		Order("liked_at DESC").
		Limit(limit).Offset(offset).
		Find(&likes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get likes: %w", err)
	}

	return likes, total, nil
}
