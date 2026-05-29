package studio

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// SearchHandler handles studio search HTTP requests.
type SearchHandler struct {
	repo *SearchRepository
}

// NewSearchHandler creates a search handler.
func NewSearchHandler(repo *SearchRepository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

// HandleSearch handles GET /v1/studio/search?q=&type=&limit=
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("query")
	}
	resultType := r.URL.Query().Get("type")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	results, err := h.repo.Search(r.Context(), SearchParams{
		TenantID:    tenantID,
		UserID:      userID,
		Environment: environment,
		Query:       query,
		Type:        resultType,
		Limit:       limit,
	})
	if err != nil {
		logrus.WithError(err).Warn("studio search: failed")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"results": []SearchResult{},
			"query":   query,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   query,
		"total":   len(results),
	})
}
