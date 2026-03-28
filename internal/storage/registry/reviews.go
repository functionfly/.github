package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistryFunctionReview is a per-user review for a function (optional text + 1-5 stars).
type RegistryFunctionReview struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	FunctionID uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`

	Stars int    `json:"stars" gorm:"not null"` // 1-5
	Title string `json:"title" gorm:"type:text"`
	Body  string `json:"body" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RegistryFunctionReview) TableName() string { return "registry_function_reviews" }

func (r *RegistryRepository) ensureReviewsTable(ctx context.Context) error {
	// We avoid relying on global migrations in dev (many environments run with --skip-migrations).
	// This uses idempotent CREATE TABLE IF NOT EXISTS.
	sql := `
CREATE TABLE IF NOT EXISTS registry_function_reviews (
  id uuid PRIMARY KEY,
  function_id uuid NOT NULL,
  user_id uuid NOT NULL,
  stars integer NOT NULL,
  title text NOT NULL DEFAULT '',
  body text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS registry_function_reviews_function_user_uidx
  ON registry_function_reviews(function_id, user_id);
CREATE INDEX IF NOT EXISTS registry_function_reviews_function_idx
  ON registry_function_reviews(function_id, created_at DESC);
`
	if err := r.db.WithContext(ctx).Exec(sql).Error; err != nil {
		return fmt.Errorf("ensure reviews table: %w", err)
	}
	return nil
}

// HasUserExecutedFunction returns true if the user has executed the function at least once.
func (r *RegistryRepository) HasUserExecutedFunction(ctx context.Context, functionID, userID uuid.UUID) (bool, error) {
	// registry_function_executions.user_id is nullable, so only count rows with a match.
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND user_id = ?", functionID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check user execution: %w", err)
	}
	return count > 0, nil
}

// UpsertFunctionReview creates or updates the user's review for a function.
func (r *RegistryRepository) UpsertFunctionReview(ctx context.Context, functionID, userID uuid.UUID, stars int, title, body string) (*RegistryFunctionReview, error) {
	if err := r.ensureReviewsTable(ctx); err != nil {
		return nil, err
	}
	if stars < 1 || stars > 5 {
		return nil, fmt.Errorf("stars must be between 1 and 5")
	}

	var existing RegistryFunctionReview
	err := r.db.WithContext(ctx).
		Where("function_id = ? AND user_id = ?", functionID, userID).
		First(&existing).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get existing review: %w", err)
	}

	now := time.Now()
	if err == nil {
		existing.Stars = stars
		existing.Title = title
		existing.Body = body
		existing.UpdatedAt = now
		if e := r.db.WithContext(ctx).Save(&existing).Error; e != nil {
			return nil, fmt.Errorf("update review: %w", e)
		}
		return &existing, nil
	}

	review := &RegistryFunctionReview{
		ID:         uuid.New(),
		FunctionID: functionID,
		UserID:     userID,
		Stars:      stars,
		Title:      title,
		Body:       body,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if e := r.db.WithContext(ctx).Create(review).Error; e != nil {
		return nil, fmt.Errorf("create review: %w", e)
	}
	return review, nil
}

// ListFunctionReviews returns recent reviews for a function.
func (r *RegistryRepository) ListFunctionReviews(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]RegistryFunctionReview, int64, error) {
	if err := r.ensureReviewsTable(ctx); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := r.db.WithContext(ctx).Model(&RegistryFunctionReview{}).Where("function_id = ?", functionID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	var rows []RegistryFunctionReview
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	return rows, total, nil
}

