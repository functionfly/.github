package templates

import (
	"fmt"
	"time"
)

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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Update Payment</a>
              </td></tr></table>
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
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Invoice</a>
              </td></tr></table>
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