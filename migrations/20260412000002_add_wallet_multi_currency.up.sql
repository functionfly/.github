-- Migration: Add multi-currency support to wallets
-- Extends wallet tables to support multiple currencies beyond USD

-- Add currency column to wallets table
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS balance_local DECIMAL(14,4) NOT NULL DEFAULT 0;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS lifetime_earnings_local DECIMAL(14,4) NOT NULL DEFAULT 0;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS lifetime_spent_local DECIMAL(14,4) NOT NULL DEFAULT 0;

-- Add exchange rate tracking
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS exchange_rate_to_usd DECIMAL(12,6);
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS exchange_rate_updated_at TIMESTAMP WITH TIME ZONE;

-- Add currency to wallet_transactions
ALTER TABLE wallet_transactions ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';
ALTER TABLE wallet_transactions ADD COLUMN IF NOT EXISTS amount_local DECIMAL(14,4);
ALTER TABLE wallet_transactions ADD COLUMN IF NOT EXISTS exchange_rate DECIMAL(12,6);

-- Add currency to wallet_balance_history
ALTER TABLE wallet_balance_history ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';
ALTER TABLE wallet_balance_history ADD COLUMN IF NOT EXISTS balance_local DECIMAL(14,4);

-- Create currency exchange rates table
CREATE TABLE IF NOT EXISTS currency_exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(12,6) NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'manual', -- 'stripe', 'openexchangerates', 'manual', 'fixed'
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(from_currency, to_currency, valid_from)
);

-- Indexes for exchange rates
CREATE INDEX idx_exchange_rates_from_to ON currency_exchange_rates(from_currency, to_currency);
CREATE INDEX idx_exchange_rates_active ON currency_exchange_rates(is_active, valid_from DESC);

-- Add index on wallet currency
CREATE INDEX idx_wallets_currency ON wallets(currency);
CREATE INDEX idx_wallet_transactions_currency ON wallet_transactions(currency);

-- Comments
COMMENT ON TABLE currency_exchange_rates IS 'Exchange rates for multi-currency wallet support';
COMMENT ON COLUMN wallets.currency IS 'Primary currency for the wallet (ISO 4217 code)';
COMMENT ON COLUMN wallet_transactions.currency IS 'Currency of the transaction amount';
