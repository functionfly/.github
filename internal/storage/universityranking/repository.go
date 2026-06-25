package universityranking

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ── Seed ──────────────────────────────────────────────────────────────────

// SeedFromCSV loads universities + aliases from a CSV with header:
//   slug,name,short_name,country_code,state_code,student_count,institution_type,website,city_slug
// city_slug is optional and resolved to cities.id when present.
func (r *Repository) SeedFromCSV(ctx context.Context, path string) (SeedResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return SeedResult{}, fmt.Errorf("open seed csv: %w", err)
	}
	defer func() { _ = f.Close() }()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return SeedResult{}, fmt.Errorf("read header: %w", err)
	}
	col := indexHeader(header)

	cityIDs, err := r.fetchCitySlugs(ctx)
	if err != nil {
		return SeedResult{}, err
	}

	inserted := 0
	aliasesInserted := 0
	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return SeedResult{}, fmt.Errorf("read row: %w", err)
		}
		if len(row) < len(header) {
			continue
		}
		slug := strings.TrimSpace(row[col["slug"]])
		if slug == "" {
			continue
		}
		students, _ := strconv.Atoi(strings.TrimSpace(row[col["student_count"]]))
		shortName := strings.TrimSpace(row[col["short_name"]])
		stateCode := strings.TrimSpace(row[col["state_code"]])
		citySlug := strings.TrimSpace(row[col["city_slug"]])
		instType := strings.TrimSpace(row[col["institution_type"]])
		if instType == "" {
			instType = "university"
		}
		website := strings.TrimSpace(row[col["website"]])
		var cityID *int64
		if citySlug != "" {
			if id, ok := cityIDs[citySlug]; ok {
				idCopy := id
				cityID = &idCopy
			}
		}

		var id int64
		err = r.pool.QueryRow(ctx, `
			INSERT INTO universities (slug, name, short_name, country_code, state_code,
				city_id, student_count, institution_type, website, is_active)
			VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''), $6, $7, $8, NULLIF($9, ''), TRUE)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				short_name = EXCLUDED.short_name,
				country_code = EXCLUDED.country_code,
				state_code = EXCLUDED.state_code,
				city_id = EXCLUDED.city_id,
				student_count = EXCLUDED.student_count,
				institution_type = EXCLUDED.institution_type,
				website = EXCLUDED.website,
				is_active = TRUE
			RETURNING id
		`, slug, strings.TrimSpace(row[col["name"]]), shortName,
			strings.ToUpper(strings.TrimSpace(row[col["country_code"]])), stateCode,
			cityID, students, instType, website).Scan(&id)
		if err != nil {
			return SeedResult{}, fmt.Errorf("insert university %s: %w", slug, err)
		}
		inserted++

		// Aliases: full lowercase name + short_name (if present) + "city uni"
		// pseudo-form. Source = "seed" so user-typed aliases don't collide.
		aliases := buildUniversityAliases(
			strings.TrimSpace(row[col["name"]]), shortName, stateCode)
		for _, a := range aliases {
			if a == "" {
				continue
			}
			_, err := r.pool.Exec(ctx, `
				INSERT INTO university_aliases (university_id, alias, source)
				VALUES ($1, $2, 'seed')
				ON CONFLICT (alias, source) DO NOTHING
			`, id, strings.ToLower(a))
			if err != nil {
				r.log.WithError(err).WithField("alias", a).Warn("insert alias failed")
				continue
			}
			aliasesInserted++
		}
	}
	return SeedResult{UniversitiesInserted: inserted, AliasesInserted: aliasesInserted}, nil
}

func buildUniversityAliases(name, shortName, stateCode string) []string {
	out := []string{name}
	if shortName != "" && !strings.EqualFold(shortName, name) {
		out = append(out, shortName)
	}
	if stateCode != "" {
		out = append(out, name+" "+stateCode)
	}
	return out
}

// SeedResult is the summary returned by SeedFromCSV.
type SeedResult struct {
	UniversitiesInserted int
	AliasesInserted      int
}

func (r *Repository) fetchCitySlugs(ctx context.Context) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug FROM cities`)
	if err != nil {
		return nil, fmt.Errorf("fetch city slugs: %w", err)
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var id int64
		var slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return nil, err
		}
		out[slug] = id
	}
	return out, rows.Err()
}

func indexHeader(header []string) map[string]int {
	out := map[string]int{}
	for i, h := range header {
		out[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return out
}

// FindSeedCSV tries a few likely locations.
func FindSeedCSV() (string, bool) {
	candidates := []string{
		"data/universities_seed.csv",
		"../data/universities_seed.csv",
		"../../data/universities_seed.csv",
		"../../../data/universities_seed.csv",
		"../../../../data/universities_seed.csv",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
}

// ── Lookups ──────────────────────────────────────────────────────────────

// GetBySlug returns a single university by its slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*University, error) {
	var u University
	var shortName, stateCode, website *string
	var cityID *int64
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, short_name, country_code, state_code, city_id,
			student_count, institution_type, website, is_active, created_at
		FROM universities WHERE slug = $1
	`, slug).Scan(&u.ID, &u.Slug, &u.Name, &shortName, &u.CountryCode, &stateCode,
		&cityID, &u.StudentCount, &u.InstitutionType, &website, &u.IsActive, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if shortName != nil {
		u.ShortName = *shortName
	}
	if stateCode != nil {
		u.StateCode = *stateCode
	}
	if website != nil {
		u.Website = *website
	}
	u.CityID = cityID
	return &u, nil
}

// LookupByAlias resolves a normalized university name to one or more
// matches. Returns up to 25 results, ranked by exact-alias match first,
// then city-prefix matches.
func (r *Repository) LookupByAlias(ctx context.Context, normalized string) ([]University, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.slug, u.name, u.short_name, u.country_code, u.state_code,
			u.city_id, u.student_count, u.institution_type, u.website, u.is_active, u.created_at
		FROM university_aliases a
		JOIN universities u ON u.id = a.university_id
		WHERE a.alias = $1
			AND u.is_active = TRUE
		LIMIT 25
	`, normalized)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []University
	for rows.Next() {
		var u University
		var shortName, stateCode, website *string
		var cityID *int64
		if err := rows.Scan(&u.ID, &u.Slug, &u.Name, &shortName, &u.CountryCode, &stateCode,
			&cityID, &u.StudentCount, &u.InstitutionType, &website, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		if shortName != nil {
			u.ShortName = *shortName
		}
		if stateCode != nil {
			u.StateCode = *stateCode
		}
		if website != nil {
			u.Website = *website
		}
		u.CityID = cityID
		out = append(out, u)
	}
	return out, rows.Err()
}

// ── Signals + recompute ──────────────────────────────────────────────────

// SignalsFor returns the activity vector for one university over the
// period [start, end). Same shape as cityranking.Signals.
func (r *Repository) SignalsFor(ctx context.Context, universityID int64, periodStart, periodEnd time.Time) (Signals, error) {
	var s Signals
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT u.id),
			COALESCE(SUM(CASE WHEN u.created_at >= $2 AND u.created_at < $3 THEN 1 ELSE 0 END), 0)
		FROM users u
		WHERE u.university_id = $1
			AND COALESCE(u.university_ranking_opted_out, FALSE) = FALSE
			AND (u.last_active_at IS NULL OR u.last_active_at >= $2)
	`, universityID, periodStart, periodEnd).Scan(&s.ActiveUsers, &s.NewUsers30d)
	if err != nil {
		return s, fmt.Errorf("active users: %w", err)
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_functions rf
		JOIN users u ON u.id = rf.owner_user_id
		WHERE u.university_id = $1
			AND COALESCE(u.university_ranking_opted_out, FALSE) = FALSE
			AND rf.created_at >= $2 AND rf.created_at < $3
	`, universityID, periodStart, periodEnd).Scan(&s.Deployments); err != nil {
		r.log.WithError(err).WithField("university_id", universityID).Debug("deployments failed, defaulting to 0")
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_function_executions rfe
		JOIN users u ON u.id = rfe.user_id
		WHERE u.university_id = $1
			AND COALESCE(u.university_ranking_opted_out, FALSE) = FALSE
			AND rfe.timestamp >= $2 AND rfe.timestamp < $3
	`, universityID, periodStart, periodEnd).Scan(&s.Executions30d); err != nil {
		r.log.WithError(err).WithField("university_id", universityID).Debug("executions failed, defaulting to 0")
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(LEAST(ac.total_earnings_cents, 1000000000000)), 0)
		FROM affiliate_codes ac
		JOIN users u ON u.id = ac.publisher_id
		WHERE u.university_id = $1
			AND COALESCE(u.university_ranking_opted_out, FALSE) = FALSE
	`, universityID).Scan(&s.FounderEarnings); err != nil {
		r.log.WithError(err).WithField("university_id", universityID).Debug("founder earnings failed, defaulting to 0")
	}
	return s, nil
}

// ListAll returns every active university. Used by the recompute job.
func (r *Repository) ListAll(ctx context.Context) ([]University, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, short_name, country_code, state_code, city_id,
			student_count, institution_type, website, is_active, created_at
		FROM universities WHERE is_active = TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []University
	for rows.Next() {
		var u University
		var shortName, stateCode, website *string
		var cityID *int64
		if err := rows.Scan(&u.ID, &u.Slug, &u.Name, &shortName, &u.CountryCode, &stateCode,
			&cityID, &u.StudentCount, &u.InstitutionType, &website, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		if shortName != nil {
			u.ShortName = *shortName
		}
		if stateCode != nil {
			u.StateCode = *stateCode
		}
		if website != nil {
			u.Website = *website
		}
		u.CityID = cityID
		out = append(out, u)
	}
	return out, rows.Err()
}

// ListUniversities returns the top-N active universities by their cached
// score. Honors the k=5 privacy threshold.
func (r *Repository) ListUniversities(ctx context.Context, country string, limit int, category Category) ([]Ranking, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{category, MinActiveUsersForPublic, limit}
	countryClause := ""
	if country != "" {
		countryClause = " AND u.country_code = $4"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT r.university_id, u.slug, u.name, COALESCE(u.short_name, ''),
			u.country_code, COALESCE(u.state_code, ''),
			COALESCE(c.slug, ''),
			r.rank_position, r.prev_rank_position,
			r.score_raw, r.score_per_capita,
			r.active_users, r.deployments, r.executions_30d,
			r.founder_earnings, r.new_users_30d,
			r.period_start, r.period_end
		FROM university_rankings r
		JOIN universities u ON u.id = r.university_id
		LEFT JOIN cities c ON c.id = u.city_id
		WHERE r.ranking_category = $1
			AND r.active_users >= $2
			AND u.is_active = TRUE`+countryClause+`
		ORDER BY r.score_per_capita DESC
		LIMIT $3
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ranking
	for rows.Next() {
		var rk Ranking
		var prev *int
		if err := rows.Scan(&rk.UniversityID, &rk.UniversitySlug, &rk.UniversityName, &rk.ShortName,
			&rk.CountryCode, &rk.StateCode, &rk.CitySlug,
			&rk.RankPosition, &prev,
			&rk.ScoreRaw, &rk.ScorePerCapita,
			&rk.ActiveUsers, &rk.Deployments, &rk.Executions30d,
			&rk.FounderEarnings, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd); err != nil {
			return nil, err
		}
		rk.PrevRank = prev
		if prev != nil {
			rk.RankDelta = *prev - rk.RankPosition
		}
		out = append(out, rk)
	}
	return out, rows.Err()
}

// GetRankingBySlug fetches the latest ranking row for a single university.
func (r *Repository) GetRankingBySlug(ctx context.Context, slug string, category Category) (*Ranking, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT r.university_id, u.slug, u.name, COALESCE(u.short_name, ''),
			u.country_code, COALESCE(u.state_code, ''),
			COALESCE(c.slug, ''),
			r.rank_position, r.prev_rank_position,
			r.score_raw, r.score_per_capita,
			r.active_users, r.deployments, r.executions_30d,
			r.founder_earnings, r.new_users_30d,
			r.period_start, r.period_end
		FROM university_rankings r
		JOIN universities u ON u.id = r.university_id
		LEFT JOIN cities c ON c.id = u.city_id
		WHERE u.slug = $1 AND r.ranking_category = $2
		ORDER BY r.period_end DESC LIMIT 1
	`, slug, category)
	var rk Ranking
	var prev *int
	if err := row.Scan(&rk.UniversityID, &rk.UniversitySlug, &rk.UniversityName, &rk.ShortName,
		&rk.CountryCode, &rk.StateCode, &rk.CitySlug,
		&rk.RankPosition, &prev,
		&rk.ScoreRaw, &rk.ScorePerCapita,
		&rk.ActiveUsers, &rk.Deployments, &rk.Executions30d,
		&rk.FounderEarnings, &rk.NewUsers30d,
		&rk.PeriodStart, &rk.PeriodEnd); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rk.PrevRank = prev
	if prev != nil {
		rk.RankDelta = *prev - rk.RankPosition
	}
	return &rk, nil
}

// LatestPeriod returns the most recent period_end in the rankings table.
func (r *Repository) LatestPeriod(ctx context.Context) (time.Time, error) {
	var t time.Time
	err := r.pool.QueryRow(ctx, `SELECT MAX(period_end) FROM university_rankings`).Scan(&t)
	return t, err
}

// IsOptedOut reports whether the user has opted out of the university
// leaderboard.
func (r *Repository) IsOptedOut(ctx context.Context, userID string) (bool, error) {
	var v bool
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(university_ranking_opted_out, FALSE) FROM users WHERE id = $1::uuid`, userID).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return v, err
}

// SetOptOut toggles the opt-out flag.
func (r *Repository) SetOptOut(ctx context.Context, userID string, optedOut bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET university_ranking_opted_out = $2 WHERE id = $1::uuid`, userID, optedOut)
	return err
}

// AssignUserUniversity writes the user's university_id (or NULL to clear).
// source is recorded in the audit log but the row is the same — schema
// constraint.
func (r *Repository) AssignUserUniversity(ctx context.Context, userID string, universityID *int64, source string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET university_id = $2 WHERE id = $1::uuid`,
		userID, universityID)
	if err != nil {
		return err
	}
	r.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"university_id": universityID,
		"source":       source,
	}).Debug("assigned user university")
	return nil
}

// GetUserUniversity returns the user's current university (slug + name) or
// nil if not set.
func (r *Repository) GetUserUniversity(ctx context.Context, userID string) (*University, error) {
	var u University
	var shortName, stateCode, website *string
	var cityID *int64
	err := r.pool.QueryRow(ctx, `
		SELECT univ.id, univ.slug, univ.name, univ.short_name, univ.country_code,
			univ.state_code, univ.city_id, univ.student_count, univ.institution_type,
			univ.website, univ.is_active, univ.created_at
		FROM users u
		JOIN universities univ ON univ.id = u.university_id
		WHERE u.id = $1::uuid
	`, userID).Scan(&u.ID, &u.Slug, &u.Name, &shortName, &u.CountryCode, &stateCode,
		&cityID, &u.StudentCount, &u.InstitutionType, &website, &u.IsActive, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if shortName != nil {
		u.ShortName = *shortName
	}
	if stateCode != nil {
		u.StateCode = *stateCode
	}
	if website != nil {
		u.Website = *website
	}
	u.CityID = cityID
	return &u, nil
}

// UpsertRanking writes (or updates) a single ranking row.
func (r *Repository) UpsertRanking(ctx context.Context, rk Ranking, category Category) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO university_rankings
			(university_id, rank_position, prev_rank_position, score_raw, score_per_capita,
			 active_users, deployments, executions_30d, founder_earnings, new_users_30d,
			 period_start, period_end, ranking_category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (university_id, period_end, ranking_category) DO UPDATE SET
			rank_position = EXCLUDED.rank_position,
			prev_rank_position = EXCLUDED.prev_rank_position,
			score_raw = EXCLUDED.score_raw,
			score_per_capita = EXCLUDED.score_per_capita,
			active_users = EXCLUDED.active_users,
			deployments = EXCLUDED.deployments,
			executions_30d = EXCLUDED.executions_30d,
			founder_earnings = EXCLUDED.founder_earnings,
			new_users_30d = EXCLUDED.new_users_30d,
			period_start = EXCLUDED.period_start,
			computed_at = NOW()
	`, rk.UniversityID, rk.RankPosition, rk.PrevRank, rk.ScoreRaw, rk.ScorePerCapita,
		rk.ActiveUsers, rk.Deployments, rk.Executions30d, rk.FounderEarnings, rk.NewUsers30d,
		rk.PeriodStart, rk.PeriodEnd, category)
	return err
}

// Pool returns the underlying pool. Used by the recompute job and tests.
func (r *Repository) Pool() *pgxpool.Pool { return r.pool }

// ListByMetro returns universities linked to cities in a metro area, ordered by per-capita rank.
func (r *Repository) ListByMetro(ctx context.Context, metroSlug string, limit int) ([]Ranking, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT r.university_id, u.slug, u.name, COALESCE(u.short_name, ''),
			u.country_code, COALESCE(u.state_code, ''),
			COALESCE(c.slug, ''),
			r.rank_position, r.prev_rank_position,
			r.score_raw, r.score_per_capita,
			r.active_users, r.deployments, r.executions_30d,
			r.founder_earnings, r.new_users_30d,
			r.period_start, r.period_end
		FROM university_rankings r
		JOIN universities u ON u.id = r.university_id
		LEFT JOIN cities c ON c.id = u.city_id
		WHERE r.ranking_category = 'composite'
			AND r.active_users >= $1
			AND u.is_active = TRUE
			AND c.metro_area_id = (SELECT id FROM metro_areas WHERE slug = $2)
		ORDER BY r.score_per_capita DESC
		LIMIT $3
	`, MinActiveUsersForPublic, metroSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ranking
	for rows.Next() {
		var rk Ranking
		var prev *int
		if err := rows.Scan(&rk.UniversityID, &rk.UniversitySlug, &rk.UniversityName, &rk.ShortName,
			&rk.CountryCode, &rk.StateCode, &rk.CitySlug,
			&rk.RankPosition, &prev,
			&rk.ScoreRaw, &rk.ScorePerCapita,
			&rk.ActiveUsers, &rk.Deployments, &rk.Executions30d,
			&rk.FounderEarnings, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd); err != nil {
			return nil, err
		}
		rk.PrevRank = prev
		if prev != nil {
			rk.RankDelta = rk.RankPosition - *prev
		}
		out = append(out, rk)
	}
	return out, rows.Err()
}
