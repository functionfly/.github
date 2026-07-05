package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BackendRepository handles backend-related database operations
type BackendRepository struct {
	db *PostgresDB
}

// NewBackendRepository creates a new backend repository
func NewBackendRepository(db *PostgresDB) *BackendRepository {
	return &BackendRepository{db: db}
}

// CreateBackend creates a new backend
func (r *BackendRepository) CreateBackend(ctx context.Context, appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error) {
	backend := &Backend{
		ID:           uuid.New(),
		AppID:        appID,
		Provider:     provider,
		Region:       region,
		URL:          url,
		SharedSecret: sharedSecret,
		Enabled:      true,
		Priority:     priority,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO backends (id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		backend.ID, backend.AppID, backend.Provider, backend.Region,
		backend.URL, backend.SharedSecret, backend.Enabled, backend.Priority,
		backend.CreatedAt, backend.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	return backend, nil
}

// ListBackendsByAppID lists backends for an app
func (r *BackendRepository) ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*Backend, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
		FROM backends WHERE app_id = $1 ORDER BY created_at`, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}
	defer rows.Close()

	var backends []*Backend
	for rows.Next() {
		backend := &Backend{}
		err := rows.Scan(&backend.ID, &backend.AppID, &backend.Provider,
			&backend.Region, &backend.URL, &backend.SharedSecret,
			&backend.Enabled, &backend.Priority, &backend.CreatedAt, &backend.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backend: %w", err)
		}
		backends = append(backends, backend)
	}

	return backends, nil
}

// CountBackendsByTenant counts all backends for a tenant across all apps
func (r *BackendRepository) CountBackendsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM backends b
		JOIN apps a ON b.app_id = a.id
		WHERE a.tenant_id = $1`, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count backends: %w", err)
	}
	return count, nil
}

// GetBackendByID retrieves a backend by ID
func (r *BackendRepository) GetBackendByID(ctx context.Context, id uuid.UUID) (*Backend, error) {
	backend := &Backend{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, app_id, provider, region, url, shared_secret, enabled, created_at, updated_at
		FROM backends WHERE id = $1`, id).Scan(
		&backend.ID, &backend.AppID, &backend.Provider, &backend.Region,
		&backend.URL, &backend.SharedSecret, &backend.Enabled,
		&backend.CreatedAt, &backend.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get backend: %w", err)
	}

	return backend, nil
}

// GetAllEnabledBackends gets all enabled backends
func (r *BackendRepository) GetAllEnabledBackends(ctx context.Context) ([]*Backend, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
		FROM backends WHERE enabled = true ORDER BY COALESCE(priority, 999) ASC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all enabled backends: %w", err)
	}
	defer rows.Close()

	var backends []*Backend
	for rows.Next() {
		backend := &Backend{}
		err := rows.Scan(&backend.ID, &backend.AppID, &backend.Provider,
			&backend.Region, &backend.URL, &backend.SharedSecret,
			&backend.Enabled, &backend.Priority, &backend.CreatedAt, &backend.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backend: %w", err)
		}
		backends = append(backends, backend)
	}

	return backends, nil
}

// GetBackendStatusByAppID gets backend status for an app
func (r *BackendRepository) GetBackendStatusByAppID(ctx context.Context, appID uuid.UUID) ([]*BackendStatus, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			-- Backend data
			b.id, b.app_id, b.provider, b.region, b.url, b.shared_secret, b.enabled, b.created_at, b.updated_at,
			-- Circuit state data (nullable)
			cs.state, cs.since_ts, cs.fail_count, cs.success_count, cs.last_failure_ts,
			-- Latest health check data (nullable)
			hc.timestamp, hc.ok, hc.status_code, hc.latency_ms, hc.error_message
		FROM backends b
		LEFT JOIN circuit_state cs ON b.id = cs.backend_id
		LEFT JOIN LATERAL (
			SELECT timestamp, ok, status_code, latency_ms, error_message
			FROM health_checks
			WHERE backend_id = b.id
			ORDER BY timestamp DESC
			LIMIT 1
		) hc ON true
		WHERE b.app_id = $1 AND b.enabled = true
		ORDER BY b.created_at`, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend status: %w", err)
	}
	defer rows.Close()

	var statuses []*BackendStatus
	for rows.Next() {
		status := &BackendStatus{
			Backend: &Backend{},
		}

		var circuitState *string
		var sinceTs *time.Time
		var failCount *int
		var successCount *int
		var lastFailureTs *time.Time
		var hcTimestamp *time.Time
		var hcOK *bool
		var hcStatusCode *int
		var hcLatencyMs *int
		var hcErrorMessage *string

		err := rows.Scan(
			// Backend fields
			&status.Backend.ID, &status.Backend.AppID, &status.Backend.Provider,
			&status.Backend.Region, &status.Backend.URL, &status.Backend.SharedSecret,
			&status.Backend.Enabled, &status.Backend.CreatedAt, &status.Backend.UpdatedAt,
			// Circuit state fields
			&circuitState, &sinceTs, &failCount, &successCount, &lastFailureTs,
			// Health check fields
			&hcTimestamp, &hcOK, &hcStatusCode, &hcLatencyMs, &hcErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backend status: %w", err)
		}

		// Build circuit state if it exists
		if circuitState != nil {
			status.CircuitState = &CircuitState{
				BackendID:     status.Backend.ID,
				State:         *circuitState,
				SinceTs:       *sinceTs,
				FailCount:     *failCount,
				SuccessCount:  *successCount,
				LastFailureTs: lastFailureTs,
			}
		}

		// Build health check if it exists
		if hcTimestamp != nil {
			status.LatestHealthCheck = &HealthCheck{
				BackendID:    status.Backend.ID,
				Timestamp:    *hcTimestamp,
				OK:           *hcOK,
				StatusCode:   *hcStatusCode,
				LatencyMs:    *hcLatencyMs,
				ErrorMessage: *hcErrorMessage,
			}
		}

		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return statuses, nil
}

// InsertHealthCheck inserts a health check result
func (r *BackendRepository) InsertHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO health_checks (backend_id, ok, status_code, latency_ms, error_message)
		VALUES ($1, $2, $3, $4, $5)`,
		backendID, ok, statusCode, latencyMs, errorMessage)

	if err != nil {
		return fmt.Errorf("failed to insert health check: %w", err)
	}

	return nil
}

// GetRecentHealthChecks gets recent health checks for a backend
func (r *BackendRepository) GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, backend_id, timestamp, ok, status_code, latency_ms, error_message
		FROM health_checks
		WHERE backend_id = $1
		ORDER BY timestamp DESC
		LIMIT $2`, backendID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get health checks: %w", err)
	}
	defer rows.Close()

	var checks []*HealthCheck
	for rows.Next() {
		check := &HealthCheck{}
		err := rows.Scan(&check.ID, &check.BackendID, &check.Timestamp,
			&check.OK, &check.StatusCode, &check.LatencyMs, &check.ErrorMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health check: %w", err)
		}
		checks = append(checks, check)
	}

	return checks, nil
}

// DeleteHealthChecksBefore deletes health check records older than the given time.
// Returns the number of deleted rows.
func (r *BackendRepository) DeleteHealthChecksBefore(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM health_checks WHERE timestamp < $1`, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old health checks: %w", err)
	}
	return result.RowsAffected()
}

// GetCircuitState gets circuit state for a backend
func (r *BackendRepository) GetCircuitState(ctx context.Context, backendID uuid.UUID) (*CircuitState, error) {
	state := &CircuitState{}
	err := r.db.QueryRowContext(ctx, `
		SELECT backend_id, state, since_ts, fail_count, success_count, last_failure_ts, last_success_ts
		FROM circuit_state WHERE backend_id = $1`, backendID).Scan(
		&state.BackendID, &state.State, &state.SinceTs,
		&state.FailCount, &state.SuccessCount,
		&state.LastFailureTs, &state.LastSuccessTs)

	if err == sql.ErrNoRows {
		// Return default closed state if no record exists
		return &CircuitState{
			BackendID: backendID,
			State:     "closed",
			SinceTs:   time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit state: %w", err)
	}

	return state, nil
}

// UpdateCircuitState updates circuit state
func (r *BackendRepository) UpdateCircuitState(ctx context.Context, state *CircuitState) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE circuit_state SET
			state = $2, since_ts = $3, fail_count = $4, success_count = $5,
			last_failure_ts = $6, last_success_ts = $7
		WHERE backend_id = $1`,
		state.BackendID, state.State, state.SinceTs,
		state.FailCount, state.SuccessCount,
		state.LastFailureTs, state.LastSuccessTs)

	if err != nil {
		return fmt.Errorf("failed to update circuit state: %w", err)
	}

	return nil
}

// UpsertCircuitState upserts circuit state
func (r *BackendRepository) UpsertCircuitState(ctx context.Context, state *CircuitState) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO circuit_state (backend_id, state, since_ts, fail_count, success_count, last_failure_ts, last_success_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (backend_id) DO UPDATE SET
			state = EXCLUDED.state,
			since_ts = EXCLUDED.since_ts,
			fail_count = EXCLUDED.fail_count,
			success_count = EXCLUDED.success_count,
			last_failure_ts = EXCLUDED.last_failure_ts,
			last_success_ts = EXCLUDED.last_success_ts`,
		state.BackendID, state.State, state.SinceTs,
		state.FailCount, state.SuccessCount,
		state.LastFailureTs, state.LastSuccessTs)

	if err != nil {
		return fmt.Errorf("failed to upsert circuit state: %w", err)
	}

	return nil
}

// InsertRoutingEvent inserts a routing event
func (r *BackendRepository) InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO routing_events (app_id, backend_id, latency_ms, outcome, request_id)
		VALUES ($1, $2, $3, $4, $5)`,
		appID, backendID, latencyMs, outcome, requestID)

	if err != nil {
		return fmt.Errorf("failed to insert routing event: %w", err)
	}

	return nil
}

// GetRecentRoutingEvents retrieves recent routing events for error rate calculation
func (r *BackendRepository) GetRecentRoutingEvents(ctx context.Context, limit int, since time.Time) ([]*RoutingEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, backend_id, timestamp, latency_ms, outcome, request_id
		FROM routing_events
		WHERE timestamp >= $1
		ORDER BY timestamp DESC
		LIMIT $2`,
		since, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query routing events: %w", err)
	}
	defer rows.Close()

	var events []*RoutingEvent
	for rows.Next() {
		var event RoutingEvent
		err := rows.Scan(
			&event.ID,
			&event.AppID,
			&event.BackendID,
			&event.Timestamp,
			&event.LatencyMs,
			&event.Outcome,
			&event.RequestID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routing event: %w", err)
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routing events: %w", err)
	}

	return events, nil
}

// GetRecentRoutingEventsByBackend returns recent routing events for a specific backend.
func (r *BackendRepository) GetRecentRoutingEventsByBackend(ctx context.Context, backendID uuid.UUID, limit int) ([]*RoutingEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, backend_id, timestamp, latency_ms, outcome, request_id
		FROM routing_events
		WHERE backend_id = $1
		ORDER BY timestamp DESC
		LIMIT $2`,
		backendID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query routing events by backend: %w", err)
	}
	defer rows.Close()

	var events []*RoutingEvent
	for rows.Next() {
		var event RoutingEvent
		err := rows.Scan(
			&event.ID,
			&event.AppID,
			&event.BackendID,
			&event.Timestamp,
			&event.LatencyMs,
			&event.Outcome,
			&event.RequestID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routing event: %w", err)
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routing events: %w", err)
	}

	return events, nil
}

// ListAllBackends lists all backends (without shared_secret for security)
func (r *BackendRepository) ListAllBackends(ctx context.Context) ([]*Backend, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, provider, region, url, enabled, priority, created_at, updated_at
		FROM backends
		ORDER BY provider ASC, region ASC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list all backends: %w", err)
	}
	defer rows.Close()

	var backends []*Backend
	for rows.Next() {
		backend := &Backend{}
		err := rows.Scan(
			&backend.ID, &backend.AppID, &backend.Provider, &backend.Region,
			&backend.URL, &backend.Enabled, &backend.Priority,
			&backend.CreatedAt, &backend.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backend: %w", err)
		}
		backends = append(backends, backend)
	}

	return backends, nil
}

// UpdateBackendEnabled updates the enabled status of a backend
func (r *BackendRepository) UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE backends
		SET enabled = $1, updated_at = NOW()
		WHERE id = $2`,
		enabled, backendID)
	if err != nil {
		return fmt.Errorf("failed to update backend enabled status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("backend not found")
	}

	return nil
}

// DeleteBackend deletes a backend by ID
func (r *BackendRepository) DeleteBackend(ctx context.Context, backendID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM backends WHERE id = $1`, backendID)
	if err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("backend not found")
	}

	return nil
}