package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// InAppChannel handles in-app notification delivery using pg_notify
type InAppChannel struct {
	repo   Repository
	db     *storage.PostgresDB
	logger *logrus.Logger
}

// NewInAppChannel creates a new in-app channel
func NewInAppChannel(repo Repository, db *storage.PostgresDB, logger *logrus.Logger) *InAppChannel {
	return &InAppChannel{
		repo:   repo,
		db:     db,
		logger: logger,
	}
}

// Name returns the channel name
func (c *InAppChannel) Name() string {
	return ChannelInApp
}

// Send sends a notification via in-app (pg_notify)
func (c *InAppChannel) Send(ctx context.Context, n *Notification, user *storage.User) error {
	// Notification is already stored in database by the service
	// We just need to trigger pg_notify for real-time updates

	payload, err := json.Marshal(map[string]interface{}{
		"type":            "notification",
		"user_id":         n.UserID,
		"notification_id": n.ID,
		"title":           n.Title,
		"body":            n.Body,
		"category":        n.Category,
		"priority":        n.Priority,
		"created_at":      n.CreatedAt.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	// Use pg_notify to broadcast to listening clients
	// The channel name includes user_id for targeted delivery
	channelName := fmt.Sprintf("user_notifications_%s", n.UserID.String())

	if err := c.pgNotify(channelName, string(payload)); err != nil {
		c.logger.WithError(err).Warn("Failed to send pg_notify, but notification is stored in database")
		// Don't return error here - the notification is already saved in DB
		// The client will pick it up on next poll
	}

	c.logger.WithFields(logrus.Fields{
		"notification_id": n.ID,
		"user_id":         n.UserID,
	}).Debug("In-app notification sent via pg_notify")

	return nil
}

// IsConfigured returns whether the channel is configured
func (c *InAppChannel) IsConfigured() bool {
	return c.db != nil
}

// pgNotify sends a PostgreSQL notification
func (c *InAppChannel) pgNotify(channel, payload string) error {
	// Use the underlying sql.DB to send NOTIFY
	// This requires direct SQL execution
	query := "SELECT pg_notify($1, $2)"
	if _, err := c.db.Exec(query, channel, payload); err != nil {
		return err
	}
	// Also fire the global broadcast channel so the WebSocket notification
	// listener (which subscribes to "notification_broadcast") receives the event.
	// The payload's "user_id" field is used by the listener to route to the
	// correct WebSocket client(s).
	if _, err := c.db.Exec(query, "notification_broadcast", payload); err != nil {
		c.logger.WithError(err).Warn("Failed to send notification_broadcast pg_notify")
	}
	return nil
}

// InAppMessage represents a message sent via the in-app channel
type InAppMessage struct {
	Type           string    `json:"type"`
	UserID         string    `json:"user_id"`
	NotificationID string    `json:"notification_id"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Category       string    `json:"category"`
	Priority       string    `json:"priority"`
	CreatedAt      time.Time `json:"created_at"`
}

// WebSocketManager manages WebSocket connections for real-time notifications
type WebSocketManager struct {
	clients    map[string][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *InAppMessage
	mu         sync.RWMutex
	logger     *logrus.Logger
}

// Client represents a WebSocket client
type Client struct {
	UserID string
	Conn   interface{} // Would be *websocket.Conn in actual implementation
	Send   chan []byte
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *logrus.Logger) *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *InAppMessage, 100),
		logger:     logger,
	}
}

// Run starts the WebSocket manager
func (m *WebSocketManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.UserID] = append(m.clients[client.UserID], client)
			m.mu.Unlock()
			m.logger.WithField("user_id", client.UserID).Debug("WebSocket client registered")

		case client := <-m.unregister:
			m.mu.Lock()
			if clients, ok := m.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						m.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(m.clients[client.UserID]) == 0 {
					delete(m.clients, client.UserID)
				}
			}
			m.mu.Unlock()
			close(client.Send)
			m.logger.WithField("user_id", client.UserID).Debug("WebSocket client unregistered")

		case message := <-m.broadcast:
			m.mu.RLock()
			clients := m.clients[message.UserID]
			m.mu.RUnlock()

			data, err := json.Marshal(message)
			if err != nil {
				m.logger.WithError(err).Error("Failed to marshal message")
				continue
			}

			for _, client := range clients {
				select {
				case client.Send <- data:
				default:
					// Client buffer full, close connection
					m.unregister <- client
				}
			}
		}
	}
}

