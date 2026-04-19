package cache

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EdgeCacheService manages edge caching for frequently executed functions
type EdgeCacheService struct {
	registryCache *RegistryRedisCache
	cdnService    *CDNService
	repository    EdgeCacheRepository // Repository for database queries
	minPopularity int                 // Minimum popularity score for edge caching
	minExecutions int64               // Minimum execution count for edge caching
	cacheDuration time.Duration       // How long to cache at edge
}

// EdgeCacheRepository interface for database operations needed by edge cache
type EdgeCacheRepository interface {
	GetEdgeCacheCandidates(ctx context.Context, minPopularity, minExecutionCount int, minTrustScore, minSuccessRate float64, maxLatencyMs int, limit int) ([]*EdgeCacheCandidate, error)
	GetFunctionEdgeCacheMetrics(ctx context.Context, functionID uuid.UUID, timeWindow time.Duration) (*EdgeCacheCandidate, error)
	UpdateFunctionPopularityScore(ctx context.Context, functionID uuid.UUID, newScore int) error
}

// EdgeCacheCandidate represents a function that could be cached at edge
type EdgeCacheCandidate struct {
	FunctionID     uuid.UUID `json:"function_id"`
	FunctionName   string    `json:"function_name"`
	Author         string    `json:"author"`
	Version        string    `json:"version"`
	PopularityScore int      `json:"popularity_score"`
	ExecutionCount int64     `json:"execution_count"`
	SuccessRate    float64   `json:"success_rate"`
	AvgLatency     int       `json:"avg_latency"`
	TrustScore     float64   `json:"trust_score"`
	LastExecuted   *time.Time `json:"last_executed"`
	CachePriority  int       `json:"cache_priority"` // Calculated priority score
}

// EdgeCacheConfig holds edge caching configuration
type EdgeCacheConfig struct {
	Enabled                bool          // Whether edge caching is enabled
	MinPopularityScore     int           // Minimum popularity score required
	MinExecutionCount      int           // Minimum execution count required
	MinTrustScore          float64       // Minimum trust score required
	MinSuccessRate         float64       // Minimum success rate required
	MaxLatencyMs           int           // Maximum average latency allowed
	CacheDuration          time.Duration // How long to cache at edge
	MaxEdgeFunctions       int           // Maximum functions to cache at edge
	RefreshInterval        time.Duration // How often to refresh edge cache list
}

// NewEdgeCacheService creates a new edge cache service
func NewEdgeCacheService(registryCache *RegistryRedisCache, cdnService *CDNService, config *EdgeCacheConfig) *EdgeCacheService {
	if config == nil {
		config = &EdgeCacheConfig{
			Enabled:            true,
			MinPopularityScore: 50,
			MinExecutionCount:  100,
			MinTrustScore:      70.0,
			MinSuccessRate:     95.0,
			MaxLatencyMs:       5000,
			CacheDuration:      1 * time.Hour,
			MaxEdgeFunctions:   100,
			RefreshInterval:    10 * time.Minute,
		}
	}

	service := &EdgeCacheService{
		registryCache: registryCache,
		cdnService:    cdnService,
		minPopularity: config.MinPopularityScore,
		minExecutions: int64(config.MinExecutionCount),
		cacheDuration: config.CacheDuration,
	}

	if config.Enabled {
		// Start background refresh of edge cache candidates
		go service.startEdgeCacheRefresh(config.RefreshInterval)
	}

	return service
}

// SetRepository sets the repository for database queries
func (e *EdgeCacheService) SetRepository(repo EdgeCacheRepository) {
	e.repository = repo
}

// IsEligibleForEdgeCaching determines if a function should be cached at edge
func (e *EdgeCacheService) IsEligibleForEdgeCaching(functionID uuid.UUID, popularityScore int, executionCount int64, trustScore float64, successRate float64, avgLatency int) bool {
	return popularityScore >= e.minPopularity &&
		   executionCount >= e.minExecutions &&
		   trustScore >= 70.0 && // Minimum trust score
		   successRate >= 95.0 && // Minimum success rate
		   avgLatency <= 5000 // Maximum latency 5 seconds
}

// CalculateCachePriority calculates a priority score for edge caching
func (e *EdgeCacheService) CalculateCachePriority(candidate *EdgeCacheCandidate) int {
	// Priority based on: popularity (40%), execution count (30%), trust score (20%), success rate (10%)
	popularityWeight := float64(candidate.PopularityScore) * 0.4
	executionWeight := float64(candidate.ExecutionCount) / 1000.0 * 0.3 // Normalize execution count
	trustWeight := candidate.TrustScore * 0.2
	successWeight := candidate.SuccessRate * 0.1

	// Recency bonus (if executed in last hour, +10 points)
	recencyBonus := 0
	if candidate.LastExecuted != nil && time.Since(*candidate.LastExecuted) < time.Hour {
		recencyBonus = 10
	}

	return int(popularityWeight + executionWeight + trustWeight + successWeight) + recencyBonus
}

// GetEdgeCacheCandidates finds functions eligible for edge caching
func (e *EdgeCacheService) GetEdgeCacheCandidates(ctx context.Context) ([]*EdgeCacheCandidate, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repository not set for edge cache service")
	}

	// Try to get from cache first
	cacheKey := "edge_cache:candidates"
	var candidates []*EdgeCacheCandidate
	if err := e.registryCache.GetJSON(ctx, cacheKey, &candidates); err == nil {
		return candidates, nil
	}

	// Query database for edge cache candidates
	minPopularity := e.minPopularity
	minExecutionCount := int(e.minExecutions)
	minTrustScore := 70.0    // Minimum trust score
	minSuccessRate := 95.0   // Minimum success rate
	maxLatencyMs := 5000     // Maximum latency 5 seconds
	limit := 200             // Get more candidates, will be filtered later

	candidates, err := e.repository.GetEdgeCacheCandidates(
		ctx,
		minPopularity,
		minExecutionCount,
		minTrustScore,
		minSuccessRate,
		maxLatencyMs,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query edge cache candidates: %w", err)
	}

	// Filter, score, and validate candidates
	var validCandidates []*EdgeCacheCandidate
	for _, candidate := range candidates {
		candidate.CachePriority = e.CalculateCachePriority(candidate)
		if e.IsEligibleForEdgeCaching(
			candidate.FunctionID,
			candidate.PopularityScore,
			candidate.ExecutionCount,
			candidate.TrustScore,
			candidate.SuccessRate,
			candidate.AvgLatency,
		) {
			validCandidates = append(validCandidates, candidate)
		}
	}

	// Cache the candidates list for faster access
	if err := e.registryCache.SetJSONWithTTL(ctx, cacheKey, validCandidates, 5*time.Minute); err != nil {
		logrus.Warnf("Failed to cache edge cache candidates: %v", err)
	}

	return validCandidates, nil
}

// RefreshEdgeCacheCandidates refreshes the list of functions cached at edge
func (e *EdgeCacheService) RefreshEdgeCacheCandidates(ctx context.Context) error {
	candidates, err := e.GetEdgeCacheCandidates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get edge cache candidates: %w", err)
	}

	// Sort by priority (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].CachePriority > candidates[j].CachePriority
	})

	// Take top N candidates
	maxCandidates := 100 // Configurable
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}

	// Update edge cache configuration
	edgeCacheKey := "edge_cache:active_functions"
	if err := e.registryCache.SetJSON(ctx, edgeCacheKey, candidates); err != nil {
		return fmt.Errorf("failed to update edge cache: %w", err)
	}

	logrus.Infof("Refreshed edge cache with %d functions", len(candidates))
	return nil
}

// IsFunctionEdgeCached checks if a function is currently cached at edge
func (e *EdgeCacheService) IsFunctionEdgeCached(ctx context.Context, functionID uuid.UUID) (bool, error) {
	edgeCacheKey := "edge_cache:active_functions"
	var candidates []*EdgeCacheCandidate

	if err := e.registryCache.GetJSON(ctx, edgeCacheKey, &candidates); err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		return false, err
	}

	for _, candidate := range candidates {
		if candidate.FunctionID == functionID {
			return true, nil
		}
	}

	return false, nil
}

// SetEdgeCacheHeaders sets appropriate headers for edge-cached functions
func (e *EdgeCacheService) SetEdgeCacheHeaders(w http.ResponseWriter, functionID uuid.UUID, version string, popularityScore int) {
	// Only set edge cache headers if function is eligible and CDN is enabled
	if !e.IsEligibleForEdgeCaching(functionID, popularityScore, 1000, 80.0, 98.0, 2000) {
		return
	}

	if e.cdnService == nil || !e.cdnService.IsCDNEnabled() {
		return
	}

	// Set edge cache headers (longer cache duration for popular functions)
	edgeMaxAge := int(e.cacheDuration.Seconds())
	if popularityScore > 200 {
		edgeMaxAge *= 2 // Double cache time for very popular functions
	}

	cacheControl := fmt.Sprintf("public, max-age=%d, s-maxage=%d, stale-while-revalidate=%d",
		edgeMaxAge/4, // Browser cache: shorter
		edgeMaxAge,   // Edge cache: longer
		edgeMaxAge/2) // Stale while revalidate

	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("X-Edge-Cache", "ENABLED")
	w.Header().Set("X-Function-Popularity", fmt.Sprintf("%d", popularityScore))
	w.Header().Set("X-Cache-Status", "EDGE") // Will be overridden by CDN if HIT
}

// startEdgeCacheRefresh starts background refresh of edge cache candidates
func (e *EdgeCacheService) startEdgeCacheRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := e.RefreshEdgeCacheCandidates(ctx); err != nil {
			logrus.Errorf("Failed to refresh edge cache candidates: %v", err)
		}
		cancel()
	}
}

// GetEdgeCacheStats returns statistics about edge caching
func (e *EdgeCacheService) GetEdgeCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"enabled": e.cdnService != nil && e.cdnService.IsCDNEnabled(),
		"min_popularity": e.minPopularity,
		"min_executions": e.minExecutions,
		"cache_duration": e.cacheDuration.String(),
	}

	// Get active edge cache functions
	edgeCacheKey := "edge_cache:active_functions"
	var candidates []*EdgeCacheCandidate
	if err := e.registryCache.GetJSON(ctx, edgeCacheKey, &candidates); err == nil {
		stats["active_functions"] = len(candidates)
		stats["total_popularity"] = func() int {
			total := 0
			for _, c := range candidates {
				total += c.PopularityScore
			}
			return total
		}()
	} else {
		stats["active_functions"] = 0
		stats["total_popularity"] = 0
	}

	return stats, nil
}

// PurgeFunctionFromEdge purges a specific function from edge cache
func (e *EdgeCacheService) PurgeFunctionFromEdge(ctx context.Context, functionID uuid.UUID) error {
	if e.cdnService == nil {
		return nil
	}

	// Purge from CDN
	path := fmt.Sprintf("/functions/%s", functionID.String())
	if err := e.cdnService.PurgeCDNCache(path); err != nil {
		return fmt.Errorf("failed to purge function from CDN: %w", err)
	}

	logrus.Infof("Purged function %s from edge cache", functionID.String())
	return nil
}

// UpdateFunctionPopularity persists a popularity change and triggers cache invalidation
// when the function crosses an edge-cache eligibility boundary.
func (e *EdgeCacheService) UpdateFunctionPopularity(ctx context.Context, functionID uuid.UUID, newPopularity int) error {
	logrus.Debugf("Function %s popularity updated to %d", functionID.String(), newPopularity)

	// Persist the new score so the next candidate recomputation uses it
	if e.repository != nil {
		if err := e.repository.UpdateFunctionPopularityScore(ctx, functionID, newPopularity); err != nil {
			logrus.Warnf("Failed to persist popularity score for %s: %v", functionID.String(), err)
		}
	}

	// Invalidate the cached candidates list so the next GetEdgeCacheCandidates
	// call recomputes with the updated score.
	if err := e.registryCache.Delete(ctx, "edge_cache:candidates"); err != nil && err != ErrNotFound {
		logrus.Warnf("Failed to invalidate edge cache candidates for %s: %v", functionID.String(), err)
	}

	// Only trigger an immediate full refresh when crossing an eligibility boundary.
	// Small drift within the same eligibility state is picked up by the next
	// periodic RefreshEdgeCacheCandidates (startEdgeCacheRefresh, every 10 min).
	if e.repository != nil {
		candidate, err := e.repository.GetFunctionEdgeCacheMetrics(ctx, functionID, time.Hour)
		if err == nil && candidate != nil {
			wasEligible := e.IsEligibleForEdgeCaching(
				functionID,
				candidate.PopularityScore,
				candidate.ExecutionCount,
				candidate.TrustScore,
				candidate.SuccessRate,
				candidate.AvgLatency,
			)
			isNowEligible := newPopularity >= e.minPopularity

			if wasEligible != isNowEligible {
				logrus.Infof("Function %s eligibility changed (was %v, now %v); triggering edge cache refresh",
					functionID.String(), wasEligible, isNowEligible)
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					if err := e.RefreshEdgeCacheCandidates(ctx); err != nil {
						logrus.Errorf("Failed to refresh edge cache after eligibility change: %v", err)
					}
				}()
			}
		}
	}

	return nil
}
