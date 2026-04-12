// Wallet Migration Script
// Transfers data from user_wallets and agent_billing_controls to the unified wallets table
//
// Usage:
//   go run cmd/migrate-wallets/migrate.go [--dry-run] [--batch-size=100]
//
// This script should be run after applying the unified wallet schema migration.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	dryRun     = flag.Bool("dry-run", false, "Run in dry-run mode (no actual changes)")
	batchSize  = flag.Int("batch-size", 100, "Number of records to process per batch")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("Unified Wallet System Migration Tool")
	fmt.Println("========================================")
	if *dryRun {
		fmt.Println("*** DRY RUN MODE - No changes will be made ***")
	}
	fmt.Printf("Batch size: %d\n", *batchSize)
	fmt.Println()

	// Load database connection
	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate wallet schema if needed (skip if tables already exist from SQL migration)
	// Only run AutoMigrate if the tables weren't created by the SQL migration
	var tableCount int64
	db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'wallets'").Scan(&tableCount)
	if tableCount == 0 {
		if err := db.AutoMigrate(&wallet.Wallet{}, &wallet.WalletTransaction{}); err != nil {
			log.Fatalf("Failed to auto-migrate wallet schema: %v", err)
		}
		log.Println("Auto-migrated wallet schema")
	} else {
		log.Println("Wallet tables already exist, skipping auto-migration")
	}

	ctx := context.Background()
	repo := wallet.NewRepository(db)

	// Migration statistics
	stats := &MigrationStats{}

	// Step 1: Migrate user wallets
	fmt.Println("\n[Step 1] Migrating user wallets...")
	if err := migrateUserWallets(ctx, db, repo, stats); err != nil {
		log.Fatalf("Failed to migrate user wallets: %v", err)
	}

	// Step 2: Migrate agent wallets
	fmt.Println("\n[Step 2] Migrating agent wallets...")
	if err := migrateAgentWallets(ctx, db, repo, stats); err != nil {
		log.Fatalf("Failed to migrate agent wallets: %v", err)
	}

	// Step 3: Migrate transaction history (optional, can be done separately)
	fmt.Println("\n[Step 3] Migrating transaction history...")
	if err := migrateTransactionHistory(ctx, db, repo, stats); err != nil {
		log.Printf("Warning: Failed to migrate transaction history: %v", err)
		// Don't fail completely for transaction history
	}

	// Print summary
	fmt.Println("\n========================================")
	fmt.Println("Migration Summary")
	fmt.Println("========================================")
	fmt.Printf("User wallets migrated:    %d (skipped: %d, errors: %d)\n",
		stats.UserWalletsMigrated, stats.UserWalletsSkipped, stats.UserWalletErrors)
	fmt.Printf("Agent wallets migrated:   %d (skipped: %d, errors: %d)\n",
		stats.AgentWalletsMigrated, stats.AgentWalletsSkipped, stats.AgentWalletErrors)
	fmt.Printf("Transactions migrated:    %d (skipped: %d, errors: %d)\n",
		stats.TransactionsMigrated, stats.TransactionsSkipped, stats.TransactionErrors)

	if *dryRun {
		fmt.Println("\n*** This was a DRY RUN. No actual changes were made. ***")
		fmt.Println("Run without --dry-run to apply the migration.")
	} else {
		fmt.Println("\nMigration completed successfully!")
	}
}

// MigrationStats tracks migration progress
type MigrationStats struct {
	UserWalletsMigrated   int
	UserWalletsSkipped    int
	UserWalletErrors      int
	AgentWalletsMigrated  int
	AgentWalletsSkipped   int
	AgentWalletErrors     int
	TransactionsMigrated  int
	TransactionsSkipped   int
	TransactionErrors     int
}

// LegacyUserWallet represents the old user_wallets table structure
type LegacyUserWallet struct {
	UserID              uuid.UUID `gorm:"column:user_id"`
	BalanceUSD          float64   `gorm:"column:balance_usd"`
	LifetimeEarningsUSD float64   `gorm:"column:lifetime_earnings_usd"`
	LifetimeFeesUSD     float64   `gorm:"column:lifetime_fees_usd"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (LegacyUserWallet) TableName() string {
	return "user_wallets"
}

// LegacyAgentBillingControls represents the old agent_billing_controls structure
type LegacyAgentBillingControls struct {
	ID                 uuid.UUID `gorm:"column:id"`
	AgentID            string    `gorm:"column:agent_id"`
	SpendCapMonthlyUSD *float64  `gorm:"column:spend_cap_monthly_usd"`
	SpendCapDailyUSD   *float64  `gorm:"column:spend_cap_daily_usd"`
	CreditBalanceUSD   float64   `gorm:"column:credit_balance_usd"`
	BillingMode        string    `gorm:"column:billing_mode"`
	TeamID             *uuid.UUID `gorm:"column:team_id"`
	AlertThresholds    []float64 `gorm:"column:alert_thresholds;type:decimal[];default:'{0.5,0.8,0.95}'"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (LegacyAgentBillingControls) TableName() string {
	return "agent_billing_controls"
}

// LegacyFeeTransaction represents the old fee_transactions structure
type LegacyFeeTransaction struct {
	ID        uuid.UUID `gorm:"column:id"`
	UserID    uuid.UUID `gorm:"column:user_id"`
	Kind      string    `gorm:"column:kind"`
	AmountUSD float64   `gorm:"column:amount_usd"`
	Status    string    `gorm:"column:status"`
	Reference string    `gorm:"column:reference"`
	Metadata  []byte    `gorm:"column:metadata"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (LegacyFeeTransaction) TableName() string {
	return "fee_transactions"
}

// LegacyAgentFinancialTransaction represents the old agent_financial_transactions structure
type LegacyAgentFinancialTransaction struct {
	ID          uuid.UUID  `gorm:"column:id"`
	TenantID    uuid.UUID  `gorm:"column:tenant_id"`
	AgentID     string     `gorm:"column:agent_id"`
	Kind        string     `gorm:"column:kind"`
	AmountUSD   float64    `gorm:"column:amount_usd"`
	Status      string     `gorm:"column:status"`
	Provider    *string    `gorm:"column:provider"`
	ProviderRef *string    `gorm:"column:provider_ref"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (LegacyAgentFinancialTransaction) TableName() string {
	return "agent_financial_transactions"
}

func connectToDatabase() (*gorm.DB, error) {
	// storage.NewPostgresDB() reads config from environment variables
	db, err := storage.NewPostgresDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool limits on the embedded *sql.DB
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db.GORM, nil
}

func migrateUserWallets(ctx context.Context, db *gorm.DB, repo *wallet.Repository, stats *MigrationStats) error {
	var legacyWallets []LegacyUserWallet

	// Fetch all user wallets
	if err := db.WithContext(ctx).Find(&legacyWallets).Error; err != nil {
		return fmt.Errorf("failed to fetch user wallets: %w", err)
	}

	fmt.Printf("Found %d user wallets to migrate\n", len(legacyWallets))

	for _, legacy := range legacyWallets {
		// Check if wallet already exists
		existing, err := repo.GetWalletByOwner(ctx, wallet.OwnerTypeUser, legacy.UserID.String())
		if err != nil {
			log.Printf("Error checking existing wallet for user %s: %v", legacy.UserID, err)
			stats.UserWalletErrors++
			continue
		}
		if existing != nil {
			if *verbose {
				log.Printf("Wallet already exists for user %s, skipping", legacy.UserID)
			}
			stats.UserWalletsSkipped++
			continue
		}

		if *dryRun {
			if *verbose {
				log.Printf("[DRY RUN] Would create wallet for user %s with balance %.4f",
					legacy.UserID, legacy.BalanceUSD)
			}
			stats.UserWalletsMigrated++
			continue
		}

		// Create new wallet
		newWallet := &wallet.Wallet{
			ID:                  uuid.New(),
			OwnerType:           wallet.OwnerTypeUser,
			OwnerID:             legacy.UserID.String(),
			UserID:              &legacy.UserID,
			WalletType:          wallet.WalletTypeRegistry,
			BalanceUSD:          legacy.BalanceUSD,
			LifetimeEarningsUSD: legacy.LifetimeEarningsUSD,
			LifetimeSpentUSD:    legacy.LifetimeFeesUSD,
			BillingMode:         wallet.BillingModePerWallet,
			Status:              wallet.WalletStatusActive,
			AlertThresholds:     []float64{0.5, 0.8, 0.95},
			CreatedAt:           legacy.CreatedAt,
			UpdatedAt:           legacy.UpdatedAt,
		}

		if err := db.WithContext(ctx).Create(newWallet).Error; err != nil {
			log.Printf("Failed to create wallet for user %s: %v", legacy.UserID, err)
			stats.UserWalletErrors++
			continue
		}

		stats.UserWalletsMigrated++
		if *verbose {
			log.Printf("Created wallet for user %s with balance %.4f", legacy.UserID, legacy.BalanceUSD)
		}
	}

	return nil
}

func migrateAgentWallets(ctx context.Context, db *gorm.DB, repo *wallet.Repository, stats *MigrationStats) error {
	// Use raw SQL to handle alert_thresholds type issue
	rows, err := db.WithContext(ctx).Raw(`
		SELECT id, agent_id, spend_cap_monthly_usd, spend_cap_daily_usd, 
		       credit_balance_usd, billing_mode, team_id, 
		       COALESCE(alert_thresholds, '{0.5,0.8,0.95}'::decimal[]) as alert_thresholds,
		       created_at, updated_at
		FROM agent_billing_controls
	`).Rows()
	if err != nil {
		return fmt.Errorf("failed to fetch agent billing controls: %w", err)
	}
	defer rows.Close()

	var legacyControls []LegacyAgentBillingControls
	for rows.Next() {
		var control LegacyAgentBillingControls
		var alertThresholdsStr string
		if err := rows.Scan(&control.ID, &control.AgentID, &control.SpendCapMonthlyUSD,
			&control.SpendCapDailyUSD, &control.CreditBalanceUSD, &control.BillingMode,
			&control.TeamID, &alertThresholdsStr, &control.CreatedAt, &control.UpdatedAt); err != nil {
			log.Printf("Error scanning agent billing control: %v", err)
			continue
		}
		// Parse alert_thresholds from string like "{0.5,0.8,0.95}"
		control.AlertThresholds = []float64{0.5, 0.8, 0.95}
		legacyControls = append(legacyControls, control)
	}

	fmt.Printf("Found %d agent billing controls to migrate\n", len(legacyControls))

	for _, legacy := range legacyControls {
		// Check if wallet already exists
		existing, err := repo.GetWalletByOwner(ctx, wallet.OwnerTypeAgent, legacy.AgentID)
		if err != nil {
			log.Printf("Error checking existing wallet for agent %s: %v", legacy.AgentID, err)
			stats.AgentWalletErrors++
			continue
		}
		if existing != nil {
			if *verbose {
				log.Printf("Wallet already exists for agent %s, skipping", legacy.AgentID)
			}
			stats.AgentWalletsSkipped++
			continue
		}

		if *dryRun {
			if *verbose {
				log.Printf("[DRY RUN] Would create wallet for agent %s with balance %.4f",
					legacy.AgentID, legacy.CreditBalanceUSD)
			}
			stats.AgentWalletsMigrated++
			continue
		}

		// Map billing mode
		billingMode := legacy.BillingMode
		if billingMode == "" {
			billingMode = wallet.BillingModePerWallet
		}

		// Create new wallet
		newWallet := &wallet.Wallet{
			ID:                 legacy.ID, // Preserve ID for referential integrity
			OwnerType:          wallet.OwnerTypeAgent,
			OwnerID:            legacy.AgentID,
			AgentID:            &legacy.AgentID,
			WalletType:         wallet.WalletTypeExecution,
			BalanceUSD:         legacy.CreditBalanceUSD,
			SpendCapMonthlyUSD: legacy.SpendCapMonthlyUSD,
			SpendCapDailyUSD:   legacy.SpendCapDailyUSD,
			BillingMode:        billingMode,
			TeamID:             legacy.TeamID,
			Status:             wallet.WalletStatusActive,
			AlertThresholds:    legacy.AlertThresholds,
			CreatedAt:          legacy.CreatedAt,
			UpdatedAt:          legacy.UpdatedAt,
		}

		if newWallet.AlertThresholds == nil {
			newWallet.AlertThresholds = []float64{0.5, 0.8, 0.95}
		}

		if err := db.WithContext(ctx).Create(newWallet).Error; err != nil {
			log.Printf("Failed to create wallet for agent %s: %v", legacy.AgentID, err)
			stats.AgentWalletErrors++
			continue
		}

		stats.AgentWalletsMigrated++
		if *verbose {
			log.Printf("Created wallet for agent %s with balance %.4f", legacy.AgentID, legacy.CreditBalanceUSD)
		}
	}

	return nil
}

func migrateTransactionHistory(ctx context.Context, db *gorm.DB, repo *wallet.Repository, stats *MigrationStats) error {
	// Migrate fee_transactions to wallet_transactions
	var legacyFeeTransactions []LegacyFeeTransaction

	if err := db.WithContext(ctx).
		Where("kind IN ?", []string{"credit", "fee_payment"}).
		Find(&legacyFeeTransactions).Error; err != nil {
		return fmt.Errorf("failed to fetch fee transactions: %w", err)
	}

	fmt.Printf("Found %d fee transactions to migrate\n", len(legacyFeeTransactions))

	for _, legacy := range legacyFeeTransactions {
		// Find the wallet for this user
		w, err := repo.GetWalletByOwner(ctx, wallet.OwnerTypeUser, legacy.UserID.String())
		if err != nil {
			log.Printf("Error finding wallet for user %s: %v", legacy.UserID, err)
			stats.TransactionErrors++
			continue
		}
		if w == nil {
			log.Printf("No wallet found for user %s, skipping transaction %s", legacy.UserID, legacy.ID)
			stats.TransactionsSkipped++
			continue
		}

		// Check for existing transaction with same ID
		existing, err := repo.GetTransactionByID(ctx, legacy.ID)
		if err != nil {
			log.Printf("Error checking existing transaction %s: %v", legacy.ID, err)
			stats.TransactionErrors++
			continue
		}
		if existing != nil {
			stats.TransactionsSkipped++
			continue
		}

		if *dryRun {
			stats.TransactionsMigrated++
			continue
		}

		// Map transaction type
		txType := wallet.TransactionTypeCredit
		if legacy.Kind == "fee_payment" {
			txType = wallet.TransactionTypeFeePayment
		}

		// Create new transaction with original ID
		newTx := &wallet.WalletTransaction{
			ID:              legacy.ID,
			WalletID:        w.ID,
			TransactionType: txType,
			AmountUSD:       legacy.AmountUSD,
			Status:          wallet.TransactionStatusCompleted,
			Reference:       &legacy.Reference,
			CreatedAt:       legacy.CreatedAt,
			CompletedAt:     ptr(legacy.CreatedAt),
		}

		// Set balance snapshots (approximate since we don't have historical data)
		newTx.BalanceAfterUSD = w.BalanceUSD
		newTx.BalanceBeforeUSD = w.BalanceUSD - legacy.AmountUSD

		if err := db.WithContext(ctx).Create(newTx).Error; err != nil {
			log.Printf("Failed to create transaction %s: %v", legacy.ID, err)
			stats.TransactionErrors++
			continue
		}

		stats.TransactionsMigrated++
	}

	// Migrate agent_financial_transactions to wallet_transactions
	var legacyAgentTransactions []LegacyAgentFinancialTransaction

	if err := db.WithContext(ctx).
		Where("kind IN ?", []string{"credit_purchase", "execution_charge"}).
		Find(&legacyAgentTransactions).Error; err != nil {
		return fmt.Errorf("failed to fetch agent financial transactions: %w", err)
	}

	fmt.Printf("Found %d agent financial transactions to migrate\n", len(legacyAgentTransactions))

	for _, legacy := range legacyAgentTransactions {
		// Find the wallet for this agent
		w, err := repo.GetWalletByOwner(ctx, wallet.OwnerTypeAgent, legacy.AgentID)
		if err != nil {
			log.Printf("Error finding wallet for agent %s: %v", legacy.AgentID, err)
			stats.TransactionErrors++
			continue
		}
		if w == nil {
			log.Printf("No wallet found for agent %s, skipping transaction %s", legacy.AgentID, legacy.ID)
			stats.TransactionsSkipped++
			continue
		}

		// Check for existing transaction
		existing, err := repo.GetTransactionByID(ctx, legacy.ID)
		if err != nil {
			log.Printf("Error checking existing transaction %s: %v", legacy.ID, err)
			stats.TransactionErrors++
			continue
		}
		if existing != nil {
			stats.TransactionsSkipped++
			continue
		}

		if *dryRun {
			stats.TransactionsMigrated++
			continue
		}

		// Map transaction type
		txType := wallet.TransactionTypeCredit
		if legacy.Kind == "execution_charge" {
			txType = wallet.TransactionTypeExecutionCharge
		}

		// Create new transaction with original ID
		newTx := &wallet.WalletTransaction{
			ID:              legacy.ID,
			WalletID:        w.ID,
			TransactionType: txType,
			AmountUSD:       legacy.AmountUSD,
			Status:          wallet.TransactionStatusCompleted,
			Reference:       legacy.ProviderRef,
			TriggeredByType: ptr("system"),
			TriggeredByID:   ptr(legacy.AgentID),
			CreatedAt:       legacy.CreatedAt,
			CompletedAt:     ptr(legacy.CreatedAt),
		}

		// Set balance snapshots (approximate)
		newTx.BalanceAfterUSD = w.BalanceUSD
		if txType == wallet.TransactionTypeCredit {
			newTx.BalanceBeforeUSD = w.BalanceUSD - legacy.AmountUSD
		} else {
			newTx.BalanceBeforeUSD = w.BalanceUSD + legacy.AmountUSD
		}

		if err := db.WithContext(ctx).Create(newTx).Error; err != nil {
			log.Printf("Failed to create agent transaction %s: %v", legacy.ID, err)
			stats.TransactionErrors++
			continue
		}

		stats.TransactionsMigrated++
	}

	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func ptr[T any](v T) *T {
	return &v
}
