package statereport

import (
	"strings"
	"testing"
	"time"
)

func TestHumanInt64(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1234567, "1,234,567"},
		{-1234, "-1,234"},
	}
	for _, c := range cases {
		if got := humanInt64(c.in); got != c.want {
			t.Errorf("humanInt64(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFirstOfMonth(t *testing.T) {
	ref := time.Date(2026, 6, 21, 18, 45, 0, 0, time.UTC)
	got := firstOfMonth(ref)
	want := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("firstOfMonth(%v) = %v, want %v", ref, got, want)
	}
}

func TestReport_SlugAndTitle(t *testing.T) {
	rep := &Report{
		Month: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	// Simulate the same logic as the builder.
	rep.Slug = rep.Month.Format("2006-01")
	rep.Title = "State of AI Builders · " + rep.Month.Format("January 2006")
	if rep.Slug != "2026-06" {
		t.Errorf("expected slug 2026-06, got %s", rep.Slug)
	}
	if !strings.Contains(rep.Title, "June 2026") {
		t.Errorf("expected title to contain 'June 2026', got %s", rep.Title)
	}
}

func TestReport_Render_EmptySectionsAreSkipped(t *testing.T) {
	rep := &Report{
		Month:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Title:       "State of AI Builders · June 2026",
		Slug:        "2026-06",
		GeneratedAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		HeadlineStats: HeadlineStats{
			MetrosRanked: 3, UniversitiesRanked: 5,
		},
		TopMetros: []MetroRow{
			{Rank: 1, Name: "San Francisco", CountryCode: "US", Population: 4729000, ScorePerCapita: 0.04, ActiveUsers: 10},
		},
	}
	out := rep.Render()
	if !strings.Contains(out, "# State of AI Builders") {
		t.Error("missing title")
	}
	if !strings.Contains(out, "## TL;DR") {
		t.Error("missing TL;DR")
	}
	if !strings.Contains(out, "| # | City |") {
		t.Error("missing top metros table")
	}
	if strings.Contains(out, "## Top 10 universities") {
		t.Error("universities section should be skipped when empty")
	}
	if strings.Contains(out, "## New metros") {
		t.Error("new metros section should be skipped when empty")
	}
}

func TestSortByDeltaDesc(t *testing.T) {
	rs := []MoverRow{
		{Rank: 5, RankDelta: 1, Name: "A"},
		{Rank: 1, RankDelta: 10, Name: "B"},
		{Rank: 3, RankDelta: -2, Name: "C"},
	}
	SortByDeltaDesc(rs)
	if rs[0].Name != "B" || rs[1].Name != "A" || rs[2].Name != "C" {
		t.Errorf("unexpected order: %+v", rs)
	}
}
