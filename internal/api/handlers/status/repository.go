package status

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RepositoryInterface defines the interface for status page data access
type RepositoryInterface interface {
	GetIncidentByID(ctx context.Context, id interface{}) (*Incident, error)
	ListIncidents(ctx context.Context, query ListIncidentsQuery) (*IncidentsListResponse, error)
	CreateIncident(ctx context.Context, req *CreateIncidentRequest, createdBy interface{}) (*Incident, error)
	UpdateIncident(ctx context.Context, id interface{}, req *UpdateIncidentRequest, updatedBy interface{}) (*Incident, error)
	GetActiveIncidents(ctx context.Context) ([]Incident, error)
	GetMaintenanceByID(ctx context.Context, id interface{}) (*MaintenanceWindow, error)
	ListMaintenance(ctx context.Context, query ListMaintenanceQuery) (*MaintenanceListResponse, error)
	CreateMaintenance(ctx context.Context, req *CreateMaintenanceRequest, createdBy interface{}) (*MaintenanceWindow, error)
	GetUpcomingMaintenance(ctx context.Context) ([]MaintenanceWindow, error)
	GetSystemHealthChecks(ctx context.Context) ([]Component, error)
	GetComponentHealthHistory(ctx context.Context, componentName string, since time.Time) ([]StatusHistoryPoint, error)
	CalculateComponentUptime(ctx context.Context, componentName string, duration time.Duration) (float64, error)
	GetLatestComponentResponseTime(ctx context.Context, componentName string) (int, error)
	GetProviderStatus(ctx context.Context) ([]ProviderStatus, error)
	GetProviderRegions(ctx context.Context, provider string) ([]RegionStatus, error)
	GetProviderBackends(ctx context.Context, provider, region string) ([]BackendStatus, error)
}

// Repository provides data access for status page functionality
type Repository struct {
	db           *sql.DB
	gormDB       interface{} // Used for GORM operations if needed
	incidentRepo *storage.IncidentRepository
}

// NewRepository creates a new status page repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:           db,
		incidentRepo: storage.NewIncidentRepository(db),
	}
}

// SetGormDB sets the GORM database for advanced queries
func (r *Repository) SetGormDB(gormDB interface{}) {
	r.gormDB = gormDB
}

// --- Incident Methods ---

// GetIncidentByID retrieves a single incident by ID with its updates
func (r *Repository) GetIncidentByID(ctx context.Context, id interface{}) (*Incident, error) {
	// Get incident
	var dbIncident DatabaseIncident
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
		FROM incidents WHERE id = $1`, id).Scan(
		&dbIncident.ID, &dbIncident.Title, &dbIncident.Severity, &dbIncident.Status,
		&dbIncident.Description, &dbIncident.CreatedAt, &dbIncident.ResolvedAt, &dbIncident.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("incident not found")
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	// Get incident updates
	incidentID, ok := id.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid incident ID type")
	}
	updates, err := r.getIncidentUpdates(ctx, incidentID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get incident updates")
	}

	return r.dbIncidentToIncident(&dbIncident, updates), nil
}

// ListIncidents retrieves incidents with filtering and pagination
func (r *Repository) ListIncidents(ctx context.Context, query ListIncidentsQuery) (*IncidentsListResponse, error) {
	// Build query
	whereClause := ""
	var args []interface{}
	argIndex := 1

	if query.Status != "" {
		if query.Status == "active" {
			whereClause += " WHERE status != 'resolved'"
		} else {
			whereClause += fmt.Sprintf(" WHERE status = $%d", argIndex)
			args = append(args, query.Status)
			argIndex++
		}
	}

	if query.Severity != "" {
		if whereClause == "" {
			whereClause = " WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += fmt.Sprintf(" severity = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.StartDate != nil {
		if whereClause == "" {
			whereClause = " WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += fmt.Sprintf(" created_at >= $%d", argIndex)
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		if whereClause == "" {
			whereClause = " WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += fmt.Sprintf(" created_at <= $%d", argIndex)
		args = append(args, *query.EndDate)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM incidents" + whereClause
	var total int
	if len(args) > 0 {
		err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("failed to count incidents: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("failed to count incidents: %w", err)
		}
	}

	// Set default limit
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	// Get incidents
	selectQuery := fmt.Sprintf(`
		SELECT id, title, severity, status, description, created_at, resolved_at, updated_at
		FROM incidents %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	finalArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var dbIncident DatabaseIncident
		err := rows.Scan(&dbIncident.ID, &dbIncident.Title, &dbIncident.Severity, &dbIncident.Status,
			&dbIncident.Description, &dbIncident.CreatedAt, &dbIncident.ResolvedAt, &dbIncident.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		// Get updates for each incident
		updates, _ := r.getIncidentUpdates(ctx, dbIncident.ID)
		incident := r.dbIncidentToIncident(&dbIncident, updates)
		incidents = append(incidents, *incident)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}

	return &IncidentsListResponse{
		Incidents: incidents,
		Pagination: Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+len(incidents) < total,
		},
	}, nil
}

// CreateIncident creates a new incident
func (r *Repository) CreateIncident(ctx context.Context, req *CreateIncidentRequest, createdBy interface{}) (*Incident, error) {
	createdByID, ok := createdBy.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid createdBy type")
	}

	id := uuid.New()
	now := time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO incidents (id, title, severity, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, req.Title, req.Severity, "investigating", req.Description, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	// Add initial update if provided
	if req.InitialUpdate != nil {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO incident_updates (id, incident_id, status, message, created_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), id, req.InitialUpdate.Status, req.InitialUpdate.Message, now, createdByID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to create initial incident update")
		}
	}

	return r.GetIncidentByID(ctx, id)
}

// UpdateIncident updates an incident and optionally adds a new update
func (r *Repository) UpdateIncident(ctx context.Context, id interface{}, req *UpdateIncidentRequest, updatedBy interface{}) (*Incident, error) {
	incidentID, ok := id.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid incident ID type")
	}

	updatedByID, ok := updatedBy.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid updatedBy type")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != "" {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, req.Title)
		argIndex++
	}

	if req.Severity != "" {
		setParts = append(setParts, fmt.Sprintf("severity = $%d", argIndex))
		args = append(args, req.Severity)
		argIndex++
	}

	if req.Status != "" {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++

		// If resolving, set resolved_at
		if req.Status == "resolved" {
			setParts = append(setParts, fmt.Sprintf("resolved_at = $%d", argIndex))
			args = append(args, time.Now())
			argIndex++
		}
	}

	if req.Description != "" {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	// Always update updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	if len(setParts) == 0 && req.NewUpdate == nil {
		return r.GetIncidentByID(ctx, incidentID)
	}

	// Execute update if there are fields to update
	if len(setParts) > 0 {
		query := fmt.Sprintf("UPDATE incidents SET %s WHERE id = $%d",
			joinStrings(setParts, ", "), argIndex)
		args = append(args, incidentID)

		_, err := r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update incident: %w", err)
		}
	}

	// Add new update if provided
	if req.NewUpdate != nil {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO incident_updates (id, incident_id, status, message, created_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), incidentID, req.NewUpdate.Status, req.NewUpdate.Message, time.Now(), updatedByID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to create incident update")
		}
	}

	return r.GetIncidentByID(ctx, id)
}

// getIncidentUpdates retrieves updates for an incident
func (r *Repository) getIncidentUpdates(ctx context.Context, incidentID uuid.UUID) ([]IncidentUpdate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name
		FROM incident_updates iu
		LEFT JOIN users u ON iu.created_by = u.id
		WHERE iu.incident_id = $1
		ORDER BY iu.created_at ASC`, incidentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []IncidentUpdate
	for rows.Next() {
		var update IncidentUpdate
		var createdBy *uuid.UUID
		var userName *string

		err := rows.Scan(&update.ID, &update.Status, &update.Message, &update.CreatedAt, &createdBy, &userName)
		if err != nil {
			continue
		}

		if createdBy != nil && userName != nil {
			update.CreatedBy = &UserRef{
				ID:   createdBy.String(),
				Name: *userName,
			}
		}

		updates = append(updates, update)
	}

	return updates, rows.Err()
}

// GetActiveIncidents retrieves currently active (non-resolved) incidents
func (r *Repository) GetActiveIncidents(ctx context.Context) ([]Incident, error) {
	query := ListIncidentsQuery{
		Status: "active",
		Limit:  10,
	}

	resp, err := r.ListIncidents(ctx, query)
	if err != nil {
		return nil, err
	}

	return resp.Incidents, nil
}

// dbIncidentToIncident converts a database incident to API model
func (r *Repository) dbIncidentToIncident(db *DatabaseIncident, updates []IncidentUpdate) *Incident {
	incident := &Incident{
		ID:          db.ID.String(),
		Title:       db.Title,
		Severity:    db.Severity,
		Status:      db.Status,
		Description: db.Description,
		CreatedAt:   db.CreatedAt,
		ResolvedAt:  db.ResolvedAt,
		UpdatedAt:   db.UpdatedAt,
		Updates:     updates,
	}

	// Calculate duration if resolved
	if db.ResolvedAt != nil {
		duration := int(db.ResolvedAt.Sub(db.CreatedAt).Minutes())
		incident.DurationMinutes = &duration
	}

	return incident
}

// --- Maintenance Methods ---

// GetMaintenanceByID retrieves a maintenance window by ID
func (r *Repository) GetMaintenanceByID(ctx context.Context, id interface{}) (*MaintenanceWindow, error) {
	maintenanceID, ok := id.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid maintenance ID type")
	}

	var dbMaint DatabaseMaintenance
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end,
		       status, affected_components, affected_providers, created_at, updated_at
		FROM maintenance_windows WHERE id = $1`, maintenanceID).Scan(
		&dbMaint.ID, &dbMaint.Title, &dbMaint.Description,
		&dbMaint.ScheduledStart, &dbMaint.ScheduledEnd, &dbMaint.ActualStart, &dbMaint.ActualEnd,
		&dbMaint.Status, &dbMaint.AffectedComponents, &dbMaint.AffectedProviders,
		&dbMaint.CreatedAt, &dbMaint.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("maintenance window not found")
		}
		return nil, fmt.Errorf("failed to get maintenance: %w", err)
	}

	return r.dbMaintenanceToMaintenance(&dbMaint), nil
}

// ListMaintenance retrieves maintenance windows with filtering
func (r *Repository) ListMaintenance(ctx context.Context, query ListMaintenanceQuery) (*MaintenanceListResponse, error) {
	whereClause := ""
	var args []interface{}
	argIndex := 1

	if query.Status != "" {
		whereClause = fmt.Sprintf("WHERE status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	} else if query.Upcoming {
		if whereClause == "" {
			whereClause = "WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += fmt.Sprintf(" status IN ('scheduled', 'in_progress') AND scheduled_end > $%d", argIndex)
		args = append(args, time.Now())
		argIndex++
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	queryStr := fmt.Sprintf(`
		SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end,
		       status, affected_components, affected_providers, created_at, updated_at
		FROM maintenance_windows
		%s
		ORDER BY scheduled_start DESC
		LIMIT $%d`, whereClause, argIndex)

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list maintenance: %w", err)
	}
	defer rows.Close()

	var windows []MaintenanceWindow
	for rows.Next() {
		var dbMaint DatabaseMaintenance
		err := rows.Scan(
			&dbMaint.ID, &dbMaint.Title, &dbMaint.Description,
			&dbMaint.ScheduledStart, &dbMaint.ScheduledEnd, &dbMaint.ActualStart, &dbMaint.ActualEnd,
			&dbMaint.Status, &dbMaint.AffectedComponents, &dbMaint.AffectedProviders,
			&dbMaint.CreatedAt, &dbMaint.UpdatedAt)
		if err != nil {
			continue
		}
		windows = append(windows, *r.dbMaintenanceToMaintenance(&dbMaint))
	}

	return &MaintenanceListResponse{MaintenanceWindows: windows}, rows.Err()
}

// CreateMaintenance creates a new maintenance window
func (r *Repository) CreateMaintenance(ctx context.Context, req *CreateMaintenanceRequest, createdBy interface{}) (*MaintenanceWindow, error) {
	createdByID, ok := createdBy.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("invalid createdBy type")
	}

	id := uuid.New()
	now := time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO maintenance_windows (id, title, description, scheduled_start, scheduled_end,
		                                  status, affected_components, affected_providers, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		id, req.Title, req.Description, req.ScheduledStart, req.ScheduledEnd,
		"scheduled", req.AffectedComponents, req.AffectedProviders, createdByID, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create maintenance: %w", err)
	}

	return r.GetMaintenanceByID(ctx, id)
}

// GetUpcomingMaintenance retrieves scheduled maintenance windows
func (r *Repository) GetUpcomingMaintenance(ctx context.Context) ([]MaintenanceWindow, error) {
	query := ListMaintenanceQuery{
		Upcoming: true,
		Limit:    10,
	}

	resp, err := r.ListMaintenance(ctx, query)
	if err != nil {
		return nil, err
	}

	return resp.MaintenanceWindows, nil
}

// dbMaintenanceToMaintenance converts a database maintenance to API model
func (r *Repository) dbMaintenanceToMaintenance(db *DatabaseMaintenance) *MaintenanceWindow {
	return &MaintenanceWindow{
		ID:                 db.ID.String(),
		Title:              db.Title,
		Description:        db.Description,
		Status:             db.Status,
		ScheduledStart:     db.ScheduledStart,
		ScheduledEnd:       db.ScheduledEnd,
		ActualStart:        db.ActualStart,
		ActualEnd:          db.ActualEnd,
		AffectedComponents: db.AffectedComponents,
		AffectedProviders:  db.AffectedProviders,
		CreatedAt:          db.CreatedAt,
		UpdatedAt:          db.UpdatedAt,
	}
}

// --- System Health Methods ---

// GetSystemHealthChecks retrieves the latest health checks for all components
func (r *Repository) GetSystemHealthChecks(ctx context.Context) ([]Component, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ON (component_name)
			id, check_type, component_name, status, response_time_ms, message, metadata, checked_at
		FROM system_health_checks
		ORDER BY component_name, checked_at DESC`)

	if err != nil {
		return nil, fmt.Errorf("failed to get system health checks: %w", err)
	}
	defer rows.Close()

	var components []Component
	for rows.Next() {
		var dbCheck DatabaseHealthCheck
		err := rows.Scan(
			&dbCheck.ID, &dbCheck.CheckType, &dbCheck.ComponentName, &dbCheck.Status,
			&dbCheck.ResponseTimeMs, &dbCheck.Message, &dbCheck.Metadata, &dbCheck.CheckedAt)
		if err != nil {
			continue
		}

		component := r.dbHealthCheckToComponent(&dbCheck)
		components = append(components, *component)
	}

	return components, rows.Err()
}

// GetComponentHealthHistory retrieves health history for a component
func (r *Repository) GetComponentHealthHistory(ctx context.Context, componentName string, since time.Time) ([]StatusHistoryPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT checked_at, status, response_time_ms
		FROM system_health_checks
		WHERE component_name = $1 AND checked_at >= $2
		ORDER BY checked_at ASC`,
		componentName, since)

	if err != nil {
		return nil, fmt.Errorf("failed to get health history: %w", err)
	}
	defer rows.Close()

	var history []StatusHistoryPoint
	for rows.Next() {
		var point StatusHistoryPoint
		err := rows.Scan(&point.Timestamp, &point.Status, &point.ResponseTimeMs)
		if err != nil {
			continue
		}
		history = append(history, point)
	}

	return history, rows.Err()
}

// CalculateComponentUptime calculates uptime percentage for a component over a duration
func (r *Repository) CalculateComponentUptime(ctx context.Context, componentName string, duration time.Duration) (float64, error) {
	since := time.Now().Add(-duration)

	history, err := r.GetComponentHealthHistory(ctx, componentName, since)
	if err != nil {
		return 0, fmt.Errorf("failed to get health history: %w", err)
	}

	if len(history) == 0 {
		return 100.0, nil // Assume 100% uptime if no data
	}

	// Sort history by timestamp (should already be sorted, but ensure it)
	for i := 0; i < len(history)-1; i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].Timestamp.After(history[j].Timestamp) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	totalDuration := duration.Seconds()
	healthyDuration := 0.0
	currentHealthy := false
	lastTimestamp := since

	for _, point := range history {
		// Calculate time spent in previous state
		timeInState := point.Timestamp.Sub(lastTimestamp).Seconds()

		if currentHealthy {
			healthyDuration += timeInState
		}

		// Update state based on current point
		currentHealthy = point.Status == "operational" || point.Status == "healthy"
		lastTimestamp = point.Timestamp
	}

	// Add remaining time until now
	timeInState := time.Now().Sub(lastTimestamp).Seconds()
	if currentHealthy {
		healthyDuration += timeInState
	}

	if totalDuration <= 0 {
		return 100.0, nil
	}

	uptimePercent := (healthyDuration / totalDuration) * 100.0

	// Clamp to reasonable bounds
	if uptimePercent < 0 {
		uptimePercent = 0
	}
	if uptimePercent > 100 {
		uptimePercent = 100
	}

	return uptimePercent, nil
}

// dbHealthCheckToComponent converts a database health check to API component model
func (r *Repository) dbHealthCheckToComponent(db *DatabaseHealthCheck) *Component {
	component := &Component{
		ID:          db.ID.String(),
		Name:        db.ComponentName,
		Type:        mapCheckTypeToComponentType(db.CheckType),
		Status:      mapHealthStatusToComponentStatus(db.Status),
		Description: db.Message,
		LastChecked: db.CheckedAt,
	}

	if db.ResponseTimeMs != nil {
		component.ResponseTime = *db.ResponseTimeMs
	}

	return component
}

// GetLatestComponentResponseTime retrieves the most recent response time for a component from health checks
func (r *Repository) GetLatestComponentResponseTime(ctx context.Context, componentName string) (int, error) {
	var responseTimeMs sql.NullInt32
	err := r.db.QueryRowContext(ctx, `
		SELECT response_time_ms
		FROM system_health_checks
		WHERE component_name = $1 AND response_time_ms IS NOT NULL
		ORDER BY checked_at DESC
		LIMIT 1`,
		componentName).Scan(&responseTimeMs)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No data yet, return 0
		}
		return 0, fmt.Errorf("failed to get latest response time: %w", err)
	}

	if responseTimeMs.Valid {
		return int(responseTimeMs.Int32), nil
	}
	return 0, nil
}

// --- Alerts Methods ---

// GetActiveAlerts retrieves active alerts
func (r *Repository) GetActiveAlerts(ctx context.Context) ([]storage.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.alert_type, a.severity, a.tenant_id, a.app_id, a.backend_id,
		       a.title, a.message, a.status, a.resolved_at, a.created_at, a.updated_at,
		       b.provider, b.region, app.name as app_name
		FROM alerts a
		LEFT JOIN backends b ON a.backend_id = b.id
		LEFT JOIN apps app ON a.app_id = app.id
		WHERE a.status = 'active'
		ORDER BY
			CASE a.severity
				WHEN 'critical' THEN 1
				WHEN 'error' THEN 2
				WHEN 'warning' THEN 3
				ELSE 4
			END,
			a.created_at DESC
		LIMIT 50`)

	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []storage.Alert
	for rows.Next() {
		var alert storage.Alert
		var provider, region, appName sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID, &alert.AlertType, &alert.Severity, &alert.TenantID, &alert.AppID, &alert.BackendID,
			&alert.Title, &alert.Message, &alert.Status, &resolvedAt, &alert.CreatedAt, &alert.UpdatedAt,
			&provider, &region, &appName)
		if err != nil {
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		// Store provider/region/app in metadata
		alert.Metadata = make(map[string]interface{})
		if provider.Valid {
			alert.Metadata["provider"] = provider.String
		}
		if region.Valid {
			alert.Metadata["region"] = region.String
		}
		if appName.Valid {
			alert.Metadata["app_name"] = appName.String
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// --- Provider Status Methods ---

// GetProviderStatus retrieves status for all providers
func (r *Repository) GetProviderStatus(ctx context.Context) ([]ProviderStatus, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			b.provider,
			COUNT(*) as total_backends,
			COUNT(*) FILTER (WHERE cs.state = 'closed') as healthy_backends,
			COUNT(*) FILTER (WHERE cs.state = 'half-open') as degraded_backends,
			COUNT(*) FILTER (WHERE cs.state = 'open') as unhealthy_backends
		FROM backends b
		LEFT JOIN circuit_state cs ON b.id = cs.backend_id
		WHERE b.enabled = true
		GROUP BY b.provider`)

	if err != nil {
		return nil, fmt.Errorf("failed to get provider status: %w", err)
	}
	defer rows.Close()

	providers := make(map[string]*ProviderStatus)
	for rows.Next() {
		var provider string
		var total, healthy, degraded, unhealthy int

		err := rows.Scan(&provider, &total, &healthy, &degraded, &unhealthy)
		if err != nil {
			continue
		}

		providers[provider] = &ProviderStatus{
			Name:          provider,
			DisplayName:   formatProviderName(provider),
			OverallStatus: calculateProviderStatus(healthy, degraded, unhealthy),
			Summary: ProviderSummary{
				TotalBackends:     total,
				HealthyBackends:   healthy,
				DegradedBackends:  degraded,
				UnhealthyBackends: unhealthy,
			},
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Get region details for each provider
	for provider := range providers {
		regions, err := r.GetProviderRegions(ctx, provider)
		if err != nil {
			logrus.WithError(err).WithField("provider", provider).Warn("Failed to get provider regions")
			continue
		}
		providers[provider].Regions = regions
	}

	// Convert map to slice
	result := make([]ProviderStatus, 0, len(providers))
	for _, ps := range providers {
		result = append(result, *ps)
	}

	return result, nil
}

// GetProviderRegions retrieves region status for a provider
func (r *Repository) GetProviderRegions(ctx context.Context, provider string) ([]RegionStatus, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			b.region,
			COUNT(*) as total_backends,
			COUNT(*) FILTER (WHERE hc.ok = true) as healthy_count,
			AVG(hc.latency_ms) FILTER (WHERE hc.ok = true) as avg_latency,
			COUNT(*) FILTER (WHERE hc.ok = false)::float / NULLIF(COUNT(*), 0) * 100 as error_rate
		FROM backends b
		LEFT JOIN LATERAL (
			SELECT * FROM health_checks
			WHERE backend_id = b.id
			ORDER BY timestamp DESC
			LIMIT 1
		) hc ON true
		WHERE b.provider = $1 AND b.enabled = true
		GROUP BY b.region`, provider)

	if err != nil {
		return nil, fmt.Errorf("failed to get provider regions: %w", err)
	}
	defer rows.Close()

	var regions []RegionStatus
	for rows.Next() {
		var region RegionStatus
		var total, healthy int
		var errorRate sql.NullFloat64

		err := rows.Scan(&region.Code, &total, &healthy, &region.LatencyMs, &errorRate)
		if err != nil {
			continue
		}

		region.Name = formatRegionName(region.Code)
		region.Status = calculateRegionStatus(healthy, total)
		if errorRate.Valid {
			region.ErrorRate = errorRate.Float64
		}

		// Calculate uptime (placeholder - would use actual historical data)
		if total > 0 {
			region.Uptime24h = float64(healthy) / float64(total) * 100
		}

		regions = append(regions, region)
	}

	return regions, rows.Err()
}

// GetProviderBackends retrieves backend details for a provider/region
func (r *Repository) GetProviderBackends(ctx context.Context, provider, region string) ([]BackendStatus, error) {
	query := `
		SELECT b.id, b.url, cs.state, cs.fail_count, cs.success_count,
		       cs.last_failure_ts, cs.last_success_ts,
		       hc.ok, hc.latency_ms, hc.status_code, hc.timestamp
		FROM backends b
		LEFT JOIN circuit_state cs ON b.id = cs.backend_id
		LEFT JOIN LATERAL (
			SELECT * FROM health_checks
			WHERE backend_id = b.id
			ORDER BY timestamp DESC
			LIMIT 1
		) hc ON true
		WHERE b.provider = $1 AND b.enabled = true`

	args := []interface{}{provider}
	if region != "" && region != "all" {
		query += ` AND b.region = $2`
		args = append(args, region)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider backends: %w", err)
	}
	defer rows.Close()

	var backends []BackendStatus
	for rows.Next() {
		var bs BackendStatus
		var circuitState string
		var failCount, successCount int
		var lastFailure, lastSuccess sql.NullTime
		var healthy sql.NullBool
		var latency sql.NullInt64
		var statusCode sql.NullInt64
		var lastCheck sql.NullTime

		err := rows.Scan(
			&bs.ID, &bs.URL, &circuitState, &failCount, &successCount,
			&lastFailure, &lastSuccess, &healthy, &latency, &statusCode, &lastCheck)
		if err != nil {
			continue
		}

		bs.CircuitState = circuitState
		if healthy.Valid {
			if healthy.Bool {
				bs.Status = "healthy"
			} else {
				bs.Status = "unhealthy"
			}
		} else {
			bs.Status = "unknown"
		}

		if latency.Valid {
			bs.LatencyMs = int(latency.Int64)
		}

		if lastCheck.Valid {
			bs.LastCheck = lastCheck.Time
		}

		bs.ConsecutiveFailures = failCount

		backends = append(backends, bs)
	}

	return backends, rows.Err()
}

// --- Helper Functions ---

// mapCheckTypeToComponentType maps database check type to component type
func mapCheckTypeToComponentType(checkType string) string {
	switch checkType {
	case "database":
		return "database"
	case "api":
		return "api"
	case "cache":
		return "cache"
	case "external_service":
		return "provider"
	case "monitoring":
		return "monitoring"
	default:
		return "infrastructure"
	}
}

// mapHealthStatusToComponentStatus maps health status to component status
func mapHealthStatusToComponentStatus(status string) string {
	switch status {
	case "healthy":
		return "operational"
	case "degraded":
		return "degraded_performance"
	case "unhealthy":
		return "major_outage"
	default:
		return "operational"
	}
}

// formatProviderName formats a provider identifier for display
func formatProviderName(provider string) string {
	switch provider {
	case "workers":
		return "Cloudflare Workers"
	case "vercel":
		return "Vercel"
	case "fly":
		return "Fly.io"
	case "deno-deploy":
		return "Deno Deploy"
	case "functionfly-edge":
		return "FunctionFly Edge"
	default:
		// Capitalize first letter
		if len(provider) > 0 {
			return string(provider[0]-32) + provider[1:]
		}
		return provider
	}
}

// formatRegionName formats a region code for display
func formatRegionName(region string) string {
	// Map common region codes to display names
	regionNames := map[string]string{
		"us-east":      "US East",
		"us-west":      "US West",
		"eu-west":      "EU West",
		"eu-central":   "EU Central",
		"ap-south":     "Asia Pacific South",
		"ap-northeast": "Asia Pacific Northeast",
		"ap-southeast": "Asia Pacific Southeast",
		"sa-east":      "South America East",
	}

	if name, ok := regionNames[region]; ok {
		return name
	}
	return region
}

// calculateProviderStatus determines overall provider status from backend counts
func calculateProviderStatus(healthy, degraded, unhealthy int) string {
	total := healthy + degraded + unhealthy
	if total == 0 {
		return "operational"
	}

	unhealthyRatio := float64(unhealthy) / float64(total)
	degradedRatio := float64(degraded) / float64(total)

	if unhealthyRatio > 0.5 {
		return "outage"
	} else if unhealthyRatio > 0.1 || degradedRatio > 0.3 {
		return "degraded"
	}
	return "operational"
}

// calculateRegionStatus determines region status from health counts
func calculateRegionStatus(healthy, total int) string {
	if total == 0 {
		return "operational"
	}

	ratio := float64(healthy) / float64(total)
	if ratio < 0.5 {
		return "major_outage"
	} else if ratio < 0.8 {
		return "partial_outage"
	} else if ratio < 0.95 {
		return "degraded_performance"
	}
	return "operational"
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
