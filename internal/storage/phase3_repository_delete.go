package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteDocument deletes a document by ID.
func (r *Phase3Repository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM documents WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}
	return nil
}
