package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ManualReviewRepositoryInterface defines the interface for manual review data access
type ManualReviewRepositoryInterface interface {
	CreateManualReviewQueue(ctx context.Context, review *ManualReviewQueue) error
	GetManualReviewQueueByID(ctx context.Context, id uuid.UUID) (*ManualReviewQueue, error)
	GetManualReviewQueueByFunctionVersion(ctx context.Context, functionVersionID uuid.UUID) (*ManualReviewQueue, error)
	UpdateManualReviewQueue(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	AssignManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID) error
	CompleteManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID, decision, reason string) error
	EscalateManualReviewQueue(ctx context.Context, id uuid.UUID, reason string) error
	GetPendingManualReviews(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]ManualReviewQueue, int64, error)
	GetOverdueManualReviews(ctx context.Context, limit int) ([]ManualReviewQueue, error)
	GetManualReviewsByReviewer(ctx context.Context, reviewerID uuid.UUID, statusFilter []string, limit, offset int) ([]ManualReviewQueue, int64, error)
	DeleteManualReviewQueue(ctx context.Context, id uuid.UUID) error
}

// CreateManualReviewQueue creates a new manual review queue entry
func (r *RegistryRepository) CreateManualReviewQueue(ctx context.Context, review *ManualReviewQueue) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("failed to create manual review queue entry: %w", err)
	}

	return nil
}

// GetManualReviewQueueByID retrieves a manual review queue entry by ID
func (r *RegistryRepository) GetManualReviewQueueByID(ctx context.Context, id uuid.UUID) (*ManualReviewQueue, error) {
	var review ManualReviewQueue
	if err := r.db.WithContext(ctx).First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("failed to get manual review queue entry: %w", err)
	}

	return &review, nil
}

// GetManualReviewQueueByFunctionVersion retrieves a manual review queue entry by function version
func (r *RegistryRepository) GetManualReviewQueueByFunctionVersion(ctx context.Context, functionVersionID uuid.UUID) (*ManualReviewQueue, error) {
	var review ManualReviewQueue
	if err := r.db.WithContext(ctx).
		Where("function_version_id = ?", functionVersionID).
		Order("created_at DESC").
		First(&review).Error; err != nil {
		return nil, fmt.Errorf("failed to get manual review queue entry by function version: %w", err)
	}

	return &review, nil
}

// UpdateManualReviewQueue updates a manual review queue entry
func (r *RegistryRepository) UpdateManualReviewQueue(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).
		Model(&ManualReviewQueue{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update manual review queue entry: %w", err)
	}

	return nil
}

// AssignManualReviewQueue assigns a manual review to a reviewer
func (r *RegistryRepository) AssignManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID) error {
	return r.UpdateManualReviewQueue(ctx, id, map[string]interface{}{
		"assigned_to": reviewerID,
		"status":      "in_review",
	})
}

// CompleteManualReviewQueue completes a manual review with a decision
func (r *RegistryRepository) CompleteManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID, decision, reason string) error {
	now := time.Now()
	return r.UpdateManualReviewQueue(ctx, id, map[string]interface{}{
		"status":        decision,
		"decision_by":   reviewerID,
		"decision_at":   now,
		"decision_reason": reason,
		"completed_at":  now,
	})
}

// EscalateManualReviewQueue escalates a manual review to higher priority
func (r *RegistryRepository) EscalateManualReviewQueue(ctx context.Context, id uuid.UUID, reason string) error {
	return r.UpdateManualReviewQueue(ctx, id, map[string]interface{}{
		"status":   "escalated",
		"priority": "urgent",
	})
}

// GetPendingManualReviews retrieves all pending manual reviews with optional filters
func (r *RegistryRepository) GetPendingManualReviews(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]ManualReviewQueue, int64, error) {
	var reviews []ManualReviewQueue
	var total int64

	query := r.db.WithContext(ctx).Model(&ManualReviewQueue{}).Where("status IN ?", []string{"pending", "in_review", "escalated"})

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"].(string); ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if reviewType, ok := filters["review_type"].(string); ok && reviewType != "" {
		query = query.Where("review_type = ?", reviewType)
	}
	if assignedTo, ok := filters["assigned_to"].(uuid.UUID); ok {
		query = query.Where("assigned_to = ?", assignedTo)
	}
	if functionID, ok := filters["function_id"].(uuid.UUID); ok {
		query = query.Where("function_id = ?", functionID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count pending manual reviews: %w", err)
	}

	// Fetch with pagination
	if err := query.
		Order("CASE priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 ELSE 4 END, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get pending manual reviews: %w", err)
	}

	return reviews, total, nil
}

// GetOverdueManualReviews retrieves manual reviews past their due date
func (r *RegistryRepository) GetOverdueManualReviews(ctx context.Context, limit int) ([]ManualReviewQueue, error) {
	var reviews []ManualReviewQueue
	if err := r.db.WithContext(ctx).
		Where("due_at < ? AND status IN ?", time.Now(), []string{"pending", "in_review"}).
		Order("due_at ASC").
		Limit(limit).
		Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue manual reviews: %w", err)
	}

	return reviews, nil
}

// GetManualReviewsByReviewer retrieves all manual reviews assigned to a specific reviewer
func (r *RegistryRepository) GetManualReviewsByReviewer(ctx context.Context, reviewerID uuid.UUID, statusFilter []string, limit, offset int) ([]ManualReviewQueue, int64, error) {
	var reviews []ManualReviewQueue
	var total int64

	query := r.db.WithContext(ctx).Model(&ManualReviewQueue{}).Where("assigned_to = ?", reviewerID)
	if len(statusFilter) > 0 {
		query = query.Where("status IN ?", statusFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count manual reviews by reviewer: %w", err)
	}

	if err := query.
		Order("CASE status WHEN 'in_review' THEN 1 WHEN 'pending' THEN 2 WHEN 'escalated' THEN 3 ELSE 4 END, priority DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get manual reviews by reviewer: %w", err)
	}

	return reviews, total, nil
}

// DeleteManualReviewQueue deletes a manual review queue entry
func (r *RegistryRepository) DeleteManualReviewQueue(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&ManualReviewQueue{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete manual review queue entry: %w", err)
	}

	return nil
}
