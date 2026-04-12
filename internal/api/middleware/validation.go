// Package middleware provides validation middleware for standardized input validation.
package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
)

// ValidatedRequest defines the interface for request bodies that support validation.
// Implement this interface on your request structs for automatic validation.
type ValidatedRequest interface {
	// Validate performs validation and returns an error if invalid.
	// The error should be user-friendly as it will be sent in API responses.
	Validate() error
}

// Validatable wraps a ValidatedRequest with the context
type Validatable interface {
	Validate() error
}

// ValidationResult holds the outcome of validation
type ValidationResult struct {
	Valid        bool
	ErrorCode    string
	ErrorMessage string
	Field        string
}

// ValidateRequestMiddleware creates middleware that validates request bodies
// against the ValidatedRequest interface.
func ValidateRequestMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only validate on write operations
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		// Check if there's a validated request in context
		if req, ok := r.Context().Value("validated_request").(ValidatedRequest); ok {
			if err := req.Validate(); err != nil {
				err := apierror.NewValidation(err.Error())
				apierror.WriteError(w, err)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

// RequireJSONValidation wraps a handler with JSON parsing and validation
// The target must implement ValidatedRequest.
func RequireJSONValidation[T ValidatedRequest](target T, handler func(http.ResponseWriter, *http.Request, T)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse JSON into a new instance
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			err := apierror.NewBadRequest("Invalid JSON: " + err.Error())
			apierror.WriteError(w, err)
			return
		}

		// Validate
		if err := req.Validate(); err != nil {
			validationErr := apierror.NewValidation(err.Error())
			apierror.WriteError(w, validationErr)
			return
		}

		handler(w, r, req)
	}
}

// ValidateWithField provides field-specific validation errors
type FieldValidator struct {
	Field string
	Value interface{}
	Rules []ValidationRule
}

// ValidationRule defines a single validation rule
type ValidationRule struct {
	Name    string
	Check   func(interface{}) bool
	Message string
}

// ValidateField performs validation on a single field
func ValidateField(field string, value interface{}, required bool, validator func(interface{}) bool, errorMsg string) *apierror.APIError {
	if required && value == nil {
		return apierror.ValidationFieldError(field, "This field is required")
	}
	if value != nil && !validator(value) {
		return apierror.ValidationFieldError(field, errorMsg)
	}
	return nil
}

// Common validation helpers
var (
	// NonEmpty validates that a string is not empty
	NonEmpty = func(v interface{}) bool {
		s, ok := v.(string)
		return ok && s != ""
	}

	// MinLength returns a validator for minimum string length
	MinLength = func(min int) func(interface{}) bool {
		return func(v interface{}) bool {
			s, ok := v.(string)
			return ok && len(s) >= min
		}
	}

	// MaxLength returns a validator for maximum string length
	MaxLength = func(max int) func(interface{}) bool {
		return func(v interface{}) bool {
			s, ok := v.(string)
			return ok && len(s) <= max
		}
	}

	// Email validates email format (basic check)
	Email = func(v interface{}) bool {
		s, ok := v.(string)
		if !ok || s == "" {
			return false
		}
		// Basic email validation
		for i := 0; i < len(s); i++ {
			if s[i] == '@' {
				return i > 0 && i < len(s)-1
			}
		}
		return false
	}

	// UUID validates UUID format (basic check)
	UUID = func(v interface{}) bool {
		s, ok := v.(string)
		if !ok || s == "" {
			return false
		}
		// Basic UUID validation (length and format)
		if len(s) != 36 {
			return false
		}
		return s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
	}

	// PositiveInt validates positive integers
	PositiveInt = func(v interface{}) bool {
		switch n := v.(type) {
		case int:
			return n > 0
		case int64:
			return n > 0
		case float64:
			return n > 0
		default:
			return false
		}
	}

	// NonNegativeInt validates non-negative integers
	NonNegativeInt = func(v interface{}) bool {
		switch n := v.(type) {
		case int:
			return n >= 0
		case int64:
			return n >= 0
		case float64:
			return n >= 0
		default:
			return false
		}
	}
)
