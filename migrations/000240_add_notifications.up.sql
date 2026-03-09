-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    data JSONB DEFAULT '{}',
    channels TEXT[] DEFAULT '{}',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    read_at TIMESTAMP,
    sent_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, status) WHERE status != 'read';

-- Notification preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    frequency VARCHAR(20) DEFAULT 'immediate',
    quiet_hours_start VARCHAR(5),
    quiet_hours_end VARCHAR(5),
    timezone VARCHAR(50) DEFAULT 'UTC',
    webhook_url VARCHAR(500),
    webhook_secret VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, channel, category)
);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Notification templates table
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(255),
    body_html TEXT,
    body_text TEXT,
    variables TEXT[],
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(type, channel)
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON notification_templates(type);
CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON notification_templates(is_active);

-- Notification analytics table
CREATE TABLE IF NOT EXISTS notification_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_analytics_notification_id ON notification_analytics(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_analytics_status ON notification_analytics(status);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_notifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;
CREATE TRIGGER trigger_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_notifications_updated_at();

DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
CREATE TRIGGER trigger_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_notifications_updated_at();

DROP TRIGGER IF EXISTS trigger_notification_templates_updated_at ON notification_templates;
CREATE TRIGGER trigger_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_notifications_updated_at();

-- Seed default notification templates
INSERT INTO notification_templates (type, channel, subject, body_html, body_text, variables) VALUES
    ('deployment.success', 'email', 'Deployment Successful - {{app_name}}',
     '<p>Your deployment of <strong>{{app_name}}</strong> was successful.</p><p>Deployed at: {{deployed_at}}</p>',
     'Your deployment of {{app_name}} was successful. Deployed at: {{deployed_at}}',
     ARRAY['app_name', 'deployed_at']),

    ('deployment.failed', 'email', 'Deployment Failed - {{app_name}}',
     '<p>Your deployment of <strong>{{app_name}}</strong> failed.</p><p>Error: {{error_message}}</p><p><a href="{{logs_url}}">View Logs</a></p>',
     'Your deployment of {{app_name}} failed. Error: {{error_message}}. View logs: {{logs_url}}',
     ARRAY['app_name', 'error_message', 'logs_url']),

    ('billing.invoice_generated', 'email', 'New Invoice Available',
     '<p>A new invoice for {{amount}} is available.</p><p>Due date: {{due_date}}</p>',
     'A new invoice for {{amount}} is available. Due date: {{due_date}}',
     ARRAY['amount', 'due_date']),

    ('security.password_changed', 'email', 'Password Changed',
     '<p>Your password was changed on {{changed_at}}.</p><p>If you did not make this change, please contact support immediately.</p>',
     'Your password was changed on {{changed_at}}. If you did not make this change, please contact support immediately.',
     ARRAY['changed_at']),

    ('team.invitation', 'email', 'You\''ve been invited to join {{team_name}}',
     '<p>You have been invited to join <strong>{{team_name}}</strong> on FunctionFly.</p><p><a href="{{invitation_url}}">Accept Invitation</a></p>',
     'You have been invited to join {{team_name}} on FunctionFly. Accept invitation: {{invitation_url}}',
     ARRAY['team_name', 'invitation_url']),

    ('system.maintenance', 'email', 'Scheduled Maintenance Notice',
     '<p>We will be performing scheduled maintenance on {{maintenance_date}} from {{start_time}} to {{end_time}} UTC.</p><p>Services may be unavailable during this time.</p>',
     'Scheduled maintenance on {{maintenance_date}} from {{start_time}} to {{end_time}} UTC. Services may be unavailable.',
     ARRAY['maintenance_date', 'start_time', 'end_time'])
ON CONFLICT (type, channel) DO NOTHING;

-- Create in-app templates
INSERT INTO notification_templates (type, channel, subject, body_html, body_text, variables) VALUES
    ('deployment.success', 'in_app', NULL,
     '<span class="text-green-600">✓</span> Deployment of <strong>{{app_name}}</strong> successful',
     'Deployment of {{app_name}} successful',
     ARRAY['app_name']),

    ('deployment.failed', 'in_app', NULL,
     '<span class="text-red-600">✗</span> Deployment of <strong>{{app_name}}</strong> failed',
     'Deployment of {{app_name}} failed',
     ARRAY['app_name']),

    ('billing.invoice_generated', 'in_app', NULL,
     'New invoice for {{amount}} available',
     'New invoice for {{amount}} available',
     ARRAY['amount']),

    ('security.password_changed', 'in_app', NULL,
     'Password changed successfully',
     'Password changed successfully',
     ARRAY[]::text[]),

    ('team.invitation', 'in_app', NULL,
     'Invitation to join {{team_name}}',
     'Invitation to join {{team_name}}',
     ARRAY['team_name']),

    ('system.maintenance', 'in_app', NULL,
     'Scheduled maintenance: {{maintenance_date}}',
     'Scheduled maintenance: {{maintenance_date}}',
     ARRAY['maintenance_date'])
ON CONFLICT (type, channel) DO NOTHING;
