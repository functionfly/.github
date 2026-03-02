package triggers

import (
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

// TeamEvent represents a team event
type TeamEvent struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TeamID        uuid.UUID
	TeamName      string
	InvitedByID   uuid.UUID
	InvitedByName string
	MemberEmail   string
	Role          string
	Type          string // "invitation", "member_added", "member_removed", "role_changed"
	InvitationURL string
	Timestamp     time.Time
}

// TeamTrigger handles team events and creates notifications
type TeamTrigger struct {
	name string
}

// NewTeamTrigger creates a new team trigger
func NewTeamTrigger() *TeamTrigger {
	return &TeamTrigger{
		name: "team",
	}
}

// Name returns the trigger name
func (t *TeamTrigger) Name() string {
	return t.name
}

// ShouldTrigger determines if this trigger should handle the event
func (t *TeamTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*TeamEvent)
	return ok
}

// BuildNotification creates a notification from a team event
func (t *TeamTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*TeamEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.Type {
	case "invitation":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeTeamInvitation,
			Category: notification.CategoryTeam,
			Title:    fmt.Sprintf("You've been invited to join %s", e.TeamName),
			Body:     fmt.Sprintf("%s has invited you to join the team %s on FunctionFly.", e.InvitedByName, e.TeamName),
			Data: notification.JSONMap{
				"team_id":        e.TeamID.String(),
				"team_name":      e.TeamName,
				"invited_by":     e.InvitedByName,
				"invitation_url": e.InvitationURL,
				"invited_at":     e.Timestamp.Format(time.RFC3339),
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "member_added":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeTeamMemberAdded,
			Category: notification.CategoryTeam,
			Title:    fmt.Sprintf("New member added to %s", e.TeamName),
			Body:     fmt.Sprintf("%s has been added to the team %s with role: %s.", e.MemberEmail, e.TeamName, e.Role),
			Data: notification.JSONMap{
				"team_id":      e.TeamID.String(),
				"team_name":    e.TeamName,
				"member_email": e.MemberEmail,
				"role":         e.Role,
				"added_at":     e.Timestamp.Format(time.RFC3339),
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "member_removed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeTeamMemberRemoved,
			Category: notification.CategoryTeam,
			Title:    fmt.Sprintf("Member removed from %s", e.TeamName),
			Body:     fmt.Sprintf("%s has been removed from the team %s.", e.MemberEmail, e.TeamName),
			Data: notification.JSONMap{
				"team_id":      e.TeamID.String(),
				"team_name":    e.TeamName,
				"member_email": e.MemberEmail,
				"removed_at":   e.Timestamp.Format(time.RFC3339),
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "role_changed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeTeamRoleChanged,
			Category: notification.CategoryTeam,
			Title:    fmt.Sprintf("Your role in %s has changed", e.TeamName),
			Body:     fmt.Sprintf("Your role in team %s has been updated to %s.", e.TeamName, e.Role),
			Data: notification.JSONMap{
				"team_id":      e.TeamID.String(),
				"team_name":    e.TeamName,
				"new_role":     e.Role,
				"changed_at":   e.Timestamp.Format(time.RFC3339),
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	default:
		return nil, fmt.Errorf("unknown team event type: %s", e.Type)
	}
}
