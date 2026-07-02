-- Deduplicate state_usage_metrics before creating unique constraints
-- For tenant_aggregate (state_id IS NULL): keep row with latest created_at per (tenant_id, metric_type, period_start)
DELETE FROM state_usage_metrics
WHERE state_id IS NULL
  AND id NOT IN (
      SELECT id FROM (
          SELECT id, ROW_NUMBER() OVER (
              PARTITION BY tenant_id, metric_type, period_start
              ORDER BY created_at DESC, id DESC
          ) AS rn
          FROM state_usage_metrics
          WHERE state_id IS NULL
      ) ranked
      WHERE rn = 1
  );

-- For per-state metrics (state_id IS NOT NULL): keep row with latest created_at per (state_id, metric_type, period_start)
DELETE FROM state_usage_metrics
WHERE state_id IS NOT NULL
  AND id NOT IN (
      SELECT id FROM (
          SELECT id, ROW_NUMBER() OVER (
              PARTITION BY state_id, metric_type, period_start
              ORDER BY created_at DESC, id DESC
          ) AS rn
          FROM state_usage_metrics
          WHERE state_id IS NOT NULL
      ) ranked
      WHERE rn = 1
  );

-- Partial unique index for tenant_aggregate metrics (no state_id)
-- One row per tenant per metric_type per hour period
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_usage_metrics_tenant_aggregate_unique
    ON state_usage_metrics(tenant_id, metric_type, period_start)
    WHERE state_id IS NULL;

-- Partial unique index for per-state metrics (has state_id)
-- One row per state per metric_type per hour period
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_usage_metrics_state_unique
    ON state_usage_metrics(state_id, metric_type, period_start)
    WHERE state_id IS NOT NULL;
