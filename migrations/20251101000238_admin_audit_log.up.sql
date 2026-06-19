-- Admin audit log table for tracking all admin operations
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    changes JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_admin_audit_log_user_id ON admin_audit_log(user_id);
CREATE INDEX idx_admin_audit_log_resource ON admin_audit_log(resource_type, resource_id);
CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log(created_at);
CREATE INDEX idx_admin_audit_log_action ON admin_audit_log(action);

-- Composite index for common query patterns
CREATE INDEX idx_admin_audit_log_user_action ON admin_audit_log(user_id, action);

-- Comment for documentation
COMMENT ON TABLE admin_audit_log IS 'Admin dashboard audit log for tracking all administrative operations';
