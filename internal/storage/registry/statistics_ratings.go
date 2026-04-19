package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CachedFunctionStats represents cached function statistics
type CachedFunctionStats struct {
	TotalCalls   int       `json:"total_calls"`
	SuccessRate  float64   `json:"success_rate"`
	AvgLatencyMs int       `json:"avg_latency_ms"`
	P95LatencyMs int       `json:"p95_latency_ms"`
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
