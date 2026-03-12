ALTER TABLE factory_config
  DROP COLUMN IF EXISTS notification_webhook_url,
  DROP COLUMN IF EXISTS rate_limit_per_hour,
  DROP COLUMN IF EXISTS max_concurrent_runs,
  DROP COLUMN IF EXISTS dry_run_mode,
  DROP COLUMN IF EXISTS discovery_sources,
  DROP COLUMN IF EXISTS feature_flags,
  DROP COLUMN IF EXISTS approval_required_above_quality,
  DROP COLUMN IF EXISTS approval_required_above_test,
  DROP COLUMN IF EXISTS log_level,
  DROP COLUMN IF EXISTS notify_on_failure,
  DROP COLUMN IF EXISTS notify_on_review_required,
  DROP COLUMN IF EXISTS discovery_cooldown_minutes,
  DROP COLUMN IF EXISTS max_versions_per_function;
