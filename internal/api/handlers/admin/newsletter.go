package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type NewsletterHandler struct {
	repo         storage.Repository
	emailService email.Service
}

func NewNewsletterHandler(repo storage.Repository, emailService email.Service) *NewsletterHandler {
	return &NewsletterHandler{
		repo:         repo,
		emailService: emailService,
	}
}

type SubscriberListResponse struct {
	Subscribers []storage.NewsletterSubscriber `json:"subscribers"`
	Total       int64                          `json:"total"`
	Limit       int                            `json:"limit"`
	Offset      int                            `json:"offset"`
}

func (h *NewsletterHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
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

	subscribers, total, err := h.repo.ListNewsletterSubscribers(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list newsletter subscribers")
		http.Error(w, `{"error": "Failed to list subscribers"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SubscriberListResponse{
		Subscribers: subscribers,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
	})
}

func (h *NewsletterHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetNewsletterStats(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get newsletter stats")
		http.Error(w, `{"error": "Failed to get newsletter stats"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

type CreateSubscriberRequest struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
}

func (h *NewsletterHandler) CreateSubscriber(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	if emailAddr == "" {
		http.Error(w, `{"error": "Email is required"}`, http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	source := req.Source
	if source == "" {
		source = "admin"
	}
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	subscriber, err := h.repo.CreateNewsletterSubscriber(r.Context(), emailAddr, name, source, ipAddress, userAgent)
	if err != nil {
		if err == storage.ErrSubscriberExists {
			http.Error(w, `{"error": "Email is already subscribed"}`, http.StatusConflict)
			return
		}
		logrus.WithError(err).Error("Failed to create newsletter subscriber")
		http.Error(w, `{"error": "Failed to create subscriber"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"subscriber": subscriber,
	})
}

func (h *NewsletterHandler) DeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid subscriber ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteNewsletterSubscriber(r.Context(), id); err != nil {
		if err == storage.ErrSubscriberNotFound {
			http.Error(w, `{"error": "Subscriber not found"}`, http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to delete newsletter subscriber")
		http.Error(w, `{"error": "Failed to delete subscriber"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Subscriber deleted successfully",
	})
}

type CampaignListResponse struct {
	Campaigns []storage.NewsletterCampaign `json:"campaigns"`
	Total     int64                        `json:"total"`
	Limit     int                          `json:"limit"`
	Offset    int                          `json:"offset"`
}

func (h *NewsletterHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
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

	campaigns, total, err := h.repo.ListNewsletterCampaigns(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list newsletter campaigns")
		http.Error(w, `{"error": "Failed to list campaigns"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CampaignListResponse{
		Campaigns: campaigns,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	})
}

type CreateCampaignRequest struct {
	Subject     string `json:"subject"`
	PreviewText string `json:"preview_text,omitempty"`
	Content     string `json:"content"`
}

func (h *NewsletterHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Subject) == "" {
		http.Error(w, `{"error": "Subject is required"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, `{"error": "Content is required"}`, http.StatusBadRequest)
		return
	}

	adminID := getAdminID(r.Context())

	campaign := &storage.NewsletterCampaign{
		Subject:     strings.TrimSpace(req.Subject),
		PreviewText: strings.TrimSpace(req.PreviewText),
		Content:     req.Content,
		HTMLContent: req.Content,
		Status:      "draft",
		CreatedBy:   &adminID,
	}

	c, err := h.repo.CreateNewsletterCampaign(r.Context(), campaign)
	if err != nil {
		logrus.WithError(err).Error("Failed to create newsletter campaign")
		http.Error(w, `{"error": "Failed to create campaign"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"campaign": c,
	})
}

func (h *NewsletterHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid campaign ID"}`, http.StatusBadRequest)
		return
	}

	campaign, err := h.repo.GetNewsletterCampaignByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrCampaignNotFound {
			http.Error(w, `{"error": "Campaign not found"}`, http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get newsletter campaign")
		http.Error(w, `{"error": "Failed to get campaign"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaign)
}

type SendCampaignRequest struct {
	ScheduleAt string `json:"schedule_at,omitempty"`
}

func (h *NewsletterHandler) SendCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid campaign ID"}`, http.StatusBadRequest)
		return
	}

	campaign, err := h.repo.GetNewsletterCampaignByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrCampaignNotFound {
			http.Error(w, `{"error": "Campaign not found"}`, http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get newsletter campaign")
		http.Error(w, `{"error": "Failed to get campaign"}`, http.StatusInternalServerError)
		return
	}

	if campaign.Status == "sent" {
		http.Error(w, `{"error": "Campaign has already been sent"}`, http.StatusBadRequest)
		return
	}

	subscribers, err := h.repo.GetActiveNewsletterSubscribers(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get active subscribers")
		http.Error(w, `{"error": "Failed to get subscribers"}`, http.StatusInternalServerError)
		return
	}

	if len(subscribers) == 0 {
		http.Error(w, `{"error": "No active subscribers to send to"}`, http.StatusBadRequest)
		return
	}

	emailAddresses := make([]string, len(subscribers))
	for i, sub := range subscribers {
		emailAddresses[i] = sub.Email
	}

	now := time.Now()
	campaign.SentAt = &now
	campaign.Status = "sent"
	if _, err := h.repo.UpdateNewsletterCampaign(r.Context(), id, map[string]interface{}{
		"status":  "sent",
		"sent_at": now,
	}); err != nil {
		logrus.WithError(err).Error("Failed to update campaign status")
	}

	go func() {
		if err := h.emailService.SendNewsletterCampaign(emailAddresses, campaign.Subject, campaign.PreviewText, campaign.HTMLContent); err != nil {
			logrus.WithError(err).Error("Failed to send newsletter campaign emails")
			h.repo.UpdateNewsletterCampaign(r.Context(), id, map[string]interface{}{
				"status":        "failed",
				"error_message": err.Error(),
			})
			return
		}

		h.repo.UpdateCampaignStats(r.Context(), id)
		logrus.Infof("Newsletter campaign %s sent to %d subscribers", id, len(subscribers))
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success:":         true,
		"message":          "Campaign is being sent",
		"subscriber_count": len(subscribers),
	})
}

func getAdminID(ctx context.Context) uuid.UUID {
	return uuid.Nil
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	forwarded = r.Header.Get("X-Real-IP")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
