package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/apierror"
)

// errorResponseWriter wraps http.ResponseWriter to capture responses and
// ensure all error responses use the structured apierror format.
// This normalizes responses from handlers that use http.Error() so that
// all API errors have consistent machine-readable error codes.
type errorResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	wrote      bool
	request    *http.Request
}

// newErrorResponseWriter creates a new error-normalizing response writer
func newErrorResponseWriter(w http.ResponseWriter) *errorResponseWriter {
	return &errorResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// Header returns the response header map
func (rw *errorResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// WriteHeader captures the status code. For error responses, the actual
// normalization happens in Write() to capture any body that http.Error() writes.
func (rw *errorResponseWriter) WriteHeader(code int) {
	if rw.wrote {
		return
	}
	rw.statusCode = code
	// Don't set wrote=true yet for 4xx/5xx - wait for Write() to capture body
	if code < 400 {
		rw.wrote = true
		rw.ResponseWriter.WriteHeader(code)
		if rw.body.Len() > 0 {
			_, _ = rw.ResponseWriter.Write(rw.body.Bytes())
		}
	}
}

// Write captures response body and normalizes error responses
func (rw *errorResponseWriter) Write(b []byte) (int, error) {
	if !rw.wrote && rw.statusCode < 400 {
		// Default to 200 if Write is called without WriteHeader
		rw.statusCode = http.StatusOK
		rw.wrote = true
		rw.ResponseWriter.WriteHeader(http.StatusOK)
		return rw.ResponseWriter.Write(b)
	}

	if !rw.wrote && rw.statusCode >= 400 {
		// First write after error status code - capture body and emit normalized response
		rw.wrote = true
		rw.body.Write(b)

		// Get the request ID from context if available
		var requestID string
		if rw.request != nil {
			requestID = GetRequestID(rw.request.Context())
		}

		// Create structured error response
		errResp := statusCodeToAPIError(rw.statusCode, rw.body.String())
		if requestID != "" {
			errResp = errResp.WithRequestID(requestID)
		}

		// Set the structured response
		rw.ResponseWriter.Header().Set("Content-Type", "application/json")
		rw.ResponseWriter.WriteHeader(rw.statusCode)
		_ = json.NewEncoder(rw.ResponseWriter).Encode(errResp)
		return len(b), nil
	}

	// Normal write
	return rw.ResponseWriter.Write(b)
}

// Flush forwards Flush calls to the underlying ResponseWriter
func (rw *errorResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// ErrorNormalizerMiddleware wraps responses to ensure all errors use structured format.
// Handlers that use http.Error() will have their responses normalized to the
// apierror.APIError JSON format with proper error codes.
func ErrorNormalizerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip normalization for WebSocket upgrades
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip normalization if explicitly disabled
		if os.Getenv("DISABLE_ERROR_NORMALIZER") == "true" {
			next.ServeHTTP(w, r)
			return
		}

		rw := newErrorResponseWriter(w)
		rw.request = r

		next.ServeHTTP(rw, r)
	})
}

// statusCodeToAPIError converts an HTTP status code and optional body to an APIError
func statusCodeToAPIError(status int, body string) *apierror.APIError {
	// Try to parse body as JSON first
	if body != "" {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(body), &parsed); err == nil {
			// If body already has a code field, use it
			if code, ok := parsed["code"].(string); ok {
				message := ""
				if m, ok := parsed["message"].(string); ok {
					message = m
				}
				detail := ""
				if d, ok := parsed["detail"].(string); ok {
					detail = d
				}
				field := ""
				if f, ok := parsed["field"].(string); ok {
					field = f
				}
				return &apierror.APIError{
					Status:  status,
					Code:    apierror.ErrorCode(code),
					Message: message,
					Detail:  detail,
					Field:   field,
				}
			}
		}
	}

	// Generate default error based on status code
	switch status {
	case http.StatusBadRequest:
		return apierror.NewBadRequest(getMessage(body, "Bad request"))
	case http.StatusUnauthorized:
		return apierror.NewUnauthorized(getMessage(body, "Unauthorized"))
	case http.StatusForbidden:
		return apierror.NewForbidden(getMessage(body, "Forbidden"))
	case http.StatusNotFound:
		return apierror.NewNotFound(getMessage(body, "Not found"))
	case http.StatusConflict:
		return apierror.NewConflict(getMessage(body, "Conflict"))
	case http.StatusUnprocessableEntity:
		return apierror.NewValidation(getMessage(body, "Validation error"))
	case http.StatusTooManyRequests:
		return apierror.NewRateLimited(getMessage(body, "Rate limit exceeded"))
	case http.StatusLocked:
		return apierror.NewLocked(getMessage(body, "Resource locked"))
	case http.StatusPreconditionRequired:
		return apierror.NewPreconditionFailed(getMessage(body, "Precondition failed"))
	case http.StatusInternalServerError:
		return apierror.NewInternal(getMessage(body, "Internal server error"))
	case http.StatusServiceUnavailable:
		return apierror.NewServiceUnavailable(getMessage(body, "Service unavailable"))
	case http.StatusNotImplemented:
		return apierror.NewNotImplemented(getMessage(body, "Not implemented"))
	default:
		return &apierror.APIError{
			Status:  status,
			Code:    apierror.ErrorCode("HTTP_" + strconv.Itoa(status)),
			Message: getMessage(body, http.StatusText(status)),
		}
	}
}

// getMessage extracts a message from the body, stripping trailing newlines
func getMessage(body, fallback string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return fallback
	}
	return body
}
