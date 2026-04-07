-- Fix tenant plan default to 'free' instead of 'starter'
-- This ensures new signups get the free tier, not the starter tier

BEGIN;

-- Drop the existing CHECK constraint (it only allows 'starter' and 'pro')
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_check;

-- Add new CHECK constraint that includes 'free'
ALTER TABLE tenants ADD CONSTRAINT tenants_plan_check
    CHECK (plan IN ('free', 'starter', 'professional', 'enterprise'));

-- Change the default to 'free'
ALTER TABLE tenants ALTER COLUMN plan SET DEFAULT 'free';

-- Update any existing tenants with NULL plan to 'free' (shouldn't happen, but be safe)
UPDATE tenants SET plan = 'free' WHERE plan IS NULL;

COMMIT;
