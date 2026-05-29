# Alerting Setup

Connect Sentry errors to notification channels for on-call response.

## Sentry + Slack

1. Go to **Sentry** → **Settings** → **Integrations** → **Slack**
2. Click **Add Workspace**
3. Authorize in Slack
4. Configure notification routes:

### Alert Rules

**Critical Errors (immediate Slack):**
- New critical error detected
- Spike in error rate (>50% increase)
- Panic/recovery events

**Warning (Slack channel):**
- Performance degradation
- Slow transactions

**Setup in Sentry:**
1. **Projects** → **Alerts** → **Create Alert Rule**
2. Choose trigger: "An error occurs"
3. Set conditions: `level: critical` OR `error.handled: false` AND `count() > 5`
4. Action: Send to Slack channel `#alerts`

## Better Stack Uptime Alerts

Better Stack monitors uptime and pages on outage:

1. Go to [betteruptime.com](https://betteruptime.com) → **Monitors**
2. Add monitor for: `https://api.functionfly.com/health`
3. Configure alerting:
   - On-call schedule
   - Email + Slack integration
   - Escalation policy

## Fly.io Health Checks

Built-in checks (already configured in `fly.toml`):

```bash
# Check status
flyctl status --app functionfly-orchestrator

# View health checks
flyctl checks list --app functionfly-orchestrator
```

## Quick Commands

```bash
# View recent errors from logs
fly logs -a functionfly-orchestrator | grep -E "(ERROR|panic)"

# Check app health
curl https://api.functionfly.com/health
```

## Escalation Path

| Severity | Notification | Response Time |
|----------|-------------|---------------|
| Critical (outage) | Slack #incidents + PagerDuty | 15 min |
| High (degraded) | Slack #alerts | 1 hour |
| Medium (warnings) | Slack #monitoring | Business hours |
| Low (info) | Sentry dashboard only | Next sprint |
