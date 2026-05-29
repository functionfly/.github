package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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
