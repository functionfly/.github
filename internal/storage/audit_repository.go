package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditRepository handles audit logging operations
type AuditRepository struct {
	db *PostgresDB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *PostgresDB) *AuditRepository {
	return &AuditRepository{db: db}
}

// LogAuditEvent logs an audit event
func (r *AuditRepository) LogAuditEvent(ctx context.Context, event *AuditEvent) error {
	event.ID = uuid.New()
	event.Timestamp = time.Now()

	var beforeState, afterState []byte
	if event.BeforeState != nil {
		beforeState, _ = json.Marshal(event.BeforeState)
	}
	if event.AfterState != nil {
		afterState, _ = json.Marshal(event.AfterState)
	}

	query := `
		INSERT INTO audit_events (id, actor_user_id, actor_email, tenant_id, action, resource_type, resource_id, request_id, before_state, after_state, ip_address, user_agent, timestamp, success)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.Exec(query, event.ID, event.ActorUserID, event.ActorEmail, event.TenantID,
		event.Action, event.ResourceType, event.ResourceID, event.RequestID,
		beforeState, afterState, event.IPAddress, event.UserAgent, event.Timestamp, event.Success)

	return err
}

// ListAuditEvents lists audit events with pagination
func (r *AuditRepository) ListAuditEvents(limit, offset int) ([]*AuditEvent, error) {
	return r.ListAuditEventsFiltered(limit, offset, nil)
}

// ListAuditEventsFiltered lists audit events with filters
func (r *AuditRepository) ListAuditEventsFiltered(limit, offset int, filters map[string]interface{}) ([]*AuditEvent, error) {
	baseQuery := `
		SELECT id, actor_user_id, actor_email, tenant_id, action, resource_type, resource_id,
			   request_id, before_state, after_state, ip_address, user_agent, timestamp, success
		FROM audit_events WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Add filters
	if actorUserID, ok := filters["actor_user_id"].(uuid.UUID); ok {
		baseQuery += fmt.Sprintf(" AND actor_user_id = $%d", argIndex)
		args = append(args, actorUserID)
		argIndex++
	}

	if actorEmail, ok := filters["actor_email"].(string); ok {
		baseQuery += fmt.Sprintf(" AND actor_email ILIKE $%d", argIndex)
		args = append(args, "%"+actorEmail+"%")
		argIndex++
	}

	if tenantID, ok := filters["tenant_id"].(uuid.UUID); ok {
		baseQuery += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	if action, ok := filters["action"].(string); ok {
		baseQuery += fmt.Sprintf(" AND action ILIKE $%d", argIndex)
		args = append(args, "%"+action+"%")
		argIndex++
	}

	if resourceType, ok := filters["resource_type"].(string); ok {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, resourceType)
		argIndex++
	}

	if resourceID, ok := filters["resource_id"].(uuid.UUID); ok {
		baseQuery += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, resourceID)
		argIndex++
	}

	if success, ok := filters["success"].(bool); ok {
		baseQuery += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, success)
		argIndex++
	}

	if startTime, ok := filters["start_time"].(time.Time); ok {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, startTime)
		argIndex++
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, endTime)
		argIndex++
	}

	// Add ordering and pagination
	baseQuery += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	var events []*AuditEvent
	for rows.Next() {
		event := &AuditEvent{}
		var beforeState, afterState []byte
		err := rows.Scan(
			&event.ID, &event.ActorUserID, &event.ActorEmail, &event.TenantID,
			&event.Action, &event.ResourceType, &event.ResourceID, &event.RequestID,
			&beforeState, &afterState, &event.IPAddress, &event.UserAgent,
			&event.Timestamp, &event.Success)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Parse JSON states
		if len(beforeState) > 0 {
			json.Unmarshal(beforeState, &event.BeforeState)
		}
		if len(afterState) > 0 {
			json.Unmarshal(afterState, &event.AfterState)
		}

		events = append(events, event)
	}

	return events, nil
}