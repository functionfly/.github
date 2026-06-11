package notification

import (
	"context"
	"fmt"
	"html"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var emailChannelMetrics = GetNotificationMetrics()

// EmailChannel handles email notification delivery
type EmailChannel struct {
	emailSvc    email.Service
	logger      *logrus.Logger
	templateEng *TemplateEngine
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(emailSvc email.Service, logger *logrus.Logger) *EmailChannel {
	return &EmailChannel{
		emailSvc:    emailSvc,
		logger:      logger,
		templateEng: NewTemplateEngine(nil),
	}
}

// Name returns the channel name
func (c *EmailChannel) Name() string {
	return ChannelEmail
}

// Send sends a notification via email
func (c *EmailChannel) Send(ctx context.Context, n *Notification, user *storage.User) error {
	traceID := uuid.New().String()
	spanID := uuid.New().String()
	ctx = tracing.WithTraceContext(ctx, traceID, spanID, "")
	defer tracing.Finish(ctx)
	tracing.SetAttribute(ctx, "notification_id", n.ID.String())
	tracing.SetAttribute(ctx, "channel", ChannelEmail)

	if c.emailSvc == nil {
		tracing.AddEvent(ctx, "channel_skipped", map[string]interface{}{"reason": "email_service_not_configured"})
		c.logger.WithField("notification_id", n.ID).Debug("Email channel skipped: service not configured")
		return nil
	}

	userEmail := user.Email
	if userEmail == "" {
		tracing.AddEvent(ctx, "channel_skipped", map[string]interface{}{"reason": "user_has_no_email"})
		c.logger.WithField("notification_id", n.ID).Debug("Email channel skipped: user has no email")
		return nil
	}

	tracing.SetAttribute(ctx, "to", userEmail)

	subject := n.Title
	bodyHTML := c.buildHTMLBody(n)
	bodyText := c.buildTextBody(n)

	start := time.Now()
	err := c.emailSvc.SendEmail(userEmail, subject, bodyText, bodyHTML)
	duration := time.Since(start)

	emailChannelMetrics.RecordEmailLatency(duration)

	if err != nil {
		tracing.RecordError(ctx, err)
		emailChannelMetrics.RecordEmailError(classifyError(err))
		emailChannelMetrics.RecordEmailResult("failure")
		c.logger.WithError(err).Error("Failed to send email notification")
		return fmt.Errorf("failed to send email: %w", err)
	}

	emailChannelMetrics.RecordEmailResult("success")
	c.logger.WithFields(logrus.Fields{
		"to":                userEmail,
		"subject":           subject,
		"notification_type": n.Type,
		"duration_ms":       duration.Milliseconds(),
	}).Info("Email notification sent successfully")

	return nil
}

// IsConfigured returns whether the channel is configured
func (c *EmailChannel) IsConfigured() bool {
	return c.emailSvc != nil
}

// buildHTMLBody builds the HTML email body with rich data support
func (c *EmailChannel) buildHTMLBody(n *Notification) string {
	if c.templateEng != nil && n.Data != nil {
		_, bodyHTML, _, _ := c.templateEng.Render(context.Background(), n.Type, ChannelEmail, n.Data)
		if bodyHTML != "" {
			return bodyHTML
		}
	}

	dataHTML := ""
	if n.Data != nil {
		for key, value := range n.Data {
			safeKey := html.EscapeString(key)
			safeValue := html.EscapeString(fmt.Sprintf("%v", value))
			dataHTML += fmt.Sprintf(`<p class="data-item"><strong>%s:</strong> %s</p>`, safeKey, safeValue)
		}
	}

	safeTitle := html.EscapeString(n.Title)
	safeBody := html.EscapeString(n.Body)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .data-item { font-size: 14px; color: #666; margin: 5px 0; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>FunctionFly</h1>
        </div>
        <div class="content">
            <h2>%s</h2>
            <p>%s</p>
            %s
        </div>
        <div class="footer">
            %s
        </div>
    </div>
</body>
</html>
`, safeTitle, safeTitle, safeBody, dataHTML, email.TransactionalEmailCopyrightHTML())
}

// buildTextBody builds the plain text email body
func (c *EmailChannel) buildTextBody(n *Notification) string {
	return fmt.Sprintf(`FunctionFly Notification

%s

%s

---
%s
`, n.Title, n.Body, email.TransactionalEmailCopyrightPlain())
}

