package users

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/sirupsen/logrus"
)

// HandleSearchUsers handles GET /v1/users/search?q=prefix&limit=8
// Authenticated users only; returns public id, username, name for autocomplete.
func (h *Handler) HandleSearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if middleware.GetUserFromContext(r) == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if strings.HasPrefix(q, "@") {
		q = strings.TrimSpace(strings.TrimPrefix(q, "@"))
	}
	if len(q) < 2 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"users": []interface{}{}})
		return
	}

	limit := 8
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}

	hits, err := h.repo.SearchUsersByUsernamePrefix(r.Context(), q, limit)
	if err != nil {
		logrus.WithError(err).WithField("q", q).Warn("user search failed")
		writeJSONError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	out := make([]map[string]string, 0, len(hits))
	for _, hit := range hits {
		out = append(out, map[string]string{
			"id":       hit.ID.String(),
			"username": hit.Username,
			"name":     hit.Name,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"users": out})
}
