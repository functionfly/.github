-- Create maintenance windows table for scheduled maintenance
-- This is different from platform_maintenance (maintenance mode) - this tracks scheduled maintenance windows
DO $$
BEGIN
    IF to_regclass('public.maintenance_windows') IS NULL THEN
        CREATE TABLE maintenance_windows (
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

        CREATE INDEX idx_maintenance_windows_status ON maintenance_windows(status);
        CREATE INDEX idx_maintenance_windows_scheduled_start ON maintenance_windows(scheduled_start);
        CREATE INDEX idx_maintenance_windows_scheduled_end ON maintenance_windows(scheduled_end);
        CREATE INDEX idx_maintenance_windows_created_at ON maintenance_windows(created_at DESC);
        CREATE INDEX idx_maintenance_windows_upcoming ON maintenance_windows(status, scheduled_end)
        WHERE status IN ('scheduled', 'in_progress');
    END IF;
END $$;