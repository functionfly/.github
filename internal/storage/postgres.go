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
	skipPreparedStatements       bool     // For migrations, skip prepared statements
	*sql.DB                               // Primary write connection
	GORM                         *gorm.DB // GORM instance for ORM operations
	userRepository               *UserRepository
	tenantRepository             *TenantRepository
	billingRepository            *BillingRepository
	revenueRepository            *RevenueRepository
	auditRepository              *AuditRepository
	appRepository                *AppRepository
	backendRepository            *BackendRepository
	deploymentRepository         *DeploymentRepository
	contentRepository            *ContentRepository
	feedbackRepository           *FeedbackRepository
	monitoringRepository         *MonitoringRepository
	sessionRepository            *SessionRepository
	refreshTokenRepository       *RefreshTokenRepository
	loginAttemptRepository       *LoginAttemptRepository
	authEventRepository          *AuthEventRepository
	localRuntimeRepository       *LocalRuntimeRepository
	functionRepository           *FunctionRepository
	registryRepository           *registry.RegistryRepository
	incidentRepository           *IncidentRepository
	featureMeasureRepository     *FeatureMeasureRepository
	teamRepository               *TeamRepository
	followRepository             *FollowRepository
	favoriteRepository           *FavoriteRepository
	adminSessionRepository       *AdminSessionRepository
	analyticsRepository          *AnalyticsRepository
	usageAlertRepository         *UsageAlertRepository
	encryptionManager            *DatabaseEncryptionManager
	exportRepository             *ExportRepository
	teamMemoryRepository         TeamMemoryRepository
	creditNoteRepository         *CreditNoteRepository
	tenantStripeConfigRepository *TenantStripeConfigRepository
	certificationRepository      *CertificationRepository
	employeeRepository           *EmployeeRepository
	ffidRepository              *FFIDRepository
	phase2Repository            *Phase2Repository
	phase3Repository           *Phase3Repository
	phase4Repository           *Phase4Repository
	phase5Repository           *Phase5Repository
	phase6Repository           *Phase6Repository
	remainingRepository          *RemainingRepository

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
	stmtInitialized     bool // Tracks whether InitPreparedStatements completed successfully
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
	postgresDB.favoriteRepository = NewFavoriteRepository(postgresDB)
	postgresDB.adminSessionRepository = NewAdminSessionRepository(postgresDB)
	postgresDB.analyticsRepository = NewAnalyticsRepository(postgresDB)
	postgresDB.usageAlertRepository = NewUsageAlertRepository(postgresDB.DB)
	postgresDB.exportRepository = NewExportRepository(postgresDB.DB)
	postgresDB.teamMemoryRepository = NewTeamMemoryRepository(postgresDB.GORM, nil)
	postgresDB.creditNoteRepository = NewCreditNoteRepositorySQL(postgresDB)
	postgresDB.tenantStripeConfigRepository = NewTenantStripeConfigRepository(postgresDB.DB)
	postgresDB.certificationRepository = NewCertificationRepository(postgresDB)

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

// ExportRepository accessor
func (db *PostgresDB) ExportRepository() *ExportRepository {
	return db.exportRepository
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

func (db *PostgresDB) EncryptField(ctx context.Context, value string) (string, error) {
	return db.encryptionManager.EncryptField(value)
}

func (db *PostgresDB) DecryptField(ctx context.Context, value string) (string, error) {
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

func (db *PostgresDB) AnalyticsRepository() *AnalyticsRepository {
	return db.analyticsRepository
}

// CertificationRepository accessor
func (db *PostgresDB) CertificationRepository() *CertificationRepository {
	return db.certificationRepository
}

// InitializeTenantAnalytics creates default analytics tracking for a tenant
func (db *PostgresDB) InitializeTenantAnalytics(tenantID uuid.UUID) error {
	return db.analyticsRepository.InitializeTenantAnalytics(tenantID)
}

// Usage Export operations - delegate to exportRepository
func (db *PostgresDB) CreateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error {
	return db.exportRepository.CreateUsageExportConfiguration(ctx, config)
}

func (db *PostgresDB) GetUsageExportConfiguration(ctx context.Context, id uuid.UUID) (*UsageExportConfiguration, error) {
	return db.exportRepository.GetUsageExportConfiguration(ctx, id)
}

func (db *PostgresDB) UpdateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error {
	return db.exportRepository.UpdateUsageExportConfiguration(ctx, config)
}

func (db *PostgresDB) DeleteUsageExportConfiguration(ctx context.Context, id uuid.UUID) error {
	return db.exportRepository.DeleteUsageExportConfiguration(ctx, id)
}

func (db *PostgresDB) ListUsageExportConfigurations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportConfiguration, error) {
	return db.exportRepository.ListUsageExportConfigurations(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) CreateUsageExportJob(ctx context.Context, job *UsageExportJob) error {
	return db.exportRepository.CreateUsageExportJob(ctx, job)
}

func (db *PostgresDB) GetUsageExportJob(ctx context.Context, id uuid.UUID) (*UsageExportJob, error) {
	return db.exportRepository.GetUsageExportJob(ctx, id)
}

func (db *PostgresDB) UpdateUsageExportJobStatus(ctx context.Context, id uuid.UUID, status UsageExportStatus, errorMessage string) error {
	return db.exportRepository.UpdateUsageExportJobStatus(ctx, id, status, errorMessage)
}

func (db *PostgresDB) CompleteUsageExportJob(ctx context.Context, id uuid.UUID, storagePath, storageURL, checksum string, recordCount, fileSize int64) error {
	return db.exportRepository.CompleteUsageExportJob(ctx, id, storagePath, storageURL, checksum, recordCount, fileSize)
}

func (db *PostgresDB) UpdateDeliveryStatus(ctx context.Context, jobID uuid.UUID, status, errorMessage string) error {
	return db.exportRepository.UpdateDeliveryStatus(ctx, jobID, status, errorMessage)
}

func (db *PostgresDB) UpdateLastExecution(ctx context.Context, configID, jobID uuid.UUID, executedAt time.Time) error {
	return db.exportRepository.UpdateLastExecution(ctx, configID, jobID, executedAt)
}

func (db *PostgresDB) ListUsageExportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportJob, error) {
	return db.exportRepository.ListUsageExportJobs(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) GetPendingScheduledConfigs(ctx context.Context, now time.Time) ([]*UsageExportConfiguration, error) {
	return db.exportRepository.GetPendingScheduledConfigs(ctx, now)
}

func (db *PostgresDB) CreateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error {
	return db.exportRepository.CreateExternalBillingSystem(ctx, system)
}

func (db *PostgresDB) GetExternalBillingSystem(ctx context.Context, id uuid.UUID) (*ExternalBillingSystem, error) {
	return db.exportRepository.GetExternalBillingSystem(ctx, id)
}

func (db *PostgresDB) UpdateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error {
	return db.exportRepository.UpdateExternalBillingSystem(ctx, system)
}

func (db *PostgresDB) DeleteExternalBillingSystem(ctx context.Context, id uuid.UUID) error {
	return db.exportRepository.DeleteExternalBillingSystem(ctx, id)
}

func (db *PostgresDB) ListExternalBillingSystems(ctx context.Context, tenantID uuid.UUID, limit, offset int, activeOnly bool) ([]*ExternalBillingSystem, error) {
	return db.exportRepository.ListExternalBillingSystems(ctx, tenantID, limit, offset, activeOnly)
}

func (db *PostgresDB) CreateBillingIntegrationSync(ctx context.Context, sync *BillingIntegrationSync) error {
	return db.exportRepository.CreateBillingIntegrationSync(ctx, sync)
}

func (db *PostgresDB) GetBillingIntegrationSync(ctx context.Context, id uuid.UUID) (*BillingIntegrationSync, error) {
	return db.exportRepository.GetBillingIntegrationSync(ctx, id)
}

func (db *PostgresDB) ListBillingIntegrationSyncs(ctx context.Context, tenantID uuid.UUID, systemID *uuid.UUID, status string, limit, offset int) ([]*BillingIntegrationSync, error) {
	return db.exportRepository.ListBillingIntegrationSyncs(ctx, tenantID, systemID, status, limit, offset)
}

func (db *PostgresDB) CreateUsageExportTemplate(ctx context.Context, template *UsageExportTemplate) error {
	return db.exportRepository.CreateUsageExportTemplate(ctx, template)
}

func (db *PostgresDB) GetUsageExportTemplate(ctx context.Context, id uuid.UUID) (*UsageExportTemplate, error) {
	return db.exportRepository.GetUsageExportTemplate(ctx, id)
}

func (db *PostgresDB) ListUsageExportTemplates(ctx context.Context, category string) ([]*UsageExportTemplate, error) {
	return db.exportRepository.ListUsageExportTemplates(ctx, category)
}

// ============================================
// Team Memory Repository Methods (Shared Brain)
// ============================================

func (db *PostgresDB) CreateTeamMemory(ctx context.Context, memory *TeamMemory) (*TeamMemory, error) {
	return db.teamMemoryRepository.Create(ctx, memory)
}

func (db *PostgresDB) GetTeamMemoryByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*TeamMemory, error) {
	return db.teamMemoryRepository.GetByID(ctx, tenantID, teamID, memoryID)
}

func (db *PostgresDB) UpdateTeamMemory(ctx context.Context, memory *TeamMemory) (*TeamMemory, error) {
	return db.teamMemoryRepository.Update(ctx, memory)
}

func (db *PostgresDB) DeleteTeamMemory(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error {
	return db.teamMemoryRepository.Delete(ctx, tenantID, teamID, memoryID)
}

func (db *PostgresDB) ListTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, filter TeamMemoryFilter) ([]*TeamMemory, int64, error) {
	return db.teamMemoryRepository.ListByTeam(ctx, tenantID, teamID, filter)
}

func (db *PostgresDB) ListTeamMemoriesByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*TeamMemory, error) {
	return db.teamMemoryRepository.ListByType(ctx, tenantID, teamID, memoryType, limit, offset)
}

func (db *PostgresDB) SearchTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, query string, limit int) ([]*TeamMemorySearchResult, error) {
	return db.teamMemoryRepository.SearchByText(ctx, tenantID, teamID, query, limit)
}

func (db *PostgresDB) SearchTeamMemoriesByVector(ctx context.Context, tenantID, teamID uuid.UUID, embedding []float32, limit int) ([]*TeamMemorySearchResult, error) {
	return db.teamMemoryRepository.SearchSimilar(ctx, tenantID, teamID, embedding, limit)
}

func (db *PostgresDB) SearchTeamMemoriesFallback(ctx context.Context, tenantID, teamID uuid.UUID, query string, memoryType, category string, limit int) ([]*TeamMemorySearchResult, error) {
	return db.teamMemoryRepository.SearchFallback(ctx, tenantID, teamID, query, memoryType, category, limit)
}

func (db *PostgresDB) ValidateTeamMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error {
	return db.teamMemoryRepository.ValidateMemory(ctx, memoryID, validatedBy)
}

func (db *PostgresDB) MarkTeamMemoryAsAccessed(ctx context.Context, memoryID uuid.UUID) error {
	return db.teamMemoryRepository.MarkAsAccessed(ctx, memoryID)
}

func (db *PostgresDB) CreateEncryptedTeamMemory(ctx context.Context, memory *TeamMemory, encryptedContent, iv, tag []byte) (*TeamMemory, error) {
	return db.teamMemoryRepository.CreateEncryptedMemory(ctx, memory, encryptedContent, iv, tag)
}

func (db *PostgresDB) GetTeamMemoryDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error) {
	return db.teamMemoryRepository.GetDecryptionPayload(ctx, memoryID)
}

func (db *PostgresDB) CreateMemoryExtraction(ctx context.Context, extraction *MemoryExtraction) (*MemoryExtraction, error) {
	return db.teamMemoryRepository.CreateExtraction(ctx, extraction)
}

func (db *PostgresDB) GetMemoryExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*MemoryExtraction, error) {
	return db.teamMemoryRepository.GetExtractionsByTeam(ctx, teamID, status, limit)
}

func (db *PostgresDB) ApproveMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*TeamMemory, error) {
	return db.teamMemoryRepository.ApproveExtraction(ctx, extractionID, reviewedBy)
}

func (db *PostgresDB) RejectMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error {
	return db.teamMemoryRepository.RejectExtraction(ctx, extractionID, reviewedBy, reason)
}

func (db *PostgresDB) ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error) {
	return db.teamMemoryRepository.ProcessAutoApplyExtractions(ctx, batchSize)
}

// Memory sharing methods (cross-team collaboration)
func (db *PostgresDB) CreateMemoryShare(ctx context.Context, share *MemoryShare) error {
	return db.GORM.WithContext(ctx).Create(share).Error
}

func (db *PostgresDB) GetMemoryShareByID(ctx context.Context, shareID uuid.UUID) (*MemoryShare, error) {
	var share MemoryShare
	err := db.GORM.WithContext(ctx).First(&share, "id = ?", shareID).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (db *PostgresDB) GetMemoryShareBetweenTeams(ctx context.Context, memoryID, sourceTeamID, targetTeamID uuid.UUID) (*MemoryShare, error) {
	var share MemoryShare
	err := db.GORM.WithContext(ctx).
		Where("memory_id = ? AND source_team_id = ? AND target_team_id = ?", memoryID, sourceTeamID, targetTeamID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (db *PostgresDB) UpdateMemoryShare(ctx context.Context, share *MemoryShare) error {
	return db.GORM.WithContext(ctx).Save(share).Error
}

func (db *PostgresDB) ListMemorySharesByTargetTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*MemoryShare, error) {
	var shares []*MemoryShare
	query := db.GORM.WithContext(ctx).Where("target_team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Limit(limit).Offset(offset).Find(&shares).Error
	return shares, err
}

func (db *PostgresDB) ListMemorySharesBySourceTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*MemoryShare, error) {
	var shares []*MemoryShare
	query := db.GORM.WithContext(ctx).Where("source_team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Limit(limit).Offset(offset).Find(&shares).Error
	return shares, err
}

func (db *PostgresDB) ListMemorySharesByMemoryID(ctx context.Context, memoryID uuid.UUID, status string) ([]*MemoryShare, error) {
	var shares []*MemoryShare
	query := db.GORM.WithContext(ctx).Where("memory_id = ?", memoryID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&shares).Error
	return shares, err
}

// ── Tenant Auth Settings (Backend-in-a-Box) ─────────────────────────────────

func (db *PostgresDB) CreateAuthSettings(ctx context.Context, settings *TenantAuthSettings) error {
	return db.billingRepository.CreateAuthSettings(ctx, settings)
}

func (db *PostgresDB) GetAuthSettings(ctx context.Context, tenantID uuid.UUID) (*TenantAuthSettings, error) {
	return db.billingRepository.GetAuthSettings(ctx, tenantID)
}

func (db *PostgresDB) UpdateAuthSettings(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error {
	return db.billingRepository.UpdateAuthSettings(ctx, tenantID, updates)
}

func (db *PostgresDB) DeleteAuthSettings(ctx context.Context, tenantID uuid.UUID) error {
	return db.billingRepository.DeleteAuthSettings(ctx, tenantID)
}

func (db *PostgresDB) CreateOAuthProvider(ctx context.Context, provider *TenantOAuthProvider) error {
	return db.billingRepository.CreateOAuthProvider(ctx, provider)
}

func (db *PostgresDB) GetOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*TenantOAuthProvider, error) {
	return db.billingRepository.GetOAuthProvider(ctx, tenantID, provider)
}

func (db *PostgresDB) ListOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error) {
	return db.billingRepository.ListOAuthProviders(ctx, tenantID)
}

func (db *PostgresDB) UpdateOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string, updates map[string]interface{}) (*TenantOAuthProvider, error) {
	return db.billingRepository.UpdateOAuthProvider(ctx, tenantID, provider, updates)
}

func (db *PostgresDB) DeleteOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) error {
	return db.billingRepository.DeleteOAuthProvider(ctx, tenantID, provider)
}

func (db *PostgresDB) GetEnabledOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error) {
	return db.billingRepository.GetEnabledOAuthProviders(ctx, tenantID)
}

func (db *PostgresDB) CreateInviteCode(ctx context.Context, invite *TenantInviteCode) error {
	return db.billingRepository.CreateInviteCode(ctx, invite)
}

func (db *PostgresDB) GetInviteCode(ctx context.Context, code string) (*TenantInviteCode, error) {
	return db.billingRepository.GetInviteCode(ctx, code)
}

func (db *PostgresDB) GetInviteCodesByTenant(ctx context.Context, tenantID uuid.UUID, includeUsed bool) ([]*TenantInviteCode, error) {
	return db.billingRepository.GetInviteCodesByTenant(ctx, tenantID, includeUsed)
}

func (db *PostgresDB) GetInviteCodeByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*TenantInviteCode, error) {
	return db.billingRepository.GetInviteCodeByEmail(ctx, tenantID, email)
}

func (db *PostgresDB) AcceptInviteCode(ctx context.Context, code string, userID uuid.UUID) error {
	return db.billingRepository.AcceptInviteCode(ctx, code, userID)
}

func (db *PostgresDB) RevokeInviteCode(ctx context.Context, code string) error {
	return db.billingRepository.RevokeInviteCode(ctx, code)
}

func (db *PostgresDB) IncrementInviteCodeUses(ctx context.Context, code string) error {
	return db.billingRepository.IncrementInviteCodeUses(ctx, code)
}

func (db *PostgresDB) DeleteExpiredInviteCodes(ctx context.Context) (int64, error) {
	return db.billingRepository.DeleteExpiredInviteCodes(ctx)
}

func (db *PostgresDB) CreateMembership(ctx context.Context, membership *TenantMembership) error {
	return db.billingRepository.CreateMembership(ctx, membership)
}

func (db *PostgresDB) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*TenantMembership, error) {
	return db.billingRepository.GetMembership(ctx, tenantID, userID)
}

func (db *PostgresDB) ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]*TenantMembership, error) {
	return db.billingRepository.ListMemberships(ctx, tenantID)
}

func (db *PostgresDB) ListMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*TenantMembership, error) {
	return db.billingRepository.ListMembershipsByRole(ctx, tenantID, role)
}

func (db *PostgresDB) UpdateMembership(ctx context.Context, tenantID, userID uuid.UUID, updates map[string]interface{}) (*TenantMembership, error) {
	return db.billingRepository.UpdateMembership(ctx, tenantID, userID, updates)
}

func (db *PostgresDB) DeleteMembership(ctx context.Context, tenantID, userID uuid.UUID) error {
	return db.billingRepository.DeleteMembership(ctx, tenantID, userID)
}

func (db *PostgresDB) UpdateMembershipLastActive(ctx context.Context, tenantID, userID uuid.UUID) error {
	return db.billingRepository.UpdateMembershipLastActive(ctx, tenantID, userID)
}

func (db *PostgresDB) CountMembershipsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return db.billingRepository.CountMembershipsByTenant(ctx, tenantID)
}

func (db *PostgresDB) CountMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) (int, error) {
	return db.billingRepository.CountMembershipsByRole(ctx, tenantID, role)
}

func (db *PostgresDB) CountActiveFounderModeRegistrations(ctx context.Context) (int, error) {
	return db.billingRepository.CountActiveFounderModeRegistrations(ctx)
}

func (db *PostgresDB) CountRecentSuccessfulDeployments(ctx context.Context) (int, error) {
	return db.billingRepository.CountRecentSuccessfulDeployments(ctx)
}

func (db *PostgresDB) CreateAuthAuditLog(ctx context.Context, log *TenantAuthAuditLog) error {
	return db.billingRepository.CreateAuthAuditLog(ctx, log)
}

func (db *PostgresDB) ListAuthAuditLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int, actions []string, userID *uuid.UUID, since *time.Time) ([]*TenantAuthAuditLog, int, error) {
	return db.billingRepository.ListAuthAuditLogs(ctx, tenantID, limit, offset, actions, userID, since)
}

func (db *PostgresDB) GetAuthAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*TenantAuthAuditLog, error) {
	return db.billingRepository.GetAuthAuditLogsByUser(ctx, tenantID, userID, limit)
}

func (db *PostgresDB) DeleteOldAuthAuditLogs(ctx context.Context, before time.Time) (int64, error) {
	return db.billingRepository.DeleteOldAuthAuditLogs(ctx, before)
}
