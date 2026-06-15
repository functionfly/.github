package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestErrorNormalizer_BadRequest verifies that http.Error responses are normalized
func TestErrorNormalizer_BadRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid input", http.StatusBadRequest)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ErrorNormalizerMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if code, ok := resp["code"].(string); !ok || code != "BAD_REQUEST" {
		t.Errorf("Expected code BAD_REQUEST, got %v", resp["code"])
	}

	if message, ok := resp["message"].(string); !ok || message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got %v", resp["message"])
	}
}

// TestErrorNormalizer_NotFound verifies 404 responses are normalized
func TestErrorNormalizer_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "User not found", http.StatusNotFound)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ErrorNormalizerMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if code, ok := resp["code"].(string); !ok || code != "NOT_FOUND" {
		t.Errorf("Expected code NOT_FOUND, got %v", resp["code"])
	}
}

// TestErrorNormalizer_Success verifies successful responses pass through unchanged
func TestErrorNormalizer_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ErrorNormalizerMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if body != `{"result":"ok"}` {
		t.Errorf("Expected body {\"result\":\"ok\"}, got %s", body)
	}
}

// TestErrorNormalizer_InternalServerError verifies 500 responses
func TestErrorNormalizer_InternalServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ErrorNormalizerMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if code, ok := resp["code"].(string); !ok || code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %v", resp["code"])
	}
}

// TestErrorNormalizer_PreservesExistingJSONCode verifies that responses with existing error codes are preserved
func TestErrorNormalizer_PreservesExistingJSONCode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"CUSTOM_CODE","message":"Custom error","field":"email"}`))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ErrorNormalizerMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if code, ok := resp["code"].(string); !ok || code != "CUSTOM_CODE" {
		t.Errorf("Expected code CUSTOM_CODE, got %v", resp["code"])
	}
}

// TestLogSampler_AlwaysLogsErrors verifies 5xx errors are always logged
func TestLogSampler_AlwaysLogsErrors(t *testing.T) {
	sampler := NewLogSampler(LogSamplerConfig{
		SuccessSampleRate:    0.0, // Don't log any successes
		ClientErrorSampleRate: 0.0, // Don't log client errors
		AlwaysLogErrors:      true,
	})

	if !sampler.ShouldLog(500, 100*time.Millisecond) {
		t.Error("Expected 500 error to be logged")
	}
	if !sampler.ShouldLog(503, 100*time.Millisecond) {
		t.Error("Expected 503 error to be logged")
	}
}

// TestLogSampler_LogsSlowRequests verifies slow requests are always logged
func TestLogSampler_LogsSlowRequests(t *testing.T) {
	sampler := NewLogSampler(LogSamplerConfig{
		SuccessSampleRate: 0.0,
		SlowThreshold:     500 * time.Millisecond,
	})

	if !sampler.ShouldLog(200, 600*time.Millisecond) {
		t.Error("Expected slow request to be logged")
	}
}
