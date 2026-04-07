package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// ReadReplicaConnection represents a read replica connection
type ReadReplicaConnection struct {
	*sql.DB
	Host            string
	Port            int
	Weight          int
	Priority        int
	Region          string
	Healthy         bool
	Failures        int
	LastHealthCheck time.Time
}

// PostgresDB wraps sql.DB and provides repository pattern implementation with read replica support
type PostgresDB struct {
	skipPreparedStatements   bool     // For migrations, skip prepared statements
	*sql.DB                           // Primary write connection
	GORM                     *gorm.DB // GORM instance for ORM operations
	userRepository           *UserRepository
	tenantRepository         *TenantRepository
	billingRepository        *BillingRepository
	revenueRepository        *RevenueRepository
	auditRepository          *AuditRepository
	appRepository            *AppRepository
	backendRepository        *BackendRepository
	deploymentRepository     *DeploymentRepository
	contentRepository        *ContentRepository
	feedbackRepository       *FeedbackRepository
	monitoringRepository     *MonitoringRepository
	sessionRepository        *SessionRepository
	refreshTokenRepository   *RefreshTokenRepository
	loginAttemptRepository   *LoginAttemptRepository
	authEventRepository      *AuthEventRepository
	localRuntimeRepository   *LocalRuntimeRepository
	functionRepository       *FunctionRepository
	registryRepository       *registry.RegistryRepository
	incidentRepository       *IncidentRepository
	featureMeasureRepository *FeatureMeasureRepository
	teamRepository           *TeamRepository
	followRepository         *FollowRepository
	adminSessionRepository   *AdminSessionRepository
	encryptionManager        *DatabaseEncryptionManager

	// Read replica connections
	readReplicas       []ReadReplicaConnection
	readReplicaEnabled bool

	// Connection health monitoring
	config            *DatabaseConfig
	healthCheckTicker *time.Ticker
	healthCheckDone   chan bool

	// Prepared statements for common queries
	stmtGetUserByEmail       *sql.Stmt
	stmtGetUserByID          *sql.Stmt
	stmtGetTenantByID        *sql.Stmt
	stmtGetAppByID           *sql.Stmt
	stmtGetAppBySlug         *sql.Stmt
	stmtListAppsByTenant     *sql.Stmt
	stmtGetBackendByID       *sql.Stmt
	stmtListBackendsByApp    *sql.Stmt
	stmtGetDeploymentByID    *sql.Stmt
	stmtListDeploymentsByApp *sql.Stmt

	// Enhanced prepared statement management
	preparedStatements  map[string]*sql.Stmt
	stmtMutex           sync.RWMutex
	stmtReprepareTicker *time.Ticker
	lastStmtReprepare   time.Time

	// Transaction management
	transactionManager *TransactionManager

	// Query performance monitoring
	queryMonitor *QueryMonitor
}

// NewPostgresDB creates a new PostgreSQL database connection with all repositories initialized
func NewPostgresDB() (*PostgresDB, error) {
	return NewPostgresDBWithOptions(false)
}

// openMigrateSQLDB opens a dedicated *sql.DB for golang-migrate. Pooled Neon/PgBouncer
// endpoints need the simple query protocol; using lib/pq here still fails
// pg_advisory_lock with "unnamed prepared statement does not exist".
func openMigrateSQLDB(config *DatabaseConfig) (*sql.DB, error) {
	if config.ConnectionString != "" {
		pgxConfig, err := pgx.ParseConfig(config.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL for migration connection: %w", err)
		}
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		dsn := stdlib.RegisterConnConfig(pgxConfig)
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open migration database: %w", err)
		}
		if err := configureConnectionPool(db, config); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to configure migration connection pool: %w", err)
		}
		return db, nil
	}
	db, err := sql.Open("postgres", buildConnectionString(config))
	if err != nil {
		return nil, fmt.Errorf("failed to open migration database: %w", err)
	}
	if err := configureConnectionPool(db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure migration connection pool: %w", err)
	}
	return db, nil
}

func NewPostgresDBWithOptions(skipPreparedStatements bool) (*PostgresDB, error) {
	config, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	connStr := buildConnectionString(config)

	var db *sql.DB
	var gormDSN string
	if config.ConnectionString != "" {
		// Use pgx with simple protocol for Neon/PgBouncer to avoid "unnamed prepared statement does not exist".
		// lib/pq does not support disabling server-side prepared statements for parameterized queries.
		pgxConfig, err := pgx.ParseConfig(config.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL for pgx: %w", err)
		}
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		gormDSN = stdlib.RegisterConnConfig(pgxConfig)
		db, err = sql.Open("pgx", gormDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}
	} else {
		gormDSN = connStr
		var err error
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}
	}

	// Configure connection pool
	if err := configureConnectionPool(db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// Test connection with context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
		"sslmode":  config.SSLMode,
	}).Info("Connected to PostgreSQL database")

	// Initialize GORM. When using pgx (DATABASE_URL), same driver and simple protocol. Otherwise lib/pq with PrepareStmt: false for pooler safety.
	gormCfg := postgres.Config{DSN: gormDSN}
	if config.ConnectionString != "" {
		gormCfg.DriverName = "pgx"
	}
	gormDB, err := gorm.Open(postgres.New(gormCfg), &gorm.Config{
		// Ignore ErrRecordNotFound so expected misses (e.g. no platform_maintenance row) do not spam logs.
		Logger: gormlog.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlog.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlog.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		PrepareStmt: false,
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// Configure GORM connection pool to match sql.DB settings
	sqlDB, err := gormDB.DB()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to get underlying sql.DB from GORM: %w", err)
	}

	// Apply same connection pool settings as the raw sql.DB
	if err := configureConnectionPool(sqlDB, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure GORM connection pool: %w", err)
	}

	postgresDB := &PostgresDB{
		DB:                     db,
		GORM:                   gormDB,
		skipPreparedStatements: skipPreparedStatements,
		config:                 config,
		readReplicaEnabled:     config.ReadReplicaEnabled,
		preparedStatements:     make(map[string]*sql.Stmt),
	}

	// Initialize transaction manager
	postgresDB.transactionManager = NewTransactionManager(postgresDB)

	// Initialize query monitor
	postgresDB.queryMonitor = NewQueryMonitor(postgresDB)
	// Enable slow query logging with 100ms threshold
	postgresDB.queryMonitor.EnableSlowQueryLogging(100 * time.Millisecond)

	// Initialize read replicas if enabled
	if config.ReadReplicaEnabled && len(config.ReadReplicas) > 0 {
		if err := postgresDB.initReadReplicas(config); err != nil {
			logrus.WithError(err).Warn("Failed to initialize read replicas, continuing with primary only")
		}
	}

	// Initialize repositories (before health monitoring so we don't compete for connections)
	postgresDB.userRepository = NewUserRepository(postgresDB)
	postgresDB.tenantRepository = NewTenantRepository(postgresDB)
	postgresDB.billingRepository = NewBillingRepository(postgresDB)
	postgresDB.revenueRepository = NewRevenueRepository(postgresDB.DB)
	postgresDB.auditRepository = NewAuditRepository(postgresDB)
	postgresDB.appRepository = NewAppRepository(postgresDB)
	postgresDB.backendRepository = NewBackendRepository(postgresDB)
	postgresDB.deploymentRepository = NewDeploymentRepository(postgresDB)
	postgresDB.contentRepository = NewContentRepository(postgresDB)
	postgresDB.feedbackRepository = NewFeedbackRepository(postgresDB)
	postgresDB.monitoringRepository = NewMonitoringRepository(postgresDB)
	postgresDB.sessionRepository = NewSessionRepository(postgresDB)
	postgresDB.refreshTokenRepository = NewRefreshTokenRepository(postgresDB)
	postgresDB.loginAttemptRepository = NewLoginAttemptRepository(postgresDB)
	postgresDB.authEventRepository = NewAuthEventRepository(postgresDB)
	postgresDB.localRuntimeRepository = NewLocalRuntimeRepository(postgresDB)
	postgresDB.functionRepository = NewFunctionRepository(postgresDB.DB)
	postgresDB.registryRepository = registry.NewRegistryRepository(postgresDB.GORM, nil)
	postgresDB.incidentRepository = NewIncidentRepository(postgresDB.DB)
	postgresDB.featureMeasureRepository = NewFeatureMeasureRepository(postgresDB.DB)
	postgresDB.teamRepository = NewTeamRepository(postgresDB.GORM)
	postgresDB.followRepository = NewFollowRepository(postgresDB)
	postgresDB.adminSessionRepository = NewAdminSessionRepository(postgresDB)

	// Initialize encryption manager
	encryptionManager, err := NewDatabaseEncryptionManager(postgresDB)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize encryption manager: %w", err)
	}
	postgresDB.encryptionManager = encryptionManager

	// Initialize prepared statements (skip for migrations). Use a fresh context with
	// sufficient timeout so encryption/repos above don't exhaust the initial ping ctx.
	if !skipPreparedStatements {
		stmtCtx, stmtCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer stmtCancel()
		if err := postgresDB.initPreparedStatements(stmtCtx); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to initialize prepared statements: %w", err)
		}
	}

	// Start health monitoring after prepared statements so it doesn't hold connections during init
	if config.HealthCheckInterval > 0 {
		postgresDB.startHealthMonitoring()
	}

	return postgresDB, nil
}

// Repository interface implementation
func (db *PostgresDB) Repository() Repository {
	return db
}

// Encryption methods
func (db *PostgresDB) EncryptionManager() *DatabaseEncryptionManager {
	return db.encryptionManager
}

// Registry repository accessor
func (db *PostgresDB) RegistryRepository() *registry.RegistryRepository {
	return db.registryRepository
}

// LoginAttemptRepository accessor
func (db *PostgresDB) LoginAttemptRepository() *LoginAttemptRepository {
	return db.loginAttemptRepository
}

func (db *PostgresDB) IsEncryptionEnabled() bool {
	return db.encryptionManager.IsEncryptionEnabled()
}

func (db *PostgresDB) EncryptField(value string) (string, error) {
	return db.encryptionManager.EncryptField(value)
}

func (db *PostgresDB) DecryptField(value string) (string, error) {
	return db.encryptionManager.DecryptField(value)
}

// AdminSessionRepository accessors
func (db *PostgresDB) CreateAdminSession(userID uuid.UUID, token string, ipAddress, userAgent, deviceFingerprint string, expiresAt time.Time) (*AdminSessionModel, error) {
	return db.adminSessionRepository.CreateAdminSession(userID, token, ipAddress, userAgent, deviceFingerprint, expiresAt)
}

func (db *PostgresDB) GetAdminSessionByToken(token string) (*AdminSessionModel, error) {
	return db.adminSessionRepository.GetAdminSessionByToken(token)
}

func (db *PostgresDB) UpdateAdminSessionLastActivity(sessionID uuid.UUID) error {
	return db.adminSessionRepository.UpdateAdminSessionLastActivity(sessionID)
}

func (db *PostgresDB) RevokeAdminSession(sessionID uuid.UUID) error {
	return db.adminSessionRepository.RevokeAdminSession(sessionID)
}

func (db *PostgresDB) RevokeAllAdminUserSessions(userID uuid.UUID) error {
	return db.adminSessionRepository.RevokeAllAdminUserSessions(userID)
}

func (db *PostgresDB) ListAdminUserSessions(userID uuid.UUID) ([]*AdminSessionModel, error) {
	return db.adminSessionRepository.ListAdminUserSessions(userID)
}

func (db *PostgresDB) DeleteExpiredAdminSessions() (int64, error) {
	return db.adminSessionRepository.DeleteExpiredAdminSessions()
}
