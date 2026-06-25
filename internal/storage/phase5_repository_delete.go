package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteFeatureFlag deletes a feature flag by ID.
func (r *Phase5Repository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM feature_flags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("feature flag not found")
	}
	return nil
}

// DeleteDataClassification deletes a data classification by ID.
func (r *Phase5Repository) DeleteDataClassification(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM data_classifications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete data classification: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("data classification not found")
	}
	return nil
}
