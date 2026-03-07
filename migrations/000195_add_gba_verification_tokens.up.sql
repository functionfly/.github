-- Add verification tokens table for email verification
CREATE TABLE IF NOT EXISTS gba_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL, -- Usually email address
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_identifier ON gba_verification_tokens(identifier);
CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_token ON gba_verification_tokens(token);
CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_tenant ON gba_verification_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_verification_tokens_expires ON gba_verification_tokens(expires_at);

-- Add comment for documentation
COMMENT ON TABLE gba_verification_tokens IS 'Email verification tokens for GoBetterAuth';