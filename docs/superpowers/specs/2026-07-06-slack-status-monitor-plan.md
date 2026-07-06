# Slack Status Monitor — Implementation Plan

**Spec:** `docs/superpowers/specs/2026-07-06-slack-status-monitor-design.md`
**Date:** 2026-07-06

---

## Phase 1: Database Schema & Repository

### Step 1.1: Create migration file

**File:** `migrations/20260706120000_slack_status_monitor.up.sql`

Create three tables:
- `slack_config` — encrypted bot token, signing secret, webhook URL, channel routing, severity config, quiet hours, enabled flag. UNIQUE on tenant_id.
- `monitored_components` — id (PK), name, type, enabled, slack_channel override. Seed with 20 components.
- `slack_alert_log` — component_id, old_status, new_status, severity, channel, message_ts, delivered, error, created_at.

Indexes:
- `idx_slack_config_tenant` on `slack_config(tenant_id)`
- `idx_slack_config_enabled` partial on `slack_config(enabled) WHERE enabled = TRUE`
- `idx_slack_alert_log_component` on `slack_alert_log(component_id, created_at DESC)`
- `idx_slack_alert_log_created` on `slack_alert_log(created_at DESC)`

**File:** `migrations/20260706120000_slack_status_monitor.down.sql`

Drop tables in reverse order (slack_alert_log, monitored_components, slack_config).

### Step 1.2: Create Slack repository

**File:** `internal/storage/sql/slack_repository.go`

Follow the `PostgresRepository` pattern from `internal/notification/repository.go`:
- Struct: `SlackRepository struct { db *sql.DB }`
- Constructor: `NewSlackRepository(db *sql.DB) *SlackRepository`

Methods:
- `GetSlackConfig(ctx, tenantID) (*SlackConfig, error)` — SELECT from slack_config
- `UpsertSlackConfig(ctx, config *SlackConfig) error` — INSERT ... ON CONFLICT (tenant_id) DO UPDATE
- `GetMonitoredComponents(ctx) ([]*MonitoredComponent, error)` — SELECT from monitored_components WHERE enabled = TRUE
- `CreateAlertLog(ctx, log *SlackAlertLog) error` — INSERT into slack_alert_log
- `GetAlertLogs(ctx, componentID string, limit int) ([]*SlackAlertLog, error)` — SELECT with LIMIT

Models (define in same file or `internal/storage/sql/slack_models.go`):
```go
type SlackConfig struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    BotTokenEnc    []byte
    SigningSecret   []byte
    WebhookURL     string
    AlertChannel   string
    ReportChannel  string
    ChannelRouting map[string]string  // component_id -> channel_id
    SeverityConfig map[string]bool    // severity -> enabled
    QuietHours     QuietHoursConfig
    Enabled        bool
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type QuietHoursConfig struct {
    Enabled  bool   `json:"enabled"`
    Start    string `json:"start"`
    End      string `json:"end"`
    Timezone string `json:"timezone"`
}

type MonitoredComponent struct {
    ID           string
    Name         string
    Type         string
    Enabled      bool
    SlackChannel string
    CreatedAt    time.Time
}

type SlackAlertLog struct {
    ID          uuid.UUID
    ComponentID string
    OldStatus   string
    NewStatus   string
    Severity    string
    Channel     string
    MessageTS   string
    Delivered   bool
    Error       string
    CreatedAt   time.Time
}
```

**File:** `internal/storage/sql/slack_repository_test.go`

Test against local Postgres (port 5432). Test CRUD for all methods.

---

## Phase 2: Slack Channel (Notification System)

### Step 2.1: Add ChannelSlack constant

**File:** `internal/notification/types.go`

Add to channel constants:
```go
ChannelSlack = "slack"
```

### Step 2.2: Create Slack channel implementation

**File:** `internal/notification/slack_channel.go`

Follow `WebhookChannel` pattern exactly:

```go
type SlackChannel struct {
    client      *http.Client
    logger      *logrus.Logger
    repo        Repository
    webhookURL  string
    mu          sync.Mutex  // rate limiting
    lastSend    time.Time
}
```

Constructor: `NewSlackChannel(webhookURL string, logger *logrus.Logger) *SlackChannel`

Methods:
- `Name() string` — returns `ChannelSlack`
- `IsConfigured() bool` — returns `c.webhookURL != ""`
- `SetRepository(repo Repository)` — setter for repo injection
- `Send(ctx, n, user) error` — main delivery logic:
  1. Rate limit: lock mutex, check if 1s has passed since lastSend, sleep if not
  2. Build Block Kit payload using `buildStatusAlertPayload(n)` from templates
  3. Marshal to JSON
  4. POST to webhook URL with `Content-Type: application/json`
  5. Retry 3x with exponential backoff (1s, 2s, 4s)
  6. On 404/410: log error, return err (webhook revoked)
  7. On 429: read Retry-After header, sleep, retry
  8. Record metrics via prometheus counters

### Step 2.3: Create Block Kit templates

**File:** `internal/notification/slack_templates.go`

Functions:
- `buildStatusAlertPayload(n *Notification) map[string]interface{}` — builds Block Kit for service status changes
- `buildIncidentPayload(n *Notification) map[string]interface{}` — builds Block Kit for incident notifications
- `buildMaintenancePayload(n *Notification) map[string]interface{}` — builds Block Kit for maintenance alerts
- `buildRecoveryPayload(n *Notification) map[string]interface{}` — builds Block Kit for recovery notifications
- `buildDailyDigestPayload(services []ServiceStatus, incidents []Incident) map[string]interface{}` — daily report
- `buildWeeklyReportPayload(services []ServiceStatus, incidents []Incident, latency LatencySummary) map[string]interface{}` — weekly report

Severity-to-format mapping (in `severityConfig` function):
```go
func severityConfig(severity string) (emoji, color string, mention string) {
    switch severity {
    case "critical": return ":red_circle:", "#FF0000", "<!channel>"
    case "high":     return ":large_orange_circle:", "#FF6600", "<!here>"
    case "medium":   return ":large_yellow_circle:", "#FFCC00", ""
    case "info":     return ":large_green_circle:", "#00CC00", ""
    case "maintenance": return ":large_blue_circle:", "#0066FF", ""
    }
}
```

Reference: `internal/consciousness/dispatcher.go` lines 209-280 for exact Block Kit structure.

### Step 2.4: Register channel in service constructor

**File:** `internal/notification/service.go`

In `NewService()`, after the existing channel registrations, add:
```go
slackChannel := NewSlackChannel(os.Getenv("SLACK_WEBHOOK_URL"), logger)
slackChannel.SetRepository(repo)
s.channels[ChannelSlack] = slackChannel
```

### Step 2.5: Tests

**File:** `internal/notification/slack_channel_test.go`

- Test `IsConfigured()` with/without webhook URL
- Test `Send()` with mock HTTP server (success, retry on 500, skip on 429 with Retry-After)
- Test rate limiting (verify 1s gap between requests)
- Test Block Kit payload structure for each severity level

---

## Phase 3: Status Monitor

### Step 3.1: Create status monitor service

**File:** `internal/services/status_monitor.go`

Follow `UsageForecaster` pattern:

```go
type StatusMonitor struct {
    slackRepo    *sql.SlackRepository
    redis        *redis.Client
    queue        *notification.Queue
    logger       *logrus.Logger
    config       *StatusMonitorConfig
    stopChan     chan struct{}
    stopOnce     sync.Once
}

type StatusMonitorConfig struct {
    Enabled       bool
    IntervalSec   int
    WebhookURL    string
}
```

Constructor: `NewStatusMonitor(slackRepo, redis, queue, logger, config) *StatusMonitor`

Methods:
- `Start(ctx context.Context)` — if enabled, spawn `go f.runLoop(ctx)`
- `Stop()` — close stopChan
- `runLoop(ctx context.Context)` — ticker loop:
  1. Get monitored components from DB
  2. For each component, query current state from `system_health_checks` + `circuit_state`
  3. Compare with Redis key `status:monitor:last_state:{component_id}`
  4. If changed: determine severity, create `Notification`, call `queue.Enqueue()`, update Redis key with 24h TTL
  5. Log summary (N checked, M changed)

State determination logic:
```go
func (m *StatusMonitor) determineState(healthCheck *HealthCheckResult, circuit *CircuitState) string {
    if circuit != nil && circuit.State == "open" { return "major_outage" }
    if healthCheck == nil { return "unknown" }
    if healthCheck.SuccessRate < 0.5 { return "major_outage" }
    if healthCheck.SuccessRate < 0.9 { return "partial_outage" }
    if healthCheck.LatencyP95 > healthCheck.BaselineLatency*2 { return "degraded_performance" }
    return "operational"
}
```

Notification type mapping:
```go
func notificationTypeForTransition(oldState, newState string) string {
    if newState == "operational" { return notification.TypeProviderOnline }
    if newState == "major_outage" { return notification.TypeProviderOffline }
    if newState == "degraded_performance" || newState == "partial_outage" { return notification.TypeProviderDegraded }
    return notification.TypeProviderDegraded
}
```

### Step 3.2: Tests

**File:** `internal/services/status_monitor_test.go`

- Test state determination with various health check + circuit combinations
- Test dedup logic (mock Redis, verify no dispatch when state unchanged)
- Test notification creation on state transition
- Test severity mapping

---

## Phase 4: Slack App Handler (Slash Commands)

### Step 4.1: Create request verification middleware

**File:** `internal/api/middleware/slack_verify.go`

Follow the pattern from `internal/agent/runtime/router.go` lines 786-903:

```go
func SlackVerifyMiddleware(signingSecret string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // Read body
            // Compute HMAC-SHA256 of "v0:" + timestamp + ":" + body
            // Compare with X-Slack-Signature header
            // Check timestamp is within 5 minutes (replay protection)
            // If invalid: return 401
            next.ServeHTTP(w, r)
        }
    }
}
```

### Step 4.2: Create Slack handler

**File:** `internal/api/handlers/slack/handler.go`

```go
type Handler struct {
    statusRepo   storage.Repository
    slackRepo    *sql.SlackRepository
    logger       *logrus.Logger
}
```

Constructor: `NewHandler(statusRepo, slackRepo, logger) *Handler`

### Step 4.3: Implement slash commands

**File:** `internal/api/handlers/slack/commands.go`

Functions:
- `HandleSlashCommand(w, r)` — parse `command` and `text` from form body, route to:
  - `handleStatusCommand(w, r, text)` — query all components, build Block Kit response
  - `handleUptimeCommand(w, r, text)` — parse service name + period, query uptime metrics
  - `handleIncidentsCommand(w, r, text)` — query active incidents

Response strategy:
1. Parse form values (`command`, `text`, `response_url`)
2. Return `200 OK` immediately with `{"response_type": "ephemeral", "text": "Fetching..."}`
3. Spawn goroutine to query DB and POST full response to `response_url`

### Step 4.4: Implement interaction handler

**File:** `internal/api/handlers/slack/handler.go`

- `HandleInteractions(w, r)` — parse `payload` JSON, switch on `action_id`:
  - `ack_incident` — update incident status to "investigating"
  - `resolve_incident` — update incident status to "resolved"
  - `view_status` — no-op (button has URL)

### Step 4.5: Register routes

**File:** `internal/api/routes.go`

Add Slack routes (public, with signing secret verification):
```go
slackHandler := slack.NewHandler(s.storageRepo, s.slackRepo, s.logger)
api.HandleFunc("/slack/commands", middleware.SlackVerifyMiddleware(slackSigningSecret)(slackHandler.HandleSlashCommand)).Methods("POST")
api.HandleFunc("/slack/interactions", middleware.SlackVerifyMiddleware(slackSigningSecret)(slackHandler.HandleInteractions)).Methods("POST")
```

### Step 4.6: Tests

**File:** `internal/api/handlers/slack/handler_test.go`

- Test request verification (valid signature, invalid signature, expired timestamp)
- Test `/status` command response format
- Test `/uptime` command with service filter
- Test `/incidents` command with active incidents
- Test interaction handler for ack/resolve actions

---

## Phase 5: Report Scheduler

### Step 5.1: Create report scheduler

**File:** `internal/services/status_reporter.go`

Follow `UsageForecaster` pattern:

```go
type StatusReporter struct {
    slackRepo    *sql.SlackRepository
    httpClient   *http.Client
    logger       *logrus.Logger
    config       *StatusReporterConfig
    stopChan     chan struct{}
    stopOnce     sync.Once
}

type StatusReporterConfig struct {
    Enabled      bool
    DailyCron    string
    WeeklyCron   string
    WebhookURL   string
    ReportChannel string
}
```

Constructor: `NewStatusReporter(slackRepo, logger, config) *StatusReporter`

Methods:
- `Start(ctx context.Context)` — if enabled, spawn `go f.runDaily(ctx)` and `go f.runWeekly(ctx)`
- `Stop()` — close stopChan
- `runDaily(ctx)` — parse cron, wait for next occurrence, query status data, build daily digest payload, POST to Slack
- `runWeekly(ctx)` — same pattern for weekly
- `buildDailyReport(ctx) (*DailyReport, error)` — query all components, uptime metrics, overnight incidents
- `buildWeeklyReport(ctx) (*WeeklyReport, error)` — query 7-day trends, latency percentiles, incident summary
- `postToSlack(payload map[string]interface{}) error` — HTTP POST to webhook URL (reuse pattern from consciousness dispatcher)

### Step 5.2: Tests

**File:** `internal/services/status_reporter_test.go`

- Test daily report payload structure
- Test weekly report payload structure
- Test cron parsing and scheduling

---

## Phase 6: Admin Dashboard Configuration

### Step 6.1: Create admin API handler

**File:** `internal/api/handlers/admin/slack_handler.go`

Methods:
- `HandleGetSlackConfig(w, r)` — GET, returns config (mask secrets)
- `HandleUpdateSlackConfig(w, r)` — PUT, upserts config
- `HandleTestSlackMessage(w, r)` — POST, sends test message to configured channel
- `HandleListSlackChannels(w, r)` — GET, calls Slack API `conversations.list` with bot token

### Step 6.2: Register admin routes

**File:** `internal/api/routes_admin.go`

Add within the admin routes section:
```go
adminRoutes.HandleFunc("/slack/config", adminSlackHandler.HandleGetSlackConfig).Methods("GET")
adminRoutes.HandleFunc("/slack/config", adminSlackHandler.HandleUpdateSlackConfig).Methods("PUT")
adminRoutes.HandleFunc("/slack/test", adminSlackHandler.HandleTestSlackMessage).Methods("POST")
adminRoutes.HandleFunc("/slack/channels", adminSlackHandler.HandleListSlackChannels).Methods("GET")
```

### Step 6.3: Create admin dashboard page

**File:** `web/admin-dashboard/src/pages/IntegrationsSlackPage.tsx`

Follow existing admin page patterns. Components:
- Connection status badge (green/red based on test message result)
- Form fields: Webhook URL, Bot Token (password input), Signing Secret (password input)
- Channel selectors: Alert Channel, Report Channel (dropdown from `/api/v1/admin/slack/channels`)
- Per-service channel routing table (20 rows, each with optional channel dropdown)
- Severity checkboxes (Critical, High, Medium, Low)
- Quiet hours time pickers with timezone selector
- Enable/disable toggle
- Test button (POST to `/api/v1/admin/slack/test`)

**File:** `web/admin-dashboard/src/api/slack.ts`

API client:
```typescript
export const slackApi = {
  getConfig: () => api.get('/v1/admin/slack/config'),
  updateConfig: (config: SlackConfig) => api.put('/v1/admin/slack/config', config),
  testMessage: () => api.post('/v1/admin/slack/test'),
  listChannels: () => api.get('/v1/admin/slack/channels'),
}
```

### Step 6.4: Add admin sidebar link

**File:** `web/admin-dashboard/src/components/layout/AdminSidebar.tsx`

Add "Slack Integration" entry under Monitoring section with `MessageSquare` icon.

### Step 6.5: Add admin route

**File:** `web/admin-dashboard/src/routes/adminRoutes.tsx`

Add route: `{ path: '/integrations/slack', element: <IntegrationsSlackPage /> }`

---

## Phase 7: Wiring & Integration

### Step 7.1: Initialize services in server.go

**File:** `internal/api/server.go`

In `NewServer()`, after notification service initialization:
```go
// Slack repository
slackRepo := sql.NewSlackRepository(db.DB)
s.slackRepo = slackRepo

// Status monitor
monitorConfig := &services.StatusMonitorConfig{
    Enabled:     os.Getenv("STATUS_MONITOR_ENABLED") != "false",
    IntervalSec: parseIntEnv("STATUS_MONITOR_INTERVAL_SEC", 30),
    WebhookURL:  os.Getenv("SLACK_WEBHOOK_URL"),
}
statusMonitor := services.NewStatusMonitor(slackRepo, redisClient, notificationQueue, logger, monitorConfig)
s.statusMonitor = statusMonitor

// Report scheduler
reporterConfig := &services.StatusReporterConfig{
    Enabled:       os.Getenv("SLACK_REPORT_ENABLED") != "false",
    DailyCron:     getEnvOrDefault("SLACK_REPORT_DAILY_CRON", "0 9 * * *"),
    WeeklyCron:    getEnvOrDefault("SLACK_REPORT_WEEKLY_CRON", "0 9 * * 1"),
    WebhookURL:    os.Getenv("SLACK_WEBHOOK_URL"),
    ReportChannel: os.Getenv("SLACK_REPORT_CHANNEL"),
}
statusReporter := services.NewStatusReporter(slackRepo, logger, reporterConfig)
s.statusReporter = statusReporter
```

### Step 7.2: Start services in ListenAndServe()

**File:** `internal/api/server.go`

In `ListenAndServe()`, after `s.notificationSvc.Start(s.serverCtx)`:
```go
s.statusMonitor.Start(s.serverCtx)
s.statusReporter.Start(s.serverCtx)
```

### Step 7.3: Graceful shutdown

In the shutdown handler, add:
```go
s.statusMonitor.Stop()
s.statusReporter.Stop()
```

---

## Phase 8: User Dashboard Slack Toggle

### Step 8.1: Add Slack fields to notification preferences

**File:** `internal/storage/sql/notification_repository.go`

Add `slack_enabled` column to `notification_preferences` table (migration + repository update).

**File:** `migrations/20260706120100_notification_preferences_slack.up.sql`
```sql
ALTER TABLE notification_preferences ADD COLUMN IF NOT EXISTS slack_enabled BOOLEAN DEFAULT FALSE;
```

### Step 8.2: Update frontend notification preferences

**File:** `web/dashboard/src/pages/SettingsPage/components/NotificationsSettingsTab.tsx`

Add a "Slack" toggle group (visible only when admin has enabled Slack for the tenant). The toggle controls `slack_enabled` per category.

**File:** `web/dashboard/src/types/notifications.ts`

Add `slack` field to `NotificationPreferences` interface:
```typescript
slack?: { enabled: boolean; categories: Record<string, boolean> };
```

---

## Build Order

Execute phases in this order (each phase builds on the previous):

1. **Phase 1** — DB schema + repository (foundation, no dependencies)
2. **Phase 2** — Slack channel (depends on Phase 1 for types)
3. **Phase 3** — Status monitor (depends on Phase 1 repo + Phase 2 channel)
4. **Phase 4** — Slash commands (depends on Phase 1 repo)
5. **Phase 5** — Report scheduler (depends on Phase 1 repo + Phase 2 templates)
6. **Phase 6** — Admin dashboard (depends on Phase 1 repo + Phase 6 API handler)
7. **Phase 7** — Wiring (depends on all previous phases)
8. **Phase 8** — User dashboard toggle (depends on Phase 1)

Phases 3, 4, and 5 can be done in parallel after Phase 2.

---

## Environment Variables Summary

| Variable | Phase | Purpose | Default |
|----------|-------|---------|---------|
| `SLACK_WEBHOOK_URL` | 2,3,5 | Incoming Webhook URL | (unset) |
| `SLACK_BOT_TOKEN` | 4,6 | Bot token for slash commands | (unset) |
| `SLACK_SIGNING_SECRET` | 4 | Request verification | (unset) |
| `STATUS_MONITOR_ENABLED` | 3 | Enable monitor | `true` |
| `STATUS_MONITOR_INTERVAL_SEC` | 3 | Poll interval | `30` |
| `SLACK_REPORT_ENABLED` | 5 | Enable reports | `true` |
| `SLACK_REPORT_DAILY_CRON` | 5 | Daily cron | `0 9 * * *` |
| `SLACK_REPORT_WEEKLY_CRON` | 5 | Weekly cron | `0 9 * * 1` |
| `SLACK_REPORT_CHANNEL` | 5 | Report channel ID | (unset) |

---

## Verification

After implementation, verify with:

1. **Unit tests:** `go test ./internal/notification/... ./internal/services/... ./internal/api/handlers/slack/...`
2. **Lint:** `golangci-lint run`
3. **Build:** `go build -o bin/orchestrator-api ./cmd/orchestrator-api`
4. **Manual test:** Configure webhook URL, POST to `/api/v1/admin/slack/test`, verify message in Slack workspace
5. **Slash command test:** Configure bot token + signing secret, type `/status` in Slack channel
