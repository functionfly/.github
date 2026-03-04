-- Create platform maintenance table for maintenance windows
-- This is different from platform maintenance mode - this tracks scheduled maintenance windows

CREATE TABLE platform_maintenance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    scheduled_start TIMESTAMPTZ NOT NULL,
    scheduled_end TIMESTAMPTZ NOT NULL,
    actual_start TIMESTAMPTZ,
    actual_end TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled')),
    affected_components TEXT[] DEFAULT '{}',
    affected_providers TEXT[] DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_platform_maintenance_status ON platform_maintenance(status);
CREATE INDEX idx_platform_maintenance_scheduled_start ON platform_maintenance(scheduled_start);
CREATE INDEX idx_platform_maintenance_scheduled_end ON platform_maintenance(scheduled_end);
CREATE INDEX idx_platform_maintenance_created_at ON platform_maintenance(created_at DESC);

-- Index for upcoming maintenance queries
CREATE INDEX idx_platform_maintenance_upcoming ON platform_maintenance(status, scheduled_end)
WHERE status IN ('scheduled', 'in_progress');