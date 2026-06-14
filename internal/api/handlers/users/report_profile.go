package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var allowedProfileReportReasons = map[string]struct{}{
	"tos_violation": {},
	"harassment":    {},
	"spam":          {},
	"impersonation": {},
	"other":         {},
}

type reportProfileRequest struct {
	Reason               string `json:"reason"`
	Details              string `json:"details"`
	AcknowledgedAccuracy bool   `json:"acknowledged_accuracy"`
}

// HandleReportProfile handles POST /v1/users/{username}/report (auth required).
// Stores a moderation report in the feedback table for staff review.
func (h *Handler) HandleReportProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	username := strings.TrimSpace(mux.Vars(r)["username"])
	if username == "" || strings.EqualFold(username, "me") {
		apierror.WriteError(w, apierror.NewBadRequest("username is required"))
		return
	}

	var req reportProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON body"))
		return
	}

	if !req.AcknowledgedAccuracy {
		apierror.WriteError(w, apierror.NewBadRequest("You must confirm this report is submitted in good faith"))
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if _, ok := allowedProfileReportReasons[reason]; !ok {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid report reason"))
		return
	}

	details := strings.TrimSpace(req.Details)
	if len(details) > 4000 {
		apierror.WriteError(w, apierror.NewBadRequest("Details must be 4000 characters or less"))
		return
	}
	if reason == "other" && len(details) < 20 {
		apierror.WriteError(w, apierror.NewBadRequest("Please describe what happened (at least 20 characters)"))
		return
	}

	reported, err := h.repo.GetUserForPublicProfile(r.Context(), username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("GetUserForPublicProfile for report failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to load profile"))
		return
	}
	if reported == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	if reported.ID == claims.UserID {
		apierror.WriteError(w, apierror.NewBadRequest("You cannot report your own profile"))
		return
	}

	reportedUsername := ""
	if reported.Username != nil {
		reportedUsername = *reported.Username
	}

	reporterID := claims.UserID
	recent, err := h.repo.GetFeedbackByUser(r.Context(), &reporterID, nil, 40, 0)
	if err == nil {
		cutoff := time.Now().Add(-1 * time.Hour)
		n := 0
		for _, f := range recent {
			if f.FeedbackType == "profile_report" && f.CreatedAt.After(cutoff) {
				n++
			}
		}
		if n >= 8 {
			apierror.WriteError(w, apierror.NewRateLimited("Too many reports submitted. Please try again later."))
			return
		}
	}

	ipAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ipAddr = host
	}
	if ipAddr != "" && net.ParseIP(ipAddr) == nil {
		ipAddr = ""
	}

	subject := fmt.Sprintf("Profile report: @%s (%s)", reportedUsername, reason)
	message := fmt.Sprintf(
		"reported_user_id: %s\nreported_username: @%s\nreason: %s\nreporter_user_id: %s\nprofile_path: /u/%s\ndetails:\n%s\n",
		reported.ID.String(),
		reportedUsername,
		reason,
		claims.UserID.String(),
		reportedUsername,
		details,
	)

	feedback := &storage.Feedback{
		FeedbackType: "profile_report",
		Subject:      subject,
		Message:      message,
		Priority:     "high",
		UserID:       &reporterID,
		IPAddress:    ipAddr,
		UserAgent:    r.Header.Get("User-Agent"),
	}

	_, err = h.repo.CreateFeedback(r.Context(), feedback)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" && strings.Contains(strings.ToLower(pqErr.Message), "feedback_type") {
			apierror.WriteError(w, apierror.NewBadRequest("Profile reporting is not enabled on this deployment (database constraint)."))
			return
		}
		logrus.WithError(err).Error("CreateFeedback profile_report failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to submit report"))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Thank you. Our team will review your report.",
	})
}
