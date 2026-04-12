-- Aggregation State Table
-- Tracks the progress of billing aggregation jobs for reliable resume/restart

CREATE TABLE IF NOT EXISTS aggregation_state (
    id VARCHAR(50) PRIMARY KEY, -- 'execution_aggregation', 'rollup_aggregation', etc.
    last_processed_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    last_processed_id UUID, -- Optional: last processed record ID for idempotency
    processed_count BIGINT DEFAULT 0, -- Total records processed
    metadata JSONB DEFAULT '{}', -- Additional state data
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for quick lookups
CREATE INDEX IF NOT EXISTS idx_aggregation_state_updated_at ON aggregation_state(updated_at DESC);

-- Insert default states
INSERT INTO aggregation_state (id, last_processed_timestamp, metadata)
VALUES
    ('execution_aggregation', NOW() - INTERVAL '1 hour', '{"description": "Last execution aggregated to usage_events"}'::jsonb),
    ('rollup_aggregation', NOW() - INTERVAL '24 hours', '{"description": "Last day rolled up to usage_rollups"}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- Comments
COMMENT ON TABLE aggregation_state IS 'Tracks aggregation job progress for reliable billing pipeline operation';
COMMENT ON COLUMN aggregation_state.id IS 'Unique identifier for the aggregation job type';
COMMENT ON COLUMN aggregation_state.last_processed_timestamp IS 'Timestamp of the last successfully processed record';
COMMENT ON COLUMN aggregation_state.processed_count IS 'Running count of total records processed by this job';
