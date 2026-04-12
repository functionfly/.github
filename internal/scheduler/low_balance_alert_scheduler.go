package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// LowBalanceAlertScheduler runs proactive low balance monitoring and alerting
type LowBalanceAlertScheduler struct {
	cron           *cron.Cron
	walletRepo     *storage.BillingRepository
	notifySvc      *notification.Service
	userRepo       storage.Repository
	redisClient    *redis.Client
	db             *sql.DB
	logger         *logrus.Logger

	// Configuration
	LowBalanceThresholdUSD float64
	CheckInterval          string
	EnableAutoTopupAlert   bool

	// Throttling configuration
	WarningThrottleDuration  time.Duration
	CriticalThrottleDuration time.Duration
	AutoTopupThrottleDuration time.Duration

	// Alert history retention
	AlertHistoryRetentionDays int

	// Metrics
	metrics *LowBalanceMetrics

	// Last result for monitoring
	lastResult     *LowBalanceCheckResult
	lastResultMu   sync.RWMutex
}

// LowBalanceWallet represents a wallet with low balance for processing
type LowBalanceWallet struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	BalanceUSD            float64
	BalanceLocal          float64
	Currency              string
	AutoTopupEnabled      bool
	AutoTopupThresholdUSD float64
	Suspended             bool
	CreatedAt             time.Time
}

// LowBalanceUser represents user details for notification
type LowBalanceUser struct {
	ID    uuid.UUID
	Email string
	Name  string
}

// LowBalanceAlertRecord represents a sent alert for throttling and audit
type LowBalanceAlertRecord struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	WalletID   uuid.UUID `json:"wallet_id" db:"wallet_id"`
	Severity   string    `json:"severity" db:"severity"`
	BalanceUSD float64   `json:"balance_usd" db:"balance_usd"`
	SentAt     time.Time `json:"sent_at" db:"sent_at"`
	Channel    string    `json:"channel" db:"channel"`
	Status     string    `json:"status" db:"status"`
}

// LowBalanceMetrics tracks scheduler performance
type LowBalanceMetrics struct {
	TotalChecks      int64
	TotalAlertsSent  int64
	TotalErrors      int64
	LastCheckTime    time.Time
	LastCheckLatency time.Duration
}

// Config holds scheduler configuration
type LowBalanceAlertConfig struct {
	LowBalanceThresholdUSD    float64
	CheckInterval             string
	EnableAutoTopupAlert      bool
	WarningThrottleDuration   time.Duration
	CriticalThrottleDuration  time.Duration
	AutoTopupThrottleDuration time.Duration
	AlertHistoryRetentionDays int
}

// DefaultConfig returns default configuration
func DefaultLowBalanceAlertConfig() *LowBalanceAlertConfig {
	return &LowBalanceAlertConfig{
		LowBalanceThresholdUSD:    getEnvFloat64("AGENT_WALLET_LOW_BALANCE_USD", 5.0),
		CheckInterval:             getEnvString("LOW_BALANCE_CHECK_INTERVAL", "*/15 * * * *"), // Every 15 minutes
		EnableAutoTopupAlert:      getEnvBool("ENABLE_AUTO_TOPUP_LOW_BALANCE_ALERT", true),
		WarningThrottleDuration:   getEnvDuration("LOW_BALANCE_WARNING_THROTTLE", 4*time.Hour),
		CriticalThrottleDuration:  getEnvDuration("LOW_BALANCE_CRITICAL_THROTTLE", 1*time.Hour),
		AutoTopupThrottleDuration: getEnvDuration("LOW_BALANCE_AUTOTOPUP_THROTTLE", 24*time.Hour),
		AlertHistoryRetentionDays: getEnvInt("LOW_BALANCE_ALERT_HISTORY_DAYS", 30),
	}
}

// NewLowBalanceAlertScheduler creates a new low balance alert scheduler
func NewLowBalanceAlertScheduler(
	walletRepo *storage.BillingRepository,
	notifySvc *notification.Service,
	userRepo storage.Repository,
	redisClient *redis.Client,
	db *sql.DB,
) *LowBalanceAlertScheduler {
	cfg := DefaultLowBalanceAlertConfig()

	return &LowBalanceAlertScheduler{
		cron:                      cron.New(cron.WithSeconds()),
		walletRepo:                walletRepo,
		notifySvc:                 notifySvc,
		userRepo:                  userRepo,
		redisClient:               redisClient,
		db:                        db,
		logger:                    logrus.New(),
		LowBalanceThresholdUSD:    cfg.LowBalanceThresholdUSD,
		CheckInterval:             cfg.CheckInterval,
		EnableAutoTopupAlert:      cfg.EnableAutoTopupAlert,
		WarningThrottleDuration:   cfg.WarningThrottleDuration,
		CriticalThrottleDuration:  cfg.CriticalThrottleDuration,
		AutoTopupThrottleDuration: cfg.AutoTopupThrottleDuration,
		AlertHistoryRetentionDays: cfg.AlertHistoryRetentionDays,
		metrics:                   &LowBalanceMetrics{},
	}
}

// WithLogger sets a custom logger
func (s *LowBalanceAlertScheduler) WithLogger(logger *logrus.Logger) *LowBalanceAlertScheduler {
	s.logger = logger
	return s
}

// Start begins the scheduled low balance monitoring
func (s *LowBalanceAlertScheduler) Start(ctx context.Context) error {
	// Create alert history table if not exists
	if err := s.initAlertHistoryTable(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to init alert history table, continuing without persistence")
	}

	_, err := s.cron.AddFunc(s.CheckInterval, func() {
		s.runLowBalanceCheck(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule low balance check: %w", err)
	}

	s.cron.Start()
	s.logger.WithFields(logrus.Fields{
		"interval":       s.CheckInterval,
		"threshold_usd":  s.LowBalanceThresholdUSD,
		"warning_throttle": s.WarningThrottleDuration,
		"critical_throttle": s.CriticalThrottleDuration,
	}).Info("Low balance alert scheduler started")
	return nil
}

// Stop halts the scheduler
func (s *LowBalanceAlertScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Low balance alert scheduler stopped")
}

// RunOnce executes a single low balance check manually (for testing or admin use)
func (s *LowBalanceAlertScheduler) RunOnce(ctx context.Context) *LowBalanceCheckResult {
	return s.runLowBalanceCheck(ctx)
}

// GetMetrics returns current scheduler metrics
func (s *LowBalanceAlertScheduler) GetMetrics() *LowBalanceMetrics {
	return s.metrics
}

// GetLastResult returns the result of the most recent check
func (s *LowBalanceAlertScheduler) GetLastResult() *LowBalanceCheckResult {
	s.lastResultMu.RLock()
	defer s.lastResultMu.RUnlock()
	if s.lastResult == nil {
		return nil
	}
	// Return a copy
	resultCopy := *s.lastResult
	return &resultCopy
}

// initAlertHistoryTable creates the alert history table
func (s *LowBalanceAlertScheduler) initAlertHistoryTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS wallet_low_balance_alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
			severity VARCHAR(20) NOT NULL,
			balance_usd DECIMAL(14,4) NOT NULL,
			threshold_usd DECIMAL(14,4) NOT NULL,
			sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			channel VARCHAR(50) NOT NULL DEFAULT 'email',
			status VARCHAR(20) NOT NULL DEFAULT 'sent',
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_wallet_low_balance_alerts_user_id ON wallet_low_balance_alerts(user_id);
		CREATE INDEX IF NOT EXISTS idx_wallet_low_balance_alerts_sent_at ON wallet_low_balance_alerts(sent_at);
	`

	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create alert history table: %w", err)
	}

	return nil
}

// LowBalanceCheckResult contains the results of a low balance check run
type LowBalanceCheckResult struct {
	CheckedAt         time.Time `json:"checked_at"`
	CheckDuration     time.Duration `json:"check_duration_ms"`
	WalletsChecked    int       `json:"wallets_checked"`
	LowBalanceCount   int       `json:"low_balance_count"`
	CriticalCount     int       `json:"critical_count"`
	AlertsSent        int       `json:"alerts_sent"`
	AlertsThrottled   int       `json:"alerts_throttled"`
	EmailsSent        int       `json:"emails_sent"`
	InAppNotifsSent   int       `json:"in_app_notifications_sent"`
	AutoTopupAlerts   int       `json:"auto_topup_alerts_sent"`
	Errors            []string  `json:"errors,omitempty"`
}

// runLowBalanceCheck performs the actual low balance monitoring
func (s *LowBalanceAlertScheduler) runLowBalanceCheck(ctx context.Context) *LowBalanceCheckResult {
	startTime := time.Now()
	result := &LowBalanceCheckResult{
		CheckedAt: time.Now(),
		Errors:    []string{},
	}

	defer func() {
		result.CheckDuration = time.Since(startTime)
		s.lastResultMu.Lock()
		s.lastResult = result
		s.lastResultMu.Unlock()
	}()

	s.metrics.TotalChecks++
	s.metrics.LastCheckTime = startTime

	s.logger.Info("Running proactive low balance check")

	// Get all active wallets with low balance
	lowBalanceWallets, err := s.getWalletsWithLowBalance(ctx, s.LowBalanceThresholdUSD)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get low balance wallets")
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get low balance wallets: %v", err))
		s.metrics.TotalErrors++
		return result
	}

	result.WalletsChecked = len(lowBalanceWallets)

	for _, wallet := range lowBalanceWallets {
		// Skip suspended wallets
		if wallet.Suspended {
			continue
		}

		// Determine severity level
		severity := s.determineSeverity(wallet.BalanceUSD)

		if severity == "critical" {
			result.CriticalCount++
		} else {
			result.LowBalanceCount++
		}

		// Get user details for notification
		user, err := s.getUserByID(ctx, wallet.UserID)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", wallet.UserID).Error("Failed to get user details")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to get user %s: %v", wallet.UserID, err))
			s.metrics.TotalErrors++
			continue
		}

		if user == nil {
			s.logger.WithField("user_id", wallet.UserID).Warn("User not found for low balance wallet")
			continue
		}

		// Check if we've already sent an alert recently (throttling)
		throttled, err := s.shouldThrottleAlert(ctx, user.ID, severity)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to check alert throttle status")
		}
		if throttled {
			s.logger.WithFields(logrus.Fields{
				"user_id":  user.ID,
				"severity": severity,
			}).Debug("Throttling low balance alert")
			result.AlertsThrottled++
			continue
		}

		// Send notifications
		alertStatus, err := s.sendLowBalanceAlert(ctx, user, wallet, severity)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to send low balance alert")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to alert user %s: %v", user.ID, err))
			s.metrics.TotalErrors++
			s.recordAlertHistory(ctx, user.ID, wallet.ID, severity, wallet.BalanceUSD, "failed", err.Error())
			continue
		}

		result.AlertsSent++
		if alertStatus.EmailSent {
			result.EmailsSent++
		}
		if alertStatus.InAppSent {
			result.InAppNotifsSent++
		}

		// Record that we sent an alert (for throttling and audit)
		if err := s.recordAlertSent(ctx, user.ID, wallet.ID, severity, wallet.BalanceUSD); err != nil {
			s.logger.WithError(err).Warn("Failed to record alert sent")
		}

		// Persist to database for history
		s.recordAlertHistory(ctx, user.ID, wallet.ID, severity, wallet.BalanceUSD, "sent", "")

		s.metrics.TotalAlertsSent++

		s.logger.WithFields(logrus.Fields{
			"user_id":       user.ID,
			"wallet_id":     wallet.ID,
			"balance_usd":   wallet.BalanceUSD,
			"threshold_usd": s.LowBalanceThresholdUSD,
			"severity":      severity,
		}).Info("Low balance alert sent")
	}

	// Check for auto-topup enabled users who might be approaching their threshold
	if s.EnableAutoTopupAlert {
		autoTopupCount, err := s.checkAutoTopupUsers(ctx, result)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to check auto-topup users")
		}
		result.AutoTopupAlerts = autoTopupCount
	}

	s.metrics.LastCheckLatency = time.Since(startTime)

	s.logger.WithFields(logrus.Fields{
		"wallets_checked":   result.WalletsChecked,
		"low_balance":     result.LowBalanceCount,
		"critical":        result.CriticalCount,
		"alerts_sent":     result.AlertsSent,
		"alerts_throttled": result.AlertsThrottled,
		"duration_ms":     result.CheckDuration.Milliseconds(),
	}).Info("Low balance check completed")

	return result
}

// determineSeverity categorizes the balance level
func (s *LowBalanceAlertScheduler) determineSeverity(balanceUSD float64) string {
	criticalThreshold := s.LowBalanceThresholdUSD * 0.2 // 20% of threshold is critical
	if balanceUSD <= criticalThreshold {
		return "critical"
	}
	return "warning"
}

// getThrottleDuration returns the throttle duration based on severity
func (s *LowBalanceAlertScheduler) getThrottleDuration(severity string) time.Duration {
	switch severity {
	case "critical":
		return s.CriticalThrottleDuration
	case "autotopup":
		return s.AutoTopupThrottleDuration
	default:
		return s.WarningThrottleDuration
	}
}

// shouldThrottleAlert checks if we should skip sending an alert (to avoid spam)
func (s *LowBalanceAlertScheduler) shouldThrottleAlert(ctx context.Context, userID uuid.UUID, severity string) (bool, error) {
	// Check Redis first if available
	if s.redisClient != nil {
		key := fmt.Sprintf("low_balance_alert:%s:%s", userID.String(), severity)
		exists, err := s.redisClient.Exists(ctx, key).Result()
		if err != nil {
			return false, fmt.Errorf("redis check failed: %w", err)
		}
		if exists > 0 {
			return true, nil
		}
	}

	// Check database for recent alert
	var lastAlertTime time.Time
	query := `
		SELECT MAX(sent_at) FROM wallet_low_balance_alerts
		WHERE user_id = $1 AND severity = $2 AND status = 'sent'
		AND sent_at > NOW() - INTERVAL '1 day'
	`

	err := s.db.QueryRowContext(ctx, query, userID, severity).Scan(&lastAlertTime)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("database check failed: %w", err)
	}

	if err == sql.ErrNoRows || lastAlertTime.IsZero() {
		return false, nil
	}

	// Check if within throttle window
	throttleDuration := s.getThrottleDuration(severity)
	if time.Since(lastAlertTime) < throttleDuration {
		return true, nil
	}

	return false, nil
}

// recordAlertSent records that an alert was sent (for throttling)
func (s *LowBalanceAlertScheduler) recordAlertSent(ctx context.Context, userID, walletID uuid.UUID, severity string, balanceUSD float64) error {
	// Set Redis key with TTL if available
	if s.redisClient != nil {
		key := fmt.Sprintf("low_balance_alert:%s:%s", userID.String(), severity)
		ttl := s.getThrottleDuration(severity)
		if err := s.redisClient.Set(ctx, key, balanceUSD, ttl).Err(); err != nil {
			s.logger.WithError(err).Warn("Failed to set Redis throttle key")
			// Continue to database record
		}
	}

	return nil
}

// recordAlertHistory persists alert to database for audit
func (s *LowBalanceAlertScheduler) recordAlertHistory(ctx context.Context, userID, walletID uuid.UUID, severity string, balanceUSD float64, status, errorMsg string) {
	query := `
		INSERT INTO wallet_low_balance_alerts (user_id, wallet_id, severity, balance_usd, threshold_usd, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query, userID, walletID, severity, balanceUSD, s.LowBalanceThresholdUSD, status, errorMsg)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to record alert history")
	}
}

// AlertStatus tracks which channels successfully sent
type AlertStatus struct {
	EmailSent bool
	InAppSent bool
}

// sendLowBalanceAlert sends notifications to the user
func (s *LowBalanceAlertScheduler) sendLowBalanceAlert(ctx context.Context, user *LowBalanceUser, wallet *LowBalanceWallet, severity string) (*AlertStatus, error) {
	status := &AlertStatus{}

	if user.Email == "" {
		return status, fmt.Errorf("user has no email address")
	}

	data := map[string]interface{}{
		"balance_usd":          wallet.BalanceUSD,
		"balance_local":        wallet.BalanceLocal,
		"currency":             wallet.Currency,
		"threshold_usd":        s.LowBalanceThresholdUSD,
		"severity":             severity,
		"user_id":              user.ID.String(),
		"user_name":            user.Name,
		"wallet_id":            wallet.ID.String(),
		"auto_topup_enabled":   wallet.AutoTopupEnabled,
		"auto_topup_threshold": wallet.AutoTopupThresholdUSD,
	}

	// Send email notification
	if err := s.notifySvc.SendLowBalance(ctx, user.Email, data); err != nil {
		return status, fmt.Errorf("failed to send email: %w", err)
	}
	status.EmailSent = true

	// Send in-app notification if available
	if err := s.notifySvc.SendLowBalanceNotification(ctx, user.ID, data); err != nil {
		s.logger.WithError(err).Warn("Failed to send in-app low balance notification")
		// Don't fail for in-app notification
	} else {
		status.InAppSent = true
	}

	return status, nil
}

// checkAutoTopupUsers checks for users with auto-topup who are below their threshold
func (s *LowBalanceAlertScheduler) checkAutoTopupUsers(ctx context.Context, result *LowBalanceCheckResult) (int, error) {
	// Get users with auto-topup enabled who are approaching their threshold
	topupUsers, err := s.getAutoTopupUsersBelowThreshold(ctx, s.LowBalanceThresholdUSD*2)
	if err != nil {
		return 0, fmt.Errorf("failed to get auto-topup users: %w", err)
	}

	count := 0
	for _, wallet := range topupUsers {
		// Check throttling
		throttled, err := s.shouldThrottleAlert(ctx, wallet.UserID, "autotopup")
		if err != nil {
			s.logger.WithError(err).WithField("user_id", wallet.UserID).Warn("Failed to check auto-topup throttle")
		}
		if throttled {
			continue
		}

		// Get user details
		user, err := s.getUserByID(ctx, wallet.UserID)
		if err != nil || user == nil {
			continue
		}

		// Send notification
		if err := s.sendAutoTopupApproachingAlert(ctx, user, wallet); err != nil {
			s.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to send auto-topup approaching alert")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to send auto-topup alert to %s: %v", user.ID, err))
		} else {
			count++
			s.recordAlertSent(ctx, user.ID, wallet.ID, "autotopup", wallet.BalanceUSD)
			s.recordAlertHistory(ctx, user.ID, wallet.ID, "autotopup", wallet.BalanceUSD, "sent", "")
		}
	}

	return count, nil
}

// sendAutoTopupApproachingAlert notifies users that auto-topup is approaching
func (s *LowBalanceAlertScheduler) sendAutoTopupApproachingAlert(ctx context.Context, user *LowBalanceUser, wallet *LowBalanceWallet) error {
	data := map[string]interface{}{
		"balance_usd":          wallet.BalanceUSD,
		"balance_local":        wallet.BalanceLocal,
		"currency":             wallet.Currency,
		"auto_topup_threshold": wallet.AutoTopupThresholdUSD,
		"user_id":              user.ID.String(),
		"user_name":            user.Name,
		"wallet_id":            wallet.ID.String(),
	}

	return s.notifySvc.SendAutoTopupApproaching(ctx, user.ID, data)
}

// getWalletsWithLowBalance retrieves wallets below the threshold
func (s *LowBalanceAlertScheduler) getWalletsWithLowBalance(ctx context.Context, threshold float64) ([]*LowBalanceWallet, error) {
	query := `
		SELECT w.id, w.user_id, w.balance_usd, w.balance_local, w.currency,
		       w.auto_topup_enabled, w.auto_topup_threshold_usd, w.suspended, w.created_at
		FROM wallets w
		WHERE w.balance_usd < $1
		AND w.suspended = false
		AND w.closed_at IS NULL
		ORDER BY w.balance_usd ASC
	`

	rows, err := s.db.QueryContext(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var wallets []*LowBalanceWallet
	for rows.Next() {
		w := &LowBalanceWallet{}
		err := rows.Scan(
			&w.ID, &w.UserID, &w.BalanceUSD, &w.BalanceLocal, &w.Currency,
			&w.AutoTopupEnabled, &w.AutoTopupThresholdUSD, &w.Suspended, &w.CreatedAt,
		)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan wallet row")
			continue
		}
		wallets = append(wallets, w)
	}

	return wallets, rows.Err()
}

// getAutoTopupUsersBelowThreshold retrieves users with auto-topup enabled approaching threshold
func (s *LowBalanceAlertScheduler) getAutoTopupUsersBelowThreshold(ctx context.Context, threshold float64) ([]*LowBalanceWallet, error) {
	query := `
		SELECT w.id, w.user_id, w.balance_usd, w.balance_local, w.currency,
		       w.auto_topup_enabled, w.auto_topup_threshold_usd, w.suspended, w.created_at
		FROM wallets w
		WHERE w.auto_topup_enabled = true
		AND w.balance_usd < $1
		AND w.balance_usd > $2
		AND w.suspended = false
		AND w.closed_at IS NULL
		ORDER BY w.balance_usd ASC
	`

	rows, err := s.db.QueryContext(ctx, query, threshold, s.LowBalanceThresholdUSD)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var wallets []*LowBalanceWallet
	for rows.Next() {
		w := &LowBalanceWallet{}
		err := rows.Scan(
			&w.ID, &w.UserID, &w.BalanceUSD, &w.BalanceLocal, &w.Currency,
			&w.AutoTopupEnabled, &w.AutoTopupThresholdUSD, &w.Suspended, &w.CreatedAt,
		)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan auto-topup wallet row")
			continue
		}
		wallets = append(wallets, w)
	}

	return wallets, rows.Err()
}

// getUserByID retrieves user details
func (s *LowBalanceAlertScheduler) getUserByID(ctx context.Context, userID uuid.UUID) (*LowBalanceUser, error) {
	query := `SELECT id, email, name FROM users WHERE id = $1`

	u := &LowBalanceUser{}
	var name sql.NullString
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.Email, &name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if name.Valid {
		u.Name = name.String
	}

	return u, nil
}

// CleanupOldAlertHistory removes old alert history records
func (s *LowBalanceAlertScheduler) CleanupOldAlertHistory(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM wallet_low_balance_alerts
		WHERE sent_at < NOW() - INTERVAL '%d days'
	`

	result, err := s.db.ExecContext(ctx, fmt.Sprintf(query, s.AlertHistoryRetentionDays))
	if err != nil {
		return 0, fmt.Errorf("cleanup failed: %w", err)
	}

	return result.RowsAffected()
}

// GetRecentAlerts retrieves recent low balance alerts (for admin dashboard)
func (s *LowBalanceAlertScheduler) GetRecentAlerts(ctx context.Context, limit int) ([]*LowBalanceAlertRecord, error) {
	query := `
		SELECT id, user_id, wallet_id, severity, balance_usd, sent_at, channel, status
		FROM wallet_low_balance_alerts
		ORDER BY sent_at DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []*LowBalanceAlertRecord
	for rows.Next() {
		a := &LowBalanceAlertRecord{}
		err := rows.Scan(&a.ID, &a.UserID, &a.WalletID, &a.Severity, &a.BalanceUSD, &a.SentAt, &a.Channel, &a.Status)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan alert row")
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// GetUserAlertHistory retrieves alert history for a specific user
func (s *LowBalanceAlertScheduler) GetUserAlertHistory(ctx context.Context, userID uuid.UUID, days int) ([]*LowBalanceAlertRecord, error) {
	query := `
		SELECT id, user_id, wallet_id, severity, balance_usd, sent_at, channel, status
		FROM wallet_low_balance_alerts
		WHERE user_id = $1 AND sent_at > NOW() - INTERVAL '%d days'
		ORDER BY sent_at DESC
	`

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(query, days), userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []*LowBalanceAlertRecord
	for rows.Next() {
		a := &LowBalanceAlertRecord{}
		err := rows.Scan(&a.ID, &a.UserID, &a.WalletID, &a.Severity, &a.BalanceUSD, &a.SentAt, &a.Channel, &a.Status)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan alert row")
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// ToJSON serializes the result to JSON (for logging/metrics)
func (r *LowBalanceCheckResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// Environment helpers
func getEnvFloat64(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
