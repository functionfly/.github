-- Drop local runtime registry tables
-- +migrate Down

-- Drop the view first
DROP VIEW IF EXISTS local_runtime_aggregated_metrics;

-- Drop tables (cascade will handle foreign keys)
DROP TABLE IF EXISTS local_runtime_health;
DROP TABLE IF EXISTS local_runtime_metrics;
DROP TABLE IF EXISTS local_runtime_instances;