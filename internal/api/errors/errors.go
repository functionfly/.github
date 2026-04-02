package apierrors

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// safeMessages maps HTTP status codes to generic client-safe messages.
// Internal error details are NEVER sent to the client.
var safeMessages = map[int]string{
	http.StatusBadRequest:          "Invalid request",
	http.StatusUnauthorized:        "Authentication required",
	http.StatusForbidden:           "Access denied",
	http.StatusNotFound:            "Resource not found",
	http.StatusConflict:            "Resource conflict",
	http.StatusTooManyRequests:     "Too many requests",
	http.StatusInternalServerError: "Internal server error",
	http.StatusServiceUnavailable:  "Service unavailable",
	http.StatusPaymentRequired:     "Payment required",
}

// WriteJSONError writes a generic JSON error response. The internal error is
// logged server-side but NEVER sent to the client.
func WriteJSONError(w http.ResponseWriter, status int, internalErr error, clientMsg string) {
	if internalErr != nil {
		logrus.WithError(internalErr).WithField("status", status).Error("API error")
	}

	msg := clientMsg
	if msg == "" {
		if m, ok := safeMessages[status]; ok {
			msg = m
		} else {
			msg = "An error occurred"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

// WriteJSONErrorWithCode writes a JSON error response with an error code for programmatic handling.
func WriteJSONErrorWithCode(w http.ResponseWriter, status int, internalErr error, clientMsg, code string) {
	if internalErr != nil {
		logrus.WithError(internalErr).WithFields(logrus.Fields{
			"status": status,
			"code":   code,
		}).Error("API error")
	}

	msg := clientMsg
	if msg == "" {
		if m, ok := safeMessages[status]; ok {
			msg = m
		} else {
			msg = "An error occurred"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
		"code":  code,
	})
}
