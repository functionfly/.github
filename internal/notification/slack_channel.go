package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	slackRateLimitDuration = 1 * time.Second
)

var slackMetrics = GetNotificationMetrics()

type SlackChannel struct {
	client      *http.Client
	logger      *logrus.Logger
	repo        Repository
	webhookURL  string
	rateLimiter *rateLimiter
}

type rateLimiter struct {
	mu       sync.Mutex
	lastSend time.Time
}

func NewSlackChannel(logger *logrus.Logger) *SlackChannel {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	return &SlackChannel{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:      logger,
		webhookURL:  webhookURL,
		rateLimiter: &rateLimiter{},
	}
}

func (c *SlackChannel) Name() string {
	return ChannelSlack
}

func (c *SlackChannel) IsConfigured() bool {
	return c.webhookURL != "" || c.repo != nil
}

func (c *SlackChannel) SetRepository(repo Repository) {
	c.repo = repo
}

func (c *SlackChannel) SetWebhookURL(url string) {
	c.webhookURL = url
}

func (c *SlackChannel) Send(ctx context.Context, n *Notification, user *storage.User) error {
	traceID := uuid.New().String()
	spanID := uuid.New().String()
	ctx = tracing.WithTraceContext(ctx, traceID, spanID, "")
	defer tracing.Finish(ctx)
	tracing.SetAttribute(ctx, "notification_id", n.ID.String())
	tracing.SetAttribute(ctx, "channel", ChannelSlack)

	webhookURL := c.webhookURL
	if webhookURL == "" && c.repo != nil {
		pref, err := c.repo.GetPreference(ctx, n.UserID, ChannelSlack, n.Category)
		if err == nil && pref != nil && pref.WebhookURL != nil {
			webhookURL = *pref.WebhookURL
		}
	}

	if webhookURL == "" {
		c.logger.Debug("No Slack webhook URL configured")
		return nil
	}

	severity := c.determineSeverity(n)
	if c.isInQuietHours(ctx, severity) && severity != "critical" {
		c.logger.WithFields(logrus.Fields{
			"notification_id": n.ID,
			"severity":        severity,
		}).Debug("Notification held during quiet hours")
		return nil
	}

	payload := BuildSlackPayload(n, severity)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	c.rateLimiter.mu.Lock()
	time.Sleep(slackRateLimitDuration)
	c.rateLimiter.mu.Unlock()

	var lastErr error
	retryBaseWait := time.Second

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoffDuration := retryBaseWait * time.Duration(1<<(attempt-1))
			time.Sleep(backoffDuration)
		}

		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = fmt.Errorf("failed to create Slack request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			lastErr = err
			slackMetrics.RecordWebhookLatency(duration)
			c.logger.WithError(err).WithField("attempt", attempt+1).Warn("Slack webhook request failed")
			continue
		}
		defer resp.Body.Close()

		slackMetrics.RecordWebhookLatency(duration)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slackMetrics.RecordWebhookResult("success")
			c.logger.WithFields(logrus.Fields{
				"notification_id": n.ID,
				"duration_ms":      duration.Milliseconds(),
			}).Debug("Slack notification delivered successfully")
			return nil
		}

		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				if sleepDur, err := time.ParseDuration(retryAfter + "s"); err == nil {
					time.Sleep(sleepDur)
				}
			}
			lastErr = fmt.Errorf("Slack rate limited")
			continue
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			slackMetrics.RecordWebhookError("client_error")
			lastErr = fmt.Errorf("Slack returned status %d", resp.StatusCode)
			break
		}

		lastErr = fmt.Errorf("Slack returned status %d", resp.StatusCode)
		slackMetrics.RecordRetry(ChannelSlack, "retry")
	}

	tracing.RecordError(ctx, lastErr)
	slackMetrics.RecordWebhookError("server_error")
	return fmt.Errorf("Slack delivery failed after retries: %w", lastErr)
}

func (c *SlackChannel) determineSeverity(n *Notification) string {
	switch n.Priority {
	case PriorityCritical:
		return "critical"
	case PriorityUrgent:
		return "high"
	case PriorityHigh:
		return "high"
	case PriorityNormal:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "medium"
	}
}

func (c *SlackChannel) isInQuietHours(ctx context.Context, severity string) bool {
	if severity == "critical" {
		return false
	}

	quietHoursConfig := os.Getenv("SLACK_QUIET_HOURS")
	if quietHoursConfig == "" {
		return false
	}

	var quietHours QuietHoursConfig
	if err := json.Unmarshal([]byte(quietHoursConfig), &quietHours); err != nil {
		return false
	}

	if !quietHours.Enabled {
		return false
	}

	loc, err := time.LoadLocation(quietHours.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentTime := now.Format("15:04")

	return currentTime >= quietHours.Start && currentTime <= quietHours.End
}

type QuietHoursConfig struct {
	Enabled  bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Timezone string `json:"timezone"`
}
