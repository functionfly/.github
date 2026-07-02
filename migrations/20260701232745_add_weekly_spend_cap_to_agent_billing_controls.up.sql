ALTER TABLE agent_billing_controls
  ADD COLUMN IF NOT EXISTS spend_cap_weekly_usd DECIMAL(10,2);

ALTER TABLE wallets
  ADD COLUMN IF NOT EXISTS spend_cap_weekly_usd DECIMAL(10,2);
