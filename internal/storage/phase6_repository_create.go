package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateEmailAccount creates a new email account.
func (r *Phase6Repository) CreateEmailAccount(ctx context.Context, ea *EmailAccount) (*EmailAccount, error) {
	ea.ID = uuid.New()
	ea.CreatedAt = time.Now()
	ea.UpdatedAt = time.Now()

	var aliasesParam, groupsParam interface{}
	if ea.Aliases != nil {
		aliasesParam = ea.Aliases
	}
	if ea.Groups != nil {
		groupsParam = ea.Groups
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO email_accounts (id, employee_id, tenant_id, email, display_name, provider, provider_account_id, aliases, groups, status, provisioned_at, last_sync_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		ea.ID, ea.EmployeeID, ea.TenantID, ea.Email, ea.DisplayName, ea.Provider, ea.ProviderAccountID, aliasesParam, groupsParam, ea.Status, ea.ProvisionedAt, ea.LastSyncAt, ea.CreatedAt, ea.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create email account: %w", err)
	}
	return ea, nil
}

// CreateDevice registers a new device.
func (r *Phase6Repository) CreateDevice(ctx context.Context, d *Device) (*Device, error) {
	d.ID = uuid.New()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()

	var metaParam interface{}
	if d.Metadata != nil {
		b, _ := json.Marshal(d.Metadata)
		metaParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO devices (id, employee_id, tenant_id, device_name, device_type, serial_number, os, os_version, manufacturer, model, last_seen_at, compliance_status, certificate_id, enrolled_at, metadata, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		d.ID, d.EmployeeID, d.TenantID, d.DeviceName, d.DeviceType, d.SerialNumber, d.OS, d.OSVersion, d.Manufacturer, d.Model, d.LastSeenAt, d.ComplianceStatus, d.CertificateID, d.EnrolledAt, metaParam, d.Status, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}
	return d, nil
}

// CreateSSOProvisioningConfig creates a new SSO provisioning config.
func (r *Phase6Repository) CreateSSOProvisioningConfig(ctx context.Context, cfg *SSOProvisioningConfig) (*SSOProvisioningConfig, error) {
	cfg.ID = uuid.New()
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()

	var mappingsParam interface{}
	if cfg.FieldMappings != nil {
		b, _ := json.Marshal(cfg.FieldMappings)
		mappingsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sso_provisioning_configs (id, tenant_id, provider, provider_url, client_id, client_secret_encrypted, scim_endpoint, scim_token_encrypted, auto_create_employee, auto_update_employee, auto_deactivate, default_department_id, default_clearance, field_mappings, is_active, last_sync_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		cfg.ID, cfg.TenantID, cfg.Provider, cfg.ProviderURL, cfg.ClientID, cfg.ClientSecretEncrypted, cfg.SCIMEndpoint, cfg.SCIMTokenEncrypted, cfg.AutoCreateEmployee, cfg.AutoUpdateEmployee, cfg.AutoDeactivate, cfg.DefaultDepartmentID, cfg.DefaultClearance, mappingsParam, cfg.IsActive, cfg.LastSyncAt, cfg.CreatedAt, cfg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSO provisioning config: %w", err)
	}
	return cfg, nil
}

// CreateSSOProvisioningLog creates a new SSO provisioning log entry.
func (r *Phase6Repository) CreateSSOProvisioningLog(ctx context.Context, log *SSOProvisioningLog) (*SSOProvisioningLog, error) {
	log.CreatedAt = time.Now()

	var detailsParam interface{}
	if log.Details != nil {
		b, _ := json.Marshal(log.Details)
		detailsParam = b
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO sso_provisioning_logs (config_id, external_user_id, employee_id, action, details, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		log.ConfigID, log.ExternalUserID, log.EmployeeID, log.Action, detailsParam, log.ErrorMessage, log.CreatedAt,
	).Scan(&log.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSO provisioning log: %w", err)
	}
	return log, nil
}

// CreateWalletPass creates a new wallet pass.
func (r *Phase6Repository) CreateWalletPass(ctx context.Context, wp *WalletPass) (*WalletPass, error) {
	wp.ID = uuid.New()
	wp.CreatedAt = time.Now()
	wp.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wallet_passes (id, employee_id, tenant_id, pass_type, platform, pass_id, qr_token, qr_expires_at, device_id, installed_at, last_presented_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		wp.ID, wp.EmployeeID, wp.TenantID, wp.PassType, wp.Platform, wp.PassID, wp.QRToken, wp.QRExpiresAt, wp.DeviceID, wp.InstalledAt, wp.LastPresentedAt, wp.Status, wp.CreatedAt, wp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet pass: %w", err)
	}
	return wp, nil
}

// CreatePushSubscription creates a new push subscription.
func (r *Phase6Repository) CreatePushSubscription(ctx context.Context, ps *PushSubscription) (*PushSubscription, error) {
	ps.ID = uuid.New()
	ps.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO push_subscriptions (id, user_id, tenant_id, endpoint, p256dh, auth, user_agent, is_active, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		ps.ID, ps.UserID, ps.TenantID, ps.Endpoint, ps.P256DH, ps.Auth, ps.UserAgent, ps.IsActive, ps.CreatedAt, ps.LastUsedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create push subscription: %w", err)
	}
	return ps, nil
}

// UpsertNotificationPreference creates or updates a notification preference.
func (r *Phase6Repository) UpsertNotificationPreference(ctx context.Context, pref *NotificationPreference) (*NotificationPreference, error) {
	pref.ID = uuid.New()
	pref.CreatedAt = time.Now()
	pref.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO notification_preferences (id, user_id, tenant_id, channel, event_type, is_enabled, quiet_hours_start, quiet_hours_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, tenant_id, channel, event_type)
		DO UPDATE SET is_enabled = EXCLUDED.is_enabled, quiet_hours_start = EXCLUDED.quiet_hours_start, quiet_hours_end = EXCLUDED.quiet_hours_end, updated_at = NOW()
		RETURNING id`,
		pref.ID, pref.UserID, pref.TenantID, pref.Channel, pref.EventType, pref.IsEnabled, pref.QuietHoursStart, pref.QuietHoursEnd, pref.CreatedAt, pref.UpdatedAt,
	).Scan(&pref.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert notification preference: %w", err)
	}
	return pref, nil
}
