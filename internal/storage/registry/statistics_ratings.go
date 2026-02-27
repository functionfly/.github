package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CachedFunctionStats represents cached function statistics
type CachedFunctionStats struct {
	TotalCalls   int     `json:"total_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	P95LatencyMs int     `json:"p95_latency_ms"`
	CachedAt     time.Time `json:"cached_at"`
}

// GetFunctionStats retrieves execution stats for a function (with caching)
func (r *RegistryRepository) GetFunctionStats(functionID uuid.UUID, since time.Time) (totalCalls int, successRate float64, avgLatencyMs int, p95LatencyMs int, err error) {
	// Try cache first if available
	if r.cache != nil {
		cacheKey := r.keyGen.FunctionStats(functionID.String(), since)
		var cached CachedFunctionStats
		if err := r.cache.GetJSON(context.Background(), cacheKey, &cached); err == nil {
			// Check if cache is still fresh (within 10 minutes for stats)
			if time.Since(cached.CachedAt) < 10*time.Minute {
				return cached.TotalCalls, cached.SuccessRate, cached.AvgLatencyMs, cached.P95LatencyMs, nil
			}
		}
	}

	var result struct {
		TotalCalls  int     `json:"total_calls"`
		SuccessRate float64 `json:"success_rate"`
		AvgLatency  int     `json:"avg_latency"`
		P95Latency  int     `json:"p95_latency"`
	}

	query := `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p95_latency
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get function stats: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil {
		cacheKey := r.keyGen.FunctionStats(functionID.String(), since)
		cachedResult := CachedFunctionStats{
			TotalCalls:   result.TotalCalls,
			SuccessRate:  result.SuccessRate,
			AvgLatencyMs: result.AvgLatency,
			P95LatencyMs: result.P95Latency,
			CachedAt:     time.Now(),
		}
		// Cache for 10 minutes for stats
		if err := r.cache.SetJSONWithTTL(context.Background(), cacheKey, cachedResult, 10*time.Minute); err != nil {
			// Log but don't fail the request
			fmt.Printf("Failed to cache function stats: %v\n", err)
		}
	}

	return result.TotalCalls, result.SuccessRate, result.AvgLatency, result.P95Latency, nil
}

// GetFunctionTrustStats retrieves extended stats including trust score components
func (r *RegistryRepository) GetFunctionTrustStats(functionID uuid.UUID, since time.Time) (
	totalCalls int, successRate float64, avgLatencyMs int, p50LatencyMs int, p95LatencyMs int,
	timeoutRate float64, errorRate float64, err error) {
	var result struct {
		TotalCalls  int     `json:"total_calls"`
		SuccessRate float64 `json:"success_rate"`
		AvgLatency  int     `json:"avg_latency"`
		P50Latency  int     `json:"p50_latency"`
		P95Latency  int     `json:"p95_latency"`
		TimeoutRate float64 `json:"timeout_rate"`
		ErrorRate   float64 `json:"error_rate"`
	}

	query := `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p50_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p95_latency,
			COALESCE(SUM(CASE WHEN outcome = 'timeout' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as timeout_rate,
			COALESCE(SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as error_rate
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to get function trust stats: %w", err)
	}

	return result.TotalCalls, result.SuccessRate, result.AvgLatency, result.P50Latency,
		result.P95Latency, result.TimeoutRate, result.ErrorRate, nil
}

// GetConsumerDiversity returns unique caller metrics for a function
func (r *RegistryRepository) GetConsumerDiversity(functionID uuid.UUID, since time.Time) (uniqueIPs int, uniqueTenants int, uniqueUsers int, err error) {
	var result struct {
		UniqueIPs     int `json:"unique_ips"`
		UniqueTenants int `json:"unique_tenants"`
		UniqueUsers   int `json:"unique_users"`
	}

	query := `
		SELECT
			COUNT(DISTINCT caller_ip) as unique_ips,
			COUNT(DISTINCT tenant_id) as unique_tenants,
			COUNT(DISTINCT user_id) as unique_users
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
			AND caller_ip IS NOT NULL
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get consumer diversity: %w", err)
	}

	return result.UniqueIPs, result.UniqueTenants, result.UniqueUsers, nil
}

// GetOrCreateRating gets or creates a rating record for a function
func (r *RegistryRepository) GetOrCreateRating(functionID uuid.UUID) (*RegistryFunctionRating, error) {
	var rating RegistryFunctionRating

	// Try to find existing rating
	err := r.db.Where("function_id = ?", functionID).First(&rating).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new rating
			rating = RegistryFunctionRating{
				ID:         uuid.New(),
				FunctionID: functionID,
				UpdatedAt:  time.Now(),
			}
			if err := r.db.Create(&rating).Error; err != nil {
				return nil, fmt.Errorf("failed to create rating: %w", err)
			}
			return &rating, nil
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return &rating, nil
}

// UpdateRating updates a function's rating
func (r *RegistryRepository) UpdateRating(rating *RegistryFunctionRating) error {
	rating.UpdatedAt = time.Now()

	if err := r.db.Where("function_id = ?", rating.FunctionID).Updates(map[string]interface{}{
		"overall_score":       rating.OverallScore,
		"reliability_score":   rating.ReliabilityScore,
		"latency_score":       rating.LatencyScore,
		"documentation_score": rating.DocumentationScore,
		"total_ratings":       rating.TotalRatings,
		"success_rate":        rating.SuccessRate,
		"p95_latency_ms":      rating.P95LatencyMs,
		"avg_latency_ms":      rating.AvgLatencyMs,
		"p50_latency_ms":      rating.P50LatencyMs,
		"timeout_rate":        rating.TimeoutRate,
		"error_rate":          rating.ErrorRate,
		"consumer_diversity":  rating.ConsumerDiversity,
		"tenant_diversity":    rating.TenantDiversity,
		"user_diversity":      rating.UserDiversity,
		"trust_score":         rating.TrustScore,
		"trust_updated_at":    rating.TrustUpdatedAt,
		"updated_at":          rating.UpdatedAt,
	}).Error; err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	// Invalidate rating cache
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(rating.FunctionID.String())
		go func() {
			if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
				fmt.Printf("Failed to invalidate rating cache: %v\n", err)
			}
		}()
	}

	return nil
}

// UpdateTrustScore updates the trust score and related metrics for a function
func (r *RegistryRepository) UpdateTrustScore(rating *RegistryFunctionRating) error {
	now := time.Now()
	trustUpdatedAt := &now

	if err := r.db.Model(&RegistryFunctionRating{}).Where("function_id = ?", rating.FunctionID).Updates(map[string]interface{}{
		"p50_latency_ms":     rating.P50LatencyMs,
		"timeout_rate":       rating.TimeoutRate,
		"error_rate":         rating.ErrorRate,
		"consumer_diversity": rating.ConsumerDiversity,
		"tenant_diversity":   rating.TenantDiversity,
		"user_diversity":     rating.UserDiversity,
		"trust_score":        rating.TrustScore,
		"trust_updated_at":   trustUpdatedAt,
		"success_rate":       rating.SuccessRate,
		"p95_latency_ms":     rating.P95LatencyMs,
		"avg_latency_ms":     rating.AvgLatencyMs,
		"reliability_score":  rating.ReliabilityScore,
		"latency_score":      rating.LatencyScore,
	}).Error; err != nil {
		return fmt.Errorf("failed to update trust score: %w", err)
	}

	return nil
}

// IncrementPopularity increments the popularity score for a function
func (r *RegistryRepository) IncrementPopularity(functionID uuid.UUID) error {
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", functionID).
		UpdateColumn("popularity_score", gorm.Expr("popularity_score + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment popularity: %w", err)
	}
	return nil
}

// GetRatingByFunctionID gets a rating by function ID (returns nil if not found)
func (r *RegistryRepository) GetRatingByFunctionID(functionID uuid.UUID) (*RegistryFunctionRating, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(functionID.String())
		var rating RegistryFunctionRating
		if err := r.cache.GetJSON(context.Background(), cacheKey, &rating); err == nil {
			return &rating, nil
		}
		// Cache miss - continue to database
	}

	var rating RegistryFunctionRating
	err := r.db.Where("function_id = ?", functionID).First(&rating).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(functionID.String())
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, rating); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache function rating: %v\n", err)
			}
		}()
	}

	return &rating, nil
}

// UpdateFunctionPopularity sets the popularity score for a function
func (r *RegistryRepository) UpdateFunctionPopularity(functionID uuid.UUID, score int) error {
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", functionID).
		UpdateColumn("popularity_score", score).Error; err != nil {
		return fmt.Errorf("failed to update popularity score: %w", err)
	}
	return nil
}

// GetExecutionCountForFunction gets the execution count for a function within a time window
func (r *RegistryRepository) GetExecutionCountForFunction(functionID uuid.UUID, timeWindow time.Duration) (int64, error) {
	windowStart := time.Now().Add(-timeWindow)

	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND timestamp > ?", functionID, windowStart).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}

	return count, nil
}