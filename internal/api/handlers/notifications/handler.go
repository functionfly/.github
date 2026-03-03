package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains notification handlers
type Handler struct {
	service *notification.Service
	repo    notification.Repository
}

// NewHandler creates a new notifications handler
func NewHandler(service *notification.Service, repo notification.Repository) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
	}
}

// ListNotificationsResponse represents the response for listing notifications
type ListNotificationsResponse struct {
	Notifications []*notification.Notification `json:"notifications"`
	Total         int                          `json:"total"`
	UnreadCount   int                          `json:"unread_count"`
}

// UpdatePreferencesRequest represents a request to update preferences
type UpdatePreferencesRequest struct {
	Preferences []notification.NotificationPreference `json:"preferences"`
}

// HandleListNotifications handles GET /v1/notifications
func (h *Handler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	opts := notification.ListOptions{}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			opts.Limit = l
		}
	}
	if opts.Limit == 0 || opts.Limit > 100 {
		opts.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			opts.Offset = o
		}
	}
	opts.Status = r.URL.Query().Get("status")
	opts.Category = r.URL.Query().Get("category")
	opts.UnreadOnly = r.URL.Query().Get("unread_only") == "true"

	// Get notifications
	notifications, err := h.service.ListNotifications(r.Context(), user.UserID, opts)
	if err != nil {
		logrus.WithError(err).Error("Failed to list notifications")
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}

	// Get unread count
	unreadCount, err := h.service.GetUnreadCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
		// Don't fail the request, just log the error
		unreadCount = 0
	}

	response := ListNotificationsResponse{
		Notifications: notifications,
		Total:         len(notifications),
		UnreadCount:   unreadCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetUnreadCount handles GET /v1/notifications/unread-count
func (h *Handler) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
		http.Error(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	response := map[string]int{
		"unread_count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMarkAsRead handles PATCH /v1/notifications/{id}/read
func (h *Handler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Verify notification belongs to user
	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	if n.UserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.MarkAsRead(r.Context(), notificationID); err != nil {
		logrus.WithError(err).Error("Failed to mark notification as read")
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "marked_as_read",
	})
}

// HandleMarkAllAsRead handles POST /v1/notifications/read-all
func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.MarkAllAsRead(r.Context(), user.UserID); err != nil {
		logrus.WithError(err).Error("Failed to mark all notifications as read")
		http.Error(w, "Failed to mark all as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "all_marked_as_read",
	})
}

// HandleDeleteNotification handles DELETE /v1/notifications/{id}
func (h *Handler) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Verify notification belongs to user
	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	if n.UserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteNotification(r.Context(), notificationID); err != nil {
		logrus.WithError(err).Error("Failed to delete notification")
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// HandleGetPreferences handles GET /v1/users/me/notification-preferences
func (h *Handler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	prefs, err := h.service.GetPreferences(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get notification preferences")
		http.Error(w, "Failed to get preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preferences": prefs,
	})
}

// HandleUpdatePreferences handles PATCH /v1/users/me/notification-preferences
func (h *Handler) HandleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update each preference
	for _, pref := range req.Preferences {
		pref.UserID = user.UserID // Ensure user can only update their own preferences
		if err := h.service.SavePreference(r.Context(), &pref); err != nil {
			logrus.WithError(err).Error("Failed to save preference")
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "preferences_updated",
	})
}

// RegisterRoutes registers the notification routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Notification routes
	router.HandleFunc("/v1/notifications", h.HandleListNotifications).Methods("GET")
	router.HandleFunc("/v1/notifications/unread-count", h.HandleGetUnreadCount).Methods("GET")
	router.HandleFunc("/v1/notifications/read-all", h.HandleMarkAllAsRead).Methods("POST")
	router.HandleFunc("/v1/notifications/{id}/read", h.HandleMarkAsRead).Methods("PATCH")
	router.HandleFunc("/v1/notifications/{id}", h.HandleDeleteNotification).Methods("DELETE")

	// Preference routes
	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleGetPreferences).Methods("GET")
	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleUpdatePreferences).Methods("PATCH")
}
