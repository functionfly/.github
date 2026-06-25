package cityranking

import (
	"time"

	crstorage "github.com/functionfly/functionfly/internal/storage/cityranking"
)

// ── Response shapes ───────────────────────────────────────────────────────

// StateRankingEntry is one row in the state leaderboard.
type StateRankingEntry struct {
	Rank          int       `json:"rank"`
	StateCode     string    `json:"state_code"`
	StateName     string    `json:"state_name"`
	CountryCode   string    `json:"country_code"`
	Population    int       `json:"population"`
	ScoreRaw      float64   `json:"score_raw"`
	ScorePerCapita float64  `json:"score_per_capita"`
	ActiveUsers   int       `json:"active_users"`
	Deployments   int       `json:"deployments"`
	Executions30d int64     `json:"executions_30d"`
	MetroCount    int       `json:"metro_count"`
	RankedMetros  int       `json:"ranked_metros"`
	PeriodEnd     time.Time `json:"period_end"`
}

type StatesLeaderboardResponse struct {
	PeriodEnd   time.Time           `json:"period_end"`
	TotalStates int                 `json:"total_states"`
	Entries     []StateRankingEntry `json:"entries"`
	Country     string              `json:"country,omitempty"`
	Category    string              `json:"category"`
	CacheHit    bool                `json:"cache_hit"`
}

type StateDetailResponse struct {
	Current               *StateRankingEntry `json:"current"`
	PrivacyMinActiveUsers int                `json:"privacy_min_active_users"`
	NotRanked             bool               `json:"not_ranked"`
	PeriodEnd             time.Time          `json:"period_end"`
	Category              string             `json:"category"`
	CacheHit              bool               `json:"cache_hit"`
}

// MapPointEntry is one ranked metro projected to a (lat, lon) for the map.
type MapPointEntry struct {
	MetroSlug      string  `json:"metro_slug"`
	MetroName      string  `json:"metro_name"`
	CountryCode    string  `json:"country_code"`
	StateCode      string  `json:"state_code"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Population     int     `json:"population"`
	RankPosition   int     `json:"rank_position"`
	ScorePerCapita float64 `json:"score_per_capita"`
	ActiveUsers    int     `json:"active_users"`
	Tier           string  `json:"tier"`
}

type MapPointsResponse struct {
	PeriodEnd time.Time       `json:"period_end"`
	Points    []MapPointEntry `json:"points"`
	Category  string          `json:"category"`
	CacheHit  bool            `json:"cache_hit"`
}

// CityRankingEntry is the public row in the leaderboard JSON.
type CityRankingEntry struct {
	Rank             int       `json:"rank"`
	PreviousRank     int       `json:"previous_rank"`
	RankDelta        int       `json:"rank_delta"`
	MetroSlug        string    `json:"metro_slug"`
	MetroName        string    `json:"metro_name"`
	CountryCode      string    `json:"country_code"`
	Population       int       `json:"population"`
	ScoreRaw         float64   `json:"score_raw"`
	ScorePerCapita   float64   `json:"score_per_capita"`
	ActiveUsers      int       `json:"active_users"`
	Deployments      int       `json:"deployments"`
	Executions30d    int64     `json:"executions_30d"`
	FounderEarningsC int64     `json:"founder_earnings_cents"`
	NewUsers30d      int       `json:"new_users_30d"`
	PeriodEnd        time.Time `json:"period_end"`
}

type LeaderboardResponse struct {
	PeriodEnd   time.Time          `json:"period_end"`
	TotalRanked int                `json:"total_ranked"`
	Entries     []CityRankingEntry `json:"entries"`
	Country     string             `json:"country,omitempty"`
	Category    string             `json:"category"`
	CacheHit    bool               `json:"cache_hit"`
}

type CategoriesResponse struct {
	Categories []crstorage.CategoryMeta `json:"categories"`
}

type MetroDetailResponse struct {
	Current               *CityRankingEntry  `json:"current"`
	History               []CityRankingEntry `json:"history"`
	NotRanked             bool               `json:"not_ranked"`
	PrivacyMinActiveUsers int                `json:"privacy_min_active_users"`
	PeriodEnd             time.Time          `json:"period_end"`
	Category              string             `json:"category"`
	CacheHit              bool               `json:"cache_hit"`
}

type MoversResponse struct {
	Direction string             `json:"direction"`
	PeriodEnd time.Time          `json:"period_end"`
	Entries   []CityRankingEntry `json:"entries"`
	Category  string             `json:"category"`
}

type MyCityResponse struct {
	HasCity  bool              `json:"has_city"`
	Metro    *CityRankingEntry `json:"metro,omitempty"`
	OptedOut bool              `json:"opted_out"`
}

type CityResolution struct {
	CityID        int64  `json:"city_id"`
	CitySlug      string `json:"city_slug"`
	CityName      string `json:"city_name"`
	MetroSlug     string `json:"metro_slug"`
	Source        string `json:"source"`
	Ambiguous     bool   `json:"ambiguous"`
	PendingReview bool   `json:"pending_review"`
}

type OptOutResponse struct {
	OptedOut bool `json:"opted_out"`
}

type SetOptOutRequest struct {
	OptedOut *bool `json:"opted_out"`
}

// CityEntry is one row in the city-proper leaderboard (toggle off the MSA
// rollup via the /cities endpoint).
type CityEntry struct {
	Rank           int       `json:"rank"`
	CityID         int64     `json:"city_id"`
	CitySlug       string    `json:"city_slug"`
	CityName       string    `json:"city_name"`
	StateCode      string    `json:"state_code"`
	StateName      string    `json:"state_name"`
	CountryCode    string    `json:"country_code"`
	CountryName    string    `json:"country_name"`
	MetroSlug      string    `json:"metro_slug,omitempty"`
	MetroName      string    `json:"metro_name,omitempty"`
	Population     int       `json:"population"`
	ScoreRaw       float64   `json:"score_raw"`
	ScorePerCapita float64   `json:"score_per_capita"`
	ActiveUsers    int       `json:"active_users"`
	Deployments    int       `json:"deployments"`
	Executions30d  int64     `json:"executions_30d"`
	NewUsers30d    int       `json:"new_users_30d"`
}

type CitiesLeaderboardResponse struct {
	PeriodEnd    time.Time    `json:"period_end"`
	TotalRanked  int          `json:"total_ranked"`
	Entries      []CityEntry  `json:"entries"`
	Country      string       `json:"country,omitempty"`
	Category     string       `json:"category"`
	CacheHit     bool         `json:"cache_hit"`
	PrivacyMin   int          `json:"privacy_min_active_users"`
}

// BuilderEntry is one anonymized top contributor in a metro. k-anonymity is
// enforced at the repository layer.
type BuilderEntry struct {
	Rank            int     `json:"rank"`
	UserID          string  `json:"user_id"`
	DisplayName     string  `json:"display_name"`
	ProfileNumber   *int    `json:"profile_number,omitempty"`
	ProfilePublic   bool    `json:"profile_public"`
	Deployments     int     `json:"deployments"`
	Executions30d   int64   `json:"executions_30d"`
	ScoreComposite  float64 `json:"score_composite"`
}

type BuildersResponse struct {
	MetroSlug         string         `json:"metro_slug"`
	PeriodEnd         time.Time      `json:"period_end"`
	Entries           []BuilderEntry `json:"entries"`
	Category          string         `json:"category"`
	PrivacySuppressed bool           `json:"privacy_suppressed"`
}

// IPGeoCity is the city portion of an IP-geo lookup.
type IPGeoCity struct {
	CityID      int64  `json:"city_id"`
	CitySlug    string `json:"city_slug"`
	CityName    string `json:"city_name"`
	StateCode   string `json:"state_code"`
	CountryCode string `json:"country_code"`
	MetroSlug   string `json:"metro_slug"`
	MetroName   string `json:"metro_name"`
	Population  int    `json:"population"`
}

// IPGeoResponse is the public payload for /city-rankings/resolve-by-ip.
type IPGeoResponse struct {
	City        *IPGeoCity `json:"city,omitempty"`
	CountryCode string     `json:"country_code"`
	Source      string     `json:"source"` // "ip" or "default"
	NotFound    bool       `json:"not_found"`
}

// AmbassadorEntry is one row in the public ambassadors list. Email is
// intentionally omitted to prevent scraping — the front-end links to
// the public profile via profile_number.
type AmbassadorEntry struct {
	MetroID       int64     `json:"metro_id"`
	MetroSlug     string    `json:"metro_slug"`
	MetroName     string    `json:"metro_name"`
	CountryCode   string    `json:"country_code"`
	StateCode     string    `json:"state_code,omitempty"`
	CitySlug      string    `json:"city_slug,omitempty"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	Username      *string   `json:"username,omitempty"`
	ProfileNumber *int      `json:"profile_number,omitempty"`
	PromotedAt    time.Time `json:"promoted_at"`
	Source        string    `json:"source"` // "auto" | "manual"
}

// AmbassadorsListResponse is the payload for /city-rankings/ambassadors.
type AmbassadorsListResponse struct {
	Total      int                `json:"total"`
	Entries    []AmbassadorEntry  `json:"entries"`
	Country    string             `json:"country,omitempty"`
	PrivacyMin int                `json:"privacy_min_active_users"`
}

// CountriesResponse is the payload for /city-rankings/countries.
type CountriesResponse struct {
	Countries []CountryEntry `json:"countries"`
}

// CountryEntry represents a single country option for filtering.
type CountryEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// AmbassadorResponse is the payload for /city-rankings/{slug}/ambassador.
type AmbassadorResponse struct {
	MetroSlug   string          `json:"metro_slug"`
	MetroName   string          `json:"metro_name"`
	CountryCode string          `json:"country_code"`
	Ambassador  AmbassadorEntry `json:"ambassador"`
}

// ToAmbassadorEntries converts storage-layer entries to public DTOs.
func ToAmbassadorEntries(rs []crstorage.AmbassadorListEntry) []AmbassadorEntry {
	out := make([]AmbassadorEntry, 0, len(rs))
	for _, r := range rs {
		out = append(out, ToAmbassadorEntryFromList(r))
	}
	return out
}

func ToAmbassadorEntryFromList(r crstorage.AmbassadorListEntry) AmbassadorEntry {
	return AmbassadorEntry{
		MetroID:       r.MetroID,
		MetroSlug:     r.MetroSlug,
		MetroName:     r.MetroName,
		CountryCode:   r.CountryCode,
		StateCode:     r.StateCode,
		CitySlug:      r.CitySlug,
		UserID:        r.UserID,
		Name:          r.FullName,
		Username:      r.Username,
		ProfileNumber: r.ProfileNumber,
		PromotedAt:    r.PromotedAt,
		Source:        r.Source,
	}
}

func ToAmbassadorEntry(a crstorage.Ambassador) AmbassadorEntry {
	return AmbassadorEntry{
		MetroID:       a.MetroID,
		UserID:        a.UserID,
		Name:          a.FullName,
		ProfileNumber: a.ProfileNumber,
		PromotedAt:    a.PromotedAt,
		Source:        a.Source,
	}
}

// ── Conversions ───────────────────────────────────────────────────────────

func ToEntry(r crstorage.Ranking) CityRankingEntry {
	delta := 0
	if r.PrevRank > 0 {
		delta = r.PrevRank - r.RankPosition
	}
	return CityRankingEntry{
		Rank:             r.RankPosition,
		PreviousRank:     r.PrevRank,
		RankDelta:        delta,
		MetroSlug:        r.MetroSlug,
		MetroName:        r.MetroName,
		CountryCode:      r.CountryCode,
		Population:       r.Population,
		ScoreRaw:         r.ScoreRaw,
		ScorePerCapita:   r.ScorePerCapita,
		ActiveUsers:      r.ActiveUsers,
		Deployments:      r.Deployments,
		Executions30d:    r.Executions30d,
		FounderEarningsC: r.FounderEarnings,
		NewUsers30d:      r.NewUsers30d,
		PeriodEnd:        r.PeriodEnd,
	}
}

func ToEntries(rs []crstorage.Ranking) []CityRankingEntry {
	out := make([]CityRankingEntry, 0, len(rs))
	for _, r := range rs {
		out = append(out, ToEntry(r))
	}
	return out
}

func ToStateEntry(s crstorage.StateRanking) StateRankingEntry {
	return StateRankingEntry{
		Rank:           s.RankPosition,
		StateCode:      s.StateCode,
		StateName:      s.StateName,
		CountryCode:    s.CountryCode,
		Population:     s.Population,
		ScoreRaw:       s.ScoreRaw,
		ScorePerCapita: s.ScorePerCapita,
		ActiveUsers:    s.ActiveUsers,
		Deployments:    s.Deployments,
		Executions30d:  s.Executions30d,
		MetroCount:     s.MetroCount,
		RankedMetros:   s.RankedMetros,
		PeriodEnd:      s.PeriodEnd,
	}
}

func ToStateEntries(ss []crstorage.StateRanking) []StateRankingEntry {
	out := make([]StateRankingEntry, 0, len(ss))
	for _, s := range ss {
		out = append(out, ToStateEntry(s))
	}
	return out
}

func ToMapPoint(p crstorage.MapPoint) MapPointEntry {
	return MapPointEntry{
		MetroSlug:      p.MetroSlug,
		MetroName:      p.MetroName,
		CountryCode:    p.CountryCode,
		StateCode:      p.StateCode,
		Latitude:       p.Latitude,
		Longitude:      p.Longitude,
		Population:     p.Population,
		RankPosition:   p.RankPosition,
		ScorePerCapita: p.ScorePerCapita,
		ActiveUsers:    p.ActiveUsers,
		Tier:           p.Tier,
	}
}

func ToMapPoints(ps []crstorage.MapPoint) []MapPointEntry {
	out := make([]MapPointEntry, 0, len(ps))
	for _, p := range ps {
		out = append(out, ToMapPoint(p))
	}
	return out
}

// ToCityEntries converts the storage-layer city-proper rankings into the
// public DTO, dropping the per-row internal IDs.
func ToCityEntries(rs []crstorage.CityRanking) []CityEntry {
	out := make([]CityEntry, 0, len(rs))
	for i, r := range rs {
		entry := CityEntry{
			Rank:           i + 1,
			CityID:         r.CityID,
			CitySlug:       r.CitySlug,
			CityName:       r.CityName,
			StateCode:      r.StateCode,
			StateName:      r.StateName,
			CountryCode:    r.CountryCode,
			CountryName:    r.CountryName,
			Population:     r.Population,
			ScoreRaw:       r.ScoreRaw,
			ScorePerCapita: r.ScorePerCapita,
			ActiveUsers:    r.ActiveUsers,
			Deployments:    r.Deployments,
			Executions30d:  r.Executions30d,
			NewUsers30d:    r.NewUsers30d,
		}
		if r.MetroSlug != nil {
			entry.MetroSlug = *r.MetroSlug
		}
		if r.MetroName != nil {
			entry.MetroName = *r.MetroName
		}
		out = append(out, entry)
	}
	return out
}

func ToBuilderEntries(bs []crstorage.Builder) []BuilderEntry {
	out := make([]BuilderEntry, 0, len(bs))
	for _, b := range bs {
		out = append(out, BuilderEntry{
			Rank:           b.Rank,
			UserID:         b.UserID,
			DisplayName:    b.DisplayName,
			ProfileNumber:  b.ProfileNumber,
			ProfilePublic:  b.ProfilePublic,
			Deployments:    b.Deployments,
			Executions30d:  b.Executions30d,
			ScoreComposite: b.ScoreComposite,
		})
	}
	return out
}

type MetroUniversitiesResponse struct {
	MetroSlug string                        `json:"metro_slug"`
	Entries   []crstorage.UniversityInMetro `json:"entries"`
	Total     int                           `json:"total"`
}

// CityReviewEntry is one city awaiting admin review.
type CityReviewEntry struct {
	CityID      int64     `json:"city_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	StateCode   string    `json:"state_code"`
	StateName   string    `json:"state_name"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Population  int       `json:"population"`
	MetroName   string    `json:"metro_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	AliasCount  int       `json:"alias_count"`
}

// CityReviewDetail is full city details for admin review.
type CityReviewDetail struct {
	CityID      int64      `json:"city_id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	StateCode   string     `json:"state_code"`
	StateName   string     `json:"state_name"`
	CountryCode string     `json:"country_code"`
	CountryName string     `json:"country_name"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Population  int        `json:"population"`
	ReviewStatus string    `json:"review_status"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy  *string    `json:"reviewed_by,omitempty"`
	ReviewNotes *string     `json:"review_notes,omitempty"`
	MetroName   string     `json:"metro_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	Aliases     []CityAliasEntry `json:"aliases"`
}

// CityAliasEntry is an alias for admin display.
type CityAliasEntry struct {
	Alias   string    `json:"alias"`
	Source  string    `json:"source"`
	Created time.Time `json:"created"`
}

// CityReviewListResponse is the payload for listing pending reviews.
type CityReviewListResponse struct {
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	Cities []CityReviewEntry `json:"cities"`
}

// CityAdminListResponse is the payload for listing all cities in admin.
type CityAdminListResponse struct {
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
	Cities []crstorage.CityAdminListEntry `json:"cities"`
}

// ReviewCityRequest is the body for approving/rejecting a city.
type ReviewCityRequest struct {
	Status string `json:"status"` // "approved" or "rejected"
	Notes  string `json:"notes,omitempty"`
}

// CityReviewResponse is the result of a review action.
type CityReviewResponse struct {
	CityID      int64  `json:"city_id"`
	Slug        string `json:"slug"`
	ReviewStatus string `json:"review_status"`
	Message     string `json:"message"`
}
