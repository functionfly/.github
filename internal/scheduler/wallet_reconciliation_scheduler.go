package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// WalletReconciliationScheduler runs automated wallet reconciliation jobs
type WalletReconciliationScheduler struct {
	cron              *cron.Cron
	reconciliationSvc *wallet.ReconciliationService
	walletRepo        *wallet.Repository
	redisClient       *redis.Client
	db                *sql.DB
	logger            *logrus.Logger
	notifySvc         *notification.Service

	// Configuration
	CronExpression   string
	Enabled          bool
	AutoFix          bool
	MaxAutoFixAmount float64
	AlertThreshold   int

	// Notification settings
	NotifyOnDiscrepancies  bool
	AdminNotificationEmail string

	// Metrics
	metrics *ReconciliationMetrics
}

// ReconciliationMetrics tracks scheduler performance
type ReconciliationMetrics struct {
	TotalRuns          int64
	SuccessfulRuns     int64
	FailedRuns         int64
	TotalDiscrepancies int64
	TotalFixed         int64
	LastRunTime        time.Time
	LastRunDuration    time.Duration
}

// ReconciliationSchedulerConfig holds scheduler configuration
type ReconciliationSchedulerConfig struct {
	CronExpression         string
	Enabled                bool
	AutoFix                bool
	MaxAutoFixAmount       float64
	AlertThreshold         int
	NotifyOnDiscrepancies  bool
	AdminNotificationEmail string
}

// DefaultReconciliationSchedulerConfig returns default configuration
func DefaultReconciliationSchedulerConfig() *ReconciliationSchedulerConfig {
	return &ReconciliationSchedulerConfig{
		CronExpression:         getEnvString("WALLET_RECONCILIATION_CRON", "0 3 * * *"), // 3 AM daily
		Enabled:                getEnvBool("WALLET_RECONCILIATION_ENABLED", true),
		AutoFix:                getEnvBool("WALLET_RECONCILIATION_AUTO_FIX", false),
		MaxAutoFixAmount:       getEnvFloat64("WALLET_RECONCILIATION_MAX_AUTO_FIX", 0.01),
		AlertThreshold:         getEnvInt("WALLET_RECONCILIATION_ALERT_THRESHOLD", 1),
		NotifyOnDiscrepancies:  getEnvBool("WALLET_RECONCILIATION_NOTIFY", true),
		AdminNotificationEmail: getEnvString("WALLET_RECONCILIATION_ADMIN_EMAIL", ""),
	}
}

// NewWalletReconciliationScheduler creates a new wallet reconciliation scheduler
func NewWalletReconciliationScheduler(
	walletRepo *wallet.Repository,
	redisClient *redis.Client,
	db *sql.DB,
	notifySvc *notification.Service,
) *WalletReconciliationScheduler {
	cfg := DefaultReconciliationSchedulerConfig()

	// Create reconciliation config
	reconciliationConfig := &wallet.ReconciliationConfig{
		CronExpression:   cfg.CronExpression,
		Enabled:          cfg.Enabled,
		AutoFix:          cfg.AutoFix,
		MaxAutoFixAmount: cfg.MaxAutoFixAmount,
		AlertThreshold:   cfg.AlertThreshold,
		Types: []wallet.ReconciliationType{
			wallet.ReconciliationTypeLedgerToBalance,
			wallet.ReconciliationTypeCrossTable,
		},
	}

	reconciliationSvc := wallet.NewReconciliationService(walletRepo, reconciliationConfig)

	return &WalletReconciliationScheduler{
		cron:                   cron.New(cron.WithSeconds()),
		reconciliationSvc:      reconciliationSvc,
		walletRepo:             walletRepo,
		redisClient:            redisClient,
		db:                     db,
		logger:                 logrus.New(),
		notifySvc:              notifySvc,
		CronExpression:         cfg.CronExpression,
		Enabled:                cfg.Enabled,
		AutoFix:                cfg.AutoFix,
		MaxAutoFixAmount:       cfg.MaxAutoFixAmount,
		AlertThreshold:         cfg.AlertThreshold,
		NotifyOnDiscrepancies:  cfg.NotifyOnDiscrepancies,
		AdminNotificationEmail: cfg.AdminNotificationEmail,
		metrics:                &ReconciliationMetrics{},
	}
}

// WithLogger sets a custom logger
func (s *WalletReconciliationScheduler) WithLogger(logger *logrus.Logger) *WalletReconciliationScheduler {
	s.logger = logger
	s.reconciliationSvc.SetLogger(logger)
	return s
}

// SetReconciliationService allows injecting a custom reconciliation service
func (s *WalletReconciliationScheduler) SetReconciliationService(svc *wallet.ReconciliationService) {
	s.reconciliationSvc = svc
}

// Start begins the scheduled reconciliation
func (s *WalletReconciliationScheduler) Start(ctx context.Context) error {
	if !s.Enabled {
		s.logger.Info("Wallet reconciliation scheduler is disabled")
		return nil
	}

	// Ensure reconciliation tables exist
	if err := s.initReconciliationTables(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to init reconciliation tables, continuing anyway")
	}

	// Schedule the reconciliation job
	_, err := s.cron.AddFunc(s.CronExpression, func() {
		s.runScheduledReconciliation(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule reconciliation: %w", err)
	}

	s.cron.Start()
	s.logger.WithFields(logrus.Fields{
		"cron":        s.CronExpression,
		"auto_fix":    s.AutoFix,
		"max_fix_usd": s.MaxAutoFixAmount,
	}).Info("Wallet reconciliation scheduler started")

	return nil
}

// Stop halts the scheduler
func (s *WalletReconciliationScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Wallet reconciliation scheduler stopped")
}

// RunOnce executes a single reconciliation run manually
func (s *WalletReconciliationScheduler) RunOnce(ctx context.Context) (*wallet.ReconciliationRun, error) {
	return s.reconciliationSvc.RunFullReconciliation(ctx, "manual", nil)
}

// RunOnceWithUser executes a single reconciliation run triggered by a specific user
func (s *WalletReconciliationScheduler) RunOnceWithUser(ctx context.Context, userID uuid.UUID) (*wallet.ReconciliationRun, error) {
	return s.reconciliationSvc.RunFullReconciliation(ctx, "manual", &userID)
}

// initReconciliationTables creates necessary database tables
func (s *WalletReconciliationScheduler) initReconciliationTables(ctx context.Context) error {
	// Tables should already exist from migrations, but ensure indexes are present
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_wallet_reconciliation_discrepancies_wallet_id 
		 ON wallet_reconciliation_discrepancies(wallet_id)`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_reconciliation_discrepancies_status 
		 ON wallet_reconciliation_discrepancies(fixed, severity)`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_reconciliation_runs_status 
		 ON wallet_reconciliation_runs(status, created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			s.logger.WithError(err).WithField("query", query).Warn("Failed to create reconciliation index")
		}
	}

	return nil
}

// runScheduledReconciliation performs the scheduled reconciliation
func (s *WalletReconciliationScheduler) runScheduledReconciliation(ctx context.Context) {
	startTime := time.Now()
	s.metrics.LastRunTime = startTime
	s.metrics.TotalRuns++

	s.logger.Info("Starting scheduled wallet reconciliation")

	// Run the reconciliation
	run, err := s.reconciliationSvc.RunFullReconciliation(ctx, "scheduler", nil)
	if err != nil {
		s.logger.WithError(err).Error("Reconciliation run failed")
		s.metrics.FailedRuns++
		s.metrics.LastRunDuration = time.Since(startTime)
		s.notifyOnFailure(ctx, err)
		return
	}

	s.metrics.SuccessfulRuns++
	s.metrics.LastRunDuration = time.Since(startTime)
	s.metrics.TotalDiscrepancies += int64(run.DiscrepanciesFound)
	s.metrics.TotalFixed += int64(run.DiscrepanciesFixed)

	s.logger.WithFields(logrus.Fields{
		"run_id":              run.ID,
		"status":              run.Status,
		"wallets_checked":     run.WalletsChecked,
		"discrepancies_found": run.DiscrepanciesFound,
		"discrepancies_fixed": run.DiscrepanciesFixed,
		"total_delta":         run.TotalAmountDelta,
		"duration":            s.metrics.LastRunDuration,
	}).Info("Scheduled reconciliation completed")

	// Notify on discrepancies if enabled
	if s.NotifyOnDiscrepancies && run.DiscrepanciesFound > 0 {
		s.notifyOnDiscrepancies(ctx, run)
	}

	// Store run result in Redis for quick access
	s.cacheRunResult(ctx, run)
}

// notifyOnFailure sends notification when reconciliation fails
func (s *WalletReconciliationScheduler) notifyOnFailure(ctx context.Context, err error) {
	if s.notifySvc == nil {
		return
	}

	data := map[string]interface{}{
		"error":     err.Error(),
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "wallet_reconciliation",
	}

	// Try to send notification to configured admin email
	if s.AdminNotificationEmail != "" {
		if err := s.notifySvc.SendBillingAlert(ctx, s.AdminNotificationEmail, "reconciliation_failed", data); err != nil {
			s.logger.WithError(err).Warn("Failed to send reconciliation failure notification")
		}
	}
}

// notifyOnDiscrepancies sends notification when discrepancies are found
func (s *WalletReconciliationScheduler) notifyOnDiscrepancies(ctx context.Context, run *wallet.ReconciliationRun) {
	if s.notifySvc == nil {
		return
	}

	// Only notify if above threshold
	if run.DiscrepanciesFound < s.AlertThreshold {
		return
	}

	data := map[string]interface{}{
		"run_id":              run.ID,
		"discrepancies_found": run.DiscrepanciesFound,
		"discrepancies_fixed": run.DiscrepanciesFixed,
		"wallets_checked":     run.WalletsChecked,
		"total_delta_usd":     run.TotalAmountDelta,
		"timestamp":           time.Now().Format(time.RFC3339),
	}

	if s.AdminNotificationEmail != "" {
		if err := s.notifySvc.SendBillingAlert(ctx, s.AdminNotificationEmail, "discrepancies_detected", data); err != nil {
			s.logger.WithError(err).Warn("Failed to send discrepancies notification")
		}
	}
}

// cacheRunResult stores reconciliation result in Redis for quick access
func (s *WalletReconciliationScheduler) cacheRunResult(ctx context.Context, run *wallet.ReconciliationRun) {
	if s.redisClient == nil {
		return
	}

	data, err := json.Marshal(map[string]interface{}{
		"id":                  run.ID,
		"status":              run.Status,
		"type":                run.Type,
		"wallets_checked":     run.WalletsChecked,
		"discrepancies_found": run.DiscrepanciesFound,
		"discrepancies_fixed": run.DiscrepanciesFixed,
		"total_delta":         run.TotalAmountDelta,
		"created_at":          run.CreatedAt,
	})
	if err != nil {
		return
	}

	key := "wallet:reconciliation:last_run"
	if err := s.redisClient.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		s.logger.WithError(err).Debug("Failed to cache reconciliation result")
	}
}

// GetLastRunResult retrieves the last reconciliation result from cache
func (s *WalletReconciliationScheduler) GetLastRunResult(ctx context.Context) (map[string]interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	data, err := s.redisClient.Get(ctx, "wallet:reconciliation:last_run").Result()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMetrics returns current scheduler metrics
func (s *WalletReconciliationScheduler) GetMetrics() *ReconciliationMetrics {
	return s.metrics
}

// GetReconciliationService returns the underlying reconciliation service
func (s *WalletReconciliationScheduler) GetReconciliationService() *wallet.ReconciliationService {
	return s.reconciliationSvc
}

// IsRunning returns true if the scheduler is currently running a reconciliation
func (s *WalletReconciliationScheduler) IsRunning(ctx context.Context) (bool, error) {
	if s.redisClient == nil {
		return false, nil
	}

	exists, err := s.redisClient.Exists(ctx, "wallet:reconciliation:running").Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// markRunning sets the running flag in Redis
func (s *WalletReconciliationScheduler) markRunning(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	return s.redisClient.Set(ctx, "wallet:reconciliation:running", "1", 30*time.Minute).Err()
}

// clearRunning clears the running flag in Redis
func (s *WalletReconciliationScheduler) clearRunning(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	return s.redisClient.Del(ctx, "wallet:reconciliation:running").Err()
}
