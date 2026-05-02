package templates

import (
	"fmt")

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