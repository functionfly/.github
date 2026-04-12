-- Migration: Create trust audit log table
-- Description: Adds audit logging for all trust-related actions

CREATE TABLE IF NOT EXISTS trust_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    actor_id UUID NOT NULL,
    actor_type VARCHAR(20) NOT NULL,
    actor_partner_id UUID,
    function_id UUID REFERENCES registry_functions(id) ON DELETE CASCADE,
    previous_state JSONB,
    new_state JSONB,
    change_summary TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit logs
CREATE INDEX IF NOT EXISTS idx_trust_audit_action ON trust_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_trust_audit_entity ON trust_audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_trust_audit_actor ON trust_audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_trust_audit_function ON trust_audit_logs(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_audit_created_at ON trust_audit_logs(created_at DESC);

-- Composite index for filtering by function and action
CREATE INDEX IF NOT EXISTS idx_trust_audit_function_action ON trust_audit_logs(function_id, action, created_at DESC);

COMMENT ON TABLE trust_audit_logs IS 'Audit trail for all trust-related actions including revocations, attestations, and policy changes';
