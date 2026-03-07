package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthAuditLog represents an authentication audit log entry
type AuthAuditLog struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	TenantID      *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	UserID        *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	EventType     string                 `json:"event_type" db:"event_type"`
	EventData     map[string]interface{} `json:"event_data,omitempty" db:"event_data"`
	IPAddress     string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     string                 `json:"user_agent,omitempty" db:"user_agent"`
	Success       bool                   `json:"success" db:"success"`
	FailureReason string                 `json:"failure_reason,omitempty" db:"failure_reason"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// AuthAuditRepository handles authentication audit log operations
type AuthAuditRepository struct {
	db *PostgresDB
}

// NewAuthAuditRepository creates a new auth audit repository
func NewAuthAuditRepository(db *PostgresDB) *AuthAuditRepository {
	return &AuthAuditRepository{db: db}
}

// Create inserts a new audit log entry
func (r *AuthAuditRepository) Create(ctx context.Context, log *AuthAuditLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	eventData, err := json.Marshal(log.EventData)
	if err != nil {
		eventData = []byte("{}")
	}

	query := `
		INSERT INTO auth_audit_log (id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.UserID, log.EventType, eventData,
		log.IPAddress, log.UserAgent, log.Success, log.FailureReason, log.CreatedAt)

	return err
}

// GetByTenant retrieves audit logs for a specific tenant
func (r *AuthAuditRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*AuthAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at
		FROM auth_audit_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs for tenant: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// GetByTenantSince retrieves audit logs for a specific tenant since the given time
func (r *AuthAuditRepository) GetByTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]*AuthAuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at
		FROM auth_audit_log
		WHERE tenant_id = $1 AND created_at > $2
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs since: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// GetByUser retrieves audit logs for a specific user
func (r *AuthAuditRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuthAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at
		FROM auth_audit_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs for user: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// GetByEventType retrieves audit logs of a specific event type
func (r *AuthAuditRepository) GetByEventType(ctx context.Context, eventType string, limit, offset int) ([]*AuthAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `
		SELECT id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at
		FROM auth_audit_log
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by event type: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// List retrieves audit logs with optional filtering
func (r *AuthAuditRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*AuthAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	baseQuery := `
		SELECT id, tenant_id, user_id, event_type, event_data, ip_address, user_agent, success, failure_reason, created_at
		FROM auth_audit_log WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Apply filters
	if tenantID, ok := filters["tenant_id"].(uuid.UUID); ok {
		baseQuery += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if eventType, ok := filters["event_type"].(string); ok {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if success, ok := filters["success"].(bool); ok {
		baseQuery += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, success)
		argIndex++
	}

	if startTime, ok := filters["start_time"].(time.Time); ok {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startTime)
		argIndex++
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endTime)
		argIndex++
	}

	// Add ordering and pagination
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// DeleteOld removes audit logs older than the specified time
func (r *AuthAuditRepository) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM auth_audit_log
		WHERE created_at < $1`, before)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// scanAuditLogs scans database rows into AuthAuditLog structs
func (r *AuthAuditRepository) scanAuditLogs(rows *sql.Rows) ([]*AuthAuditLog, error) {
	var logs []*AuthAuditLog

	for rows.Next() {
		var log AuthAuditLog
		var eventData []byte
		var tenantID, userID sql.NullString
		var failureReason sql.NullString

		err := rows.Scan(
			&log.ID, &tenantID, &userID, &log.EventType, &eventData,
			&log.IPAddress, &log.UserAgent, &log.Success, &failureReason, &log.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Handle nullable fields
		if tenantID.Valid {
			if tid, err := uuid.Parse(tenantID.String); err == nil {
				log.TenantID = &tid
			}
		}
		if userID.Valid {
			if uid, err := uuid.Parse(userID.String); err == nil {
				log.UserID = &uid
			}
		}
		if failureReason.Valid {
			log.FailureReason = failureReason.String
		}

		// Unmarshal JSON event data
		if len(eventData) > 0 {
			json.Unmarshal(eventData, &log.EventData)
		} else {
			log.EventData = make(map[string]interface{})
		}

		logs = append(logs, &log)
	}

	return logs, nil
}
