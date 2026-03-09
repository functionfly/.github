-- Create platform maintenance table for maintenance windows
-- This is different from platform maintenance mode - this tracks scheduled maintenance windows
-- Only create when table does not exist (may already exist from 000183 or 000190)

DO $$
BEGIN
    IF to_regclass('public.platform_maintenance') IS NULL THEN
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
    END IF;
END $$;

-- Indexes only when this migration's schema exists (status/scheduled_start columns)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'platform_maintenance' AND column_name = 'status')
     AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'platform_maintenance' AND column_name = 'scheduled_start') THEN
    CREATE INDEX IF NOT EXISTS idx_platform_maintenance_status ON platform_maintenance(status);
    CREATE INDEX IF NOT EXISTS idx_platform_maintenance_scheduled_start ON platform_maintenance(scheduled_start);
    CREATE INDEX IF NOT EXISTS idx_platform_maintenance_scheduled_end ON platform_maintenance(scheduled_end);
    CREATE INDEX IF NOT EXISTS idx_platform_maintenance_created_at ON platform_maintenance(created_at DESC);
    CREATE INDEX IF NOT EXISTS idx_platform_maintenance_upcoming ON platform_maintenance(status, scheduled_end)
    WHERE status IN ('scheduled', 'in_progress');
  END IF;
END $$;
