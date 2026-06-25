-- Create rotation_schedules table for managing automated and scheduled secret rotation
-- Supports immediate, scheduled, and automatic rotation with grace periods

CREATE TABLE IF NOT EXISTS rotation_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    rotation_type VARCHAR(20) NOT NULL, -- 'scheduled', 'automatic'
    scheduled_at TIMESTAMP WITH TIME ZONE,
    cron_expression VARCHAR(100),
    interval_days INT,
    grace_period_hours INT NOT NULL DEFAULT 24,
    auto_rotate_interval INT NOT NULL DEFAULT 90,
    last_rotated_at TIMESTAMP WITH TIME ZONE,
    next_rotation_at TIMESTAMP WITH TIME ZONE,
    notify_stakeholders BOOLEAN NOT NULL DEFAULT true,
    require_approval BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'pending', 'paused', 'cancelled'
    last_error TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_rotation_secret FOREIGN KEY (secret_id) REFERENCES secrets_vault(id) ON DELETE CASCADE
);

CREATE INDEX idx_rotation_schedules_secret_id ON rotation_schedules(secret_id);
CREATE INDEX idx_rotation_schedules_tenant_id ON rotation_schedules(tenant_id);
CREATE INDEX idx_rotation_schedules_next_rotation ON rotation_schedules(next_rotation_at) WHERE enabled = true AND status = 'active';
CREATE INDEX idx_rotation_schedules_status ON rotation_schedules(status);

COMMENT ON TABLE rotation_schedules IS 'Manages automated and scheduled secret rotation with grace period support';
COMMENT ON COLUMN rotation_schedules.rotation_type IS 'Type of rotation: scheduled (one-time) or automatic (recurring)';
COMMENT ON COLUMN rotation_schedules.grace_period_hours IS 'How long the old secret remains valid after rotation';
COMMENT ON COLUMN rotation_schedules.auto_rotate_interval IS 'Days between automatic rotation cycles';
