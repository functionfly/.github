package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// userLookup is the minimal interface the Dispatcher needs to resolve full user details.
type userLookup interface {
	GetUserByID(userID uuid.UUID) (*storage.User, error)
	GetUserSettings(userID uuid.UUID) (map[string]interface{}, error)
}

// Dispatcher handles notification delivery to channels
type Dispatcher struct {
	channels   map[string]Channel
	repo       Repository
	userLookup userLookup
	logger     *logrus.Logger
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(channels map[string]Channel, repo Repository, ul userLookup, logger *logrus.Logger) *Dispatcher {
	return &Dispatcher{
		channels:   channels,
		repo:       repo,
		userLookup: ul,
		logger:     logger,
	}
}

// Dispatch sends a notification through configured channels
func (d *Dispatcher) Dispatch(ctx context.Context, n *Notification) error {
	if err := d.repo.UpdateNotificationStatus(ctx, n.ID, StatusProcessing); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	user := &storage.User{ID: n.UserID}
	if d.userLookup != nil {
		if u, err := d.userLookup.GetUserByID(n.UserID); err != nil {
			d.logger.WithError(err).WithField("user_id", n.UserID).Warn("failed to look up user for notification dispatch; email channel will be skipped")
		} else if u != nil {
			user = u
		}
	}

	successCount := 0
	failedCount := 0

	settings, _ := d.userLookup.GetUserSettings(n.UserID)

	for _, channelName := range n.Channels {
		channel, ok := d.channels[channelName]
		if !ok {
			d.logger.WithField("channel", channelName).Warn("Unknown notification channel")
			continue
		}

		if !channel.IsConfigured() {
			d.logger.WithField("channel", channelName).Debug("Channel not configured")
			continue
		}

		pref, err := d.repo.GetPreference(ctx, n.UserID, channelName, n.Category)
		if err != nil {
			d.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":  n.UserID,
				"channel":  channelName,
				"category": n.Category,
			}).Warn("Failed to get preference; proceeding with send")
		}

		if !ShouldDeliverChannel(settings, pref, n.Category, n.Type, channelName) {
			d.logger.WithFields(logrus.Fields{
				"user_id":  n.UserID,
				"channel":  channelName,
				"category": n.Category,
				"type":     n.Type,
			}).Debug("Notification skipped by preference rules")
			continue
		}

		if err := channel.Send(ctx, n, user); err != nil {
			d.logger.WithError(err).WithFields(logrus.Fields{
				"notification_id": n.ID,
				"channel":         channelName,
			}).Error("Failed to send notification")

			analytics := &NotificationAnalytics{
				NotificationID: n.ID,
				Channel:        channelName,
				Status:         AnalyticsStatusFailed,
				ErrorMessage:   strPtr(err.Error()),
			}
			d.repo.TrackAnalytics(ctx, analytics)
			failedCount++
		} else {
			now := time.Now()
			analytics := &NotificationAnalytics{
				NotificationID: n.ID,
				Channel:        channelName,
				Status:         AnalyticsStatusDelivered,
				DeliveredAt:    &now,
			}
			d.repo.TrackAnalytics(ctx, analytics)
			successCount++
		}
	}

	finalStatus := StatusSent
	if failedCount > 0 && successCount == 0 {
		finalStatus = StatusFailed
	}

	if err := d.repo.UpdateNotificationStatus(ctx, n.ID, finalStatus); err != nil {
		return fmt.Errorf("failed to update final notification status: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"notification_id": n.ID,
		"success_count":   successCount,
		"failed_count":    failedCount,
		"final_status":    finalStatus,
	}).Info("Notification dispatched")

	return nil
}

func strPtr(s string) *string {
	return &s
}
