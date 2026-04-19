package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ReadReplicaConfig holds configuration for a read replica
type ReadReplicaConfig struct {
	Host     string
	Port     int
	Weight   int    // Load balancing weight (higher = more load)
	Priority int    // Priority level (lower = higher priority)
	Region   string // Geographic region for latency-based routing
}

// DatabaseConfig holds the database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int
	MaxIdle         int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Read replica configuration
	ReadReplicas       []ReadReplicaConfig
	ReadReplicaEnabled bool

	// Health monitoring
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	MaxHealthFailures   int

	// Connection recovery
	RetryAttempts         int
	RetryDelay            time.Duration
	CircuitBreakerTimeout time.Duration

	// ConnectionString is set when DATABASE_URL is used (e.g. Neon); takes precedence over Host/Port/User/Password/Database/SSLMode
	ConnectionString string
}

// loadDatabaseConfig loads database configuration from environment variables.
// If DATABASE_URL is set (e.g. for Neon), it is used as the connection string and Host/Port/User/Database/SSLMode are parsed for logging.
func loadDatabaseConfig() (*DatabaseConfig, error) {
	port, err := strconv.Atoi(getEnvOrDefault("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	// Dynamic connection pool sizing based on CPU cores and load
	maxConns, maxIdle := calculateConnectionPoolSize()

	connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 10*time.Minute)
	connMaxIdleTime := getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	// Load read replica configuration
	readReplicas := loadReadReplicaConfig()
	readReplicaEnabled := getEnvOrDefault("DB_READ_REPLICA_ENABLED", "false") == "true"

	// Load health monitoring configuration
	healthCheckInterval := getEnvDuration("DB_HEALTH_CHECK_INTERVAL", 30*time.Second)
	healthCheckTimeout := getEnvDuration("DB_HEALTH_CHECK_TIMEOUT", 5*time.Second)
	maxHealthFailures, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_HEALTH_FAILURES", "3"))

	// Load connection recovery configuration
	retryAttempts, _ := strconv.Atoi(getEnvOrDefault("DB_RETRY_ATTEMPTS", "3"))
	retryDelay := getEnvDuration("DB_RETRY_DELAY", 1*time.Second)
	circuitBreakerTimeout := getEnvDuration("DB_CIRCUIT_BREAKER_TIMEOUT", 60*time.Second)

	// Determine if we're in development mode (allows insecure defaults)
	isDevelopment := os.Getenv("DEVELOPMENT") == "true"

	// In production, require explicit database configuration - no insecure defaults
	var dbHost, dbUser, dbPassword, dbName, dbSSLMode string
	if isDevelopment {
		// Development: use safe defaults
		dbHost = getEnvOrDefault("DB_HOST", "localhost")
		dbUser = getEnvOrDefault("DB_USER", "postgres")
		dbPassword = getEnvOrDefault("DB_PASSWORD", "postgres")
		dbName = getEnvOrDefault("DB_NAME", "functionfly")
		dbSSLMode = getEnvOrDefault("DB_SSLMODE", "require")
	} else {
		// Production: if DATABASE_URL is set, individual DB_* vars are optional
		if os.Getenv("DATABASE_URL") == "" {
			// No DATABASE_URL — require individual DB_* vars
			dbHost = os.Getenv("DB_HOST")
			dbUser = os.Getenv("DB_USER")
			dbPassword = os.Getenv("DB_PASSWORD")
			dbName = os.Getenv("DB_NAME")
			dbSSLMode = os.Getenv("DB_SSLMODE")
			if dbSSLMode == "" {
				dbSSLMode = "require" // Default to require for security
			}

			missingVars := []string{}
			if dbHost == "" {
				missingVars = append(missingVars, "DB_HOST")
			}
			if dbUser == "" {
				missingVars = append(missingVars, "DB_USER")
			}
			if dbPassword == "" {
				missingVars = append(missingVars, "DB_PASSWORD")
			}
			if dbName == "" {
				missingVars = append(missingVars, "DB_NAME")
			}
			if len(missingVars) > 0 {
				return nil, fmt.Errorf("production database configuration missing required environment variables: %s", strings.Join(missingVars, ", "))
			}
		}
		// Otherwise DATABASE_URL is set — individual vars are not required
	}

	cfg := &DatabaseConfig{
		Host:            dbHost,
		Port:            port,
		User:            dbUser,
		Password:        dbPassword,
		Database:        dbName,
		SSLMode:         dbSSLMode,
		MaxConns:        maxConns,
		MaxIdle:         maxIdle,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,

		ReadReplicas:       readReplicas,
		ReadReplicaEnabled: readReplicaEnabled,

		HealthCheckInterval: healthCheckInterval,
		HealthCheckTimeout:  healthCheckTimeout,
		MaxHealthFailures:   maxHealthFailures,

		RetryAttempts:         retryAttempts,
		RetryDelay:            retryDelay,
		CircuitBreakerTimeout: circuitBreakerTimeout,
	}

	// DATABASE_URL takes precedence in production (e.g. Neon connection string from Console)
	// In development, if DB_HOST is explicitly set, prefer individual vars over DATABASE_URL
	raw := os.Getenv("DATABASE_URL")
	if raw != "" {
		if isDevelopment && os.Getenv("DB_HOST") != "" {
			// Development mode with explicit DB_HOST: ignore DATABASE_URL and use local Postgres
			logrus.Info("DEVELOPMENT=true with DB_HOST set: ignoring DATABASE_URL, using local Postgres")
		} else {
			cfg.ConnectionString = raw
			if u, err := parsePostgresURL(raw); err == nil {
				cfg.Host = u.Host
				cfg.Port = u.Port
				cfg.User = u.User
				cfg.Database = u.Database
				if u.SSLMode != "" {
					cfg.SSLMode = u.SSLMode
				}
			}
		}
	}

	return cfg, nil
}

// parsedPostgresURL holds components parsed from a postgresql:// or postgres:// URL for logging
type parsedPostgresURL struct {
	Host     string
	Port     int
	User     string
	Database string
	SSLMode  string
}

func parsePostgresURL(s string) (*parsedPostgresURL, error) {
	// lib/pq accepts postgres:// and postgresql://
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "postgres://") && !strings.HasPrefix(s, "postgresql://") {
		return nil, fmt.Errorf("invalid scheme")
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	port := 5432
	if p := u.Port(); p != "" {
		if pi, err := strconv.Atoi(p); err == nil {
			port = pi
		}
	}
	db := strings.TrimPrefix(u.Path, "/")
	if db == "" {
		db = "functionfly"
	}
	sslmode := "require"
	if q := u.Query().Get("sslmode"); q != "" {
		sslmode = q
	}
	return &parsedPostgresURL{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Database: db,
		SSLMode:  sslmode,
	}, nil
}

// buildConnectionString creates a PostgreSQL connection string from config.
// If config.ConnectionString is set (e.g. from DATABASE_URL for Neon), it is used and
// prefer_simple_protocol=1 is appended so lib/pq does not use prepared statements,
// which break with PgBouncer/Neon pooled connections (e.g. "unnamed prepared statement does not exist").
func buildConnectionString(config *DatabaseConfig) string {
	if config.ConnectionString != "" {
		s := config.ConnectionString
		if !strings.Contains(s, "prefer_simple_protocol=") {
			if strings.Contains(s, "?") {
				s += "&prefer_simple_protocol=1"
			} else {
				s += "?prefer_simple_protocol=1"
			}
		}
		return s
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
}

// configureConnectionPool sets up the database connection pool parameters with optimized settings
func configureConnectionPool(db *sql.DB, config *DatabaseConfig) error {
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Additional optimizations for better performance
	// Set maximum time spent waiting for a connection (default is no timeout)
	// This prevents indefinite blocking when all connections are in use
	// db.SetConnMaxWaitTime(30 * time.Second) // Uncomment if needed

	return nil
}

// GetConnectionString returns a PostgreSQL connection string from environment variables.
// This is a package-level helper that wraps loadDatabaseConfig + buildConnectionString
// so callers (e.g. the notification WebSocket LISTEN pool) can obtain a connection string
// without duplicating the env-var logic.
func GetConnectionString() string {
	cfg, err := loadDatabaseConfig()
	if err != nil {
		logrus.WithError(err).Error("GetConnectionString: failed to load config, returning empty string")
		return ""
	}
	return buildConnectionString(cfg)
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loadReadReplicaConfig loads read replica configuration from environment variables
func loadReadReplicaConfig() []ReadReplicaConfig {
	var replicas []ReadReplicaConfig

	// Support up to 5 read replicas for now
	for i := 1; i <= 5; i++ {
		hostKey := fmt.Sprintf("DB_REPLICA_%d_HOST", i)
		portKey := fmt.Sprintf("DB_REPLICA_%d_PORT", i)
		weightKey := fmt.Sprintf("DB_REPLICA_%d_WEIGHT", i)
		priorityKey := fmt.Sprintf("DB_REPLICA_%d_PRIORITY", i)
		regionKey := fmt.Sprintf("DB_REPLICA_%d_REGION", i)

		host := os.Getenv(hostKey)
		if host == "" {
			continue // No replica configured at this index
		}

		port, err := strconv.Atoi(getEnvOrDefault(portKey, "5432"))
		if err != nil {
			port = 5432
		}

		weight, err := strconv.Atoi(getEnvOrDefault(weightKey, "1"))
		if err != nil {
			weight = 1
		}

		priority, err := strconv.Atoi(getEnvOrDefault(priorityKey, "1"))
		if err != nil {
			priority = 1
		}

		region := getEnvOrDefault(regionKey, "default")

		replicas = append(replicas, ReadReplicaConfig{
			Host:     host,
			Port:     port,
			Weight:   weight,
			Priority: priority,
			Region:   region,
		})
	}

	return replicas
}

// getEnvDuration parses a duration from environment variable or returns default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// calculateConnectionPoolSize dynamically calculates optimal connection pool size based on system resources and load
func calculateConnectionPoolSize() (maxOpen, maxIdle int) {
	// Get CPU cores for baseline calculation
	cpuCores := runtime.NumCPU()

	// Base calculations
	// Max open connections: 4x CPU cores for moderate load, can be overridden by env var
	baseMaxOpen := cpuCores * 4
	// Max idle connections: 25% of max open connections
	baseMaxIdle := baseMaxOpen / 4

	// Check for explicit environment overrides
	if maxOpenStr := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpenStr != "" {
		if val, err := strconv.Atoi(maxOpenStr); err == nil && val > 0 {
			maxOpen = val
		} else {
			maxOpen = baseMaxOpen
		}
	} else {
		maxOpen = baseMaxOpen
	}

	if maxIdleStr := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdleStr != "" {
		if val, err := strconv.Atoi(maxIdleStr); err == nil && val > 0 {
			maxIdle = val
		} else {
			maxIdle = baseMaxIdle
		}
	} else {
		maxIdle = baseMaxIdle
	}

	// Apply reasonable bounds
	if maxOpen < 1 {
		maxOpen = 1
	}
	if maxOpen > 200 { // Prevent excessive connections
		maxOpen = 200
	}
	if maxIdle < 1 {
		maxIdle = 1
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	return maxOpen, maxIdle
}
