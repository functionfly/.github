package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetEmailAccountByID retrieves an email account by ID.
func (r *Phase6Repository) GetEmailAccountByID(ctx context.Context, id uuid.UUID) (*EmailAccount, error) {
	ea := &EmailAccount{}
	var aliases, groups []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, email, display_name, provider, provider_account_id, aliases, groups, status, provisioned_at, last_sync_at, created_at, updated_at
		FROM email_accounts WHERE id = $1`, id).Scan(
		&ea.ID, &ea.EmployeeID, &ea.TenantID, &ea.Email, &ea.DisplayName, &ea.Provider, &ea.ProviderAccountID, &aliases, &groups, &ea.Status, &ea.ProvisionedAt, &ea.LastSyncAt, &ea.CreatedAt, &ea.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get email account: %w", err)
	}
	if aliases != nil {
		json.Unmarshal(aliases, &ea.Aliases)
	}
	if groups != nil {
		json.Unmarshal(groups, &ea.Groups)
	}
	return ea, nil
}

// ListEmailAccounts lists email accounts for a tenant.
func (r *Phase6Repository) ListEmailAccounts(ctx context.Context, tenantID uuid.UUID, opts ListEmailAccountsOpts) ([]*EmailAccount, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.EmployeeID != nil {
		where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
		args = append(args, *opts.EmployeeID)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM email_accounts %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count email accounts: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, email, display_name, provider, provider_account_id, aliases, groups, status, provisioned_at, last_sync_at, created_at, updated_at
		FROM email_accounts %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list email accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*EmailAccount
	for rows.Next() {
		ea := &EmailAccount{}
		var aliases, groups []byte
		if err := rows.Scan(&ea.ID, &ea.EmployeeID, &ea.TenantID, &ea.Email, &ea.DisplayName, &ea.Provider, &ea.ProviderAccountID, &aliases, &groups, &ea.Status, &ea.ProvisionedAt, &ea.LastSyncAt, &ea.CreatedAt, &ea.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan email account: %w", err)
		}
		if aliases != nil {
			json.Unmarshal(aliases, &ea.Aliases)
		}
		if groups != nil {
			json.Unmarshal(groups, &ea.Groups)
		}
		accounts = append(accounts, ea)
	}
	return accounts, total, nil
}

// GetDeviceByID retrieves a device by ID.
func (r *Phase6Repository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*Device, error) {
	d := &Device{}
	var metaBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, device_name, device_type, serial_number, os, os_version, manufacturer, model, last_seen_at, compliance_status, certificate_id, enrolled_at, metadata, status, created_at, updated_at
		FROM devices WHERE id = $1`, id).Scan(
		&d.ID, &d.EmployeeID, &d.TenantID, &d.DeviceName, &d.DeviceType, &d.SerialNumber, &d.OS, &d.OSVersion, &d.Manufacturer, &d.Model, &d.LastSeenAt, &d.ComplianceStatus, &d.CertificateID, &d.EnrolledAt, &metaBytes, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if metaBytes != nil {
		var meta JSONMap
		if err := json.Unmarshal(metaBytes, &meta); err == nil {
			d.Metadata = meta
		}
	}
	return d, nil
}

// ListDevices lists devices for a tenant.
func (r *Phase6Repository) ListDevices(ctx context.Context, tenantID uuid.UUID, opts ListDevicesOpts) ([]*Device, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.EmployeeID != nil {
		where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
		args = append(args, *opts.EmployeeID)
		argIdx++
	}
	if opts.DeviceType != nil {
		where += fmt.Sprintf(" AND device_type = $%d", argIdx)
		args = append(args, *opts.DeviceType)
		argIdx++
	}
	if opts.ComplianceStatus != nil {
		where += fmt.Sprintf(" AND compliance_status = $%d", argIdx)
		args = append(args, *opts.ComplianceStatus)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM devices %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count devices: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, device_name, device_type, serial_number, os, os_version, manufacturer, model, last_seen_at, compliance_status, certificate_id, enrolled_at, metadata, status, created_at, updated_at
		FROM devices %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		d := &Device{}
		var metaBytes []byte
		if err := rows.Scan(&d.ID, &d.EmployeeID, &d.TenantID, &d.DeviceName, &d.DeviceType, &d.SerialNumber, &d.OS, &d.OSVersion, &d.Manufacturer, &d.Model, &d.LastSeenAt, &d.ComplianceStatus, &d.CertificateID, &d.EnrolledAt, &metaBytes, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan device: %w", err)
		}
		if metaBytes != nil {
			var meta JSONMap
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				d.Metadata = meta
			}
		}
		devices = append(devices, d)
	}
	return devices, total, nil
}

// GetSSOProvisioningConfigByID retrieves an SSO config by ID.
func (r *Phase6Repository) GetSSOProvisioningConfigByID(ctx context.Context, id uuid.UUID) (*SSOProvisioningConfig, error) {
	cfg := &SSOProvisioningConfig{}
	var mappingsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, provider, provider_url, client_id, client_secret_encrypted, scim_endpoint, scim_token_encrypted, auto_create_employee, auto_update_employee, auto_deactivate, default_department_id, default_clearance, field_mappings, is_active, last_sync_at, created_at, updated_at
		FROM sso_provisioning_configs WHERE id = $1`, id).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.Provider, &cfg.ProviderURL, &cfg.ClientID, &cfg.ClientSecretEncrypted, &cfg.SCIMEndpoint, &cfg.SCIMTokenEncrypted, &cfg.AutoCreateEmployee, &cfg.AutoUpdateEmployee, &cfg.AutoDeactivate, &cfg.DefaultDepartmentID, &cfg.DefaultClearance, &mappingsBytes, &cfg.IsActive, &cfg.LastSyncAt, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provisioning config: %w", err)
	}
	if mappingsBytes != nil {
		var m JSONMap
		if err := json.Unmarshal(mappingsBytes, &m); err == nil {
			cfg.FieldMappings = m
		}
	}
	return cfg, nil
}

// ListSSOProvisioningConfigs lists SSO configs for a tenant.
func (r *Phase6Repository) ListSSOProvisioningConfigs(ctx context.Context, tenantID uuid.UUID, opts ListSSOProvisioningConfigsOpts) ([]*SSOProvisioningConfig, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Provider != nil {
		where += fmt.Sprintf(" AND provider = $%d", argIdx)
		args = append(args, *opts.Provider)
		argIdx++
	}
	if opts.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *opts.IsActive)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM sso_provisioning_configs %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count SSO configs: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, provider, provider_url, client_id, client_secret_encrypted, scim_endpoint, scim_token_encrypted, auto_create_employee, auto_update_employee, auto_deactivate, default_department_id, default_clearance, field_mappings, is_active, last_sync_at, created_at, updated_at
		FROM sso_provisioning_configs %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list SSO configs: %w", err)
	}
	defer rows.Close()

	var configs []*SSOProvisioningConfig
	for rows.Next() {
		cfg := &SSOProvisioningConfig{}
		var mappingsBytes []byte
		if err := rows.Scan(&cfg.ID, &cfg.TenantID, &cfg.Provider, &cfg.ProviderURL, &cfg.ClientID, &cfg.ClientSecretEncrypted, &cfg.SCIMEndpoint, &cfg.SCIMTokenEncrypted, &cfg.AutoCreateEmployee, &cfg.AutoUpdateEmployee, &cfg.AutoDeactivate, &cfg.DefaultDepartmentID, &cfg.DefaultClearance, &mappingsBytes, &cfg.IsActive, &cfg.LastSyncAt, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan SSO config: %w", err)
		}
		if mappingsBytes != nil {
			var m JSONMap
			if err := json.Unmarshal(mappingsBytes, &m); err == nil {
				cfg.FieldMappings = m
			}
		}
		configs = append(configs, cfg)
	}
	return configs, total, nil
}

// ListSSOProvisioningLogs lists provisioning logs for a config.
func (r *Phase6Repository) ListSSOProvisioningLogs(ctx context.Context, configID uuid.UUID, opts ListSSOProvisioningLogsOpts) ([]*SSOProvisioningLog, int, error) {
	where := "WHERE config_id = $1"
	args := []interface{}{configID}
	argIdx := 2

	if opts.Action != nil {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *opts.Action)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM sso_provisioning_logs %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count SSO provisioning logs: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, config_id, external_user_id, employee_id, action, details, error_message, created_at
		FROM sso_provisioning_logs %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list SSO provisioning logs: %w", err)
	}
	defer rows.Close()

	var logs []*SSOProvisioningLog
	for rows.Next() {
		l := &SSOProvisioningLog{}
		var detailsBytes []byte
		if err := rows.Scan(&l.ID, &l.ConfigID, &l.ExternalUserID, &l.EmployeeID, &l.Action, &detailsBytes, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan SSO provisioning log: %w", err)
		}
		if detailsBytes != nil {
			var d JSONMap
			if err := json.Unmarshal(detailsBytes, &d); err == nil {
				l.Details = d
			}
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

// GetWalletPassByID retrieves a wallet pass by ID.
func (r *Phase6Repository) GetWalletPassByID(ctx context.Context, id uuid.UUID) (*WalletPass, error) {
	wp := &WalletPass{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, pass_type, platform, pass_id, qr_token, qr_expires_at, device_id, installed_at, last_presented_at, status, created_at, updated_at
		FROM wallet_passes WHERE id = $1`, id).Scan(
		&wp.ID, &wp.EmployeeID, &wp.TenantID, &wp.PassType, &wp.Platform, &wp.PassID, &wp.QRToken, &wp.QRExpiresAt, &wp.DeviceID, &wp.InstalledAt, &wp.LastPresentedAt, &wp.Status, &wp.CreatedAt, &wp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet pass: %w", err)
	}
	return wp, nil
}

// GetWalletPassByPassID retrieves a wallet pass by its pass_id.
func (r *Phase6Repository) GetWalletPassByPassID(ctx context.Context, passID string) (*WalletPass, error) {
	wp := &WalletPass{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, pass_type, platform, pass_id, qr_token, qr_expires_at, device_id, installed_at, last_presented_at, status, created_at, updated_at
		FROM wallet_passes WHERE pass_id = $1`, passID).Scan(
		&wp.ID, &wp.EmployeeID, &wp.TenantID, &wp.PassType, &wp.Platform, &wp.PassID, &wp.QRToken, &wp.QRExpiresAt, &wp.DeviceID, &wp.InstalledAt, &wp.LastPresentedAt, &wp.Status, &wp.CreatedAt, &wp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet pass by pass_id: %w", err)
	}
	return wp, nil
}

// GetWalletPassByQRToken retrieves a wallet pass by its QR token.
func (r *Phase6Repository) GetWalletPassByQRToken(ctx context.Context, qrToken string) (*WalletPass, error) {
	wp := &WalletPass{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, pass_type, platform, pass_id, qr_token, qr_expires_at, device_id, installed_at, last_presented_at, status, created_at, updated_at
		FROM wallet_passes WHERE qr_token = $1`, qrToken).Scan(
		&wp.ID, &wp.EmployeeID, &wp.TenantID, &wp.PassType, &wp.Platform, &wp.PassID, &wp.QRToken, &wp.QRExpiresAt, &wp.DeviceID, &wp.InstalledAt, &wp.LastPresentedAt, &wp.Status, &wp.CreatedAt, &wp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet pass by QR token: %w", err)
	}
	return wp, nil
}

// ListWalletPasses lists wallet passes for an employee.
func (r *Phase6Repository) ListWalletPasses(ctx context.Context, employeeID uuid.UUID, opts ListWalletPassesOpts) ([]*WalletPass, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM wallet_passes %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count wallet passes: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, pass_type, platform, pass_id, qr_token, qr_expires_at, device_id, installed_at, last_presented_at, status, created_at, updated_at
		FROM wallet_passes %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list wallet passes: %w", err)
	}
	defer rows.Close()

	var passes []*WalletPass
	for rows.Next() {
		wp := &WalletPass{}
		if err := rows.Scan(&wp.ID, &wp.EmployeeID, &wp.TenantID, &wp.PassType, &wp.Platform, &wp.PassID, &wp.QRToken, &wp.QRExpiresAt, &wp.DeviceID, &wp.InstalledAt, &wp.LastPresentedAt, &wp.Status, &wp.CreatedAt, &wp.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan wallet pass: %w", err)
		}
		passes = append(passes, wp)
	}
	return passes, total, nil
}

// ListPushSubscriptions lists push subscriptions for a user.
func (r *Phase6Repository) ListPushSubscriptions(ctx context.Context, userID uuid.UUID) ([]*PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, endpoint, p256dh, auth, user_agent, is_active, created_at, last_used_at
		FROM push_subscriptions WHERE user_id = $1 AND is_active = true ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list push subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*PushSubscription
	for rows.Next() {
		ps := &PushSubscription{}
		if err := rows.Scan(&ps.ID, &ps.UserID, &ps.TenantID, &ps.Endpoint, &ps.P256DH, &ps.Auth, &ps.UserAgent, &ps.IsActive, &ps.CreatedAt, &ps.LastUsedAt); err != nil {
			return nil, fmt.Errorf("failed to scan push subscription: %w", err)
		}
		subs = append(subs, ps)
	}
	return subs, nil
}

// ListNotificationPreferences lists notification preferences for a user.
func (r *Phase6Repository) ListNotificationPreferences(ctx context.Context, userID uuid.UUID, opts ListNotificationPreferencesOpts) ([]*NotificationPreference, int, error) {
	where := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if opts.Channel != nil {
		where += fmt.Sprintf(" AND channel = $%d", argIdx)
		args = append(args, *opts.Channel)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM notification_preferences %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notification preferences: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY channel, event_type LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, user_id, tenant_id, channel, event_type, is_enabled, quiet_hours_start, quiet_hours_end, created_at, updated_at
		FROM notification_preferences %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notification preferences: %w", err)
	}
	defer rows.Close()

	var prefs []*NotificationPreference
	for rows.Next() {
		p := &NotificationPreference{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.TenantID, &p.Channel, &p.EventType, &p.IsEnabled, &p.QuietHoursStart, &p.QuietHoursEnd, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification preference: %w", err)
		}
		prefs = append(prefs, p)
	}
	return prefs, total, nil
}
