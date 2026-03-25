package validation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInputValidator_Validate(t *testing.T) {
	validator := DefaultInputValidator()

	tests := []struct {
		name      string
		input     json.RawMessage
		expectErr error
	}{
		{
			name:      "valid input",
			input:     json.RawMessage(`{"data": "test"}`),
			expectErr: nil,
		},
		{
			name:      "input too large",
			input:     json.RawMessage(strings.Repeat("a", 1024*1024+1)),
			expectErr: ErrInputTooLarge,
		},
		{
			name:      "email detected",
			input:     json.RawMessage(`{"email": "user@example.com"}`),
			expectErr: ErrPIIDetected,
		},
		{
			name:      "SSN detected",
			input:     json.RawMessage(`{"ssn": "123-45-6789"}`),
			expectErr: ErrPIIDetected,
		},
		{
			name:      "credit card detected",
			input:     json.RawMessage(`{"card": "1234-5678-9012-3456"}`),
			expectErr: ErrPIIDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if err != tt.expectErr {
				t.Errorf("expected %v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestInputValidator_ValidateString(t *testing.T) {
	validator := DefaultInputValidator()

	tests := []struct {
		name      string
		input     string
		expectErr error
	}{
		{
			name:      "valid input",
			input:     "test data",
			expectErr: nil,
		},
		{
			name:      "input too large",
			input:     strings.Repeat("a", 1024*1024+1),
			expectErr: ErrInputTooLarge,
		},
		{
			name:      "email detected",
			input:     "user@example.com",
			expectErr: ErrPIIDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateString(tt.input)
			if err != tt.expectErr {
				t.Errorf("expected %v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestOutputValidator_Validate(t *testing.T) {
	validator := DefaultOutputValidator()

	tests := []struct {
		name        string
		input       json.RawMessage
		expectErr   error
		checkOutput func(json.RawMessage) bool
	}{
		{
			name:      "valid output",
			input:     json.RawMessage(`{"data": "test"}`),
			expectErr: nil,
			checkOutput: func(output json.RawMessage) bool {
				return string(output) == `{"data": "test"}`
			},
		},
		{
			name:      "output too large",
			input:     json.RawMessage(strings.Repeat("a", 10*1024*1024+1)),
			expectErr: ErrOutputTooLarge,
		},
		{
			name:      "password redacted",
			input:     json.RawMessage(`{"password": "secret123"}`),
			expectErr: nil,
			checkOutput: func(output json.RawMessage) bool {
				return strings.Contains(string(output), "[REDACTED]")
			},
		},
		{
			name:      "token redacted",
			input:     json.RawMessage(`{"token": "abc123"}`),
			expectErr: nil,
			checkOutput: func(output json.RawMessage) bool {
				return strings.Contains(string(output), "[REDACTED]")
			},
		},
		{
			name:      "api_key redacted",
			input:     json.RawMessage(`{"api_key": "key123"}`),
			expectErr: nil,
			checkOutput: func(output json.RawMessage) bool {
				return strings.Contains(string(output), "[REDACTED]")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := validator.Validate(tt.input)
			if err != tt.expectErr {
				t.Errorf("expected %v, got %v", tt.expectErr, err)
			}
			if tt.checkOutput != nil && err == nil {
				if !tt.checkOutput(output) {
					t.Errorf("output validation failed")
				}
			}
		})
	}
}

func TestOutputValidator_ValidateString(t *testing.T) {
	validator := DefaultOutputValidator()

	tests := []struct {
		name        string
		input       string
		expectErr   error
		checkOutput func(string) bool
	}{
		{
			name:      "valid output",
			input:     "test data",
			expectErr: nil,
			checkOutput: func(output string) bool {
				return output == "test data"
			},
		},
		{
			name:      "output too large",
			input:     strings.Repeat("a", 10*1024*1024+1),
			expectErr: ErrOutputTooLarge,
		},
		{
			name:      "password redacted",
			input:     `password: secret123`,
			expectErr: nil,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "[REDACTED]")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := validator.ValidateString(tt.input)
			if err != tt.expectErr {
				t.Errorf("expected %v, got %v", tt.expectErr, err)
			}
			if tt.checkOutput != nil && err == nil {
				if !tt.checkOutput(output) {
					t.Errorf("output validation failed")
				}
			}
		})
	}
}

func TestOutputValidator_CheckPII(t *testing.T) {
	validator := DefaultOutputValidator()

	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "no PII",
			output:   "test data",
			expected: false,
		},
		{
			name:     "email detected",
			output:   "user@example.com",
			expected: true,
		},
		{
			name:     "SSN detected",
			output:   "123-45-6789",
			expected: true,
		},
		{
			name:     "credit card detected",
			output:   "1234-5678-9012-3456",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.CheckPII(tt.output)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no HTML",
			input:    "test data",
			expected: "test data",
		},
		{
			name:     "script tag removed",
			input:    `<script>alert("xss")</script>test`,
			expected: "test",
		},
		{
			name:     "event handler removed",
			input:    `<div onclick="alert('xss')">test</div>`,
			expected: `<div>test</div>`,
		},
		{
			name:     "javascript URL removed",
			input:    `<a href="javascript:alert('xss')">test</a>`,
			expected: `<a href="alert('xss')">test</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     json.RawMessage
		expectErr bool
	}{
		{
			name:      "valid JSON",
			input:     json.RawMessage(`{"key": "value"}`),
			expectErr: false,
		},
		{
			name:      "invalid JSON",
			input:     json.RawMessage(`{"key": "value"`),
			expectErr: true,
		},
		{
			name:      "empty JSON",
			input:     json.RawMessage(`{}`),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error=%v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestValidateJSONSchema(t *testing.T) {
	schema := map[string]interface{}{
		"required": []interface{}{"name", "age"},
	}

	tests := []struct {
		name      string
		input     json.RawMessage
		expectErr bool
	}{
		{
			name:      "valid input",
			input:     json.RawMessage(`{"name": "test", "age": 25}`),
			expectErr: false,
		},
		{
			name:      "missing required field",
			input:     json.RawMessage(`{"name": "test"}`),
			expectErr: true,
		},
		{
			name:      "invalid JSON",
			input:     json.RawMessage(`{"name": "test"`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONSchema(tt.input, schema)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error=%v, got %v", tt.expectErr, err)
			}
		})
	}
}
