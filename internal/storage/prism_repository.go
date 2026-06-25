package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PrismRepository handles Prism runtime persistence.
type PrismRepository struct {
	db *gorm.DB
}

// NewPrismRepository creates a new Prism repository.
func NewPrismRepository(db *gorm.DB) *PrismRepository {
	return &PrismRepository{db: db}
}

// AutoMigrate creates or updates the Prism tables.
func (r *PrismRepository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&PrismCell{},
		&PrismHeartbeat{},
		&PrismExecutionResult{},
		&PrismCapability{},
		&PrismRuntimeStatus{},
	)
}

// UpsertCell registers or updates a Prism cell.
func (r *PrismRepository) UpsertCell(ctx context.Context, cell *PrismCell) error {
	now := time.Now()
	cell.UpdatedAt = now
	if cell.RegisteredAt.IsZero() {
		cell.RegisteredAt = now
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cell_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "status", "capabilities", "metadata", "updated_at"}),
		}).
		Create(cell).Error
}

// UpdateCellHeartbeat updates the last heartbeat time and status for a cell.
func (r *PrismRepository) UpdateCellHeartbeat(ctx context.Context, cellID, status string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&PrismCell{}).
		Where("cell_id = ?", cellID).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"status":         status,
			"updated_at":     now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update cell heartbeat: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("prism cell not found: %s", cellID)
	}
	return nil
}

// GetCell retrieves a Prism cell by its cell_id.
func (r *PrismRepository) GetCell(ctx context.Context, cellID string) (*PrismCell, error) {
	var cell PrismCell
	err := r.db.WithContext(ctx).Where("cell_id = ?", cellID).First(&cell).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get prism cell: %w", err)
	}
	return &cell, nil
}

// ListCells returns all Prism cells, optionally filtered by status.
func (r *PrismRepository) ListCells(ctx context.Context, status string) ([]*PrismCell, error) {
	var cells []*PrismCell
	q := r.db.WithContext(ctx).Order("registered_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&cells).Error; err != nil {
		return nil, fmt.Errorf("failed to list prism cells: %w", err)
	}
	return cells, nil
}

// ListActiveCells returns cells that have sent a heartbeat within the given timeout.
func (r *PrismRepository) ListActiveCells(ctx context.Context, timeout time.Duration) ([]*PrismCell, error) {
	var cells []*PrismCell
	cutoff := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).
		Where("last_heartbeat > ? AND status != 'terminated'", cutoff).
		Order("last_heartbeat DESC").
		Find(&cells).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list active prism cells: %w", err)
	}
	return cells, nil
}

// TerminateCell marks a cell as terminated.
func (r *PrismRepository) TerminateCell(ctx context.Context, cellID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&PrismCell{}).
		Where("cell_id = ?", cellID).
		Updates(map[string]interface{}{
			"status":        "terminated",
			"terminated_at": now,
			"updated_at":    now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to terminate prism cell: %w", result.Error)
	}
	return nil
}

// RecordHeartbeat persists a heartbeat record.
func (r *PrismRepository) RecordHeartbeat(ctx context.Context, hb *PrismHeartbeat) error {
	if hb.ReceivedAt.IsZero() {
		hb.ReceivedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(hb).Error
}

// RecordExecutionResult persists an execution result.
func (r *PrismRepository) RecordExecutionResult(ctx context.Context, result *PrismExecutionResult) error {
	if result.ReceivedAt.IsZero() {
		result.ReceivedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(result).Error
}

// UpsertCapability registers or updates a capability for a cell.
func (r *PrismRepository) UpsertCapability(ctx context.Context, cap *PrismCapability) error {
	if cap.AnnouncedAt.IsZero() {
		cap.AnnouncedAt = time.Now()
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cell_id"}, {Name: "capability"}},
			DoUpdates: clause.AssignmentColumns([]string{"trust_score", "metadata", "announced_at"}),
		}).
		Create(cap).Error
}

// GetCapabilities returns all capabilities for a given cell.
func (r *PrismRepository) GetCapabilities(ctx context.Context, cellID string) ([]*PrismCapability, error) {
	var caps []*PrismCapability
	err := r.db.WithContext(ctx).
		Where("cell_id = ?", cellID).
		Order("capability ASC").
		Find(&caps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get capabilities: %w", err)
	}
	return caps, nil
}

// RecordRuntimeStatus persists a runtime status report.
func (r *PrismRepository) RecordRuntimeStatus(ctx context.Context, status *PrismRuntimeStatus) error {
	if status.ReceivedAt.IsZero() {
		status.ReceivedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(status).Error
}

// GetLatestRuntimeStatus returns the most recent runtime status report.
func (r *PrismRepository) GetLatestRuntimeStatus(ctx context.Context) (*PrismRuntimeStatus, error) {
	var status PrismRuntimeStatus
	err := r.db.WithContext(ctx).
		Order("received_at DESC").
		First(&status).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime status: %w", err)
	}
	return &status, nil
}

// CleanupOldHeartbeats deletes heartbeat records older than the given duration.
func (r *PrismRepository) CleanupOldHeartbeats(ctx context.Context, maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)
	result := r.db.WithContext(ctx).
		Where("received_at < ?", cutoff).
		Delete(&PrismHeartbeat{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup heartbeats: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// CleanupOldExecutionResults deletes execution results older than the given duration.
func (r *PrismRepository) CleanupOldExecutionResults(ctx context.Context, maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)
	result := r.db.WithContext(ctx).
		Where("received_at < ?", cutoff).
		Delete(&PrismExecutionResult{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup execution results: %w", result.Error)
	}
	return result.RowsAffected, nil
}
