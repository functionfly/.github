package apierror

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogAndInternal_SanitizesErrText(t *testing.T) {
	const internalMsg = "pq: duplicate key value violates unique constraint \"users_email_key\""

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/users", strings.NewReader("{}"))
	req.Header.Set("X-Request-ID", "req-abc123")

	LogAndInternal(rr, req, errors.New(internalMsg), "create user")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	body := rr.Body.String()
	if strings.Contains(body, internalMsg) {
		t.Fatalf("INTERNAL ERROR LEAK: response body contains err text: %s", body)
	}
	if strings.Contains(body, "pq:") {
		t.Fatalf("INTERNAL ERROR LEAK: response body contains driver prefix: %s", body)
	}

	var resp APIError
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Code != ErrCodeInternal {
		t.Errorf("Expected code %q, got %q", ErrCodeInternal, resp.Code)
	}
	if resp.Message != "Internal server error" {
		t.Errorf("Expected generic message, got %q", resp.Message)
	}
}

func TestLogAndInternal_HandlesNilErr(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	LogAndInternal(rr, req, nil, "test")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	var resp APIError
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Message != "Internal server error" {
		t.Errorf("Expected generic message, got %q", resp.Message)
	}
}

func TestLogAndInternal_HandlesNilRequest(t *testing.T) {
	rr := httptest.NewRecorder()

	LogAndInternal(rr, nil, errors.New("boom"), "test")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestLogAndXxx_ProducesCorrectStatuses(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(http.ResponseWriter, *http.Request, error, string)
		wantCode int
		wantAPIC ErrorCode
	}{
		{"Internal", LogAndInternal, http.StatusInternalServerError, ErrCodeInternal},
		{"BadRequest", LogAndBadRequest, http.StatusBadRequest, ErrCodeBadRequest},
		{"NotFound", LogAndNotFound, http.StatusNotFound, ErrCodeNotFound},
		{"Conflict", LogAndConflict, http.StatusConflict, ErrCodeConflict},
		{"Forbidden", LogAndForbidden, http.StatusForbidden, ErrCodeForbidden},
		{"Unauthorized", LogAndUnauthorized, http.StatusUnauthorized, ErrCodeUnauthorized},
		{"ServiceUnavailable", LogAndServiceUnavailable, http.StatusServiceUnavailable, ErrCodeServiceUnavailable},
		{"Unprocessable", LogAndUnprocessable, http.StatusUnprocessableEntity, ErrCodeValidation},
		{"GatewayTimeout", LogAndGatewayTimeout, http.StatusGatewayTimeout, ErrCodeGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("X-Request-ID", "req-test")

			tt.fn(rr, req, errors.New("sensitive: pq: unique violation"), "test ctx")

			if rr.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, rr.Code)
			}

			body := rr.Body.String()
			if strings.Contains(body, "sensitive") || strings.Contains(body, "pq:") {
				t.Errorf("LEAK in %s: body=%s", tt.name, body)
			}

			var resp APIError
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if resp.Code != tt.wantAPIC {
				t.Errorf("Expected code %q, got %q", tt.wantAPIC, resp.Code)
			}
		})
	}
}

func TestLogAnd_CustomStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	apiErr := LogAnd(req, errors.New("boom"), http.StatusTeapot, "IM_A_TEAPOT", "test")
	WriteError(rr, apiErr)

	if rr.Code != http.StatusTeapot {
		t.Errorf("Expected status %d, got %d", http.StatusTeapot, rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "boom") {
		t.Errorf("LEAK: %s", body)
	}
}

func TestFromError_NilErr(t *testing.T) {
	apiErr := FromError(nil)
	if apiErr == nil {
		t.Fatal("FromError(nil) returned nil")
	}
	if apiErr.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", apiErr.Status)
	}
	if apiErr.Message != "Internal server error" {
		t.Errorf("Expected generic message, got %q", apiErr.Message)
	}
}

func TestFromError_GenericErr(t *testing.T) {
	apiErr := FromError(errors.New("pq: connection refused"))
	if apiErr == nil {
		t.Fatal("FromError returned nil")
	}
	if apiErr.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", apiErr.Status)
	}
	if strings.Contains(apiErr.Message, "pq:") || strings.Contains(apiErr.Message, "refused") {
		t.Errorf("LEAK in FromError: %q", apiErr.Message)
	}
}

func TestFromError_APIErrorPassesThrough(t *testing.T) {
	src := &APIError{
		Status:  http.StatusBadRequest,
		Code:    ErrCodeValidation,
		Message: "this should be sanitized",
		Field:   "email",
	}

	apiErr := FromError(src)
	if apiErr == nil {
		t.Fatal("FromError returned nil")
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", apiErr.Status)
	}
	if apiErr.Code != ErrCodeValidation {
		t.Errorf("Expected code %q, got %q", ErrCodeValidation, apiErr.Code)
	}
	if apiErr.Message == "this should be sanitized" {
		t.Error("FromError did not sanitize message from APIError source")
	}
	if apiErr.Field != "email" {
		t.Errorf("Expected field 'email', got %q", apiErr.Field)
	}
}

func TestSanitizeMessage_SafeHandWritten4xx(t *testing.T) {
	cases := []string{
		"User not found",
		"Invalid API version",
		"Email is required",
		"Authentication required",
	}
	for _, msg := range cases {
		got := SanitizeMessage(http.StatusNotFound, msg)
		if got != msg {
			t.Errorf("SanitizeMessage(%q) = %q; want %q (hand-written message should pass through)", msg, got, msg)
		}
	}
}

func TestSanitizeMessage_5xxAlwaysGeneric(t *testing.T) {
	cases := []string{
		"pq: duplicate key value",
		"connection refused",
		"User not found",
		"anything",
	}
	for _, msg := range cases {
		got := SanitizeMessage(http.StatusInternalServerError, msg)
		if got != "Internal server error" {
			t.Errorf("SanitizeMessage(500, %q) = %q; want %q", msg, got, "Internal server error")
		}
	}
}

func TestSanitizeMessage_4xxWithErrText(t *testing.T) {
	cases := []string{
		"pq: invalid input syntax",
		"json: cannot unmarshal string into int",
		"/home/user/secret.txt: no such file",
		"goroutine 1 [running]: panic",
	}
	for _, msg := range cases {
		got := SanitizeMessage(http.StatusBadRequest, msg)
		if got != "Bad request" {
			t.Errorf("SanitizeMessage(400, %q) = %q; want %q (err text should be sanitized)", msg, got, "Bad request")
		}
	}
}

func TestSanitizeMessage_LongBody(t *testing.T) {
	long := strings.Repeat("a", 250)
	got := SanitizeMessage(http.StatusBadRequest, long)
	if got != "Bad request" {
		t.Errorf("SanitizeMessage should reject long body, got %q", got)
	}
}

func TestSanitizeMessage_EmptyBody(t *testing.T) {
	got := SanitizeMessage(http.StatusNotFound, "")
	if got != "Not found" {
		t.Errorf("SanitizeMessage(empty) = %q; want 'Not found'", got)
	}
}

func TestGenericCodeMessages_AllCodesCovered(t *testing.T) {
	allCodes := []ErrorCode{
		ErrCodeBadRequest, ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeNotFound,
		ErrCodeConflict, ErrCodeValidation, ErrCodeRateLimited, ErrCodeUnprocessable,
		ErrCodeUpgradeRequired, ErrCodeLocked, ErrCodePrecondition, ErrCodeTooManyRequests,
		ErrCodeInternal, ErrCodeServiceUnavailable, ErrCodeGatewayTimeout, ErrCodeNotImplemented,
		ErrCodeDependencyFailed, ErrCodeBilling, ErrCodeAuth, ErrCodeResourceExhausted,
		ErrCodeQuotaExceeded, ErrCodeInvalidState,
	}
	for _, code := range allCodes {
		if _, ok := genericCodeMessages[code]; !ok {
			t.Errorf("genericCodeMessages missing entry for %q", code)
		}
	}
}
