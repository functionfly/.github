// Package statereport generates the monthly "State of AI Builders" report.
// The data is sourced from the city and university ranking tables plus the
// city_ambassadors table; no separate aggregation layer is needed because
// the hourly recompute job already maintains the rows the report consumes.
//
// The report is a single Markdown file (one per month) with a stable
// filename `YYYY-MM.md`. The marketing site renders the file under
// `/blog/[month]` so the same artifact powers both the static site and
// any email export.
//
// Run on the 1st of each month at 09:00 UTC by the cron in
// internal/jobs/statereport. The script at scripts/generate_report
// exposes the same code path for ad-hoc / back-fill runs.
package statereport

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/functionfly/functionfly/internal/storage/universityranking"
)

// Report is the in-memory representation of one month's state.
type Report struct {
	Month            time.Time         // first day of the month (UTC)
	Title            string            // "State of AI Builders · 2026 · June"
	Slug             string            // "2026-06" (used in URL paths)
	GeneratedAt      time.Time         // when this Report was built
	HeadlineStats    HeadlineStats     // top-of-page numbers
	TopMetros        []MetroRow        // top 10 metros (composite)
	TopUniversities  []UniversityRow   // top 10 universities (composite)
	NewAmbassadors   []AmbassadorRow   // ambassadors promoted during the month
	BiggestGainers   []MoverRow        // top 5 metros that climbed the most
	BiggestLosers    []MoverRow        // top 5 metros that fell the most
	NewMetros        []MetroRow        // metros that crossed the privacy threshold this month
	PeriodStart      time.Time
	PeriodEnd        time.Time
}

// HeadlineStats is the top-of-page summary.
type HeadlineStats struct {
	MetrosRanked         int
	UniversitiesRanked   int
	TotalActiveUsers     int
	TotalNewUsers        int
	TotalDeployments     int
	TotalExecutions      int64
	NewAmbassadorsCount  int
	NewMetrosCount       int
}

// MetroRow is one row in the top-metros / new-metros / movers tables.
type MetroRow struct {
	Rank          int
	Slug          string
	Name          string
	CountryCode   string
	StateCode     string
	Population    int
	ScorePerCapita float64
	ActiveUsers   int
	NewUsers      int
	Deployments   int
	Executions    int64
}

// UniversityRow is one row in the top-universities table.
type UniversityRow struct {
	Rank          int
	Slug          string
	Name          string
	CountryCode   string
	ScorePerCapita float64
	ActiveUsers   int
	StudentCount  int
}

// AmbassadorRow is one row in the new-ambassadors table.
type AmbassadorRow struct {
	MetroSlug   string
	MetroName   string
	CountryCode string
	UserName    string
	PromotedAt  time.Time
	Source      string
}

// MoverRow is one row in the gainers/losers table.
type MoverRow struct {
	Rank         int
	PrevRank     *int
	RankDelta    int
	Slug         string
	Name         string
	CountryCode  string
	ScorePerCapita float64
}

// Builder wires a Report from the existing city + university repos.
type Builder struct {
	city  *cityranking.Repository
	univ  *universityranking.Repository
	now   func() time.Time
}

// New returns a Builder. now is injectable for tests; pass nil for time.Now.
func New(city *cityranking.Repository, univ *universityranking.Repository, now func() time.Time) *Builder {
	if now == nil {
		now = time.Now
	}
	return &Builder{city: city, univ: univ, now: now}
}

// BuildForMonth returns the Report for the calendar month that ended at
// `reference`. The period_start is the first of the previous month, the
// period_end is the first of the current month (exclusive). If reference
// is June 15, 2026, the report is for June 2026.
func (b *Builder) BuildForMonth(ctx context.Context, reference time.Time) (*Report, error) {
	monthStart := firstOfMonth(reference).UTC()
	monthEnd := monthStart.AddDate(0, 1, 0) // exclusive
	prevPeriodStart := monthStart.AddDate(0, -1, 0)

	rep := &Report{
		Month:       monthStart,
		Title:       fmt.Sprintf("State of AI Builders · %s", monthStart.Format("January 2006")),
		Slug:        monthStart.Format("2006-01"),
		GeneratedAt: b.now().UTC(),
		PeriodStart: prevPeriodStart,
		PeriodEnd:   monthEnd,
	}

	if err := b.populateHeadline(ctx, rep); err != nil {
		return nil, fmt.Errorf("headline: %w", err)
	}
	if err := b.populateMetros(ctx, rep); err != nil {
		return nil, fmt.Errorf("metros: %w", err)
	}
	if err := b.populateUniversities(ctx, rep); err != nil {
		return nil, fmt.Errorf("universities: %w", err)
	}
	if err := b.populateAmbassadors(ctx, rep); err != nil {
		return nil, fmt.Errorf("ambassadors: %w", err)
	}
	if err := b.populateMovers(ctx, rep); err != nil {
		return nil, fmt.Errorf("movers: %w", err)
	}
	return rep, nil
}

func (b *Builder) populateHeadline(ctx context.Context, rep *Report) error {
	// Use the latest period for the headline — the report is about the
	// final state of the month, not the average.
	period, err := b.city.LatestPeriod(ctx)
	if err != nil {
		return err
	}
	rep.HeadlineStats = HeadlineStats{}

	rows, err := b.city.Pool().Query(ctx, `
		SELECT active_users, new_users_30d, deployments, executions_30d
		FROM city_rankings
		WHERE period_end = $1 AND ranking_category = 'composite' AND active_users >= $2
	`, period, cityranking.MinActiveUsersForPublic)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var active, newUsers, deploys int
		var execs int64
		if err := rows.Scan(&active, &newUsers, &deploys, &execs); err != nil {
			return err
		}
		rep.HeadlineStats.MetrosRanked++
		rep.HeadlineStats.TotalActiveUsers += active
		rep.HeadlineStats.TotalNewUsers += newUsers
		rep.HeadlineStats.TotalDeployments += deploys
		rep.HeadlineStats.TotalExecutions += execs
	}

	// Universities.
	uniRows, err := b.univ.Pool().Query(ctx, `
		SELECT active_users FROM university_rankings
		WHERE period_end = (SELECT MAX(period_end) FROM university_rankings WHERE ranking_category = 'composite')
			AND ranking_category = 'composite' AND active_users >= $1
	`, universityranking.MinActiveUsersForPublic)
	if err == nil {
		defer uniRows.Close()
		for uniRows.Next() {
			var a int
			if err := uniRows.Scan(&a); err == nil {
				rep.HeadlineStats.UniversitiesRanked++
			}
		}
	}

	rep.HeadlineStats.NewMetrosCount = len(rep.NewMetros)
	return nil
}

func (b *Builder) populateMetros(ctx context.Context, rep *Report) error {
	period, err := b.city.LatestPeriod(ctx)
	if err != nil {
		return err
	}
	rows, err := b.city.Pool().Query(ctx, `
		SELECT cr.rank_position, m.slug, m.name, m.country_code, m.population,
			cr.score_per_capita, cr.active_users, cr.new_users_30d,
			cr.deployments, cr.executions_30d
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.period_end = $1 AND cr.ranking_category = 'composite'
			AND cr.active_users >= $2
		ORDER BY cr.rank_position ASC
		LIMIT 10
	`, period, cityranking.MinActiveUsersForPublic)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r MetroRow
		if err := rows.Scan(&r.Rank, &r.Slug, &r.Name, &r.CountryCode, &r.Population,
			&r.ScorePerCapita, &r.ActiveUsers, &r.NewUsers, &r.Deployments, &r.Executions); err != nil {
			return err
		}
		rep.TopMetros = append(rep.TopMetros, r)
	}

	// New metros: any ranking that didn't exist in the previous month.
	prevPeriod, err := b.city.Pool().Query(ctx, `
		SELECT MAX(period_end) FROM city_rankings WHERE period_end < $1
	`, period)
	if err != nil {
		return nil // non-fatal
	}
	defer prevPeriod.Close()
	var prevEnd *time.Time
	if prevPeriod.Next() {
		var t time.Time
		if err := prevPeriod.Scan(&t); err == nil {
			prevEnd = &t
		}
	} else {
		// No previous period — every ranked metro is "new".
		_ = prevEnd
	}

	rows2, err := b.city.Pool().Query(ctx, `
		SELECT cr.rank_position, m.slug, m.name, m.country_code, m.population,
			cr.score_per_capita, cr.active_users, cr.new_users_30d,
			cr.deployments, cr.executions_30d
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.period_end = $1 AND cr.ranking_category = 'composite'
			AND cr.active_users >= $2
		ORDER BY cr.rank_position ASC
	`, period, cityranking.MinActiveUsersForPublic)
	if err != nil {
		return nil
	}
	defer rows2.Close()
	for rows2.Next() {
		var r MetroRow
		if err := rows2.Scan(&r.Rank, &r.Slug, &r.Name, &r.CountryCode, &r.Population,
			&r.ScorePerCapita, &r.ActiveUsers, &r.NewUsers, &r.Deployments, &r.Executions); err != nil {
			continue
		}
		// "New" = present this month but absent (or below threshold) last month.
		if prevEnd == nil {
			rep.NewMetros = append(rep.NewMetros, r)
			continue
		}
		var existed bool
		_ = b.city.Pool().QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM city_rankings
				WHERE period_end = $1 AND metro_area_id = (
					SELECT id FROM metro_areas WHERE slug = $2
				) AND active_users >= $3
			)
		`, *prevEnd, r.Slug, cityranking.MinActiveUsersForPublic).Scan(&existed)
		if !existed {
			rep.NewMetros = append(rep.NewMetros, r)
		}
	}
	return nil
}

func (b *Builder) populateUniversities(ctx context.Context, rep *Report) error {
	rows, err := b.univ.Pool().Query(ctx, `
		SELECT r.rank_position, univ.slug, univ.name, univ.country_code,
			r.score_per_capita, r.active_users, univ.student_count
		FROM university_rankings r
		JOIN universities univ ON univ.id = r.university_id
		WHERE r.period_end = (SELECT MAX(period_end) FROM university_rankings WHERE ranking_category = 'composite')
			AND r.ranking_category = 'composite' AND r.active_users >= $1
		ORDER BY r.rank_position ASC
		LIMIT 10
	`, universityranking.MinActiveUsersForPublic)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r UniversityRow
		if err := rows.Scan(&r.Rank, &r.Slug, &r.Name, &r.CountryCode,
			&r.ScorePerCapita, &r.ActiveUsers, &r.StudentCount); err != nil {
			return err
		}
		rep.TopUniversities = append(rep.TopUniversities, r)
	}
	return nil
}

func (b *Builder) populateAmbassadors(ctx context.Context, rep *Report) error {
	rows, err := b.city.Pool().Query(ctx, `
		SELECT m.slug, m.name, m.country_code, u.name, a.promoted_at, a.source
		FROM city_ambassadors a
		JOIN metro_areas m ON m.id = a.metro_id
		JOIN users u ON u.id = a.user_id
		WHERE a.promoted_at >= $1 AND a.promoted_at < $2
		ORDER BY a.promoted_at DESC
	`, rep.PeriodStart, rep.PeriodEnd)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r AmbassadorRow
		if err := rows.Scan(&r.MetroSlug, &r.MetroName, &r.CountryCode, &r.UserName, &r.PromotedAt, &r.Source); err != nil {
			return err
		}
		rep.NewAmbassadors = append(rep.NewAmbassadors, r)
	}
	rep.HeadlineStats.NewAmbassadorsCount = len(rep.NewAmbassadors)
	return nil
}

func (b *Builder) populateMovers(ctx context.Context, rep *Report) error {
	// Top 5 gainers.
	gainers, err := b.city.ListMovers(ctx, "gainers", 5, cityranking.CategoryComposite)
	if err != nil {
		return err
	}
	for _, r := range gainers {
		rep.BiggestGainers = append(rep.BiggestGainers, toMoverRow(r))
	}
	losers, err := b.city.ListMovers(ctx, "losers", 5, cityranking.CategoryComposite)
	if err != nil {
		return err
	}
	for _, r := range losers {
		rep.BiggestLosers = append(rep.BiggestLosers, toMoverRow(r))
	}
	return nil
}

func toMoverRow(r cityranking.Ranking) MoverRow {
	out := MoverRow{
		Rank:          r.RankPosition,
		Slug:          r.MetroSlug,
		Name:          r.MetroName,
		CountryCode:   r.CountryCode,
		ScorePerCapita: r.ScorePerCapita,
	}
	if r.PrevRankPosition.Valid {
		prev := int(r.PrevRankPosition.Int64)
		out.PrevRank = &prev
		out.RankDelta = prev - r.RankPosition
	}
	return out
}

// firstOfMonth returns the first instant of the month containing t, in UTC.
func firstOfMonth(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// ── Markdown rendering ───────────────────────────────────────────────────

// Render returns the Markdown source for the report. The output is
// self-contained: no external CSS or JS. The marketing site renders it
// as-is.
func (r *Report) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", r.Title)
	fmt.Fprintf(&b, "_Period: %s → %s · Generated %s_\n\n",
		r.PeriodStart.Format("2006-01-02"), r.PeriodEnd.Format("2006-01-02"),
		r.GeneratedAt.Format("2006-01-02 15:04 MST"))
	fmt.Fprintf(&b, "**Slug:** `%s`\n\n", r.Slug)

	fmt.Fprintf(&b, "## TL;DR\n\n")
	fmt.Fprintf(&b, "- **%d metros** ranked, **%d universities** ranked.\n", r.HeadlineStats.MetrosRanked, r.HeadlineStats.UniversitiesRanked)
	fmt.Fprintf(&b, "- **%s** total active builders in the last 30 days.\n", humanInt(r.HeadlineStats.TotalActiveUsers))
	fmt.Fprintf(&b, "- **%s** new builders joined in the last 30 days.\n", humanInt(r.HeadlineStats.TotalNewUsers))
	fmt.Fprintf(&b, "- **%s** deployments and **%s** function executions.\n",
		humanInt(r.HeadlineStats.TotalDeployments), humanInt64(r.HeadlineStats.TotalExecutions))
	if r.HeadlineStats.NewMetrosCount > 0 {
		fmt.Fprintf(&b, "- **%d new metros** crossed the privacy threshold this month.\n", r.HeadlineStats.NewMetrosCount)
	}
	if r.HeadlineStats.NewAmbassadorsCount > 0 {
		fmt.Fprintf(&b, "- **%d new ambassadors** promoted.\n", r.HeadlineStats.NewAmbassadorsCount)
	}
	fmt.Fprintln(&b)

	if len(r.TopMetros) > 0 {
		fmt.Fprintln(&b, "## Top 10 metros")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| # | City | Country | Pop. | Per-capita | Builders |")
		fmt.Fprintln(&b, "|---|------|---------|-----:|-----------:|---------:|")
		for _, m := range r.TopMetros {
			fmt.Fprintf(&b, "| %d | %s | %s | %s | %.4f | %d |\n",
				m.Rank, m.Name, m.CountryCode, humanInt(m.Population), m.ScorePerCapita, m.ActiveUsers)
		}
		fmt.Fprintln(&b)
	}

	if len(r.TopUniversities) > 0 {
		fmt.Fprintln(&b, "## Top 10 universities")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| # | University | Country | Students | Per-capita | Builders |")
		fmt.Fprintln(&b, "|---|------------|---------|---------:|-----------:|---------:|")
		for _, u := range r.TopUniversities {
			fmt.Fprintf(&b, "| %d | %s | %s | %s | %.4f | %d |\n",
				u.Rank, u.Name, u.CountryCode, humanInt(u.StudentCount), u.ScorePerCapita, u.ActiveUsers)
		}
		fmt.Fprintln(&b)
	}

	if len(r.NewMetros) > 0 {
		fmt.Fprintln(&b, "## New metros this month")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "These metros crossed the k=5 privacy threshold for the first time:")
		fmt.Fprintln(&b)
		for _, m := range r.NewMetros {
			fmt.Fprintf(&b, "- **%s**, %s — %d builders\n", m.Name, m.CountryCode, m.ActiveUsers)
		}
		fmt.Fprintln(&b)
	}

	if len(r.NewAmbassadors) > 0 {
		fmt.Fprintln(&b, "## New ambassadors")
		fmt.Fprintln(&b)
		for _, a := range r.NewAmbassadors {
			fmt.Fprintf(&b, "- **%s** — %s, %s (%s)\n",
				a.UserName, a.MetroName, a.CountryCode, a.PromotedAt.Format("2006-01-02"))
		}
		fmt.Fprintln(&b)
	}

	if len(r.BiggestGainers) > 0 {
		fmt.Fprintln(&b, "## Biggest gainers")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Δ | City | New rank | Per-capita |")
		fmt.Fprintln(&b, "|---|------|---------:|-----------:|")
		for _, m := range r.BiggestGainers {
			prev := "—"
			if m.PrevRank != nil {
				prev = fmt.Sprintf("%d", *m.PrevRank)
			}
			fmt.Fprintf(&b, "| +%d | %s, %s (#%s) | %d | %.4f |\n",
				m.RankDelta, m.Name, m.CountryCode, prev, m.Rank, m.ScorePerCapita)
		}
		fmt.Fprintln(&b)
	}

	if len(r.BiggestLosers) > 0 {
		fmt.Fprintln(&b, "## Biggest losers")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Δ | City | New rank | Per-capita |")
		fmt.Fprintln(&b, "|---|------|---------:|-----------:|")
		for _, m := range r.BiggestLosers {
			prev := "—"
			if m.PrevRank != nil {
				prev = fmt.Sprintf("%d", *m.PrevRank)
			}
			fmt.Fprintf(&b, "| %d | %s, %s (#%s) | %d | %.4f |\n",
				m.RankDelta, m.Name, m.CountryCode, prev, m.Rank, m.ScorePerCapita)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "---")
	fmt.Fprintf(&b, "\n_Generated by the [FunctionFly](https://functionfly.com) state-report job. Source: live city + university + ambassador tables. Privacy threshold: k≥%d active builders._\n",
		cityranking.MinActiveUsersForPublic)
	return b.String()
}

func humanInt(n int) string {
	return humanInt64(int64(n))
}

func humanInt64(n int64) string {
	// Simple thousands-separator (no intl for now).
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	// Insert commas every 3 digits from the right.
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// SortByDeltaDesc sorts movers by rank delta descending.
func SortByDeltaDesc(rows []MoverRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].RankDelta > rows[j].RankDelta
	})
}
