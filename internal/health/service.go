// Package health provides production-ready health monitoring services
package health

import (
	"context"
	"database/sql"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Service provides health monitoring services
type Service struct {
	logger      *logrus.Logger
	redisClient *redis.Client
	db          *sql.DB

	cpuMu          sync.Mutex
	lastCPUSeconds float64
	lastCPUCheckAt time.Time
}

// NewService creates a new health monitoring service
func NewService(logger *logrus.Logger, redisClient *redis.Client) *Service {
	return &Service{
		logger:     logger,
		redisClient: redisClient,
	}
}

// SetDatabase provides an optional DB handle for health checks.
func (s *Service) SetDatabase(db *sql.DB) {
	s.db = db
}

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]HealthCheck `json:"services"`
	Checks    []HealthCheck          `json:"checks"`
}

// CheckMemoryUsage performs memory usage health check
func (s *Service) CheckMemoryUsage() HealthCheck {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage metrics
	allocMB := float64(m.Alloc) / 1024 / 1024
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	heapAllocMB := float64(m.HeapAlloc) / 1024 / 1024
	heapSysMB := float64(m.HeapSys) / 1024 / 1024

	// Calculate heap utilization percentage
	heapUtilization := float64(0)
	if m.HeapSys > 0 {
		heapUtilization = float64(m.HeapAlloc) / float64(m.HeapSys) * 100
	}

	// Determine status based on heap utilization
	status := HealthStatusHealthy
	description := "Memory usage within acceptable limits"

	if heapUtilization > 90 {
		status = HealthStatusUnhealthy
		description = "High memory usage detected"
	} else if heapUtilization > 80 {
		status = HealthStatusDegraded
		description = "Elevated memory usage"
	}

	return HealthCheck{
		Name:        "memory_usage",
		Status:      status,
		Description: description,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"current_alloc_mb":     allocMB,
			"total_alloc_mb":       totalAllocMB,
			"system_memory_mb":     sysMB,
			"heap_alloc_mb":        heapAllocMB,
			"heap_sys_mb":          heapSysMB,
			"heap_utilization_pct": heapUtilization,
			"gc_cycles":            m.NumGC,
			"next_gc_target_mb":    float64(m.NextGC) / 1024 / 1024,
		},
	}
}

// CheckDiskSpace performs disk space health check
func (s *Service) CheckDiskSpace() HealthCheck {
	// Get disk usage statistics for the current working directory
	wd, err := syscall.Getwd()
	if err != nil {
		return HealthCheck{
			Name:        "disk_space",
			Status:      HealthStatusUnknown,
			Description: "Unable to determine current working directory",
			Timestamp:   time.Now(),
			Error:       err.Error(),
		}
	}

	var stat syscall.Statfs_t
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return HealthCheck{
			Name:        "disk_space",
			Status:      HealthStatusUnknown,
			Description: "Unable to get filesystem statistics",
			Timestamp:   time.Now(),
			Error:       err.Error(),
		}
	}

	// Calculate disk usage metrics
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - availableBytes

	totalGB := float64(totalBytes) / 1024 / 1024 / 1024
	usedGB := float64(usedBytes) / 1024 / 1024 / 1024
	availableGB := float64(availableBytes) / 1024 / 1024 / 1024

	// Calculate usage percentage
	usagePercent := float64(0)
	if totalBytes > 0 {
		usagePercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	// Determine status based on usage percentage
	status := HealthStatusHealthy
	description := "Disk space usage within acceptable limits"

	if usagePercent > 95 {
		status = HealthStatusUnhealthy
		description = "Critical disk space usage - immediate action required"
	} else if usagePercent > 90 {
		status = HealthStatusDegraded
		description = "High disk space usage detected"
	} else if usagePercent > 80 {
		status = HealthStatusDegraded
		description = "Elevated disk space usage"
	}

	return HealthCheck{
		Name:        "disk_space",
		Status:      status,
		Description: description,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"total_gb":      totalGB,
			"used_gb":       usedGB,
			"available_gb":  availableGB,
			"usage_percent": usagePercent,
			"filesystem":    wd,
		},
	}
}

// CheckGoroutines performs goroutine health check
func (s *Service) CheckGoroutines() HealthCheck {
	numGoroutines := runtime.NumGoroutine()

	status := HealthStatusHealthy
	description := "Goroutine count within acceptable limits"

	if numGoroutines > 10000 {
		status = HealthStatusUnhealthy
		description = "High goroutine count detected"
	} else if numGoroutines > 5000 {
		status = HealthStatusDegraded
		description = "Elevated goroutine count"
	}

	return HealthCheck{
		Name:        "goroutines",
		Status:      status,
		Description: description,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"count": numGoroutines,
		},
	}
}

// CheckCPUUsage performs CPU usage health check
func (s *Service) CheckCPUUsage() HealthCheck {
	// Get CPU usage statistics
	var rusage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	if err != nil {
		return HealthCheck{
			Name:        "cpu_usage",
			Status:      HealthStatusUnknown,
			Description: "Unable to get CPU usage statistics",
			Timestamp:   time.Now(),
			Error:       err.Error(),
		}
	}

	// Calculate CPU usage (user + system time)
	userTime := float64(rusage.Utime.Sec) + float64(rusage.Utime.Usec)/1000000
	systemTime := float64(rusage.Stime.Sec) + float64(rusage.Stime.Usec)/1000000
	totalTime := userTime + systemTime

	now := time.Now()

	usagePercent := float64(-1)
	windowSeconds := float64(0)
	numCPU := runtime.NumCPU()

	s.cpuMu.Lock()
	if !s.lastCPUCheckAt.IsZero() {
		deltaCPU := totalTime - s.lastCPUSeconds
		deltaWall := now.Sub(s.lastCPUCheckAt).Seconds()
		if deltaCPU >= 0 && deltaWall > 0 && numCPU > 0 {
			// deltaCPU/deltaWall is "CPU cores used" on average over the interval.
			// Divide by NumCPU to get % of total machine capacity.
			usagePercent = (deltaCPU / deltaWall) / float64(numCPU) * 100
			if usagePercent < 0 {
				usagePercent = 0
			}
			if usagePercent > 100 {
				usagePercent = 100
			}
			windowSeconds = deltaWall
		}
	}
	s.lastCPUSeconds = totalTime
	s.lastCPUCheckAt = now
	s.cpuMu.Unlock()

	status := HealthStatusHealthy
	description := "CPU usage within acceptable limits"

	if usagePercent >= 0 {
		if usagePercent > 90 {
			status = HealthStatusUnhealthy
			description = "High CPU usage detected"
		} else if usagePercent > 75 {
			status = HealthStatusDegraded
			description = "Elevated CPU usage"
		}
	}

	return HealthCheck{
		Name:        "cpu_usage",
		Status:      status,
		Description: description,
		Timestamp:   now,
		Details: map[string]interface{}{
			"user_time_seconds":   userTime,
			"system_time_seconds": systemTime,
			"total_time_seconds":  totalTime,
			"usage_percent":       usagePercent,
			"sample_window_sec":   windowSeconds,
			"num_cpu":             numCPU,
		},
	}
}

// CheckDatabaseHealth performs database health check
func (s *Service) CheckDatabaseHealth(ctx context.Context) HealthCheck {
	if s.db == nil {
		return HealthCheck{
			Name:        "database",
			Status:      HealthStatusUnknown,
			Description: "Database not configured for health service",
			Timestamp:   time.Now(),
			Error:       "db is nil",
			Details: map[string]interface{}{
				"type": "postgresql",
			},
		}
	}

	status := HealthStatusHealthy
	description := "Database is healthy"

	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(pingCtx); err != nil {
		return HealthCheck{
			Name:        "database",
			Status:      HealthStatusUnhealthy,
			Description: "Database ping failed",
			Timestamp:   time.Now(),
			Error:       err.Error(),
			Details: map[string]interface{}{
				"type":           "postgresql",
				"latency_ms":     time.Since(start).Milliseconds(),
				"check":          "ping",
				"timeout_ms":     2000,
				"context_cancel": pingCtx.Err() != nil,
			},
		}
	}

	// Basic query to validate connectivity beyond ping.
	var one int
	if err := s.db.QueryRowContext(pingCtx, "SELECT 1").Scan(&one); err != nil || one != 1 {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = "unexpected result"
		}
		return HealthCheck{
			Name:        "database",
			Status:      HealthStatusDegraded,
			Description: "Database query check failed",
			Timestamp:   time.Now(),
			Error:       errMsg,
			Details: map[string]interface{}{
				"type":       "postgresql",
				"latency_ms": time.Since(start).Milliseconds(),
				"check":      "select_1",
			},
		}
	}

	return HealthCheck{
		Name:        "database",
		Status:      status,
		Description: description,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"type": "postgresql",
			"latency_ms": time.Since(start).Milliseconds(),
			"check":      "ping+select_1",
		},
	}
}

// CheckRedisHealth performs Redis health check
func (s *Service) CheckRedisHealth(ctx context.Context) HealthCheck {
	if s.redisClient == nil {
		return HealthCheck{
			Name:        "redis",
			Status:      HealthStatusUnhealthy,
			Description: "Redis client not configured",
			Timestamp:   time.Now(),
			Error:       "redis client is nil",
			Details: map[string]interface{}{
				"type": "redis",
			},
		}
	}

	// Perform actual Redis PING
	start := time.Now()
	result, err := s.redisClient.Ping(ctx).Result()
	duration := time.Since(start)

	if err != nil {
		s.logger.WithError(err).Error("Redis health check failed")
		return HealthCheck{
			Name:        "redis",
			Status:      HealthStatusUnhealthy,
			Description: "Redis ping failed",
			Timestamp:   time.Now(),
			Error:       err.Error(),
			Details: map[string]interface{}{
				"type":         "redis",
				"latency_ms":   duration.Milliseconds(),
				"ping_result":  result,
			},
		}
	}

	return HealthCheck{
		Name:        "redis",
		Status:      HealthStatusHealthy,
		Description: "Redis is healthy",
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"type":        "redis",
			"latency_ms":  duration.Milliseconds(),
			"ping_result": result,
		},
	}
}

// GetSystemHealth returns the overall system health
func (s *Service) GetSystemHealth(ctx context.Context) SystemHealth {
	checks := []HealthCheck{
		s.CheckMemoryUsage(),
		s.CheckDiskSpace(),
		s.CheckGoroutines(),
		s.CheckCPUUsage(),
		s.CheckDatabaseHealth(ctx),
		s.CheckRedisHealth(ctx),
	}

	// Determine overall status
	overallStatus := HealthStatusHealthy
	for _, check := range checks {
		if check.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if check.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return SystemHealth{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services: map[string]HealthCheck{
			"database": s.CheckDatabaseHealth(ctx),
			"redis":    s.CheckRedisHealth(ctx),
		},
		Checks: checks,
	}
}

// LogHealthCheck logs a health check result
func (s *Service) LogHealthCheck(check HealthCheck) {
	fields := logrus.Fields{
		"check_name":  check.Name,
		"status":      string(check.Status),
		"description": check.Description,
		"timestamp":   check.Timestamp,
	}

	if check.Error != "" {
		fields["error"] = check.Error
	}

	if check.Status == HealthStatusHealthy {
		s.logger.WithFields(fields).Info("Health check passed")
	} else if check.Status == HealthStatusDegraded {
		s.logger.WithFields(fields).Warn("Health check degraded")
	} else {
		s.logger.WithFields(fields).Error("Health check failed")
	}
}
