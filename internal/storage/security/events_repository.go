package security

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EventRepository handles persistent security event storage.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new security event repository.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// SecurityEvent represents a persisted security event.
type SecurityEvent struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	FunctionID *string       `json:"function_id,omitempty"`
	EventType string         `json:"event_type"`
	Severity  string         `json:"severity"`
	Source    string         `json:"source"`
	Details   map[string]any `json:"details,omitempty"`
	IPAddress *string        `json:"ip_address,omitempty"`
	UserAgent *string        `json:"user_agent,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// InsertEvent persists a security event to the database.
func (r *EventRepository) InsertEvent(ctx context.Context, event *SecurityEvent) error {
	event.ID = uuid.New().String()

	detailsJSON, _ := json.Marshal(event.Details)

	var funcID *string
	if event.FunctionID != nil && *event.FunctionID != "" {
		funcID = event.FunctionID
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO security_events
			(id, tenant_id, function_id, event_type, severity, source, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, event.ID, event.TenantID, funcID, event.EventType, event.Severity,
		event.Source, detailsJSON, event.IPAddress, event.UserAgent)

	return err
}

// GetEventsByTenant returns security events for a tenant within a time window.
func (r *EventRepository) GetEventsByTenant(ctx context.Context, tenantID string, since time.Duration, limit int) ([]*SecurityEvent, error) {
	cutoff := time.Now().Add(-since)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, function_id, event_type, severity, source, details,
			ip_address, user_agent, created_at
		FROM security_events
		WHERE tenant_id = $1 AND created_at > $2
		ORDER BY created_at DESC
		LIMIT $3
	`, tenantID, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("query security events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*SecurityEvent
	for rows.Next() {
		e := &SecurityEvent{}
		var detailsJSON sql.NullString
		var funcID, ipAddr, userAgent sql.NullString
		if err := rows.Scan(&e.ID, &e.TenantID, &funcID, &e.EventType, &e.Severity,
			&e.Source, &detailsJSON, &ipAddr, &userAgent, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan security event: %w", err)
		}
		if funcID.Valid {
			e.FunctionID = &funcID.String
		}
		if ipAddr.Valid {
			e.IPAddress = &ipAddr.String
		}
		if userAgent.Valid {
			e.UserAgent = &userAgent.String
		}
		if detailsJSON.Valid {
			if err := json.Unmarshal([]byte(detailsJSON.String), &e.Details); err != nil {
				logrus.WithError(err).Warn("failed to unmarshal event details")
			}
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetEventsByFunction returns security events for a function within a time window.
func (r *EventRepository) GetEventsByFunction(ctx context.Context, functionID string, since time.Duration, limit int) ([]*SecurityEvent, error) {
	cutoff := time.Now().Add(-since)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, function_id, event_type, severity, source, details,
			ip_address, user_agent, created_at
		FROM security_events
		WHERE function_id = $1 AND created_at > $2
		ORDER BY created_at DESC
		LIMIT $3
	`, functionID, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("query security events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*SecurityEvent
	for rows.Next() {
		e := &SecurityEvent{}
		var detailsJSON sql.NullString
		var funcID, ipAddr, userAgent sql.NullString
		if err := rows.Scan(&e.ID, &e.TenantID, &funcID, &e.EventType, &e.Severity,
			&e.Source, &detailsJSON, &ipAddr, &userAgent, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan security event: %w", err)
		}
		if funcID.Valid {
			e.FunctionID = &funcID.String
		}
		if ipAddr.Valid {
			e.IPAddress = &ipAddr.String
		}
		if userAgent.Valid {
			e.UserAgent = &userAgent.String
		}
		if detailsJSON.Valid {
			if err := json.Unmarshal([]byte(detailsJSON.String), &e.Details); err != nil {
				logrus.WithError(err).Warn("failed to unmarshal event details")
			}
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// CountCriticalEvents counts critical events for a function within a time window.
func (r *EventRepository) CountCriticalEvents(ctx context.Context, functionID string, since time.Duration) (int, error) {
	cutoff := time.Now().Add(-since)

	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM security_events
		WHERE function_id = $1 AND severity = 'critical' AND created_at > $2
	`, functionID, cutoff).Scan(&count)

	return count, err
}

// CleanupOldEvents deletes events older than the retention period.
func (r *EventRepository) CleanupOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM security_events WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
