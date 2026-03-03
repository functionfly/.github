package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// RealtimeMonitor provides real-time monitoring capabilities using Supabase's real-time features
type RealtimeMonitor struct {
	service           *Service
	connections       map[string]*RealtimeConnection
	dbChangeListeners map[string]chan *DatabaseChangeEvent
	mu                sync.RWMutex
}

// DatabaseChangeEvent represents a database change notification
type DatabaseChangeEvent struct {
	Schema          string                 `json:"schema"`
	Table           string                 `json:"table"`
	EventType       string                 `json:"eventType"` // INSERT, UPDATE, DELETE
	CommitTimestamp string                 `json:"commit_timestamp"`
	New             map[string]interface{} `json:"new,omitempty"`
	Old             map[string]interface{} `json:"old,omitempty"`
	IDs             []string               `json:"ids"`
	Errors          *string                `json:"errors"`
}

// RealtimeConnection represents a real-time WebSocket connection
type RealtimeConnection struct {
	ID         string
	UserID     *uuid.UUID
	TenantID   *uuid.UUID
	Conn       *websocket.Conn
	SendChan   chan []byte
	Subscribed map[string]bool // Channels this connection is subscribed to
	mu         sync.Mutex
}

// NewRealtimeMonitor creates a new real-time monitor
func NewRealtimeMonitor(service *Service) *RealtimeMonitor {
	rtm := &RealtimeMonitor{
		service:           service,
		connections:       make(map[string]*RealtimeConnection),
		dbChangeListeners: make(map[string]chan *DatabaseChangeEvent),
	}

	// Start database change listeners for monitored tables
	rtm.startDatabaseChangeListeners()

	return rtm
}

// HandleRealtimeConnection handles a new WebSocket connection for real-time monitoring
func (rtm *RealtimeMonitor) HandleRealtimeConnection(ctx context.Context, conn *websocket.Conn, userID, tenantID *uuid.UUID) {
	connectionID := uuid.New().String()

	rtConn := &RealtimeConnection{
		ID:         connectionID,
		UserID:     userID,
		TenantID:   tenantID,
		Conn:       conn,
		SendChan:   make(chan []byte, 256),
		Subscribed: make(map[string]bool),
	}

	// Register connection
	rtm.mu.Lock()
	rtm.connections[connectionID] = rtConn
	rtm.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"connection_id": connectionID,
		"user_id":       userID,
		"tenant_id":     tenantID,
	}).Info("Real-time monitoring connection established")

	// Start goroutines for handling the connection
	go rtConn.writePump()
	go rtConn.readPump(rtm)

	// Subscribe to default channels
	rtm.subscribeToDefaultChannels(rtConn)

	// Send connection established message to client
	rtConn.sendJSON(map[string]interface{}{
		"type": "connection_established",
	})
}

// subscribeToDefaultChannels subscribes a connection to default monitoring channels
func (rtm *RealtimeMonitor) subscribeToDefaultChannels(rtConn *RealtimeConnection) {
	// Subscribe to global monitoring events
	rtm.subscribeConnection(rtConn, "monitoring_events")
	rtm.subscribeConnection(rtConn, "monitoring_alerts")

	// Subscribe to deployment and function execution channels
	rtm.subscribeConnection(rtConn, "deployments")
	rtm.subscribeConnection(rtConn, "function_executions")
	rtm.subscribeConnection(rtConn, "registry_updates")

	// Subscribe to tenant-specific channels if tenant ID is provided
	if rtConn.TenantID != nil {
		tenantEventsChannel := fmt.Sprintf("tenant_%s_events", rtConn.TenantID.String())
		tenantAlertsChannel := fmt.Sprintf("tenant_%s_alerts", rtConn.TenantID.String())
		tenantDeploymentsChannel := fmt.Sprintf("tenant_%s_deployments", rtConn.TenantID.String())
		tenantExecutionsChannel := fmt.Sprintf("tenant_%s_executions", rtConn.TenantID.String())
		tenantRegistryChannel := fmt.Sprintf("tenant_%s_registry", rtConn.TenantID.String())

		rtm.subscribeConnection(rtConn, tenantEventsChannel)
		rtm.subscribeConnection(rtConn, tenantAlertsChannel)
		rtm.subscribeConnection(rtConn, tenantDeploymentsChannel)
		rtm.subscribeConnection(rtConn, tenantExecutionsChannel)
		rtm.subscribeConnection(rtConn, tenantRegistryChannel)
	}

	// Subscribe to database change channels for real-time data updates
	// These channels receive notifications when database tables are modified
	databaseTables := []string{
		"users", "apps", "backends", "alerts",
		"performance_metrics", "monitoring_events", "user_notifications",
		"functions", "deployments", "function_executions",
	}

	for _, table := range databaseTables {
		channelName := "db_changes_" + table
		rtm.subscribeConnection(rtConn, channelName)

		// Also subscribe to tenant-specific database change channels
		if rtConn.TenantID != nil {
			tenantChannelName := fmt.Sprintf("db_changes_%s_%s", table, rtConn.TenantID.String())
			rtm.subscribeConnection(rtConn, tenantChannelName)
		}
	}

	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type":      "connection_established",
		"timestamp": time.Now(),
		"channels":  rtConn.Subscribed,
	}

	rtConn.sendJSON(welcomeMsg)
}

// subscribeConnection subscribes a connection to a channel
func (rtm *RealtimeMonitor) subscribeConnection(rtConn *RealtimeConnection, channel string) {
	rtConn.mu.Lock()
	rtConn.Subscribed[channel] = true
	rtConn.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"connection_id": rtConn.ID,
		"channel":       channel,
	}).Debug("Connection subscribed to channel")
}

// unsubscribeConnection unsubscribes a connection from a channel
func (rtm *RealtimeMonitor) unsubscribeConnection(rtConn *RealtimeConnection, channel string) {
	rtConn.mu.Lock()
	delete(rtConn.Subscribed, channel)
	rtConn.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"connection_id": rtConn.ID,
		"channel":       channel,
	}).Debug("Connection unsubscribed from channel")
}

// BroadcastToChannel broadcasts a message to all connections subscribed to a channel
func (rtm *RealtimeMonitor) BroadcastToChannel(channel string, message interface{}) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal broadcast message")
		return
	}

	rtm.mu.RLock()
	defer rtm.mu.RUnlock()

	broadcastCount := 0
	for _, rtConn := range rtm.connections {
		rtConn.mu.Lock()
		if rtConn.Subscribed[channel] {
			select {
			case rtConn.SendChan <- messageBytes:
				broadcastCount++
			default:
				// Channel is full, connection might be slow
				logrus.WithField("connection_id", rtConn.ID).Warn("Connection send channel full")
			}
		}
		rtConn.mu.Unlock()
	}

	logrus.WithFields(logrus.Fields{
		"channel":           channel,
		"broadcast_count":   broadcastCount,
		"total_connections": len(rtm.connections),
	}).Debug("Message broadcasted to channel")
}

// RemoveConnection removes a connection when it disconnects
func (rtm *RealtimeMonitor) RemoveConnection(connectionID string) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()

	if rtConn, exists := rtm.connections[connectionID]; exists {
		rtConn.mu.Lock()
		close(rtConn.SendChan)
		rtConn.mu.Unlock()
		delete(rtm.connections, connectionID)

		logrus.WithField("connection_id", connectionID).Info("Real-time monitoring connection removed")
	}
}

// GetConnectionStats returns statistics about active connections
func (rtm *RealtimeMonitor) GetConnectionStats() map[string]interface{} {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()

	totalConnections := len(rtm.connections)
	channelSubscriptions := make(map[string]int)

	for _, rtConn := range rtm.connections {
		rtConn.mu.Lock()
		for channel := range rtConn.Subscribed {
			channelSubscriptions[channel]++
		}
		rtConn.mu.Unlock()
	}

	return map[string]interface{}{
		"total_connections":     totalConnections,
		"channel_subscriptions": channelSubscriptions,
		"timestamp":             time.Now(),
	}
}

// writePump handles writing messages to the WebSocket connection
func (rtConn *RealtimeConnection) writePump() {
	defer func() {
		rtConn.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-rtConn.SendChan:
			if !ok {
				rtConn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := rtConn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logrus.WithError(err).WithField("connection_id", rtConn.ID).Error("Failed to write to WebSocket")
				return
			}
		}
	}
}

// readPump handles reading messages from the WebSocket connection
func (rtConn *RealtimeConnection) readPump(rtm *RealtimeMonitor) {
	defer func() {
		rtm.RemoveConnection(rtConn.ID)
		rtConn.Conn.Close()
	}()

	// Set read deadline and pong handler
	rtConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	rtConn.Conn.SetPongHandler(func(string) error {
		rtConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := rtConn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).WithField("connection_id", rtConn.ID).Error("WebSocket error")
			}
			break
		}

		// Handle client messages (e.g., subscribe/unsubscribe requests)
		rtConn.handleClientMessage(rtm, message)
	}
}

// handleClientMessage processes messages from the client
func (rtConn *RealtimeConnection) handleClientMessage(rtm *RealtimeMonitor, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		logrus.WithError(err).WithField("connection_id", rtConn.ID).Warn("Invalid client message")
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		logrus.WithField("connection_id", rtConn.ID).Warn("Client message missing type")
		return
	}

	switch msgType {
	case "subscribe":
		if channel, ok := msg["channel"].(string); ok {
			rtm.subscribeConnection(rtConn, channel)
			rtConn.sendJSON(map[string]interface{}{
				"type":    "subscribed",
				"channel": channel,
			})
		}
	case "unsubscribe":
		if channel, ok := msg["channel"].(string); ok {
			rtm.unsubscribeConnection(rtConn, channel)
			rtConn.sendJSON(map[string]interface{}{
				"type":    "unsubscribed",
				"channel": channel,
			})
		}
	case "ping":
		// Extend read deadline when ping is received (like pong handler)
		rtConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		rtConn.sendJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now(),
		})
	default:
		logrus.WithFields(logrus.Fields{
			"connection_id": rtConn.ID,
			"message_type":  msgType,
		}).Warn("Unknown client message type")
	}
}

// sendJSON sends a JSON message to the client
func (rtConn *RealtimeConnection) sendJSON(data interface{}) {
	messageBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).WithField("connection_id", rtConn.ID).Error("Failed to marshal message")
		return
	}

	select {
	case rtConn.SendChan <- messageBytes:
	default:
		logrus.WithField("connection_id", rtConn.ID).Warn("Send channel full")
	}
}

// startDatabaseChangeListeners starts listeners for database change notifications
func (rtm *RealtimeMonitor) startDatabaseChangeListeners() {
	// Tables to monitor for changes
	tables := []string{
		"users",
		"apps",
		"backends",
		"alerts",
		"performance_metrics",
		"monitoring_events",
		"user_notifications",
	}

	for _, table := range tables {
		channelName := "db_changes_" + table
		listenerChan := make(chan *DatabaseChangeEvent, 100)

		rtm.dbChangeListeners[channelName] = listenerChan

		// Start goroutine to listen for notifications and broadcast them
		go rtm.listenAndBroadcastDatabaseChanges(channelName, listenerChan)
	}
}

// listenAndBroadcastDatabaseChanges listens for database changes and broadcasts them to WebSocket clients
func (rtm *RealtimeMonitor) listenAndBroadcastDatabaseChanges(channelName string, listenerChan chan *DatabaseChangeEvent) {
	logrus.WithField("channel", channelName).Info("Starting database change listener and broadcaster")

	ctx := context.Background()

	for {
		// Listen for notifications from the database
		notification, err := rtm.service.ListenForNotification(ctx, channelName)
		if err != nil {
			logrus.WithError(err).WithField("channel", channelName).Error("Failed to listen for database notification")
			time.Sleep(5 * time.Second) // Retry after delay
			continue
		}

		// Skip if no notification
		if notification == "" {
			time.Sleep(1 * time.Second) // Small delay before checking again
			continue
		}

		// Parse the notification payload
		var dbEvent DatabaseChangeEvent
		if err := json.Unmarshal([]byte(notification), &dbEvent); err != nil {
			logrus.WithError(err).WithField("channel", channelName).Error("Failed to parse database change event")
			continue
		}

		// Broadcast the database change to all WebSocket clients subscribed to this table
		rtm.BroadcastDatabaseChange(&dbEvent)
	}
}

// ListenForNotification listens for a PostgreSQL notification on the specified channel
func (s *Service) ListenForNotification(ctx context.Context, channel string) (string, error) {
	// Start listening on the channel
	if err := s.db.PgListen(ctx, channel); err != nil {
		return "", fmt.Errorf("failed to listen on channel %s: %w", channel, err)
	}

	// Wait for notification
	notification, err := s.db.PgWaitForNotification(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for notification on channel %s: %w", channel, err)
	}

	if notification != nil {
		return notification.Payload, nil
	}

	return "", nil // No notification received
}

// SubscribeToDatabaseChanges allows subscribing to database changes for a specific table
// Note: Database changes are automatically broadcast to WebSocket clients.
// This method is for programmatic subscriptions if needed in the future.
func (rtm *RealtimeMonitor) SubscribeToDatabaseChanges(table string, tenantID *uuid.UUID, callback func(*DatabaseChangeEvent)) {
	channelName := "db_changes_" + table
	if tenantID != nil {
		channelName = fmt.Sprintf("db_changes_%s_%s", table, tenantID.String())
	}

	// For now, database changes are handled automatically through WebSocket broadcasting
	// This method is kept for future programmatic subscription needs
	logrus.WithFields(logrus.Fields{
		"table":     table,
		"tenant_id": tenantID,
		"channel":   channelName,
	}).Info("Database change subscription requested (WebSocket broadcasting is automatic)")

	// If programmatic callback is needed in the future, it could be implemented here
	_ = callback // Suppress unused parameter warning
}

// BroadcastDatabaseChange broadcasts a database change event to all subscribed WebSocket connections
func (rtm *RealtimeMonitor) BroadcastDatabaseChange(event *DatabaseChangeEvent) {
	// Create the broadcast payload in the format expected by the frontend
	payload := map[string]interface{}{
		"type":      "broadcast",
		"event":     "db_change",
		"payload":   event,
		"channel":   fmt.Sprintf("db_changes_%s", event.Table),
		"timestamp": time.Now(),
	}

	rtm.BroadcastToChannel(fmt.Sprintf("db_changes_%s", event.Table), payload)
}

// BroadcastDeploymentUpdate broadcasts deployment status updates
func (rtm *RealtimeMonitor) BroadcastDeploymentUpdate(tenantID *uuid.UUID, deploymentID uuid.UUID, status string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"type":          "broadcast",
		"event":         "deployment_update",
		"deployment_id": deploymentID,
		"status":        status,
		"details":       details,
		"timestamp":     time.Now(),
	}

	// Broadcast to global deployments channel
	rtm.BroadcastToChannel("deployments", payload)

	// Broadcast to tenant-specific channel if tenant ID provided
	if tenantID != nil {
		tenantChannel := fmt.Sprintf("tenant_%s_deployments", tenantID.String())
		rtm.BroadcastToChannel(tenantChannel, payload)
	}
}

// BroadcastFunctionExecution broadcasts function execution events
func (rtm *RealtimeMonitor) BroadcastFunctionExecution(tenantID *uuid.UUID, functionID uuid.UUID, executionID uuid.UUID, eventType string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"type":         "broadcast",
		"event":        "function_execution",
		"function_id":  functionID,
		"execution_id": executionID,
		"event_type":   eventType, // "started", "completed", "failed", "log"
		"details":      details,
		"timestamp":    time.Now(),
	}

	// Broadcast to global function executions channel
	rtm.BroadcastToChannel("function_executions", payload)

	// Broadcast to tenant-specific channel if tenant ID provided
	if tenantID != nil {
		tenantChannel := fmt.Sprintf("tenant_%s_executions", tenantID.String())
		rtm.BroadcastToChannel(tenantChannel, payload)
	}
}

// BroadcastRegistryUpdate broadcasts registry function updates (ratings, popularity, etc.)
func (rtm *RealtimeMonitor) BroadcastRegistryUpdate(functionID uuid.UUID, updateType string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"type":        "broadcast",
		"event":       "registry_update",
		"function_id": functionID,
		"update_type": updateType, // "rating", "popularity", "download", "new_version"
		"details":     details,
		"timestamp":   time.Now(),
	}

	rtm.BroadcastToChannel("registry_updates", payload)
}

// BroadcastTeamUpdate broadcasts team-related updates (member changes, permissions, etc.)
func (rtm *RealtimeMonitor) BroadcastTeamUpdate(tenantID uuid.UUID, eventType string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"type":       "broadcast",
		"event":      "team_update",
		"event_type": eventType, // "member_added", "member_removed", "role_changed", "permissions_updated"
		"details":    details,
		"timestamp":  time.Now(),
	}

	tenantChannel := fmt.Sprintf("tenant_%s_team", tenantID.String())
	rtm.BroadcastToChannel(tenantChannel, payload)
}
