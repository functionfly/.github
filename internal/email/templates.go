package email

import (
	"fmt"
	"strings"
	"time"
)

// EmailTemplate holds the HTML and plain-text body for an email type
type EmailTemplate struct {
	HTML string
	Text string
}

// buildEmailURL constructs an email verification or auth URL
func buildEmailURL(base, path, token string) string {
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s%s=%s", base, separator, path, token)
}

// baseURL resolves the appropriate base URL for a given override
func baseURL(authURL, baseURL string) string {
	if authURL != "" {
		return authURL
	}
	return baseURL
}

// withDefault returns the first non-empty string or a fallback
func withDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// defaultRiskLevel returns a default risk level if none is provided
func defaultRiskLevel(level string) string {
	return withDefault(level, "high")
}

// extractString safely extracts a string from a map
func extractString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// extractInt safely extracts an int from a map
func extractInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

// VerificationEmailTemplate returns the email verification template set
func VerificationEmailTemplate(verifyURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
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
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <table role="presentation" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding-right:10px;vertical-align:middle;">
                <table role="presentation" width="32" height="32" cellpadding="0" cellspacing="0">
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">&#9993;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Verify your email</h1>
              <p style="margin:0 0 8px;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Thanks for signing up for FunctionFly. Click the button below to verify your email address and activate your account.
              </p>
            </td></tr>
            <tr><td style="padding:28px 40px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:260px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Verify email address</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Verify email address</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
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
                <a href="%s" style="color:#f97316;text-decoration:none;word-break:break-all;">%s</a>
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
</html>`, verifyURL, verifyURL, verifyURL, verifyURL, TransactionalEmailCopyrightHTML())

	text := fmt.Sprintf(`Welcome to FunctionFly!

Thank you for signing up. Please verify your email address to complete your registration.

Click this link to verify your email: %s

This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

--
FunctionFly Team
`, verifyURL)

	return EmailTemplate{HTML: html, Text: text}
}

// PasswordResetEmailTemplate returns the password reset template set
func PasswordResetEmailTemplate(resetURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Reset your password — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr><td style="width:56px;height:56px;background:rgba(239,68,68,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
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
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:220px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Reset password</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Reset password</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
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
                <a href="%s" style="color:#f97316;text-decoration:none;word-break:break-all;">%s</a>
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

	text := fmt.Sprintf(`Reset your password

We received a request to reset your FunctionFly password. Click this link to choose a new one:

%s

This link expires in 1 hour. If you didn't request this, ignore this email.
`, resetURL)

	return EmailTemplate{HTML: html, Text: text}
}

// WaitlistEmailTemplate returns the waitlist confirmation template set
func WaitlistEmailTemplate() EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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

	text := fmt.Sprintf(`You're on the list — FunctionFly

Thanks for requesting early access to FunctionFly. We've received your request and will review it shortly.

We'll send you an invite code as soon as we're ready for more users. Hang tight!

We're rolling out access gradually to ensure the best experience for everyone. You'll be among the first to know when your spot is ready.

--
%s`, TransactionalEmailCopyrightPlain())

	return EmailTemplate{HTML: html, Text: text}
}

// NewsletterSubscriptionTemplate returns the newsletter subscription confirmation template set
func NewsletterSubscriptionTemplate(name string) EmailTemplate {
	greeting := "Thanks for subscribing"
	if name != "" {
		greeting = "Hi " + name
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>You're subscribed — FunctionFly</title>
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">You're subscribed!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                %s! You've successfully subscribed to the FunctionFly newsletter.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Get ready for product updates, feature announcements, and more. We promise to only send you things worth reading.
              </p>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    You can unsubscribe at any time using the link in our emails.
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
</html>`, greeting, TransactionalEmailCopyrightHTML())

	text := fmt.Sprintf(`You're subscribed — FunctionFly

%s! You've successfully subscribed to the FunctionFly newsletter.

Get ready for product updates, feature announcements, and more. We promise to only send you things worth reading.

You can unsubscribe at any time using the link in our emails.

--
%s`, greeting, TransactionalEmailCopyrightPlain())

	return EmailTemplate{HTML: html, Text: text}
}

// BreachNotificationTemplate returns the GDPR breach notification template set
func BreachNotificationTemplate(breachType, detectionTime string, affectedUsers int, riskLevel string) EmailTemplate {
	if breachType == "" {
		breachType = "Data Breach"
	}
	if detectionTime == "" {
		detectionTime = "N/A"
	}
	if riskLevel == "" {
		riskLevel = "high"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
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
</html>`, breachType, detectionTime, affectedUsers, riskLevel, TransactionalEmailCopyrightHTML())

	text := fmt.Sprintf(`DATA BREACH NOTIFICATION - GDPR Article 33 Compliance

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

	return EmailTemplate{HTML: html, Text: text}
}

// MagicLinkEmailTemplate returns the magic link (existing user) template set
func MagicLinkEmailTemplate(magicURL string, expiryMinutes int) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Sign in to FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">✨</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Sign in to FunctionFly</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Click the button below to sign in to your FunctionFly account. This magic link is secure and expires in %d minutes.
              </p>
            </td></tr>
            <tr><td style="padding:28px 40px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:220px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Sign in</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Sign in</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    This link expires in <strong style="color:#a1a1aa;">%d minutes</strong> and can only be used once. If you didn't request this link, you can safely ignore this email.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 40px;">
              <p style="margin:0;font-size:12px;color:#52525b;line-height:1.5;">
                Having trouble with the button? Copy and paste this link into your browser:<br>
                <a href="%s" style="color:#f97316;text-decoration:none;word-break:break-all;">%s</a>
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
</html>`, expiryMinutes, magicURL, magicURL, expiryMinutes, magicURL, magicURL, TransactionalEmailCopyrightHTML())

	text := fmt.Sprintf(`Sign in to FunctionFly

Click this link to sign in to your account:

%s

This link expires in %d minutes and can only be used once.
If you didn't request this, you can safely ignore this email.
`, magicURL, expiryMinutes)

	return EmailTemplate{HTML: html, Text: text}
}

// MagicLinkSignupEmailTemplate returns the magic link (new user signup) template set
func MagicLinkSignupEmailTemplate(magicURL string, expiryMinutes int) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Welcome to FunctionFly</title>
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Welcome to FunctionFly!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Click the button below to create your FunctionFly account. No password needed!
              </p>
            </td></tr>
            <tr><td style="padding:28px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Create Account</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:16px 20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                    This link expires in <strong style="color:#a1a1aa;">%d minutes</strong> and can only be used once.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 40px;">
              <p style="margin:0;font-size:12px;color:#52525b;line-height:1.5;">
                Having trouble? Copy and paste this link:<br>
                <a href="%s" style="color:#f97316;text-decoration:none;word-break:break-all;">%s</a>
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
</html>`, magicURL, expiryMinutes, magicURL, magicURL, TransactionalEmailCopyrightHTML())

	text := fmt.Sprintf(`Welcome to FunctionFly!

Click this link to create your account:

%s

This link expires in %d minutes and can only be used once.
`, magicURL, expiryMinutes)

	return EmailTemplate{HTML: html, Text: text}
}

// TestBannerHTML prepends a test warning banner to HTML content
func TestBannerHTML(html string) string {
	banner := `<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
  <tr><td style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
    ⚠️ TEST EMAIL — FunctionFly Development Environment
  </td></tr>
</table>`
	return strings.Replace(html, `<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">`, banner, 1)
}

// TestBannerText prepends a test warning banner to plain text content
func TestBannerText(text string) string {
	return "[TEST EMAIL - FunctionFly Development Environment]\n\n" + text
}

// ═══════════════════════════════════════════════════════════════════════════════
// VELOCITY ORANGE BRAND COLOR: #f97316
// Brand consistency across all email templates
// ═══════════════════════════════════════════════════════════════════════════════

// Copyright helpers using Velocity Orange brand color
func TransactionalEmailCopyrightOrangeHTML() string {
	return fmt.Sprintf(`© %d FunctionFly. All rights reserved.<br>
<a href="https://functionfly.com/privacy" style="color:#f97316;text-decoration:none;">Privacy Policy</a> · 
<a href="https://functionfly.com/terms" style="color:#f97316;text-decoration:none;">Terms of Service</a>`, time.Now().Year())
}

func TransactionalEmailCopyrightOrangePlain() string {
	return fmt.Sprintf("© %d FunctionFly. All rights reserved.\nhttps://functionfly.com/privacy | https://functionfly.com/terms", time.Now().Year())
}

// AgentWalletLowBalanceTemplate alerts when agent wallet balance is below threshold
func AgentWalletLowBalanceTemplate(balance float64, threshold float64, walletID string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Wallet Low Balance Alert — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">⚠️</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Wallet Balance Low</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your agent wallet balance has dropped below the alert threshold.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 8px;font-size:13px;color:#71717a;">Current Balance</p>
                  <p style="margin:0;font-size:28px;font-weight:700;color:#f97316;">$%.2f</p>
                  <p style="margin:8px 0 0;font-size:13px;color:#71717a;">Alert threshold: $%.2f</p>
                  <p style="margin:8px 0 0;font-size:12px;color:#52525b;">Wallet ID: %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/wallet" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Add Funds</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/wallet" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Add Funds</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Agents may stop executing functions if the balance reaches zero. Add funds to avoid interruption.
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
</html>`, balance, threshold, walletID, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Wallet Low Balance Alert — FunctionFly

Your agent wallet balance ($%.2f) has dropped below the alert threshold ($%.2f).

Wallet ID: %s

Add funds to avoid agent execution interruptions:
https://dashboard.functionfly.com/wallet

--
%s`, balance, threshold, walletID, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// NewDeviceLoginTemplate notifies of login from unrecognized device/location
func NewDeviceLoginTemplate(deviceInfo, location, ipAddress string, loginTime time.Time) EmailTemplate {
	timeStr := loginTime.Format("Jan 2, 2006 at 3:04 PM MST")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>New Device Login — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">💻</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">New Device Login Detected</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We noticed a login to your FunctionFly account from a new device or location.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">When:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Device:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Location:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">IP Address:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/security" style="height:48px;v-text-anchor:middle;width:280px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Review Account Security</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/security" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Review Account Security</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                If this was you, no action is needed. If you don't recognize this activity, secure your account immediately.
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
</html>`, timeStr, deviceInfo, location, ipAddress, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`New Device Login Detected — FunctionFly

We noticed a login to your FunctionFly account from a new device or location.

Login Details:
- When: %s
- Device: %s
- Location: %s
- IP Address: %s

If this was you, no action is needed.
If you don't recognize this activity, secure your account:
https://dashboard.functionfly.com/security

--
%s`, timeStr, deviceInfo, location, ipAddress, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// PasswordChangedTemplate confirms password change with revert instructions
func PasswordChangedTemplate(changedAt time.Time, deviceInfo string) EmailTemplate {
	timeStr := changedAt.Format("Jan 2, 2006 at 3:04 PM MST")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Password Changed — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(34,197,94,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🔐</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Password Changed</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your FunctionFly account password was successfully changed.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Changed:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Device:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:rgba(249,115,22,0.05);border:1px solid rgba(249,115,22,0.2);border-radius:8px;">
                <tr><td style="padding:20px;">
                  <p style="margin:0 0 8px;font-size:14px;font-weight:600;color:#f97316;">Don't recognize this change?</p>
                  <p style="margin:0 0 16px;font-size:13px;color:#a1a1aa;line-height:1.5;">
                    If you didn't change your password, your account may be compromised. Secure it immediately.
                  </p>
                  <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/security/reset" style="height:40px;v-text-anchor:middle;width:180px;" arcsize="17%%" stroke="f" fillcolor="#dc2626"><center style="color:#fff;font-size:14px;font-weight:600;">Secure Account</center></v:roundrect><![endif]-->
                  <!--[if !mso]><!-->
                  <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#dc2626;border-radius:8px;">
                    <a href="https://dashboard.functionfly.com/security/reset" target="_blank" style="display:inline-block;padding:12px 24px;font-size:14px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Secure Account</a>
                  </td></tr></table>
                  <!--<![endif]-->
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
</html>`, timeStr, deviceInfo, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Password Changed — FunctionFly

Your FunctionFly account password was successfully changed.

Details:
- Changed: %s
- Device: %s

If you didn't change your password, secure your account immediately:
https://dashboard.functionfly.com/security/reset

--
%s`, timeStr, deviceInfo, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// SecurityAlertTemplate generic security event notification
func SecurityAlertTemplate(alertType, description, actionRequired string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Security Alert — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🚨</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Security Alert: %s</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                %s
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:rgba(249,115,22,0.05);border:1px solid rgba(249,115,22,0.2);border-radius:8px;">
                <tr><td style="padding:20px;">
                  <p style="margin:0 0 8px;font-size:14px;font-weight:600;color:#f97316;">Action Required</p>
                  <p style="margin:0 0 16px;font-size:13px;color:#a1a1aa;line-height:1.5;">
                    %s
                  </p>
                  <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/security" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Review Security</center></v:roundrect><![endif]-->
                  <!--[if !mso]><!-->
                  <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                    <a href="https://dashboard.functionfly.com/security" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Review Security</a>
                  </td></tr></table>
                  <!--<![endif]-->
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
</html>`, alertType, description, actionRequired, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Security Alert: %s — FunctionFly

%s

Action Required:
%s

Review your account security:
https://dashboard.functionfly.com/security

--
%s`, alertType, description, actionRequired, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// WaitlistInviteTemplate sends invite code to accepted waitlist users
func WaitlistInviteTemplate(inviteCode, signupURL string, expiresAt time.Time) EmailTemplate {
	expiresStr := expiresAt.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>You're In! — FunctionFly Early Access</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🎉</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">You're In!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your early access request was approved. Welcome to FunctionFly!
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:24px;text-align:center;">
                  <p style="margin:0 0 8px;font-size:13px;color:#71717a;">Your Invite Code</p>
                  <p style="margin:0;font-size:32px;font-weight:700;color:#f97316;letter-spacing:0.1em;">%s</p>
                  <p style="margin:12px 0 0;font-size:12px;color:#52525b;">Expires: %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:260px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Create Your Account</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Create Your Account</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Use your invite code to claim your spot. Codes expire after 14 days and can only be used once.
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
</html>`, inviteCode, expiresStr, signupURL, signupURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`You're In! — FunctionFly Early Access

Your early access request was approved. Welcome to FunctionFly!

Your Invite Code: %s
Expires: %s

Create your account:
%s

--
%s`, inviteCode, expiresStr, signupURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// FunctionDeploySuccessTemplate confirms successful function deployment
func FunctionDeploySuccessTemplate(functionName, version, runtime string, deployTime time.Time) EmailTemplate {
	timeStr := deployTime.Format("Jan 2, 2006 at 3:04 PM MST")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Deployment Successful — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(34,197,94,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🚀</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Deployment Successful</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your function has been deployed and is now live.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Function:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Version:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Runtime:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Deployed:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/functions/%s" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">View Function</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/functions/%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Function</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Your function is now serving requests. Monitor logs and metrics from the dashboard.
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
</html>`, functionName, version, runtime, timeStr, functionName, functionName, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Deployment Successful — FunctionFly

Your function has been deployed and is now live.

Function: %s
Version: %s
Runtime: %s
Deployed: %s

View function:
https://dashboard.functionfly.com/functions/%s

--
%s`, functionName, version, runtime, timeStr, functionName, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// FunctionDeployFailureTemplate alerts on failed deployment with logs link
func FunctionDeployFailureTemplate(functionName, errorMsg string, retryCount int) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Deployment Failed — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(220,38,38,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">❌</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Deployment Failed</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We couldn't deploy your function. Here's what went wrong:
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Function:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#ef4444;"><strong style="color:#a1a1aa;">Error:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Retry Attempt:</strong> %d/3</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/functions/%s/logs" style="height:48px;v-text-anchor:middle;width:180px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">View Logs</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/functions/%s/logs" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Logs</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Check the deployment logs for details. Fix the issue and redeploy when ready.
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
</html>`, functionName, errorMsg, retryCount, functionName, functionName, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Deployment Failed — FunctionFly

We couldn't deploy your function.

Function: %s
Error: %s
Retry Attempt: %d/3

View logs:
https://dashboard.functionfly.com/functions/%s/logs

--
%s`, functionName, errorMsg, retryCount, functionName, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// UsageExportReadyTemplate notifies when billing export is complete
func UsageExportReadyTemplate(exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) EmailTemplate {
	expiresStr := expiresAt.Format("Jan 2, 2006 at 3:04 PM MST")
	sizeMB := float64(sizeBytes) / 1024 / 1024

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Export Ready — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">📁</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Export Ready</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your usage data export is complete and ready for download.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Export ID:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Size:</strong> %.2f MB</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Expires:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Download Export</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Download Export</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Download links expire after 7 days. After expiration, you'll need to generate a new export.
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
</html>`, exportID, sizeMB, expiresStr, downloadURL, downloadURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Export Ready — FunctionFly

Your usage data export is complete.

Export ID: %s
Size: %.2f MB
Expires: %s

Download:
%s

--
%s`, exportID, sizeMB, expiresStr, downloadURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// PaymentFailedTemplate alerts on failed payment with retry info
func PaymentFailedTemplate(amount float64, dueDate time.Time, retryURL string) EmailTemplate {
	dueStr := dueDate.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Payment Failed — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(220,38,38,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">💳</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Payment Failed</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We couldn't process your payment. Please update your billing information.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Amount Due:</strong> $%.2f</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Due Date:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:220px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Update Payment</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Update Payment</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                We'll retry the payment in 24 hours. If it fails again, your service may be suspended.
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
</html>`, amount, dueStr, retryURL, retryURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Payment Failed — FunctionFly

We couldn't process your payment. Please update your billing information.

Amount Due: $%.2f
Due Date: %s

Update payment method:
%s

We'll retry in 24 hours. After that, service may be suspended.

--
%s`, amount, dueStr, retryURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// InvoiceReadyTemplate monthly invoice notification
func InvoiceReadyTemplate(period string, amount float64, invoiceURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Invoice Ready — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">📋</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Invoice Ready</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your invoice for %s is ready.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Period:</strong> %s</p>
                  <p style="margin:0;font-size:28px;font-weight:700;color:#f97316;">$%.2f</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">View Invoice</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Invoice</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Payment will be automatically charged to your default payment method on file.
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
</html>`, period, period, amount, invoiceURL, invoiceURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Invoice Ready — FunctionFly

Your invoice for %s is ready.

Period: %s
Amount: $%.2f

View invoice:
%s

Payment will be charged automatically.

--
%s`, period, period, amount, invoiceURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// TrustRevocationTemplate alerts when function trust status changes
func TrustRevocationTemplate(functionName, reason string, revokedAt time.Time) EmailTemplate {
	timeStr := revokedAt.Format("Jan 2, 2006 at 3:04 PM MST")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Trust Status Changed — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(220,38,38,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🔒</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Trust Status Revoked</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                A function's trust status has been revoked. It will no longer execute in trusted mode.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Function:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Revoked:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#ef4444;"><strong style="color:#a1a1aa;">Reason:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="https://dashboard.functionfly.com/functions/%s/trust" style="height:48px;v-text-anchor:middle;width:240px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Review Trust Status</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/functions/%s/trust" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Review Trust Status</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                The function will continue to run in untrusted mode. Re-apply for trust verification to restore privileges.
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
</html>`, functionName, timeStr, reason, functionName, functionName, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Trust Status Revoked — FunctionFly

A function's trust status has been revoked.

Function: %s
Revoked: %s
Reason: %s

Review trust status:
https://dashboard.functionfly.com/functions/%s/trust

--
%s`, functionName, timeStr, reason, functionName, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// DataRequestConfirmationTemplate confirms GDPR data request received
func DataRequestConfirmationTemplate(requestType, requestID string, estimatedCompletion time.Time) EmailTemplate {
	estStr := estimatedCompletion.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Data Request Received — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">📝</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Data Request Received</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We've received your %s request and are processing it.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Request ID:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Type:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Expected Completion:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                We'll notify you when your request is complete. You can check status anytime using your request ID.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>This request is handled per GDPR Article 12-22.</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, requestType, requestID, requestType, estStr, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Data Request Received — FunctionFly

We've received your %s request.

Request ID: %s
Type: %s
Expected Completion: %s

We'll notify you when complete. Check status anytime with your request ID.

--
%s
This request is handled per GDPR Article 12-22.`, requestType, requestID, requestType, estStr, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// AccountDeletionScheduledTemplate confirms account deletion with grace period
func AccountDeletionScheduledTemplate(deletionDate time.Time, cancelURL string) EmailTemplate {
	dateStr := deletionDate.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Account Deletion Scheduled — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(220,38,38,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🐛</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Account Deletion Scheduled</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your account and all associated data will be permanently deleted.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Deletion Date:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#ef4444;">This action cannot be undone.</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:240px;" arcsize="17%%" stroke="f" fillcolor="#52525b"><center style="color:#fff;font-size:15px;font-weight:600;">Cancel Deletion</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#52525b;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Cancel Deletion</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                You have 14 days to cancel. After %s, all data will be permanently erased.
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
</html>`, dateStr, cancelURL, cancelURL, dateStr, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Account Deletion Scheduled — FunctionFly

Your account will be permanently deleted on %s.

You have 14 days to cancel:
%s

After %s, all data will be permanently erased. This cannot be undone.

--
%s`, dateStr, cancelURL, dateStr, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// TeamInviteTemplate invites user to join organization
func TeamInviteTemplate(orgName, invitedBy, role string, acceptURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Team Invitation — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🤝</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">You're Invited!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                <strong style="color:#fafafa;">%s</strong> invited you to join <strong style="color:#fafafa;">%s</strong> on FunctionFly.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Organization:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Role:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Invited By:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Accept Invite</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Accept Invite</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                This invite expires in 7 days. If you weren't expecting this, you can safely ignore it.
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
</html>`, invitedBy, orgName, orgName, role, invitedBy, acceptURL, acceptURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Team Invitation — FunctionFly

%s invited you to join %s.

Organization: %s
Role: %s
Invited By: %s

Accept invitation:
%s

Expires in 7 days.

--
%s`, invitedBy, orgName, orgName, role, invitedBy, acceptURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// VaultSecretSharedTemplate notifies when secret is shared with user
func VaultSecretSharedTemplate(secretName, sharedBy, accessLevel string, viewURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Secret Shared — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🔐</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Secret Shared</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                A vault secret has been shared with you.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Secret:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Shared By:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Access Level:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:200px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">View Secret</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Secret</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Secrets are encrypted and can only be decrypted with your passphrase. FunctionFly never sees plaintext values.
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
</html>`, secretName, sharedBy, accessLevel, viewURL, viewURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Secret Shared — FunctionFly

A vault secret has been shared with you.

Secret: %s
Shared By: %s
Access Level: %s

View secret:
%s

--
%s`, secretName, sharedBy, accessLevel, viewURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// KeyRotationReminderTemplate warns of upcoming API key expiration
func KeyRotationReminderTemplate(keyName, keyID string, expiresAt time.Time, rotationURL string) EmailTemplate {
	daysUntil := int(time.Until(expiresAt).Hours() / 24)
	expiresStr := expiresAt.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>API Key Expiring — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🔑</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">API Key Expiring Soon</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                One of your API keys will expire in %d days. Rotate it now to avoid service interruption.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Key Name:</strong> %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Key ID:</strong> %s</p>
                  <p style="margin:0;font-size:13px;color:#f97316;"><strong style="color:#a1a1aa;">Expires:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <!--[if mso]><v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" href="%s" style="height:48px;v-text-anchor:middle;width:220px;" arcsize="17%%" stroke="f" fillcolor="#f97316"><center style="color:#fff;font-size:15px;font-weight:600;">Rotate Key</center></v:roundrect><![endif]-->
              <!--[if !mso]><!-->
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Rotate Key</a>
              </td></tr></table>
              <!--<![endif]-->
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                After rotation, update your applications with the new key. The old key will stop working on the expiration date.
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
</html>`, daysUntil, keyName, keyID, expiresStr, rotationURL, rotationURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`API Key Expiring — FunctionFly

Your API key expires in %d days.

Key Name: %s
Key ID: %s
Expires: %s

Rotate your key:
%s

--
%s`, daysUntil, keyName, keyID, expiresStr, rotationURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// MaintenanceNoticeTemplate advance notice of scheduled maintenance
func MaintenanceNoticeTemplate(windowStart, windowEnd time.Time, affectedServices []string) EmailTemplate {
	startStr := windowStart.Format("Jan 2, 3:04 PM MST")
	endStr := windowEnd.Format("3:04 PM MST")
	duration := windowEnd.Sub(windowStart).Minutes()

	servicesList := strings.Join(affectedServices, ", ")
	if len(affectedServices) == 0 {
		servicesList = "All services"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Scheduled Maintenance — FunctionFly</title>
  <!--[if mso]><style>table,td{font-family:Arial,sans-serif!important}</style><![endif]-->
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
                  <tr><td style="background:#7c2d12;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#f97316;transform:rotate(45deg);margin:0 auto;"></div>
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
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">🔧</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Scheduled Maintenance</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                We'll be performing maintenance to improve our services. Brief downtime may occur.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Date:</strong> %s - %s</p>
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Duration:</strong> %.0f minutes</p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Affected:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:rgba(249,115,22,0.05);border:1px solid rgba(249,115,22,0.2);border-radius:8px;">
                <tr><td style="padding:20px;">
                  <p style="margin:0;font-size:13px;color:#a1a1aa;line-height:1.5;">
                    <strong style="color:#f97316;">No action required.</strong> Your functions will automatically resume after maintenance.
                  </p>
                </td></tr>
              </table>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Status: <a href="https://status.functionfly.com" style="color:#f97316;text-decoration:none;">status.functionfly.com</a></div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, startStr, endStr, duration, servicesList, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Scheduled Maintenance — FunctionFly

We'll be performing maintenance to improve our services.

Date: %s - %s
Duration: %.0f minutes
Affected: %s

No action required. Functions will resume automatically.

Status: https://status.functionfly.com

--
%s`, startStr, endStr, duration, servicesList, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}
