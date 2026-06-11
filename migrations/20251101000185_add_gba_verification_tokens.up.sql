-- Add verification tokens table for email verification
-- Only runs when gba_tenants exists (add_gba_base_tables may run after this in version order)

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'gba_tenants') THEN
    CREATE TABLE IF NOT EXISTS gba_verification_tokens (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      identifier VARCHAR(255) NOT NULL,
      token VARCHAR(255) NOT NULL UNIQUE,
      expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
      tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_identifier ON gba_verification_tokens(identifier);
    CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_token ON gba_verification_tokens(token);
    CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_tenant ON gba_verification_tokens(tenant_id);
    CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_expires ON gba_verification_tokens(expires_at);
    COMMENT ON TABLE gba_verification_tokens IS 'Email verification tokens for GoBetterAuth';
  END IF;
END $$;
