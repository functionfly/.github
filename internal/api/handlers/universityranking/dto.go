package universityrankinghandler

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage/universityranking"
)

// LeaderboardResponse is the payload for GET /university-rankings.
type LeaderboardResponse struct {
	PeriodEnd   time.Time     `json:"period_end"`
	TotalRanked int           `json:"total_ranked"`
	Entries     []RankingEntry `json:"entries"`
	Country     string        `json:"country,omitempty"`
	Category    string        `json:"category"`
	CacheHit    bool          `json:"cache_hit"`
	PrivacyMin  int           `json:"privacy_min_active_users"`
}

// RankingEntry is one row in the public leaderboard.
type RankingEntry struct {
	Rank           int       `json:"rank"`
	PreviousRank   *int      `json:"previous_rank,omitempty"`
	RankDelta      int       `json:"rank_delta"`
	UniversityID   int64     `json:"university_id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	ShortName      string    `json:"short_name,omitempty"`
	CountryCode    string    `json:"country_code"`
	StateCode      string    `json:"state_code,omitempty"`
	CitySlug       string    `json:"city_slug,omitempty"`
	StudentCount   int       `json:"student_count"`
	ScoreRaw       float64   `json:"score_raw"`
	ScorePerCapita float64   `json:"score_per_capita"`
	ActiveUsers    int       `json:"active_users"`
	Deployments    int       `json:"deployments"`
	Executions30d  int64     `json:"executions_30d"`
	NewUsers30d    int       `json:"new_users_30d"`
	PeriodEnd      time.Time `json:"period_end"`
}

// ToRankingEntries converts the storage-layer rankings to DTOs.
func ToRankingEntries(rs []universityranking.Ranking) []RankingEntry {
	out := make([]RankingEntry, 0, len(rs))
	for _, r := range rs {
		out = append(out, ToRankingEntry(r))
	}
	return out
}

func ToRankingEntry(r universityranking.Ranking) RankingEntry {
	return RankingEntry{
		Rank:           r.RankPosition,
		PreviousRank:   r.PrevRank,
		RankDelta:      r.RankDelta,
		UniversityID:   r.UniversityID,
		Slug:           r.UniversitySlug,
		Name:           r.UniversityName,
		ShortName:      r.ShortName,
		CountryCode:    r.CountryCode,
		StateCode:      r.StateCode,
		CitySlug:       r.CitySlug,
		StudentCount:   0, // not stored on Ranking; left for callers that join
		ScoreRaw:       r.ScoreRaw,
		ScorePerCapita: r.ScorePerCapita,
		ActiveUsers:    r.ActiveUsers,
		Deployments:    r.Deployments,
		Executions30d:  r.Executions30d,
		NewUsers30d:    r.NewUsers30d,
		PeriodEnd:      r.PeriodEnd,
	}
}

// UniversityEntry is the static part of a university (used in /me, /resolve).
type UniversityEntry struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	ShortName       string `json:"short_name,omitempty"`
	CountryCode     string `json:"country_code"`
	StateCode       string `json:"state_code,omitempty"`
	CityID          *int64 `json:"city_id,omitempty"`
	StudentCount    int    `json:"student_count"`
	InstitutionType string `json:"institution_type"`
	Website         string `json:"website,omitempty"`
}

func ToUniversityEntry(u universityranking.University) UniversityEntry {
	return UniversityEntry{
		ID:              u.ID,
		Slug:            u.Slug,
		Name:            u.Name,
		ShortName:       u.ShortName,
		CountryCode:     u.CountryCode,
		StateCode:       u.StateCode,
		CityID:          u.CityID,
		StudentCount:    u.StudentCount,
		InstitutionType: u.InstitutionType,
		Website:         u.Website,
	}
}

// UniversityDetailResponse is the payload for /university-rankings/{slug}.
type UniversityDetailResponse struct {
	Slug     string       `json:"slug"`
	Entry    RankingEntry `json:"entry"`
	CacheHit bool         `json:"cache_hit"`
}

func ToDetailResponse(slug string, rk *universityranking.Ranking, cacheHit bool) UniversityDetailResponse {
	return UniversityDetailResponse{
		Slug:     slug,
		Entry:    ToRankingEntry(*rk),
		CacheHit: cacheHit,
	}
}

// MyUniversityResponse is the payload for /university-rankings/me.
type MyUniversityResponse struct {
	Source     string          `json:"source"`
	University UniversityEntry `json:"university,omitempty"`
	Ranking    RankingEntry    `json:"ranking,omitempty"`
}

// ResolutionResponse is the payload for POST /university-rankings/resolve.
type ResolutionResponse struct {
	University UniversityEntry `json:"university"`
	Source     string          `json:"source"`
	Ambiguous  bool            `json:"ambiguous"`
	MatchCount int             `json:"match_count"`
}

type OptOutResponse struct {
	OptedOut bool `json:"opted_out"`
}

type SetOptOutRequest struct {
	OptedOut *bool `json:"opted_out"`
}
