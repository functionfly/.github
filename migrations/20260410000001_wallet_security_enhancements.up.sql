-- Wallet Security Enhancements Migration
-- Adds encrypted balance field and security audit columns

-- Add encrypted balance column for at-rest encryption verification
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS balance_encrypted TEXT;

-- Add index for efficient lookups of encrypted wallets
CREATE INDEX IF NOT EXISTS idx_wallets_balance_encrypted 
    ON wallets(balance_encrypted) 
    WHERE balance_encrypted IS NOT NULL;

-- Add admin adjustment audit log table for tracking large adjustments
CREATE TABLE IF NOT EXISTS wallet_admin_adjustment_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES users(id),
    amount_usd DECIMAL(14,4) NOT NULL,
    previous_balance_usd DECIMAL(14,4) NOT NULL,
    new_balance_usd DECIMAL(14,4) NOT NULL,
    reason TEXT NOT NULL,
    reference TEXT,
    requires_secondary_approval BOOLEAN NOT NULL DEFAULT false,
    secondary_approved_by UUID REFERENCES users(id),
    secondary_approved_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add index for admin adjustment lookups
CREATE INDEX IF NOT EXISTS idx_wallet_admin_adjustment_audit_wallet_id 
    ON wallet_admin_adjustment_audit(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_admin_adjustment_audit_admin_id 
    ON wallet_admin_adjustment_audit(admin_id);
CREATE INDEX IF NOT EXISTS idx_wallet_admin_adjustment_audit_created_at 
    ON wallet_admin_adjustment_audit(created_at DESC);

-- Add secondary approval tracking table
CREATE TABLE IF NOT EXISTS wallet_secondary_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    requesting_admin_id UUID NOT NULL REFERENCES users(id),
    amount_usd DECIMAL(14,4) NOT NULL,
    reason TEXT NOT NULL,
    reference TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    approving_admin_id UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes for secondary approval lookups
CREATE INDEX IF NOT EXISTS idx_wallet_secondary_approvals_status 
    ON wallet_secondary_approvals(status);
CREATE INDEX IF NOT EXISTS idx_wallet_secondary_approvals_wallet_id 
    ON wallet_secondary_approvals(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_secondary_approvals_requesting_admin 
    ON wallet_secondary_approvals(requesting_admin_id);

-- Add comment explaining encrypted balance column
COMMENT ON COLUMN wallets.balance_encrypted IS 
    'Encrypted balance verification data (AES-256-GCM) for detecting tampering at rest';
