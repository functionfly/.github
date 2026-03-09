package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	// Invalidate rating cache so next GET returns updated trust score
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(rating.FunctionID.String())
		if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
			fmt.Printf("Failed to invalidate rating cache: %v\n", err)
		}
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

// TrustDistribution holds counts of functions per trust bucket (DB trust_score 0-1: excellent >= 0.8, good >= 0.6, fair >= 0.4, poor < 0.4).
type TrustDistribution struct {
	Excellent int
	Good      int
	Fair      int
	Poor      int
}

// GetTrustDistribution returns the count of registry functions in each trust bucket.
// Uses reliability_score only (base schema: 0-100 scale, normalized to 0-1 for buckets). Run migration 000036 for trust_score.
func (r *RegistryRepository) GetTrustDistribution() (TrustDistribution, error) {
	var out TrustDistribution
	query := `
		SELECT
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.8) AS excellent,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.6 AND COALESCE(rat.reliability_score, 0) / 100.0 < 0.8) AS good,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 >= 0.4 AND COALESCE(rat.reliability_score, 0) / 100.0 < 0.6) AS fair,
			COUNT(*) FILTER (WHERE COALESCE(rat.reliability_score, 0) / 100.0 < 0.4) AS poor
		FROM registry_functions f
		LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
	`
	if err := r.db.Raw(query).Scan(&out).Error; err != nil {
		return TrustDistribution{}, fmt.Errorf("failed to get trust distribution: %w", err)
	}
	return out, nil
}

// HighRiskFunctionRow is one row for admin trust dashboard high-risk list (low trust + risk metrics).
type HighRiskFunctionRow struct {
	FunctionID      uuid.UUID
	Author          string
	Name            string
	TrustScore      float64
	ErrorRate       float64
	TimeoutRate     float64
	TenantDiversity int
	TrustUpdatedAt  *time.Time
}

// GetHighRiskFunctions returns functions with low trust score (e.g. < 0.35), ordered by trust ascending, limit n.
// Uses reliability_score only (0-100 → 0-1) so base schema works; run migration 000036 for trust_score/error_rate/timeout_rate/tenant_diversity.
func (r *RegistryRepository) GetHighRiskFunctions(limit int) ([]HighRiskFunctionRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []HighRiskFunctionRow
	query := `
		SELECT
			f.id AS function_id,
			f.author,
			f.name,
			COALESCE(rat.reliability_score, 0) / 100.0 AS trust_score,
			0::float AS error_rate,
			0::float AS timeout_rate,
			0::integer AS tenant_diversity,
			rat.updated_at AS trust_updated_at
		FROM registry_functions f
		LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
		WHERE COALESCE(rat.reliability_score, 0) / 100.0 < 0.35
		ORDER BY COALESCE(rat.reliability_score, 0) ASC NULLS LAST, f.created_at ASC
		LIMIT ?
	`
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get high-risk functions: %w", err)
	}
	return rows, nil
}

// ExecutionAuditRow represents a row for execution audit with function details.
type ExecutionAuditRow struct {
	ID                string
	FunctionID        uuid.UUID
	FunctionName      string
	Author            string
	Version           string
	Timestamp         time.Time
	DurationMs        int
	Outcome           string
	ErrorCode         sql.NullString
	TenantID          *uuid.UUID
	InputSize         int            // Placeholder, not in current schema
	OutputSize        int            // Placeholder, not in current schema
	ExecutionRootHash sql.NullString // From certificates if available
	NodeSignature     sql.NullString // From certificates if available
}

// GetExecutionAuditData returns paginated execution audit data with function details and filtering.
// Supports search (function name/author), tenant filter, status filter, pagination.
func (r *RegistryRepository) GetExecutionAuditData(
	searchTerm, tenantFilter, statusFilter string,
	offset, limit int,
) ([]ExecutionAuditRow, int64, error) {
	// Build the base query with joins
	query := r.db.Table("registry_function_executions e").
		Joins("JOIN registry_functions f ON e.function_id = f.id").
		Select(`
			e.id::text as id,
			e.function_id,
			f.name as function_name,
			f.author,
			e.version,
			e.timestamp,
			e.duration_ms,
			e.outcome,
			e.error_code,
			e.tenant_id,
			0 as input_size,  -- Placeholder
			0 as output_size,  -- Placeholder
			'' as execution_root_hash,  -- Placeholder
			'' as node_signature        -- Placeholder
		`)

	// Apply filters
	if searchTerm != "" {
		q := "%" + searchTerm + "%"
		query = query.Where("(f.name ILIKE ? OR f.author ILIKE ?)", q, q)
	}
	if tenantFilter != "" {
		query = query.Where("f.author = ?", tenantFilter)
	}
	if statusFilter != "" {
		query = query.Where("e.outcome = ?", statusFilter)
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	// Apply ordering and pagination
	var rows []ExecutionAuditRow
	err := query.Order("e.timestamp DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get execution audit data: %w", err)
	}

	return rows, total, nil
}

// FraudDetectionResult holds detected fraud patterns.
type FraudDetectionResult struct {
	BotPatterns       []BotPatternRow
	IPClusters        []IPClusterRow
	WashUsagePatterns []WashUsageRow
	Summary           FraudSummaryCounts
}

type BotPatternRow struct {
	PatternType       string
	ConfidenceScore   int
	AffectedFunctions []string
	AffectedTenants   []string
	Pattern           string
	DetectedAt        time.Time
}

type IPClusterRow struct {
	IPRange           string
	AssociatedTenants []string
	RiskLevel         string
	CommonPatterns    []string
	FirstSeen         time.Time
	LastSeen          time.Time
}

type WashUsageRow struct {
	TenantA              string
	TenantB              string
	Function             string
	Pattern              string
	Confidence           int
	ReciprocalExecutions int
	DetectedAt           time.Time
}

type FraudSummaryCounts struct {
	TotalBotPatterns  int
	HighRiskClusters  int
	SuspiciousTenants int
	WashUsageDetected int
}

// DetectFraudPatterns runs basic fraud detection algorithms on registry data.
func (r *RegistryRepository) DetectFraudPatterns() (*FraudDetectionResult, error) {
	result := &FraudDetectionResult{}

	// Detect bot patterns: functions with low diversity but high execution counts
	botPatterns, err := r.detectBotPatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to detect bot patterns: %w", err)
	}
	result.BotPatterns = botPatterns

	// Detect IP clusters: group by IP ranges and look for suspicious patterns
	ipClusters, err := r.detectIPClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to detect IP clusters: %w", err)
	}
	result.IPClusters = ipClusters

	// Detect wash usage patterns (simplified)
	washPatterns, err := r.detectWashUsagePatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to detect wash usage patterns: %w", err)
	}
	result.WashUsagePatterns = washPatterns

	// Calculate summary counts
	result.Summary = FraudSummaryCounts{
		TotalBotPatterns:  len(botPatterns),
		HighRiskClusters:  countHighRiskClusters(ipClusters),
		SuspiciousTenants: calculateSuspiciousTenants(botPatterns, ipClusters),
		WashUsageDetected: len(washPatterns),
	}

	return result, nil
}

// detectBotPatterns looks for functions with suspicious execution patterns.
func (r *RegistryRepository) detectBotPatterns() ([]BotPatternRow, error) {
	var patterns []BotPatternRow

	// Simple pattern: functions with very low tenant diversity (< 3 unique tenants)
	// but high execution counts (> 100 executions in last 30 days)
	query := `
		WITH function_stats AS (
			SELECT
				e.function_id,
				f.name as function_name,
				f.author as tenant_name,
				COUNT(DISTINCT e.tenant_id) as unique_tenants,
				COUNT(*) as total_executions,
				SUM(CASE WHEN e.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as success_rate
			FROM registry_function_executions e
			JOIN registry_functions f ON e.function_id = f.id
			WHERE e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY e.function_id, f.name, f.author
			HAVING COUNT(*) > 100 AND COUNT(DISTINCT e.tenant_id) < 3
		)
		SELECT
			function_id,
			function_name,
			tenant_name,
			unique_tenants,
			total_executions,
			success_rate
		FROM function_stats
		ORDER BY total_executions DESC
		LIMIT 10
	`

	var rows []struct {
		FunctionID      uuid.UUID `json:"function_id"`
		FunctionName    string    `json:"function_name"`
		TenantName      string    `json:"tenant_name"`
		UniqueTenants   int       `json:"unique_tenants"`
		TotalExecutions int       `json:"total_executions"`
		SuccessRate     float64   `json:"success_rate"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 50
		if row.UniqueTenants == 1 {
			confidence = 90
		} else if row.UniqueTenants == 2 {
			confidence = 70
		}

		patterns = append(patterns, BotPatternRow{
			PatternType:       "low_diversity_execution",
			ConfidenceScore:   confidence,
			AffectedFunctions: []string{row.FunctionName},
			AffectedTenants:   []string{row.TenantName},
			Pattern:           fmt.Sprintf("High execution volume (%d) with low tenant diversity (%d unique tenants)", row.TotalExecutions, row.UniqueTenants),
			DetectedAt:        time.Now(),
		})
	}

	return patterns, nil
}

// detectIPClusters groups executions by IP ranges and identifies suspicious clusters.
func (r *RegistryRepository) detectIPClusters() ([]IPClusterRow, error) {
	var clusters []IPClusterRow

	// Simple IP clustering: group by /24 subnets and look for patterns
	query := `
		WITH ip_clusters AS (
			SELECT
				SUBSTRING(caller_ip FROM '^(\d+\.\d+\.\d+)\.') || '.0/24' as ip_range,
				COUNT(DISTINCT tenant_id) as tenant_count,
				COUNT(DISTINCT function_id) as function_count,
				COUNT(*) as execution_count,
				MIN(timestamp) as first_seen,
				MAX(timestamp) as last_seen
			FROM registry_function_executions
			WHERE caller_ip IS NOT NULL AND caller_ip != ''
				AND timestamp > NOW() - INTERVAL '7 days'
			GROUP BY SUBSTRING(caller_ip FROM '^(\d+\.\d+\.\d+)\.')
			HAVING COUNT(DISTINCT tenant_id) > 5 OR COUNT(*) > 1000
		)
		SELECT * FROM ip_clusters
		ORDER BY execution_count DESC
		LIMIT 5
	`

	var rows []struct {
		IPRange        string    `json:"ip_range"`
		TenantCount    int       `json:"tenant_count"`
		FunctionCount  int       `json:"function_count"`
		ExecutionCount int       `json:"execution_count"`
		FirstSeen      time.Time `json:"first_seen"`
		LastSeen       time.Time `json:"last_seen"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		riskLevel := "medium"
		if row.TenantCount > 10 || row.ExecutionCount > 5000 {
			riskLevel = "high"
		}

		patterns := []string{"High execution volume"}
		if row.TenantCount > 5 {
			patterns = append(patterns, "Many tenants from same IP range")
		}

		clusters = append(clusters, IPClusterRow{
			IPRange:           row.IPRange,
			AssociatedTenants: []string{}, // Would need separate query to get tenant names
			RiskLevel:         riskLevel,
			CommonPatterns:    patterns,
			FirstSeen:         row.FirstSeen,
			LastSeen:          row.LastSeen,
		})
	}

	return clusters, nil
}

// detectWashUsagePatterns looks for reciprocal execution patterns between tenants.
func (r *RegistryRepository) detectWashUsagePatterns() ([]WashUsageRow, error) {
	var patterns []WashUsageRow

	// Simplified wash trading detection: look for tenants that execute each other's functions frequently
	query := `
		WITH tenant_pairs AS (
			SELECT
				e1.tenant_id as tenant_a,
				e2.tenant_id as tenant_b,
				e1.function_id,
				COUNT(*) as reciprocal_count
			FROM registry_function_executions e1
			JOIN registry_function_executions e2 ON e1.function_id = e2.function_id
				AND e1.tenant_id = e2.user_id
				AND e2.tenant_id = e1.user_id
			WHERE e1.timestamp > NOW() - INTERVAL '30 days'
				AND e2.timestamp > NOW() - INTERVAL '30 days'
				AND e1.tenant_id != e2.tenant_id
			GROUP BY e1.tenant_id, e2.tenant_id, e1.function_id
			HAVING COUNT(*) > 10
		)
		SELECT
			tp.tenant_a,
			tp.tenant_b,
			f.name as function_name,
			tp.reciprocal_count
		FROM tenant_pairs tp
		JOIN registry_functions f ON tp.function_id = f.id
		ORDER BY tp.reciprocal_count DESC
		LIMIT 5
	`

	var rows []struct {
		TenantA         uuid.UUID `json:"tenant_a"`
		TenantB         uuid.UUID `json:"tenant_b"`
		FunctionName    string    `json:"function_name"`
		ReciprocalCount int       `json:"reciprocal_count"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 60
		if row.ReciprocalCount > 50 {
			confidence = 85
		}

		patterns = append(patterns, WashUsageRow{
			TenantA:              row.TenantA.String(),
			TenantB:              row.TenantB.String(),
			Function:             row.FunctionName,
			Pattern:              "Reciprocal execution pattern detected",
			Confidence:           confidence,
			ReciprocalExecutions: row.ReciprocalCount,
			DetectedAt:           time.Now(),
		})
	}

	return patterns, nil
}

// countHighRiskClusters counts clusters with high risk level.
func countHighRiskClusters(clusters []IPClusterRow) int {
	count := 0
	for _, cluster := range clusters {
		if cluster.RiskLevel == "high" {
			count++
		}
	}
	return count
}

// calculateSuspiciousTenants calculates total suspicious tenants from patterns.
func calculateSuspiciousTenants(botPatterns []BotPatternRow, ipClusters []IPClusterRow) int {
	tenantSet := make(map[string]bool)

	for _, pattern := range botPatterns {
		for _, tenant := range pattern.AffectedTenants {
			tenantSet[tenant] = true
		}
	}

	// For IP clusters, we can't easily get tenant names without additional queries
	// So we'll estimate based on cluster sizes
	for _, cluster := range ipClusters {
		if cluster.RiskLevel == "high" {
			// Assume most tenants in high-risk clusters are suspicious
			// This is a simplification - in reality we'd need to get actual tenant lists
		}
	}

	return len(tenantSet)
}

// EconomicAnalysisResult holds economic monitoring data.
type EconomicAnalysisResult struct {
	TopRevenueGenerators []RevenueGeneratorRow
	SuspiciousGrowth     []SuspiciousGrowthRow
	ArtificialBoosting   []ArtificialBoostingRow
}

type RevenueGeneratorRow struct {
	FunctionID     uuid.UUID
	FunctionName   string
	Author         string
	Revenue30d     float64
	ExecutionCount int
	GrowthRate     float64
}

type SuspiciousGrowthRow struct {
	FunctionID   uuid.UUID
	FunctionName string
	Author       string
	Pattern      string
	Details      string
	DetectedAt   time.Time
}

type ArtificialBoostingRow struct {
	FunctionID      uuid.UUID
	FunctionName    string
	DetectedPattern string
	Confidence      int
	RelatedAccounts []string
}

// AnalyzeEconomicData performs economic monitoring and analysis.
func (r *RegistryRepository) AnalyzeEconomicData() (*EconomicAnalysisResult, error) {
	result := &EconomicAnalysisResult{}

	// Get top revenue generators
	revenueGenerators, err := r.getTopRevenueGenerators(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue generators: %w", err)
	}
	result.TopRevenueGenerators = revenueGenerators

	// Detect suspicious growth patterns
	suspiciousGrowth, err := r.detectSuspiciousGrowth()
	if err != nil {
		return nil, fmt.Errorf("failed to detect suspicious growth: %w", err)
	}
	result.SuspiciousGrowth = suspiciousGrowth

	// Detect artificial boosting
	artificialBoosting, err := r.detectArtificialBoosting()
	if err != nil {
		return nil, fmt.Errorf("failed to detect artificial boosting: %w", err)
	}
	result.ArtificialBoosting = artificialBoosting

	return result, nil
}

// getTopRevenueGenerators returns functions ranked by revenue in the last 30 days.
func (r *RegistryRepository) getTopRevenueGenerators(limit int) ([]RevenueGeneratorRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		WITH function_revenue AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				COALESCE(f.price_per_call, 0) as price_per_call,
				COUNT(e.id) as execution_count,
				COUNT(e.id) * COALESCE(f.price_per_call, 0) as revenue_30d
			FROM registry_functions f
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY f.id, f.name, f.author, f.price_per_call
		),
		function_growth AS (
			SELECT
				fr.function_id,
				fr.function_name,
				fr.author,
				fr.revenue_30d,
				fr.execution_count,
				-- Calculate growth rate: (current - previous) / previous * 100
				CASE
					WHEN prev.revenue_30d > 0 THEN
						((fr.revenue_30d - prev.revenue_30d) / prev.revenue_30d) * 100
					ELSE 0
				END as growth_rate
			FROM function_revenue fr
			LEFT JOIN (
				-- Previous 30-day period revenue
				SELECT
					f.id as function_id,
					COUNT(e.id) * COALESCE(f.price_per_call, 0) as revenue_30d
				FROM registry_functions f
				LEFT JOIN registry_function_executions e ON e.function_id = f.id
					AND e.timestamp BETWEEN NOW() - INTERVAL '60 days' AND NOW() - INTERVAL '30 days'
				GROUP BY f.id, f.price_per_call
			) prev ON fr.function_id = prev.function_id
			WHERE fr.revenue_30d > 0  -- Only include functions with revenue
			ORDER BY fr.revenue_30d DESC
			LIMIT ?
		)
		SELECT * FROM function_growth
	`

	var rows []RevenueGeneratorRow
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

// detectSuspiciousGrowth identifies functions with anomalous growth patterns.
func (r *RegistryRepository) detectSuspiciousGrowth() ([]SuspiciousGrowthRow, error) {
	var suspicious []SuspiciousGrowthRow

	// Look for functions with sudden spikes in the last 7 days
	query := `
		WITH recent_growth AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				-- Recent 7-day revenue
				COALESCE(SUM(CASE WHEN e.timestamp > NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) * f.price_per_call, 0) as recent_revenue,
				-- Previous 7-day revenue (week before)
				COALESCE(SUM(CASE WHEN e.timestamp BETWEEN NOW() - INTERVAL '14 days' AND NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) * f.price_per_call, 0) as prev_revenue,
				-- Recent execution count
				SUM(CASE WHEN e.timestamp > NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) as recent_executions
			FROM registry_functions f
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '14 days'
			WHERE f.price_per_call > 0
			GROUP BY f.id, f.name, f.author, f.price_per_call
		)
		SELECT
			function_id,
			function_name,
			author,
			recent_revenue,
			prev_revenue,
			recent_executions
		FROM recent_growth
		WHERE recent_revenue > 0 AND prev_revenue > 0
			AND (recent_revenue / NULLIF(prev_revenue, 0)) > 5.0  -- 500% growth
			AND recent_executions > 100  -- Significant volume
		ORDER BY (recent_revenue / NULLIF(prev_revenue, 0)) DESC
		LIMIT 10
	`

	var rows []struct {
		FunctionID       uuid.UUID `json:"function_id"`
		FunctionName     string    `json:"function_name"`
		Author           string    `json:"author"`
		RecentRevenue    float64   `json:"recent_revenue"`
		PrevRevenue      float64   `json:"prev_revenue"`
		RecentExecutions int       `json:"recent_executions"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		growthMultiplier := row.RecentRevenue / row.PrevRevenue
		suspicious = append(suspicious, SuspiciousGrowthRow{
			FunctionID:   row.FunctionID,
			FunctionName: row.FunctionName,
			Author:       row.Author,
			Pattern:      "Sudden revenue spike",
			Details: fmt.Sprintf("%.1fx revenue increase (%.2f -> %.2f) in 7 days with %d executions",
				growthMultiplier, row.PrevRevenue, row.RecentRevenue, row.RecentExecutions),
			DetectedAt: time.Now(),
		})
	}

	return suspicious, nil
}

// detectArtificialBoosting identifies functions that appear to be artificially boosted.
func (r *RegistryRepository) detectArtificialBoosting() ([]ArtificialBoostingRow, error) {
	var artificial []ArtificialBoostingRow

	// Look for functions with coordinated rating patterns that don't match execution volume
	query := `
		WITH function_metrics AS (
			SELECT
				f.id as function_id,
				f.name as function_name,
				f.author,
				-- Rating metrics
				COALESCE(rat.total_ratings, 0) as total_ratings,
				COALESCE(rat.overall_score, 0) as overall_score,
				-- Execution metrics
				COUNT(e.id) as execution_count,
				COUNT(DISTINCT e.tenant_id) as unique_tenants,
				COUNT(DISTINCT e.caller_ip) as unique_ips
			FROM registry_functions f
			LEFT JOIN registry_function_ratings rat ON rat.function_id = f.id
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
				AND e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY f.id, f.name, f.author, rat.total_ratings, rat.overall_score
		)
		SELECT
			function_id,
			function_name,
			author,
			total_ratings,
			overall_score,
			execution_count,
			unique_tenants,
			unique_ips
		FROM function_metrics
		WHERE total_ratings > 10  -- Has some ratings
			AND execution_count < 100  -- Low execution volume
			AND unique_tenants < 5  -- Low tenant diversity
			AND overall_score > 4.0  -- High rating score
		ORDER BY overall_score DESC
		LIMIT 5
	`

	var rows []struct {
		FunctionID     uuid.UUID `json:"function_id"`
		FunctionName   string    `json:"function_name"`
		Author         string    `json:"author"`
		TotalRatings   int       `json:"total_ratings"`
		OverallScore   float64   `json:"overall_score"`
		ExecutionCount int       `json:"execution_count"`
		UniqueTenants  int       `json:"unique_tenants"`
		UniqueIPs      int       `json:"unique_ips"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 60
		if row.UniqueTenants == 1 {
			confidence = 95
		} else if row.UniqueTenants <= 3 {
			confidence = 80
		}

		relatedAccounts := []string{row.Author}
		pattern := fmt.Sprintf("High rating (%.1f/5) with low usage (%d executions, %d tenants) - possible artificial boosting",
			row.OverallScore, row.ExecutionCount, row.UniqueTenants)

		artificial = append(artificial, ArtificialBoostingRow{
			FunctionID:      row.FunctionID,
			FunctionName:    row.FunctionName,
			DetectedPattern: pattern,
			Confidence:      confidence,
			RelatedAccounts: relatedAccounts,
		})
	}

	return artificial, nil
}
