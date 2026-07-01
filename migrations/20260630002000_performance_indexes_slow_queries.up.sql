-- Composite index for state_usage_metrics aggregation query
-- Covers: WHERE period_start >= ? AND period_end <= ? AND state_id IS NOT NULL GROUP BY tenant_id, metric_type
CREATE INDEX IF NOT EXISTS idx_state_usage_metrics_agg
    ON state_usage_metrics(period_start, period_end, tenant_id, metric_type)
    WHERE state_id IS NOT NULL;

-- Partial index for trust_attestations expiration sweep
-- Covers: WHERE status = 'valid' AND valid_until IS NOT NULL AND valid_until < ?
CREATE INDEX IF NOT EXISTS idx_trust_attestations_expire_sweep
    ON trust_attestations(valid_until)
    WHERE status = 'valid' AND valid_until IS NOT NULL;
