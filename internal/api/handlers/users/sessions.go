package users

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// sessionResponseItem is the safe session payload returned to the client (no token).
type sessionResponseItem struct {
	ID             string `json:"id"`
	Device         string `json:"device"`
	IP             string `json:"ip"`
	Location       string `json:"location"`
	LastActive     string `json:"lastActive"`
	CurrentSession bool   `json:"currentSession"`
}

func parseUserAgent(ua string) string {
	if ua == "" {
		return "Unknown device"
	}
	// Simple heuristic: look for known browsers/OS
	if strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Edg") {
		return "Chrome"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
		return "Safari"
	}
	if strings.Contains(ua, "Edg") {
		return "Edge"
	}
	if strings.Contains(ua, "Mac") {
		return "Desktop (macOS)"
	}
	if strings.Contains(ua, "Windows") {
		return "Desktop (Windows)"
	}
	if strings.Contains(ua, "Linux") {
		return "Desktop (Linux)"
	}
	if strings.Contains(ua, "iPhone") || strings.Contains(ua, "Android") {
		return "Mobile"
	}
	return "Unknown device"
}

func formatLastActive(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "Now"
	}
	if diff < time.Hour {
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if diff < 24*time.Hour {
		hr := int(diff.Hours())
		if hr == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hr)
	}
	d := int(diff.Hours() / 24)
	if d == 1 {
		return "1 day ago"
	}
	if d < 7 {
		return fmt.Sprintf("%d days ago", d)
	}
	return t.Format("2006-01-02")
}

// HandleListSessions returns GET /v1/users/me/sessions - list active sessions for the current user
func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	var currentSessionID uuid.UUID
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			if sess, err := h.repo.GetSessionByToken(ctx, parts[1]); err == nil && sess != nil {
				currentSessionID = sess.ID
			}
		}
	}

	sessions, err := h.repo.ListUserSessions(ctx, claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to list sessions")
		apierror.WriteError(w, apierror.NewInternal("Failed to load sessions"))
		return
	}

	items := make([]sessionResponseItem, 0, len(sessions))
	for _, s := range sessions {
		lastActive := formatLastActive(s.LastActivity)
		items = append(items, sessionResponseItem{
			ID:             s.ID.String(),
			Device:         parseUserAgent(s.UserAgent),
			IP:             s.IPAddress,
			Location:       "", // optional: derive from IP later
			LastActive:     lastActive,
			CurrentSession: s.ID == currentSessionID,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": items})
}

// HandleRevokeSession handles DELETE /v1/users/me/sessions/{id} - revoke one session
func (h *Handler) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	sessionIDStr := vars["id"]
	if sessionIDStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("session id required"))
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid session id"))
		return
	}

	if err := h.repo.DeleteSessionByID(r.Context(), sessionID, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			apierror.WriteError(w, apierror.NewNotFound("Session not found"))
			return
		}
		logrus.WithError(err).WithField("sessionID", sessionID).Error("Failed to revoke session")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke session"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Session revoked"})
}

// HandleRevokeOtherSessions handles POST /v1/users/me/sessions/revoke-others - revoke all other sessions
func (h *Handler) HandleRevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	var currentSessionID uuid.UUID
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			if sess, err := h.repo.GetSessionByToken(ctx, parts[1]); err == nil && sess != nil {
				currentSessionID = sess.ID
			}
		}
	}

	sessions, err := h.repo.ListUserSessions(ctx, claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to list sessions for revoke-others")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke sessions"))
		return
	}

	for _, s := range sessions {
		if s.ID != currentSessionID {
			_ = h.repo.DeleteSessionByID(ctx, s.ID, claims.UserID)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "All other sessions revoked"})
}

// loginHistoryResponseItem is the safe login history payload returned to the client.
type loginHistoryResponseItem struct {
	ID          string `json:"id"`
	EventType   string `json:"eventType"`
	IP          string `json:"ip"`
	Device      string `json:"device"`
	Location    string `json:"location,omitempty"`
	LoginMethod string `json:"loginMethod,omitempty"`
	MFAUsed     bool   `json:"mfaUsed"`
	CreatedAt   string `json:"createdAt"`
}

// HandleListLoginHistory returns GET /v1/users/me/login-history - list login history for the current user
func (h *Handler) HandleListLoginHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()

	// Parse pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parsePositiveInt(limitStr); err == nil && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := parsePositiveInt(offsetStr); err == nil {
			offset = o
		}
	}

	history, err := h.repo.ListUserLoginHistory(ctx, claims.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to list login history")
		apierror.WriteError(w, apierror.NewInternal("Failed to load login history"))
		return
	}

	total, err := h.repo.CountUserLoginHistory(ctx, claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Error("Failed to count login history")
	}

	items := make([]loginHistoryResponseItem, 0, len(history))
	for _, h := range history {
		items = append(items, loginHistoryResponseItem{
			ID:          h.ID.String(),
			EventType:   h.EventType,
			IP:          h.IPAddress,
			Device:      parseUserAgent(h.UserAgent),
			Location:    h.Location,
			LoginMethod: h.LoginMethod,
			MFAUsed:     h.MFAUsed,
			CreatedAt:   formatLoginHistoryTime(h.CreatedAt),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"history": items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func parsePositiveInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid number")
	}
	return n, nil
}

func formatLoginHistoryTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if diff < 24*time.Hour {
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	if diff < 7*24*time.Hour {
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}
	return t.Format("Jan 2, 2006")
}
