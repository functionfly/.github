package cityranking

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CityReviewStatus string

const (
	CityReviewStatusSeed     CityReviewStatus = "seed"
	CityReviewStatusApproved CityReviewStatus = "approved"
	CityReviewStatusPending  CityReviewStatus = "pending"
	CityReviewStatusRejected CityReviewStatus = "rejected"
)

const DefaultAutoReviewPopulationThreshold = 100000

// City is the public-facing record for a single city, including the metro
// area it belongs to (nullable for cities without a metro).
type City struct {
	ID                       int64           `json:"id"`
	Slug                     string          `json:"slug"`
	Name                     string          `json:"name"`
	StateCode                string          `json:"state_code"`
	StateName                string          `json:"state_name"`
	CountryCode              string          `json:"country_code"`
	CountryName              string          `json:"country_name"`
	Latitude                 float64         `json:"latitude"`
	Longitude                float64         `json:"longitude"`
	Population               int             `json:"population"`
	ReviewStatus             CityReviewStatus `json:"review_status"`
	AutoReviewPopThreshold   int              `json:"auto_review_pop_threshold,omitempty"`
	ReviewedAt               *time.Time       `json:"reviewed_at,omitempty"`
	ReviewedBy               *string          `json:"reviewed_by,omitempty"`
	ReviewNotes              string          `json:"review_notes,omitempty"`
	MetroAreaID              *int64          `json:"metro_area_id,omitempty"`
	MetroSlug                *string         `json:"metro_slug,omitempty"`
	MetroName                *string         `json:"metro_name,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
}

// MetroArea is the ranking aggregation unit (typically an MSA or non-US
// equivalent).
type MetroArea struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	CountryCode string    `json:"country_code"`
	Population  int       `json:"population"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// Ranking is one row in the city_rankings table — the materialized, hourly
// computed score for a metro under a single category.
type Ranking struct {
	ID                int64         `json:"id"`
	MetroAreaID       int64         `json:"metro_area_id"`
	MetroSlug         string        `json:"metro_slug"`
	MetroName         string        `json:"metro_name"`
	CountryCode       string        `json:"country_code"`
	Population        int           `json:"population"`
	Category          Category      `json:"category"`
	RankPosition      int           `json:"rank_position"`
	PrevRankPosition  sql.NullInt64 `json:"-"`
	PrevRank          int           `json:"prev_rank_position"`
	ScoreRaw          float64       `json:"score_raw"`
	ScorePerCapita    float64       `json:"score_per_capita"`
	ActiveUsers       int           `json:"active_users"`
	Deployments       int           `json:"deployments"`
	Executions30d     int64         `json:"executions_30d"`
	FounderEarnings   int64         `json:"founder_earnings_cents"`
	NewUsers30d       int           `json:"new_users_30d"`
	PeriodStart       time.Time     `json:"period_start"`
	PeriodEnd         time.Time     `json:"period_end"`
	ComputedAt        time.Time     `json:"computed_at"`
}

// StateRanking is the rolled-up score for a (state_code, country_code) pair.
// ScoreRaw is the sum of metro scores; ScorePerCapita is normalized to a
// 100k-population base across the whole state.
type StateRanking struct {
	StateCode      string    `json:"state_code"`
	StateName      string    `json:"state_name"`
	CountryCode    string    `json:"country_code"`
	Population     int       `json:"population"`
	RankPosition   int       `json:"rank_position"`
	ScoreRaw       float64   `json:"score_raw"`
	ScorePerCapita float64   `json:"score_per_capita"`
	ActiveUsers    int       `json:"active_users"`
	Deployments    int       `json:"deployments"`
	Executions30d  int64     `json:"executions_30d"`
	MetroCount     int       `json:"metro_count"`
	RankedMetros   int       `json:"ranked_metros"`
	PeriodEnd      time.Time `json:"period_end"`
}

// MapPoint is a ranked metro projected to a single (lat, lon) for the
// AI World Map visualization. The tier is computed on the read path from
// the per-capita score.
type MapPoint struct {
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

// Country represents a distinct country with active metros.
type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Category is a ranking sub-axis. The v1 categories are:
//   - composite   : the original weighted sum of every signal
//   - agents      : emphasizes active users, executions, and growth (agent adoption)
//   - automation  : emphasizes deployments and executions (workflow volume)
//   - startups     : emphasizes founder earnings and new users (early-stage velocity)
//   - open_source  : emphasizes deployments and executions (registry activity)
//   - robotics     : emphasizes robotics functions, robot deployments, hardware integrations
type Category string

const (
	CategoryComposite  Category = "composite"
	CategoryAgents     Category = "agents"
	CategoryAutomation Category = "automation"
	CategoryStartups   Category = "startups"
	CategoryOpenSource Category = "open_source"
	CategoryRobotics  Category = "robotics"
)

// AllCategories is the canonical, ordered list of categories the recompute
// job iterates over. Order matters: composite is first because it is the
// "default" leaderboard.
var AllCategories = []Category{
	CategoryComposite,
	CategoryAgents,
	CategoryAutomation,
	CategoryStartups,
	CategoryOpenSource,
	CategoryRobotics,
}

// IsValidCategory reports whether s is one of the shipped category slugs.
func IsValidCategory(s string) bool {
	switch Category(s) {
	case CategoryComposite, CategoryAgents, CategoryAutomation, CategoryStartups, CategoryOpenSource, CategoryRobotics:
		return true
	}
	return false
}

// CategoryMeta is the public-facing metadata for a ranking category.
type CategoryMeta struct {
	Slug        string `json:"slug"`
	Label       string `json:"label"`
	Description string `json:"description"`
	// Default weights used by the recompute job.
	Weights Weights `json:"weights"`
}

// UserCityAssignment is the cached record of which city a user is counted in.
type UserCityAssignment struct {
	UserID   uuid.UUID `json:"user_id"`
	CityID   int64     `json:"city_id"`
	CitySlug string    `json:"city_slug"`
	CityName string    `json:"city_name"`
	MetroID  *int64    `json:"metro_area_id,omitempty"`
	Source   string    `json:"source"`
}
