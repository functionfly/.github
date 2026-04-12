-- Down migration: Remove multi-currency support
DROP TABLE IF EXISTS currency_exchange_rates;

ALTER TABLE wallets DROP COLUMN IF EXISTS currency;
ALTER TABLE wallets DROP COLUMN IF EXISTS balance_local;
ALTER TABLE wallets DROP COLUMN IF EXISTS lifetime_earnings_local;
ALTER TABLE wallets DROP COLUMN IF EXISTS lifetime_spent_local;
ALTER TABLE wallets DROP COLUMN IF EXISTS exchange_rate_to_usd;
ALTER TABLE wallets DROP COLUMN IF EXISTS exchange_rate_updated_at;

ALTER TABLE wallet_transactions DROP COLUMN IF EXISTS currency;
ALTER TABLE wallet_transactions DROP COLUMN IF EXISTS amount_local;
ALTER TABLE wallet_transactions DROP COLUMN IF EXISTS exchange_rate;

ALTER TABLE wallet_balance_history DROP COLUMN IF EXISTS currency;
ALTER TABLE wallet_balance_history DROP COLUMN IF EXISTS balance_local;
