-- Add default_app_id column to bundle_subscriptions for one-click deploy tracking
-- This column stores the app ID created during bundle provisioning so the dashboard
-- can redirect users to their pre-configured app after activating a bundle

ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS default_app_id uuid REFERENCES apps(id);

CREATE INDEX IF NOT EXISTS idx_bundle_subs_default_app ON bundle_subscriptions(default_app_id) WHERE default_app_id IS NOT NULL;