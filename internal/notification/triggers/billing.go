package triggers

import (
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

// BillingEvent represents a billing event
type BillingEvent struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TenantID    uuid.UUID
	Type        string // "invoice_generated", "payment_failed", "payment_success", "subscription_expiring"
	Amount      string
	Currency    string
	InvoiceID   string
	DueDate     *time.Time
	ErrorMsg    string
	Timestamp   time.Time
}

// BillingTrigger handles billing events and creates notifications
type BillingTrigger struct {
	name string
}

// NewBillingTrigger creates a new billing trigger
func NewBillingTrigger() *BillingTrigger {
	return &BillingTrigger{
		name: "billing",
	}
}

// Name returns the trigger name
func (t *BillingTrigger) Name() string {
	return t.name
}

// ShouldTrigger determines if this trigger should handle the event
func (t *BillingTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*BillingEvent)
	return ok
}

// BuildNotification creates a notification from a billing event
func (t *BillingTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*BillingEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.Type {
	case "invoice_generated":
		dueDate := "N/A"
		if e.DueDate != nil {
			dueDate = e.DueDate.Format("2006-01-02")
		}
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeBillingInvoiceGenerated,
			Category: notification.CategoryBilling,
			Title:    "New Invoice Available",
			Body:     fmt.Sprintf("A new invoice for %s %s is available. Due date: %s", e.Currency, e.Amount, dueDate),
			Data: notification.JSONMap{
				"invoice_id": e.InvoiceID,
				"amount":     e.Amount,
				"currency":   e.Currency,
				"due_date":   dueDate,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "payment_failed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeBillingPaymentFailed,
			Category: notification.CategoryBilling,
			Title:    "Payment Failed",
			Body:     fmt.Sprintf("Your payment of %s %s failed. %s", e.Currency, e.Amount, e.ErrorMsg),
			Data: notification.JSONMap{
				"invoice_id":   e.InvoiceID,
				"amount":       e.Amount,
				"currency":     e.Currency,
				"error":        e.ErrorMsg,
				"retry_needed": true,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "payment_success":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeBillingPaymentSuccess,
			Category: notification.CategoryBilling,
			Title:    "Payment Successful",
			Body:     fmt.Sprintf("Your payment of %s %s was successful.", e.Currency, e.Amount),
			Data: notification.JSONMap{
				"invoice_id": e.InvoiceID,
				"amount":     e.Amount,
				"currency":   e.Currency,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "subscription_expiring":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeBillingSubscriptionExpiring,
			Category: notification.CategoryBilling,
			Title:    "Subscription Expiring Soon",
			Body:     "Your subscription will expire soon. Please renew to avoid service interruption.",
			Data: notification.JSONMap{
				"tenant_id": e.TenantID.String(),
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	default:
		return nil, fmt.Errorf("unknown billing event type: %s", e.Type)
	}
}
