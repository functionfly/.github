package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/google/uuid"
)

// CachedFunctionList represents a cached function list result
type CachedFunctionList struct {
	Functions []RegistryFunction `json:"functions"`
	Total     int                `json:"total"`
	CachedAt  time.Time          `json:"cached_at"`
}

// ListFunctions lists functions with filters (with caching)
func (r *RegistryRepository) ListFunctions(author, category string, tags []string, visibility string, limit, offset int) ([]RegistryFunction, int, error) {
	// Try cache first if available
	if r.cache != nil {
		cacheKey := r.keyGen.ListFunctions(author, category, visibility, limit, offset)
		var cached CachedFunctionList
		if err := r.cache.GetJSON(context.Background(), cacheKey, &cached); err == nil {
			// Check if cache is still fresh (within 5 minutes for list queries)
			if time.Since(cached.CachedAt) < 5*time.Minute {
				return cached.Functions, cached.Total, nil
			}
		}
	}

	query := r.db.Model(&RegistryFunction{})

	if author != "" {
		query = query.Where("author = ?", author)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if visibility == "" {
		visibility = "public"
	}
	query = query.Where("visibility = ?", visibility)

	// Count query
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	// Add ordering and pagination
	var functions []RegistryFunction
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list functions: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil {
		cacheKey := r.keyGen.ListFunctions(author, category, visibility, limit, offset)
		cachedResult := CachedFunctionList{
			Functions: functions,
			Total:     int(total),
			CachedAt:  time.Now(),
		}
		// Cache for 5 minutes for list queries
		if err := r.cache.SetJSONWithTTL(context.Background(), cacheKey, cachedResult, 5*time.Minute); err != nil {
			// Log but don't fail the request
			fmt.Printf("Failed to cache function list: %v\n", err)
		}
	}

	return functions, int(total), nil
}

// GetEdgeCacheCandidates retrieves functions eligible for edge caching based on metrics
func (r *RegistryRepository) GetEdgeCacheCandidates(ctx context.Context, minPopularity, minExecutionCount int, minTrustScore, minSuccessRate float64, maxLatencyMs int, limit int) ([]*cache.EdgeCacheCandidate, error) {
	// Query functions that meet edge cache criteria
	query := r.db.Table("registry_functions").
		Select(`
			rf.id,
			rf.author,
			rf.name,
			rf.popularity_score,
			rf.trust_score,
			COALESCE(rf.latest_version, '') as version,
			COUNT(re.id) as execution_count,
			AVG(re.duration_ms) as avg_latency,
			(SUM(CASE WHEN re.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(re.id)) as success_rate,
			MAX(re.created_at) as last_executed
		`).
		Joins("LEFT JOIN registry_function_executions re ON rf.id = re.function_id").
		Where("rf.visibility = ?", "public").
		Where("rf.popularity_score >= ?", minPopularity).
		Where("rf.trust_score >= ?", minTrustScore).
		Group("rf.id, rf.author, rf.name, rf.popularity_score, rf.trust_score, rf.latest_version").
		Having("COUNT(re.id) >= ?", minExecutionCount).
		Having("AVG(re.duration_ms) <= ?", maxLatencyMs).
		Having("(SUM(CASE WHEN re.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(re.id)) >= ?", minSuccessRate).
		Order("rf.popularity_score DESC, COUNT(re.id) DESC").
		Limit(limit)

	rows, err := query.Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query edge cache candidates: %w", err)
	}
	defer rows.Close()

	var candidates []*cache.EdgeCacheCandidate

	for rows.Next() {
		var candidate cache.EdgeCacheCandidate
		var lastExecuted sql.NullTime

		err := rows.Scan(
			&candidate.FunctionID,
			&candidate.Author,
			&candidate.FunctionName,
			&candidate.PopularityScore,
			&candidate.TrustScore,
			&candidate.Version,
			&candidate.ExecutionCount,
			&candidate.AvgLatency,
			&candidate.SuccessRate,
			&lastExecuted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge cache candidate: %w", err)
		}

		if lastExecuted.Valid {
			candidate.LastExecuted = &lastExecuted.Time
		}

		// Calculate cache priority using the edge cache service logic
		edgeCacheSvc := &cache.EdgeCacheService{}
		candidate.CachePriority = edgeCacheSvc.CalculateCachePriority(&candidate)

		candidates = append(candidates, &candidate)
	}

	return candidates, nil
}

// GetFunctionEdgeCacheMetrics retrieves metrics for a specific function for edge cache eligibility
func (r *RegistryRepository) GetFunctionEdgeCacheMetrics(ctx context.Context, functionID uuid.UUID, timeWindow time.Duration) (*cache.EdgeCacheCandidate, error) {
	// Calculate time window start
	windowStart := time.Now().Add(-timeWindow)

	// Query function metrics within the time window
	query := r.db.Table("registry_functions").
		Select(`
			rf.id,
			rf.author,
			rf.name,
			rf.popularity_score,
			rf.trust_score,
			COALESCE(rf.latest_version, '') as version,
			COUNT(re.id) as execution_count,
			AVG(re.duration_ms) as avg_latency,
			(SUM(CASE WHEN re.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(re.id)) as success_rate,
			MAX(re.created_at) as last_executed
		`).
		Joins("LEFT JOIN registry_function_executions re ON rf.id = re.function_id AND re.created_at >= ?", windowStart).
		Where("rf.id = ?", functionID).
		Group("rf.id, rf.author, rf.name, rf.popularity_score, rf.trust_score, rf.latest_version")

	var candidate cache.EdgeCacheCandidate
	var lastExecuted sql.NullTime

	row := query.Row()
	err := row.Scan(
		&candidate.FunctionID,
		&candidate.Author,
		&candidate.FunctionName,
		&candidate.PopularityScore,
		&candidate.TrustScore,
		&candidate.Version,
		&candidate.ExecutionCount,
		&candidate.AvgLatency,
		&candidate.SuccessRate,
		&lastExecuted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get function edge cache metrics: %w", err)
	}

	if lastExecuted.Valid {
		candidate.LastExecuted = &lastExecuted.Time
	}

	// Calculate cache priority
	edgeCacheSvc := &cache.EdgeCacheService{}
	candidate.CachePriority = edgeCacheSvc.CalculateCachePriority(&candidate)

	return &candidate, nil
}

// UpdateFunctionPopularityScore updates the popularity score for a function
func (r *RegistryRepository) UpdateFunctionPopularityScore(ctx context.Context, functionID uuid.UUID, newScore int) error {
	if err := r.db.Model(&RegistryFunction{}).
		Where("id = ?", functionID).
		Update("popularity_score", newScore).Error; err != nil {
		return fmt.Errorf("failed to update popularity score: %w", err)
	}

	// Invalidate related caches
	if r.cache != nil {
		go func() {
			if err := r.cache.InvalidateFunction(context.Background(), functionID.String()); err != nil {
				fmt.Printf("Failed to invalidate function cache after popularity update: %v\n", err)
			}
		}()
	}

	return nil
}

// GetPopularFunctionsByCategory gets popular functions in a specific category
func (r *RegistryRepository) GetPopularFunctionsByCategory(ctx context.Context, category string, limit int) ([]*cache.EdgeCacheCandidate, error) {
	// Query popular functions in a category
	query := r.db.Table("registry_functions").
		Select(`
			rf.id,
			rf.author,
			rf.name,
			rf.popularity_score,
			rf.trust_score,
			COALESCE(rf.latest_version, '') as version,
			COUNT(re.id) as execution_count,
			AVG(re.duration_ms) as avg_latency,
			(SUM(CASE WHEN re.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(re.id)) as success_rate,
			MAX(re.created_at) as last_executed
		`).
		Joins("LEFT JOIN registry_function_executions re ON rf.id = re.function_id").
		Where("rf.category = ?", category).
		Where("rf.visibility = ?", "public").
		Group("rf.id, rf.author, rf.name, rf.popularity_score, rf.trust_score, rf.latest_version").
		Order("rf.popularity_score DESC, COUNT(re.id) DESC").
		Limit(limit)

	rows, err := query.Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query popular functions by category: %w", err)
	}
	defer rows.Close()

	var candidates []*cache.EdgeCacheCandidate

	for rows.Next() {
		var candidate cache.EdgeCacheCandidate
		var lastExecuted sql.NullTime

		err := rows.Scan(
			&candidate.FunctionID,
			&candidate.Author,
			&candidate.FunctionName,
			&candidate.PopularityScore,
			&candidate.TrustScore,
			&candidate.Version,
			&candidate.ExecutionCount,
			&candidate.AvgLatency,
			&candidate.SuccessRate,
			&lastExecuted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan popular function candidate: %w", err)
		}

		if lastExecuted.Valid {
			candidate.LastExecuted = &lastExecuted.Time
		}

		// Calculate cache priority
		edgeCacheSvc := &cache.EdgeCacheService{}
		candidate.CachePriority = edgeCacheSvc.CalculateCachePriority(&candidate)

		candidates = append(candidates, &candidate)
	}

	return candidates, nil
}

// SearchFunctions searches functions by text query (with caching)
func (r *RegistryRepository) SearchFunctions(query string, category, runtime string, minRating float64, limit, offset int) ([]RegistryFunction, int, error) {
	// Try cache first if available
	if r.cache != nil {
		cacheKey := r.keyGen.SearchFunctions(query, category, runtime, minRating, limit, offset)
		var cached CachedFunctionList
		if err := r.cache.GetJSON(context.Background(), cacheKey, &cached); err == nil {
			// Check if cache is still fresh (within 2 minutes for search queries)
			if time.Since(cached.CachedAt) < 2*time.Minute {
				return cached.Functions, cached.Total, nil
			}
		}
	}

	dbQuery := r.db.Model(&RegistryFunction{}).
		Where("visibility = ?", "public").
		Where("name ILIKE ? OR title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	if minRating > 0 {
		dbQuery = dbQuery.Where("reliability_score >= ?", minRating)
	}

	// Count query
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Add ordering and pagination - sort by trust_score first, then popularity
	var functions []RegistryFunction
	if err := dbQuery.
		Order("trust_score DESC, popularity_score DESC, reliability_score DESC").
		Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search functions: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil {
		cacheKey := r.keyGen.SearchFunctions(query, category, runtime, minRating, limit, offset)
		cachedResult := CachedFunctionList{
			Functions: functions,
			Total:     int(total),
			CachedAt:  time.Now(),
		}
		// Cache for 2 minutes for search queries (shorter TTL due to dynamic nature)
		if err := r.cache.SetJSONWithTTL(context.Background(), cacheKey, cachedResult, 2*time.Minute); err != nil {
			// Log but don't fail the request
			fmt.Printf("Failed to cache search results: %v\n", err)
		}
	}

	return functions, int(total), nil
}

