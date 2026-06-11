package trustapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TrustScoreRepository defines the interface for trust score operations
type TrustScoreRepository interface {
	UpdateSlidingWindowScores(config registry.SlidingWindowConfig) ([]registry.TrustScoreDelta, error)
}

// TrustScoreStreamer handles Server-Sent Events for real-time trust score updates
type TrustScoreStreamer struct {
	repo            TrustScoreRepository
	logger          *logrus.Logger
	clients         map[string]*SSEClient // function_id -> client
	register        chan *SSEClient
	unregister      chan *SSEClient
	broadcast       chan *registry.TrustScoreStreamEvent
	shutdown        chan struct{}
	windowConfig    registry.SlidingWindowConfig
	thresholdConfig registry.TrustScoreThresholdConfig
}

// SSEClient represents a connected SSE client
type SSEClient struct {
	FunctionIDs []string           // Functions being watched (empty = all)
	FilterTier  registry.TrustTier // Filter by minimum tier
	Send        chan string        // Channel for sending SSE data
	Done        <-chan struct{}    // Client disconnect signal (read-only)
}

// NewTrustScoreStreamer creates a new trust score streaming handler
func NewTrustScoreStreamer(repo TrustScoreRepository, logger *logrus.Logger) *TrustScoreStreamer {
	return &TrustScoreStreamer{
		repo:            repo,
		logger:          logger,
		clients:         make(map[string]*SSEClient),
		register:        make(chan *SSEClient),
		unregister:      make(chan *SSEClient),
		broadcast:       make(chan *registry.TrustScoreStreamEvent, 100),
		shutdown:        make(chan struct{}),
		windowConfig:    registry.DefaultSlidingWindowConfig(),
		thresholdConfig: registry.DefaultThresholdConfig(),
	}
}

// SetWindowConfig updates the sliding window configuration
func (s *TrustScoreStreamer) SetWindowConfig(config registry.SlidingWindowConfig) {
	s.windowConfig = config
}

// SetThresholdConfig updates the threshold configuration
func (s *TrustScoreStreamer) SetThresholdConfig(config registry.TrustScoreThresholdConfig) {
	s.thresholdConfig = config
}

// Run starts the streaming hub
func (s *TrustScoreStreamer) Run() {
	s.logger.Info("Starting TrustScoreStreamer hub")

	for {
		select {
		case client := <-s.register:
			clientID := s.generateClientID(client)
			s.clients[clientID] = client
			s.logger.WithField("client_id", clientID).Debug("SSE client registered")

		case client := <-s.unregister:
			clientID := s.generateClientID(client)
			if _, ok := s.clients[clientID]; ok {
				delete(s.clients, clientID)
				close(client.Send)
				s.logger.WithField("client_id", clientID).Debug("SSE client unregistered")
			}

		case event := <-s.broadcast:
			s.handleBroadcast(event)

		case <-s.shutdown:
			s.logger.Info("Shutting down TrustScoreStreamer hub")
			for _, client := range s.clients {
				close(client.Send)
			}
			return
		}
	}
}

// generateClientID creates a unique client identifier
func (s *TrustScoreStreamer) generateClientID(client *SSEClient) string {
	// Simple hash of function IDs for client identification
	if len(client.FunctionIDs) == 0 {
		return fmt.Sprintf("all-%p", client)
	}
	return fmt.Sprintf("%s-%p", client.FunctionIDs[0], client)
}

// handleBroadcast sends events to relevant clients
func (s *TrustScoreStreamer) handleBroadcast(event *registry.TrustScoreStreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal stream event")
		return
	}

	sseData := fmt.Sprintf("event: %s\ndata: %s\n\n", event.EventType, string(data))

	for _, client := range s.clients {
		// Check if client is watching this function
		if !s.shouldSendToClient(client, event) {
			continue
		}

		select {
		case client.Send <- sseData:
		default:
			// Client buffer full, close connection
			s.unregister <- client
		}
	}
}

// shouldSendToClient determines if an event should be sent to a specific client
func (s *TrustScoreStreamer) shouldSendToClient(client *SSEClient, event *registry.TrustScoreStreamEvent) bool {
	// If client watches specific functions, check membership
	if len(client.FunctionIDs) > 0 {
		found := false
		for _, fid := range client.FunctionIDs {
			if fid == event.FunctionID.String() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tier filter if set
	if client.FilterTier != "" && event.Score != nil {
		if event.Score.TrustTier != client.FilterTier {
			// For minimum tier filtering, check if score is below filter
			tierValues := map[registry.TrustTier]int{
				registry.TrustTierUntrusted:     0,
				registry.TrustTierTrusted:       50,
				registry.TrustTierVerified:      70,
				registry.TrustTierHighlyTrusted: 90,
			}
			if tierValues[event.Score.TrustTier] < tierValues[client.FilterTier] {
				return false
			}
		}
	}

	return true
}

// BroadcastDelta sends a trust score delta to all relevant clients
func (s *TrustScoreStreamer) BroadcastDelta(delta *registry.TrustScoreDelta) {
	// Determine event type based on changes
	eventType := "score_update"
	if delta.TierChanged {
		eventType = "tier_change"
	}
	if delta.CurrentScore < s.thresholdConfig.CriticalThreshold {
		eventType = "threshold_breach"
	} else if delta.CurrentScore < s.thresholdConfig.WarningThreshold && delta.PreviousScore >= s.thresholdConfig.WarningThreshold {
		eventType = "threshold_breach"
	}

	event := &registry.TrustScoreStreamEvent{
		EventType:  eventType,
		FunctionID: delta.FunctionID,
		Delta:      delta,
		Timestamp:  time.Now(),
		WindowType: delta.WindowType,
	}

	select {
	case s.broadcast <- event:
	default:
		s.logger.Warn("Broadcast channel full, dropping event")
	}
}

// BroadcastScore sends a full trust score update
func (s *TrustScoreStreamer) BroadcastScore(score *registry.TrustHistory) {
	event := &registry.TrustScoreStreamEvent{
		EventType:  "score_update",
		FunctionID: score.FunctionID,
		Score:      score,
		Timestamp:  time.Now(),
		WindowType: registry.WindowTypeDiscrete,
	}

	select {
	case s.broadcast <- event:
	default:
		s.logger.Warn("Broadcast channel full, dropping event")
	}
}

// HandleSSE handles Server-Sent Events connections
func (s *TrustScoreStreamer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	middleware.SetCORSHeaders(w, r)

	// Parse query parameters
	functionIDs := r.URL.Query()["function_id"]
	filterTier := r.URL.Query().Get("min_tier")

	client := &SSEClient{
		FunctionIDs: functionIDs,
		Send:        make(chan string, 100),
		Done:        r.Context().Done(),
	}

	if filterTier != "" {
		client.FilterTier = registry.TrustTier(filterTier)
	}

	// Register client
	s.register <- client

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: %s\n\n", `{"status":"connected","watching":`+toJSONStringArray(functionIDs)+`}`)
	w.(http.Flusher).Flush()

	// Stream events
	for {
		select {
		case data, ok := <-client.Send:
			if !ok {
				return
			}
			fmt.Fprint(w, data)
			w.(http.Flusher).Flush()

		case <-client.Done:
			s.unregister <- client
			return
		}
	}
}

// HandleSSEFunction handles SSE for a single function via URL path
func (s *TrustScoreStreamer) HandleSSEFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID := vars["function_id"]

	// Validate function ID
	if _, err := uuid.Parse(functionID); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, "INVALID_FUNCTION_ID", "Invalid function ID")
		return
	}

	// Modify request to include function_id in query
	q := r.URL.Query()
	q.Add("function_id", functionID)
	r.URL.RawQuery = q.Encode()

	s.HandleSSE(w, r)
}

// StartSlidingWindowUpdates begins periodic sliding window calculations
func (s *TrustScoreStreamer) StartSlidingWindowUpdates(ctx context.Context) {
	ticker := time.NewTicker(s.windowConfig.UpdateInterval)
	defer ticker.Stop()

	s.logger.WithField("interval", s.windowConfig.UpdateInterval).Info("Starting sliding window updates")

	for {
		select {
		case <-ticker.C:
			s.performSlidingWindowUpdate()

		case <-ctx.Done():
			s.logger.Info("Stopping sliding window updates")
			return
		}
	}
}

// performSlidingWindowUpdate recalculates sliding window scores and broadcasts changes
func (s *TrustScoreStreamer) performSlidingWindowUpdate() {
	deltas, err := s.repo.UpdateSlidingWindowScores(s.windowConfig)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update sliding window scores")
		return
	}

	// Broadcast significant changes
	for _, delta := range deltas {
		// Only broadcast if change exceeds threshold or tier changed
		if abs(delta.ScoreChange) >= s.thresholdConfig.MinChangeForNotify ||
			delta.TierChanged ||
			delta.CurrentScore < s.thresholdConfig.WarningThreshold {
			s.BroadcastDelta(&delta)
		}
	}

	s.logger.WithField("deltas", len(deltas)).Debug("Sliding window update complete")
}

// TriggerImmediateUpdate forces an immediate sliding window recalculation
func (s *TrustScoreStreamer) TriggerImmediateUpdate() error {
	s.logger.Info("Triggering immediate sliding window update")
	go s.performSlidingWindowUpdate()
	return nil
}

// Stop stops the streaming hub
func (s *TrustScoreStreamer) Stop() {
	close(s.shutdown)
}

// Helper function to convert string slice to JSON array
func toJSONStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(arr)
	return string(data)
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
