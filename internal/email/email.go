package email

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Service defines the interface for email operations
type Service interface {
	SendVerificationEmail(user *storage.User, verificationToken string) error
	SendPasswordResetEmail(user *storage.User, resetToken string) error
	SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error
	SendWaitlistConfirmationEmail(email string) error
	SendEmail(to, subject, textBody, htmlBody string) error
	ValidateConfiguration() error
}

// Config holds email service configuration
type Config struct {
	Provider      string // "resend" or "smtp"
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	ResendAPIKey string
	FromEmail    string
	FromName     string
	BaseURL      string
	AuthURL      string // Auth frontend URL (e.g. https://auth.functionfly.com) for verification/reset links
	ReplyToEmail string
}

// NewService creates an email service based on configuration
func NewService(config Config) Service {
	if config.Provider == "resend" && config.ResendAPIKey != "" {
		return NewResendService(ResendConfig{
			APIKey:       config.ResendAPIKey,
			FromEmail:    config.FromEmail,
			FromName:     config.FromName,
			BaseURL:      config.BaseURL,
			AuthURL:      config.AuthURL,
			ReplyToEmail: config.ReplyToEmail,
		})
	}
	// Default to SMTP
	return NewSMTPService(Config{
		SMTPHost:     config.SMTPHost,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.SMTPUsername,
		SMTPPassword: config.SMTPPassword,
		FromEmail:    config.FromEmail,
		FromName:     config.FromName,
		BaseURL:      config.BaseURL,
		AuthURL:      config.AuthURL,
	})
}

// SMTPService implements the Service interface using SMTP
type SMTPService struct {
	config Config
}

// NewSMTPService creates a new SMTP email service
func NewSMTPService(config Config) *SMTPService {
	return &SMTPService{
		config: config,
	}
}

// SendVerificationEmail sends an email verification email to a user
func (s *SMTPService) SendVerificationEmail(user *storage.User, verificationToken string) error {
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
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <!-- Card -->
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <!-- Logo -->
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
        <!-- Main card -->
        <tr><td>
          <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#18181b;border:1px solid #27272a;border-radius:12px;overflow:hidden;">
            <!-- Header -->
            <tr><td style="padding:40px 40px 0;">
              <table role="presentation" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="width:56px;height:56px;background:rgba(99,102,241,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto;">
                      <tr><td style="text-align:center;">
                        <div style="font-size:24px;line-height:56px;">&#9993;</div>
                      </td></tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td></tr>
            <!-- Body -->
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Verify your email</h1>
              <p style="margin:0 0 8px;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Thanks for signing up for FunctionFly. Click the button below to verify your email address and activate your account.
              </p>
            </td></tr>
            <!-- CTA button -->
            <tr><td style="padding:28px 40px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:260px;" arcsize="17%%" stroke="f" fillcolor="#6366F1"><center style="color:#fff;font-size:15px;font-weight:600;">Verify email address</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#6366F1;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Verify email address</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <!-- Expiry note -->
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    This link expires in <strong style="color:#a1a1aa;">24 hours</strong>. If you didn't create an account, you can safely ignore this email.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <!-- Fallback link -->
            <tr><td style="padding:24px 40px 40px;">
              <p style="margin:0;font-size:12px;color:#52525b;line-height:1.5;">
                Having trouble with the button? Copy and paste this link into your browser:<br>
                <a href="%s" style="color:#6366F1;text-decoration:none;word-break:break-all;">%s</a>
              </p>
            </td></tr>
          </table>
        </td></tr>
        <!-- Footer -->
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, verificationURL, verificationURL, verificationURL, verificationURL, TransactionalEmailCopyrightHTML())

	// Plain text version
	textBody := fmt.Sprintf(`Welcome to FunctionFly!

Thank you for signing up. Please verify your email address to complete your registration.

Click this link to verify your email: %s

This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

--
FunctionFly Team
`, verificationURL)

	return s.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendPasswordResetEmail sends a password reset email to a user
func (s *SMTPService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
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

	return s.sendEmail(user.Email, subject, textBody, htmlBody)
}

// sendEmail sends an email using SMTP
func (s *SMTPService) sendEmail(to, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if s.config.SMTPUsername != "" || s.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		err = smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, s.config.FromEmail, []string{to}, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockService implements the Service interface using SMTP (for testing with real email sending)
type MockService struct {
	config Config
}

// NewMockService creates a new mock email service with real SMTP functionality
func NewMockService(config Config) *MockService {
	return &MockService{
		config: config,
	}
}

// SendBreachNotification sends a GDPR breach notification email
func (s *SMTPService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	// Extract breach details with defaults
	breachType := breachDetails["type"].(string)
	if breachType == "" {
		breachType = "Data Breach"
	}
	detectionTime := breachDetails["detectionTime"].(string)
	if detectionTime == "" {
		detectionTime = time.Now().Format(time.RFC3339)
	}
	affectedUsers := breachDetails["affectedUsers"].(int)
	riskLevel := breachDetails["riskLevel"].(string)
	if riskLevel == "" {
		riskLevel = "high"
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

	return s.sendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

// SendEmail sends a generic email with the given subject and body
func (s *SMTPService) SendEmail(to, subject, textBody, htmlBody string) error {
	return s.sendEmail(to, subject, textBody, htmlBody)
}

// SendWaitlistConfirmationEmail sends a waitlist confirmation email
func (s *SMTPService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
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
	return s.sendEmail(email, subject, textBody, htmlBody)
}

// ValidateConfiguration checks if the email service is properly configured
func (s *SMTPService) ValidateConfiguration() error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	if s.config.SMTPPort == 0 {
		return fmt.Errorf("SMTP port not configured")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	// Test SMTP connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construct SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Test basic TCP connection first
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SMTP server unreachable: %w", err)
	}
	conn.Close()

	// Test SMTP handshake
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer client.Close()

	// Get server greeting
	if err := client.Hello("functionfly-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	// Test authentication if credentials are provided
	if s.config.SMTPUsername != "" && s.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Send QUIT to gracefully close connection
	if err := client.Quit(); err != nil {
		// This is not a critical error, just log it
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}

// sendEmailToMultiple sends an email to multiple recipients
func (s *SMTPService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if s.config.SMTPUsername != "" || s.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		err = smtp.SendMail(addr, auth, s.config.FromEmail, to, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, s.config.FromEmail, to, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send breach notification email: %w", err)
	}

	return nil
}

// SendVerificationEmail sends an email verification email to a user (with real SMTP for testing)
func (m *MockService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "[TEST] Verify Your Email Address - FunctionFly"
	verifyBase := m.config.AuthURL
	if verifyBase == "" {
		verifyBase = m.config.BaseURL
	}
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", verifyBase, *user.VerificationToken)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verify your email — FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter','Segoe UI',system-ui,-apple-system,sans-serif;">
  <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <!-- Test banner -->
    <tr><td align="center">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%%%;">
        <tr><td style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
          ⚠️ TEST EMAIL — FunctionFly Development Environment
        </td></tr>
      </table>
    </td></tr>
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%%%;">
        <!-- Logo -->
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
        <!-- Main card -->
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#6366F1;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;">Verify email address</a>
              </td></tr></table>
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
          <p style="margin:8px 0 0;font-size:11px;color:#3f3f46;">This is a test email from the development environment.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, verificationURL, verificationURL, verificationURL, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

Verify your email

Thanks for signing up for FunctionFly. Click this link to verify your email and activate your account:

%s

This link expires in 24 hours. If you didn't create an account, ignore this email.
`, verificationURL)

	return m.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendPasswordResetEmail sends a password reset email (mock/dev version)
func (m *MockService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "[TEST] Reset Your Password - FunctionFly"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", m.config.AuthURL, resetToken)
	if m.config.AuthURL == "" {
		resetURL = fmt.Sprintf("%s/auth/reset-password?token=%s", m.config.BaseURL, resetToken)
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Reset your password — FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter','Segoe UI',system-ui,-apple-system,sans-serif;">
  <table role="presentation" width="100%%%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <!-- Test banner -->
    <tr><td align="center">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%%%;">
        <tr><td style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
          ⚠️ TEST EMAIL — FunctionFly Development Environment
        </td></tr>
      </table>
    </td></tr>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#6366F1;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;">Reset password</a>
              </td></tr></table>
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
          <p style="margin:8px 0 0;font-size:11px;color:#3f3f46;">This is a test email from the development environment.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, resetURL, resetURL, resetURL, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

Reset your password

We received a request to reset your FunctionFly password. Click this link to choose a new one:

%s

This link expires in 1 hour. If you didn't request this, ignore this email.
`, resetURL)

	return m.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendBreachNotification sends a GDPR breach notification email (with real SMTP for testing)
func (m *MockService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "[TEST] URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	// Extract breach details with defaults
	breachType := "Data Breach"
	if bt, ok := breachDetails["type"].(string); ok && bt != "" {
		breachType = bt
	}
	detectionTime := time.Now().Format(time.RFC3339)
	if dt, ok := breachDetails["detectionTime"].(string); ok && dt != "" {
		detectionTime = dt
	}
	affectedUsers := 0
	if au, ok := breachDetails["affectedUsers"].(int); ok {
		affectedUsers = au
	}
	riskLevel := "high"
	if rl, ok := breachDetails["riskLevel"].(string); ok && rl != "" {
		riskLevel = rl
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
        .test-notice { background-color: #f39c12; color: white; padding: 10px; text-align: center; font-weight: bold; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .urgent { color: #dc3545; font-weight: bold; font-size: 18px; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 20px; margin: 20px 0; border-left: 4px solid #dc3545; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="test-notice">
            ⚠️ TEST EMAIL - FunctionFly Development Environment
        </div>
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
            <p><em>This is a test email from the development environment.</em></p>
        </div>
    </div>
</body>
</html>
`, breachType, detectionTime, affectedUsers, riskLevel, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

DATA BREACH NOTIFICATION - GDPR Article 33 Compliance

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

* This is a test email from the development environment.
`, breachType, detectionTime, affectedUsers, riskLevel)

	return m.sendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

// SendEmail sends a generic email with the given subject and body
func (m *MockService) SendEmail(to, subject, textBody, htmlBody string) error {
	return m.sendEmail(to, subject, textBody, htmlBody)
}

// SendWaitlistConfirmationEmail sends a waitlist confirmation email (uses SMTP like real sends)
func (m *MockService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
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
	return m.sendEmail(email, subject, textBody, htmlBody)
}

// sendEmail sends an email using SMTP for the mock service
func (m *MockService) sendEmail(to, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", m.config.FromName, m.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if m.config.SMTPUsername != "" || m.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		err = smtp.SendMail(addr, auth, m.config.FromEmail, []string{to}, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, m.config.FromEmail, []string{to}, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendEmailToMultiple sends an email to multiple recipients for the mock service
func (m *MockService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", m.config.FromName, m.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if m.config.SMTPUsername != "" || m.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		err = smtp.SendMail(addr, auth, m.config.FromEmail, to, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, m.config.FromEmail, to, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send breach notification email: %w", err)
	}

	return nil
}

// ValidateConfiguration checks if the mock email service is properly configured
func (m *MockService) ValidateConfiguration() error {
	if m.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	if m.config.SMTPPort == 0 {
		return fmt.Errorf("SMTP port not configured")
	}
	if m.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	// Test SMTP connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construct SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Test basic TCP connection first
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SMTP server unreachable: %w", err)
	}
	conn.Close()

	// Test SMTP handshake
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer client.Close()

	// Get server greeting
	if err := client.Hello("functionfly-mock-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	// Test authentication if credentials are provided
	if m.config.SMTPUsername != "" && m.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Send QUIT to gracefully close connection
	if err := client.Quit(); err != nil {
		// This is not a critical error, just log it
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}
