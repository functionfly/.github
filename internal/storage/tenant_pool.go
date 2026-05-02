package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// TenantPoolManager manages connection pools for all tenant databases
type TenantPoolManager struct {
	config     *TenantDatabaseConfig
	provisioner *TenantDBProvisioner
	pools      sync.Map // map[uuid.UUID]*ManagedPool
	stats      TenantPoolStats
	closed     atomic.Bool
	cleanupInterval time.Duration
	stopCleanup  chan struct{}
}

// ManagedPool wraps a pgxpool.Pool with metadata and health tracking
type ManagedPool struct {
	TenantID    uuid.UUID
	Pool        *pgxpool.Pool
	CreatedAt   time.Time
	LastUsedAt  time.Time
	HealthCheck time.Time
	Status      PoolStatus
	ErrorCount  int
}

// PoolStatus represents the health state of a pool
type PoolStatus string

const (
	PoolStatusHealthy   PoolStatus = "healthy"
	PoolStatusDegraded  PoolStatus = "degraded"
	PoolStatusUnhealthy PoolStatus = "unhealthy"
	PoolStatusClosed    PoolStatus = "closed"
)

// TenantPoolStats holds aggregated statistics for all tenant pools
type TenantPoolStats struct {
	TotalPools     int32
	HealthyPools   int32
	DegradedPools  int32
	UnhealthyPools int32
	TotalConns     int32
	IdleConns      int32
	AcquiredConns  int32
}

// Pool metrics
var (
	tenantPoolConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_tenant_pool_connections",
			Help: "Number of connections in tenant database pools",
		},
		[]string{"tenant_id", "pool_status"},
	)

	tenantPoolErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_tenant_pool_errors_total",
			Help: "Total number of tenant pool errors",
		},
		[]string{"tenant_id", "error_type"},
	)

	tenantPoolHealthDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_tenant_pool_health_check_duration_seconds",
			Help:    "Duration of tenant pool health checks",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tenant_id", "status"},
	)
)

// NewTenantPoolManager creates a new pool manager
func NewTenantPoolManager(config *TenantDatabaseConfig, provisioner *TenantDBProvisioner) *TenantPoolManager {
	m := &TenantPoolManager{
		config:          config,
		provisioner:    provisioner,
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	if config.Enabled {
		go m.cleanupLoop()
	}

	return m
}

// GetPool retrieves or creates a connection pool for a tenant
func (m *TenantPoolManager) GetPool(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Pool, error) {
	if m.closed.Load() {
		return nil, fmt.Errorf("pool manager is closed")
	}

	// Check if pool already exists
	if managed, ok := m.pools.Load(tenantID); ok {
		mp := managed.(*ManagedPool)
		if mp.Status != PoolStatusClosed {
			mp.LastUsedAt = time.Now()
			return mp.Pool, nil
		}
		// Pool was closed, recreate it
		m.pools.Delete(tenantID)
	}

	// Check if tenant DB config exists
	dbConfig, err := m.provisioner.GetTenantDBStatus(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant database not found: %w", err)
	}
	if dbConfig == "suspended" {
		return nil, fmt.Errorf("tenant database is suspended")
	}

	// Create new pool via provisioner
	pool, err := m.provisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant pool: %w", err)
	}

	// Wrap with managed pool
	mp := &ManagedPool{
		TenantID:   tenantID,
		Pool:       pool,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		HealthCheck: time.Now(),
		Status:     PoolStatusHealthy,
	}

	m.pools.Store(tenantID, mp)
	atomic.AddInt32(&m.stats.TotalPools, 1)
	atomic.AddInt32(&m.stats.HealthyPools, 1)

	// Start health monitoring for this pool
	go m.monitorPool(mp)

	return pool, nil
}

// GetPoolByDBName retrieves a pool by database name (for admin operations)
func (m *TenantPoolManager) GetPoolByDBName(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	// Look up tenant ID from database name
	tenantID, err := m.resolveTenantIDFromDBName(ctx, dbName)
	if err != nil {
		return nil, err
	}
	return m.GetPool(ctx, tenantID)
}

// resolveTenantIDFromDBName extracts tenant ID from database name
func (m *TenantPoolManager) resolveTenantIDFromDBName(ctx context.Context, dbName string) (uuid.UUID, error) {
	// Database name format: prefix + first 8 chars of tenant ID
	prefix := m.config.Prefix
	if len(dbName) <= len(prefix) {
		return uuid.Nil, fmt.Errorf("invalid database name format")
	}

	tenantIDStr := dbName[len(prefix):]
	return uuid.Parse(tenantIDStr)
}

// ClosePool closes and removes a specific tenant's pool
func (m *TenantPoolManager) ClosePool(tenantID uuid.UUID) error {
	if managed, ok := m.pools.LoadAndDelete(tenantID); ok {
		mp := managed.(*ManagedPool)
		mp.Close()

		// Update stats
		atomic.AddInt32(&m.stats.TotalPools, -1)
		if mp.Status == PoolStatusHealthy {
			atomic.AddInt32(&m.stats.HealthyPools, -1)
		}

		// Remove metrics
		tenantPoolConnections.DeleteLabelValues(tenantID.String(), string(mp.Status))
	}

	return nil
}

// Close closes the managed pool
func (mp *ManagedPool) Close() {
	if mp.Pool != nil {
		mp.Pool.Close()
		mp.Status = PoolStatusClosed
	}
}

// Close closes all pools and stops the manager
func (m *TenantPoolManager) Close() error {
	if m.closed.Swap(true) {
		return nil // Already closed
	}

	close(m.stopCleanup)

	m.pools.Range(func(key, value interface{}) bool {
		mp := value.(*ManagedPool)
		mp.Pool.Close()
		return true
	})

	m.pools = sync.Map{} // Clear all pools

	return nil
}

// GetStats returns current pool statistics
func (m *TenantPoolManager) GetStats() TenantPoolStats {
	return TenantPoolStats{
		TotalPools:     atomic.LoadInt32(&m.stats.TotalPools),
		HealthyPools:   atomic.LoadInt32(&m.stats.HealthyPools),
		DegradedPools:  atomic.LoadInt32(&m.stats.DegradedPools),
		UnhealthyPools: atomic.LoadInt32(&m.stats.UnhealthyPools),
		TotalConns:     atomic.LoadInt32(&m.stats.TotalConns),
		IdleConns:      atomic.LoadInt32(&m.stats.IdleConns),
		AcquiredConns:  atomic.LoadInt32(&m.stats.AcquiredConns),
	}
}

// ListPools returns information about all managed pools
func (m *TenantPoolManager) ListPools() []PoolInfo {
	var pools []PoolInfo

	m.pools.Range(func(key, value interface{}) bool {
		mp := value.(*ManagedPool)
		stats := mp.Pool.Stat()

		pools = append(pools, PoolInfo{
			TenantID:     mp.TenantID,
			Status:       string(mp.Status),
			CreatedAt:    mp.CreatedAt,
			LastUsedAt:   mp.LastUsedAt,
			TotalConns:   int32(stats.TotalConns()),
			IdleConns:    int32(stats.IdleConns()),
			AcquiredConns: int32(stats.AcquiredConns()),
			ErrorCount:   mp.ErrorCount,
		})
		return true
	})

	return pools
}

// PoolInfo holds information about a managed pool
type PoolInfo struct {
	TenantID      uuid.UUID
	Status        string
	CreatedAt     time.Time
	LastUsedAt    time.Time
	TotalConns    int32
	IdleConns     int32
	AcquiredConns int32
	ErrorCount    int
}

// cleanupLoop periodically closes idle pools
func (m *TenantPoolManager) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupIdlePools()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanupIdlePools closes pools that haven't been used for the configured idle timeout
func (m *TenantPoolManager) cleanupIdlePools() {
	idleTimeout := m.config.MaxIdleTime
	if idleTimeout == 0 {
		idleTimeout = 5 * time.Minute // Default
	}

	m.pools.Range(func(key, value interface{}) bool {
		mp := value.(*ManagedPool)

		// Skip closed pools
		if mp.Status == PoolStatusClosed {
			return true
		}

		// Check if pool is idle
		if time.Since(mp.LastUsedAt) > idleTimeout {
			logrus.Infof("Closing idle pool for tenant %s (idle for %v)", mp.TenantID, time.Since(mp.LastUsedAt))
			m.ClosePool(mp.TenantID)
		}

		return true
	})
}

// monitorPool performs continuous health monitoring for a pool
func (m *TenantPoolManager) monitorPool(mp *ManagedPool) {
	interval := 30 * time.Second // Check every 30 seconds

	for {
		select {
		case <-time.After(interval):
			if m.closed.Load() || mp.Status == PoolStatusClosed {
				return
			}

			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := mp.Pool.Ping(ctx)
			cancel()

			duration := time.Since(start).Seconds()
			tenantPoolHealthDuration.WithLabelValues(mp.TenantID.String(), string(mp.Status)).Observe(duration)

			mp.HealthCheck = time.Now()

			if err != nil {
				mp.ErrorCount++
				tenantPoolErrors.WithLabelValues(mp.TenantID.String(), "ping_failed").Inc()

				logrus.Warnf("Tenant %s pool health check failed: %v (errors: %d)", mp.TenantID, err, mp.ErrorCount)

				if mp.ErrorCount >= 3 {
					mp.Status = PoolStatusUnhealthy
					atomic.AddInt32(&m.stats.HealthyPools, -1)
					atomic.AddInt32(&m.stats.UnhealthyPools, 1)
				} else {
					mp.Status = PoolStatusDegraded
					atomic.AddInt32(&m.stats.HealthyPools, -1)
					atomic.AddInt32(&m.stats.DegradedPools, 1)
				}

				tenantPoolConnections.WithLabelValues(mp.TenantID.String(), string(mp.Status)).Set(float64(mp.Pool.Stat().TotalConns()))
			} else {
				if mp.ErrorCount >= 3 {
					// Recovering
					atomic.AddInt32(&m.stats.UnhealthyPools, -1)
					if mp.ErrorCount < 6 {
						atomic.AddInt32(&m.stats.DegradedPools, 1)
					} else {
						atomic.AddInt32(&m.stats.HealthyPools, 1)
					}
				}
				mp.ErrorCount = 0
				if mp.Status != PoolStatusHealthy {
					mp.Status = PoolStatusHealthy
					atomic.AddInt32(&m.stats.DegradedPools, -1)
					atomic.AddInt32(&m.stats.HealthyPools, 1)
				}
			}

			// Update connection metrics
			stats := mp.Pool.Stat()
			tenantPoolConnections.WithLabelValues(mp.TenantID.String(), string(mp.Status)).Set(float64(stats.TotalConns()))

		case <-m.stopCleanup:
			return
		}
	}
}

// GetPoolHealth returns the health status of a specific tenant's pool
func (m *TenantPoolManager) GetPoolHealth(tenantID uuid.UUID) (PoolStatus, error) {
	if managed, ok := m.pools.Load(tenantID); ok {
		mp := managed.(*ManagedPool)
		return mp.Status, nil
	}
	return PoolStatusClosed, fmt.Errorf("pool not found for tenant %s", tenantID)
}

// PrewarmPool pre-activates a pool for a tenant (for proactive scaling)
func (m *TenantPoolManager) PrewarmPool(ctx context.Context, tenantID uuid.UUID) error {
	_, err := m.GetPool(ctx, tenantID)
	return err
}

// TenantPoolManagerInterface defines the interface for tenant pool management
type TenantPoolManagerInterface interface {
	GetPool(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Pool, error)
	ClosePool(tenantID uuid.UUID) error
	Close() error
	GetStats() TenantPoolStats
	ListPools() []PoolInfo
	GetPoolHealth(tenantID uuid.UUID) (PoolStatus, error)
	PrewarmPool(ctx context.Context, tenantID uuid.UUID) error
}

// Verify TenantPoolManager implements TenantPoolManagerInterface
var _ TenantPoolManagerInterface = (*TenantPoolManager)(nil)