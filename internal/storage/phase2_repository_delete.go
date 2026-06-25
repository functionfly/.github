package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteChatSession soft-deletes a chat session.
func (r *Phase2Repository) DeleteChatSession(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `UPDATE ai_chat_sessions SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("chat session not found")
	}
	return nil
}

// DeleteTimeEntry deletes a time entry by ID.
func (r *Phase2Repository) DeleteTimeEntry(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM time_entries WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("time entry not found")
	}
	return nil
}
