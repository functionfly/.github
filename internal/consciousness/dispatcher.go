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
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// NotificationDispatcher sends consciousness insights through configured channels.
type NotificationDispatcher struct {
	db         *sql.DB
	repo       *Repository
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewNotificationDispatcher creates a new notification dispatcher.
func NewNotificationDispatcher(db *sql.DB, repo *Repository, logger *logrus.Logger) *NotificationDispatcher {
	return &NotificationDispatcher{
		db:         db,
		repo:       repo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// Dispatch sends an insight through all channels enabled in the tenant's preferences.
func (d *NotificationDispatcher) Dispatch(ctx context.Context, insight *Insight, prefs *Preferences) []string {
	var sent []string

	if !severityMeetsThreshold(insight.Severity, prefs.MinNotifySeverity) {
		return sent
	}

	if !categoryEnabled(string(insight.Category), prefs.EnabledCategories) {
		return sent
	}

	// In-app (always first — fast, local DB write)
	if prefs.InAppEnabled {
		if err := d.sendInApp(ctx, insight); err != nil {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "in_app", "failed", err)
			d.logger.WithError(err).WithFields(logrus.Fields{
				"insight_id": insight.ID,
				"tenant_id":  insight.TenantID,
				"channel":    "in_app",
			}).Error("Failed to send consciousness in-app notification")
		} else {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "in_app", "sent", nil)
			sent = append(sent, "in_app")
		}
	}

	// Email
	if prefs.EmailEnabled {
		if err := d.sendEmail(ctx, insight); err != nil {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "email", "failed", err)
			d.logger.WithError(err).WithFields(logrus.Fields{
				"insight_id": insight.ID,
				"tenant_id":  insight.TenantID,
				"channel":    "email",
			}).Error("Failed to send consciousness email")
		} else {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "email", "sent", nil)
			sent = append(sent, "email")
		}
	}

	// Slack
	if prefs.SlackEnabled && prefs.SlackWebhookURL != nil && *prefs.SlackWebhookURL != "" {
		if err := d.sendSlack(ctx, insight, *prefs.SlackWebhookURL); err != nil {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "slack", "failed", err)
			d.logger.WithError(err).WithFields(logrus.Fields{
				"insight_id": insight.ID,
				"tenant_id":  insight.TenantID,
				"channel":    "slack",
			}).Error("Failed to send consciousness Slack notification")
		} else {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "slack", "sent", nil)
			sent = append(sent, "slack")
		}
	}

	// Webhook
	if prefs.WebhookEnabled && prefs.WebhookURL != nil && *prefs.WebhookURL != "" {
		secret := ""
		if prefs.WebhookSecret != nil {
			secret = *prefs.WebhookSecret
		}
		if err := d.sendWebhook(ctx, insight, *prefs.WebhookURL, secret); err != nil {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "webhook", "failed", err)
			d.logger.WithError(err).WithFields(logrus.Fields{
				"insight_id": insight.ID,
				"tenant_id":  insight.TenantID,
				"channel":    "webhook",
			}).Error("Failed to send consciousness webhook")
		} else {
			d.logDelivery(ctx, insight.ID, insight.TenantID, "webhook", "sent", nil)
			sent = append(sent, "webhook")
		}
	}

	return sent
}

// DispatchDigest sends a compiled digest of insights through configured channels.
func (d *NotificationDispatcher) DispatchDigest(ctx context.Context, digest *Digest, prefs *Preferences) []string {
	var sent []string

	if prefs.InAppEnabled {
		if err := d.sendDigestInApp(ctx, digest); err != nil {
			d.logger.WithError(err).WithField("tenant_id", digest.TenantID).Error("Failed to send consciousness digest in-app")
		} else {
			sent = append(sent, "in_app")
		}
	}

	if prefs.EmailEnabled {
		if err := d.sendDigestEmail(ctx, digest); err != nil {
			d.logger.WithError(err).WithField("tenant_id", digest.TenantID).Error("Failed to send consciousness digest email")
		} else {
			sent = append(sent, "email")
		}
	}

	if prefs.SlackEnabled && prefs.SlackWebhookURL != nil && *prefs.SlackWebhookURL != "" {
		if err := d.sendDigestSlack(ctx, digest, *prefs.SlackWebhookURL); err != nil {
			d.logger.WithError(err).WithField("tenant_id", digest.TenantID).Error("Failed to send consciousness digest Slack")
		} else {
			sent = append(sent, "slack")
		}
	}

	return sent
}

// ── In-App ──────────────────────────────────────────────────────────────────

func (d *NotificationDispatcher) sendInApp(ctx context.Context, insight *Insight) error {
	// Use PostgreSQL NOTIFY for real-time delivery via WebSocket hub.
	payload := map[string]interface{}{
		"type":      "consciousness.insight",
		"tenant_id": insight.TenantID.String(),
		"insight": map[string]interface{}{
			"id":        insight.ID.String(),
			"category":  insight.Category,
			"severity":  insight.Severity,
			"title":     insight.Title,
			"message":   insight.Message,
			"action":    insight.ActionType,
			"created_at": insight.CreatedAt.Format(time.RFC3339),
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// pg_notify on the tenant's channel
	_, err := d.db.ExecContext(ctx,
		"SELECT pg_notify($1, $2)",
		fmt.Sprintf("consciousness:%s", insight.TenantID.String()),
		string(payloadBytes),
	)
	return err
}

// ── Email ───────────────────────────────────────────────────────────────────

func (d *NotificationDispatcher) sendEmail(ctx context.Context, insight *Insight) error {
	// Look up the tenant owner's email
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

	// Use the existing email sending infrastructure
	return d.sendEmailRaw(ctx, email, subject, textBody, htmlBody)
}

// ── Slack ───────────────────────────────────────────────────────────────────

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

// ── Webhook (HMAC-SHA256 signed) ────────────────────────────────────────────

func (d *NotificationDispatcher) sendWebhook(ctx context.Context, insight *Insight, webhookURL, secret string) error {
	payload := map[string]interface{}{
		"event":     "consciousness.insight",
		"tenant_id": insight.TenantID.String(),
		"insight": map[string]interface{}{
			"id":        insight.ID.String(),
			"category":  insight.Category,
			"severity":  insight.Severity,
			"title":     insight.Title,
			"message":   insight.Message,
			"action":    insight.ActionType,
			"data":      insight.InsightData,
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Consciousness/1.0")

	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-FunctionFly-Signature", "sha256="+sig)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// ── Digest delivery ─────────────────────────────────────────────────────────

func (d *NotificationDispatcher) sendDigestInApp(ctx context.Context, digest *Digest) error {
	payload := map[string]interface{}{
		"type":      "consciousness.digest",
		"tenant_id": digest.TenantID.String(),
		"period":    digest.Period,
		"insight_count": len(digest.Insights),
		"score":     digest.Score,
		"generated_at": digest.GeneratedAt.Format(time.RFC3339),
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

// ── Internal helpers ────────────────────────────────────────────────────────

func (d *NotificationDispatcher) postSlackWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %s", resp.Status)
	}
	return nil
}

func (d *NotificationDispatcher) sendEmailRaw(ctx context.Context, to, subject, textBody, htmlBody string) error {
	// Uses the same SMTP/Resend infrastructure as the rest of the platform.
	// For now, we insert into the email queue table for async delivery.
	query := `
		INSERT INTO email_events (to_email, subject, text_body, html_body, status, created_at)
		VALUES ($1, $2, $3, $4, 'pending', NOW())
		RETURNING id`

	var id uuid.UUID
	err := d.db.QueryRowContext(ctx, query, to, subject, textBody, htmlBody).Scan(&id)
	if err != nil {
		// Fallback: if email_events table doesn't exist, log and skip
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

func (d *NotificationDispatcher) logDelivery(ctx context.Context, insightID, tenantID uuid.UUID, channel, status string, err error) {
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
	if logErr := d.repo.LogDelivery(ctx, log); logErr != nil {
		d.logger.WithError(logErr).Error("Failed to log delivery attempt")
	}
}

// ── Email HTML templates ────────────────────────────────────────────────────

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

// ── Utility functions ───────────────────────────────────────────────────────

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

// RetryDelivery attempts to deliver a previously failed notification.
func (d *NotificationDispatcher) RetryDelivery(ctx context.Context, retry *DeliveryRetry) error {
	// Reconstruct the insight from the stored payload
	var insight struct {
		ID        uuid.UUID
		TenantID  uuid.UUID
		Category  InsightCategory
		Severity  InsightSeverity
		Title     string
		Message   string
		ActionType ActionType
		CreatedAt time.Time
	}

	if err := json.Unmarshal(retry.Payload, &insight); err != nil {
		return fmt.Errorf("unmarshal retry payload: %w", err)
	}

	ins := &Insight{
		ID:         insight.ID,
		TenantID:   insight.TenantID,
		Category:   insight.Category,
		Severity:   insight.Severity,
		Title:      insight.Title,
		Message:    insight.Message,
		ActionType: insight.ActionType,
		CreatedAt:  insight.CreatedAt,
	}

	// Get tenant preferences
	prefs, err := d.repo.GetPreferences(ctx, retry.TenantID)
	if err != nil {
		return fmt.Errorf("get preferences: %w", err)
	}

	// Retry the specific channel
	switch retry.Channel {
	case "in_app":
		return d.sendInApp(ctx, ins)
	case "email":
		return d.sendEmail(ctx, ins)
	case "slack":
		if prefs.SlackWebhookURL != nil && *prefs.SlackWebhookURL != "" {
			return d.sendSlack(ctx, ins, *prefs.SlackWebhookURL)
		}
		return fmt.Errorf("slack webhook not configured")
	case "webhook":
		secret := ""
		if prefs.WebhookSecret != nil {
			secret = *prefs.WebhookSecret
		}
		if prefs.WebhookURL != nil && *prefs.WebhookURL != "" {
			return d.sendWebhook(ctx, ins, *prefs.WebhookURL, secret)
		}
		return fmt.Errorf("webhook URL not configured")
	default:
		return fmt.Errorf("unknown channel: %s", retry.Channel)
	}
}

// ScheduleRetryIfFailed checks if delivery failed and schedules a retry if appropriate.
func (d *NotificationDispatcher) ScheduleRetryIfFailed(ctx context.Context, insight *Insight, channel string, err error, maxRetries int) {
	if err == nil {
		return
	}

	// Serialize the insight for retry
	payload, marshalErr := json.Marshal(map[string]interface{}{
		"id":         insight.ID,
		"tenant_id":  insight.TenantID,
		"category":   insight.Category,
		"severity":   insight.Severity,
		"title":      insight.Title,
		"message":    insight.Message,
		"action":     insight.ActionType,
		"created_at": insight.CreatedAt,
	})
	if marshalErr != nil {
		d.logger.WithError(marshalErr).WithField("insight_id", insight.ID).Error("Failed to marshal insight for retry")
		return
	}

	errMsg := err.Error()
	retry := &DeliveryRetry{
		InsightID:    insight.ID,
		TenantID:     insight.TenantID,
		Channel:      channel,
		Payload:      payload,
		AttemptCount: 1,
		MaxAttempts:  maxRetries,
		NextRetryAt:  time.Now().Add(5 * time.Minute), // Initial delay
		LastError:    &errMsg,
	}

	if scheduleErr := d.repo.ScheduleRetry(ctx, retry); scheduleErr != nil {
		d.logger.WithError(scheduleErr).WithField("insight_id", insight.ID).Error("Failed to schedule retry")
	} else {
		d.logger.WithFields(logrus.Fields{
			"insight_id": insight.ID,
			"channel":    channel,
			"next_retry": retry.NextRetryAt,
		}).Info("Scheduled delivery retry")
	}
}
