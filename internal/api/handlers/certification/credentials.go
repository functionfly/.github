package certification

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ListMyCredentials handles GET /v1/certification/credentials
// Returns the authenticated user's credentials
func (h *Handler) ListMyCredentials(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	creds, err := h.repo.ListCredentialsByUser(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list credentials")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve credentials")
		return
	}

	result := make([]map[string]interface{}, 0, len(creds))
	for _, c := range creds {
		entry := map[string]interface{}{
			"id":                 c.ID,
			"tier_id":            c.TierID,
			"credential_number":  c.CredentialNumber,
			"status":             c.Status,
			"issued_at":          c.IssuedAt,
			"expires_at":         c.ExpiresAt,
			"verification_url":   c.VerificationURL,
			"verification_hash":  c.VerificationHash,
		}
		if c.Tier != nil {
			entry["tier"] = map[string]interface{}{
				"slug": c.Tier.Slug,
				"name": c.Tier.Name,
			}
		}
		result = append(result, entry)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"credentials": result,
	})
}

// GetCredential handles GET /v1/certification/credentials/{credentialId}
// Returns a single credential with full details including Open Badges JSON
func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	credID, err := uuid.Parse(mux.Vars(r)["credentialId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid credential ID")
		return
	}

	cred, err := h.repo.GetCredentialByID(r.Context(), credID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get credential")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve credential")
		return
	}
	if cred == nil {
		writeJSONError(w, http.StatusNotFound, "Credential not found")
		return
	}
	if cred.UserID != claims.UserID {
		writeJSONError(w, http.StatusForbidden, "Access denied")
		return
	}

	result := map[string]interface{}{
		"id":                 cred.ID,
		"credential_number":  cred.CredentialNumber,
		"status":             cred.Status,
		"issued_at":          cred.IssuedAt,
		"expires_at":         cred.ExpiresAt,
		"verification_url":   cred.VerificationURL,
		"verification_hash":  cred.VerificationHash,
		"oba_credential":     cred.OBACredential,
	}
	if cred.Tier != nil {
		result["tier"] = map[string]interface{}{
			"slug": cred.Tier.Slug,
			"name": cred.Tier.Name,
		}
	}
	if cred.RevokedAt != nil {
		result["revoked_at"] = cred.RevokedAt
		result["revoked_reason"] = cred.RevokedReason
	}

	writeJSON(w, http.StatusOK, result)
}
