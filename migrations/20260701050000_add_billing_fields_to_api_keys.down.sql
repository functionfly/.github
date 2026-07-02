ALTER TABLE api_keys DROP COLUMN IF EXISTS billing_budget_cents;
ALTER TABLE api_keys DROP COLUMN IF EXISTS is_high_value;
ALTER TABLE api_keys DROP COLUMN IF EXISTS cost_center;
