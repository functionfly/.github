package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

func (r *BillingRepository) CreateAffiliateCode(ctx context.Context, code *AffiliateCode) (*AffiliateCode, error) {
	code.ID = uuid.New()
	code.CreatedAt = time.Now()
	code.UpdatedAt = time.Now()
	code.TotalReferrals = 0
	code.TotalCommissions = 0
	code.PendingCommissions = 0
	code.PendingEarningsCents = 0
	code.TotalEarningsCents = 0
	code.PaidOutEarningsCents = 0

	query := `
		INSERT INTO affiliate_codes (id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at`

	err := r.db.QueryRow(query, code.ID, code.Code, code.PublisherID, code.TenantID, code.Name, code.Description,
		code.CommissionType, code.CommissionValue, code.MaxCommissions, code.MaxReferrals, code.TotalReferrals,
		code.TotalCommissions, code.PendingCommissions, code.PendingEarningsCents, code.TotalEarningsCents,
		code.PaidOutEarningsCents, code.ValidFrom, code.ValidUntil, code.IsActive, code.UTMSource, code.UTMCampaign,
		code.CreatedAt, code.UpdatedAt).Scan(
		&code.ID, &code.Code, &code.PublisherID, &code.TenantID, &code.Name, &code.Description,
		&code.CommissionType, &code.CommissionValue, &code.MaxCommissions, &code.MaxReferrals, &code.TotalReferrals,
		&code.TotalCommissions, &code.PendingCommissions, &code.PendingEarningsCents, &code.TotalEarningsCents,
		&code.PaidOutEarningsCents, &code.ValidFrom, &code.ValidUntil, &code.IsActive, &code.UTMSource, &code.UTMCampaign,
		&code.CreatedAt, &code.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create affiliate code: %w", err)
	}

	return code, nil
}

func (r *BillingRepository) GetAffiliateCodeByID(ctx context.Context, id uuid.UUID) (*AffiliateCode, error) {
	query := `SELECT id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at
			  FROM affiliate_codes WHERE id = $1`

	code := &AffiliateCode{}
	err := r.db.QueryRow(query, id).Scan(&code.ID, &code.Code, &code.PublisherID, &code.TenantID, &code.Name, &code.Description,
		&code.CommissionType, &code.CommissionValue, &code.MaxCommissions, &code.MaxReferrals, &code.TotalReferrals,
		&code.TotalCommissions, &code.PendingCommissions, &code.PendingEarningsCents, &code.TotalEarningsCents,
		&code.PaidOutEarningsCents, &code.ValidFrom, &code.ValidUntil, &code.IsActive, &code.UTMSource, &code.UTMCampaign,
		&code.CreatedAt, &code.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliate code: %w", err)
	}

	return code, nil
}

func (r *BillingRepository) GetAffiliateCodeByCode(ctx context.Context, code string) (*AffiliateCode, error) {
	query := `SELECT id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at
			  FROM affiliate_codes WHERE UPPER(code) = UPPER($1) AND is_active = true`

	affiliateCode := &AffiliateCode{}
	err := r.db.QueryRow(query, code).Scan(&affiliateCode.ID, &affiliateCode.Code, &affiliateCode.PublisherID, &affiliateCode.TenantID, &affiliateCode.Name, &affiliateCode.Description,
		&affiliateCode.CommissionType, &affiliateCode.CommissionValue, &affiliateCode.MaxCommissions, &affiliateCode.MaxReferrals, &affiliateCode.TotalReferrals,
		&affiliateCode.TotalCommissions, &affiliateCode.PendingCommissions, &affiliateCode.PendingEarningsCents, &affiliateCode.TotalEarningsCents,
		&affiliateCode.PaidOutEarningsCents, &affiliateCode.ValidFrom, &affiliateCode.ValidUntil, &affiliateCode.IsActive, &affiliateCode.UTMSource, &affiliateCode.UTMCampaign,
		&affiliateCode.CreatedAt, &affiliateCode.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliate code: %w", err)
	}

	return affiliateCode, nil
}

func (r *BillingRepository) ListAffiliateCodesByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCode, error) {
	query := `SELECT id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at
			  FROM affiliate_codes WHERE publisher_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, publisherID)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliate codes: %w", err)
	}
	defer rows.Close()

	var codes []*AffiliateCode
	for rows.Next() {
		code := &AffiliateCode{}
		err := rows.Scan(&code.ID, &code.Code, &code.PublisherID, &code.TenantID, &code.Name, &code.Description,
			&code.CommissionType, &code.CommissionValue, &code.MaxCommissions, &code.MaxReferrals, &code.TotalReferrals,
			&code.TotalCommissions, &code.PendingCommissions, &code.PendingEarningsCents, &code.TotalEarningsCents,
			&code.PaidOutEarningsCents, &code.ValidFrom, &code.ValidUntil, &code.IsActive, &code.UTMSource, &code.UTMCampaign,
			&code.CreatedAt, &code.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affiliate code: %w", err)
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func (r *BillingRepository) ListAffiliateCodes(ctx context.Context) ([]*AffiliateCode, error) {
	query := `SELECT id, code, publisher_id, tenant_id, name, description, commission_type, commission_value, max_commissions, max_referrals, total_referrals, total_commissions, pending_commissions, pending_earnings_cents, total_earnings_cents, paid_out_earnings_cents, valid_from, valid_until, is_active, utm_source, utm_campaign, created_at, updated_at
			  FROM affiliate_codes ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliate codes: %w", err)
	}
	defer rows.Close()

	var codes []*AffiliateCode
	for rows.Next() {
		code := &AffiliateCode{}
		err := rows.Scan(&code.ID, &code.Code, &code.PublisherID, &code.TenantID, &code.Name, &code.Description,
			&code.CommissionType, &code.CommissionValue, &code.MaxCommissions, &code.MaxReferrals, &code.TotalReferrals,
			&code.TotalCommissions, &code.PendingCommissions, &code.PendingEarningsCents, &code.TotalEarningsCents,
			&code.PaidOutEarningsCents, &code.ValidFrom, &code.ValidUntil, &code.IsActive, &code.UTMSource, &code.UTMCampaign,
			&code.CreatedAt, &code.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affiliate code: %w", err)
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func (r *BillingRepository) UpdateAffiliateCode(ctx context.Context, code *AffiliateCode) error {
	code.UpdatedAt = time.Now()

	query := `UPDATE affiliate_codes SET
		name = $2, description = $3, commission_type = $4, commission_value = $5,
		max_commissions = $6, max_referrals = $7, valid_from = $8, valid_until = $9,
		is_active = $10, utm_source = $11, utm_campaign = $12, updated_at = $13
		WHERE id = $1`

	_, err := r.db.Exec(query, code.ID, code.Name, code.Description, code.CommissionType, code.CommissionValue,
		code.MaxCommissions, code.MaxReferrals, code.ValidFrom, code.ValidUntil, code.IsActive,
		code.UTMSource, code.UTMCampaign, code.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update affiliate code: %w", err)
	}

	return nil
}

func (r *BillingRepository) CreateAffiliateReferral(ctx context.Context, referral *AffiliateReferral) (*AffiliateReferral, error) {
	referral.ID = uuid.New()
	referral.CreatedAt = time.Now()
	referral.UpdatedAt = time.Now()

	query := `INSERT INTO affiliate_referrals (id, affiliate_code_id, referred_tenant_id, subscription_id, utm_source, utm_campaign, utm_content, utm_term, ip_address, user_agent, status, referred_at, converted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	err := r.db.QueryRow(query, referral.ID, referral.AffiliateCodeID, referral.ReferredTenantID, referral.SubscriptionID,
		referral.UTMSource, referral.UTMCampaign, referral.UTContent, referral.UTMTerm, referral.IPAddress,
		referral.UserAgent, referral.Status, referral.ReferredAt, referral.ConvertedAt, referral.CreatedAt, referral.UpdatedAt).Scan(&referral.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create affiliate referral: %w", err)
	}

	// Update affiliate code referral count
	_, err = r.db.Exec("UPDATE affiliate_codes SET total_referrals = total_referrals + 1 WHERE id = $1", referral.AffiliateCodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to update referral count: %w", err)
	}

	return referral, nil
}

func (r *BillingRepository) GetAffiliateReferralByID(ctx context.Context, id uuid.UUID) (*AffiliateReferral, error) {
	query := `SELECT id, affiliate_code_id, referred_tenant_id, subscription_id, utm_source, utm_campaign, utm_content, utm_term, ip_address, user_agent, status, referred_at, converted_at, created_at, updated_at
			  FROM affiliate_referrals WHERE id = $1`

	referral := &AffiliateReferral{}
	err := r.db.QueryRow(query, id).Scan(&referral.ID, &referral.AffiliateCodeID, &referral.ReferredTenantID,
		&referral.SubscriptionID, &referral.UTMSource, &referral.UTMCampaign, &referral.UTContent, &referral.UTMTerm,
		&referral.IPAddress, &referral.UserAgent, &referral.Status, &referral.ReferredAt, &referral.ConvertedAt,
		&referral.CreatedAt, &referral.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliate referral: %w", err)
	}

	return referral, nil
}

func (r *BillingRepository) GetAffiliateReferralByTenant(ctx context.Context, tenantID uuid.UUID) (*AffiliateReferral, error) {
	query := `SELECT id, affiliate_code_id, referred_tenant_id, subscription_id, utm_source, utm_campaign, utm_content, utm_term, ip_address, user_agent, status, referred_at, converted_at, created_at, updated_at
			  FROM affiliate_referrals WHERE referred_tenant_id = $1 ORDER BY created_at DESC LIMIT 1`

	referral := &AffiliateReferral{}
	err := r.db.QueryRow(query, tenantID).Scan(&referral.ID, &referral.AffiliateCodeID, &referral.ReferredTenantID,
		&referral.SubscriptionID, &referral.UTMSource, &referral.UTMCampaign, &referral.UTContent, &referral.UTMTerm,
		&referral.IPAddress, &referral.UserAgent, &referral.Status, &referral.ReferredAt, &referral.ConvertedAt,
		&referral.CreatedAt, &referral.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliate referral: %w", err)
	}

	return referral, nil
}

func (r *BillingRepository) ListAffiliateReferralsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateReferral, error) {
	query := `SELECT id, affiliate_code_id, referred_tenant_id, subscription_id, utm_source, utm_campaign, utm_content, utm_term, ip_address, user_agent, status, referred_at, converted_at, created_at, updated_at
			  FROM affiliate_referrals WHERE affiliate_code_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, codeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliate referrals: %w", err)
	}
	defer rows.Close()

	var referrals []*AffiliateReferral
	for rows.Next() {
		referral := &AffiliateReferral{}
		err := rows.Scan(&referral.ID, &referral.AffiliateCodeID, &referral.ReferredTenantID,
			&referral.SubscriptionID, &referral.UTMSource, &referral.UTMCampaign, &referral.UTContent, &referral.UTMTerm,
			&referral.IPAddress, &referral.UserAgent, &referral.Status, &referral.ReferredAt, &referral.ConvertedAt,
			&referral.CreatedAt, &referral.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affiliate referral: %w", err)
		}
		referrals = append(referrals, referral)
	}

	return referrals, nil
}

func (r *BillingRepository) UpdateAffiliateReferralStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec("UPDATE affiliate_referrals SET status = $2, updated_at = $3 WHERE id = $1", id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update referral status: %w", err)
	}
	return nil
}

func (r *BillingRepository) CreateAffiliateCommission(ctx context.Context, commission *AffiliateCommission) (*AffiliateCommission, error) {
	commission.ID = uuid.New()
	commission.CreatedAt = time.Now()
	commission.UpdatedAt = time.Now()

	query := `INSERT INTO affiliate_commissions (id, affiliate_code_id, referral_id, commission_type, commission_value, base_amount_cents, base_amount_usd, commission_cents, commission_usd, status, paid_at, payment_batch_id, payment_batch, subscription_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id`

	err := r.db.QueryRow(query, commission.ID, commission.AffiliateCodeID, commission.ReferralID,
		commission.CommissionType, commission.CommissionValue, commission.BaseAmountCents, commission.BaseAmountUSD,
		commission.CommissionCents, commission.CommissionUSD, commission.Status, commission.PaidAt,
		commission.PaymentBatchID, commission.PaymentBatch, commission.SubscriptionID, commission.Notes,
		commission.CreatedAt, commission.UpdatedAt).Scan(&commission.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create affiliate commission: %w", err)
	}

	// Update affiliate code commission tracking
	_, err = r.db.Exec(`UPDATE affiliate_codes SET 
		pending_commissions = pending_commissions + 1,
		pending_earnings_cents = pending_earnings_cents + $2
		WHERE id = $1`, commission.AffiliateCodeID, commission.CommissionCents)
	if err != nil {
		return nil, fmt.Errorf("failed to update affiliate code earnings: %w", err)
	}

	return commission, nil
}

func (r *BillingRepository) GetAffiliateCommissionByID(ctx context.Context, id uuid.UUID) (*AffiliateCommission, error) {
	query := `SELECT id, affiliate_code_id, referral_id, commission_type, commission_value, base_amount_cents, base_amount_usd, commission_cents, commission_usd, status, paid_at, payment_batch_id, payment_batch, subscription_id, notes, created_at, updated_at
			  FROM affiliate_commissions WHERE id = $1`

	commission := &AffiliateCommission{}
	err := r.db.QueryRow(query, id).Scan(&commission.ID, &commission.AffiliateCodeID, &commission.ReferralID,
		&commission.CommissionType, &commission.CommissionValue, &commission.BaseAmountCents, &commission.BaseAmountUSD,
		&commission.CommissionCents, &commission.CommissionUSD, &commission.Status, &commission.PaidAt,
		&commission.PaymentBatchID, &commission.PaymentBatch, &commission.SubscriptionID, &commission.Notes,
		&commission.CreatedAt, &commission.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliate commission: %w", err)
	}

	return commission, nil
}

func (r *BillingRepository) ListAffiliateCommissionsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateCommission, error) {
	query := `SELECT id, affiliate_code_id, referral_id, commission_type, commission_value, base_amount_cents, base_amount_usd, commission_cents, commission_usd, status, paid_at, payment_batch_id, payment_batch, subscription_id, notes, created_at, updated_at
			  FROM affiliate_commissions WHERE affiliate_code_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, codeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliate commissions: %w", err)
	}
	defer rows.Close()

	var commissions []*AffiliateCommission
	for rows.Next() {
		commission := &AffiliateCommission{}
		err := rows.Scan(&commission.ID, &commission.AffiliateCodeID, &commission.ReferralID,
			&commission.CommissionType, &commission.CommissionValue, &commission.BaseAmountCents, &commission.BaseAmountUSD,
			&commission.CommissionCents, &commission.CommissionUSD, &commission.Status, &commission.PaidAt,
			&commission.PaymentBatchID, &commission.PaymentBatch, &commission.SubscriptionID, &commission.Notes,
			&commission.CreatedAt, &commission.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affiliate commission: %w", err)
		}
		commissions = append(commissions, commission)
	}

	return commissions, nil
}

func (r *BillingRepository) ListAffiliateCommissionsByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCommission, error) {
	query := `SELECT ac.id, ac.affiliate_code_id, ac.referral_id, ac.commission_type, ac.commission_value, ac.base_amount_cents, ac.base_amount_usd, ac.commission_cents, ac.commission_usd, ac.status, ac.paid_at, ac.payment_batch_id, ac.payment_batch, ac.subscription_id, ac.notes, ac.created_at, ac.updated_at
			  FROM affiliate_commissions ac
			  INNER JOIN affiliate_codes acode ON ac.affiliate_code_id = acode.id
			  WHERE acode.publisher_id = $1
			  ORDER BY ac.created_at DESC`

	rows, err := r.db.Query(query, publisherID)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliate commissions: %w", err)
	}
	defer rows.Close()

	var commissions []*AffiliateCommission
	for rows.Next() {
		commission := &AffiliateCommission{}
		err := rows.Scan(&commission.ID, &commission.AffiliateCodeID, &commission.ReferralID,
			&commission.CommissionType, &commission.CommissionValue, &commission.BaseAmountCents, &commission.BaseAmountUSD,
			&commission.CommissionCents, &commission.CommissionUSD, &commission.Status, &commission.PaidAt,
			&commission.PaymentBatchID, &commission.PaymentBatch, &commission.SubscriptionID, &commission.Notes,
			&commission.CreatedAt, &commission.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affiliate commission: %w", err)
		}
		commissions = append(commissions, commission)
	}

	return commissions, nil
}

func (r *BillingRepository) UpdateAffiliateCommissionStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	var query string
	var args []interface{}

	switch status {
	case CommissionStatusPaid:
		query = `UPDATE affiliate_commissions SET status = $2, paid_at = $3, updated_at = $4 WHERE id = $1`
		args = []interface{}{id, status, now, now}
	case CommissionStatusCanceled:
		query = `UPDATE affiliate_commissions SET status = $2, updated_at = $3 WHERE id = $1`
		args = []interface{}{id, status, now}
	default:
		query = `UPDATE affiliate_commissions SET status = $2, updated_at = $3 WHERE id = $1`
		args = []interface{}{id, status, now}
	}

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update commission status: %w", err)
	}
	return nil
}

func (r *BillingRepository) CalculateCommission(ctx context.Context, commissionType string, commissionValue, baseAmountUSD float64) (commissionCents int64, commissionUSD float64) {
	switch commissionType {
	case "percent":
		commissionUSD = baseAmountUSD * (commissionValue / 100.0)
	case "fixed":
		commissionUSD = commissionValue
	default:
		commissionUSD = 0
	}
	commissionCents = int64(commissionUSD * 100)
	return commissionCents, commissionUSD
}
// Status constant aliases from the types package.
const CommissionStatusCanceled = types.CommissionStatusCanceled
const CommissionStatusPaid = types.CommissionStatusPaid
