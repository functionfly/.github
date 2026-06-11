package lifecycle

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type identityVerifier interface {
	VerifyAPIKey(ctx context.Context, agentID, apiKey string) (bool, error)
}

type LifecycleStatus string

const (
	LifecycleStatusRegistered             LifecycleStatus = "registered"
	LifecycleStatusActive                LifecycleStatus = "active"
	LifecycleStatusGracefulShutdownStart LifecycleStatus = "graceful_shutdown_start"
	LifecycleStatusGracefulShutdownDone  LifecycleStatus = "graceful_shutdown_complete"
	LifecycleStatusOrphaned              LifecycleStatus = "orphaned"
	LifecycleStatusCrashed               LifecycleStatus = "crashed"
)

type EventType string

const (
	EventTypeRegistered            EventType = "registered"
	EventTypeHeartbeat             EventType = "heartbeat"
	EventTypeGracefulShutdownStart EventType = "graceful_shutdown_start"
	EventTypeGracefulShutdownDone  EventType = "graceful_shutdown_complete"
	EventTypeForcedShutdown        EventType = "forced_shutdown"
	EventTypeOrphanDetected        EventType = "orphan_detected"
	EventTypeCrashRecovery         EventType = "crash_recovery"
	EventTypeStateCheckpoint       EventType = "state_checkpoint"
)

type AgentLifecycleEvent struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID   string         `json:"agent_id" gorm:"not null;index"`
	EventType EventType      `json:"event_type" gorm:"not null;index"`
	EventData map[string]any `json:"event_data" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;index"`
}

func (AgentLifecycleEvent) TableName() string {
	return "agent_lifecycle_events"
}

type AgentLifecycleState struct {
	LifecycleStatus           string     `json:"lifecycle_status"`
	LastHeartbeatAt           *time.Time `json:"last_heartbeat_at"`
	LastActiveAt              *time.Time `json:"last_active_at"`
	GracefulShutdownAt        *time.Time `json:"graceful_shutdown_at"`
	ForcedShutdownAt          *time.Time `json:"forced_shutdown_at"`
	OrphanDetectedAt          *time.Time `json:"orphan_detected_at"`
	ShutdownGracePeriodSeconds int        `json:"shutdown_grace_period_seconds"`
	StateSnapshot             JSONMap    `json:"state_snapshot" gorm:"type:jsonb"`
}

type JSONMap map[string]any

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for JSONB")
	}
	if len(b) == 0 {
		*m = make(JSONMap)
		return nil
	}
	return json.Unmarshal(b, m)
}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

type Config struct {
	HeartbeatTimeout       time.Duration
	OrphanCheckInterval    time.Duration
	CheckpointInterval     time.Duration
	DefaultGracePeriodSecs int
}

func DefaultConfig() Config {
	return Config{
		HeartbeatTimeout:       2 * time.Minute,
		OrphanCheckInterval:    1 * time.Minute,
		CheckpointInterval:     30 * time.Second,
		DefaultGracePeriodSecs: 30,
	}
}

type Service struct {
	db           *gorm.DB
	redis        *redis.Client
	config       Config
	log          *logrus.Logger
	mu           sync.RWMutex
	activeAgents map[string]*AgentContext
	stopCh       chan struct{}
	wg           sync.WaitGroup
	identityVerifier identityVerifier
}

type AgentContext struct {
	AgentID       string
	Status        LifecycleStatus
	LastHeartbeat time.Time
	StateSnapshot JSONMap
	inFlight      int32
	shutdownSig   chan struct{}
}

func NewService(db *gorm.DB, redis *redis.Client, log *logrus.Logger, identityVerifier identityVerifier) *Service {
	return &Service{
		db:           db,
		redis:        redis,
		config:       DefaultConfig(),
		log:          log,
		activeAgents: make(map[string]*AgentContext),
		stopCh:       make(chan struct{}),
		identityVerifier: identityVerifier,
	}
}

func (s *Service) Start() {
	s.wg.Add(3)
	go s.heartbeatMonitor()
	go s.orphanDetector()
	go s.checkpointWorker()
}

func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.log.Info("Agent lifecycle service stopped")
}

func (s *Service) RegisterAgent(ctx context.Context, agentID string) error {
	s.log.WithField("agent_id", agentID).Info("Registering agent for lifecycle management")

	now := time.Now()
	agentCtx := &AgentContext{
		AgentID:       agentID,
		Status:        LifecycleStatusActive,
		LastHeartbeat: now,
		StateSnapshot: make(JSONMap),
		shutdownSig:    make(chan struct{}),
	}

	s.mu.Lock()
	s.activeAgents[agentID] = agentCtx
	s.mu.Unlock()

	result := s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET lifecycle_status = ?, last_heartbeat_at = ?, last_active_at = ? WHERE agent_id = ?`,
			LifecycleStatusActive, now, now, agentID)
	return result.Error
}

func (s *Service) RecordHeartbeat(ctx context.Context, agentID, apiKey string, stateSnapshot JSONMap) error {
	if s.identityVerifier != nil {
		valid, err := s.identityVerifier.VerifyAPIKey(ctx, agentID, apiKey)
		if err != nil {
			s.log.WithError(err).WithField("agent_id", agentID).Warn("API key verification error")
			return fmt.Errorf("failed to verify API key: %w", err)
		}
		if !valid {
			s.log.WithField("agent_id", agentID).Warn("Invalid API key for heartbeat")
			return fmt.Errorf("invalid API key for agent: %s", agentID)
		}
	}

	now := time.Now()

	s.mu.Lock()
	if agent, ok := s.activeAgents[agentID]; ok {
		agent.LastHeartbeat = now
		agent.Status = LifecycleStatusActive
		if len(stateSnapshot) > 0 {
			agent.StateSnapshot = stateSnapshot
		}
	}
	s.mu.Unlock()

	if err := s.recordEvent(ctx, agentID, EventTypeHeartbeat, map[string]any{
		"timestamp":       now,
		"state_snapshot":  stateSnapshot,
	}); err != nil {
		s.log.WithError(err).WithField("agent_id", agentID).Warn("Failed to record heartbeat event")
	}

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET last_heartbeat_at = ?, last_active_at = ?, lifecycle_status = ?, state_snapshot = ? WHERE agent_id = ?`,
			now, now, LifecycleStatusActive, stateSnapshot, agentID).Error
}

func (s *Service) InitiateGracefulShutdown(ctx context.Context, agentID string) error {
	s.log.WithField("agent_id", agentID).Info("Initiating graceful shutdown for agent")

	s.mu.Lock()
	if agent, ok := s.activeAgents[agentID]; ok {
		agent.Status = LifecycleStatus(LifecycleStatusGracefulShutdownStart)
		close(agent.shutdownSig)
		agent.shutdownSig = make(chan struct{})
	}
	s.mu.Unlock()

	now := time.Now()
	if err := s.recordEvent(ctx, agentID, EventTypeGracefulShutdownStart, map[string]any{
		"initiated_at": now,
	}); err != nil {
		s.log.WithError(err).WithField("agent_id", agentID).Warn("Failed to record shutdown event")
	}

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET lifecycle_status = ?, graceful_shutdown_at = ? WHERE agent_id = ?`,
			LifecycleStatusGracefulShutdownStart, now, agentID).Error
}

func (s *Service) CompleteGracefulShutdown(ctx context.Context, agentID string) error {
	s.log.WithField("agent_id", agentID).Info("Completing graceful shutdown for agent")

	s.mu.Lock()
	delete(s.activeAgents, agentID)
	s.mu.Unlock()

	now := time.Now()
	if err := s.recordEvent(ctx, agentID, EventTypeGracefulShutdownDone, map[string]any{
		"completed_at": now,
	}); err != nil {
		s.log.WithError(err).WithField("agent_id", agentID).Warn("Failed to record shutdown complete event")
	}

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET lifecycle_status = ?, graceful_shutdown_at = ?, status = 'suspended' WHERE agent_id = ?`,
			LifecycleStatusGracefulShutdownDone, now, agentID).Error
}

func (s *Service) ForceShutdown(ctx context.Context, agentID string) error {
	s.log.WithField("agent_id", agentID).Warn("Forcing shutdown for agent")

	s.mu.Lock()
	delete(s.activeAgents, agentID)
	s.mu.Unlock()

	now := time.Now()
	if err := s.recordEvent(ctx, agentID, EventTypeForcedShutdown, map[string]any{
		"forced_at": now,
	}); err != nil {
		s.log.WithError(err).WithField("agent_id", agentID).Warn("Failed to record forced shutdown event")
	}

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET lifecycle_status = ?, forced_shutdown_at = ?, status = 'suspended' WHERE agent_id = ?`,
			LifecycleStatus(LifecycleStatusOrphaned), now, agentID).Error
}

func (s *Service) RecordOrphanDetection(ctx context.Context, agentID string) error {
	s.log.WithField("agent_id", agentID).Warn("Agent marked as orphaned")

	s.mu.Lock()
	if agent, ok := s.activeAgents[agentID]; ok {
		agent.Status = LifecycleStatus(LifecycleStatusOrphaned)
	}
	s.mu.Unlock()

	now := time.Now()
	if err := s.recordEvent(ctx, agentID, EventTypeOrphanDetected, map[string]any{
		"detected_at":     now,
		"last_heartbeat":  s.getLastHeartbeat(agentID),
	}); err != nil {
		s.log.WithError(err).WithField("agent_id", agentID).Warn("Failed to record orphan event")
	}

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET lifecycle_status = ?, orphan_detected_at = ? WHERE agent_id = ?`,
			LifecycleStatusOrphaned, now, agentID).Error
}

func (s *Service) SaveStateSnapshot(ctx context.Context, agentID string, state JSONMap) error {
	s.mu.Lock()
	if agent, ok := s.activeAgents[agentID]; ok {
		agent.StateSnapshot = state
	}
	s.mu.Unlock()

	return s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET state_snapshot = ? WHERE agent_id = ?`, state, agentID).Error
}

func (s *Service) GetStateSnapshot(ctx context.Context, agentID string) (JSONMap, error) {
	var agent struct {
		StateSnapshot JSONMap `gorm:"column:state_snapshot"`
	}
	err := s.db.WithContext(ctx).Model(&struct{}{}).
		Select("state_snapshot").
		Where("agent_id = ?", agentID).
		Find(&agent).Error
	if err != nil {
		return nil, err
	}
	return agent.StateSnapshot, nil
}

func (s *Service) GetLifecycleStatus(ctx context.Context, agentID string) (*AgentLifecycleState, error) {
	var state AgentLifecycleState
	err := s.db.WithContext(ctx).
		Raw(`SELECT lifecycle_status, last_heartbeat_at, last_active_at,
			graceful_shutdown_at, forced_shutdown_at, orphan_detected_at,
			shutdown_grace_period_seconds, state_snapshot
			FROM agent_identities WHERE agent_id = ?`, agentID).
		Scan(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Service) heartbeatMonitor() {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			s.cleanupStaleEntries(ctx)
			cancel()
		}
	}
}

func (s *Service) cleanupStaleEntries(ctx context.Context) {
	s.mu.Lock()
	for agentID, agent := range s.activeAgents {
		if time.Since(agent.LastHeartbeat) > s.config.HeartbeatTimeout {
			s.log.WithField("agent_id", agentID).Debug("Removing stale agent from active tracking")
		}
	}
	s.mu.Unlock()
}

func (s *Service) orphanDetector() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.config.OrphanCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s.detectOrphans(ctx)
			cancel()
		}
	}
}

func (s *Service) detectOrphans(ctx context.Context) {
	cutoff := time.Now().Add(-s.config.HeartbeatTimeout)

	var agentIDs []string
	err := s.db.WithContext(ctx).
		Raw(`SELECT agent_id FROM agent_identities
			WHERE lifecycle_status IN (?, ?)
			AND last_heartbeat_at < ?`, LifecycleStatusActive, LifecycleStatusRegistered, cutoff).
		Pluck("agent_id", &agentIDs).Error
	if err != nil {
		s.log.WithError(err).Error("Failed to detect orphaned agents")
		return
	}

	for _, agentID := range agentIDs {
		s.log.WithField("agent_id", agentID).Warn("Detected orphaned agent")
		if err := s.RecordOrphanDetection(ctx, agentID); err != nil {
			s.log.WithError(err).WithField("agent_id", agentID).Error("Failed to record orphan detection")
		}
	}
}

func (s *Service) checkpointWorker() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.config.CheckpointInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			s.saveAllCheckpoints()
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s.saveAllCheckpointsCtx(ctx)
			cancel()
		}
	}
}

func (s *Service) saveAllCheckpoints() {
	s.mu.RLock()
	for agentID, agent := range s.activeAgents {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.saveSnapshot(ctx, agentID, agent.StateSnapshot)
		cancel()
	}
	s.mu.RUnlock()
}

func (s *Service) saveAllCheckpointsCtx(ctx context.Context) {
	s.mu.RLock()
	var wg sync.WaitGroup
	for agentID, agent := range s.activeAgents {
		wg.Add(1)
		go func(agentID string, snapshot JSONMap) {
			defer wg.Done()
			s.saveSnapshot(ctx, agentID, snapshot)
		}(agentID, agent.StateSnapshot)
	}
	s.mu.RUnlock()
	wg.Wait()
}

func (s *Service) saveSnapshot(ctx context.Context, agentID string, snapshot JSONMap) {
	if len(snapshot) == 0 {
		return
	}
	result := s.db.WithContext(ctx).Model(&struct{}{}).
		Exec(`UPDATE agent_identities SET state_snapshot = ? WHERE agent_id = ?`, snapshot, agentID)
	if result.Error != nil {
		s.log.WithError(result.Error).WithField("agent_id", agentID).Warn("Failed to save checkpoint")
	}
}

func (s *Service) recordEvent(ctx context.Context, agentID string, eventType EventType, data map[string]any) error {
	event := &AgentLifecycleEvent{
		ID:        uuid.New(),
		AgentID:   agentID,
		EventType: eventType,
		EventData: data,
		CreatedAt: time.Now(),
	}
	return s.db.WithContext(ctx).Create(event).Error
}

func (s *Service) getLastHeartbeat(agentID string) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if agent, ok := s.activeAgents[agentID]; ok {
		return agent.LastHeartbeat
	}
	return time.Time{}
}

func (s *Service) GetEvents(ctx context.Context, agentID string, limit int) ([]*AgentLifecycleEvent, error) {
	var events []*AgentLifecycleEvent
	query := s.db.WithContext(ctx).Where("agent_id = ?", agentID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&events).Error
	return events, err
}

func (s *Service) IsAgentAlive(ctx context.Context, agentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if agent, ok := s.activeAgents[agentID]; ok {
		return time.Since(agent.LastHeartbeat) < s.config.HeartbeatTimeout && agent.Status != LifecycleStatus(LifecycleStatusOrphaned)
	}
	return false
}

func (s *Service) GetActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeAgents)
}