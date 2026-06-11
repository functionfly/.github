package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var webhookChannelMetrics = GetNotificationMetrics()

// WebhookChannel handles webhook notification delivery
type WebhookChannel struct {
	client *http.Client
	logger *logrus.Logger
	repo   Repository
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

// validateWebhookURL validates a webhook URL for security
// Prevents SSRF attacks by blocking private/internal IP ranges and requiring HTTPS
func validateWebhookURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS for security")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL must have a valid host")
	}

	host := parsedURL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("webhook URL cannot point to localhost")
	}

	if isPrivateOrInternalIP(host) {
		return fmt.Errorf("webhook URL cannot point to private or internal IP addresses")
	}

	return nil
}

// isPrivateOrInternalIP checks if a host is a private or internal IP
func isPrivateOrInternalIP(host string) bool {
	lowerHost := strings.ToLower(host)

	privatePrefixes := []string{
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"127.",
		"0.",
		"::1",
		"fc00:",
		"fe80:",
		"169.254.", // AWS metadata endpoint
		"metadata.google.internal.", // GCP metadata
	}

	for _, prefix := range privatePrefixes {
		if strings.HasPrefix(lowerHost, prefix) {
			return true
		}
	}

	return false
}

// Name returns the channel name
func (c *WebhookChannel) Name() string {
	return ChannelWebhook
}

// Send sends a notification via webhook
func (c *WebhookChannel) Send(ctx context.Context, n *Notification, user *storage.User) error {
	traceID := uuid.New().String()
	spanID := uuid.New().String()
	ctx = tracing.WithTraceContext(ctx, traceID, spanID, "")
	defer tracing.Finish(ctx)
	tracing.SetAttribute(ctx, "notification_id", n.ID.String())
	tracing.SetAttribute(ctx, "channel", ChannelWebhook)

	pref, err := c.repo.GetPreference(ctx, n.UserID, ChannelWebhook, n.Category)
	if err != nil {
		return fmt.Errorf("failed to get webhook preference: %w", err)
	}

	if pref == nil || pref.WebhookURL == nil || *pref.WebhookURL == "" {
		c.logger.Debug("No webhook URL configured for user")
		return nil
	}

	if !pref.Enabled {
		c.logger.Debug("Webhook disabled for this category")
		return nil
	}

	if err := validateWebhookURL(*pref.WebhookURL); err != nil {
		c.logger.WithError(err).WithField("webhook_url", *pref.WebhookURL).Warn("Invalid webhook URL rejected")
		return fmt.Errorf("webhook URL validation failed: %w", err)
	}

	tracing.SetAttribute(ctx, "webhook_url", *pref.WebhookURL)

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

	req, err := http.NewRequestWithContext(ctx, "POST", *pref.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-FunctionFly-Event", n.Type)
	req.Header.Set("X-FunctionFly-Delivery", n.ID.String())

	if pref.WebhookSecret != nil && *pref.WebhookSecret != "" {
		signature := c.signPayload(payloadBytes, *pref.WebhookSecret)
		req.Header.Set("X-FunctionFly-Signature", signature)
	}

	var lastErr error
	retryBaseWait := time.Second
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoffDuration := retryBaseWait * time.Duration(1<<(attempt-1))
			time.Sleep(backoffDuration)
		}

		start := time.Now()
		resp, err := c.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			lastErr = err
			webhookChannelMetrics.RecordWebhookLatency(duration)
			webhookChannelMetrics.RecordRetry(ChannelWebhook, "retry")
			c.logger.WithError(err).WithField("attempt", attempt+1).Warn("Webhook request failed")
			continue
		}
		defer resp.Body.Close()

		webhookChannelMetrics.RecordWebhookLatency(duration)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			contentType := resp.Header.Get("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") {
				c.logger.WithFields(logrus.Fields{
					"webhook_url":     *pref.WebhookURL,
					"content_type":   contentType,
					"notification_id": n.ID,
				}).Warn("Webhook response Content-Type is not application/json")
			}
			webhookChannelMetrics.RecordWebhookResult("success")
			c.logger.WithFields(logrus.Fields{
				"webhook_url":     *pref.WebhookURL,
				"status_code":     resp.StatusCode,
				"notification_id": n.ID,
				"duration_ms":      duration.Milliseconds(),
			}).Debug("Webhook delivered successfully")
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		webhookChannelMetrics.RecordWebhookResult("failure")

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			webhookChannelMetrics.RecordWebhookError("client_error")
			return lastErr
		}

		webhookChannelMetrics.RecordRetry(ChannelWebhook, "retry")
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"attempt":     attempt + 1,
		}).Warn("Webhook returned non-success status")
	}

	tracing.RecordError(ctx, lastErr)
	webhookChannelMetrics.RecordWebhookError("server_error")
	return fmt.Errorf("webhook delivery failed after retries: %w", lastErr)
}

// IsConfigured returns whether the channel is configured
func (c *WebhookChannel) IsConfigured() bool {
	return c.client != nil
}

// SetRepository sets the repository for the channel
func (c *WebhookChannel) SetRepository(repo Repository) {
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
