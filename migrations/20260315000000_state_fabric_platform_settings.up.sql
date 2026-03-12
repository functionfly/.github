-- Platform-wide state fabric settings (single row, admin-configurable)
CREATE TABLE IF NOT EXISTS state_fabric_platform_settings (
    id         INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config     JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO state_fabric_platform_settings (id, config)
VALUES (1, '{"maxFabricsPerTenant": 10, "defaultSnapshotRetentionDays": 30, "allowPublicPipelines": false, "maintenanceMode": false}')
ON CONFLICT (id) DO NOTHING;
