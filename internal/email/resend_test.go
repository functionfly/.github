package email

import (
	"errors"
	"testing"
)

// TestResendService_isRetryableError tests retry logic for different error types
func TestResendService_isRetryableError(t *testing.T) {
	svc := &ResendService{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "429 rate limit",
			err:      errors.New("HTTP 429: rate limit exceeded"),
			expected: true,
		},
		{
			name:     "500 server error",
			err:      errors.New("HTTP 500: internal server error"),
			expected: true,
		},
		{
			name:     "503 service unavailable",
			err:      errors.New("HTTP 503: service unavailable"),
			expected: true,
		},
		{
			name:     "400 bad request (not retryable)",
			err:      errors.New("HTTP 400: bad request"),
			expected: false,
		},
		{
			name:     "401 unauthorized (not retryable)",
			err:      errors.New("HTTP 401: unauthorized"),
			expected: false,
		},
		{
			name:     "generic error (not retryable)",
			err:      errors.New("something went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, expected %v for error: %v", result, tt.expected, tt.err)
			}
		})
	}
}

// TestResendService_ValidateConfiguration tests configuration validation
func TestResendService_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      ResendConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: ResendConfig{
				APIKey:    "re_test_key",
				FromEmail: "noreply@functionfly.com",
				FromName:  "FunctionFly",
				BaseURL:   "https://functionfly.com",
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: ResendConfig{
				APIKey:    "",
				FromEmail: "noreply@functionfly.com",
				BaseURL:   "https://functionfly.com",
			},
			expectError: true,
		},
		{
			name: "invalid email format",
			config: ResendConfig{
				APIKey:    "re_test_key",
				FromEmail: "invalid-email",
				BaseURL:   "https://functionfly.com",
			},
			expectError: true,
		},
		{
			name: "missing base URL",
			config: ResendConfig{
				APIKey:    "re_test_key",
				FromEmail: "noreply@functionfly.com",
				BaseURL:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewResendService(tt.config)
			err := svc.ValidateConfiguration()

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
