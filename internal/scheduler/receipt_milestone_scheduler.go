// Package scheduler — receipt milestone sweep scheduler.
//
// Runs once a day and asks the receipt milestone worker to back-fill any
// thresholds we may have missed during downtime. The real-time hook in
// HandleExecute is the primary path; this scheduler is a defensive
// belt-and-braces that catches edge cases like:
//   - Orchestrator restart between an execution being recorded and the
//     milestone hook firing.
//   - Database failover where the milestone row was written but the
//     notification fan-out failed.
//   - Any future code path that creates a public-execution row without
//     going through the HandleExecute hook (e.g. bulk imports).
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/receipt"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// ReceiptMilestoneSchedulerConfig is the scheduler's runtime configuration.
type ReceiptMilestoneSchedulerConfig struct {
	// Enabled turns the scheduler on. Default: false.
	Enabled bool
	// Cron is a standard 5-field cron expression (UTC). Default: "0 3 * * *"
	// (3am daily — same as the data retention scheduler to spread load).
	Cron string
	// Lookback is how far back the sweep should look for missed milestones.
	// Default: 48h. Should be > the maximum orchestrator downtime you
	// expect between sweep runs.
	Lookback time.Duration
	// Logger for the scheduler.
	Logger *logrus.Logger
}

// DefaultReceiptMilestoneSchedulerConfig returns a safe default.
func DefaultReceiptMilestoneSchedulerConfig() ReceiptMilestoneSchedulerConfig {
	return ReceiptMilestoneSchedulerConfig{
		Enabled:  false,
		Cron:     "0 3 * * *",
		Lookback: 48 * time.Hour,
		Logger:   nil,
	}
}

// ReceiptMilestoneScheduler runs a daily sweep that re-checks milestone
// thresholds for functions with new receipts since the last sweep.
type ReceiptMilestoneScheduler struct {
	cfg     ReceiptMilestoneSchedulerConfig
	worker  *receipt.Milestone
	cron    *cron.Cron
	entryID cron.EntryID

	mu        sync.Mutex
	isRunning bool
}

// NewReceiptMilestoneScheduler constructs a scheduler. It does not start
// any goroutines — call Start.
func NewReceiptMilestoneScheduler(worker *receipt.Milestone, cfg ReceiptMilestoneSchedulerConfig) *ReceiptMilestoneScheduler {
	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
	}
	return &ReceiptMilestoneScheduler{
		cfg:    cfg,
		worker: worker,
		cron:   cron.New(),
	}
}

// Start schedules the daily sweep. Safe to call multiple times — subsequent
// calls are no-ops.
func (s *ReceiptMilestoneScheduler) Start(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if !s.cfg.Enabled {
		s.cfg.Logger.Info("Receipt milestone sweep scheduler is DISABLED")
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		return nil
	}

	expr := s.cfg.Cron
	if expr == "" {
		expr = "0 3 * * *"
	}
	entryID, err := s.cron.AddFunc(expr, func() {
		s.runSweep(ctx)
	})
	if err != nil {
		return err
	}
	s.entryID = entryID
	s.cron.Start()
	s.isRunning = true
	s.cfg.Logger.WithField("cron", expr).Info("Receipt milestone sweep scheduler started")
	return nil
}

// Stop halts the scheduler. Call this in graceful-shutdown paths.
func (s *ReceiptMilestoneScheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isRunning {
		return
	}
	s.cron.Stop()
	s.isRunning = false
}

// runSweep executes one sweep iteration.
func (s *ReceiptMilestoneScheduler) runSweep(parent context.Context) {
	lookback := s.cfg.Lookback
	if lookback <= 0 {
		lookback = 48 * time.Hour
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()

	since := time.Now().Add(-lookback)
	fired, err := s.worker.SweepMissedMilestones(ctx, since)
	if err != nil {
		s.cfg.Logger.WithError(err).Error("receipt milestone sweep failed")
		return
	}
	s.cfg.Logger.WithFields(logrus.Fields{
		"functions_checked": fired,
		"since":            since.Format(time.RFC3339),
	}).Info("Receipt milestone sweep complete")
}
