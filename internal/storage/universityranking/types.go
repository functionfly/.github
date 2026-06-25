// Package universityranking stores and ranks universities by the activity
// of their affiliated users. It mirrors the cityranking package: same
// scoring math, same privacy threshold, same Redis cache layer, but keyed
// on `university_id` instead of `metro_area_id`.
//
// See docs/UNIVERSITY_RANKINGS.md for the architecture rationale and
// plan .kilo/plans/1782018734195-city-rankings-plan.md §8 #5 for the
// future-work item this implements.
package universityranking

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// MinActiveUsersForPublic mirrors cityranking.MinActiveUsersForPublic. We
// keep the same k=5 threshold so the two leaderboards have consistent
// privacy guarantees for users.
const MinActiveUsersForPublic = 5

// Category is the ranking category (composite + the four sub-categories).
// Same shape as cityranking.Category so the scorer / cache / jobs code can
// be unified later.
type Category string

const (
	CategoryComposite  Category = "composite"
	CategoryAgents     Category = "agents"
	CategoryAutomation Category = "automation"
	CategoryStartups   Category = "startups"
	CategoryOpenSource Category = "open_source"
	CategoryRobotics  Category = "robotics"
)

// AllCategories returns every ranking category the recompute job writes.
func AllCategories() []Category {
	return []Category{CategoryComposite, CategoryAgents, CategoryAutomation, CategoryStartups, CategoryOpenSource, CategoryRobotics}
}

// ValidCategory reports whether s is a known category.
func ValidCategory(s string) bool {
	for _, c := range AllCategories() {
		if string(c) == s {
			return true
		}
	}
	return false
}

// CategoryWeights returns the per-category weight vector. Identical to
// cityranking.CategoryWeights so the same scoring math works. Kept in sync
// manually — the future refactor is to share the weights via a single
// `rankingscore` package.
func CategoryWeights(c Category) (active, deploy, exec, founders, newUsers float64) {
	switch c {
	case CategoryComposite:
		return 1.0, 1.5, 0.5, 2.0, 1.2
	case CategoryAgents:
		return 1.5, 1.0, 2.0, 1.0, 0.8
	case CategoryAutomation:
		return 1.0, 2.0, 1.5, 0.5, 0.8
	case CategoryStartups:
		return 0.8, 1.0, 0.5, 3.0, 2.0
	case CategoryOpenSource:
		return 1.2, 1.5, 1.0, 0.5, 1.0
	case CategoryRobotics:
		return 1.0, 2.0, 2.0, 0.5, 0.8
	}
	return 1, 1, 1, 1, 1
}

// University is the core record.
type University struct {
	ID             int64     `json:"id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	ShortName      string    `json:"short_name,omitempty"`
	CountryCode    string    `json:"country_code"`
	StateCode      string    `json:"state_code,omitempty"`
	CityID         *int64    `json:"city_id,omitempty"`
	StudentCount   int       `json:"student_count"`
	InstitutionType string   `json:"institution_type"`
	Website        string    `json:"website,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// Ranking is one materialized row in university_rankings.
type Ranking struct {
	UniversityID    int64     `json:"university_id"`
	UniversitySlug  string    `json:"university_slug"`
	UniversityName  string    `json:"university_name"`
	ShortName       string    `json:"short_name,omitempty"`
	CountryCode     string    `json:"country_code"`
	StateCode       string    `json:"state_code,omitempty"`
	CitySlug        string    `json:"city_slug,omitempty"`
	RankPosition    int       `json:"rank_position"`
	PrevRank        *int      `json:"prev_rank_position,omitempty"`
	RankDelta       int       `json:"rank_delta"`
	ScoreRaw        float64   `json:"score_raw"`
	ScorePerCapita  float64   `json:"score_per_capita"`
	ActiveUsers     int       `json:"active_users"`
	Deployments     int       `json:"deployments"`
	Executions30d   int64     `json:"executions_30d"`
	FounderEarnings int64     `json:"founder_earnings"`
	NewUsers30d     int       `json:"new_users_30d"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
}

// Signals is the raw activity vector for a single university. Same shape
// as cityranking.Signals so the scoring formula transfers one-to-one.
type Signals struct {
	ActiveUsers    int
	Deployments    int
	Executions30d  int64
	FounderEarnings int64
	NewUsers30d    int
}

// Score is the output of the scoring function.
type Score struct {
	Raw         float64
	PerCapita   float64
	ActiveUsers int
	Deployments int
	Executions30d int64
	NewUsers30d int
}

// Period is (start, end) tuple for the rolling 30-day window.
type Period struct {
	Start time.Time
	End   time.Time
}

// Repository wraps a pgx pool + logger.
type Repository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewRepository wires a Repository.
func NewRepository(pool *pgxpool.Pool, log *logrus.Logger) *Repository {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Repository{pool: pool, log: log}
}
