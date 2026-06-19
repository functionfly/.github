package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthMiddleware creates a mock auth middleware for testing
// Uses the exported NewAuthService constructor
func mockAuthMiddleware() *AuthMiddleware {
	authSvc, err := auth.NewAuthService(nil, "c3d1148ec940122ae79123fe3f6f21ca")
	if err != nil {
		panic(err)
	}
	return NewAuthMiddleware(authSvc)
}

// mockUser creates a mock user for testing
func mockUser() *storage.User {
	return &storage.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
		Role:     "user",
	}
}

// TestGetClaimsFromContext tests extracting claims from context
func TestGetClaimsFromContext(t *testing.T) {
	claims := &auth.Claims{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		TenantID: uuid.New(),
		Role:     "user",
	}

	ctx := context.WithValue(context.Background(), contextKeyUser, claims)
	result := GetClaimsFromContext(ctx)

	assert.NotNil(t, result)
	assert.Equal(t, claims.UserID, result.UserID)
	assert.Equal(t, claims.Email, result.Email)
}

// TestGetClaimsFromContext_Empty tests extracting claims from empty context
func TestGetClaimsFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	result := GetClaimsFromContext(ctx)

	assert.Nil(t, result)
}

// TestGetUserFromContext tests extracting user from request context
func TestGetUserFromContext(t *testing.T) {
	claims := &auth.Claims{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		TenantID: uuid.New(),
		Role:     "user",
	}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), contextKeyUser, claims)
	req = req.WithContext(ctx)

	result := GetUserFromContext(req)

	assert.NotNil(t, result)
	assert.Equal(t, claims.UserID, result.UserID)
}

// TestGetUserFromContext_Empty tests extracting user from request without context
func TestGetUserFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	result := GetUserFromContext(req)

	assert.Nil(t, result)
}

// TestSetUserInContext tests setting user claims in context
func TestSetUserInContext(t *testing.T) {
	claims := &auth.Claims{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		TenantID: uuid.New(),
		Role:     "admin",
	}

	req := httptest.NewRequest("GET", "/", nil)
	resultReq := SetUserInContext(req, claims)

	result := GetUserFromContext(resultReq)
	assert.NotNil(t, result)
	assert.Equal(t, claims.UserID, result.UserID)
	assert.Equal(t, claims.Role, result.Role)
}

// TestGetActingTenantID tests getting acting tenant ID from request
func TestGetActingTenantID(t *testing.T) {
	tenantID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), contextKeyActingTenantID, &tenantID)
	req = req.WithContext(ctx)

	result := GetActingTenantID(req)

	assert.NotNil(t, result)
	assert.Equal(t, tenantID, *result)
}

// TestGetActingTenantID_Empty tests getting acting tenant ID when not set
func TestGetActingTenantID_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	result := GetActingTenantID(req)

	assert.Nil(t, result)
}

// TestSetActingTenantID tests setting acting tenant ID in request
func TestSetActingTenantID(t *testing.T) {
	tenantID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	resultReq := SetActingTenantID(req, &tenantID)

	result := GetActingTenantID(resultReq)
	assert.NotNil(t, result)
	assert.Equal(t, tenantID, *result)
}

// TestRequireAuth_NoAuthHeader tests RequireAuth with missing auth header
func TestRequireAuth_NoAuthHeader(t *testing.T) {
	authMiddleware := mockAuthMiddleware()

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler := authMiddleware.RequireAuth(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestRequireAuth_InvalidAuthHeaderFormat tests RequireAuth with invalid auth header format
func TestRequireAuth_InvalidAuthHeaderFormat(t *testing.T) {
	authMiddleware := mockAuthMiddleware()

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler := authMiddleware.RequireAuth(nextHandler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"no bearer prefix", "some-token"},
		{"empty bearer", "Bearer "},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"multiple spaces", "Bearer token with spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.False(t, handlerCalled)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

// TestRequireAuth_ValidToken tests RequireAuth with valid token
func TestRequireAuth_ValidToken(t *testing.T) {
	authMiddleware := mockAuthMiddleware()

	// Create a test user and generate a valid token
	user := mockUser()
	authSvc, err := auth.NewAuthService(nil, "c3d1148ec940122ae79123fe3f6f21ca")
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}
	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Verify user context is set
		claims := GetUserFromContext(r)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.UserID)
	})

	handler := authMiddleware.RequireAuth(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRequireAuth_ExpiredToken tests RequireAuth with invalid/expired token
func TestRequireAuth_ExpiredToken(t *testing.T) {
	authMiddleware := mockAuthMiddleware()

	// Use a malformed token
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler := authMiddleware.RequireAuth(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestNewAuthMiddleware tests the constructor
func TestNewAuthMiddleware(t *testing.T) {
	authSvc, err := auth.NewAuthService(nil, "c3d1148ec940122ae79123fe3f6f21ca")
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}
	authMiddleware := NewAuthMiddleware(authSvc)

	assert.NotNil(t, authMiddleware)
	assert.Equal(t, authSvc, authMiddleware.authSvc)
}

// TestContextKey_Uniqueness tests that context keys don't collide
func TestContextKey_Uniqueness(t *testing.T) {
	// Ensure the context key constants have unique values
	assert.NotEqual(t, contextKeyUser, contextKeyActingTenantID)
}
