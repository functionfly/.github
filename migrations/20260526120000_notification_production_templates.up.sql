-- Additional notification email templates for operational alerts

INSERT INTO notification_templates (type, channel, subject, body_html, body_text, variables, is_active) VALUES
('deployment.success', 'email', 'Deployment Successful: {{.AppName}}', '<p>Your deployment of <strong>{{.AppName}}</strong> was successful.</p>', 'Your deployment of {{.AppName}} was successful.', ARRAY['AppName'], true),
('deployment.failed', 'email', 'Deployment Failed: {{.AppName}}', '<p>Your deployment of <strong>{{.AppName}}</strong> failed.</p><p>{{.ErrorMessage}}</p>', 'Your deployment of {{.AppName}} failed. {{.ErrorMessage}}', ARRAY['AppName','ErrorMessage'], true),
('failover.triggered', 'email', 'Failover Triggered: {{.FunctionName}}', '<p>Failover was triggered for {{.FunctionName}}.</p>', 'Failover was triggered for {{.FunctionName}}.', ARRAY['FunctionName'], true),
('failover.resolved', 'email', 'Failover Resolved: {{.FunctionName}}', '<p>Normal operation resumed for {{.FunctionName}}.</p>', 'Normal operation resumed for {{.FunctionName}}.', ARRAY['FunctionName'], true),
('provider.offline', 'email', 'Provider Offline: {{.ProviderName}}', '<p>Provider {{.ProviderName}} is offline.</p>', 'Provider {{.ProviderName}} is offline.', ARRAY['ProviderName'], true),
('provider.online', 'email', 'Provider Online: {{.ProviderName}}', '<p>Provider {{.ProviderName}} is back online.</p>', 'Provider {{.ProviderName}} is back online.', ARRAY['ProviderName'], true),
('provider.degraded', 'email', 'Provider Degraded: {{.ProviderName}}', '<p>Provider {{.ProviderName}} is degraded.</p>', 'Provider {{.ProviderName}} is degraded.', ARRAY['ProviderName'], true)
ON CONFLICT (type, channel) DO UPDATE SET
  subject = EXCLUDED.subject,
  body_html = EXCLUDED.body_html,
  body_text = EXCLUDED.body_text,
  variables = EXCLUDED.variables,
  is_active = EXCLUDED.is_active,
  updated_at = CURRENT_TIMESTAMP;
