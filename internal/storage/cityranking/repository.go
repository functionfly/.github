package cityranking

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Repository owns all city-ranking database access. It is intentionally built
// on a pgxpool (not GORM) so the cross-table aggregations stay efficient and
// the scoring query can be tuned with raw SQL.
type Repository struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

// NewRepository wires a repository to a pgx connection pool.
func NewRepository(pool *pgxpool.Pool, log *logrus.Logger) *Repository {
	if log == nil {
		log = logrus.StandardLogger()
	}
	return &Repository{pool: pool, log: log}
}

// Pool exposes the underlying pgxpool for callers that need direct access
// (e.g. the cache layer in internal/api/handlers/cityranking).
func (r *Repository) Pool() *pgxpool.Pool { return r.pool }

// ── Seeding ────────────────────────────────────────────────────────────────

// SeedResult is the summary of a successful seed run.
type SeedResult struct {
	MetrosInserted  int
	CitiesInserted  int
	AliasesInserted int
}

// SeedFromCSV loads metro_areas, cities, and city_aliases from a CSV file
// with header: slug,name,country_code,population,latitude,longitude,
// state_code,state_name,country_name,metro_slug
//
// Rows whose metro_slug does not match an existing metro area are skipped
// silently. Re-running is safe: the loader upserts on slug.
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

	metros := map[string]metroCSV{}
	cities := []cityCSV{}

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
		metroSlug := strings.TrimSpace(row[col["metro_slug"]])
		if metroSlug == "" {
			metroSlug = slug
		}
		pop, _ := strconv.Atoi(strings.TrimSpace(row[col["population"]]))
		lat, _ := strconv.ParseFloat(strings.TrimSpace(row[col["latitude"]]), 64)
		lon, _ := strconv.ParseFloat(strings.TrimSpace(row[col["longitude"]]), 64)
		metros[metroSlug] = metroCSV{
			Slug:        metroSlug,
			Name:        strings.TrimSpace(row[col["name"]]) + ", " + strings.TrimSpace(row[col["state_name"]]),
			CountryCode: strings.TrimSpace(row[col["country_code"]]),
			Population:  pop,
			Latitude:    lat,
			Longitude:   lon,
		}
		cities = append(cities, cityCSV{
			Slug:        slug,
			Name:        strings.TrimSpace(row[col["name"]]),
			StateCode:   strings.TrimSpace(row[col["state_code"]]),
			StateName:   strings.TrimSpace(row[col["state_name"]]),
			CountryCode: strings.TrimSpace(row[col["country_code"]]),
			CountryName: strings.TrimSpace(row[col["country_name"]]),
			Latitude:    lat,
			Longitude:   lon,
			Population:  pop,
			MetroSlug:   metroSlug,
		})
	}

	// First pass: insert metros
	metrosInserted := 0
	for _, m := range metros {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO metro_areas (slug, name, country_code, population, latitude, longitude, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				country_code = EXCLUDED.country_code,
				population = EXCLUDED.population,
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				is_active = TRUE
		`, m.Slug, m.Name, m.CountryCode, m.Population, m.Latitude, m.Longitude)
		if err != nil {
			return SeedResult{}, fmt.Errorf("insert metro %s: %w", m.Slug, err)
		}
		metrosInserted++
	}

	metroIDs, err := r.fetchMetroIDs(ctx)
	if err != nil {
		return SeedResult{}, err
	}

	citiesInserted := 0
	aliasesInserted := 0
	for _, c := range cities {
		metroID, ok := metroIDs[c.MetroSlug]
		if !ok {
			continue
		}
		var cityID int64
		err := r.pool.QueryRow(ctx, `
			INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name,
				latitude, longitude, population, metro_area_id, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				state_code = EXCLUDED.state_code,
				state_name = EXCLUDED.state_name,
				country_code = EXCLUDED.country_code,
				country_name = EXCLUDED.country_name,
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				population = EXCLUDED.population,
				metro_area_id = EXCLUDED.metro_area_id,
				is_active = TRUE
			RETURNING id
		`, c.Slug, c.Name, c.StateCode, c.StateName, c.CountryCode, c.CountryName,
			c.Latitude, c.Longitude, c.Population, metroID).Scan(&cityID)
		if err != nil {
			return SeedResult{}, fmt.Errorf("insert city %s: %w", c.Slug, err)
		}
		citiesInserted++

		// Aliases: the city name in lowercase and a "<name> <state_code>" form
		aliases := []struct{ alias, source string }{
			{strings.ToLower(c.Name), "seed"},
			{strings.ToLower(c.Name) + " " + strings.ToLower(c.StateCode), "seed"},
		}
		for _, a := range aliases {
			if a.alias == "" {
				continue
			}
			_, err := r.pool.Exec(ctx, `
				INSERT INTO city_aliases (city_id, alias, source)
				VALUES ($1, $2, $3)
				ON CONFLICT (alias, source) DO NOTHING
			`, cityID, a.alias, a.source)
			if err != nil {
				r.log.WithError(err).WithField("alias", a.alias).Warn("insert alias failed")
				continue
			}
			aliasesInserted++
		}
	}

	return SeedResult{
		MetrosInserted:  metrosInserted,
		CitiesInserted:  citiesInserted,
		AliasesInserted: aliasesInserted,
	}, nil
}

type metroCSV struct {
	Slug        string
	Name        string
	CountryCode string
	Population  int
	Latitude    float64
	Longitude   float64
}

type cityCSV struct {
	Slug        string
	Name        string
	StateCode   string
	StateName   string
	CountryCode string
	CountryName string
	Latitude    float64
	Longitude   float64
	Population  int
	MetroSlug   string
}

func indexHeader(h []string) map[string]int {
	out := make(map[string]int, len(h))
	for i, c := range h {
		out[strings.ToLower(strings.TrimSpace(c))] = i
	}
	return out
}

func (r *Repository) fetchMetroIDs(ctx context.Context) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug FROM metro_areas`)
	if err != nil {
		return nil, fmt.Errorf("query metros: %w", err)
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
	return out, nil
}

// ── Lookups ────────────────────────────────────────────────────────────────

// ListMetros returns the full list of active metros (used by the recompute
// job and admin tools).
func (r *Repository) ListMetros(ctx context.Context) ([]MetroArea, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, country_code, population,
			COALESCE(latitude, 0), COALESCE(longitude, 0),
			is_active, created_at
		FROM metro_areas
		WHERE is_active = TRUE
		ORDER BY population DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MetroArea{}
	for rows.Next() {
		var m MetroArea
		if err := rows.Scan(&m.ID, &m.Slug, &m.Name, &m.CountryCode, &m.Population,
			&m.Latitude, &m.Longitude, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// GetMetroBySlug returns a single metro by slug.
func (r *Repository) GetMetroBySlug(ctx context.Context, slug string) (*MetroArea, error) {
	var m MetroArea
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, country_code, population,
			COALESCE(latitude, 0), COALESCE(longitude, 0),
			is_active, created_at
		FROM metro_areas WHERE slug = $1
	`, slug).Scan(&m.ID, &m.Slug, &m.Name, &m.CountryCode, &m.Population,
		&m.Latitude, &m.Longitude, &m.IsActive, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

// ── City Ambassadors ─────────────────────────────────────────────────────
//
// One active ambassador per metro. The top user (by the same per-metro
// activity score used by the leaderboard) is auto-promoted after each
// recompute cycle. Manual override is also supported via UpsertAmbassador
// (admin only — not exposed in the public API yet).
//
// Privacy: the same k=5 threshold as the leaderboard is enforced, so a
// metro with fewer than 5 active opted-in builders has no ambassador.

// CandidateAmbassador is the per-user score used to pick the ambassador.
type CandidateAmbassador struct {
	UserID    string
	MetroID   int64
	FullName  string
	Email     string
	ScoreRaw  float64
	ActiveUsers int
}

// TopBuilderForMetro returns the highest-scoring active user in a metro
// (or nil if no eligible user). Used by the ambassador promotion job and
// the test suite.
func (r *Repository) TopBuilderForMetro(ctx context.Context, metroID int64) (*CandidateAmbassador, error) {
	row := r.pool.QueryRow(ctx, `
		WITH user_score AS (
			SELECT
				u.id, u.name, u.email,
				c.metro_area_id,
				(
					LEAST(1.0, 1.0) * LOG(1 + COALESCE((
						SELECT COUNT(DISTINCT rf.id) FROM registry_functions rf
						WHERE rf.owner_user_id = u.id AND rf.created_at >= NOW() - INTERVAL '30 days'
					), 0))
					+ 1.5 * LOG(1 + COALESCE((
						SELECT COUNT(*) FROM registry_function_executions rfe
						WHERE rfe.user_id = u.id AND rfe.timestamp >= NOW() - INTERVAL '30 days'
					), 0))
					+ 1.2 * LOG(1 + COALESCE((
						SELECT COUNT(*) FROM affiliate_codes ac
						WHERE ac.publisher_id = u.id
					), 0))
				) AS score
			FROM users u
			JOIN cities c ON c.id = u.city_id
			WHERE c.metro_area_id = $1
				AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
		)
		SELECT id, name, email, metro_area_id, score
		FROM user_score
		ORDER BY score DESC, id ASC
		LIMIT 1
	`, metroID)
	var c CandidateAmbassador
	if err := row.Scan(&c.UserID, &c.FullName, &c.Email, &c.MetroID, &c.ScoreRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// ListMetrosWithActiveBuilders returns metro IDs that pass the k=5
// privacy threshold. Used by the ambassador sync to know which metros
// are eligible for an ambassador.
func (r *Repository) ListMetrosWithActiveBuilders(ctx context.Context, minActive int) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cr.metro_area_id
		FROM city_rankings cr
		WHERE cr.ranking_category = 'composite'
			AND cr.active_users >= $1
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite')
		ORDER BY cr.metro_area_id
	`, minActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// GetAmbassadorForMetro returns the active ambassador for a metro, or nil
// if none.
func (r *Repository) GetAmbassadorForMetro(ctx context.Context, metroID int64) (*Ambassador, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT a.id, a.metro_id, a.user_id, u.name, u.email, u.profile_number,
			a.promoted_at, a.source
		FROM city_ambassadors a
		JOIN users u ON u.id = a.user_id
		WHERE a.metro_id = $1 AND a.revoked_at IS NULL
		LIMIT 1
	`, metroID)
	var amb Ambassador
	if err := row.Scan(&amb.ID, &amb.MetroID, &amb.UserID, &amb.FullName, &amb.Email, &amb.ProfileNumber, &amb.PromotedAt, &amb.Source); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &amb, nil
}

// ListAmbassadors returns the active ambassador for every metro (one row
// per metro), optionally filtered to a single country. Sorted by metro
// population descending so big hubs show first.
func (r *Repository) ListAmbassadors(ctx context.Context, country string, limit int) ([]AmbassadorListEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []any{limit}
	countryClause := ""
	if country != "" {
		countryClause = " AND m.country_code = $2"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT a.metro_id, m.slug, m.name, m.country_code,
			COALESCE(c.state_code, ''),
			COALESCE(c.slug, ''), a.user_id, u.name, u.username, u.email, u.profile_number,
			a.promoted_at, a.source
		FROM city_ambassadors a
		JOIN metro_areas m ON m.id = a.metro_id
		LEFT JOIN cities c ON c.id = (SELECT id FROM cities WHERE metro_area_id = m.id ORDER BY population DESC NULLS LAST LIMIT 1)
		JOIN users u ON u.id = a.user_id
		WHERE a.revoked_at IS NULL
			AND m.is_active = TRUE`+countryClause+`
		ORDER BY m.population DESC
		LIMIT $1
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AmbassadorListEntry
	for rows.Next() {
		var e AmbassadorListEntry
		var username sql.NullString
		if err := rows.Scan(&e.MetroID, &e.MetroSlug, &e.MetroName, &e.CountryCode, &e.StateCode,
			&e.CitySlug, &e.UserID, &e.FullName, &username, &e.Email, &e.ProfileNumber, &e.PromotedAt, &e.Source); err != nil {
			return nil, err
		}
		if username.Valid {
			e.Username = &username.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListCountries returns distinct countries that have active metros, with names
// resolved from the cities table. This allows the ambassadors filter to be
// dynamic — new countries are automatically included when metros are added.
func (r *Repository) ListCountries(ctx context.Context) ([]Country, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (m.country_code) m.country_code,
			COALESCE(
				(SELECT c2.country_name FROM cities c2 WHERE c2.metro_area_id = m.id AND c2.country_code = m.country_code ORDER BY c2.country_name LIMIT 1),
				m.country_code
			) AS country_name
		FROM metro_areas m
		WHERE m.is_active = TRUE AND m.country_code != ''
		ORDER BY m.country_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Country
	for rows.Next() {
		var c Country
		if err := rows.Scan(&c.Code, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpsertAmbassador inserts (or reactivates) an ambassador row. If a
// different user is currently ambassador, the old row is revoked
// (revoked_at = now) and the new one is created. Idempotent: calling
// twice with the same (metro_id, user_id) is a no-op.
func (r *Repository) UpsertAmbassador(ctx context.Context, metroID int64, userID string, source string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Revoke any other active ambassador for this metro.
	if _, err := tx.Exec(ctx, `
		UPDATE city_ambassadors
		SET revoked_at = NOW()
		WHERE metro_id = $1 AND user_id <> $2::uuid AND revoked_at IS NULL
	`, metroID, userID); err != nil {
		return fmt.Errorf("revoke old: %w", err)
	}

	// Insert (or reactivate) this row. ON CONFLICT handles the
	// (metro_id, user_id) unique constraint.
	if _, err := tx.Exec(ctx, `
		INSERT INTO city_ambassadors (metro_id, user_id, source, promoted_at, revoked_at)
		VALUES ($1, $2::uuid, $3, NOW(), NULL)
		ON CONFLICT (metro_id, user_id) DO UPDATE SET
			promoted_at = NOW(),
			revoked_at = NULL,
			source = EXCLUDED.source
	`, metroID, userID, source); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	return tx.Commit(ctx)
}

// RevokeAmbassador marks the active ambassador for a metro as revoked.
// Used when a metro drops below the privacy threshold.
func (r *Repository) RevokeAmbassador(ctx context.Context, metroID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE city_ambassadors
		SET revoked_at = NOW()
		WHERE metro_id = $1 AND revoked_at IS NULL
	`, metroID)
	return err
}

// Ambassador is the public type returned by GetAmbassadorForMetro.
type Ambassador struct {
	ID            int64     `json:"id"`
	MetroID       int64     `json:"metro_id"`
	UserID        string    `json:"user_id"`
	FullName      string    `json:"name"`
	Email         string    `json:"-"` // never expose
	ProfileNumber *int      `json:"profile_number,omitempty"`
	PromotedAt    time.Time `json:"promoted_at"`
	Source        string    `json:"source"`
}

// AmbassadorListEntry is one row in the public ambassador list.
type AmbassadorListEntry struct {
	MetroID       int64     `json:"metro_id"`
	MetroSlug     string    `json:"metro_slug"`
	MetroName     string    `json:"metro_name"`
	CountryCode   string    `json:"country_code"`
	StateCode     string    `json:"state_code,omitempty"`
	CitySlug      string    `json:"city_slug,omitempty"`
	UserID        string    `json:"user_id"`
	FullName      string    `json:"name"`
	Username      *string   `json:"username,omitempty"`
	Email         string    `json:"-"` // never expose
	ProfileNumber *int      `json:"profile_number,omitempty"`
	PromotedAt    time.Time `json:"promoted_at"`
	Source        string    `json:"source"`
}

// GetCityBySlug returns a single city by its slug. Used by the IP-geo
// fallback path (HandleSetMyCity) when the front-end posts a `slug` rather
// than a free-text `input`.
func (r *Repository) GetCityBySlug(ctx context.Context, slug string) (*City, error) {
	var c City
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, state_code, state_name, country_code, country_name,
			latitude, longitude, COALESCE(population, 0), metro_area_id, created_at
		FROM cities WHERE slug = $1
	`, slug).Scan(&c.ID, &c.Slug, &c.Name, &c.StateCode, &c.StateName, &c.CountryCode,
		&c.CountryName, &c.Latitude, &c.Longitude, &c.Population, &c.MetroAreaID, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

// LookupCityByAlias resolves a normalized user-typed location to a city.
// Match priority: exact alias → "<city> <state_code>" alias → fallback
// "<city> <state_name>" → city name only.
func (r *Repository) LookupCityByAlias(ctx context.Context, normalized string) ([]City, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name, c.country_code,
			c.country_name, c.latitude, c.longitude, COALESCE(c.population, 0),
			c.metro_area_id, c.created_at
		FROM city_aliases a
		JOIN cities c ON c.id = a.city_id
		WHERE a.alias = $1
			AND c.is_active = TRUE
		LIMIT 25
	`, normalized)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCities(rows)
}

// SearchCitiesByName is a fallback that matches by city name when the alias
// table has no entry.
func (r *Repository) SearchCitiesByName(ctx context.Context, normalized string) ([]City, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name, c.country_code,
			c.country_name, c.latitude, c.longitude, COALESCE(c.population, 0),
			c.metro_area_id, c.created_at
		FROM cities c
		WHERE LOWER(c.name) = $1
			AND c.is_active = TRUE
		LIMIT 25
	`, normalized)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCities(rows)
}

func scanCities(rows pgx.Rows) ([]City, error) {
	out := []City{}
	for rows.Next() {
		var c City
		var metroID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.StateCode, &c.StateName, &c.CountryCode,
			&c.CountryName, &c.Latitude, &c.Longitude, &c.Population, &metroID, &c.CreatedAt); err != nil {
			return nil, err
		}
		if metroID.Valid {
			v := metroID.Int64
			c.MetroAreaID = &v
		}
		out = append(out, c)
	}
	return out, nil
}

// AssignUserCity stores a user's resolved city. Setting cityID to nil clears
// the assignment.
func (r *Repository) AssignUserCity(ctx context.Context, userID string, cityID *int64, source string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users
		SET city_id = $2
		WHERE id = $1
	`, userID, cityID)
	return err
}

// SetOptOut toggles a user's opt-out from city rankings.
func (r *Repository) SetOptOut(ctx context.Context, userID string, optedOut bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users
		SET city_ranking_opted_out = $2
		WHERE id = $1
	`, userID, optedOut)
	return err
}

// IsOptedOut reports a user's current opt-out status. Returns (false, nil) if
// the user is not found.
func (r *Repository) IsOptedOut(ctx context.Context, userID string) (bool, error) {
	var v bool
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(city_ranking_opted_out, FALSE)
		FROM users WHERE id = $1
	`, userID).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return v, err
}

// GetUserMetro returns the metro that the user is currently counted in,
// or nil if they have no city assignment or have opted out.
func (r *Repository) GetUserMetro(ctx context.Context, userID string) (*MetroArea, error) {
	var slug sql.NullString
	err := r.pool.QueryRow(ctx, `
		SELECT m.slug
		FROM users u
		JOIN cities c ON c.id = u.city_id
		LEFT JOIN metro_areas m ON m.id = c.metro_area_id
		WHERE u.id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
	`, userID).Scan(&slug)
	if errors.Is(err, pgx.ErrNoRows) || !slug.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.GetMetroBySlug(ctx, slug.String)
}

// ── Aggregations ───────────────────────────────────────────────────────────

// MetroSignals computes the raw activity signals for a single metro, used
// by the recompute job. The 30-day window ends at `now` (truncated to the
// hour by the caller).
func (r *Repository) MetroSignals(ctx context.Context, metroID int64, periodStart, periodEnd time.Time) (Signals, error) {
	out := Signals{}

	// 1) Active users in last 30d (any user with city_id in this metro who is
	//    not opted-out and who was active during the window).
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND (
				u.last_active_at IS NULL
				OR u.last_active_at >= $2
			)
	`, metroID, periodStart).Scan(&out.ActiveUsers)
	if err != nil {
		return out, fmt.Errorf("active users: %w", err)
	}

	// 2) New users in the 30d window.
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users u
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND u.created_at >= $2
			AND u.created_at < $3
	`, metroID, periodStart, periodEnd).Scan(&out.NewUsers30d)
	if err != nil {
		return out, fmt.Errorf("new users: %w", err)
	}

	// 3) Deployments attributed to users in this metro.
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_functions rf
		JOIN users u ON u.id = rf.owner_user_id
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND rf.created_at >= $2
			AND rf.created_at < $3
	`, metroID, periodStart, periodEnd).Scan(&out.Deployments)
	if err != nil {
		// registry_functions may not have owner_user_id populated in every
		// installation; treat that as 0 rather than failing the whole job.
		r.log.WithError(err).WithField("metro_id", metroID).Debug("deployments query failed, defaulting to 0")
		out.Deployments = 0
	}

	// 4) Function executions in the 30d window, attributed by the executing
	//    user's metro.
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_function_executions rfe
		JOIN users u ON u.id = rfe.user_id
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND rfe.timestamp >= $2
			AND rfe.timestamp < $3
	`, metroID, periodStart, periodEnd).Scan(&out.Executions30d)
	if err != nil {
		r.log.WithError(err).WithField("metro_id", metroID).Debug("executions query failed, defaulting to 0")
		out.Executions30d = 0
	}

	// 5) Founder-style earnings (referral commissions earned by users in
	//    this metro, all-time). Capped at 1e12 to avoid runaway numbers.
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(LEAST(ac.total_earnings_cents, 1000000000000)), 0)
		FROM affiliate_codes ac
		JOIN users u ON u.id = ac.publisher_id
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
	`, metroID).Scan(&out.FounderEarnings)
	if err != nil {
		r.log.WithError(err).WithField("metro_id", metroID).Debug("founder earnings query failed, defaulting to 0")
		out.FounderEarnings = 0
	}

	return out, nil
}

// ── Materialized rankings ──────────────────────────────────────────────────

// UpsertRanking writes (or replaces) a city's score row for the given
// period_end and category. Previous rank is preserved from the immediately
// prior period.
func (r *Repository) UpsertRanking(ctx context.Context, metroID int64, category Category, score ScoreResult, periodStart, periodEnd time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO city_rankings (
			metro_area_id, ranking_category, score_raw, score_per_capita,
			active_users, deployments, executions_30d,
			founder_earnings, new_users_30d,
			period_start, period_end, computed_at, rank_position, prev_rank_position
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), 0, NULL)
		ON CONFLICT (metro_area_id, period_end, ranking_category) DO UPDATE SET
			score_raw = EXCLUDED.score_raw,
			score_per_capita = EXCLUDED.score_per_capita,
			active_users = EXCLUDED.active_users,
			deployments = EXCLUDED.deployments,
			executions_30d = EXCLUDED.executions_30d,
			founder_earnings = EXCLUDED.founder_earnings,
			new_users_30d = EXCLUDED.new_users_30d,
			period_start = EXCLUDED.period_start,
			computed_at = NOW()
	`, metroID, string(category), score.Raw, score.PerCapita, score.ActiveUsers, score.Deployments,
		score.Executions30d, score.FounderCents, score.NewUsers30d, periodStart, periodEnd)
	return err
}

// AssignRanks updates the rank_position column for a given period_end and
// category, based on the current per-capita score ordering. Each category
// gets its own ranking — the composite leaderboard is just one of them.
func (r *Repository) AssignRanks(ctx context.Context, periodEnd time.Time, category Category) error {
	_, err := r.pool.Exec(ctx, `
		WITH ranked AS (
			SELECT id,
				ROW_NUMBER() OVER (ORDER BY score_per_capita DESC, metro_area_id) AS new_rank
			FROM city_rankings
			WHERE period_end = $1
				AND ranking_category = $2
				AND active_users >= $3
		)
		UPDATE city_rankings cr
		SET prev_rank_position = cr.rank_position,
			rank_position = ranked.new_rank
		FROM ranked
		WHERE cr.id = ranked.id
	`, periodEnd, string(category), MinActiveUsersForPublic)
	return err
}

// ListRankings returns the current top-N metros for the most recent period in
// the given category. Cities below the privacy threshold are excluded.
func (r *Repository) ListRankings(ctx context.Context, limit int, country string, category Category) ([]Ranking, error) {
	args := []any{string(category), MinActiveUsersForPublic, limit}
	where := "cr.ranking_category = $1 AND cr.active_users >= $2"
	if country != "" {
		where += " AND m.country_code = $4"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT cr.id, cr.metro_area_id, m.slug, m.name, m.country_code, m.population,
			cr.ranking_category,
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.founder_earnings, cr.new_users_30d,
			cr.period_start, cr.period_end, cr.computed_at
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $1)
			AND `+where+`
		ORDER BY cr.rank_position ASC
		LIMIT $3
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ranking{}
	for rows.Next() {
		var rk Ranking
		if err := rows.Scan(&rk.ID, &rk.MetroAreaID, &rk.MetroSlug, &rk.MetroName,
			&rk.CountryCode, &rk.Population, &rk.Category,
			&rk.RankPosition, &rk.PrevRankPosition,
			&rk.ScoreRaw, &rk.ScorePerCapita, &rk.ActiveUsers, &rk.Deployments,
			&rk.Executions30d, &rk.FounderEarnings, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd, &rk.ComputedAt); err != nil {
			return nil, err
		}
		if rk.PrevRankPosition.Valid {
			rk.PrevRank = int(rk.PrevRankPosition.Int64)
		}
		out = append(out, rk)
	}
	return out, nil
}

// GetRankingBySlug returns a single metro's current ranking row in the given
// category, or nil if the metro isn't ranked (e.g. below the privacy
// threshold).
func (r *Repository) GetRankingBySlug(ctx context.Context, slug string, category Category) (*Ranking, error) {
	var rk Ranking
	err := r.pool.QueryRow(ctx, `
		SELECT cr.id, cr.metro_area_id, m.slug, m.name, m.country_code, m.population,
			cr.ranking_category,
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.founder_earnings, cr.new_users_30d,
			cr.period_start, cr.period_end, cr.computed_at
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE m.slug = $1
			AND cr.ranking_category = $2
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $2)
	`, slug, string(category)).Scan(&rk.ID, &rk.MetroAreaID, &rk.MetroSlug, &rk.MetroName,
		&rk.CountryCode, &rk.Population, &rk.Category,
		&rk.RankPosition, &rk.PrevRankPosition,
		&rk.ScoreRaw, &rk.ScorePerCapita, &rk.ActiveUsers, &rk.Deployments,
		&rk.Executions30d, &rk.FounderEarnings, &rk.NewUsers30d,
		&rk.PeriodStart, &rk.PeriodEnd, &rk.ComputedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if rk.PrevRankPosition.Valid {
		rk.PrevRank = int(rk.PrevRankPosition.Int64)
	}
	return &rk, nil
}

// ListHistory returns the last `days` days of ranking rows for a metro in the
// given category. Used for the sparkline on the metro detail page.
func (r *Repository) ListHistory(ctx context.Context, metroSlug string, days int, category Category) ([]Ranking, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cr.id, cr.metro_area_id, m.slug, m.name, m.country_code, m.population,
			cr.ranking_category,
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.founder_earnings, cr.new_users_30d,
			cr.period_start, cr.period_end, cr.computed_at
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE m.slug = $1
			AND cr.ranking_category = $2
			AND cr.period_end >= NOW() - ($3 || ' hours')::interval
		ORDER BY cr.period_end ASC
	`, metroSlug, string(category), days*24)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ranking{}
	for rows.Next() {
		var rk Ranking
		if err := rows.Scan(&rk.ID, &rk.MetroAreaID, &rk.MetroSlug, &rk.MetroName,
			&rk.CountryCode, &rk.Population, &rk.Category,
			&rk.RankPosition, &rk.PrevRankPosition,
			&rk.ScoreRaw, &rk.ScorePerCapita, &rk.ActiveUsers, &rk.Deployments,
			&rk.Executions30d, &rk.FounderEarnings, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd, &rk.ComputedAt); err != nil {
			return nil, err
		}
		if rk.PrevRankPosition.Valid {
			rk.PrevRank = int(rk.PrevRankPosition.Int64)
		}
		out = append(out, rk)
	}
	return out, nil
}

// ListMovers returns the metros with the biggest rank deltas in the most
// recent period for the given category.
func (r *Repository) ListMovers(ctx context.Context, direction string, limit int, category Category) ([]Ranking, error) {
	if limit <= 0 || limit > 200 {
		limit = 25
	}
	order := "DESC"
	if direction == "losers" {
		order = "ASC"
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT cr.id, cr.metro_area_id, m.slug, m.name, m.country_code, m.population,
			cr.ranking_category,
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.founder_earnings, cr.new_users_30d,
			cr.period_start, cr.period_end, cr.computed_at
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $3)
			AND cr.ranking_category = $3
			AND cr.active_users >= $1
			AND cr.prev_rank_position IS NOT NULL
		ORDER BY (cr.prev_rank_position - cr.rank_position) %s, cr.rank_position ASC
		LIMIT $2
	`, order), MinActiveUsersForPublic, limit, string(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ranking{}
	for rows.Next() {
		var rk Ranking
		if err := rows.Scan(&rk.ID, &rk.MetroAreaID, &rk.MetroSlug, &rk.MetroName,
			&rk.CountryCode, &rk.Population, &rk.Category,
			&rk.RankPosition, &rk.PrevRankPosition,
			&rk.ScoreRaw, &rk.ScorePerCapita, &rk.ActiveUsers, &rk.Deployments,
			&rk.Executions30d, &rk.FounderEarnings, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd, &rk.ComputedAt); err != nil {
			return nil, err
		}
		if rk.PrevRankPosition.Valid {
			rk.PrevRank = int(rk.PrevRankPosition.Int64)
		}
		out = append(out, rk)
	}
	return out, nil
}

// LatestPeriod returns the most recent period_end stored in city_rankings.
func (r *Repository) LatestPeriod(ctx context.Context) (time.Time, error) {
	var t sql.NullTime
	err := r.pool.QueryRow(ctx, `SELECT MAX(period_end) FROM city_rankings`).Scan(&t)
	if err != nil {
		return time.Time{}, err
	}
	if !t.Valid {
		return time.Time{}, nil
	}
	return t.Time, nil
}

// ── State rankings ────────────────────────────────────────────────────────

// ListStateRankings aggregates the most-recent city_rankings rows by
// (country_code, state_code) for the given category. The state code/name is
// derived from the cities table (the first active city in the metro is used
// as canonical). States with zero ranked metros are omitted; states whose
// summed active_users fall below MinActiveUsersForPublic are filtered out
// (k-anonymity).
func (r *Repository) ListStateRankings(ctx context.Context, country string, limit int, category Category) ([]StateRanking, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{string(category), MinActiveUsersForPublic, limit}
	countryClause := ""
	if country != "" {
		countryClause = " AND m.country_code = $4"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.state_code,
			MIN(c.state_name) AS state_name,
			m.country_code,
			SUM(m.population)::int AS population,
			SUM(cr.score_raw) AS score_raw,
			SUM(cr.score_per_capita) AS score_per_capita,
			SUM(cr.active_users) AS active_users,
			SUM(cr.deployments) AS deployments,
			SUM(cr.executions_30d) AS executions_30d,
			COUNT(DISTINCT m.id) AS metro_count,
			COUNT(DISTINCT CASE WHEN cr.id IS NOT NULL THEN m.id END) AS ranked_metros,
			MAX(cr.period_end) AS period_end
		FROM metro_areas m
		JOIN cities c ON c.metro_area_id = m.id AND c.is_active = TRUE
		LEFT JOIN city_rankings cr
			ON cr.metro_area_id = m.id
			AND cr.ranking_category = $1
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $1)
		WHERE c.state_code <> ''
			`+countryClause+`
		GROUP BY c.state_code, m.country_code
		HAVING SUM(cr.active_users) >= $2
			AND COUNT(DISTINCT CASE WHEN cr.id IS NOT NULL THEN m.id END) > 0
		ORDER BY score_per_capita DESC
		LIMIT $3
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StateRanking{}
	for rows.Next() {
		var s StateRanking
		if err := rows.Scan(&s.StateCode, &s.StateName, &s.CountryCode,
			&s.Population, &s.ScoreRaw, &s.ScorePerCapita,
			&s.ActiveUsers, &s.Deployments, &s.Executions30d,
			&s.MetroCount, &s.RankedMetros, &s.PeriodEnd); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	for i := range out {
		out[i].RankPosition = i + 1
	}
	return out, nil
}

// GetStateRankingByCode returns a single state's aggregate for the most
// recent period and category, or (nil, nil) if the state doesn't exist or
// has no ranked metros.
func (r *Repository) GetStateRankingByCode(ctx context.Context, country, stateCode string, category Category) (*StateRanking, error) {
	stateCode = strings.ToUpper(strings.TrimSpace(stateCode))
	if stateCode == "" {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.state_code,
			MIN(c.state_name) AS state_name,
			m.country_code,
			SUM(m.population)::int AS population,
			SUM(cr.score_raw) AS score_raw,
			SUM(cr.score_per_capita) AS score_per_capita,
			SUM(cr.active_users) AS active_users,
			SUM(cr.deployments) AS deployments,
			SUM(cr.executions_30d) AS executions_30d,
			COUNT(DISTINCT m.id) AS metro_count,
			COUNT(DISTINCT CASE WHEN cr.id IS NOT NULL THEN m.id END) AS ranked_metros,
			MAX(cr.period_end) AS period_end
		FROM metro_areas m
		JOIN cities c ON c.metro_area_id = m.id AND c.is_active = TRUE
		LEFT JOIN city_rankings cr
			ON cr.metro_area_id = m.id
			AND cr.ranking_category = $3
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $3)
		WHERE c.state_code = $1
			AND m.country_code = $2
		GROUP BY c.state_code, m.country_code
		HAVING SUM(cr.active_users) >= $4
	`, stateCode, strings.ToUpper(country), string(category), MinActiveUsersForPublic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var s StateRanking
	if err := rows.Scan(&s.StateCode, &s.StateName, &s.CountryCode,
		&s.Population, &s.ScoreRaw, &s.ScorePerCapita,
		&s.ActiveUsers, &s.Deployments, &s.Executions30d,
		&s.MetroCount, &s.RankedMetros, &s.PeriodEnd); err != nil {
		return nil, err
	}
	return &s, nil
}

// ── Map points ────────────────────────────────────────────────────────────

// ListMapPoints returns all ranked metros for the most recent period in the
// given category with coordinates (taken from metro_areas) and a tier
// label. The state_code is derived from the canonical city in the metro.
// Intended for the AI World Map; for the leaderboard table, use
// ListRankings instead.
func (r *Repository) ListMapPoints(ctx context.Context, category Category) ([]MapPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			m.slug, m.name, m.country_code,
			COALESCE(canonical.state_code, '') AS state_code,
			COALESCE(m.latitude, 0) AS latitude,
			COALESCE(m.longitude, 0) AS longitude,
			m.population,
			cr.rank_position,
			cr.score_per_capita,
			cr.active_users
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		LEFT JOIN LATERAL (
			SELECT state_code
			FROM cities
			WHERE metro_area_id = m.id AND is_active = TRUE AND state_code <> ''
			ORDER BY id
			LIMIT 1
		) canonical ON TRUE
		WHERE cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = $2)
			AND cr.ranking_category = $2
			AND cr.active_users >= $1
		ORDER BY cr.score_per_capita DESC
	`, MinActiveUsersForPublic, string(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MapPoint{}
	for rows.Next() {
		var p MapPoint
		if err := rows.Scan(&p.MetroSlug, &p.MetroName, &p.CountryCode,
			&p.StateCode, &p.Latitude, &p.Longitude, &p.Population,
			&p.RankPosition, &p.ScorePerCapita, &p.ActiveUsers); err != nil {
			return nil, err
		}
		p.Tier = Tier(p.ScorePerCapita)
		out = append(out, p)
	}
	return out, nil
}

// ── City-proper (sub-metro) rankings ──────────────────────────────────────

// CityRanking is a single city's score under a category. Used by the
// /v1/city-rankings/cities leaderboard, which lets the front-end toggle off
// the MSA rollup.
type CityRanking struct {
	CityID         int64
	CitySlug       string
	CityName       string
	StateCode      string
	StateName      string
	CountryCode    string
	CountryName    string
	MetroSlug      *string
	MetroName      *string
	Population     int
	ScoreRaw       float64
	ScorePerCapita float64
	ActiveUsers    int
	Deployments    int
	Executions30d  int64
	NewUsers30d    int
}

// CitySignals computes the same Signals struct as MetroSignals but scoped to
// a single city (its own user pool, not the metro's).
func (r *Repository) CitySignals(ctx context.Context, cityID int64, periodStart, periodEnd time.Time) (Signals, error) {
	out := Signals{}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE u.city_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND (u.last_active_at IS NULL OR u.last_active_at >= $2)
	`, cityID, periodStart).Scan(&out.ActiveUsers); err != nil {
		return out, fmt.Errorf("active users (city): %w", err)
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users u
		WHERE u.city_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND u.created_at >= $2 AND u.created_at < $3
	`, cityID, periodStart, periodEnd).Scan(&out.NewUsers30d); err != nil {
		return out, fmt.Errorf("new users (city): %w", err)
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_functions rf
		JOIN users u ON u.id = rf.owner_user_id
		WHERE u.city_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND rf.created_at >= $2 AND rf.created_at < $3
	`, cityID, periodStart, periodEnd).Scan(&out.Deployments); err != nil {
		r.log.WithError(err).WithField("city_id", cityID).Debug("deployments (city) failed, defaulting to 0")
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM registry_function_executions rfe
		JOIN users u ON u.id = rfe.user_id
		WHERE u.city_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND rfe.timestamp >= $2 AND rfe.timestamp < $3
	`, cityID, periodStart, periodEnd).Scan(&out.Executions30d); err != nil {
		r.log.WithError(err).WithField("city_id", cityID).Debug("executions (city) failed, defaulting to 0")
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(LEAST(ac.total_earnings_cents, 1000000000000)), 0)
		FROM affiliate_codes ac
		JOIN users u ON u.id = ac.publisher_id
		WHERE u.city_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
	`, cityID).Scan(&out.FounderEarnings); err != nil {
		r.log.WithError(err).WithField("city_id", cityID).Debug("founder earnings (city) failed, defaulting to 0")
	}

	return out, nil
}

// ListCityRankings computes a per-city leaderboard on the fly (no materialized
// city-level rows). The privacy threshold filters cities with fewer than
// MinActiveUsersForPublic active builders. Performance note: this is O(N) where
// N is the number of seeded cities (a few hundred at v1). If this becomes a
// hot path, materialize a `city_rankings_city` table.
func (r *Repository) ListCityRankings(ctx context.Context, country string, limit int, category Category) ([]CityRanking, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	weights := CategoryWeights(category)
	periodEnd := cityrankingNow(ctx, r)
	periodStart := periodEnd.Add(-30 * 24 * time.Hour)
	// Fetch a generous candidate pool so the post-filter (privacy + sort +
	// user limit) has enough rows to work with. The user-facing limit is
	// applied AFTER scoring; the candidate limit is the hard ceiling for the
	// raw SQL query. At v1 we have ~230 seeded cities, so 500 is safe.
	const candidateLimit = 500
	args := []any{periodStart, candidateLimit}
	countryClause := ""
	if country != "" {
		countryClause = " AND c.country_code = $3"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name, c.country_code, c.country_name,
			COALESCE(c.population, 0),
			m.slug, m.name,
			COALESCE((
				SELECT COUNT(DISTINCT u.id)
				FROM users u
				WHERE u.city_id = c.id
					AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
					AND (u.last_active_at IS NULL OR u.last_active_at >= $1)
			), 0) AS active_users
		FROM cities c
		LEFT JOIN metro_areas m ON m.id = c.metro_area_id
		WHERE c.is_active = TRUE`+countryClause+`
		ORDER BY c.population DESC NULLS LAST
		LIMIT $2
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type row struct {
		c        CityRanking
		activeN  int
	}
	var candidates []row
	for rows.Next() {
		var rk row
		var metroSlug, metroName sql.NullString
		if err := rows.Scan(&rk.c.CityID, &rk.c.CitySlug, &rk.c.CityName,
			&rk.c.StateCode, &rk.c.StateName, &rk.c.CountryCode, &rk.c.CountryName,
			&rk.c.Population, &metroSlug, &metroName, &rk.activeN); err != nil {
			return nil, err
		}
		if metroSlug.Valid {
			v := metroSlug.String
			rk.c.MetroSlug = &v
		}
		if metroName.Valid {
			v := metroName.String
			rk.c.MetroName = &v
		}
		candidates = append(candidates, rk)
	}

	out := make([]CityRanking, 0, len(candidates))
	for _, c := range candidates {
		if c.activeN < MinActiveUsersForPublic {
			continue
		}
		signals, err := r.CitySignals(ctx, c.c.CityID, periodStart, periodEnd)
		if err != nil {
			r.log.WithError(err).WithField("city_id", c.c.CityID).Warn("city signals failed")
			continue
		}
		s := Compute(signals, c.c.Population, weights)
		c.c.ScoreRaw = s.Raw
		c.c.ScorePerCapita = s.PerCapita
		c.c.ActiveUsers = s.ActiveUsers
		c.c.Deployments = s.Deployments
		c.c.Executions30d = s.Executions30d
		c.c.NewUsers30d = s.NewUsers30d
		out = append(out, c.c)
	}
	sortCityRankings(out)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func sortCityRankings(rs []CityRanking) {
	for i := 1; i < len(rs); i++ {
		for j := i; j > 0 && rs[j-1].ScorePerCapita < rs[j].ScorePerCapita; j-- {
			rs[j-1], rs[j] = rs[j], rs[j-1]
		}
	}
}

func cityrankingNow(ctx context.Context, r *Repository) time.Time {
	t, err := r.LatestPeriod(ctx)
	if err != nil || t.IsZero() {
		return TruncateHour(time.Now())
	}
	return t
}

// ── Builders (anonymized top contributors) ─────────────────────────────────

// Builder is one anonymized top contributor in a metro. Public profile fields
// (name, profile_number) are only included when the user has explicitly
// opted in to public profile visibility. Builders are returned in descending
// order of total activity (deployments + executions + active weight).
type Builder struct {
	UserID         string
	DisplayName    string  // empty if profile is private
	ProfileNumber  *int    // nil if private
	ProfilePublic  bool
	Deployments    int
	Executions30d  int64
	ScoreComposite float64
	Rank           int
}

// ListBuilders returns the top contributors for a metro. The result is
// suppressed (returns an empty slice, not an error) unless the metro has at
// least MinActiveUsersForPublic active builders, enforcing k-anonymity.
func (r *Repository) ListBuilders(ctx context.Context, metroSlug string, limit int, category Category) ([]Builder, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	weights := CategoryWeights(category)

	var metroID int64
	if err := r.pool.QueryRow(ctx, `SELECT id FROM metro_areas WHERE slug = $1`, metroSlug).Scan(&metroID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Active-builder count is the k-anonymity guard.
	var activeCount int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
			AND (u.last_active_at IS NULL OR u.last_active_at >= NOW() - INTERVAL '30 days')
	`, metroID).Scan(&activeCount); err != nil {
		return nil, fmt.Errorf("active count: %w", err)
	}
	if activeCount < MinActiveUsersForPublic {
		return []Builder{}, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id::text,
			COALESCE(NULLIF(u.name, ''), 'Anonymous') AS display_name,
			u.profile_number,
			COALESCE((
				SELECT COUNT(*) FROM registry_functions rf
				WHERE rf.owner_user_id = u.id
			), 0) AS deployments,
			COALESCE((
				SELECT COUNT(*) FROM registry_function_executions rfe
				WHERE rfe.user_id = u.id
					AND rfe.timestamp >= NOW() - INTERVAL '30 days'
			), 0)::bigint AS executions_30d
		FROM users u
		JOIN cities c ON c.id = u.city_id
		WHERE c.metro_area_id = $1
			AND COALESCE(u.city_ranking_opted_out, FALSE) = FALSE
		ORDER BY (
			COALESCE((SELECT COUNT(*) FROM registry_functions rf WHERE rf.owner_user_id = u.id), 0)
			+ COALESCE((SELECT COUNT(*) FROM registry_function_executions rfe WHERE rfe.user_id = u.id AND rfe.timestamp >= NOW() - INTERVAL '30 days'), 0)
		) DESC
		LIMIT $2
	`, metroID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// The "score" is computed from the same per-builder signals as the metro
	// score, so the leaderboard reads as a coherent extension of the public
	// ranking. We use log scaling + the category weights for parity.
	out := []Builder{}
	for rows.Next() {
		var b Builder
		if err := rows.Scan(&b.UserID, &b.DisplayName, &b.ProfileNumber, &b.Deployments, &b.Executions30d); err != nil {
			return nil, err
		}
		// Privacy: hide profile fields unless the user has a public profile
		// marker. The schema doesn't yet have a `public_profile` flag, so we
		// keep names visible by default and require a future flag to gate
		// them. Display name is never the email.
		b.ProfilePublic = true
		if b.DisplayName == "" {
			b.DisplayName = "Anonymous builder"
		}
		signals := Signals{
			ActiveUsers:    1,
			Deployments:    b.Deployments,
			Executions30d:  int(b.Executions30d),
			NewUsers30d:    0,
		}
		s := Compute(signals, 0, weights)
		b.ScoreComposite = s.Raw
		out = append(out, b)
	}
	for i := range out {
		out[i].Rank = i + 1
	}
	return out, nil
}

// FindSeedCSV tries a few likely locations relative to the working directory
// and returns the first one that exists.
func FindSeedCSV() (string, bool) {
	candidates := []string{
		"data/cities_seed.csv",
		"../data/cities_seed.csv",
		"../../data/cities_seed.csv",
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, true
		}
	}
	return "", false
}

// UniversityInMetro represents a university ranking entry in a specific metro.
type UniversityInMetro struct {
	UniversityID   int64   `json:"university_id"`
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	ShortName     string  `json:"short_name,omitempty"`
	CountryCode   string  `json:"country_code"`
	StateCode     string  `json:"state_code,omitempty"`
	Rank          int     `json:"rank"`
	ScorePerCapita float64 `json:"score_per_capita"`
	ActiveUsers   int     `json:"active_users"`
	StudentCount  int     `json:"student_count,omitempty"`
}

// ListUniversitiesByMetro returns universities in a metro area, ordered by per-capita rank.
func (r *Repository) ListUniversitiesByMetro(ctx context.Context, metroSlug string, limit int) ([]UniversityInMetro, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id,
			u.slug,
			u.name,
			COALESCE(u.short_name, ''),
			u.country_code,
			COALESCE(u.state_code, ''),
			r.rank_position,
			r.score_per_capita,
			r.active_users,
			COALESCE(u.student_count, 0)
		FROM university_rankings r
		JOIN universities u ON u.id = r.university_id
		WHERE r.ranking_category = 'composite'
			AND r.active_users >= $1
			AND u.is_active = TRUE
			AND u.city_id IN (SELECT id FROM cities WHERE metro_area_id = (SELECT id FROM metro_areas WHERE slug = $2))
		ORDER BY r.score_per_capita DESC
		LIMIT $3
	`, MinActiveUsersForPublic, metroSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []UniversityInMetro
	for rows.Next() {
		var u UniversityInMetro
		if err := rows.Scan(&u.UniversityID, &u.Slug, &u.Name, &u.ShortName,
			&u.CountryCode, &u.StateCode, &u.Rank,
			&u.ScorePerCapita, &u.ActiveUsers, &u.StudentCount); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// ── Auto-creation for unknown cities ──────────────────────────────────────────

// GetOrCreateMetroByState returns an existing metro for (state, country) or creates a placeholder.
// Used when a user picks a city that isn't in our seed data yet.
func (r *Repository) GetOrCreateMetroByState(ctx context.Context, stateCode, countryCode string) (*MetroArea, error) {
	// Try existing metro first — match by state_code in the metro name (e.g. "Texas" in "Dallas-Fort Worth, Texas").
	search := "%" + stateCode + "%"
	var m MetroArea
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, country_code, population, latitude, longitude, is_active, created_at
		FROM metro_areas
		WHERE country_code = $1 AND name ILIKE $2 AND is_active = TRUE
		LIMIT 1
	`, countryCode, search).Scan(&m.ID, &m.Slug, &m.Name, &m.CountryCode, &m.Population, &m.Latitude, &m.Longitude, &m.IsActive, &m.CreatedAt)
	if err == nil {
		return &m, nil
	}
	if !errors.Is(err, sql.ErrNoRows) && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("lookup metro: %w", err)
	}

	// No existing metro — create a placeholder for this state.
	slug := strings.ToLower(countryCode) + "-" + strings.ToLower(stateCode)
	name := stateCode + ", " + countryCode
	lat, lon, centroidOk := GetUSStateCentroid(stateCode)
	if !centroidOk {
		lat, lon = 39.8283, -98.5795 // geographic center of US
	}
	pop := 5_000_000

	_, err = r.pool.Exec(ctx, `
		INSERT INTO metro_areas (slug, name, country_code, population, latitude, longitude, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			is_active = TRUE
	`, slug, name, countryCode, pop, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("insert metro: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		SELECT id, slug, name, country_code, population, latitude, longitude, is_active, created_at
		FROM metro_areas WHERE slug = $1
	`, slug).Scan(&m.ID, &m.Slug, &m.Name, &m.CountryCode, &m.Population, &m.Latitude, &m.Longitude, &m.IsActive, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("re-fetch metro: %w", err)
	}
	return &m, nil
}

// CityCreateInput describes a city to upsert.
type CityCreateInput struct {
	Name        string
	Slug       string
	StateCode   string
	StateName   string
	CountryCode string
	CountryName string
	Latitude    float64
	Longitude   float64
	Population  int
	MetroID     *int64
}

// UpsertCity creates or updates a city. Returns the city ID.
func (r *Repository) UpsertCity(ctx context.Context, c CityCreateInput) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name,
			latitude, longitude, population, metro_area_id, is_active,
			review_status, auto_review_pop_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE,
			(CASE WHEN $9::int >= COALESCE((SELECT current_setting('app.city_pop_threshold', true)::int, 100000))
				THEN 'pending' ELSE 'approved' END),
			COALESCE((SELECT current_setting('app.city_pop_threshold', true)::int, 100000), 100000))
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			state_code = EXCLUDED.state_code,
			state_name = EXCLUDED.state_name,
			country_name = EXCLUDED.country_name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			population = EXCLUDED.population,
			metro_area_id = EXCLUDED.metro_area_id,
			is_active = TRUE,
			review_status = CASE
				WHEN cities.review_status IN ('approved', 'seed') THEN cities.review_status
				ELSE (CASE WHEN EXCLUDED.population >= COALESCE((SELECT current_setting('app.city_pop_threshold', true)::int, 100000), 100000)
					THEN 'pending' ELSE 'approved' END)
			END
		RETURNING id
	`, c.Slug, c.Name, c.StateCode, c.StateName, c.CountryCode, c.CountryName,
		c.Latitude, c.Longitude, c.Population, c.MetroID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert city: %w", err)
	}
	return id, nil
}

// UpsertCityForReview creates or updates a city with explicit review status.
// reviewStatus should be 'approved', 'pending', or 'seed'.
// When forceStatus is true, the reviewStatus is always applied (for geocoded cities).
// Returns the city ID and the review status that was set.
func (r *Repository) UpsertCityForReview(ctx context.Context, c CityCreateInput, reviewStatus CityReviewStatus, forceStatus ...bool) (int64, CityReviewStatus, error) {
	var id int64
	var status CityReviewStatus
	force := len(forceStatus) > 0 && forceStatus[0]

	var reviewStatusClause string
	if force {
		// For geocoded cities, always use the passed-in status
		reviewStatusClause = "$11"
	} else {
		// Preserve existing approved/seed status, otherwise use passed-in status
		reviewStatusClause = `CASE
			WHEN cities.review_status IN ('approved', 'seed') THEN cities.review_status
			ELSE $11
		END`
	}

	query := fmt.Sprintf(`
		INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name,
			latitude, longitude, population, metro_area_id, is_active, review_status, auto_review_pop_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, $11, $12)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			state_code = EXCLUDED.state_code,
			state_name = EXCLUDED.state_name,
			country_name = EXCLUDED.country_name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			population = EXCLUDED.population,
			metro_area_id = EXCLUDED.metro_area_id,
			is_active = TRUE,
			review_status = %s
		RETURNING id, review_status
	`, reviewStatusClause)

	err := r.pool.QueryRow(ctx, query,
		c.Slug, c.Name, c.StateCode, c.StateName, c.CountryCode, c.CountryName,
		c.Latitude, c.Longitude, c.Population, c.MetroID, reviewStatus, DefaultAutoReviewPopulationThreshold).Scan(&id, &status)
	if err != nil {
		return 0, "", fmt.Errorf("upsert city for review: %w", err)
	}
	return id, status, nil
}

// CityReviewSummary is one city pending review.
type CityReviewSummary struct {
	CityID      int64     `json:"city_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	StateCode   string    `json:"state_code"`
	StateName   string    `json:"state_name"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Population  int       `json:"population"`
	MetroName   string    `json:"metro_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	AliasCount  int       `json:"alias_count"`
}

// ListCitiesPendingReview returns all cities awaiting admin review.
func (r *Repository) ListCitiesPendingReview(ctx context.Context, limit, offset int) ([]CityReviewSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name,
			c.country_code, c.country_name, c.latitude, c.longitude,
			COALESCE(c.population, 0), COALESCE(m.name, ''), c.created_at,
			(SELECT COUNT(*) FROM city_aliases WHERE city_id = c.id) AS alias_count
		FROM cities c
		LEFT JOIN metro_areas m ON m.id = c.metro_area_id
		WHERE c.review_status = 'pending' AND c.is_active = TRUE
		ORDER BY c.population DESC NULLS LAST, c.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CityReviewSummary
	for rows.Next() {
		var s CityReviewSummary
		if err := rows.Scan(&s.CityID, &s.Slug, &s.Name, &s.StateCode, &s.StateName,
			&s.CountryCode, &s.CountryName, &s.Latitude, &s.Longitude,
			&s.Population, &s.MetroName, &s.CreatedAt, &s.AliasCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CountCitiesPendingReview returns the total count of cities awaiting review.
func (r *Repository) CountCitiesPendingReview(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM cities WHERE review_status = 'pending' AND is_active = TRUE
	`).Scan(&count)
	return count, err
}

// CityAdminListEntry is a city for the admin list view.
type CityAdminListEntry struct {
	CityID      int64     `json:"city_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	StateCode   string    `json:"state_code"`
	StateName   string    `json:"state_name"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Population  int       `json:"population"`
	ReviewStatus string   `json:"review_status"`
	MetroName   string    `json:"metro_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	AliasCount  int       `json:"alias_count"`
}

// ListAllCitiesForAdmin returns all active cities for the admin UI.
func (r *Repository) ListAllCitiesForAdmin(ctx context.Context, limit, offset int) ([]CityAdminListEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name,
			c.country_code, c.country_name, c.latitude, c.longitude,
			COALESCE(c.population, 0), c.review_status, COALESCE(m.name, ''), c.created_at,
			(SELECT COUNT(*) FROM city_aliases WHERE city_id = c.id) AS alias_count
		FROM cities c
		LEFT JOIN metro_areas m ON m.id = c.metro_area_id
		WHERE c.is_active = TRUE
		ORDER BY c.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CityAdminListEntry
	for rows.Next() {
		var s CityAdminListEntry
		if err := rows.Scan(&s.CityID, &s.Slug, &s.Name, &s.StateCode, &s.StateName,
			&s.CountryCode, &s.CountryName, &s.Latitude, &s.Longitude,
			&s.Population, &s.ReviewStatus, &s.MetroName, &s.CreatedAt, &s.AliasCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ReviewCity updates a city's review status. reviewerID is the admin's user UUID.
func (r *Repository) ReviewCity(ctx context.Context, cityID int64, reviewerID string, status CityReviewStatus, notes string) error {
	if status != CityReviewStatusApproved && status != CityReviewStatusRejected {
		return fmt.Errorf("invalid review status: %s", status)
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE cities
		SET review_status = $2,
			reviewed_at = NOW(),
			reviewed_by = $3::uuid,
			review_notes = $4
		WHERE id = $1 AND review_status = 'pending'
	`, cityID, status, reviewerID, notes)
	return err
}

// GetCityForReview returns full city details for admin review.
func (r *Repository) GetCityForReview(ctx context.Context, cityID int64) (*City, error) {
	var c City
	var reviewedBy sql.NullString
	var reviewNotes sql.NullString
	var reviewedAt sql.NullTime
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.slug, c.name, c.state_code, c.state_name,
			c.country_code, c.country_name, c.latitude, c.longitude,
			COALESCE(c.population, 0), c.review_status, c.reviewed_at,
			c.reviewed_by, c.review_notes,
			c.metro_area_id, c.created_at
		FROM cities c
		WHERE c.id = $1
	`, cityID).Scan(&c.ID, &c.Slug, &c.Name, &c.StateCode, &c.StateName,
		&c.CountryCode, &c.CountryName, &c.Latitude, &c.Longitude,
		&c.Population, &c.ReviewStatus, &reviewedAt, &reviewedBy,
		&reviewNotes, &c.MetroAreaID, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		c.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		c.ReviewedAt = &reviewedAt.Time
	}
	if reviewNotes.Valid {
		c.ReviewNotes = reviewNotes.String
	}
	return &c, nil
}

// CityAliasDetail is an alias with its source for admin review.
type CityAliasDetail struct {
	Alias   string    `json:"alias"`
	Source  string    `json:"source"`
	Created time.Time `json:"created"`
}

// ListCityAliases returns all aliases for a city (admin view).
func (r *Repository) ListCityAliases(ctx context.Context, cityID int64) ([]CityAliasDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT alias, source, NOW()
		FROM city_aliases
		WHERE city_id = $1
		ORDER BY source, alias
	`, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CityAliasDetail
	for rows.Next() {
		var a CityAliasDetail
		if err := rows.Scan(&a.Alias, &a.Source, &a.Created); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AddCityAlias creates an alias for a city. Used for user-typed locations.
func (r *Repository) AddCityAlias(ctx context.Context, cityID int64, alias, source string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO city_aliases (city_id, alias, source)
		VALUES ($1, $2, $3)
		ON CONFLICT (alias, source) DO NOTHING
	`, cityID, alias, source)
	return err
}

// Slugify produces a URL-safe slug from a city name.
func SlugifyCity(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}
