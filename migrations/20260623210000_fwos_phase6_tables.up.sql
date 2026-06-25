-- Email Accounts (provisioned @functionfly.com addresses)
CREATE TABLE IF NOT EXISTS email_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  email TEXT NOT NULL,
  display_name TEXT,
  provider TEXT NOT NULL DEFAULT 'spacemail',
  provider_account_id TEXT,
  aliases TEXT[] DEFAULT '{}',
  groups TEXT[] DEFAULT '{}',
  status TEXT NOT NULL DEFAULT 'active',
  provisioned_at TIMESTAMPTZ,
  last_sync_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT email_accounts_status_check CHECK (status = ANY (ARRAY['active'::text, 'suspended'::text, 'deactivated'::text]))
);
CREATE INDEX IF NOT EXISTS idx_email_accounts_employee ON email_accounts(employee_id);
CREATE INDEX IF NOT EXISTS idx_email_accounts_email ON email_accounts(email);
CREATE INDEX IF NOT EXISTS idx_email_accounts_tenant ON email_accounts(tenant_id);

-- Devices
CREATE TABLE IF NOT EXISTS devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
  tenant_id UUID NOT NULL,
  device_name TEXT NOT NULL,
  device_type TEXT NOT NULL DEFAULT 'laptop',
  serial_number TEXT,
  os TEXT,
  os_version TEXT,
  manufacturer TEXT,
  model TEXT,
  last_seen_at TIMESTAMPTZ,
  compliance_status TEXT DEFAULT 'compliant',
  certificate_id UUID,
  enrolled_at TIMESTAMPTZ,
  metadata JSONB DEFAULT '{}',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT devices_type_check CHECK (device_type = ANY (ARRAY['laptop'::text, 'desktop'::text, 'phone'::text, 'tablet'::text, 'node'::text])),
  CONSTRAINT devices_compliance_check CHECK (compliance_status = ANY (ARRAY['compliant'::text, 'non_compliant'::text, 'unknown'::text, 'quarantined'::text])),
  CONSTRAINT devices_status_check CHECK (status = ANY (ARRAY['active'::text, 'lost'::text, 'stolen'::text, 'retired'::text]))
);
CREATE INDEX IF NOT EXISTS idx_devices_employee ON devices(employee_id);
CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_serial ON devices(serial_number);

-- SSO Provisioning Configs
CREATE TABLE IF NOT EXISTS sso_provisioning_configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  provider TEXT NOT NULL,
  provider_url TEXT,
  client_id TEXT,
  client_secret_encrypted TEXT,
  scim_endpoint TEXT,
  scim_token_encrypted TEXT,
  auto_create_employee BOOLEAN DEFAULT TRUE,
  auto_update_employee BOOLEAN DEFAULT TRUE,
  auto_deactivate BOOLEAN DEFAULT FALSE,
  default_department_id BIGINT REFERENCES departments(id),
  default_clearance TEXT DEFAULT 'standard',
  field_mappings JSONB DEFAULT '{}',
  is_active BOOLEAN DEFAULT TRUE,
  last_sync_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT sso_provisioning_provider_check CHECK (provider = ANY (ARRAY['saml_scim'::text, 'oidc'::text, 'google_workspace'::text, 'azure_ad'::text, 'okta'::text]))
);
CREATE INDEX IF NOT EXISTS idx_sso_provisioning_tenant ON sso_provisioning_configs(tenant_id);

-- SSO Provisioning Logs
CREATE TABLE IF NOT EXISTS sso_provisioning_logs (
  id BIGSERIAL PRIMARY KEY,
  config_id UUID NOT NULL REFERENCES sso_provisioning_configs(id) ON DELETE CASCADE,
  external_user_id TEXT,
  employee_id UUID REFERENCES employees(id),
  action TEXT NOT NULL,
  details JSONB DEFAULT '{}',
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sso_provisioning_logs_config ON sso_provisioning_logs(config_id);

-- Wallet Passes
CREATE TABLE IF NOT EXISTS wallet_passes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  pass_type TEXT NOT NULL DEFAULT 'employee_badge',
  platform TEXT NOT NULL,
  pass_id TEXT UNIQUE NOT NULL,
  qr_token TEXT NOT NULL,
  qr_expires_at TIMESTAMPTZ NOT NULL,
  device_id TEXT,
  installed_at TIMESTAMPTZ,
  last_presented_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT wallet_passes_type_check CHECK (pass_type = ANY (ARRAY['employee_badge'::text, 'founder_badge'::text, 'visitor'::text])),
  CONSTRAINT wallet_passes_platform_check CHECK (platform = ANY (ARRAY['apple_wallet'::text, 'google_wallet'::text])),
  CONSTRAINT wallet_passes_status_check CHECK (status = ANY (ARRAY['active'::text, 'revoked'::text, 'expired'::text]))
);
CREATE INDEX IF NOT EXISTS idx_wallet_passes_employee ON wallet_passes(employee_id);
CREATE INDEX IF NOT EXISTS idx_wallet_passes_pass_id ON wallet_passes(pass_id);

-- Push Notification Subscriptions
CREATE TABLE IF NOT EXISTS push_subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  tenant_id UUID NOT NULL,
  endpoint TEXT NOT NULL,
  p256dh TEXT NOT NULL,
  auth TEXT NOT NULL,
  user_agent TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user ON push_subscriptions(user_id);

-- Notification Preferences
CREATE TABLE IF NOT EXISTS notification_preferences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  tenant_id UUID NOT NULL,
  channel TEXT NOT NULL,
  event_type TEXT NOT NULL,
  is_enabled BOOLEAN DEFAULT TRUE,
  quiet_hours_start TIME,
  quiet_hours_end TIME,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, tenant_id, channel, event_type)
);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user ON notification_preferences(user_id);
