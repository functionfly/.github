package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	statusMonitorLastStateKey = "status:monitor:last_state:%s"
	statusMonitorTTL         = 24 * time.Hour
)

type StatusMonitorConfig struct {
	Enabled       bool
	Interval      time.Duration
	RedisAddr    string
	WebhookURL   string
}

func DefaultStatusMonitorConfig() *StatusMonitorConfig {
	return &StatusMonitorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		RedisAddr: os.Getenv("REDIS_ADDR"),
		WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
	}
}

func LoadStatusMonitorConfig() *StatusMonitorConfig {
	config := DefaultStatusMonitorConfig()

	if v := os.Getenv("STATUS_MONITOR_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("STATUS_MONITOR_INTERVAL_SEC"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			config.Interval = time.Duration(sec) * time.Second
		}
	}

	return config
}

type StatusMonitor struct {
	config        *StatusMonitorConfig
	repo          storage.Repository
	notificationSvc *notification.Service
	redisClient  *redis.Client
	logger       *logrus.Logger
	stopChan     chan struct{}
	stopOnce     sync.Once
	restartCount int
	maxRestarts  int
}

func NewStatusMonitor(
	config *StatusMonitorConfig,
	repo storage.Repository,
	notificationSvc *notification.Service,
	redisClient *redis.Client,
	logger *logrus.Logger,
) *StatusMonitor {
	return &StatusMonitor{
		config:           config,
		repo:            repo,
		notificationSvc: notificationSvc,
		redisClient:    redisClient,
		logger:         logger,
		stopChan:       make(chan struct{}),
		maxRestarts:    5,
	}
}

func (m *StatusMonitor) Start(ctx context.Context) {
	if !m.config.Enabled {
		m.logger.Info("Status monitor is disabled")
		return
	}

	m.logger.WithFields(logrus.Fields{
		"interval": m.config.Interval,
	}).Info("Starting status monitor")

	go m.runLoop(ctx)
}

func (m *StatusMonitor) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
	m.logger.Info("Status monitor stopped")
}

func (m *StatusMonitor) runLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Status monitor stopping due to context cancellation")
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.checkComponents(ctx); err != nil {
				m.logger.WithError(err).Error("Failed to check components")
				m.handleCrash(ctx)
			}
		}
	}
}

func (m *StatusMonitor) handleCrash(ctx context.Context) {
	m.restartCount++
	if m.restartCount > m.maxRestarts {
		m.logger.Error("Status monitor exceeded max restarts, stopping")
		m.Stop()
		return
	}

	m.logger.WithField("restart_count", m.restartCount).Warn("Status monitor restarting after crash")
	time.Sleep(time.Duration(m.restartCount) * time.Second)
}

func (m *StatusMonitor) checkComponents(ctx context.Context) error {
	components, err := m.getMonitoredComponents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get monitored components: %w", err)
	}

	for _, comp := range components {
		if !comp.Enabled {
			continue
		}

		currentState := comp.Status
		lastState, err := m.getLastState(ctx, comp.ID)

		if err != nil {
			m.logger.WithError(err).WithField("component", comp.ID).Warn("Failed to get last state, dispatching anyway")
			m.dispatchNotification(ctx, comp, "", currentState)
		} else if lastState != currentState {
			m.dispatchNotification(ctx, comp, lastState, currentState)
		}

		if err := m.setLastState(ctx, comp.ID, currentState); err != nil {
			m.logger.WithError(err).WithField("component", comp.ID).Warn("Failed to set last state")
		}
	}

	return nil
}

type MonitoredComponent struct {
	ID          string
	Name        string
	Type        string
	Status      string
	Enabled     bool
	SlackChannel string
}

func (m *StatusMonitor) getMonitoredComponents(ctx context.Context) ([]MonitoredComponent, error) {
	components := []MonitoredComponent{
		{ID: "api", Name: "API", Type: "api", Status: "operational", Enabled: true},
		{ID: "database", Name: "Database", Type: "database", Status: "operational", Enabled: true},
		{ID: "cache", Name: "Cache", Type: "cache", Status: "operational", Enabled: true},
		{ID: "ai-service", Name: "AI Service", Type: "ai", Status: "operational", Enabled: true},
		{ID: "embeddings", Name: "Embeddings", Type: "ai", Status: "operational", Enabled: true},
		{ID: "state-fabric", Name: "State Fabric", Type: "storage", Status: "operational", Enabled: true},
		{ID: "microvm", Name: "MicroVM Runtime", Type: "runtime", Status: "operational", Enabled: true},
		{ID: "queue", Name: "Queue Worker", Type: "worker", Status: "operational", Enabled: true},
		{ID: "function-backup", Name: "Function Backup", Type: "backup", Status: "operational", Enabled: true},
		{ID: "email", Name: "Email Delivery", Type: "email", Status: "operational", Enabled: true},
		{ID: "billing", Name: "Billing", Type: "billing", Status: "operational", Enabled: true},
		{ID: "storage", Name: "Object Storage", Type: "storage", Status: "operational", Enabled: true},
		{ID: "cdn", Name: "CDN", Type: "cdn", Status: "operational", Enabled: true},
		{ID: "pgbouncer", Name: "Connection Pool", Type: "infrastructure", Status: "operational", Enabled: true},
		{ID: "recommendations", Name: "Recommendations", Type: "ai", Status: "operational", Enabled: true},
		{ID: "verification", Name: "Verification Pipeline", Type: "security", Status: "operational", Enabled: true},
		{ID: "trust-api", Name: "Trust API", Type: "security", Status: "operational", Enabled: true},
		{ID: "support", Name: "Support System", Type: "service", Status: "operational", Enabled: true},
		{ID: "registry", Name: "Function Registry", Type: "service", Status: "operational", Enabled: true},
		{ID: "health-monitor", Name: "Health Monitor", Type: "monitoring", Status: "operational", Enabled: true},
	}

	if m.redisClient == nil {
		return components, nil
	}

	keys, err := m.redisClient.Keys(ctx, "system:health:*").Result()
	if err != nil {
		return components, nil
	}

	for _, key := range keys {
		componentID := key[13:]
		val, err := m.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		for i := range components {
			if components[i].ID == componentID {
				components[i].Status = val
				break
			}
		}
	}

	return components, nil
}

func (m *StatusMonitor) getLastState(ctx context.Context, componentID string) (string, error) {
	if m.redisClient == nil {
		return "", fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf(statusMonitorLastStateKey, componentID)
	val, err := m.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (m *StatusMonitor) setLastState(ctx context.Context, componentID, state string) error {
	if m.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf(statusMonitorLastStateKey, componentID)
	return m.redisClient.Set(ctx, key, state, statusMonitorTTL).Err()
}

func (m *StatusMonitor) dispatchNotification(ctx context.Context, comp MonitoredComponent, oldState, newState string) {
	notificationType := m.getNotificationType(oldState, newState)
	severity := m.getSeverity(newState)
	title := fmt.Sprintf("%s — %s", comp.Name, m.getStatusTitle(newState))

	body := fmt.Sprintf("Status changed from %s to %s", oldState, newState)
	if oldState == "" {
		body = fmt.Sprintf("Current status: %s", newState)
	}

	users, err := m.getAdminUsers(ctx)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to get admin users for notification")
		return
	}

	for _, user := range users {
		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notificationType,
			Category: notification.CategoryProvider,
			Title:    title,
			Body:     body,
			Data: notification.JSONMap{
				"component_id":    comp.ID,
				"component_name": comp.Name,
				"component_type": comp.Type,
				"old_status":     oldState,
				"new_status":     newState,
				"severity":       severity,
			},
			Channels: []string{notification.ChannelSlack, notification.ChannelInApp},
			Priority: m.getPriority(severity),
		})
		if err != nil {
			m.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to send status notification")
		}
	}

	m.logger.WithFields(logrus.Fields{
		"component": comp.ID,
		"old_state": oldState,
		"new_state": newState,
	}).Info("Status change notification dispatched")
}

func (m *StatusMonitor) getNotificationType(oldState, newState string) string {
	if newState == "down" || newState == "major_outage" {
		return notification.TypeProviderOffline
	}
	if newState == "degraded" {
		return notification.TypeProviderDegraded
	}
	if oldState != "" && (newState == "operational" || newState == "healthy") {
		return notification.TypeProviderOnline
	}
	if newState == "maintenance" {
		return notification.TypeSystemMaintenance
	}
	return notification.TypeProviderDegraded
}

func (m *StatusMonitor) getSeverity(status string) string {
	switch status {
	case "major_outage", "down":
		return "critical"
	case "degraded", "slow":
		return "high"
	case "maintenance":
		return "maintenance"
	case "operational", "healthy":
		return "info"
	default:
		return "low"
	}
}

func (m *StatusMonitor) getStatusTitle(status string) string {
	switch status {
	case "major_outage", "down":
		return "Major Outage"
	case "degraded":
		return "Degraded"
	case "slow":
		return "Slow Response"
	case "maintenance":
		return "Under Maintenance"
	case "operational", "healthy":
		return "Operational"
	default:
		return "Unknown"
	}
}

func (m *StatusMonitor) getPriority(severity string) string {
	switch severity {
	case "critical":
		return notification.PriorityCritical
	case "high":
		return notification.PriorityHigh
	case "maintenance":
		return notification.PriorityNormal
	case "info":
		return notification.PriorityNormal
	default:
		return notification.PriorityLow
	}
}

func (m *StatusMonitor) getAdminUsers(ctx context.Context) ([]storage.User, error) {
	return []storage.User{}, nil
}
