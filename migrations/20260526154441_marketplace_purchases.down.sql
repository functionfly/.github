BEGIN;

DROP INDEX IF EXISTS idx_marketplace_plan_subscriptions_subscriber_user;
DROP INDEX IF EXISTS idx_marketplace_plan_subscriptions_subscriber_tenant;
DROP INDEX IF EXISTS idx_marketplace_license_grants_purchaser_user;
DROP INDEX IF EXISTS idx_marketplace_license_grants_purchaser_tenant;

DROP TABLE IF EXISTS marketplace_purchase_audit_log;
DROP TABLE IF EXISTS marketplace_purchase_idempotency;
DROP TABLE IF EXISTS function_purchases;
DROP TABLE IF EXISTS agent_hirings;

COMMIT;
