package users

import (
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetUserAchievements handles GET /v1/users/{username}/achievements
// Returns earned badges/achievements with progress
func (h *Handler) HandleGetUserAchievements(w http.ResponseWriter, r *http.Request) {
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

	// Join date: user.CreatedAt is the user's actual account creation time from the users table (when they joined).
	joinDate := user.CreatedAt
	if joinDate.IsZero() {
		joinDate = time.Now() // fallback only if created_at was never set
	}
	earnedAtISO := joinDate.Format("2006-01-02T15:04:05Z07:00")

	// Get user achievements (on error still return "Joined FunctionFly" so something shows)
	achievements, err := h.repo.GetUserAchievements(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get user achievements, returning joined achievement only")
		joinedOnly := []map[string]interface{}{{
			"id":          "joined-" + user.ID.String(),
			"slug":        "joined_functionfly",
			"name":        "Member",
			"description": "Joined FunctionFly",
			"icon":        "UserPlus",
			"color":       "blue",
			"category":    "milestone",
			"points":      0,
			"earnedAt":    earnedAtISO,
			"progress":    100,
			"isCompleted": true,
			"metadata":    map[string]interface{}{},
		}}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"achievements": joinedOnly,
			"totalPoints":  0,
			"available":    1,
		})
		return
	}

	// Transform to response format
	response := make([]map[string]interface{}, 0, len(achievements))
	for _, ua := range achievements {
		if ua.Achievement == nil {
			continue
		}
		response = append(response, map[string]interface{}{
			"id":          ua.ID,
			"slug":        ua.Achievement.Slug,
			"name":        ua.Achievement.Name,
			"description": ua.Achievement.Description,
			"icon":        ua.Achievement.Icon,
			"color":       ua.Achievement.Color,
			"category":    ua.Achievement.Category,
			"points":      ua.Achievement.Points,
			"earnedAt":    ua.EarnedAt.Format("2006-01-02T15:04:05Z07:00"),
			"progress":    ua.Progress,
			"isCompleted": ua.IsCompleted,
			"metadata":    ua.Metadata,
		})
	}

	// Ensure "Joined FunctionFly" achievement is always included as earned (earnedAt = real join date from users.created_at)
	hasJoined := false
	for _, m := range response {
		if s, _ := m["slug"].(string); s == "joined_functionfly" {
			hasJoined = true
			break
		}
	}
	if !hasJoined {
		joinedDef, _ := h.repo.GetAchievementBySlug("joined_functionfly")
		name, desc, icon, color, category, points := "Member", "Joined FunctionFly", "UserPlus", "blue", "milestone", 0
		if joinedDef != nil {
			name, desc, icon, color, category, points = joinedDef.Name, joinedDef.Description, joinedDef.Icon, joinedDef.Color, joinedDef.Category, joinedDef.Points
		}
		response = append(response, map[string]interface{}{
			"id":          "joined-" + user.ID.String(),
			"slug":        "joined_functionfly",
			"name":        name,
			"description": desc,
			"icon":        icon,
			"color":       color,
			"category":    category,
			"points":      points,
			"earnedAt":    earnedAtISO,
			"progress":    100,
			"isCompleted": true,
			"metadata":    map[string]interface{}{},
		})
	}

	// Get all available achievements for progress tracking
	allAchievements, err := h.repo.ListAchievements()
	if err != nil {
		logrus.WithError(err).Error("Failed to list all achievements")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"achievements": response,
		"totalPoints":  h.calculateTotalPoints(achievements),
		"available":    len(allAchievements),
	})
}

// calculateTotalPoints calculates the sum of points from earned achievements
func (h *Handler) calculateTotalPoints(achievements []*storage.UserAchievement) int {
	total := 0
	for _, ua := range achievements {
		if ua.IsCompleted && ua.Achievement != nil {
			total += ua.Achievement.Points
		}
	}
	return total
}