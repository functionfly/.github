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
)

// HandleCreateTeamInvite creates team invitations during onboarding
func (h *Handler) HandleCreateTeamInvite(w http.ResponseWriter, r *http.Request) {
	var req TeamInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user for team invite")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	team, err := h.repo.GetTeamByUserID(user.ID)
	if err != nil {
		team = &storage.Team{
			Name:      fmt.Sprintf("%s's Team", user.Email),
			TenantID:  user.TenantID,
			CreatedBy: user.ID,
		}
		if err := h.repo.CreateTeam(team); err != nil {
			logrus.WithError(err).Error("Failed to create team")
			http.Error(w, "Failed to create team", http.StatusInternalServerError)
			return
		}

		teamMember := &storage.TeamMembership{
			TeamID: team.ID,
			UserID: user.ID,
			Role:   "admin",
		}
		if err := h.repo.AddTeamMember(teamMember); err != nil {
			logrus.WithError(err).Error("Failed to add user to team")
			http.Error(w, "Failed to setup team", http.StatusInternalServerError)
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
			invitedUser, err := h.repo.GetUserByEmail(email)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider, err := h.repo.GetProviderByID(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	if provider.UserID != claims.UserID {
		isAdmin, err := h.repo.IsTeamAdmin(claims.UserID, req.TeamID)
		if err != nil || !isAdmin {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.repo.ShareProviderWithTeam(providerID, req.TeamID); err != nil {
		logrus.WithError(err).Error("Failed to share provider")
		http.Error(w, "Failed to share provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "shared"})
}
