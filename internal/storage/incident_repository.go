package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
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
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
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