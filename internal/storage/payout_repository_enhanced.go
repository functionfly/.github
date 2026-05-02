package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Payout Schedule Preferences
// =============================================================================

// PayoutSchedulePreference stores a user's auto-payout schedule settings.
type PayoutSchedulePreference struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	ScheduleEnabled    bool       `json:"schedule_enabled"`
	Frequency          string     `json:"frequency"` // weekly, biweekly, monthly
	MinimumAmountCents int        `json:"minimum_amount_cents"`
	DayOfWeek          *int       `json:"day_of_week,omitempty"`   // 0-6 for weekly
	DayOfMonth         *int       `json:"day_of_month,omitempty"` // 1-28 for monthly
	Currency           string     `json:"currency"`
	LastAutoPayoutAt   *time.Time `json:"last_auto_payout_at,omitempty"`
	NextScheduledAt    *time.Time `json:"next_scheduled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// GetSchedulePreference retrieves a user's payout schedule preference.
func (r *PayoutRepository) GetSchedulePreference(ctx context.Context, userID uuid.UUID) (*PayoutSchedulePreference, error) {
	query := `
		SELECT id, user_id, schedule_enabled, frequency, minimum_amount_cents,
		       day_of_week, day_of_month, currency, last_auto_payout_at,
		       next_scheduled_at, created_at, updated_at
		FROM payout_schedule_preferences
		WHERE user_id = $1`

	pref := &PayoutSchedulePreference{}
	var dayOfWeek, dayOfMonth sql.NullInt32
	var lastPayout, nextScheduled sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&pref.ID, &pref.UserID, &pref.ScheduleEnabled, &pref.Frequency,
		&pref.MinimumAmountCents, &dayOfWeek, &dayOfMonth, &pref.Currency,
		&lastPayout, &nextScheduled, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get schedule preference: %w", err)
	}

	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int32)
		pref.DayOfWeek = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int32)
		pref.DayOfMonth = &v
	}
	if lastPayout.Valid {
		pref.LastAutoPayoutAt = &lastPayout.Time
	}
	if nextScheduled.Valid {
		pref.NextScheduledAt = &nextScheduled.Time
	}

	return pref, nil
}

// UpsertSchedulePreference creates or updates a user's payout schedule preference.
func (r *PayoutRepository) UpsertSchedulePreference(ctx context.Context, pref *PayoutSchedulePreference) error {
	query := `
		INSERT INTO payout_schedule_preferences
			(id, user_id, schedule_enabled, frequency, minimum_amount_cents,
			 day_of_week, day_of_month, currency, next_scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id) DO UPDATE SET
			schedule_enabled = EXCLUDED.schedule_enabled,
			frequency = EXCLUDED.frequency,
			minimum_amount_cents = EXCLUDED.minimum_amount_cents,
			day_of_week = EXCLUDED.day_of_week,
			day_of_month = EXCLUDED.day_of_month,
			currency = EXCLUDED.currency,
			next_scheduled_at = EXCLUDED.next_scheduled_at,
			updated_at = NOW()`

	now := time.Now()
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}
	pref.CreatedAt = now
	pref.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		pref.ID, pref.UserID, pref.ScheduleEnabled, pref.Frequency,
		pref.MinimumAmountCents, pref.DayOfWeek, pref.DayOfMonth,
		pref.Currency, pref.NextScheduledAt, pref.CreatedAt, pref.UpdatedAt,
	)
	return err
}

// GetUsersWithScheduledPayouts returns all users with auto-payout enabled and due.
func (r *PayoutRepository) GetUsersWithScheduledPayouts(ctx context.Context, before time.Time) ([]*PayoutSchedulePreference, error) {
	query := `
		SELECT id, user_id, schedule_enabled, frequency, minimum_amount_cents,
		       day_of_week, day_of_month, currency, last_auto_payout_at,
		       next_scheduled_at, created_at, updated_at
		FROM payout_schedule_preferences
		WHERE schedule_enabled = TRUE
		  AND next_scheduled_at <= $1
		ORDER BY next_scheduled_at ASC`

	rows, err := r.db.QueryContext(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("failed to query scheduled payouts: %w", err)
	}
	defer rows.Close()

	var prefs []*PayoutSchedulePreference
	for rows.Next() {
		pref := &PayoutSchedulePreference{}
		var dayOfWeek, dayOfMonth sql.NullInt32
		var lastPayout, nextScheduled sql.NullTime

		err := rows.Scan(
			&pref.ID, &pref.UserID, &pref.ScheduleEnabled, &pref.Frequency,
			&pref.MinimumAmountCents, &dayOfWeek, &dayOfMonth, &pref.Currency,
			&lastPayout, &nextScheduled, &pref.CreatedAt, &pref.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int32)
			pref.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int32)
			pref.DayOfMonth = &v
		}
		if lastPayout.Valid {
			pref.LastAutoPayoutAt = &lastPayout.Time
		}
		if nextScheduled.Valid {
			pref.NextScheduledAt = &nextScheduled.Time
		}
		prefs = append(prefs, pref)
	}
	return prefs, rows.Err()
}

// UpdateSchedulePreferenceAfterPayout records the last auto-payout and computes the next one.
func (r *PayoutRepository) UpdateSchedulePreferenceAfterPayout(ctx context.Context, userID uuid.UUID, payoutAt time.Time) error {
	pref, err := r.GetSchedulePreference(ctx, userID)
	if err != nil || pref == nil {
		return err
	}

	nextAt := computeNextPayoutTime(pref, payoutAt)

	query := `
		UPDATE payout_schedule_preferences
		SET last_auto_payout_at = $2, next_scheduled_at = $3, updated_at = NOW()
		WHERE user_id = $1`

	_, err = r.db.ExecContext(ctx, query, userID, payoutAt, nextAt)
	return err
}

func computeNextPayoutTime(pref *PayoutSchedulePreference, from time.Time) time.Time {
	switch pref.Frequency {
	case "weekly":
		day := time.Sunday
		if pref.DayOfWeek != nil {
			day = time.Weekday(*pref.DayOfWeek)
		}
		daysUntil := int(day - from.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7
		}
		return from.AddDate(0, 0, daysUntil)
	case "biweekly":
		return from.AddDate(0, 0, 14)
	case "monthly":
		day := 1
		if pref.DayOfMonth != nil {
			day = *pref.DayOfMonth
		}
		next := from.AddDate(0, 1, 0)
		if day > 28 {
			day = 28
		}
		return time.Date(next.Year(), next.Month(), day, 0, 0, 0, 0, time.UTC)
	default:
		return from.AddDate(0, 0, 7)
	}
}

// =============================================================================
// Velocity Tracking
// =============================================================================

// PayoutVelocityRecord tracks payout patterns for fraud prevention.
type PayoutVelocityRecord struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	WindowStart      time.Time  `json:"window_start"`
	WindowEnd        time.Time  `json:"window_end"`
	PayoutCount      int        `json:"payout_count"`
	TotalAmountCents int        `json:"total_amount_cents"`
	LastPayoutAt     *time.Time `json:"last_payout_at,omitempty"`
	Flags            map[string]interface{} `json:"flags,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// RecordPayoutVelocity inserts or updates the 24-hour velocity window for a user.
func (r *PayoutRepository) RecordPayoutVelocity(ctx context.Context, userID uuid.UUID, amountCents int) (*PayoutVelocityRecord, error) {
	now := time.Now()
	windowStart := now.Truncate(24 * time.Hour)
	windowEnd := windowStart.Add(24 * time.Hour)

	query := `
		INSERT INTO payout_velocity_tracking
			(id, user_id, window_start, window_end, payout_count, total_amount_cents, last_payout_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 1, $5, $6, $7, $8)
		ON CONFLICT (user_id, window_start, window_end) DO UPDATE SET
			payout_count = payout_velocity_tracking.payout_count + 1,
			total_amount_cents = payout_velocity_tracking.total_amount_cents + EXCLUDED.total_amount_cents,
			last_payout_at = EXCLUDED.last_payout_at,
			updated_at = NOW()
		RETURNING id, payout_count, total_amount_cents`

	record := &PayoutVelocityRecord{
		UserID:           userID,
		WindowStart:      windowStart,
		WindowEnd:        windowEnd,
		TotalAmountCents: amountCents,
	}

	err := r.db.QueryRowContext(ctx, query,
		uuid.New(), userID, windowStart, windowEnd, amountCents, now, now, now,
	).Scan(&record.ID, &record.PayoutCount, &record.TotalAmountCents)
	if err != nil {
		return nil, fmt.Errorf("failed to record velocity: %w", err)
	}

	record.LastPayoutAt = &now
	return record, nil
}

// GetVelocityRecord retrieves the current 24-hour velocity window for a user.
func (r *PayoutRepository) GetVelocityRecord(ctx context.Context, userID uuid.UUID) (*PayoutVelocityRecord, error) {
	now := time.Now()
	windowStart := now.Truncate(24 * time.Hour)
	windowEnd := windowStart.Add(24 * time.Hour)

	query := `
		SELECT id, user_id, window_start, window_end, payout_count, total_amount_cents,
		       last_payout_at, created_at, updated_at
		FROM payout_velocity_tracking
		WHERE user_id = $1 AND window_start = $2 AND window_end = $3`

	record := &PayoutVelocityRecord{}
	var lastPayout sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, windowStart, windowEnd).Scan(
		&record.ID, &record.UserID, &record.WindowStart, &record.WindowEnd,
		&record.PayoutCount, &record.TotalAmountCents, &lastPayout,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if lastPayout.Valid {
		record.LastPayoutAt = &lastPayout.Time
	}
	return record, nil
}

// =============================================================================
// Fee Configuration & Deductions
// =============================================================================

// PayoutFeeConfig holds fee configuration for a payout type.
type PayoutFeeConfig struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	FeeType        string     `json:"fee_type"` // percentage, flat, tiered
	FeePercent     float64    `json:"fee_percent"`
	FlatFeeCents   int        `json:"flat_fee_cents"`
	MinimumFeeCents int       `json:"minimum_fee_cents"`
	MaximumFeeCents *int      `json:"maximum_fee_cents,omitempty"`
	IsActive       bool       `json:"is_active"`
	AppliesTo      string     `json:"applies_to"` // all, international, domestic
}

// PayoutFeeDeduction records a fee deducted from a payout.
type PayoutFeeDeduction struct {
	ID              uuid.UUID `json:"id"`
	PayoutRequestID uuid.UUID `json:"payout_request_id"`
	UserID          uuid.UUID `json:"user_id"`
	GrossAmountCents int      `json:"gross_amount_cents"`
	FeeAmountCents  int       `json:"fee_amount_cents"`
	NetAmountCents  int       `json:"net_amount_cents"`
	FeeConfigID     *uuid.UUID `json:"fee_config_id,omitempty"`
	FeeType         string    `json:"fee_type"`
	FeeRate         *float64  `json:"fee_rate,omitempty"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetActiveFeeConfig returns the active fee configuration for a given type.
func (r *PayoutRepository) GetActiveFeeConfig(ctx context.Context, name string) (*PayoutFeeConfig, error) {
	query := `
		SELECT id, name, description, fee_type, fee_percent, flat_fee_cents,
		       minimum_fee_cents, maximum_fee_cents, is_active, applies_to
		FROM payout_fee_config
		WHERE name = $1 AND is_active = TRUE
		  AND (effective_until IS NULL OR effective_until > NOW())`

	cfg := &PayoutFeeConfig{}
	var maxFee sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cfg.ID, &cfg.Name, &cfg.Description, &cfg.FeeType, &cfg.FeePercent,
		&cfg.FlatFeeCents, &cfg.MinimumFeeCents, &maxFee, &cfg.IsActive, &cfg.AppliesTo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if maxFee.Valid {
		v := int(maxFee.Int32)
		cfg.MaximumFeeCents = &v
	}
	return cfg, nil
}

// RecordFeeDeduction creates an immutable fee deduction record.
func (r *PayoutRepository) RecordFeeDeduction(ctx context.Context, tx *sql.Tx, deduction *PayoutFeeDeduction) error {
	query := `
		INSERT INTO payout_fee_deductions
			(id, payout_request_id, user_id, gross_amount_cents, fee_amount_cents,
			 net_amount_cents, fee_config_id, fee_type, fee_rate, currency, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	deduction.ID = uuid.New()
	if deduction.CreatedAt.IsZero() {
		deduction.CreatedAt = time.Now()
	}

	_, err := tx.ExecContext(ctx, query,
		deduction.ID, deduction.PayoutRequestID, deduction.UserID,
		deduction.GrossAmountCents, deduction.FeeAmountCents, deduction.NetAmountCents,
		deduction.FeeConfigID, deduction.FeeType, deduction.FeeRate,
		deduction.Currency, deduction.CreatedAt,
	)
	return err
}

// GetFeeDedByPayoutRequestID retrieves the fee deduction for a payout request.
func (r *PayoutRepository) GetFeeDedByPayoutRequestID(ctx context.Context, payoutRequestID uuid.UUID) (*PayoutFeeDeduction, error) {
	query := `
		SELECT id, payout_request_id, user_id, gross_amount_cents, fee_amount_cents,
		       net_amount_cents, fee_config_id, fee_type, fee_rate, currency, created_at
		FROM payout_fee_deductions
		WHERE payout_request_id = $1`

	d := &PayoutFeeDeduction{}
	var configID sql.NullString
	var feeRate sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, payoutRequestID).Scan(
		&d.ID, &d.PayoutRequestID, &d.UserID, &d.GrossAmountCents,
		&d.FeeAmountCents, &d.NetAmountCents, &configID, &d.FeeType,
		&feeRate, &d.Currency, &d.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if configID.Valid {
		uid, _ := uuid.Parse(configID.String)
		d.FeeConfigID = &uid
	}
	if feeRate.Valid {
		d.FeeRate = &feeRate.Float64
	}
	return d, nil
}

// =============================================================================
// Admin / Reporting Queries
// =============================================================================

// PayoutSummary provides an aggregated view for admin dashboards.
type PayoutSummary struct {
	TotalPayouts        int     `json:"total_payouts"`
	TotalAmountCents    int     `json:"total_amount_cents"`
	PendingCount        int     `json:"pending_count"`
	ProcessingCount     int     `json:"processing_count"`
	CompletedCount      int     `json:"completed_count"`
	FailedCount         int     `json:"failed_count"`
	CancelledCount      int     `json:"cancelled_count"`
	AverageAmountCents  float64 `json:"average_amount_cents"`
	TotalFeesCents      int     `json:"total_fees_cents"`
}

// GetAllPayoutRequests returns all payout requests with filtering for admin view.
func (r *PayoutRepository) GetAllPayoutRequests(ctx context.Context, status string, limit, offset int) ([]*PayoutRequest, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	countQuery := `SELECT COUNT(*) FROM payout_requests`
	dataQuery := `
		SELECT id, user_id, connect_account_id, amount_cents, currency, status,
		       stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM payout_requests`

	var args []interface{}
	if status != "" {
		countQuery += " WHERE status = $1"
		dataQuery += " WHERE status = $1"
		args = append(args, status)
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payout requests: %w", err)
	}

	dataQuery += " ORDER BY created_at DESC"
	if len(args) > 0 {
		dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	} else {
		dataQuery += " LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
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

// GetPayoutSummary returns aggregated payout statistics.
func (r *PayoutRepository) GetPayoutSummary(ctx context.Context, since time.Time) (*PayoutSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_payouts,
			COALESCE(SUM(amount_cents), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END), 0) as processing,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failed,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled,
			COALESCE(AVG(amount_cents), 0) as avg_amount
		FROM payout_requests
		WHERE created_at >= $1`

	s := &PayoutSummary{}
	err := r.db.QueryRowContext(ctx, query, since).Scan(
		&s.TotalPayouts, &s.TotalAmountCents, &s.PendingCount,
		&s.ProcessingCount, &s.CompletedCount, &s.FailedCount,
		&s.CancelledCount, &s.AverageAmountCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout summary: %w", err)
	}

	// Get total fees
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(fee_amount_cents), 0) FROM payout_fee_deductions WHERE created_at >= $1`, since,
	).Scan(&s.TotalFeesCents)
	if err != nil {
		s.TotalFeesCents = 0
	}

	return s, nil
}

// CancelPayoutRequest cancels a pending payout request and reverses the ledger debit.
func (r *PayoutRepository) CancelPayoutRequest(ctx context.Context, payoutID, userID uuid.UUID, reason string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		// Verify the payout is in a cancellable state
		var status string
		var amountCents int
		err := tx.QueryRowContext(ctx,
			`SELECT status, amount_cents FROM payout_requests WHERE id = $1 AND user_id = $2 FOR UPDATE`,
			payoutID, userID,
		).Scan(&status, &amountCents)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("payout request not found")
			}
			return err
		}

		if status != "pending" && status != "processing" {
			return fmt.Errorf("cannot cancel payout in status: %s", status)
		}

		// Update status to cancelled
		_, err = tx.ExecContext(ctx,
			`UPDATE payout_requests SET status = 'cancelled', failure_reason = $2, updated_at = NOW() WHERE id = $1`,
			payoutID, reason,
		)
		if err != nil {
			return fmt.Errorf("failed to cancel payout: %w", err)
		}

		// Reverse the ledger debit
		balance, err := getBalanceForUpdateTx(ctx, tx, userID)
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
			ReferenceID:       &payoutID,
			BalanceAfterCents: newBalance,
			Description:       &reason,
		}
		return addLedgerEntryTx(ctx, tx, entry)
	})
}

// getBalanceForUpdateTx is a transaction-scoped balance helper.
func getBalanceForUpdateTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	var balance int
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents), 0) FROM payout_ledger WHERE user_id = $1`,
		userID,
	).Scan(&balance)
	return balance, err
}

// addLedgerEntryTx inserts a ledger entry within an existing transaction.
func addLedgerEntryTx(ctx context.Context, tx *sql.Tx, entry *PayoutLedgerEntry) error {
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

// GetConnectAccountByStripeID retrieves a connect account by Stripe account ID.
func (r *PayoutRepository) GetConnectAccountByStripeID(ctx context.Context, stripeAccountID string) (*StripeConnectAccount, error) {
	query := `
		SELECT id, user_id, stripe_account_id, account_status, payouts_enabled, details_submitted,
		       charges_enabled, COALESCE(country, ''), currency,
		       bank_last4, bank_name, onboarding_url, onboarding_url_expires_at,
		       created_at, updated_at
		FROM stripe_connect_accounts
		WHERE stripe_account_id = $1`

	account := &StripeConnectAccount{}
	var country, bankLast4, bankName, onboardingURL sql.NullString
	var onboardingExpires sql.NullTime

	err := r.db.QueryRowContext(ctx, query, stripeAccountID).Scan(
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

// GetPayoutRequestsByStatus retrieves payout requests by status (admin use).
func (r *PayoutRepository) GetPayoutRequestsByStatus(ctx context.Context, status string, limit, offset int) ([]*PayoutRequest, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM payout_requests WHERE status = $1`, status,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, connect_account_id, amount_cents, currency, status,
		       stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM payout_requests
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
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

// ForceProcessPayout admin override to force-process a pending payout.
func (r *PayoutRepository) ForceProcessPayout(ctx context.Context, payoutID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payout_requests SET status = 'processing', updated_at = NOW() WHERE id = $1 AND status IN ('pending', 'failed')`,
		payoutID,
	)
	return err
}
