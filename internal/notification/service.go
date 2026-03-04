package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service is the main notification service
type Service struct {
	repo       Repository
	channels   map[string]Channel
	templates  *TemplateEngine
	queue      *Queue
	dispatcher *Dispatcher
	logger     *logrus.Logger
	db         *storage.PostgresDB
	mu         sync.RWMutex
}

// NewService creates a new notification service
func NewService(repo Repository, db *storage.PostgresDB, emailSvc email.Service, logger *logrus.Logger) *Service {
	s := &Service{
		repo:     repo,
		db:       db,
		logger:   logger,
		channels: make(map[string]Channel),
	}

	// Register channels
	emailChannel := NewEmailChannel(emailSvc, logger)
	inAppChannel := NewInAppChannel(repo, db, logger)
	webhookChannel := NewWebhookChannel(logger)
	webhookChannel.SetRepository(repo)

	s.channels[ChannelEmail] = emailChannel
	s.channels[ChannelInApp] = inAppChannel
	s.channels[ChannelWebhook] = webhookChannel

	// Initialize queue and dispatcher
	s.queue = NewQueue(repo, logger)
	s.dispatcher = NewDispatcher(s.channels, repo, logger)
	s.templates = NewTemplateEngine(repo)

	return s
}

// Send creates and sends a notification
func (s *Service) Send(ctx context.Context, req SendRequest) (*Notification, error) {
	// Determine channels if not specified
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{ChannelInApp, ChannelEmail}
	}

	// Determine priority if not specified
	priority := req.Priority
	if priority == "" {
		priority = PriorityNormal
	}

	// Build notification from request
	notification := &Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Category: req.Category,
		Title:    req.Title,
		Body:     req.Body,
		Data:     req.Data,
		Channels: channels,
		Priority: priority,
		Status:   StatusPending,
	}

	// Save to database
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"type":            notification.Type,
	}).Info("Notification created")

	// Add to processing queue
	s.queue.Enqueue(notification)

	return notification, nil
}

// Broadcast sends a notification to multiple users
func (s *Service) Broadcast(ctx context.Context, req BroadcastRequest) error {
	for _, userID := range req.UserIDs {
		sendReq := SendRequest{
			UserID:   userID,
			Type:     req.Type,
			Category: req.Category,
			Title:    req.Title,
			Body:     req.Body,
			Data:     req.Data,
			Channels: req.Channels,
			Priority: req.Priority,
		}
		if _, err := s.Send(ctx, sendReq); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send broadcast notification")
		}
	}
	return nil
}

// SendWelcome sends a welcome notification to a new user (in-app only so everyone sees it when they open the app).
func (s *Service) SendWelcome(ctx context.Context, userID uuid.UUID) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeWelcome,
		Category: CategorySystem,
		Title:    "Welcome to FunctionFly",
		Body:     "We're glad you're here. Deploy your first function or explore the docs to get started.",
		Channels: []string{ChannelInApp},
		Priority: PriorityNormal,
	})
	return err
}

// GetUnreadCount returns unread notification count for user
func (s *Service) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// GetUnreadCountsByCategory returns unread notification counts grouped by category
func (s *Service) GetUnreadCountsByCategory(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	return s.repo.GetUnreadCountsByCategory(ctx, userID)
}

// GetTotalCount returns total notification count for user
func (s *Service) GetTotalCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetTotalCount(ctx, userID)
}

// ListNotifications lists notifications for a user
func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error) {
	return s.repo.ListNotifications(ctx, userID, opts)
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
	return s.repo.GetNotification(ctx, id)
}

// DeleteNotification deletes a notification
func (s *Service) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteNotification(ctx, id)
}

// GetPreferences retrieves all notification preferences for a user
func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If no preferences exist, create defaults
	if len(prefs) == 0 {
		if err := s.repo.CreateDefaultPreferences(ctx, userID); err != nil {
			return nil, err
		}
		return s.repo.GetPreferences(ctx, userID)
	}

	return prefs, nil
}

// SavePreference saves a notification preference
func (s *Service) SavePreference(ctx context.Context, pref *NotificationPreference) error {
	return s.repo.SavePreference(ctx, pref)
}

// RegisterChannel registers a custom channel
func (s *Service) RegisterChannel(name string, channel Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[name] = channel
}

// Start begins processing the notification queue
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting notification service")
	s.queue.Start(ctx, s.dispatcher)
}

// Stop stops the notification service
func (s *Service) Stop() {
	s.logger.Info("Stopping notification service")
	s.queue.Stop()
}

// TemplateEngine handles template rendering
type TemplateEngine struct {
	repo Repository
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(repo Repository) *TemplateEngine {
	return &TemplateEngine{repo: repo}
}

// Render renders a template with data
func (e *TemplateEngine) Render(ctx context.Context, notificationType, channel string, data JSONMap) (subject, bodyHTML, bodyText string, err error) {
	template, err := e.repo.GetTemplate(ctx, notificationType, channel)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		// Fallback: return empty strings, let channel handle default rendering
		return "", "", "", nil
	}

	// Simple variable substitution
	subject = template.Subject
	bodyHTML = template.BodyHTML
	bodyText = template.BodyText

	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		strValue := fmt.Sprintf("%v", value)
		subject = replaceAll(subject, placeholder, strValue)
		bodyHTML = replaceAll(bodyHTML, placeholder, strValue)
		bodyText = replaceAll(bodyText, placeholder, strValue)
	}

	return subject, bodyHTML, bodyText, nil
}

// replaceAll is a simple string replacement helper
func replaceAll(s, old, new string) string {
	// Simple implementation - in production you might want to use a proper template engine
	result := ""
	for {
		idx := 0
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == 0 && (len(s) < len(old) || s[:len(old)] != old) {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

// Queue manages the notification processing queue
type Queue struct {
	repo    Repository
	logger  *logrus.Logger
	queue   chan *Notification
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	workers int
}

// NewQueue creates a new notification queue
func NewQueue(repo Repository, logger *logrus.Logger) *Queue {
	return &Queue{
		repo:    repo,
		logger:  logger,
		queue:   make(chan *Notification, 1000),
		workers: 5,
	}
}

// Enqueue adds a notification to the queue
func (q *Queue) Enqueue(n *Notification) {
	select {
	case q.queue <- n:
		q.logger.WithField("notification_id", n.ID).Debug("Notification queued")
	default:
		q.logger.WithField("notification_id", n.ID).Warn("Notification queue full, dropping notification")
	}
}

// Start begins processing the queue
func (q *Queue) Start(ctx context.Context, dispatcher *Dispatcher) {
	q.ctx, q.cancel = context.WithCancel(ctx)

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(dispatcher)
	}

	q.logger.WithField("workers", q.workers).Info("Notification queue started")
}

// Stop stops the queue processing
func (q *Queue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	close(q.queue)
	q.wg.Wait()
	q.logger.Info("Notification queue stopped")
}

// worker processes notifications from the queue
func (q *Queue) worker(dispatcher *Dispatcher) {
	defer q.wg.Done()

	for {
		select {
		case n, ok := <-q.queue:
			if !ok {
				return
			}
			if err := dispatcher.Dispatch(q.ctx, n); err != nil {
				q.logger.WithError(err).WithField("notification_id", n.ID).Error("Failed to dispatch notification")
			}
		case <-q.ctx.Done():
			return
		}
	}
}

// Dispatcher handles notification delivery to channels
type Dispatcher struct {
	channels map[string]Channel
	repo     Repository
	logger   *logrus.Logger
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(channels map[string]Channel, repo Repository, logger *logrus.Logger) *Dispatcher {
	return &Dispatcher{
		channels: channels,
		repo:     repo,
		logger:   logger,
	}
}

// Dispatch sends a notification through configured channels
func (d *Dispatcher) Dispatch(ctx context.Context, n *Notification) error {
	// Update status to processing
	if err := d.repo.UpdateNotificationStatus(ctx, n.ID, StatusProcessing); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	// Get user from storage
	// Note: We need to get the user to pass to channels
	// For now, we'll create a minimal user struct with just the ID
	user := &storage.User{
		ID: n.UserID,
	}

	successCount := 0
	failedCount := 0

	for _, channelName := range n.Channels {
		channel, ok := d.channels[channelName]
		if !ok {
			d.logger.WithField("channel", channelName).Warn("Unknown notification channel")
			continue
		}

		if !channel.IsConfigured() {
			d.logger.WithField("channel", channelName).Debug("Channel not configured")
			continue
		}

		if err := channel.Send(ctx, n, user); err != nil {
			d.logger.WithError(err).WithFields(logrus.Fields{
				"notification_id": n.ID,
				"channel":         channelName,
			}).Error("Failed to send notification")

			// Track failure analytics
			analytics := &NotificationAnalytics{
				NotificationID: n.ID,
				Channel:        channelName,
				Status:         AnalyticsStatusFailed,
				ErrorMessage:   strPtr(err.Error()),
			}
			d.repo.TrackAnalytics(ctx, analytics)
			failedCount++
		} else {
			// Track success analytics
			now := time.Now()
			analytics := &NotificationAnalytics{
				NotificationID: n.ID,
				Channel:        channelName,
				Status:         AnalyticsStatusDelivered,
				DeliveredAt:    &now,
			}
			d.repo.TrackAnalytics(ctx, analytics)
			successCount++
		}
	}

	// Update final status
	finalStatus := StatusSent
	if failedCount > 0 && successCount == 0 {
		finalStatus = StatusFailed
	}

	if err := d.repo.UpdateNotificationStatus(ctx, n.ID, finalStatus); err != nil {
		return fmt.Errorf("failed to update final notification status: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"notification_id": n.ID,
		"success_count":   successCount,
		"failed_count":    failedCount,
		"final_status":    finalStatus,
	}).Info("Notification dispatched")

	return nil
}

// strPtr is a helper to create a string pointer
func strPtr(s string) *string {
	return &s
}
