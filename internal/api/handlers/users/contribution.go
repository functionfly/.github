package users

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const contributionGraphDays = 365

// HandleGetUserContributions handles GET /v1/users/{username}/contributions
// Returns a GitHub-style daily series (UTC) plus streak stats for the profile heatmap.
func (h *Handler) HandleGetUserContributions(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("Failed to get user by username")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	start := today.AddDate(0, 0, -(contributionGraphDays - 1))

	counts, err := h.repo.GetUserContributionDailyCounts(user.ID, start)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to get contribution counts, returning empty graph")
		writeJSON(w, http.StatusOK, emptyContributionResponse(today))
		return
	}

	maxCount := int64(0)
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	days := make([]map[string]interface{}, 0, contributionGraphDays)
	var lastContrib *time.Time

	for i := contributionGraphDays - 1; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		cnt := counts[dateStr]
		if cnt > 0 {
			t := day
			lastContrib = &t
		}
		lvl := contributionLevel(cnt, maxCount)
		days = append(days, map[string]interface{}{
			"date":  dateStr,
			"count": cnt,
			"level": lvl,
		})
	}

	currentStreak := contributionCurrentStreak(today, counts)
	longestStreak := contributionLongestStreak(today, counts)

	var lastStr interface{}
	if lastContrib != nil {
		lastStr = lastContrib.Format("2006-01-02")
	} else {
		lastStr = nil
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"days":                 days,
		"currentStreak":        currentStreak,
		"longestStreak":        longestStreak,
		"lastContributionDate": lastStr,
	})
}

func emptyContributionResponse(today time.Time) map[string]interface{} {
	days := make([]map[string]interface{}, 0, contributionGraphDays)
	for i := contributionGraphDays - 1; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		days = append(days, map[string]interface{}{
			"date":  dateStr,
			"count": int64(0),
			"level": 0,
		})
	}
	return map[string]interface{}{
		"days":                 days,
		"currentStreak":        0,
		"longestStreak":        0,
		"lastContributionDate": nil,
	}
}

func contributionLevel(count, maxCount int64) int {
	if count <= 0 || maxCount <= 0 {
		return 0
	}
	// Map 1..maxCount to levels 1..4 (linear by intensity).
	lvl := int((count*4 + maxCount - 1) / maxCount)
	if lvl < 1 {
		lvl = 1
	}
	if lvl > 4 {
		lvl = 4
	}
	return lvl
}

func contributionCurrentStreak(today time.Time, counts map[string]int64) int {
	streak := 0
	for i := 0; i < contributionGraphDays; i++ {
		d := today.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		if counts[key] > 0 {
			streak++
		} else {
			break
		}
	}
	return streak
}

func contributionLongestStreak(today time.Time, counts map[string]int64) int {
	best := 0
	run := 0
	for i := contributionGraphDays - 1; i >= 0; i-- {
		d := today.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		if counts[key] > 0 {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 0
		}
	}
	return best
}
