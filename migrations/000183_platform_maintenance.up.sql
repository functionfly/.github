-- Platform Maintenance Mode Tables
-- Provides platform-wide maintenance mode functionality

-- Main maintenance mode configuration table
CREATE TABLE IF NOT EXISTS platform_maintenance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enabled BOOLEAN NOT NULL DEFAULT false,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    message TEXT,
    page_template VARCHAR(100) DEFAULT 'default',
    retry_after_seconds INTEGER DEFAULT 3600,

    -- Rollout control (for gradual enablement)
    rollout_percentage INTEGER DEFAULT 100 CHECK (rollout_percentage BETWEEN 0 AND 100),
    rollout_seed VARCHAR(50),

    -- Scheduling
    scheduled_start TIMESTAMPTZ,
    scheduled_end TIMESTAMPTZ,
    is_scheduled BOOLEAN DEFAULT false,

    -- Recurring maintenance window (optional)
    recurrence_rule VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',

    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Customizable maintenance page templates
CREATE TABLE IF NOT EXISTS maintenance_page_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255),
    message_html TEXT,
    logo_url VARCHAR(500),
    background_color VARCHAR(20) DEFAULT '#1a1a2e',
    text_color VARCHAR(20) DEFAULT '#ffffff',
    accent_color VARCHAR(20) DEFAULT '#4ecdc4',
    show_contact_info BOOLEAN DEFAULT true,
    contact_email VARCHAR(255),
    show_social_links BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log for maintenance mode changes
CREATE TABLE IF NOT EXISTS maintenance_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    maintenance_id UUID REFERENCES platform_maintenance(id),
    action VARCHAR(50) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_platform_maintenance_enabled ON platform_maintenance(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_platform_maintenance_scheduled ON platform_maintenance(scheduled_start) WHERE is_scheduled = true;
CREATE INDEX IF NOT EXISTS idx_maintenance_audit_log_maintenance_id ON maintenance_audit_log(maintenance_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_audit_log_changed_at ON maintenance_audit_log(changed_at DESC);

-- Insert default maintenance template
INSERT INTO maintenance_page_templates (
    name,
    title,
    message_html,
    background_color,
    text_color,
    accent_color,
    show_contact_info,
    show_social_links,
    is_default
) VALUES (
    'default',
    'We''ll be back soon!',
    '<p>We''re performing scheduled maintenance. We''ll be back shortly.</p>',
    '#1a1a2e',
    '#ffffff',
    '#4ecdc4',
    true,
    true,
    true
) ON CONFLICT (name) DO NOTHING;

-- Insert default maintenance record (disabled by default)
INSERT INTO platform_maintenance (
    name,
    description,
    message,
    enabled,
    rollout_percentage
)
SELECT
    'Default Maintenance',
    'Default maintenance configuration',
    'We''re performing scheduled maintenance. We''ll be back shortly.',
    false,
    100
WHERE NOT EXISTS (SELECT 1 FROM platform_maintenance LIMIT 1);
