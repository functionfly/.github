package users

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetUserSkills handles GET /v1/users/{username}/skills
// Returns user skills/expertise
func (h *Handler) HandleGetUserSkills(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	// Get user by username
	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("Failed to get user by username")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Get user skills (return empty if tables not yet migrated)
	skills, err := h.repo.GetUserSkills(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get user skills, returning empty")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"skills": []interface{}{},
		})
		return
	}

	// Transform to response format
	response := make([]map[string]interface{}, 0, len(skills))
	for _, skill := range skills {
		response = append(response, map[string]interface{}{
			"id":       skill.ID,
			"name":     skill.Name,
			"level":    skill.Level,
			"category": skill.Category,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"skills": response,
	})
}

// HandleAddUserSkill handles POST /v1/users/me/skills
// Adds a new skill for the authenticated user
func (h *Handler) HandleAddUserSkill(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type addSkillRequest struct {
		Name     string `json:"name"`
		Level    string `json:"level"`    // beginner, intermediate, advanced, expert
		Category string `json:"category"` // language, framework, tool, platform, soft
	}

	var req addSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Validate level
	validLevels := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
	if req.Level != "" && !validLevels[req.Level] {
		writeJSONError(w, http.StatusBadRequest, "level must be one of: beginner, intermediate, advanced, expert")
		return
	}
	if req.Level == "" {
		req.Level = "intermediate"
	}

	skill := &storage.UserSkill{
		UserID:   claims.UserID,
		Name:     req.Name,
		Level:    req.Level,
		Category: req.Category,
	}

	if err := h.repo.AddUserSkill(skill); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeJSONError(w, http.StatusConflict, "Skill already exists")
			return
		}
		if strings.Contains(err.Error(), "does not exist") {
			logrus.WithError(err).WithField("userID", claims.UserID).Warn("user_skills table missing; run migrations")
			writeJSONError(w, http.StatusServiceUnavailable, "Profile features are not available yet. Run database migrations.")
			return
		}
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to add user skill")
		writeJSONError(w, http.StatusInternalServerError, "Failed to add skill")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      skill.ID,
		"name":    skill.Name,
		"level":   skill.Level,
		"message": "Skill added successfully",
	})
}

// HandleRemoveUserSkill handles DELETE /v1/users/me/skills/{id}
// Removes a skill for the authenticated user
func (h *Handler) HandleRemoveUserSkill(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	skillIDStr := mux.Vars(r)["id"]
	if skillIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid skill id")
		return
	}

	// First verify the skill belongs to this user (by checking if we can get it)
	userSkills, err := h.repo.GetUserSkills(claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			logrus.WithError(err).WithField("userID", claims.UserID).Warn("user_skills table missing; run migrations")
			writeJSONError(w, http.StatusServiceUnavailable, "Profile features are not available yet. Run database migrations.")
			return
		}
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to get user skills")
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify skill ownership")
		return
	}

	found := false
	for _, s := range userSkills {
		if s.ID == skillID {
			found = true
			break
		}
	}

	if !found {
		writeJSONError(w, http.StatusNotFound, "Skill not found")
		return
	}

	if err := h.repo.RemoveUserSkill(skillID); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			logrus.WithError(err).WithField("skillID", skillID).Warn("user_skills table missing; run migrations")
			writeJSONError(w, http.StatusServiceUnavailable, "Profile features are not available yet. Run database migrations.")
			return
		}
		logrus.WithError(err).WithField("skillID", skillID).Error("Failed to remove user skill")
		writeJSONError(w, http.StatusInternalServerError, "Failed to remove skill")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Skill removed successfully"})
}