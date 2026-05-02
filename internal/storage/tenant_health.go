package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TenantDBHealthChecker provides health monitoring for all tenant databases
type TenantDBHealthChecker struct {
	provisioner *TenantDBProvisioner
	poolManager *TenantPoolManager
	config      *TenantDatabaseConfig
	// per-tenant health state
	healthStatus sync.Map // map[uuid.UUID]*TenantHealthStatus
	// monitoring config
	checkInterval   time.Duration
	checkTimeout    time.Duration
	maxFailures     int
	stopChan        chan struct{}
}

// TenantHealthStatus holds health information for a tenant's database
type TenantHealthStatus struct {
	TenantID       uuid.UUID
	Status         HealthStatus
	LatencyMs      int64
	Failures        int
	LastCheck       time.Time
	LastSuccess     time.Time
	LastFailure     time.Time
	ErrorMessage    string
	IsolateOnFailure bool // If true, isolate the tenant DB on repeated failures
}

// HealthStatus represents the health state of a tenant database
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// NewTenantDBHealthChecker creates a new tenant health checker
func NewTenantDBHealthChecker(
	provisioner *TenantDBProvisioner,
	poolManager *TenantPoolManager,
	config *TenantDatabaseConfig,
) *TenantDBHealthChecker {
	return &TenantDBHealthChecker{
		provisioner:     provisioner,
		poolManager:     poolManager,
		config:          config,
		checkInterval:   30 * time.Second,
		checkTimeout:    5 * time.Second,
		maxFailures:     3,
		stopChan:        make(chan struct{}),
	}
}

// Start begins background health monitoring for all tenant databases
func (h *TenantDBHealthChecker) Start(ctx context.Context) error {
	if !h.config.Enabled {
		logrus.Info("TenantDBHealthChecker: disabled")
		return nil
	}

	// Initial scan of all tenant databases
	if err := h.scanTenantDatabases(ctx); err != nil {
		logrus.Warnf("Initial tenant database scan failed: %v", err)
	}

	// Start periodic health checks
	go h.healthCheckLoop()

	logrus.Info("TenantDBHealthChecker: started")
	return nil
}

// Stop stops the health checker
func (h *TenantDBHealthChecker) Stop() {
	close(h.stopChan)
	logrus.Info("TenantDBHealthChecker: stopped")
}

// scanTenantDatabases discovers all tenant databases and initializes health tracking
func (h *TenantDBHealthChecker) scanTenantDatabases(ctx context.Context) error {
	entries, err := h.provisioner.ListTenantDatabases(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tenant databases: %w", err)
	}

	for _, entry := range entries {
		status := &TenantHealthStatus{
			TenantID: entry.TenantID,
			Status:   HealthStatusUnknown,
		}
		h.healthStatus.Store(entry.TenantID, status)
	}

	logrus.Infof("TenantDBHealthChecker: tracking %d tenant databases", len(entries))
	return nil
}

// healthCheckLoop runs continuous health checks
func (h *TenantDBHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.performHealthChecks()
		case <-h.stopChan:
			return
		}
	}
}

// performHealthChecks checks health of all tracked tenant databases
func (h *TenantDBHealthChecker) performHealthChecks() {
	ctx := context.Background()

	h.healthStatus.Range(func(key, value interface{}) bool {
		tenantID := key.(uuid.UUID)
		status := value.(*TenantHealthStatus)

		go h.checkTenantHealth(ctx, tenantID, status)

		return true
	})
}

// checkTenantHealth performs a single health check for a tenant
func (h *TenantDBHealthChecker) checkTenantHealth(ctx context.Context, tenantID uuid.UUID, status *TenantHealthStatus) {
	start := time.Now()

	// Create a timeout context
	checkCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()

	// Get the tenant pool
	pool, err := h.poolManager.GetPool(checkCtx, tenantID)
	if err != nil {
		h.recordFailure(status, tenantID, fmt.Sprintf("pool unavailable: %v", err))
		return
	}

	// Perform health check query
	err = pool.Ping(checkCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		h.recordFailure(status, tenantID, fmt.Sprintf("ping failed: %v", err))
		return
	}

	// Record success
	h.recordSuccess(status, tenantID, latency)
}

// recordFailure records a health check failure
func (h *TenantDBHealthChecker) recordFailure(status *TenantHealthStatus, tenantID uuid.UUID, errMsg string) {
	status.Failures++
	status.LastFailure = time.Now()
	status.ErrorMessage = errMsg

	if status.Failures >= h.maxFailures {
		if status.Status != HealthStatusUnhealthy {
			status.Status = HealthStatusUnhealthy
			logrus.Errorf("Tenant %s database health: UNHEALTHY (failures: %d, error: %s)",
				tenantID, status.Failures, errMsg)

			// Trigger automatic isolation if configured
			if h.config.Enabled && status.IsolateOnFailure {
				h.isolateTenant(tenantID, errMsg)
			}
		}
	} else if status.Failures >= 1 {
		status.Status = HealthStatusDegraded
	}

	status.LastCheck = time.Now()

	// Update registry status if unhealthy
	if status.Status == HealthStatusUnhealthy {
		registry := NewTenantDBRegistry(nil, "") // Will be injected in production
		if registry != nil {
			_ = registry.UpdateStatus(context.Background(), tenantID, "unhealthy")
		}
	}

	// Store updated status
	h.healthStatus.Store(tenantID, status)
}

// recordSuccess records a successful health check
func (h *TenantDBHealthChecker) recordSuccess(status *TenantHealthStatus, tenantID uuid.UUID, latencyMs int64) {
	status.LatencyMs = latencyMs
	status.LastSuccess = time.Now()
	status.LastCheck = time.Now()
	status.Failures = 0
	status.ErrorMessage = ""

	// Upgrade from degraded/unhealthy to healthy if all checks pass
	if status.Failures == 0 && status.Status != HealthStatusHealthy {
		logrus.Infof("Tenant %s database health: RECOVERED (latency: %dms)", tenantID, latencyMs)
		status.Status = HealthStatusHealthy

		// Update registry status
		if registry := NewTenantDBRegistry(nil, ""); registry != nil {
			_ = registry.UpdateStatus(context.Background(), tenantID, "active")
		}
	}

	// Store updated status
	h.healthStatus.Store(tenantID, status)
}

// isolateTenant isolates an unhealthy tenant database (stops accepting connections)
func (h *TenantDBHealthChecker) isolateTenant(tenantID uuid.UUID, reason string) {
	logrus.Warnf("Isolating tenant %s database: %s", tenantID, reason)

	// Close the pool to stop accepting new connections
	_ = h.poolManager.ClosePool(tenantID)

	// Mark as suspended in registry
	registry := NewTenantDBRegistry(nil, "")
	if registry != nil {
		_ = registry.UpdateStatus(context.Background(), tenantID, "isolated")
	}

	// Emit event for alerting (in production, this would go to event system)
	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID.String(),
		"reason":    reason,
	}).Error("Tenant database isolated due to health failure")
}

// GetTenantHealth returns the health status of a specific tenant
func (h *TenantDBHealthChecker) GetTenantHealth(tenantID uuid.UUID) (*TenantHealthStatus, error) {
	if status, ok := h.healthStatus.Load(tenantID); ok {
		return status.(*TenantHealthStatus), nil
	}
	return nil, fmt.Errorf("health status not found for tenant %s", tenantID)
}

// GetAllTenantHealth returns health status for all tracked tenants
func (h *TenantDBHealthChecker) GetAllTenantHealth() []*TenantHealthStatus {
	var statuses []*TenantHealthStatus

	h.healthStatus.Range(func(key, value interface{}) bool {
		statuses = append(statuses, value.(*TenantHealthStatus))
		return true
	})

	return statuses
}

// GetUnhealthyTenants returns all tenants with unhealthy databases
func (h *TenantDBHealthChecker) GetUnhealthyTenants() []*TenantHealthStatus {
	var unhealthy []*TenantHealthStatus

	h.healthStatus.Range(func(key, value interface{}) bool {
		status := value.(*TenantHealthStatus)
		if status.Status == HealthStatusUnhealthy {
			unhealthy = append(unhealthy, status)
		}
		return true
	})

	return unhealthy
}

// CheckTenantExplicit performs an immediate health check for a specific tenant
func (h *TenantDBHealthChecker) CheckTenantExplicit(ctx context.Context, tenantID uuid.UUID) (*TenantHealthStatus, error) {
	status := &TenantHealthStatus{
		TenantID: tenantID,
	}

	h.checkTenantHealth(ctx, tenantID, status)

	return status, nil
}

// TenantHealthReport provides a comprehensive health report for all tenant databases
type TenantHealthReport struct {
	Timestamp       time.Time
	TotalTenants    int
	HealthyCount    int
	DegradedCount   int
	UnhealthyCount  int
	AvgLatencyMs    int64
	HealthByTenant  map[uuid.UUID]*TenantHealthStatus
}

// GenerateReport creates a comprehensive health report
func (h *TenantDBHealthChecker) GenerateReport() *TenantHealthReport {
	report := &TenantHealthReport{
		Timestamp:      time.Now(),
		HealthByTenant: make(map[uuid.UUID]*TenantHealthStatus),
	}

	var totalLatency int64
	var degraded, unhealthy int

	h.healthStatus.Range(func(key, value interface{}) bool {
		tenantID := key.(uuid.UUID)
		status := value.(*TenantHealthStatus)

		report.HealthByTenant[tenantID] = status
		report.TotalTenants++

		switch status.Status {
		case HealthStatusHealthy:
			report.HealthyCount++
		case HealthStatusDegraded:
			report.DegradedCount++
			degraded++
		case HealthStatusUnhealthy:
			report.UnhealthyCount++
			unhealthy++
		}

		totalLatency += status.LatencyMs

		return true
	})

	if report.TotalTenants > 0 {
		report.AvgLatencyMs = totalLatency / int64(report.TotalTenants)
	}

	return report
}

// HealthCheckResult represents a single health check result
type HealthCheckResult struct {
	TenantID   uuid.UUID
	Healthy    bool
	LatencyMs   int64
	Error       string
	CheckedAt   time.Time
}

// CheckTenantsConcurrently performs health checks on multiple tenants in parallel
func (h *TenantDBHealthChecker) CheckTenantsConcurrently(ctx context.Context, tenantIDs []uuid.UUID) []HealthCheckResult {
	results := make([]HealthCheckResult, len(tenantIDs))
	var wg sync.WaitGroup

	for i, tenantID := range tenantIDs {
		wg.Add(1)
		go func(idx int, id uuid.UUID) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
			defer cancel()

			result := HealthCheckResult{
				TenantID: id,
				CheckedAt: time.Now(),
			}

			pool, err := h.poolManager.GetPool(checkCtx, id)
			if err != nil {
				result.Healthy = false
				result.Error = err.Error()
				results[idx] = result
				return
			}

			start := time.Now()
			err = pool.Ping(checkCtx)
			result.LatencyMs = time.Since(start).Milliseconds()

			if err != nil {
				result.Healthy = false
				result.Error = err.Error()
			} else {
				result.Healthy = true
			}

			results[idx] = result
		}(i, tenantID)
	}

	wg.Wait()
	return results
}

// TenantDBHealthCheckerInterface defines the interface for tenant health checking
type TenantDBHealthCheckerInterface interface {
	Start(ctx context.Context) error
	Stop()
	GetTenantHealth(tenantID uuid.UUID) (*TenantHealthStatus, error)
	GetAllTenantHealth() []*TenantHealthStatus
	GetUnhealthyTenants() []*TenantHealthStatus
	CheckTenantExplicit(ctx context.Context, tenantID uuid.UUID) (*TenantHealthStatus, error)
	GenerateReport() *TenantHealthReport
	CheckTenantsConcurrently(ctx context.Context, tenantIDs []uuid.UUID) []HealthCheckResult
}

// Verify implementation
var _ TenantDBHealthCheckerInterface = (*TenantDBHealthChecker)(nil)