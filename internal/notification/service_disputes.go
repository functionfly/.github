package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SendDisputeCreated notifies admins when a new payment dispute is created
func (s *Service) SendDisputeCreated(ctx context.Context, adminUserIDs []uuid.UUID, disputeID, amountUSD, currency, reason, evidenceDueBy string) error {
	for _, userID := range adminUserIDs {
		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeBillingDisputeCreated,
			Category: CategoryBilling,
			Title:    "New Payment Dispute",
			Body:     fmt.Sprintf("A chargeback dispute of %s %s was filed. Evidence due: %s. Reason: %s.", amountUSD, currency, evidenceDueBy, reason),
			Data: JSONMap{
				"dispute_id":      disputeID,
				"amount_usd":      amountUSD,
				"currency":        currency,
				"reason":          reason,
				"evidence_due_by": evidenceDueBy,
			},
			Channels: []string{ChannelInApp, ChannelEmail, ChannelWebhook},
			Priority: PriorityUrgent,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send dispute created notification")
		}
	}
	return nil
}

// SendDisputeEvidenceDueSoon warns admins that evidence submission deadline is approaching
func (s *Service) SendDisputeEvidenceDueSoon(ctx context.Context, adminUserIDs []uuid.UUID, disputeID string, daysRemaining int) error {
	urgency := "today"
	if daysRemaining > 1 {
		urgency = fmt.Sprintf("in %d days", daysRemaining)
	}

	for _, userID := range adminUserIDs {
		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeBillingDisputeEvidenceDue,
			Category: CategoryBilling,
			Title:    fmt.Sprintf("URGENT: Dispute Evidence Due %s", urgency),
			Body:     fmt.Sprintf("Evidence for dispute %s must be submitted %s to avoid an automatic loss.", disputeID, urgency),
			Data: JSONMap{
				"dispute_id":     disputeID,
				"days_remaining": daysRemaining,
			},
			Channels: []string{ChannelInApp, ChannelEmail, ChannelWebhook},
			Priority: PriorityUrgent,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send dispute evidence due notification")
		}
	}
	return nil
}

// SendDisputeResolved notifies admins about a dispute resolution
func (s *Service) SendDisputeResolved(ctx context.Context, adminUserIDs []uuid.UUID, disputeID, outcome string, amountUSD float64, won bool) error {
	outcomeText := "lost"
	if won {
		outcomeText = "won"
	}

	for _, userID := range adminUserIDs {
		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeBillingDisputeResolved,
			Category: CategoryBilling,
			Title:    fmt.Sprintf("Dispute %s: %s", outcomeText, disputeID),
			Body:     fmt.Sprintf("The dispute %s has been resolved. Outcome: %s. Amount: $%.2f.", disputeID, outcome, amountUSD),
			Data: JSONMap{
				"dispute_id": disputeID,
				"outcome":    outcome,
				"amount_usd": amountUSD,
				"won":        won,
			},
			Channels: []string{ChannelInApp, ChannelEmail},
			Priority: PriorityHigh,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send dispute resolved notification")
		}
	}
	return nil
}

// SendRefundProcessed notifies admins about a processed refund
func (s *Service) SendRefundProcessed(ctx context.Context, adminUserIDs []uuid.UUID, refundID string, amountUSD float64, reason string, tenantID *string) error {
	for _, userID := range adminUserIDs {
		data := JSONMap{
			"refund_id":  refundID,
			"amount_usd": amountUSD,
			"reason":     reason,
		}
		if tenantID != nil {
			data["tenant_id"] = *tenantID
		}

		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeBillingRefundProcessed,
			Category: CategoryBilling,
			Title:    fmt.Sprintf("Refund Processed: $%.2f", amountUSD),
			Body:     fmt.Sprintf("A refund of $%.2f has been processed. Reason: %s.", amountUSD, reason),
			Data:     data,
			Channels: []string{ChannelInApp, ChannelWebhook},
			Priority: PriorityNormal,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send refund processed notification")
		}
	}
	return nil
}

// SendChargebackFundsWithdrawn notifies admins when funds are withdrawn due to a lost chargeback
func (s *Service) SendChargebackFundsWithdrawn(ctx context.Context, adminUserIDs []uuid.UUID, disputeID string, amountUSD float64) error {
	for _, userID := range adminUserIDs {
		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeBillingChargebackFundsWithdrawn,
			Category: CategoryBilling,
			Title:    fmt.Sprintf("Chargeback Funds Withdrawn: $%.2f", amountUSD),
			Body:     fmt.Sprintf("Funds of $%.2f were withdrawn from your account due to lost chargeback %s. A $15 dispute fee also applies.", amountUSD, disputeID),
			Data: JSONMap{
				"dispute_id":      disputeID,
				"amount_usd":      amountUSD,
				"dispute_fee_usd": 15.0,
			},
			Channels: []string{ChannelInApp, ChannelEmail, ChannelWebhook},
			Priority: PriorityHigh,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send chargeback funds withdrawn notification")
		}
	}
	return nil
}
