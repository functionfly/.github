package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SendFounderModeThresholdWarning notifies a user they're approaching a founder mode threshold
func (s *Service) SendFounderModeThresholdWarning(ctx context.Context, userID uuid.UUID, bundleName string, progressPercent float64, threshold string, current int) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFounderModeThresholdWarning,
		Category: CategoryBilling,
		Title:    fmt.Sprintf("🚀 %s: You're Building Momentum!", bundleName),
		Body:     fmt.Sprintf("You're at %d%% of the %s threshold for your %s bundle (%s). Great progress! Consider converting to paid to keep all features.", int(progressPercent), threshold, bundleName, formatMetric(threshold, current)),
		Data: JSONMap{
			"bundle_name":      bundleName,
			"progress_percent": progressPercent,
			"threshold_type":   threshold,
			"current_value":    current,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendFounderModeThresholdReached notifies a user they've hit a founder mode threshold and grace period has started
func (s *Service) SendFounderModeThresholdReached(ctx context.Context, userID uuid.UUID, bundleName string, threshold string, gracePeriodDays int) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFounderModeThresholdReached,
		Category: CategoryBilling,
		Title:    fmt.Sprintf("🎉 %s Threshold Reached!", bundleName),
		Body:     fmt.Sprintf("Congratulations! You've hit the %s threshold for your %s bundle. You have %d days to convert to a paid subscription before the grace period ends.", threshold, bundleName, gracePeriodDays),
		Data: JSONMap{
			"bundle_name":       bundleName,
			"threshold_type":    threshold,
			"grace_period_days": gracePeriodDays,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendFounderModeGracePeriodEnding warns that grace period is ending soon
func (s *Service) SendFounderModeGracePeriodEnding(ctx context.Context, userID uuid.UUID, bundleName string, daysLeft int) error {
	urgency := "soon"
	priority := PriorityNormal
	if daysLeft <= 1 {
		urgency = "within 24 hours"
		priority = PriorityUrgent
	} else if daysLeft <= 3 {
		urgency = fmt.Sprintf("in %d days", daysLeft)
		priority = PriorityHigh
	}

	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFounderModeGracePeriodEnding,
		Category: CategoryBilling,
		Title:    fmt.Sprintf("⚠️ %s: Grace Period Ending %s", bundleName, urgency),
		Body:     fmt.Sprintf("Your %s grace period ends %s. Please convert to a paid subscription to keep all bundle features and data access.", bundleName, urgency),
		Data: JSONMap{
			"bundle_name": bundleName,
			"days_left":   daysLeft,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: priority,
	})
	return err
}

// SendFounderModeConverted confirms successful conversion from founder mode to paid
func (s *Service) SendFounderModeConverted(ctx context.Context, userID uuid.UUID, bundleName string, priceUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFounderModeConverted,
		Category: CategoryBilling,
		Title:    fmt.Sprintf("✅ %s Subscription Active", bundleName),
		Body:     fmt.Sprintf("Your %s bundle has been successfully converted to a paid subscription at $%.2f/month. You now have full access to all features!", bundleName, priceUSD),
		Data: JSONMap{
			"bundle_name": bundleName,
			"price_usd":   priceUSD,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

func formatMetric(thresholdType string, value int) string {
	switch thresholdType {
	case "user_count":
		return fmt.Sprintf("%d users", value)
	case "mrr_cents", "revenue_cents":
		return fmt.Sprintf("$%.2f MRR", float64(value)/100)
	case "api_calls":
		return fmt.Sprintf("%d API calls", value)
	case "days_elapsed":
		return fmt.Sprintf("%d days", value)
	default:
		return fmt.Sprintf("%d", value)
	}
}
