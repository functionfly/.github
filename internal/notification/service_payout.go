package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SendPayoutCompleted notifies a user that their payout was processed successfully.
func (s *Service) SendPayoutCompleted(ctx context.Context, userID uuid.UUID, amountUSD float64, payoutRequestID string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypePayoutCompleted,
		Category: CategoryBilling,
		Title:    "Payout Completed",
		Body:     fmt.Sprintf("Your payout of $%.2f has been processed successfully and will arrive in your bank account shortly.", amountUSD),
		Data: JSONMap{
			"amount_usd":        amountUSD,
			"payout_request_id": payoutRequestID,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendPayoutFailed notifies a user that their payout failed.
func (s *Service) SendPayoutFailed(ctx context.Context, userID uuid.UUID, amountUSD float64, payoutRequestID, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypePayoutFailed,
		Category: CategoryBilling,
		Title:    "Payout Failed",
		Body:     fmt.Sprintf("Your payout of $%.2f could not be processed: %s. The funds have been returned to your balance.", amountUSD, reason),
		Data: JSONMap{
			"amount_usd":        amountUSD,
			"payout_request_id": payoutRequestID,
			"reason":            reason,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendPayoutCancelled notifies a user that their payout was cancelled.
func (s *Service) SendPayoutCancelled(ctx context.Context, userID uuid.UUID, amountUSD float64, payoutRequestID, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypePayoutCancelled,
		Category: CategoryBilling,
		Title:    "Payout Cancelled",
		Body:     "Your payout request has been cancelled and funds have been returned to your balance.",
		Data: JSONMap{
			"amount_usd":        amountUSD,
			"payout_request_id": payoutRequestID,
			"reason":            reason,
		},
		Channels: []string{ChannelInApp},
		Priority: PriorityNormal,
	})
	return err
}

// SendPayoutReversed notifies a user that a completed payout was reversed.
func (s *Service) SendPayoutReversed(ctx context.Context, userID uuid.UUID, amountUSD float64, stripeTransferID string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypePayoutReversed,
		Category: CategoryBilling,
		Title:    "Payout Reversed",
		Body:     fmt.Sprintf("A payout of $%.2f has been reversed by the payment processor. The funds have been returned to your balance.", amountUSD),
		Data: JSONMap{
			"amount_usd":         amountUSD,
			"stripe_transfer_id": stripeTransferID,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}
