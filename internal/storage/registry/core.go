package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common repository errors
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrReviewNotFound = errors.New("manual review not found")
)

// RegistryRepository handles function registry database operations
type RegistryRepository struct {
	db     *gorm.DB
	cache  *cache.RegistryRedisCache
	keyGen *cache.RegistryCacheKey
}

// NewRegistryRepository creates a new registry repository
func NewRegistryRepository(db *gorm.DB, cacheClient *cache.RegistryRedisCache) *RegistryRepository {
	var keyGen *cache.RegistryCacheKey

	if cacheClient != nil {
		keyGen = cache.NewRegistryCacheKey()
	}

	return &RegistryRepository{
		db:     db,
		cache:  cacheClient,
		keyGen: keyGen,
	}
}

// InvalidateListCache invalidates function list caches so description/category updates appear after publish.
func (r *RegistryRepository) InvalidateListCache(ctx context.Context) {
	if r.cache != nil {
		_ = r.cache.InvalidateListResults(ctx)
	}
}

// GetRecentExecutions retrieves the most recent executions for a function
func (r *RegistryRepository) GetRecentExecutions(functionID uuid.UUID, limit int) ([]RegistryFunctionExecution, error) {
	var executions []RegistryFunctionExecution
	err := r.db.Where("function_id = ?", functionID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// GetRecentFailedExecutions retrieves the most recent non-success executions for a function (outcome != 'success').
func (r *RegistryRepository) GetRecentFailedExecutions(functionID uuid.UUID, limit int) ([]RegistryFunctionExecution, error) {
	var executions []RegistryFunctionExecution
	err := r.db.Where("function_id = ? AND outcome != ?", functionID, "success").
		Order("timestamp DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// GetRecentPublicExecutions retrieves the most recent public executions for a function
func (r *RegistryRepository) GetRecentPublicExecutions(functionID uuid.UUID, limit int) ([]RegistryExecutionPublic, error) {
	var executions []RegistryExecutionPublic
	err := r.db.Where("function_id = ?", functionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// CreateExecutionPublic creates a new shareable execution record for playground/replay
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

// GetExecutionCountForVersion returns the total number of executions for a specific function version
func (r *RegistryRepository) GetExecutionCountForVersion(functionID uuid.UUID, version string) (int, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).Where("function_id = ? AND version = ?", functionID, version).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}
	return int(count), nil
}

// GetLastVerificationTimeForVersion returns the most recent verification time for a specific function version
func (r *RegistryRepository) GetLastVerificationTimeForVersion(functionID uuid.UUID, version string) (*time.Time, error) {
	var exec RegistryFunctionExecution
	if err := r.db.Where("function_id = ? AND version = ? AND verified_at IS NOT NULL", functionID, version).
		Order("verified_at DESC").First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
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

	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ?",
			functionID, version, sevenDaysAgo).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verifications: %w", err)
	}

	if total == 0 {
		return 0, nil
	}

	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ? AND verification_status = ?",
			functionID, version, sevenDaysAgo, "failed").Count(&failed).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verification failures: %w", err)
	}

	return float64(failed) / float64(total), nil
}

// RecordExecution records a function execution
func (r *RegistryRepository) RecordExecution(exec *RegistryFunctionExecution) error {
	exec.ID = uuid.New()
	exec.Timestamp = time.Now()
	if err := r.db.Create(exec).Error; err != nil {
		return fmt.Errorf("failed to record execution: %w", err)
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
