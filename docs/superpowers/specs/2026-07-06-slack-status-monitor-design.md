# Slack Status Monitor — Design Spec

**Date:** 2026-07-06
**Status:** Draft
**Scope:** Production-grade Slack integration for platform status monitoring, incident alerting, and on-demand status queries.

---

## Overview

Add a Slack integration to FunctionFly that monitors all platform services (20 components) and sends real-time alerts to Slack when service status changes, incidents occur, or maintenance is scheduled. Includes slash commands for on-demand status queries and scheduled uptime/latency reports.

The integration extends the existing notification system (`internal/notification/`) as a new channel, following the established `Channel` interface pattern used by email, in-app, and webhook channels.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Status Monitor (goroutine)                  │
│  Reads from system_health_checks + circuit_state tables     │
│  Runs every 30s via ticker                                  │
│  Redis dedup: SET component:last_state:{id} with TTL        │
│  Only dispatches on state transitions                       │
└─────────────────┬───────────────────────────────────────────┘
                  │ emits StatusChangeEvent
┌─────────────────▼───────────────────────────────────────────┐
│              Notification Dispatcher (existing)              │
│  Routes to channels based on admin-configured preferences   │
└───┬──────────┬──────────────┬──────────────┬────────────────┘
    │          │              │              │
┌───▼───┐ ┌───▼────┐  ┌─────▼─────┐  ┌────▼──────┐
│ Email │ │ In-App │  │  Webhook  │  │  Slack    │
│(exist)│ │(exist) │  │ (exist)   │  │  (new)    │
└───────┘ └────────┘  └───────────┘  └───────────┘
                                           │
                                    ┌──────▼──────┐
                                    │ Slack API   │
                                    │ Incoming    │
                                    │ Webhook     │
                                    └─────────────┘

┌─────────────────────────────────────────────────────────────┐
│                Slack App (new subsystem)                     │
│  /status → current health of all 20 services                │
│  /uptime [service] [period] → uptime % and latency          │
│  /incidents → active incidents list                         │
│  Interactive buttons: Acknowledge, Update, Resolve          │
│  Request verification via SLACK_SIGNING_SECRET              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│           Report Scheduler (new subsystem)                   │
│  Daily digest: 9 AM, uptime % per service, overnight events │
│  Weekly report: Monday 9 AM, 7-day trends, latency p95/p99 │
│  Uses existing cron pattern from internal/services/         │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **DB-based monitoring** — Status Monitor reads directly from `system_health_checks` and `circuit_state` tables instead of polling the HTTP API. Avoids internal HTTP roundtrip and survives orchestrator restarts.

2. **Incoming Webhooks for alerts** — Uses Slack's simple Incoming Webhook URLs for sending alerts (matches existing consciousness Slack pattern). Bot token only needed for slash commands.

3. **Redis deduplication** — Caches last known state per component in Redis. Only dispatches notifications on state transitions, preventing alert storms during flapping.

4. **Severity mapping** — Component status maps to notification severity levels that control Slack message formatting and mention behavior.

5. **Admin-only config** — All Slack configuration lives in the admin dashboard. User dashboard gets a Slack toggle that inherits from admin config.

---

## Components

### 1. Status Monitor (`internal/services/status_monitor.go`)

**Purpose:** Detects service state transitions and emits notifications.

**Lifecycle:** Started as a goroutine by `cmd/orchestrator-api/main.go` alongside existing services (usage forecaster, dunning scheduler, etc.)

**Data source:** Reads from existing database tables:
- `system_health_checks` — latest health check results per component
- `circuit_state` — circuit breaker state per backend
- `health_checks` — probe results

**State tracking:** Redis keys `status:monitor:last_state:{component_id}` with 24h TTL. On each tick:
1. Query current state of all monitored components
2. For each component, compare with Redis-cached last state
3. If state changed: create notification via `Queue.Enqueue()`, update Redis key
4. If state unchanged: skip

**Component discovery:** Reads from `monitored_components` table (seeded with the 20 components from `getComponentSummaries()`).

**Configuration:**

| Variable | Purpose | Default |
|----------|---------|---------|
| `STATUS_MONITOR_ENABLED` | Enable/disable the monitor | `true` |
| `STATUS_MONITOR_INTERVAL_SEC` | Poll interval in seconds | `30` |
| `SLACK_WEBHOOK_URL` | Default Slack webhook for alerts | (required) |

**Notification types used:**
- `TypeProviderOffline` — component went down
- `TypeProviderOnline` — component recovered
- `TypeProviderDegraded` — component degraded
- `TypeSystemMaintenance` — maintenance started/ended

---

### 2. Slack Channel (`internal/notification/slack_channel.go`)

**Purpose:** Implements the `Channel` interface for Slack delivery.

**Interface:**
```go
type Channel interface {
    Name() string
    Send(ctx context.Context, notification *Notification, user *storage.User) error
    IsConfigured() bool
}
```

**Implementation:**
- `Name()` returns `"slack"`
- `IsConfigured()` returns whether `SLACK_WEBHOOK_URL` is set
- `Send()` builds Block Kit payload and POSTs to the webhook URL

**Block Kit message structure:**
```json
{
  "blocks": [
    {
      "type": "header",
      "text": { "type": "plain_text", "text": "🔴 API — Major Outage" }
    },
    {
      "type": "section",
      "fields": [
        { "type": "mrkdwn", "text": "*Status:*\nMajor Outage" },
        { "type": "mrkdwn", "text": "*Uptime (24h):*\n99.2%" },
        { "type": "mrkdwn", "text": "*Response Time:*\n450ms" },
        { "type": "mrkdwn", "text": "*Detected:*\n2 minutes ago" }
      ]
    },
    {
      "type": "actions",
      "elements": [
        { "type": "button", "text": { "type": "plain_text", "text": "View Status" }, "url": "https://status.functionfly.com" },
        { "type": "button", "text": { "type": "plain_text", "text": "Acknowledge" }, "action_id": "ack_incident", "value": "incident_id" }
      ]
    }
  ]
}
```

**Severity-to-format mapping:**

| Severity | Header Emoji | Color | Mention |
|----------|-------------|-------|---------|
| Critical | 🔴 | `#FF0000` | `@channel` |
| High | 🟠 | `#FF6600` | `@here` |
| Medium | 🟡 | `#FFCC00` | none |
| Info (recovery) | 🟢 | `#00CC00` | none |
| Maintenance | 🔵 | `#0066FF` | none |

**Retry:** 3 attempts with exponential backoff (1s, 2s, 4s). Matches `WebhookChannel` pattern.

**Rate limiting:** Mutex + 1s sleep to respect Slack's 1 req/sec per webhook URL limit.

**Metrics:**
- `slack_send_total` (counter)
- `slack_send_latency_ms` (histogram)
- `slack_send_errors_total` (counter, labeled by error type)

---

### 3. Slack App Handler (`internal/api/handlers/slack/`)

**Purpose:** Handles slash commands and interactive components from Slack.

**Endpoints:**

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/slack/commands` | POST | Slash command handler |
| `/api/v1/slack/interactions` | POST | Button/action handler |

**Request verification:** Validates `X-Slack-Signature` header using HMAC-SHA256 with `SLACK_SIGNING_SECRET`. Matches the pattern in `internal/agent/runtime/router.go` lines 786-903.

**Slash commands:**

| Command | Response | Example |
|---------|----------|---------|
| `/status` | All 20 services with status emoji, uptime % | `🟢 API — 99.9% \| 🔴 Database — Down \| ...` |
| `/status api` | Detailed status for one service | Uptime, latency, circuit state, recent incidents |
| `/uptime` | 24h uptime summary for all services | Table with uptime % per service |
| `/uptime api 7d` | 7-day uptime for one service | Uptime chart (text-based) |
| `/incidents` | Active incidents list | Title, severity, status, affected services |
| `/incidents abc123` | Incident detail with timeline | Full incident timeline |

**Response strategy:** Slack requires <3s initial response. For `/status` and `/uptime` (which query the DB), use `response_url` for deferred response:
1. Return `200 OK` with `{"response_type": "ephemeral", "text": "Fetching status..."}` immediately
2. Query DB in background goroutine
3. POST full response to `response_url` within 15 minutes

**Interactive components:**
- `ack_incident` button — Sets incident status to "investigating" and adds update
- `resolve_incident` button — Sets incident status to "resolved"
- `view_status` button — Opens status page URL

---

### 4. Report Scheduler (`internal/services/status_reporter.go`)

**Purpose:** Sends scheduled uptime and latency reports to Slack.

**Reports:**

| Report | Schedule | Content |
|--------|----------|---------|
| Daily digest | 9 AM UTC (configurable) | Overnight incidents, current status of all 20 services, uptime % change vs previous day |
| Weekly report | Monday 9 AM UTC | 7-day uptime trends, latency p95/p99 per service, incident summary, notable changes |

**Data source:** Queries status handler logic internally (reuses `HandleGetPlatformStatus`, `HandleGetUptimeMetrics` handler functions).

**Block Kit format:**
- Header with date range
- Status grid: 4 columns of service status cards
- Incident summary section
- Uptime trend chart (using Slack's `chart_image` or text-based bar chart)
- Latency table with p50/p95/p99

**Configuration:**

| Variable | Purpose | Default |
|----------|---------|---------|
| `SLACK_REPORT_ENABLED` | Enable scheduled reports | `true` |
| `SLACK_REPORT_DAILY_CRON` | Daily report cron expression | `0 9 * * *` |
| `SLACK_REPORT_WEEKLY_CRON` | Weekly report cron expression | `0 9 * * 1` |
| `SLACK_REPORT_CHANNEL` | Slack channel for reports | (required if enabled) |

---

### 5. Admin Dashboard Configuration

**Route:** `/integrations/slack` in admin dashboard (`web/admin-dashboard/`)

**UI components:**
- Connection status indicator (connected/disconnected)
- Bot Token input (masked, encrypted at rest)
- Signing Secret input (masked, encrypted at rest)
- Default Alert Channel ID input
- Report Channel ID input
- Per-service channel routing table (20 rows, each with optional channel override)
- Severity threshold config (which severities trigger alerts)
- Quiet hours config (start/end time, timezone)
- Enable/disable master toggle
- Test button (sends test message to configured channel)

**User notification preferences:** The existing `NotificationsSettingsTab` in the user dashboard (`/u/:username/settings#notifications`) will gain a "Slack" toggle group. When the admin has enabled Slack for the tenant, users can opt in/out of Slack notifications per category (deployment, billing, security, etc.). The toggle is hidden if the admin has not configured Slack. The preferences are stored in the existing `notification_preferences` table with a new `slack_enabled` column (reusing the pattern from `consciousness_preferences`).

**API endpoints:**

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/admin/slack/config` | GET | Get Slack config |
| `/api/v1/admin/slack/config` | PUT | Update Slack config |
| `/api/v1/admin/slack/test` | POST | Send test message |
| `/api/v1/admin/slack/channels` | GET | List available channels (via Slack API) |

---

### 6. Database Schema

**New table: `slack_config`**
```sql
CREATE TABLE slack_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID REFERENCES tenants(id),
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

CREATE INDEX idx_slack_config_tenant ON slack_config(tenant_id);
CREATE INDEX idx_slack_config_enabled ON slack_config(enabled) WHERE enabled = TRUE;
```

**New table: `monitored_components`**
```sql
CREATE TABLE monitored_components (
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
    ('health-monitor', 'Health Monitor', 'monitoring');
```

**New table: `slack_alert_log`**
```sql
CREATE TABLE slack_alert_log (
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

CREATE INDEX idx_slack_alert_log_component ON slack_alert_log(component_id, created_at DESC);
CREATE INDEX idx_slack_alert_log_created ON slack_alert_log(created_at DESC);
```

---

## File Structure

```
internal/
├── notification/
│   ├── slack_channel.go          # NEW: Channel implementation
│   ├── slack_channel_test.go     # NEW: Tests
│   ├── slack_templates.go        # NEW: Block Kit message builders
│   ├── types.go                  # MODIFY: Add ChannelSlack constant
│   ├── dispatcher.go             # MODIFY: Register slack channel
│   └── ...
├── services/
│   ├── status_monitor.go         # NEW: Status monitoring goroutine
│   ├── status_monitor_test.go    # NEW: Tests
│   ├── status_reporter.go        # NEW: Scheduled reports
│   └── status_reporter_test.go   # NEW: Tests
├── api/
│   ├── handlers/
│   │   └── slack/
│   │       ├── handler.go        # NEW: Slash command + interaction handler
│   │       ├── handler_test.go   # NEW: Tests
│   │       ├── verification.go   # NEW: Slack request verification
│   │       └── commands.go       # NEW: Command implementations
│   ├── routes.go                 # MODIFY: Register /api/v1/slack/* routes
│   └── middleware/
│       └── slack_verify.go       # NEW: Slack signing secret middleware
├── storage/
│   └── sql/
│       ├── slack_repository.go   # NEW: Slack config + alert log queries
│       └── slack_repository_test.go
cmd/
└── orchestrator-api/
    └── main.go                   # MODIFY: Start status monitor + reporter

web/admin-dashboard/
├── src/
│   ├── pages/
│   │   └── IntegrationsSlackPage.tsx  # NEW: Slack config page
│   ├── api/
│   │   └── slack.ts                   # NEW: Slack config API client
│   └── components/
│       └── slack/
│           ├── SlackConfigForm.tsx    # NEW: Config form
│           ├── ChannelRoutingTable.tsx # NEW: Per-service routing
│           └── TestMessageButton.tsx  # NEW: Test button

migrations/
└── 20260706120000_slack_status_monitor.up.sql   # NEW
└── 20260706120000_slack_status_monitor.down.sql # NEW
```

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `SLACK_WEBHOOK_URL` | Default Incoming Webhook URL for alerts | (unset) |
| `SLACK_BOT_TOKEN` | Bot token for slash commands (xoxb-...) | (unset) |
| `SLACK_SIGNING_SECRET` | Request verification for slash commands | (unset) |
| `STATUS_MONITOR_ENABLED` | Enable status monitoring goroutine | `true` |
| `STATUS_MONITOR_INTERVAL_SEC` | Monitor poll interval | `30` |
| `SLACK_REPORT_ENABLED` | Enable scheduled reports | `true` |
| `SLACK_REPORT_DAILY_CRON` | Daily report schedule | `0 9 * * *` |
| `SLACK_REPORT_WEEKLY_CRON` | Weekly report schedule | `0 9 * * 1` |
| `SLACK_REPORT_CHANNEL` | Channel ID for reports | (unset) |

---

## Error Handling

- **Slack API errors:** Retry 3x with exponential backoff. Log failures to `slack_alert_log` table. After 3 failures, mark as undelivered and continue.
- **Rate limiting:** Slack returns `429` with `Retry-After` header. Respect it and retry.
- **Webhook URL rotation:** If a webhook URL is revoked, Slack returns `404` or `410`. Mark config as `enabled = FALSE` and log an admin alert.
- **Monitor crashes:** The goroutine is wrapped in a `recover()` with automatic restart (max 5 restarts per minute). Logs to structured logger.
- **Redis unavailable:** Fall back to always-dispatch mode (no dedup). Log warning.
- **Quiet hours:** During quiet hours, only Critical severity alerts are sent. Others are held in the notification queue (existing `Queue.Enqueue()` already supports delayed delivery) and flushed when quiet hours end. The quiet hours check happens in `SlackChannel.Send()` — if current time is within quiet hours and severity < Critical, the notification is re-enqueued with a `deliver_at` timestamp set to quiet hours end.

---

## Testing Strategy

- **Unit tests:** `slack_channel_test.go` — mock HTTP server for Slack webhook, test Block Kit payload generation, retry logic, rate limiting
- **Unit tests:** `status_monitor_test.go` — mock DB queries, test state transition detection, dedup logic
- **Integration tests:** Slash command handler with mock Slack requests, request verification
- **E2E:** Manual test with a real Slack workspace using `/api/v1/admin/slack/test` endpoint

---

## Migration Path

1. Deploy DB migration (slack_config, monitored_components, slack_alert_log tables)
2. Seed monitored_components with 20 components
3. Deploy Go code (slack channel, monitor, reporter, slash command handler)
4. Admin configures Slack in dashboard (webhook URL, signing secret, channels)
5. Enable monitor via `STATUS_MONITOR_ENABLED=true`
6. Test with `/api/v1/admin/slack/test` endpoint
7. Slash commands available immediately after bot token is configured

---

## Security Considerations

- Bot token and signing secret encrypted at rest using existing encryption infrastructure
- Slack request verification (HMAC-SHA256) on all incoming Slack requests
- Webhook URLs are HTTPS-only (enforced by Slack)
- SSRF protection: webhook URLs validated against Slack domain allowlist (`hooks.slack.com`, `*.slack.com`)
- Admin-only access: all Slack config endpoints require admin auth middleware
- Alert log: all sent alerts logged for audit trail
- No secrets in logs: webhook URLs and tokens masked in application logs
