package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// RegistryQuerier is an optional interface for querying published functions by author.
// The PostgresDB implements this if the registry repository is available.
type RegistryQuerier interface {
	QueryPublishedFunctionsByAuthor(author string) ([]map[string]interface{}, error)
}

// Handler contains user-related HTTP handlers
type Handler struct {
	repo    storage.Repository
	authSvc *auth.AuthService
}

// NewHandler creates a new users handler
func NewHandler(repo storage.Repository, authSvc *auth.AuthService) *Handler {
	return &Handler{repo: repo, authSvc: authSvc}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// boolSetting returns the boolean value for a key in user settings (e.g. showLocation).
// Defaults to true if missing or not a bool so existing users keep visible fields.
func boolSetting(settings map[string]interface{}, key string) bool {
	if settings == nil {
		return true
	}
	v, ok := settings[key]
	if !ok {
		return true
	}
	b, ok := v.(bool)
	if !ok {
		return true
	}
	return b
}

// applyProfileVisibility overwrites profile fields (location, companyName, jobTitle) to empty when the profile owner's visibility settings hide them.
func (h *Handler) applyProfileVisibility(ctx context.Context, profile map[string]interface{}, userID uuid.UUID) {
	settings, err := h.repo.GetUserSettings(ctx, userID)
	if err != nil || settings == nil {
		return
	}
	if !boolSetting(settings, "showLocation") {
		profile["location"] = ""
	}
	if !boolSetting(settings, "showCompany") {
		profile["companyName"] = ""
		profile["jobTitle"] = ""
	}
	if !boolSetting(settings, "showFounderBadge") {
		profile["founderNumber"] = 0
	}
}

// getPublishedFunctions fetches published registry functions for a given author/username.
// Uses the optional RegistryQuerier interface; returns an empty list if not available.
func (h *Handler) getPublishedFunctions(username string) []map[string]interface{} {
	if username == "" {
		return []map[string]interface{}{}
	}

	querier, ok := h.repo.(RegistryQuerier)
	if !ok {
		return []map[string]interface{}{}
	}

	rows, err := querier.QueryPublishedFunctionsByAuthor(username)
	if err != nil || rows == nil {
		return []map[string]interface{}{}
	}
	return rows
}

// requireSelfUsername ensures the request is authenticated and the path username matches the current user. Returns the user and true on success; on failure it writes the error response and returns nil, false.
func (h *Handler) requireSelfUsername(w http.ResponseWriter, r *http.Request, pathUsername string) (*storage.User, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return nil, false
	}
	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return nil, false
	}
	usernameStr := ""
	if user.Username != nil {
		usernameStr = *user.Username
	}
	if pathUsername == "" || !strings.EqualFold(pathUsername, usernameStr) {
		apierror.WriteError(w, apierror.NewForbidden("You can only access your own settings"))
		return nil, false
	}
	return user, true
}
