package scheduler

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConsciousnessSchedulerConfig(t *testing.T) {
	config := DefaultConsciousnessSchedulerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "*/30 * * * *", config.Cron)
	assert.True(t, config.Enabled)
	assert.Equal(t, 10*time.Minute, config.Timeout)
}

func TestLoadConsciousnessSchedulerConfig(t *testing.T) {
	os.Setenv("CONSCIOUSNESS_CRON", "0 */6 * * *")
	os.Setenv("CONSCIOUSNESS_ENABLED", "false")
	os.Setenv("CONSCIOUSNESS_TIMEOUT", "30m")
	defer func() {
		os.Unsetenv("CONSCIOUSNESS_CRON")
		os.Unsetenv("CONSCIOUSNESS_ENABLED")
		os.Unsetenv("CONSCIOUSNESS_TIMEOUT")
	}()

	config := LoadConsciousnessSchedulerConfig()

	assert.Equal(t, "0 */6 * * *", config.Cron)
	assert.False(t, config.Enabled)
	assert.Equal(t, 30*time.Minute, config.Timeout)
}

func TestLoadConsciousnessSchedulerConfig_InvalidBool(t *testing.T) {
	os.Setenv("CONSCIOUSNESS_ENABLED", "not-a-bool")
	defer os.Unsetenv("CONSCIOUSNESS_ENABLED")

	config := LoadConsciousnessSchedulerConfig()
	assert.True(t, config.Enabled)
}

func TestLoadConsciousnessSchedulerConfig_InvalidTimeout(t *testing.T) {
	os.Setenv("CONSCIOUSNESS_TIMEOUT", "invalid")
	defer os.Unsetenv("CONSCIOUSNESS_TIMEOUT")

	config := LoadConsciousnessSchedulerConfig()
	assert.Equal(t, 10*time.Minute, config.Timeout)
}

func TestConsciousnessSchedulerConfig_Constants(t *testing.T) {
	assert.Equal(t, "*/30 * * * *", DefaultConsciousnessCron)
	assert.Equal(t, 10*time.Minute, DefaultConsciousnessTimeout)
}

func TestNewConsciousnessSchedulerWithConfig(t *testing.T) {
	config := &ConsciousnessSchedulerConfig{
		Cron:    "0 */6 * * *",
		Enabled: true,
		Timeout: 30 * time.Minute,
	}

	scheduler := NewConsciousnessSchedulerWithConfig(nil, config)

	assert.NotNil(t, scheduler)
	assert.Equal(t, config, scheduler.config)
	assert.NotNil(t, scheduler.cron)
	assert.NotNil(t, scheduler.engine)
	assert.NotNil(t, scheduler.stopCh)
	assert.NotNil(t, scheduler.stoppedCh)
}

func TestConsciousnessScheduler_IsRunning(t *testing.T) {
	scheduler := &ConsciousnessScheduler{
		runningCh: make(chan struct{}, 1),
	}

	assert.False(t, scheduler.IsRunning())

	scheduler.runningCh <- struct{}{}
	assert.True(t, scheduler.IsRunning())

	<-scheduler.runningCh
	assert.False(t, scheduler.IsRunning())
}

func TestConsciousnessScheduler_GetStatus(t *testing.T) {
	scheduler := &ConsciousnessScheduler{
		config: &ConsciousnessSchedulerConfig{
			Cron:    "*/30 * * * *",
			Enabled: true,
			Timeout: 10 * time.Minute,
		},
		cron:      nil,
		runningCh: make(chan struct{}, 1),
	}

	status := scheduler.GetStatus()

	assert.Equal(t, true, status["enabled"])
	assert.Equal(t, "*/30 * * * *", status["cron"])
	assert.Equal(t, "10m0s", status["timeout"])
	assert.Equal(t, "unknown", status["next_run"])
	assert.NotNil(t, status["is_running"])
}