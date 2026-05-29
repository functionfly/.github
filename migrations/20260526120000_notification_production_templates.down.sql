DELETE FROM notification_templates WHERE type IN (
  'deployment.success',
  'deployment.failed',
  'failover.triggered',
  'failover.resolved',
  'provider.offline',
  'provider.online',
  'provider.degraded'
) AND channel = 'email';
