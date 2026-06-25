package companyrankinghandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/companyranking"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo  *companyranking.Repository
	cache *companyranking.Cache
	log   *logrus.Logger
}

func NewHandler(repo *companyranking.Repository, cache *companyranking.Cache, log *logrus.Logger) *Handler {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Handler{repo: repo, cache: cache, log: log}
}

func (h *Handler) HandleListLeaderboard(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetLeaderboard(r.Context(), country, string(category)); err == nil && hit {
		entries := toEntries(cached)
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, LeaderboardResponse{
			TotalRanked: len(cached),
			Entries:     entries,
			Country:     country,
			Category:    string(category),
			CacheHit:   true,
		})
		return
	}

	rows, err := h.repo.ListRankings(r.Context(), country, limit, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list rankings")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = h.cache.SetLeaderboard(r.Context(), country, string(category), rows)
	entries := toEntries(rows)
	if len(entries) > limit {
		entries = entries[:limit]
	}
	writeJSON(w, LeaderboardResponse{
		TotalRanked: len(rows),
		Entries:    entries,
		Country:    country,
		Category:   string(category),
		CacheHit:   false,
	})
}

func (h *Handler) HandleGetCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	category := parseCategory(r)

	rk, err := h.repo.GetRankingBySlug(r.Context(), slug, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to get company ranking")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if rk == nil {
		http.Error(w, "company not found", http.StatusNotFound)
		return
	}
	if rk.ActiveUsers < companyranking.MinActiveUsersForPublic {
		http.Error(w, "company below privacy threshold", http.StatusNotFound)
		return
	}
	writeJSON(w, CompanyResponse{Entry: toEntry(*rk)})
}

func parseLimit(raw string, def, max int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}

func parseCategory(r *http.Request) companyranking.Category {
	c := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	if c == "" || !companyranking.ValidCategory(c) {
		return companyranking.CategoryComposite
	}
	return companyranking.Category(c)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

type RankingEntry struct {
	Rank            int     `json:"rank"`
	PreviousRank    *int    `json:"previous_rank,omitempty"`
	RankDelta       int     `json:"rank_delta"`
	CompanyID       int64   `json:"company_id"`
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	CountryCode     string  `json:"country_code"`
	CitySlug        string  `json:"city_slug,omitempty"`
	EmployeeCount   int     `json:"employee_count"`
	ScoreRaw        float64 `json:"score_raw"`
	ScorePerCapita  float64 `json:"score_per_capita"`
	ActiveUsers     int     `json:"active_users"`
	Deployments     int     `json:"deployments"`
	Executions30d   int64   `json:"executions_30d"`
	RevenueCents    int64    `json:"revenue_cents"`
	NewUsers30d    int      `json:"new_users_30d"`
	Category       string   `json:"category"`
}

type LeaderboardResponse struct {
	TotalRanked int            `json:"total_ranked"`
	Entries    []RankingEntry  `json:"entries"`
	Country   string          `json:"country,omitempty"`
	Category  string          `json:"category"`
	CacheHit  bool            `json:"cache_hit"`
}

type CompanyResponse struct {
	Entry   RankingEntry `json:"entry"`
	CacheHit bool         `json:"cache_hit"`
}

func toEntry(r companyranking.Ranking) RankingEntry {
	prev := r.PrevRank
	return RankingEntry{
		Rank:          r.RankPosition,
		PreviousRank:   prev,
		RankDelta:     r.RankDelta,
		CompanyID:     r.CompanyID,
		Slug:          r.CompanySlug,
		Name:          r.CompanyName,
		CountryCode:   r.CountryCode,
		CitySlug:      r.CitySlug,
		EmployeeCount: r.EmployeeCount,
		ScoreRaw:      r.ScoreRaw,
		ScorePerCapita: r.ScorePerCapita,
		ActiveUsers:   r.ActiveUsers,
		Deployments:   r.Deployments,
		Executions30d: r.Executions30d,
		RevenueCents:  r.RevenueCents,
		NewUsers30d:  r.NewUsers30d,
		Category:     string(r.Category),
	}
}

func toEntries(rs []companyranking.Ranking) []RankingEntry {
	out := make([]RankingEntry, 0, len(rs))
	for _, r := range rs {
		out = append(out, toEntry(r))
	}
	return out
}
