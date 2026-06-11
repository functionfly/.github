package consciousness

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	maxRetries        = 3
	retryBaseInterval = 5 * time.Minute
	httpTimeout       = 30 * time.Second
)

func validateWebhookURL(webhookURL string) error {
	parsedURL, err := url.Parse(webhookURL)
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

	if isPrivateIP(host) {
		return fmt.Errorf("webhook URL cannot point to private IP addresses")
	}

	return nil
}

func isPrivateIP(host string) bool {
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
	}

	for _, prefix := range privatePrefixes {
		if len(host) >= len(prefix) && host[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

type NotificationDispatcher struct {
	db         *sql.DB
	repo       *Repository
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewNotificationDispatcher(db *sql.DB, repo *Repository, logger *logrus.Logger) *NotificationDispatcher {
	return &NotificationDispatcher{
		db:   db,
		repo: repo,
		httpClient: &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

func (d *NotificationDispatcher) Close() error {
	d.httpClient.CloseIdleConnections()
	return nil
}

func (d *NotificationDispatcher) Dispatch(ctx context.Context, insight *Insight, prefs *Preferences) []string {
	var sent []string

	if !severityMeetsThreshold(insight.Severity, prefs.MinNotifySeverity) {
		return sent
	}

	if !categoryEnabled(string(insight.Category), prefs.EnabledCategories) {
		return sent
	}

	if prefs.InAppEnabled {
		if err := d.sendInApp(ctx, insight); err != nil {
			d.logDelivery(insight.ID, insight.TenantID, "in_app", "failed", err)
			d.logger.WithError(err).Warn("Failed to send consciousness in-app notification")
			d.enqueueRetry(ctx, insight, "in_app", err)
		} else {
			d.logDelivery(insight.ID, insight.TenantID, "in_app", "sent", nil)
			sent = append(sent, "in_app")
			dispatchTotal.WithLabelValues("in_app", "sent").Inc()
		}
	}

	if prefs.EmailEnabled {
		if err := d.sendEmail(ctx, insight); err != nil {
			d.logDelivery(insight.ID, insight.TenantID, "email", "failed", err)
			d.logger.WithError(err).Warn("Failed to send consciousness email")
			d.enqueueRetry(ctx, insight, "email", err)
		} else {
			d.logDelivery(insight.ID, insight.TenantID, "email", "sent", nil)
			sent = append(sent, "email")
			dispatchTotal.WithLabelValues("email", "sent").Inc()
		}
	}

	if prefs.SlackEnabled && prefs.SlackWebhookURL != nil && *prefs.SlackWebhookURL != "" {
		if err := d.sendSlack(ctx, insight, *prefs.SlackWebhookURL); err != nil {
			d.logDelivery(insight.ID, insight.TenantID, "slack", "failed", err)
			d.logger.WithError(err).Warn("Failed to send consciousness Slack notification")
			d.enqueueRetry(ctx, insight, "slack", err)
		} else {
			d.logDelivery(insight.ID, insight.TenantID, "slack", "sent", nil)
			sent = append(sent, "slack")
			dispatchTotal.WithLabelValues("slack", "sent").Inc()
		}
	}

	if prefs.WebhookEnabled && prefs.WebhookURL != nil && *prefs.WebhookURL != "" {
		secret := ""
		if prefs.WebhookSecret != nil {
			secret = *prefs.WebhookSecret
		}
		if err := d.sendWebhook(ctx, insight, *prefs.WebhookURL, secret); err != nil {
			d.logDelivery(insight.ID, insight.TenantID, "webhook", "failed", err)
			d.logger.WithError(err).Warn("Failed to send consciousness webhook")
			d.enqueueRetry(ctx, insight, "webhook", err)
		} else {
			d.logDelivery(insight.ID, insight.TenantID, "webhook", "sent", nil)
			sent = append(sent, "webhook")
			dispatchTotal.WithLabelValues("webhook", "sent").Inc()
		}
	}

	return sent
}

func (d *NotificationDispatcher) DispatchDigest(ctx context.Context, digest *Digest, prefs *Preferences) []string {
	var sent []string

	if prefs.InAppEnabled {
		if err := d.sendDigestInApp(ctx, digest); err != nil {
			d.logger.WithError(err).Warn("Failed to send consciousness digest in-app")
			dispatchTotal.WithLabelValues("in_app", "failed").Inc()
		} else {
			sent = append(sent, "in_app")
			dispatchTotal.WithLabelValues("in_app", "sent").Inc()
		}
	}

	if prefs.EmailEnabled {
		if err := d.sendDigestEmail(ctx, digest); err != nil {
			d.logger.WithError(err).Warn("Failed to send consciousness digest email")
			dispatchTotal.WithLabelValues("email", "failed").Inc()
		} else {
			sent = append(sent, "email")
			dispatchTotal.WithLabelValues("email", "sent").Inc()
		}
	}

	if prefs.SlackEnabled && prefs.SlackWebhookURL != nil && *prefs.SlackWebhookURL != "" {
		if err := d.sendDigestSlack(ctx, digest, *prefs.SlackWebhookURL); err != nil {
			d.logger.WithError(err).Warn("Failed to send consciousness digest Slack")
			dispatchTotal.WithLabelValues("slack", "failed").Inc()
		} else {
			sent = append(sent, "slack")
			dispatchTotal.WithLabelValues("slack", "sent").Inc()
		}
	}

	return sent
}

func (d *NotificationDispatcher) enqueueRetry(ctx context.Context, insight *Insight, channel string, err error) {
	payload, _ := json.Marshal(insight)
	nextRetry := time.Now().Add(retryBaseInterval)

	query := `
		INSERT INTO consciousness_dispatch_retry (
			insight_id, tenant_id, channel, payload, attempt_count, next_retry_at, last_error
		) VALUES ($1, $2, $3, $4, 1, $5, $6)
		ON CONFLICT (insight_id, channel) DO UPDATE SET
			attempt_count = consciousness_dispatch_retry.attempt_count + 1,
			next_retry_at = $5,
			last_error = $6,
			updated_at = NOW()`

	errMsg := err.Error()
	_, dbErr := d.db.ExecContext(ctx, query, insight.ID, insight.TenantID, channel, payload, nextRetry, errMsg)
	if dbErr != nil {
		d.logger.WithError(dbErr).Warn("Failed to enqueue dispatch retry")
	}

	dispatchTotal.WithLabelValues(channel, "failed").Inc()
}

func (d *NotificationDispatcher) sendInApp(ctx context.Context, insight *Insight) error {
	payload := map[string]interface{}{
		"type":      "consciousness.insight",
		"tenant_id": insight.TenantID.String(),
		"insight": map[string]interface{}{
			"id":         insight.ID.String(),
			"category":   insight.Category,
			"severity":   insight.Severity,
			"title":      insight.Title,
			"message":    insight.Message,
			"action":     insight.ActionType,
			"created_at": insight.CreatedAt.Format(time.RFC3339),
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	_, err := d.db.ExecContext(ctx,
		"SELECT pg_notify($1, $2)",
		fmt.Sprintf("consciousness:%s", insight.TenantID.String()),
		string(payloadBytes),
	)
	return err
}

func (d *NotificationDispatcher) sendEmail(ctx context.Context, insight *Insight) error {
	email, err := d.getTenantOwnerEmail(ctx, insight.TenantID)
	if err != nil || email == "" {
		return fmt.Errorf("tenant owner email not found: %w", err)
	}

	subject := fmt.Sprintf("[%s] %s — FunctionFly Consciousness",
		strings.ToUpper(string(insight.Severity)), insight.Title)

	textBody := fmt.Sprintf(
		"FunctionFly Consciousness Alert\n\n"+
			"Severity: %s\n"+
			"Category: %s\n\n"+
			"%s\n\n"+
			"View in dashboard: https://app.functionfly.dev/consciousness\n\n"+
			"— FunctionFly Consciousness",
		insight.Severity, insight.Category, insight.Message,
	)

	htmlBody := buildInsightEmailHTML(insight)

	return d.sendEmailRaw(ctx, email, subject, textBody, htmlBody)
}

func (d *NotificationDispatcher) sendSlack(ctx context.Context, insight *Insight, webhookURL string) error {
	severityEmoji := map[InsightSeverity]string{
		SeverityCritical:    ":red_circle:",
		SeverityWarning:     ":large_orange_circle:",
		SeverityOpportunity: ":large_green_circle:",
		SeverityInfo:        ":large_blue_circle:",
	}

	emoji := severityEmoji[insight.Severity]
	if emoji == "" {
		emoji = ":white_circle:"
	}

	payload := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": fmt.Sprintf("%s %s", emoji, insight.Title),
				},
			},
			{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": insight.Message,
				},
			},
			{
				"type": "section",
				"fields": []map[string]interface{}{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Severity:*\n%s", insight.Severity)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Category:*\n%s", insight.Category)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Confidence:*\n%.0f%%", derefFloat64(insight.Confidence)*100)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Action:*\n%s", actionLabel(insight.ActionType))},
				},
			},
			{
				"type": "actions",
				"elements": []map[string]interface{}{
					{
						"type":      "button",
						"text":      map[string]interface{}{"type": "plain_text", "text": "View in Dashboard"},
						"url":       "https://app.functionfly.dev/consciousness",
						"style":     "primary",
						"action_id": "view_consciousness",
					},
				},
			},
			{
				"type": "context",
				"elements": []map[string]interface{}{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("FunctionFly Consciousness • %s", time.Now().UTC().Format(time.RFC3339)),
					},
				},
			},
		},
	}

	return d.postSlackWebhook(ctx, webhookURL, payload)
}

func (d *NotificationDispatcher) sendWebhook(ctx context.Context, insight *Insight, webhookURL, secret string) error {
	if err := validateWebhookURL(webhookURL); err != nil {
		return fmt.Errorf("webhook URL validation failed: %w", err)
	}

	payload := map[string]interface{}{
		"event":     "consciousness.insight",
		"tenant_id": insight.TenantID.String(),
		"insight": map[string]interface{}{
			"id":         insight.ID.String(),
			"category":   insight.Category,
			"severity":   insight.Severity,
			"title":      insight.Title,
			"message":    insight.Message,
			"action":     insight.ActionType,
			"data":       insight.InsightData,
			"confidence": insight.Confidence,
			"trajectory": insight.Trajectory,
			"created_at": insight.CreatedAt.Format(time.RFC3339),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "FunctionFly-Consciousness/1.0")

		timestamp := time.Now().UTC().Format(time.RFC3339)
		if secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			mac.Write([]byte(timestamp))
			sig := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-FunctionFly-Signature", "sha256="+sig)
			req.Header.Set("X-FunctionFly-Timestamp", timestamp)
		}

		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			d.logger.WithError(err).Warnf("Webhook delivery attempt %d failed", attempt+1)
			continue
		}
		defer resp.Body.Close()
		defer io.Copy(io.Discard, resp.Body)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
			d.logger.WithError(lastErr).Warnf("Webhook delivery attempt %d failed", attempt+1)
			continue
		}
		return nil
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxRetries, lastErr)
}

func (d *NotificationDispatcher) sendDigestInApp(ctx context.Context, digest *Digest) error {
	payload := map[string]interface{}{
		"type":          "consciousness.digest",
		"tenant_id":     digest.TenantID.String(),
		"period":        digest.Period,
		"insight_count": len(digest.Insights),
		"score":         digest.Score,
		"generated_at":  digest.GeneratedAt.Format(time.RFC3339),
	}
	payloadBytes, _ := json.Marshal(payload)
	_, err := d.db.ExecContext(ctx,
		"SELECT pg_notify($1, $2)",
		fmt.Sprintf("consciousness:%s", digest.TenantID.String()),
		string(payloadBytes),
	)
	return err
}

func (d *NotificationDispatcher) sendDigestEmail(ctx context.Context, digest *Digest) error {
	email, err := d.getTenantOwnerEmail(ctx, digest.TenantID)
	if err != nil || email == "" {
		return fmt.Errorf("tenant owner email not found: %w", err)
	}

	scoreLabel := "N/A"
	var scoreVal float64
	if digest.Score != nil {
		scoreVal = digest.Score.OverallScore
		scoreLabel = ScoreLabel(scoreVal)
	}

	subject := fmt.Sprintf("[Digest] %d insights — Score: %.0f (%s) — FunctionFly",
		len(digest.Insights), scoreVal, scoreLabel)

	textBody := buildDigestText(digest)
	htmlBody := buildDigestHTML(digest)

	return d.sendEmailRaw(ctx, email, subject, textBody, htmlBody)
}

func (d *NotificationDispatcher) sendDigestSlack(ctx context.Context, digest *Digest, webhookURL string) error {
	var summaryLines []string
	for i, insight := range digest.Insights {
		if i >= 5 {
			summaryLines = append(summaryLines, fmt.Sprintf("...and %d more insights", len(digest.Insights)-5))
			break
		}
		severityEmoji := map[InsightSeverity]string{
			SeverityCritical: ":red_circle:", SeverityWarning: ":large_orange_circle:",
			SeverityOpportunity: ":large_green_circle:", SeverityInfo: ":large_blue_circle:",
		}
		emoji := severityEmoji[insight.Severity]
		if emoji == "" {
			emoji = ":white_circle:"
		}
		summaryLines = append(summaryLines, fmt.Sprintf("%s %s", emoji, insight.Title))
	}

	scoreText := "Not yet computed"
	if digest.Score != nil {
		scoreText = fmt.Sprintf("%.0f/100 (%s)", digest.Score.OverallScore, ScoreLabel(digest.Score.OverallScore))
	}

	payload := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": fmt.Sprintf(":brain: FunctionFly %s Digest — %d insights", digest.Period, len(digest.Insights)),
				},
			},
			{
				"type": "section",
				"fields": []map[string]interface{}{
					{"type": "mrkdwn", "text": fmt.Sprintf("*System Awareness Score:*\n%s", scoreText)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Period:*\n%s", digest.Period)},
				},
			},
			{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": strings.Join(summaryLines, "\n"),
				},
			},
			{
				"type": "actions",
				"elements": []map[string]interface{}{
					{
						"type":      "button",
						"text":      map[string]interface{}{"type": "plain_text", "text": "View All Insights"},
						"url":       "https://app.functionfly.dev/consciousness",
						"style":     "primary",
						"action_id": "view_digest",
					},
				},
			},
			{
				"type": "context",
				"elements": []map[string]interface{}{
					{"type": "mrkdwn", "text": fmt.Sprintf("FunctionFly Consciousness • %s", time.Now().UTC().Format(time.RFC3339))},
				},
			},
		},
	}

	return d.postSlackWebhook(ctx, webhookURL, payload)
}

func (d *NotificationDispatcher) postSlackWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	if err := validateWebhookURL(webhookURL); err != nil {
		return fmt.Errorf("slack webhook URL validation failed: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			d.logger.WithError(err).Warnf("Slack webhook delivery attempt %d failed", attempt+1)
			continue
		}
		defer resp.Body.Close()
		defer io.Copy(io.Discard, resp.Body)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("slack webhook returned status %s", resp.Status)
			d.logger.WithError(lastErr).Warnf("Slack webhook delivery attempt %d failed", attempt+1)
			continue
		}
		return nil
	}

	return fmt.Errorf("slack webhook delivery failed after %d attempts: %w", maxRetries, lastErr)
}

func (d *NotificationDispatcher) sendEmailRaw(ctx context.Context, to, subject, textBody, htmlBody string) error {
	query := `
		INSERT INTO email_events (to_email, subject, text_body, html_body, status, created_at)
		VALUES ($1, $2, $3, $4, 'pending', NOW())
		RETURNING id`

	var id uuid.UUID
	err := d.db.QueryRowContext(ctx, query, to, subject, textBody, htmlBody).Scan(&id)
	if err != nil {
		if isRelationNotExist(err) {
			d.logger.WithField("to", to).Info("Email queued (no email_events table, logged only)")
			return nil
		}
		return fmt.Errorf("queue email: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"email_id": id,
		"to":       to,
		"subject":  subject,
	}).Info("Consciousness email queued")
	return nil
}

func (d *NotificationDispatcher) getTenantOwnerEmail(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `
		SELECT u.email FROM users u
		JOIN tenants t ON t.id = u.tenant_id
		WHERE u.tenant_id = $1
		ORDER BY u.created_at ASC
		LIMIT 1`

	var email string
	err := d.db.QueryRowContext(ctx, query, tenantID).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func (d *NotificationDispatcher) logDelivery(insightID, tenantID uuid.UUID, channel, status string, err error) {
	log := &DeliveryLog{
		InsightID: insightID,
		TenantID:  tenantID,
		Channel:   channel,
		Status:    status,
		SentAt:    time.Now(),
	}
	if err != nil {
		errMsg := err.Error()
		log.ErrorMsg = &errMsg
	}
	if logErr := d.repo.LogDelivery(context.Background(), log); logErr != nil {
		d.logger.WithError(logErr).Warn("Failed to log delivery attempt")
	}
}

func buildInsightEmailHTML(insight *Insight) string {
	severityColor := map[InsightSeverity]string{
		SeverityCritical:    "#ef4444",
		SeverityWarning:     "#f97316",
		SeverityOpportunity: "#22c55e",
		SeverityInfo:        "#3b82f6",
	}
	color := severityColor[insight.Severity]
	if color == "" {
		color = "#6b7280"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#0a0a0b;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#0a0a0b;padding:40px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;">
  <tr><td style="padding:20px;text-align:center;">
    <span style="font-size:24px;font-weight:700;color:#f97316;">&#9670; FunctionFly</span>
  </td></tr>
  <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:32px;">
    <div style="display:inline-block;padding:4px 12px;border-radius:6px;background:%s20;color:%s;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:0.5px;">%s</div>
    <h2 style="color:#fafafa;font-size:20px;margin:16px 0 8px;">%s</h2>
    <p style="color:#a1a1aa;font-size:14px;line-height:1.6;margin:0 0 24px;">%s</p>
    <table width="100%%" style="border-collapse:collapse;">
      <tr>
        <td style="padding:8px 0;color:#71717a;font-size:13px;">Category</td>
        <td style="padding:8px 0;color:#fafafa;font-size:13px;text-align:right;">%s</td>
      </tr>
      <tr>
        <td style="padding:8px 0;color:#71717a;font-size:13px;">Severity</td>
        <td style="padding:8px 0;color:%s;font-size:13px;text-align:right;font-weight:600;">%s</td>
      </tr>
      <tr>
        <td style="padding:8px 0;color:#71717a;font-size:13px;">Action</td>
        <td style="padding:8px 0;color:#fafafa;font-size:13px;text-align:right;">%s</td>
      </tr>
    </table>
    <div style="margin-top:24px;text-align:center;">
      <a href="https://app.functionfly.dev/consciousness" style="display:inline-block;padding:12px 32px;background:#f97316;color:#fff;text-decoration:none;border-radius:8px;font-weight:600;font-size:14px;">View in Dashboard</a>
    </div>
  </td></tr>
  <tr><td style="padding:16px;text-align:center;">
    <p style="color:#52525b;font-size:12px;">FunctionFly Consciousness &bull; %s</p>
  </td></tr>
</table>
</td></tr></table>
</body></html>`,
		color, color, strings.ToUpper(string(insight.Severity)),
		insight.Title,
		insight.Message,
		string(insight.Category),
		color, string(insight.Severity),
		actionLabel(insight.ActionType),
		time.Now().UTC().Format("Jan 2, 2006"),
	)
}

func buildDigestText(digest *Digest) string {
	var b strings.Builder
	b.WriteString("FunctionFly Consciousness Digest\n")
	b.WriteString("================================\n\n")

	if digest.Score != nil {
		b.WriteString(fmt.Sprintf("System Awareness Score: %.0f/100 (%s)\n\n",
			digest.Score.OverallScore, ScoreLabel(digest.Score.OverallScore)))
	}

	b.WriteString(fmt.Sprintf("Active Insights: %d\n\n", len(digest.Insights)))

	for i, insight := range digest.Insights {
		b.WriteString(fmt.Sprintf("%d. [%s] %s\n   %s\n\n",
			i+1, strings.ToUpper(string(insight.Severity)),
			insight.Title, insight.Message))
	}

	b.WriteString("\nView all insights: https://app.functionfly.dev/consciousness\n")
	b.WriteString("\n— FunctionFly Consciousness\n")
	return b.String()
}

func buildDigestHTML(digest *Digest) string {
	var insightRows string
	for i, insight := range digest.Insights {
		if i >= 10 {
			insightRows += fmt.Sprintf(`<tr><td colspan="2" style="padding:12px;color:#71717a;font-size:13px;text-align:center;">...and %d more insights</td></tr>`, len(digest.Insights)-10)
			break
		}
		severityColor := map[InsightSeverity]string{
			SeverityCritical: "#ef4444", SeverityWarning: "#f97316",
			SeverityOpportunity: "#22c55e", SeverityInfo: "#3b82f6",
		}
		c := severityColor[insight.Severity]
		if c == "" {
			c = "#6b7280"
		}
		insightRows += fmt.Sprintf(`
		<tr>
			<td style="padding:12px 8px;border-bottom:1px solid #27272a;"><span style="display:inline-block;width:8px;height:8px;border-radius:50%%;background:%s;margin-right:8px;"></span><span style="color:#fafafa;font-size:13px;">%s</span></td>
			<td style="padding:12px 8px;border-bottom:1px solid #27272a;color:#a1a1aa;font-size:12px;text-align:right;">%s</td>
		</tr>`, c, insight.Title, strings.ToUpper(string(insight.Severity)))
	}

	scoreText := "Not yet computed"
	if digest.Score != nil {
		scoreText = fmt.Sprintf("%.0f/100 (%s)", digest.Score.OverallScore, ScoreLabel(digest.Score.OverallScore))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#0a0a0b;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#0a0a0b;padding:40px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;">
  <tr><td style="padding:20px;text-align:center;">
    <span style="font-size:24px;font-weight:700;color:#f97316;">&#9670; FunctionFly</span>
    <p style="color:#a1a1aa;font-size:14px;margin:8px 0 0;">%s Consciousness Digest</p>
  </td></tr>
  <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:32px;">
    <table width="100%%" style="margin-bottom:24px;">
      <tr>
        <td style="padding:16px;background:#0f0f12;border-radius:8px;text-align:center;">
          <div style="color:#f97316;font-size:32px;font-weight:700;">%s</div>
          <div style="color:#71717a;font-size:12px;margin-top:4px;">System Awareness Score</div>
        </td>
        <td width="16"></td>
        <td style="padding:16px;background:#0f0f12;border-radius:8px;text-align:center;">
          <div style="color:#fafafa;font-size:32px;font-weight:700;">%d</div>
          <div style="color:#71717a;font-size:12px;margin-top:4px;">Active Insights</div>
        </td>
      </tr>
    </table>
    <h3 style="color:#fafafa;font-size:16px;margin:0 0 12px;">Top Insights</h3>
    <table width="100%%" style="border-collapse:collapse;">%s</table>
    <div style="margin-top:24px;text-align:center;">
      <a href="https://app.functionfly.dev/consciousness" style="display:inline-block;padding:12px 32px;background:#f97316;color:#fff;text-decoration:none;border-radius:8px;font-weight:600;font-size:14px;">View All Insights</a>
    </div>
  </td></tr>
  <tr><td style="padding:16px;text-align:center;">
    <p style="color:#52525b;font-size:12px;">FunctionFly Consciousness &bull; %s</p>
  </td></tr>
</table>
</td></tr></table>
</body></html>`,
		strings.Title(digest.Period),
		scoreText,
		len(digest.Insights),
		insightRows,
		time.Now().UTC().Format("Jan 2, 2006"),
	)
}

func severityMeetsThreshold(severity InsightSeverity, threshold string) bool {
	weight := map[string]int{
		"info": 1, "warning": 2, "opportunity": 2, "critical": 3,
	}
	return weight[string(severity)] >= weight[threshold]
}

func categoryEnabled(category string, enabled []string) bool {
	for _, c := range enabled {
		if c == category {
			return true
		}
	}
	return false
}

func actionLabel(action ActionType) string {
	switch action {
	case ActionMergeFunctions:
		return "Merge Functions"
	case ActionScaleConfig:
		return "Adjust Scaling"
	case ActionSwapMarketplace:
		return "Swap to Marketplace"
	case ActionOptimize:
		return "Optimize"
	default:
		return "None"
	}
}

func derefFloat64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
