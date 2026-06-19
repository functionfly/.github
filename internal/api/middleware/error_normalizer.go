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

// statusCodeToAPIError converts an HTTP status code and optional body to an APIError.
// For 5xx responses, the body is treated as untrusted (often raw err.Error() text)
// and a generic message is always returned. For 4xx responses, the body is usually
// a hand-written client-visible message and is passed through unless it looks
// like raw error output.
func statusCodeToAPIError(status int, body string) *apierror.APIError {
	body = strings.TrimSpace(body)

	// Try to parse body as JSON first - this is the trusted path
	if body != "" {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(body), &parsed); err == nil {
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
				if status >= 500 && message != "" {
					message = genericMessageForStatus(status)
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

	// 5xx: body is untrusted err text - always use generic message
	if status >= 500 {
		return genericAPIError(status)
	}

	// 4xx: body is usually a hand-written client-visible message.
	// Pass through, but sanitize if it looks like raw err output.
	safeBody := sanitizeClientMessage(body)
	switch status {
	case http.StatusBadRequest:
		return apierror.NewBadRequest(orDefault(safeBody, "Bad request"))
	case http.StatusUnauthorized:
		return apierror.NewUnauthorized(orDefault(safeBody, "Unauthorized"))
	case http.StatusForbidden:
		return apierror.NewForbidden(orDefault(safeBody, "Forbidden"))
	case http.StatusNotFound:
		return apierror.NewNotFound(orDefault(safeBody, "Not found"))
	case http.StatusConflict:
		return apierror.NewConflict(orDefault(safeBody, "Conflict"))
	case http.StatusUnprocessableEntity:
		return apierror.NewValidation(orDefault(safeBody, "Validation error"))
	case http.StatusTooManyRequests:
		return apierror.NewRateLimited(orDefault(safeBody, "Rate limit exceeded"))
	case http.StatusLocked:
		return apierror.NewLocked(orDefault(safeBody, "Resource locked"))
	case http.StatusPreconditionRequired:
		return apierror.NewPreconditionFailed(orDefault(safeBody, "Precondition failed"))
	default:
		return &apierror.APIError{
			Status:  status,
			Code:    apierror.ErrorCode("HTTP_" + strconv.Itoa(status)),
			Message: orDefault(safeBody, http.StatusText(status)),
		}
	}
}

// genericAPIError returns a generic APIError for a 5xx status code.
func genericAPIError(status int) *apierror.APIError {
	switch status {
	case http.StatusInternalServerError:
		return apierror.NewInternal("Internal server error")
	case http.StatusBadGateway:
		return apierror.NewInternal("Bad gateway")
	case http.StatusServiceUnavailable:
		return apierror.NewServiceUnavailable("Service unavailable")
	case http.StatusGatewayTimeout:
		return &apierror.APIError{Status: status, Code: apierror.ErrCodeGatewayTimeout, Message: "Gateway timeout"}
	default:
		return &apierror.APIError{
			Status:  status,
			Code:    apierror.ErrorCode("HTTP_" + strconv.Itoa(status)),
			Message: genericMessageForStatus(status),
		}
	}
}

// genericMessageForStatus returns a generic human-readable message for a 5xx status.
func genericMessageForStatus(status int) string {
	switch status {
	case http.StatusInternalServerError:
		return "Internal server error"
	case http.StatusBadGateway:
		return "Bad gateway"
	case http.StatusServiceUnavailable:
		return "Service unavailable"
	case http.StatusGatewayTimeout:
		return "Gateway timeout"
	case http.StatusNotImplemented:
		return "Not implemented"
	default:
		return http.StatusText(status)
	}
}

// errLeakPrefixes are substrings that indicate the body is a raw error message
// (e.g. SQL driver output, Go panic output, JSON parse errors) that must not
// be forwarded to the client.
var errLeakPrefixes = []string{
	"pq:",
	"sql:",
	"pgx:",
	"json:",
	"yaml:",
	"xml:",
	"goroutine ",
	"panic:",
	"runtime error:",
	"stack overflow",
	"connection refused",
	"dial tcp",
	"no such host",
	"tls:",
	"x509:",
	"context deadline exceeded",
	"context canceled",
	"invalid character",
	"unexpected end of JSON",
	"unmarshal",
	"redis:",
	"nats:",
	"kafka:",
	"s3:",
	"r2:",
	"open ",
	"read ",
	"write ",
	"stat ",
	"/Users/",
	"/home/",
	"/var/",
	"/etc/",
	"\\Users\\",
	"\\home\\",
}

// sanitizeClientMessage returns the body if it looks like a safe, hand-written
// client-visible message; otherwise returns an empty string so the caller falls
// back to a generic message.
func sanitizeClientMessage(body string) string {
	if body == "" {
		return ""
	}
	lower := strings.ToLower(body)
	for _, prefix := range errLeakPrefixes {
		if strings.Contains(lower, strings.ToLower(prefix)) {
			return ""
		}
	}
	if len(body) > 200 {
		return ""
	}
	return body
}

// orDefault returns s if non-empty, otherwise fallback.
func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

