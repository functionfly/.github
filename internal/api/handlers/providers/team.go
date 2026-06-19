package providers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleCreateTeamInvite creates team invitations during onboarding
func (h *Handler) HandleCreateTeamInvite(w http.ResponseWriter, r *http.Request) {
	var req TeamInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user for team invite")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	team, err := h.repo.GetTeamByUserID(r.Context(), user.ID)
	if err != nil {
		team = &storage.Team{
			Name:      fmt.Sprintf("%s's Team", user.Email),
			TenantID:  user.TenantID,
			CreatedBy: user.ID,
		}
		if err := h.repo.CreateTeam(r.Context(), team); err != nil {
			logrus.WithError(err).Error("Failed to create team")
			apierror.WriteError(w, apierror.NewInternal("Failed to create team"))
			return
		}

		teamMember := &storage.TeamMembership{
			TeamID: team.ID,
			UserID: user.ID,
			Role:   "admin",
		}
		if err := h.repo.AddTeamMember(r.Context(), teamMember); err != nil {
			logrus.WithError(err).Error("Failed to add user to team")
			apierror.WriteError(w, apierror.NewInternal("Failed to setup team"))
			return
		}
	}

	var invites []TeamInvite
	for _, email := range req.Emails {
		token, expires := auth.GenerateInviteToken()

		invite := &storage.TeamInvite{
			TeamID:    team.ID,
			Email:     email,
			Token:     token,
			Role:      req.Role,
			InvitedBy: user.ID,
			ExpiresAt: expires,
			Message:   req.Message,
		}

		if err := h.repo.CreateTeamInvite(invite); err != nil {
			logrus.WithError(err).WithField("email", email).Error("Failed to create team invite")
			continue
		}

		if h.notify != nil {
			invitedUser, err := h.repo.GetUserByEmail(r.Context(), email)
			if err == nil && invitedUser != nil {
				invitedByName := user.Email
				if user.Username != nil && *user.Username != "" {
					invitedByName = *user.Username
				}
				if err := h.notify.SendTeamInviteSent(r.Context(), invitedUser.ID, team.ID, team.Name, invitedByName, req.Role); err != nil {
					logrus.WithError(err).WithField("email", email).Warn("Failed to send team invite notification")
				}
			}
		}

		invites = append(invites, TeamInvite{
			Email:   email,
			Token:   token,
			Expires: expires.Unix(),
		})
	}

	response := TeamInviteResponse{Invites: invites}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleShareProvider shares a provider configuration with team members
func (h *Handler) HandleShareProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["providerId"]

	var req struct {
		TeamID string `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	provider, err := h.repo.GetProviderByID(r.Context(), providerID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Provider not found"))
		return
	}

	if provider.UserID != claims.UserID {
		isAdmin, err := h.repo.IsTeamAdmin(r.Context(), claims.UserID, req.TeamID)
		if err != nil || !isAdmin {
			apierror.WriteError(w, apierror.NewForbidden("Unauthorized"))
			return
		}
	}

	if err := h.repo.ShareProviderWithTeam(r.Context(), providerID, req.TeamID); err != nil {
		logrus.WithError(err).Error("Failed to share provider")
		apierror.WriteError(w, apierror.NewInternal("Failed to share provider"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "shared"})
}
