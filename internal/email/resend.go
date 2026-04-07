package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/resend/resend-go/v2"
)

const (
	maxRetries    = 3
	retryBaseWait = 1 * time.Second
)

type ResendConfig struct {
	APIKey       string
	FromEmail    string
	FromName     string
	BaseURL      string
	AuthURL      string // Auth frontend URL for verification/reset links (e.g. https://auth.functionfly.com)
	ReplyToEmail string
}

type ResendService struct {
	client *resend.Client
	config ResendConfig
}

func NewResendService(config ResendConfig) *ResendService {
	return &ResendService{
		client: resend.NewClient(config.APIKey),
		config: config,
	}
}

// isRetryableError reports whether the error is a transient one (e.g. 429, 5xx) that callers may retry.
func (s *ResendService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Use typed error from Resend SDK for rate-limit detection
	if errors.Is(err, resend.ErrRateLimit) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "500") || strings.Contains(msg, "503")
}

func (s *ResendService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "Verify Your Email Address - FunctionFly"
	verifyBase := s.config.AuthURL
	if verifyBase == "" {
		verifyBase = s.config.BaseURL
	}
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", verifyBase, *user.VerificationToken)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Verify your email — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter','Segoe UI',system-ui,-apple-system,sans-serif;">
  <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <table role="presentation" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding-right:10px;vertical-align:middle;">
                <table role="presentation" width="32" height="32" cellpadding="0" cellspacing="0">
                  <tr><td style="background:#0F172A;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#6366F1;transform:rotate(45deg);margin:0 auto;"></div>
                  </td></tr>
                </table>
              </td>
              <td style="vertical-align:middle;font-size:18px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">FunctionFly</td>
            </tr>
          </table>
        </td></tr>
        <tr><td>
          <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background:#18181b;border:1px solid #27272a;border-radius:12px;overflow:hidden;">
            <tr><td style="padding:40px 40px 0;">
              <table role="presentation" cellpadding="0" cellspacing="0">
                <tr><td style="width:56px;height:56px;background:rgba(99,102,241,0.1);border-radius:50%%%%;text-align:center;vertical-align:middle;">
                  <div style="font-size:24px;line-height:56px;">&#9993;</div>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Verify your email</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Thanks for signing up for FunctionFly. Click the button below to verify your email address and activate your account.
              </p>
            </td></tr>
            <tr><td style="padding:28px 40px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:260px;" arcsize="17%%%%" stroke="f" fillcolor="#6366F1"><center style="color:#fff;font-size:15px;font-weight:600;">Verify email address</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#6366F1;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Verify email address</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    This link expires in <strong style="color:#a1a1aa;">24 hours</strong>. If you didn't create an account, you can safely ignore this email.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 40px;">
              <p style="margin:0;font-size:12px;color:#52525b;line-height:1.5;">
                Having trouble with the button? Copy and paste this link into your browser:<br>
                <a href="%s" style="color:#6366F1;text-decoration:none;word-break:break-all;">%s</a>
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, verificationURL, verificationURL, verificationURL, verificationURL, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`Verify your email

Thanks for signing up for FunctionFly. Click this link to verify your email and activate your account:

%s

This link expires in 24 hours. If you didn't create an account, ignore this email.
`, verificationURL)

	return s.SendEmail(user.Email, subject, textBody, htmlBody)
}

func (s *ResendService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Reset Your Password - FunctionFly"
	resetBase := s.config.AuthURL
	if resetBase == "" {
		resetBase = s.config.BaseURL
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", resetBase, resetToken)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Reset your password — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter','Segoe UI',system-ui,-apple-system,sans-serif;">
  <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <table role="presentation" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding-right:10px;vertical-align:middle;">
                <table role="presentation" width="32" height="32" cellpadding="0" cellspacing="0">
                  <tr><td style="background:#0F172A;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#6366F1;transform:rotate(45deg);margin:0 auto;"></div>
                  </td></tr>
                </table>
              </td>
              <td style="vertical-align:middle;font-size:18px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">FunctionFly</td>
            </tr>
          </table>
        </td></tr>
        <tr><td>
          <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background:#18181b;border:1px solid #27272a;border-radius:12px;overflow:hidden;">
            <tr><td style="padding:40px 40px 0;">
              <table role="presentation" cellpadding="0" cellspacing="0">
                <tr><td style="width:56px;height:56px;background:rgba(239,68,68,0.1);border-radius:50%%%%;text-align:center;vertical-align:middle;">
                  <div style="font-size:24px;line-height:56px;">&#128274;</div>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Reset your password</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We received a request to reset your password. Click the button below to choose a new one.
              </p>
            </td></tr>
            <tr><td style="padding:28px 40px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:220px;" arcsize="17%%%%" stroke="f" fillcolor="#6366F1"><center style="color:#fff;font-size:15px;font-weight:600;">Reset password</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#6366F1;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Reset password</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    This link expires in <strong style="color:#a1a1aa;">1 hour</strong>. If you didn't request a password reset, you can safely ignore this email.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 40px;">
              <p style="margin:0;font-size:12px;color:#52525b;line-height:1.5;">
                Having trouble with the button? Copy and paste this link into your browser:<br>
                <a href="%s" style="color:#6366F1;text-decoration:none;word-break:break-all;">%s</a>
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, resetURL, resetURL, resetURL, resetURL, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`Reset your password

We received a request to reset your FunctionFly password. Click this link to choose a new one:

%s

This link expires in 1 hour. If you didn't request this, ignore this email.
`, resetURL)

	return s.SendEmail(user.Email, subject, textBody, htmlBody)
}

func (s *ResendService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	breachType := "Data Breach"
	if v, ok := breachDetails["type"].(string); ok && v != "" {
		breachType = v
	}

	detectionTime := time.Now().Format(time.RFC3339)
	if v, ok := breachDetails["detectionTime"].(string); ok && v != "" {
		detectionTime = v
	}

	affectedUsers := 0
	if v, ok := breachDetails["affectedUsers"].(int); ok {
		affectedUsers = v
	}

	riskLevel := "high"
	if v, ok := breachDetails["riskLevel"].(string); ok && v != "" {
		riskLevel = v
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Data Breach Notification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .urgent { color: #dc3545; font-weight: bold; font-size: 18px; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 20px; margin: 20px 0; border-left: 4px solid #dc3545; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DATA BREACH NOTIFICATION</h1>
            <p class="urgent">GDPR Article 33 Compliance</p>
        </div>
        <div class="content">
            <h2>Urgent Security Incident</h2>
            <p>This is an official notification regarding a data breach that affects personal data processing activities.</p>

            <div class="details">
                <h3>Breach Details:</h3>
                <ul>
                    <li><strong>Type:</strong> %s</li>
                    <li><strong>Detection Time:</strong> %s</li>
                    <li><strong>Estimated Affected Users:</strong> %d</li>
                    <li><strong>Risk Level:</strong> %s</li>
                </ul>
            </div>

            <p><strong>Next Steps:</strong></p>
            <ul>
                <li>Notification to supervisory authority within 72 hours</li>
                <li>Communication to affected individuals (if high risk)</li>
                <li>Implementation of remedial measures</li>
                <li>Documentation of the incident</li>
            </ul>

            <p>For additional information or concerns, please contact our Data Protection Officer.</p>
        </div>
        <div class="footer">
            %s
            <p>This notification is sent in compliance with GDPR Article 33.</p>
        </div>
    </div>
</body>
</html>
`, breachType, detectionTime, affectedUsers, riskLevel, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`DATA BREACH NOTIFICATION - GDPR Article 33 Compliance

Urgent Security Incident

This is an official notification regarding a data breach that affects personal data processing activities.

Breach Details:
- Type: %s
- Detection Time: %s
- Estimated Affected Users: %d
- Risk Level: %s

Next Steps:
- Notification to supervisory authority within 72 hours
- Communication to affected individuals (if high risk)
- Implementation of remedial measures
- Documentation of the incident

For additional information or concerns, please contact our Data Protection Officer.

--
FunctionFly Security Team
This notification is sent in compliance with GDPR Article 33.
`, breachType, detectionTime, affectedUsers, riskLevel)

	return s.SendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

func (s *ResendService) SendEmail(to, subject, textBody, htmlBody string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = s.config.ReplyToEmail
	}

	return s.sendWithRetry(ctx, params)
}

func (s *ResendService) SendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      to,
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = s.config.ReplyToEmail
	}

	return s.sendWithRetry(ctx, params)
}

// sendWithRetry sends an email with exponential backoff retry for transient errors.
func (s *ResendService) sendWithRetry(ctx context.Context, params *resend.SendEmailRequest) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := s.client.Emails.SendWithContext(ctx, params)
		if err == nil {
			slog.Info("email sent via Resend",
				"to", params.To,
				"subject", params.Subject,
			)
			return nil
		}

		lastErr = err
		if !s.isRetryableError(err) {
			slog.Error("email send failed (non-retryable)",
				"to", params.To,
				"subject", params.Subject,
				"error", err,
			)
			return fmt.Errorf("failed to send email: %w", err)
		}

		wait := retryBaseWait * (1 << attempt) // 1s, 2s, 4s
		slog.Warn("email send failed, retrying",
			"to", params.To,
			"subject", params.Subject,
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"retry_in", wait,
			"error", err,
		)
		time.Sleep(wait)
	}

	slog.Error("email send failed after retries",
		"to", params.To,
		"subject", params.Subject,
		"attempts", maxRetries,
		"error", lastErr,
	)
	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, lastErr)
}

func (s *ResendService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>You're on the list — FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter','Segoe UI',system-ui,-apple-system,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <table role="presentation" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding-right:10px;vertical-align:middle;">
                <table role="presentation" width="32" height="32" cellpadding="0" cellspacing="0">
                  <tr><td style="background:#0F172A;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#6366F1;transform:rotate(45deg);margin:0 auto;"></div>
                  </td></tr>
                </table>
              </td>
              <td style="vertical-align:middle;font-size:18px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">FunctionFly</td>
            </tr>
          </table>
        </td></tr>
        <tr><td>
          <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#18181b;border:1px solid #27272a;border-radius:12px;overflow:hidden;">
            <tr><td style="padding:40px 40px 0;">
              <table role="presentation" cellpadding="0" cellspacing="0">
                <tr><td style="width:56px;height:56px;background:rgba(99,102,241,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                  <div style="font-size:24px;line-height:56px;">&#128227;</div>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">You're on the list!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Thanks for requesting early access to FunctionFly. We've received your request and will review it shortly.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We'll send you an invite code as soon as we're ready for more users. Hang tight!
              </p>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    We're rolling out access gradually to ensure the best experience for everyone. You'll be among the first to know when your spot is ready.
                  </p>
                </td></tr>
              </table>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`You're on the list — FunctionFly

Thanks for requesting early access to FunctionFly. We've received your request and will review it shortly.

We'll send you an invite code as soon as we're ready for more users. Hang tight!

We're rolling out access gradually to ensure the best experience for everyone. You'll be among the first to know when your spot is ready.
`, TransactionalEmailCopyrightPlain())

	return s.SendEmail(email, subject, textBody, htmlBody)
}

func (s *ResendService) ValidateConfiguration() error {
	if s.config.APIKey == "" {
		return fmt.Errorf("Resend API key not configured")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.ApiKeys.ListWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate Resend API key: %w", err)
	}

	return nil
}
