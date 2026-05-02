package payment

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/transfer"
)

// VelocityLimits defines configurable payout velocity limits.
type VelocityLimits struct {
	MaxPayoutsPerDay   int   // Maximum number of payouts per 24-hour window
	MaxAmountPerDay    int   // Maximum total amount in cents per 24-hour window
	MinIntervalMinutes int   // Minimum minutes between payouts
}

// DefaultVelocityLimits returns sensible defaults.
func DefaultVelocityLimits() VelocityLimits {
	return VelocityLimits{
		MaxPayoutsPerDay:   getEnvInt("PAYOUT_MAX_PER_DAY", 5),
		MaxAmountPerDay:    getEnvInt("PAYOUT_MAX_AMOUNT_PER_DAY_CENTS", 1000000), // $10,000
		MinIntervalMinutes: getEnvInt("PAYOUT_MIN_INTERVAL_MINUTES", 60),
	}
}

// PayoutFeeResult holds the calculated fee for a payout.
type PayoutFeeResult struct {
	GrossAmountCents int     `json:"gross_amount_cents"`
	FeeAmountCents   int     `json:"fee_amount_cents"`
	NetAmountCents   int     `json:"net_amount_cents"`
	FeeType          string  `json:"fee_type"`
	FeeRate          float64 `json:"fee_rate"`
}

// PayoutServiceExtended wraps PayoutService with additional production features.
type PayoutServiceExtended struct {
	*PayoutService
	payoutRepo     *storage.PayoutRepository
	notifySvc      *notification.Service
	velocityLimits VelocityLimits
}

// NewPayoutServiceExtended creates the extended payout service.
func NewPayoutServiceExtended(
	payoutRepo *storage.PayoutRepository,
	notifySvc *notification.Service,
) *PayoutServiceExtended {
	return &PayoutServiceExtended{
		PayoutService:  NewPayoutService(payoutRepo),
		payoutRepo:     payoutRepo,
		notifySvc:      notifySvc,
		velocityLimits: DefaultVelocityLimits(),
	}
}

// RequestPayoutWithChecks performs a full payout request with fee calculation,
// velocity checks, and notifications.
func (s *PayoutServiceExtended) RequestPayoutWithChecks(
	ctx context.Context,
	userID uuid.UUID,
	amountCents int,
	idempotencyKey string,
	feeType string,
) (*PayoutResult, *PayoutFeeResult, error) {
	if stripeKey() == "" {
		return nil, nil, fmt.Errorf("Stripe is not configured")
	}

	if amountCents < MinPayoutAmountCents {
		return nil, nil, fmt.Errorf("minimum payout is $%.2f", float64(MinPayoutAmountCents)/100.0)
	}
	if amountCents > MaxPayoutAmountCents {
		return nil, nil, fmt.Errorf("maximum payout is $%.2f", float64(MaxPayoutAmountCents)/100.0)
	}

	// Velocity checks
	if err := s.checkVelocityLimits(ctx, userID, amountCents); err != nil {
		return nil, nil, err
	}

	// Calculate fees
	feeResult, err := s.calculateFee(ctx, amountCents, feeType)
	if err != nil {
		logrus.WithError(err).Warn("payout: failed to calculate fee, using zero fee")
		feeResult = &PayoutFeeResult{
			GrossAmountCents: amountCents,
			FeeAmountCents:   0,
			NetAmountCents:   amountCents,
			FeeType:          "none",
		}
	}

	// Check balance against gross amount (fee is deducted from the transfer, not the balance)
	connectAccount, err := s.payoutRepo.GetConnectAccountByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connect account: %w", err)
	}
	if connectAccount == nil {
		return nil, nil, fmt.Errorf("no connected account found; please complete onboarding first")
	}
	if !connectAccount.PayoutsEnabled {
		return nil, nil, fmt.Errorf("payouts are not enabled; please complete onboarding first")
	}

	balance, err := s.payoutRepo.GetPayoutBalance(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get balance: %w", err)
	}
	availableCents := int(balance.AvailableBalanceUSD * 100)
	if availableCents < amountCents {
		return nil, nil, fmt.Errorf("insufficient balance: available $%.2f, requested $%.2f",
			balance.AvailableBalanceUSD, float64(amountCents)/100.0)
	}

	// Create payout request
	payoutReq := &storage.PayoutRequest{
		UserID:           userID,
		ConnectAccountID: connectAccount.ID,
		AmountCents:      amountCents,
		Currency:         "usd",
		Status:           "pending",
		IdempotencyKey:   idempotencyKey,
	}
	if err := s.payoutRepo.CreatePayoutRequest(ctx, payoutReq); err != nil {
		return nil, nil, fmt.Errorf("failed to create payout request: %w", err)
	}

	// Debit the user's ledger balance
	if err := s.payoutRepo.DebitForPayout(ctx, userID, amountCents, payoutReq.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to debit balance: %w", err)
	}

	// Record fee deduction if applicable
	if feeResult.FeeAmountCents > 0 {
		_ = s.payoutRepo.RecordFeeDeduction(ctx, nil, &storage.PayoutFeeDeduction{
			PayoutRequestID:  payoutReq.ID,
			UserID:           userID,
			GrossAmountCents: feeResult.GrossAmountCents,
			FeeAmountCents:   feeResult.FeeAmountCents,
			NetAmountCents:   feeResult.NetAmountCents,
			FeeType:          feeResult.FeeType,
			FeeRate:          &feeResult.FeeRate,
			Currency:         "usd",
		})
	}

	// Update status to processing
	if err := s.payoutRepo.UpdatePayoutRequestStatus(ctx, payoutReq.ID, "processing", nil, nil); err != nil {
		logrus.WithError(err).Error("failed to update payout status to processing")
	}

	// Create Stripe Transfer for the net amount
	transferAmount := feeResult.NetAmountCents
	transferParams := &stripe.TransferParams{
		Amount:      stripe.Int64(int64(transferAmount)),
		Currency:    stripe.String("usd"),
		Destination: stripe.String(connectAccount.StripeAccountID),
	}
	transferParams.IdempotencyKey = stripe.String(idempotencyKey)

	stripeTransfer, err := transfer.New(transferParams)
	if err != nil {
		reverseErr := s.payoutRepo.ReversePayoutDebit(ctx, userID, amountCents, payoutReq.ID,
			fmt.Sprintf("Transfer failed: %s", err.Error()))
		if reverseErr != nil {
			logrus.WithError(reverseErr).Error("CRITICAL: failed to reverse payout debit after transfer failure")
		}
		_ = s.payoutRepo.MarkPayoutRequestFailed(ctx, payoutReq.ID, err.Error())

	if s.notifySvc != nil {
		_, _ = s.notifySvc.Send(ctx, notification.SendRequest{
			UserID:   userID,
			Type:     "payout.failed",
			Category: notification.CategoryBilling,
			Title:    "Payout Failed",
			Body:     fmt.Sprintf("Your payout of $%.2f has failed: %s", float64(amountCents)/100.0, err.Error()),
			Data: map[string]interface{}{
				"payout_request_id": payoutReq.ID.String(),
				"amount_cents":      amountCents,
				"error":             err.Error(),
			},
		})
		}

		return nil, nil, fmt.Errorf("stripe transfer failed: %w", err)
	}

	// Record velocity
	_, _ = s.payoutRepo.RecordPayoutVelocity(ctx, userID, amountCents)

	// Mark completed
	transferID := stripeTransfer.ID
	if err := s.payoutRepo.UpdatePayoutRequestStatus(ctx, payoutReq.ID, "completed", &transferID, nil); err != nil {
		logrus.WithError(err).Error("failed to update payout request after successful transfer")
	}

	// Send success notification
	if s.notifySvc != nil {
	_, _ = s.notifySvc.Send(ctx, notification.SendRequest{
		UserID:   userID,
		Type:     "payout.completed",
		Category: notification.CategoryBilling,
		Title:    "Payout Completed",
		Body:     fmt.Sprintf("Your payout of $%.2f has been processed successfully.", float64(amountCents)/100.0),
		Data: map[string]interface{}{
			"payout_request_id": payoutReq.ID.String(),
			"amount_cents":      amountCents,
			"fee_cents":         feeResult.FeeAmountCents,
			"net_cents":         feeResult.NetAmountCents,
			"transfer_id":       transferID,
		},
	})
	}

	return &PayoutResult{
		PayoutRequestID: payoutReq.ID,
		AmountCents:     amountCents,
		Currency:        "usd",
		Status:          "completed",
	}, feeResult, nil
}

// CancelPayout cancels a pending payout request and reverses the ledger debit.
func (s *PayoutServiceExtended) CancelPayout(ctx context.Context, userID, payoutID uuid.UUID, reason string) error {
	if reason == "" {
		reason = "Cancelled by user"
	}

	if err := s.payoutRepo.CancelPayoutRequest(ctx, payoutID, userID, reason); err != nil {
		return fmt.Errorf("failed to cancel payout: %w", err)
	}

	if s.notifySvc != nil {
		_, _ = s.notifySvc.Send(ctx, notification.SendRequest{
			UserID:   userID,
			Type:     "payout.cancelled",
			Category: notification.CategoryBilling,
			Title:    "Payout Cancelled",
			Body:     "Your payout request has been cancelled and funds have been returned to your balance.",
			Data: map[string]interface{}{
				"payout_request_id": payoutID.String(),
				"reason":            reason,
			},
		})
	}

	return nil
}

// checkVelocityLimits enforces payout frequency and amount limits.
func (s *PayoutServiceExtended) checkVelocityLimits(ctx context.Context, userID uuid.UUID, amountCents int) error {
	record, err := s.payoutRepo.GetVelocityRecord(ctx, userID)
	if err != nil {
		logrus.WithError(err).Warn("payout: failed to check velocity, allowing")
		return nil
	}
	if record == nil {
		return nil // No payouts yet today
	}

	if record.PayoutCount >= s.velocityLimits.MaxPayoutsPerDay {
		return fmt.Errorf("daily payout limit reached (%d per day); please try again tomorrow",
			s.velocityLimits.MaxPayoutsPerDay)
	}

	if record.TotalAmountCents+amountCents > s.velocityLimits.MaxAmountPerDay {
		return fmt.Errorf("daily payout amount limit reached ($%.2f per day); please try again tomorrow",
			float64(s.velocityLimits.MaxAmountPerDay)/100.0)
	}

	if record.LastPayoutAt != nil {
		minInterval := time.Duration(s.velocityLimits.MinIntervalMinutes) * time.Minute
		if time.Since(*record.LastPayoutAt) < minInterval {
			remaining := minInterval - time.Since(*record.LastPayoutAt)
			return fmt.Errorf("please wait %s between payout requests", remaining.Round(time.Minute))
		}
	}

	return nil
}

// calculateFee computes the fee for a payout based on the fee configuration.
func (s *PayoutServiceExtended) calculateFee(ctx context.Context, amountCents int, feeType string) (*PayoutFeeResult, error) {
	if feeType == "" {
		feeType = "standard"
	}

	cfg, err := s.payoutRepo.GetActiveFeeConfig(ctx, feeType)
	if err != nil || cfg == nil {
		// Default to zero fee
		return &PayoutFeeResult{
			GrossAmountCents: amountCents,
			FeeAmountCents:   0,
			NetAmountCents:   amountCents,
			FeeType:          feeType,
		}, nil
	}

	feeCents := 0
	switch cfg.FeeType {
	case "percentage":
		feeCents = int(float64(amountCents) * cfg.FeePercent / 100.0)
	case "flat":
		feeCents = cfg.FlatFeeCents
	}

	// Apply minimum fee floor
	if cfg.MinimumFeeCents > 0 && feeCents < cfg.MinimumFeeCents {
		feeCents = cfg.MinimumFeeCents
	}

	// Apply fee cap
	if cfg.MaximumFeeCents != nil && feeCents > *cfg.MaximumFeeCents {
		feeCents = *cfg.MaximumFeeCents
	}

	// Ensure fee doesn't exceed the payout amount
	if feeCents > amountCents {
		feeCents = amountCents
	}

	return &PayoutFeeResult{
		GrossAmountCents: amountCents,
		FeeAmountCents:   feeCents,
		NetAmountCents:   amountCents - feeCents,
		FeeType:          cfg.FeeType,
		FeeRate:          cfg.FeePercent,
	}, nil
}

// ProcessScheduledPayouts runs auto-payouts for users with schedule preferences enabled.
func (s *PayoutServiceExtended) ProcessScheduledPayouts(ctx context.Context) error {
	if stripeKey() == "" {
		return nil
	}

	prefs, err := s.payoutRepo.GetUsersWithScheduledPayouts(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get scheduled payouts: %w", err)
	}

	processed := 0
	for _, pref := range prefs {
		if err := s.processScheduledPayout(ctx, pref); err != nil {
			logrus.WithError(err).WithField("user_id", pref.UserID).Error("payout scheduler: failed to process")
			continue
		}
		processed++
	}

	if processed > 0 {
		logrus.Infof("Payout scheduler: processed %d scheduled payouts", processed)
	}
	return nil
}

func (s *PayoutServiceExtended) processScheduledPayout(ctx context.Context, pref *storage.PayoutSchedulePreference) error {
	balance, err := s.payoutRepo.GetPayoutBalance(ctx, pref.UserID)
	if err != nil {
		return err
	}

	availableCents := int(balance.AvailableBalanceUSD * 100)
	if availableCents < pref.MinimumAmountCents {
		logrus.WithFields(logrus.Fields{
			"user_id":  pref.UserID,
			"available": balance.AvailableBalanceUSD,
			"minimum":  float64(pref.MinimumAmountCents) / 100.0,
		}).Debug("payout scheduler: balance below minimum, skipping")
		return nil
	}

	if availableCents < MinPayoutAmountCents {
		return nil
	}

	idempotencyKey := fmt.Sprintf("auto-payout-%s-%s", pref.UserID.String(), time.Now().Format("20060102"))

	_, _, err = s.RequestPayoutWithChecks(ctx, pref.UserID, availableCents, idempotencyKey, "standard")
	if err != nil {
		return err
	}

	_ = s.payoutRepo.UpdateSchedulePreferenceAfterPayout(ctx, pref.UserID, time.Now())
	return nil
}

// ProcessTransferReversed handles a transfer.reversed Stripe webhook event.
func (s *PayoutServiceExtended) ProcessTransferReversedNotify(ctx context.Context, stripeTransferID string) error {
	if err := s.PayoutService.ProcessTransferReversed(ctx, stripeTransferID); err != nil {
		return err
	}

	// Find the payout request to get the user for notification
	var payoutReqID, userIDStr string
	var amountCents int
	if err := s.payoutRepo.FindPayoutRequestByTransferID(ctx, stripeTransferID, &payoutReqID, &userIDStr, &amountCents); err != nil {
		return nil // already processed
	}

	userID, _ := uuid.Parse(userIDStr)
	if s.notifySvc != nil && userID != uuid.Nil {
		_, _ = s.notifySvc.Send(ctx, notification.SendRequest{
			UserID:   userID,
			Type:     "payout.reversed",
			Category: notification.CategoryBilling,
			Title:    "Payout Reversed",
			Body:     fmt.Sprintf("A payout of $%.2f has been reversed by Stripe. The funds have been returned to your balance.", float64(amountCents)/100.0),
			Data: map[string]interface{}{
				"stripe_transfer_id": stripeTransferID,
				"amount_cents":       amountCents,
			},
		})
	}

	return nil
}

func getEnvInt(key string, defaultVal int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return v
}
