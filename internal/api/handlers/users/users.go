package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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

// HandleGetPublicProfile returns a user's public profile by username.
// This endpoint is public — it never exposes email, tenantId, role, or any sensitive fields.
func (h *Handler) HandleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("Failed to get user by username")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Build safe public profile — never expose email, tenantId, role, password, MFA, etc.
	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}

	var avatar string
	if user.ProviderData != nil {
		if a, ok := user.ProviderData["avatar_url"].(string); ok {
			avatar = a
		}
	}

	var bio string
	if user.Bio != nil {
		bio = *user.Bio
	}

	usernameStr := ""
	if user.Username != nil {
		usernameStr = *user.Username
	}

	// Fetch published functions from registry
	publishedFunctions := h.getPublishedFunctions(usernameStr)

	profile := map[string]interface{}{
		"id":                 user.ID,
		"username":           usernameStr,
		"name":               name,
		"avatar":             avatar,
		"bio":                bio,
		"createdAt":          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"publishedFunctions": publishedFunctions,
	}

	writeJSON(w, http.StatusOK, profile)
}

// HandleGetMe returns the current authenticated user's profile
func (h *Handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to get user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}

	var avatar string
	if user.ProviderData != nil {
		if a, ok := user.ProviderData["avatar_url"].(string); ok {
			avatar = a
		}
	}

	usernameStr := ""
	if user.Username != nil {
		usernameStr = *user.Username
	}

	companyName := ""
	if user.CompanyName != nil {
		companyName = *user.CompanyName
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"name":        name,
		"username":    usernameStr,
		"companyName": companyName,
		"avatar":      avatar,
		"updatedAt":   user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateMeRequest represents the request body for updating the current user's profile
type UpdateMeRequest struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	CompanyName string `json:"companyName"`
	Bio         string `json:"bio"`
}

// HandleUpdateMe updates the current authenticated user's profile
func (h *Handler) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Username != "" {
		// Validate username format: lowercase alphanumeric, hyphens, underscores
		clean := strings.ToLower(strings.TrimSpace(req.Username))
		for _, c := range clean {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				writeJSONError(w, http.StatusBadRequest, "Username may only contain lowercase letters, numbers, hyphens, and underscores")
				return
			}
		}
		if len(clean) < 3 {
			writeJSONError(w, http.StatusBadRequest, "Username must be at least 3 characters")
			return
		}
		if len(clean) > 30 {
			writeJSONError(w, http.StatusBadRequest, "Username must be 30 characters or fewer")
			return
		}
		updates["username"] = clean
	}
	if req.CompanyName != "" {
		updates["company_name"] = strings.TrimSpace(req.CompanyName)
	}
	if req.Bio != "" {
		updates["bio"] = strings.TrimSpace(req.Bio)
	}

	if len(updates) == 0 {
		writeJSONError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updatedUser, err := h.repo.UpdateUser(context.Background(), claims.UserID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeJSONError(w, http.StatusConflict, "Username is already taken")
			return
		}
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	name := updatedUser.Name
	if name == "" && updatedUser.ProviderData != nil {
		if n, ok := updatedUser.ProviderData["name"].(string); ok {
			name = n
		}
	}

	usernameStr := ""
	if updatedUser.Username != nil {
		usernameStr = *updatedUser.Username
	}

	companyName := ""
	if updatedUser.CompanyName != nil {
		companyName = *updatedUser.CompanyName
	}

	var avatar string
	if updatedUser.ProviderData != nil {
		if a, ok := updatedUser.ProviderData["avatar_url"].(string); ok {
			avatar = a
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Profile updated successfully",
		"user": map[string]interface{}{
			"id":          updatedUser.ID,
			"name":        name,
			"username":    usernameStr,
			"companyName": companyName,
			"email":       updatedUser.Email,
			"avatar":      avatar,
			"updatedAt":   updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
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
