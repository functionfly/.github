# Notification System Implementation Plan

## Detailed Implementation Steps

### Phase 1: Core Database & Models

#### 1.1 Create Database Models
**Files to create:**
- `internal/notification/models.go`

**Key structures:**
```go
// Notification - stores individual notifications
type Notification struct {
    ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID      uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
    Type        string          `json:"type" gorm:"not null;index"`      // e.g., "deployment.success"
    Category    string          `json:"category" gorm:"not null;index"`  // e.g., "deployments"
    Title       string          `json:"title" gorm:"not null"`
    Body        string          `json:"body" gorm:"type:text"`
    Data        JSONMap         `json:"data" gorm:"type:jsonb"`
    Channels    StringArray     `json:"channels" gorm:"type:text[]"`
    Priority    string          `json:"priority" gorm:"not null;default:'normal'"` // low, normal, high, urgent
    Status      string          `json:"status" gorm:"not null;default:'pending'"`  // pending, processing, sent, failed, read
    ReadAt      *time.Time      `json:"read_at,omitempty"`
    SentAt      *time.Time      `json:"sent_at,omitempty"`
    ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
    CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationPreference - user preferences per channel/category
type NotificationPreference struct {
    ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID          uuid.UUID   `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_channel_category"`
    Channel         string      `json:"channel" gorm:"not null;uniqueIndex:idx_user_channel_category"` // email, in_app, webhook, push
    Category        string      `json:"category" gorm:"not null;uniqueIndex:idx_user_channel_category"` // system, security, billing, etc.
    Enabled         bool        `json:"enabled" gorm:"default:true"`
    Frequency       string      `json:"frequency" gorm:"default:'immediate'"` // immediate, digest_daily, digest_weekly
    QuietHoursStart *string     `json:"quiet_hours_start,omitempty"` // HH:MM format
    QuietHoursEnd   *string     `json:"quiet_hours_end,omitempty"`   // HH:MM format
    Timezone        string      `json:"timezone" gorm:"default:'UTC'"`
    WebhookURL      *string     `json:"webhook_url,omitempty"`
    WebhookSecret   *string     `json:"webhook_secret,omitempty"`
    CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationTemplate - templates for each notification type/channel
type NotificationTemplate struct {
    ID          uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Type        string      `json:"type" gorm:"not null;uniqueIndex:idx_type_channel"` // e.g., "deployment.success"
    Channel     string      `json:"channel" gorm:"not null;uniqueIndex:idx_type_channel"` // email, in_app, webhook
    Subject     string      `json:"subject"` // For email notifications
    BodyHTML    string      `json:"body_html" gorm:"type:text"`
    BodyText    string      `json:"body_text" gorm:"type:text"`
    Variables   StringArray `json:"variables" gorm:"type:text[]"`
    IsActive    bool        `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationAnalytics - tracking delivery and engagement
type NotificationAnalytics struct {
    ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    NotificationID  uuid.UUID   `json:"notification_id" gorm:"type:uuid;not null;index"`
    Channel         string      `json:"channel" gorm:"not null"`
    Status          string      `json:"status" gorm:"not null"` // delivered, failed, bounced, opened, clicked
    ErrorMessage    *string     `json:"error_message,omitempty"`
    DeliveredAt     *time.Time  `json:"delivered_at,omitempty"`
    OpenedAt        *time.Time  `json:"opened_at,omitempty"`
    ClickedAt       *time.Time  `json:"clicked_at,omitempty"`
    IPAddress       *string     `json:"ip_address,omitempty"`
    UserAgent       *string     `json:"user_agent,omitempty"`
    CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
}
```

#### 1.2 Create Repository Layer
**Files to create:**
- `internal/notification/repository.go`

**Methods to implement:**
```go
type Repository interface {
    // Notifications
    CreateNotification(ctx context.Context, n *Notification) error
    GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error)
    ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
    MarkAsRead(ctx context.Context, id uuid.UUID) error
    MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
    DeleteNotification(ctx context.Context, id uuid.UUID) error
    UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string) error

    // Preferences
    GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)
    GetPreference(ctx context.Context, userID uuid.UUID, channel, category string) (*NotificationPreference, error)
    SavePreference(ctx context.Context, p *NotificationPreference) error
    CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error

    // Templates
    GetTemplate(ctx context.Context, notificationType, channel string) (*NotificationTemplate, error)
    ListTemplates(ctx context.Context) ([]*NotificationTemplate, error)
    SaveTemplate(ctx context.Context, t *NotificationTemplate) error
    DeleteTemplate(ctx context.Context, id uuid.UUID) error

    // Analytics
    TrackAnalytics(ctx context.Context, a *NotificationAnalytics) error
    GetAnalytics(ctx context.Context, notificationID uuid.UUID) ([]*NotificationAnalytics, error)
}
```

#### 1.3 Create Migrations
**Files to create:**
- `internal/storage/migrations/xxx_add_notifications.up.sql`
- `internal/storage/migrations/xxx_add_notifications.down.sql`

**Migration SQL:**
```sql
-- Notifications table
CREATE TABLE notifications (
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

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, status) WHERE status != 'read';

-- Notification preferences table
CREATE TABLE notification_preferences (
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

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Notification templates table
CREATE TABLE notification_templates (
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

CREATE INDEX idx_notification_templates_type ON notification_templates(type);
CREATE INDEX idx_notification_templates_active ON notification_templates(is_active);

-- Notification analytics table
CREATE TABLE notification_analytics (
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

CREATE INDEX idx_notification_analytics_notification_id ON notification_analytics(notification_id);
CREATE INDEX idx_notification_analytics_status ON notification_analytics(status);
```

### Phase 2: Core Service Implementation

#### 2.1 Types and Interfaces
**File:** `internal/notification/types.go`

```go
package notification

import "context"

// Channel is the interface for notification channels
type Channel interface {
    Name() string
    Send(ctx context.Context, notification *Notification, user *storage.User) error
    IsConfigured() bool
}

// Trigger is the interface for event triggers
type Trigger interface {
    Name() string
    ShouldTrigger(event interface{}) bool
    BuildNotification(event interface{}) (*Notification, error)
}

// Priority levels
const (
    PriorityLow    = "low"
    PriorityNormal = "normal"
    PriorityHigh   = "high"
    PriorityUrgent = "urgent"
)

// Status values
const (
    StatusPending    = "pending"
    StatusProcessing = "processing"
    StatusSent       = "sent"
    StatusFailed     = "failed"
    StatusRead       = "read"
    StatusDelivered  = "delivered"
)

// Channel types
const (
    ChannelEmail   = "email"
    ChannelInApp   = "in_app"
    ChannelWebhook = "webhook"
    ChannelPush    = "push"
)

// Categories
const (
    CategorySystem     = "system"
    CategorySecurity   = "security"
    CategoryBilling    = "billing"
    CategoryDeployment = "deployment"
    CategoryFunction   = "function"
    CategoryTeam       = "team"
    CategoryRegistry   = "registry"
)

// Frequencies
const (
    FrequencyImmediate    = "immediate"
    FrequencyDigestDaily  = "digest_daily"
    FrequencyDigestWeekly = "digest_weekly"
)
```

#### 2.2 Core Service
**File:** `internal/notification/service.go`

```go
package notification

type Service struct {
    repo       Repository
    channels   map[string]Channel
    templates  *TemplateEngine
    queue      *Queue
    dispatcher *Dispatcher
    logger     *logrus.Logger
    db         *storage.PostgresDB
}

func NewService(repo Repository, db *storage.PostgresDB, emailSvc email.Service, logger *logrus.Logger) *Service {
    s := &Service{
        repo:   repo,
        db:     db,
        logger: logger,
        channels: make(map[string]Channel),
    }

    // Register channels
    s.channels[ChannelEmail] = NewEmailChannel(emailSvc, logger)
    s.channels[ChannelInApp] = NewInAppChannel(repo, db, logger)
    s.channels[ChannelWebhook] = NewWebhookChannel(logger)

    // Initialize queue and dispatcher
    s.queue = NewQueue(repo, logger)
    s.dispatcher = NewDispatcher(s.channels, repo, logger)
    s.templates = NewTemplateEngine(repo)

    return s
}

// Send creates and sends a notification
func (s *Service) Send(ctx context.Context, req SendRequest) error {
    // Build notification from request
    notification := &Notification{
        UserID:   req.UserID,
        Type:     req.Type,
        Category: req.Category,
        Title:    req.Title,
        Body:     req.Body,
        Data:     req.Data,
        Channels: req.Channels,
        Priority: req.Priority,
        Status:   StatusPending,
    }

    // Save to database
    if err := s.repo.CreateNotification(ctx, notification); err != nil {
        return fmt.Errorf("failed to create notification: %w", err)
    }

    // Add to processing queue
    s.queue.Enqueue(notification)

    return nil
}

// Broadcast sends to multiple users
func (s *Service) Broadcast(ctx context.Context, req BroadcastRequest) error {
    // Implementation for broadcasting
}

// GetUnreadCount returns unread notification count for user
func (s *Service) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
    return s.repo.GetUnreadCount(ctx, userID)
}

// Start begins processing the notification queue
func (s *Service) Start(ctx context.Context) {
    s.queue.Start(ctx, s.dispatcher)
}
```

### Phase 3: Channel Implementations

#### 3.1 Email Channel
**File:** `internal/notification/channels/email.go`

Leverages existing `internal/email/email.go` service.

```go
package channels

type EmailChannel struct {
    emailSvc email.Service
    logger   *logrus.Logger
}

func NewEmailChannel(emailSvc email.Service, logger *logrus.Logger) *EmailChannel {
    return &EmailChannel{
        emailSvc: emailSvc,
        logger:   logger,
    }
}

func (c *EmailChannel) Name() string {
    return "email"
}

func (c *EmailChannel) Send(ctx context.Context, notification *Notification, user *storage.User) error {
    // Use template engine to render email
    // Call email service
    // Track analytics
}

func (c *EmailChannel) IsConfigured() bool {
    return c.emailSvc != nil
}
```

#### 3.2 In-App Channel
**File:** `internal/notification/channels/inapp.go`

Uses PostgreSQL pg_notify for real-time delivery.

```go
package channels

type InAppChannel struct {
    repo   notification.Repository
    db     *storage.PostgresDB
    logger *logrus.Logger
}

func (c *InAppChannel) Send(ctx context.Context, notification *Notification, user *storage.User) error {
    // Notification already stored in DB
    // Trigger pg_notify for real-time updates
    payload, _ := json.Marshal(map[string]interface{}{
        "type": "notification",
        "user_id": notification.UserID,
        "notification_id": notification.ID,
        "title": notification.Title,
        "body": notification.Body,
        "created_at": notification.CreatedAt,
    })

    return c.db.PgNotify("user_notifications", string(payload))
}
```

#### 3.3 Webhook Channel
**File:** `internal/notification/channels/webhook.go`

```go
package channels

type WebhookChannel struct {
    client *http.Client
    logger *logrus.Logger
}

func (c *WebhookChannel) Send(ctx context.Context, notification *Notification, user *storage.User) error {
    // Get webhook URL from user preferences
    // Sign payload if secret is configured
    // Send HTTP POST
    // Handle retries
}
```

### Phase 4: API Handlers

#### 4.1 REST Handlers
**File:** `internal/api/handlers/notifications/handler.go`

```go
package notifications

type Handler struct {
    service notification.Service
    repo    notification.Repository
}

func (h *Handler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
    // GET /v1/notifications
    // Support pagination, filtering by status, category
}

func (h *Handler) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
    // GET /v1/notifications/unread-count
}

func (h *Handler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
    // PATCH /v1/notifications/{id}/read
}

func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
    // POST /v1/notifications/read-all
}

func (h *Handler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
    // GET /v1/users/me/notification-preferences
}

func (h *Handler) HandleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
    // PATCH /v1/users/me/notification-preferences
}
```

#### 4.2 WebSocket Handler
**File:** `internal/api/handlers/notifications/websocket.go`

```go
package notifications

// HandleWebSocket upgrades connection and streams notifications
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Authenticate user
    // Upgrade to WebSocket
    // Subscribe to pg_notify channel
    // Stream notifications to client
}
```

### Phase 5: Event Triggers

#### 5.1 Deployment Triggers
**File:** `internal/notification/triggers/deployment.go`

```go
package triggers

func (t *DeploymentTrigger) BuildNotification(event *DeploymentEvent) (*notification.Notification, error) {
    switch event.Status {
    case "success":
        return &notification.Notification{
            Type:     "deployment.success",
            Category: notification.CategoryDeployment,
            Title:    "Deployment Successful",
            Body:     fmt.Sprintf("Your deployment of %s was successful.", event.AppName),
            Priority: notification.PriorityNormal,
            Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
        }, nil
    case "failed":
        return &notification.Notification{
            Type:     "deployment.failed",
            Category: notification.CategoryDeployment,
            Title:    "Deployment Failed",
            Body:     fmt.Sprintf("Your deployment of %s failed. Click to view logs.", event.AppName),
            Priority: notification.PriorityHigh,
            Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
        }, nil
    }
}
```

### Phase 6: Integration

#### 6.1 Wire into API Server
**File:** `internal/api/server.go`

Add notification service initialization and routes.

#### 6.2 Wire into Existing Services
Update existing services to trigger notifications:
- `internal/deployment/service.go` - Deployment events
- `internal/billing/service.go` - Billing events
- `internal/auth/service.go` - Security events

## Testing Strategy

1. **Unit Tests**: Mock channels and repository
2. **Integration Tests**: Test with test database
3. **End-to-End Tests**: Full flow from trigger to delivery

## Deployment Steps

1. Run database migrations
2. Seed default templates
3. Deploy code changes
4. Enable notification service
5. Monitor delivery rates
