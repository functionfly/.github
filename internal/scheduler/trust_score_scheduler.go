package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// TrustScoreScheduler manages scheduled trust score recalculations
type TrustScoreScheduler struct {
	cron       *cron.Cron
	repo       *registry.RegistryRepository
	scheduleID cron.EntryID
	mu         sync.RWMutex
	isRunning  bool
	nextRun    time.Time

	// Real-time streaming support
	streamingConfig     TrustScoreStreamingConfig
	onScoreUpdate       func(*registry.TrustScoreDelta)
	slidingWindowConfig registry.SlidingWindowConfig
}

// TrustScoreStreamingConfig controls real-time streaming behavior
type TrustScoreStreamingConfig struct {
	Enabled              bool          `json:"enabled"`
	BroadcastSignificant bool          `json:"broadcast_significant"` // Only broadcast significant changes
	MinChangeThreshold   float64       `json:"min_change_threshold"`  // Minimum score change to broadcast
	EnableSlidingWindow  bool          `json:"enable_sliding_window"` // Use sliding window instead of discrete
	UpdateInterval       time.Duration `json:"update_interval"`       // For sliding window updates
}

// TrustScoreScheduleConfig represents the trust score schedule configuration
type TrustScoreScheduleConfig struct {
	Enabled bool
	Cron    string // e.g., "0 * * * *" for hourly
}

// NewTrustScoreScheduler creates a new trust score scheduler
func NewTrustScoreScheduler(repo *registry.RegistryRepository) *TrustScoreScheduler {
	return &TrustScoreScheduler{
		cron: cron.New(),
		repo: repo,
		streamingConfig: TrustScoreStreamingConfig{
			Enabled:              false,
			BroadcastSignificant: true,
			MinChangeThreshold:   5.0,
			EnableSlidingWindow:  false,
			UpdateInterval:       5 * time.Minute,
		},
		slidingWindowConfig: registry.DefaultSlidingWindowConfig(),
	}
}

// SetOnScoreUpdate sets the callback for trust score updates
func (s *TrustScoreScheduler) SetOnScoreUpdate(callback func(*registry.TrustScoreDelta)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onScoreUpdate = callback
}

// SetStreamingConfig updates the streaming configuration
func (s *TrustScoreScheduler) SetStreamingConfig(config TrustScoreStreamingConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streamingConfig = config
}

// SetSlidingWindowConfig updates the sliding window configuration
func (s *TrustScoreScheduler) SetSlidingWindowConfig(config registry.SlidingWindowConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slidingWindowConfig = config
}

// StartSlidingWindowUpdates starts continuous sliding window updates in a separate goroutine
func (s *TrustScoreScheduler) StartSlidingWindowUpdates(ctx context.Context) {
	if !s.streamingConfig.EnableSlidingWindow {
		return
	}

	go func() {
		ticker := time.NewTicker(s.streamingConfig.UpdateInterval)
		defer ticker.Stop()

		logrus.WithField("interval", s.streamingConfig.UpdateInterval).Info("Starting sliding window updates")

		for {
			select {
			case <-ticker.C:
				s.performSlidingWindowUpdate()
			case <-ctx.Done():
				logrus.Info("Stopping sliding window updates")
				return
			}
		}
	}()
}

// performSlidingWindowUpdate recalculates all sliding window scores
func (s *TrustScoreScheduler) performSlidingWindowUpdate() {
	s.mu.RLock()
	config := s.slidingWindowConfig
	onUpdate := s.onScoreUpdate
	streamConfig := s.streamingConfig
	s.mu.RUnlock()

	deltas, err := s.repo.UpdateSlidingWindowScores(config)
	if err != nil {
		logrus.WithError(err).Error("Failed to update sliding window scores")
		return
	}

	// Notify callback for significant changes
	if onUpdate != nil {
		for _, delta := range deltas {
			if streamConfig.BroadcastSignificant {
				if abs(delta.ScoreChange) < streamConfig.MinChangeThreshold && !delta.TierChanged {
					continue
				}
			}
			onUpdate(&delta)
		}
	}

	logrus.WithFields(logrus.Fields{
		"deltas":              len(deltas),
		"significant_changes": countSignificant(deltas, streamConfig.MinChangeThreshold),
	}).Debug("Sliding window update complete")
}

// countSignificant counts deltas that exceed the threshold or involve tier changes
func countSignificant(deltas []registry.TrustScoreDelta, threshold float64) int {
	count := 0
	for _, d := range deltas {
		if abs(d.ScoreChange) >= threshold || d.TierChanged {
			count++
		}
	}
	return count
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Start starts the scheduler with the given configuration
func (s *TrustScoreScheduler) Start(ctx context.Context, config TrustScoreScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		logrus.Warn("Trust score scheduler already running")
		return nil
	}

	s.isRunning = true

	if config.Enabled && config.Cron != "" {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executeTrustScoreRecalculation(context.Background())
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to add trust score schedule")
			return err
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Trust score scheduler started with cron: %s", config.Cron)
	} else {
		logrus.Info("Trust score scheduler initialized but scheduling is disabled")
	}

	s.cron.Start()
	return nil
}

// Stop stops the scheduler
func (s *TrustScoreScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	logrus.Info("Stopping trust score scheduler")
	s.cron.Stop()
	s.isRunning = false

	return nil
}

// UpdateConfig updates the schedule configuration
func (s *TrustScoreScheduler) UpdateConfig(ctx context.Context, config TrustScoreScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing schedule
	if s.scheduleID > 0 {
		s.cron.Remove(s.scheduleID)
		s.scheduleID = 0
	}

	// Add new schedule if enabled
	if config.Enabled && config.Cron != "" {
		entryID, err := s.cron.AddFunc(config.Cron, func() {
			s.executeTrustScoreRecalculation(ctx)
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to update trust score schedule")
			return err
		}
		s.scheduleID = entryID

		// Calculate next run time
		schedule, err := cron.ParseStandard(config.Cron)
		if err == nil {
			s.nextRun = schedule.Next(time.Now())
		}

		logrus.Infof("Trust score schedule updated: cron=%s, enabled=%v", config.Cron, config.Enabled)
	}

	return nil
}

// GetStatus returns the current scheduler status
func (s *TrustScoreScheduler) GetStatus() TrustScoreSchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return TrustScoreSchedulerStatus{
		IsRunning: s.isRunning,
		NextRun:   s.nextRun,
		Enabled:   s.scheduleID > 0,
	}
}

// TrustScoreSchedulerStatus represents the current scheduler status
type TrustScoreSchedulerStatus struct {
	IsRunning bool      `json:"is_running"`
	NextRun   time.Time `json:"next_run,omitempty"`
	Enabled   bool      `json:"enabled"`
}

// executeTrustScoreRecalculation runs the trust score recalculation
func (s *TrustScoreScheduler) executeTrustScoreRecalculation(ctx context.Context) {
	s.mu.RLock()
	streamingEnabled := s.streamingConfig.Enabled
	slidingWindowEnabled := s.streamingConfig.EnableSlidingWindow
	onUpdate := s.onScoreUpdate
	s.mu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"streaming_enabled":      streamingEnabled,
		"sliding_window_enabled": slidingWindowEnabled,
	}).Info("Starting scheduled trust score recalculation")

	// First, aggregate hourly metrics for the past hour
	hour := time.Now().Add(-1 * time.Hour).Truncate(time.Hour)
	if err := s.repo.AggregateHourlyMetrics(hour); err != nil {
		logrus.WithError(err).Error("Failed to aggregate hourly metrics")
	}

	var deltas []registry.TrustScoreDelta

	// Use sliding window calculations if enabled
	if slidingWindowEnabled {
		var err error
		deltas, err = s.repo.UpdateSlidingWindowScores(s.slidingWindowConfig)
		if err != nil {
			logrus.WithError(err).Error("Failed to update sliding window scores")
		}
	} else {
		// Traditional discrete window recalculation
		// We need to track deltas manually since RefreshAllTrustScores doesn't return them
		deltas = s.recalculateWithDeltas()
	}

	// Broadcast updates if streaming is enabled
	if streamingEnabled && onUpdate != nil {
		for _, delta := range deltas {
			// Only broadcast significant changes
			threshold := s.streamingConfig.MinChangeThreshold
			if s.streamingConfig.BroadcastSignificant {
				if abs(delta.ScoreChange) < threshold && !delta.TierChanged {
					continue
				}
			}
			onUpdate(&delta)
		}

		logrus.WithField("broadcast_count", len(deltas)).Debug("Broadcast trust score updates")
	}

	// Also run the full recalculation job for history tracking
	job, err := s.repo.RefreshAllTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh trust scores")
		return
	}

	logrus.WithFields(logrus.Fields{
		"job_id":              job.ID,
		"functions_processed": job.FunctionsProcessed,
		"functions_total":     job.FunctionsTotal,
		"status":              job.Status,
	}).Info("Completed scheduled trust score recalculation")

	// Update next run time
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scheduleID > 0 {
		entries := s.cron.Entries()
		for _, entry := range entries {
			if entry.ID == s.scheduleID {
				s.nextRun = entry.Next
				break
			}
		}
	}
}

// recalculateWithDeltas performs traditional recalculation and returns deltas
// This is used when sliding window is disabled but streaming is enabled
func (s *TrustScoreScheduler) recalculateWithDeltas() []registry.TrustScoreDelta {
	// Query current scores before recalculation
	functions, err := s.repo.GetAllFunctionsWithTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to get functions for delta tracking")
		return []registry.TrustScoreDelta{}
	}

	// Store previous scores
	previousScores := make(map[string]struct {
		Score float64
		Tier  registry.TrustTier
	})
	for _, fn := range functions {
		previousScores[fn.ID.String()] = struct {
			Score float64
			Tier  registry.TrustTier
		}{
			Score: fn.TrustScore,
			Tier:  fn.TrustTier,
		}
	}

	// Recalculate all trust scores
	_, err = s.repo.RefreshAllTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to refresh trust scores for delta tracking")
		return []registry.TrustScoreDelta{}
	}

	// Query new scores and generate deltas
	updatedFunctions, err := s.repo.GetAllFunctionsWithTrustScores()
	if err != nil {
		logrus.WithError(err).Error("Failed to get updated functions for delta tracking")
		return []registry.TrustScoreDelta{}
	}

	var deltas []registry.TrustScoreDelta
	now := time.Now()
	for _, fn := range updatedFunctions {
		prev, exists := previousScores[fn.ID.String()]
		if !exists {
			// New function or no previous data
			continue
		}

		// Only include if score changed
		if fn.TrustScore != prev.Score {
			scoreChange := fn.TrustScore - prev.Score
			scoreChangePercent := 0.0
			if prev.Score > 0 {
				scoreChangePercent = (scoreChange / prev.Score) * 100
			}

			delta := registry.TrustScoreDelta{
				FunctionID:         fn.ID,
				PreviousScore:      prev.Score,
				CurrentScore:       fn.TrustScore,
				ScoreChange:        scoreChange,
				ScoreChangePercent: scoreChangePercent,
				PreviousTier:       prev.Tier,
				CurrentTier:        fn.TrustTier,
				TierChanged:        prev.Tier != fn.TrustTier,
				CalculatedAt:       now,
				WindowType:         registry.WindowTypeDiscrete,
			}
			deltas = append(deltas, delta)
		}
	}

	return deltas
}

// TriggerImmediate triggers an immediate trust score recalculation
func (s *TrustScoreScheduler) TriggerImmediate(ctx context.Context) error {
	logrus.Info("Triggering immediate trust score recalculation")
	go s.executeTrustScoreRecalculation(ctx)
	return nil
}
