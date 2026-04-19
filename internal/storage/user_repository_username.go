package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateUsernameChangeHistory records a username change in the history
func (r *UserRepository) CreateUsernameChangeHistory(ctx context.Context, history *UsernameChangeHistory) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO username_change_history (
			id, user_id, old_username, new_username, changed_at, changed_by,
			was_early_change, fee_paid_cents, fee_currency, stripe_payment_id,
			ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		history.ID, history.UserID, history.OldUsername, history.NewUsername,
		history.ChangedAt, history.ChangedBy, history.WasEarlyChange,
		history.FeePaidCents, history.FeeCurrency, history.StripePaymentID,
		history.IPAddress, history.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to create username change history: %w", err)
	}
	return nil
}

// GetUsernameChangeHistory returns all username changes for a user, ordered by most recent first
func (r *UserRepository) GetUsernameChangeHistory(ctx context.Context, userID uuid.UUID) ([]*UsernameChangeHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, old_username, new_username, changed_at, changed_by,
			   was_early_change, fee_paid_cents, fee_currency, stripe_payment_id,
			   ip_address, user_agent
		FROM username_change_history
		WHERE user_id = $1
		ORDER BY changed_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get username change history: %w", err)
	}
	defer rows.Close()

	var history []*UsernameChangeHistory
	for rows.Next() {
		h := &UsernameChangeHistory{}
		var changedBy uuid.NullUUID
		var stripePaymentID, ipAddress, userAgent *string
		err := rows.Scan(
			&h.ID, &h.UserID, &h.OldUsername, &h.NewUsername, &h.ChangedAt, &changedBy,
			&h.WasEarlyChange, &h.FeePaidCents, &h.FeeCurrency, &stripePaymentID,
			&ipAddress, &userAgent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan username change history: %w", err)
		}
		if changedBy.Valid {
			h.ChangedBy = changedBy.UUID
		}
		h.StripePaymentID = stripePaymentID
		h.IPAddress = derefString(ipAddress)
		h.UserAgent = derefString(userAgent)
		history = append(history, h)
	}
	return history, nil
}

// CountUsernameChangesInWindow counts how many username changes a user made in a given time window
func (r *UserRepository) CountUsernameChangesInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM username_change_history
		WHERE user_id = $1 AND changed_at >= $2`,
		userID, windowStart).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count username changes: %w", err)
	}
	return count, nil
}

// GetLastUsernameChange returns the most recent username change for a user
// Returns nil, nil if no changes exist
func (r *UserRepository) GetLastUsernameChange(ctx context.Context, userID uuid.UUID) (*UsernameChangeHistory, error) {
	h := &UsernameChangeHistory{}
	var changedBy uuid.NullUUID
	var stripePaymentID, ipAddress, userAgent *string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, old_username, new_username, changed_at, changed_by,
			   was_early_change, fee_paid_cents, fee_currency, stripe_payment_id,
			   ip_address, user_agent
		FROM username_change_history
		WHERE user_id = $1
		ORDER BY changed_at DESC
		LIMIT 1`,
		userID).Scan(
		&h.ID, &h.UserID, &h.OldUsername, &h.NewUsername, &h.ChangedAt, &changedBy,
		&h.WasEarlyChange, &h.FeePaidCents, &h.FeeCurrency, &stripePaymentID,
		&ipAddress, &userAgent,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last username change: %w", err)
	}
	if changedBy.Valid {
		h.ChangedBy = changedBy.UUID
	}
	h.StripePaymentID = stripePaymentID
	h.IPAddress = derefString(ipAddress)
	h.UserAgent = derefString(userAgent)
	return h, nil
}

// HasUsernameChangedInWindow checks if user has any username changes in the given time window
func (r *UserRepository) HasUsernameChangedInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM username_change_history
			WHERE user_id = $1 AND changed_at >= $2
		)`,
		userID, windowStart).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username change history: %w", err)
	}
	return exists, nil
}

// derefString safely dereferences a string pointer
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
