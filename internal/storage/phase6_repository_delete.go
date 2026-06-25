package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeletePushSubscription deletes a push subscription by ID.
func (r *Phase6Repository) DeletePushSubscription(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete push subscription: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("push subscription not found")
	}
	return nil
}

// DeleteNotificationPreference deletes a notification preference by ID.
func (r *Phase6Repository) DeleteNotificationPreference(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM notification_preferences WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification preference: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notification preference not found")
	}
	return nil
}
