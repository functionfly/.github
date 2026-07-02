-- Remove bundle function templates
DROP TABLE IF EXISTS bundle_function_templates;

-- Remove deployment columns from bundle_subscriptions
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS script_name;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS provider_id;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS deployed_at;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS deploy_error;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS deploy_attempts;
ALTER TABLE bundle_subscriptions DROP COLUMN IF EXISTS deploy_status;

DROP INDEX IF EXISTS idx_bundle_subscriptions_deploy_status;
