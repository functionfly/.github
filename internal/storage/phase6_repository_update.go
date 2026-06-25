package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateEmailAccount updates an email account dynamically.
func (r *Phase6Repository) UpdateEmailAccount(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if displayName, ok := updates["display_name"]; ok {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, displayName)
		argIdx++
	}
	if provider, ok := updates["provider"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("provider = $%d", argIdx))
		args = append(args, provider)
		argIdx++
	}
	if providerAccountID, ok := updates["provider_account_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("provider_account_id = $%d", argIdx))
		args = append(args, providerAccountID)
		argIdx++
	}
	if aliases, ok := updates["aliases"]; ok {
		setParts = append(setParts, fmt.Sprintf("aliases = $%d", argIdx))
		args = append(args, aliases)
		argIdx++
	}
	if groups, ok := updates["groups"]; ok {
		setParts = append(setParts, fmt.Sprintf("groups = $%d", argIdx))
		args = append(args, groups)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if provisionedAt, ok := updates["provisioned_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("provisioned_at = $%d", argIdx))
		args = append(args, provisionedAt)
		argIdx++
	}
	if lastSyncAt, ok := updates["last_sync_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("last_sync_at = $%d", argIdx))
		args = append(args, lastSyncAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE email_accounts SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update email account: %w", err)
	}
	return nil
}

// UpdateDevice updates a device dynamically.
func (r *Phase6Repository) UpdateDevice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if deviceName, ok := updates["device_name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("device_name = $%d", argIdx))
		args = append(args, deviceName)
		argIdx++
	}
	if deviceType, ok := updates["device_type"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("device_type = $%d", argIdx))
		args = append(args, deviceType)
		argIdx++
	}
	if serialNumber, ok := updates["serial_number"]; ok {
		setParts = append(setParts, fmt.Sprintf("serial_number = $%d", argIdx))
		args = append(args, serialNumber)
		argIdx++
	}
	if os, ok := updates["os"]; ok {
		setParts = append(setParts, fmt.Sprintf("os = $%d", argIdx))
		args = append(args, os)
		argIdx++
	}
	if osVersion, ok := updates["os_version"]; ok {
		setParts = append(setParts, fmt.Sprintf("os_version = $%d", argIdx))
		args = append(args, osVersion)
		argIdx++
	}
	if manufacturer, ok := updates["manufacturer"]; ok {
		setParts = append(setParts, fmt.Sprintf("manufacturer = $%d", argIdx))
		args = append(args, manufacturer)
		argIdx++
	}
	if model, ok := updates["model"]; ok {
		setParts = append(setParts, fmt.Sprintf("model = $%d", argIdx))
		args = append(args, model)
		argIdx++
	}
	if complianceStatus, ok := updates["compliance_status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("compliance_status = $%d", argIdx))
		args = append(args, complianceStatus)
		argIdx++
	}
	if lastSeenAt, ok := updates["last_seen_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("last_seen_at = $%d", argIdx))
		args = append(args, lastSeenAt)
		argIdx++
	}
	if enrolledAt, ok := updates["enrolled_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("enrolled_at = $%d", argIdx))
		args = append(args, enrolledAt)
		argIdx++
	}
	if metadata, ok := updates["metadata"]; ok {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, metadata)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE devices SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	return nil
}

// UpdateSSOProvisioningConfig updates an SSO config dynamically.
func (r *Phase6Repository) UpdateSSOProvisioningConfig(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if provider, ok := updates["provider"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("provider = $%d", argIdx))
		args = append(args, provider)
		argIdx++
	}
	if providerURL, ok := updates["provider_url"]; ok {
		setParts = append(setParts, fmt.Sprintf("provider_url = $%d", argIdx))
		args = append(args, providerURL)
		argIdx++
	}
	if clientID, ok := updates["client_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("client_id = $%d", argIdx))
		args = append(args, clientID)
		argIdx++
	}
	if clientSecret, ok := updates["client_secret_encrypted"]; ok {
		setParts = append(setParts, fmt.Sprintf("client_secret_encrypted = $%d", argIdx))
		args = append(args, clientSecret)
		argIdx++
	}
	if scimEndpoint, ok := updates["scim_endpoint"]; ok {
		setParts = append(setParts, fmt.Sprintf("scim_endpoint = $%d", argIdx))
		args = append(args, scimEndpoint)
		argIdx++
	}
	if scimToken, ok := updates["scim_token_encrypted"]; ok {
		setParts = append(setParts, fmt.Sprintf("scim_token_encrypted = $%d", argIdx))
		args = append(args, scimToken)
		argIdx++
	}
	if autoCreate, ok := updates["auto_create_employee"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("auto_create_employee = $%d", argIdx))
		args = append(args, autoCreate)
		argIdx++
	}
	if autoUpdate, ok := updates["auto_update_employee"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("auto_update_employee = $%d", argIdx))
		args = append(args, autoUpdate)
		argIdx++
	}
	if autoDeactivate, ok := updates["auto_deactivate"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("auto_deactivate = $%d", argIdx))
		args = append(args, autoDeactivate)
		argIdx++
	}
	if deptID, ok := updates["default_department_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("default_department_id = $%d", argIdx))
		args = append(args, deptID)
		argIdx++
	}
	if clearance, ok := updates["default_clearance"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("default_clearance = $%d", argIdx))
		args = append(args, clearance)
		argIdx++
	}
	if mappings, ok := updates["field_mappings"]; ok {
		setParts = append(setParts, fmt.Sprintf("field_mappings = $%d", argIdx))
		args = append(args, mappings)
		argIdx++
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, isActive)
		argIdx++
	}
	if lastSyncAt, ok := updates["last_sync_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("last_sync_at = $%d", argIdx))
		args = append(args, lastSyncAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE sso_provisioning_configs SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update SSO config: %w", err)
	}
	return nil
}

// UpdateWalletPass updates a wallet pass dynamically.
func (r *Phase6Repository) UpdateWalletPass(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if platform, ok := updates["platform"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("platform = $%d", argIdx))
		args = append(args, platform)
		argIdx++
	}
	if deviceID, ok := updates["device_id"]; ok {
		setParts = append(setParts, fmt.Sprintf("device_id = $%d", argIdx))
		args = append(args, deviceID)
		argIdx++
	}
	if installedAt, ok := updates["installed_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("installed_at = $%d", argIdx))
		args = append(args, installedAt)
		argIdx++
	}
	if lastPresentedAt, ok := updates["last_presented_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("last_presented_at = $%d", argIdx))
		args = append(args, lastPresentedAt)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE wallet_passes SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update wallet pass: %w", err)
	}
	return nil
}

// UpdateNotificationPreferenceLastUsed updates the last_used_at of a push subscription.
func (r *Phase6Repository) UpdatePushSubscriptionLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE push_subscriptions SET last_used_at = $1 WHERE id = $2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update push subscription last used: %w", err)
	}
	return nil
}
