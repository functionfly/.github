package founders

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo              storage.Repository
	notificationSvc   *notification.Service
	log               *logrus.Logger
}

func NewHandler(repo storage.Repository, notificationSvc *notification.Service, log *logrus.Logger) *Handler {
	return &Handler{
		repo:            repo,
		notificationSvc: notificationSvc,
		log:             log,
	}
}

func (h *Handler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	isFounder := user.IsFounder
	var founderNumber *int
	if isFounder {
		founderNumber = user.FounderNumber
	}

	founderCount, _ := h.repo.GetFounderCount(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_founder":     isFounder,
		"founder_number": founderNumber,
		"total_founders": founderCount,
		"max_founders":   10000,
		"benefits": map[string]interface{}{
			"permanent_badge":    true,
			"voting_rights":      isFounder,
			"lifetime_commissions": isFounder,
			"early_access":       isFounder,
		},
	})
}

func (h *Handler) HandleAssignFounder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if user.IsFounder {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"already_founder": true,
			"founder_number":  user.FounderNumber,
		})
		return
	}

	founderNumber, err := h.repo.AssignFounderStatus(r.Context(), claims.UserID)
	if err != nil {
		h.log.WithError(err).Error("Failed to assign founder status")
		http.Error(w, "failed to assign founder status", http.StatusInternalServerError)
		return
	}

	if founderNumber == 0 {
		http.Error(w, "no founder slots available", http.StatusConflict)
		return
	}

	achievement, _ := h.repo.GetAchievementBySlug(r.Context(), "founders_club")
	if achievement != nil {
		h.repo.AwardAchievement(claims.UserID, achievement.ID, map[string]interface{}{
			"founder_number": founderNumber,
		})
	}

	if h.notificationSvc != nil {
		title := fmt.Sprintf("🎉 You're Founder #%d!", founderNumber)
		body := "Your founder status is permanent and never expires. Enjoy lifetime benefits!"
		_, _ = h.notificationSvc.Send(r.Context(), notification.SendRequest{
			UserID:   claims.UserID,
			Type:     "founder_status",
			Category: "founder",
			Title:    title,
			Body:     body,
			Channels: []string{notification.ChannelInApp},
			Priority: notification.PriorityHigh,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"founder_number": founderNumber,
		"message":        "Congratulations! You are now a FunctionFly Founder.",
	})
}

func (h *Handler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 50
	founders, total, err := h.repo.GetFoundersLeaderboard(r.Context(), limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to get founders leaderboard")
		http.Error(w, "failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	type FounderEntry struct {
		ID            uuid.UUID `json:"id"`
		Name          string    `json:"name"`
		Username      string    `json:"username,omitempty"`
		FounderNumber int       `json:"founder_number"`
		AvatarURL     string    `json:"avatar_url,omitempty"`
	}

	entries := make([]FounderEntry, 0, len(founders))
	for _, founder := range founders {
		entry := FounderEntry{
			ID:            founder.ID,
			FounderNumber: 0,
		}
		if founder.FounderNumber != nil {
			entry.FounderNumber = *founder.FounderNumber
		}
		if founder.Name != "" {
			entry.Name = founder.Name
		}
		if founder.Username != nil {
			entry.Username = *founder.Username
		}
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"founders":      entries,
		"total_founders": total,
		"limit":         limit,
	})
}

type PublicLeaderboardEntry struct {
	FounderNumber   int    `json:"founder_number"`
	DisplayName     string `json:"display_name"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	TotalEarnings   int64  `json:"total_earnings_cents"`
	ReferralCount   int    `json:"referral_count"`
	MemberSince     string `json:"member_since"`
}

func (h *Handler) HandleGetPublicLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 50

	founders, total, err := h.repo.GetFoundersLeaderboard(r.Context(), limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to get public leaderboard")
		http.Error(w, "failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	entries := make([]PublicLeaderboardEntry, 0, len(founders))
	for _, founder := range founders {
		entry := PublicLeaderboardEntry{
			FounderNumber: 0,
		}
		if founder.FounderNumber != nil {
			entry.FounderNumber = *founder.FounderNumber
		}
		entry.DisplayName = fmt.Sprintf("Founder #%d", entry.FounderNumber)
		if founder.AvatarURL != nil && *founder.AvatarURL != "" {
			entry.AvatarURL = *founder.AvatarURL
		}
		entry.MemberSince = founder.CreatedAt.Format("2006-01-02")
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":        entries,
		"total_founders": total,
	})
}

func (h *Handler) HandleGetMyRank(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if !user.IsFounder || user.FounderNumber == nil {
		http.Error(w, "not a founder", http.StatusNotFound)
		return
	}

	rank, _ := h.repo.GetFounderRank(r.Context(), claims.UserID)
	founderCount, _ := h.repo.GetFounderCount(r.Context())

	var percentile float64
	if founderCount > 0 && rank > 0 {
		percentile = (1 - float64(rank)/float64(founderCount)) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"founder_number": *user.FounderNumber,
		"rank":          rank,
		"percentile":    percentile,
		"total_founders": founderCount,
	})
}
