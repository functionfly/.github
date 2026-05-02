-- Create agent dead letter table for retry exhaustion handling
CREATE TABLE IF NOT EXISTS agent_dead_letter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL,
    function_uri VARCHAR(500),
    execution_id VARCHAR(100),
    session_id VARCHAR(100),
    input_payload JSONB DEFAULT '{}',
    output_payload JSONB,
    final_error VARCHAR(2000) NOT NULL,
    error_code VARCHAR(100) NOT NULL DEFAULT 'unknown',
    attempts INTEGER NOT NULL DEFAULT 0,
    first_attempt_at TIMESTAMP WITH TIME ZONE,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    first_attempt_error VARCHAR(2000),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    can_retry BOOLEAN NOT NULL DEFAULT true,
    retry_count INTEGER NOT NULL DEFAULT 0,
    retried_at TIMESTAMP WITH TIME ZONE,
    retry_success BOOLEAN DEFAULT false,
    retry_error VARCHAR(2000),
    alert_sent BOOLEAN DEFAULT false,
    alert_sent_at TIMESTAMP WITH TIME ZONE,
    alert_threshold INTEGER DEFAULT 3,
    metadata JSONB DEFAULT '{}',
    trace VARCHAR(5000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_agent_id ON agent_dead_letter(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_tenant_id ON agent_dead_letter(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_function_id ON agent_dead_letter(function_id);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_status ON agent_dead_letter(status);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_created_at ON agent_dead_letter(created_at);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_execution_id ON agent_dead_letter(execution_id);
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_session_id ON agent_dead_letter(session_id);

-- Index for finding pending entries needing retry
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_pending ON agent_dead_letter(status, can_retry) WHERE status = 'pending' AND can_retry = true;

-- Index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_created_status ON agent_dead_letter(created_at, status);

-- Composite index for agent-specific queries
CREATE INDEX IF NOT EXISTS idx_agent_dead_letter_agent_status ON agent_dead_letter(agent_id, status, created_at DESC);