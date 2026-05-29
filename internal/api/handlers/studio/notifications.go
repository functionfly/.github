package studio

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StudioNotification represents an in-app Studio notification.
type StudioNotification struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      *string   `json:"user_id,omitempty"`
	Environment string    `json:"environment"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Read        bool      `json:"read"`
	ActionURL   *string   `json:"action_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateNotificationParams holds fields for creating a notification.
type CreateNotificationParams struct {
	TenantID    string
	UserID      *string
	Environment string
	Type        string
	Category    string
	Title       string
	Message     string
	ActionURL   *string
}

// NotificationRepository handles studio_notifications persistence.
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a notification repository.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ListNotifications returns notifications for a tenant/user/environment.
func (r *NotificationRepository) ListNotifications(ctx context.Context, tenantID, userID, environment string, unreadOnly bool, limit, offset int) ([]StudioNotification, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, user_id, environment, type, category, title, message, read, action_url, created_at
		FROM studio_notifications
		WHERE tenant_id = $1
		  AND (user_id IS NULL OR user_id = $2)
		  AND COALESCE(environment, '') = $3`
	args := []interface{}{tenantID, userID, environment}
	argN := 4

	if unreadOnly {
		query += fmt.Sprintf(" AND read = FALSE")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []StudioNotification
	for rows.Next() {
		var n StudioNotification
		var userIDRaw sql.NullString
		var actionURL sql.NullString
		if err := rows.Scan(
			&n.ID, &n.TenantID, &userIDRaw, &n.Environment, &n.Type, &n.Category,
			&n.Title, &n.Message, &n.Read, &actionURL, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		if userIDRaw.Valid {
			n.UserID = &userIDRaw.String
		}
		if actionURL.Valid {
			n.ActionURL = &actionURL.String
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// CountUnread returns unread notification count for the user scope.
func (r *NotificationRepository) CountUnread(ctx context.Context, tenantID, userID, environment string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM studio_notifications
		WHERE tenant_id = $1
		  AND (user_id IS NULL OR user_id = $2)
		  AND COALESCE(environment, '') = $3
		  AND read = FALSE`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID, userID, environment).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

// CreateNotification inserts a new notification.
func (r *NotificationRepository) CreateNotification(ctx context.Context, params CreateNotificationParams) (*StudioNotification, error) {
	id := uuid.New().String()
	if params.Type == "" {
		params.Type = "info"
	}
	if params.Category == "" {
		params.Category = "system"
	}

	query := `
		INSERT INTO studio_notifications (
			id, tenant_id, user_id, environment, type, category, title, message, read, action_url, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, FALSE, $9, NOW())
		RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, query,
		id, params.TenantID, params.UserID, params.Environment,
		params.Type, params.Category, params.Title, params.Message, params.ActionURL,
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return &StudioNotification{
		ID:          id,
		TenantID:    params.TenantID,
		UserID:      params.UserID,
		Environment: params.Environment,
		Type:        params.Type,
		Category:    params.Category,
		Title:       params.Title,
		Message:     params.Message,
		Read:        false,
		ActionURL:   params.ActionURL,
		CreatedAt:   createdAt,
	}, nil
}

// MarkRead marks a single notification as read.
func (r *NotificationRepository) MarkRead(ctx context.Context, tenantID, userID, environment, notificationID string) (*StudioNotification, error) {
	query := `
		UPDATE studio_notifications
		SET read = TRUE
		WHERE tenant_id = $1 AND id = $2
		  AND (user_id IS NULL OR user_id = $3)
		  AND COALESCE(environment, '') = $4
		RETURNING id, tenant_id, user_id, environment, type, category, title, message, read, action_url, created_at`
	return r.scanNotification(r.db.QueryRowContext(ctx, query, tenantID, notificationID, userID, environment))
}

// MarkAllRead marks all notifications as read for the user scope.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, tenantID, userID, environment string) (int64, error) {
	query := `
		UPDATE studio_notifications
		SET read = TRUE
		WHERE tenant_id = $1
		  AND (user_id IS NULL OR user_id = $2)
		  AND COALESCE(environment, '') = $3
		  AND read = FALSE`
	res, err := r.db.ExecContext(ctx, query, tenantID, userID, environment)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications read: %w", err)
	}
	return res.RowsAffected()
}

// DeleteNotification removes a notification.
func (r *NotificationRepository) DeleteNotification(ctx context.Context, tenantID, userID, environment, notificationID string) error {
	query := `
		DELETE FROM studio_notifications
		WHERE tenant_id = $1 AND id = $2
		  AND (user_id IS NULL OR user_id = $3)
		  AND COALESCE(environment, '') = $4`
	res, err := r.db.ExecContext(ctx, query, tenantID, notificationID, userID, environment)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteAllNotifications removes all notifications for the user scope.
func (r *NotificationRepository) DeleteAllNotifications(ctx context.Context, tenantID, userID, environment string) (int64, error) {
	query := `
		DELETE FROM studio_notifications
		WHERE tenant_id = $1
		  AND (user_id IS NULL OR user_id = $2)
		  AND COALESCE(environment, '') = $3`
	res, err := r.db.ExecContext(ctx, query, tenantID, userID, environment)
	if err != nil {
		return 0, fmt.Errorf("delete all notifications: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return n, nil
}

func (r *NotificationRepository) scanNotification(row *sql.Row) (*StudioNotification, error) {
	var n StudioNotification
	var userIDRaw sql.NullString
	var actionURL sql.NullString
	err := row.Scan(
		&n.ID, &n.TenantID, &userIDRaw, &n.Environment, &n.Type, &n.Category,
		&n.Title, &n.Message, &n.Read, &actionURL, &n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan notification: %w", err)
	}
	if userIDRaw.Valid {
		n.UserID = &userIDRaw.String
	}
	if actionURL.Valid {
		n.ActionURL = &actionURL.String
	}
	return &n, nil
}
