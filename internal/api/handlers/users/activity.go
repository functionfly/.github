package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/apierror"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetUserActivity handles GET /v1/users/{username}/activity
// Returns timeline of user actions (publish, update, earn badge, etc.)
func (h *Handler) HandleGetUserActivity(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		apierror.WriteError(w, apierror.NewBadRequest("username is required"))
		return
	}

	// Get query params for pagination
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := fmt.Sscanf(l, "%d", &limit); err == nil && parsed == 1 {
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := fmt.Sscanf(o, "%d", &offset); err == nil && parsed == 1 {
			if offset < 0 {
				offset = 0
			}
		}
	}

	// Get user by username
	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("Failed to get user by username")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve user"))
		return
	}
	if user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	// Get user activity (return empty if tables not yet migrated)
	activities, err := h.repo.GetUserActivity(user.ID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get user activity, returning empty")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"activities": []interface{}{},
			"limit":      limit,
			"offset":     offset,
			"total":      0,
		})
		return
	}

	// Transform to response format
	response := make([]map[string]interface{}, 0, len(activities))
	for _, activity := range activities {
		response = append(response, map[string]interface{}{
			"id":          activity.ID,
			"type":        activity.ActivityType,
			"title":       activity.Title,
			"description": activity.Description,
			"metadata":    activity.Metadata,
			"isPublic":    activity.IsPublic,
			"createdAt":   activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"activities": response,
		"limit":      limit,
		"offset":     offset,
		"total":      len(response),
	})
}

// HandleCreateUserActivity handles POST /v1/users/me/activity (for authenticated users)
// Creates a new activity feed item
func (h *Handler) HandleCreateUserActivity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	type createActivityRequest struct {
		ActivityType string                 `json:"activityType"`
		Title        string                 `json:"title"`
		Description  string                 `json:"description"`
		Metadata     map[string]interface{} `json:"metadata"`
		IsPublic     bool                   `json:"isPublic"`
	}

	var req createActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.ActivityType == "" || req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("activityType and title are required"))
		return
	}

	activity := &storage.UserActivity{
		UserID:       claims.UserID,
		ActivityType: req.ActivityType,
		Title:        req.Title,
		Description:  req.Description,
		Metadata:     req.Metadata,
		IsPublic:     req.IsPublic,
	}

	if err := h.repo.CreateUserActivity(activity); err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to create user activity")
		apierror.WriteError(w, apierror.NewInternal("Failed to create activity"))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        activity.ID,
		"message":   "Activity created successfully",
		"createdAt": activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
