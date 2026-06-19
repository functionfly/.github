// Package apierror logging helpers.
//
// These helpers close the most common error-handling bug in the API:
// returning err.Error() text directly to clients. They log the full error
// server-side (with request id, handler context, and stack-relevant fields)
// and emit a sanitized, generic apierror.APIError response to the caller.
//
// Use them in place of `http.Error(w, err.Error(), ...)`,
// `apierror.WriteError(w, apierror.NewInternal(err.Error()))`, and any
// local-package `writeError/respondError/writeJSONError` helper that
// forwards err.Error() to the client.
package apierror

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// logAPIError emits a structured server-side log line for an error and
// returns the appropriate APIError to write to the client.
//
// `status` is the HTTP status to return. `ctx` is a short, handler-scoped
// description (e.g. "create user", "list agents") included in the log so
// operators can correlate the user-visible response to the server log.
//
// The returned APIError always has a generic message. The original err
// is never forwarded to the client.
func logAPIError(r *http.Request, err error, status int, code ErrorCode, ctx string) *APIError {
	requestID := requestIDFromContext(r)

	fields := logrus.Fields{
		"status":     status,
		"code":       string(code),
		"context":    ctx,
		"method":     "",
		"path":       "",
		"request_id": requestID,
	}
	if r != nil {
		fields["method"] = r.Method
		if r.URL != nil {
			fields["path"] = r.URL.Path
		}
	}

	entry := logrus.WithError(err).WithFields(fields)

	switch {
	case status >= 500:
		entry.Error("server error")
	case status == http.StatusTooManyRequests, status == http.StatusLocked:
		entry.Warn("client throttled")
	case status >= 400:
		entry.Info("client error")
	default:
		entry.Info("handled error")
	}

	return &APIError{
		Status:  status,
		Code:    code,
		Message: genericMessageForStatusCode(status, code),
	}
}

// genericMessageForStatusCode returns a safe, generic message for a status
// and code pair. The message never contains err text or other untrusted input.
func genericMessageForStatusCode(status int, code ErrorCode) string {
	if msg, ok := genericCodeMessages[code]; ok {
		return msg
	}
	switch status {
	case http.StatusBadRequest:
		return "Bad request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusUnprocessableEntity:
		return "Validation error"
	case http.StatusTooManyRequests:
		return "Too many requests"
	case http.StatusLocked:
		return "Resource locked"
	case http.StatusPreconditionRequired:
		return "Precondition failed"
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

var genericCodeMessages = map[ErrorCode]string{
	ErrCodeBadRequest:         "Bad request",
	ErrCodeUnauthorized:       "Unauthorized",
	ErrCodeForbidden:          "Forbidden",
	ErrCodeNotFound:           "Not found",
	ErrCodeConflict:           "Conflict",
	ErrCodeValidation:         "Validation error",
	ErrCodeRateLimited:        "Too many requests",
	ErrCodeUnprocessable:      "Unprocessable entity",
	ErrCodeUpgradeRequired:    "Upgrade required",
	ErrCodeLocked:             "Resource locked",
	ErrCodePrecondition:       "Precondition failed",
	ErrCodeTooManyRequests:    "Too many requests",
	ErrCodeInternal:           "Internal server error",
	ErrCodeServiceUnavailable: "Service unavailable",
	ErrCodeGatewayTimeout:     "Gateway timeout",
	ErrCodeNotImplemented:     "Not implemented",
	ErrCodeDependencyFailed:   "Upstream dependency failed",
	ErrCodeBilling:            "Billing error",
	ErrCodeAuth:               "Authentication error",
	ErrCodeResourceExhausted:  "Resource exhausted",
	ErrCodeQuotaExceeded:      "Quota exceeded",
	ErrCodeInvalidState:       "Invalid state",
}

// requestIDFromContext extracts the request id from a request, if available.
// Falls back to "-" if no id is set, so log aggregation tools always have a value.
//
// We use the X-Request-ID header which is the standard convention used by
// the request-id middleware. We cannot look up the typed RequestIDKey from
// the middleware package (it is unexported) without creating a circular
// dependency, and the header is always set by the middleware on the way
// through, so the lookup is equivalent in practice.
func requestIDFromContext(r *http.Request) string {
	if r == nil {
		return "-"
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return "-"
}

// ── Public helpers ──────────────────────────────────────────────────────────

// LogAndInternal logs err server-side (with context + request id) and writes
// a generic 500 response to the client. Use this for any 5xx error where
// the underlying err is from a dependency (DB, network, third-party API,
// or any internal subsystem that may include sensitive details in err text).
//
// Replaces:
//
//	http.Error(w, err.Error(), http.StatusInternalServerError)
//	apierror.WriteError(w, apierror.NewInternal(err.Error()))
//	apierror.WriteError(w, apierror.NewInternal("ctx: "+err.Error()))
//	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
func LogAndInternal(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusInternalServerError, ErrCodeInternal, contextMsg)
	WriteError(w, apiErr)
}

// LogAndBadRequest logs err and writes a generic 400 response. Use when the
// err represents a client-side input problem (e.g. JSON unmarshal failure)
// where leaking the parse error text could help an attacker probe the API.
func LogAndBadRequest(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusBadRequest, ErrCodeBadRequest, contextMsg)
	WriteError(w, apiErr)
}

// LogAndNotFound logs err and writes a generic 404 response. Use when the
// underlying err from a lookup may contain sensitive details (e.g. raw SQL).
func LogAndNotFound(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusNotFound, ErrCodeNotFound, contextMsg)
	WriteError(w, apiErr)
}

// LogAndConflict logs err and writes a generic 409 response.
func LogAndConflict(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusConflict, ErrCodeConflict, contextMsg)
	WriteError(w, apiErr)
}

// LogAndForbidden logs err and writes a generic 403 response.
func LogAndForbidden(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusForbidden, ErrCodeForbidden, contextMsg)
	WriteError(w, apiErr)
}

// LogAndUnauthorized logs err and writes a generic 401 response.
func LogAndUnauthorized(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusUnauthorized, ErrCodeUnauthorized, contextMsg)
	WriteError(w, apiErr)
}

// LogAndServiceUnavailable logs err and writes a generic 503 response.
func LogAndServiceUnavailable(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusServiceUnavailable, ErrCodeServiceUnavailable, contextMsg)
	WriteError(w, apiErr)
}

// LogAndUnprocessable logs err and writes a generic 422 response.
func LogAndUnprocessable(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusUnprocessableEntity, ErrCodeValidation, contextMsg)
	WriteError(w, apiErr)
}

// LogAndGatewayTimeout logs err and writes a generic 504 response.
func LogAndGatewayTimeout(w http.ResponseWriter, r *http.Request, err error, contextMsg string) {
	apiErr := logAPIError(r, err, http.StatusGatewayTimeout, ErrCodeGatewayTimeout, contextMsg)
	WriteError(w, apiErr)
}

// LogAnd writes a server-side log line for err and returns the appropriate
// APIError. The caller is responsible for writing the response. Useful when
// the caller wants to add extra context (e.g. a Field for validation errors)
// before emitting the response.
func LogAnd(r *http.Request, err error, status int, code ErrorCode, contextMsg string) *APIError {
	return logAPIError(r, err, status, code, contextMsg)
}

// FromError inspects err and returns the best-fit *APIError. The returned
// APIError always uses a generic Message; the caller is expected to log the
// original err via LogAndInternal (or one of its siblings) before writing
// the response.
//
// This is a convenience mapper. It does not log the error itself. Handlers
// that need to log the err should call LogAndInternal (or appropriate sibling)
// directly. FromError is useful when a handler wants to add extra context
// (e.g. a Field) to the APIError before writing it.
func FromError(err error) *APIError {
	if err == nil {
		return NewInternal("Internal server error")
	}
	if e, ok := err.(*APIError); ok && e != nil {
		sanitized := &APIError{
			Status:    e.Status,
			Code:      e.Code,
			Message:   genericMessageForStatusCode(e.Status, e.Code),
			Field:     e.Field,
			RequestID: e.RequestID,
		}
		return sanitized
	}
	return NewInternal("Internal server error")
}

// SanitizeMessage returns a generic, status-appropriate message. If `body`
// looks like a safe, hand-written client-visible message (no err-driver
// prefixes, no file paths, no panic stacks, length under 200), it is
// returned unchanged. Otherwise the appropriate generic fallback is
// returned. Used by the error normalizer middleware to decide whether a
// pre-existing body should be trusted or replaced.
func SanitizeMessage(status int, body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return genericMessageForStatus(status)
	}
	if status >= 500 {
		return genericMessageForStatus(status)
	}
	if looksLikeErrorText(body) {
		return genericMessageForStatus(status)
	}
	return body
}

func genericMessageForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "Bad request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusUnprocessableEntity:
		return "Validation error"
	case http.StatusTooManyRequests:
		return "Too many requests"
	case http.StatusLocked:
		return "Resource locked"
	case http.StatusPreconditionRequired:
		return "Precondition failed"
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

// looksLikeErrorText returns true if body contains substrings typical of
// raw err output (SQL drivers, panic stacks, file paths, etc.) or is
// suspiciously long.
func looksLikeErrorText(body string) bool {
	if len(body) > 200 {
		return true
	}
	lower := strings.ToLower(body)
	for _, prefix := range errLeakPrefixes {
		if strings.Contains(lower, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// errLeakPrefixes is the shared allowlist of substrings that indicate the
// body is raw error output that must not be forwarded to the client.
// Mirrors the list in middleware/error_normalizer.go.
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
	"unexpected end of json",
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
	"/users/",
	"/home/",
	"/var/",
	"/etc/",
	`\users\`,
	`\home\`,
	"exception:",
	"traceback",
	"stacktrace",
}

// Format is a convenience for handlers that want to include err text in a
// log line but not in the response. Equivalent to fmt.Sprintf but exists
// as a stable symbol so future log-format changes are localized.
func Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
