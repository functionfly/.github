package monitoring

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
)

// LocalRuntimeMetricsCollector collects runtime-specific performance metrics
type LocalRuntimeMetricsCollector struct {
	runtimeType     string
	process         *process.Process
	mu              sync.RWMutex
	activeRequests  int64
	totalRequests   int64
	lastThroughput  float64
	startTime       time.Time
	requestCount    int64
	throughputWindow time.Time
}

// RuntimeMetrics represents the current metrics for a local runtime
type RuntimeMetrics struct {
	Runtime           string    `json:"runtime"`
	MemoryUsage       MemoryStats `json:"memory_usage"`
	CPUUsage          float64   `json:"cpu_usage_percent"`
	ActiveConnections int       `json:"active_connections"`
	RequestThroughput float64   `json:"request_throughput_per_second"`
	Uptime            time.Duration `json:"uptime"`
	TotalRequests     int64     `json:"total_requests"`
	ErrorRate         float64   `json:"error_rate_percent"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Heap   uint64 `json:"heap_bytes"`
	Stack  uint64 `json:"stack_bytes"`
	System uint64 `json:"system_bytes"`
}

// NewLocalRuntimeMetricsCollector creates a new metrics collector for a local runtime
func NewLocalRuntimeMetricsCollector(runtimeType string) (*LocalRuntimeMetricsCollector, error) {
	// Get current process for system metrics
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("failed to create process monitor: %w", err)
	}

	collector := &LocalRuntimeMetricsCollector{
		runtimeType:     runtimeType,
		process:         proc,
		startTime:       time.Now(),
		throughputWindow: time.Now(),
	}

	return collector, nil
}

// Start begins periodic metrics collection
func (c *LocalRuntimeMetricsCollector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collectMetrics()
		}
	}
}

// collectMetrics gathers current runtime metrics
func (c *LocalRuntimeMetricsCollector) collectMetrics() {
	// Collect system metrics
	c.collectSystemMetrics()

	// Calculate request throughput
	c.calculateThroughput()

	// Record metrics
	c.recordPrometheusMetrics()
}

// collectSystemMetrics collects CPU and memory usage
func (c *LocalRuntimeMetricsCollector) collectSystemMetrics() {
	// Memory usage from process
	memInfo, err := c.process.MemoryInfo()
	if err != nil {
		logrus.WithError(err).Warn("Failed to collect memory info")
		return
	}

	// CPU usage percentage
	cpuPercent, err := c.process.CPUPercent()
	if err != nil {
		logrus.WithError(err).Warn("Failed to collect CPU info")
		return
	}

	// Record memory usage
	RecordLocalRuntimeMemoryUsage(c.runtimeType, memInfo.VMS, memInfo.RSS, memInfo.Data)

	// Record CPU usage
	RecordLocalRuntimeCPUUsage(c.runtimeType, cpuPercent)
}

// calculateThroughput calculates requests per second
func (c *LocalRuntimeMetricsCollector) calculateThroughput() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	windowDuration := now.Sub(c.throughputWindow)

	if windowDuration >= time.Minute {
		// Calculate throughput over the last minute
		requests := c.requestCount
		throughput := float64(requests) / windowDuration.Seconds()

		c.lastThroughput = throughput
		c.requestCount = 0
		c.throughputWindow = now

		// Record throughput
		RecordLocalRuntimeRequestThroughput(c.runtimeType, throughput)
	}
}

// recordPrometheusMetrics records current state to Prometheus metrics
func (c *LocalRuntimeMetricsCollector) recordPrometheusMetrics() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// This will be called periodically to update gauge metrics
	// Active connections and other gauges are updated when they change
}

// RecordRequestStart records the start of a request
func (c *LocalRuntimeMetricsCollector) RecordRequestStart(port int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeRequests++
	c.totalRequests++

	// Record active connections
	RecordLocalRuntimeActiveConnections(c.runtimeType, port, int(c.activeRequests))
}

// RecordRequestEnd records the completion of a request
func (c *LocalRuntimeMetricsCollector) RecordRequestEnd(port int, success bool, duration time.Duration, functionName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeRequests--
	if c.activeRequests < 0 {
		c.activeRequests = 0 // Safety check
	}

	c.requestCount++

	// Record execution metrics
	status := "success"
	if !success {
		status = "error"
	}
	RecordLocalRuntimeExecution(c.runtimeType, status, functionName)
	RecordLocalRuntimeExecutionDuration(c.runtimeType, functionName, duration)

	// Record active connections
	RecordLocalRuntimeActiveConnections(c.runtimeType, port, int(c.activeRequests))
}

// RecordError records a runtime error
func (c *LocalRuntimeMetricsCollector) RecordError(errorType, functionName string) {
	RecordLocalRuntimeError(c.runtimeType, errorType, functionName)
}

// GetCurrentMetrics returns the current runtime metrics
func (c *LocalRuntimeMetricsCollector) GetCurrentMetrics() RuntimeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get memory stats
	var memStats MemoryStats
	if memInfo, err := c.process.MemoryInfo(); err == nil {
		memStats = MemoryStats{
			Heap:   memInfo.VMS,
			Stack:  memInfo.RSS,
			System: memInfo.Data,
		}
	}

	// Get CPU usage
	var cpuUsage float64
	if cpuPercent, err := c.process.CPUPercent(); err == nil {
		cpuUsage = cpuPercent
	}

	return RuntimeMetrics{
		Runtime:           c.runtimeType,
		MemoryUsage:       memStats,
		CPUUsage:          cpuUsage,
		ActiveConnections: int(c.activeRequests),
		RequestThroughput: c.lastThroughput,
		Uptime:            time.Since(c.startTime),
		TotalRequests:     c.totalRequests,
		ErrorRate:         c.calculateErrorRate(),
	}
}

// calculateErrorRate calculates the current error rate (simplified version)
func (c *LocalRuntimeMetricsCollector) calculateErrorRate() float64 {
	// This is a simplified calculation - in production you'd want to track
	// errors over a sliding window
	// For now, return 0 as we need to implement proper error tracking
	return 0.0
}

// GetSystemInfo returns system-level information
func (c *LocalRuntimeMetricsCollector) GetSystemInfo() (map[string]interface{}, error) {
	// Get virtual memory info
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percent: %w", err)
	}

	return map[string]interface{}{
		"system_memory_total":     vmem.Total,
		"system_memory_used":      vmem.Used,
		"system_memory_available": vmem.Available,
		"system_memory_used_percent": vmem.UsedPercent,
		"cpu_cores":               len(cpuInfo),
		"cpu_usage_percent":       cpuPercent[0],
		"go_routines":             runtime.NumGoroutine(),
		"go_version":              runtime.Version(),
	}, nil
}