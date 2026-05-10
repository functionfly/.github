package provisioning

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthProvisioner creates the isolated authentication infrastructure for a tenant.
// All auth data lives in the tenant's own dedicated database — never in the platform DB.
//
// What it provisions:
//   - JWT signing key (HS256, 32-byte, rotatable via tenant_auth_keys)
//   - Default session policy (7-day max, 30-min idle, 5 concurrent)
//   - Pre-configured OAuth providers (Google, GitHub — disabled by default, ready to enable)
//   - Default auth audit event marking provisioning complete
type AuthProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewAuthProvisioner creates a new Auth provisioner
func NewAuthProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *AuthProvisioner {
	return &AuthProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision sets up the complete isolated auth infrastructure in the tenant's database.
func (ap *AuthProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "auth",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := ap.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Generate JWT signing key
	keyID := fmt.Sprintf("key_%s", uuid.New().String()[:8])
	keyMaterial, err := generateSigningKey(32)
	if err != nil {
		return state, fmt.Errorf("failed to generate JWT signing key: %w", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_auth_keys (id, tenant_id, key_id, algorithm, key_material, is_active, activated_at)
		 VALUES ($1, $2, $3, 'HS256', $4, true, NOW())
		 ON CONFLICT (tenant_id, key_id) DO NOTHING`,
		uuid.New(), tenantID, keyID, keyMaterial)
	if err != nil {
		return state, fmt.Errorf("failed to store JWT signing key: %w", err)
	}
	log.WithField("key_id", keyID).Info("JWT signing key generated")

	// 3. Seed OAuth provider configurations (disabled, ready for customer to enable)
	oauthProviders := []struct {
		provider string
		scopes   string
	}{
		{"google", `["openid","email","profile"]`},
		{"github", `["user:email","read:user"]`},
	}

	for _, p := range oauthProviders {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_oauth_configs (id, tenant_id, provider, client_id, encrypted_client_secret, callback_url, scopes, enabled)
			 VALUES ($1, $2, $3, '', '', '', $4, false)
			 ON CONFLICT (tenant_id, provider) DO NOTHING`,
			uuid.New(), tenantID, p.provider, p.scopes)
		if err != nil {
			log.WithError(err).WithField("provider", p.provider).Warn("Failed to seed OAuth provider (non-fatal)")
		}
	}
	log.Info("OAuth providers seeded (Google, GitHub — disabled)")

	// 4. Seed default email templates for auth flows
	authTemplates := []struct {
		slug     string
		name     string
		subject  string
		category string
	}{
		{"verify-email", "Email Verification", "Verify your email address", "auth"},
		{"reset-password", "Password Reset", "Reset your password", "auth"},
		{"magic-link", "Magic Link", "Your sign-in link", "auth"},
		{"welcome", "Welcome", "Welcome to {{.AppName}}!", "auth"},
		{"mfa-setup", "MFA Setup", "Set up two-factor authentication", "auth"},
	}

	for _, t := range authTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject,
			fmt.Sprintf("<!-- %s template — customize in dashboard -->", t.name),
			fmt.Sprintf("%s template - customize in dashboard", t.name),
			`[{"name":"AppName","type":"string","default":"My App","required":false},{"name":"Link","type":"string","default":"","required":true}]`,
			t.category)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed auth template (non-fatal)")
		}
	}
	log.Info("Auth email templates seeded")

	// 5. Log provisioning event in tenant's isolated audit log
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_auth_audit (id, tenant_id, event_type, success, metadata, created_at)
		 VALUES ($1, $2, 'system_provision', true, '{"component":"auth","action":"provisioned"}', NOW())`,
		uuid.New(), tenantID)
	if err != nil {
		log.WithError(err).Warn("Failed to log auth provisioning audit (non-fatal)")
	}

	state.Status = StatusActive
	state.ResourceID = keyID
	log.Info("Auth provisioning complete")
	return state, nil
}

// generateSigningKey creates a cryptographically secure random key
func generateSigningKey(bytes int) (string, error) {
	key := make([]byte, bytes)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
