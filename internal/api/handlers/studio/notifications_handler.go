package studio

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// NotificationsHandler handles studio notification HTTP requests.
type NotificationsHandler struct {
	repo *NotificationRepository
}

// NewNotificationsHandler creates a notifications handler.
func NewNotificationsHandler(repo *NotificationRepository) *NotificationsHandler {
	return &NotificationsHandler{repo: repo}
}

// HandleListNotifications handles GET /v1/studio/notifications
func (h *NotificationsHandler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	notifications, err := h.repo.ListNotifications(r.Context(), tenantID, userID, environment, unreadOnly, limit, offset)
	if err != nil {
		logrus.WithError(err).Warn("studio notifications: failed to list")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"notifications": []StudioNotification{},
			"unread_count":  0,
		})
		return
	}

	unreadCount, err := h.repo.CountUnread(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Warn("studio notifications: failed to count unread")
		unreadCount = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// HandleMarkRead handles PATCH /v1/studio/notifications/{id}
func (h *NotificationsHandler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	environment := getEnvironment(r)

	notification, err := h.repo.MarkRead(r.Context(), tenantID, userID, environment, id)
	if err != nil {
		logrus.WithError(err).Error("studio notifications: failed to mark read")
		writeJSONError(w, http.StatusInternalServerError, "Failed to mark notification read")
		return
	}
	if notification == nil {
		writeJSONError(w, http.StatusNotFound, "Notification not found")
		return
	}

	writeJSON(w, http.StatusOK, notification)
}

// HandleMarkAllRead handles POST /v1/studio/notifications/mark-all-read
func (h *NotificationsHandler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	count, err := h.repo.MarkAllRead(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Error("studio notifications: failed to mark all read")
		writeJSONError(w, http.StatusInternalServerError, "Failed to mark notifications read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notifications marked read",
		"count":   count,
	})
}

// HandleDeleteNotification handles DELETE /v1/studio/notifications/{id}
func (h *NotificationsHandler) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	environment := getEnvironment(r)

	if err := h.repo.DeleteNotification(r.Context(), tenantID, userID, environment, id); err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "Notification not found")
			return
		}
		logrus.WithError(err).Error("studio notifications: failed to delete")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

// HandleClearNotifications handles DELETE /v1/studio/notifications
func (h *NotificationsHandler) HandleClearNotifications(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	count, err := h.repo.DeleteAllNotifications(r.Context(), tenantID, userID, environment)
	if err != nil {
		logrus.WithError(err).Error("studio notifications: failed to clear")
		writeJSONError(w, http.StatusInternalServerError, "Failed to clear notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notifications cleared",
		"count":   count,
	})
}

// HandleCreateNotification handles POST /v1/studio/notifications (development only)
func (h *NotificationsHandler) HandleCreateNotification(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEVELOPMENT") != "true" {
		writeJSONError(w, http.StatusForbidden, "Creating studio notifications via API is disabled in production")
		return
	}
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		Type      string  `json:"type"`
		Category  string  `json:"category"`
		Title     string  `json:"title"`
		Message   string  `json:"message"`
		ActionURL *string `json:"action_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Title == "" || req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "title and message are required")
		return
	}

	uid := userID
	notification, err := h.repo.CreateNotification(r.Context(), CreateNotificationParams{
		TenantID:    tenantID,
		UserID:      &uid,
		Environment: environment,
		Type:        req.Type,
		Category:    req.Category,
		Title:       req.Title,
		Message:     req.Message,
		ActionURL:   req.ActionURL,
	})
	if err != nil {
		logrus.WithError(err).Error("studio notifications: failed to create")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create notification")
		return
	}

	writeJSON(w, http.StatusCreated, notification)
}
