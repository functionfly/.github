// Package flywheel provides HTTP handlers for the Flywheel Network
package flywheel

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler handles Flywheel Network HTTP requests
type Handler struct {
	service *flywheel.Service
	wsHub   *WebSocketHub
	logger  *logrus.Logger
}

// NewHandler creates a new Flywheel handler
func NewHandler(service *flywheel.Service, wsHub *WebSocketHub, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		wsHub:   wsHub,
		logger:  logger,
	}
}

// getUser is a helper to get user from context and handle unauthorized
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) *auth.Claims {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return nil
	}
	return user
}

// parseUUID is a helper to parse a UUID string from URL vars
func (h *Handler) parseUUID(w http.ResponseWriter, r *http.Request, value string, fieldName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		h.logger.WithField(fieldName, value).Warn("Invalid UUID format")
		http.Error(w, `{"error":"Invalid `+fieldName+`"}`, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

// parseUUIDStr is a helper to parse a UUID string without HTTP context
func parseUUIDStr(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
