package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
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

// HandleGetPublicProfileByAt returns a user's public profile by @username.
// This endpoint is public — it never exposes email, tenantId, role, or any sensitive fields.
// This is the new clean URL format: /@/username
func (h *Handler) HandleGetPublicProfileByAt(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	// Remove @ prefix if present
	if strings.HasPrefix(username, "@") {
		username = strings.TrimPrefix(username, "@")
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

	// Build SEO-enhanced profile with additional metadata
	profile := map[string]interface{}{
		"id":                 user.ID,
		"username":           usernameStr,
		"name":               name,
		"avatar":             avatar,
		"bio":                bio,
		"createdAt":          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"publishedFunctions": publishedFunctions,
		// SEO enhancement fields
		"profileUrl":     "/@" + usernameStr,
		"totalFunctions": len(publishedFunctions),
	}

	// Add verification fields if available
	if user.EmailVerified {
		profile["isVerified"] = true
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

	// Load tenant plan for billing/UI (authoritative source)
	plan := ""
	if tenant, err := h.repo.GetTenantByID(user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
		plan = tenant.Plan
	}

	resp := map[string]interface{}{
		"id":          user.ID,
		"tenantId":    user.TenantID,
		"email":       user.Email,
		"name":        name,
		"username":    usernameStr,
		"companyName": companyName,
		"avatar":      avatar,
		"plan":        plan,
		"updatedAt":   user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	writeJSON(w, http.StatusOK, resp)
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

// sessionResponseItem is the safe session payload returned to the client (no token).
type sessionResponseItem struct {
	ID             string `json:"id"`
	Device         string `json:"device"`
	IP             string `json:"ip"`
	Location       string `json:"location"`
	LastActive     string `json:"lastActive"`
	CurrentSession bool   `json:"currentSession"`
}

func parseUserAgent(ua string) string {
	if ua == "" {
		return "Unknown device"
	}
	// Simple heuristic: look for known browsers/OS
	if strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Edg") {
		return "Chrome"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
		return "Safari"
	}
	if strings.Contains(ua, "Edg") {
		return "Edge"
	}
	if strings.Contains(ua, "Mac") {
		return "Desktop (macOS)"
	}
	if strings.Contains(ua, "Windows") {
		return "Desktop (Windows)"
	}
	if strings.Contains(ua, "Linux") {
		return "Desktop (Linux)"
	}
	if strings.Contains(ua, "iPhone") || strings.Contains(ua, "Android") {
		return "Mobile"
	}
	return "Unknown device"
}

func formatLastActive(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "Now"
	}
	if diff < time.Hour {
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if diff < 24*time.Hour {
		hr := int(diff.Hours())
		if hr == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hr)
	}
	d := int(diff.Hours() / 24)
	if d == 1 {
		return "1 day ago"
	}
	if d < 7 {
		return fmt.Sprintf("%d days ago", d)
	}
	return t.Format("2006-01-02")
}

// HandleListSessions returns GET /v1/users/me/sessions - list active sessions for the current user
func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var currentSessionID uuid.UUID
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			if sess, err := h.repo.GetSessionByToken(parts[1]); err == nil && sess != nil {
				currentSessionID = sess.ID
			}
		}
	}

	sessions, err := h.repo.ListUserSessions(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to list sessions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load sessions")
		return
	}

	items := make([]sessionResponseItem, 0, len(sessions))
	for _, s := range sessions {
		lastActive := formatLastActive(s.LastActivity)
		items = append(items, sessionResponseItem{
			ID:             s.ID.String(),
			Device:         parseUserAgent(s.UserAgent),
			IP:             s.IPAddress,
			Location:       "", // optional: derive from IP later
			LastActive:     lastActive,
			CurrentSession: s.ID == currentSessionID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": items})
}

// HandleRevokeSession handles DELETE /v1/users/me/sessions/{id} - revoke one session
func (h *Handler) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sessionIDStr := vars["id"]
	if sessionIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "session id required")
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := h.repo.DeleteSessionByID(sessionID, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			writeJSONError(w, http.StatusNotFound, "Session not found")
			return
		}
		logrus.WithError(err).WithField("sessionID", sessionID).Error("Failed to revoke session")
		writeJSONError(w, http.StatusInternalServerError, "Failed to revoke session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Session revoked"})
}

// HandleRevokeOtherSessions handles POST /v1/users/me/sessions/revoke-others - revoke all other sessions
func (h *Handler) HandleRevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var currentSessionID uuid.UUID
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			if sess, err := h.repo.GetSessionByToken(parts[1]); err == nil && sess != nil {
				currentSessionID = sess.ID
			}
		}
	}

	sessions, err := h.repo.ListUserSessions(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to list sessions for revoke-others")
		writeJSONError(w, http.StatusInternalServerError, "Failed to revoke sessions")
		return
	}

	for _, s := range sessions {
		if s.ID != currentSessionID {
			_ = h.repo.DeleteSessionByID(s.ID, claims.UserID)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "All other sessions revoked"})
}

// requireSelfUsername ensures the request is authenticated and the path username matches the current user. Returns the user and true on success; on failure it writes the error response and returns nil, false.
func (h *Handler) requireSelfUsername(w http.ResponseWriter, r *http.Request, pathUsername string) (*storage.User, bool) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return nil, false
	}
	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return nil, false
	}
	usernameStr := ""
	if user.Username != nil {
		usernameStr = *user.Username
	}
	if pathUsername == "" || !strings.EqualFold(pathUsername, usernameStr) {
		writeJSONError(w, http.StatusForbidden, "You can only access your own settings")
		return nil, false
	}
	return user, true
}

// HandleGetUserSettings returns GET /v1/users/{username}/settings — full settings payload for the current user (username must match).
func (h *Handler) HandleGetUserSettings(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	user, ok := h.requireSelfUsername(w, r, pathUsername)
	if !ok {
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
	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	payload := map[string]interface{}{
		"id":       user.ID.String(),
		"username": usernameStr,
		"name":     name,
		"email":    user.Email,
		"avatar":   avatar,
		"bio":      bio,
		"website":  "",
		"twitter":  "",
		"github":   "",
		"settings": map[string]interface{}{
			"emailNotifications": true,
			"marketingEmails":    false,
			"publicProfile":      true,
			"allowMessaging":     false,
		},
	}
	if user.CreatedAt.IsZero() == false {
		payload["createdAt"] = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	writeJSON(w, http.StatusOK, payload)
}

// SettingsProfilePatchRequest is the body for PATCH /v1/users/{username}/settings/profile
type SettingsProfilePatchRequest struct {
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Website  string `json:"website"`
	Twitter  string `json:"twitter"`
	Github   string `json:"github"`
	Username string `json:"username"`
}

// HandlePatchUserSettingsProfile handles PATCH /v1/users/{username}/settings/profile
func (h *Handler) HandlePatchUserSettingsProfile(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	_, ok := h.requireSelfUsername(w, r, pathUsername)
	if !ok {
		return
	}
	claims := middleware.GetUserFromContext(r)

	var req SettingsProfilePatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Bio != "" {
		updates["bio"] = strings.TrimSpace(req.Bio)
	}
	if req.Username != "" {
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

	if len(updates) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"message": "No changes to save"})
		return
	}

	_, err := h.repo.UpdateUser(context.Background(), claims.UserID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeJSONError(w, http.StatusConflict, "Username is already taken")
			return
		}
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update profile settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Profile updated"})
}

// HandlePatchUserSettingsNotifications handles PATCH /v1/users/{username}/settings/notifications (no-op for now)
func (h *Handler) HandlePatchUserSettingsNotifications(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	if _, ok := h.requireSelfUsername(w, r, pathUsername); !ok {
		return
	}
	// Accept body for future use; no persistence yet
	_ = json.NewDecoder(r.Body)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Notification preferences updated"})
}

// HandlePatchUserSettingsPrivacy handles PATCH /v1/users/{username}/settings/privacy (no-op for now)
func (h *Handler) HandlePatchUserSettingsPrivacy(w http.ResponseWriter, r *http.Request) {
	pathUsername := mux.Vars(r)["username"]
	if _, ok := h.requireSelfUsername(w, r, pathUsername); !ok {
		return
	}
	// Accept body for future use; no persistence yet
	_ = json.NewDecoder(r.Body)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Privacy settings updated"})
}
