// Package versioning provides tests for version middleware.
package versioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ==================== Version Repository Interface ====================

// versionRepoInterface defines the interface for version repository operations used by middleware
type versionRepoInterface interface {
	GetAPIVersionByVersion(ctx context.Context, version string) (*APIVersion, error)
}

// ==================== concrete Repository for testing ====================

// testRepo wraps Repository for testing
type testRepo struct {
	*Repository
}

// ==================== Middleware Constants Tests ====================

func TestMiddlewareConstants(t *testing.T) {
	assert.Equal(t, "v1", DefaultAPIVersion)
}

func TestContextKey(t *testing.T) {
	key := versionInfoKey
	assert.Equal(t, contextKey("version_info"), key)
}

// ==================== VersionInfo Tests ====================

func TestVersionInfo_ActiveVersion(t *testing.T) {
	info := VersionInfo{
		Version:      "v1",
		IsActive:     true,
		IsDeprecated: false,
		IsSunset:     false,
		IsDefault:    true,
	}

	assert.Equal(t, "v1", info.Version)
	assert.True(t, info.IsActive)
	assert.False(t, info.IsDeprecated)
	assert.False(t, info.IsSunset)
	assert.True(t, info.IsDefault)
}

func TestVersionInfo_DeprecatedVersion(t *testing.T) {
	deprecatedAt := time.Now()
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)

	info := VersionInfo{
		Version:      "v1",
		IsActive:     false,
		IsDeprecated: true,
		IsSunset:     false,
		IsDefault:    false,
		DeprecationInfo: &DeprecationWarning{
			DeprecatedAt:     deprecatedAt,
			SunsetAt:         &sunsetAt,
			SunsetMessage:    "Use v2 instead",
			SuccessorVersion: "v2",
		},
	}

	assert.True(t, info.IsDeprecated)
	assert.NotNil(t, info.DeprecationInfo)
	assert.Equal(t, "Use v2 instead", info.DeprecationInfo.SunsetMessage)
}

func TestVersionInfo_SunsetVersion(t *testing.T) {
	info := VersionInfo{
		Version:      "v1",
		IsActive:     false,
		IsDeprecated: true,
		IsSunset:     true,
		IsDefault:    false,
	}

	assert.True(t, info.IsSunset)
}

// ==================== DeprecationWarning Tests ====================

func TestDeprecationWarning_Fields(t *testing.T) {
	deprecatedAt := time.Now()
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)

	warning := DeprecationWarning{
		DeprecatedAt:     deprecatedAt,
		SunsetAt:         &sunsetAt,
		SunsetMessage:    "Migration required",
		SuccessorVersion: "v2",
	}

	assert.Equal(t, deprecatedAt, warning.DeprecatedAt)
	assert.NotNil(t, warning.SunsetAt)
	assert.Equal(t, "Migration required", warning.SunsetMessage)
	assert.Equal(t, "v2", warning.SuccessorVersion)
}

// ==================== Create middleware for testing ====================

// testVersionMiddleware creates a VersionMiddleware for testing without a real DB
func testVersionMiddleware() *VersionMiddleware {
	// We can't use a real Repository without a DB, so we'll test the parts
	// that don't require the repository
	vm := &VersionMiddleware{
		repo:           nil, // Will be nil but we'll pre-populate cache for tests
		cache:          make(map[string]*APIVersion),
		defaultVersion: DefaultAPIVersion,
	}
	return vm
}

// ==================== Extract Version Tests ====================

func TestExtractVersion_FromPath(t *testing.T) {
	vm := testVersionMiddleware()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"v1 in path", "/v1/users", "v1"},
		{"v2 in path", "/v2/functions", "v2"},
		{"v10 in path", "/v10/endpoint", "v10"},
		{"no version uses default", "/users", "v1"},
		{"version at root", "/v1/", "v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			result := vm.extractVersion(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractVersion_FromHeader(t *testing.T) {
	vm := testVersionMiddleware()

	// Create request without version in path but with Accept-Version header
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Accept-Version", "v2")

	result := vm.extractVersion(req)
	assert.Equal(t, "v2", result)
}

func TestExtractVersion_HeaderWithoutVPrefix(t *testing.T) {
	vm := testVersionMiddleware()

	// Header without v prefix should still work
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Accept-Version", "2")

	result := vm.extractVersion(req)
	assert.Equal(t, "v2", result)
}

func TestExtractVersion_HeaderWithV(t *testing.T) {
	vm := testVersionMiddleware()

	// Header with v prefix should use as-is
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Accept-Version", "v3")

	result := vm.extractVersion(req)
	assert.Equal(t, "v3", result)
}

func TestExtractVersion_PathTakesPrecedence(t *testing.T) {
	vm := testVersionMiddleware()

	// Path should take precedence over header
	req := httptest.NewRequest("GET", "/v1/users", nil)
	req.Header.Set("Accept-Version", "v2")

	result := vm.extractVersion(req)
	assert.Equal(t, "v1", result)
}

// ==================== Get Version Info Tests (with cache) ====================

func TestGetVersionInfo_FromCache(t *testing.T) {
	vm := testVersionMiddleware()

	// Pre-populate cache
	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	vm.cache["v1"] = apiVersion

	ctx := context.Background()
	info := vm.getVersionInfo(ctx, "v1")

	assert.NotNil(t, info)
	assert.Equal(t, "v1", info.Version)
	assert.True(t, info.IsActive)
}

func TestGetVersionInfo_NotFound_UsesDefault(t *testing.T) {
	vm := testVersionMiddleware()
	vm.defaultVersion = "v1"

	ctx := context.Background()
	info := vm.getVersionInfo(ctx, "v3") // Not in cache, no repo

	// Should fall back to default
	assert.NotNil(t, info)
	assert.Equal(t, "v1", info.Version)
	assert.True(t, info.IsDefault)
}

func TestGetVersionInfo_NoCache_NoRepo(t *testing.T) {
	vm := testVersionMiddleware()
	vm.defaultVersion = "v1"

	ctx := context.Background()
	info := vm.getVersionInfo(ctx, "v1")

	// No cache, no repo - should use default
	assert.NotNil(t, info)
	assert.Equal(t, DefaultAPIVersion, info.Version)
	assert.True(t, info.IsDefault)
}

// ==================== To Version Info Tests ====================

func TestToVersionInfo_Active(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v2",
		PathPrefix: "/v2",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	info := vm.toVersionInfo(apiVersion)

	assert.Equal(t, "v2", info.Version)
	assert.True(t, info.IsActive)
	assert.False(t, info.IsDeprecated)
	assert.False(t, info.IsSunset)
}

func TestToVersionInfo_Deprecated(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	deprecatedAt := now.Add(-30 * 24 * time.Hour)
	apiVersion := &APIVersion{
		ID:            uuid.New(),
		Version:       "v1",
		PathPrefix:    "/v1",
		Status:        APIVersionStatusDeprecated,
		ReleasedAt:    now,
		DeprecatedAt:  &deprecatedAt,
		SunsetMessage: "Use v2 instead",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	info := vm.toVersionInfo(apiVersion)

	assert.Equal(t, "v1", info.Version)
	assert.True(t, info.IsDeprecated)
	assert.NotNil(t, info.DeprecationInfo)
	assert.Equal(t, "Use v2 instead", info.DeprecationInfo.SunsetMessage)
}

func TestToVersionInfo_SunsetByDate(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	sunsetAt := now.Add(-24 * time.Hour) // Past date
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusDeprecated,
		ReleasedAt: now,
		SunsetAt:   &sunsetAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	info := vm.toVersionInfo(apiVersion)

	assert.True(t, info.IsSunset)
	assert.True(t, info.IsDeprecated)
}

func TestToVersionInfo_Archived(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusArchived,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	info := vm.toVersionInfo(apiVersion)

	assert.Equal(t, "v1", info.Version)
	assert.False(t, info.IsActive)
}

// ==================== Get Successor Version Tests ====================

func TestGetSuccessorVersion(t *testing.T) {
	vm := testVersionMiddleware()

	tests := []struct {
		version  string
		expected string
	}{
		{"v1", "v2"},
		{"v2", "v3"},
		{"v3", ""},  // No known successor
		{"v99", ""}, // No known successor
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			apiVersion := &APIVersion{Version: tt.version}
			result := vm.getSuccessorVersion(apiVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSuccessorVersion_InvalidFormat(t *testing.T) {
	vm := testVersionMiddleware()

	apiVersion := &APIVersion{Version: "invalid"}
	result := vm.getSuccessorVersion(apiVersion)
	assert.Equal(t, "", result)
}

// ==================== Handler Integration Tests ====================

func TestHandler_DeprecationHeaders_ActiveVersion(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	vm.cache["v1"] = apiVersion

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := vm.Handler()

	req := httptest.NewRequest("GET", "/v1/test", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))
	assert.Equal(t, "", w.Header().Get("Deprecation"))
}

func TestHandler_DeprecationHeaders_DeprecatedVersion(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	deprecatedAt := now.Add(-30 * 24 * time.Hour)
	sunsetAt := now.Add(30 * 24 * time.Hour)

	apiVersion := &APIVersion{
		ID:            uuid.New(),
		Version:       "v1",
		PathPrefix:    "/v1",
		Status:        APIVersionStatusDeprecated,
		ReleasedAt:    now,
		DeprecatedAt:  &deprecatedAt,
		SunsetAt:      &sunsetAt,
		SunsetMessage: "Use v2 instead",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	vm.cache["v1"] = apiVersion

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := vm.Handler()

	req := httptest.NewRequest("GET", "/v1/test", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("Deprecation"))
	assert.Equal(t, "This API version is deprecated", w.Header().Get("X-API-Warning"))
	assert.NotEmpty(t, w.Header().Get("Sunset"))
	assert.Equal(t, "</v2/>; rel=\"successor-version\"", w.Header().Get("Link"))
}

func TestHandler_SunsetVersion_Returns410(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	sunsetAt := now.Add(-24 * time.Hour) // Past date

	apiVersion := &APIVersion{
		ID:            uuid.New(),
		Version:       "v1",
		PathPrefix:    "/v1",
		Status:        APIVersionStatusSunset,
		ReleasedAt:    now,
		SunsetAt:      &sunsetAt,
		SunsetMessage: "This version is no longer available",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	vm.cache["v1"] = apiVersion

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := vm.Handler()

	req := httptest.NewRequest("GET", "/v1/test", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
	assert.Equal(t, "true", w.Header().Get("Deprecation"))

	// Check response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), assert.AnError)
	assert.NoError(t, err)
	assert.Equal(t, "VERSION_SUNSET", response["error"])
}

func TestHandler_SunsetMessage_Header(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	sunsetAt := now.Add(-24 * time.Hour)

	apiVersion := &APIVersion{
		ID:            uuid.New(),
		Version:       "v1",
		PathPrefix:    "/v1",
		Status:        APIVersionStatusSunset,
		ReleasedAt:    now,
		SunsetAt:      &sunsetAt,
		SunsetMessage: "Use the new API",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	vm.cache["v1"] = apiVersion

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := vm.Handler()

	req := httptest.NewRequest("GET", "/v1/test", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, "Use the new API", w.Header().Get("X-API-Sunset-Message"))
}

// ==================== Context Functions Tests ====================

func TestGetVersionInfo_FromContext(t *testing.T) {
	info := &VersionInfo{
		Version:   "v1",
		IsActive:  true,
		IsDefault: true,
	}
	ctx := context.WithValue(context.Background(), versionInfoKey, info)

	result := GetVersionInfo(ctx)
	assert.NotNil(t, result)
	assert.Equal(t, "v1", result.Version)
}

func TestGetVersionInfo_EmptyContext(t *testing.T) {
	ctx := context.Background()

	result := GetVersionInfo(ctx)
	assert.Nil(t, result)
}

func TestGetAPIVersionFromContext(t *testing.T) {
	info := &VersionInfo{
		Version:   "v2",
		IsActive:  true,
		IsDefault: false,
	}
	ctx := context.WithValue(context.Background(), versionInfoKey, info)

	result := GetAPIVersionFromContext(ctx)
	assert.Equal(t, "v2", result)
}

func TestGetAPIVersionFromContext_Empty(t *testing.T) {
	ctx := context.Background()

	result := GetAPIVersionFromContext(ctx)
	assert.Equal(t, DefaultAPIVersion, result)
}

// ==================== RequireAPIVersion Tests ====================

func TestRequireAPIVersion_ValidVersion(t *testing.T) {
	middleware := RequireAPIVersion("v1")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	info := &VersionInfo{Version: "v1"}
	ctx := context.WithValue(context.Background(), versionInfoKey, info)
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAPIVersion_InvalidVersion(t *testing.T) {
	middleware := RequireAPIVersion("v1")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	// Use different version in context
	info := &VersionInfo{Version: "v2"}
	ctx := context.WithValue(context.Background(), versionInfoKey, info)
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireAPIVersion_EmptyContext(t *testing.T) {
	middleware := RequireAPIVersion("v1")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== ResponseWithVersionInfo Tests ====================

func TestResponseWithVersionInfo(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "success"}
	ResponseWithVersionInfo(w, "v1", data)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), assert.AnError)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

// ==================== GetUserFromContext Tests ====================

func TestGetUserFromContext(t *testing.T) {
	ctx := context.Background()
	result := GetUserFromContext(ctx)
	assert.Nil(t, result)
}

// ==================== Integration Tests ====================

func TestMiddleware_FullFlow(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	vm.cache["v1"] = apiVersion

	callCount := 0
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Verify version info in context
		info := GetVersionInfo(r.Context())
		assert.NotNil(t, info)
		assert.Equal(t, "v1", info.Version)

		w.Header().Set("X-Custom-Header", "test")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler := vm.Handler()

	req := httptest.NewRequest("GET", "/v1/api/endpoint", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, 1, callCount)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))
	assert.Equal(t, "test", w.Header().Get("X-Custom-Header"))
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
}

func TestMiddleware_VersionPropagation(t *testing.T) {
	vm := testVersionMiddleware()

	// Test with different version paths
	versions := []string{"v1", "v2", "v3"}

	for _, version := range versions {
		now := time.Now()
		apiVersion := &APIVersion{
			ID:         uuid.New(),
			Version:    version,
			PathPrefix: "/" + version,
			Status:     APIVersionStatusActive,
			ReleasedAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		vm.cache[version] = apiVersion

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := GetVersionInfo(r.Context())
			assert.NotNil(t, info)
			w.WriteHeader(http.StatusOK)
		})

		handler := vm.Handler()
		req := httptest.NewRequest("GET", "/"+version+"/test", nil)
		w := httptest.NewRecorder()

		handler(nextHandler).ServeHTTP(w, req)

		assert.Equal(t, version, w.Header().Get("X-API-Version"))
	}
}

// ==================== NewVersionMiddleware Tests ====================

func TestNewVersionMiddleware(t *testing.T) {
	// Test that we can create middleware - repo will be nil but that's OK for basic tests
	vm := &VersionMiddleware{
		repo:           nil,
		cache:          make(map[string]*APIVersion),
		defaultVersion: DefaultAPIVersion,
	}

	assert.NotNil(t, vm)
	assert.Equal(t, DefaultAPIVersion, vm.defaultVersion)
	assert.NotNil(t, vm.cache)
}

func TestVersionMiddleware_Handler_ReturnsMiddleware(t *testing.T) {
	vm := testVersionMiddleware()

	handler := vm.Handler()
	assert.NotNil(t, handler)
}

// ==================== Edge Cases ====================

func TestExtractVersion_EmptyPath(t *testing.T) {
	vm := testVersionMiddleware()

	req := httptest.NewRequest("GET", "/", nil)
	result := vm.extractVersion(req)
	assert.Equal(t, "v1", result)
}

func TestExtractVersion_RootPath(t *testing.T) {
	vm := testVersionMiddleware()

	req := httptest.NewRequest("GET", "", nil)
	result := vm.extractVersion(req)
	assert.Equal(t, "v1", result)
}

func TestToVersionInfo_NilDeprecationInfo(t *testing.T) {
	vm := testVersionMiddleware()

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	info := vm.toVersionInfo(apiVersion)

	assert.NotNil(t, info)
	assert.Nil(t, info.DeprecationInfo)
}

func TestGetSuccessorVersion_ZeroValue(t *testing.T) {
	vm := testVersionMiddleware()

	apiVersion := &APIVersion{Version: ""}
	result := vm.getSuccessorVersion(apiVersion)
	assert.Equal(t, "", result)
}
