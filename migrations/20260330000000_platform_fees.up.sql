-- Platform fees: wallet, fee records, and RegistryFunction columns

-- Add platform fee columns to registry_functions table
ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS platform_fee_paid BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS platform_fee_amount_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_fee_charged_at TIMESTAMPTZ;

-- Create user_wallets table
CREATE TABLE IF NOT EXISTS user_wallets (
    user_id UUID PRIMARY KEY,
    balance_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,
    lifetime_earnings_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,
    lifetime_fees_usd DECIMAL(14, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create platform_fees table (audit trail)
CREATE TABLE IF NOT EXISTS platform_fees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    fee_type TEXT NOT NULL CHECK (fee_type IN ('publish', 'version_update', 'commission')),
    amount_usd DECIMAL(14, 4) NOT NULL,
    charged_at TIMESTAMPTZ NOT NULL,
    stripe_payment_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for platform_fees
CREATE INDEX IF NOT EXISTS idx_platform_fees_function_id ON platform_fees (function_id);
CREATE INDEX IF NOT EXISTS idx_platform_fees_user_id ON platform_fees (user_id);
CREATE INDEX IF NOT EXISTS idx_platform_fees_fee_type ON platform_fees (fee_type);
CREATE INDEX IF NOT EXISTS idx_platform_fees_charged_at ON platform_fees (charged_at DESC);
CREATE INDEX IF NOT EXISTS idx_platform_fees_status ON platform_fees (status);
CREATE INDEX IF NOT EXISTS idx_platform_fees_stripe_payment_id ON platform_fees (stripe_payment_id) WHERE stripe_payment_id IS NOT NULL;

-- Create fee_transactions table (wallet ledger)
CREATE TABLE IF NOT EXISTS fee_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('credit', 'debit', 'fee_payment', 'commission')),
    amount_usd DECIMAL(14, 4) NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed')),
    reference TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for fee_transactions
CREATE INDEX IF NOT EXISTS idx_fee_transactions_user_id ON fee_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_fee_transactions_kind ON fee_transactions (kind);
CREATE INDEX IF NOT EXISTS idx_fee_transactions_created_at ON fee_transactions (created_at DESC);
