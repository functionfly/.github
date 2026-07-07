-- Slack Status Monitor Schema
-- Migration: 20260706120000_slack_status_monitor.up.sql

-- Slack configuration per tenant
CREATE TABLE IF NOT EXISTS slack_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID REFERENCES tenants(id) ON DELETE CASCADE,
    bot_token_enc   BYTEA,
    signing_secret  BYTEA,
    webhook_url     VARCHAR(1000),
    alert_channel   VARCHAR(100),
    report_channel  VARCHAR(100),
    channel_routing JSONB DEFAULT '{}',
    severity_config JSONB DEFAULT '{"critical": true, "high": true, "medium": true, "low": false}',
    quiet_hours     JSONB DEFAULT '{"enabled": false, "start": "22:00", "end": "08:00", "timezone": "UTC"}',
    enabled         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_slack_config_tenant ON slack_config(tenant_id);
CREATE INDEX IF NOT EXISTS idx_slack_config_enabled ON slack_config(enabled) WHERE enabled = TRUE;

-- Monitored components table
CREATE TABLE IF NOT EXISTS monitored_components (
    id            VARCHAR(100) PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    type          VARCHAR(50) NOT NULL,
    enabled       BOOLEAN DEFAULT TRUE,
    slack_channel VARCHAR(100),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Seed with the 20 components from getComponentSummaries()
INSERT INTO monitored_components (id, name, type) VALUES
    ('api', 'API', 'api'),
    ('database', 'Database', 'database'),
    ('cache', 'Cache', 'cache'),
    ('ai-service', 'AI Service', 'ai'),
    ('embeddings', 'Embeddings', 'ai'),
    ('state-fabric', 'State Fabric', 'storage'),
    ('microvm', 'MicroVM Runtime', 'runtime'),
    ('queue', 'Queue Worker', 'worker'),
    ('function-backup', 'Function Backup', 'backup'),
    ('email', 'Email Delivery', 'email'),
    ('billing', 'Billing', 'billing'),
    ('storage', 'Object Storage', 'storage'),
    ('cdn', 'CDN', 'cdn'),
    ('pgbouncer', 'Connection Pool', 'infrastructure'),
    ('recommendations', 'Recommendations', 'ai'),
    ('verification', 'Verification Pipeline', 'security'),
    ('trust-api', 'Trust API', 'security'),
    ('support', 'Support System', 'service'),
    ('registry', 'Function Registry', 'service'),
    ('health-monitor', 'Health Monitor', 'monitoring')
ON CONFLICT (id) DO NOTHING;

-- Slack alert log for audit trail
CREATE TABLE IF NOT EXISTS slack_alert_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    component_id    VARCHAR(100) NOT NULL,
    old_status      VARCHAR(50),
    new_status      VARCHAR(50) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    channel         VARCHAR(100) NOT NULL,
    message_ts      VARCHAR(100),
    delivered       BOOLEAN DEFAULT FALSE,
    error           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_slack_alert_log_component ON slack_alert_log(component_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_slack_alert_log_created ON slack_alert_log(created_at DESC);

-- Add slack_enabled column to notification_preferences if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'notification_preferences' AND column_name = 'slack_enabled') THEN
        ALTER TABLE notification_preferences ADD COLUMN slack_enabled BOOLEAN DEFAULT FALSE;
    END IF;
END $$;
