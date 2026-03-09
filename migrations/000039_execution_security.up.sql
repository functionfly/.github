-- Add execution security tables for rate limiting, abuse prevention, and input validation

-- User execution quotas table
CREATE TABLE IF NOT EXISTS user_execution_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,

    -- Quota limits
    daily_execution_limit INTEGER NOT NULL DEFAULT 1000,
    hourly_execution_limit INTEGER NOT NULL DEFAULT 100,
    minute_execution_limit INTEGER NOT NULL DEFAULT 10,

    -- Current usage counters
    daily_executions INTEGER NOT NULL DEFAULT 0,
    hourly_executions INTEGER NOT NULL DEFAULT 0,
    minute_executions INTEGER NOT NULL DEFAULT 0,

    -- Reset timestamps
    daily_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours'),
    hourly_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '1 hour'),
    minute_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '1 minute'),

    -- Abuse detection
    suspicious_activity_score INTEGER NOT NULL DEFAULT 0,
    last_suspicious_activity TIMESTAMP WITH TIME ZONE,
    is_throttled BOOLEAN NOT NULL DEFAULT FALSE,
    throttle_until TIMESTAMP WITH TIME ZONE,
    block_until TIMESTAMP WITH TIME ZONE,

    -- CAPTCHA requirements
    captcha_required BOOLEAN NOT NULL DEFAULT FALSE,
    last_captcha_completed TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure only one record per user/IP combination
    UNIQUE(user_id, ip_address)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_user_id ON user_execution_quotas(user_id);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_ip_address ON user_execution_quotas(ip_address);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_daily_reset ON user_execution_quotas(daily_reset_at);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_hourly_reset ON user_execution_quotas(hourly_reset_at);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_minute_reset ON user_execution_quotas(minute_reset_at);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_throttle_until ON user_execution_quotas(throttle_until);
CREATE INDEX IF NOT EXISTS idx_user_execution_quotas_block_until ON user_execution_quotas(block_until);

-- Partial unique index for anonymous users (user_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_execution_quotas_anonymous_ip ON user_execution_quotas(ip_address) WHERE user_id IS NULL;

-- Abuse patterns table
CREATE TABLE IF NOT EXISTS abuse_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),

    -- Pattern identification
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    function_id UUID REFERENCES registry_functions(id) ON DELETE SET NULL,

    -- Pattern data
    pattern_data JSONB,
    description TEXT NOT NULL,

    -- Action taken
    action_taken VARCHAR(50) NOT NULL DEFAULT 'none',
    action_data JSONB,

    -- Timing
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for abuse patterns
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_pattern_type ON abuse_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_user_id ON abuse_patterns(user_id);
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_ip_address ON abuse_patterns(ip_address);
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_function_id ON abuse_patterns(function_id);
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_detected_at ON abuse_patterns(detected_at);
CREATE INDEX IF NOT EXISTS idx_abuse_patterns_severity ON abuse_patterns(severity);

-- Execution security events table
CREATE TABLE IF NOT EXISTS execution_security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES registry_function_executions(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,

    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL,
    event_data JSONB,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for security events
CREATE INDEX IF NOT EXISTS idx_execution_security_events_execution_id ON execution_security_events(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_security_events_user_id ON execution_security_events(user_id);
CREATE INDEX IF NOT EXISTS idx_execution_security_events_ip_address ON execution_security_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_execution_security_events_event_type ON execution_security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_execution_security_events_created_at ON execution_security_events(created_at);

-- Function input schemas table
CREATE TABLE IF NOT EXISTS function_input_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    schema JSONB NOT NULL,
    is_strict BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(function_version_id)
);

-- Create indexes for input schemas
CREATE INDEX IF NOT EXISTS idx_function_input_schemas_function_version_id ON function_input_schemas(function_version_id);

-- Execution resource usage table
CREATE TABLE IF NOT EXISTS execution_resource_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES registry_function_executions(id) ON DELETE SET NULL,

    -- Resource limits (from function configuration)
    max_memory_mb INTEGER,
    max_cpu_time_ms INTEGER,

    -- Actual usage
    memory_used_mb DOUBLE PRECISION,
    cpu_time_used_ms INTEGER,
    wall_time_used_ms INTEGER,

    -- Termination reason
    terminated_by VARCHAR(50),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for resource usage
CREATE INDEX IF NOT EXISTS idx_execution_resource_usage_execution_id ON execution_resource_usage(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_resource_usage_created_at ON execution_resource_usage(created_at);

-- Add function to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
DROP TRIGGER IF EXISTS update_user_execution_quotas_updated_at ON user_execution_quotas;
CREATE TRIGGER update_user_execution_quotas_updated_at BEFORE UPDATE ON user_execution_quotas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_abuse_patterns_updated_at ON abuse_patterns;
CREATE TRIGGER update_abuse_patterns_updated_at BEFORE UPDATE ON abuse_patterns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_function_input_schemas_updated_at ON function_input_schemas;
CREATE TRIGGER update_function_input_schemas_updated_at BEFORE UPDATE ON function_input_schemas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();