package mailchimpadmin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/mailchimp"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AdminHandler struct {
	mcClient    *mailchimp.Client
	syncService *mailchimp.SyncService
	repo        storage.Repository
	logger      *logrus.Logger
}

func NewAdminHandler(mcClient *mailchimp.Client, syncService *mailchimp.SyncService, repo storage.Repository, logger *logrus.Logger) *AdminHandler {
	return &AdminHandler{
		mcClient:    mcClient,
		syncService: syncService,
		repo:        repo,
		logger:      logger,
	}
}

type MailchimpStatsResponse struct {
	PlatformSubscribers int64                  `json:"platform_subscribers"`
	MailchimpStats      *mailchimp.ListStats   `json:"mailchimp_stats,omitempty"`
	SyncQueueLength    int64                  `json:"sync_queue_length"`
	SyncEnabled        bool                   `json:"sync_enabled"`
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.repo.GetNewsletterStats(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get platform newsletter stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get newsletter stats"))
		return
	}

	platformSubscribers, _ := stats["active_subscribers"].(int64)

	response := MailchimpStatsResponse{
		PlatformSubscribers: platformSubscribers,
		SyncEnabled:        h.mcClient != nil && h.mcClient.IsSyncEnabled(),
	}

	if h.mcClient != nil && h.mcClient.IsConfigured() {
		mcStats, err := h.mcClient.GetListStats(ctx)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to get Mailchimp stats")
		} else {
			response.MailchimpStats = mcStats
		}
	}

	if h.syncService != nil {
		queueLen, err := h.syncService.GetQueueLength(ctx)
		if err == nil {
			response.SyncQueueLength = queueLen
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type MailchimpSubscriberResponse struct {
	Email              string  `json:"email"`
	Name               string  `json:"name,omitempty"`
	Status             string  `json:"status"`
	MailchimpStatus    string  `json:"mailchimp_status,omitempty"`
	MailchimpSyncStatus string `json:"mailchimp_sync_status,omitempty"`
	SubscribedAt       string  `json:"subscribed_at,omitempty"`
	MailchimpLastSynced string `json:"mailchimp_last_synced,omitempty"`
	EmailFrequency     string  `json:"email_frequency,omitempty"`
}

type SubscriberListResponse struct {
	Subscribers []MailchimpSubscriberResponse `json:"subscribers"`
	Total       int64                       `json:"total"`
	Limit       int                         `json:"limit"`
	Offset      int                         `json:"offset"`
}

func (h *AdminHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()

	status := query.Get("status")
	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	subscribers, total, err := h.repo.ListNewsletterSubscribers(ctx, status, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list newsletter subscribers")
		apierror.WriteError(w, apierror.NewInternal("Failed to list subscribers"))
		return
	}

	responseSubscribers := make([]MailchimpSubscriberResponse, len(subscribers))
	for i, sub := range subscribers {
		responseSubscribers[i] = MailchimpSubscriberResponse{
			Email:               sub.Email,
			Name:                sub.Name,
			Status:              sub.Status,
			MailchimpSyncStatus: sub.MailchimpSyncStatus,
			SubscribedAt:        sub.SubscribedAt.Format("2006-01-02T15:04:05Z07:00"),
			EmailFrequency:      sub.EmailFrequency,
		}
		if sub.MailchimpLastSyncedAt != nil {
			responseSubscribers[i].MailchimpLastSynced = sub.MailchimpLastSyncedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		if h.mcClient != nil && h.mcClient.IsConfigured() && sub.Email != "" {
			mcSub, err := h.mcClient.GetSubscriber(ctx, sub.Email)
			if err == nil && mcSub != nil {
				responseSubscribers[i].MailchimpStatus = string(mcSub.Status)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SubscriberListResponse{
		Subscribers: responseSubscribers,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
	})
}

type SyncRequest struct {
	FullSync bool `json:"full_sync"`
}

func (h *AdminHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.syncService == nil {
		apierror.WriteError(w, apierror.NewInternal("Sync service not configured"))
		return
	}

	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	synced := 0
	failed := 0

	if req.FullSync {
		subscribers, err := h.repo.GetNewsletterSubscribersNeedingSync(ctx, 1000)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get subscribers needing sync")
			apierror.WriteError(w, apierror.NewInternal("Failed to get subscribers for sync"))
			return
		}

		for _, sub := range subscribers {
			mergeFields := mailchimp.MergeFields{
				"FNAME": sub.Name,
			}

			if err := h.syncService.EnqueueSubscribe(ctx, sub.ID.String(), sub.Email, mergeFields); err != nil {
				failed++
				continue
			}
			synced++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"synced_count": synced,
		"failed_count": failed,
	})
}

func (h *AdminHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var response struct {
		SyncEnabled    bool  `json:"sync_enabled"`
		QueueLength   int64 `json:"queue_length"`
		PendingCount  int64 `json:"pending_count"`
		FailedCount   int64 `json:"failed_count"`
	}

	response.SyncEnabled = h.mcClient != nil && h.mcClient.IsSyncEnabled()

	if h.syncService != nil {
		queueLen, err := h.syncService.GetQueueLength(ctx)
		if err == nil {
			response.QueueLength = queueLen
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) GetSubscriberActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Email is required"))
		return
	}

	if h.mcClient == nil || !h.mcClient.IsConfigured() {
		apierror.WriteError(w, apierror.NewInternal("Mailchimp not configured"))
		return
	}

	activity, err := h.mcClient.GetSubscriberActivity(ctx, email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subscriber activity")
		apierror.WriteError(w, apierror.NewInternal("Failed to get activity"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

func (h *AdminHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/admin/mailchimp/stats", h.GetStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/admin/mailchimp/subscribers", h.ListSubscribers).Methods("GET", "OPTIONS")
	router.HandleFunc("/admin/mailchimp/sync", h.TriggerSync).Methods("POST", "OPTIONS")
	router.HandleFunc("/admin/mailchimp/sync-status", h.GetSyncStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/admin/mailchimp/subscribers/{email}/activity", h.GetSubscriberActivity).Methods("GET", "OPTIONS")
}
