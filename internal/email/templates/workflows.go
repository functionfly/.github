package templates

import (
	"fmt"
	"time"
)

// EmailWorkflow represents an automated email workflow configuration
type EmailWorkflow struct {
	ID          string
	Name        string
	Description string
	Trigger     string // "on_signup", "on_payment", "on_milestone", "on_inactivity", "manual"
	Category    string // "onboarding", "billing", "engagement", "retention", "security"
	DelayDays   int    // 0 = immediate, positive = delay in days
	Active      bool
}

// BundleEmailWorkflows returns the pre-configured email workflows for each bundle type
// These workflows are auto-provisioned when a bundle is deployed
func BundleEmailWorkflows(bundleSlug string) []EmailWorkflow {
	switch bundleSlug {
	case "saas-starter":
		return []EmailWorkflow{
			{
				ID:          "saas-starter-welcome",
				Name:        "Welcome Series",
				Description: "Onboarding sequence for new SaaS customers",
				Trigger:     "on_signup",
				Category:    "onboarding",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "saas-starter-payment-reminder",
				Name:        "Payment Reminder",
				Description: "Reminds users before payment is due",
				Trigger:     "on_payment",
				Category:    "billing",
				DelayDays:   -3, // 3 days before due
				Active:      true,
			},
			{
				ID:          "saas-starter-payment-failed",
				Name:        "Payment Failed Alert",
				Description: "Notifies users when payment fails",
				Trigger:     "on_payment",
				Category:    "billing",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "saas-starter-invoice-ready",
				Name:        "Invoice Ready",
				Description: "Sends invoice when billing cycle completes",
				Trigger:     "on_payment",
				Category:    "billing",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "saas-starter-reengagement",
				Name:        "Re-engagement Campaign",
				Description: "Brings back inactive users",
				Trigger:     "on_inactivity",
				Category:    "retention",
				DelayDays:   14, // 14 days of inactivity
				Active:      true,
			},
		}

	case "marketplace":
		return []EmailWorkflow{
			{
				ID:          "marketplace-welcome",
				Name:        "Marketplace Welcome",
				Description: "Welcome and seller onboarding",
				Trigger:     "on_signup",
				Category:    "onboarding",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "marketplace-first-sale",
				Name:        "First Sale Celebration",
				Description: "Celebrates when seller makes first sale",
				Trigger:     "on_milestone",
				Category:    "engagement",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "marketplace-payout-ready",
				Name:        "Payout Ready",
				Description: "Notifies when payout is available",
				Trigger:     "on_payment",
				Category:    "billing",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "marketplace-new-order",
				Name:        "New Order Alert",
				Description: "Notifies seller of new order",
				Trigger:     "on_milestone",
				Category:    "engagement",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "marketplace-review-request",
				Name:        "Review Request",
				Description: "Asks buyer to review purchase",
				Trigger:     "on_milestone",
				Category:    "retention",
				DelayDays:   3,
				Active:      true,
			},
		}

	case "ai-app":
		return []EmailWorkflow{
			{
				ID:          "ai-app-welcome",
				Name:        "AI App Welcome",
				Description: "Welcome and AI onboarding guide",
				Trigger:     "on_signup",
				Category:    "onboarding",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "ai-app-model-ready",
				Name:        "Model Ready",
				Description: "Notifies when AI model is configured",
				Trigger:     "on_milestone",
				Category:    "engagement",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "ai-app-usage-milestone",
				Name:        "Usage Milestone",
				Description: "Celebrates API usage milestones",
				Trigger:     "on_milestone",
				Category:    "engagement",
				DelayDays:   0,
				Active:      true,
			},
			{
				ID:          "ai-app-quota-warning",
				Name:        "Quota Warning",
				Description: "Warns before API quota is exhausted",
				Trigger:     "on_payment",
				Category:    "retention",
				DelayDays:   -2,
				Active:      true,
			},
			{
				ID:          "ai-app-security-alert",
				Name:        "Security Alert",
				Description: "Critical security notifications",
				Trigger:     "on_payment",
				Category:    "security",
				DelayDays:   0,
				Active:      true,
			},
		}
	}

	return nil
}

// SaaSStarterWelcomeEmailTemplate is the welcome email for SaaS Starter bundles
func SaaSStarterWelcomeEmailTemplate(bundleName, dashboardURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Welcome to %s — FunctionFly</title>
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
                    <div style="font-size:24px;line-height:56px;">&#127881;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Welcome to %s!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your <strong style="color:#f97316;">%s</strong> bundle is ready. You're now set up with everything you need to launch your SaaS product.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Your SaaS Starter includes:</strong></p>
                  <ul style="margin:0;padding-left:20px;font-size:13px;color:#a1a1aa;line-height:1.8;">
                    <li>Pre-configured app with Stripe billing</li>
                    <li>User authentication with optional MFA</li>
                    <li>Email workflow templates ready to customize</li>
                    <li>Analytics dashboard for tracking growth</li>
                    <li>Webhook handlers for payment events</li>
                  </ul>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Start Building</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Next steps: Configure your Stripe keys, customize your email templates, and add your first function. Check the docs for step-by-step guides.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Questions? Visit <a href="https://functionfly.com/docs" style="color:#f97316;text-decoration:none;">functionfly.com/docs</a></div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, bundleName, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Welcome to %s — FunctionFly

Your %s bundle is ready. You're now set up with everything you need to launch your SaaS product.

Your SaaS Starter includes:
- Pre-configured app with Stripe billing
- User authentication with optional MFA
- Email workflow templates ready to customize
- Analytics dashboard for tracking growth
- Webhook handlers for payment events

Start building: %s

Next steps: Configure your Stripe keys, customize your email templates, and add your first function.

--
%s
Questions? Visit functionfly.com/docs`, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// MarketplaceWelcomeEmailTemplate is the welcome email for Marketplace bundles
func MarketplaceWelcomeEmailTemplate(bundleName, dashboardURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Welcome to %s — FunctionFly Marketplace</title>
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
                  <td style="width:56px;height:56px;background:rgba(168,85,247,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">&#128722;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Welcome to %s!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your <strong style="color:#a855f7;">%s</strong> bundle is ready. Start selling your products or services today.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Your Marketplace includes:</strong></p>
                  <ul style="margin:0;padding-left:20px;font-size:13px;color:#a1a1aa;line-height:1.8;">
                    <li>Product listing management</li>
                    <li>Stripe Connect for marketplace payments</li>
                    <li>Order management and tracking</li>
                    <li>Automated payout workflows</li>
                    <li>Review and rating system</li>
                  </ul>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#a855f7;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Open Marketplace Dashboard</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Next steps: Create your first product listing, set up your Stripe Connect account, and customize your storefront.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Questions? Visit <a href="https://functionfly.com/docs/marketplace" style="color:#a855f7;text-decoration:none;">functionfly.com/docs/marketplace</a></div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, bundleName, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Welcome to %s — FunctionFly Marketplace

Your %s bundle is ready. Start selling your products or services today.

Your Marketplace includes:
- Product listing management
- Stripe Connect for marketplace payments
- Order management and tracking
- Automated payout workflows
- Review and rating system

Open Marketplace Dashboard: %s

Next steps: Create your first product listing, set up your Stripe Connect account, and customize your storefront.

--
%s
Questions? Visit functionfly.com/docs/marketplace`, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// AIAppWelcomeEmailTemplate is the welcome email for AI App bundles
func AIAppWelcomeEmailTemplate(bundleName, dashboardURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Welcome to %s — FunctionFly AI</title>
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
                  <tr><td style="background:#0c0c10;border-radius:6px;width:32px;height:32px;text-align:center;vertical-align:middle;">
                    <div style="width:14px;height:14px;background:#00D4FF;transform:rotate(45deg);margin:0 auto;"></div>
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
                  <td style="width:56px;height:56px;background:rgba(0,212,255,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">&#129504;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Welcome to %s!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                Your <strong style="color:#00D4FF;">%s</strong> bundle is ready. Build your AI-powered application with built-in RAG, embeddings, and vector storage.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Your AI App includes:</strong></p>
                  <ul style="margin:0;padding-left:20px;font-size:13px;color:#a1a1aa;line-height:1.8;">
                    <li>Chat completion API endpoints</li>
                    <li>Text embedding and vector storage</li>
                    <li>RAG-ready knowledge base setup</li>
                    <li>OpenRouter provider integration</li>
                    <li>Usage tracking and quota management</li>
                  </ul>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#00D4FF;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#09090b;text-decoration:none;font-family:inherit;">Build Your AI App</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Next steps: Configure your AI provider (OpenRouter or custom), add documents to your knowledge base, and test your first chat completion.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Questions? Visit <a href="https://functionfly.com/docs/ai-apps" style="color:#00D4FF;text-decoration:none;">functionfly.com/docs/ai-apps</a></div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, bundleName, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Welcome to %s — FunctionFly AI

Your %s bundle is ready. Build your AI-powered application with built-in RAG, embeddings, and vector storage.

Your AI App includes:
- Chat completion API endpoints
- Text embedding and vector storage
- RAG-ready knowledge base setup
- OpenRouter provider integration
- Usage tracking and quota management

Build Your AI App: %s

Next steps: Configure your AI provider (OpenRouter or custom), add documents to your knowledge base, and test your first chat completion.

--
%s
Questions? Visit functionfly.com/docs/ai-apps`, bundleName, bundleName, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// PaymentReminderEmailTemplate sends a reminder before payment is due
func PaymentReminderEmailTemplate(amount float64, dueDate time.Time, dashboardURL string) EmailTemplate {
	dueStr := dueDate.Format("Jan 2, 2006")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Payment Reminder — FunctionFly</title>
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
                  <td style="width:56px;height:56px;background:rgba(245,158,11,0.1);border-radius:50%%;text-align:center;vertical-align:middle;">
                    <div style="font-size:24px;line-height:56px;">&#128680;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">Payment Due Soon</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                This is a friendly reminder that your payment is due soon. Please ensure your payment method is up to date.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Amount Due:</strong> <span style="color:#f97316;font-weight:700;">$%.2f</span></p>
                  <p style="margin:0;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">Due Date:</strong> %s</p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Update Payment Method</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                If your payment method is already updated, you can safely ignore this email. Questions? Reply to this email or visit our support page.
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
</html>`, amount, dueStr, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`Payment Reminder — FunctionFly

This is a friendly reminder that your payment is due soon. Please ensure your payment method is up to date.

Amount Due: $%.2f
Due Date: %s

Update payment method: %s

If your payment method is already updated, you can safely ignore this email.

--
%s`, amount, dueStr, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// ReengagementEmailTemplate brings back inactive users
func ReengagementEmailTemplate(daysInactive int, dashboardURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>We miss you — FunctionFly</title>
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
                    <div style="font-size:24px;line-height:56px;">&#128075;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">We miss you!</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                It's been %d days since your last visit. We've been busy adding new features and would love to have you back.
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0 0 12px;font-size:13px;color:#71717a;"><strong style="color:#a1a1aa;">What's new:</strong></p>
                  <ul style="margin:0;padding-left:20px;font-size:13px;color:#a1a1aa;line-height:1.8;">
                    <li>New email workflow templates</li>
                    <li>Improved deployment experience</li>
                    <li>Enhanced analytics dashboard</li>
                  </ul>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#f97316;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">Come Back</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Still too busy? No worries — we understand. Reply to let us know what would help you get back on track.
              </p>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding:24px 16px;text-align:center;">
          <div style="margin:0;font-size:12px;color:#52525b;line-height:1.6;">%s<br>Questions? Reply to this email — we read every reply.</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, daysInactive, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`We miss you! — FunctionFly

It's been %d days since your last visit. We've been busy adding new features and would love to have you back.

What's new:
- New email workflow templates
- Improved deployment experience
- Enhanced analytics dashboard

Come back: %s

Still too busy? No worries — we understand. Reply to let us know what would help you get back on track.

--
%s
Questions? Reply to this email — we read every reply.`, daysInactive, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}

// MilestoneEmailTemplate celebrates user milestones
func MilestoneEmailTemplate(milestoneName, milestoneDescription, dashboardURL string) EmailTemplate {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>%s — FunctionFly</title>
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
                    <div style="font-size:24px;line-height:56px;">&#127942;</div>
                  </td>
                </tr>
              </table>
            </td></tr>
            <tr><td style="padding:24px 40px 0;">
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#fafafa;letter-spacing:-0.02em;">%s</h1>
              <p style="margin:0;font-size:15px;color:#a1a1aa;line-height:1.6;">
                %s
              </p>
            </td></tr>
            <tr><td style="padding:24px 40px;">
              <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                <tr><td style="background:#111113;border:1px solid #27272a;border-radius:8px;padding:20px;">
                  <p style="margin:0;font-size:13px;color:#71717a;line-height:1.6;">
                    Keep up the great work! Your progress is inspiring.
                  </p>
                </td></tr>
              </table>
            </td></tr>
            <tr><td style="padding:0 40px 28px;">
              <table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="background:#22c55e;border-radius:8px;">
                <a href="%s" target="_blank" style="display:inline-block;padding:14px 32px;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;font-family:inherit;">View Your Progress</a>
              </td></tr></table>
            </td></tr>
            <tr><td style="padding:0 40px 40px;">
              <p style="margin:0;font-size:13px;color:#71717a;line-height:1.5;">
                Thank you for building with FunctionFly!
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
</html>`, milestoneName, milestoneName, milestoneDescription, dashboardURL, TransactionalEmailCopyrightOrangeHTML())

	text := fmt.Sprintf(`%s — FunctionFly

%s

Keep up the great work! Your progress is inspiring.

View Your Progress: %s

Thank you for building with FunctionFly!

--
%s`, milestoneName, milestoneDescription, dashboardURL, TransactionalEmailCopyrightOrangePlain())

	return EmailTemplate{HTML: html, Text: text}
}
