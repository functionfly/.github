package channels

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// WebhookChannel handles webhook notification delivery
type WebhookChannel struct {
	client *http.Client
	logger *logrus.Logger
	repo   notification.Repository
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(logger *logrus.Logger) *WebhookChannel {
	return &WebhookChannel{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Name returns the channel name
func (c *WebhookChannel) Name() string {
	return notification.ChannelWebhook
}

// Send sends a notification via webhook
func (c *WebhookChannel) Send(ctx context.Context, n *notification.Notification, user *storage.User) error {
	// Get user's webhook preferences
	pref, err := c.repo.GetPreference(ctx, n.UserID, notification.ChannelWebhook, n.Category)
	if err != nil {
		return fmt.Errorf("failed to get webhook preference: %w", err)
	}

	// If no preference or webhook not configured, skip
	if pref == nil || pref.WebhookURL == nil || *pref.WebhookURL == "" {
		c.logger.Debug("No webhook URL configured for user")
		return nil
	}

	// Check if webhook is enabled for this category
	if !pref.Enabled {
		c.logger.Debug("Webhook disabled for this category")
		return nil
	}

	// Build webhook payload
	payload := WebhookPayload{
		ID:        n.ID.String(),
		Type:      n.Type,
		Category:  n.Category,
		Title:     n.Title,
		Body:      n.Body,
		Data:      n.Data,
		Priority:  n.Priority,
		CreatedAt: n.CreatedAt,
		UserID:    n.UserID.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", *pref.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-FunctionFly-Event", n.Type)
	req.Header.Set("X-FunctionFly-Delivery", n.ID.String())

	// Sign payload if secret is configured
	if pref.WebhookSecret != nil && *pref.WebhookSecret != "" {
		signature := c.signPayload(payloadBytes, *pref.WebhookSecret)
		req.Header.Set("X-FunctionFly-Signature", signature)
	}

	// Send request with retries
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			c.logger.WithError(err).WithField("attempt", attempt+1).Warn("Webhook request failed")
			continue
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.logger.WithFields(logrus.Fields{
				"webhook_url": *pref.WebhookURL,
				"status_code": resp.StatusCode,
				"notification_id": n.ID,
			}).Debug("Webhook delivered successfully")
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"attempt":     attempt + 1,
		}).Warn("Webhook returned non-success status")

		// Don't retry on 4xx errors (client errors)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return fmt.Errorf("webhook delivery failed after retries: %w", lastErr)
}

// IsConfigured returns whether the channel is configured
func (c *WebhookChannel) IsConfigured() bool {
	return c.client != nil
}

// SetRepository sets the repository for the channel
func (c *WebhookChannel) SetRepository(repo notification.Repository) {
	c.repo = repo
}

// signPayload signs the payload with the webhook secret
func (c *WebhookChannel) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Category  string                 `json:"category"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	Priority  string                 `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
	UserID    string                 `json:"user_id"`
}

// WebhookResponse represents the expected response from webhook endpoints
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
