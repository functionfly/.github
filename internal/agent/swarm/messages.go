package swarm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/otel"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const (
	// asyncSendBufferSize is the buffer size for async message sending
	// This allows handling bursts without blocking the caller
	asyncSendBufferSize = 1000
	// asyncWorkerCount is the number of background workers processing messages
	asyncWorkerCount = 4
)

type MessageService struct {
	db             *gorm.DB
	signingService *SigningService
	redisClient    *redis.Client
	propagator     *otel.Propagator

	// Async message sending for burst handling
	asyncChan   chan *identity.AgentMessage
	asyncWg     sync.WaitGroup
	asyncStopCh chan struct{}
}

func NewMessageService(db *gorm.DB, redisClient *redis.Client) *MessageService {
	s := &MessageService{
		db:             db,
		signingService: NewSigningService(redisClient),
		redisClient:    redisClient,
		propagator:     otel.NewPropagator("agent-message"),
		asyncChan:      make(chan *identity.AgentMessage, asyncSendBufferSize),
		asyncStopCh:    make(chan struct{}),
	}
	// Start async workers
	s.startAsyncWorkers()
	return s
}

// NewMessageServiceWithoutWorkers creates a MessageService without starting background workers.
// Use this for testing when you don't want async workers running.
func NewMessageServiceWithoutWorkers(db *gorm.DB, redisClient *redis.Client) *MessageService {
	s := &MessageService{
		db:             db,
		signingService: NewSigningService(redisClient),
		redisClient:    redisClient,
		propagator:     otel.NewPropagator("agent-message"),
		asyncChan:      make(chan *identity.AgentMessage, asyncSendBufferSize),
		asyncStopCh:    make(chan struct{}),
	}
	// Don't start async workers
	return s
}

// startAsyncWorkers starts background workers for async message processing
func (s *MessageService) startAsyncWorkers() {
	for i := 0; i < asyncWorkerCount; i++ {
		s.asyncWg.Add(1)
		go s.asyncMessageWorker(i)
	}
}

// asyncMessageWorker processes messages from the async channel.
// Messages are accumulated and flushed either when the batch is full
// or when the periodic ticker fires.  The previous "flush on empty channel"
// heuristic is removed — it defeated batching for low-frequency messages
// (heartbeats, capability discovery) by flushing each message individually.
func (s *MessageService) asyncMessageWorker(workerID int) {
	defer s.asyncWg.Done()

	batch := make([]*identity.AgentMessage, 0, 100)
	// 50ms ticker balances latency vs. batching efficiency.
	// Heartbeats (sent every ~5s) will be batched with any co-occurring messages;
	// worst-case added latency is 50ms which is negligible for async work.
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.asyncStopCh:
			// Flush anything still buffered
			if len(batch) > 0 {
				s.flushBatch(batch)
			}
			// Drain remaining messages in channel
			for msg := range s.asyncChan {
				s.sendMessageSync(msg)
			}
			return
		case msg := <-s.asyncChan:
			batch = append(batch, msg)
			if len(batch) >= 100 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			// Periodic flush for small batches
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch sends a batch of messages to the database
func (s *MessageService) flushBatch(batch []*identity.AgentMessage) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()
	err := s.db.CreateInBatches(batch, 100).Error
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).WithField("count", len(batch)).Error("Failed to flush message batch")
		return
	}

	if duration > 50*time.Millisecond {
		logrus.WithFields(logrus.Fields{
			"duration_ms": duration.Milliseconds(),
			"batch_size":   len(batch),
		}).Warn("Slow batch message create")
	}
}

// sendMessageSync sends a message synchronously (used internally and for async fallback)
func (s *MessageService) sendMessageSync(msg *identity.AgentMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	if msg.TTLSeconds == 0 {
		msg.TTLSeconds = 3600
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.Nonce == "" {
		nonce, _ := generateNonce()
		msg.Nonce = nonce
	}

	createStart := time.Now()
	err := s.db.Create(msg).Error
	createDuration := time.Since(createStart)

	if createDuration > 50*time.Millisecond {
		sqlDB, _ := s.db.DB()
		var poolStats string
		if sqlDB != nil {
			stats := sqlDB.Stats()
			poolStats = fmt.Sprintf("open=%d inUse=%d idle=%d waitCount=%d",
				stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
		}
		logrus.WithFields(logrus.Fields{
			"duration_ms":   createDuration.Milliseconds(),
			"from_agent_id": msg.FromAgentID,
			"to_agent_id":   msg.ToAgentID,
			"message_type":  msg.MessageType,
			"table":         "agent_messages",
			"pool_stats":    poolStats,
		}).Warn("Slow agent message create")
	}

	return err
}

// Stop gracefully stops the async workers
func (s *MessageService) Stop() {
	close(s.asyncStopCh)
	s.asyncWg.Wait()
}

func (s *MessageService) SendMessage(ctx context.Context, msg *identity.AgentMessage, signingKey string) error {
	validTypes := []string{
		identity.MessageTypeTaskDelegation,
		identity.MessageTypeTaskResult,
		identity.MessageTypeQuery,
		identity.MessageTypeResponse,
		identity.MessageTypeCapabilityDiscovery,
		identity.MessageTypeHeartbeat,
		identity.MessageTypeEvolutionProposal,
		identity.MessageTypeBudgetRequest,
	}

	valid := false
	for _, t := range validTypes {
		if msg.MessageType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid message type: %s", msg.MessageType)
	}

	if signingKey == "" {
		return fmt.Errorf("signing key is required for agent-to-agent messages")
	}

	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	if msg.TTLSeconds == 0 {
		msg.TTLSeconds = 3600
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	msg.CreatedAt = time.Now()

	nonce, _ := generateNonce()
	seq, _ := s.signingService.IncrementSequence(ctx, msg.FromAgentID)
	msg.Nonce = nonce
	msg.SequenceNumber = seq

	payloadBytes, _ := json.Marshal(msg.Payload)
	msg.Signature = s.signingService.SignMessage(msg.FromAgentID, signingKey, payloadBytes, nonce, seq)

	s.injectTraceContext(ctx, msg)

	return s.db.WithContext(ctx).Create(msg).Error
}

func (s *MessageService) injectTraceContext(ctx context.Context, msg *identity.AgentMessage) {
	if s.propagator == nil {
		return
	}

	carrier := make(map[string]string)
	s.propagator.InjectTraceContext(ctx, carrier)

	if traceID, ok := carrier["traceparent"]; ok {
		parsed := parseTraceParent(traceID)
		msg.TraceID = parsed.traceID
		msg.SpanID = parsed.spanID
		msg.TraceFlags = parsed.traceFlags
		msg.TraceState = parsed.traceState
	}
}

type traceParentParts struct {
	traceID    string
	spanID     string
	traceFlags string
	traceState string
}

func parseTraceParent(tp string) traceParentParts {
	var parts traceParentParts
	if len(tp) < 55 || tp[:3] != "00-" {
		return parts
	}

	rest := tp[3:]
	endTraceID := 32 + 1

	if len(rest) > endTraceID {
		parts.traceID = rest[:32]
	}
	if len(rest) > endTraceID+1 {
		remaining := rest[endTraceID+1:]
		if len(remaining) >= 16 {
			parts.spanID = remaining[:16]
		}
		if len(remaining) > 16 {
			afterSpan := remaining[16+1:]
			if len(afterSpan) > 0 {
				parts.traceFlags = afterSpan[:2]
			}
			if len(afterSpan) > 2 {
				parts.traceState = afterSpan[3:]
			}
		}
	}
	return parts
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetInbox retrieves pending messages for an agent
func (s *MessageService) GetInbox(ctx context.Context, agentID string, limit int) ([]identity.AgentMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	var messages []identity.AgentMessage
	err := s.db.WithContext(ctx).
		Where("to_agent_id = ? AND status IN ?", agentID, []string{"pending", "delivered"}).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetInboxForAgents fetches pending messages for multiple agents in a single query.
// Returns a map of agentID -> messages, capped at limitPerAgent per agent.
func (s *MessageService) GetInboxForAgents(ctx context.Context, agentIDs []string, limitPerAgent int) (map[string][]identity.AgentMessage, error) {
	if len(agentIDs) == 0 {
		return nil, nil
	}
	if limitPerAgent <= 0 {
		limitPerAgent = 10
	}

	var messages []identity.AgentMessage
	err := s.db.WithContext(ctx).
		Where("to_agent_id IN ? AND status IN ?", agentIDs, []string{"pending", "delivered"}).
		Order("created_at ASC").
		Limit(len(agentIDs) * limitPerAgent).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Group by agent ID, preserving per-agent limit
	result := make(map[string][]identity.AgentMessage, len(agentIDs))
	for _, id := range agentIDs {
		result[id] = nil
	}
	for _, msg := range messages {
		if len(result[msg.ToAgentID]) < limitPerAgent {
			result[msg.ToAgentID] = append(result[msg.ToAgentID], msg)
		}
	}
	return result, nil
}

// GetOutbox retrieves messages sent by an agent
func (s *MessageService) GetOutbox(ctx context.Context, agentID string, limit int) ([]identity.AgentMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	var messages []identity.AgentMessage
	err := s.db.WithContext(ctx).
		Where("from_agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// MarkDelivered marks a message as delivered
func (s *MessageService) MarkDelivered(ctx context.Context, messageID uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("id = ? AND status = ?", messageID, "pending").
		Updates(map[string]any{
			"status":       "delivered",
			"delivered_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found or already delivered")
	}
	return nil
}

// MarkRead marks a single message as read.
func (s *MessageService) MarkRead(ctx context.Context, messageID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("id = ?", messageID).
		Update("status", "read")
	return result.Error
}

// MarkReadBatch marks multiple messages as read in a single UPDATE.
// This is significantly faster than calling MarkRead in a loop (1 round-trip vs N).
func (s *MessageService) MarkReadBatch(ctx context.Context, messageIDs []uuid.UUID) error {
	if len(messageIDs) == 0 {
		return nil
	}
	result := s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("id IN ?", messageIDs).
		Update("status", "read")
	return result.Error
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(ctx context.Context, messageID uuid.UUID, agentID string) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND (from_agent_id = ? OR to_agent_id = ?)", messageID, agentID, agentID).
		Delete(&identity.AgentMessage{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found or not authorized")
	}
	return nil
}

// CleanupExpired deletes expired messages
func (s *MessageService) CleanupExpired(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "pending", time.Now().Add(-24*time.Hour)).
		Delete(&identity.AgentMessage{})
	return result.RowsAffected, result.Error
}

// SendTaskDelegation sends a task from one agent to another with mandatory signing
func (s *MessageService) SendTaskDelegation(ctx context.Context, fromAgentID, toAgentID string, taskData map[string]any, sessionID string, signingKey string) error {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Payload:     taskData,
		SessionID:   &sessionID,
		TTLSeconds:  3600,
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg, signingKey)
}

// SendTaskResult sends a task result back to the delegating agent with mandatory signing
func (s *MessageService) SendTaskResult(ctx context.Context, fromAgentID, toAgentID string, resultData map[string]any, sessionID, parentExecutionID string, signingKey string) error {
	resultData["parent_execution_id"] = parentExecutionID
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeTaskResult,
		Payload:     resultData,
		SessionID:   &sessionID,
		TTLSeconds:  3600,
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg, signingKey)
}

// SendHeartbeat sends a heartbeat message between agents with mandatory signing
func (s *MessageService) SendHeartbeat(ctx context.Context, fromAgentID, toAgentID string, signingKey string) error {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: fromAgentID,
		ToAgentID:   toAgentID,
		MessageType: identity.MessageTypeHeartbeat,
		Payload:     map[string]any{"timestamp": time.Now().Unix()},
		TTLSeconds:  300,
		Status:      "pending",
	}
	return s.SendMessage(ctx, msg, signingKey)
}

// VerifyAndReceiveMessage verifies signature and receives a signed message
func (s *MessageService) VerifyAndReceiveMessage(ctx context.Context, msg *identity.AgentMessage, recipientSigningKey string) error {
	if msg.Signature == "" {
		return fmt.Errorf("message signature is required")
	}

	if msg.Nonce == "" {
		return fmt.Errorf("missing nonce for signed message")
	}

	allowed, err := s.signingService.ValidateReplay(ctx, msg.FromAgentID, msg.Nonce)
	if err != nil || !allowed {
		return fmt.Errorf("replay detected or nonce validation failed")
	}

	payloadBytes, _ := json.Marshal(msg.Payload)
	valid := s.signingService.VerifySignature(
		msg.FromAgentID,
		msg.Signature,
		recipientSigningKey,
		payloadBytes,
		msg.Nonce,
		msg.SequenceNumber,
	)
	if !valid {
		return fmt.Errorf("invalid message signature")
	}

	msg.Status = "delivered"

	s.extractAndContinueTrace(ctx, msg)

	return s.db.WithContext(ctx).Create(msg).Error
}

func (s *MessageService) extractAndContinueTrace(ctx context.Context, msg *identity.AgentMessage) context.Context {
	if s.propagator == nil || msg.TraceID == "" {
		return ctx
	}

	carrier := make(map[string]string)
	if msg.TraceID != "" {
		traceParent := fmt.Sprintf("00-%s-%s-01", msg.TraceID, msg.SpanID)
		carrier["traceparent"] = traceParent
	}
	if msg.TraceState != "" {
		carrier["tracestate"] = msg.TraceState
	}

	return s.propagator.ExtractTraceContext(ctx, carrier)
}

func (s *MessageService) StartMessageProcessingSpan(ctx context.Context, msg *identity.AgentMessage) (context.Context, func()) {
	if s.propagator == nil {
		return ctx, func() {}
	}

	carrier := make(map[string]string)
	if msg.TraceID != "" {
		traceParent := fmt.Sprintf("00-%s-%s-01", msg.TraceID, msg.SpanID)
		carrier["traceparent"] = traceParent
	}

	ctx = s.propagator.ExtractTraceContext(ctx, carrier)
	ctx, _ = s.propagator.StartSpan(ctx, fmt.Sprintf("agent.message.%s", msg.MessageType),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)

	return ctx, func() {
		s.propagator.EndSpan(ctx)
	}
}

// ReceiveMessage is a simplified version that stores unsigned messages (for backwards compatibility with trusted internal sources)
func (s *MessageService) ReceiveMessage(ctx context.Context, msg *identity.AgentMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	if msg.TTLSeconds == 0 {
		msg.TTLSeconds = 3600
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	msg.CreatedAt = time.Now()
	return s.db.WithContext(ctx).Create(msg).Error
}

// SendSystemMessage sends a message without signing (for internal system agents only)
// This should only be used for platform controller and other internal system agents
// Messages are sent asynchronously via a buffered channel to handle bursts
func (s *MessageService) SendSystemMessage(ctx context.Context, msg *identity.AgentMessage) error {
	// Prepare the message
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Payload == nil {
		msg.Payload = map[string]any{}
	}
	if msg.TTLSeconds == 0 {
		msg.TTLSeconds = 3600
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	msg.CreatedAt = time.Now()
	nonce, _ := generateNonce()
	msg.Nonce = nonce

	// Try to send to async channel (non-blocking)
	select {
	case s.asyncChan <- msg:
		return nil
	default:
		// Channel full, fall back to synchronous sending
		return s.sendMessageSync(msg)
	}
}
