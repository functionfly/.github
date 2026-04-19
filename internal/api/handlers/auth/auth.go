package auth

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
)

// Handler contains authentication handlers
type Handler struct {
	authSvc *auth.AuthService
}

// NewHandler creates a new auth handler
func NewHandler(authSvc *auth.AuthService) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeJSONError writes a JSON error body and status code so the frontend can parse it
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONErrorDetail(w, status, message, "")
}

// writeJSONErrorDetail writes a JSON error with optional detail (e.g. root cause for debugging)
func writeJSONErrorDetail(w http.ResponseWriter, status int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]string{"message": message}
	if detail != "" {
		body["detail"] = detail
	}
	_ = json.NewEncoder(w).Encode(body)
}

// isLoginInternalError returns true if the error is a server-side failure (DB, token, etc.) rather than bad credentials.
func isLoginInternalError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if msg == "invalid credentials" || strings.Contains(msg, "invalid credentials") {
		return false
	}
	if strings.Contains(msg, "sql: no rows in result set") {
		return false
	}
	if strings.Contains(msg, "email not verified") {
		return false
	}
	if msg == "internal error" {
		return true
	}
	if strings.Contains(msg, "failed to generate token") || strings.Contains(msg, "failed to verify password") {
		return true
	}
	if strings.Contains(msg, "failed to get user:") {
		return true
	}
	return false
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// getClientIP extracts the real client IP address from the request
func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
