package registry

import (
	"fmt"
	"time"
	"context"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckExecutionCache checks if there's a cached result for the given function execution
func (r *RegistryRepository) CheckExecutionCache(ctx context.Context, functionID, version string, input interface{}) (*types.ExecutionResult, error) {
	if r.executionCache == nil {
		return nil, nil // No cache available
	}

	return r.executionCache.GetExecutionResult(ctx, functionID, version, input)
}

// CacheExecutionResult caches the result of a function execution
func (r *RegistryRepository) CacheExecutionResult(ctx context.Context, functionID, version string, input interface{}, result *types.ExecutionResult) error {
	if r.executionCache == nil {
		return nil // No cache available
	}

	return r.executionCache.SetExecutionResult(ctx, functionID, version, input, result)
}

// RecordExecution records a function execution
func (r *RegistryRepository) RecordExecution(exec *RegistryFunctionExecution) error {
	exec.ID = uuid.New()
	exec.Timestamp = time.Now()

	if err := r.db.Create(exec).Error; err != nil {
		return fmt.Errorf("failed to record execution: %w", err)
	}

	// Also record in cache for analytics if available
	if r.executionCache != nil {
		ctx := context.Background()
		if err := r.executionCache.RecordExecution(ctx, exec.ID.String(), exec.FunctionID.String(), exec.Version, exec.Outcome, exec.DurationMs, exec.Cached, exec.Timestamp); err != nil {
			// Log but don't fail the execution
			fmt.Printf("Failed to record execution in cache: %v\n", err)
		}
	}

	return nil
}

// RecordResourceUsage records resource usage for an execution
func (r *RegistryRepository) RecordResourceUsage(usage *ExecutionResourceUsage) error {
	usage.ID = uuid.New()
	usage.CreatedAt = time.Now()

	if err := r.db.Create(usage).Error; err != nil {
		return fmt.Errorf("failed to record resource usage: %w", err)
	}

	return nil
}

// CreateExecutionPublic creates a new shareable execution record
func (r *RegistryRepository) CreateExecutionPublic(exec *RegistryExecutionPublic) error {
	exec.ID = uuid.New()
	exec.CreatedAt = time.Now()

	if err := r.db.Create(exec).Error; err != nil {
		return fmt.Errorf("failed to create public execution: %w", err)
	}

	return nil
}

// GetExecutionPublicByID retrieves a shareable execution by its public ID
func (r *RegistryRepository) GetExecutionPublicByID(publicID string) (*RegistryExecutionPublic, error) {
	var exec RegistryExecutionPublic
	if err := r.db.Where("public_id = ? AND shareable = ?", publicID, true).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found or not shareable")
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &exec, nil
}

// CachedExecutionCount represents cached execution count
type CachedExecutionCount struct {
	Count    int       `json:"count"`
	CachedAt time.Time `json:"cached_at"`
}

// GetExecutionCountForVersion returns the total number of executions for a specific function version (with caching)
func (r *RegistryRepository) GetExecutionCountForVersion(functionID uuid.UUID, version string) (int, error) {
	// Try cache first if available
	if r.cache != nil {
		cacheKey := r.keyGen.ExecutionCount(functionID.String(), version)
		var cached CachedExecutionCount
		if err := r.cache.GetJSON(context.Background(), cacheKey, &cached); err == nil {
			// Check if cache is still fresh (within 15 minutes for execution counts)
			if time.Since(cached.CachedAt) < 15*time.Minute {
				return cached.Count, nil
			}
		}
	}

	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).Where("function_id = ? AND version = ?", functionID, version).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil {
		cacheKey := r.keyGen.ExecutionCount(functionID.String(), version)
		cachedResult := CachedExecutionCount{
			Count:    int(count),
			CachedAt: time.Now(),
		}
		// Cache for 15 minutes for execution counts
		if err := r.cache.SetJSONWithTTL(context.Background(), cacheKey, cachedResult, 15*time.Minute); err != nil {
			// Log but don't fail the request
			fmt.Printf("Failed to cache execution count: %v\n", err)
		}
	}

	return int(count), nil
}

// GetLastVerificationTimeForVersion returns the most recent verification time for a specific function version
func (r *RegistryRepository) GetLastVerificationTimeForVersion(functionID uuid.UUID, version string) (*time.Time, error) {
	var exec RegistryFunctionExecution
	if err := r.db.Where("function_id = ? AND version = ? AND verified_at IS NOT NULL", functionID, version).
		Order("verified_at DESC").First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No verification found
		}
		return nil, fmt.Errorf("failed to get last verification time: %w", err)
	}
	return &exec.VerifiedAt.Time, nil
}

// GetRecentVerificationFailureRate returns the failure rate for verifications in the last 7 days for a function version
func (r *RegistryRepository) GetRecentVerificationFailureRate(functionID uuid.UUID, version string) (float64, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	var total int64
	var failed int64

	// Count total verifications in the last 7 days
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ?",
			functionID, version, sevenDaysAgo).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verifications: %w", err)
	}

	if total == 0 {
		return 0, nil // No recent verifications
	}

	// Count failed verifications in the last 7 days
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ? AND verification_status = ?",
			functionID, version, sevenDaysAgo, "failed").Count(&failed).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verification failures: %w", err)
	}

	return float64(failed) / float64(total), nil
}

// GetExecutionCountByUserInTimeWindow returns the number of executions by a user in the specified time window
func (r *RegistryRepository) GetExecutionCountByUserInTimeWindow(userID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("user_id = ? AND timestamp >= ?", userID, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions by user: %w", err)
	}
	return count, nil
}

// GetExecutionCountByIPInTimeWindow returns the number of executions by IP address in the specified time window
func (r *RegistryRepository) GetExecutionCountByIPInTimeWindow(ipAddress string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("caller_ip = ? AND timestamp >= ?", ipAddress, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions by IP: %w", err)
	}
	return count, nil
}

// GetRecentExecutionRate returns the execution rate (executions per minute) for a user over the last time window
func (r *RegistryRepository) GetRecentExecutionRate(userID uuid.UUID, windowMinutes int) (float64, error) {
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("user_id = ? AND timestamp >= ?", userID, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent executions: %w", err)
	}
	return float64(count) / float64(windowMinutes), nil
}

// GetExecutionSpikes returns executions that exceed normal patterns for rate spike detection
func (r *RegistryRepository) GetExecutionSpikes(userID *uuid.UUID, ipAddress string, thresholdMultiplier float64, windowMinutes int) ([]RegistryFunctionExecution, error) {
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	// Get baseline rate (average executions per minute over the last 24 hours)
	baselineWindow := time.Now().Add(-24 * time.Hour)
	var baselineCount int64
	query := r.db.Model(&RegistryFunctionExecution{}).Where("timestamp >= ?", baselineWindow)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("caller_ip = ?", ipAddress)
	}
	if err := query.Count(&baselineCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get baseline count: %w", err)
	}
	baselineRate := float64(baselineCount) / (24 * 60) // executions per minute over 24 hours

	// Get current window executions
	var executions []RegistryFunctionExecution
	query = r.db.Model(&RegistryFunctionExecution{}).Where("timestamp >= ?", since)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("caller_ip = ?", ipAddress)
	}
	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}

	// Filter for spikes (current rate exceeds baseline by threshold)
	currentRate := float64(len(executions)) / float64(windowMinutes)
	if currentRate > baselineRate*thresholdMultiplier {
		return executions, nil
	}

	return []RegistryFunctionExecution{}, nil
}
