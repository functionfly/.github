// Package version provides integration tests for version handlers.
package version

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup ====================

type handlerTestSuite struct {
	db      *sql.DB
	mock    sqlmock.Sqlmock
	repo    *versioning.Repository
	handler *Handler
	router  *mux.Router
}

func setupHandlerTest(t *testing.T) *handlerTestSuite {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := versioning.NewRepository(db)
	handler := NewHandler(repo)

	router := mux.NewRouter()

	return &handlerTestSuite{
		db:      db,
		mock:    mock,
		repo:    repo,
		handler: handler,
		router:  router,
	}
}

func (s *handlerTestSuite) tearDown() {
	if s.db != nil {
		s.db.Close()
	}
}

// ==================== Handler Route Setup ====================

func (s *handlerTestSuite) setupRoutes() {
	s.router.HandleFunc("/api/versions", s.handler.HandleListVersions).Methods("GET")
	s.router.HandleFunc("/api/versions/{version}", s.handler.HandleGetVersion).Methods("GET")
	s.router.HandleFunc("/api/versions", s.handler.HandleCreateAPIVersion).Methods("POST")
	s.router.HandleFunc("/api/versions/{version}/deprecate", s.handler.HandleDeprecateVersion).Methods("POST")
	s.router.HandleFunc("/functions/{functionId}/versions", s.handler.HandleListFunctionVersions).Methods("GET")
	s.router.HandleFunc("/functions/{functionId}/versions/{version}/publish", s.handler.HandlePublishVersion).Methods("POST")
	s.router.HandleFunc("/functions/{functionId}/versions/{version}/deprecate", s.handler.HandleDeprecateFunctionVersion).Methods("POST")
	s.router.HandleFunc("/functions/{functionId}/rollback", s.handler.HandleRollbackLatest).Methods("POST")
	s.router.HandleFunc("/functions/{functionId}/versions/{version}/alias", s.handler.HandleSetAlias).Methods("POST")
}

// ==================== API Version Handler Tests ====================

func TestHandleListVersions_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	now := time.Now()
	s.setupRoutes()

	// Mock the database query
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "released_at",
	}).AddRow(uuid.New(), "v1", "/v1", "active", now).
		AddRow(uuid.New(), "v2", "/v2", "deprecated", now)

	s.mock.ExpectQuery("SELECT .* FROM api_versions").
		WillReturnRows(rows)

	// Create request
	req := httptest.NewRequest("GET", "/api/versions", nil)
	w := httptest.NewRecorder()

	// Serve request
	s.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	versions := response["versions"].([]interface{})
	assert.Len(t, versions, 2)
}

func TestHandleListVersions_Empty(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	// Mock empty result
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "released_at",
	})

	s.mock.ExpectQuery("SELECT .* FROM api_versions").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/api/versions", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetVersion_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	versionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "released_at",
		"deprecated_at", "sunset_at", "sunset_message", "metadata",
		"openapi_spec_url", "changelog_url", "created_at", "updated_at",
	}).AddRow(
		versionID, "v1", "/v1", "active", now,
		nil, nil, "", nil, "", "", now, now,
	)

	s.mock.ExpectQuery("SELECT .* FROM api_versions WHERE version = \\$1").
		WithArgs("v1").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/api/versions/v1", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetVersion_NotFound(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	s.mock.ExpectQuery("SELECT .* FROM api_versions WHERE version = \\$1").
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/versions/v3", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCreateAPIVersion_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	s.mock.ExpectExec("INSERT INTO api_versions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"version":    "v3",
		"pathPrefix": "/v3",
		"status":     "active",
		"releasedAt": time.Now().Format(time.RFC3339),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/versions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// May return 201 Created or 200 OK depending on implementation
	assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK)
}

func TestHandleDeprecateVersion_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	versionID := uuid.New()
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)

	// First query to get the version
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status",
	}).AddRow(versionID, "v1", "/v1", "active")

	s.mock.ExpectQuery("SELECT .* FROM api_versions WHERE version = \\$1").
		WithArgs("v1").
		WillReturnRows(rows)

	// Update query
	s.mock.ExpectExec("UPDATE api_versions SET status = \\$2").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"sunsetAt":      sunsetAt.Format(time.RFC3339),
		"sunsetMessage": "Use v2 instead",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/versions/v1/deprecate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Function Version Handler Tests ====================

func TestHandleListFunctionVersions_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	functionID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	}).AddRow(uuid.New(), functionID, "v2.0.0", "published").
		AddRow(uuid.New(), functionID, "v1.0.0", "deprecated")

	s.mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1").
		WithArgs(functionID).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/functions/"+functionID.String()+"/versions", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleListFunctionVersions_Empty(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	functionID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	})

	s.mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1").
		WithArgs(functionID).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/functions/"+functionID.String()+"/versions", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlePublishVersion_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	functionID := uuid.New()
	versionID := uuid.New()
	now := time.Now()

	// First get the version
	getRows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	}).AddRow(versionID, functionID, "v1.0.0", "draft")

	s.mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1 AND version = \\$2").
		WithArgs(functionID, "v1.0.0").
		WillReturnRows(getRows)

	// Update to published
	updateRows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "published_at",
	}).AddRow(versionID, functionID, "v1.0.0", "published", now)

	s.mock.ExpectQuery("UPDATE registry_function_versions SET version_state = \\$2, published_at = \\$3").
		WillReturnRows(updateRows)

	body := map[string]interface{}{
		"setAsLatest": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/functions/"+functionID.String()+"/versions/v1.0.0/publish", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDeprecateFunctionVersion_Success(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	functionID := uuid.New()
	versionID := uuid.New()

	// First get the version
	getRows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	}).AddRow(versionID, functionID, "v1.0.0", "published")

	s.mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1 AND version = \\$2").
		WithArgs(functionID, "v1.0.0").
		WillReturnRows(getRows)

	// Update to deprecated
	s.mock.ExpectExec("UPDATE registry_function_versions SET version_state = \\$2").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"reason":          "Security issue",
		"replacedBy":      "v2.0.0",
		"migrationGuide":  "https://example.com/migration",
		"gracePeriodDays": 30,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/functions/"+functionID.String()+"/versions/v1.0.0/deprecate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Context Tests ====================

func TestHandler_UsesContext(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "released_at",
	}).AddRow(uuid.New(), "v1", "/v1", "active", now)

	s.mock.ExpectQuery("SELECT .* FROM api_versions").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/api/versions", nil)
	ctx := context.WithValue(req.Context(), "testKey", "testValue")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Error Handling Tests ====================

func TestHandleGetVersion_DatabaseError(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	s.mock.ExpectQuery("SELECT .* FROM api_versions WHERE version = \\$1").
		WillReturnError(assert.AnError)

	req := httptest.NewRequest("GET", "/api/versions/v1", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Request Validation Tests ====================

func TestHandleCreateAPIVersion_InvalidBody(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	req := httptest.NewRequest("POST", "/api/versions", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Should return an error (400 or 500 depending on validation)
	assert.True(t, w.Code >= http.StatusBadRequest)
}

func TestHandlePublishVersion_MissingBody(t *testing.T) {
	s := setupHandlerTest(t)
	defer s.tearDown()

	s.setupRoutes()

	functionID := uuid.New()

	req := httptest.NewRequest("POST", "/functions/"+functionID.String()+"/versions/v1.0.0/publish", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Should handle missing body gracefully
	assert.True(t, w.Code >= http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}
