-- Migration: 000253_security_alert_rules.up.sql
-- Description: Create security_alert_rules table for configurable security alerts
-- Created: 2026-03-22

-- Create security_alert_rules table
CREATE TABLE IF NOT EXISTS security_alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    threshold INTEGER NOT NULL,
    window_seconds INTEGER DEFAULT 300,
    severity VARCHAR(20) DEFAULT 'medium',
    is_enabled BOOLEAN DEFAULT TRUE,
    notification_channels JSONB DEFAULT '[]',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Constraints
    CONSTRAINT security_alert_rules_alert_type CHECK (
        alert_type IN ('failed_login_threshold', 'rate_limit_exceeded', 'ip_blocked', 'suspicious_activity', 'session_anomaly')
    ),
    CONSTRAINT security_alert_rules_severity CHECK (
        severity IN ('low', 'medium', 'high', 'critical')
    ),
    CONSTRAINT security_alert_rules_threshold_positive CHECK (
        threshold > 0
    ),
    CONSTRAINT security_alert_rules_window_positive CHECK (
        window_seconds > 0
    )
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_security_alert_rules_alert_type ON security_alert_rules(alert_type);
CREATE INDEX IF NOT EXISTS idx_security_alert_rules_is_enabled ON security_alert_rules(is_enabled);
CREATE INDEX IF NOT EXISTS idx_security_alert_rules_severity ON security_alert_rules(severity);
CREATE INDEX IF NOT EXISTS idx_security_alert_rules_created_by ON security_alert_rules(created_by);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_security_alert_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-update updated_at
DROP TRIGGER IF EXISTS trigger_security_alert_rules_updated_at ON security_alert_rules;
CREATE TRIGGER trigger_security_alert_rules_updated_at
    BEFORE UPDATE ON security_alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_security_alert_rules_updated_at();

-- Insert default security alert rules
INSERT INTO security_alert_rules (name, alert_type, threshold, window_seconds, severity, is_enabled, notification_channels, created_by)
VALUES
    ('Failed Login Threshold', 'failed_login_threshold', 5, 900, 'high', true, '["email", "slack"]', NULL),
    ('Rate Limit Exceeded', 'rate_limit_exceeded', 100, 60, 'medium', true, '["slack"]', NULL),
    ('IP Blocked', 'ip_blocked', 1, 3600, 'critical', true, '["email", "slack", "pagerduty"]', NULL),
    ('Suspicious Activity', 'suspicious_activity', 10, 300, 'high', true, '["email", "slack"]', NULL),
    ('Session Anomaly', 'session_anomaly', 3, 600, 'critical', true, '["email", "slack", "pagerduty"]', NULL)
ON CONFLICT DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE security_alert_rules IS 'Configurable security alert rules for admin dashboard monitoring';
COMMENT ON COLUMN security_alert_rules.alert_type IS 'Type of alert: failed_login_threshold, rate_limit_exceeded, ip_blocked, suspicious_activity, session_anomaly';
COMMENT ON COLUMN security_alert_rules.threshold IS 'Number of events that trigger the alert';
COMMENT ON COLUMN security_alert_rules.window_seconds IS 'Time window in seconds for counting events';
COMMENT ON COLUMN security_alert_rules.severity IS 'Alert severity: low, medium, high, critical';
COMMENT ON COLUMN security_alert_rules.notification_channels IS 'JSON array of notification channels: email, slack, pagerduty';
