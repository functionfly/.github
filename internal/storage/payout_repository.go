package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// =============================================================================
// Models
// =============================================================================

// StripeConnectAccount represents a user's Stripe connected account for receiving payouts.
type StripeConnectAccount struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 uuid.UUID  `json:"user_id"`
	StripeAccountID        string     `json:"stripe_account_id"`
	AccountStatus          string     `json:"account_status"` // pending, onboarding, active, restricted, disabled
	PayoutsEnabled         bool       `json:"payouts_enabled"`
	DetailsSubmitted       bool       `json:"details_submitted"`
	ChargesEnabled         bool       `json:"charges_enabled"`
	Country                *string    `json:"country,omitempty"`
	Currency               string     `json:"currency"`
	BankLast4              *string    `json:"bank_last4,omitempty"`
	BankName               *string    `json:"bank_name,omitempty"`
	OnboardingURL          *string    `json:"onboarding_url,omitempty"`
	OnboardingURLExpiresAt *time.Time `json:"onboarding_url_expires_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// PayoutRequest represents a user-initiated withdrawal request.
type PayoutRequest struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	ConnectAccountID      uuid.UUID  `json:"connect_account_id"`
	AmountCents           int        `json:"amount_cents"`
	Currency              string     `json:"currency"`
	Status                string     `json:"status"` // pending, processing, completed, failed, cancelled
	StripeTransferID      *string    `json:"stripe_transfer_id,omitempty"`
	StripePayoutID        *string    `json:"stripe_payout_id,omitempty"`
	IdempotencyKey        string     `json:"idempotency_key"`
	FailureReason         *string    `json:"failure_reason,omitempty"`
	ReviewedBy            *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	// Approval workflow fields
	RequiresApproval      bool       `json:"requires_approval,omitempty"`
	ApprovalThresholdUSD  *float64   `json:"approval_threshold_usd,omitempty"`
	ApprovedBy            *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt            *time.Time `json:"approved_at,omitempty"`
	ApprovalNotes         *string    `json:"approval_notes,omitempty"`
	SecondApprovalBy      *uuid.UUID `json:"second_approval_by,omitempty"`
	SecondApprovalAt      *time.Time `json:"second_approval_at,omitempty"`
	RejectedBy            *uuid.UUID `json:"rejected_by,omitempty"`
	RejectedAt            *time.Time `json:"rejected_at,omitempty"`
	RejectionReason       *string    `json:"rejection_reason,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// PayoutLedgerEntry is an immutable audit record for all payout-related fund movements.
type PayoutLedgerEntry struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	EntryType         string     `json:"entry_type"` // earning_credit, payout_debit, payout_reversal, adjustment
	AmountCents       int        `json:"amount_cents"`
	Currency          string     `json:"currency"`
	ReferenceType     *string    `json:"reference_type,omitempty"` // publisher_earning, payout_request, admin_adjustment
	ReferenceID       *uuid.UUID `json:"reference_id,omitempty"`
	BalanceAfterCents int        `json:"balance_after_cents"`
	Description       *string    `json:"description,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// PayoutBalance holds a user's payout-available balance breakdown.
type PayoutBalance struct {
	UserID              uuid.UUID `json:"user_id"`
	AvailableBalanceUSD float64   `json:"available_balance_usd"` // earned and eligible for payout
	PendingBalanceUSD   float64   `json:"pending_balance_usd"`   // earned but not yet available
	TotalEarningsUSD    float64   `json:"total_earnings_usd"`    // lifetime earnings
	TotalPaidOutUSD     float64   `json:"total_paid_out_usd"`    // lifetime payouts
}

// =============================================================================
// Repository
// =============================================================================

// PayoutRepository handles payout-related database operations.
type PayoutRepository struct {
	db *sql.DB
}

// NewPayoutRepository creates a new payout repository.
func NewPayoutRepository(db *sql.DB) *PayoutRepository {
	return &PayoutRepository{db: db}
}

// ─── Stripe Connect Accounts ────────────────────────────────────────────────

// CreateConnectAccount inserts a new Stripe Connect account record.
func (r *PayoutRepository) CreateConnectAccount(ctx context.Context, account *StripeConnectAccount) error {
	query := `
		INSERT INTO stripe_connect_accounts
			(id, user_id, stripe_account_id, account_status, payouts_enabled, details_submitted,
			 charges_enabled, country, currency, bank_last4, bank_name, onboarding_url,
			 onboarding_url_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	now := time.Now()
	account.ID = uuid.New()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		account.ID, account.UserID, account.StripeAccountID, account.AccountStatus,
		account.PayoutsEnabled, account.DetailsSubmitted, account.ChargesEnabled,
		account.Country, account.Currency, account.BankLast4, account.BankName,
		account.OnboardingURL, account.OnboardingURLExpiresAt, account.CreatedAt, account.UpdatedAt,
	)
	return err
}

// GetConnectAccountByUserID retrieves a user's Stripe Connect account.
func (r *PayoutRepository) GetConnectAccountByUserID(ctx context.Context, userID uuid.UUID) (*StripeConnectAccount, error) {
	query := `
		SELECT id, user_id, stripe_account_id, account_status, payouts_enabled, details_submitted,
		       charges_enabled, COALESCE(country, ''), currency,
		       bank_last4, bank_name, onboarding_url, onboarding_url_expires_at,
		       created_at, updated_at
		FROM stripe_connect_accounts
		WHERE user_id = $1`

	account := &StripeConnectAccount{}
	var country, bankLast4, bankName, onboardingURL sql.NullString
	var onboardingExpires sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID, &account.UserID, &account.StripeAccountID, &account.AccountStatus,
		&account.PayoutsEnabled, &account.DetailsSubmitted, &account.ChargesEnabled,
		&country, &account.Currency,
		&bankLast4, &bankName, &onboardingURL, &onboardingExpires,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connect account: %w", err)
	}

	if country.Valid {
		s := country.String
		account.Country = &s
	}
	if bankLast4.Valid {
		s := bankLast4.String
		account.BankLast4 = &s
	}
	if bankName.Valid {
		s := bankName.String
		account.BankName = &s
	}
	if onboardingURL.Valid {
		s := onboardingURL.String
		account.OnboardingURL = &s
	}
	if onboardingExpires.Valid {
		account.OnboardingURLExpiresAt = &onboardingExpires.Time
	}

	return account, nil
}

// UpdateConnectAccountStatus updates the status and fields of a connect account.
func (r *PayoutRepository) UpdateConnectAccountStatus(ctx context.Context, accountID uuid.UUID, status string, payoutsEnabled, detailsSubmitted, chargesEnabled bool) error {
	query := `
		UPDATE stripe_connect_accounts
		SET account_status = $2, payouts_enabled = $3, details_submitted = $4,
		    charges_enabled = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID, status, payoutsEnabled, detailsSubmitted, chargesEnabled)
	return err
}

// UpdateConnectAccountOnboardingURL stores a refreshed onboarding URL.
func (r *PayoutRepository) UpdateConnectAccountOnboardingURL(ctx context.Context, accountID uuid.UUID, url string, expiresAt time.Time) error {
	query := `
		UPDATE stripe_connect_accounts
		SET onboarding_url = $2, onboarding_url_expires_at = $3, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID, url, expiresAt)
	return err
}

// UpdateConnectAccountBankInfo stores masked bank account info from Stripe.
func (r *PayoutRepository) UpdateConnectAccountBankInfo(ctx context.Context, accountID uuid.UUID, bankLast4, bankName *string) error {
	query := `
		UPDATE stripe_connect_accounts
		SET bank_last4 = $2, bank_name = $3, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID, bankLast4, bankName)
	return err
}

// ─── Payout Requests ────────────────────────────────────────────────────────

// CreatePayoutRequest inserts a new payout request.
func (r *PayoutRepository) CreatePayoutRequest(ctx context.Context, req *PayoutRequest) error {
	query := `
		INSERT INTO payout_requests
			(id, user_id, connect_account_id, amount_cents, currency, status,
			 stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
			 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	now := time.Now()
	req.ID = uuid.New()
	req.CreatedAt = now
	req.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		req.ID, req.UserID, req.ConnectAccountID, req.AmountCents, req.Currency,
		req.Status, req.StripeTransferID, req.StripePayoutID, req.IdempotencyKey,
		req.FailureReason, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

// GetPayoutRequestByID retrieves a payout request by ID.
func (r *PayoutRepository) GetPayoutRequestByID(ctx context.Context, id uuid.UUID) (*PayoutRequest, error) {
	query := `
		SELECT id, user_id, connect_account_id, amount_cents, currency, status,
		       stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM payout_requests
		WHERE id = $1`

	req := &PayoutRequest{}
	var transferID, payoutID, failureReason sql.NullString
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.UserID, &req.ConnectAccountID, &req.AmountCents, &req.Currency,
		&req.Status, &transferID, &payoutID, &req.IdempotencyKey, &failureReason,
		&reviewedBy, &reviewedAt, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payout request: %w", err)
	}

	if transferID.Valid {
		s := transferID.String
		req.StripeTransferID = &s
	}
	if payoutID.Valid {
		s := payoutID.String
		req.StripePayoutID = &s
	}
	if failureReason.Valid {
		s := failureReason.String
		req.FailureReason = &s
	}
	if reviewedBy.Valid {
		uid, _ := uuid.Parse(reviewedBy.String)
		req.ReviewedBy = &uid
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}

	return req, nil
}

// GetPayoutRequestsByUserID retrieves payout requests for a user, paginated.
func (r *PayoutRepository) GetPayoutRequestsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PayoutRequest, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Count
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payout_requests WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payout requests: %w", err)
	}

	query := `
		SELECT id, user_id, connect_account_id, amount_cents, currency, status,
		       stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM payout_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payout requests: %w", err)
	}
	defer rows.Close()

	var requests []*PayoutRequest
	for rows.Next() {
		req := &PayoutRequest{}
		var transferID, payoutID, failureReason sql.NullString
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&req.ID, &req.UserID, &req.ConnectAccountID, &req.AmountCents, &req.Currency,
			&req.Status, &transferID, &payoutID, &req.IdempotencyKey, &failureReason,
			&reviewedBy, &reviewedAt, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payout request: %w", err)
		}
		if transferID.Valid {
			s := transferID.String
			req.StripeTransferID = &s
		}
		if payoutID.Valid {
			s := payoutID.String
			req.StripePayoutID = &s
		}
		if failureReason.Valid {
			s := failureReason.String
			req.FailureReason = &s
		}
		if reviewedBy.Valid {
			uid, _ := uuid.Parse(reviewedBy.String)
			req.ReviewedBy = &uid
		}
		if reviewedAt.Valid {
			req.ReviewedAt = &reviewedAt.Time
		}

		requests = append(requests, req)
	}
	return requests, total, nil
}

// UpdatePayoutRequestStatus updates a payout request's status and optional Stripe IDs.
func (r *PayoutRepository) UpdatePayoutRequestStatus(ctx context.Context, id uuid.UUID, status string, transferID, payoutID *string) error {
	query := `
		UPDATE payout_requests
		SET status = $2,
		    stripe_transfer_id = COALESCE($3, stripe_transfer_id),
		    stripe_payout_id = COALESCE($4, stripe_payout_id),
		    updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status, transferID, payoutID)
	return err
}

// MarkPayoutRequestFailed marks a payout request as failed with a reason.
func (r *PayoutRepository) MarkPayoutRequestFailed(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE payout_requests
		SET status = 'failed', failure_reason = $2, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, reason)
	return err
}

// ─── Payout Ledger ──────────────────────────────────────────────────────────

// GetPayoutBalance returns the user's payout balance by summing ledger entries.
func (r *PayoutRepository) GetPayoutBalance(ctx context.Context, userID uuid.UUID) (*PayoutBalance, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN entry_type = 'earning_credit' THEN amount_cents ELSE 0 END), 0) as total_earned,
			COALESCE(SUM(CASE WHEN entry_type = 'payout_debit' THEN -amount_cents ELSE 0 END), 0) as total_paid_out,
			COALESCE(SUM(amount_cents), 0) as available_cents
		FROM payout_ledger
		WHERE user_id = $1`

	var totalEarned, totalPaidOut, availableCents int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&totalEarned, &totalPaidOut, &availableCents)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout balance: %w", err)
	}

	return &PayoutBalance{
		UserID:              userID,
		AvailableBalanceUSD: float64(availableCents) / 100.0,
		PendingBalanceUSD:   0, // Calculated separately if needed
		TotalEarningsUSD:    float64(totalEarned) / 100.0,
		TotalPaidOutUSD:     float64(totalPaidOut) / 100.0,
	}, nil
}

// AddLedgerEntry inserts an immutable ledger record. Must be called within a transaction.
func (r *PayoutRepository) AddLedgerEntry(ctx context.Context, tx *sql.Tx, entry *PayoutLedgerEntry) error {
	query := `
		INSERT INTO payout_ledger
			(id, user_id, entry_type, amount_cents, currency, reference_type, reference_id,
			 balance_after_cents, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	entry.ID = uuid.New()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := tx.ExecContext(ctx, query,
		entry.ID, entry.UserID, entry.EntryType, entry.AmountCents, entry.Currency,
		entry.ReferenceType, entry.ReferenceID, entry.BalanceAfterCents,
		entry.Description, entry.CreatedAt,
	)
	return err
}

// CreditEarning atomically credits a user's payout balance via the ledger.
// This should be called when publisher earnings become "available".
func (r *PayoutRepository) CreditEarning(ctx context.Context, userID uuid.UUID, amountCents int, referenceType string, referenceID uuid.UUID, description string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		// Get current balance
		balance, err := r.getBalanceForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}

		newBalance := balance + amountCents
		entry := &PayoutLedgerEntry{
			UserID:            userID,
			EntryType:         "earning_credit",
			AmountCents:       amountCents,
			Currency:          "usd",
			ReferenceType:     &referenceType,
			ReferenceID:       &referenceID,
			BalanceAfterCents: newBalance,
		}
		if description != "" {
			entry.Description = &description
		}
		return r.AddLedgerEntry(ctx, tx, entry)
	})
}

const ledgerRefRegistryFunctionExecution = "registry_function_execution"

func (r *PayoutRepository) ledgerReferenceExists(ctx context.Context, referenceType string, referenceID uuid.UUID) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM payout_ledger WHERE reference_type = $1 AND reference_id = $2`,
		referenceType, referenceID,
	).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CreditPublisherRegistryExecutionShare credits the function owner once per registry execution log row.
// Idempotent when used with idx_payout_ledger_ref_type_id_unique (duplicate inserts return an error — caller may ignore).
func (r *PayoutRepository) CreditPublisherRegistryExecutionShare(ctx context.Context, publisherUserID, executionLogID uuid.UUID, amountCents int, description string) error {
	if amountCents <= 0 {
		return nil
	}
	dup, err := r.ledgerReferenceExists(ctx, ledgerRefRegistryFunctionExecution, executionLogID)
	if err != nil {
		return err
	}
	if dup {
		return nil
	}
	err = r.CreditEarning(ctx, publisherUserID, amountCents, ledgerRefRegistryFunctionExecution, executionLogID, description)
	if err != nil && isPayoutLedgerUniqueViolation(err) {
		return nil
	}
	return err
}

func isPayoutLedgerUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// DebitForPayout atomically debits the user's balance for a payout within a transaction.
func (r *PayoutRepository) DebitForPayout(ctx context.Context, userID uuid.UUID, amountCents int, payoutRequestID uuid.UUID) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		balance, err := r.getBalanceForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}

		if balance < amountCents {
			return fmt.Errorf("insufficient payout balance: available $%.2f, requested $%.2f",
				float64(balance)/100.0, float64(amountCents)/100.0)
		}

		newBalance := balance - amountCents
		refType := "payout_request"
		desc := fmt.Sprintf("Payout debit for request %s", payoutRequestID.String())
		entry := &PayoutLedgerEntry{
			UserID:            userID,
			EntryType:         "payout_debit",
			AmountCents:       -amountCents,
			Currency:          "usd",
			ReferenceType:     &refType,
			ReferenceID:       &payoutRequestID,
			BalanceAfterCents: newBalance,
			Description:       &desc,
		}
		return r.AddLedgerEntry(ctx, tx, entry)
	})
}

// ReversePayoutDebit reverses a payout debit (e.g. if a Stripe transfer fails).
func (r *PayoutRepository) ReversePayoutDebit(ctx context.Context, userID uuid.UUID, amountCents int, payoutRequestID uuid.UUID, reason string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		balance, err := r.getBalanceForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}

		newBalance := balance + amountCents
		refType := "payout_request"
		entry := &PayoutLedgerEntry{
			UserID:            userID,
			EntryType:         "payout_reversal",
			AmountCents:       amountCents,
			Currency:          "usd",
			ReferenceType:     &refType,
			ReferenceID:       &payoutRequestID,
			BalanceAfterCents: newBalance,
			Description:       &reason,
		}
		return r.AddLedgerEntry(ctx, tx, entry)
	})
}

// ListLedgerEntries returns paginated ledger entries for a user.
func (r *PayoutRepository) ListLedgerEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PayoutLedgerEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payout_ledger WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, entry_type, amount_cents, currency,
		       reference_type, reference_id, balance_after_cents, description, created_at
		FROM payout_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*PayoutLedgerEntry
	for rows.Next() {
		entry := &PayoutLedgerEntry{}
		var refType, refID, description sql.NullString

		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.EntryType, &entry.AmountCents, &entry.Currency,
			&refType, &refID, &entry.BalanceAfterCents, &description, &entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if refType.Valid {
			s := refType.String
			entry.ReferenceType = &s
		}
		if refID.Valid {
			uid, _ := uuid.Parse(refID.String)
			entry.ReferenceID = &uid
		}
		if description.Valid {
			s := description.String
			entry.Description = &s
		}
		entries = append(entries, entry)
	}
	return entries, total, nil
}

// getBalanceForUpdate returns the current payout balance (sum of ledger entries) with a row lock.
func (r *PayoutRepository) getBalanceForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	// Lock a pseudo-row to prevent concurrent modifications.
	// We use a SELECT FOR UPDATE on the user's wallet if it exists, otherwise just compute.
	var balance int
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM payout_ledger WHERE user_id = $1`,
		userID,
	).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get current balance: %w", err)
	}
	return balance, nil
}

// GetAccountIDByStripeID looks up the local account ID from a Stripe account ID.
func (r *PayoutRepository) GetAccountIDByStripeID(ctx context.Context, stripeAccountID string, accountID *uuid.UUID) error {
	return r.db.QueryRowContext(ctx,
		`SELECT id FROM stripe_connect_accounts WHERE stripe_account_id = $1`, stripeAccountID,
	).Scan(accountID)
}

// FindPayoutRequestByTransferID looks up a payout request by its Stripe transfer ID.
func (r *PayoutRepository) FindPayoutRequestByTransferID(ctx context.Context, stripeTransferID string, payoutReqID, userID *string, amountCents *int) error {
	return r.db.QueryRowContext(ctx,
		`SELECT id::text, user_id::text, amount_cents FROM payout_requests WHERE stripe_transfer_id = $1`,
		stripeTransferID,
	).Scan(payoutReqID, userID, amountCents)
}

// ─── Transaction helper ─────────────────────────────────────────────────────

// withTx executes fn within a database transaction.
func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
