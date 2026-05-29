package notification

import (
	"context"

	"github.com/google/uuid"
)

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
