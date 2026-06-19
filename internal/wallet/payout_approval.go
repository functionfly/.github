package wallet

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// =============================================================================
// Payout Approval Models
// =============================================================================

// PayoutApprovalStatus represents the status of a payout in the approval workflow
type PayoutApprovalStatus string

const (
	PayoutApprovalStatusPending       PayoutApprovalStatus = "pending"
	PayoutApprovalStatusFirstApproved PayoutApprovalStatus = "first_approved"
	PayoutApprovalStatusFullyApproved PayoutApprovalStatus = "fully_approved"
	PayoutApprovalStatusRejected      PayoutApprovalStatus = "rejected"
	PayoutApprovalStatusProcessing    PayoutApprovalStatus = "processing"
	PayoutApprovalStatusCompleted     PayoutApprovalStatus = "completed"
	PayoutApprovalStatusFailed        PayoutApprovalStatus = "failed"
)

// PayoutApprovalRule defines when payouts require approval
type PayoutApprovalRule struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Name                   string    `json:"name" db:"name"`
	Description            string    `json:"description" db:"description"`
	MinAmountUSD           float64   `json:"min_amount_usd" db:"min_amount_usd"`
	MaxAmountUSD           *float64  `json:"max_amount_usd,omitempty" db:"max_amount_usd"`
	RequiresFirstApproval  bool      `json:"requires_first_approval" db:"requires_first_approval"`
	RequiresSecondApproval bool      `json:"requires_second_approval" db:"requires_second_approval"`
	FirstApproverRoles     []string  `json:"first_approver_roles" db:"first_approver_roles"`
	SecondApproverRoles    []string  `json:"second_approver_roles" db:"second_approver_roles"`
	IsActive               bool      `json:"is_active" db:"is_active"`
	Priority               int       `json:"priority" db:"priority"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// PayoutApprovalRecord tracks the approval workflow for a payout
type PayoutApprovalRecord struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	PayoutRequestID  uuid.UUID              `json:"payout_request_id" db:"payout_request_id"`
	Status           PayoutApprovalStatus   `json:"status" db:"status"`
	RuleID           *uuid.UUID             `json:"rule_id,omitempty" db:"rule_id"`
	AmountUSD        float64                `json:"amount_usd" db:"amount_usd"`
	SubmittedBy      uuid.UUID              `json:"submitted_by" db:"submitted_by"`
	SubmittedAt      time.Time              `json:"submitted_at" db:"submitted_at"`
	FirstApprovedBy  *uuid.UUID             `json:"first_approved_by,omitempty" db:"first_approved_by"`
	FirstApprovedAt  *time.Time             `json:"first_approved_at,omitempty" db:"first_approved_at"`
	SecondApprovedBy *uuid.UUID             `json:"second_approved_by,omitempty" db:"second_approved_by"`
	SecondApprovedAt *time.Time             `json:"second_approved_at,omitempty" db:"second_approved_at"`
	RejectedBy       *uuid.UUID             `json:"rejected_by,omitempty" db:"rejected_by"`
	RejectedAt       *time.Time             `json:"rejected_at,omitempty" db:"rejected_at"`
	RejectionReason  *string                `json:"rejection_reason,omitempty" db:"rejection_reason"`
	ApprovalNotes    *string                `json:"approval_notes,omitempty" db:"approval_notes"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// PayoutApprovalAuditEntry tracks all actions in the approval workflow
type PayoutApprovalAuditEntry struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ApprovalRecordID uuid.UUID `json:"approval_record_id" db:"approval_record_id"`
	Action           string    `json:"action" db:"action"` // submitted, first_approved, second_approved, rejected, etc.
	PerformedBy      uuid.UUID `json:"performed_by" db:"performed_by"`
	PerformedAt      time.Time `json:"performed_at" db:"performed_at"`
	PreviousStatus   *string   `json:"previous_status,omitempty" db:"previous_status"`
	NewStatus        string    `json:"new_status" db:"new_status"`
	Notes            *string   `json:"notes,omitempty" db:"notes"`
	IPAddress        *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string   `json:"user_agent,omitempty" db:"user_agent"`
}

// PayoutApprovalSummary provides a summary for admin dashboards
type PayoutApprovalSummary struct {
	TotalPending          int     `json:"total_pending"`
	TotalFirstApproved    int     `json:"total_first_approved"`
	TotalFullyApproved    int     `json:"total_fully_approved"`
	TotalRejected         int     `json:"total_rejected"`
	TotalAmountPendingUSD float64 `json:"total_amount_pending_usd"`
	RequiresAttention     int     `json:"requires_attention"`
}

// =============================================================================
// Payout Approval Service
// =============================================================================

// PayoutApprovalService manages the payout approval workflow
type PayoutApprovalService struct {
	db                     *sql.DB
	logger                 *logrus.Logger
	notifySvc              *notification.Service
	payoutRepo             *storage.PayoutRepository
	approvalThreshold      float64
	adminNotificationEmail string
}

// PayoutApprovalConfig holds configuration for the approval service
type PayoutApprovalConfig struct {
	ApprovalThresholdUSD       float64
	RequireSecondApprovalUSD   float64
	DefaultFirstApproverRoles  []string
	DefaultSecondApproverRoles []string
	AdminNotificationEmail     string
}

// DefaultPayoutApprovalConfig returns default configuration
func DefaultPayoutApprovalConfig() *PayoutApprovalConfig {
	adminEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL")
	if adminEmail == "" {
		logrus.Warn("ADMIN_NOTIFICATION_EMAIL not set - payout notifications will be disabled")
	}
	return &PayoutApprovalConfig{
		ApprovalThresholdUSD:       getEnvFloat64("PAYOUT_APPROVAL_THRESHOLD_USD", 1000.0),
		RequireSecondApprovalUSD:   getEnvFloat64("PAYOUT_SECOND_APPROVAL_THRESHOLD_USD", 10000.0),
		DefaultFirstApproverRoles:  []string{"finance", "admin"},
		DefaultSecondApproverRoles: []string{"finance_manager", "admin"},
		AdminNotificationEmail:     adminEmail,
	}
}

// NewPayoutApprovalService creates a new payout approval service
func NewPayoutApprovalService(db *sql.DB, payoutRepo *storage.PayoutRepository, notifySvc *notification.Service) *PayoutApprovalService {
	cfg := DefaultPayoutApprovalConfig()

	return &PayoutApprovalService{
		db:                     db,
		logger:                 logrus.New(),
		notifySvc:              notifySvc,
		payoutRepo:             payoutRepo,
		approvalThreshold:     cfg.ApprovalThresholdUSD,
		adminNotificationEmail: cfg.AdminNotificationEmail,
	}
}

// SetLogger sets the logger
func (s *PayoutApprovalService) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

// InitializeTables creates the necessary database tables
func (s *PayoutApprovalService) InitializeTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS payout_approval_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			min_amount_usd DECIMAL(14,4) NOT NULL DEFAULT 0,
			max_amount_usd DECIMAL(14,4),
			requires_first_approval BOOLEAN NOT NULL DEFAULT false,
			requires_second_approval BOOLEAN NOT NULL DEFAULT false,
			first_approver_roles TEXT[],
			second_approver_roles TEXT[],
			is_active BOOLEAN NOT NULL DEFAULT true,
			priority INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS payout_approval_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			payout_request_id UUID NOT NULL REFERENCES payout_requests(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			rule_id UUID REFERENCES payout_approval_rules(id),
			amount_usd DECIMAL(14,4) NOT NULL,
			submitted_by UUID NOT NULL REFERENCES users(id),
			submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			first_approved_by UUID REFERENCES users(id),
			first_approved_at TIMESTAMP WITH TIME ZONE,
			second_approved_by UUID REFERENCES users(id),
			second_approved_at TIMESTAMP WITH TIME ZONE,
			rejected_by UUID REFERENCES users(id),
			rejected_at TIMESTAMP WITH TIME ZONE,
			rejection_reason TEXT,
			approval_notes TEXT,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS payout_approval_audit (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			approval_record_id UUID NOT NULL REFERENCES payout_approval_records(id) ON DELETE CASCADE,
			action VARCHAR(50) NOT NULL,
			performed_by UUID NOT NULL REFERENCES users(id),
			performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			previous_status VARCHAR(50),
			new_status VARCHAR(50) NOT NULL,
			notes TEXT,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_payout_approval_records_status 
		 ON payout_approval_records(status)`,
		`CREATE INDEX IF NOT EXISTS idx_payout_approval_records_payout_request 
		 ON payout_approval_records(payout_request_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payout_approval_records_submitted 
		 ON payout_approval_records(submitted_by)`,
		`CREATE INDEX IF NOT EXISTS idx_payout_approval_audit_record 
		 ON payout_approval_audit(approval_record_id)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Insert default rule if none exists
	if err := s.createDefaultRule(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to create default approval rule")
	}

	return nil
}

// createDefaultRule creates the default approval rule if no rules exist
func (s *PayoutApprovalService) createDefaultRule(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payout_approval_rules`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	cfg := DefaultPayoutApprovalConfig()

	query := `
		INSERT INTO payout_approval_rules 
		(name, description, min_amount_usd, requires_first_approval, requires_second_approval, 
		 first_approver_roles, second_approver_roles, is_active, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.ExecContext(ctx, query,
		"Large Payout Approval",
		"Requires approval for payouts over $1,000",
		cfg.ApprovalThresholdUSD,
		true,
		cfg.RequireSecondApprovalUSD > 0,
		pq.Array(cfg.DefaultFirstApproverRoles),
		pq.Array(cfg.DefaultSecondApproverRoles),
		true,
		1,
	)

	return err
}

// EvaluatePayout determines if a payout requires approval and creates approval record
func (s *PayoutApprovalService) EvaluatePayout(ctx context.Context, payoutRequest *storage.PayoutRequest, submittedBy uuid.UUID) (*PayoutApprovalRecord, error) {
	amountUSD := float64(payoutRequest.AmountCents) / 100.0

	// Find applicable rule
	rule, err := s.findApplicableRule(ctx, amountUSD)
	if err != nil {
		return nil, err
	}

	// If no rule applies, no approval needed
	if rule == nil || (!rule.RequiresFirstApproval && !rule.RequiresSecondApproval) {
		return nil, nil
	}

	// Create approval record
	record := &PayoutApprovalRecord{
		PayoutRequestID: payoutRequest.ID,
		Status:          PayoutApprovalStatusPending,
		RuleID:          &rule.ID,
		AmountUSD:       amountUSD,
		SubmittedBy:     submittedBy,
		SubmittedAt:     time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO payout_approval_records 
		(payout_request_id, status, rule_id, amount_usd, submitted_by, submitted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err = s.db.QueryRowContext(ctx, query,
		record.PayoutRequestID,
		record.Status,
		record.RuleID,
		record.AmountUSD,
		record.SubmittedBy,
		record.SubmittedAt,
		record.CreatedAt,
		record.UpdatedAt,
	).Scan(&record.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create approval record: %w", err)
	}

	// Audit log
	if err := s.createAuditEntry(ctx, record.ID, "submitted", submittedBy, "", string(record.Status), nil, nil, nil); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit entry")
	}

	// Notify approvers
	if s.notifySvc != nil {
		s.notifyApprovers(ctx, record, rule)
	}

	return record, nil
}

// FirstApprove approves a payout (first approval)
func (s *PayoutApprovalService) FirstApprove(ctx context.Context, approvalRecordID uuid.UUID, approverID uuid.UUID, notes string, ipAddress, userAgent *string) error {
	record, err := s.getApprovalRecord(ctx, approvalRecordID)
	if err != nil {
		return err
	}

	if record.Status != PayoutApprovalStatusPending {
		return fmt.Errorf("payout cannot be approved - current status: %s", record.Status)
	}

	// Check if approver is the submitter
	if record.SubmittedBy == approverID {
		return fmt.Errorf("submitter cannot approve their own payout")
	}

	previousStatus := string(record.Status)
	now := time.Now()

	query := `
		UPDATE payout_approval_records
		SET status = $1, first_approved_by = $2, first_approved_at = $3, approval_notes = $4, updated_at = $5
		WHERE id = $6
	`

	newStatus := PayoutApprovalStatusFirstApproved
	if record.RuleID == nil {
		// No rule requiring second approval
		newStatus = PayoutApprovalStatusFullyApproved
	} else {
		rule, err := s.getRule(ctx, *record.RuleID)
		if err != nil || !rule.RequiresSecondApproval {
			newStatus = PayoutApprovalStatusFullyApproved
		}
	}

	_, err = s.db.ExecContext(ctx, query,
		newStatus,
		approverID,
		now,
		notes,
		now,
		approvalRecordID,
	)

	if err != nil {
		return fmt.Errorf("failed to approve payout: %w", err)
	}

	// Audit log
	if err := s.createAuditEntry(ctx, approvalRecordID, "first_approved", approverID, previousStatus, string(newStatus), &notes, ipAddress, userAgent); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit entry")
	}

	s.logger.WithFields(logrus.Fields{
		"approval_record_id": approvalRecordID,
		"approver_id":        approverID,
		"new_status":         newStatus,
	}).Info("Payout first approval completed")

	// Notify
	if s.notifySvc != nil {
		s.notifyApprovalStatusChange(ctx, record, string(newStatus))
	}

	return nil
}

// SecondApprove provides second approval for high-value payouts
func (s *PayoutApprovalService) SecondApprove(ctx context.Context, approvalRecordID uuid.UUID, approverID uuid.UUID, notes string, ipAddress, userAgent *string) error {
	record, err := s.getApprovalRecord(ctx, approvalRecordID)
	if err != nil {
		return err
	}

	if record.Status != PayoutApprovalStatusFirstApproved {
		return fmt.Errorf("payout cannot receive second approval - current status: %s", record.Status)
	}

	// Second approver must be different from first approver
	if record.FirstApprovedBy != nil && *record.FirstApprovedBy == approverID {
		return fmt.Errorf("second approver must be different from first approver")
	}

	// Also cannot be the submitter
	if record.SubmittedBy == approverID {
		return fmt.Errorf("submitter cannot approve their own payout")
	}

	previousStatus := string(record.Status)
	now := time.Now()

	query := `
		UPDATE payout_approval_records
		SET status = $1, second_approved_by = $2, second_approved_at = $3, approval_notes = COALESCE(approval_notes, '') || '\n' || $4, updated_at = $5
		WHERE id = $6
	`

	_, err = s.db.ExecContext(ctx, query,
		PayoutApprovalStatusFullyApproved,
		approverID,
		now,
		notes,
		now,
		approvalRecordID,
	)

	if err != nil {
		return fmt.Errorf("failed to second approve payout: %w", err)
	}

	// Audit log
	if err := s.createAuditEntry(ctx, approvalRecordID, "second_approved", approverID, previousStatus, string(PayoutApprovalStatusFullyApproved), &notes, ipAddress, userAgent); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit entry")
	}

	s.logger.WithFields(logrus.Fields{
		"approval_record_id": approvalRecordID,
		"approver_id":        approverID,
	}).Info("Payout fully approved")

	// Notify
	if s.notifySvc != nil {
		s.notifyApprovalStatusChange(ctx, record, string(PayoutApprovalStatusFullyApproved))
	}

	return nil
}

// Reject rejects a payout
func (s *PayoutApprovalService) Reject(ctx context.Context, approvalRecordID uuid.UUID, approverID uuid.UUID, reason string, ipAddress, userAgent *string) error {
	record, err := s.getApprovalRecord(ctx, approvalRecordID)
	if err != nil {
		return err
	}

	if record.Status != PayoutApprovalStatusPending && record.Status != PayoutApprovalStatusFirstApproved {
		return fmt.Errorf("payout cannot be rejected - current status: %s", record.Status)
	}

	previousStatus := string(record.Status)
	now := time.Now()

	query := `
		UPDATE payout_approval_records
		SET status = $1, rejected_by = $2, rejected_at = $3, rejection_reason = $4, updated_at = $5
		WHERE id = $6
	`

	_, err = s.db.ExecContext(ctx, query,
		PayoutApprovalStatusRejected,
		approverID,
		now,
		reason,
		now,
		approvalRecordID,
	)

	if err != nil {
		return fmt.Errorf("failed to reject payout: %w", err)
	}

	// Audit log
	if err := s.createAuditEntry(ctx, approvalRecordID, "rejected", approverID, previousStatus, string(PayoutApprovalStatusRejected), &reason, ipAddress, userAgent); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit entry")
	}

	// Update the payout request status
	if err := s.payoutRepo.MarkPayoutRequestFailed(ctx, record.PayoutRequestID, fmt.Sprintf("Rejected: %s", reason)); err != nil {
		s.logger.WithError(err).Warn("Failed to mark payout request as failed")
	}

	s.logger.WithFields(logrus.Fields{
		"approval_record_id": approvalRecordID,
		"rejected_by":        approverID,
		"reason":             reason,
	}).Warn("Payout rejected")

	// Notify
	if s.notifySvc != nil {
		s.notifyApprovalStatusChange(ctx, record, string(PayoutApprovalStatusRejected))
	}

	return nil
}

// GetPendingApprovals returns all pending approval records
func (s *PayoutApprovalService) GetPendingApprovals(ctx context.Context, limit, offset int) ([]*PayoutApprovalRecord, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Count total
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payout_approval_records WHERE status IN ('pending', 'first_approved')`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, payout_request_id, status, rule_id, amount_usd, submitted_by, submitted_at,
		       first_approved_by, first_approved_at, second_approved_by, second_approved_at,
		       rejected_by, rejected_at, rejection_reason, approval_notes, created_at, updated_at
		FROM payout_approval_records
		WHERE status IN ('pending', 'first_approved')
		ORDER BY 
			CASE status 
				WHEN 'pending' THEN 1 
				WHEN 'first_approved' THEN 2 
				ELSE 3 
			END,
			amount_usd DESC,
			submitted_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return s.scanApprovalRecords(rows, total)
}

// GetApprovalSummary returns a summary of approval status
func (s *PayoutApprovalService) GetApprovalSummary(ctx context.Context) (*PayoutApprovalSummary, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
			COUNT(*) FILTER (WHERE status = 'first_approved') as first_approved_count,
			COUNT(*) FILTER (WHERE status = 'fully_approved') as fully_approved_count,
			COUNT(*) FILTER (WHERE status = 'rejected') as rejected_count,
			COALESCE(SUM(amount_usd) FILTER (WHERE status = 'pending'), 0) as pending_amount
		FROM payout_approval_records
		WHERE created_at > NOW() - INTERVAL '30 days'
	`

	summary := &PayoutApprovalSummary{}
	var pendingAmount float64

	err := s.db.QueryRowContext(ctx, query).Scan(
		&summary.TotalPending,
		&summary.TotalFirstApproved,
		&summary.TotalFullyApproved,
		&summary.TotalRejected,
		&pendingAmount,
	)
	if err != nil {
		return nil, err
	}

	summary.TotalAmountPendingUSD = pendingAmount
	summary.RequiresAttention = summary.TotalPending + summary.TotalFirstApproved

	return summary, nil
}

// GetApprovalHistory returns the approval history for a payout request
func (s *PayoutApprovalService) GetApprovalHistory(ctx context.Context, payoutRequestID uuid.UUID) ([]*PayoutApprovalAuditEntry, error) {
	query := `
		SELECT a.id, a.approval_record_id, a.action, a.performed_by, a.performed_at,
		       a.previous_status, a.new_status, a.notes, a.ip_address, a.user_agent
		FROM payout_approval_audit a
		JOIN payout_approval_records r ON a.approval_record_id = r.id
		WHERE r.payout_request_id = $1
		ORDER BY a.performed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, payoutRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*PayoutApprovalAuditEntry
	for rows.Next() {
		entry := &PayoutApprovalAuditEntry{}
		var prevStatus, notes, ipAddr, userAgent sql.NullString

		err := rows.Scan(
			&entry.ID, &entry.ApprovalRecordID, &entry.Action, &entry.PerformedBy, &entry.PerformedAt,
			&prevStatus, &entry.NewStatus, &notes, &ipAddr, &userAgent,
		)
		if err != nil {
			continue
		}

		if prevStatus.Valid {
			s := prevStatus.String
			entry.PreviousStatus = &s
		}
		if notes.Valid {
			s := notes.String
			entry.Notes = &s
		}
		if ipAddr.Valid {
			s := ipAddr.String
			entry.IPAddress = &s
		}
		if userAgent.Valid {
			s := userAgent.String
			entry.UserAgent = &s
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Helper methods

func (s *PayoutApprovalService) findApplicableRule(ctx context.Context, amountUSD float64) (*PayoutApprovalRule, error) {
	query := `
		SELECT id, name, description, min_amount_usd, max_amount_usd,
		       requires_first_approval, requires_second_approval,
		       first_approver_roles, second_approver_roles,
		       is_active, priority, created_at, updated_at
		FROM payout_approval_rules
		WHERE is_active = true
		  AND min_amount_usd <= $1
		  AND (max_amount_usd IS NULL OR max_amount_usd >= $1)
		ORDER BY priority DESC, min_amount_usd ASC
		LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, query, amountUSD)
	return s.scanRule(row)
}

func (s *PayoutApprovalService) getRule(ctx context.Context, ruleID uuid.UUID) (*PayoutApprovalRule, error) {
	query := `
		SELECT id, name, description, min_amount_usd, max_amount_usd,
		       requires_first_approval, requires_second_approval,
		       first_approver_roles, second_approver_roles,
		       is_active, priority, created_at, updated_at
		FROM payout_approval_rules
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, ruleID)
	return s.scanRule(row)
}

func (s *PayoutApprovalService) getApprovalRecord(ctx context.Context, id uuid.UUID) (*PayoutApprovalRecord, error) {
	query := `
		SELECT id, payout_request_id, status, rule_id, amount_usd, submitted_by, submitted_at,
		       first_approved_by, first_approved_at, second_approved_by, second_approved_at,
		       rejected_by, rejected_at, rejection_reason, approval_notes, created_at, updated_at
		FROM payout_approval_records
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanApprovalRecord(row)
}

func (s *PayoutApprovalService) createAuditEntry(ctx context.Context, approvalRecordID uuid.UUID, action string, performedBy uuid.UUID, previousStatus, newStatus string, notes, ipAddress, userAgent *string) error {
	query := `
		INSERT INTO payout_approval_audit 
		(approval_record_id, action, performed_by, performed_at, previous_status, new_status, notes, ip_address, user_agent)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6, $7, $8)
	`

	_, err := s.db.ExecContext(ctx, query,
		approvalRecordID, action, performedBy,
		sql.NullString{String: previousStatus, Valid: previousStatus != ""},
		newStatus,
		sql.NullString{String: *notes, Valid: notes != nil},
		sql.NullString{String: *ipAddress, Valid: ipAddress != nil},
		sql.NullString{String: *userAgent, Valid: userAgent != nil},
	)

	return err
}

func (s *PayoutApprovalService) notifyApprovers(ctx context.Context, record *PayoutApprovalRecord, rule *PayoutApprovalRule) {
	if s.notifySvc == nil {
		return
	}

	data := map[string]interface{}{
		"approval_record_id": record.ID,
		"payout_request_id":  record.PayoutRequestID,
		"amount_usd":         record.AmountUSD,
		"submitted_by":       record.SubmittedBy,
		"submitted_at":       record.SubmittedAt,
		"status":             record.Status,
	}

	// Notify admins with approval roles - use billing alert for admin notifications
	if err := s.notifySvc.SendBillingAlert(ctx, s.adminNotificationEmail, "payout_approval_needed", data); err != nil {
		s.logger.WithError(err).Warn("Failed to send approval notification")
	}
}

func (s *PayoutApprovalService) notifyApprovalStatusChange(ctx context.Context, record *PayoutApprovalRecord, newStatus string) {
	if s.notifySvc == nil {
		return
	}

	// Log status change - notification service doesn't have a direct Send method for arbitrary types
	// This can be expanded to use Send() with a proper request if needed
	s.logger.WithFields(logrus.Fields{
		"approval_record_id": record.ID,
		"payout_request_id":  record.PayoutRequestID,
		"amount_usd":         record.AmountUSD,
		"new_status":         newStatus,
		"submitted_by":       record.SubmittedBy,
	}).Info("Payout approval status changed")
}

// Scan helpers

func (s *PayoutApprovalService) scanRule(row *sql.Row) (*PayoutApprovalRule, error) {
	rule := &PayoutApprovalRule{}
	var maxAmount sql.NullFloat64
	var firstRoles, secondRoles []string

	err := row.Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.MinAmountUSD, &maxAmount,
		&rule.RequiresFirstApproval, &rule.RequiresSecondApproval,
		pq.Array(&firstRoles), pq.Array(&secondRoles),
		&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if maxAmount.Valid {
		rule.MaxAmountUSD = &maxAmount.Float64
	}
	rule.FirstApproverRoles = firstRoles
	rule.SecondApproverRoles = secondRoles

	return rule, nil
}

func (s *PayoutApprovalService) scanApprovalRecord(row *sql.Row) (*PayoutApprovalRecord, error) {
	record := &PayoutApprovalRecord{}
	var ruleID, firstApprovedBy, secondApprovedBy, rejectedBy uuid.NullUUID
	var firstApprovedAt, secondApprovedAt, rejectedAt sql.NullTime
	var rejectionReason, approvalNotes sql.NullString
	var metadata []byte

	err := row.Scan(
		&record.ID, &record.PayoutRequestID, &record.Status, &ruleID, &record.AmountUSD,
		&record.SubmittedBy, &record.SubmittedAt,
		&firstApprovedBy, &firstApprovedAt, &secondApprovedBy, &secondApprovedAt,
		&rejectedBy, &rejectedAt, &rejectionReason, &approvalNotes,
		&record.CreatedAt, &record.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if ruleID.Valid {
		record.RuleID = &ruleID.UUID
	}
	if firstApprovedBy.Valid {
		record.FirstApprovedBy = &firstApprovedBy.UUID
	}
	if firstApprovedAt.Valid {
		record.FirstApprovedAt = &firstApprovedAt.Time
	}
	if secondApprovedBy.Valid {
		record.SecondApprovedBy = &secondApprovedBy.UUID
	}
	if secondApprovedAt.Valid {
		record.SecondApprovedAt = &secondApprovedAt.Time
	}
	if rejectedBy.Valid {
		record.RejectedBy = &rejectedBy.UUID
	}
	if rejectedAt.Valid {
		record.RejectedAt = &rejectedAt.Time
	}
	if rejectionReason.Valid {
		s := rejectionReason.String
		record.RejectionReason = &s
	}
	if approvalNotes.Valid {
		s := approvalNotes.String
		record.ApprovalNotes = &s
	}
	if len(metadata) > 0 {
		json.Unmarshal(metadata, &record.Metadata)
	}

	return record, nil
}

func (s *PayoutApprovalService) scanApprovalRecords(rows *sql.Rows, total int) ([]*PayoutApprovalRecord, int, error) {
	var records []*PayoutApprovalRecord

	for rows.Next() {
		record := &PayoutApprovalRecord{}
		var ruleID, firstApprovedBy, secondApprovedBy, rejectedBy uuid.NullUUID
		var firstApprovedAt, secondApprovedAt, rejectedAt sql.NullTime
		var rejectionReason, approvalNotes sql.NullString

		err := rows.Scan(
			&record.ID, &record.PayoutRequestID, &record.Status, &ruleID, &record.AmountUSD,
			&record.SubmittedBy, &record.SubmittedAt,
			&firstApprovedBy, &firstApprovedAt, &secondApprovedBy, &secondApprovedAt,
			&rejectedBy, &rejectedAt, &rejectionReason, &approvalNotes,
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan approval record")
			continue
		}

		if ruleID.Valid {
			record.RuleID = &ruleID.UUID
		}
		if firstApprovedBy.Valid {
			record.FirstApprovedBy = &firstApprovedBy.UUID
		}
		if firstApprovedAt.Valid {
			record.FirstApprovedAt = &firstApprovedAt.Time
		}
		if secondApprovedBy.Valid {
			record.SecondApprovedBy = &secondApprovedBy.UUID
		}
		if secondApprovedAt.Valid {
			record.SecondApprovedAt = &secondApprovedAt.Time
		}
		if rejectedBy.Valid {
			record.RejectedBy = &rejectedBy.UUID
		}
		if rejectedAt.Valid {
			record.RejectedAt = &rejectedAt.Time
		}
		if rejectionReason.Valid {
			s := rejectionReason.String
			record.RejectionReason = &s
		}
		if approvalNotes.Valid {
			s := approvalNotes.String
			record.ApprovalNotes = &s
		}

		records = append(records, record)
	}

	return records, total, rows.Err()
}

// GetApprovalRecordByPayoutRequestID retrieves the approval record for a specific payout request
func (s *PayoutApprovalService) GetApprovalRecordByPayoutRequestID(ctx context.Context, payoutRequestID uuid.UUID) (*PayoutApprovalRecord, error) {
	query := `
		SELECT id, payout_request_id, status, rule_id, amount_usd, submitted_by, submitted_at,
		       first_approved_by, first_approved_at, second_approved_by, second_approved_at,
		       rejected_by, rejected_at, rejection_reason, approval_notes, created_at, updated_at
		FROM payout_approval_records
		WHERE payout_request_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, query, payoutRequestID)
	return s.scanApprovalRecord(row)
}

// ListApprovalRules returns all active approval rules
func (s *PayoutApprovalService) ListApprovalRules(ctx context.Context) ([]*PayoutApprovalRule, error) {
	query := `
		SELECT id, name, description, min_amount_usd, max_amount_usd,
		       requires_first_approval, requires_second_approval,
		       first_approver_roles, second_approver_roles,
		       is_active, priority, created_at, updated_at
		FROM payout_approval_rules
		WHERE is_active = true
		ORDER BY priority DESC, min_amount_usd ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*PayoutApprovalRule
	for rows.Next() {
		rule, err := s.scanRuleFromRow(rows)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan approval rule")
			continue
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// CreateApprovalRule creates a new payout approval rule
func (s *PayoutApprovalService) CreateApprovalRule(ctx context.Context, rule *PayoutApprovalRule) (*PayoutApprovalRule, error) {
	query := `
		INSERT INTO payout_approval_rules 
		(name, description, min_amount_usd, max_amount_usd, requires_first_approval, requires_second_approval,
		 first_approver_roles, second_approver_roles, is_active, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	rule.ID = uuid.New()
	err := s.db.QueryRowContext(ctx, query,
		rule.Name, rule.Description, rule.MinAmountUSD, rule.MaxAmountUSD,
		rule.RequiresFirstApproval, rule.RequiresSecondApproval,
		pq.Array(rule.FirstApproverRoles), pq.Array(rule.SecondApproverRoles),
		rule.IsActive, rule.Priority,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create approval rule: %w", err)
	}

	return rule, nil
}

// scanRuleFromRow scans a rule from a sql.Rows (for use with QueryContext)
func (s *PayoutApprovalService) scanRuleFromRow(rows *sql.Rows) (*PayoutApprovalRule, error) {
	rule := &PayoutApprovalRule{}
	var maxAmount sql.NullFloat64
	var firstRoles, secondRoles []string

	err := rows.Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.MinAmountUSD, &maxAmount,
		&rule.RequiresFirstApproval, &rule.RequiresSecondApproval,
		pq.Array(&firstRoles), pq.Array(&secondRoles),
		&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if maxAmount.Valid {
		rule.MaxAmountUSD = &maxAmount.Float64
	}
	rule.FirstApproverRoles = firstRoles
	rule.SecondApproverRoles = secondRoles

	return rule, nil
}
