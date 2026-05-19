package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIPRateLimiter(t *testing.T) {
	limiter := newIPRateLimiter()

	for i := 0; i < rateLimitMaxRequests; i++ {
		if !limiter.allow("192.168.1.1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	if limiter.allow("192.168.1.1") {
		t.Error("Request over limit should be blocked")
	}

	if !limiter.allow("192.168.1.2") {
		t.Error("Different IP should be allowed")
	}
}

func TestValidateAuthToken(t *testing.T) {
	tests := []struct {
		token  string
		valid  bool
	}{
		{"valid_token_12345", true},
		{"ab", false},
		{strings.Repeat("a", 129), false},
		{strings.Repeat("a", 16), true},
		{"invalid token!", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := validateAuthToken(tt.token); got != tt.valid {
			t.Errorf("validateAuthToken(%q) = %v, want %v", tt.token, got, tt.valid)
		}
	}
}

func TestHandleExecute_MethodNotAllowed(t *testing.T) {
	handler := HandleExecute(nil, "")

	req := httptest.NewRequest(http.MethodGet, "/execute", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleExecute_CodeRequired(t *testing.T) {
	handler := HandleExecute(nil, "")

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{"code": ""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleExecute_CodeTooLarge(t *testing.T) {
	handler := HandleExecute(nil, "")

	largeCode := strings.Repeat("x", maxCodeSizeBytes+1)
	reqBody := `{"code": "` + largeCode + `"}`

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleExecute_InvalidJSON(t *testing.T) {
	handler := HandleExecute(nil, "")

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleExecute_AuthTokenRequired(t *testing.T) {
	handler := HandleExecute(nil, "required_token")

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{"code": "print(1)"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandleExecute_AuthTokenInvalid(t *testing.T) {
	handler := HandleExecute(nil, "required_token")

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{"code": "print(1)"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", "short")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}