// Package universityrankinghandler exposes the HTTP surface for the
// university leaderboard. It mirrors internal/api/handlers/cityranking.
package universityrankinghandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage/universityranking"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler holds the dependencies for the university-ranking HTTP surface.
type Handler struct {
	repo  *universityranking.Repository
	cache *universityranking.Cache
	log   *logrus.Logger
}

// NewHandler wires a Handler with the repository and cache. cache may be
// nil (handler will skip caching).
func NewHandler(repo *universityranking.Repository, cache *universityranking.Cache, log *logrus.Logger) *Handler {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Handler{repo: repo, cache: cache, log: log}
}

// HandleListLeaderboard: GET /university-rankings
func (h *Handler) HandleListLeaderboard(w http.ResponseWriter, r *http.Request) {
	country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("country")))
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetLeaderboard(r.Context(), country, category); err == nil && hit && cached != nil {
		writeJSON(w, LeaderboardResponse{
			PeriodEnd: latestPeriod(cached),
			TotalRanked: len(cached),
			Entries: ToRankingEntries(cached),
			Country:   country,
			Category:  string(category),
			CacheHit:  true,
			PrivacyMin: universityranking.MinActiveUsersForPublic,
		})
		return
	}

	rows, err := h.repo.ListUniversities(r.Context(), country, limit, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to list university rankings")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	period, _ := h.repo.LatestPeriod(r.Context())
	_ = h.cache.SetLeaderboard(r.Context(), country, category, rows)
	writeJSON(w, LeaderboardResponse{
		PeriodEnd: period,
		TotalRanked: len(rows),
		Entries:   ToRankingEntries(rows),
		Country:   country,
		Category:  string(category),
		CacheHit:  false,
		PrivacyMin: universityranking.MinActiveUsersForPublic,
	})
}

// HandleGetUniversity: GET /university-rankings/{slug}
func (h *Handler) HandleGetUniversity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.TrimSpace(vars["slug"])
	slug = strings.TrimPrefix(slug, "/")
	if slug == "" {
		http.Error(w, "slug required", http.StatusBadRequest)
		return
	}
	category := parseCategory(r)

	if cached, hit, err := h.cache.GetDetail(r.Context(), slug, category); err == nil && hit && cached != nil {
		writeJSON(w, ToDetailResponse(slug, cached, true))
		return
	}

	rk, err := h.repo.GetRankingBySlug(r.Context(), slug, category)
	if err != nil {
		h.log.WithError(err).Error("Failed to get university ranking")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if rk == nil {
		http.Error(w, "university not found", http.StatusNotFound)
		return
	}
	if rk.ActiveUsers < universityranking.MinActiveUsersForPublic {
		http.Error(w, "university below privacy threshold", http.StatusNotFound)
		return
	}
	_ = h.cache.SetDetail(r.Context(), slug, category, rk)
	writeJSON(w, ToDetailResponse(slug, rk, false))
}

// HandleGetMyUniversity: GET /university-rankings/me
func (h *Handler) HandleGetMyUniversity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())
	category := parseCategory(r)

	uni, err := h.repo.GetUserUniversity(r.Context(), userID.String())
	if err != nil {
		h.log.WithError(err).Error("Failed to get user university")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if uni == nil {
		writeJSON(w, MyUniversityResponse{Source: "none"})
		return
	}
	rk, err := h.repo.GetRankingBySlug(r.Context(), uni.Slug, category)
	if err != nil || rk == nil || rk.ActiveUsers < universityranking.MinActiveUsersForPublic {
		// University is set but no public ranking yet (e.g., new signup).
		writeJSON(w, MyUniversityResponse{
			Source:     "set_no_ranking",
			University: ToUniversityEntry(*uni),
		})
		return
	}
	writeJSON(w, MyUniversityResponse{
		Source:     "ranking",
		University: ToUniversityEntry(*uni),
		Ranking:    ToRankingEntry(*rk),
	})
}

// HandleResolveUniversity: POST /university-rankings/resolve (auth)
func (h *Handler) HandleResolveUniversity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())

	var body struct {
		Input string `json:"input"`
		Slug  string `json:"slug"`
	}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
	}

	// Direct slug short-circuit.
	if body.Slug != "" {
		uni, err := h.repo.GetBySlug(r.Context(), body.Slug)
		if err != nil {
			h.log.WithError(err).Error("Failed to get university by slug")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if uni == nil {
			http.Error(w, "university not found", http.StatusNotFound)
			return
		}
		if err := h.repo.AssignUserUniversity(r.Context(), userID.String(), &uni.ID, "user_input"); err != nil {
			h.log.WithError(err).Warn("AssignUserUniversity (slug) failed")
		}
		writeJSON(w, ResolutionResponse{
			University: ToUniversityEntry(*uni),
			Source:     "slug",
		})
		return
	}

	if body.Input == "" {
		http.Error(w, "input or slug required", http.StatusBadRequest)
		return
	}
	norm := normalize(body.Input)
	matches, err := h.repo.LookupByAlias(r.Context(), norm)
	if err != nil {
		h.log.WithError(err).Error("LookupByAlias failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(matches) == 0 {
		http.Error(w, "no university match", http.StatusNotFound)
		return
	}
	first := matches[0]
	ambiguous := len(matches) > 1
	var idPtr *int64
	if !ambiguous {
		id := first.ID
		idPtr = &id
	}
	if err := h.repo.AssignUserUniversity(r.Context(), userID.String(), idPtr, "user_input"); err != nil {
		h.log.WithError(err).Warn("AssignUserUniversity (input) failed")
	}
	writeJSON(w, ResolutionResponse{
		University: ToUniversityEntry(first),
		Source:     "alias",
		Ambiguous:  ambiguous,
		MatchCount: len(matches),
	})
}

// HandleGetOptOut: GET /users/me/university-ranking-opt-out
func (h *Handler) HandleGetOptOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())
	v, err := h.repo.IsOptedOut(r.Context(), userID.String())
	if err != nil {
		h.log.WithError(err).Error("Failed to read opt-out")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, OptOutResponse{OptedOut: v})
}

// HandleSetOptOut: POST /users/me/university-ranking-opt-out
func (h *Handler) HandleSetOptOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := uuid.MustParse(claims.UserID.String())
	var body SetOptOutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OptedOut == nil {
		http.Error(w, "opted_out (bool) required", http.StatusBadRequest)
		return
	}
	if err := h.repo.SetOptOut(r.Context(), userID.String(), *body.OptedOut); err != nil {
		h.log.WithError(err).Error("Failed to set opt-out")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, OptOutResponse{OptedOut: *body.OptedOut})
}

// ── helpers ───────────────────────────────────────────────────────────────

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

func parseCategory(r *http.Request) universityranking.Category {
	c := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	if c == "" || !universityranking.ValidCategory(c) {
		return universityranking.CategoryComposite
	}
	return universityranking.Category(c)
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Strip common punctuation: commas, periods, dashes.
	for _, c := range []string{",", ".", "-", "'", "\""} {
		s = strings.ReplaceAll(s, c, " ")
	}
	return strings.Join(strings.Fields(s), " ")
}

func latestPeriod(rs []universityranking.Ranking) time.Time {
	var t time.Time
	for _, r := range rs {
		if r.PeriodEnd.After(t) {
			t = r.PeriodEnd
		}
	}
	return t
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
