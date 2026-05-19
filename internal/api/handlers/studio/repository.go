package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CollabEvent represents a generic studio collaboration event stored in studio_collab_events.
type CollabEvent struct {
	ID         string                 `json:"id"`
	TenantID   string                 `json:"tenant_id"`
	EventType  string                 `json:"event_type"`
	CreatedBy  string                 `json:"created_by"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Environment string                `json:"environment"`
}

// CollabRepository handles CRUD for studio_collab_events.
type CollabRepository struct {
	db *sql.DB
}

// NewCollabRepository creates a new CollabRepository.
func NewCollabRepository(db *sql.DB) *CollabRepository {
	return &CollabRepository{db: db}
}

// ListEvents returns events filtered by tenant, optional event_type, and environment.
func (r *CollabRepository) ListEvents(ctx context.Context, tenantID, eventType, environment string, limit, offset int) ([]CollabEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var query string
	var args []interface{}
	if eventType != "" && environment != "" {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1 AND event_type = $2 AND COALESCE(environment, '') = $3
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5`
		args = []interface{}{tenantID, eventType, environment, limit, offset}
	} else if eventType != "" {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1 AND event_type = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`
		args = []interface{}{tenantID, eventType, limit, offset}
	} else if environment != "" {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1 AND COALESCE(environment, '') = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`
		args = []interface{}{tenantID, environment, limit, offset}
	} else {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{tenantID, limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list collab events: %w", err)
	}
	defer rows.Close()

	var events []CollabEvent
	for rows.Next() {
		var ev CollabEvent
		var metaRaw []byte
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.EventType, &ev.CreatedBy, &metaRaw, &ev.CreatedAt, &ev.UpdatedAt, &ev.Environment); err != nil {
			return nil, fmt.Errorf("scan collab event: %w", err)
		}
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &ev.Metadata)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// GetEvent returns a single event by ID scoped to tenant and environment.
func (r *CollabRepository) GetEvent(ctx context.Context, tenantID, eventID, environment string) (*CollabEvent, error) {
	var query string
	var args []interface{}
	if environment != "" {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1 AND id = $2 AND COALESCE(environment, '') = $3`
		args = []interface{}{tenantID, eventID, environment}
	} else {
		query = `
			SELECT id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')
			FROM studio_collab_events
			WHERE tenant_id = $1 AND id = $2`
		args = []interface{}{tenantID, eventID}
	}
	var ev CollabEvent
	var metaRaw []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&ev.ID, &ev.TenantID, &ev.EventType, &ev.CreatedBy, &metaRaw, &ev.CreatedAt, &ev.UpdatedAt, &ev.Environment,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get collab event: %w", err)
	}
	if len(metaRaw) > 0 {
		_ = json.Unmarshal(metaRaw, &ev.Metadata)
	}
	return &ev, nil
}

// CreateEvent inserts a new collab event.
func (r *CollabRepository) CreateEvent(ctx context.Context, tenantID, eventType, createdBy, environment string, metadata map[string]interface{}) (*CollabEvent, error) {
	id := uuid.New().String()
	metaRaw, _ := json.Marshal(metadata)

	query := `
		INSERT INTO studio_collab_events (id, tenant_id, event_type, created_by, metadata, environment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, tenantID, eventType, createdBy, metaRaw, environment).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("create collab event: %w", err)
	}

	return &CollabEvent{
		ID:         id,
		TenantID:   tenantID,
		EventType:  eventType,
		CreatedBy:  createdBy,
		Metadata:   metadata,
		Environment: environment,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// UpdateEvent updates metadata for an existing event scoped to tenant and environment.
func (r *CollabRepository) UpdateEvent(ctx context.Context, tenantID, eventID, environment string, metadata map[string]interface{}) (*CollabEvent, error) {
	metaRaw, _ := json.Marshal(metadata)
	var query string
	var args []interface{}
	if environment != "" {
		query = `
			UPDATE studio_collab_events
			SET metadata = $1, updated_at = NOW()
			WHERE tenant_id = $2 AND id = $3 AND COALESCE(environment, '') = $4
			RETURNING id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')`
		args = []interface{}{metaRaw, tenantID, eventID, environment}
	} else {
		query = `
			UPDATE studio_collab_events
			SET metadata = $1, updated_at = NOW()
			WHERE tenant_id = $2 AND id = $3
			RETURNING id, tenant_id, event_type, created_by, metadata, created_at, updated_at, COALESCE(environment, '')`
		args = []interface{}{metaRaw, tenantID, eventID}
	}
	var ev CollabEvent
	var metaRawOut []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&ev.ID, &ev.TenantID, &ev.EventType, &ev.CreatedBy, &metaRawOut, &ev.CreatedAt, &ev.UpdatedAt, &ev.Environment,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update collab event: %w", err)
	}
	if len(metaRawOut) > 0 {
		_ = json.Unmarshal(metaRawOut, &ev.Metadata)
	}
	return &ev, nil
}

// DeleteEvent removes an event scoped to tenant and environment.
func (r *CollabRepository) DeleteEvent(ctx context.Context, tenantID, eventID, environment string) error {
	var query string
	var args []interface{}
	if environment != "" {
		query = `DELETE FROM studio_collab_events WHERE tenant_id = $1 AND id = $2 AND COALESCE(environment, '') = $3`
		args = []interface{}{tenantID, eventID, environment}
	} else {
		query = `DELETE FROM studio_collab_events WHERE tenant_id = $1 AND id = $2`
		args = []interface{}{tenantID, eventID}
	}
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete collab event: %w", err)
	}
	return nil
}

// TelemetryMetric represents one point in a time-series telemetry feed.
type TelemetryMetric struct {
	Timestamp          time.Time `json:"timestamp"`
	Requests           int64     `json:"requests"`
	SuccessfulRequests int64     `json:"successfulRequests"`
	FailedRequests     int64     `json:"failedRequests"`
	AverageLatencyMs   float64   `json:"averageLatencyMs"`
	P50LatencyMs       float64   `json:"p50LatencyMs"`
	P95LatencyMs       float64   `json:"p95LatencyMs"`
	P99LatencyMs       float64   `json:"p99LatencyMs"`
	ErrorRate          float64   `json:"errorRate"`
	Throughput         float64   `json:"throughput"`
}

// GetTelemetryMetrics returns hourly telemetry metrics for a tenant over the last N hours.
func (r *CollabRepository) GetTelemetryMetrics(ctx context.Context, tenantID, environment string, hours int) ([]TelemetryMetric, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168
	}

	var query string
	if environment != "" {
		query = `
			WITH hours AS (
				SELECT generate_series(
					NOW() - INTERVAL '1 hour' * $2,
					NOW(),
					INTERVAL '1 hour'
				) AS hour
			)
			SELECT
				h.hour,
				COALESCE(COUNT(*)::bigint, 0) AS requests,
				COALESCE(COUNT(*) FILTER (WHERE fl.level != 'error')::bigint, 0) AS successful_requests,
				COALESCE(COUNT(*) FILTER (WHERE fl.level = 'error')::bigint, 0) AS failed_requests,
				COALESCE(AVG((fl.metadata->>'duration_ms')::double precision), 0) AS avg_latency_ms,
				COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p50_latency_ms,
				COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p95_latency_ms,
				COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p99_latency_ms
			FROM hours h
			LEFT JOIN function_logs fl ON fl.timestamp >= h.hour AND fl.timestamp < h.hour + INTERVAL '1 hour'
				INNER JOIN functions f ON f.id = fl.function_id AND f.tenant_id = $1 AND f.environment = $3
			GROUP BY h.hour
			ORDER BY h.hour`
	} else {
		query = `
			WITH hours AS (
				SELECT generate_series(
					NOW() - INTERVAL '1 hour' * $2,
					NOW(),
					INTERVAL '1 hour'
				) AS hour
			)
			SELECT
				h.hour,
				COALESCE(COUNT(*)::bigint, 0) AS requests,
				COALESCE(COUNT(*) FILTER (WHERE fl.level != 'error')::bigint, 0) AS successful_requests,
				COALESCE(COUNT(*) FILTER (WHERE fl.level = 'error')::bigint, 0) AS failed_requests,
				COALESCE(AVG((fl.metadata->>'duration_ms')::double precision), 0) AS avg_latency_ms,
				COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p50_latency_ms,
				COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p95_latency_ms,
				COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY (fl.metadata->>'duration_ms')::double precision), 0) AS p99_latency_ms
			FROM hours h
			LEFT JOIN function_logs fl ON fl.timestamp >= h.hour AND fl.timestamp < h.hour + INTERVAL '1 hour'
				INNER JOIN functions f ON f.id = fl.function_id AND f.tenant_id = $1
			GROUP BY h.hour
			ORDER BY h.hour`
	}

	var rows *sql.Rows
	var err error
	if environment != "" {
		rows, err = r.db.QueryContext(ctx, query, tenantID, hours, environment)
	} else {
		rows, err = r.db.QueryContext(ctx, query, tenantID, hours)
	}
	if err != nil {
		return nil, fmt.Errorf("get telemetry metrics: %w", err)
	}
	defer rows.Close()

	var metrics []TelemetryMetric
	for rows.Next() {
		var m TelemetryMetric
		var avgLat, p50, p95, p99 sql.NullFloat64
		if err := rows.Scan(&m.Timestamp, &m.Requests, &m.SuccessfulRequests, &m.FailedRequests, &avgLat, &p50, &p95, &p99); err != nil {
			return nil, fmt.Errorf("scan telemetry metric: %w", err)
		}
		if avgLat.Valid {
			m.AverageLatencyMs = avgLat.Float64
		}
		if p50.Valid {
			m.P50LatencyMs = p50.Float64
		}
		if p95.Valid {
			m.P95LatencyMs = p95.Float64
		}
		if p99.Valid {
			m.P99LatencyMs = p99.Float64
		}
		if m.Requests > 0 {
			m.ErrorRate = float64(m.FailedRequests) / float64(m.Requests)
		}
		m.Throughput = float64(m.Requests)
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}
