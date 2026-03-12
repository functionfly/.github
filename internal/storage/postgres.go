package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	skipPreparedStatements bool     // For migrations, skip prepared statements
	*sql.DB                         // Primary write connection
	GORM                   *gorm.DB // GORM instance for ORM operations
	userRepository         *UserRepository
	tenantRepository       *TenantRepository
	billingRepository      *BillingRepository
	auditRepository        *AuditRepository
	appRepository          *AppRepository
	backendRepository      *BackendRepository
	deploymentRepository   *DeploymentRepository
	contentRepository      *ContentRepository
	feedbackRepository     *FeedbackRepository
	monitoringRepository   *MonitoringRepository
	sessionRepository       *SessionRepository
	refreshTokenRepository  *RefreshTokenRepository
	loginAttemptRepository  *LoginAttemptRepository
	authEventRepository     *AuthEventRepository
	localRuntimeRepository  *LocalRuntimeRepository
	functionRepository     *FunctionRepository
	registryRepository     *registry.RegistryRepository
	incidentRepository       *IncidentRepository
	featureMeasureRepository *FeatureMeasureRepository
	teamRepository           *TeamRepository
	followRepository         *FollowRepository
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

func NewPostgresDBWithOptions(skipPreparedStatements bool) (*PostgresDB, error) {
	config, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	connStr := buildConnectionString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
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

	// Initialize GORM
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
	// Initialize repositories
	postgresDB.userRepository = NewUserRepository(postgresDB)
	postgresDB.tenantRepository = NewTenantRepository(postgresDB)
	postgresDB.billingRepository = NewBillingRepository(postgresDB)
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

func (db *PostgresDB) IsEncryptionEnabled() bool {
	return db.encryptionManager.IsEncryptionEnabled()
}

func (db *PostgresDB) EncryptField(value string) (string, error) {
	return db.encryptionManager.EncryptField(value)
}

func (db *PostgresDB) DecryptField(value string) (string, error) {
	return db.encryptionManager.DecryptField(value)
}

// InitPreparedStatements initializes prepared statements. Call after migrations when the DB
// was created with skipPreparedStatements=true. No-op if already initialized.
func (db *PostgresDB) InitPreparedStatements(ctx context.Context) error {
	db.stmtMutex.Lock()
	already := len(db.preparedStatements) > 0
	db.stmtMutex.Unlock()
	if already {
		return nil
	}
	return db.initPreparedStatements(ctx)
}

// initPreparedStatements initializes commonly used prepared statements with enhanced management
func (db *PostgresDB) initPreparedStatements(ctx context.Context) error {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()

	// Define prepared statements with their queries
	statements := map[string]string{
		"getUserByEmail": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio
			FROM users WHERE email = $1`,
		"getUserByID": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio
			FROM users WHERE id = $1`,
		"getTenantByID": `
			SELECT id, name, plan, status, created_at, updated_at
			FROM tenants WHERE id = $1`,
		"getAppByID": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE id = $1`,
		"getAppBySlug": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE slug = $1`,
		"listAppsByTenant": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE tenant_id = $1 ORDER BY created_at DESC`,
		"getBackendByID": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE id = $1`,
		"listBackendsByApp": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE app_id = $1 ORDER BY priority ASC NULLS LAST, created_at DESC`,
		"getDeploymentByID": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE id = $1`,
		"listDeploymentsByApp": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2`,
		// Registry function queries
		"getRegistryFunctionByID": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE id = $1`,
		"getRegistryFunctionByAuthorName": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE author = $1 AND name = $2`,
		"searchRegistryFunctions": `
			SELECT id, author, name, description, popularity_score, COALESCE(reliability_score, 0) AS trust_score, created_at
			FROM registry_functions
			WHERE visibility = 'public' AND (name ILIKE $1 OR description ILIKE $1)
			ORDER BY COALESCE(reliability_score, 0) DESC, popularity_score DESC
			LIMIT $2 OFFSET $3`,
	}

	// Prepare all statements
	for name, query := range statements {
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			db.closePreparedStatementsLocked() // Clean up on error; already hold stmtMutex
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
		db.preparedStatements[name] = stmt

		// Set legacy statement pointers for backward compatibility
		switch name {
		case "getUserByEmail":
			db.stmtGetUserByEmail = stmt
		case "getUserByID":
			db.stmtGetUserByID = stmt
		case "getTenantByID":
			db.stmtGetTenantByID = stmt
		case "getAppByID":
			db.stmtGetAppByID = stmt
		case "getAppBySlug":
			db.stmtGetAppBySlug = stmt
		case "listAppsByTenant":
			db.stmtListAppsByTenant = stmt
		case "getBackendByID":
			db.stmtGetBackendByID = stmt
		case "listBackendsByApp":
			db.stmtListBackendsByApp = stmt
		case "getDeploymentByID":
			db.stmtGetDeploymentByID = stmt
		case "listDeploymentsByApp":
			db.stmtListDeploymentsByApp = stmt
		}
	}

	// Start prepared statement re-preparation ticker
	db.stmtReprepareTicker = time.NewTicker(1 * time.Hour) // Reprepare every hour
	db.lastStmtReprepare = time.Now()

	go db.reprepareStatementsRoutine()

	logrus.Info("Initialized enhanced prepared statement management")
	return nil
}

// reprepareStatementsRoutine periodically re-prepares statements to prevent staleness
func (db *PostgresDB) reprepareStatementsRoutine() {
	for range db.stmtReprepareTicker.C {
		if err := db.reprepareStatements(); err != nil {
			logrus.WithError(err).Warn("Failed to re-prepare statements")
		} else {
			logrus.Debug("Successfully re-prepared database statements")
		}
	}
}

// reprepareStatements re-prepares all cached statements
func (db *PostgresDB) reprepareStatements() error {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Re-prepare each statement
	for name, oldStmt := range db.preparedStatements {
		newStmt, err := db.PrepareContext(ctx, db.getStatementQuery(name))
		if err != nil {
			return fmt.Errorf("failed to re-prepare statement %s: %w", name, err)
		}

		// Close old statement and replace
		oldStmt.Close()
		db.preparedStatements[name] = newStmt

		// Update legacy pointers
		switch name {
		case "getUserByEmail":
			db.stmtGetUserByEmail = newStmt
		case "getUserByID":
			db.stmtGetUserByID = newStmt
		case "getTenantByID":
			db.stmtGetTenantByID = newStmt
		case "getAppByID":
			db.stmtGetAppByID = newStmt
		case "getAppBySlug":
			db.stmtGetAppBySlug = newStmt
		case "listAppsByTenant":
			db.stmtListAppsByTenant = newStmt
		case "getBackendByID":
			db.stmtGetBackendByID = newStmt
		case "listBackendsByApp":
			db.stmtListBackendsByApp = newStmt
		case "getDeploymentByID":
			db.stmtGetDeploymentByID = newStmt
		case "listDeploymentsByApp":
			db.stmtListDeploymentsByApp = newStmt
		}
	}

	db.lastStmtReprepare = time.Now()
	return nil
}

// GetStatementQuery returns the SQL query for a given statement name (for fallback when prepared stmt is nil)
func (db *PostgresDB) GetStatementQuery(name string) string {
	return db.getStatementQuery(name)
}

// getStatementQuery returns the SQL query for a given statement name
func (db *PostgresDB) getStatementQuery(name string) string {
	queries := map[string]string{
		"getUserByEmail": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio
			FROM users WHERE email = $1`,
		"getUserByID": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio
			FROM users WHERE id = $1`,
		"getTenantByID": `
			SELECT id, name, plan, status, created_at, updated_at
			FROM tenants WHERE id = $1`,
		"getAppByID": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE id = $1`,
		"getAppBySlug": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE slug = $1`,
		"listAppsByTenant": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE tenant_id = $1 ORDER BY created_at DESC`,
		"getBackendByID": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE id = $1`,
		"listBackendsByApp": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE app_id = $1 ORDER BY priority ASC NULLS LAST, created_at DESC`,
		"getDeploymentByID": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE id = $1`,
		"listDeploymentsByApp": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2`,
		"getRegistryFunctionByID": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE id = $1`,
		"getRegistryFunctionByAuthorName": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE author = $1 AND name = $2`,
		"searchRegistryFunctions": `
			SELECT id, author, name, description, popularity_score, COALESCE(reliability_score, 0) AS trust_score, created_at
			FROM registry_functions
			WHERE visibility = 'public' AND (name ILIKE $1 OR description ILIKE $1)
			ORDER BY COALESCE(reliability_score, 0) DESC, popularity_score DESC
			LIMIT $2 OFFSET $3`,
	}

	return queries[name]
}

// GetPreparedStatement returns a prepared statement by name (thread-safe)
func (db *PostgresDB) GetPreparedStatement(name string) (*sql.Stmt, error) {
	db.stmtMutex.RLock()
	stmt, exists := db.preparedStatements[name]
	db.stmtMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("prepared statement %s not found", name)
	}

	return stmt, nil
}

// closePreparedStatements closes all prepared statements (takes stmtMutex).
func (db *PostgresDB) closePreparedStatements() {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()
	db.closePreparedStatementsLocked()
}

// closePreparedStatementsLocked does the actual close; caller must hold db.stmtMutex.
// Used from initPreparedStatements on error to avoid deadlock (init already holds the lock).
func (db *PostgresDB) closePreparedStatementsLocked() {
	// Stop re-preparation ticker
	if db.stmtReprepareTicker != nil {
		db.stmtReprepareTicker.Stop()
	}

	// Close all prepared statements
	for name, stmt := range db.preparedStatements {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				logrus.WithError(err).Warnf("Failed to close prepared statement: %s", name)
			}
		}
	}

	// Clear the map
	db.preparedStatements = make(map[string]*sql.Stmt)

	// Clear legacy pointers
	db.stmtGetUserByEmail = nil
	db.stmtGetUserByID = nil
	db.stmtGetTenantByID = nil
	db.stmtGetAppByID = nil
	db.stmtGetAppBySlug = nil
	db.stmtListAppsByTenant = nil
	db.stmtGetBackendByID = nil
	db.stmtListBackendsByApp = nil
	db.stmtGetDeploymentByID = nil
	db.stmtListDeploymentsByApp = nil
}

// Incident operations
func (db *PostgresDB) CreateIncident(ctx context.Context, incident *Incident) (*Incident, error) {
	return db.incidentRepository.CreateIncident(ctx, incident)
}

func (db *PostgresDB) GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	return db.incidentRepository.GetIncidentByID(ctx, incidentID)
}

func (db *PostgresDB) ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*Incident, error) {
	return db.incidentRepository.ListIncidents(ctx, limit, offset, status)
}

func (db *PostgresDB) ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*Incident, error) {
	return db.incidentRepository.ListIncidentsSince(ctx, since, limit)
}

func (db *PostgresDB) CountIncidentsSince(ctx context.Context, since time.Time) (int, error) {
	return db.incidentRepository.CountIncidentsSince(ctx, since)
}

func (db *PostgresDB) CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]DailyIncidentCount, error) {
	return db.incidentRepository.CountIncidentsGroupedByDay(ctx, since)
}

func (db *PostgresDB) UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error) {
	return db.incidentRepository.UpdateIncident(ctx, incidentID, updates)
}

func (db *PostgresDB) ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error) {
	return db.incidentRepository.ResolveIncident(ctx, incidentID)
}

// initReadReplicas initializes connections to read replica databases
func (db *PostgresDB) initReadReplicas(config *DatabaseConfig) error {
	for _, replicaConfig := range config.ReadReplicas {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			replicaConfig.Host, replicaConfig.Port, config.User, config.Password, config.Database, config.SSLMode)

		replicaDB, err := sql.Open("postgres", connStr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Failed to connect to read replica")
			continue
		}

		// Configure connection pool for replica
		if err := configureConnectionPool(replicaDB, config); err != nil {
			replicaDB.Close()
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Failed to configure read replica connection pool")
			continue
		}

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := replicaDB.PingContext(ctx); err != nil {
			cancel()
			replicaDB.Close()
			logrus.WithFields(logrus.Fields{
				"host": replicaConfig.Host,
				"port": replicaConfig.Port,
			}).WithError(err).Error("Read replica health check failed")
			continue
		}
		cancel()

		replicaConn := ReadReplicaConnection{
			DB:              replicaDB,
			Host:            replicaConfig.Host,
			Port:            replicaConfig.Port,
			Weight:          replicaConfig.Weight,
			Priority:        replicaConfig.Priority,
			Region:          replicaConfig.Region,
			Healthy:         true,
			Failures:        0,
			LastHealthCheck: time.Now(),
		}

		db.readReplicas = append(db.readReplicas, replicaConn)

		logrus.WithFields(logrus.Fields{
			"host":     replicaConfig.Host,
			"port":     replicaConfig.Port,
			"weight":   replicaConfig.Weight,
			"priority": replicaConfig.Priority,
			"region":   replicaConfig.Region,
		}).Info("Connected to read replica")
	}

	if len(db.readReplicas) == 0 {
		return fmt.Errorf("no read replicas could be initialized")
	}

	logrus.WithField("count", len(db.readReplicas)).Info("Successfully initialized read replicas")
	return nil
}

// startHealthMonitoring starts background health checking for all database connections
// startHealthMonitoring starts background health checking for all database connections
func (db *PostgresDB) startHealthMonitoring() {
	db.healthCheckDone = make(chan bool)
	db.healthCheckTicker = time.NewTicker(db.config.HealthCheckInterval)

	go func() {
		defer db.healthCheckTicker.Stop()
		for {
			select {
			case <-db.healthCheckDone:
				return
			case <-db.healthCheckTicker.C:
				db.performHealthChecks()
			}
		}
	}()

	logrus.WithField("interval", db.config.HealthCheckInterval).Info("Started database health monitoring")
}

// performHealthChecks performs health checks on primary and all read replicas
func (db *PostgresDB) performHealthChecks() {
	// Check primary connection
	db.checkConnectionHealth(db.DB, "primary", db.config.Host, db.config.Port)

	// Check read replicas
	for i := range db.readReplicas {
		replica := &db.readReplicas[i]
		healthy := db.checkConnectionHealth(replica.DB, "replica", replica.Host, replica.Port)
		replica.LastHealthCheck = time.Now()

		if healthy {
			replica.Healthy = true
			replica.Failures = 0
		} else {
			replica.Failures++
			if replica.Failures >= db.config.MaxHealthFailures {
				replica.Healthy = false
				logrus.WithFields(logrus.Fields{
					"host":     replica.Host,
					"port":     replica.Port,
					"failures": replica.Failures,
				}).Warn("Read replica marked as unhealthy")
			}
		}
	}
}

// checkConnectionHealth performs a health check on a single database connection
func (db *PostgresDB) checkConnectionHealth(conn *sql.DB, connType, host string, port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), db.config.HealthCheckTimeout)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": connType,
			"host": host,
			"port": port,
		}).WithError(err).Warn("Database health check failed")
		return false
	}

	return true
}

// getReadReplica returns a healthy read replica connection using weighted load balancing
func (db *PostgresDB) getReadReplica() *sql.DB {
	if !db.readReplicaEnabled || len(db.readReplicas) == 0 {
		return db.DB // Fall back to primary
	}

	// Filter healthy replicas
	var healthyReplicas []ReadReplicaConnection
	for _, replica := range db.readReplicas {
		if replica.Healthy {
			healthyReplicas = append(healthyReplicas, replica)
		}
	}

	if len(healthyReplicas) == 0 {
		logrus.Warn("No healthy read replicas available, falling back to primary")
		return db.DB
	}

	// Simple weighted random selection (could be enhanced with more sophisticated load balancing)
	totalWeight := 0
	for _, replica := range healthyReplicas {
		totalWeight += replica.Weight
	}

	if totalWeight == 0 {
		// If no weights specified, use round-robin
		return healthyReplicas[0].DB
	}

	// Simple random weighted selection
	randomWeight := time.Now().UnixNano() % int64(totalWeight)
	currentWeight := 0

	for _, replica := range healthyReplicas {
		currentWeight += replica.Weight
		if int64(currentWeight) > randomWeight {
			return replica.DB
		}
	}

	// Fallback to first healthy replica
	return healthyReplicas[0].DB
}

// GetReadDB returns a read-optimized database connection (replica if available, primary otherwise)
func (db *PostgresDB) GetReadDB() *sql.DB {
	if db.readReplicaEnabled && len(db.readReplicas) > 0 {
		return db.getReadReplica()
	}
	return db.DB
}

// GetStats returns connection pool statistics for monitoring
func (db *PostgresDB) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Primary connection stats
	primaryStats := db.DB.Stats()
	stats["primary"] = map[string]interface{}{
		"open_connections":    primaryStats.OpenConnections,
		"in_use":              primaryStats.InUse,
		"idle":                primaryStats.Idle,
		"wait_count":          primaryStats.WaitCount,
		"wait_duration":       primaryStats.WaitDuration,
		"max_idle_closed":     primaryStats.MaxIdleClosed,
		"max_lifetime_closed": primaryStats.MaxLifetimeClosed,
	}

	// Read replica stats
	replicas := make([]map[string]interface{}, len(db.readReplicas))
	for i, replica := range db.readReplicas {
		replicaStats := replica.DB.Stats()
		replicas[i] = map[string]interface{}{
			"host":              replica.Host,
			"port":              replica.Port,
			"region":            replica.Region,
			"weight":            replica.Weight,
			"priority":          replica.Priority,
			"healthy":           replica.Healthy,
			"failures":          replica.Failures,
			"last_health_check": replica.LastHealthCheck,
			"open_connections":  replicaStats.OpenConnections,
			"in_use":            replicaStats.InUse,
			"idle":              replicaStats.Idle,
			"wait_count":        replicaStats.WaitCount,
			"wait_duration":     replicaStats.WaitDuration,
		}
	}
	stats["read_replicas"] = replicas

	return stats
}

// TransactionManager returns the transaction manager for advanced transaction handling
func (db *PostgresDB) TransactionManager() *TransactionManager {
	return db.transactionManager
}

// ExecuteInTransaction executes a function within a transaction with timeout
func (db *PostgresDB) ExecuteInTransaction(ctx context.Context, opts *TransactionOptions, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteInTransaction(ctx, opts, fn)
}

// ExecuteInReadTransaction executes read-only operations with snapshot isolation
func (db *PostgresDB) ExecuteInReadTransaction(ctx context.Context, timeout time.Duration, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteInReadTransaction(ctx, timeout, fn)
}

// ExecuteSaga executes a saga pattern transaction
func (db *PostgresDB) ExecuteSaga(ctx context.Context, opts *TransactionOptions, steps []SagaStep) error {
	return db.transactionManager.ExecuteSaga(ctx, opts, steps)
}

// ExecuteWithRetry executes a transaction with retry logic
func (db *PostgresDB) ExecuteWithRetry(ctx context.Context, opts *TransactionOptions, maxRetries int, fn func(*gorm.DB) error) error {
	return db.transactionManager.ExecuteWithRetry(ctx, opts, maxRetries, fn)
}

// NewTransactionScope creates a transaction scope builder
func (db *PostgresDB) NewTransactionScope(ctx context.Context) *TransactionScope {
	return db.transactionManager.NewTransactionScope(ctx)
}

// QueryMonitor returns the query performance monitor
func (db *PostgresDB) QueryMonitor() *QueryMonitor {
	return db.queryMonitor
}

// EnableSlowQueryLogging enables slow query logging
func (db *PostgresDB) EnableSlowQueryLogging(threshold time.Duration) {
	db.queryMonitor.EnableSlowQueryLogging(threshold)
}

// GetQueryStats returns current query performance statistics
func (db *PostgresDB) GetQueryStats() map[string]*QueryStats {
	return db.queryMonitor.GetQueryStats()
}

// GetSlowQueries returns queries exceeding the slow query threshold
func (db *PostgresDB) GetSlowQueries() []*QueryStats {
	return db.queryMonitor.GetSlowQueries()
}

// ============================================================================
// User Profile Operations
// ============================================================================

// GetUserSkills retrieves all skills for a user
func (db *PostgresDB) GetUserSkills(userID uuid.UUID) ([]*UserSkill, error) {
	var skills []*UserSkill
	if err := db.GORM.Where("user_id = ?", userID).Find(&skills).Error; err != nil {
		return nil, fmt.Errorf("failed to get user skills: %w", err)
	}
	return skills, nil
}

// AddUserSkill adds a new skill for a user
func (db *PostgresDB) AddUserSkill(skill *UserSkill) error {
	if err := db.GORM.Create(skill).Error; err != nil {
		return fmt.Errorf("failed to add user skill: %w", err)
	}
	return nil
}

// RemoveUserSkill removes a skill by ID
func (db *PostgresDB) RemoveUserSkill(skillID uuid.UUID) error {
	if err := db.GORM.Delete(&UserSkill{}, skillID).Error; err != nil {
		return fmt.Errorf("failed to remove user skill: %w", err)
	}
	return nil
}

// GetUserAchievements retrieves all achievements for a user
func (db *PostgresDB) GetUserAchievements(userID uuid.UUID) ([]*UserAchievement, error) {
	var achievements []*UserAchievement
	if err := db.GORM.Where("user_id = ?", userID).
		Preload("Achievement").
		Order("earned_at DESC").
		Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}
	return achievements, nil
}

// GetAchievementBySlug retrieves an achievement by its slug
func (db *PostgresDB) GetAchievementBySlug(slug string) (*Achievement, error) {
	var achievement Achievement
	if err := db.GORM.Where("slug = ?", slug).First(&achievement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get achievement by slug: %w", err)
	}
	return &achievement, nil
}

// ListAchievements retrieves all achievement definitions
func (db *PostgresDB) ListAchievements() ([]*Achievement, error) {
	var achievements []*Achievement
	if err := db.GORM.Order("category, name").Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to list achievements: %w", err)
	}
	return achievements, nil
}

// AwardAchievement awards an achievement to a user
func (db *PostgresDB) AwardAchievement(userID, achievementID uuid.UUID, metadata map[string]interface{}) error {
	ua := &UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		EarnedAt:      time.Now(),
		Progress:      100,
		IsCompleted:   true,
		Metadata:      metadata,
	}
	if err := db.GORM.Create(ua).Error; err != nil {
		return fmt.Errorf("failed to award achievement: %w", err)
	}
	return nil
}

// UpdateAchievementProgress updates the progress of a user achievement
func (db *PostgresDB) UpdateAchievementProgress(userAchievementID uuid.UUID, progress int, isCompleted bool) error {
	updates := map[string]interface{}{
		"progress": progress,
	}
	if isCompleted {
		updates["is_completed"] = true
		updates["earned_at"] = time.Now()
	}
	if err := db.GORM.Model(&UserAchievement{}).Where("id = ?", userAchievementID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update achievement progress: %w", err)
	}
	return nil
}

// GetUserActivity retrieves activity feed for a user
func (db *PostgresDB) GetUserActivity(userID uuid.UUID, limit, offset int) ([]*UserActivity, error) {
	var activities []*UserActivity
	if err := db.GORM.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get user activity: %w", err)
	}
	return activities, nil
}

// CreateUserActivity creates a new activity feed item
func (db *PostgresDB) CreateUserActivity(activity *UserActivity) error {
	if err := db.GORM.Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create user activity: %w", err)
	}
	return nil
}

// GetUserExecutionStats retrieves execution statistics for a user
func (db *PostgresDB) GetUserExecutionStats(userID uuid.UUID) (map[string]interface{}, error) {
	// Get user's published functions from registry
	var functions []struct {
		ID             uuid.UUID
		ExecutionCount int64
		UniqueUsers    int64
	}

	if err := db.GORM.Raw(`
		SELECT id, COALESCE(execution_count, 0) as execution_count, COALESCE(unique_users, 0) as unique_users
		FROM registry_functions
		WHERE author_id = ?
	`, userID).Scan(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user functions: %w", err)
	}

	var totalExecutions, totalUniqueUsers int64
	for _, fn := range functions {
		totalExecutions += fn.ExecutionCount
		totalUniqueUsers += fn.UniqueUsers
	}

	// Get execution history for last 30 days
	var history []struct {
		Date       string `json:"date"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as executions
		FROM registry_executions
		WHERE author_id = ? AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date
	`, userID).Scan(&history).Error; err != nil {
		// Continue even if no data
		history = []struct {
			Date       string `json:"date"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"totalExecutions":  totalExecutions,
		"totalUniqueUsers": totalUniqueUsers,
		"functionCount":    len(functions),
		"executionHistory": history,
	}, nil
}

// GetUserPopularFunctions retrieves most popular functions for a user
func (db *PostgresDB) GetUserPopularFunctions(userID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	var functions []map[string]interface{}

	if err := db.GORM.Raw(`
		SELECT
			rf.id,
			rf.name,
			rf.description,
			COALESCE(rf.execution_count, 0) as execution_count,
			COALESCE(rf.rating, 0) as rating,
			COALESCE(rf.total_ratings, 0) as total_ratings
		FROM registry_functions rf
		WHERE rf.author_id = ?
		ORDER BY rf.execution_count DESC NULLS LAST
		LIMIT ?
	`, userID, limit).Scan(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular functions: %w", err)
	}

	return functions, nil
}

// GetUserGeographicStats retrieves geographic distribution of executions
func (db *PostgresDB) GetUserGeographicStats(userID uuid.UUID) (map[string]interface{}, error) {
	var regions []struct {
		Region     string `json:"region"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.Raw(`
		SELECT
			COALESCE(re.region, 'unknown') as region,
			COUNT(*) as executions
		FROM registry_executions re
		WHERE re.author_id = ? AND re.created_at >= NOW() - INTERVAL '30 days'
		GROUP BY re.region
		ORDER BY executions DESC
	`, userID).Scan(&regions).Error; err != nil {
		regions = []struct {
			Region     string `json:"region"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"regions": regions,
	}, nil
}

// GetUserDeviceStats retrieves device/browser statistics
func (db *PostgresDB) GetUserDeviceStats(userID uuid.UUID) (map[string]interface{}, error) {
	var devices []struct {
		Device     string `json:"device"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.Raw(`
		SELECT
			COALESCE(re.user_agent, 'unknown') as device,
			COUNT(*) as executions
		FROM registry_executions re
		WHERE re.author_id = ? AND re.created_at >= NOW() - INTERVAL '30 days'
		GROUP BY re.user_agent
		ORDER BY executions DESC
		LIMIT 5
	`, userID).Scan(&devices).Error; err != nil {
		devices = []struct {
			Device     string `json:"device"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"devices": devices,
	}, nil
}

// Close gracefully shuts down the database connections and health monitoring
func (db *PostgresDB) Close() error {
	// Stop health monitoring
	if db.healthCheckDone != nil {
		db.healthCheckDone <- true
		close(db.healthCheckDone)
	}

	// Close prepared statements
	db.closePreparedStatements()

	// Close read replica connections
	for _, replica := range db.readReplicas {
		if err := replica.DB.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"host": replica.Host,
				"port": replica.Port,
			}).WithError(err).Warn("Failed to close read replica connection")
		}
	}

	return db.DB.Close()
}
