-- Create local runtime registry tables
-- +migrate Up

-- Table for registered local runtime instances
CREATE TABLE IF NOT EXISTS local_runtime_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_id VARCHAR(255) UNIQUE NOT NULL,
    runtime_type VARCHAR(50) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    manifest_path TEXT NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    pid INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    uptime BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for efficient heartbeat queries
CREATE INDEX IF NOT EXISTS idx_local_runtime_instances_last_heartbeat ON local_runtime_instances(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_local_runtime_instances_status ON local_runtime_instances(status);
CREATE INDEX IF NOT EXISTS idx_local_runtime_instances_runtime_type ON local_runtime_instances(runtime_type);

-- Table for local runtime metrics snapshots
CREATE TABLE IF NOT EXISTS local_runtime_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    runtime_instance_id UUID NOT NULL REFERENCES local_runtime_instances(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Performance metrics as JSON
    memory_usage JSONB NOT NULL,
    cpu_usage DOUBLE PRECISION NOT NULL,
    active_connections INTEGER NOT NULL DEFAULT 0,
    request_throughput DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_requests BIGINT NOT NULL DEFAULT 0,
    error_rate DOUBLE PRECISION NOT NULL DEFAULT 0,

    -- Function execution metrics
    execution_count BIGINT NOT NULL DEFAULT 0,
    average_latency INTERVAL NOT NULL DEFAULT '0 seconds',
    error_count BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient metrics queries
CREATE INDEX IF NOT EXISTS idx_local_runtime_metrics_instance_timestamp ON local_runtime_metrics(runtime_instance_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_local_runtime_metrics_timestamp ON local_runtime_metrics(timestamp DESC);

-- Table for local runtime health checks
CREATE TABLE IF NOT EXISTS local_runtime_health (
    runtime_instance_id UUID PRIMARY KEY REFERENCES local_runtime_instances(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL,
    response_time INTERVAL NOT NULL DEFAULT '0 seconds',
    checks JSONB NOT NULL DEFAULT '{}',
    error TEXT
);

-- Index for health queries
CREATE INDEX IF NOT EXISTS idx_local_runtime_health_timestamp ON local_runtime_health(timestamp DESC);

-- Create a view for aggregated runtime metrics
DROP VIEW IF EXISTS local_runtime_aggregated_metrics;
CREATE VIEW local_runtime_aggregated_metrics AS
SELECT
    DATE_TRUNC('minute', m.timestamp) as time_bucket,
    COUNT(DISTINCT m.runtime_instance_id) as active_instances,
    AVG(m.cpu_usage) as avg_cpu_usage,
    SUM(m.active_connections) as total_connections,
    AVG(m.request_throughput) as avg_throughput,
    SUM(m.total_requests) as total_requests,
    AVG(m.error_rate) as avg_error_rate,
    SUM(m.execution_count) as total_executions,
    AVG(EXTRACT(epoch FROM m.average_latency)) as avg_latency_seconds,
    SUM(m.error_count) as total_errors
FROM local_runtime_metrics m
JOIN local_runtime_instances i ON m.runtime_instance_id = i.id
WHERE m.timestamp >= NOW() - INTERVAL '1 hour'
  AND i.last_heartbeat >= NOW() - INTERVAL '5 minutes'
GROUP BY DATE_TRUNC('minute', m.timestamp)
ORDER BY time_bucket DESC;