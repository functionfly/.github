package scheduler

import (
	"context"
	"os"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripeMeterConfigDefaults(t *testing.T) {
	cfg := DefaultStripeMeterUsageSchedulerConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "0 * * * *", cfg.Cron)
	assert.True(t, cfg.Enabled)
	assert.False(t, cfg.DryRun)
	assert.Equal(t, 1, cfg.LookbackHours)
	assert.Equal(t, 100, cfg.MaxTenantsPerRun)
	assert.Equal(t, 3, cfg.ReportRetryAttempts)
	assert.Equal(t, 5, cfg.ReportRetryDelaySeconds)
}

func TestStripeMeterConfigEnvOverrides(t *testing.T) {
	os.Setenv("STRIPE_METER_SCHEDULER_ENABLED", "false")
	os.Setenv("STRIPE_METER_SCHEDULER_CRON", "0 0 * * *")
	os.Setenv("STRIPE_METER_SCHEDULER_DRY_RUN", "true")
	os.Setenv("STRIPE_METER_SCHEDULER_LOOKBACK_HOURS", "2")
	os.Setenv("STRIPE_METER_SCHEDULER_MAX_TENANTS", "50")
	defer func() {
		os.Unsetenv("STRIPE_METER_SCHEDULER_ENABLED")
		os.Unsetenv("STRIPE_METER_SCHEDULER_CRON")
		os.Unsetenv("STRIPE_METER_SCHEDULER_DRY_RUN")
		os.Unsetenv("STRIPE_METER_SCHEDULER_LOOKBACK_HOURS")
		os.Unsetenv("STRIPE_METER_SCHEDULER_MAX_TENANTS")
	}()

	cfg := LoadStripeMeterUsageSchedulerConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "0 0 * * *", cfg.Cron)
	assert.False(t, cfg.Enabled)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, 2, cfg.LookbackHours)
	assert.Equal(t, 50, cfg.MaxTenantsPerRun)
}

func TestStripeMeterSchedulerStartDisabled(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: &StripeMeterUsageSchedulerConfig{Enabled: false},
		logger: logrus.New(),
	}

	ctx := context.Background()
	err := s.Start(ctx)
	require.NoError(t, err)
	assert.False(t, s.IsRunning())
}

func TestStripeMeterSchedulerStartInvalidCron(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: &StripeMeterUsageSchedulerConfig{Enabled: true, Cron: "invalid"},
		logger: logrus.New(),
	}

	ctx := context.Background()
	err := s.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron expression")
	assert.False(t, s.IsRunning())
}

func TestStripeMeterSchedulerStopNotRunning(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: DefaultStripeMeterUsageSchedulerConfig(),
		logger: logrus.New(),
	}
	s.Stop()
	assert.False(t, s.IsRunning())
}

func TestStripeMeterSchedulerGetSchedule(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: &StripeMeterUsageSchedulerConfig{
			Cron:           "0 * * * *",
			Enabled:        true,
			DryRun:         true,
			LookbackHours:  2,
			MaxTenantsPerRun: 50,
		},
		logger: logrus.New(),
	}

	schedule := s.GetSchedule()
	assert.Equal(t, true, schedule["enabled"])
	assert.Equal(t, "0 * * * *", schedule["cron"])
	assert.Equal(t, true, schedule["dry_run"])
	assert.Equal(t, 2, schedule["lookback_hours"])
	assert.Equal(t, 50, schedule["max_tenants"])
}

func TestStripeMeterSchedulerIsRunning(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: DefaultStripeMeterUsageSchedulerConfig(),
		logger: logrus.New(),
	}

	assert.False(t, s.IsRunning())

	s.mu.Lock()
	s.isRunning = true
	s.mu.Unlock()

	assert.True(t, s.IsRunning())
}

func TestStripeMeterSchedulerTriggerNowNoop(t *testing.T) {
	s := &StripeMeterUsageScheduler{
		cron:   cron.New(),
		config: &StripeMeterUsageSchedulerConfig{Enabled: true, Cron: "0 * * * *"},
		logger: logrus.New(),
	}

	ctx := context.Background()
	err := s.TriggerNow(ctx)
	require.NoError(t, err)
}