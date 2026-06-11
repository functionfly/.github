package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for notification data access
type Repository interface {
	// Notifications
	CreateNotification(ctx context.Context, n *Notification) error
	GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error)
	CountNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) (int, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUnreadCountsByCategory(ctx context.Context, userID uuid.UUID) (map[string]int, error)
	GetTotalCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, id uuid.UUID) error
	ArchiveNotification(ctx context.Context, id uuid.UUID) error
	UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
	CheckDuplicate(ctx context.Context, userID uuid.UUID, notificationType, category string, data JSONMap, within time.Duration) (*Notification, error)

	// Preferences
	GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)
	GetPreference(ctx context.Context, userID uuid.UUID, channel, category string) (*NotificationPreference, error)
	SavePreference(ctx context.Context, p *NotificationPreference) error
	CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error

	// Templates
	GetTemplate(ctx context.Context, notificationType, channel string) (*NotificationTemplate, error)
	ListTemplates(ctx context.Context) ([]*NotificationTemplate, error)
	SaveTemplate(ctx context.Context, t *NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// Analytics
	TrackAnalytics(ctx context.Context, a *NotificationAnalytics) error
	GetAnalytics(ctx context.Context, notificationID uuid.UUID) ([]*NotificationAnalytics, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// notificationDataJSON returns a JSON text for the data column (valid UTF-8 for ::jsonb).
func notificationDataJSON(data JSONMap) (string, error) {
	if len(data) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(map[string]interface{}(data))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CreateNotification creates a new notification
func (r *PostgresRepository) CreateNotification(ctx context.Context, n *Notification) error {
	dataStr, err := notificationDataJSON(n.Data)
	if err != nil {
		return fmt.Errorf("notification data: %w", err)
	}
	// Explicit ::jsonb and ::text[] so PostgreSQL does not interpret a text[] literal like {"in_app"}
	// as json (invalid JSON → 22P02) during unknown-literal coercion under pgx simple protocol.
	query := `
		INSERT INTO notifications (user_id, type, category, title, body, data, channels, priority, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::text[], $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		n.UserID, n.Type, n.Category, n.Title, n.Body,
		dataStr,
		[]string(n.Channels),
		n.Priority, n.Status, n.ExpiresAt,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
}

// GetNotification retrieves a notification by ID
func (r *PostgresRepository) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
	query := `
		SELECT id, user_id, type, category, title, body, data, channels, priority, status, read_at, sent_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`
	n := &Notification{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Category, &n.Title, &n.Body, &n.Data, &n.Channels,
		&n.Priority, &n.Status, &n.ReadAt, &n.SentAt, &n.ExpiresAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}
	return n, err
}

// ListNotifications lists notifications for a user with options
func (r *PostgresRepository) ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error) {
	query := `
		SELECT id, user_id, type, category, title, body, data, channels, priority, status, read_at, sent_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if opts.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, opts.Status)
	}
	if opts.Category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, opts.Category)
	}
	if opts.UnreadOnly {
		query += " AND status NOT IN ('read', 'archived')"
	}

	query += " ORDER BY created_at DESC"

	if opts.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, opts.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Category, &n.Title, &n.Body, &n.Data, &n.Channels,
			&n.Priority, &n.Status, &n.ReadAt, &n.SentAt, &n.ExpiresAt, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// CountNotifications returns the total number of notifications matching the filters
func (r *PostgresRepository) CountNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) (int, error) {
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	if opts.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, opts.Status)
	}
	if opts.Category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, opts.Category)
	}
	if opts.UnreadOnly {
		query += " AND status NOT IN ('read', 'archived')"
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// GetUnreadCount returns the number of unread notifications for a user
func (r *PostgresRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND status NOT IN ('read', 'archived')
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// GetUnreadCountsByCategory returns unread notification counts grouped by category
func (r *PostgresRepository) GetUnreadCountsByCategory(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT category, COUNT(*) as count
		FROM notifications
		WHERE user_id = $1 AND status NOT IN ('read', 'archived')
		GROUP BY category
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		counts[category] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

// GetTotalCount returns the total number of notifications for a user (both read and unread)
func (r *PostgresRepository) GetTotalCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// MarkAsRead marks a notification as read
func (r *PostgresRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1, updated_at = $1
		WHERE id = $2
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (r *PostgresRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1, updated_at = $1
		WHERE user_id = $2 AND status NOT IN ('read', 'archived')
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, userID)
	return err
}

// DeleteNotification deletes a notification
func (r *PostgresRepository) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ArchiveNotification sets status to archived and clears from unread counts.
func (r *PostgresRepository) ArchiveNotification(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE notifications
		SET status = 'archived', read_at = $1, updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}

// UpdateNotificationStatus updates the status of a notification
func (r *PostgresRepository) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE notifications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, now, id)
	return err
}

// GetPreferences retrieves all preferences for a user
func (r *PostgresRepository) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	query := `
		SELECT id, user_id, channel, category, enabled, frequency, quiet_hours_start, quiet_hours_end, timezone, webhook_url, webhook_secret, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []*NotificationPreference
	for rows.Next() {
		p := &NotificationPreference{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Channel, &p.Category, &p.Enabled, &p.Frequency,
			&p.QuietHoursStart, &p.QuietHoursEnd, &p.Timezone, &p.WebhookURL, &p.WebhookSecret,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}

// GetPreference retrieves a specific preference for a user
func (r *PostgresRepository) GetPreference(ctx context.Context, userID uuid.UUID, channel, category string) (*NotificationPreference, error) {
	query := `
		SELECT id, user_id, channel, category, enabled, frequency, quiet_hours_start, quiet_hours_end, timezone, webhook_url, webhook_secret, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND channel = $2 AND category = $3
	`
	p := &NotificationPreference{}
	err := r.db.QueryRowContext(ctx, query, userID, channel, category).Scan(
		&p.ID, &p.UserID, &p.Channel, &p.Category, &p.Enabled, &p.Frequency,
		&p.QuietHoursStart, &p.QuietHoursEnd, &p.Timezone, &p.WebhookURL, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// SavePreference saves a notification preference
func (r *PostgresRepository) SavePreference(ctx context.Context, p *NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (user_id, channel, category, enabled, frequency, quiet_hours_start, quiet_hours_end, timezone, webhook_url, webhook_secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, channel, category) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			frequency = EXCLUDED.frequency,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			timezone = EXCLUDED.timezone,
			webhook_url = EXCLUDED.webhook_url,
			webhook_secret = EXCLUDED.webhook_secret,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		p.UserID, p.Channel, p.Category, p.Enabled, p.Frequency,
		p.QuietHoursStart, p.QuietHoursEnd, p.Timezone, p.WebhookURL, p.WebhookSecret,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

// CreateDefaultPreferences creates default preferences for a new user
func (r *PostgresRepository) CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error {
	categories := []string{CategorySystem, CategorySecurity, CategoryBilling, CategoryDeployment, CategoryFunction, CategoryTeam, CategoryMessages, CategoryFailover, CategoryProvider}
	channels := []string{ChannelEmail, ChannelInApp}

	for _, category := range categories {
		for _, channel := range channels {
			pref := &NotificationPreference{
				UserID:    userID,
				Channel:   channel,
				Category:  category,
				Enabled:   true,
				Frequency: FrequencyImmediate,
				Timezone:  "UTC",
			}
			if err := r.SavePreference(ctx, pref); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetTemplate retrieves a template by type and channel
func (r *PostgresRepository) GetTemplate(ctx context.Context, notificationType, channel string) (*NotificationTemplate, error) {
	query := `
		SELECT id, type, channel, subject, body_html, body_text, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE type = $1 AND channel = $2 AND is_active = true
	`
	t := &NotificationTemplate{}
	err := r.db.QueryRowContext(ctx, query, notificationType, channel).Scan(
		&t.ID, &t.Type, &t.Channel, &t.Subject, &t.BodyHTML, &t.BodyText, &t.Variables,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// ListTemplates lists all templates
func (r *PostgresRepository) ListTemplates(ctx context.Context) ([]*NotificationTemplate, error) {
	query := `
		SELECT id, type, channel, subject, body_html, body_text, variables, is_active, created_at, updated_at
		FROM notification_templates
		ORDER BY type, channel
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*NotificationTemplate
	for rows.Next() {
		t := &NotificationTemplate{}
		if err := rows.Scan(
			&t.ID, &t.Type, &t.Channel, &t.Subject, &t.BodyHTML, &t.BodyText, &t.Variables,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// SaveTemplate saves a notification template
func (r *PostgresRepository) SaveTemplate(ctx context.Context, t *NotificationTemplate) error {
	query := `
		INSERT INTO notification_templates (type, channel, subject, body_html, body_text, variables, is_active)
		VALUES ($1, $2, $3, $4, $5, $6::text[], $7)
		ON CONFLICT (type, channel) DO UPDATE SET
			subject = EXCLUDED.subject,
			body_html = EXCLUDED.body_html,
			body_text = EXCLUDED.body_text,
			variables = EXCLUDED.variables,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		t.Type, t.Channel, t.Subject, t.BodyHTML, t.BodyText, []string(t.Variables), t.IsActive,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

// DeleteTemplate deletes a template
func (r *PostgresRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_templates WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// TrackAnalytics tracks notification analytics
func (r *PostgresRepository) TrackAnalytics(ctx context.Context, a *NotificationAnalytics) error {
	query := `
		INSERT INTO notification_analytics (notification_id, channel, status, error_message, delivered_at, opened_at, clicked_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		a.NotificationID, a.Channel, a.Status, a.ErrorMessage, a.DeliveredAt,
		a.OpenedAt, a.ClickedAt, a.IPAddress, a.UserAgent,
	).Scan(&a.ID, &a.CreatedAt)
}

// GetAnalytics retrieves analytics for a notification
func (r *PostgresRepository) GetAnalytics(ctx context.Context, notificationID uuid.UUID) ([]*NotificationAnalytics, error) {
	query := `
		SELECT id, notification_id, channel, status, error_message, delivered_at, opened_at, clicked_at, ip_address, user_agent, created_at
		FROM notification_analytics
		WHERE notification_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, notificationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []*NotificationAnalytics
	for rows.Next() {
		a := &NotificationAnalytics{}
		if err := rows.Scan(
			&a.ID, &a.NotificationID, &a.Channel, &a.Status, &a.ErrorMessage,
			&a.DeliveredAt, &a.OpenedAt, &a.ClickedAt, &a.IPAddress, &a.UserAgent, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		analytics = append(analytics, a)
	}
	return analytics, rows.Err()
}

// DeleteExpired removes all expired notifications that are not already read.
// Returns the number of deleted notifications.
func (r *PostgresRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM notifications
		WHERE expires_at < $1 AND status NOT IN ('read', 'archived')
	`
	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CheckDuplicate checks if a similar notification was created within the time window.
// Returns the existing notification if found, nil otherwise.
// Deduplication is based on user_id, type, category, and data hash within the time window.
func (r *PostgresRepository) CheckDuplicate(ctx context.Context, userID uuid.UUID, notificationType, category string, data JSONMap, within time.Duration) (*Notification, error) {
	// Create a simple hash of the data for comparison
	dataHash := ""
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		dataHash = string(dataBytes)
	}

	query := `
		SELECT id, user_id, type, category, title, body, data, channels, priority, status, read_at, sent_at, expires_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
			AND type = $2
			AND category = $3
			AND created_at > $4
			AND status NOT IN ('read', 'archived')
		ORDER BY created_at DESC
		LIMIT 1
	`
	 cutoff := time.Now().Add(-within)
	n := &Notification{}
	err := r.db.QueryRowContext(ctx, query, userID, notificationType, category, cutoff).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Category, &n.Title, &n.Body, &n.Data, &n.Channels,
		&n.Priority, &n.Status, &n.ReadAt, &n.SentAt, &n.ExpiresAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Verify data matches (simple string comparison)
	if dataHash != "" {
		existingDataBytes, _ := json.Marshal(n.Data)
		if string(existingDataBytes) != dataHash {
			return nil, nil // Data doesn't match, not a duplicate
		}
	}

	return n, nil
}
