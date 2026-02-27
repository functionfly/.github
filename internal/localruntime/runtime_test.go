package localruntime

import (
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRuntimeID(t *testing.T) {
	id1, err := generateRuntimeID()
	require.NoError(t, err)
	assert.NotEmpty(t, id1)
	assert.Contains(t, id1, "runtime-")
	assert.Len(t, id1, 24) // "runtime-" + 16 hex chars

	id2, err := generateRuntimeID()
	require.NoError(t, err)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should generate unique IDs
}

func TestNewRuntime(t *testing.T) {
	// Create a test manifest
	testManifest := &manifest.Manifest{
		Name:    "test-function",
		Runtime: "node18",
	}

	// Test creating a new runtime with a function to avoid file loading
	testFunc := func(interface{}) (interface{}, error) { return "test", nil }
	runtime, err := New(testManifest, WithFunction(testFunc))
	require.NoError(t, err)
	assert.NotNil(t, runtime)
	assert.NotEmpty(t, runtime.runtimeID)
	assert.Contains(t, runtime.runtimeID, "runtime-")
	assert.NotNil(t, runtime.shutdownChan)
	assert.Equal(t, testManifest, runtime.manifest)
	assert.NotNil(t, runtime.function)
}

func TestRuntimeOptions(t *testing.T) {
	// Create a test manifest
	testManifest := &manifest.Manifest{
		Name:    "test-function",
		Runtime: "node18",
	}

	// Test WithFunction option
	testFunc := func(interface{}) (interface{}, error) { return "test", nil }
	runtime, err := New(testManifest, WithFunction(testFunc))
	require.NoError(t, err)
	assert.NotNil(t, runtime.function)
	assert.Equal(t, testManifest, runtime.manifest)
}

func TestResponseStruct(t *testing.T) {
	response := Response{
		Result:   "test result",
		Input:    "test input",
		Runtime:  "node18",
		ExecTime: 150,
		Error:    "",
	}

	assert.Equal(t, "test result", response.Result)
	assert.Equal(t, "test input", response.Input)
	assert.Equal(t, "node18", response.Runtime)
	assert.Equal(t, int64(150), response.ExecTime)
	assert.Empty(t, response.Error)
}

func TestMonitoringTypes(t *testing.T) {
	// Test MemoryStats struct
	stats := monitoring.MemoryStats{
		Heap:   1024 * 1024, // 1MB
		Stack:  512 * 1024,  // 512KB
		System: 2048 * 1024, // 2MB
	}

	assert.Equal(t, uint64(1024*1024), stats.Heap)
	assert.Equal(t, uint64(512*1024), stats.Stack)
	assert.Equal(t, uint64(2048*1024), stats.System)
}

func TestRuntimeMetrics(t *testing.T) {
	// Test RuntimeMetrics struct
	metrics := monitoring.RuntimeMetrics{
		Runtime:           "node18",
		MemoryUsage:       monitoring.MemoryStats{Heap: 1024, Stack: 512, System: 2048},
		CPUUsage:          45.5,
		ActiveConnections: 3,
		RequestThroughput: 12.5,
		Uptime:            5 * time.Minute,
		TotalRequests:     150,
		ErrorRate:         2.5,
	}

	assert.Equal(t, "node18", metrics.Runtime)
	assert.Equal(t, uint64(1024), metrics.MemoryUsage.Heap)
	assert.Equal(t, 45.5, metrics.CPUUsage)
	assert.Equal(t, 3, metrics.ActiveConnections)
	assert.Equal(t, 12.5, metrics.RequestThroughput)
	assert.Equal(t, 5*time.Minute, metrics.Uptime)
	assert.Equal(t, int64(150), metrics.TotalRequests)
	assert.Equal(t, 2.5, metrics.ErrorRate)
}

func TestLocalRuntimeHealth(t *testing.T) {
	// Test LocalRuntimeHealth struct
	health := storage.LocalRuntimeHealth{
		RuntimeInstanceID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, // mock UUID
		Timestamp:         time.Now(),
		Status:            "healthy",
		ResponseTime:      50 * time.Millisecond,
		Checks: map[string]interface{}{
			"function_loaded": map[string]interface{}{
				"status": "healthy",
				"detail": "Function code is loaded",
			},
			"server_running": map[string]interface{}{
				"status": "healthy",
				"detail": "HTTP server is running",
			},
		},
	}

	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, 50*time.Millisecond, health.ResponseTime)
	assert.NotNil(t, health.Checks)
	assert.Len(t, health.Checks, 2)
}