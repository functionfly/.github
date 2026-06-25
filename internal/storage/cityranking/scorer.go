package cityranking

import (
	"math"
	"time"
)

// Weights controls the relative importance of each activity signal in the
// composite score. Sum is normalized to 100 at the call sites; default values
// match the v1 weights in docs/CITY_RANKINGS.md.
type Weights struct {
	ActiveUsers    int
	Deployments    int
	Executions30d  int
	Founder        int
	NewUsersGrowth int
}

// DefaultWeights returns the v1 weights from the plan.
func DefaultWeights() Weights {
	return Weights{
		ActiveUsers:    30,
		Deployments:    25,
		Executions30d:  20,
		Founder:        15,
		NewUsersGrowth: 10,
	}
}

// Signals captures the raw counts that feed into a city's composite score.
type Signals struct {
	ActiveUsers     int
	Deployments     int
	Executions30d   int
	FounderEarnings int64 // cents
	NewUsers30d     int
}

// ScoreResult is the output of computing a city's score for one period.
type ScoreResult struct {
	Raw           float64
	PerCapita     float64
	ActiveUsers   int
	Deployments   int
	Executions30d int64
	FounderCents  int64
	NewUsers30d   int
}

// Compute returns the raw and per-capita scores for a metro using log10 scaling.
// Per-capita is normalized to a 100k-population base so the headline score is
// comparable across cities of very different sizes. A population <= 0 falls
// back to the raw score to avoid division by zero.
func Compute(signals Signals, population int, w Weights) ScoreResult {
	totalWeight := float64(w.ActiveUsers + w.Deployments + w.Executions30d + w.Founder + w.NewUsersGrowth)
	if totalWeight <= 0 {
		// All-zero weights: treat every signal as equal weight 1.
		w = Weights{ActiveUsers: 1, Deployments: 1, Executions30d: 1, Founder: 1, NewUsersGrowth: 1}
		totalWeight = 5
	}

	raw := (float64(w.ActiveUsers) / totalWeight) * log1p(float64(signals.ActiveUsers)) +
		(float64(w.Deployments) / totalWeight) * log1p(float64(signals.Deployments)) +
		(float64(w.Executions30d) / totalWeight) * log1p(float64(signals.Executions30d)) +
		(float64(w.Founder) / totalWeight) * log1p(float64(signals.FounderEarnings)) +
		(float64(w.NewUsersGrowth) / totalWeight) * log1p(float64(signals.NewUsers30d))

	perCapita := raw
	if population > 0 {
		perCapita = raw * 100000.0 / float64(population)
	}

	return ScoreResult{
		Raw:           raw,
		PerCapita:     perCapita,
		ActiveUsers:   signals.ActiveUsers,
		Deployments:   signals.Deployments,
		Executions30d: int64(signals.Executions30d),
		FounderCents:  signals.FounderEarnings,
		NewUsers30d:   signals.NewUsers30d,
	}
}

func log1p(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Log10(x + 1)
}

// TruncateHour rounds t down to the start of its hour, used as a deterministic
// period_end boundary.
func TruncateHour(t time.Time) time.Time {
	return t.UTC().Truncate(time.Hour)
}

// TierThresholds maps a per-capita score to a tier label. The boundaries are
// the v1 marketing defaults — tune in code if the live distribution shifts.
var TierThresholds = []struct {
	Min   float64
	Label string
}{
	{0.20, "gold"},
	{0.05, "blue"},
	{0.0, "green"},
}

// Tier returns the visual tier for a per-capita score. "gold" beats "blue"
// beats "green".
func Tier(perCapita float64) string {
	for _, t := range TierThresholds {
		if perCapita >= t.Min {
			return t.Label
		}
	}
	return "green"
}

// CategoryWeights returns the Weights profile for a category. Unknown
// categories fall back to the composite weights.
func CategoryWeights(c Category) Weights {
	switch c {
	case CategoryAgents:
		// agents — emphasize executions and growth (proxy for agent adoption).
		return Weights{ActiveUsers: 20, Deployments: 20, Executions30d: 30, Founder: 10, NewUsersGrowth: 20}
	case CategoryAutomation:
		// automation — emphasize deployments and executions (workflow volume).
		return Weights{ActiveUsers: 15, Deployments: 40, Executions30d: 25, Founder: 5, NewUsersGrowth: 15}
	case CategoryStartups:
		// startups — emphasize founder earnings and new users (early-stage velocity).
		return Weights{ActiveUsers: 10, Deployments: 15, Executions30d: 15, Founder: 40, NewUsersGrowth: 20}
	case CategoryOpenSource:
		// open_source — emphasize deployments and executions (registry activity).
		return Weights{ActiveUsers: 20, Deployments: 35, Executions30d: 30, Founder: 5, NewUsersGrowth: 10}
	case CategoryRobotics:
		// robotics — emphasize deployments and executions (hardware/robotics function activity).
		return Weights{ActiveUsers: 15, Deployments: 35, Executions30d: 35, Founder: 5, NewUsersGrowth: 10}
	default:
		return DefaultWeights()
	}
}

// CategoryMeta returns the public-facing metadata for a category, including
// the weights it uses. Used by /v1/city-rankings/categories and the
// marketing-site filter chips.
func CategoryMetaFor(c Category) CategoryMeta {
	switch c {
	case CategoryAgents:
		return CategoryMeta{
			Slug:        string(CategoryAgents),
			Label:       "Agent Capital",
			Description: "Cities where AI agents are the dominant workload. Weights active users, executions, and growth above all else.",
			Weights:     CategoryWeights(CategoryAgents),
		}
	case CategoryAutomation:
		return CategoryMeta{
			Slug:        string(CategoryAutomation),
			Label:       "Automation Capital",
			Description: "Cities where workflow automation is heaviest. Weights deployments and executions.",
			Weights:     CategoryWeights(CategoryAutomation),
		}
	case CategoryStartups:
		return CategoryMeta{
			Slug:        string(CategoryStartups),
			Label:       "Startup Capital",
			Description: "Cities with the most founder velocity. Weights referral earnings and new-user growth.",
			Weights:     CategoryWeights(CategoryStartups),
		}
	case CategoryOpenSource:
		return CategoryMeta{
			Slug:        string(CategoryOpenSource),
			Label:       "Open Source Capital",
			Description: "Cities leading on public registry activity. Weights deployments and executions on public functions.",
			Weights:     CategoryWeights(CategoryOpenSource),
		}
	case CategoryRobotics:
		return CategoryMeta{
			Slug:        string(CategoryRobotics),
			Label:       "Robotics Capital",
			Description: "Cities with the most robotics function deployments and hardware integrations. Weights deployments and executions.",
			Weights:     CategoryWeights(CategoryRobotics),
		}
	default:
		return CategoryMeta{
			Slug:        string(CategoryComposite),
			Label:       "Composite",
			Description: "The default leaderboard: a balanced weighted sum of every signal.",
			Weights:     DefaultWeights(),
		}
	}
}
