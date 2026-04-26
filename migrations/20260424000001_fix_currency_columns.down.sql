-- Revert currency column widths to VARCHAR(3)
ALTER TABLE supported_currencies ALTER COLUMN code TYPE VARCHAR(3);
ALTER TABLE currency_exchange_rates ALTER COLUMN base_currency TYPE VARCHAR(3);
ALTER TABLE currency_exchange_rates ALTER COLUMN quote_currency TYPE VARCHAR(3);

-- Remove added columns
ALTER TABLE currency_exchange_rates DROP COLUMN IF EXISTS rate_numerator;
ALTER TABLE currency_exchange_rates DROP COLUMN IF EXISTS rate_denominator;

-- Restore original convert_currency function (direct rates only)
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
