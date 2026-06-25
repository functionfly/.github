package cityranking

import (
	"math"
	"testing"
	"time"
)

func TestCompute_BasicRawAndPerCapita(t *testing.T) {
	w := DefaultWeights()
	signals := Signals{
		ActiveUsers:     100,
		Deployments:     50,
		Executions30d:   10000,
		FounderEarnings: 50000,
		NewUsers30d:     20,
	}
	// Austin-ish population
	got := Compute(signals, 1000000, w)
	if got.Raw <= 0 {
		t.Fatalf("expected raw > 0, got %v", got.Raw)
	}
	if got.PerCapita <= 0 {
		t.Fatalf("expected per-capita > 0, got %v", got.PerCapita)
	}
	// Per-capita should be raw * 100000 / population
	want := got.Raw * 100000.0 / 1000000.0
	if math.Abs(got.PerCapita-want) > 1e-6 {
		t.Fatalf("per-capita mismatch: got %v want %v", got.PerCapita, want)
	}
}

func TestCompute_ZeroPopulationFallsBackToRaw(t *testing.T) {
	signals := Signals{ActiveUsers: 10}
	got := Compute(signals, 0, DefaultWeights())
	if got.PerCapita != got.Raw {
		t.Fatalf("expected per-capita to equal raw when pop=0, got raw=%v pc=%v", got.Raw, got.PerCapita)
	}
}

func TestCompute_LogScalingFlattensLargeCounts(t *testing.T) {
	w := DefaultWeights()
	small := Compute(Signals{ActiveUsers: 10}, 1000000, w)
	large := Compute(Signals{ActiveUsers: 1000000}, 1000000, w)
	// log10(10) = 1, log10(1e6+1) ≈ 6 — much less than 100,000x raw
	if large.Raw/small.Raw > 7 {
		t.Fatalf("log scaling not applied: small=%v large=%v ratio=%v", small.Raw, large.Raw, large.Raw/small.Raw)
	}
}

func TestCompute_EmptySignalsAreZero(t *testing.T) {
	got := Compute(Signals{}, 1000000, DefaultWeights())
	if got.Raw != 0 || got.PerCapita != 0 {
		t.Fatalf("expected zero, got raw=%v pc=%v", got.Raw, got.PerCapita)
	}
}

func TestCompute_ZeroWeightsFallbackToEqual(t *testing.T) {
	got := Compute(Signals{ActiveUsers: 100}, 1000000, Weights{})
	if got.Raw <= 0 {
		t.Fatalf("expected raw > 0 even with zero weights, got %v", got.Raw)
	}
}

func TestCompute_RelativeWeightsAffectResult(t *testing.T) {
	// Weights are normalized (relative). Increasing ALL weights by the same
	// factor shouldn't change the score.
	all1 := Compute(Signals{ActiveUsers: 100, Deployments: 10}, 1000000, Weights{ActiveUsers: 30, Deployments: 25})
	all100 := Compute(Signals{ActiveUsers: 100, Deployments: 10}, 1000000, Weights{ActiveUsers: 3000, Deployments: 2500})
	if math.Abs(all1.Raw-all100.Raw) > 1e-6 {
		t.Fatalf("expected equal results for scaled weights, got %v vs %v", all1.Raw, all100.Raw)
	}

	// Shifting weight from active_users to deployments should change the result.
	biased := Compute(Signals{ActiveUsers: 1000, Deployments: 10}, 1000000, Weights{ActiveUsers: 1, Deployments: 99})
	unbiased := Compute(Signals{ActiveUsers: 1000, Deployments: 10}, 1000000, Weights{ActiveUsers: 99, Deployments: 1})
	if math.Abs(biased.Raw-unbiased.Raw) < 1e-6 {
		t.Fatalf("expected different results for different weight splits, got %v vs %v", biased.Raw, unbiased.Raw)
	}
}

func TestNormalizeInput(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Austin, TX", "austin tx"},
		{"  austin  TX  ", "austin tx"},
		{"", ""},
		{"São Paulo", "sao paulo"},
		{"Zürich", "zurich"},
		{"New York, NY", "new york ny"},
		{"", ""},
	}
	for _, c := range cases {
		got := NormalizeInput(c.in)
		if got != c.want {
			t.Errorf("NormalizeInput(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExpandStateAbbreviations(t *testing.T) {
	c, s := ExpandStateAbbreviations("austin tx")
	if c != "austin" || s != "Texas" {
		t.Errorf("got city=%q state=%q, want austin/Texas", c, s)
	}
	c, s = ExpandStateAbbreviations("berlin")
	if c != "berlin" || s != "" {
		t.Errorf("got city=%q state=%q, want berlin/(empty)", c, s)
	}
}

func TestTruncateHour(t *testing.T) {
	now := time.Date(2026, 6, 21, 14, 37, 42, 123, time.UTC)
	got := TruncateHour(now)
	want := time.Date(2026, 6, 21, 14, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("TruncateHour(%v) = %v, want %v", now, got, want)
	}
}

func TestTier(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.5, "gold"},
		{0.2, "gold"},
		{0.19999, "blue"},
		{0.05, "blue"},
		{0.04999, "green"},
		{0, "green"},
		{-1, "green"},
	}
	for _, c := range cases {
		if got := Tier(c.score); got != c.want {
			t.Errorf("Tier(%v) = %q, want %q", c.score, got, c.want)
		}
	}
}
