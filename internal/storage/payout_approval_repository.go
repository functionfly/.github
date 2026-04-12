package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PayoutApprovalRule defines rules for payout approval requirements
type PayoutApprovalRule struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	MinAmountUSD    float64   `json:"min_amount_usd"`
	MaxAmountUSD    *float64  `json:"max_amount_usd,omitempty"`
	RequiredApprovals int     `json:"required_approvals"` // 1 or 2
	ApproverRoles   []string  `json:"approver_roles"`      // e.g., ['admin', 'billing_manager']
	IsActive        bool      `json:"is_active"`
	Priority        int       `json:"priority"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PayoutApprovalAudit represents an audit log entry for payout approval actions
type PayoutApprovalAudit struct {
	ID              uuid.UUID  `json:"id"`
	PayoutRequestID uuid.UUID  `json:"payout_request_id"`
	Action          string     `json:"action"` // 'submitted', 'approved', 'rejected', 'second_approved', 'cancelled'
	PerformedBy     uuid.UUID  `json:"performed_by"`
	PerformedAt     time.Time  `json:"performed_at"`
	PreviousStatus  *string    `json:"previous_status,omitempty"`
	NewStatus       *string    `json:"new_status,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	IPAddress       *string    `json:"ip_address,omitempty"`
	UserAgent       *string    `json:"user_agent,omitempty"`
}

// PayoutApprovalRepository handles payout approval related database operations
type PayoutApprovalRepository struct {
	db *sql.DB
}

// NewPayoutApprovalRepository creates a new payout approval repository
func NewPayoutApprovalRepository(db *sql.DB) *PayoutApprovalRepository {
	return &PayoutApprovalRepository{db: db}
}

// GetApprovalRules retrieves all active approval rules ordered by priority
func (r *PayoutApprovalRepository) GetApprovalRules(ctx context.Context) ([]PayoutApprovalRule, error) {
	query := `
		SELECT id, name, description, min_amount_usd, max_amount_usd,
		       required_approvals, approver_roles, is_active, priority,
		       created_at, updated_at
		FROM payout_approval_rules
		WHERE is_active = true
		ORDER BY priority DESC, min_amount_usd ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval rules: %w", err)
	}
	defer rows.Close()

	var rules []PayoutApprovalRule
	for rows.Next() {
		var rule PayoutApprovalRule
		var maxAmount sql.NullFloat64
		var approverRoles []string

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.MinAmountUSD,
			&maxAmount, &rule.RequiredApprovals, &approverRoles,
			&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if maxAmount.Valid {
			rule.MaxAmountUSD = &maxAmount.Float64
		}
		rule.ApproverRoles = approverRoles

		rules = append(rules, rule)
	}

	return rules, nil
}

// GetApprovalRuleByID retrieves a specific approval rule
func (r *PayoutApprovalRepository) GetApprovalRuleByID(ctx context.Context, ruleID uuid.UUID) (*PayoutApprovalRule, error) {
	query := `
		SELECT id, name, description, min_amount_usd, max_amount_usd,
		       required_approvals, approver_roles, is_active, priority,
		       created_at, updated_at
		FROM payout_approval_rules
		WHERE id = $1`

	var rule PayoutApprovalRule
	var maxAmount sql.NullFloat64
	var approverRoles []string

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.MinAmountUSD,
		&maxAmount, &rule.RequiredApprovals, &approverRoles,
		&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval rule: %w", err)
	}

	if maxAmount.Valid {
		rule.MaxAmountUSD = &maxAmount.Float64
	}
	rule.ApproverRoles = approverRoles

	return &rule, nil
}

// CreateApprovalRule creates a new approval rule
func (r *PayoutApprovalRepository) CreateApprovalRule(ctx context.Context, rule *PayoutApprovalRule) error {
	query := `
		INSERT INTO payout_approval_rules
			(id, name, description, min_amount_usd, max_amount_usd,
			 required_approvals, approver_roles, is_active, priority,
			 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now()
	rule.ID = uuid.New()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	var maxAmount interface{}
	if rule.MaxAmountUSD != nil {
		maxAmount = *rule.MaxAmountUSD
	} else {
		maxAmount = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.MinAmountUSD,
		maxAmount, rule.RequiredApprovals, rule.ApproverRoles,
		rule.IsActive, rule.Priority, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create approval rule: %w", err)
	}

	return nil
}

// UpdateApprovalRule updates an existing approval rule
func (r *PayoutApprovalRepository) UpdateApprovalRule(ctx context.Context, ruleID uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE payout_approval_rules SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	if name, ok := updates["name"]; ok {
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, name)
		argCount++
	}
	if desc, ok := updates["description"]; ok {
		query += fmt.Sprintf(", description = $%d", argCount)
		args = append(args, desc)
		argCount++
	}
	if minAmt, ok := updates["min_amount_usd"]; ok {
		query += fmt.Sprintf(", min_amount_usd = $%d", argCount)
		args = append(args, minAmt)
		argCount++
	}
	if maxAmt, ok := updates["max_amount_usd"]; ok {
		query += fmt.Sprintf(", max_amount_usd = $%d", argCount)
		args = append(args, maxAmt)
		argCount++
	}
	if reqApprovals, ok := updates["required_approvals"]; ok {
		query += fmt.Sprintf(", required_approvals = $%d", argCount)
		args = append(args, reqApprovals)
		argCount++
	}
	if roles, ok := updates["approver_roles"]; ok {
		query += fmt.Sprintf(", approver_roles = $%d", argCount)
		args = append(args, roles)
		argCount++
	}
	if isActive, ok := updates["is_active"]; ok {
		query += fmt.Sprintf(", is_active = $%d", argCount)
		args = append(args, isActive)
		argCount++
	}
	if priority, ok := updates["priority"]; ok {
		query += fmt.Sprintf(", priority = $%d", argCount)
		args = append(args, priority)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, ruleID)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update approval rule: %w", err)
	}

	return nil
}

// DeleteApprovalRule marks an approval rule as inactive
func (r *PayoutApprovalRepository) DeleteApprovalRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `UPDATE payout_approval_rules SET is_active = false, updated_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete approval rule: %w", err)
	}

	return nil
}

// GetRequiredApprovalLevel determines if a payout requires approval and how many
func (r *PayoutApprovalRepository) GetRequiredApprovalLevel(ctx context.Context, amountCents int) (requiresApproval bool, requiredApprovals int, thresholdUSD float64, err error) {
	amountUSD := float64(amountCents) / 100.0

	rules, err := r.GetApprovalRules(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	// Find matching rule (rules are ordered by priority)
	for _, rule := range rules {
		if amountUSD >= rule.MinAmountUSD {
			if rule.MaxAmountUSD == nil || amountUSD <= *rule.MaxAmountUSD {
				return true, rule.RequiredApprovals, rule.MinAmountUSD, nil
			}
		}
	}

	// No matching rule - no approval required
	return false, 0, 0, nil
}

// ListPendingPayouts lists all payout requests awaiting approval
func (r *PayoutApprovalRepository) ListPendingPayouts(ctx context.Context, limit, offset int) ([]*PayoutRequest, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Count total
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM payout_requests WHERE status = 'pending' AND requires_approval = true`,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count pending payouts: %w", err)
	}

	query := `
		SELECT id, user_id, connect_account_id, amount_cents, currency, status,
		       stripe_transfer_id, stripe_payout_id, idempotency_key, failure_reason,
		       requires_approval, approval_threshold_usd, approved_by, approved_at,
		       approval_notes, second_approval_by, second_approval_at,
		       reviewed_by, reviewed_at, rejection_reason, created_at, updated_at
		FROM payout_requests
		WHERE status = 'pending' AND requires_approval = true
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list pending payouts: %w", err)
	}
	defer rows.Close()

	var requests []*PayoutRequest
	for rows.Next() {
		req := &PayoutRequest{}
		var transferID, payoutID, failureReason, approvalNotes, rejectionReason sql.NullString
		var reviewedBy, approvedBy, secondApprovedBy sql.NullString
		var reviewedAt, secondApprovedAt sql.NullTime
		var requiresApproval bool
		var approvalThreshold sql.NullFloat64

		err := rows.Scan(
			&req.ID, &req.UserID, &req.ConnectAccountID, &req.AmountCents, &req.Currency, &req.Status,
			&transferID, &payoutID, &req.IdempotencyKey, &failureReason,
			&requiresApproval, &approvalThreshold, &approvedBy, &req.ApprovedAt,
			&approvalNotes, &secondApprovedBy, &secondApprovedAt,
			&reviewedBy, &reviewedAt, &rejectionReason, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			continue
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
		req.RequiresApproval = requiresApproval
		if approvalThreshold.Valid {
			req.ApprovalThresholdUSD = &approvalThreshold.Float64
		}
		if rejectionReason.Valid {
			s := rejectionReason.String
			req.RejectionReason = &s
		}

		requests = append(requests, req)
	}

	return requests, total, nil
}

// ApprovePayout records an approval for a payout request
func (r *PayoutApprovalRepository) ApprovePayout(ctx context.Context, payoutID, adminID uuid.UUID, notes string, isSecondApproval bool) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		// Get current payout state
		var currentStatus string
		var approvedBy interface{}

		err := tx.QueryRowContext(ctx,
			`SELECT status, approved_by FROM payout_requests WHERE id = $1`,
			payoutID,
		).Scan(&currentStatus, &approvedBy)
		if err != nil {
			return fmt.Errorf("failed to get payout status: %w", err)
		}

		if currentStatus != "pending" {
			return fmt.Errorf("payout is not in pending status: %s", currentStatus)
		}

		// Determine if this is first or second approval
		if approvedBy == nil {
			// First approval
			_, err = tx.ExecContext(ctx, `
				UPDATE payout_requests
				SET approved_by = $1, approved_at = NOW(), approval_notes = $2, reviewed_by = $1, reviewed_at = NOW(), updated_at = NOW()
				WHERE id = $3`,
				adminID, notes, payoutID)
			if err != nil {
				return fmt.Errorf("failed to record first approval: %w", err)
			}
		} else if isSecondApproval {
			// Second approval - mark as processing to proceed with transfer
			_, err = tx.ExecContext(ctx, `
				UPDATE payout_requests
				SET second_approval_by = $1, second_approval_at = NOW(),
				    reviewed_by = $1, reviewed_at = NOW(), status = 'processing', updated_at = NOW()
				WHERE id = $2`,
				adminID, payoutID)
			if err != nil {
				return fmt.Errorf("failed to record second approval: %w", err)
			}
		} else {
			return fmt.Errorf("payout already has first approval, second approval required")
		}

		// Add audit log entry
		_, err = tx.ExecContext(ctx, `
			INSERT INTO payout_approval_audit
				(id, payout_request_id, action, performed_by, performed_at, previous_status, new_status, notes)
			VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7)`,
			uuid.New(), payoutID, map[bool]string{true: "second_approved", false: "approved"}[isSecondApproval],
			adminID, currentStatus, map[bool]string{true: "processing", false: "pending"}[isSecondApproval], notes)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return nil
	})
}

// RejectPayout rejects a payout request
func (r *PayoutApprovalRepository) RejectPayout(ctx context.Context, payoutID, adminID uuid.UUID, reason string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		var currentStatus string
		err := tx.QueryRowContext(ctx,
			`SELECT status FROM payout_requests WHERE id = $1`,
			payoutID,
		).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("failed to get payout status: %w", err)
		}

		if currentStatus != "pending" {
			return fmt.Errorf("payout is not in pending status: %s", currentStatus)
		}

		// Update payout status
		_, err = tx.ExecContext(ctx, `
			UPDATE payout_requests
			SET status = 'failed', failure_reason = $1, rejection_reason = $1, rejected_by = $2, rejected_at = NOW(), updated_at = NOW()
			WHERE id = $3`,
			reason, adminID, payoutID)
		if err != nil {
			return fmt.Errorf("failed to reject payout: %w", err)
		}

		// Add audit log entry
		_, err = tx.ExecContext(ctx, `
			INSERT INTO payout_approval_audit
				(id, payout_request_id, action, performed_by, performed_at, previous_status, new_status, notes)
			VALUES ($1, $2, 'rejected', $3, NOW(), $4, 'failed', $5)`,
			uuid.New(), payoutID, adminID, currentStatus, reason)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return nil
	})
}

// GetPayoutAuditLog retrieves the audit log for a payout request
func (r *PayoutApprovalRepository) GetPayoutAuditLog(ctx context.Context, payoutID uuid.UUID) ([]PayoutApprovalAudit, error) {
	query := `
		SELECT id, payout_request_id, action, performed_by, performed_at,
		       previous_status, new_status, notes, ip_address, user_agent
		FROM payout_approval_audit
		WHERE payout_request_id = $1
		ORDER BY performed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, payoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout audit log: %w", err)
	}
	defer rows.Close()

	var audits []PayoutApprovalAudit
	for rows.Next() {
		var audit PayoutApprovalAudit
		var prevStatus, newStatus, notes, ipAddr, userAgent sql.NullString

		err := rows.Scan(
			&audit.ID, &audit.PayoutRequestID, &audit.Action, &audit.PerformedBy, &audit.PerformedAt,
			&prevStatus, &newStatus, &notes, &ipAddr, &userAgent,
		)
		if err != nil {
			continue
		}

		if prevStatus.Valid {
			audit.PreviousStatus = &prevStatus.String
		}
		if newStatus.Valid {
			audit.NewStatus = &newStatus.String
		}
		if notes.Valid {
			audit.Notes = &notes.String
		}
		if ipAddr.Valid {
			audit.IPAddress = &ipAddr.String
		}
		if userAgent.Valid {
			audit.UserAgent = &userAgent.String
		}

		audits = append(audits, audit)
	}

	return audits, nil
}
