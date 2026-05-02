package templates

import (
	"fmt"
	"time"
)

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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/wallet" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Add Funds</a>
              </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/security" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Review Account Security</a>
              </td></tr></table>
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
                  <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#dc2626;border-radius:8px;">
                    <a href="https://dashboard.functionfly.com/security/reset" target="_blank" style="display:inline-block;padding:12px 24px;font-size:14px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Secure Account</a>
                  </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0>
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
                  <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                    <a href="https://dashboard.functionfly.com/security" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Review Security</a>
                  </td></tr></table>
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