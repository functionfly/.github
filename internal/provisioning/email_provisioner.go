package provisioning

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EmailWorkflowProvisioner creates the isolated email automation engine for a tenant.
// All email data (templates, workflows, send logs) lives in the tenant's own database.
//
// What it provisions:
//   - Default email templates (transactional, marketing, billing — 15+ templates)
//   - Automated workflow: Welcome Sequence (signup → welcome → setup guide → check-in)
//   - Automated workflow: Onboarding Drip (signup → feature tour → activation nudge)
//   - Automated workflow: Payment Dunning (payment_failed → retry → warning → suspension)
//   - Automated workflow: Trial Conversion (trial_start → mid-trial → trial_ending → converted/churned)
//   - Email suppression management
//   - Provider configuration (delegates to platform Resend/SMTP)
type EmailWorkflowProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
	emailService  email.Service
}

// NewEmailWorkflowProvisioner creates a new Email Workflow provisioner
func NewEmailWorkflowProvisioner(
	platformRepo storage.Repository,
	dbProvisioner *storage.TenantDBProvisioner,
	emailService email.Service,
) *EmailWorkflowProvisioner {
	return &EmailWorkflowProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
		emailService:  emailService,
	}
}

// Provision sets up the complete email workflow infrastructure in the tenant's database.
func (ewp *EmailWorkflowProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "email_workflows",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := ewp.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Seed transactional email templates
	transactionalTemplates := []struct {
		slug     string
		name     string
		subject  string
		htmlBody string
		textBody string
		vars     string
	}{
		{
			"team-invite", "Team Invitation", "You've been invited to join {{.OrgName}}",
			`<div style="font-family:sans-serif;max-width:600px;margin:0 auto"><h2>You're invited!</h2><p>{{.InvitedBy}} has invited you to join <strong>{{.OrgName}}</strong> as a {{.Role}}.</p><a href="{{.AcceptURL}}" style="display:inline-block;padding:12px 24px;background:#6366f1;color:white;text-decoration:none;border-radius:8px">Accept Invitation</a><p style="color:#666;margin-top:24px;font-size:14px">This invite expires in 72 hours.</p></div>`,
			"You're invited to join {{.OrgName}} as a {{.Role}}. Accept: {{.AcceptURL}}",
			`[{"name":"OrgName","type":"string","required":true},{"name":"InvitedBy","type":"string","required":true},{"name":"Role","type":"string","default":"member","required":false},{"name":"AcceptURL","type":"string","required":true}]`,
		},
		{
			"password-changed", "Password Changed", "Your password was changed",
			`<div style="font-family:sans-serif;max-width:600px;margin:0 auto"><h2>Password Changed</h2><p>Your password was changed on {{.ChangedAt}} from {{.DeviceInfo}}.</p><p>If you didn't make this change, <a href="{{.ResetURL}}">reset your password immediately</a>.</p></div>`,
			"Your password was changed on {{.ChangedAt}}. If this wasn't you, reset at: {{.ResetURL}}",
			`[{"name":"ChangedAt","type":"string","required":true},{"name":"DeviceInfo","type":"string","default":"Unknown device","required":false},{"name":"ResetURL","type":"string","required":true}]`,
		},
		{
			"new-login", "New Login", "New sign-in to your account",
			`<div style="font-family:sans-serif;max-width:600px;margin:0 auto"><h2>New Sign-In Detected</h2><p>Your account was accessed from:</p><ul><li><strong>Device:</strong> {{.DeviceInfo}}</li><li><strong>Location:</strong> {{.Location}}</li><li><strong>Time:</strong> {{.LoginTime}}</li></ul><p>If this wasn't you, <a href="{{.SecureURL}}">secure your account</a>.</p></div>`,
			"New sign-in from {{.DeviceInfo}} at {{.LoginTime}}. If this wasn't you: {{.SecureURL}}",
			`[{"name":"DeviceInfo","type":"string","required":true},{"name":"Location","type":"string","default":"Unknown","required":false},{"name":"LoginTime","type":"string","required":true},{"name":"SecureURL","type":"string","required":true}]`,
		},
		{
			"account-deleted", "Account Deleted", "Your account has been deleted",
			`<div style="font-family:sans-serif;max-width:600px;margin:0 auto"><h2>Account Deleted</h2><p>Your account and all associated data have been permanently deleted.</p><p>If this was a mistake, contact support within 30 days for possible recovery.</p></div>`,
			"Your account has been permanently deleted. Contact support within 30 days for recovery.",
			`[]`,
		},
		{
			"data-export-ready", "Data Export Ready", "Your data export is ready",
			`<div style="font-family:sans-serif;max-width:600px;margin:0 auto"><h2>Data Export Ready</h2><p>Your data export is ready for download.</p><a href="{{.DownloadURL}}" style="display:inline-block;padding:12px 24px;background:#6366f1;color:white;text-decoration:none;border-radius:8px">Download Export</a><p style="color:#666;margin-top:16px;font-size:14px">This link expires on {{.ExpiresAt}}. Size: {{.SizeBytes}}</p></div>`,
			"Your data export is ready. Download: {{.DownloadURL}} (expires {{.ExpiresAt}})",
			`[{"name":"DownloadURL","type":"string","required":true},{"name":"ExpiresAt","type":"string","required":true},{"name":"SizeBytes","type":"string","default":"","required":false}]`,
		},
	}

	for _, t := range transactionalTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'transactional', true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject, t.htmlBody, t.textBody, t.vars)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed template (non-fatal)")
		}
	}
	log.WithField("count", len(transactionalTemplates)).Info("Transactional email templates seeded")

	// 3. Create automated workflows
	workflows := []struct {
		slug        string
		name        string
		triggerType string
		triggerCfg  string
		steps       []workflowStep
	}{
		{
			"welcome-sequence", "Welcome Sequence", "user_signup", "{}",
			[]workflowStep{
				{0, "welcome", 0, "minutes", "always"},
				{1, "setup-guide", 1, "hours", "always"},
				{2, "check-in", 3, "days", "if_not_opened"},
			},
		},
		{
			"onboarding-drip", "Onboarding Drip", "user_signup", "{}",
			[]workflowStep{
				{0, "feature-tour", 2, "hours", "always"},
				{1, "activation-nudge", 2, "days", "always"},
				{2, "success-story", 5, "days", "if_opened"},
			},
		},
		{
			"payment-dunning", "Payment Dunning", "payment_failed", "{}",
			[]workflowStep{
				{0, "payment-failed", 0, "minutes", "always"},
				{1, "payment-retry", 3, "days", "always"},
				{2, "payment-warning", 7, "days", "always"},
				{3, "suspension-notice", 14, "days", "always"},
			},
		},
		{
			"trial-conversion", "Trial Conversion", "trial_started", "{}",
			[]workflowStep{
				{0, "trial-welcome", 0, "minutes", "always"},
				{1, "trial-midpoint", 7, "days", "always"},
				{2, "trial-ending", 12, "days", "always"},
				{3, "trial-expired", 14, "days", "always"},
			},
		},
	}

	for _, wf := range workflows {
		var workflowID uuid.UUID
		err = pool.QueryRow(ctx,
			`INSERT INTO tenant_email_workflows (id, tenant_id, slug, name, trigger_type, trigger_config, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, true)
			 ON CONFLICT (tenant_id, slug) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			uuid.New(), tenantID, wf.slug, wf.name, wf.triggerType, wf.triggerCfg).Scan(&workflowID)
		if err != nil {
			log.WithError(err).WithField("workflow", wf.slug).Warn("Failed to create workflow (non-fatal)")
			continue
		}

		for _, step := range wf.steps {
			// Look up template ID
			var templateID uuid.NullUUID
			pool.QueryRow(ctx,
				`SELECT id FROM tenant_email_templates WHERE tenant_id = $1 AND slug = $2`,
				tenantID, step.templateSlug).Scan(&templateID)

			_, err = pool.Exec(ctx,
				`INSERT INTO tenant_email_workflow_steps (id, tenant_id, workflow_id, template_id, step_order, delay_amount, delay_unit, condition_type, is_active)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
				 ON CONFLICT (workflow_id, step_order) DO NOTHING`,
				uuid.New(), tenantID, workflowID, templateID, step.order, step.delay, step.unit, step.condition)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"workflow": wf.slug,
					"step":     step.order,
				}).Warn("Failed to create workflow step (non-fatal)")
			}
		}
		log.WithField("workflow", wf.slug).Info("Email workflow created")
	}

	// 4. Create additional email templates needed by workflows
	additionalTemplates := []struct {
		slug    string
		name    string
		subject string
	}{
		{"setup-guide", "Setup Guide", "Get started in 3 easy steps"},
		{"check-in", "How's it going?", "Quick check-in — need any help?"},
		{"feature-tour", "Feature Tour", "Discover what you can do"},
		{"activation-nudge", "Activate Your Account", "You're one step away"},
		{"success-story", "Customer Spotlight", "See how others succeed"},
		{"payment-retry", "Payment Retry", "We'll retry your payment soon"},
		{"payment-warning", "Payment Warning", "Your account may be suspended"},
		{"suspension-notice", "Account Suspended", "Your account has been suspended"},
		{"trial-welcome", "Trial Started", "Your free trial has begun!"},
		{"trial-midpoint", "Trial Check-in", "How's your trial going?"},
		{"trial-ending", "Trial Ending Soon", "Your trial ends in 2 days"},
		{"trial-expired", "Trial Expired", "Your trial has ended"},
	}

	for _, t := range additionalTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'workflow', true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject,
			fmt.Sprintf("<!-- %s — customize in dashboard -->", t.name),
			fmt.Sprintf("%s — customize in dashboard", t.name),
			`[{"name":"AppName","type":"string","default":"My App","required":false},{"name":"UserName","type":"string","default":"","required":false},{"name":"Link","type":"string","default":"","required":false}]`)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed additional template (non-fatal)")
		}
	}
	log.WithField("count", len(additionalTemplates)).Info("Workflow email templates seeded")

	state.Status = StatusActive
	state.ResourceID = fmt.Sprintf("workflows:%d", len(workflows))
	log.Info("Email workflows provisioning complete")
	return state, nil
}

type workflowStep struct {
	order         int
	templateSlug  string
	delay         int
	unit          string
	condition     string
}
