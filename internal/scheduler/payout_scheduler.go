package scheduler

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/payment"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// PayoutScheduleConfig represents configuration for the payout processing scheduler.
type PayoutScheduleConfig struct {
	Enabled bool
	Cron    string
	Timezone string
}

// PayoutScheduler processes scheduled/auto-payouts on a cron schedule.
type PayoutScheduler struct {
	cron       *cron.Cron
	payoutSvc  *payment.PayoutServiceExtended
	logger     *logrus.Logger
	stopOnce   sync.Once
	cancel     context.CancelFunc
	enabled    bool
}

// NewPayoutScheduler creates a new payout processing scheduler.
func NewPayoutScheduler(payoutSvc *payment.PayoutServiceExtended) *PayoutScheduler {
	return &PayoutScheduler{
		cron:      cron.New(),
		payoutSvc: payoutSvc,
		logger:    logrus.New(),
	}
}

// Start begins the payout scheduler.
func (s *PayoutScheduler) Start(ctx context.Context, config PayoutScheduleConfig) error {
	if !config.Enabled {
		s.logger.Info("Payout scheduler disabled")
		return nil
	}

	if config.Cron == "" {
		config.Cron = "0 2 * * 1" // Default: every Monday at 2 AM UTC
	}

	if config.Timezone == "" {
		config.Timezone = "UTC"
	}

	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		s.logger.WithError(err).Warn("Invalid timezone, using UTC")
		loc = time.UTC
	}

	s.cron = cron.New(cron.WithLocation(loc))

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err = s.cron.AddFunc(config.Cron, func() {
		s.runProcessing(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add payout cron job: %w", err)
	}

	s.cron.Start()
	s.enabled = true
	s.logger.Infof("Payout scheduler started with cron: %s (timezone: %s)", config.Cron, config.Timezone)
	return nil
}

// Stop stops the payout scheduler.
func (s *PayoutScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.cron != nil {
			s.cron.Stop()
		}
	})
	return nil
}

// IsEnabled returns whether the scheduler is enabled.
func (s *PayoutScheduler) IsEnabled() bool {
	return s.enabled
}

func (s *PayoutScheduler) runProcessing(ctx context.Context) {
	s.logger.Info("Payout scheduler: starting processing run")

	if err := s.payoutSvc.ProcessScheduledPayouts(ctx); err != nil {
		s.logger.WithError(err).Error("Payout scheduler: processing failed")
		return
	}

	s.logger.Info("Payout scheduler: processing run completed")
}

// EnvPayoutScheduleConfig loads payout scheduler configuration from environment variables.
func EnvPayoutScheduleConfig() PayoutScheduleConfig {
	enabled := os.Getenv("PAYOUT_SCHEDULER_ENABLED") == "true"
	cron := os.Getenv("PAYOUT_SCHEDULER_CRON")
	if cron == "" {
		cron = "0 2 * * 1" // Every Monday at 2 AM
	}
	timezone := os.Getenv("PAYOUT_SCHEDULER_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}
	return PayoutScheduleConfig{
		Enabled:  enabled,
		Cron:     cron,
		Timezone: timezone,
	}
}
