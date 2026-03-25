package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ConnectionPoolConfig holds connection pool configuration
type ConnectionPoolConfig struct {
	// MaxConns is the maximum number of connections
	MaxConns int32
	// MinConns is the minimum number of connections
	MinConns int32
	// MaxConnLifetime is the maximum lifetime of a connection
	MaxConnLifetime time.Duration
	// MaxConnIdleTime is the maximum idle time of a connection
	MaxConnIdleTime time.Duration
	// HealthCheckPeriod is how often to check connection health
	HealthCheckPeriod time.Duration
}

// DefaultConnectionPoolConfig returns production-ready defaults
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxConns:          20,
		MinConns:          5,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

// ConnectionPool manages database connection pooling
type ConnectionPool struct {
	pool   *pgxpool.Pool
	config *ConnectionPoolConfig
	mu     sync.RWMutex
	stats  ConnectionPoolStats
}

// ConnectionPoolStats tracks connection pool statistics
type ConnectionPoolStats struct {
	TotalConns      int32
	IdleConns       int32
	ActiveConns     int32
	MaxConns        int32
	AcquireCount    int64
	AcquireDuration time.Duration
	mu              sync.RWMutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(ctx context.Context, connString string, config *ConnectionPoolConfig) (*ConnectionPool, error) {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"max_conns": config.MaxConns,
		"min_conns": config.MinConns,
	}).Info("Connection pool created")

	return &ConnectionPool{
		pool:   pool,
		config: config,
	}, nil
}

// GetPool returns the underlying connection pool
func (cp *ConnectionPool) GetPool() *pgxpool.Pool {
	return cp.pool
}

// Acquire acquires a connection from the pool
func (cp *ConnectionPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	start := time.Now()
	conn, err := cp.pool.Acquire(ctx)
	duration := time.Since(start)

	cp.mu.Lock()
	cp.stats.AcquireCount++
	cp.stats.AcquireDuration += duration
	cp.mu.Unlock()

	if err != nil {
		logrus.WithError(err).Error("Failed to acquire connection")
		return nil, err
	}

	return conn, nil
}

// Release releases a connection back to the pool
func (cp *ConnectionPool) Release(conn *pgxpool.Conn) {
	conn.Release()
}

// Stats returns connection pool statistics
func (cp *ConnectionPool) Stats() ConnectionPoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	poolStats := cp.pool.Stat()

	return ConnectionPoolStats{
		TotalConns:      poolStats.TotalConns(),
		IdleConns:       poolStats.IdleConns(),
		ActiveConns:     poolStats.TotalConns() - poolStats.IdleConns(),
		MaxConns:        cp.config.MaxConns,
		AcquireCount:    cp.stats.AcquireCount,
		AcquireDuration: cp.stats.AcquireDuration,
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() {
	cp.pool.Close()
	logrus.Info("Connection pool closed")
}

// HealthCheck checks the health of the connection pool
func (cp *ConnectionPool) HealthCheck(ctx context.Context) error {
	if err := cp.pool.Ping(ctx); err != nil {
		return fmt.Errorf("connection pool health check failed: %w", err)
	}

	stats := cp.Stats()
	if stats.ActiveConns > int32(float64(stats.MaxConns)*0.8) {
		logrus.WithFields(logrus.Fields{
			"active_conns": stats.ActiveConns,
			"max_conns":    stats.MaxConns,
		}).Warn("Connection pool utilization high")
	}

	return nil
}
