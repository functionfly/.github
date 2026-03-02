# Notification System

A comprehensive multi-channel notification system for the FunctionFly platform supporting email, in-app, webhook, and push notifications with user preferences, templates, and real-time delivery.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│ Event       │────▶│ Notification │────▶│ Channel         │
│ Sources     │     │ Service      │     │ Dispatcher      │
└─────────────┘     └──────────────┘     └─────────────────┘
                            │                       │
                            ▼                       ▼
                     ┌──────────────┐     ┌─────────────────┐
                     │ PostgreSQL   │     │ Email, In-App,  │
                     │ (Queue)      │     │ Webhook, Push   │
                     └──────────────┘     └─────────────────┘
```

## Components

### Core Components

1. **Models** ([`models.go`](models.go)) - Database models for notifications, preferences, templates, and analytics
2. **Repository** ([`repository.go`](repository.go)) - Data access layer for notification operations
3. **Types** ([`types.go`](types.go)) - Constants, interfaces, and shared types
4. **Service** ([`service.go`](service.go)) - Core notification service with queue and dispatcher

### Channels

Located in [`channels/`](channels/):

- **Email** ([`email.go`](channels/email.go)) - Email notifications via SMTP
- **In-App** ([`inapp.go`](channels/inapp.go)) - Real-time notifications via pg_notify and WebSocket
- **Webhook** ([`webhook.go`](channels/webhook.go)) - HTTP webhook notifications with HMAC signing

### Triggers

Located in [`triggers/`](triggers/):

- **Deployment** ([`deployment.go`](triggers/deployment.go)) - Deployment success/failure notifications
- **Billing** ([`billing.go`](triggers/billing.go)) - Invoice, payment, and subscription notifications
- **Security** ([`security.go`](triggers/security.go)) - Password change, MFA, and security alerts
- **Team** ([`team.go`](triggers/team.go)) - Team invitation and membership notifications
- **Registry** ([`registry.go`](triggers/registry.go)) - Function registry notifications

### API Handlers

Located in [`../api/handlers/notifications/`](../api/handlers/notifications/):

- **Handler** ([`handler.go`](../api/handlers/notifications/handler.go)) - REST API endpoints
- **WebSocket** ([`websocket.go`](../api/handlers/notifications/websocket.go)) - Real-time WebSocket connections

## Database Schema

### Tables

1. **notifications** - Stores individual notifications
2. **notification_preferences** - User preferences per channel/category
3. **notification_templates** - Templates for each notification type
4. **notification_analytics** - Delivery and engagement tracking

See migration file: [`migrations/000076_add_notifications.up.sql`](../../../migrations/000076_add_notifications.up.sql)

## Usage

### Sending a Notification

```go
import (
    "github.com/functionfly/functionfly/internal/notification"
    "github.com/functionfly/functionfly/internal/notification/triggers"
)

// Create notification service
repo := notification.NewPostgresRepository(db)
svc := notification.NewService(repo, db, emailSvc, logger)

// Start the service
svc.Start(ctx)

// Send a notification
req := notification.SendRequest{
    UserID:   userID,
    Type:     notification.TypeDeploymentSuccess,
    Category: notification.CategoryDeployment,
    Title:    "Deployment Successful",
    Body:     "Your deployment was successful!",
    Priority: notification.PriorityNormal,
    Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
}
notification, err := svc.Send(ctx, req)
```

### Using Triggers

```go
// Create trigger registry
registry := triggers.NewTriggerRegistry()

// Create and process an event
event := &triggers.DeploymentEvent{
    AppID:   appID,
    AppName: "my-app",
    UserID:  userID,
    Status:  "success",
}

notification, err := registry.ProcessEvent(event)
if err != nil {
    log.Fatal(err)
}

// Send the notification
_, err = svc.Send(ctx, notification.SendRequest())
```

### Updating Preferences

```go
pref := &notification.NotificationPreference{
    UserID:   userID,
    Channel:  notification.ChannelEmail,
    Category: notification.CategoryDeployment,
    Enabled:  true,
    Frequency: notification.FrequencyImmediate,
}

err := svc.SavePreference(ctx, pref)
```

## API Endpoints

### REST Endpoints

- `GET /v1/notifications` - List notifications
- `GET /v1/notifications/unread-count` - Get unread count
- `POST /v1/notifications/read-all` - Mark all as read
- `PATCH /v1/notifications/{id}/read` - Mark as read
- `DELETE /v1/notifications/{id}` - Delete notification
- `GET /v1/users/me/notification-preferences` - Get preferences
- `PATCH /v1/users/me/notification-preferences` - Update preferences

### WebSocket

- `WS /v1/notifications/stream` - Real-time notification stream

## Configuration

### Environment Variables

- `NOTIFICATION_EMAIL_ENABLED` - Enable email notifications (default: true)
- `NOTIFICATION_WEBHOOK_ENABLED` - Enable webhook notifications (default: true)
- `NOTIFICATION_QUEUE_WORKERS` - Number of queue workers (default: 5)
- `NOTIFICATION_QUEUE_SIZE` - Queue buffer size (default: 1000)

## Testing

```bash
# Run notification tests
go test ./internal/notification/...

# Run with coverage
go test -cover ./internal/notification/...
```

## Migration

To apply the notification system migrations:

```bash
# Using the migration tool
make migrate-up

# Or directly with PostgreSQL
psql -d functionfly -f migrations/000076_add_notifications.up.sql
```
