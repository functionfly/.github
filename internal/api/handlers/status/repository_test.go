package status

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a new sqlmock database connection for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, mock, cleanup
}

// mustParseTime parses a time string or fails the test
func mustParseTime(t *testing.T, s string) time.Time {
	parsed, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return parsed
}

func TestNewRepository(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
	assert.NotNil(t, repo.incidentRepo)
}

func TestRepository_GetIncidentByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	incidentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now()

	t.Run("existing incident", func(t *testing.T) {
		// Setup expectations for incident query
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, "Test Incident", "high", "investigating", "Test description", now, nil, now))

		// Setup expectations for incident updates query
		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu LEFT JOIN users u ON iu.created_by = u.id WHERE iu.incident_id = \$1 ORDER BY iu.created_at ASC`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.GetIncidentByID(ctx, incidentID)

		require.NoError(t, err)
		assert.NotNil(t, incident)
		assert.Equal(t, incidentID.String(), incident.ID)
		assert.Equal(t, "Test Incident", incident.Title)
		assert.Equal(t, "high", incident.Severity)
		assert.Equal(t, "investigating", incident.Status)
		assert.Equal(t, "Test description", incident.Description)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("incident not found", func(t *testing.T) {
		notFoundID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(notFoundID).
			WillReturnError(sql.ErrNoRows)

		incident, err := repo.GetIncidentByID(ctx, notFoundID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incident not found")
		assert.Nil(t, incident)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		errorID := uuid.MustParse("770e8400-e29b-41d4-a716-446655440002")

		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(errorID).
			WillReturnError(sql.ErrConnDone)

		incident, err := repo.GetIncidentByID(ctx, errorID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get incident")
		assert.Nil(t, incident)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_ListIncidents(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	t.Run("list all incidents", func(t *testing.T) {
		// Count query
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Select query
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(uuid.New(), "Incident 1", "critical", "investigating", "Desc 1", now, nil, now).
				AddRow(uuid.New(), "Incident 2", "high", "resolved", "Desc 2", now, now, now))

		// Expect updates queries for each incident
		// Expect updates query twice (once for each incident)
		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))
		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		result, err := repo.ListIncidents(ctx, ListIncidentsQuery{})

		require.NoError(t, err)
		assert.Len(t, result.Incidents, 2)
		assert.Equal(t, 2, result.Pagination.Total)
		assert.False(t, result.Pagination.HasMore)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by status active", func(t *testing.T) {
		// Count query with filter
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents WHERE status != 'resolved'`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Select query with filter
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE status != 'resolved' ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(uuid.New(), "Active Incident", "high", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		result, err := repo.ListIncidents(ctx, ListIncidentsQuery{Status: "active"})

		require.NoError(t, err)
		assert.Len(t, result.Incidents, 1)
		assert.Equal(t, "investigating", result.Incidents[0].Status)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by severity", func(t *testing.T) {
		// Count query
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents WHERE severity = \$1`).
			WithArgs("critical").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Select query
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE severity = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs("critical", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(uuid.New(), "Critical Incident", "critical", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		result, err := repo.ListIncidents(ctx, ListIncidentsQuery{Severity: "critical"})

		require.NoError(t, err)
		assert.Len(t, result.Incidents, 1)
		assert.Equal(t, "critical", result.Incidents[0].Severity)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("pagination", func(t *testing.T) {
		// Count query
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

		// Select query with limit and offset
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(10, 10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(uuid.New(), "Incident", "high", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		result, err := repo.ListIncidents(ctx, ListIncidentsQuery{Limit: 10, Offset: 10})

		require.NoError(t, err)
		assert.Len(t, result.Incidents, 1)
		assert.Equal(t, 25, result.Pagination.Total)
		assert.Equal(t, 10, result.Pagination.Limit)
		assert.Equal(t, 10, result.Pagination.Offset)
		assert.True(t, result.Pagination.HasMore) // 10 + 1 < 25

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by date range", func(t *testing.T) {
		startDate := mustParseTime(t, "2024-01-01T00:00:00Z")
		endDate := mustParseTime(t, "2024-01-31T23:59:59Z")

		// Count query with date range
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents WHERE created_at >= \$1 AND created_at <= \$2`).
			WithArgs(startDate, endDate).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		// Select query
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE created_at >= \$1 AND created_at <= \$2 ORDER BY created_at DESC LIMIT \$3 OFFSET \$4`).
			WithArgs(startDate, endDate, 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(uuid.New(), "Incident", "high", "investigating", "Desc", startDate.Add(24*time.Hour), nil, startDate.Add(24*time.Hour)))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		result, err := repo.ListIncidents(ctx, ListIncidentsQuery{
			StartDate: &startDate,
			EndDate:   &endDate,
		})

		require.NoError(t, err)
		assert.Len(t, result.Incidents, 1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_CreateIncident(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	createdBy := uuid.MustParse("880e8400-e29b-41d4-a716-446655440003")
	incidentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now()

	t.Run("create incident without initial update", func(t *testing.T) {
		req := &CreateIncidentRequest{
			Title:       "New Incident",
			Severity:    "high",
			Description: "Test description",
		}

		// Insert incident
		mock.ExpectExec(`INSERT INTO incidents \(id, title, severity, status, description, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)`).
			WithArgs(sqlmock.AnyArg(), req.Title, req.Severity, "investigating", req.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Get incident by ID (for return)
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, req.Title, req.Severity, "investigating", req.Description, now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.CreateIncident(ctx, req, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, incident)
		assert.Equal(t, req.Title, incident.Title)
		assert.Equal(t, req.Severity, incident.Severity)
		assert.Equal(t, "investigating", incident.Status)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create incident with initial update", func(t *testing.T) {
		req := &CreateIncidentRequest{
			Title:       "New Incident with Update",
			Severity:    "critical",
			Description: "Test description",
			InitialUpdate: &InitialUpdate{
				Message: "Initial investigation started",
				Status:  "investigating",
			},
		}

		// Insert incident
		mock.ExpectExec(`INSERT INTO incidents \(id, title, severity, status, description, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)`).
			WithArgs(sqlmock.AnyArg(), req.Title, req.Severity, "investigating", req.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Insert initial update
		mock.ExpectExec(`INSERT INTO incident_updates \(id, incident_id, status, message, created_at, created_by\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), req.InitialUpdate.Status, req.InitialUpdate.Message, sqlmock.AnyArg(), createdBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Get incident by ID
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, req.Title, req.Severity, "investigating", req.Description, now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.CreateIncident(ctx, req, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_UpdateIncident(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	incidentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	updatedBy := uuid.MustParse("880e8400-e29b-41d4-a716-446655440003")
	now := time.Now()

	t.Run("update incident title", func(t *testing.T) {
		req := &UpdateIncidentRequest{
			Title: "Updated Title",
		}

		// Update query
		mock.ExpectExec(`UPDATE incidents SET title = \$1, updated_at = \$2 WHERE id = \$3`).
			WithArgs(req.Title, sqlmock.AnyArg(), incidentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Get updated incident
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, req.Title, "high", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.UpdateIncident(ctx, incidentID, req, updatedBy)

		require.NoError(t, err)
		assert.Equal(t, req.Title, incident.Title)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("resolve incident", func(t *testing.T) {
		req := &UpdateIncidentRequest{
			Status: "resolved",
		}

		// Update query with resolved_at
		mock.ExpectExec(`UPDATE incidents SET status = \$1, resolved_at = \$2, updated_at = \$3 WHERE id = \$4`).
			WithArgs(req.Status, sqlmock.AnyArg(), sqlmock.AnyArg(), incidentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Get updated incident
		resolvedAt := now
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, "Title", "high", "resolved", "Desc", now.Add(-1*time.Hour), &resolvedAt, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.UpdateIncident(ctx, incidentID, req, updatedBy)

		require.NoError(t, err)
		assert.Equal(t, "resolved", incident.Status)
		assert.NotNil(t, incident.ResolvedAt)
		assert.NotNil(t, incident.DurationMinutes)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("add update to incident", func(t *testing.T) {
		req := &UpdateIncidentRequest{
			NewUpdate: &IncidentUpdateRequest{
				Message: "We are monitoring the situation",
				Status:  "monitoring",
			},
		}

		// Insert new update
		mock.ExpectExec(`INSERT INTO incident_updates \(id, incident_id, status, message, created_at, created_by\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
			WithArgs(sqlmock.AnyArg(), incidentID, req.NewUpdate.Status, req.NewUpdate.Message, sqlmock.AnyArg(), updatedBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Get updated incident
		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, "Title", "high", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}).
				AddRow(uuid.New(), "monitoring", req.NewUpdate.Message, now, &updatedBy, "Test User"))

		incident, err := repo.UpdateIncident(ctx, incidentID, req, updatedBy)

		require.NoError(t, err)
		assert.Len(t, incident.Updates, 1)
		assert.Equal(t, req.NewUpdate.Message, incident.Updates[0].Message)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no fields to update returns current incident", func(t *testing.T) {
		req := &UpdateIncidentRequest{}

		mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE id = \$1`).
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
				AddRow(incidentID, "Title", "high", "investigating", "Desc", now, nil, now))

		mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

		incident, err := repo.UpdateIncident(ctx, incidentID, req, updatedBy)

		require.NoError(t, err)
		assert.NotNil(t, incident)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetActiveIncidents(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM incidents WHERE status != 'resolved'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Select query
	mock.ExpectQuery(`SELECT id, title, severity, status, description, created_at, resolved_at, updated_at FROM incidents WHERE status != 'resolved' ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "severity", "status", "description", "created_at", "resolved_at", "updated_at"}).
			AddRow(uuid.New(), "Active 1", "critical", "investigating", "Desc", now, nil, now).
			AddRow(uuid.New(), "Active 2", "high", "identified", "Desc", now, nil, now))

	// Expect updates query twice (once for each incident)
	mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))
	mock.ExpectQuery(`SELECT iu.id, iu.status, iu.message, iu.created_at, iu.created_by, u.name as user_name FROM incident_updates iu`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "message", "created_at", "created_by", "user_name"}))

	incidents, err := repo.GetActiveIncidents(ctx)

	require.NoError(t, err)
	assert.Len(t, incidents, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetMaintenanceByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	maintenanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	t.Run("existing maintenance", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE id = \$1`).
			WithArgs(maintenanceID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
				AddRow(maintenanceID, "Maintenance Window", "System update", startTime, endTime, nil, nil, "scheduled", []string{"api"}, []string{"fly"}, now, now))

		maintenance, err := repo.GetMaintenanceByID(ctx, maintenanceID)

		require.NoError(t, err)
		assert.NotNil(t, maintenance)
		assert.Equal(t, "Maintenance Window", maintenance.Title)
		assert.Equal(t, "scheduled", maintenance.Status)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("maintenance not found", func(t *testing.T) {
		notFoundID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

		mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE id = \$1`).
			WithArgs(notFoundID).
			WillReturnError(sql.ErrNoRows)

		maintenance, err := repo.GetMaintenanceByID(ctx, notFoundID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maintenance window not found")
		assert.Nil(t, maintenance)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_ListMaintenance(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("list all maintenance", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows ORDER BY scheduled_start DESC LIMIT \$1`).
			WithArgs(20).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
				AddRow(uuid.New(), "Maintenance 1", "Desc 1", now.Add(24*time.Hour), now.Add(26*time.Hour), nil, nil, "scheduled", []string{"api"}, []string{"fly"}, now, now).
				AddRow(uuid.New(), "Maintenance 2", "Desc 2", now.Add(48*time.Hour), now.Add(50*time.Hour), nil, nil, "scheduled", []string{"db"}, []string{"vercel"}, now, now))

		result, err := repo.ListMaintenance(ctx, ListMaintenanceQuery{})

		require.NoError(t, err)
		assert.Len(t, result.MaintenanceWindows, 2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by status", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE status = \$1 ORDER BY scheduled_start DESC LIMIT \$2`).
			WithArgs("in_progress", 20).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
				AddRow(uuid.New(), "Ongoing Maintenance", "Desc", now.Add(-1*time.Hour), now.Add(time.Hour), now.Add(-1*time.Hour), nil, "in_progress", []string{"api"}, []string{"fly"}, now, now))

		result, err := repo.ListMaintenance(ctx, ListMaintenanceQuery{Status: "in_progress"})

		require.NoError(t, err)
		assert.Len(t, result.MaintenanceWindows, 1)
		assert.Equal(t, "in_progress", result.MaintenanceWindows[0].Status)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("upcoming maintenance", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE status IN \('scheduled', 'in_progress'\) AND scheduled_end > \$1 ORDER BY scheduled_start DESC LIMIT \$2`).
			WithArgs(sqlmock.AnyArg(), 10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
				AddRow(uuid.New(), "Upcoming", "Desc", now.Add(24*time.Hour), now.Add(26*time.Hour), nil, nil, "scheduled", []string{"api"}, []string{"fly"}, now, now))

		result, err := repo.ListMaintenance(ctx, ListMaintenanceQuery{Upcoming: true, Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result.MaintenanceWindows, 1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_CreateMaintenance(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	createdBy := uuid.MustParse("880e8400-e29b-41d4-a716-446655440003")
	maintenanceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	req := &CreateMaintenanceRequest{
		Title:              "New Maintenance",
		Description:        "System update",
		ScheduledStart:     startTime,
		ScheduledEnd:       endTime,
		AffectedComponents: []string{"api", "database"},
		AffectedProviders:  []string{"fly"},
	}

	// Insert maintenance
	mock.ExpectExec(`INSERT INTO maintenance_windows \(id, title, description, scheduled_start, scheduled_end, status, affected_components, affected_providers, created_by, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11\)`).
		WithArgs(sqlmock.AnyArg(), req.Title, req.Description, req.ScheduledStart, req.ScheduledEnd, "scheduled", req.AffectedComponents, req.AffectedProviders, createdBy, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get maintenance by ID
	mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
			AddRow(maintenanceID, req.Title, req.Description, req.ScheduledStart, req.ScheduledEnd, nil, nil, "scheduled", req.AffectedComponents, req.AffectedProviders, now, now))

	maintenance, err := repo.CreateMaintenance(ctx, req, createdBy)

	require.NoError(t, err)
	assert.NotNil(t, maintenance)
	assert.Equal(t, req.Title, maintenance.Title)
	assert.Equal(t, "scheduled", maintenance.Status)
	assert.Equal(t, req.AffectedComponents, maintenance.AffectedComponents)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetUpcomingMaintenance(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, title, description, scheduled_start, scheduled_end, actual_start, actual_end, status, affected_components, affected_providers, created_at, updated_at FROM maintenance_windows WHERE status IN \('scheduled', 'in_progress'\) AND scheduled_end > \$1 ORDER BY scheduled_start DESC LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "scheduled_start", "scheduled_end", "actual_start", "actual_end", "status", "affected_components", "affected_providers", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Upcoming", "Desc", now.Add(24*time.Hour), now.Add(26*time.Hour), nil, nil, "scheduled", []string{"api"}, []string{"fly"}, now, now))

	maintenance, err := repo.GetUpcomingMaintenance(ctx)

	require.NoError(t, err)
	assert.Len(t, maintenance, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSystemHealthChecks(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()
	responseTime := 45

	mock.ExpectQuery(`SELECT DISTINCT ON \(component_name\) id, check_type, component_name, status, response_time_ms, message, metadata, checked_at FROM system_health_checks ORDER BY component_name, checked_at DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "check_type", "component_name", "status", "response_time_ms", "message", "metadata", "checked_at"}).
			AddRow(uuid.New(), "api", "API Gateway", "healthy", &responseTime, "All systems operational", []byte("{}"), now).
			AddRow(uuid.New(), "database", "PostgreSQL", "healthy", &responseTime, "Database healthy", []byte("{}"), now).
			AddRow(uuid.New(), "cache", "Redis", "degraded", &responseTime, "High memory usage", []byte("{}"), now))

	components, err := repo.GetSystemHealthChecks(ctx)

	require.NoError(t, err)
	assert.Len(t, components, 3)

	// Check component mapping
	assert.Equal(t, "API Gateway", components[0].Name)
	assert.Equal(t, "api", components[0].Type)
	assert.Equal(t, "operational", components[0].Status)

	assert.Equal(t, "PostgreSQL", components[1].Name)
	assert.Equal(t, "database", components[1].Type)

	assert.Equal(t, "Redis", components[2].Name)
	assert.Equal(t, "cache", components[2].Type)
	assert.Equal(t, "degraded_performance", components[2].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetComponentHealthHistory(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()
	since := now.Add(-24 * time.Hour)

	mock.ExpectQuery(`SELECT checked_at, status, response_time_ms FROM system_health_checks WHERE component_name = \$1 AND checked_at >= \$2 ORDER BY checked_at ASC`).
		WithArgs("API Gateway", since).
		WillReturnRows(sqlmock.NewRows([]string{"checked_at", "status", "response_time_ms"}).
			AddRow(now.Add(-6*time.Hour), "healthy", 45).
			AddRow(now.Add(-4*time.Hour), "healthy", 42).
			AddRow(now.Add(-2*time.Hour), "degraded", 120).
			AddRow(now, "healthy", 38))

	history, err := repo.GetComponentHealthHistory(ctx, "API Gateway", since)

	require.NoError(t, err)
	assert.Len(t, history, 4)
	assert.Equal(t, "healthy", history[0].Status)
	assert.Equal(t, 45, history[0].ResponseTimeMs)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetProviderStatus(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	// First query: provider summary
	mock.ExpectQuery(`SELECT b.provider, COUNT\(\*\) as total_backends, COUNT\(\*\) FILTER \(WHERE cs.state = 'closed'\) as healthy_backends, COUNT\(\*\) FILTER \(WHERE cs.state = 'half-open'\) as degraded_backends, COUNT\(\*\) FILTER \(WHERE cs.state = 'open'\) as unhealthy_backends FROM backends b LEFT JOIN circuit_state cs ON b.id = cs.backend_id WHERE b.enabled = true GROUP BY b.provider`).
		WillReturnRows(sqlmock.NewRows([]string{"provider", "total", "healthy", "degraded", "unhealthy"}).
			AddRow("fly", 10, 8, 1, 1).
			AddRow("vercel", 5, 5, 0, 0))

	// Second query: region details for fly
	mock.ExpectQuery(`SELECT b.region, COUNT\(\*\) as total_backends, COUNT\(\*\) FILTER \(WHERE hc.ok = true\) as healthy_count, AVG\(hc.latency_ms\) FILTER \(WHERE hc.ok = true\) as avg_latency, COUNT\(\*\) FILTER \(WHERE hc.ok = false\)::float \/ NULLIF\(COUNT\(\*\), 0\) \* 100 as error_rate FROM backends b LEFT JOIN LATERAL`).
		WithArgs("fly").
		WillReturnRows(sqlmock.NewRows([]string{"region", "total", "healthy", "avg_latency", "error_rate"}).
			AddRow("iad", 5, 4, 45.5, 20.0).
			AddRow("fra", 5, 4, 65.2, 0.0))

	// Second query: region details for vercel
	mock.ExpectQuery(`SELECT b.region, COUNT\(\*\) as total_backends, COUNT\(\*\) FILTER \(WHERE hc.ok = true\) as healthy_count, AVG\(hc.latency_ms\) FILTER \(WHERE hc.ok = true\) as avg_latency, COUNT\(\*\) FILTER \(WHERE hc.ok = false\)::float \/ NULLIF\(COUNT\(\*\), 0\) \* 100 as error_rate FROM backends b LEFT JOIN LATERAL`).
		WithArgs("vercel").
		WillReturnRows(sqlmock.NewRows([]string{"region", "total", "healthy", "avg_latency", "error_rate"}).
			AddRow("iad", 5, 5, 32.1, 0.0))

	providers, err := repo.GetProviderStatus(ctx)

	require.NoError(t, err)
	assert.Len(t, providers, 2)

	// Check Fly provider
	flyProvider := providers[0]
	assert.Equal(t, "fly", flyProvider.Name)
	assert.Equal(t, 10, flyProvider.Summary.TotalBackends)
	assert.Equal(t, 8, flyProvider.Summary.HealthyBackends)
	assert.Len(t, flyProvider.Regions, 2)

	// Check Vercel provider
	vercelProvider := providers[1]
	assert.Equal(t, "vercel", vercelProvider.Name)
	assert.Equal(t, 5, vercelProvider.Summary.TotalBackends)
	assert.Equal(t, "operational", vercelProvider.OverallStatus)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_dbIncidentToIncident(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewRepository(db)
	now := time.Now()
	resolvedAt := now.Add(30 * time.Minute)

	t.Run("unresolved incident", func(t *testing.T) {
		dbIncident := &DatabaseIncident{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Title:       "Test Incident",
			Severity:    "high",
			Status:      "investigating",
			Description: "Test description",
			CreatedAt:   now,
			ResolvedAt:  nil,
			UpdatedAt:   now,
		}

		updates := []IncidentUpdate{
			{ID: "update-1", Status: "investigating", Message: "Investigating", CreatedAt: now},
		}

		incident := repo.dbIncidentToIncident(dbIncident, updates)

		assert.Equal(t, dbIncident.ID.String(), incident.ID)
		assert.Equal(t, dbIncident.Title, incident.Title)
		assert.Nil(t, incident.ResolvedAt)
		assert.Nil(t, incident.DurationMinutes)
		assert.Len(t, incident.Updates, 1)
	})

	t.Run("resolved incident with duration", func(t *testing.T) {
		dbIncident := &DatabaseIncident{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Title:       "Resolved Incident",
			Severity:    "low",
			Status:      "resolved",
			Description: "Test description",
			CreatedAt:   now,
			ResolvedAt:  &resolvedAt,
			UpdatedAt:   resolvedAt,
		}

		incident := repo.dbIncidentToIncident(dbIncident, []IncidentUpdate{})

		assert.NotNil(t, incident.ResolvedAt)
		assert.NotNil(t, incident.DurationMinutes)
		assert.Equal(t, 30, *incident.DurationMinutes)
	})
}
