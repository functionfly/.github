-- Vault client-encrypted dynamic credentials (2026-06-16 design)
-- Additive: no destructive changes. Safe to apply on a live DB.

-- ============================================================================
-- 1. Per-tenant DEK storage (one row per user per tenant)
-- ============================================================================
CREATE TABLE IF NOT EXISTS vault_tenant_keys (
  tenant_id      UUID         NOT NULL,
  user_id        UUID         NOT NULL,
  wrapped_dek    BYTEA        NOT NULL,
  dek_iv         BYTEA        NOT NULL,
  dek_auth_tag   BYTEA        NOT NULL,
  dek_salt       BYTEA        NOT NULL,
  key_version    INT          NOT NULL DEFAULT 3,
  kdf_params     JSONB        NOT NULL DEFAULT '{"t":3,"m":65536,"p":4}'::jsonb,
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  rotated_at     TIMESTAMPTZ,
  PRIMARY KEY (tenant_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_vault_tenant_keys_tenant ON vault_tenant_keys(tenant_id);

-- ============================================================================
-- 2. Reserved for v2 asymmetric key wrapping (unused in v1)
-- ============================================================================
CREATE TABLE IF NOT EXISTS vault_user_keys (
  tenant_id    UUID         NOT NULL,
  user_id      UUID         NOT NULL,
  public_key   BYTEA        NOT NULL,
  key_type     TEXT         NOT NULL DEFAULT 'x25519',
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id)
);

-- ============================================================================
-- 3. dynamic_secret_targets: add envelope + namespace columns
-- ============================================================================
ALTER TABLE dynamic_secret_targets
  ADD COLUMN IF NOT EXISTS encryption_mode      TEXT  NOT NULL DEFAULT 'server',
  ADD COLUMN IF NOT EXISTS key_version          INT   NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS wrap_iv              BYTEA,
  ADD COLUMN IF NOT EXISTS wrap_auth_tag        BYTEA,
  ADD COLUMN IF NOT EXISTS namespace            TEXT  NOT NULL DEFAULT '/';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_dynamic_secret_targets_encryption_mode'
  ) THEN
    ALTER TABLE dynamic_secret_targets
      ADD CONSTRAINT chk_dynamic_secret_targets_encryption_mode
      CHECK (encryption_mode IN ('server', 'client'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_dst_namespace ON dynamic_secret_targets(tenant_id, namespace);

-- ============================================================================
-- 4. CI/CD agent tokens for dynamic credentials (mirrors static-secret AccessToken)
-- ============================================================================
CREATE TABLE IF NOT EXISTS dynamic_wrapped_access_tokens (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID         NOT NULL,
  credential_id   UUID         NOT NULL,
  token_hash      TEXT         NOT NULL UNIQUE,
  name            TEXT         NOT NULL,
  scopes          JSONB        NOT NULL DEFAULT '[]'::jsonb,
  expires_at      TIMESTAMPTZ  NOT NULL,
  is_revoked      BOOLEAN      NOT NULL DEFAULT false,
  revoked_at      TIMESTAMPTZ,
  revoked_reason  TEXT,
  allowed_ips     JSONB        NOT NULL DEFAULT '[]'::jsonb,
  denied_ips      JSONB        NOT NULL DEFAULT '[]'::jsonb,
  ip_restriction_enabled BOOLEAN NOT NULL DEFAULT false,
  last_used_at    TIMESTAMPTZ,
  use_count       INT          NOT NULL DEFAULT 0,
  created_by      UUID         NOT NULL,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dwat_tenant_credential
  ON dynamic_wrapped_access_tokens(tenant_id, credential_id);

-- ============================================================================
-- 5. Cross-tenant target shares (stub for v2; unused in v1)
-- ============================================================================
CREATE TABLE IF NOT EXISTS dynamic_target_shares (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  target_id            UUID         NOT NULL,
  source_tenant_id     UUID         NOT NULL,
  granted_to_tenant_id UUID         NOT NULL,
  granted_by_user      UUID         NOT NULL,
  permissions          TEXT         NOT NULL DEFAULT 'read',
  expires_at           TIMESTAMPTZ,
  revoked_at           TIMESTAMPTZ,
  revoked_by           UUID,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dts_grantee_active
  ON dynamic_target_shares(granted_to_tenant_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dts_target
  ON dynamic_target_shares(target_id);
