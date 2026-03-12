-- Extended factory settings for complex admin configuration
ALTER TABLE factory_config
  ADD COLUMN IF NOT EXISTS notification_webhook_url TEXT,
  ADD COLUMN IF NOT EXISTS rate_limit_per_hour INTEGER NOT NULL DEFAULT 10,
  ADD COLUMN IF NOT EXISTS max_concurrent_runs INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS dry_run_mode BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS discovery_sources JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS feature_flags JSONB NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS approval_required_above_quality INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS approval_required_above_test INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS log_level TEXT NOT NULL DEFAULT 'info',
  ADD COLUMN IF NOT EXISTS notify_on_failure BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS notify_on_review_required BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS discovery_cooldown_minutes INTEGER NOT NULL DEFAULT 60,
  ADD COLUMN IF NOT EXISTS max_versions_per_function INTEGER NOT NULL DEFAULT 5;
