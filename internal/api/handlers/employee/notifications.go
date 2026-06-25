package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

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

	notifications, total, err := h.repo.ListNotifications(r.Context(), claims.UserID, unreadOnly, limit, offset)
	if err != nil {
		h.log.WithError(err).Warn("employee notifications: failed to list")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"notifications": []interface{}{},
			"total":        0,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"total":        total,
	})
}

func (h *Handler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	count, err := h.repo.CountUnreadNotifications(r.Context(), claims.UserID)
	if err != nil {
		h.log.WithError(err).Warn("employee notifications: failed to count unread")
		count = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"unread_count": count,
	})
}

func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid notification ID"))
		return
	}

	if err := h.repo.MarkNotificationRead(r.Context(), id); err != nil {
		h.log.WithError(err).WithField("notification_id", id).Warn("employee notifications: failed to mark read")
		apierror.WriteError(w, apierror.NewInternal("Failed to mark notification read"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if err := h.repo.MarkAllNotificationsRead(r.Context(), claims.UserID); err != nil {
		h.log.WithError(err).WithField("user_id", claims.UserID).Warn("employee notifications: failed to mark all read")
		apierror.WriteError(w, apierror.NewInternal("Failed to mark all notifications read"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
