package certification

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// logCertVerificationAudit records a public verification lookup for compliance/forensics.
func (h *Handler) logCertVerificationAudit(ctx context.Context, ip, userAgent, eventType, target string) {
	logrus.WithFields(logrus.Fields{
		"event_type":  eventType,
		"target":      target,
		"ip":          ip,
		"user_agent":  userAgent,
		"timestamp":   time.Now().UTC(),
	}).Info("cert_verification_audit")
}

// VerifyCredential handles GET /v1/certification/verify/{username}
// Public endpoint — returns a user's active credentials for verification
func (h *Handler) VerifyCredential(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "Username is required")
		return
	}

	h.logCertVerificationAudit(r.Context(), getClientIP(r), r.UserAgent(), "verify_by_username", username)

	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Error("Failed to get user for verification")
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify credentials")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	creds, err := h.repo.ListActiveCredentialsByUser(r.Context(), user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list credentials for verification")
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify credentials")
		return
	}

	result := make([]map[string]interface{}, 0, len(creds))
	for _, c := range creds {
		entry := map[string]interface{}{
			"credential_number": c.CredentialNumber,
			"status":            c.Status,
			"issued_at":         c.IssuedAt,
			"expires_at":        c.ExpiresAt,
			"verification_hash": c.VerificationHash,
		}
		if c.Tier != nil {
			entry["tier"] = map[string]interface{}{
				"slug": c.Tier.Slug,
				"name": c.Tier.Name,
			}
		}
		result = append(result, entry)
	}

	userData := map[string]interface{}{
		"id": user.ID,
	}
	if user.Username != nil {
		userData["username"] = *user.Username
	}
	if user.Name != "" {
		userData["name"] = user.Name
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":        userData,
		"credentials": result,
	})
}

// VerifyByNumber handles GET /v1/certification/verify/number/{credentialNumber}
// Public endpoint — verifies a specific credential by its number
func (h *Handler) VerifyByNumber(w http.ResponseWriter, r *http.Request) {
	number := mux.Vars(r)["credentialNumber"]
	if number == "" {
		writeJSONError(w, http.StatusBadRequest, "Credential number is required")
		return
	}

	h.logCertVerificationAudit(r.Context(), getClientIP(r), r.UserAgent(), "verify_by_number", number)

	cred, err := h.repo.GetCredentialByNumber(r.Context(), number)
	if err != nil {
		logrus.WithError(err).Error("Failed to get credential by number")
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify credential")
		return
	}
	if cred == nil {
		writeJSONError(w, http.StatusNotFound, "Credential not found")
		return
	}

	result := map[string]interface{}{
		"credential_number": cred.CredentialNumber,
		"status":            cred.Status,
		"issued_at":         cred.IssuedAt,
		"expires_at":        cred.ExpiresAt,
		"verification_hash": cred.VerificationHash,
	}
	if cred.Tier != nil {
		result["tier"] = map[string]interface{}{
			"slug": cred.Tier.Slug,
			"name": cred.Tier.Name,
		}
	}
	if cred.User != nil {
		userData := map[string]interface{}{"id": cred.User.ID}
		if cred.User.Username != nil {
			userData["username"] = *cred.User.Username
		}
		if cred.User.Name != "" {
			userData["name"] = cred.User.Name
		}
		result["user"] = userData
	}

	writeJSON(w, http.StatusOK, result)
}

// PublicBadges handles GET /v1/certification/credentials/{username}/badges
// Public endpoint — returns badge data for embedding in profiles
func (h *Handler) PublicBadges(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "Username is required")
		return
	}

	h.logCertVerificationAudit(r.Context(), getClientIP(r), r.UserAgent(), "public_badges", username)

	user, err := h.userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	creds, err := h.repo.ListActiveCredentialsByUser(r.Context(), user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list badges")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve badges")
		return
	}

	badges := make([]map[string]interface{}, 0, len(creds))
	for _, c := range creds {
		badge := map[string]interface{}{
			"tier_slug":         "",
			"tier_name":         "",
			"tier_color":        "",
			"tier_icon":         "",
			"credential_number": c.CredentialNumber,
			"issued_at":         c.IssuedAt,
			"expires_at":        c.ExpiresAt,
		}
		if c.Tier != nil {
			badge["tier_slug"] = c.Tier.Slug
			badge["tier_name"] = c.Tier.Name
			badge["tier_color"] = c.Tier.Color
			badge["tier_icon"] = c.Tier.Icon
		}
		badges = append(badges, badge)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"username": username,
		"badges":   badges,
		"count":    len(badges),
	})
}
