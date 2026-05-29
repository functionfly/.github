package studio

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/functionfly/functionfly/internal/payment"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/transfer"
)

// RequestConnectPayout transfers pending creator earnings to the tenant's Stripe Connect account.
func (r *MarketplaceRepository) RequestConnectPayout(ctx context.Context, tenantID, userID string) (string, error) {
	if !payment.IsConfigured() {
		return "", fmt.Errorf("Stripe is not configured")
	}

	pendingCents, err := r.sumPendingPublisherEarningsCents(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if pendingCents <= 0 {
		return "", fmt.Errorf("no pending royalties to pay out")
	}

	connectAccountID, err := r.getTenantConnectAccountID(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if connectAccountID == "" {
		return "", fmt.Errorf("complete Stripe Connect onboarding in Settings → Payouts before requesting a payout")
	}

	idempotencyKey := uuid.NewSHA1(uuid.NameSpaceOID, []byte("marketplace-payout:"+tenantID+":"+userID)).String()

	transferParams := &stripe.TransferParams{
		Amount:      stripe.Int64(int64(pendingCents)),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Destination: stripe.String(connectAccountID),
		Metadata: map[string]string{
			"tenant_id": tenantID,
			"purpose":   "marketplace_creator_payout",
		},
	}
	transferParams.IdempotencyKey = stripe.String(idempotencyKey)

	stripeTransfer, err := transfer.New(transferParams)
	if err != nil {
		return "", fmt.Errorf("stripe transfer failed: %w", err)
	}

	payoutID := uuid.New().String()
	amountUSD := float64(pendingCents) / 100.0
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO marketplace_payout_requests (id, tenant_id, amount, currency, status, requested_by_user_id, notes)
		VALUES ($1::uuid, $2::uuid, $3, 'USD', 'processing', NULLIF($4, '')::uuid, $5)`,
		payoutID, tenantID, amountUSD, userID, "stripe_transfer:"+stripeTransfer.ID,
	)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return "", fmt.Errorf("record payout request: %w", err)
	}

	if markErr := r.markPublisherEarningsWithdrawn(ctx, tenantID, stripeTransfer.ID); markErr != nil {
		return stripeTransfer.ID, fmt.Errorf("transfer succeeded but failed to update earnings: %w", markErr)
	}

	return stripeTransfer.ID, nil
}

func (r *MarketplaceRepository) sumPendingPublisherEarningsCents(ctx context.Context, tenantID string) (int, error) {
	var cents sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(net_amount_cents), 0)::bigint
		FROM publisher_earnings
		WHERE tenant_id = $1::uuid
		  AND status IN ('pending', 'available')`,
		tenantID,
	).Scan(&cents)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			_, pending, _, listErr := r.ListRoyalties(ctx, tenantID)
			if listErr != nil {
				return 0, listErr
			}
			return int(pending * 100), nil
		}
		return 0, fmt.Errorf("sum pending earnings: %w", err)
	}
	return int(cents.Int64), nil
}

func (r *MarketplaceRepository) getTenantConnectAccountID(ctx context.Context, tenantID string) (string, error) {
	var accountID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT stripe_connect_account_id
		FROM tenant_stripe_configs
		WHERE tenant_id = $1::uuid`,
		tenantID,
	).Scan(&accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		if strings.Contains(err.Error(), "does not exist") {
			return "", nil
		}
		return "", fmt.Errorf("load connect account: %w", err)
	}
	if accountID.Valid {
		return accountID.String, nil
	}
	return "", nil
}

func (r *MarketplaceRepository) markPublisherEarningsWithdrawn(ctx context.Context, tenantID, stripeTransferID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE publisher_earnings
		SET status = 'withdrawn',
		    stripe_payout_id = $2,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid
		  AND status IN ('pending', 'available')`,
		tenantID, stripeTransferID,
	)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}
