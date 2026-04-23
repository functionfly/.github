#!/usr/bin/env python3
"""
Script to add magic link email methods to the email service.
This adds:
1. SendMagicLinkEmail and SendMagicLinkSignupEmail to the Service interface
2. SendMagicLinkEmail and SendMagicLinkSignupEmail to SMTPService
3. SendMagicLinkEmail and SendMagicLinkSignupEmail to MockService
"""

import re

# Read the file
with open('internal/email/email.go', 'r') as f:
    content = f.read()

# 1. Update Service interface to add magic link methods
old_interface_end = '\tSendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error\n\tSendEmail(to, subject, textBody, htmlBody string) error\n\tValidateConfiguration() error\n}'
new_interface_end = '''\tSendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error
\tSendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error
\tSendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error
\tSendEmail(to, subject, textBody, htmlBody string) error
\tValidateConfiguration() error
}'''
content = content.replace(old_interface_end, new_interface_end)

# 2. Add SMTPService magic link methods before MockService definition
smtp_magic_link_methods = '''
// SendMagicLinkEmail sends a magic link email to an existing user
func (s *SMTPService) SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Sign in to FunctionFly — Magic Link"
	authBase := s.config.AuthURL
	if authBase == "" {
		authBase = s.config.BaseURL
	}
	magicLinkURL := fmt.Sprintf("%s/auth/magic-link/verify?token=%s", authBase, token)
	expiryMinutes := int(expiry.Minutes())

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Sign in to FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter',system-ui,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <span style="font-size:18px;font-weight:700;color:#fafafa;">FunctionFly</span>
        </td></tr>
        <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:40px;">
          <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;">Sign in to FunctionFly</h1>
          <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
            Click the button below to sign in. This magic link expires in %%d minutes.
          </p>
          <div style="padding:28px 0;">
            <a href="%%s" style="display:inline-block;padding:14px 32px;background:#6366F1;border-radius:8px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">Sign in</a>
          </div>
          <p style="margin:0;font-size:13px;color:#71717a;">
            This link expires in <strong style="color:#a1a1aa;">%%d minutes</strong> and can only be used once.
          </p>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="font-size:12px;color:#52525b;">%%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, expiryMinutes, magicLinkURL, expiryMinutes, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`Sign in to FunctionFly

Click this link to sign in:
%%s

This link expires in %%d minutes and can only be used once.
`, magicLinkURL, expiryMinutes)

	return s.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendMagicLinkSignupEmail sends a magic link email for new user signup
func (s *SMTPService) SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error {
	subject := "Welcome to FunctionFly — Complete Your Sign Up"
	authBase := s.config.AuthURL
	if authBase == "" {
		authBase = s.config.BaseURL
	}
	magicLinkURL := fmt.Sprintf("%%s/auth/magic-link/verify?token=%%s", authBase, token)
	expiryMinutes := int(expiry.Minutes())

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Welcome to FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter',system-ui,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="padding-bottom:32px;">
          <span style="font-size:18px;font-weight:700;color:#fafafa;">FunctionFly</span>
        </td></tr>
        <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:40px;">
          <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;">Welcome to FunctionFly!</h1>
          <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
            Click the button below to create your account. No password needed!
          </p>
          <div style="padding:28px 0;">
            <a href="%%s" style="display:inline-block;padding:14px 32px;background:#6366F1;border-radius:8px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">Create Account</a>
          </div>
          <p style="margin:0;font-size:13px;color:#71717a;">
            This link expires in <strong style="color:#a1a1aa;">%%d minutes</strong>.
          </p>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="font-size:12px;color:#52525b;">%%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, magicLinkURL, expiryMinutes, magicLinkURL, TransactionalEmailCopyrightHTML())

	textBody := fmt.Sprintf(`Welcome to FunctionFly!

Click this link to create your account:
%%s

This link expires in %%d minutes.
`, magicLinkURL, expiryMinutes)

	return s.sendEmail(email, subject, textBody, htmlBody)
}

'''

mockservice_marker = '// MockService implements the Service interface using SMTP (for testing with real email sending)'
content = content.replace(mockservice_marker, smtp_magic_link_methods + mockservice_marker)

# 3. Add MockService magic link methods - find the MockService SendEmail method and add before it
mock_magic_link_methods = '''// SendMagicLinkEmail sends a magic link email (mock implementation)
func (m *MockService) SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "[TEST] Sign in to FunctionFly — Magic Link"
	authBase := m.config.AuthURL
	if authBase == "" {
		authBase = m.config.BaseURL
	}
	magicLinkURL := fmt.Sprintf("%%s/auth/magic-link/verify?token=%%s", authBase, token)
	expiryMinutes := int(expiry.Minutes())

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>[TEST] Sign in to FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter',system-ui,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
          ⚠️ TEST EMAIL — FunctionFly Development Environment
        </td></tr>
        <tr><td align="center" style="padding-bottom:32px;padding-top:24px;">
          <span style="font-size:18px;font-weight:700;color:#fafafa;">FunctionFly</span>
        </td></tr>
        <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:40px;">
          <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;">Sign in to FunctionFly</h1>
          <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
            Click the button below to sign in. This magic link expires in %%d minutes.
          </p>
          <div style="padding:28px 0;">
            <a href="%%s" style="display:inline-block;padding:14px 32px;background:#6366F1;border-radius:8px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">Sign in</a>
          </div>
          <p style="margin:0;font-size:13px;color:#71717a;">
            This link expires in <strong style="color:#a1a1aa;">%%d minutes</strong> and can only be used once.
          </p>
          <p style="margin:24px 0 0;font-size:11px;color:#3f3f46;">This is a test email from the development environment.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, expiryMinutes, magicLinkURL, expiryMinutes)

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

Sign in to FunctionFly

Click this link to sign in:
%%s

This link expires in %%d minutes.
`, magicLinkURL, expiryMinutes)

	return m.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendMagicLinkSignupEmail sends a magic link email for new user signup (mock implementation)
func (m *MockService) SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error {
	subject := "[TEST] Welcome to FunctionFly — Complete Your Sign Up"
	authBase := m.config.AuthURL
	if authBase == "" {
		authBase = m.config.BaseURL
	}
	magicLinkURL := fmt.Sprintf("%%s/auth/magic-link/verify?token=%%s", authBase, token)
	expiryMinutes := int(expiry.Minutes())

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>[TEST] Welcome to FunctionFly</title>
</head>
<body style="margin:0;padding:0;background-color:#0a0a0b;font-family:'Inter',system-ui,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0a0a0b;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
        <tr><td align="center" style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
          ⚠️ TEST EMAIL — FunctionFly Development Environment
        </td></tr>
        <tr><td align="center" style="padding-bottom:32px;padding-top:24px;">
          <span style="font-size:18px;font-weight:700;color:#fafafa;">FunctionFly</span>
        </td></tr>
        <tr><td style="background:#18181b;border:1px solid #27272a;border-radius:12px;padding:40px;">
          <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;">Welcome to FunctionFly!</h1>
          <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
            Click the button below to create your account. No password needed!
          </p>
          <div style="padding:28px 0;">
            <a href="%%s" style="display:inline-block;padding:14px 32px;background:#6366F1;border-radius:8px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">Create Account</a>
          </div>
          <p style="margin:0;font-size:13px;color:#71717a;">
            This link expires in <strong style="color:#a1a1aa;">%%d minutes</strong>.
          </p>
          <p style="margin:24px 0 0;font-size:11px;color:#3f3f46;">This is a test email from the development environment.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, magicLinkURL, expiryMinutes)

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

Welcome to FunctionFly!

Click this link to create your account:
%%s

This link expires in %%d minutes.
`, magicLinkURL, expiryMinutes)

	return m.sendEmail(email, subject, textBody, htmlBody)
}

'''

# Find the MockService SendEmail method and add before it
mock_sendemail_marker = '// SendEmail sends a generic email with the given subject and body\nfunc (m *MockService) SendEmail'
content = content.replace(mock_sendemail_marker, mock_magic_link_methods + mock_sendemail_marker)

# Write the updated content back
with open('internal/email/email.go', 'w') as f:
    f.write(content)

print("Successfully updated email.go with magic link methods!")
EOF
python3 add_magic_link_email.py
