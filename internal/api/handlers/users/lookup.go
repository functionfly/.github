package users

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const maxLookupUserIDs = 50

type lookupByIDsRequest struct {
	UserIDs []string `json:"user_ids"`
}

// HandleLookupUsersByIDs resolves user UUIDs to public display fields for the authenticated client.
// POST /v1/users/lookup-by-ids  body: { "user_ids": ["uuid", ...] }  (max 50, deduped)
func (h *Handler) HandleLookupUsersByIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if middleware.GetUserFromContext(r) == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req lookupByIDsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(req.UserIDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "user_ids is required")
		return
	}
	if len(req.UserIDs) > maxLookupUserIDs {
		writeJSONError(w, http.StatusBadRequest, "too many user_ids")
		return
	}

	seen := make(map[string]struct{}, len(req.UserIDs))
	out := make([]map[string]string, 0, len(req.UserIDs))

	for _, idStr := range req.UserIDs {
		if idStr == "" {
			continue
		}
		if _, dup := seen[idStr]; dup {
			continue
		}
		seen[idStr] = struct{}{}

		uid, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}

		u, err := h.repo.GetUserByID(uid)
		if err != nil {
			logrus.WithError(err).WithField("user_id", idStr).Warn("lookup-by-ids: GetUserByID failed")
			continue
		}
		if u == nil {
			continue
		}

		username := ""
		if u.Username != nil {
			username = *u.Username
		}
		name := u.Name
		if name == "" && u.ProviderData != nil {
			if n, ok := u.ProviderData["name"].(string); ok {
				name = n
			}
		}

		out = append(out, map[string]string{
			"id":       u.ID.String(),
			"username": username,
			"name":     name,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"users": out})
}
