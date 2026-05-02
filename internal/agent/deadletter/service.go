package deadletter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/agent/metrics"
)

type AlertConfig struct {
	RetryStormThreshold  int
	RetryStormWindow     time.Duration
	RetryStormCooldown   time.Duration
	AlertCooldownMinutes int
}

func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		RetryStormThreshold:  5,
		RetryStormWindow:      5 * time.Minute,
		RetryStormCooldown:    10 * time.Minute,
		AlertCooldownMinutes:  15,
	}
}

type Service struct {
	repo   *Repository
	logger *logrus.Logger
	cfg    AlertConfig

	stormTracker map[string]*stormEntry
	stormMu      sync.RWMutex

	alertCooldown map[string]time.Time
	cooldownMu    sync.RWMutex

	stopCh chan struct{}
	wg     sync.WaitGroup
}

type stormEntry struct {
	Count     int
	StartTime time.Time
}

func NewService(repo *Repository, logger *logrus.Logger, cfg AlertConfig) *Service {
	if cfg.RetryStormThreshold == 0 {
		cfg = DefaultAlertConfig()
	}
	return &Service{
		repo:          repo,
		logger:        logger,
		cfg:           cfg,
		stormTracker:  make(map[string]*stormEntry),
		alertCooldown: make(map[string]time.Time),
		stopCh:        make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.monitorLoop(ctx)
	s.logger.Info("Dead letter service started")
}

func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("Dead letter service stopped")
}

func (s *Service) RecordFailure(ctx context.Context, req *FailureRecordRequest) error {
	now := time.Now()
	dl := &AgentDeadLetter{
		ID:                uuid.New(),
		AgentID:          req.AgentID,
		TenantID:         req.TenantID,
		FunctionID:       req.FunctionID,
		FunctionURI:      req.FunctionURI,
		ExecutionID:      req.ExecutionID,
		SessionID:        req.SessionID,
		InputPayload:     req.InputPayload,
		FinalError:       req.FinalError,
		ErrorCode:        req.ErrorCode,
		Attempts:         req.Attempts,
		FirstAttemptAt:   req.FirstAttemptAt,
		LastAttemptAt:    &now,
		FirstAttemptError: req.FirstAttemptError,
		Status:           StatusPending,
		CanRetry:         true,
		AlertThreshold:   s.cfg.RetryStormThreshold,
		Trace:            req.Trace,
		CreatedAt:        now,
	}

	if err := s.repo.Create(ctx, dl); err != nil {
		return fmt.Errorf("failed to create dead letter entry: %w", err)
	}

	metrics.AgentDeadLetterTotal.WithLabelValues(req.AgentID, req.FunctionURI, req.ErrorCode).Inc()
	metrics.AgentDeadLetterAttemptsHistogram.WithLabelValues(req.AgentID, req.FunctionURI).Observe(float64(req.Attempts))
	metrics.AgentDeadLetterPending.WithLabelValues(req.AgentID).Inc()

	if s.isRetryStorm(req.AgentID, req.FunctionURI) {
		s.logger.WithFields(logrus.Fields{
			"agent_id":     req.AgentID,
			"function_uri": req.FunctionURI,
			"attempts":     req.Attempts,
		}).Warn("Retry storm detected")
		metrics.AgentDeadLetterRetryStorm.WithLabelValues(req.AgentID, req.FunctionURI).Inc()
		s.sendRetryStormAlert(ctx, req)
	}

	return nil
}

type FailureRecordRequest struct {
	AgentID          string
	TenantID         uuid.UUID
	FunctionID       uuid.UUID
	FunctionURI      string
	ExecutionID      string
	SessionID        string
	InputPayload     JSONMap
	FinalError       string
	ErrorCode        string
	Attempts         int
	FirstAttemptAt   *time.Time
	FirstAttemptError string
	Trace            string
}

func (s *Service) GetEntry(ctx context.Context, id uuid.UUID) (*AgentDeadLetter, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListEntries(ctx context.Context, agentID string, status Status, limit, offset int) ([]*AgentDeadLetter, int64, error) {
	return s.repo.List(ctx, agentID, status, limit, offset)
}

func (s *Service) RetryEntry(ctx context.Context, id uuid.UUID, fn func(ctx context.Context, req *RetryRequest) error) (*RetryResult, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !entry.CanRetry {
		return nil, fmt.Errorf("entry is not retryable")
	}

	req := &RetryRequest{
		ExecutionID: entry.ExecutionID,
		InputPayload: entry.InputPayload,
		AgentID:     entry.AgentID,
		TenantID:    entry.TenantID,
		FunctionID:  entry.FunctionID,
	}

	err = fn(ctx, req)
	if err != nil {
		if retryErr := s.repo.MarkRetried(ctx, id, false, err.Error()); retryErr != nil {
			s.logger.WithError(retryErr).Warn("Failed to mark retry as failed")
		}
		metrics.AgentDeadLetterRetryOutcome.WithLabelValues(entry.AgentID, "failed").Inc()
		return &RetryResult{Success: false, Error: err.Error()}, err
	}

	if err := s.repo.MarkRetried(ctx, id, true, ""); err != nil {
		s.logger.WithError(err).Warn("Failed to mark retry as successful")
	}
	metrics.AgentDeadLetterRetryOutcome.WithLabelValues(entry.AgentID, "success").Inc()
	metrics.AgentDeadLetterPending.WithLabelValues(entry.AgentID).Dec()

	return &RetryResult{Success: true}, nil
}

type RetryRequest struct {
	ExecutionID string
	InputPayload JSONMap
	AgentID     string
	TenantID    uuid.UUID
	FunctionID  uuid.UUID
}

type RetryResult struct {
	Success bool
	Error   string
}

func (s *Service) MarkInspected(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkInspected(ctx, id)
}

func (s *Service) MarkDiscarded(ctx context.Context, id uuid.UUID) error {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.MarkDiscarded(ctx, id); err != nil {
		return err
	}
	metrics.AgentDeadLetterPending.WithLabelValues(entry.AgentID).Dec()
	return nil
}

func (s *Service) GetStats(ctx context.Context, agentID string) (*Stats, error) {
	return s.repo.GetStats(ctx, agentID)
}

func (s *Service) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	return s.repo.DeleteOlderThan(ctx, olderThan)
}

func (s *Service) isRetryStorm(agentID, functionURI string) bool {
	key := fmt.Sprintf("%s:%s", agentID, functionURI)

	s.stormMu.Lock()
	defer s.stormMu.Unlock()

	now := time.Now()
	entry, exists := s.stormTracker[key]

	if !exists {
		s.stormTracker[key] = &stormEntry{
			Count:     1,
			StartTime: now,
		}
		return false
	}

	if now.Sub(entry.StartTime) > s.cfg.RetryStormWindow {
		entry.Count = 1
		entry.StartTime = now
		return false
	}

	entry.Count++
	return entry.Count >= s.cfg.RetryStormThreshold
}

func (s *Service) sendRetryStormAlert(ctx context.Context, req *FailureRecordRequest) {
	key := fmt.Sprintf("%s:%s", req.AgentID, req.FunctionURI)

	s.cooldownMu.Lock()
	defer s.cooldownMu.Unlock()

	if lastAlert, exists := s.alertCooldown[key]; exists {
		if time.Since(lastAlert) < s.cfg.RetryStormCooldown {
			return
		}
	}

	s.logger.WithFields(logrus.Fields{
		"agent_id":     req.AgentID,
		"function_uri": req.FunctionURI,
		"attempts":     req.Attempts,
		"error":        req.FinalError,
	}).Error("RETRY STORM ALERT: Multiple dead letter entries in short period")

	s.alertCooldown[key] = time.Now()
}

func (s *Service) monitorLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupStaleStormEntries()
			s.updatePendingMetrics(ctx)
		}
	}
}

func (s *Service) cleanupStaleStormEntries() {
	s.stormMu.Lock()
	defer s.stormMu.Unlock()

	now := time.Now()
	for key, entry := range s.stormTracker {
		if now.Sub(entry.StartTime) > s.cfg.RetryStormWindow*2 {
			delete(s.stormTracker, key)
		}
	}
}

func (s *Service) updatePendingMetrics(ctx context.Context) {
	var agents []string
	s.repo.db.WithContext(ctx).Model(&AgentDeadLetter{}).Distinct("agent_id").Where("status = ?", StatusPending).Pluck("agent_id", &agents)

	for _, agentID := range agents {
		count, err := s.repo.CountPendingForAgent(ctx, agentID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to count pending dead letters")
			continue
		}
		metrics.AgentDeadLetterPending.WithLabelValues(agentID).Set(float64(count))
	}
}

func (r *Repository) CountPendingForAgent(ctx context.Context, agentID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgentDeadLetter{}).Where("agent_id = ? AND status = ?", agentID, StatusPending).Count(&count).Error
	return count, err
}

type RetryHandler func(ctx context.Context, req *RetryRequest) error

func (s *Service) HandleDeadLetter(ctx context.Context, agentID string, tenantID uuid.UUID, functionID uuid.UUID, functionURI string, executionID string, inputPayload JSONMap, err error, attempts int, firstErr error, trace string) error {
	req := &FailureRecordRequest{
		AgentID:          agentID,
		TenantID:         tenantID,
		FunctionID:       functionID,
		FunctionURI:      functionURI,
		ExecutionID:      executionID,
		InputPayload:     inputPayload,
		FinalError:       err.Error(),
		ErrorCode:        classifyError(err),
		Attempts:         attempts,
		FirstAttemptError: firstErr.Error(),
		Trace:            trace,
	}
	return s.RecordFailure(ctx, req)
}

func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection"):
		return "connection_error"
	case contains(errStr, "memory"):
		return "memory_error"
	case contains(errStr, "quota"):
		return "quota_exceeded"
	default:
		return "execution_error"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}