-- Migration: Create trigger execution engine tables
-- This migration creates the tables needed for the statsfabric Trigger Execution Engine

-- 1. Trigger Event Queue - stores pending trigger events
CREATE TABLE IF NOT EXISTS trigger_event_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_id UUID NOT NULL,
    state_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    key VARCHAR(500),
    event_type VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    correlation_id VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    last_error_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_trigger_queue_status ON trigger_event_queue(status);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_tenant ON trigger_event_queue(tenant_id);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_trigger ON trigger_event_queue(trigger_id);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_state ON trigger_event_queue(state_id);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_correlation ON trigger_event_queue(correlation_id);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_next_attempt ON trigger_event_queue(next_attempt_at) WHERE status = 'pending' AND next_attempt_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trigger_queue_created ON trigger_event_queue(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_trigger_queue_composite ON trigger_event_queue(status, created_at) WHERE status = 'pending';

-- Composite index for polling queries
CREATE INDEX IF NOT EXISTS idx_trigger_queue_poll ON trigger_event_queue(status, next_attempt_at, created_at) 
    WHERE status = 'pending' AND (next_attempt_at IS NULL OR next_attempt_at <= NOW());

-- 2. Trigger Execution Logs - comprehensive observability
CREATE TABLE IF NOT EXISTS trigger_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_id UUID NOT NULL,
    state_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    queued_event_id UUID,
    event_type VARCHAR(50),
    key VARCHAR(500),
    target_type VARCHAR(50) NOT NULL DEFAULT 'function',
    target_url VARCHAR(1000),
    payload_size_bytes INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    http_status_code INTEGER,
    response_size INTEGER,
    duration_ms BIGINT NOT NULL,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    correlation_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for logs
CREATE INDEX IF NOT EXISTS idx_exec_log_trigger ON trigger_execution_logs(trigger_id);
CREATE INDEX IF NOT EXISTS idx_exec_log_tenant ON trigger_execution_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_exec_log_state ON trigger_execution_logs(state_id);
CREATE INDEX IF NOT EXISTS idx_exec_log_status ON trigger_execution_logs(status);
CREATE INDEX IF NOT EXISTS idx_exec_log_correlation ON trigger_execution_logs(correlation_id);
CREATE INDEX IF NOT EXISTS idx_exec_log_created ON trigger_execution_logs(created_at DESC);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_exec_log_tenant_status ON trigger_execution_logs(tenant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exec_log_trigger_status ON trigger_execution_logs(trigger_id, status, created_at DESC);

-- 3. Trigger Dead Letter Queue - failed events that exhausted retries
CREATE TABLE IF NOT EXISTS trigger_dead_letter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_event_id UUID NOT NULL,
    trigger_id UUID NOT NULL,
    state_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    event_type VARCHAR(50),
    key VARCHAR(500),
    payload JSONB NOT NULL DEFAULT '{}',
    final_error VARCHAR(2000) NOT NULL,
    failed_attempts INTEGER NOT NULL,
    correlation_id VARCHAR(255) NOT NULL,
    can_retry BOOLEAN NOT NULL DEFAULT true,
    retried_at TIMESTAMP WITH TIME ZONE,
    retried_success BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for DLQ
CREATE INDEX IF NOT EXISTS idx_dlq_trigger ON trigger_dead_letter(trigger_id);
CREATE INDEX IF NOT EXISTS idx_dlq_tenant ON trigger_dead_letter(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dlq_state ON trigger_dead_letter(state_id);
CREATE INDEX IF NOT EXISTS idx_dlq_original_event ON trigger_dead_letter(original_event_id);
CREATE INDEX IF NOT EXISTS idx_dlq_correlation ON trigger_dead_letter(correlation_id);
CREATE INDEX IF NOT EXISTS idx_dlq_can_retry ON trigger_dead_letter(can_retry, created_at) WHERE can_retry = true;
CREATE INDEX IF NOT EXISTS idx_dlq_retried ON trigger_dead_letter(retried_success, created_at) WHERE retried_success = false;

-- 4. Add helper functions

-- Function to get queue stats for a tenant
CREATE OR REPLACE FUNCTION get_tenant_trigger_queue_stats(tenant_uuid UUID)
RETURNS TABLE (
    pending_count BIGINT,
    processing_count BIGINT,
    failed_count BIGINT,
    dead_letter_count BIGINT,
    avg_wait_time_ms DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) FILTER (WHERE status = 'pending')::BIGINT as pending_count,
        COUNT(*) FILTER (WHERE status = 'processing')::BIGINT as processing_count,
        COUNT(*) FILTER (WHERE status = 'failed')::BIGINT as failed_count,
        COUNT(*) FILTER (WHERE status = 'dead_letter')::BIGINT as dead_letter_count,
        COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) * 1000), 0)::DOUBLE PRECISION as avg_wait_time_ms
    FROM trigger_event_queue
    WHERE tenant_id = tenant_uuid;
END;
$$ LANGUAGE plpgsql;

-- Function to get execution stats for a trigger
CREATE OR REPLACE FUNCTION get_trigger_execution_stats(trigger_uuid UUID, hours_back INTEGER DEFAULT 24)
RETURNS TABLE (
    total_executions BIGINT,
    successful_executions BIGINT,
    failed_executions BIGINT,
    rate_limited_count BIGINT,
    avg_duration_ms DOUBLE PRECISION,
    p95_duration_ms DOUBLE PRECISION,
    p99_duration_ms DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_executions,
        COUNT(*) FILTER (WHERE status = 'success')::BIGINT as successful_executions,
        COUNT(*) FILTER (WHERE status = 'error')::BIGINT as failed_executions,
        COUNT(*) FILTER (WHERE status = 'rate_limited')::BIGINT as rate_limited_count,
        COALESCE(AVG(duration_ms), 0)::DOUBLE PRECISION as avg_duration_ms,
        COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 0)::DOUBLE PRECISION as p95_duration_ms,
        COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms), 0)::DOUBLE PRECISION as p99_duration_ms
    FROM trigger_execution_logs
    WHERE trigger_id = trigger_uuid
    AND created_at > NOW() - (hours_back || ' hours')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Function to retry a dead letter event
CREATE OR REPLACE FUNCTION retry_dead_letter_event(dlq_id UUID)
RETURNS UUID AS $$
DECLARE
    new_event_id UUID;
    dlq_record trigger_dead_letter%ROWTYPE;
BEGIN
    -- Get the DLQ record
    SELECT * INTO dlq_record FROM trigger_dead_letter WHERE id = dlq_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Dead letter entry not found: %', dlq_id;
    END IF;
    
    IF NOT dlq_record.can_retry THEN
        RAISE EXCEPTION 'Dead letter entry is marked as non-retryable: %', dlq_id;
    END IF;
    
    -- Create new event
    INSERT INTO trigger_event_queue (
        trigger_id, state_id, tenant_id, key, event_type, 
        old_value, new_value, correlation_id, payload,
        status, max_attempts, correlation_id
    ) VALUES (
        dlq_record.trigger_id, dlq_record.state_id, dlq_record.tenant_id,
        dlq_record.key, dlq_record.event_type,
        dlq_record.payload->'old_value', dlq_record.payload->'new_value',
        gen_random_uuid()::TEXT, dlq_record.payload,
        'pending', 3, dlq_record.correlation_id
    ) RETURNING id INTO new_event_id;
    
    -- Update DLQ record
    UPDATE trigger_dead_letter 
    SET retried_at = NOW(), 
        retried_success = false  -- Will be updated when processed
    WHERE id = dlq_id;
    
    RETURN new_event_id;
END;
$$ LANGUAGE plpgsql;

-- Add retention policy function for cleanup
CREATE OR REPLACE FUNCTION cleanup_old_trigger_logs(retention_days INTEGER DEFAULT 30)
RETURNS TABLE (deleted_queue BIGINT, deleted_logs BIGINT, deleted_dlq BIGINT) AS $$
DECLARE
    queue_count BIGINT;
    logs_count BIGINT;
    dlq_count BIGINT;
BEGIN
    -- Clean up completed queue entries
    DELETE FROM trigger_event_queue 
    WHERE status IN ('completed', 'dead_letter') 
    AND completed_at < NOW() - (retention_days || ' days')::INTERVAL;
    GET DIAGNOSTICS queue_count = ROW_COUNT;
    
    -- Clean up old execution logs
    DELETE FROM trigger_execution_logs 
    WHERE created_at < NOW() - (retention_days || ' days')::INTERVAL;
    GET DIAGNOSTICS logs_count = ROW_COUNT;
    
    -- Clean up dead letter entries that have been retried successfully or are older than retention
    DELETE FROM trigger_dead_letter 
    WHERE (retried_success = true OR can_retry = false)
    AND created_at < NOW() - (retention_days || ' days')::INTERVAL;
    GET DIAGNOSTICS dlq_count = ROW_COUNT;
    
    RETURN QUERY SELECT queue_count, logs_count, dlq_count;
END;
$$ LANGUAGE plpgsql;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_trigger_event_queue_updated_at
    BEFORE UPDATE ON trigger_event_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
