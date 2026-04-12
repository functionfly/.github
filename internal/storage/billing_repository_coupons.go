package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateCoupon creates a new coupon
func (r *BillingRepository) CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	coupon.ID = uuid.New()
	coupon.CreatedAt = time.Now()
	coupon.UpdatedAt = time.Now()
	coupon.TimesRedeemed = 0

	query := `
		INSERT INTO coupons (id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at`

	err := r.db.QueryRow(query, coupon.ID, coupon.Code, coupon.Name, coupon.Description,
		coupon.DiscountType, coupon.DiscountValue, coupon.MaxRedemptions, coupon.TimesRedeemed,
		coupon.ValidFrom, coupon.ValidUntil, coupon.IsActive, coupon.CreatedAt, coupon.UpdatedAt).Scan(
		&coupon.ID, &coupon.Code, &coupon.Name, &coupon.Description, &coupon.DiscountType,
		&coupon.DiscountValue, &coupon.MaxRedemptions, &coupon.TimesRedeemed,
		&coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}

	return coupon, nil
}

// ListCoupons lists all coupons
func (r *BillingRepository) ListCoupons() ([]*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list coupons: %w", err)
	}
	defer rows.Close()

	var coupons []*Coupon
	for rows.Next() {
		coupon := &Coupon{}
		err := rows.Scan(&coupon.ID, &coupon.Code, &coupon.Name, &coupon.Description,
			&coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
			&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil,
			&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coupon: %w", err)
		}
		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// GetCouponByCode retrieves a coupon by code
func (r *BillingRepository) GetCouponByCode(code string) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE UPPER(code) = UPPER($1) AND is_active = true`

	coupon := &Coupon{}
	err := r.db.QueryRow(query, code).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
		&coupon.Description, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
		&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive,
		&coupon.CreatedAt, &coupon.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return coupon, nil
}

// GetCouponByID retrieves a coupon by ID
func (r *BillingRepository) GetCouponByID(id uuid.UUID) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE id = $1`

	coupon := &Coupon{}
	err := r.db.QueryRow(query, id).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
		&coupon.Description, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
		&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive,
		&coupon.CreatedAt, &coupon.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return coupon, nil
}

// RedeemCoupon redeems a coupon for a tenant
func (r *BillingRepository) RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error) {
	// Check if coupon exists and is valid
	coupon, err := r.GetCouponByID(couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	// Check if already redeemed by this tenant
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND tenant_id = $2", couponID, tenantID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check redemption: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("coupon already redeemed by this tenant")
	}

	// Check redemption limits
	if coupon.MaxRedemptions != nil && coupon.TimesRedeemed >= *coupon.MaxRedemptions {
		return nil, fmt.Errorf("coupon redemption limit exceeded")
	}

	// Create redemption record
	redemption := &CouponRedemption{
		ID:             uuid.New(),
		CouponID:       couponID,
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		RedeemedAt:     time.Now(),
		Coupon:         coupon,
	}

	query := `INSERT INTO coupon_redemptions (id, coupon_id, tenant_id, subscription_id, redeemed_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err = r.db.Exec(query, redemption.ID, redemption.CouponID, redemption.TenantID,
		redemption.SubscriptionID, redemption.RedeemedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	// Update coupon redemption count
	_, err = r.db.Exec("UPDATE coupons SET times_redeemed = times_redeemed + 1 WHERE id = $1", couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to update coupon count: %w", err)
	}

	return redemption, nil
}
