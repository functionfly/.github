// Package apierror provides standardized error handling for the API.
// All handlers should use this package for consistent error responses.
package apierror

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents a machine-readable error code
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeRateLimited      ErrorCode = "RATE_LIMITED"
	ErrCodeUnprocessable    ErrorCode = "UNPROCESSABLE_ENTITY"
	ErrCodeUpgradeRequired  ErrorCode = "UPGRADE_REQUIRED"
	ErrCodeLocked           ErrorCode = "LOCKED"
	ErrCodePrecondition     ErrorCode = "PRECONDITION_FAILED"
	ErrCodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"

	// Server errors (5xx)
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeGatewayTimeout     ErrorCode = "GATEWAY_TIMEOUT"
	ErrCodeNotImplemented     ErrorCode = "NOT_IMPLEMENTED"
	ErrCodeDependencyFailed   ErrorCode = "DEPENDENCY_FAILED"

	// Domain-specific errors
	ErrCodeBilling           ErrorCode = "BILLING_ERROR"
	ErrCodeAuth              ErrorCode = "AUTH_ERROR"
	ErrCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	ErrCodeQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeInvalidState      ErrorCode = "INVALID_STATE"
)

// APIError represents a standardized API error response
type APIError struct {
	// Status is the HTTP status code (not serialized)
	Status int `json:"-"`
	// Code is a machine-readable error code
	Code ErrorCode `json:"code"`
	// Message is a human-readable error message
	Message string `json:"message"`
	// Detail provides additional context (optional, may be omitted in production)
	Detail string `json:"detail,omitempty"`
	// RequestID for tracing (optional)
	RequestID string `json:"request_id,omitempty"`
	// Field is used for validation errors to indicate which field failed
	Field string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// WithDetail adds detail to the error (returns copy for chaining)
func (e *APIError) WithDetail(detail string) *APIError {
	return &APIError{
		Status:    e.Status,
		Code:      e.Code,
		Message:   e.Message,
		Detail:    detail,
		RequestID: e.RequestID,
		Field:     e.Field,
	}
}

// WithField adds field information for validation errors
func (e *APIError) WithField(field string) *APIError {
	return &APIError{
		Status:    e.Status,
		Code:      e.Code,
		Message:   e.Message,
		Detail:    e.Detail,
		RequestID: e.RequestID,
		Field:     field,
	}
}

// WithRequestID adds request ID for tracing
func (e *APIError) WithRequestID(requestID string) *APIError {
	return &APIError{
		Status:    e.Status,
		Code:      e.Code,
		Message:   e.Message,
		Detail:    e.Detail,
		RequestID: requestID,
		Field:     e.Field,
	}
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

// WriteErrorWithStatus writes an error with a specific status code
func WriteErrorWithStatus(w http.ResponseWriter, status int, code ErrorCode, message string) {
	WriteError(w, &APIError{Status: status, Code: code, Message: message})
}

// Common error constructors
func NewBadRequest(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: ErrCodeBadRequest, Message: message}
}

func NewUnauthorized(message string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: ErrCodeUnauthorized, Message: message}
}

func NewForbidden(message string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: ErrCodeForbidden, Message: message}
}

func NewNotFound(message string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: ErrCodeNotFound, Message: message}
}

func NewConflict(message string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: ErrCodeConflict, Message: message}
}

func NewValidation(message string) *APIError {
	return &APIError{Status: http.StatusUnprocessableEntity, Code: ErrCodeValidation, Message: message}
}

func NewRateLimited(message string) *APIError {
	return &APIError{Status: http.StatusTooManyRequests, Code: ErrCodeRateLimited, Message: message}
}

func NewInternal(message string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: ErrCodeInternal, Message: message}
}

func NewServiceUnavailable(message string) *APIError {
	return &APIError{Status: http.StatusServiceUnavailable, Code: ErrCodeServiceUnavailable, Message: message}
}

func NewNotImplemented(message string) *APIError {
	return &APIError{Status: http.StatusNotImplemented, Code: ErrCodeNotImplemented, Message: message}
}

func NewQuotaExceeded(message string) *APIError {
	return &APIError{Status: http.StatusTooManyRequests, Code: ErrCodeQuotaExceeded, Message: message}
}

func NewBillingError(message string) *APIError {
	return &APIError{Status: http.StatusUnprocessableEntity, Code: ErrCodeBilling, Message: message}
}

// ValidationFieldError creates a validation error for a specific field
func ValidationFieldError(field, message string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    ErrCodeValidation,
		Message: message,
		Field:   field,
	}
}

// NewLocked creates a 423 Locked error for locked resources
func NewLocked(message string) *APIError {
	return &APIError{Status: http.StatusLocked, Code: ErrCodeLocked, Message: message}
}

// NewPreconditionFailed creates a 428 Precondition Required error
func NewPreconditionFailed(message string) *APIError {
	return &APIError{Status: http.StatusPreconditionRequired, Code: ErrCodePrecondition, Message: message}
}

// NewTooManyRequests creates a 429 Too Many Requests error
func NewTooManyRequests(message string) *APIError {
	return &APIError{Status: http.StatusTooManyRequests, Code: ErrCodeTooManyRequests, Message: message}
}

// WriteAccepted writes a 202 Accepted response for async operations
func WriteAccepted(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(data)
}

// WriteNoContent writes a 204 No Content response for DELETE operations
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WritePartialContent writes a 206 Partial Content response for pagination
func WritePartialContent(w http.ResponseWriter, data interface{}, contentRange string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Range", contentRange)
	w.WriteHeader(http.StatusPartialContent)
	json.NewEncoder(w).Encode(data)
}
