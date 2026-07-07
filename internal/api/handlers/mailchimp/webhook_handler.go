package mailchimpadmin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/mailchimp"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type WebhookHandler struct {
	mcClient *mailchimp.Client
	repo     storage.Repository
	logger   *logrus.Logger
}

func NewWebhookHandler(mcClient *mailchimp.Client, repo storage.Repository, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		mcClient: mcClient,
		repo:     repo,
		logger:   logger,
	}
}

type WebhookPayload struct {
	Type   string                 `json:"type"`
	Data   map[string]interface{} `json:"data"`
	FiredAt string               `json:"fired_at"`
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apierror.WriteError(w, apierror.NewBadRequest("Only POST allowed"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read webhook body")
		apierror.WriteError(w, apierror.NewBadRequest("Failed to read request body"))
		return
	}

	signature := r.Header.Get("X-Mailchimp-Signature")
	timestamp := r.Header.Get("X-Mailchimp-Timestamp")

	if !h.validateWebhook(signature, timestamp, string(body)) {
		h.logger.Warn("Invalid Mailchimp webhook signature")
		apierror.WriteError(w, apierror.NewUnauthorized("Invalid webhook signature"))
		return
	}

	payload, err := h.parsePayload(string(body))
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse webhook payload")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid webhook payload"))
		return
	}

	if err := h.processWebhook(r.Context(), payload); err != nil {
		h.logger.WithError(err).WithField("type", payload.Type).Error("Failed to process webhook")
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) validateWebhook(signature, timestamp, body string) bool {
	if h.mcClient == nil || h.mcClient.GetWebhookSecret() == "" {
		h.logger.Warn("Webhook validation skipped: no secret configured")
		return true
	}

	if signature == "" {
		return false
	}

	expectedSig := mailchimp.ComputeHMACSHA1(timestamp+body, h.mcClient.GetWebhookSecret())
	return signature == expectedSig
}

func (h *WebhookHandler) parsePayload(body string) (*WebhookPayload, error) {
	var payload WebhookPayload

	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()

	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

func (h *WebhookHandler) processWebhook(ctx context.Context, payload *WebhookPayload) error {
	switch payload.Type {
	case "subscribe":
		return h.handleSubscribe(ctx, payload)
	case "unsubscribe":
		return h.handleUnsubscribe(ctx, payload)
	case "campaign":
		return h.handleCampaign(ctx, payload)
	case "cleaned":
		return h.handleCleaned(ctx, payload)
	case "profile":
		return h.handleProfileUpdate(ctx, payload)
	case "email_type_change":
		return h.handleEmailTypeChange(ctx, payload)
	default:
		h.logger.WithField("type", payload.Type).Debug("Ignoring unknown webhook type")
		return nil
	}
}

func (h *WebhookHandler) handleSubscribe(ctx context.Context, payload *WebhookPayload) error {
	email := h.extractEmail(payload)
	if email == "" {
		return nil
	}

	h.logger.WithField("email", email).Info("Mailchimp webhook: subscribe event")
	return nil
}

func (h *WebhookHandler) handleUnsubscribe(ctx context.Context, payload *WebhookPayload) error {
	email := h.extractEmail(payload)
	if email == "" {
		return nil
	}

	h.logger.WithField("email", email).Info("Mailchimp webhook: unsubscribe event")

	if err := h.repo.UnsubscribeNewsletterSubscriber(ctx, email); err != nil {
		if err != storage.ErrSubscriberNotFound {
			return err
		}
	}

	return nil
}

func (h *WebhookHandler) handleCampaign(ctx context.Context, payload *WebhookPayload) error {
	campaignID, _ := payload.Data["campaign_id"].(string)
	subject := h.extractString(payload.Data, "subject")

	h.logger.WithFields(logrus.Fields{
		"campaign_id": campaignID,
		"subject":     subject,
	}).Debug("Mailchimp webhook: campaign event")

	return nil
}

func (h *WebhookHandler) handleCleaned(ctx context.Context, payload *WebhookPayload) error {
	email := h.extractEmail(payload)
	if email == "" {
		return nil
	}

	reason, _ := payload.Data["reason"].(string)

	h.logger.WithFields(logrus.Fields{
		"email":  email,
		"reason": reason,
	}).Info("Mailchimp webhook: cleaned (bounce/complaint) event")

	subscriber, err := h.repo.GetNewsletterSubscriberByEmail(ctx, email)
	if err != nil {
		if err == storage.ErrSubscriberNotFound {
			return nil
		}
		return err
	}

	if err := h.repo.MarkNewsletterSubscriberBounced(ctx, email); err != nil {
		return err
	}

	_ = subscriber
	return nil
}

func (h *WebhookHandler) handleProfileUpdate(ctx context.Context, payload *WebhookPayload) error {
	email := h.extractEmail(payload)
	if email == "" {
		return nil
	}

	h.logger.WithField("email", email).Info("Mailchimp webhook: profile update event")
	return nil
}

func (h *WebhookHandler) handleEmailTypeChange(ctx context.Context, payload *WebhookPayload) error {
	email := h.extractEmail(payload)
	if email == "" {
		return nil
	}

	h.logger.WithField("email", email).Debug("Mailchimp webhook: email type change event")
	return nil
}

func (h *WebhookHandler) extractEmail(payload *WebhookPayload) string {
	if email, ok := payload.Data["email"].(string); ok {
		return strings.ToLower(strings.TrimSpace(email))
	}
	return ""
}

func (h *WebhookHandler) extractString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func (h *WebhookHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/webhooks/mailchimp", h.HandleWebhook).Methods("POST", "OPTIONS")
}
