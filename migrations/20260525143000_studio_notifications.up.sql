-- Studio in-app notifications for workflow, billing, plugins, and system events

CREATE TABLE IF NOT EXISTS studio_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID,
    environment TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'info',
    category TEXT NOT NULL DEFAULT 'system',
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    action_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_studio_notifications_tenant_user_env
    ON studio_notifications(tenant_id, user_id, environment, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_studio_notifications_unread
    ON studio_notifications(tenant_id, user_id, environment, read, created_at DESC);

COMMENT ON TABLE studio_notifications IS 'In-app Studio notifications surfaced in the dashboard notification center';
