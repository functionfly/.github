package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreatePendingUsernameChange creates a new pending username change record
func (r *UserRepository) CreatePendingUsernameChange(ctx context.Context, pending *PendingUsernameChange) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pending_username_changes (
			id, user_id, old_username, new_username, status, checkout_session_id,
			fee_cents, fee_currency, created_at, updated_at, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		pending.ID, pending.UserID, pending.OldUsername, pending.NewUsername,
		pending.Status, pending.CheckoutSessionID, pending.FeeCents,
		pending.FeeCurrency, pending.CreatedAt, pending.UpdatedAt,
		pending.IPAddress, pending.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to create pending username change: %w", err)
	}
	return nil
}

// GetPendingUsernameChangeByID retrieves a pending username change by ID
func (r *UserRepository) GetPendingUsernameChangeByID(ctx context.Context, id uuid.UUID) (*PendingUsernameChange, error) {
	p := &PendingUsernameChange{}
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, old_username, new_username, status, checkout_session_id,
			   fee_cents, fee_currency, created_at, updated_at, completed_at,
			   ip_address, user_agent
		FROM pending_username_changes
		WHERE id = $1`, id).Scan(
		&p.ID, &p.UserID, &p.OldUsername, &p.NewUsername, &p.Status,
		&p.CheckoutSessionID, &p.FeeCents, &p.FeeCurrency,
		&p.CreatedAt, &p.UpdatedAt, &completedAt,
		&p.IPAddress, &p.UserAgent,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pending username change: %w", err)
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Time
	}
	return p, nil
}

// GetPendingUsernameChangeByCheckoutSession retrieves a pending change by checkout session ID
func (r *UserRepository) GetPendingUsernameChangeByCheckoutSession(ctx context.Context, sessionID string) (*PendingUsernameChange, error) {
	p := &PendingUsernameChange{}
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, old_username, new_username, status, checkout_session_id,
			   fee_cents, fee_currency, created_at, updated_at, completed_at,
			   ip_address, user_agent
		FROM pending_username_changes
		WHERE checkout_session_id = $1`, sessionID).Scan(
		&p.ID, &p.UserID, &p.OldUsername, &p.NewUsername, &p.Status,
		&p.CheckoutSessionID, &p.FeeCents, &p.FeeCurrency,
		&p.CreatedAt, &p.UpdatedAt, &completedAt,
		&p.IPAddress, &p.UserAgent,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pending username change by session: %w", err)
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Time
	}
	return p, nil
}

// UpdatePendingUsernameChangeStatus updates the status of a pending username change
func (r *UserRepository) UpdatePendingUsernameChangeStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pending_username_changes
		SET status = $1, updated_at = $2,
		    completed_at = CASE WHEN $1 = 'completed' THEN $2 ELSE completed_at END
		WHERE id = $3`,
		status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update pending username change status: %w", err)
	}
	return nil
}

// DeleteExpiredPendingUsernameChanges removes expired pending changes
func (r *UserRepository) DeleteExpiredPendingUsernameChanges(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM pending_username_changes
		WHERE status = 'pending' AND created_at < $1`,
		time.Now().Add(-24*time.Hour))
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired pending changes: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ListPendingUsernameChangesForUser returns all pending changes for a user
func (r *UserRepository) ListPendingUsernameChangesForUser(ctx context.Context, userID uuid.UUID) ([]*PendingUsernameChange, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, old_username, new_username, status, checkout_session_id,
			   fee_cents, fee_currency, created_at, updated_at, completed_at,
			   ip_address, user_agent
		FROM pending_username_changes
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending username changes: %w", err)
	}
	defer rows.Close()

	var pending []*PendingUsernameChange
	for rows.Next() {
		p := &PendingUsernameChange{}
		var completedAt sql.NullTime
		err := rows.Scan(
			&p.ID, &p.UserID, &p.OldUsername, &p.NewUsername, &p.Status,
			&p.CheckoutSessionID, &p.FeeCents, &p.FeeCurrency,
			&p.CreatedAt, &p.UpdatedAt, &completedAt,
			&p.IPAddress, &p.UserAgent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending username change: %w", err)
		}
		if completedAt.Valid {
			p.CompletedAt = &completedAt.Time
		}
		pending = append(pending, p)
	}
	return pending, nil
}
