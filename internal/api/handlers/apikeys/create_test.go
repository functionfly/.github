package apikeys

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/stretchr/testify/assert"
)

// TestValidateCreateRequest tests request validation
func TestValidateCreateRequest(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name    string
		req     *apikey.CreateAPIKeyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &apikey.CreateAPIKeyRequest{
				Name:    "test-key",
				KeyType: apikey.KeyTypePlatform,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: &apikey.CreateAPIKeyRequest{
				KeyType: apikey.KeyTypePlatform,
			},
			wantErr: true,
		},
		{
			name: "missing key type",
			req: &apikey.CreateAPIKeyRequest{
				Name: "test-key",
			},
			wantErr: true,
		},
		{
			name: "invalid key type",
			req: &apikey.CreateAPIKeyRequest{
				Name:    "test-key",
				KeyType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative rotation days",
			req: &apikey.CreateAPIKeyRequest{
				Name:                  "test-key",
				KeyType:               apikey.KeyTypePlatform,
				RotationFrequencyDays: -1,
			},
			wantErr: true,
		},
		{
			name: "negative rate limit RPM",
			req: &apikey.CreateAPIKeyRequest{
				Name:    "test-key",
				KeyType: apikey.KeyTypePlatform,
				RateLimit: &apikey.RateLimitConfig{
					RPM: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.validateCreateRequest(tt.req, "")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestJoinErrors tests error message joining
func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
		expect string
	}{
		{
			name:   "single error",
			errors: []string{"error1"},
			expect: "error1",
		},
		{
			name:   "multiple errors",
			errors: []string{"error1", "error2"},
			expect: "error1; error2",
		},
		{
			name:   "empty errors",
			errors: []string{},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinErrors(tt.errors)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// TestWriteJSON tests JSON writing helper
func TestWriteJSON(t *testing.T) {
	handler := &Handler{}

	w := httptest.NewRecorder()
	handler.writeJSON(w, http.StatusOK, map[string]string{"test": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

// TestWriteSuccess tests success response writing
func TestWriteSuccess(t *testing.T) {
	handler := &Handler{}

	w := httptest.NewRecorder()
	handler.writeSuccess(w, map[string]string{"test": "value"})

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestWriteError tests error response writing
func TestWriteError(t *testing.T) {
	handler := &Handler{}

	w := httptest.NewRecorder()
	handler.writeError(w, http.StatusBadRequest, "invalid_request", "Test error message")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestUserClaims tests UserClaims struct
func TestUserClaims(t *testing.T) {
	claims := UserClaims{}

	// Just verify the struct can be created
	assert.NotNil(t, &claims)
}

// TestExtractPathVars tests path variable extraction
func TestExtractPathVars(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	// Test that extractPathVars returns a map
	vars := extractPathVars(req)
	assert.NotNil(t, vars)
}

// TestHandlerWriteErrorContentType tests error response content type
func TestHandlerWriteErrorContentType(t *testing.T) {
	handler := &Handler{}

	w := httptest.NewRecorder()
	handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	errObj, ok := response["error"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "unauthorized", errObj["code"])
	assert.Equal(t, "Authentication required", errObj["message"])
}

// TestHandlerWriteSuccessWithMeta tests success response with metadata
func TestHandlerWriteSuccessWithMeta(t *testing.T) {
	handler := &Handler{}

	w := httptest.NewRecorder()
	meta := map[string]interface{}{"total": 100}
	handler.writeSuccess(w, []string{"item1", "item2"}, meta)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "meta")
}

// TestValidationError tests ValidationError struct
func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Message: "test error",
	}

	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, "test error", err.Message)
}
