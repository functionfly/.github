package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ListActiveUsersByTenant lists all active (non-deactivated) users for a tenant
func (r *UserRepository) ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*User, error) {
	query := `SELECT id, tenant_id, username, email, password_hash, role, email_verified,
		company_name, deactivated_at, deactivated_by, created_at, updated_at
		FROM users WHERE tenant_id = $1 AND deactivated_at IS NULL ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var username, companyName, role sql.NullString
		var deactivatedAt sql.NullTime
		var deactivatedBy uuid.NullUUID

		if err := rows.Scan(
			&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role,
			&user.EmailVerified, &companyName, &deactivatedAt, &deactivatedBy,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if username.Valid {
			user.Username = &username.String
		}
		if companyName.Valid {
			user.CompanyName = &companyName.String
		}
		if role.Valid {
			user.Role = role.String
		}
		if deactivatedAt.Valid {
			user.DeactivatedAt = &deactivatedAt.Time
		}
		if deactivatedBy.Valid {
			user.DeactivatedBy = &deactivatedBy.UUID
		}

		users = append(users, user)
	}

	return users, nil
}

// CountActiveUsersByTenant counts all active (non-deactivated) users for a tenant
func (r *UserRepository) CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deactivated_at IS NULL`,
		tenantID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}

	return count, nil
}

// DeactivateUser soft-deletes a user (sets deactivated_at and deactivated_by)
func (r *UserRepository) DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET deactivated_at = $1, deactivated_by = $2, updated_at = $3
		WHERE id = $4 AND deactivated_at IS NULL`,
		now, deactivatedBy, now, userID)

	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// ReactivateUser reactivates a previously deactivated user
func (r *UserRepository) ReactivateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET deactivated_at = NULL, deactivated_by = NULL, updated_at = $1
		WHERE id = $2`,
		time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	return nil
}

// UpdateUserLastActive updates the user's last_active_at timestamp to now
// This should be called periodically to track user online status.
// Throttled: only writes if last update was >30s ago to avoid hot-path DB churn.
func (r *UserRepository) UpdateUserLastActive(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET last_active_at = $1
		WHERE id = $2
		  AND (last_active_at IS NULL OR last_active_at < NOW() - INTERVAL '30 seconds')`,
		now, userID)

	if err != nil {
		return fmt.Errorf("failed to update user last active: %w", err)
	}

	return nil
}

// GetUserLastActive retrieves the last_active_at timestamp for a user
func (r *UserRepository) GetUserLastActive(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	var lastActive sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT last_active_at FROM users WHERE id = $1`, userID).Scan(&lastActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user last active: %w", err)
	}

	if lastActive.Valid {
		return &lastActive.Time, nil
	}
	return nil, nil
}

// GetOnlineUsers returns a list of user IDs who were active within the last N minutes
func (r *UserRepository) GetOnlineUsers(ctx context.Context, withinMinutes int) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM users
		WHERE last_active_at > NOW() - INTERVAL '1 minute' * $1`,
		withinMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan online user: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
