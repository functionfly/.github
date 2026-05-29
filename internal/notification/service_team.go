package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

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
