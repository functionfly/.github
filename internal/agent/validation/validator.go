package validation

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInputTooLarge is returned when input exceeds size limit
	ErrInputTooLarge = errors.New("input exceeds maximum size")
	// ErrOutputTooLarge is returned when output exceeds size limit
	ErrOutputTooLarge = errors.New("output exceeds maximum size")
	// ErrPIIDetected is returned when PII is detected in input
	ErrPIIDetected = errors.New("PII detected in input")
	// ErrSensitiveDataDetected is returned when sensitive data is detected in output
	ErrSensitiveDataDetected = errors.New("sensitive data detected in output")
)

// InputValidator validates agent inputs
type InputValidator struct {
	MaxSizeBytes int
	PIIPatterns  []*regexp.Regexp
}

// OutputValidator validates and sanitizes agent outputs
type OutputValidator struct {
	MaxSizeBytes   int
	RedactPatterns []*regexp.Regexp
	PIIPatterns    []*regexp.Regexp
}

// DefaultInputValidator returns a default input validator
func DefaultInputValidator() InputValidator {
	return InputValidator{
		MaxSizeBytes: 1024 * 1024, // 1MB
		PIIPatterns: []*regexp.Regexp{
			// Email pattern
			regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			// SSN pattern
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			// Credit card pattern
			regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),
			// Phone number pattern
			regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		},
	}
}

// DefaultOutputValidator returns a default output validator
func DefaultOutputValidator() OutputValidator {
	return OutputValidator{
		MaxSizeBytes: 10 * 1024 * 1024, // 10MB
		RedactPatterns: []*regexp.Regexp{
			// Password patterns
			regexp.MustCompile(`(?i)password["\s:=]+[^\s"]+`),
			regexp.MustCompile(`(?i)passwd["\s:=]+[^\s"]+`),
			// Secret patterns
			regexp.MustCompile(`(?i)secret["\s:=]+[^\s"]+`),
			regexp.MustCompile(`(?i)secret_key["\s:=]+[^\s"]+`),
			// Token patterns
			regexp.MustCompile(`(?i)token["\s:=]+[^\s"]+`),
			regexp.MustCompile(`(?i)access_token["\s:=]+[^\s"]+`),
			regexp.MustCompile(`(?i)api_key["\s:=]+[^\s"]+`),
			// Private key patterns
			regexp.MustCompile(`(?i)private_key["\s:=]+[^\s"]+`),
			regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----`),
		},
		PIIPatterns: []*regexp.Regexp{
			// Email pattern
			regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			// SSN pattern
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			// Credit card pattern
			regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),
		},
	}
}

// Validate validates agent input
func (v *InputValidator) Validate(input json.RawMessage) error {
	// Check size
	if len(input) > v.MaxSizeBytes {
		return ErrInputTooLarge
	}

	// Check for PII
	inputStr := string(input)
	for _, pattern := range v.PIIPatterns {
		if pattern.MatchString(inputStr) {
			return ErrPIIDetected
		}
	}

	return nil
}

// ValidateString validates a string input
func (v *InputValidator) ValidateString(input string) error {
	// Check size
	if len(input) > v.MaxSizeBytes {
		return ErrInputTooLarge
	}

	// Check for PII
	for _, pattern := range v.PIIPatterns {
		if pattern.MatchString(input) {
			return ErrPIIDetected
		}
	}

	return nil
}

// Validate validates and sanitizes agent output
func (v *OutputValidator) Validate(output json.RawMessage) (json.RawMessage, error) {
	// Check size
	if len(output) > v.MaxSizeBytes {
		return nil, ErrOutputTooLarge
	}

	// Redact sensitive data
	outputStr := string(output)
	for _, pattern := range v.RedactPatterns {
		outputStr = pattern.ReplaceAllString(outputStr, "[REDACTED]")
	}

	return json.RawMessage(outputStr), nil
}

// ValidateString validates and sanitizes a string output
func (v *OutputValidator) ValidateString(output string) (string, error) {
	// Check size
	if len(output) > v.MaxSizeBytes {
		return "", ErrOutputTooLarge
	}

	// Redact sensitive data
	for _, pattern := range v.RedactPatterns {
		output = pattern.ReplaceAllString(output, "[REDACTED]")
	}

	return output, nil
}

// CheckPII checks if output contains PII
func (v *OutputValidator) CheckPII(output string) bool {
	for _, pattern := range v.PIIPatterns {
		if pattern.MatchString(output) {
			return true
		}
	}
	return false
}

// SanitizeHTML sanitizes HTML content
func SanitizeHTML(input string) string {
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// Remove event handlers
	eventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*"[^"]*"`)
	input = eventRegex.ReplaceAllString(input, "")

	// Remove javascript: URLs
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	input = jsRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

// ValidateJSON validates JSON structure
func ValidateJSON(input json.RawMessage) error {
	var js json.RawMessage
	if err := json.Unmarshal(input, &js); err != nil {
		return err
	}
	return nil
}

// ValidateJSONSchema validates JSON against a schema (simplified)
func ValidateJSONSchema(input json.RawMessage, schema map[string]interface{}) error {
	// Parse input
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return err
	}

	// Check required fields
	if required, ok := schema["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				if _, exists := data[fieldName]; !exists {
					return errors.New("missing required field: " + fieldName)
				}
			}
		}
	}

	return nil
}
