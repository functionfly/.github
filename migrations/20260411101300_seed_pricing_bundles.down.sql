-- Rollback: Remove seeded bundle data
DELETE FROM deferred_billing_configs WHERE is_default = true;
DELETE FROM pricing_bundles WHERE slug IN ('saas-starter', 'marketplace', 'ai-app');
