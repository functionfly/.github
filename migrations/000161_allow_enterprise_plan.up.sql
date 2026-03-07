-- Allow 'enterprise' in tenants.plan (best plan for admins)
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_plan_check CHECK (plan IN ('starter', 'pro', 'enterprise'));
