package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ReconciliationStatus represents the status of a reconciliation run
type ReconciliationStatus string

const (
	ReconciliationStatusPending    ReconciliationStatus = "pending"
	ReconciliationStatusRunning    ReconciliationStatus = "running"
	ReconciliationStatusCompleted ReconciliationStatus = "completed"
	ReconciliationStatusFailed    ReconciliationStatus = "failed"
	ReconciliationStatusPartial  ReconciliationStatus = "partial"
)

// ReconciliationType represents what kind of reconciliation was performed
type ReconciliationType string

const (
	ReconciliationTypeLedgerToBalance ReconciliationType = "ledger_to_balance"
	ReconciliationTypeStripeToInternal  ReconciliationType = "stripe_to_internal"
	ReconciliationTypeCrossTable      ReconciliationType = "cross_table"
	ReconciliationTypeFull            ReconciliationType = "full"
)

// ReconciliationRun tracks a reconciliation execution
type ReconciliationRun struct {
	ID                 uuid.UUID            `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type               ReconciliationType   `json:"type" gorm:"not null"`
	Status             ReconciliationStatus `json:"status" gorm:"not null;default:'pending'"`
	StartedAt          time.Time            `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt        *time.Time           `json:"completed_at,omitempty"`
	WalletsChecked     int                  `json:"wallets_checked" gorm:"default:0"`
	WalletsWithIssues  int                  `json:"wallets_with_issues" gorm:"default:0"`
	DiscrepanciesFound int                  `json:"discrepancies_found" gorm:"default:0"`
	DiscrepanciesFixed int                  `json:"discrepancies_fixed" gorm:"default:0"`
	TotalAmountDelta   float64              `json:"total_amount_delta" gorm:"type:decimal(14,4);default:0"`
	ErrorMessage       *string              `json:"error_message,omitempty"`
	Details            []byte               `json:"details,omitempty" gorm:"type:jsonb;default:'{}'"`
	TriggeredBy        string               `json:"triggered_by" gorm:"not null;default:'scheduler'"` // 'scheduler', 'manual', 'webhook'
	TriggeredByUserID  *uuid.UUID           `json:"triggered_by_user_id,omitempty"`
	CreatedAt          time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the database table name
func (ReconciliationRun) TableName() string {
	return "wallet_reconciliation_runs"
}

// ReconciliationDiscrepancy represents a specific issue found during reconciliation
type ReconciliationDiscrepancy struct {
	ID                  uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RunID               uuid.UUID    `json:"run_id" gorm:"type:uuid;not null;index"`
	WalletID            uuid.UUID    `json:"wallet_id" gorm:"type:uuid;not null;index"`
	DiscrepancyType     string       `json:"discrepancy_type" gorm:"not null"` // 'balance_mismatch', 'missing_transaction', 'duplicate_transaction', 'orphan_transaction'
	Severity            string       `json:"severity" gorm:"not null;default:'warning'"` // 'warning', 'error', 'critical'
	ExpectedBalance     float64      `json:"expected_balance" gorm:"type:decimal(14,4)"`
	ActualBalance       float64      `json:"actual_balance" gorm:"type:decimal(14,4)"`
	DeltaAmount         float64      `json:"delta_amount" gorm:"type:decimal(14,4)"`
	TransactionID       *uuid.UUID   `json:"transaction_id,omitempty"`
	StripePaymentID     *string      `json:"stripe_payment_id,omitempty"`
	Description         string       `json:"description" gorm:"not null"`
	Fixed               bool         `json:"fixed" gorm:"default:false"`
	FixedAt             *time.Time   `json:"fixed_at,omitempty"`
	FixedBy             *uuid.UUID   `json:"fixed_by,omitempty"`
	FixMethod           *string      `json:"fix_method,omitempty"` // 'auto_adjust', 'manual_review', 'create_missing_tx'
	Metadata            []byte       `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt           time.Time    `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the database table name
func (ReconciliationDiscrepancy) TableName() string {
	return "wallet_reconciliation_discrepancies"
}

// ReconciliationConfig holds configuration for the reconciliation scheduler
type ReconciliationConfig struct {
	// Cron expression for scheduled reconciliation (default: "0 3 * * *" - 3 AM daily)
	CronExpression string

	// Enabled controls whether reconciliation is active
	Enabled bool

	// AutoFix controls whether to automatically fix minor discrepancies
	AutoFix bool

	// MaxAutoFixAmount is the maximum delta that can be auto-fixed (default: $0.01)
	MaxAutoFixAmount float64

	// AlertThreshold is the minimum number of discrepancies to trigger an alert
	AlertThreshold int

	// Types to reconcile (default: all)
	Types []ReconciliationType
}

// DefaultReconciliationConfig returns default configuration
func DefaultReconciliationConfig() *ReconciliationConfig {
	return &ReconciliationConfig{
		CronExpression:   "0 3 * * *", // 3 AM daily
		Enabled:          true,
		AutoFix:          false,
		MaxAutoFixAmount: 0.01,
		AlertThreshold:   1,
		Types: []ReconciliationType{
			ReconciliationTypeLedgerToBalance,
			ReconciliationTypeStripeToInternal,
			ReconciliationTypeCrossTable,
		},
	}
}

// ReconciliationService performs wallet reconciliation operations
type ReconciliationService struct {
	repo       *Repository
	stripeRepo *StripeReconciliationRepository // Would interface with Stripe
	logger     *logrus.Logger
	config     *ReconciliationConfig
}

// StripeReconciliationRepository interfaces with Stripe for reconciliation
type StripeReconciliationRepository struct {
	// This would contain methods to query Stripe for payment intents,
	// charges, and other financial data for reconciliation
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService(repo *Repository, config *ReconciliationConfig) *ReconciliationService {
	if config == nil {
		config = DefaultReconciliationConfig()
	}
	return &ReconciliationService{
		repo:   repo,
		logger: logrus.New(),
		config: config,
	}
}

// SetLogger sets the logger for the service
func (s *ReconciliationService) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

// RunFullReconciliation performs a full reconciliation of all wallet types
func (s *ReconciliationService) RunFullReconciliation(ctx context.Context, triggeredBy string, userID *uuid.UUID) (*ReconciliationRun, error) {
	run := &ReconciliationRun{
		Type:              ReconciliationTypeFull,
		Status:            ReconciliationStatusRunning,
		TriggeredBy:       triggeredBy,
		TriggeredByUserID: userID,
	}

	// Create run record
	if err := s.repo.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, fmt.Errorf("failed to create reconciliation run: %w", err)
	}

	s.logger.WithField("run_id", run.ID).Info("Starting full wallet reconciliation")

	// Run each reconciliation type
	results := make(map[string]interface{})

	for _, recType := range s.config.Types {
		s.logger.WithField("type", recType).Info("Running reconciliation type")

		switch recType {
		case ReconciliationTypeLedgerToBalance:
			ledgerResult, err := s.reconcileLedgerToBalance(ctx, run.ID)
			if err != nil {
				s.logger.WithError(err).Error("Ledger to balance reconciliation failed")
				results[string(recType)] = map[string]interface{}{"error": err.Error()}
			} else {
				results[string(recType)] = ledgerResult
				run.WalletsChecked += ledgerResult.WalletsChecked
				run.WalletsWithIssues += ledgerResult.WalletsWithIssues
				run.DiscrepanciesFound += ledgerResult.DiscrepanciesFound
				run.TotalAmountDelta += ledgerResult.TotalDelta
			}

		case ReconciliationTypeCrossTable:
			crossResult, err := s.reconcileCrossTable(ctx, run.ID)
			if err != nil {
				s.logger.WithError(err).Error("Cross-table reconciliation failed")
				results[string(recType)] = map[string]interface{}{"error": err.Error()}
			} else {
				results[string(recType)] = crossResult
			}

		case ReconciliationTypeStripeToInternal:
			// This would require Stripe API integration
			s.logger.Info("Stripe reconciliation skipped - requires Stripe API integration")
			results[string(recType)] = map[string]interface{}{"status": "skipped", "reason": "requires_stripe_integration"}
		}
	}

	// Update run with results
	details, _ := json.Marshal(results)
	run.Details = details

	now := time.Now()
	run.CompletedAt = &now

	if run.DiscrepanciesFound > 0 {
		run.Status = ReconciliationStatusPartial
	} else {
		run.Status = ReconciliationStatusCompleted
	}

	if err := s.repo.db.WithContext(ctx).Save(run).Error; err != nil {
		s.logger.WithError(err).Error("Failed to update reconciliation run")
	}

	s.logger.WithFields(logrus.Fields{
		"run_id":              run.ID,
		"wallets_checked":     run.WalletsChecked,
		"discrepancies_found": run.DiscrepanciesFound,
		"status":              run.Status,
	}).Info("Reconciliation completed")

	return run, nil
}

// ReconcileLedgerToBalanceResult holds the result of ledger-to-balance reconciliation
type ReconcileLedgerToBalanceResult struct {
	WalletsChecked     int
	WalletsWithIssues  int
	DiscrepanciesFound int
	DiscrepanciesFixed int
	TotalDelta         float64
}

// reconcileLedgerToBalance reconciles wallet balances against the transaction ledger
func (s *ReconciliationService) reconcileLedgerToBalance(ctx context.Context, runID uuid.UUID) (*ReconcileLedgerToBalanceResult, error) {
	result := &ReconcileLedgerToBalanceResult{}

	// Get all active wallets
	var wallets []Wallet
	if err := s.repo.db.WithContext(ctx).Where("status = ?", WalletStatusActive).Find(&wallets).Error; err != nil {
		return result, fmt.Errorf("failed to list wallets: %w", err)
	}

	result.WalletsChecked = len(wallets)

	for _, wallet := range wallets {
		// Calculate expected balance from transactions
		var ledgerBalance struct {
			Credits float64
			Debits  float64
		}

		err := s.repo.db.WithContext(ctx).Model(&WalletTransaction{}).
			Select(`
				COALESCE(SUM(CASE WHEN transaction_type = 'credit' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as credits,
				COALESCE(SUM(CASE WHEN transaction_type IN ('debit', 'fee_payment', 'execution_charge', 'commission') AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as debits
			`).
			Where("wallet_id = ?", wallet.ID).
			Scan(&ledgerBalance).Error

		if err != nil {
			s.logger.WithError(err).WithField("wallet_id", wallet.ID).Error("Failed to calculate ledger balance")
			continue
		}

		expectedBalance := ledgerBalance.Credits - ledgerBalance.Debits
		delta := expectedBalance - wallet.BalanceUSD

		// Check for discrepancy (allowing for small floating point differences)
		if abs(delta) > 0.001 {
			result.WalletsWithIssues++
			result.TotalDelta += abs(delta)

			discrepancy := &ReconciliationDiscrepancy{
				RunID:           runID,
				WalletID:        wallet.ID,
				DiscrepancyType: "balance_mismatch",
				Severity:        s.determineSeverity(abs(delta)),
				ExpectedBalance: expectedBalance,
				ActualBalance:   wallet.BalanceUSD,
				DeltaAmount:     delta,
				Description:     fmt.Sprintf("Ledger balance ($%.4f) doesn't match wallet balance ($%.4f)", expectedBalance, wallet.BalanceUSD),
			}

			if err := s.repo.db.WithContext(ctx).Create(discrepancy).Error; err != nil {
				s.logger.WithError(err).Error("Failed to create discrepancy record")
			} else {
				result.DiscrepanciesFound++
			}

			// Auto-fix if enabled and delta is small
			if s.config.AutoFix && abs(delta) <= s.config.MaxAutoFixAmount {
				if err := s.autoFixBalance(ctx, wallet.ID, expectedBalance, discrepancy.ID); err != nil {
					s.logger.WithError(err).Error("Auto-fix failed")
				} else {
					result.DiscrepanciesFixed++
				}
			}
		}
	}

	return result, nil
}

// CrossTableReconcileResult holds cross-table reconciliation results
type CrossTableReconcileResult struct {
	Status                  string                  `json:"status"`
	Message                 string                  `json:"message"`
	WalletsChecked          int                     `json:"wallets_checked"`
	WalletsWithIssues       int                     `json:"wallets_with_issues"`
	DiscrepanciesFound      int                     `json:"discrepancies_found"`
	DiscrepanciesFixed      int                     `json:"discrepancies_fixed"`
	LegacyRecordsChecked    int                     `json:"legacy_records_checked"`
	OrphanedWallets         int                     `json:"orphaned_wallets"`
	OrphanedBillingControls int                     `json:"orphaned_billing_controls"`
	TotalAmountDelta        float64                 `json:"total_amount_delta"`
	Discrepancies           []CrossTableDiscrepancy `json:"discrepancies,omitempty"`
}

// CrossTableDiscrepancy represents a specific cross-table inconsistency
type CrossTableDiscrepancy struct {
	Type        string    `json:"type"`
	WalletID    uuid.UUID `json:"wallet_id"`
	LegacyID    uuid.UUID `json:"legacy_id,omitempty"`
	Field       string    `json:"field"`
	Expected    float64   `json:"expected"`
	Actual      float64   `json:"actual"`
	Delta       float64   `json:"delta"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// LegacyBillingControls represents the legacy agent_billing_controls table structure
type LegacyBillingControls struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	AgentID             uuid.UUID  `json:"agent_id" gorm:"type:uuid;not null;index"`
	CreditBalanceUSD    float64    `json:"credit_balance_usd" gorm:"type:decimal(14,4);default:0"`
	TotalCreditsUSD     float64    `json:"total_credits_usd" gorm:"type:decimal(14,4);default:0"`
	TotalDebitsUSD      float64    `json:"total_debits_usd" gorm:"type:decimal(14,4);default:0"`
	LifetimeEarningsUSD float64    `json:"lifetime_earnings_usd" gorm:"type:decimal(14,4);default:0"`
	LifetimeSpentUSD    float64    `json:"lifetime_spent_usd" gorm:"type:decimal(14,4);default:0"`
	SpendCapDailyUSD    float64    `json:"spend_cap_daily_usd" gorm:"type:decimal(14,4);default:0"`
	SpendCapMonthlyUSD  float64    `json:"spend_cap_monthly_usd" gorm:"type:decimal(14,4);default:0"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// TableName returns the legacy table name
func (LegacyBillingControls) TableName() string {
	return "agent_billing_controls"
}

// reconcileCrossTable reconciles between wallets and legacy agent_billing_controls
func (s *ReconciliationService) reconcileCrossTable(ctx context.Context, runID uuid.UUID) (*CrossTableReconcileResult, error) {
	result := &CrossTableReconcileResult{
		Status:        "running",
		Discrepancies: []CrossTableDiscrepancy{},
	}

	s.logger.WithField("run_id", runID).Info("Starting cross-table reconciliation")

	// Check if legacy table exists
	var legacyTableExists bool
	err := s.repo.db.WithContext(ctx).Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'agent_billing_controls')",
	).Scan(&legacyTableExists).Error

	if err != nil || !legacyTableExists {
		result.Status = "skipped"
		result.Message = "Legacy agent_billing_controls table does not exist, nothing to reconcile"
		s.logger.Info("Legacy table not found, skipping cross-table reconciliation")
		return result, nil
	}

	// Get all wallets with their owner mapping
	type WalletOwner struct {
		WalletID uuid.UUID
		UserID   uuid.UUID
		AgentID  *uuid.UUID
	}

	var walletOwners []WalletOwner
	err = s.repo.db.WithContext(ctx).Model(&Wallet{}).
		Select("wallets.id as wallet_id, wallets.user_id, wallets.agent_id").
		Where("wallets.status = ?", WalletStatusActive).
		Scan(&walletOwners).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list wallet owners: %w", err)
	}

	result.WalletsChecked = len(walletOwners)

	// Build maps for quick lookup
	walletByUser := make(map[uuid.UUID]uuid.UUID)
	walletByAgent := make(map[uuid.UUID]uuid.UUID)

	for _, wo := range walletOwners {
		walletByUser[wo.UserID] = wo.WalletID
		if wo.AgentID != nil {
			walletByAgent[*wo.AgentID] = wo.WalletID
		}
	}

	// Get all legacy billing controls
	var legacyControls []LegacyBillingControls
	err = s.repo.db.WithContext(ctx).Find(&legacyControls).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list legacy billing controls: %w", err)
	}

	result.LegacyRecordsChecked = len(legacyControls)

	// Check each legacy record against wallet
	for _, legacy := range legacyControls {
		// Try to find matching wallet by agent_id or user lookup
		var wallet *Wallet

		if legacy.AgentID != uuid.Nil {
			if walletID, ok := walletByAgent[legacy.AgentID]; ok {
				err = s.repo.db.WithContext(ctx).First(&wallet, "id = ?", walletID).Error
			}
		}

		// If no wallet found, check if there's a wallet directly linked to the agent's user
		if wallet == nil {
			// Try to find agent's user and their wallet
			type AgentUser struct {
				UserID uuid.UUID
			}
			var agentUser AgentUser
			err := s.repo.db.WithContext(ctx).Raw(
				"SELECT user_id FROM agents WHERE id = ?",
				legacy.AgentID,
			).Scan(&agentUser).Error

			if err == nil && agentUser.UserID != uuid.Nil {
				if walletID, ok := walletByUser[agentUser.UserID]; ok {
					err = s.repo.db.WithContext(ctx).First(&wallet, "id = ?", walletID).Error
				}
			}
		}

		if wallet == nil {
			// Orphaned legacy record - no matching wallet
			result.OrphanedBillingControls++
			disc := CrossTableDiscrepancy{
				Type:        "orphaned_legacy_record",
				LegacyID:    legacy.ID,
				Field:       "wallet_mapping",
				Severity:    "warning",
				Description: fmt.Sprintf("Legacy billing control for agent %s has no matching wallet", legacy.AgentID),
			}
			result.Discrepancies = append(result.Discrepancies, disc)
			result.DiscrepanciesFound++

			// Create discrepancy record
			s.createCrossTableDiscrepancy(ctx, runID, disc)
			continue
		}

		// Compare fields
		s.compareCrossTableFields(wallet, legacy, result)
	}

	// Check for orphaned wallets (wallets with no legacy record)
	type AgentIDResult struct {
		AgentID uuid.UUID
	}
	var agentIDs []AgentIDResult
	err = s.repo.db.WithContext(ctx).Raw("SELECT id as agent_id FROM agents").Scan(&agentIDs).Error
	if err == nil {
		for _, agent := range agentIDs {
			if _, ok := walletByAgent[agent.AgentID]; ok {
				// Check if legacy record exists
				var count int64
				s.repo.db.WithContext(ctx).Model(&LegacyBillingControls{}).
					Where("agent_id = ?", agent.AgentID).
					Count(&count)

				if count == 0 {
					result.OrphanedWallets++
				}
			}
		}
	}

	// Update summary
	if len(result.Discrepancies) > 0 {
		result.Status = "completed_with_discrepancies"
		result.Message = fmt.Sprintf("Found %d discrepancies between wallets and legacy billing controls", len(result.Discrepancies))
	} else {
		result.Status = "completed"
		result.Message = "All wallets match legacy billing controls"
	}

	s.logger.WithFields(logrus.Fields{
		"wallets_checked":           result.WalletsChecked,
		"legacy_records_checked":    result.LegacyRecordsChecked,
		"discrepancies_found":       result.DiscrepanciesFound,
		"orphaned_wallets":          result.OrphanedWallets,
		"orphaned_billing_controls": result.OrphanedBillingControls,
	}).Info("Cross-table reconciliation completed")

	return result, nil
}

// compareCrossTableFields compares wallet fields with legacy billing controls
func (s *ReconciliationService) compareCrossTableFields(
	wallet *Wallet,
	legacy LegacyBillingControls,
	result *CrossTableReconcileResult,
) {
	fieldComparisons := []struct {
		field      string
		walletVal  float64
		legacyVal  float64
		tolerance  float64
	}{
		{"balance_usd", wallet.BalanceUSD, legacy.CreditBalanceUSD, 0.01},
		{"lifetime_earnings_usd", wallet.LifetimeEarningsUSD, legacy.LifetimeEarningsUSD, 0.01},
		{"lifetime_spent_usd", wallet.LifetimeSpentUSD, legacy.LifetimeSpentUSD, 0.01},
		{"spend_cap_daily_usd", getFloatPtrValue(wallet.SpendCapDailyUSD), legacy.SpendCapDailyUSD, 0.01},
		{"spend_cap_monthly_usd", getFloatPtrValue(wallet.SpendCapMonthlyUSD), legacy.SpendCapMonthlyUSD, 0.01},
	}

	hasDiscrepancy := false
	for _, cmp := range fieldComparisons {
		delta := cmp.walletVal - cmp.legacyVal
		if abs(delta) > cmp.tolerance {
			hasDiscrepancy = true
			result.TotalAmountDelta += abs(delta)

			severity := "warning"
			if abs(delta) >= 100 {
				severity = "error"
			}
			if abs(delta) >= 1000 {
				severity = "critical"
			}

			disc := CrossTableDiscrepancy{
				Type:        "field_mismatch",
				WalletID:    wallet.ID,
				LegacyID:    legacy.ID,
				Field:       cmp.field,
				Expected:    cmp.legacyVal,
				Actual:      cmp.walletVal,
				Delta:       delta,
				Severity:    severity,
				Description: fmt.Sprintf("Field %s mismatch: wallet=$%.4f, legacy=$%.4f, delta=$%.4f", cmp.field, cmp.walletVal, cmp.legacyVal, delta),
			}

			result.Discrepancies = append(result.Discrepancies, disc)
			result.DiscrepanciesFound++
		}
	}

	if hasDiscrepancy {
		result.WalletsWithIssues++
	}
}

// createCrossTableDiscrepancy persists a cross-table discrepancy to the database
func (s *ReconciliationService) createCrossTableDiscrepancy(
	ctx context.Context,
	runID uuid.UUID,
	disc CrossTableDiscrepancy,
) error {
	// Store as a general discrepancy record with metadata
	metadata, _ := json.Marshal(map[string]interface{}{
		"cross_table_type": disc.Type,
		"legacy_id":        disc.LegacyID,
		"field":            disc.Field,
	})

	dbDisc := &ReconciliationDiscrepancy{
		RunID:           runID,
		WalletID:        disc.WalletID,
		DiscrepancyType: "cross_table_" + disc.Type,
		Severity:        disc.Severity,
		ExpectedBalance: disc.Expected,
		ActualBalance:   disc.Actual,
		DeltaAmount:     disc.Delta,
		Description:     disc.Description,
		Metadata:        metadata,
	}

	if err := s.repo.db.WithContext(ctx).Create(dbDisc).Error; err != nil {
		s.logger.WithError(err).Error("Failed to create cross-table discrepancy record")
		return err
	}

	return nil
}

// autoFixBalance automatically adjusts a wallet balance to match the ledger
func (s *ReconciliationService) autoFixBalance(ctx context.Context, walletID uuid.UUID, correctBalance float64, discrepancyID uuid.UUID) error {
	return s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update wallet balance
		if err := tx.Model(&Wallet{}).
			Where("id = ?", walletID).
			Update("balance_usd", correctBalance).Error; err != nil {
			return err
		}

		// Mark discrepancy as fixed
		now := time.Now()
		fixMethod := "auto_adjust"
		if err := tx.Model(&ReconciliationDiscrepancy{}).
			Where("id = ?", discrepancyID).
			Updates(map[string]interface{}{
				"fixed":      true,
				"fixed_at":   now,
				"fix_method": fixMethod,
			}).Error; err != nil {
			return err
		}

		s.logger.WithFields(logrus.Fields{
			"wallet_id":       walletID,
			"correct_balance": correctBalance,
		}).Info("Auto-fixed wallet balance")

		return nil
	})
}

// determineSeverity determines the severity of a discrepancy based on amount
func (s *ReconciliationService) determineSeverity(delta float64) string {
	switch {
	case delta >= 1000:
		return "critical"
	case delta >= 100:
		return "error"
	default:
		return "warning"
	}
}

// GetReconciliationRun retrieves a reconciliation run by ID
func (s *ReconciliationService) GetReconciliationRun(ctx context.Context, runID uuid.UUID) (*ReconciliationRun, error) {
	var run ReconciliationRun
	if err := s.repo.db.WithContext(ctx).First(&run, "id = ?", runID).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

// ListReconciliationRuns lists reconciliation runs with pagination
func (s *ReconciliationService) ListReconciliationRuns(ctx context.Context, limit, offset int) ([]ReconciliationRun, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	if err := s.repo.db.WithContext(ctx).Model(&ReconciliationRun{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var runs []ReconciliationRun
	if err := s.repo.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&runs).Error; err != nil {
		return nil, 0, err
	}

	return runs, total, nil
}

// GetDiscrepanciesForRun retrieves discrepancies for a reconciliation run
func (s *ReconciliationService) GetDiscrepanciesForRun(ctx context.Context, runID uuid.UUID) ([]ReconciliationDiscrepancy, error) {
	var discrepancies []ReconciliationDiscrepancy
	if err := s.repo.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("severity DESC, created_at DESC").
		Find(&discrepancies).Error; err != nil {
		return nil, err
	}
	return discrepancies, nil
}

// getFloatPtrValue safely dereferences a float64 pointer, returning 0 if nil
func getFloatPtrValue(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
