package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type notificationRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newNotificationRateLimiter() *notificationRateLimiter {
	return &notificationRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    100,
		window:   time.Minute,
	}
}

func (rl *notificationRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	if requests, ok := rl.requests[key]; ok {
		var validRequests []time.Time
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		rl.requests[key] = validRequests
		if len(validRequests) >= rl.limit {
			return false
		}
	}

	rl.requests[key] = append(rl.requests[key], now)
	return true
}

type notificationAuditLogger struct {
	logger *logrus.Logger
}

func newNotificationAuditLogger() *notificationAuditLogger {
	return &notificationAuditLogger{
		logger: logrus.New(),
	}
}

func (l *notificationAuditLogger) log(userID uuid.UUID, action string, notificationID uuid.UUID, ipAddress, userAgent string, success bool, err error) {
	fields := logrus.Fields{
		"event":            "notification_access",
		"user_id":          userID.String(),
		"action":           action,
		"notification_id": notificationID.String(),
		"ip_address":       ipAddress,
		"user_agent":       userAgent,
		"success":          success,
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.logger.WithFields(fields).Info("Notification audit event")
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}

func getUserAgent(r *http.Request) string {
	return r.UserAgent()
}

// Handler contains notification handlers
type Handler struct {
	service     *notification.Service
	repo        notification.Repository
	rateLimiter *notificationRateLimiter
	auditLogger *notificationAuditLogger
}

// NewHandler creates a new notifications handler
func NewHandler(service *notification.Service, repo notification.Repository) *Handler {
	return &Handler{
		service:     service,
		repo:        repo,
		rateLimiter: newNotificationRateLimiter(),
		auditLogger: newNotificationAuditLogger(),
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

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

	notifications, err := h.service.ListNotifications(r.Context(), user.UserID, opts)
	if err != nil {
		logrus.WithError(err).Error("Failed to list notifications")
		apierror.WriteError(w, apierror.NewInternal("Failed to list notifications"))
		return
	}

	unreadCount, err := h.service.GetUnreadCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	totalCount, err := h.service.GetTotalCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get total count")
		apierror.WriteError(w, apierror.NewInternal("Failed to get notification counts"))
		return
	}

	unreadCount, err := h.service.GetUnreadCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
		apierror.WriteError(w, apierror.NewInternal("Failed to get notification counts"))
		return
	}

	categoryCounts, err := h.service.GetUnreadCountsByCategory(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get category counts, using defaults")
		categoryCounts = make(map[string]int)
	}

	if categoryCounts == nil {
		categoryCounts = make(map[string]int)
	}

	frontendCategories := map[string]int{
		"trust":    categoryCounts["system"],
		"revenue":  categoryCounts["billing"],
		"issues":   categoryCounts["deployment"],
		"messages": categoryCounts["team"],
		"security": categoryCounts["security"],
	}

	response := map[string]interface{}{
		"total":      totalCount,
		"unread":     unreadCount,
		"byCategory": frontendCategories,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode notification counts response")
		apierror.WriteError(w, apierror.NewInternal("Failed to encode response"))
		return
	}
}

// HandleMarkAsRead handles PATCH /v1/notifications/{id}/read
func (h *Handler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid notification ID"))
		return
	}

	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Notification not found"))
		return
	}

	if n.UserID != user.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	if err := h.service.MarkAsRead(r.Context(), notificationID); err != nil {
		logrus.WithError(err).Error("Failed to mark notification as read")
		apierror.WriteError(w, apierror.NewInternal("Failed to mark as read"))
		return
	}

	h.auditLogger.log(user.UserID, "mark_as_read", notificationID, getClientIP(r), getUserAgent(r), true, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "marked_as_read",
	})
}

// HandleMarkAllAsRead handles POST /v1/notifications/read-all
func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	if err := h.service.MarkAllAsRead(r.Context(), user.UserID); err != nil {
		logrus.WithError(err).Error("Failed to mark all notifications as read")
		apierror.WriteError(w, apierror.NewInternal("Failed to mark all as read"))
		return
	}

	h.auditLogger.log(user.UserID, "mark_all_as_read", uuid.Nil, getClientIP(r), getUserAgent(r), true, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "all_marked_as_read",
	})
}

// patchNotificationRequest is a minimal PATCH body for /v1/notifications/{id}.
type patchNotificationRequest struct {
	Status string `json:"status"`
}

// HandlePatchNotification handles PATCH /v1/notifications/{id} (e.g. archive).
func (h *Handler) HandlePatchNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid notification ID"))
		return
	}

	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Notification not found"))
		return
	}

	if n.UserID != user.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req patchNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	switch req.Status {
	case "archived":
		if err := h.service.ArchiveNotification(r.Context(), notificationID); err != nil {
			logrus.WithError(err).Error("Failed to archive notification")
			apierror.WriteError(w, apierror.NewInternal("Failed to archive notification"))
			return
		}
		h.auditLogger.log(user.UserID, "archive", notificationID, getClientIP(r), getUserAgent(r), true, nil)
	default:
		apierror.WriteError(w, apierror.NewBadRequest("Unsupported status"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
}

// HandleDeleteNotification handles DELETE /v1/notifications/{id}
func (h *Handler) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid notification ID"))
		return
	}

	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Notification not found"))
		return
	}

	if n.UserID != user.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	if err := h.service.DeleteNotification(r.Context(), notificationID); err != nil {
		logrus.WithError(err).Error("Failed to delete notification")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete notification"))
		return
	}

	h.auditLogger.log(user.UserID, "delete", notificationID, getClientIP(r), getUserAgent(r), true, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// HandleGetPreferences handles GET /v1/users/me/notification-preferences
func (h *Handler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	prefs, err := h.service.GetPreferences(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get notification preferences")
		apierror.WriteError(w, apierror.NewInternal("Failed to get preferences"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.rateLimiter.Allow(user.UserID.String()) {
		apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if len(req.Preferences) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("No preferences provided"))
		return
	}

	validChannels := map[string]bool{
		"email":   true,
		"in_app":  true,
		"webhook": true,
		"push":    true,
	}
	validCategories := map[string]bool{
		"system":     true,
		"security":   true,
		"billing":    true,
		"deployment": true,
		"function":   true,
		"team":       true,
		"messages":   true,
		"registry":   true,
		"failover":   true,
		"provider":   true,
	}
	validFrequencies := map[string]bool{
		"immediate":     true,
		"digest_daily":  true,
		"digest_weekly": true,
	}

	for _, pref := range req.Preferences {
		if !validChannels[pref.Channel] {
			http.Error(w, fmt.Sprintf("Invalid channel: %s", pref.Channel), http.StatusBadRequest)
			return
		}
		if !validCategories[pref.Category] {
			http.Error(w, fmt.Sprintf("Invalid category: %s", pref.Category), http.StatusBadRequest)
			return
		}
		if pref.Frequency != "" && !validFrequencies[pref.Frequency] {
			http.Error(w, fmt.Sprintf("Invalid frequency: %s", pref.Frequency), http.StatusBadRequest)
			return
		}
		pref.UserID = user.UserID
		if err := h.service.SavePreference(r.Context(), &pref); err != nil {
			logrus.WithError(err).Error("Failed to save preference")
			apierror.WriteError(w, apierror.NewInternal("Failed to save preferences"))
			return
		}
		h.auditLogger.log(user.UserID, "update_preference", uuid.Nil, getClientIP(r), getUserAgent(r), true, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "preferences_updated",
	})
}

// RegisterRoutes registers the notification routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/notifications", h.HandleListNotifications).Methods("GET")
	router.HandleFunc("/v1/notifications/unread-count", h.HandleGetUnreadCount).Methods("GET")
	router.HandleFunc("/v1/notifications/read-all", h.HandleMarkAllAsRead).Methods("POST")
	router.HandleFunc("/v1/notifications/{id}/read", h.HandleMarkAsRead).Methods("PATCH")
	router.HandleFunc("/v1/notifications/{id}", h.HandlePatchNotification).Methods("PATCH")
	router.HandleFunc("/v1/notifications/{id}", h.HandleDeleteNotification).Methods("DELETE")

	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleGetPreferences).Methods("GET")
	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleUpdatePreferences).Methods("PATCH")
}
