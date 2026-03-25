package notification

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// EmailChannel handles email notification delivery
type EmailChannel struct {
	emailSvc email.Service
	logger   *logrus.Logger
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(emailSvc email.Service, logger *logrus.Logger) *EmailChannel {
	return &EmailChannel{
		emailSvc: emailSvc,
		logger:   logger,
	}
}

// Name returns the channel name
func (c *EmailChannel) Name() string {
	return ChannelEmail
}

// Send sends a notification via email
func (c *EmailChannel) Send(ctx context.Context, n *Notification, user *storage.User) error {
	if c.emailSvc == nil {
		return fmt.Errorf("email service not configured")
	}

	// Get user email if not provided
	userEmail := user.Email
	if userEmail == "" {
		return fmt.Errorf("user has no email address")
	}

	// Build email content
	subject := n.Title
	bodyHTML := c.buildHTMLBody(n)
	bodyText := c.buildTextBody(n)

	// Send the email using the email service
	err := c.emailSvc.SendEmail(userEmail, subject, bodyText, bodyHTML)
	if err != nil {
		c.logger.WithError(err).Error("Failed to send email notification")
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"to":                userEmail,
		"subject":           subject,
		"notification_type": n.Type,
	}).Info("Email notification sent successfully")

	return nil
}

// IsConfigured returns whether the channel is configured
func (c *EmailChannel) IsConfigured() bool {
	return c.emailSvc != nil
}

// buildHTMLBody builds the HTML email body
func (c *EmailChannel) buildHTMLBody(n *Notification) string {
	// Default HTML template
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
        </div>
        <div class="footer">
            %s
        </div>
    </div>
</body>
</html>
`, n.Title, n.Title, n.Body, email.TransactionalEmailCopyrightHTML())
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

