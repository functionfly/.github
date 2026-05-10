-- Persist security events for audit and analysis
-- Replaces in-memory-only security state with durable storage

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id UUID,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    source VARCHAR(50) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_tenant ON security_events(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_function ON security_events(function_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events(severity, created_at DESC);

-- Retention: partition by month for efficient cleanup
-- Events older than 90 days are automatically purged by retention policy
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);

-- Add comment for documentation
COMMENT ON TABLE security_events IS 'Persistent security event log for audit, analysis, and compliance. Replaces in-memory security state.';
COMMENT ON COLUMN security_events.event_type IS 'Type of security event: violation, attack_pattern, rate_limit, input_validation, sandboxing, capability_check';
COMMENT ON COLUMN security_events.severity IS 'Event severity: low, medium, high, critical';
COMMENT ON COLUMN security_events.source IS 'Source of the event: runtime, dna, api, sandbox, enterprise_security';
