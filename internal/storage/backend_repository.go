package storage

import (
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
func (r *BackendRepository) CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error) {
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

	_, err := r.db.Exec(`
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
func (r *BackendRepository) ListBackendsByAppID(appID uuid.UUID) ([]*Backend, error) {
	rows, err := r.db.Query(`
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

// GetBackendByID retrieves a backend by ID
func (r *BackendRepository) GetBackendByID(id uuid.UUID) (*Backend, error) {
	backend := &Backend{}
	err := r.db.QueryRow(`
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
func (r *BackendRepository) GetAllEnabledBackends() ([]*Backend, error) {
	rows, err := r.db.Query(`
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
func (r *BackendRepository) GetBackendStatusByAppID(appID uuid.UUID) ([]*BackendStatus, error) {
	rows, err := r.db.Query(`
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
func (r *BackendRepository) InsertHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error {
	_, err := r.db.Exec(`
		INSERT INTO health_checks (backend_id, ok, status_code, latency_ms, error_message)
		VALUES ($1, $2, $3, $4, $5)`,
		backendID, ok, statusCode, latencyMs, errorMessage)

	if err != nil {
		return fmt.Errorf("failed to insert health check: %w", err)
	}

	return nil
}

// GetRecentHealthChecks gets recent health checks for a backend
func (r *BackendRepository) GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
	rows, err := r.db.Query(`
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

// GetCircuitState gets circuit state for a backend
func (r *BackendRepository) GetCircuitState(backendID uuid.UUID) (*CircuitState, error) {
	state := &CircuitState{}
	err := r.db.QueryRow(`
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
func (r *BackendRepository) UpdateCircuitState(state *CircuitState) error {
	_, err := r.db.Exec(`
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
func (r *BackendRepository) UpsertCircuitState(state *CircuitState) error {
	_, err := r.db.Exec(`
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
func (r *BackendRepository) InsertRoutingEvent(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	_, err := r.db.Exec(`
		INSERT INTO routing_events (app_id, backend_id, latency_ms, outcome, request_id)
		VALUES ($1, $2, $3, $4, $5)`,
		appID, backendID, latencyMs, outcome, requestID)

	if err != nil {
		return fmt.Errorf("failed to insert routing event: %w", err)
	}

	return nil
}

// GetRecentRoutingEvents retrieves recent routing events for error rate calculation
func (r *BackendRepository) GetRecentRoutingEvents(limit int, since time.Time) ([]*RoutingEvent, error) {
	rows, err := r.db.Query(`
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