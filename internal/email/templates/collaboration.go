package templates

import (
	"fmt"
)

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
              <table role="presentation" cellpadding="0" cellspacing="0>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Accept Invite</a>
              </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Secret</a>
              </td></tr></table>
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

func InviteEmailTemplate(inviterName, orgName, role, acceptURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>You've Been Invited — FunctionFly</title>
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
              <table role="presentation" cellpadding="0" cellspacing="0>
                <tr>
                  <td style="width:56px;height:56px;background:rgba(249,115,22,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">📬</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">You're Invited to Join %s</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                <strong style="color:#fafafa;">%s</strong> has invited you to join <strong style="color:#fafafa;">%s</strong> as a <strong style="color:#f97316;">%s</strong>.
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Accept Invitation</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                This invitation expires in 7 days. If you weren't expecting this, you can safely ignore it.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Questions? Visit <a href="https://functionfly.com/support" style="color:#f97316;text-decoration:none;">functionfly.com/support</a></div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, orgName, inviterName, orgName, role, orgName, role, inviterName, acceptURL, acceptURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`You're Invited to Join %s — FunctionFly

%s has invited you to join %s as a %s.

Organization: %s
Role: %s
Invited By: %s

Accept your invitation: %s

This invitation expires in 7 days. If you weren't expecting this, you can safely ignore it.

--
%s
Questions? Visit functionfly.com/support`, orgName, inviterName, orgName, role, orgName, role, inviterName, acceptURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}