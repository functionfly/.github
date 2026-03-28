package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var errMissingOrInvalidAuth = errors.New("missing or invalid authorization")

// isUserOnline determines if a user is online based on their last activity time
// A user is considered online if they were active within the last 5 minutes
func isUserOnline(lastActiveAt *time.Time) bool {
	if lastActiveAt == nil {
		return false
	}
	return time.Since(*lastActiveAt) < 5*time.Minute
}

// formatLastActivePointer formats the last active time for display (pointer version)
func formatLastActivePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatLastActive(*t)
}

// extractClaimsFromRequest parses the Bearer token from r and validates it via authSvc.
// Returns claims on success, or an error if missing/invalid.
func extractClaimsFromRequest(r *http.Request, authSvc *auth.AuthService) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errMissingOrInvalidAuth
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errMissingOrInvalidAuth
	}
	return authSvc.ValidateToken(parts[1])
}

// HandleGetPublicProfile returns a user's public profile by username.
// This endpoint is public — it never exposes email, tenantId, role, or any sensitive fields.
// If username is "me", we require auth and return the current user (same as GET /users/me), so that
// GET /v1/users/me works even when the router matches /users/{username} instead of /users/me.
func (h *Handler) HandleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	if username == "me" {
		claims, err := extractClaimsFromRequest(r, h.authSvc)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		r = middleware.SetUserInContext(r, claims)
		h.HandleGetMe(w, r)
		return
	}

	user, err := h.repo.GetUserForPublicProfile(username)
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

	// Helper to get string from pointer
	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// Helper to get int from pointer
	getInt := func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	}

	// Fetch published functions from registry
	publishedFunctions := h.getPublishedFunctions(usernameStr)

	// Determine online status
	isOnline := isUserOnline(user.LastActiveAt)
	lastActive := formatLastActivePointer(user.LastActiveAt)

	profile := map[string]interface{}{
		"id":                 user.ID,
		"username":           usernameStr,
		"name":               name,
		"avatar":             avatar,
		"bio":                bio,
		"location":           getString(user.Location),
		"website":            getString(user.Website),
		"jobTitle":           getString(user.JobTitle),
		"companyName":        getString(user.CompanyName),
		"twitterUrl":         getString(user.TwitterURL),
		"githubUrl":          getString(user.GithubURL),
		"linkedinUrl":        getString(user.LinkedInURL),
		"createdAt":          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"publishedFunctions": publishedFunctions,
		"isOnline":           isOnline,
		"lastActive":         lastActive,
		"profileNumber":      getInt(user.ProfileNumber),
		"role":               user.Role, // Platform admin role for badge display
	}
	h.attachProfileStats(profile, user.ID)
	h.applyProfileVisibility(profile, user.ID)

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

	user, err := h.repo.GetUserForPublicProfile(username)
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

	// Helper to get string from pointer
	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// Helper to get int from pointer
	getInt := func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	}

	// Fetch published functions from registry
	publishedFunctions := h.getPublishedFunctions(usernameStr)

	// Determine online status
	isOnline := isUserOnline(user.LastActiveAt)
	lastActive := formatLastActivePointer(user.LastActiveAt)

	// Build SEO-enhanced profile with additional metadata
	profile := map[string]interface{}{
		"id":                 user.ID,
		"username":           usernameStr,
		"name":               name,
		"avatar":             avatar,
		"bio":                bio,
		"location":           getString(user.Location),
		"website":            getString(user.Website),
		"jobTitle":           getString(user.JobTitle),
		"companyName":        getString(user.CompanyName),
		"twitterUrl":         getString(user.TwitterURL),
		"githubUrl":          getString(user.GithubURL),
		"linkedinUrl":        getString(user.LinkedInURL),
		"createdAt":          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"publishedFunctions": publishedFunctions,
		"isOnline":           isOnline,
		"lastActive":         lastActive,
		"profileNumber":      getInt(user.ProfileNumber),
		"role":               user.Role, // Platform admin role for badge display
		// SEO enhancement fields
		"profileUrl":     "/@" + usernameStr,
		"totalFunctions": len(publishedFunctions),
	}
	h.attachProfileStats(profile, user.ID)
	h.applyProfileVisibility(profile, user.ID)

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

	// Helper to get string from pointer
	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// Helper to get int from pointer
	getInt := func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	}

	// Determine online status - for own profile, always show as online
	// since they're currently making this request
	isOnline := true
	lastActive := "Just now"

	resp := map[string]interface{}{
		"id":            user.ID,
		"tenantId":      user.TenantID,
		"email":         user.Email,
		"name":          name,
		"username":      usernameStr,
		"companyName":   companyName,
		"avatar":        avatar,
		"plan":          plan,
		"bio":           getString(user.Bio),
		"location":      getString(user.Location),
		"website":       getString(user.Website),
		"jobTitle":      getString(user.JobTitle),
		"socialLinks":   user.SocialLinks,
		"twitterUrl":    getString(user.TwitterURL),
		"githubUrl":     getString(user.GithubURL),
		"linkedinUrl":   getString(user.LinkedInURL),
		"dateOfBirth": func() interface{} {
			if user.DateOfBirth == nil {
				return nil
			}
			return user.DateOfBirth.Format("2006-01-02")
		}(),
		"isOnline":      isOnline,
		"lastActive":    lastActive,
		"updatedAt":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"profileNumber": getInt(user.ProfileNumber),
		"role":          user.Role, // Platform admin role for badge display
		"createdAt":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	h.attachProfileStats(resp, user.ID)
	writeJSON(w, http.StatusOK, resp)
}

// UpdateMeRequest represents the request body for updating the current user's profile.
// Optional fields use *string so we can distinguish "not sent" (no change) from "sent empty" (clear).
type UpdateMeRequest struct {
	Name        string  `json:"name"`
	Username    string  `json:"username"`
	CompanyName *string `json:"companyName,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Avatar      *string `json:"avatar,omitempty"` // Profile picture URL (nil = no change, "" = clear); stored in provider_data.avatar_url
	// Extended profile fields (pointer = key present in JSON; nil = omit, "" = clear)
	Location    *string                `json:"location,omitempty"`
	Website     *string                `json:"website,omitempty"`
	JobTitle    *string                `json:"jobTitle,omitempty"`
	SocialLinks map[string]interface{} `json:"socialLinks,omitempty"`
	TwitterURL  *string                `json:"twitterUrl,omitempty"`
	GithubURL   *string                `json:"githubUrl,omitempty"`
	LinkedInURL *string                `json:"linkedinUrl,omitempty"`
	DateOfBirth *string                `json:"dateOfBirth,omitempty"` // YYYY-MM-DD, "" clears
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
	if req.CompanyName != nil {
		updates["company_name"] = strings.TrimSpace(*req.CompanyName)
	}
	if req.Bio != nil {
		updates["bio"] = strings.TrimSpace(*req.Bio)
	}
	if req.Location != nil {
		updates["location"] = strings.TrimSpace(*req.Location)
	}
	if req.Website != nil {
		website := strings.TrimSpace(*req.Website)
		if website != "" && !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		updates["website"] = website
	}
	if req.JobTitle != nil {
		updates["job_title"] = strings.TrimSpace(*req.JobTitle)
	}
	if req.TwitterURL != nil {
		updates["twitter_url"] = strings.TrimSpace(*req.TwitterURL)
	}
	if req.GithubURL != nil {
		updates["github_url"] = strings.TrimSpace(*req.GithubURL)
	}
	if req.LinkedInURL != nil {
		updates["linkedin_url"] = strings.TrimSpace(*req.LinkedInURL)
	}
	if req.DateOfBirth != nil {
		dobStr := strings.TrimSpace(*req.DateOfBirth)
		if dobStr == "" {
			updates["date_of_birth"] = (*time.Time)(nil)
		} else {
			dob, err := time.Parse("2006-01-02", dobStr)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "dateOfBirth must be in YYYY-MM-DD format")
				return
			}
			updates["date_of_birth"] = &dob
		}
	}
	if len(req.SocialLinks) > 0 {
		updates["social_links"] = req.SocialLinks
	}

	// Persist avatar URL in provider_data when client sends avatar (nil = no change, "" = clear)
	var responseAvatar string
	if req.Avatar != nil {
		currentUser, err := h.repo.GetUserByID(claims.UserID)
		if err != nil || currentUser == nil {
			// Continue without updating avatar if we can't load user
		} else {
			merged := make(map[string]interface{})
			if currentUser.ProviderData != nil {
				for k, v := range currentUser.ProviderData {
					merged[k] = v
				}
			}
			merged["avatar_url"] = *req.Avatar
			if err := h.repo.UpdateUserProviderData(claims.UserID, merged); err != nil {
				logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to update avatar in provider_data")
			} else {
				responseAvatar = *req.Avatar
			}
		}
	}

	// No-op: no profile fields and no avatar change — return current user (200) instead of 400
	if len(updates) == 0 && req.Avatar == nil {
		updatedUser, _ := h.repo.GetUserByID(claims.UserID)
		if updatedUser == nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to load profile")
			return
		}
		avatar := ""
		if updatedUser.ProviderData != nil {
			if a, ok := updatedUser.ProviderData["avatar_url"].(string); ok {
				avatar = a
			}
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
		getString := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
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
				"bio":         getString(updatedUser.Bio),
				"location":    getString(updatedUser.Location),
				"website":     getString(updatedUser.Website),
				"jobTitle":    getString(updatedUser.JobTitle),
				"socialLinks": updatedUser.SocialLinks,
				"twitterUrl":  getString(updatedUser.TwitterURL),
				"githubUrl":   getString(updatedUser.GithubURL),
				"linkedinUrl": getString(updatedUser.LinkedInURL),
				"dateOfBirth": func() interface{} {
					if updatedUser.DateOfBirth == nil {
						return nil
					}
					return updatedUser.DateOfBirth.Format("2006-01-02")
				}(),
				"updatedAt":   updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"role":        updatedUser.Role, // Platform admin role for badge display
			},
		})
		return
	}

	var updatedUser *storage.User
	if len(updates) > 0 {
		var err error
		updatedUser, err = h.repo.UpdateUser(context.Background(), claims.UserID, updates)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				writeJSONError(w, http.StatusConflict, "Username is already taken")
				return
			}
			logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to update user")
			writeJSONError(w, http.StatusInternalServerError, "Failed to update profile")
			return
		}
	} else {
		updatedUser, _ = h.repo.GetUserByID(claims.UserID)
	}
	if updatedUser == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load profile")
		return
	}

	// Build avatar for response: use responseAvatar if we just set it, else from provider_data
	avatar := responseAvatar
	if avatar == "" && updatedUser.ProviderData != nil {
		if a, ok := updatedUser.ProviderData["avatar_url"].(string); ok {
			avatar = a
		}
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

	// Helper to get string from pointer
	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
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
			"bio":         getString(updatedUser.Bio),
			"location":    getString(updatedUser.Location),
			"website":     getString(updatedUser.Website),
			"jobTitle":    getString(updatedUser.JobTitle),
			"socialLinks": updatedUser.SocialLinks,
			"twitterUrl":  getString(updatedUser.TwitterURL),
			"githubUrl":   getString(updatedUser.GithubURL),
			"linkedinUrl": getString(updatedUser.LinkedInURL),
			"dateOfBirth": func() interface{} {
				if updatedUser.DateOfBirth == nil {
					return nil
				}
				return updatedUser.DateOfBirth.Format("2006-01-02")
			}(),
			"updatedAt":   updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"role":        updatedUser.Role, // Platform admin role for badge display
		},
	})
}
