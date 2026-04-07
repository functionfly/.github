-- Fix existing tenants that were incorrectly assigned 'starter' plan
-- Downgrade to 'free' if they have never paid (no stripe_customer_id)
-- Keep 'starter' only for tenants that have a stripe_customer_id (billing engaged)

BEGIN;

-- Preview what will be changed (excludes admin account)
SELECT 
    t.id,
    t.name,
    t.plan,
    t.stripe_customer_id,
    t.created_at,
    u.email as owner_email,
    'WILL BE DOWNGRADED TO FREE' as action
FROM tenants t
JOIN users u ON u.tenant_id = t.id
WHERE t.plan = 'starter'
  AND t.stripe_customer_id IS NULL
  AND u.email != 'thefunctionfly@gmail.com'
ORDER BY t.created_at DESC;

-- Update tenants from 'starter' to 'free' if they have no Stripe customer (never paid)
-- Excludes the admin account (thefunctionfly@gmail.com)
UPDATE tenants 
SET plan = 'free',
    updated_at = NOW()
WHERE plan = 'starter'
  AND stripe_customer_id IS NULL
  AND id NOT IN (
      SELECT tenant_id FROM users WHERE email = 'thefunctionfly@gmail.com'
  );

-- Show summary of what remains as 'starter' (those with Stripe billing set up)
SELECT 
    COUNT(*) as remaining_starter_count,
    'Tenants with stripe_customer_id keep starter plan' as note
FROM tenants 
WHERE plan = 'starter';

COMMIT;
