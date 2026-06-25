package companyranking

import (
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// MinActiveUsersForPublic is the k-anonymity threshold for company rankings.
const MinActiveUsersForPublic = 3 // companies typically have fewer users than metros

// Category is the ranking category (mirrors city/university for consistent scoring).
type Category string

const (
	CategoryComposite  Category = "composite"
	CategoryAgents     Category = "agents"
	CategoryAutomation Category = "automation"
	CategoryStartups   Category = "startups"
	CategoryOpenSource Category = "open_source"
	CategoryRobotics  Category = "robotics"
)

var AllCategories = []Category{
	CategoryComposite,
	CategoryAgents,
	CategoryAutomation,
	CategoryStartups,
	CategoryOpenSource,
	CategoryRobotics,
}

func ValidCategory(s string) bool {
	for _, c := range AllCategories {
		if string(c) == s {
			return true
		}
	}
	return false
}

func CategoryWeights(c Category) (active, deploy, exec, revenue, newUsers float64) {
	switch c {
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
	return 1.0, 1.5, 0.5, 2.0, 1.2 // composite
}

// Company is a business on FunctionFly.
type Company struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	CityID       *int64    `json:"city_id,omitempty"`
	CountryCode  string    `json:"country_code"`
	EmployeeCount int      `json:"employee_count,omitempty"`
	Industry     string    `json:"industry,omitempty"`
	Website      string    `json:"website,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Ranking is one materialized row in company_rankings.
type Ranking struct {
	CompanyID     int64     `json:"company_id"`
	CompanySlug  string    `json:"company_slug"`
	CompanyName  string    `json:"company_name"`
	CountryCode  string    `json:"country_code"`
	CitySlug     string    `json:"city_slug,omitempty"`
	EmployeeCount int      `json:"employee_count,omitempty"`
	RankPosition int       `json:"rank_position"`
	PrevRank     *int      `json:"prev_rank,omitempty"`
	RankDelta    int       `json:"rank_delta"`
	ScoreRaw     float64   `json:"score_raw"`
	ScorePerCapita float64 `json:"score_per_capita"`
	ActiveUsers  int       `json:"active_users"`
	Deployments  int       `json:"deployments"`
	Executions30d int64    `json:"executions_30d"`
	RevenueCents int64    `json:"revenue_cents"`
	NewUsers30d  int       `json:"new_users_30d"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Category    Category   `json:"category"`
}

// Signals contains the raw activity signals used for scoring.
type Signals struct {
	ActiveUsers    int
	Deployments   int
	Executions30d int64
	RevenueCents  int64
	NewUsers30d   int
}

// Weights defines the per-category weight vector.
type Weights struct {
	ActiveUsers   float64
	Deployments   float64
	Executions30d float64
	Revenue      float64
	NewUsersGrowth float64
}

// DefaultWeights returns the composite weights.
func DefaultWeights() Weights {
	return Weights{ActiveUsers: 1.0, Deployments: 1.5, Executions30d: 0.5, Revenue: 2.0, NewUsersGrowth: 1.2}
}

// Compute calculates the raw and per-capita scores from signals and weights.
func Compute(s Signals, employeeCount int, w Weights) (rawScore, perCapita float64) {
	rawScore = w.ActiveUsers * math.Log10(float64(s.ActiveUsers)+1) +
		w.Deployments * math.Log10(float64(s.Deployments)+1) +
		w.Executions30d * math.Log10(float64(s.Executions30d)+1) +
		w.Revenue * math.Log10(float64(s.RevenueCents)/100.0+1) +
		w.NewUsersGrowth * math.Log10(float64(s.NewUsers30d)+1)

	if employeeCount > 0 {
		perCapita = rawScore * 100000.0 / float64(employeeCount)
	} else {
		perCapita = rawScore
	}
	return rawScore, perCapita
}

// Tier returns the display tier for a per-capita score.
func Tier(perCapita float64) string {
	if perCapita >= 0.20 {
		return "gold"
	}
	if perCapita >= 0.05 {
		return "blue"
	}
	return "green"
}

// Repository handles company ranking database access.
type Repository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewRepository creates a company ranking repository.
func NewRepository(pool *pgxpool.Pool, log *logrus.Logger) *Repository {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Repository{pool: pool, log: log}
}

func (r *Repository) Pool() *pgxpool.Pool { return r.pool }
