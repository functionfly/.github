package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

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

	// Get user settings from database
	settings, err := h.repo.GetUserSettings(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get user settings, using defaults")
		settings = getDefaultSettings()
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
		"settings": settings,
	}
	if user.CreatedAt.IsZero() == false {
		payload["createdAt"] = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	writeJSON(w, http.StatusOK, payload)
}

// getDefaultSettings returns the default profile settings
func getDefaultSettings() map[string]interface{} {
	return map[string]interface{}{
		"profileVisibility":     "public",
		"showEmail":             false,
		"showLocation":          true,
		"showCompany":           true,
		"showActivity":          true,
		"showAnalytics":         true,
		"emailNotifications":    true,
		"pushNotifications":     false,
		"notifyOnFollow":        true,
		"notifyOnMention":       true,
		"notifyOnFunctionUsage": true,
		"notifyOnReviews":       true,
		"weeklyDigest":          true,
		"allowTagging":          true,
		"allowIndexing":         true,
		"showLastActive":        true,
		"deploymentSuccess":     true,
		"deploymentFailure":     true,
		"failoverEvents":        true,
		"providerIssues":        true,
		"active_environment":    "production",
	}
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

// HandlePatchUserSettingsNotifications handles PATCH /v1/users/{username}/settings/notifications
func (h *Handler) HandlePatchUserSettingsNotifications(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update notification-related fields
	notificationFields := []string{
		"emailNotifications",
		"pushNotifications",
		"notifyOnFollow",
		"notifyOnMention",
		"notifyOnFunctionUsage",
		"notifyOnReviews",
		"weeklyDigest",
		"deploymentSuccess",
		"deploymentFailure",
		"failoverEvents",
		"providerIssues",
	}

	for _, field := range notificationFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update notification settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update notification settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Notification preferences updated"})
}

// HandlePatchUserSettingsPrivacy handles PATCH /v1/users/{username}/settings/privacy
func (h *Handler) HandlePatchUserSettingsPrivacy(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update privacy-related fields
	privacyFields := []string{
		"allowTagging",
		"allowIndexing",
		"showLastActive",
	}

	for _, field := range privacyFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update privacy settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update privacy settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Privacy settings updated"})
}

// HandlePatchUserSettingsVisibility handles PATCH /v1/users/{username}/settings/visibility
func (h *Handler) HandlePatchUserSettingsVisibility(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update visibility-related fields
	visibilityFields := []string{
		"profileVisibility",
		"showEmail",
		"showLocation",
		"showCompany",
		"showActivity",
		"showAnalytics",
	}

	for _, field := range visibilityFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update visibility settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update visibility settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Visibility settings updated"})
}

// HandleGetUserSettingsMe returns GET /v1/users/me/settings — full settings payload for the current authenticated user
func (h *Handler) HandleGetUserSettingsMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to get user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Get the username for the response
	usernameStr := ""
	if user.Username != nil {
		usernameStr = *user.Username
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
	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	// Get user settings from database
	settings, err := h.repo.GetUserSettings(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get user settings, using defaults")
		settings = getDefaultSettings()
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
		"settings": settings,
	}
	if user.CreatedAt.IsZero() == false {
		payload["createdAt"] = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	writeJSON(w, http.StatusOK, payload)
}

// HandlePatchUserSettingsProfileMe handles PATCH /v1/users/me/settings/profile
func (h *Handler) HandlePatchUserSettingsProfileMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

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

// HandlePatchUserSettingsNotificationsMe handles PATCH /v1/users/me/settings/notifications
func (h *Handler) HandlePatchUserSettingsNotificationsMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update notification-related fields
	notificationFields := []string{
		"emailNotifications",
		"pushNotifications",
		"notifyOnFollow",
		"notifyOnMention",
		"notifyOnFunctionUsage",
		"notifyOnReviews",
		"weeklyDigest",
		"deploymentSuccess",
		"deploymentFailure",
		"failoverEvents",
		"providerIssues",
	}

	for _, field := range notificationFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update notification settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update notification settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Notification preferences updated"})
}

// HandlePatchUserSettingsPrivacyMe handles PATCH /v1/users/me/settings/privacy
func (h *Handler) HandlePatchUserSettingsPrivacyMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update privacy-related fields
	privacyFields := []string{
		"allowTagging",
		"allowIndexing",
		"showLastActive",
	}

	for _, field := range privacyFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update privacy settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update privacy settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Privacy settings updated"})
}

// HandlePatchUserSettingsVisibilityMe handles PATCH /v1/users/me/settings/visibility
func (h *Handler) HandlePatchUserSettingsVisibilityMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update visibility-related fields
	visibilityFields := []string{
		"profileVisibility",
		"showEmail",
		"showLocation",
		"showCompany",
		"showActivity",
		"showAnalytics",
	}

	for _, field := range visibilityFields {
		if val, ok := req[field]; ok {
			currentSettings[field] = val
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update visibility settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update visibility settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Visibility settings updated"})
}

// ValidEnvironmentValues represents the allowed environment values
type ValidEnvironmentValues struct {
	Environment string `json:"environment"`
}

// HandleGetActiveEnvironment returns GET /v1/users/me/environment — returns the user's currently selected active environment
func (h *Handler) HandleGetActiveEnvironment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user settings from database
	settings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user settings, using defaults")
		settings = getDefaultSettings()
	}

	// Get active environment with fallback to production
	activeEnv := "production"
	if env, ok := settings["active_environment"].(string); ok && env != "" {
		activeEnv = env
	}

	// Validate it's one of the allowed values
	validEnvs := map[string]bool{"production": true, "staging": true, "development": true}
	if !validEnvs[activeEnv] {
		activeEnv = "production"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"environment": activeEnv,
		"available":   []string{"production", "staging", "development"},
	})
}

// HandleSetActiveEnvironment handles PATCH /v1/users/me/environment — updates the user's active environment
func (h *Handler) HandleSetActiveEnvironment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Environment string `json:"environment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate environment value
	validEnvs := map[string]bool{"production": true, "staging": true, "development": true}
	if !validEnvs[req.Environment] {
		writeJSONError(w, http.StatusBadRequest, "Invalid environment. Must be: production, staging, or development")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update active environment
	currentSettings["active_environment"] = req.Environment

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update active environment")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update environment preference")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Environment preference updated",
		"environment": req.Environment,
	})
}

// SessionSecuritySettingsRequest is the body for PATCH /v1/users/me/settings/security
type SessionSecuritySettingsRequest struct {
	SessionTimeout  *string `json:"sessionTimeout"`  // "1h", "24h", "7d", "30d", "never"
	RememberDevices *bool   `json:"rememberDevices"` // allow sessions on recognized devices for 30 days
}

// HandlePatchUserSettingsSecurityMe handles PATCH /v1/users/me/settings/security
func (h *Handler) HandlePatchUserSettingsSecurityMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SessionSecuritySettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate session timeout value
	if req.SessionTimeout != nil {
		validTimeouts := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true, "never": true}
		if !validTimeouts[*req.SessionTimeout] {
			writeJSONError(w, http.StatusBadRequest, "Invalid session timeout value. Must be: 1h, 24h, 7d, 30d, or never")
			return
		}
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update security-related fields
	if req.SessionTimeout != nil {
		currentSettings["sessionTimeout"] = *req.SessionTimeout
	}
	if req.RememberDevices != nil {
		currentSettings["rememberDevices"] = *req.RememberDevices
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update security settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update security settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Security settings updated"})
}

// CustomStatusRequest is the body for PATCH /v1/users/me/settings/status
type CustomStatusRequest struct {
	CustomStatus   string `json:"customStatus"`   // "online", "away", "busy", "offline", or ""
	CustomStatusEmoji string `json:"customStatusEmoji"` // Optional emoji
}

// ValidCustomStatusValues are the allowed custom status values
var ValidCustomStatusValues = map[string]bool{
	"":       true, // Clear status
	"online":  true,
	"away":    true,
	"busy":    true,
	"offline": true,
}

// HandlePatchUserSettingsStatusMe handles PATCH /v1/users/me/settings/status
// This allows users to set their own presence status (online, away, busy, offline)
func (h *Handler) HandlePatchUserSettingsStatusMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CustomStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate custom status value
	if !ValidCustomStatusValues[req.CustomStatus] {
		writeJSONError(w, http.StatusBadRequest, "Invalid custom status value. Must be: online, away, busy, offline, or empty string to clear")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Update custom status fields
	if req.CustomStatus == "" {
		currentSettings["customStatus"] = ""
		currentSettings["customStatusEmoji"] = ""
	} else {
		currentSettings["customStatus"] = req.CustomStatus
		if req.CustomStatusEmoji != "" {
			currentSettings["customStatusEmoji"] = req.CustomStatusEmoji
		}
	}

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update custom status")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update custom status")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":             "Custom status updated",
		"customStatus":        currentSettings["customStatus"],
		"customStatusEmoji":   currentSettings["customStatusEmoji"],
	})
}

// HandleGetUserSettingsStatusMe handles GET /v1/users/me/settings/status
// Returns the user's current custom status
func (h *Handler) HandleGetUserSettingsStatusMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current settings
	settings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		settings = getDefaultSettings()
	}

	customStatus, _ := settings["customStatus"].(string)
	customStatusEmoji, _ := settings["customStatusEmoji"].(string)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"customStatus":      customStatus,
		"customStatusEmoji": customStatusEmoji,
	})
}

// PlatformSettingsRequest is the body for PATCH /v1/users/me/settings/platform
type PlatformSettingsRequest struct {
	AutoEvolve               *bool   `json:"auto_evolve,omitempty"`
	RequireApproval          *bool   `json:"require_approval,omitempty"`
	SandboxValidation        *bool   `json:"sandbox_validation,omitempty"`
	DefaultCanaryPct         *int    `json:"default_canary_pct,omitempty"`
	MaxMutationsPerDay       *int    `json:"max_mutations_per_day,omitempty"`
	NotifyOnProposal         *bool   `json:"notify_on_proposal,omitempty"`
	NotifyOnDeploy           *bool   `json:"notify_on_deploy,omitempty"`
	NotifyOnRollback         *bool   `json:"notify_on_rollback,omitempty"`
	AutoRollbackOnError       *bool   `json:"auto_rollback_on_error,omitempty"`
	AutoRollbackErrorThreshold *int   `json:"auto_rollback_error_threshold,omitempty"`
}

// HandlePatchUserSettingsPlatformMe handles PATCH /v1/users/me/settings/platform
func (h *Handler) HandlePatchUserSettingsPlatformMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req PlatformSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current settings
	currentSettings, err := h.repo.GetUserSettings(claims.UserID)
	if err != nil {
		currentSettings = getDefaultSettings()
	}

	// Ensure platform key exists
	if currentSettings["platform"] == nil {
		currentSettings["platform"] = map[string]interface{}{}
	}
	platform, ok := currentSettings["platform"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Apply valid fields
	if req.AutoEvolve != nil {
		platform["auto_evolve"] = *req.AutoEvolve
	}
	if req.RequireApproval != nil {
		platform["require_approval"] = *req.RequireApproval
	}
	if req.SandboxValidation != nil {
		platform["sandbox_validation"] = *req.SandboxValidation
	}
	if req.DefaultCanaryPct != nil {
		v := *req.DefaultCanaryPct
		if v < 1 {
			v = 1
		}
		if v > 100 {
			v = 100
		}
		platform["default_canary_pct"] = v
	}
	if req.MaxMutationsPerDay != nil {
		v := *req.MaxMutationsPerDay
		if v < 1 {
			v = 1
		}
		if v > 100 {
			v = 100
		}
		platform["max_mutations_per_day"] = v
	}
	if req.NotifyOnProposal != nil {
		platform["notify_on_proposal"] = *req.NotifyOnProposal
	}
	if req.NotifyOnDeploy != nil {
		platform["notify_on_deploy"] = *req.NotifyOnDeploy
	}
	if req.NotifyOnRollback != nil {
		platform["notify_on_rollback"] = *req.NotifyOnRollback
	}
	if req.AutoRollbackOnError != nil {
		platform["auto_rollback_on_error"] = *req.AutoRollbackOnError
	}
	if req.AutoRollbackErrorThreshold != nil {
		v := *req.AutoRollbackErrorThreshold
		if v < 1 {
			v = 1
		}
		if v > 50 {
			v = 50
		}
		platform["auto_rollback_error_threshold"] = v
	}

	currentSettings["platform"] = platform

	if err := h.repo.UpdateUserSettings(claims.UserID, currentSettings); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update platform settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update platform settings")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":          claims.UserID,
		"auto_evolve":      platform["auto_evolve"],
		"require_approval":  platform["require_approval"],
		"default_canary":    platform["default_canary_pct"],
	}).Info("users: platform settings updated")

	writeJSON(w, http.StatusOK, map[string]string{"message": "Platform settings updated"})
}
