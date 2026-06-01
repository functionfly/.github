package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service is the main notification service
type Service struct {
	repo       Repository
	channels   map[string]Channel
	templates  *TemplateEngine
	queue      *Queue
	dispatcher *Dispatcher
	logger     *logrus.Logger
	db         *storage.PostgresDB
	mu         sync.RWMutex
}

// NewService creates a new notification service
func NewService(repo Repository, db *storage.PostgresDB, emailSvc email.Service, logger *logrus.Logger) *Service {
	s := &Service{
		repo:     repo,
		db:       db,
		logger:   logger,
		channels: make(map[string]Channel),
	}

	// Register channels
	emailChannel := NewEmailChannel(emailSvc, logger)
	inAppChannel := NewInAppChannel(repo, db, logger)
	webhookChannel := NewWebhookChannel(logger)
	webhookChannel.SetRepository(repo)

	s.channels[ChannelEmail] = emailChannel
	s.channels[ChannelInApp] = inAppChannel
	s.channels[ChannelWebhook] = webhookChannel

	// Initialize queue and dispatcher
	s.queue = NewQueue(repo, logger)
	s.dispatcher = NewDispatcher(s.channels, repo, db, logger)
	s.templates = NewTemplateEngine(repo)

	return s
}

// Send creates and sends a notification
func (s *Service) Send(ctx context.Context, req SendRequest) (*Notification, error) {
	// Determine channels if not specified
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{ChannelInApp, ChannelEmail}
	}

	// Determine priority if not specified
	priority := req.Priority
	if priority == "" {
		priority = PriorityNormal
	}

	// Build notification from request
	notification := &Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Category: req.Category,
		Title:    req.Title,
		Body:     req.Body,
		Data:     req.Data,
		Channels: channels,
		Priority: priority,
		Status:   StatusPending,
	}

	// Save to database
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"type":            notification.Type,
	}).Info("Notification created")

	// Add to processing queue
	s.queue.Enqueue(notification)

	return notification, nil
}

// Broadcast sends a notification to multiple users
func (s *Service) Broadcast(ctx context.Context, req BroadcastRequest) error {
	for _, userID := range req.UserIDs {
		sendReq := SendRequest{
			UserID:   userID,
			Type:     req.Type,
			Category: req.Category,
			Title:    req.Title,
			Body:     req.Body,
			Data:     req.Data,
			Channels: req.Channels,
			Priority: req.Priority,
		}
		if _, err := s.Send(ctx, sendReq); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send broadcast notification")
		}
	}
	return nil
}

// SendWelcome sends a welcome notification to a new user (in-app only so everyone sees it when they open the app).
func (s *Service) SendWelcome(ctx context.Context, userID uuid.UUID) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeWelcome,
		Category: CategorySystem,
		Title:    "Welcome to FunctionFly",
		Body:     "We're glad you're here. Deploy your first function or explore the docs to get started.",
		Channels: []string{ChannelInApp},
		Priority: PriorityNormal,
	})
	return err
}

// SendWalletTopUp notifies a user after credits are added to an agent wallet.
func (s *Service) SendWalletTopUp(ctx context.Context, userID uuid.UUID, agentID string, amountUSD, newBalanceUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletToppedUp,
		Category: CategoryBilling,
		Title:    "Wallet Top-Up Successful",
		Body:     fmt.Sprintf("Your wallet was topped up by $%.2f. New balance: $%.2f.", amountUSD, newBalanceUSD),
		Data: JSONMap{
			"agent_id":        agentID,
			"amount_usd":      amountUSD,
			"new_balance_usd": newBalanceUSD,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendRegistryWalletTopUp notifies a user after credits are added to their registry wallet.
func (s *Service) SendRegistryWalletTopUp(ctx context.Context, userID uuid.UUID, amountUSD, newBalanceUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletToppedUp,
		Category: CategoryBilling,
		Title:    "Registry Wallet Top-Up Successful",
		Body:     fmt.Sprintf("Your registry wallet was topped up by $%.2f. New balance: $%.2f.", amountUSD, newBalanceUSD),
		Data: JSONMap{
			"amount_usd":      amountUSD,
			"new_balance_usd": newBalanceUSD,
			"wallet_type":     "registry",
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendWalletLowBalance notifies a user when wallet balance drops below threshold.
func (s *Service) SendWalletLowBalance(ctx context.Context, userID uuid.UUID, agentID string, balanceUSD, thresholdUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    "Low Wallet Balance",
		Body: fmt.Sprintf(
			"Wallet balance is low ($%.2f). It is now at or below your alert threshold of $%.2f.",
			balanceUSD,
			thresholdUSD,
		),
		Data: JSONMap{
			"agent_id":      agentID,
			"balance_usd":   balanceUSD,
			"threshold_usd": thresholdUSD,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendBillingInvoiceGenerated sends an invoice generated notification when a new invoice is created.
func (s *Service) SendBillingInvoiceGenerated(ctx context.Context, userEmail, period string, amountDueUSD float64, invoiceURL, invoiceID string) error {
	// Look up the user by email to get their ID
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingInvoiceGenerated,
		Category: CategoryBilling,
		Title:    "New Invoice Available",
		Body:     fmt.Sprintf("Your invoice for %s ($%.2f) is ready for payment.", period, amountDueUSD),
		Data: JSONMap{
			"period":      period,
			"amount_usd":  amountDueUSD,
			"invoice_url": invoiceURL,
			"invoice_id":  invoiceID,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendBillingPaymentSuccess sends a payment success notification for subscription payments.
func (s *Service) SendBillingPaymentSuccess(ctx context.Context, userEmail, period string, amountPaidUSD float64, invoiceID string) error {
	// Look up the user by email to get their ID
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingPaymentSuccess,
		Category: CategoryBilling,
		Title:    "Payment Successful",
		Body:     fmt.Sprintf("Your payment of $%.2f for %s has been successfully processed. Thank you!", amountPaidUSD, period),
		Data: JSONMap{
			"period":       period,
			"amount_usd":   amountPaidUSD,
			"invoice_id":   invoiceID,
			"payment_date": time.Now().Format(time.RFC3339),
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendUpcomingRenewalNotice sends a notification about an upcoming subscription renewal.
func (s *Service) SendUpcomingRenewalNotice(ctx context.Context, userID uuid.UUID, period string, amountUSD float64, renewalDate time.Time, daysUntil int) error {
	var title, body string
	priority := PriorityNormal

	switch daysUntil {
	case 7:
		title = "Subscription Renewal in 1 Week"
		body = fmt.Sprintf("Your subscription will renew on %s for $%.2f. Ensure your payment method is up to date.", renewalDate.Format("Jan 2, 2006"), amountUSD)
	case 3:
		title = "Subscription Renewal in 3 Days"
		body = fmt.Sprintf("Your subscription renewal ($%.2f for %s) is approaching. Update your payment method if needed.", amountUSD, period)
		priority = PriorityHigh
	case 1:
		title = "Subscription Renews Tomorrow"
		body = fmt.Sprintf("Your subscription for %s ($%.2f) will renew tomorrow. Make sure your payment method is valid.", period, amountUSD)
		priority = PriorityHigh
	default:
		title = "Upcoming Subscription Renewal"
		body = fmt.Sprintf("Your subscription for %s ($%.2f) will renew on %s.", period, amountUSD, renewalDate.Format("Jan 2, 2006"))
	}

	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingSubscriptionExpiring,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data: JSONMap{
			"period":       period,
			"amount_usd":   amountUSD,
			"renewal_date": renewalDate.Format(time.RFC3339),
			"days_until":   daysUntil,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: priority,
	})
	return err
}

// SendBillingAlert sends a billing-related alert notification (e.g., payment failed, invoice overdue).
func (s *Service) SendBillingAlert(ctx context.Context, userEmail string, alertType string, data map[string]interface{}) error {
	// First, look up the user by email to get their ID
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	var title, body string
	priority := PriorityHigh

	switch alertType {
	case "payment_failed":
		amountDue, _ := data["amount_due"].(float64)
		currency, _ := data["currency"].(string)
		attemptCount, _ := data["attempt_count"].(int64)
		title = "Payment Failed"
		body = fmt.Sprintf("Your payment of %.2f %s failed (attempt %d). Please update your payment method to avoid service interruption.", amountDue, currency, attemptCount)
	case "invoice_overdue":
		title = "Invoice Overdue"
		body = "You have an overdue invoice. Please make a payment to avoid service suspension."
	case "subscription_cancelled":
		title = "Subscription Cancelled"
		body = "Your subscription has been cancelled. You will be downgraded to the free plan at the end of your billing period."
	default:
		title = "Billing Alert"
		body = "There is an issue with your billing account. Please review your payment settings."
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingAlert,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data:     data,
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: priority,
	})
	return err
}

// GetUnreadCount returns unread notification count for user
func (s *Service) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// GetUnreadCountsByCategory returns unread notification counts grouped by category
func (s *Service) GetUnreadCountsByCategory(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	return s.repo.GetUnreadCountsByCategory(ctx, userID)
}

// GetTotalCount returns total notification count for user
func (s *Service) GetTotalCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetTotalCount(ctx, userID)
}

// ListNotifications lists notifications for a user
func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error) {
	return s.repo.ListNotifications(ctx, userID, opts)
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
	return s.repo.GetNotification(ctx, id)
}

// DeleteNotification deletes a notification
func (s *Service) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteNotification(ctx, id)
}

// ArchiveNotification marks a notification as archived (excluded from unread counts).
func (s *Service) ArchiveNotification(ctx context.Context, id uuid.UUID) error {
	return s.repo.ArchiveNotification(ctx, id)
}

// GetPreferences retrieves all notification preferences for a user
func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If no preferences exist, create defaults
	if len(prefs) == 0 {
		if err := s.repo.CreateDefaultPreferences(ctx, userID); err != nil {
			return nil, err
		}
		return s.repo.GetPreferences(ctx, userID)
	}

	return prefs, nil
}

// SavePreference saves a notification preference
func (s *Service) SavePreference(ctx context.Context, pref *NotificationPreference) error {
	return s.repo.SavePreference(ctx, pref)
}

// RegisterChannel registers a custom channel
func (s *Service) RegisterChannel(name string, channel Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[name] = channel
}

// Start begins processing the notification queue
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting notification service")
	s.queue.Start(ctx, s.dispatcher)
}

// Stop stops the notification service
func (s *Service) Stop() {
	s.logger.Info("Stopping notification service")
	s.queue.Stop()
}



// SendLowBalance sends a low balance alert to a user via email
func (s *Service) SendLowBalance(ctx context.Context, userEmail string, data map[string]interface{}) error {
	// Look up the user by email to get their ID
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	thresholdUSD, _ := data["threshold_usd"].(float64)
	severity, _ := data["severity"].(string)
	autoTopupEnabled, _ := data["auto_topup_enabled"].(bool)

	title := "Low Wallet Balance"
	body := fmt.Sprintf("Your wallet balance is low ($%.2f USD). It is now at or below your alert threshold of $%.2f USD.", balanceUSD, thresholdUSD)

	if severity == "critical" {
		title = "Critical: Very Low Wallet Balance"
		body = fmt.Sprintf("URGENT: Your wallet balance is critically low ($%.2f USD). Add funds immediately to avoid service interruption.", balanceUSD)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data: JSONMap{
			"balance_usd":        balanceUSD,
			"threshold_usd":      thresholdUSD,
			"severity":           severity,
			"auto_topup_enabled": autoTopupEnabled,
			"currency":           data["currency"],
			"balance_local":      data["balance_local"],
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendLowBalanceNotification sends an in-app low balance notification
func (s *Service) SendLowBalanceNotification(ctx context.Context, userID interface{}, data map[string]interface{}) error {
	uid, ok := userID.(uuid.UUID)
	if !ok {
		if idStr, ok := userID.(string); ok {
			parsed, err := uuid.Parse(idStr)
			if err != nil {
				return fmt.Errorf("invalid user ID format: %w", err)
			}
			uid = parsed
		} else {
			return fmt.Errorf("userID must be uuid.UUID or string")
		}
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	thresholdUSD, _ := data["threshold_usd"].(float64)
	severity, _ := data["severity"].(string)

	title := "Low Wallet Balance"
	body := fmt.Sprintf("Your wallet balance ($%.2f) is below the threshold ($%.2f).", balanceUSD, thresholdUSD)

	if severity == "critical" {
		title = "Critical: Low Balance"
		body = fmt.Sprintf("URGENT: Your balance ($%.2f) is critically low. Add funds now.", balanceUSD)
	}

	_, err := s.Send(ctx, SendRequest{
		UserID:   uid,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data:     data,
		Channels: []string{ChannelInApp},
		Priority: PriorityHigh,
	})
	return err
}

// SendAutoTopupApproaching sends a notification that auto-topup threshold is approaching
func (s *Service) SendAutoTopupApproaching(ctx context.Context, userID interface{}, data map[string]interface{}) error {
	uid, ok := userID.(uuid.UUID)
	if !ok {
		if idStr, ok := userID.(string); ok {
			parsed, err := uuid.Parse(idStr)
			if err != nil {
				return fmt.Errorf("invalid user ID format: %w", err)
			}
			uid = parsed
		} else {
			return fmt.Errorf("userID must be uuid.UUID or string")
		}
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	autoTopupThreshold, _ := data["auto_topup_threshold"].(float64)

	_, err := s.Send(ctx, SendRequest{
		UserID:   uid,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    "Auto-Topup Approaching",
		Body:     fmt.Sprintf("Your balance ($%.2f) is approaching your auto-topup threshold ($%.2f). Funds will be added automatically soon.", balanceUSD, autoTopupThreshold),
		Data:     data,
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

// Team Notifications

// SendTeamCreated notifies a user when they successfully create a team
func (s *Service) SendTeamCreated(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, teamName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeTeamCreated,
		Category: CategoryTeam,
		Title:    fmt.Sprintf("Team '%s' Created", teamName),
		Body:     fmt.Sprintf("You have successfully created the team '%s'. You can now invite members and manage team resources.", teamName),
		Data: JSONMap{
			"team_id":   teamID.String(),
			"team_name": teamName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendTeamDeleted notifies team members when a team is deleted
func (s *Service) SendTeamDeleted(ctx context.Context, userIDs []uuid.UUID, teamName string, deletedByName string) error {
	for _, userID := range userIDs {
		_, err := s.Send(ctx, SendRequest{
			UserID:   userID,
			Type:     TypeTeamDeleted,
			Category: CategoryTeam,
			Title:    fmt.Sprintf("Team '%s' Deleted", teamName),
			Body:     fmt.Sprintf("The team '%s' has been deleted by %s. All team resources and access have been removed.", teamName, deletedByName),
			Data: JSONMap{
				"team_name":       teamName,
				"deleted_by_name": deletedByName,
			},
			Channels: []string{ChannelInApp, ChannelEmail},
			Priority: PriorityHigh,
		})
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to send team deletion notification")
		}
	}
	return nil
}

// SendTeamInviteSent notifies a user when they are invited to join a team
func (s *Service) SendTeamInviteSent(ctx context.Context, inviteeUserID uuid.UUID, teamID uuid.UUID, teamName string, invitedByName string, role string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   inviteeUserID,
		Type:     TypeTeamInviteSent,
		Category: CategoryTeam,
		Title:    fmt.Sprintf("Invitation to Join '%s'", teamName),
		Body:     fmt.Sprintf("%s has invited you to join the team '%s' with the role: %s. Accept or decline in your notifications.", invitedByName, teamName, role),
		Data: JSONMap{
			"team_id":         teamID.String(),
			"team_name":       teamName,
			"invited_by_name": invitedByName,
			"role":            role,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendTeamInviteAccepted notifies the inviter when their invitation is accepted
func (s *Service) SendTeamInviteAccepted(ctx context.Context, inviterUserID uuid.UUID, teamID uuid.UUID, teamName string, inviteeName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   inviterUserID,
		Type:     TypeTeamInviteAccepted,
		Category: CategoryTeam,
		Title:    fmt.Sprintf("%s Joined '%s'", inviteeName, teamName),
		Body:     fmt.Sprintf("%s has accepted your invitation and is now a member of the team '%s'.", inviteeName, teamName),
		Data: JSONMap{
			"team_id":      teamID.String(),
			"team_name":    teamName,
			"invitee_name": inviteeName,
		},
		Channels: []string{ChannelInApp},
		Priority: PriorityNormal,
	})
	return err
}

// SendTeamMemberAdded notifies a user when they are added to a team (direct add, not invite)
func (s *Service) SendTeamMemberAdded(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, teamName string, addedByName string, role string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeTeamMemberAdded,
		Category: CategoryTeam,
		Title:    fmt.Sprintf("Added to Team '%s'", teamName),
		Body:     fmt.Sprintf("%s has added you to the team '%s' with the role: %s. You now have access to team resources.", addedByName, teamName, role),
		Data: JSONMap{
			"team_id":       teamID.String(),
			"team_name":     teamName,
			"added_by_name": addedByName,
			"role":          role,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendTeamMemberRemoved notifies a user when they are removed from a team
func (s *Service) SendTeamMemberRemoved(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, teamName string, removedByName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeTeamMemberRemoved,
		Category: CategoryTeam,
		Title:    fmt.Sprintf("Removed from Team '%s'", teamName),
		Body:     fmt.Sprintf("You have been removed from the team '%s' by %s. You no longer have access to team resources.", teamName, removedByName),
		Data: JSONMap{
			"team_id":         teamID.String(),
			"team_name":       teamName,
			"removed_by_name": removedByName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendUsernameChanged notifies a user when their username has been successfully changed.
func (s *Service) SendUsernameChanged(ctx context.Context, userID uuid.UUID, oldUsername, newUsername string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeSecurityUsernameChanged,
		Category: CategorySecurity,
		Title:    "Username Changed Successfully",
		Body:     fmt.Sprintf("Your username has been changed from @%s to @%s.", oldUsername, newUsername),
		Data: JSONMap{
			"old_username": oldUsername,
			"new_username": newUsername,
			"changed_at":   time.Now().UTC(),
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

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

// SendDeploymentSuccess notifies a user that a deployment succeeded.
func (s *Service) SendDeploymentSuccess(ctx context.Context, userID uuid.UUID, appID, appName string, deployedAt int64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeDeploymentSuccess,
		Category: CategoryDeployment,
		Title:    fmt.Sprintf("Deployment Successful: %s", appName),
		Body:     fmt.Sprintf("Your deployment of %s was successful.", appName),
		Data: JSONMap{
			"app_id":      appID,
			"app_name":    appName,
			"deployed_at": deployedAt,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendDeploymentFailure notifies a user that a deployment failed.
func (s *Service) SendDeploymentFailure(ctx context.Context, userID uuid.UUID, appID, appName, errorMsg, logsURL string, failedAt int64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeDeploymentFailed,
		Category: CategoryDeployment,
		Title:    fmt.Sprintf("Deployment Failed: %s", appName),
		Body:     fmt.Sprintf("Your deployment of %s failed. %s", appName, errorMsg),
		Data: JSONMap{
			"app_id":        appID,
			"app_name":      appName,
			"error_message": errorMsg,
			"logs_url":      logsURL,
			"failed_at":     failedAt,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendFailoverTriggered notifies a user when failover is triggered for their function.
func (s *Service) SendFailoverTriggered(ctx context.Context, userID uuid.UUID, functionID, functionName, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFailoverTriggered,
		Category: CategoryFailover,
		Title:    fmt.Sprintf("Failover Triggered: %s", functionName),
		Body:     fmt.Sprintf("Failover was triggered for %s. Reason: %s.", functionName, reason),
		Data: JSONMap{
			"function_id": functionID,
			"function_name": functionName,
			"reason":      reason,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendFailoverResolved notifies a user when failover resolves and normal operation resumes.
func (s *Service) SendFailoverResolved(ctx context.Context, userID uuid.UUID, functionID, functionName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeFailoverResolved,
		Category: CategoryFailover,
		Title:    fmt.Sprintf("Failover Resolved: %s", functionName),
		Body:     fmt.Sprintf("Failover has resolved and normal operation has resumed for %s.", functionName),
		Data: JSONMap{
			"function_id":   functionID,
			"function_name": functionName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendProviderOffline notifies a user when a provider they use goes offline.
func (s *Service) SendProviderOffline(ctx context.Context, userID uuid.UUID, providerID, providerName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderOffline,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Offline: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is now offline. Some operations may be affected.", providerName),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendProviderOnline notifies a user when a provider comes back online.
func (s *Service) SendProviderOnline(ctx context.Context, userID uuid.UUID, providerID, providerName string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderOnline,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Online: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is now back online.", providerName),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendProviderDegraded notifies a user when a provider they use is degraded.
func (s *Service) SendProviderDegraded(ctx context.Context, userID uuid.UUID, providerID, providerName, reason string) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeProviderDegraded,
		Category: CategoryProvider,
		Title:    fmt.Sprintf("Provider Degraded: %s", providerName),
		Body:     fmt.Sprintf("Provider %s is experiencing degraded performance. %s", providerName, reason),
		Data: JSONMap{
			"provider_id":   providerID,
			"provider_name": providerName,
			"reason":        reason,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

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
			"amount_usd":        amountUSD,
			"stripe_transfer_id": stripeTransferID,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}
