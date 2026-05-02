package sandbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultExecutionConfig(t *testing.T) {
	t.Parallel()

	config := DefaultExecutionConfig()

	assert.Equal(t, int64(256), config.MemoryMB)
	assert.Equal(t, 0.5, config.CPULimit)
	assert.Equal(t, 30, int(config.Timeout.Seconds()))
	assert.False(t, config.EnableNetwork)
	assert.Empty(t, config.Capabilities)
	assert.NotNil(t, config.EnvVars)
}

func TestNewSecureSandboxExecutor(t *testing.T) {
	t.Parallel()

	executor, err := NewSecureSandboxExecutor()
	require.NoError(t, err)
	assert.NotNil(t, executor)
	// gVisor may or may not be available depending on installation
	_ = executor.IsGvisorAvailable() // Should not panic
}

func TestNewSecureSandboxExecutorWithConfig(t *testing.T) {
	t.Parallel()

	config := &ExecutionConfig{
		MemoryMB:   512,
		CPULimit:   1.0,
		Timeout:    60,
		EnableNetwork: false,
		Capabilities: []string{},
		EnvVars:   map[string]string{},
	}

	executor, err := NewSecureSandboxExecutorWithConfig(config)
	require.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestTernaryString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "true", ternaryString(true, "true", "false"))
	assert.Equal(t, "false", ternaryString(false, "true", "false"))
}