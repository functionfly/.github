package universityranking

import (
	"math"
	"sort"
	"time"
)

// Compute is the pure-math scoring function. It is shared with the city
// leaderboard (same Signals, same scoring, same per-capita formula) so the
// two leaderboards can never drift in how they weight activity. Active
// users and new users are log-scaled to dampen heavy-tail effects;
// deployments and executions are linear (with log dampening past 1000 to
// keep a single mega-deployer from dominating); founder earnings is
// square-root scaled.
func Compute(s Signals, population int, wActive, wDeploy, wExec, wFounders, wNewUsers float64) Score {
	if population < 1 {
		population = 1
	}
	activeTerm := math.Log1p(float64(s.ActiveUsers)) * wActive
	deployTerm := math.Log1p(float64(s.Deployments)) * wDeploy
	// Sub-linear past 1000: full credit for the first 1000 executions, then
	// log-shaped growth so a 10x multiplier doesn't yield 10x score.
	execTerm := math.Log1p(float64(s.Executions30d)/1000) * wExec
	foundersTerm := math.Sqrt(float64(s.FounderEarnings)/1e6) * wFounders
	newUsersTerm := math.Log1p(float64(s.NewUsers30d)) * wNewUsers
	raw := activeTerm + deployTerm + execTerm + foundersTerm + newUsersTerm
	perCapita := raw * 100000.0 / float64(population)
	return Score{
		Raw:          raw,
		PerCapita:    perCapita,
		ActiveUsers:  s.ActiveUsers,
		Deployments:  s.Deployments,
		Executions30d: s.Executions30d,
		NewUsers30d:  s.NewUsers30d,
	}
}

// DefaultWeights returns the category-default weight vector. Equivalent to
// `CategoryWeights(CategoryComposite)`.
func DefaultWeights() (active, deploy, exec, founders, newUsers float64) {
	return CategoryWeights(CategoryComposite)
}

// SortRankings sorts rankings in-place by score_per_capita descending. Used
// after the privacy filter so the in-memory result matches the rank
// numbers we send down the wire.
func SortRankings(rs []Ranking) {
	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].ScorePerCapita > rs[j].ScorePerCapita
	})
}

// TruncateHour is shared with the city recompute job so the two stay on
// the same time grid (e.g. 18:00:00 UTC vs 18:00:00.123 UTC). Without it
// the period_end values would drift and the cache key would churn.
func TruncateHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}
