package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthEventRepository handles authentication event logging operations
type AuthEventRepository struct {
	db *PostgresDB
}

// NewAuthEventRepository creates a new auth event repository
func NewAuthEventRepository(db *PostgresDB) *AuthEventRepository {
	return &AuthEventRepository{db: db}
}

// LogAuthEvent logs an authentication event
func (r *AuthEventRepository) LogAuthEvent(ctx context.Context, event *AuthEvent) error {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()

	var locationInfo, metadata, securityFlags []byte
	if event.LocationInfo != nil {
		locationInfo, _ = json.Marshal(event.LocationInfo)
	}
	if event.Metadata != nil {
		metadata, _ = json.Marshal(event.Metadata)
	}
	if event.SecurityFlags != nil {
		securityFlags, _ = json.Marshal(event.SecurityFlags)
	}

	query := `
		INSERT INTO auth_events (id, user_id, tenant_id, event_type, success, failure_reason,
			ip_address, user_agent, location_info, session_id, provider, metadata, security_flags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, query, event.ID, event.UserID, event.TenantID, event.EventType,
		event.Success, event.FailureReason, event.IPAddress, event.UserAgent,
		locationInfo, event.SessionID, event.Provider, metadata, securityFlags, event.CreatedAt)

	return err
}

// GetAuthEventsForUser retrieves auth events for a specific user
func (r *AuthEventRepository) GetAuthEventsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuthEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, event_type, success, failure_reason,
			ip_address, user_agent, location_info, session_id, provider, metadata, security_flags, created_at
		FROM auth_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query auth events for user: %w", err)
	}
	defer rows.Close()

	return r.scanAuthEvents(rows)
}

// GetAuthEventsByType retrieves auth events of a specific type
func (r *AuthEventRepository) GetAuthEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*AuthEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, event_type, success, failure_reason,
			ip_address, user_agent, location_info, session_id, provider, metadata, security_flags, created_at
		FROM auth_events
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		eventType, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query auth events by type: %w", err)
	}
	defer rows.Close()

	return r.scanAuthEvents(rows)
}

// GetRecentAuthEvents retrieves recent auth events across all users
func (r *AuthEventRepository) GetRecentAuthEvents(ctx context.Context, since time.Time, limit int) ([]*AuthEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, event_type, success, failure_reason,
			ip_address, user_agent, location_info, session_id, provider, metadata, security_flags, created_at
		FROM auth_events
		WHERE created_at >= $1
		ORDER BY created_at DESC
		LIMIT $2`,
		since, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query recent auth events: %w", err)
	}
	defer rows.Close()

	return r.scanAuthEvents(rows)
}

// DeleteOldAuthEvents removes auth events older than the specified time
func (r *AuthEventRepository) DeleteOldAuthEvents(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM auth_events
		WHERE created_at < $1`,
		before)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old auth events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// scanAuthEvents scans database rows into AuthEvent structs
func (r *AuthEventRepository) scanAuthEvents(rows *sql.Rows) ([]*AuthEvent, error) {
	var events []*AuthEvent

	for rows.Next() {
		var event AuthEvent
		var locationInfo, metadata, securityFlags []byte
		var failureReason sql.NullString
		var sessionID, userID, tenantID sql.NullString
		var provider sql.NullString

		err := rows.Scan(
			&event.ID, &userID, &tenantID, &event.EventType, &event.Success, &failureReason,
			&event.IPAddress, &event.UserAgent, &locationInfo, &sessionID, &provider,
			&metadata, &securityFlags, &event.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan auth event: %w", err)
		}

		// Handle nullable fields
		if userID.Valid {
			if uid, err := uuid.Parse(userID.String); err == nil {
				event.UserID = &uid
			}
		}
		if tenantID.Valid {
			if tid, err := uuid.Parse(tenantID.String); err == nil {
				event.TenantID = &tid
			}
		}
		if failureReason.Valid {
			event.FailureReason = &failureReason.String
		}
		if sessionID.Valid {
			if sid, err := uuid.Parse(sessionID.String); err == nil {
				event.SessionID = &sid
			}
		}
		if provider.Valid {
			event.Provider = &provider.String
		}

		// Unmarshal JSON fields
		if len(locationInfo) > 0 {
			json.Unmarshal(locationInfo, &event.LocationInfo)
		}
		if len(metadata) > 0 {
			json.Unmarshal(metadata, &event.Metadata)
		}
		if len(securityFlags) > 0 {
			json.Unmarshal(securityFlags, &event.SecurityFlags)
		}

		events = append(events, &event)
	}

	return events, nil
}