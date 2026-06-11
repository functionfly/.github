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

	err := r.db.QueryRowContext(ctx, query, coupon.ID, coupon.Code, coupon.Name, coupon.Description,
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
func (r *BillingRepository) ListCoupons(ctx context.Context) ([]*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
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
func (r *BillingRepository) GetCouponByCode(ctx context.Context, code string) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE UPPER(code) = UPPER($1) AND is_active = true`

	coupon := &Coupon{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
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
func (r *BillingRepository) GetCouponByID(ctx context.Context, id uuid.UUID) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE id = $1`

	coupon := &Coupon{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
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
	coupon, err := r.GetCouponByID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	var count int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND tenant_id = $2", couponID, tenantID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check redemption: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("coupon already redeemed by this tenant")
	}

	if coupon.MaxRedemptions != nil && coupon.TimesRedeemed >= *coupon.MaxRedemptions {
		return nil, fmt.Errorf("coupon redemption limit exceeded")
	}

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

	_, err = r.db.ExecContext(ctx, query, redemption.ID, redemption.CouponID, redemption.TenantID,
		redemption.SubscriptionID, redemption.RedeemedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	_, err = r.db.ExecContext(ctx, "UPDATE coupons SET times_redeemed = times_redeemed + 1 WHERE id = $1", couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to update coupon count: %w", err)
	}

	return redemption, nil
}