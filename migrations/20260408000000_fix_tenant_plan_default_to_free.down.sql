-- Revert tenant plan default back to 'starter'
-- WARNING: This will cause new signups to get starter tier instead of free

BEGIN;

-- Revert default to 'starter'
ALTER TABLE tenants ALTER COLUMN plan SET DEFAULT 'starter';

-- Revert CHECK constraint to only allow 'starter' and 'pro'
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_plan_check
    CHECK (plan IN ('starter', 'pro'));

COMMIT;
