package trustapi

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// TrustDeltaNotifier manages trust score change notifications via WebSocket
type TrustDeltaNotifier struct {
	repo            *registry.RegistryRepository
	logger          *logrus.Logger
	hub             *TrustDeltaHub
	thresholdConfig registry.TrustScoreThresholdConfig
	cooldownManager *CooldownManager
	webhookService  WebhookServiceInterface
}

// TrustDeltaHub manages WebSocket connections for trust delta notifications
type TrustDeltaHub struct {
	clients    map[string][]*TrustDeltaClient // user_id -> clients
	register   chan *TrustDeltaClient
	unregister chan *TrustDeltaClient
	broadcast  chan *TrustDeltaNotification
	logger     *logrus.Logger
	mu         sync.RWMutex
}

// TrustDeltaClient represents a connected WebSocket client
type TrustDeltaClient struct {
	UserID      string
	FunctionIDs []string // Functions being watched (empty = all)
	Conn        *websocket.Conn
	Send        chan []byte
	hub         *TrustDeltaHub
}

// TrustDeltaNotification represents a trust score change notification
type TrustDeltaNotification struct {
	UserID        string                 `json:"-"`    // Routing only
	Type          string                 `json:"type"` // "trust_delta", "threshold_breach", "tier_change"
	FunctionID    string                 `json:"function_id"`
	FunctionName  string                 `json:"function_name,omitempty"`
	PreviousScore float64                `json:"previous_score"`
	CurrentScore  float64                `json:"current_score"`
	ScoreChange   float64                `json:"score_change"`
	PreviousTier  string                 `json:"previous_tier"`
	CurrentTier   string                 `json:"current_tier"`
	Severity      string                 `json:"severity"` // "info", "warning", "critical"
	Message       string                 `json:"message"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// CooldownManager tracks last notification times to prevent spam
type CooldownManager struct {
	lastNotified map[string]time.Time // function_id -> last notification time
	cooldown     time.Duration
	mu           sync.RWMutex
}

// WebhookServiceInterface defines webhook service methods needed by notifier
type WebhookServiceInterface interface {
	DeliverEvent(eventType string, payload map[string]interface{}) error
}

// WebSocket upgrader for trust delta connections
var trustDeltaUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

// NewCooldownManager creates a new cooldown manager
func NewCooldownManager(cooldown time.Duration) *CooldownManager {
	return &CooldownManager{
		lastNotified: make(map[string]time.Time),
		cooldown:     cooldown,
	}
}

// CanNotify checks if enough time has passed since the last notification for a function
func (cm *CooldownManager) CanNotify(functionID string) bool {
	cm.mu.RLock()
	last, ok := cm.lastNotified[functionID]
	cm.mu.RUnlock()

	if !ok {
		return true
	}

	return time.Since(last) >= cm.cooldown
}

// RecordNotification records that a notification was sent for a function
func (cm *CooldownManager) RecordNotification(functionID string) {
	cm.mu.Lock()
	cm.lastNotified[functionID] = time.Now()
	cm.mu.Unlock()
}

// NewTrustDeltaHub creates a new WebSocket hub for trust delta notifications
func NewTrustDeltaHub(logger *logrus.Logger) *TrustDeltaHub {
	return &TrustDeltaHub{
		clients:    make(map[string][]*TrustDeltaClient),
		register:   make(chan *TrustDeltaClient),
		unregister: make(chan *TrustDeltaClient),
		broadcast:  make(chan *TrustDeltaNotification, 100),
		logger:     logger,
	}
}

// Run starts the trust delta hub
func (h *TrustDeltaHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()
			h.logger.WithField("user_id", client.UserID).Debug("Trust delta client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()
			close(client.Send)
			h.logger.WithField("user_id", client.UserID).Debug("Trust delta client unregistered")

		case notification := <-h.broadcast:
			h.broadcastToClients(notification)
		}
	}
}

// broadcastToClients sends a notification to relevant clients
func (h *TrustDeltaHub) broadcastToClients(notification *TrustDeltaNotification) {
	data, err := json.Marshal(notification)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal trust delta notification")
		return
	}

	h.mu.RLock()
	clients := h.clients[notification.UserID]
	h.mu.RUnlock()

	for _, client := range clients {
		// Check if client is watching this function
		if len(client.FunctionIDs) > 0 {
			found := false
			for _, fid := range client.FunctionIDs {
				if fid == notification.FunctionID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		select {
		case client.Send <- data:
		default:
			// Client buffer full, close connection
			h.unregister <- client
			client.Conn.Close()
		}
	}
}

// Broadcast sends a notification to all connected clients for a user
func (h *TrustDeltaHub) Broadcast(userID string, notification *TrustDeltaNotification) {
	notification.UserID = userID
	select {
	case h.broadcast <- notification:
	default:
		h.logger.Warn("Trust delta broadcast channel full, dropping notification")
	}
}

// NewTrustDeltaNotifier creates a new trust delta notifier
func NewTrustDeltaNotifier(repo *registry.RegistryRepository, webhookService WebhookServiceInterface, logger *logrus.Logger) *TrustDeltaNotifier {
	hub := NewTrustDeltaHub(logger)
	go hub.Run()

	return &TrustDeltaNotifier{
		repo:            repo,
		logger:          logger,
		hub:             hub,
		thresholdConfig: registry.DefaultThresholdConfig(),
		cooldownManager: NewCooldownManager(registry.DefaultThresholdConfig().CooldownPeriod),
		webhookService:  webhookService,
	}
}

// SetThresholdConfig updates the threshold configuration
func (n *TrustDeltaNotifier) SetThresholdConfig(config registry.TrustScoreThresholdConfig) {
	n.thresholdConfig = config
	n.cooldownManager.cooldown = config.CooldownPeriod
}

// HandleWebSocket handles WebSocket connections for trust delta notifications
func (n *TrustDeltaNotifier) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	functionIDs := r.URL.Query()["function_id"]

	// Upgrade connection
	conn, err := trustDeltaUpgrader.Upgrade(w, r, nil)
	if err != nil {
		n.logger.WithError(err).Error("Failed to upgrade WebSocket connection for trust deltas")
		return
	}

	client := &TrustDeltaClient{
		UserID:      user.UserID.String(),
		FunctionIDs: functionIDs,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		hub:         n.hub,
	}

	// Register client
	client.hub.register <- client

	// Start pumps
	go client.writePump()
	go client.readPump()

	// Send welcome message
	welcome := map[string]interface{}{
		"type":     "connected",
		"message":  "Connected to trust delta stream",
		"watching": functionIDs,
	}
	data, _ := json.Marshal(welcome)
	client.Send <- data

	n.logger.WithField("user_id", user.UserID).Info("Trust delta WebSocket connection established")
}

// ProcessTrustDelta processes a trust score delta and sends notifications
func (n *TrustDeltaNotifier) ProcessTrustDelta(delta *registry.TrustScoreDelta, functionOwnerID uuid.UUID, functionName string) {
	// Determine severity and notification type
	severity, notifType, message := n.determineNotification(delta)

	// Skip if below minimum change threshold and not significant
	if abs(delta.ScoreChange) < n.thresholdConfig.MinChangeForNotify &&
		!delta.TierChanged &&
		severity != "critical" {
		return
	}

	// Check cooldown
	if !n.cooldownManager.CanNotify(delta.FunctionID.String()) && severity != "critical" {
		return
	}

	notification := &TrustDeltaNotification{
		UserID:        functionOwnerID.String(),
		Type:          notifType,
		FunctionID:    delta.FunctionID.String(),
		FunctionName:  functionName,
		PreviousScore: delta.PreviousScore,
		CurrentScore:  delta.CurrentScore,
		ScoreChange:   delta.ScoreChange,
		PreviousTier:  string(delta.PreviousTier),
		CurrentTier:   string(delta.CurrentTier),
		Severity:      severity,
		Message:       message,
		Timestamp:     delta.CalculatedAt,
		Data: map[string]interface{}{
			"score_change_percent": delta.ScoreChangePercent,
			"tier_changed":         delta.TierChanged,
			"window_type":          delta.WindowType,
			"component_changes":    delta.ComponentChanges,
		},
	}

	// Broadcast via WebSocket
	n.hub.Broadcast(functionOwnerID.String(), notification)

	// Deliver webhooks for critical changes
	if severity == "critical" || delta.TierChanged {
		n.deliverWebhook(notification)
	}

	// Record cooldown
	n.cooldownManager.RecordNotification(delta.FunctionID.String())
}

// determineNotification determines notification type, severity, and message
func (n *TrustDeltaNotifier) determineNotification(delta *registry.TrustScoreDelta) (severity, notifType, message string) {
	// Check critical threshold
	if delta.CurrentScore < n.thresholdConfig.CriticalThreshold {
		return "critical", "threshold_breach",
			"Trust score dropped below critical threshold - immediate attention required"
	}

	// Check warning threshold
	if delta.CurrentScore < n.thresholdConfig.WarningThreshold &&
		delta.PreviousScore >= n.thresholdConfig.WarningThreshold {
		return "warning", "threshold_breach",
			"Trust score dropped below warning threshold"
	}

	// Check tier change
	if delta.TierChanged {
		if delta.CurrentTier < delta.PreviousTier {
			return "warning", "tier_change",
				"Trust tier downgraded from " + string(delta.PreviousTier) + " to " + string(delta.CurrentTier)
		}
		return "info", "tier_change",
			"Trust tier upgraded from " + string(delta.PreviousTier) + " to " + string(delta.CurrentTier)
	}

	// Regular score change
	if delta.ScoreChange > 0 {
		return "info", "trust_delta",
			"Trust score increased by " + formatChange(delta.ScoreChange)
	}
	return "info", "trust_delta",
		"Trust score decreased by " + formatChange(delta.ScoreChange)
}

// deliverWebhook sends webhook notifications for significant events
func (n *TrustDeltaNotifier) deliverWebhook(notification *TrustDeltaNotification) {
	if n.webhookService == nil {
		return
	}

	payload := map[string]interface{}{
		"event":          notification.Type,
		"function_id":    notification.FunctionID,
		"function_name":  notification.FunctionName,
		"severity":       notification.Severity,
		"previous_score": notification.PreviousScore,
		"current_score":  notification.CurrentScore,
		"score_change":   notification.ScoreChange,
		"message":        notification.Message,
		"timestamp":      notification.Timestamp,
		"data":           notification.Data,
	}

	if err := n.webhookService.DeliverEvent("trust.score."+notification.Type, payload); err != nil {
		n.logger.WithError(err).Warn("Failed to deliver trust score webhook")
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *TrustDeltaClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("Trust delta WebSocket read error")
			}
			break
		}

		// Handle client messages (subscription updates)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok && msgType == "subscribe" {
				if funcs, ok := msg["function_ids"].([]interface{}); ok {
					c.FunctionIDs = make([]string, 0, len(funcs))
					for _, f := range funcs {
						if fid, ok := f.(string); ok {
							c.FunctionIDs = append(c.FunctionIDs, fid)
						}
					}
				}
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *TrustDeltaClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper functions
func formatChange(change float64) string {
	if change > 0 {
		return "+" + formatScore(change)
	}
	return formatScore(change)
}

func formatScore(score float64) string {
	return jsonNumber(score)
}

func jsonNumber(n float64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
