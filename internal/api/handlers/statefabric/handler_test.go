package statefabric

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create authenticated request
func createAuthRequest(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// Add auth claims to context
	tenantID := uuid.New()
	userID := uuid.New()
	claims := &auth.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
		Role:     "admin",
	}

	// Use the middleware to set the user in context
	req = middleware.SetUserInContext(req, claims)

	return req
}

// Helper to create unauthenticated request
func createUnauthenticatedRequest(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	return req
}

// TestTenantAndUser tests the tenantAndUser helper with authenticated request
func TestTenantAndUser(t *testing.T) {
	// Test with authenticated request
	tenantID := uuid.New()
	userID := uuid.New()
	claims := &auth.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req = middleware.SetUserInContext(req, claims)

	w := httptest.NewRecorder()

	returnedTenantID, returnedUserID, ok := tenantAndUser(req, w)

	assert.True(t, ok)
	assert.Equal(t, tenantID, returnedTenantID)
	assert.Equal(t, userID, returnedUserID)
}

// TestTenantAndUser_Unauthenticated tests with unauthenticated request
func TestTenantAndUser_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	_, _, ok := tenantAndUser(req, w)

	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestParseID tests the parseID helper with valid ID
func TestParseID(t *testing.T) {
	validID := uuid.New()

	// Test valid ID
	w := httptest.NewRecorder()
	id, ok := parseID(w, validID.String(), "test")

	assert.True(t, ok)
	assert.Equal(t, validID, id)
}

// TestParseID tests the parseID helper with invalid ID
func TestParseID_Invalid(t *testing.T) {
	// Test invalid ID
	w := httptest.NewRecorder()
	id, ok := parseID(w, "invalid-uuid", "test")

	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWriteJSON tests the writeJSON helper
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

// TestWriteJSON_NilData tests writeJSON with nil data
func TestWriteJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSON(w, http.StatusOK, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

// TestWriteErr tests the writeErr helper
func TestWriteErr(t *testing.T) {
	w := httptest.NewRecorder()

	writeErr(w, http.StatusInternalServerError, "test error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test error", result["error"])
}

// TestWriteErr_NotFound tests writeErr with not found status
func TestWriteErr_NotFound(t *testing.T) {
	w := httptest.NewRecorder()

	writeErr(w, http.StatusNotFound, "not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestListFabrics_Unauthorized tests list endpoint without auth
func TestListFabrics_Unauthorized(t *testing.T) {
	req := createUnauthenticatedRequest(t, "GET", "/api/v1/state-fabrics", nil)
	req = mux.SetURLVars(req, map[string]string{})

	w := httptest.NewRecorder()

	// Create a handler (we're only testing the auth part)
	handler := &Handler{}

	handler.HandleList(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateFabric_Unauthorized tests create endpoint without auth
func TestCreateFabric_Unauthorized(t *testing.T) {
	body := []byte(`{"name": "test", "type": "custom"}`)

	req := createUnauthenticatedRequest(t, "POST", "/api/v1/state-fabrics", body)
	req = mux.SetURLVars(req, map[string]string{})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleCreate(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetFabric_Unauthorized tests get endpoint without auth
func TestGetFabric_Unauthorized(t *testing.T) {
	fabricID := uuid.New()

	req := createUnauthenticatedRequest(t, "GET", "/api/v1/state-fabrics/"+fabricID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleGet(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestUpdateFabric_Unauthorized tests update endpoint without auth
func TestUpdateFabric_Unauthorized(t *testing.T) {
	fabricID := uuid.New()
	body := []byte(`{"name": "updated"}`)

	req := createUnauthenticatedRequest(t, "PUT", "/api/v1/state-fabrics/"+fabricID.String(), body)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleUpdate(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDeleteFabric_Unauthorized tests delete endpoint without auth
func TestDeleteFabric_Unauthorized(t *testing.T) {
	fabricID := uuid.New()

	req := createUnauthenticatedRequest(t, "DELETE", "/api/v1/state-fabrics/"+fabricID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleDelete(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetFabric_InvalidID tests get endpoint with invalid ID
func TestGetFabric_InvalidID(t *testing.T) {
	req := createAuthRequest(t, "GET", "/api/v1/state-fabrics/invalid-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleGet(w, req)

	// Should return 400 for invalid ID
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateFabric_InvalidID tests update endpoint with invalid ID
func TestUpdateFabric_InvalidID(t *testing.T) {
	body := []byte(`{"name": "updated"}`)

	req := createAuthRequest(t, "PUT", "/api/v1/state-fabrics/invalid-id", body)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleUpdate(w, req)

	// Should return 400 for invalid ID
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDeleteFabric_InvalidID tests delete endpoint with invalid ID
func TestDeleteFabric_InvalidID(t *testing.T) {
	req := createAuthRequest(t, "DELETE", "/api/v1/state-fabrics/invalid-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleDelete(w, req)

	// Should return 400 for invalid ID
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateFabric_BadRequest tests create with invalid JSON body
func TestCreateFabric_BadRequest(t *testing.T) {
	body := []byte(`{invalid json}`)

	req := createAuthRequest(t, "POST", "/api/v1/state-fabrics", body)
	req = mux.SetURLVars(req, map[string]string{})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleCreate(w, req)

	// Should return 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateFabric_BadRequest tests update with invalid JSON body
func TestUpdateFabric_BadRequest(t *testing.T) {
	fabricID := uuid.New()
	body := []byte(`{invalid json}`)

	req := createAuthRequest(t, "PUT", "/api/v1/state-fabrics/"+fabricID.String(), body)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleUpdate(w, req)

	// Should return 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListStores_Unauthorized tests list stores without auth
func TestListStores_Unauthorized(t *testing.T) {
	fabricID := uuid.New()

	req := createUnauthenticatedRequest(t, "GET", "/api/v1/state-fabrics/"+fabricID.String()+"/stores", nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleListStores(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListPipelines_Unauthorized tests list pipelines without auth
func TestListPipelines_Unauthorized(t *testing.T) {
	fabricID := uuid.New()

	req := createUnauthenticatedRequest(t, "GET", "/api/v1/state-fabrics/"+fabricID.String()+"/pipelines", nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleListPipelines(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListEvents_Unauthorized tests list events without auth
func TestListEvents_Unauthorized(t *testing.T) {
	fabricID := uuid.New()

	req := createUnauthenticatedRequest(t, "GET", "/api/v1/state-fabrics/"+fabricID.String()+"/events", nil)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleListEvents(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateStore_Unauthorized tests create store without auth
func TestCreateStore_Unauthorized(t *testing.T) {
	fabricID := uuid.New()
	body := []byte(`{"name": "store", "type": "persistent"}`)

	req := createUnauthenticatedRequest(t, "POST", "/api/v1/state-fabrics/"+fabricID.String()+"/stores", body)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleCreateStore(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreatePipeline_Unauthorized tests create pipeline without auth
func TestCreatePipeline_Unauthorized(t *testing.T) {
	fabricID := uuid.New()
	body := []byte(`{"name": "pipeline", "steps": []}`)

	req := createUnauthenticatedRequest(t, "POST", "/api/v1/state-fabrics/"+fabricID.String()+"/pipelines", body)
	req = mux.SetURLVars(req, map[string]string{"id": fabricID.String()})

	w := httptest.NewRecorder()

	handler := &Handler{}

	handler.HandleCreatePipeline(w, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestContextWithTenantID tests that context is properly set up with tenant ID
func TestContextWithTenantID(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	req := httptest.NewRequest("GET", "/test", nil)

	claims := &auth.Claims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       "test@example.com",
		Role:        "admin",
		Permissions: []string{"statefabric:read", "statefabric:write"},
	}

	req = middleware.SetUserInContext(req, claims)

	w := httptest.NewRecorder()

	returnedTenantID, _, ok := tenantAndUser(req, w)
	assert.True(t, ok)
	assert.Equal(t, tenantID, returnedTenantID)
}
