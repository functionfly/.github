package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteDigitalBadge deletes a digital badge by ID.
func (r *Phase4Repository) DeleteDigitalBadge(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM digital_badges WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete digital badge: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("digital badge not found")
	}
	return nil
}

// DeleteLivingMemoryEntry deletes a living memory entry by ID.
func (r *Phase4Repository) DeleteLivingMemoryEntry(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM living_memory WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete living memory entry: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("living memory entry not found")
	}
	return nil
}

// RevokeEmployeeBadge removes a badge from an employee.
func (r *Phase4Repository) RevokeEmployeeBadge(ctx context.Context, employeeID, badgeID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM employee_badges WHERE employee_id = $1 AND badge_id = $2`, employeeID, badgeID)
	if err != nil {
		return fmt.Errorf("failed to revoke employee badge: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("employee badge not found")
	}
	return nil
}
