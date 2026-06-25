package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
)

type subscribePushRequest struct {
	Endpoint string  `json:"endpoint"`
	P256DH   string  `json:"p256dh"`
	Auth     string  `json:"auth"`
	UserAgent *string `json:"user_agent,omitempty"`
}

func (h *Handler) HandleSubscribePush(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req subscribePushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Endpoint == "" || req.P256DH == "" || req.Auth == "" {
		apierror.WriteError(w, apierror.NewBadRequest("endpoint, p256dh, and auth are required"))
		return
	}

	ps := &storage.PushSubscription{
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		Endpoint:  req.Endpoint,
		P256DH:    req.P256DH,
		Auth:      req.Auth,
		UserAgent: req.UserAgent,
		IsActive:  true,
	}

	created, err := h.repo.CreatePushSubscription(r.Context(), ps)
	if err != nil {
		h.log.WithError(err).Error("Failed to create push subscription")
		apierror.WriteError(w, apierror.NewInternal("Failed to create push subscription"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscription": created,
	})
}

type upsertNotificationPrefRequest struct {
	Channel         string  `json:"channel"`
	EventType       string  `json:"event_type"`
	IsEnabled       *bool   `json:"is_enabled,omitempty"`
	QuietHoursStart *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty"`
}

func (h *Handler) HandleUpdateNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req upsertNotificationPrefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Channel == "" || req.EventType == "" {
		apierror.WriteError(w, apierror.NewBadRequest("channel and event_type are required"))
		return
	}

	pref := &storage.NotificationPreference{
		UserID:          claims.UserID,
		TenantID:        claims.TenantID,
		Channel:         req.Channel,
		EventType:       req.EventType,
		IsEnabled:       true,
		QuietHoursStart: req.QuietHoursStart,
		QuietHoursEnd:   req.QuietHoursEnd,
	}
	if req.IsEnabled != nil {
		pref.IsEnabled = *req.IsEnabled
	}

	created, err := h.repo.UpsertNotificationPreference(r.Context(), pref)
	if err != nil {
		h.log.WithError(err).Error("Failed to update notification preference")
		apierror.WriteError(w, apierror.NewInternal("Failed to update notification preference"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preference": created,
	})
}

func (h *Handler) HandleGetNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListNotificationPreferencesOpts{
		Limit:  50,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if c := q.Get("channel"); c != "" {
		opts.Channel = &c
	}

	prefs, total, err := h.repo.ListNotificationPreferences(r.Context(), claims.UserID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list notification preferences")
		apierror.WriteError(w, apierror.NewInternal("Failed to list notification preferences"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preferences": prefs,
		"total":       total,
		"limit":       opts.Limit,
		"offset":      opts.Offset,
	})
}
