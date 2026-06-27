package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/receipt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReceiptMilestoneScheduler(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)
	assert.NotNil(t, s)
}

func TestStart_DisabledIsNoop(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	cfg.Enabled = false
	cfg.Logger = logrus.New()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)

	ctx := context.Background()
	err := s.Start(ctx)
	require.NoError(t, err)
	s.mu.Lock()
	running := s.isRunning
	s.mu.Unlock()
	assert.False(t, running)
}

func TestStart_EnabledSchedulesCron(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	cfg.Enabled = true
	cfg.Cron = "0 0 * * *"
	cfg.Logger = logrus.New()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)

	ctx := context.Background()
	err := s.Start(ctx)
	require.NoError(t, err)
	s.mu.Lock()
	running := s.isRunning
	s.mu.Unlock()
	assert.True(t, running)
	s.Stop()
}

func TestStart_Idempotent(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	cfg.Enabled = true
	cfg.Cron = "0 0 * * *"
	cfg.Logger = logrus.New()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)

	ctx := context.Background()
	err1 := s.Start(ctx)
	require.NoError(t, err1)
	err2 := s.Start(ctx)
	require.NoError(t, err2)
	s.Stop()
}

func TestStop_NotRunning(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	cfg.Logger = logrus.New()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)
	s.Stop()
}

func TestStop_MultipleTimes(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	cfg.Enabled = true
	cfg.Cron = "0 0 * * *"
	cfg.Logger = logrus.New()
	w := &receipt.Milestone{}
	s := NewReceiptMilestoneScheduler(w, cfg)
	ctx := context.Background()
	_ = s.Start(ctx)
	s.Stop()
	s.Stop()
}

func TestDefaultReceiptMilestoneSchedulerConfig(t *testing.T) {
	cfg := DefaultReceiptMilestoneSchedulerConfig()
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "0 3 * * *", cfg.Cron)
	assert.Equal(t, 48*time.Hour, cfg.Lookback)
	assert.NotNil(t, cfg.Logger)
}
