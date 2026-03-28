package users

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// attachProfileStats adds authoritative registry/follow aggregates for profile stat cards.
func (h *Handler) attachProfileStats(profile map[string]interface{}, userID uuid.UUID) {
	stats, err := h.repo.GetUserProfileStats(userID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("Failed to load profile stats")
		profile["stats"] = map[string]interface{}{
			"functionsCount":  0,
			"totalExecutions":   int64(0),
			"trustScore":        0,
			"followersCount":    0,
			"followingCount":    0,
		}
		return
	}
	profile["stats"] = stats
}
