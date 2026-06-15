package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DefaultBatchSize is the default number of records to delete per batch
const DefaultBatchSize = 1000

// BatchDeleteExpiredSessions deletes expired sessions in batches to avoid long-running transactions
func (r *SessionRepository) BatchDeleteExpiredSessions(ctx context.Context, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	var totalDeleted int64

	for {
		var ids []uuid.UUID
		err := r.db.GORM.WithContext(ctx).
			Table("sessions").
			Where("expires_at <= ?", time.Now()).
			Limit(batchSize).
			Pluck("id", &ids).Error

		if err != nil {
			return totalDeleted, fmt.Errorf("failed to query expired sessions: %w", err)
		}

		if len(ids) == 0 {
			break
		}

		result := r.db.GORM.WithContext(ctx).
			Table("sessions").
			Where("id IN ?", ids).
			Delete(nil)

		if result.Error != nil {
			return totalDeleted, fmt.Errorf("failed to delete session batch: %w", result.Error)
		}

		totalDeleted += result.RowsAffected

		if result.RowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

// BatchDeleteOldLoginAttempts deletes old login attempts in batches
func (r *LoginAttemptRepository) BatchDeleteOldLoginAttempts(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	var totalDeleted int64

	for {
		var ids []uuid.UUID
		err := r.db.GORM.WithContext(ctx).
			Table("login_attempts").
			Where("created_at < ?", before).
			Limit(batchSize).
			Pluck("id", &ids).Error

		if err != nil {
			return totalDeleted, fmt.Errorf("failed to query old login attempts: %w", err)
		}

		if len(ids) == 0 {
			break
		}

		result := r.db.GORM.WithContext(ctx).
			Table("login_attempts").
			Where("id IN ?", ids).
			Delete(nil)

		if result.Error != nil {
			return totalDeleted, fmt.Errorf("failed to delete login attempt batch: %w", result.Error)
		}

		totalDeleted += result.RowsAffected

		if result.RowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

// BatchDeleteOldAuthEvents deletes old auth events in batches
func (r *AuthEventRepository) BatchDeleteOldAuthEvents(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	var totalDeleted int64

	for {
		var ids []uuid.UUID
		err := r.db.GORM.WithContext(ctx).
			Table("auth_events").
			Where("created_at < ?", before).
			Limit(batchSize).
			Pluck("id", &ids).Error

		if err != nil {
			return totalDeleted, fmt.Errorf("failed to query old auth events: %w", err)
		}

		if len(ids) == 0 {
			break
		}

		result := r.db.GORM.WithContext(ctx).
			Table("auth_events").
			Where("id IN ?", ids).
			Delete(nil)

		if result.Error != nil {
			return totalDeleted, fmt.Errorf("failed to delete auth event batch: %w", result.Error)
		}

		totalDeleted += result.RowsAffected

		if result.RowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}
