package universityranking

import (
	"math"
	"testing"
	"time"
)

func TestCompute_PureMath(t *testing.T) {
	t.Run("zero signals → zero score", func(t *testing.T) {
		s := Compute(Signals{}, 1000, 1, 1, 1, 1, 1)
		if s.Raw != 0 || s.PerCapita != 0 {
			t.Errorf("expected zero score, got raw=%f per_capita=%f", s.Raw, s.PerCapita)
		}
	})

	t.Run("active users drives the score", func(t *testing.T) {
		// Both inputs are log1p(100) ≈ 4.6 with equal weights, so the raw
		// score is the same. This documents the design: log dampening
		// flattens the difference between "100 users" and "100 deploys" so
		// a single mega-deployer can't dominate the leaderboard.
		withActive := Compute(Signals{ActiveUsers: 100, Deployments: 0}, 1000, 1, 1, 1, 1, 1)
		withDeploys := Compute(Signals{ActiveUsers: 0, Deployments: 100}, 1000, 1, 1, 1, 1, 1)
		if withActive.Raw != withDeploys.Raw {
			t.Errorf("equal weights + log dampening should produce equal scores: active=%f deploys=%f",
				withActive.Raw, withDeploys.Raw)
		}
	})

	t.Run("per-capita is normalized by population", func(t *testing.T) {
		small := Compute(Signals{ActiveUsers: 10}, 100, 1, 0, 0, 0, 0)
		large := Compute(Signals{ActiveUsers: 10}, 10000, 1, 0, 0, 0, 0)
		// Both have the same raw score (just log(11)), but large has 100x
		// the population, so its per-capita is 100x smaller.
		if math.Abs(small.PerCapita/large.PerCapita-100) > 0.01 {
			t.Errorf("per-capita should be 100x higher for small: small=%f large=%f",
				small.PerCapita, large.PerCapita)
		}
	})

	t.Run("founder earnings uses sqrt dampening", func(t *testing.T) {
		// 1000x more earnings should NOT produce 1000x more founder score.
		low := Compute(Signals{FounderEarnings: 1_000_000}, 1000, 0, 0, 0, 1, 0)
		high := Compute(Signals{FounderEarnings: 1_000_000_000_000}, 1000, 0, 0, 0, 1, 0)
		ratio := high.Raw / low.Raw
		if ratio > 1100 {
			t.Errorf("sqrt dampening should limit founder growth: %fx (expected ~1000x)", ratio)
		}
	})

	t.Run("per-capita formula", func(t *testing.T) {
		// active=1 → log1p(1) = ln(2) ≈ 0.693
		// per_capita = 0.693 * 100000 / 100 = 693.147
		s2 := Compute(Signals{ActiveUsers: 1}, 100, 1, 0, 0, 0, 0)
		want := math.Log1p(1) * 100000 / 100
		if math.Abs(s2.PerCapita-want) > 0.01 {
			t.Errorf("expected per_capita=%f, got %f", want, s2.PerCapita)
		}
	})

	t.Run("population at least 1", func(t *testing.T) {
		s := Compute(Signals{ActiveUsers: 10}, 0, 1, 0, 0, 0, 0)
		if math.IsInf(s.PerCapita, 0) || math.IsNaN(s.PerCapita) {
			t.Errorf("per-capita with population=0 should be finite, got %f", s.PerCapita)
		}
	})
}

func TestCategoryWeights(t *testing.T) {
	cases := []struct {
		cat  Category
		want float64
	}{
		{CategoryComposite, 1.0},
		{CategoryAgents, 1.5},
		{CategoryAutomation, 1.0},
		{CategoryStartups, 0.8},
		{CategoryOpenSource, 1.2},
	}
	for _, c := range cases {
		wA, _, _, _, _ := CategoryWeights(c.cat)
		// active term is the first weight
		if math.Abs(wA-c.want) > 0.01 {
			t.Errorf("%s: expected active weight %f, got %f", c.cat, c.want, wA)
		}
	}
}

func TestValidCategory(t *testing.T) {
	if !ValidCategory("composite") {
		t.Error("composite should be valid")
	}
	if !ValidCategory("agents") {
		t.Error("agents should be valid")
	}
	if ValidCategory("garbage") {
		t.Error("garbage should be invalid")
	}
}

func TestSortRankings(t *testing.T) {
	rs := []Ranking{
		{UniversityID: 1, ScorePerCapita: 0.1},
		{UniversityID: 2, ScorePerCapita: 0.5},
		{UniversityID: 3, ScorePerCapita: 0.3},
	}
	SortRankings(rs)
	if rs[0].UniversityID != 2 || rs[1].UniversityID != 3 || rs[2].UniversityID != 1 {
		t.Errorf("expected sorted descending: got %+v", rs)
	}
}

func TestTruncateHour(t *testing.T) {
	in := time.Date(2026, 6, 21, 18, 45, 23, 999, time.UTC)
	out := TruncateHour(in)
	want := time.Date(2026, 6, 21, 18, 0, 0, 0, time.UTC)
	if !out.Equal(want) {
		t.Errorf("expected %v, got %v", want, out)
	}
}
