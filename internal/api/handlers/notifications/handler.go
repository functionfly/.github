package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains notification handlers
type Handler struct {
	service   *notification.Service
	repo      notification.Repository
	auditRepo *storage.AuditRepository
	logger    *logrus.Logger
}

// NewHandler creates a new notifications handler
func NewHandler(service *notification.Service, repo notification.Repository, auditRepo *storage.AuditRepository, logger *logrus.Logger) *Handler {
	return &Handler{
		service:   service,
		repo:       repo,
		auditRepo: auditRepo,
		logger:    logger,
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

	// Get total count with filters applied
	totalCount, err := h.service.CountNotifications(r.Context(), user.UserID, opts)
	if err != nil {
		logrus.WithError(err).Error("Failed to count notifications")
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}

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
		Total:         totalCount,
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

	// Get total count (all notifications)
	totalCount, err := h.service.GetTotalCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get total count")
		http.Error(w, "Failed to get notification counts", http.StatusInternalServerError)
		return
	}

	// Get unread count
	unreadCount, err := h.service.GetUnreadCount(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
		http.Error(w, "Failed to get notification counts", http.StatusInternalServerError)
		return
	}

	// Get unread counts by category
	categoryCounts, err := h.service.GetUnreadCountsByCategory(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get category counts, using defaults")
		// Don't fail the request, just use empty counts
		categoryCounts = make(map[string]int)
	}

	// Ensure categoryCounts is not nil
	if categoryCounts == nil {
		categoryCounts = make(map[string]int)
	}

	// Map backend categories to frontend expected categories
	frontendCategories := map[string]int{
		"trust":    categoryCounts["system"] + categoryCounts["function"] + categoryCounts["registry"], // System/trust + function/registry categories
		"revenue":  categoryCounts["billing"],                                                                                       // Billing/revenue notifications
		"issues":   categoryCounts["deployment"],                                                                                     // Deployment issues
		"messages": categoryCounts["team"],                                                                                          // Team messages/invitations
		"security": categoryCounts["security"],                                                                                      // Security notifications
	}

	response := map[string]interface{}{
		"total":      totalCount,
		"unread":     unreadCount,
		"byCategory": frontendCategories,
	}

	logrus.WithFields(logrus.Fields{
		"total":      totalCount,
		"unread":     unreadCount,
		"byCategory": frontendCategories,
	}).Debug("Sending notification counts response")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode notification counts response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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

	// Audit log for mark as read
	h.logAuditEvent(r, "notification.mark_read", "notification", &notificationID, nil, map[string]interface{}{
		"notification_id": notificationID.String(),
		"user_id":         user.UserID.String(),
	}, true)

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

	// Audit log for mark all as read
	h.logAuditEvent(r, "notification.mark_all_read", "notification", nil, nil, map[string]interface{}{
		"user_id": user.UserID.String(),
	}, true)

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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	n, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	if n.UserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req patchNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	switch req.Status {
	case "archived":
		if err := h.service.ArchiveNotification(r.Context(), notificationID); err != nil {
			logrus.WithError(err).Error("Failed to archive notification")
			http.Error(w, "Failed to archive notification", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unsupported status", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
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

	// Audit log for delete notification
	h.logAuditEvent(r, "notification.delete", "notification", &notificationID, map[string]interface{}{
		"notification_id": notificationID.String(),
		"user_id":         user.UserID.String(),
		"notification_type": n.Type,
		"notification_category": n.Category,
	}, nil, true)

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

	// Fetch existing preferences to validate ownership when IDs are provided
	existingPrefs, err := h.service.GetPreferences(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get existing preferences for validation")
		http.Error(w, "Failed to validate preferences", http.StatusInternalServerError)
		return
	}

	// Build a map of existing preference IDs for quick lookup
	existingPrefIDs := make(map[uuid.UUID]bool)
	for _, p := range existingPrefs {
		existingPrefIDs[p.ID] = true
	}

	// Update each preference
	for _, pref := range req.Preferences {
		// Validate: if preference ID is provided, ensure it belongs to this user
		if pref.ID != uuid.Nil && !existingPrefIDs[pref.ID] {
			http.Error(w, "Forbidden: preference does not belong to user", http.StatusForbidden)
			return
		}
		pref.UserID = user.UserID // Ensure user can only update their own preferences
		if err := h.service.SavePreference(r.Context(), &pref); err != nil {
			logrus.WithError(err).Error("Failed to save preference")
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	// Audit log for preference update
	prefChannels := make([]string, len(req.Preferences))
	prefCategories := make([]string, len(req.Preferences))
	for i, p := range req.Preferences {
		prefChannels[i] = p.Channel
		prefCategories[i] = p.Category
	}
	h.logAuditEvent(r, "notification.preferences.update", "notification_preferences", nil, nil, map[string]interface{}{
		"user_id":     user.UserID.String(),
		"channels":    prefChannels,
		"categories":  prefCategories,
		"pref_count":  len(req.Preferences),
	}, true)

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
	router.HandleFunc("/v1/notifications/{id}", h.HandlePatchNotification).Methods("PATCH")
	router.HandleFunc("/v1/notifications/{id}", h.HandleDeleteNotification).Methods("DELETE")

	// Preference routes
	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleGetPreferences).Methods("GET")
	router.HandleFunc("/v1/users/me/notification-preferences", h.HandleUpdatePreferences).Methods("PATCH")
}

// logAuditEvent creates an audit log entry for notification operations
func (h *Handler) logAuditEvent(r *http.Request, action, resourceType string, resourceID *uuid.UUID, beforeState, afterState interface{}, success bool) {
	if h.auditRepo == nil {
		h.logger.Warn("No audit repository configured, skipping audit log")
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		h.logger.Warn("No authenticated user for audit logging")
		return
	}

	event := &storage.AuditEvent{
		ActorUserID:  &user.UserID,
		ActorEmail:   user.Email,
		TenantID:     &user.TenantID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestID:    r.Header.Get("X-Request-ID"),
		BeforeState:  beforeState,
		AfterState:   afterState,
		IPAddress:    middleware.GetRealIP(r),
		UserAgent:    r.UserAgent(),
		Timestamp:    time.Now(),
		Success:      success,
	}

	if err := h.auditRepo.LogAuditEvent(r.Context(), event); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"action":        action,
			"resource_type": resourceType,
		}).Error("Failed to log audit event")
	}
}
