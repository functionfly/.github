-- Multi-currency exchange rates table
-- Supports real-time and daily exchange rates for pricing conversions
CREATE TABLE IF NOT EXISTS currency_exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Currency pair (always stored as BASE/QUOTE, e.g., USD/EUR = 0.92)
    base_currency VARCHAR(3) NOT NULL,
    quote_currency VARCHAR(3) NOT NULL,
    
    -- Exchange rate (how much quote currency you get for 1 base currency)
    rate DECIMAL(18, 8) NOT NULL,
    
    -- Rate source
    source VARCHAR(50) NOT NULL DEFAULT 'manual', -- 'ecb', 'openexchange', 'manual', 'stripe'
    source_url VARCHAR(200),
    
    -- Effective date (allows historical rates)
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Metadata for financial audit
    fetched_at TIMESTAMP WITH TIME ZONE,
    is_manual_override BOOLEAN DEFAULT false,
    override_reason TEXT,
    
    -- For Stripe-specific rates (they add ~1-2% markup)
    is_stripe_rate BOOLEAN DEFAULT false,
    stripe_precision VARCHAR(10) DEFAULT '2', -- '0', '2', '4' decimal places
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(base_currency, quote_currency, effective_date, is_stripe_rate)
);

-- Pre-populate with common rates (approximate, should be updated via API)
-- USD base rates (as of ~2024)
INSERT INTO currency_exchange_rates (base_currency, quote_currency, rate, source, effective_date) VALUES
    ('USD', 'EUR', 0.92000000, 'seed', CURRENT_DATE),
    ('USD', 'GBP', 0.79000000, 'seed', CURRENT_DATE),
    ('USD', 'JPY', 148.50000000, 'seed', CURRENT_DATE),
    ('USD', 'CAD', 1.35000000, 'seed', CURRENT_DATE),
    ('USD', 'AUD', 1.52000000, 'seed', CURRENT_DATE),
    ('USD', 'CHF', 0.88000000, 'seed', CURRENT_DATE),
    ('USD', 'SEK', 10.40000000, 'seed', CURRENT_DATE),
    ('USD', 'NOK', 10.60000000, 'seed', CURRENT_DATE),
    ('USD', 'DKK', 6.89000000, 'seed', CURRENT_DATE),
    ('USD', 'PLN', 4.02000000, 'seed', CURRENT_DATE),
    ('USD', 'CZK', 23.40000000, 'seed', CURRENT_DATE),
    ('USD', 'HUF', 362.00000000, 'seed', CURRENT_DATE),
    ('USD', 'RON', 4.60000000, 'seed', CURRENT_DATE),
    ('USD', 'MXN', 17.10000000, 'seed', CURRENT_DATE),
    ('USD', 'BRL', 4.95000000, 'seed', CURRENT_DATE),
    ('USD', 'SGD', 1.34000000, 'seed', CURRENT_DATE),
    ('USD', 'HKD', 7.82000000, 'seed', CURRENT_DATE),
    ('USD', 'NZD', 1.64000000, 'seed', CURRENT_DATE),
    ('USD', 'INR', 83.10000000, 'seed', CURRENT_DATE),
    ('USD', 'KRW', 1330.00000000, 'seed', CURRENT_DATE),
    ('USD', 'TWD', 31.20000000, 'seed', CURRENT_DATE),
    ('USD', 'THB', 35.80000000, 'seed', CURRENT_DATE),
    ('USD', 'MYR', 4.72000000, 'seed', CURRENT_DATE),
    ('USD', 'PHP', 56.10000000, 'seed', CURRENT_DATE),
    ('USD', 'IDR', 15600.00000000, 'seed', CURRENT_DATE)
ON CONFLICT (base_currency, quote_currency, effective_date, is_stripe_rate) DO NOTHING;

-- Inverse rates (for EUR to USD, etc.)
INSERT INTO currency_exchange_rates (base_currency, quote_currency, rate, source, effective_date)
SELECT quote_currency, base_currency, 1.0 / rate, 'seed', effective_date
FROM currency_exchange_rates
WHERE base_currency = 'USD'
ON CONFLICT (base_currency, quote_currency, effective_date, is_stripe_rate) DO NOTHING;

-- Supported currencies configuration table
CREATE TABLE IF NOT EXISTS supported_currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    symbol_position VARCHAR(10) DEFAULT 'before', -- 'before' ($100) or 'after' (100 €)
    decimal_places INTEGER DEFAULT 2,
    thousands_separator VARCHAR(5) DEFAULT ',',
    decimal_separator VARCHAR(5) DEFAULT '.',
    is_active BOOLEAN DEFAULT true,
    is_stablecoin BOOLEAN DEFAULT false,
    
    -- For crypto/stablecoin
    contract_address VARCHAR(100),
    chain_id INTEGER,
    
    -- Regional settings
    default_country VARCHAR(2),
    supported_countries VARCHAR(2)[],
    
    -- Display settings
    rounding_mode VARCHAR(20) DEFAULT 'half_up', -- 'half_up', 'half_down', 'up', 'down'
    minimum_charge_cents INTEGER DEFAULT 50, -- Stripe minimum for most currencies
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed supported currencies
INSERT INTO supported_currencies (code, name, symbol, is_active) VALUES
    ('USD', 'US Dollar', '$', true),
    ('EUR', 'Euro', '€', true),
    ('GBP', 'British Pound', '£', true),
    ('JPY', 'Japanese Yen', '¥', true),
    ('CAD', 'Canadian Dollar', 'C$', true),
    ('AUD', 'Australian Dollar', 'A$', true),
    ('CHF', 'Swiss Franc', 'Fr', true),
    ('SEK', 'Swedish Krona', 'kr', true),
    ('NOK', 'Norwegian Krone', 'kr', true),
    ('DKK', 'Danish Krone', 'kr', true),
    ('PLN', 'Polish Złoty', 'zł', true),
    ('CZK', 'Czech Koruna', 'Kč', true),
    ('MXN', 'Mexican Peso', '$', true),
    ('BRL', 'Brazilian Real', 'R$', true),
    ('SGD', 'Singapore Dollar', 'S$', true),
    ('HKD', 'Hong Kong Dollar', 'HK$', true),
    ('NZD', 'New Zealand Dollar', 'NZ$', true),
    ('INR', 'Indian Rupee', '₹', true),
    ('KRW', 'South Korean Won', '₩', true),
    ('TWD', 'New Taiwan Dollar', 'NT$', true),
    ('THB', 'Thai Baht', '฿', true),
    ('MYR', 'Malaysian Ringgit', 'RM', true),
    ('PHP', 'Philippine Peso', '₱', true),
    -- Stablecoins
    ('USDC', 'USD Coin', 'USDC', true),
    ('USDT', 'Tether', 'USDT', true)
ON CONFLICT (code) DO NOTHING;

-- Update USDC/USDT as stablecoins
UPDATE supported_currencies SET is_stablecoin = true WHERE code IN ('USDC', 'USDT');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_exchange_rates_base_quote ON currency_exchange_rates(base_currency, quote_currency);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_effective ON currency_exchange_rates(effective_date);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_stripe ON currency_exchange_rates(is_stripe_rate);
CREATE INDEX IF NOT EXISTS idx_supported_currencies_active ON supported_currencies(is_active);

-- Currency conversion helper function
CREATE OR REPLACE FUNCTION convert_currency(
    amount_cents INTEGER,
    from_currency VARCHAR(3),
    to_currency VARCHAR(3),
    conversion_date DATE DEFAULT CURRENT_DATE
) RETURNS INTEGER AS $$
DECLARE
    rate DECIMAL(18, 8);
    result INTEGER;
BEGIN
    IF from_currency = to_currency THEN
        RETURN amount_cents;
    END IF;
    
    SELECT er.rate INTO rate
    FROM currency_exchange_rates er
    WHERE er.base_currency = from_currency
      AND er.quote_currency = to_currency
      AND er.effective_date = conversion_date
      AND er.is_stripe_rate = false
    ORDER BY er.fetched_at DESC NULLS LAST
    LIMIT 1;
    
    IF rate IS NULL THEN
        RAISE EXCEPTION 'No exchange rate found for % to % on %', from_currency, to_currency, conversion_date;
    END IF;
    
    result := ROUND(amount_cents * rate);
    RETURN result;
END;
$$ LANGUAGE plpgsql;
