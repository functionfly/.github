package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// IncidentRepository handles incident-related database operations
type IncidentRepository struct {
	db *sql.DB
}

// NewIncidentRepository creates a new incident repository
func NewIncidentRepository(db *sql.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

// CreateIncident creates a new incident
func (r *IncidentRepository) CreateIncident(ctx context.Context, incident *Incident) (*Incident, error) {
	if incident.ID == uuid.Nil {
		incident.ID = uuid.New()
	}
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = time.Now()
	}
	if incident.UpdatedAt.IsZero() {
		incident.UpdatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO incidents (id, title, severity, status, description, created_at, resolved_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		incident.ID, incident.Title, incident.Severity, incident.Status, incident.Description,
		incident.CreatedAt, incident.ResolvedAt, incident.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return incident, nil
}

// GetIncidentByID retrieves an incident by ID
func (r *IncidentRepository) GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	var incident Incident
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
		FROM incidents WHERE id = $1`, incidentID).Scan(
		&incident.ID, &incident.Title, &incident.Severity, &incident.Status,
		&incident.Description, &incident.CreatedAt, &incident.ResolvedAt, &incident.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("incident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

// ListIncidents retrieves incidents with optional filtering
func (r *IncidentRepository) ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*Incident, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
				 FROM incidents WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
				 FROM incidents ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var incident Incident
		err := rows.Scan(&incident.ID, &incident.Title, &incident.Severity, &incident.Status,
			&incident.Description, &incident.CreatedAt, &incident.ResolvedAt, &incident.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}

	return incidents, nil
}

// UpdateIncident updates an incident
func (r *IncidentRepository) UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error) {
	// Get current incident
	incident, err := r.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(key), argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(setParts) == 0 {
		return incident, nil
	}

	// Always update updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf("UPDATE incidents SET %s WHERE id = $%d",
		fmt.Sprintf("%s", setParts), argIndex)
	args = append(args, incidentID)

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update incident: %w", err)
	}

	// Return updated incident
	return r.GetIncidentByID(ctx, incidentID)
}

// ResolveIncident marks an incident as resolved
func (r *IncidentRepository) ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "resolved",
		"resolved_at": &now,
		"updated_at":  now,
	}

	return r.UpdateIncident(ctx, incidentID, updates)
}

// ListIncidentsSince returns incidents created on or after the given time, newest first.
func (r *IncidentRepository) ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*Incident, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
		FROM incidents WHERE created_at >= $1 ORDER BY created_at DESC LIMIT $2`,
		since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents since: %w", err)
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var incident Incident
		if err := rows.Scan(&incident.ID, &incident.Title, &incident.Severity, &incident.Status,
			&incident.Description, &incident.CreatedAt, &incident.ResolvedAt, &incident.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}
	return incidents, nil
}

// CountIncidentsSince returns the number of incidents created on or after the given time.
func (r *IncidentRepository) CountIncidentsSince(ctx context.Context, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM incidents WHERE created_at >= $1`, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count incidents since: %w", err)
	}
	return count, nil
}

// DailyIncidentCount is the number of incidents that occurred on a single calendar day (UTC).
type DailyIncidentCount struct {
	Date  string // YYYY-MM-DD
	Count int
}

// CountIncidentsGroupedByDay returns incident counts per day for the period starting at since (inclusive).
// Days with zero incidents are not included; callers can merge with a full date range to get zeros.
func (r *IncidentRepository) CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]DailyIncidentCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT (created_at AT TIME ZONE 'UTC')::date::text AS day, COUNT(*)
		FROM incidents
		WHERE created_at >= $1
		GROUP BY (created_at AT TIME ZONE 'UTC')::date
		ORDER BY day`,
		since)
	if err != nil {
		return nil, fmt.Errorf("failed to count incidents by day: %w", err)
	}
	defer rows.Close()

	var result []DailyIncidentCount
	for rows.Next() {
		var d DailyIncidentCount
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, fmt.Errorf("failed to scan day count: %w", err)
		}
		result = append(result, d)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incident day counts: %w", err)
	}
	return result, nil
}
