-- Fix column widths for multi-currency support
-- Some currency codes like HUF exceed the original VARCHAR(3) limit

ALTER TABLE supported_currencies ALTER COLUMN code TYPE VARCHAR(10);
ALTER TABLE currency_exchange_rates ALTER COLUMN base_currency TYPE VARCHAR(10);
ALTER TABLE currency_exchange_rates ALTER COLUMN quote_currency TYPE VARCHAR(10);

-- Add rate_numerator and rate_denominator columns if they don't exist
ALTER TABLE currency_exchange_rates ADD COLUMN IF NOT EXISTS rate_numerator BIGINT DEFAULT 0;
ALTER TABLE currency_exchange_rates ADD COLUMN IF NOT EXISTS rate_denominator BIGINT DEFAULT 1000000;

-- Update numerator/denominator from existing rates
UPDATE currency_exchange_rates
SET rate_numerator = (rate::numeric * 1000000)::bigint,
    rate_denominator = 1000000
WHERE rate_numerator = 0 AND rate > 0;

-- Update convert_currency function to support cross-rates
CREATE OR REPLACE FUNCTION convert_currency(
    amount_cents INTEGER,
    from_currency VARCHAR(10),
    to_currency VARCHAR(10),
    conversion_date DATE DEFAULT CURRENT_DATE
) RETURNS INTEGER AS $$
DECLARE
    rate DECIMAL(18, 8);
    result INTEGER;
BEGIN
    IF from_currency = to_currency THEN
        RETURN amount_cents;
    END IF;

    -- Try direct rate first
    SELECT er.rate INTO rate
    FROM currency_exchange_rates er
    WHERE er.base_currency = from_currency
      AND er.quote_currency = to_currency
      AND er.effective_date = conversion_date
      AND er.is_stripe_rate = false
    ORDER BY er.fetched_at DESC NULLS LAST
    LIMIT 1;

    -- If no direct rate, try cross-rate via USD
    IF rate IS NULL AND from_currency != 'USD' AND to_currency != 'USD' THEN
        SELECT (er2.rate / er1.rate) INTO rate
        FROM currency_exchange_rates er1,
             currency_exchange_rates er2
        WHERE er1.base_currency = 'USD' AND er1.quote_currency = from_currency
          AND er2.base_currency = 'USD' AND er2.quote_currency = to_currency
          AND er1.effective_date = conversion_date
          AND er2.effective_date = conversion_date
          AND er1.is_stripe_rate = false
          AND er2.is_stripe_rate = false
        ORDER BY er1.fetched_at DESC NULLS LAST, er2.fetched_at DESC NULLS LAST
        LIMIT 1;
    END IF;

    IF rate IS NULL THEN
        RAISE EXCEPTION 'No exchange rate found for % to % on %', from_currency, to_currency, conversion_date;
    END IF;

    result := ROUND(amount_cents * rate);
    RETURN result;
END;
$$ LANGUAGE plpgsql;
