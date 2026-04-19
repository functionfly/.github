package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PaymentDispute represents a Stripe chargeback or dispute
type PaymentDispute struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           *uuid.UUID     `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	UserID             *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid;index"`
	StripeDisputeID    string         `json:"stripe_dispute_id" gorm:"not null;uniqueIndex;size:255"`
	StripePaymentID    string         `json:"stripe_payment_id" gorm:"not null;index;size:255"`
	StripeChargeID     string         `json:"stripe_charge_id" gorm:"index;size:255"`
	AmountCents        int            `json:"amount_cents" gorm:"not null"`
	Currency           string         `json:"currency" gorm:"not null;size:3;default:'USD'"`
	Reason             string         `json:"reason" gorm:"not null;size:100"`
	Status             string         `json:"status" gorm:"not null;size:50;index"`
	EvidenceDueBy      *time.Time     `json:"evidence_due_by,omitempty"`
	EvidenceSubmitted  bool           `json:"evidence_submitted" gorm:"default:false"`
	EvidenceData       datatypes.JSON `json:"evidence_data,omitempty" gorm:"type:jsonb;default:'{}'"`
	Outcome            string         `json:"outcome,omitempty" gorm:"size:50"`
	OutcomeReason      string         `json:"outcome_reason,omitempty" gorm:"size:255"`
	NetworkReasonCode  string         `json:"network_reason_code,omitempty" gorm:"size:50"`
	IsChargeRefundable bool           `json:"is_charge_refundable" gorm:"default:false"`
	RefundID           *string        `json:"refund_id,omitempty" gorm:"index;size:255"`
	Metadata           datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt          time.Time      `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"not null;default:now()"`
	ResolvedAt         *time.Time     `json:"resolved_at,omitempty"`
	// Related entities (not stored, populated on query)
	TenantName string `json:"tenant_name,omitempty" gorm:"-"`
	UserEmail  string `json:"user_email,omitempty" gorm:"-"`
	UserName   string `json:"user_name,omitempty" gorm:"-"`
}

// TableName specifies the table name for PaymentDispute
func (PaymentDispute) TableName() string {
	return "payment_disputes"
}

// IsOpen returns true if the dispute is in an open state requiring action
func (d *PaymentDispute) IsOpen() bool {
	switch d.Status {
	case "needs_response", "warning_needs_response", "needs_review":
		return true
	default:
		return false
	}
}

// IsResolved returns true if the dispute has been resolved
func (d *PaymentDispute) IsResolved() bool {
	switch d.Status {
	case "won", "lost", "closed", "warning_closed":
		return true
	default:
		return false
	}
}

// EvidenceDetails contains structured evidence data for dispute submission
type EvidenceDetails struct {
	ProductDescription    string `json:"product_description,omitempty"`
	CustomerEmail         string `json:"customer_email,omitempty"`
	CustomerName          string `json:"customer_name,omitempty"`
	CustomerPurchaseIP    string `json:"customer_purchase_ip,omitempty"`
	BillingAddress        string `json:"billing_address,omitempty"`
	ReceiptURL            string `json:"receipt_url,omitempty"`
	ServiceDate           string `json:"service_date,omitempty"`
	ServiceDocument       string `json:"service_document,omitempty"`
	ShippingAddress       string `json:"shipping_address,omitempty"`
	ShippingDate          string `json:"shipping_date,omitempty"`
	ShippingTracking      string `json:"shipping_tracking,omitempty"`
	ShippingCarrier       string `json:"shipping_carrier,omitempty"`
	RefundPolicyURL       string `json:"refund_policy_url,omitempty"`
	RefundPolicyDisclosed bool   `json:"refund_policy_disclosed,omitempty"`
	CancellationReason    string `json:"cancellation_reason,omitempty"`
	AccessActivityLog     string `json:"access_activity_log,omitempty"`
	CustomerCommunication string `json:"customer_communication,omitempty"`
	UncategorizedText     string `json:"uncategorized_text,omitempty"`
	UncategorizedFile     string `json:"uncategorized_file,omitempty"`
}

// ToJSON converts EvidenceDetails to JSON for storage
func (e *EvidenceDetails) ToJSON() datatypes.JSON {
	data, _ := json.Marshal(e)
	return datatypes.JSON(data)
}

// FromJSON parses EvidenceDetails from JSON
func EvidenceDetailsFromJSON(data datatypes.JSON) *EvidenceDetails {
	var e EvidenceDetails
	if err := json.Unmarshal(data, &e); err != nil {
		return nil
	}
	return &e
}

// PaymentRefund represents a Stripe refund
type PaymentRefund struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        *uuid.UUID     `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	UserID          *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid;index"`
	StripeRefundID  string         `json:"stripe_refund_id" gorm:"not null;uniqueIndex;size:255"`
	StripePaymentID string         `json:"stripe_payment_id" gorm:"not null;index;size:255"`
	StripeChargeID  string         `json:"stripe_charge_id" gorm:"index;size:255"`
	AmountCents     int            `json:"amount_cents" gorm:"not null"`
	Currency        string         `json:"currency" gorm:"not null;size:3;default:'USD'"`
	Status          string         `json:"status" gorm:"not null;size:50;index"`
	Reason          string         `json:"reason,omitempty" gorm:"size:50"`
	ReceiptNumber   string         `json:"receipt_number,omitempty" gorm:"size:100"`
	ReceiptURL      string         `json:"receipt_url,omitempty" gorm:"size:500"`
	Description     string         `json:"description,omitempty" gorm:"size:500"`
	Metadata        datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	FailureReason   string         `json:"failure_reason,omitempty" gorm:"size:255"`
	CreatedAt       time.Time      `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"not null;default:now()"`
	// Related entities (not stored, populated on query)
	TenantName string `json:"tenant_name,omitempty" gorm:"-"`
	UserEmail  string `json:"user_email,omitempty" gorm:"-"`
	UserName   string `json:"user_name,omitempty" gorm:"-"`
}

// TableName specifies the table name for PaymentRefund
func (PaymentRefund) TableName() string {
	return "payment_refunds"
}

// DisputeFilter provides filtering options for dispute queries
type DisputeFilter struct {
	TenantID       *uuid.UUID
	UserID         *uuid.UUID
	Status         string
	Reason         string
	Outcome        string
	StartDate      *time.Time
	EndDate        *time.Time
	IsOpen         *bool
	RequiresAction *bool
}

// RefundFilter provides filtering options for refund queries
type RefundFilter struct {
	TenantID  *uuid.UUID
	UserID    *uuid.UUID
	Status    string
	Reason    string
	StartDate *time.Time
	EndDate   *time.Time
}

// DisputeRepository handles dispute database operations
type DisputeRepository struct {
	db *gorm.DB
}

// NewDisputeRepository creates a new dispute repository
func NewDisputeRepository(db *gorm.DB) *DisputeRepository {
	return &DisputeRepository{db: db}
}

// DB returns the underlying database connection for raw queries
func (r *DisputeRepository) DB() *gorm.DB {
	return r.db
}

// UpsertDispute creates or updates a dispute record (idempotent by stripe_dispute_id)
func (r *DisputeRepository) UpsertDispute(ctx context.Context, dispute *PaymentDispute) error {
	if dispute.ID == uuid.Nil {
		dispute.ID = uuid.New()
	}
	if dispute.CreatedAt.IsZero() {
		dispute.CreatedAt = time.Now().UTC()
	}
	dispute.UpdatedAt = time.Now().UTC()

	return r.db.WithContext(ctx).
		Where("stripe_dispute_id = ?", dispute.StripeDisputeID).
		Assign(dispute).
		FirstOrCreate(dispute).Error
}

// GetDisputeByStripeID retrieves a dispute by Stripe dispute ID
func (r *DisputeRepository) GetDisputeByStripeID(ctx context.Context, stripeDisputeID string) (*PaymentDispute, error) {
	var dispute PaymentDispute
	err := r.db.WithContext(ctx).
		Where("stripe_dispute_id = ?", stripeDisputeID).
		First(&dispute).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

// GetDisputeByID retrieves a dispute by internal ID
func (r *DisputeRepository) GetDisputeByID(ctx context.Context, id uuid.UUID) (*PaymentDispute, error) {
	var dispute PaymentDispute
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&dispute).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

// ListDisputes retrieves disputes with optional filtering
func (r *DisputeRepository) ListDisputes(ctx context.Context, filter *DisputeFilter, limit, offset int) ([]*PaymentDispute, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	query := r.db.WithContext(ctx).Model(&PaymentDispute{})

	if filter != nil {
		if filter.TenantID != nil {
			query = query.Where("tenant_id = ?", *filter.TenantID)
		}
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Reason != "" {
			query = query.Where("reason = ?", filter.Reason)
		}
		if filter.Outcome != "" {
			query = query.Where("outcome = ?", filter.Outcome)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
		if filter.IsOpen != nil {
			if *filter.IsOpen {
				query = query.Where("status IN ?", []string{"needs_response", "warning_needs_response", "needs_review", "under_review"})
			} else {
				query = query.Where("status IN ?", []string{"won", "lost", "closed", "warning_closed"})
			}
		}
		if filter.RequiresAction != nil && *filter.RequiresAction {
			query = query.Where("status IN ?", []string{"needs_response", "warning_needs_response"})
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var disputes []*PaymentDispute
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&disputes).Error
	if err != nil {
		return nil, 0, err
	}

	return disputes, total, nil
}

// ListDisputesWithRelations retrieves disputes with tenant and user info
func (r *DisputeRepository) ListDisputesWithRelations(ctx context.Context, filter *DisputeFilter, limit, offset int) ([]*PaymentDispute, int64, error) {
	disputes, total, err := r.ListDisputes(ctx, filter, limit, offset)
	if err != nil || len(disputes) == 0 {
		return disputes, total, err
	}

	// Populate tenant and user info
	for _, d := range disputes {
		if d.TenantID != nil {
			var tenantName string
			r.db.WithContext(ctx).Raw("SELECT name FROM tenants WHERE id = ?", *d.TenantID).Scan(&tenantName)
			d.TenantName = tenantName
		}
		if d.UserID != nil {
			type UserInfo struct {
				Email    string
				Username string
			}
			var userInfo UserInfo
			r.db.WithContext(ctx).Raw("SELECT email, username FROM users WHERE id = ?", *d.UserID).Scan(&userInfo)
			d.UserEmail = userInfo.Email
			d.UserName = userInfo.Username
		}
	}

	return disputes, total, nil
}

// UpdateDisputeStatus updates the status of a dispute
func (r *DisputeRepository) UpdateDisputeStatus(ctx context.Context, id uuid.UUID, status string, outcome string, outcomeReason string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}

	if outcome != "" {
		updates["outcome"] = outcome
	}
	if outcomeReason != "" {
		updates["outcome_reason"] = outcomeReason
	}

	if status == "won" || status == "lost" || status == "closed" {
		now := time.Now().UTC()
		updates["resolved_at"] = &now
	}

	return r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateDisputeEvidence marks evidence as submitted
func (r *DisputeRepository) UpdateDisputeEvidence(ctx context.Context, id uuid.UUID, evidenceData *EvidenceDetails) error {
	updates := map[string]interface{}{
		"evidence_submitted": true,
		"updated_at":         time.Now().UTC(),
	}

	if evidenceData != nil {
		updates["evidence_data"] = evidenceData.ToJSON()
	}

	return r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetDisputeStats returns aggregate statistics about disputes
func (r *DisputeRepository) GetDisputeStats(ctx context.Context) (*DisputeStats, error) {
	stats := &DisputeStats{}

	// Total disputes
	var totalDisputes int64
	if err := r.db.WithContext(ctx).Model(&PaymentDispute{}).Count(&totalDisputes).Error; err != nil {
		return nil, err
	}
	stats.TotalDisputes = totalDisputes

	// By status
	type statusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []statusCount
	if err := r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	stats.ByStatus = make(map[string]int64)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Open disputes (need response)
	var openDisputes int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("status IN ?", []string{"needs_response", "warning_needs_response", "needs_review"}).
		Count(&openDisputes).Error; err != nil {
		return nil, err
	}
	stats.OpenDisputes = openDisputes

	// Total disputed amount
	var totalDisputedCents int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Select("COALESCE(SUM(amount_cents), 0)").
		Scan(&totalDisputedCents).Error; err != nil {
		return nil, err
	}
	stats.TotalDisputedCents = totalDisputedCents

	// Won disputes
	var wonDisputes int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("status = ? OR outcome = ?", "won", "won").
		Count(&wonDisputes).Error; err != nil {
		return nil, err
	}
	stats.WonDisputes = wonDisputes

	// Lost disputes
	var lostDisputes int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("status = ? OR outcome = ?", "lost", "lost").
		Count(&lostDisputes).Error; err != nil {
		return nil, err
	}
	stats.LostDisputes = lostDisputes

	return stats, nil
}

// DisputeStats contains aggregate dispute statistics
type DisputeStats struct {
	TotalDisputes      int64            `json:"total_disputes"`
	OpenDisputes       int64            `json:"open_disputes"`
	WonDisputes        int64            `json:"won_disputes"`
	LostDisputes       int64            `json:"lost_disputes"`
	TotalDisputedCents int64            `json:"total_disputed_cents"`
	ByStatus           map[string]int64 `json:"by_status"`
}

// TotalDisputedUSD returns the total disputed amount in USD
func (s *DisputeStats) TotalDisputedUSD() float64 {
	return float64(s.TotalDisputedCents) / 100.0
}

// RefundRepository handles refund database operations
type RefundRepository struct {
	db *gorm.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *gorm.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

// DB returns the underlying database connection for raw queries
func (r *RefundRepository) DB() *gorm.DB {
	return r.db
}

// UpsertRefund creates or updates a refund record (idempotent by stripe_refund_id)
func (r *RefundRepository) UpsertRefund(ctx context.Context, refund *PaymentRefund) error {
	if refund.ID == uuid.Nil {
		refund.ID = uuid.New()
	}
	if refund.CreatedAt.IsZero() {
		refund.CreatedAt = time.Now().UTC()
	}
	refund.UpdatedAt = time.Now().UTC()

	return r.db.WithContext(ctx).
		Where("stripe_refund_id = ?", refund.StripeRefundID).
		Assign(refund).
		FirstOrCreate(refund).Error
}

// GetRefundByStripeID retrieves a refund by Stripe refund ID
func (r *RefundRepository) GetRefundByStripeID(ctx context.Context, stripeRefundID string) (*PaymentRefund, error) {
	var refund PaymentRefund
	err := r.db.WithContext(ctx).
		Where("stripe_refund_id = ?", stripeRefundID).
		First(&refund).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// GetRefundByID retrieves a refund by internal ID
func (r *RefundRepository) GetRefundByID(ctx context.Context, id uuid.UUID) (*PaymentRefund, error) {
	var refund PaymentRefund
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&refund).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// ListRefunds retrieves refunds with optional filtering
func (r *RefundRepository) ListRefunds(ctx context.Context, filter *RefundFilter, limit, offset int) ([]*PaymentRefund, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	query := r.db.WithContext(ctx).Model(&PaymentRefund{})

	if filter != nil {
		if filter.TenantID != nil {
			query = query.Where("tenant_id = ?", *filter.TenantID)
		}
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Reason != "" {
			query = query.Where("reason = ?", filter.Reason)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var refunds []*PaymentRefund
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&refunds).Error
	if err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

// ListRefundsWithRelations retrieves refunds with tenant and user info
func (r *RefundRepository) ListRefundsWithRelations(ctx context.Context, filter *RefundFilter, limit, offset int) ([]*PaymentRefund, int64, error) {
	refunds, total, err := r.ListRefunds(ctx, filter, limit, offset)
	if err != nil || len(refunds) == 0 {
		return refunds, total, err
	}

	// Populate tenant and user info
	for _, ref := range refunds {
		if ref.TenantID != nil {
			var tenantName string
			r.db.WithContext(ctx).Raw("SELECT name FROM tenants WHERE id = ?", *ref.TenantID).Scan(&tenantName)
			ref.TenantName = tenantName
		}
		if ref.UserID != nil {
			type UserInfo struct {
				Email    string
				Username string
			}
			var userInfo UserInfo
			r.db.WithContext(ctx).Raw("SELECT email, username FROM users WHERE id = ?", *ref.UserID).Scan(&userInfo)
			ref.UserEmail = userInfo.Email
			ref.UserName = userInfo.Username
		}
	}

	return refunds, total, nil
}

// GetRefundStats returns aggregate statistics about refunds
func (r *RefundRepository) GetRefundStats(ctx context.Context) (*RefundStats, error) {
	stats := &RefundStats{}

	// Total refunds
	var totalRefunds int64
	if err := r.db.WithContext(ctx).Model(&PaymentRefund{}).Count(&totalRefunds).Error; err != nil {
		return nil, err
	}
	stats.TotalRefunds = totalRefunds

	// By status
	type statusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []statusCount
	if err := r.db.WithContext(ctx).
		Model(&PaymentRefund{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	stats.ByStatus = make(map[string]int64)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// By reason
	type reasonCount struct {
		Reason string
		Count  int64
	}
	var reasonCounts []reasonCount
	if err := r.db.WithContext(ctx).
		Model(&PaymentRefund{}).
		Select("reason, COUNT(*) as count").
		Group("reason").
		Scan(&reasonCounts).Error; err != nil {
		return nil, err
	}

	stats.ByReason = make(map[string]int64)
	for _, rc := range reasonCounts {
		stats.ByReason[rc.Reason] = rc.Count
	}

	// Total refunded amount
	var totalRefundedCents int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentRefund{}).
		Select("COALESCE(SUM(amount_cents), 0)").
		Scan(&totalRefundedCents).Error; err != nil {
		return nil, err
	}
	stats.TotalRefundedCents = totalRefundedCents

	// Recent refunds (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentRefundCount int64
	if err := r.db.WithContext(ctx).
		Model(&PaymentRefund{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&recentRefundCount).Error; err != nil {
		return nil, err
	}
	stats.RecentRefundCount = recentRefundCount

	return stats, nil
}

// RefundStats contains aggregate refund statistics
type RefundStats struct {
	TotalRefunds       int64            `json:"total_refunds"`
	TotalRefundedCents int64            `json:"total_refunded_cents"`
	RecentRefundCount  int64            `json:"recent_refund_count"`
	ByStatus           map[string]int64 `json:"by_status"`
	ByReason           map[string]int64 `json:"by_reason"`
}

// TotalRefundedUSD returns the total refunded amount in USD
func (s *RefundStats) TotalRefundedUSD() float64 {
	return float64(s.TotalRefundedCents) / 100.0
}

// ChargebackReconciliation tracks the financial impact of chargebacks
type ChargebackReconciliation struct {
	TotalDisputedCents       int64   `json:"total_disputed_cents"`
	TotalRefundedCents       int64   `json:"total_refunded_cents"`
	NetChargebackImpactCents int64   `json:"net_chargeback_impact_cents"`
	DisputeFeeCents          int64   `json:"dispute_fee_cents"` // Stripe's $15 dispute fee
	RecoveryRate             float64 `json:"recovery_rate"`     // Percentage won
}

// GetChargebackReconciliation returns the financial reconciliation for a period
func (r *DisputeRepository) GetChargebackReconciliation(ctx context.Context, startDate, endDate *time.Time) (*ChargebackReconciliation, error) {
	recon := &ChargebackReconciliation{}

	query := r.db.WithContext(ctx).Model(&PaymentDispute{})
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var totalDisputedCents int64
	if err := query.Select("COALESCE(SUM(amount_cents), 0)").Scan(&totalDisputedCents).Error; err != nil {
		return nil, err
	}
	recon.TotalDisputedCents = totalDisputedCents

	// Calculate disputes won
	var wonCents, lostCents int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'won' OR outcome = 'won' THEN amount_cents ELSE 0 END), 0) as won_cents,
			COALESCE(SUM(CASE WHEN status = 'lost' OR outcome = 'lost' THEN amount_cents ELSE 0 END), 0) as lost_cents
		FROM payment_disputes
		WHERE ($1::timestamp IS NULL OR created_at >= $1)
		AND ($2::timestamp IS NULL OR created_at <= $2)
	`, startDate, endDate).Row().Scan(&wonCents, &lostCents)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Calculate recovery rate
	if totalDisputedCents > 0 {
		recon.RecoveryRate = float64(wonCents) / float64(totalDisputedCents) * 100
	}

	// Stripe dispute fee is $15 per dispute (1,500 cents)
	var disputeCount int64
	if err := r.db.WithContext(ctx).Model(&PaymentDispute{}).
		Where("status NOT IN ?", []string{"warning_needs_response", "warning_closed"}).
		Count(&disputeCount).Error; err != nil {
		return nil, err
	}
	recon.DisputeFeeCents = disputeCount * 1500

	// Net impact = Lost disputes + fees - Won disputes (recovered)
	recon.NetChargebackImpactCents = lostCents + recon.DisputeFeeCents - wonCents

	return recon, nil
}

// LinkDisputeToRefund links a dispute to a refund record
func (r *DisputeRepository) LinkDisputeToRefund(ctx context.Context, disputeID uuid.UUID, refundID string) error {
	return r.db.WithContext(ctx).
		Model(&PaymentDispute{}).
		Where("id = ?", disputeID).
		Updates(map[string]interface{}{
			"refund_id":  refundID,
			"updated_at": time.Now().UTC(),
		}).Error
}
