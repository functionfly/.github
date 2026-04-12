package execution

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/sirupsen/logrus"
)

// writeError writes a structured error response for execution errors
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	errResp := functionregistry.ExecutionError{
		OK: false,
		Error: functionregistry.ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logrus.WithError(err).Error("Failed to encode error response")
	}
}

// toNullString converts a string pointer to sql.NullString
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-Ip header
	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// determineOutcome determines the execution outcome from error and status code
func determineOutcome(err error, statusCode int) (ExecutionOutcome, string) {
	if err == nil && statusCode >= 200 && statusCode < 300 {
		return OutcomeSuccess, ""
	}

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") {
			return OutcomeTimeout, "TIMEOUT"
		}
		if strings.Contains(errStr, "memory") || strings.Contains(errStr, "OOM") {
			return OutcomeOOM, "OUT_OF_MEMORY"
		}
		if strings.Contains(errStr, "cancelled") || strings.Contains(errStr, "canceled") {
			return OutcomeCancelled, "CANCELLED"
		}
	}

	return OutcomeError, "RUNTIME_ERROR"
}
