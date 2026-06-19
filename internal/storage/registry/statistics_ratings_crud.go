package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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
		"overall_score":        rating.OverallScore,
		"reliability_score":    rating.ReliabilityScore,
		"latency_score":        rating.LatencyScore,
		"documentation_score":  rating.DocumentationScore,
		"total_ratings":        rating.TotalRatings,
		"success_rate":         rating.SuccessRate,
		"p95_latency_ms":       rating.P95LatencyMs,
		"avg_latency_ms":       rating.AvgLatencyMs,
		"p50_latency_ms":       rating.P50LatencyMs,
		"timeout_rate":         rating.TimeoutRate,
		"error_rate":           rating.ErrorRate,
		"consumer_diversity":   rating.ConsumerDiversity,
		"tenant_diversity":     rating.TenantDiversity,
		"user_diversity":       rating.UserDiversity,
		"trust_score":          rating.TrustScore,
		"trust_updated_at":     rating.TrustUpdatedAt,
		"updated_at":           rating.UpdatedAt,
	}).Error; err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	// Invalidate rating cache
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(rating.FunctionID.String())
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					logrus.WithFields(logrus.Fields{
						"panic": rec,
						"stack": fmt.Sprintf("%v", rec),
					}).Error("UpdateRating cache invalidation goroutine panicked")
				}
			}()
			if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
				logrus.WithError(err).Warn("Failed to invalidate rating cache")
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
		"p50_latency_ms":      rating.P50LatencyMs,
		"timeout_rate":        rating.TimeoutRate,
		"error_rate":          rating.ErrorRate,
		"consumer_diversity": rating.ConsumerDiversity,
		"tenant_diversity":    rating.TenantDiversity,
		"user_diversity":      rating.UserDiversity,
		"trust_score":         rating.TrustScore,
		"trust_updated_at":    trustUpdatedAt,
		"success_rate":        rating.SuccessRate,
		"p95_latency_ms":      rating.P95LatencyMs,
		"avg_latency_ms":      rating.AvgLatencyMs,
		"reliability_score":   rating.ReliabilityScore,
		"latency_score":       rating.LatencyScore,
	}).Error; err != nil {
		return fmt.Errorf("failed to update trust score: %w", err)
	}

	// Invalidate rating cache so next GET returns updated trust score
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(rating.FunctionID.String())
		if err := r.cache.Delete(context.Background(), cacheKey); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate rating cache")
		}
	}
	return nil
}

// GetRatingByFunctionID gets a rating by function ID (returns nil if not found)
func (r *RegistryRepository) GetRatingByFunctionID(ctx context.Context, functionID uuid.UUID) (*RegistryFunctionRating, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(functionID.String())
		var rating RegistryFunctionRating
		if err := r.cache.GetJSON(context.Background(), cacheKey, &rating); err == nil {
			return &rating, nil
		}
		// Cache miss - continue to database
	}

	// Use Limit(1).Find to avoid GORM logging "record not found" when no rating exists (expected)
	var ratings []RegistryFunctionRating
	if err := r.db.Where("function_id = ?", functionID).Limit(1).Find(&ratings).Error; err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	if len(ratings) == 0 {
		return nil, nil
	}
	rating := ratings[0]

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionRating(functionID.String())
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, rating); err != nil {
				logrus.WithError(err).Warn("Failed to cache function rating")
			}
		}()
	}

	return &rating, nil
}

// GetRatingsByFunctionIDs loads ratings for many functions in one query (used by registry list/search).
func (r *RegistryRepository) GetRatingsByFunctionIDs(functionIDs []uuid.UUID) (map[uuid.UUID]*RegistryFunctionRating, error) {
	if len(functionIDs) == 0 {
		return map[uuid.UUID]*RegistryFunctionRating{}, nil
	}
	var ratings []RegistryFunctionRating
	if err := r.db.Where("function_id IN ?", functionIDs).Find(&ratings).Error; err != nil {
		return nil, fmt.Errorf("failed to batch-load ratings: %w", err)
	}
	out := make(map[uuid.UUID]*RegistryFunctionRating, len(ratings))
	for i := range ratings {
		out[ratings[i].FunctionID] = &ratings[i]
	}
	return out, nil
}

// UpdateFunctionPopularity sets the popularity score for a function
func (r *RegistryRepository) UpdateFunctionPopularity(functionID uuid.UUID, score int) error {
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", functionID).
		UpdateColumn("popularity_score", score).Error; err != nil {
		return fmt.Errorf("failed to update popularity score: %w", err)
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
