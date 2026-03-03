package users

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetUserAnalytics handles GET /v1/users/{username}/analytics
// Returns execution history, popular functions, geographic stats, device/browser stats
func (h *Handler) HandleGetUserAnalytics(w http.ResponseWriter, r *http.Request) {
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

	// Get execution stats
	executionStats, err := h.repo.GetUserExecutionStats(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to get user execution stats")
		// Continue without stats
		executionStats = map[string]interface{}{}
	}

	// Get popular functions
	popularFunctions, err := h.repo.GetUserPopularFunctions(user.ID, 5)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to get popular functions")
		popularFunctions = []map[string]interface{}{}
	}

	// Get geographic stats
	geoStats, err := h.repo.GetUserGeographicStats(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to get geographic stats")
		geoStats = map[string]interface{}{"regions": []interface{}{}}
	}

	// Get device stats
	deviceStats, err := h.repo.GetUserDeviceStats(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to get device stats")
		deviceStats = map[string]interface{}{"devices": []interface{}{}}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"executionStats":   executionStats,
		"popularFunctions": popularFunctions,
		"geographicStats":  geoStats,
		"deviceStats":      deviceStats,
	})
}
