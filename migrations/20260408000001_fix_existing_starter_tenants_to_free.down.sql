-- Revert: Upgrade 'free' tenants back to 'starter' (use with caution!)
-- WARNING: This will affect all tenants that were downgraded

BEGIN;

-- Preview what will be reverted (excludes admin account)
SELECT
    t.id,
    t.name,
    t.plan,
    t.created_at,
    u.email as owner_email,
    'WILL BE UPGRADED TO STARTER' as action
FROM tenants t
JOIN users u ON u.tenant_id = t.id
WHERE t.plan = 'free'
  AND t.stripe_customer_id IS NULL
  AND u.email != 'thefunctionfly@gmail.com'
ORDER BY t.created_at DESC;

-- Update 'free' tenants back to 'starter' (if they were previously 'starter')
-- Note: We can't know for sure which were previously starter, so this affects all free non-paying tenants
-- Excludes the admin account (thefunctionfly@gmail.com)
UPDATE tenants
SET plan = 'starter',
    updated_at = NOW()
WHERE plan = 'free'
  AND stripe_customer_id IS NULL
  AND id NOT IN (
      SELECT tenant_id FROM users WHERE email = 'thefunctionfly@gmail.com'
  );

COMMIT;
