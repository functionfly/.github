package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

// BalanceSnapshotType represents the type of balance snapshot
type BalanceSnapshotType string

const (
	// SnapshotTypeTransactional records balance after each transaction
	SnapshotTypeTransactional BalanceSnapshotType = "transactional"
	// SnapshotTypeScheduled records periodic balance snapshots
	SnapshotTypeScheduled BalanceSnapshotType = "scheduled"
	// SnapshotTypeManual records manually triggered snapshots
	SnapshotTypeManual BalanceSnapshotType = "manual"
	// SnapshotTypeReconciliation records post-reconciliation snapshots
	SnapshotTypeReconciliation BalanceSnapshotType = "reconciliation"
)

// BalanceHistoryEntry represents a single balance history record
type BalanceHistoryEntry struct {
	ID               uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID         uuid.UUID           `json:"wallet_id" gorm:"type:uuid;not null;index"`
	BalanceUSD       float64             `json:"balance_usd" gorm:"type:decimal(14,4);not null"`
	BalanceLocal     *float64            `json:"balance_local,omitempty" gorm:"type:decimal(14,4)"`
	Currency         string              `json:"currency" gorm:"not null;default:'USD'"`
	ChangeAmountUSD  float64             `json:"change_amount_usd" gorm:"type:decimal(14,4);not null;default:0"`
	TransactionID    *uuid.UUID          `json:"transaction_id,omitempty" gorm:"type:uuid;index"`
	RecordedAt       time.Time           `json:"recorded_at" gorm:"not null;default:NOW()"`
	RecordedDate     time.Time           `json:"recorded_date" gorm:"type:date;not null;default:CURRENT_DATE"`
	SnapshotType     BalanceSnapshotType `json:"snapshot_type" gorm:"type:varchar(20);not null;default:'transactional'"`
	Metadata         []byte              `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time           `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the database table name
func (BalanceHistoryEntry) TableName() string {
	return "wallet_balance_history"
}

// BalanceHistoryQuery represents filters for querying balance history
type BalanceHistoryQuery struct {
	WalletID     *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	SnapshotType *BalanceSnapshotType
	Limit        int
	Offset       int
}

// BalanceHistoryResult holds the result of a balance history query
type BalanceHistoryResult struct {
	Entries []BalanceHistoryEntry `json:"entries"`
	Total   int64                 `json:"total"`
	HasMore bool                  `json:"has_more"`
}

// BalanceAnalytics provides time-series analytics for wallet balances
type BalanceAnalytics struct {
	WalletID           uuid.UUID `json:"wallet_id"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	StartingBalance    float64   `json:"starting_balance"`
	EndingBalance      float64   `json:"ending_balance"`
	HighestBalance     float64   `json:"highest_balance"`
	LowestBalance      float64   `json:"lowest_balance"`
	AverageBalance     float64   `json:"average_balance"`
	TotalCredits       float64   `json:"total_credits"`
	TotalDebits        float64   `json:"total_debits"`
	NetChange          float64   `json:"net_change"`
	TransactionCount   int       `json:"transaction_count"`
	DaysWithActivity   int       `json:"days_with_activity"`
}

// RecordBalanceChange records a balance change after a transaction
func (r *Repository) RecordBalanceChange(ctx context.Context, walletID uuid.UUID, newBalance, changeAmount float64, transactionID uuid.UUID, currency string) error {
	entry := &BalanceHistoryEntry{
		WalletID:        walletID,
		BalanceUSD:      newBalance,
		ChangeAmountUSD: changeAmount,
		TransactionID:   &transactionID,
		Currency:        currency,
		RecordedAt:      time.Now(),
		RecordedDate:    time.Now(),
		SnapshotType:    SnapshotTypeTransactional,
	}

	return r.db.WithContext(ctx).Create(entry).Error
}

// RecordScheduledSnapshot creates a scheduled balance snapshot for a wallet
func (r *Repository) RecordScheduledSnapshot(ctx context.Context, walletID uuid.UUID) error {
	// Get current wallet state
	wallet, err := r.GetWalletByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found: %s", walletID)
	}

	entry := &BalanceHistoryEntry{
		WalletID:       walletID,
		BalanceUSD:     wallet.BalanceUSD,
		Currency:       wallet.Currency,
		RecordedAt:     time.Now(),
		RecordedDate:   time.Now(),
		SnapshotType:   SnapshotTypeScheduled,
	}

	return r.db.WithContext(ctx).Create(entry).Error
}

// GetBalanceHistory retrieves balance history for a wallet
func (r *Repository) GetBalanceHistory(ctx context.Context, query BalanceHistoryQuery) (*BalanceHistoryResult, error) {
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 500 {
		query.Limit = 500
	}

	db := r.db.WithContext(ctx).Model(&BalanceHistoryEntry{})

	// Apply filters
	if query.WalletID != nil {
		db = db.Where("wallet_id = ?", *query.WalletID)
	}
	if query.StartDate != nil {
		db = db.Where("recorded_at >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("recorded_at <= ?", *query.EndDate)
	}
	if query.SnapshotType != nil {
		db = db.Where("snapshot_type = ?", *query.SnapshotType)
	}

	// Count total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count balance history: %w", err)
	}

	// Get entries
	var entries []BalanceHistoryEntry
	if err := db.Order("recorded_at DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to list balance history: %w", err)
	}

	return &BalanceHistoryResult{
		Entries: entries,
		Total:   total,
		HasMore: int64(query.Offset+len(entries)) < total,
	}, nil
}

// GetBalanceAtTime retrieves the wallet balance at a specific point in time
func (r *Repository) GetBalanceAtTime(ctx context.Context, walletID uuid.UUID, at time.Time) (*BalanceHistoryEntry, error) {
	var entry BalanceHistoryEntry
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND recorded_at <= ?", walletID, at).
		Order("recorded_at DESC").
		First(&entry).Error

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// CalculateBalanceAnalytics calculates time-series analytics for a wallet
func (r *Repository) CalculateBalanceAnalytics(ctx context.Context, walletID uuid.UUID, periodStart, periodEnd time.Time) (*BalanceAnalytics, error) {
	analytics := &BalanceAnalytics{
		WalletID:    walletID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	// Get starting balance
	startEntry, err := r.GetBalanceAtTime(ctx, walletID, periodStart)
	if err != nil {
		// If no history before period start, get current wallet balance
		wallet, err := r.GetWalletByID(ctx, walletID)
		if err != nil {
			return nil, fmt.Errorf("failed to get starting balance: %w", err)
		}
		if wallet != nil {
			analytics.StartingBalance = wallet.BalanceUSD
		}
	} else {
		analytics.StartingBalance = startEntry.BalanceUSD
	}

	// Get ending balance
	endEntry, err := r.GetBalanceAtTime(ctx, walletID, periodEnd)
	if err != nil {
		// Use starting balance if no later history
		analytics.EndingBalance = analytics.StartingBalance
	} else {
		analytics.EndingBalance = endEntry.BalanceUSD
	}

	// Calculate aggregates from history entries
	var stats struct {
		HighestBalance   float64
		LowestBalance    float64
		AverageBalance   float64
		TotalCredits     float64
		TotalDebits      float64
		EntryCount       int64
		DaysWithActivity int
	}

	// Get balance history in period for statistics
	historyQuery := BalanceHistoryQuery{
		WalletID:  &walletID,
		StartDate: &periodStart,
		EndDate:   &periodEnd,
		Limit:     10000,
	}

	history, err := r.GetBalanceHistory(ctx, historyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}

	if len(history.Entries) > 0 {
		stats.HighestBalance = history.Entries[0].BalanceUSD
		stats.LowestBalance = history.Entries[0].BalanceUSD
		var sumBalance float64

		seenDates := make(map[string]bool)

		for _, entry := range history.Entries {
			if entry.BalanceUSD > stats.HighestBalance {
				stats.HighestBalance = entry.BalanceUSD
			}
			if entry.BalanceUSD < stats.LowestBalance {
				stats.LowestBalance = entry.BalanceUSD
			}
			sumBalance += entry.BalanceUSD

			if entry.ChangeAmountUSD > 0 {
				stats.TotalCredits += entry.ChangeAmountUSD
			} else if entry.ChangeAmountUSD < 0 {
				stats.TotalDebits += -entry.ChangeAmountUSD
			}

			dateKey := entry.RecordedDate.Format("2006-01-02")
			seenDates[dateKey] = true
		}

		stats.EntryCount = int64(len(history.Entries))
		stats.AverageBalance = sumBalance / float64(len(history.Entries))
		stats.DaysWithActivity = len(seenDates)
	}

	analytics.HighestBalance = stats.HighestBalance
	analytics.LowestBalance = stats.LowestBalance
	analytics.AverageBalance = stats.AverageBalance
	analytics.TotalCredits = stats.TotalCredits
	analytics.TotalDebits = stats.TotalDebits
	analytics.NetChange = analytics.EndingBalance - analytics.StartingBalance
	analytics.TransactionCount = int(stats.EntryCount)
	analytics.DaysWithActivity = stats.DaysWithActivity

	return analytics, nil
}

// GetDailyBalanceReport generates a daily balance report for analytics
func (r *Repository) GetDailyBalanceReport(ctx context.Context, walletID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT
			recorded_date,
			balance_usd,
			change_amount_usd,
			snapshot_type,
			transaction_id
		FROM wallet_balance_history
		WHERE wallet_id = ?
			AND recorded_date BETWEEN ? AND ?
		ORDER BY recorded_date ASC, recorded_at ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, walletID, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query daily balance: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var balance, change float64
		var snapshotType string
		var txID *uuid.UUID

		if err := rows.Scan(&date, &balance, &change, &snapshotType, &txID); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"date":           date.Format("2006-01-02"),
			"balance_usd":    balance,
			"change_usd":     change,
			"snapshot_type":  snapshotType,
			"transaction_id": txID,
		})
	}

	return results, nil
}

// BalanceHistoryScheduler creates periodic balance snapshots
type BalanceHistoryScheduler struct {
	repo   *Repository
	logger *logrus.Logger
	cron   string
}

// NewBalanceHistoryScheduler creates a new balance history scheduler
func NewBalanceHistoryScheduler(repo *Repository) *BalanceHistoryScheduler {
	return &BalanceHistoryScheduler{
		repo:   repo,
		logger: logrus.New(),
		cron:   "0 */6 * * *", // Every 6 hours
	}
}

// RunDailySnapshot creates balance snapshots for all active wallets
func (s *BalanceHistoryScheduler) RunDailySnapshot(ctx context.Context) error {
	// Get all active wallets
	var wallets []Wallet
	if err := s.repo.db.WithContext(ctx).Where("status = ?", WalletStatusActive).Find(&wallets).Error; err != nil {
		return fmt.Errorf("failed to list wallets: %w", err)
	}

	s.logger.WithField("wallet_count", len(wallets)).Info("Creating balance snapshots")

	successCount := 0
	failCount := 0

	for _, wallet := range wallets {
		if err := s.repo.RecordScheduledSnapshot(ctx, wallet.ID); err != nil {
			s.logger.WithError(err).WithField("wallet_id", wallet.ID).Error("Failed to create balance snapshot")
			failCount++
		} else {
			successCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"success": successCount,
		"failed":  failCount,
	}).Info("Balance snapshot creation completed")

	return nil
}

// CleanOldSnapshots removes old balance history entries based on retention policy
func (r *Repository) CleanOldSnapshots(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 365 // Default: keep 1 year
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := r.db.WithContext(ctx).
		Where("recorded_at < ? AND snapshot_type = ?", cutoffDate, SnapshotTypeScheduled).
		Delete(&BalanceHistoryEntry{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to clean old snapshots: %w", result.Error)
	}

	return result.RowsAffected, nil
}
