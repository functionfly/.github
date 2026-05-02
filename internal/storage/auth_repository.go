package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateAuthSettings creates new auth settings for a tenant
func (r *BillingRepository) CreateAuthSettings(ctx context.Context, settings *TenantAuthSettings) error {
	if settings.ID == uuid.Nil {
		settings.ID = uuid.New()
	}
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_auth_settings (
			id, tenant_id, mfa_required, mfa_mode, password_policy,
			session_timeout_minutes, ip_allowlist_enabled, ip_allowlist,
			allowed_domains, sso_provider, saml_metadata_url, saml_entity_id,
			saml_certificate, saml_private_key, use_custom_branding,
			email_from_name, email_from_address, require_email_verification,
			allow_password_login, allow_magic_link, max_login_attempts,
			lockout_duration_minutes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)`
	_, err := r.db.ExecContext(ctx, query,
		settings.ID, settings.TenantID, settings.MFARequired, settings.MFAMode, settings.PasswordPolicy,
		settings.SessionTimeoutMinutes, settings.IPAllowlistEnabled, settings.IPAllowlist,
		settings.AllowedDomains, settings.SSOProvider, settings.SAMLMetadataURL, settings.SAMLEntityID,
		settings.SAMLCertificate, settings.SAMLPrivateKey, settings.UseCustomBranding,
		settings.EmailFromName, settings.EmailFromAddress, settings.RequireEmailVerification,
		settings.AllowPasswordLogin, settings.AllowMagicLink, settings.MaxLoginAttempts,
		settings.LockoutDurationMinutes, settings.CreatedAt, settings.UpdatedAt,
	)
	return err
}

// GetAuthSettings retrieves auth settings for a tenant
func (r *BillingRepository) GetAuthSettings(ctx context.Context, tenantID uuid.UUID) (*TenantAuthSettings, error) {
	settings := &TenantAuthSettings{}
	query := `
		SELECT id, tenant_id, mfa_required, mfa_mode, password_policy,
			session_timeout_minutes, ip_allowlist_enabled, ip_allowlist,
			allowed_domains, sso_provider, saml_metadata_url, saml_entity_id,
			saml_certificate, saml_private_key, use_custom_branding,
			email_from_name, email_from_address, require_email_verification,
			allow_password_login, allow_magic_link, max_login_attempts,
			lockout_duration_minutes, created_at, updated_at
		FROM tenant_auth_settings WHERE tenant_id = $1`
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&settings.ID, &settings.TenantID, &settings.MFARequired, &settings.MFAMode, &settings.PasswordPolicy,
		&settings.SessionTimeoutMinutes, &settings.IPAllowlistEnabled, &settings.IPAllowlist,
		&settings.AllowedDomains, &settings.SSOProvider, &settings.SAMLMetadataURL, &settings.SAMLEntityID,
		&settings.SAMLCertificate, &settings.SAMLPrivateKey, &settings.UseCustomBranding,
		&settings.EmailFromName, &settings.EmailFromAddress, &settings.RequireEmailVerification,
		&settings.AllowPasswordLogin, &settings.AllowMagicLink, &settings.MaxLoginAttempts,
		&settings.LockoutDurationMinutes, &settings.CreatedAt, &settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// UpdateAuthSettings updates auth settings for a tenant
func (r *BillingRepository) UpdateAuthSettings(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	query := "UPDATE tenant_auth_settings SET updated_at = NOW(), "
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		query += fmt.Sprintf("%s = $%d, ", key, argIndex)
		args = append(args, value)
		argIndex++
	}
	query = query[:len(query)-2] // Remove trailing comma
	query += fmt.Sprintf(" WHERE tenant_id = $%d", argIndex)
	args = append(args, tenantID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteAuthSettings deletes auth settings for a tenant
func (r *BillingRepository) DeleteAuthSettings(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tenant_auth_settings WHERE tenant_id = $1", tenantID)
	return err
}

// CreateOAuthProvider creates an OAuth provider configuration for a tenant
func (r *BillingRepository) CreateOAuthProvider(ctx context.Context, provider *TenantOAuthProvider) error {
	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_oauth_providers (
			id, tenant_id, provider, client_id, encrypted_client_secret,
			encrypted_client_secret_iv, encrypted_client_secret_tag,
			enabled, callback_url, scopes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		provider.ID, provider.TenantID, provider.Provider, provider.ClientID, provider.EncryptedClientSecret,
		provider.EncryptedClientSecretIV, provider.EncryptedClientSecretTag,
		provider.Enabled, provider.CallbackURL, provider.Scopes, provider.CreatedAt, provider.UpdatedAt,
	)
	return err
}

// GetOAuthProvider retrieves an OAuth provider for a tenant
func (r *BillingRepository) GetOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*TenantOAuthProvider, error) {
	p := &TenantOAuthProvider{}
	query := `
		SELECT id, tenant_id, provider, client_id, encrypted_client_secret,
			encrypted_client_secret_iv, encrypted_client_secret_tag,
			enabled, callback_url, scopes, created_at, updated_at
		FROM tenant_oauth_providers WHERE tenant_id = $1 AND provider = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, provider).Scan(
		&p.ID, &p.TenantID, &p.Provider, &p.ClientID, &p.EncryptedClientSecret,
		&p.EncryptedClientSecretIV, &p.EncryptedClientSecretTag,
		&p.Enabled, &p.CallbackURL, &p.Scopes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// ListOAuthProviders lists all OAuth providers for a tenant
func (r *BillingRepository) ListOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error) {
	query := `
		SELECT id, tenant_id, provider, client_id, encrypted_client_secret,
			encrypted_client_secret_iv, encrypted_client_secret_tag,
			enabled, callback_url, scopes, created_at, updated_at
		FROM tenant_oauth_providers WHERE tenant_id = $1 ORDER BY provider`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*TenantOAuthProvider
	for rows.Next() {
		p := &TenantOAuthProvider{}
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Provider, &p.ClientID, &p.EncryptedClientSecret,
			&p.EncryptedClientSecretIV, &p.EncryptedClientSecretTag,
			&p.Enabled, &p.CallbackURL, &p.Scopes, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, nil
}

// UpdateOAuthProvider updates an OAuth provider configuration
func (r *BillingRepository) UpdateOAuthProvider(ctx context.Context, tenantID uuid.UUID, providerName string, updates map[string]interface{}) (*TenantOAuthProvider, error) {
	if len(updates) == 0 {
		return r.GetOAuthProvider(ctx, tenantID, providerName)
	}

	query := "UPDATE tenant_oauth_providers SET updated_at = NOW(), "
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		query += fmt.Sprintf("%s = $%d, ", key, argIndex)
		args = append(args, value)
		argIndex++
	}
	query = query[:len(query)-2]
	query += " WHERE tenant_id = $%d AND provider = $%d"
	args = append(args, tenantID, providerName)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	return r.GetOAuthProvider(ctx, tenantID, providerName)
}

// DeleteOAuthProvider removes an OAuth provider configuration
func (r *BillingRepository) DeleteOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tenant_oauth_providers WHERE tenant_id = $1 AND provider = $2", tenantID, provider)
	return err
}

// GetEnabledOAuthProviders returns only enabled OAuth providers for a tenant
func (r *BillingRepository) GetEnabledOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error) {
	query := `
		SELECT id, tenant_id, provider, client_id, encrypted_client_secret,
			encrypted_client_secret_iv, encrypted_client_secret_tag,
			enabled, callback_url, scopes, created_at, updated_at
		FROM tenant_oauth_providers WHERE tenant_id = $1 AND enabled = true ORDER BY provider`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*TenantOAuthProvider
	for rows.Next() {
		p := &TenantOAuthProvider{}
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Provider, &p.ClientID, &p.EncryptedClientSecret,
			&p.EncryptedClientSecretIV, &p.EncryptedClientSecretTag,
			&p.Enabled, &p.CallbackURL, &p.Scopes, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, nil
}

// CreateInviteCode creates an invite code for a tenant
func (r *BillingRepository) CreateInviteCode(ctx context.Context, invite *TenantInviteCode) error {
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	invite.CreatedAt = time.Now()
	invite.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_invite_codes (
			id, tenant_id, code, email, role, invited_by, expires_at,
			max_uses, uses, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, query,
		invite.ID, invite.TenantID, invite.Code, invite.Email, invite.Role, invite.InvitedBy,
		invite.ExpiresAt, invite.MaxUses, invite.Uses, invite.CreatedAt, invite.UpdatedAt,
	)
	return err
}

// GetInviteCode retrieves an invite code by its code value
func (r *BillingRepository) GetInviteCode(ctx context.Context, code string) (*TenantInviteCode, error) {
	invite := &TenantInviteCode{}
	query := `
		SELECT id, tenant_id, code, email, role, invited_by, expires_at,
			accepted_at, accepted_by, max_uses, uses, created_at, updated_at
		FROM tenant_invite_codes WHERE code = $1`
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&invite.ID, &invite.TenantID, &invite.Code, &invite.Email, &invite.Role, &invite.InvitedBy,
		&invite.ExpiresAt, &invite.AcceptedAt, &invite.AcceptedBy, &invite.MaxUses, &invite.Uses,
		&invite.CreatedAt, &invite.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return invite, err
}

// GetInviteCodesByTenant lists all invite codes for a tenant
func (r *BillingRepository) GetInviteCodesByTenant(ctx context.Context, tenantID uuid.UUID, includeUsed bool) ([]*TenantInviteCode, error) {
	query := `
		SELECT id, tenant_id, code, email, role, invited_by, expires_at,
			accepted_at, accepted_by, max_uses, uses, created_at, updated_at
		FROM tenant_invite_codes WHERE tenant_id = $1`
	if !includeUsed {
		query += " AND accepted_at IS NULL AND expires_at > NOW()"
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*TenantInviteCode
	for rows.Next() {
		invite := &TenantInviteCode{}
		if err := rows.Scan(
			&invite.ID, &invite.TenantID, &invite.Code, &invite.Email, &invite.Role, &invite.InvitedBy,
			&invite.ExpiresAt, &invite.AcceptedAt, &invite.AcceptedBy, &invite.MaxUses, &invite.Uses,
			&invite.CreatedAt, &invite.UpdatedAt,
		); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

// GetInviteCodeByEmail finds an active invite for a specific email
func (r *BillingRepository) GetInviteCodeByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*TenantInviteCode, error) {
	invite := &TenantInviteCode{}
	query := `
		SELECT id, tenant_id, code, email, role, invited_by, expires_at,
			accepted_at, accepted_by, max_uses, uses, created_at, updated_at
		FROM tenant_invite_codes
		WHERE tenant_id = $1 AND email = $2 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&invite.ID, &invite.TenantID, &invite.Code, &invite.Email, &invite.Role, &invite.InvitedBy,
		&invite.ExpiresAt, &invite.AcceptedAt, &invite.AcceptedBy, &invite.MaxUses, &invite.Uses,
		&invite.CreatedAt, &invite.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return invite, err
}

// AcceptInviteCode marks an invite code as accepted by a user
func (r *BillingRepository) AcceptInviteCode(ctx context.Context, code string, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE tenant_invite_codes
		SET accepted_at = $1, accepted_by = $2, updated_at = $1
		WHERE code = $3 AND accepted_at IS NULL AND expires_at > $1`
	result, err := r.db.ExecContext(ctx, query, now, userID, code)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("invite code not found or already used")
	}
	return nil
}

// RevokeInviteCode marks an invite code as revoked (by setting expires_at to now)
func (r *BillingRepository) RevokeInviteCode(ctx context.Context, code string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE tenant_invite_codes SET expires_at = $1, updated_at = $1 WHERE code = $2",
		now, code)
	return err
}

// IncrementInviteCodeUses increments the use counter for an invite code
func (r *BillingRepository) IncrementInviteCodeUses(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tenant_invite_codes SET uses = uses + 1, updated_at = NOW() WHERE code = $1",
		code)
	return err
}

// DeleteExpiredInviteCodes removes expired invite codes
func (r *BillingRepository) DeleteExpiredInviteCodes(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM tenant_invite_codes WHERE expires_at < NOW() AND accepted_at IS NULL")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CreateMembership creates a membership record
func (r *BillingRepository) CreateMembership(ctx context.Context, membership *TenantMembership) error {
	if membership.ID == uuid.Nil {
		membership.ID = uuid.New()
	}
	membership.JoinedAt = time.Now()

	query := `
		INSERT INTO tenant_memberships (
			id, tenant_id, user_id, role, invited_by, invited_at, joined_at,
			last_active_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id, user_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query,
		membership.ID, membership.TenantID, membership.UserID, membership.Role,
		membership.InvitedBy, membership.InvitedAt, membership.JoinedAt,
		membership.LastActiveAt, membership.Status,
	)
	return err
}

// GetMembership retrieves a membership record
func (r *BillingRepository) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*TenantMembership, error) {
	m := &TenantMembership{}
	query := `
		SELECT id, tenant_id, user_id, role, invited_by, invited_at, joined_at,
			last_active_at, status
		FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.InvitedBy, &m.InvitedAt,
		&m.JoinedAt, &m.LastActiveAt, &m.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

// ListMemberships lists all memberships for a tenant
func (r *BillingRepository) ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]*TenantMembership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, invited_by, invited_at, joined_at,
			last_active_at, status
		FROM tenant_memberships WHERE tenant_id = $1 ORDER BY joined_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*TenantMembership
	for rows.Next() {
		m := &TenantMembership{}
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.InvitedBy, &m.InvitedAt,
			&m.JoinedAt, &m.LastActiveAt, &m.Status,
		); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

// ListMembershipsByRole lists all memberships with a specific role
func (r *BillingRepository) ListMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*TenantMembership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, invited_by, invited_at, joined_at,
			last_active_at, status
		FROM tenant_memberships WHERE tenant_id = $1 AND role = $2 ORDER BY joined_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*TenantMembership
	for rows.Next() {
		m := &TenantMembership{}
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.InvitedBy, &m.InvitedAt,
			&m.JoinedAt, &m.LastActiveAt, &m.Status,
		); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

// UpdateMembership updates a membership record
func (r *BillingRepository) UpdateMembership(ctx context.Context, tenantID, userID uuid.UUID, updates map[string]interface{}) (*TenantMembership, error) {
	if len(updates) == 0 {
		return r.GetMembership(ctx, tenantID, userID)
	}

	query := "UPDATE tenant_memberships SET "
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		query += fmt.Sprintf("%s = $%d, ", key, argIndex)
		args = append(args, value)
		argIndex++
	}
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE tenant_id = $%d AND user_id = $%d", argIndex, argIndex+1)
	args = append(args, tenantID, userID)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	return r.GetMembership(ctx, tenantID, userID)
}

// DeleteMembership removes a membership record
func (r *BillingRepository) DeleteMembership(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2", tenantID, userID)
	return err
}

// UpdateMembershipLastActive updates the last active timestamp
func (r *BillingRepository) UpdateMembershipLastActive(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tenant_memberships SET last_active_at = NOW() WHERE tenant_id = $1 AND user_id = $2",
		tenantID, userID)
	return err
}

// CountMembershipsByTenant counts the total memberships for a tenant
func (r *BillingRepository) CountMembershipsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1 AND status = 'active'",
		tenantID).Scan(&count)
	return count, err
}

// CountMembershipsByRole counts memberships for a tenant with a specific role
func (r *BillingRepository) CountMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tenant_memberships WHERE tenant_id = $1 AND role = $2 AND status = 'active'",
		tenantID, role).Scan(&count)
	return count, err
}

// CreateAuthAuditLog creates an auth audit log entry
func (r *BillingRepository) CreateAuthAuditLog(ctx context.Context, log *TenantAuthAuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO tenant_auth_audit_log (
			id, tenant_id, user_id, action, resource_type, resource_id,
			ip_address, user_agent, metadata, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.IPAddress, log.UserAgent, log.Metadata, log.Success, log.ErrorMessage, log.CreatedAt,
	)
	return err
}

// ListAuthAuditLogs retrieves auth audit logs with filtering and pagination
func (r *BillingRepository) ListAuthAuditLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int, actions []string, userID *uuid.UUID, since *time.Time) ([]*TenantAuthAuditLog, int, error) {
	query := `SELECT id, tenant_id, user_id, action, resource_type, resource_id,
		ip_address, user_agent, metadata, success, error_message, created_at
		FROM tenant_auth_audit_log WHERE tenant_id = $1`
	countQuery := `SELECT COUNT(*) FROM tenant_auth_audit_log WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}
	if since != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *since)
		argIndex++
	}
	if len(actions) > 0 {
		query += fmt.Sprintf(" AND action = ANY($%d)", argIndex)
		countQuery += fmt.Sprintf(" AND action = ANY($%d)", argIndex)
		args = append(args, pq.Array(actions))
		argIndex++
	}

	// Get total count
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*TenantAuthAuditLog
	for rows.Next() {
		log := &TenantAuthAuditLog{}
		if err := rows.Scan(
			&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.IPAddress, &log.UserAgent, &log.Metadata, &log.Success, &log.ErrorMessage, &log.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	return logs, total, nil
}

// GetAuthAuditLogsByUser retrieves recent auth logs for a specific user
func (r *BillingRepository) GetAuthAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*TenantAuthAuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource_type, resource_id,
			ip_address, user_agent, metadata, success, error_message, created_at
		FROM tenant_auth_audit_log
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY created_at DESC LIMIT $3`
	rows, err := r.db.QueryContext(ctx, query, tenantID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*TenantAuthAuditLog
	for rows.Next() {
		log := &TenantAuthAuditLog{}
		if err := rows.Scan(
			&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.IPAddress, &log.UserAgent, &log.Metadata, &log.Success, &log.ErrorMessage, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// DeleteOldAuthAuditLogs removes old auth audit logs
func (r *BillingRepository) DeleteOldAuthAuditLogs(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM tenant_auth_audit_log WHERE created_at < $1",
		before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetActiveSessionsByTenant retrieves active session events for a tenant
func (r *BillingRepository) GetActiveSessionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantAuthAuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource_type, resource_id,
			ip_address, user_agent, metadata, success, error_message, created_at
		FROM tenant_auth_audit_log
		WHERE tenant_id = $1 AND action IN ('session.created', 'login.success')
		AND created_at > NOW() - INTERVAL '24 hours'
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*TenantAuthAuditLog
	for rows.Next() {
		log := &TenantAuthAuditLog{}
		if err := rows.Scan(
			&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.IPAddress, &log.UserAgent, &log.Metadata, &log.Success, &log.ErrorMessage, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// RevokeAllSessions revokes all sessions for a tenant (marks them in audit log)
func (r *BillingRepository) RevokeAllSessions(ctx context.Context, tenantID uuid.UUID) error {
	// We don't actually delete sessions - we log a session.revoked event
	// Real session revocation happens at the session/JWT level
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenant_auth_audit_log (id, tenant_id, action, success, created_at)
		SELECT gen_random_uuid(), $1, 'session.revoked', true, $2`,
		tenantID, now)
	return err
}
