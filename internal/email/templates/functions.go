package templates

import (
	"fmt"
	"time"
)

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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/functions/%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Function</a>
              </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="https://dashboard.functionfly.com/functions/%s/logs" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Logs</a>
              </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Download Export</a>
              </td></tr></table>
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