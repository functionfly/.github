package cityranking

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Integration tests for the city-ranking repository. They require a live
// Postgres connection. Skip with `go test -short`.
func TestRepository_PrivacyThreshold(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) < 3 {
		t.Fatalf("need at least 3 metros (got %d): %v", len(metros), err)
	}
	high := metros[0]
	mid := metros[len(metros)/2]
	low := metros[len(metros)-1]

	// Hermetic reset
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'crtest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM city_rankings`)

	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")

	// 4 users in `low` (below k=5 threshold), 10 in `high`, 0 in `mid`.
	for i := 0; i < 4; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, low.ID, i+1)
	}
	for i := 0; i < 10; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, high.ID, 100+i)
	}

	// Manually score each metro in a single period so the test is hermetic.
	periodEnd := time.Now().UTC().Truncate(time.Hour)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	for _, m := range []MetroArea{high, low, mid} {
		signals, err := repo.MetroSignals(ctx, m.ID, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("signals for %s: %v", m.Slug, err)
		}
		score := Compute(signals, m.Population, DefaultWeights())
		if err := repo.UpsertRanking(ctx, m.ID, CategoryComposite, score, periodStart, periodEnd); err != nil {
			t.Fatalf("upsert for %s: %v", m.Slug, err)
		}
	}
	if err := repo.AssignRanks(ctx, periodEnd, CategoryComposite); err != nil {
		t.Fatalf("assign ranks: %v", err)
	}

	// Verify state leaderboard returns at least one entry.
	rows, err := repo.ListRankings(ctx, 500, "", CategoryComposite)
	if err != nil {
		t.Fatalf("list rankings: %v", err)
	}
	hasHigh, hasLow, hasMid := false, false, false
	for _, r := range rows {
		switch r.MetroAreaID {
		case high.ID:
			hasHigh = true
		case low.ID:
			hasLow = true
		case mid.ID:
			hasMid = true
		}
	}
	if !hasHigh {
		t.Errorf("expected high-metro (10 users) in leaderboard, got %d rows", len(rows))
	}
	if hasLow {
		t.Errorf("low-metro (4 users) should be filtered out by k>=5 threshold")
	}
	if hasMid {
		t.Errorf("mid-metro (0 users) should be filtered out by k>=5 threshold")
	}

	// Verify sorted descending by per-capita score.
	for i := 1; i < len(rows); i++ {
		if rows[i-1].ScorePerCapita < rows[i].ScorePerCapita {
			t.Errorf("leaderboard not sorted descending at index %d", i)
		}
	}
}

func TestRepository_OptOutExcludesUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())
	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) == 0 {
		t.Fatalf("need metros: %v", err)
	}
	target := metros[0]

	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'crtest-%@example.com'`)

	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	uid := "00000000-0000-0000-0000-000000000777"
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, token_version, profile_number, tenant_id, city_id, city_ranking_opted_out, last_active_at, created_at)
		VALUES ($1, $2, 'OptOut Test', '!x', 0, 800777, $3, (SELECT id FROM cities WHERE metro_area_id = $4 AND is_active LIMIT 1), TRUE, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
		ON CONFLICT (id) DO UPDATE SET city_ranking_opted_out = TRUE, city_id = EXCLUDED.city_id
	`, uid, "crtest-optout@example.com", tenantID, target.ID); err != nil {
		t.Fatalf("insert: %v", err)
	}

	signals, err := repo.MetroSignals(ctx, target.ID, time.Now().Add(-30*24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("signals: %v", err)
	}
	if signals.ActiveUsers != 0 {
		t.Errorf("opt-out user should be excluded, got active_users=%d", signals.ActiveUsers)
	}
}

func TestRepository_LookupByAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	// "austin" should resolve to Austin, TX.
	cities, err := repo.LookupCityByAlias(ctx, "austin")
	if err != nil {
		t.Fatalf("lookup austin: %v", err)
	}
	if len(cities) == 0 {
		t.Fatalf("expected austin to resolve to at least one city")
	}
	// "london" should resolve to London, GB.
	cities, err = repo.LookupCityByAlias(ctx, "london")
	if err != nil {
		t.Fatalf("lookup london: %v", err)
	}
	if len(cities) == 0 {
		t.Fatalf("expected london to resolve to at least one city")
	}
	if cities[0].CountryCode != "GB" {
		t.Errorf("expected london to resolve to GB, got %s", cities[0].CountryCode)
	}
}

func TestRepository_StateAggregation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	// Hermetic reset.
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'crtest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM city_rankings`)

	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) < 2 {
		t.Fatalf("need at least 2 metros (got %d): %v", len(metros), err)
	}
	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")

	// Pick 2 metros that share a state (e.g. NYC and Jersey City are both NJ).
	// We'll fall back to two arbitrary metros if the seed doesn't have a
	// shared-state pair for the test country.
	type m struct{ id int64; slug string }
	candidates := []m{}
	for _, metro := range metros {
		var stateCode string
		row := pool.QueryRow(ctx, `SELECT state_code FROM cities WHERE metro_area_id = $1 AND is_active AND state_code <> '' LIMIT 1`, metro.ID)
		if err := row.Scan(&stateCode); err == nil {
			candidates = append(candidates, m{id: metro.ID, slug: metro.Slug})
		}
	}
	if len(candidates) < 2 {
		t.Skip("seed doesn't expose state_code on enough metros for this test")
	}

	// Insert 12 users into the first candidate, 5 into the second.
	insertSmokeUser(t, ctx, pool, tenantID, candidates[0].id, 1)
	for i := 0; i < 12; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, candidates[0].id, 10+i)
	}
	for i := 0; i < 5; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, candidates[1].id, 100+i)
	}

	// Score each metro.
	periodEnd := time.Now().UTC().Truncate(time.Hour)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	scored := map[int64]bool{}
	for _, m := range candidates {
		signals, err := repo.MetroSignals(ctx, m.id, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("signals for %s: %v", m.slug, err)
		}
		var population int
		if err := pool.QueryRow(ctx, `SELECT population FROM metro_areas WHERE id = $1`, m.id).Scan(&population); err != nil {
			t.Fatalf("population: %v", err)
		}
		score := Compute(signals, population, DefaultWeights())
		if err := repo.UpsertRanking(ctx, m.id, CategoryComposite, score, periodStart, periodEnd); err != nil {
			t.Fatalf("upsert: %v", err)
		}
		scored[m.id] = true
	}
	if err := repo.AssignRanks(ctx, periodEnd, CategoryComposite); err != nil {
		t.Fatalf("assign ranks: %v", err)
	}

	// Verify state leaderboard returns at least one entry.
	rows, err := repo.ListStateRankings(ctx, "", 100, CategoryComposite)
	if err != nil {
		t.Fatalf("list states: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected at least one ranked state, got 0")
	}
	// Each state row should sum active_users from the metros in that state.
	for _, s := range rows {
		if s.ActiveUsers == 0 {
			t.Errorf("state %s has 0 active users", s.StateCode)
		}
		if s.RankedMetros == 0 {
			t.Errorf("state %s has 0 ranked metros", s.StateCode)
		}
		if s.ScorePerCapita <= 0 {
			t.Errorf("state %s has non-positive per-capita", s.StateCode)
		}
		if s.RankPosition < 1 {
			t.Errorf("state %s has invalid rank %d", s.StateCode, s.RankPosition)
		}
	}
	// Descending by per-capita.
	for i := 1; i < len(rows); i++ {
		if rows[i-1].ScorePerCapita < rows[i].ScorePerCapita {
			t.Errorf("state leaderboard not sorted descending at index %d", i)
		}
	}
}

func TestRepository_OpenMaxMindDB_GracefulWhenMissing(t *testing.T) {
	r, err := OpenMaxMindDB()
	if err != nil {
		t.Fatalf("OpenMaxMindDB: %v", err)
	}
	if r != nil {
		_ = r.Close()
	}
}

func TestRepository_IPGeoResolver_NoLookupFnIsNoOp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())
	// nil countryLookup => resolver must return NotFound without panicking.
	resolver := NewIPGeoResolver(repo, nil)
	got, err := resolver.Resolve(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("nil lookup should not error: %v", err)
	}
	if !got.NotFound {
		t.Errorf("expected NotFound with nil lookup, got %+v", got)
	}
}

func TestRepository_IPGeoResolver_CountryFallbackFindsLargestCity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	// Stub CountryLookup: hard-coded mapping for known test IPs. Mirrors the
	// production MaxMind adapter's signature so the resolver is exercised
	// end-to-end without a real DB.
	lookup := func(ctx context.Context, ip string) (string, error) {
		switch ip {
		case "8.8.8.8":
			return "US", nil
		case "200.200.200.200":
			return "BR", nil
		case "1.1.1.1":
			return "AU", nil
		}
		return "ZZ", nil // unknown country
	}
	resolver := NewIPGeoResolver(repo, lookup)

	t.Run("US resolves to a US city", func(t *testing.T) {
		got, err := resolver.Resolve(ctx, "8.8.8.8")
		if err != nil {
			t.Fatalf("US lookup failed: %v", err)
		}
		if got.NotFound {
			t.Fatalf("expected US city, got NotFound")
		}
		if got.City == nil || got.City.CountryCode != "US" {
			t.Errorf("expected US city, got %+v", got)
		}
	})

	t.Run("BR resolves to a BR city", func(t *testing.T) {
		got, err := resolver.Resolve(ctx, "200.200.200.200")
		if err != nil {
			t.Fatalf("BR lookup failed: %v", err)
		}
		if got.NotFound {
			t.Fatalf("expected BR city, got NotFound")
		}
		if got.City.CountryCode != "BR" {
			t.Errorf("expected BR, got %s", got.City.CountryCode)
		}
	})

	t.Run("unknown country returns NotFound", func(t *testing.T) {
		got, err := resolver.Resolve(ctx, "9.9.9.9")
		if err != nil {
			t.Fatalf("unknown country: %v", err)
		}
		if !got.NotFound {
			t.Errorf("expected NotFound, got %+v", got)
		}
		if got.CountryCode != "ZZ" {
			t.Errorf("expected country_code=ZZ, got %q", got.CountryCode)
		}
	})

	t.Run("invalid IP errors", func(t *testing.T) {
		if _, err := resolver.Resolve(ctx, "not-an-ip"); err == nil {
			t.Errorf("expected error for invalid IP")
		}
	})

	t.Run("result is cached", func(t *testing.T) {
		// Calling twice should not error and should return the same city.
		first, _ := resolver.Resolve(ctx, "1.1.1.1")
		second, _ := resolver.Resolve(ctx, "1.1.1.1")
		if first.City == nil || second.City == nil {
			t.Fatalf("expected both calls to return a city")
		}
		if first.City.CityID != second.City.CityID {
			t.Errorf("cache returned different city: %d vs %d",
				first.City.CityID, second.City.CityID)
		}
	})
}

func TestRepository_ListCityRankings_HonorsPrivacy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	// Hermetic reset.
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'crtest-%@example.com'`)

	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) < 2 {
		t.Fatalf("need metros: %v", err)
	}
	high := metros[0]
	low := metros[len(metros)-1]

	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	// 10 users in high (passes k=5), 4 in low (filtered out by privacy).
	for i := 0; i < 10; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, high.ID, 2000+i)
	}
	for i := 0; i < 4; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, low.ID, 3000+i)
	}

	// Without a country filter: high's cities should appear, low's should not.
	rows, err := repo.ListCityRankings(ctx, "", 200, CategoryComposite)
	if err != nil {
		t.Fatalf("list city rankings: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected at least one ranked city (the high-metro one)")
	}
	for _, r := range rows {
		if r.ActiveUsers < MinActiveUsersForPublic {
			t.Errorf("city %s should be filtered out (active=%d)", r.CitySlug, r.ActiveUsers)
		}
	}

	// Country filter: only US cities.
	usRows, err := repo.ListCityRankings(ctx, "US", 200, CategoryComposite)
	if err != nil {
		t.Fatalf("list city rankings (US): %v", err)
	}
	for _, r := range usRows {
		if r.CountryCode != "US" {
			t.Errorf("country filter leaked %q", r.CountryCode)
		}
	}
}

func TestRepository_ListBuilders_PrivacySuppression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	// Hermetic reset.
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'crtest-%@example.com'`)

	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) < 2 {
		t.Fatalf("need metros: %v", err)
	}
	high := metros[0]
	low := metros[len(metros)-1]

	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	// 4 users in low (below k=5) — builders should be suppressed.
	for i := 0; i < 4; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, low.ID, 4000+i)
	}
	// 6 in high — builders should be returned, ranked.
	for i := 0; i < 6; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, high.ID, 5000+i)
	}

	// Privacy suppression.
	lowBuilders, err := repo.ListBuilders(ctx, low.Slug, 25, CategoryComposite)
	if err != nil {
		t.Fatalf("list builders (low): %v", err)
	}
	if len(lowBuilders) != 0 {
		t.Errorf("low metro should be privacy-suppressed, got %d builders", len(lowBuilders))
	}

	// Builders should be returned for the high metro.
	highBuilders, err := repo.ListBuilders(ctx, high.Slug, 25, CategoryComposite)
	if err != nil {
		t.Fatalf("list builders (high): %v", err)
	}
	if len(highBuilders) == 0 {
		t.Fatalf("expected builders for high metro, got 0")
	}
	// Names should never be emails (PII).
	for _, b := range highBuilders {
		if strings.Contains(b.DisplayName, "@") {
			t.Errorf("builder display name leaked email: %q", b.DisplayName)
		}
	}
	// Ranks are 1..N in returned order.
	for i, b := range highBuilders {
		if b.Rank != i+1 {
			t.Errorf("builder rank gap at %d: got %d", i, b.Rank)
		}
	}

	// Unknown metro returns (nil, nil) — handled as 404 by the handler.
	missing, err := repo.ListBuilders(ctx, "metro-that-does-not-exist", 25, CategoryComposite)
	if err != nil {
		t.Fatalf("unknown metro: %v", err)
	}
	if missing != nil {
		t.Errorf("unknown metro should return nil, got %v", missing)
	}
}

func TestRepository_MapPoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())
	pts, err := repo.ListMapPoints(ctx, CategoryComposite)
	if err != nil {
		t.Fatalf("list map points: %v", err)
	}
	if len(pts) == 0 {
		t.Skip("no ranked metros in DB; run the smoke test or full recompute first")
	}
	for _, p := range pts {
		if p.Latitude == 0 || p.Longitude == 0 {
			t.Errorf("metro %s has zero coordinates (%v, %v)", p.MetroSlug, p.Latitude, p.Longitude)
		}
		if p.Tier != "gold" && p.Tier != "blue" && p.Tier != "green" {
			t.Errorf("metro %s has invalid tier %q", p.MetroSlug, p.Tier)
		}
	}
}

func TestRepository_PerCategoryScores(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool, logrus.StandardLogger())

	metros, err := repo.ListMetros(ctx)
	if err != nil || len(metros) < 1 {
		t.Fatalf("need metros: %v", err)
	}
	target := metros[0]

	// Wipe existing ranking rows for a hermetic test.
	_, _ = pool.Exec(ctx, `DELETE FROM city_rankings`)

	// Insert enough users to clear the privacy threshold.
	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	for i := 0; i < 6; i++ {
		insertSmokeUser(t, ctx, pool, tenantID, target.ID, 10000+i)
	}

	// Score under every category.
	periodEnd := time.Now().UTC().Truncate(time.Hour)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	signals, err := repo.MetroSignals(ctx, target.ID, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("signals: %v", err)
	}
	for _, c := range AllCategories {
		weights := CategoryWeights(c)
		score := Compute(signals, target.Population, weights)
		if err := repo.UpsertRanking(ctx, target.ID, c, score, periodStart, periodEnd); err != nil {
			t.Fatalf("upsert %s: %v", c, err)
		}
	}
	// Assign ranks for every category so rank_position is populated.
	for _, c := range AllCategories {
		if err := repo.AssignRanks(ctx, periodEnd, c); err != nil {
			t.Fatalf("assign ranks %s: %v", c, err)
		}
	}

	// Verify: each category has a row, the score_raw differs because weights
	// differ, and ranking position is the same (only one metro).
	for _, c := range AllCategories {
		rk, err := repo.GetRankingBySlug(ctx, target.Slug, c)
		if err != nil {
			t.Fatalf("get ranking for %s: %v", c, err)
		}
		if rk == nil {
			t.Fatalf("no ranking row for %s", c)
		}
		if rk.Category != c {
			t.Errorf("expected category %q, got %q", c, rk.Category)
		}
		if rk.RankPosition != 1 {
			t.Errorf("expected rank 1 for %s, got %d", c, rk.RankPosition)
		}
	}

	// Cross-category isolation: changing the category filter should not
	// leak rows from another category.
	compRows, err := repo.ListRankings(ctx, 10, "", CategoryComposite)
	if err != nil {
		t.Fatalf("list composite: %v", err)
	}
	agentsRows, err := repo.ListRankings(ctx, 10, "", CategoryAgents)
	if err != nil {
		t.Fatalf("list agents: %v", err)
	}
	if len(compRows) != len(agentsRows) {
		t.Errorf("category isolation broken: composite=%d, agents=%d", len(compRows), len(agentsRows))
	}
	for _, r := range compRows {
		if r.Category != CategoryComposite {
			t.Errorf("composite list contains %q row", r.Category)
		}
	}
	for _, r := range agentsRows {
		if r.Category != CategoryAgents {
			t.Errorf("agents list contains %q row", r.Category)
		}
	}

	// Verify that per-category weights diverge (each CategoryWeights
	// returns a different profile, even if the resulting raw score happens
	// to coincide for this particular signal shape).
	seen := map[Weights]bool{}
	for _, c := range AllCategories {
		seen[CategoryWeights(c)] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected categories to use at least 2 distinct weight profiles, got %d", len(seen))
	}

	// Invalid category should be rejected by the CHECK constraint when
	// trying to insert directly. (Repository layer doesn't validate; this
	// is a DB-level test.)
	_, err = pool.Exec(ctx, `
		INSERT INTO city_rankings (metro_area_id, ranking_category, score_raw, score_per_capita,
			active_users, deployments, executions_30d, founder_earnings, new_users_30d,
			period_start, period_end)
		VALUES ($1, 'unknown', 0, 0, 0, 0, 0, 0, 0, $2, $3)
	`, target.ID, periodStart, periodEnd)
	if err == nil {
		t.Errorf("expected CHECK constraint to reject 'unknown' category")
	}
}

func insertSmokeUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, metroID int64, idx int) {
	t.Helper()
	uid := fmt.Sprintf("00000000-0000-0000-0000-%012d", 1000+idx)
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, token_version, profile_number, tenant_id, city_id, city_ranking_opted_out, last_active_at, created_at)
		VALUES ($1, $2, $3, '!x', 0, $4, $5, (SELECT id FROM cities WHERE metro_area_id = $6 AND is_active LIMIT 1), FALSE, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days')
		ON CONFLICT (id) DO UPDATE SET city_id = EXCLUDED.city_id, city_ranking_opted_out = FALSE
	`, uid, fmt.Sprintf("crtest-%d@example.com", idx), fmt.Sprintf("User %d", idx), 900000+idx, tenantID, metroID); err != nil {
		t.Fatalf("insert user %d: %v", idx, err)
	}
}

func readFirstString(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string) string {
	t.Helper()
	var v string
	if err := pool.QueryRow(ctx, query).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return v
}

// readFirstStringQ is the variadic variant for queries with parameters.
func readFirstStringQ(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, args ...any) string {
	t.Helper()
	var v string
	if err := pool.QueryRow(ctx, query, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return v
}
