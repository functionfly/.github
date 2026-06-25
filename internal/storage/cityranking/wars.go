package cityranking

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ── City Wars (plan §8 #8) ────────────────────────────────────────────────
//
// A "City War" is a single-elimination bracket of 8 metros, paired by their
// current leaderboard rank (1 vs 8, 2 vs 7, 3 vs 6, 4 vs 5). The war
// runs over ~3 weeks: one week per round (quarterfinal, semifinal, final).
// At the end of each round the higher per-capita score wins; the winner
// of the final gets a "Champion" badge on the metro detail page.
//
// Scoring happens in the admin path (`/v1/admin/...`) — public endpoints
// expose the bracket, scores, and champion. The k=5 privacy threshold is
// inherited from the leaderboard so only metros that pass it are
// eligible to be seeded.

// War is the public view of a single war.
type War struct {
	ID               int64      `json:"id"`
	Slug             string     `json:"slug"`
	Name             string     `json:"name"`
	Season           string     `json:"season"` // "2026-Q3"
	Status           string     `json:"status"` // "scheduled" | "active" | "complete" | "cancelled"
	Round            string     `json:"round"`  // "quarterfinal" | "semifinal" | "final" | "complete"
	StartsAt         time.Time  `json:"starts_at"`
	EndsAt           time.Time  `json:"ends_at"`
	ChampionMetroID  *int64     `json:"champion_metro_id,omitempty"`
	TotalMatches     int        `json:"total_matches"`
	TotalActiveUsers int        `json:"total_active_users"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	// Champion is filled in only when the war is complete.
	ChampionSlug  string `json:"champion_slug,omitempty"`
	ChampionName  string `json:"champion_name,omitempty"`
	ChampionCountry string `json:"champion_country,omitempty"`
	// Quarterfinal metros — denormalized so the bracket page renders
	// without joining the matches table.
	Quarterfinals []WarMatch `json:"quarterfinals,omitempty"`
	Semifinals    []WarMatch `json:"semifinals,omitempty"`
	Final         *WarMatch   `json:"final,omitempty"`
}

// WarMatch is one row in a war's bracket.
type WarMatch struct {
	ID            int64      `json:"id"`
	WarID         int64      `json:"war_id"`
	Round         string     `json:"round"`
	Position      int        `json:"position"`
	MetroAID      int64      `json:"metro_a_id"`
	MetroASlug    string     `json:"metro_a_slug"`
	MetroAName    string     `json:"metro_a_name"`
	MetroACountry string     `json:"metro_a_country"`
	MetroBID      int64      `json:"metro_b_id"`
	MetroBSlug    string     `json:"metro_b_slug"`
	MetroBName    string     `json:"metro_b_name"`
	MetroBCountry string     `json:"metro_b_country"`
	ScoreA        float64    `json:"score_a"`
	ScoreB        float64    `json:"score_b"`
	ActiveUsersA  int        `json:"active_users_a"`
	ActiveUsersB  int        `json:"active_users_b"`
	WinnerMetroID *int64     `json:"winner_metro_id,omitempty"`
	DecidedAt     *time.Time `json:"decided_at,omitempty"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────

// CreateWar inserts a new war (status='scheduled'). Use GenerateBracket
// after to insert the quarterfinal matches.
func (r *Repository) CreateWar(ctx context.Context, w *War) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO city_wars (slug, name, season, status, round, starts_at, ends_at, total_matches)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, w.Slug, w.Name, w.Season, w.Status, w.Round, w.StartsAt, w.EndsAt, w.TotalMatches).
		Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

// GetWar returns a single war with its full bracket (all rounds).
func (r *Repository) GetWar(ctx context.Context, slug string) (*War, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, season, status, round, starts_at, ends_at,
			champion_metro_id, total_matches, total_active_users, created_at, updated_at
		FROM city_wars WHERE slug = $1
	`, slug)
	var w War
	if err := row.Scan(&w.ID, &w.Slug, &w.Name, &w.Season, &w.Status, &w.Round,
		&w.StartsAt, &w.EndsAt, &w.ChampionMetroID, &w.TotalMatches,
		&w.TotalActiveUsers, &w.CreatedAt, &w.UpdatedAt); err != nil {
		if errors.Is(err, errNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if w.ChampionMetroID != nil {
		if err := r.pool.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`,
			*w.ChampionMetroID).Scan(&w.ChampionSlug, &w.ChampionName, &w.ChampionCountry); err != nil {
			return nil, err
		}
	}
	matches, err := r.matchesForWar(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	for i := range matches {
		switch matches[i].Round {
		case "quarterfinal":
			w.Quarterfinals = append(w.Quarterfinals, matches[i])
		case "semifinal":
			w.Semifinals = append(w.Semifinals, matches[i])
		case "final":
			m := matches[i]
			w.Final = &m
		}
	}
	return &w, nil
}

// ListWars returns the most recent wars (active first, then completed).
func (r *Repository) ListWars(ctx context.Context, limit int) ([]War, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, season, status, round, starts_at, ends_at,
			champion_metro_id, total_matches, total_active_users, created_at, updated_at
		FROM city_wars
		ORDER BY
			CASE status WHEN 'active' THEN 0 WHEN 'scheduled' THEN 1 WHEN 'complete' THEN 2 ELSE 3 END,
			starts_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []War
	for rows.Next() {
		var w War
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name, &w.Season, &w.Status, &w.Round,
			&w.StartsAt, &w.EndsAt, &w.ChampionMetroID, &w.TotalMatches,
			&w.TotalActiveUsers, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GenerateBracket seeds the quarterfinal matches for a war. Pairs are
// (1 vs 8), (2 vs 7), (3 vs 6), (4 vs 5) by current per-capita rank.
// Returns the 4 matches created. Idempotent: deletes existing quarterfinal
// matches first so the function is safe to re-run after a seed correction.
func (r *Repository) GenerateBracket(ctx context.Context, warID int64) ([]WarMatch, error) {
	// Find the top 8 active metros for this war. Use the latest
	// per-capita row per metro.
	rows, err := r.pool.Query(ctx, `
		SELECT cr.metro_area_id, m.slug, m.name, m.country_code, cr.score_per_capita
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.ranking_category = 'composite'
			AND cr.active_users >= $1
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite')
		ORDER BY cr.score_per_capita DESC
		LIMIT 8
	`, MinActiveUsersForPublic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type seed struct {
		metroID   int64
		slug      string
		name      string
		country   string
		perCapita float64
	}
	var seeds []seed
	for rows.Next() {
		var s seed
		if err := rows.Scan(&s.metroID, &s.slug, &s.name, &s.country, &s.perCapita); err != nil {
			return nil, err
		}
		seeds = append(seeds, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(seeds) < 8 {
		return nil, fmt.Errorf("only %d metros above the privacy threshold; need 8 to seed a bracket", len(seeds))
	}
	// Pair 1v8, 2v7, 3v6, 4v5 by current rank.
	pairs := [][2]int{{0, 7}, {1, 6}, {2, 5}, {3, 4}}
	// Idempotent: clear any existing quarterfinal matches.
	if _, err := r.pool.Exec(ctx, `DELETE FROM city_war_matches WHERE war_id = $1 AND round = 'quarterfinal'`, warID); err != nil {
		return nil, err
	}
	out := make([]WarMatch, 0, 4)
	for i, p := range pairs {
		row := r.pool.QueryRow(ctx, `
			INSERT INTO city_war_matches
				(war_id, round, position, metro_a_id, metro_b_id, score_a, score_b,
				 active_users_a, active_users_b)
			VALUES ($1, 'quarterfinal', $2, $3, $4, $5, $6, 0, 0)
			RETURNING id
		`, warID, i+1, seeds[p[0]].metroID, seeds[p[1]].metroID,
			seeds[p[0]].perCapita, seeds[p[1]].perCapita)
		var m WarMatch
		if err := row.Scan(&m.ID); err != nil {
			return nil, err
		}
		m.Round = "quarterfinal"
		m.Position = i + 1
		m.MetroAID = seeds[p[0]].metroID
		m.MetroASlug = seeds[p[0]].slug
		m.MetroAName = seeds[p[0]].name
		m.MetroACountry = seeds[p[0]].country
		m.MetroBID = seeds[p[1]].metroID
		m.MetroBSlug = seeds[p[1]].slug
		m.MetroBName = seeds[p[1]].name
		m.MetroBCountry = seeds[p[1]].country
		m.ScoreA = seeds[p[0]].perCapita
		m.ScoreB = seeds[p[1]].perCapita
		out = append(out, m)
	}
	return out, nil
}

// RecordMatchResult closes a single match: sets the winner, the per-capita
// score, the active-user count, and the decided-at timestamp. If this is
// the final match, also sets the war's champion_metro_id and status='complete'.
func (r *Repository) RecordMatchResult(ctx context.Context, matchID int64, scoreA, scoreB float64, activeA, activeB int, winnerID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var warID int64
	var round string
	if err := tx.QueryRow(ctx, `SELECT war_id, round FROM city_war_matches WHERE id = $1 FOR UPDATE`, matchID).
		Scan(&warID, &round); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE city_war_matches
		SET winner_metro_id = $1, score_a = $2, score_b = $3,
		    active_users_a = $4, active_users_b = $5, decided_at = NOW()
		WHERE id = $6
	`, winnerID, scoreA, scoreB, activeA, activeB, matchID); err != nil {
		return err
	}
	if round == "final" {
		if _, err := tx.Exec(ctx, `
			UPDATE city_wars
			SET status = 'complete', round = 'complete', champion_metro_id = $1, updated_at = NOW()
			WHERE id = $2
		`, winnerID, warID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AdvanceWar creates the next round's matches from the winners of the
// current round. Pass round="semifinal" after quarterfinal is complete,
// round="final" after semifinal is complete. Returns the new matches.
func (r *Repository) AdvanceWar(ctx context.Context, warID int64, nextRound string) ([]WarMatch, error) {
	var prevRound string
	switch nextRound {
	case "semifinal":
		prevRound = "quarterfinal"
	case "final":
		prevRound = "semifinal"
	default:
		return nil, fmt.Errorf("unsupported next round %q", nextRound)
	}
	// Read the previous round's winners in position order.
	rows, err := r.pool.Query(ctx, `
		SELECT winner_metro_id FROM city_war_matches
		WHERE war_id = $1 AND round = $2 AND winner_metro_id IS NOT NULL
		ORDER BY position
	`, warID, prevRound)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var winnerIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		winnerIDs = append(winnerIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var expectedCount int
	if nextRound == "semifinal" {
		expectedCount = 4
	} else {
		expectedCount = 2
	}
	if len(winnerIDs) != expectedCount {
		return nil, fmt.Errorf("expected %d winners, got %d", expectedCount, len(winnerIDs))
	}
	// Idempotent: clear existing matches for the next round.
	if _, err := r.pool.Exec(ctx, `DELETE FROM city_war_matches WHERE war_id = $1 AND round = $2`, warID, nextRound); err != nil {
		return nil, err
	}
	// Build the next round's pairs: (W1 vs W2), (W3 vs W4) for
	// semifinal; (W1 vs W2) for final. Look up metro slug/name for
	// denormalization.
	out := make([]WarMatch, 0, expectedCount/2)
	for i := 0; i < expectedCount/2; i++ {
		winnerA := winnerIDs[2*i]
		winnerB := winnerIDs[2*i+1]
		var slugA, nameA, countryA, slugB, nameB, countryB string
		if err := r.pool.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`, winnerA).
			Scan(&slugA, &nameA, &countryA); err != nil {
			return nil, err
		}
		if err := r.pool.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`, winnerB).
			Scan(&slugB, &nameB, &countryB); err != nil {
			return nil, err
		}
		row := r.pool.QueryRow(ctx, `
			INSERT INTO city_war_matches
				(war_id, round, position, metro_a_id, metro_b_id, score_a, score_b,
				 active_users_a, active_users_b)
			VALUES ($1, $2, $3, $4, $5, 0, 0, 0, 0)
			RETURNING id
		`, warID, nextRound, i+1, winnerA, winnerB)
		var m WarMatch
		if err := row.Scan(&m.ID); err != nil {
			return nil, err
		}
		m.Round = nextRound
		m.Position = i + 1
		m.MetroAID = winnerA
		m.MetroASlug = slugA
		m.MetroAName = nameA
		m.MetroACountry = countryA
		m.MetroBID = winnerB
		m.MetroBSlug = slugB
		m.MetroBName = nameB
		m.MetroBCountry = countryB
		out = append(out, m)
	}
	// Move the war to the next round.
	if _, err := r.pool.Exec(ctx, `UPDATE city_wars SET round = $1, updated_at = NOW() WHERE id = $2`, nextRound, warID); err != nil {
		return nil, err
	}
	return out, nil
}

// matchesForWar returns every match for a war, in (round, position) order.
func (r *Repository) matchesForWar(ctx context.Context, warID int64) ([]WarMatch, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.war_id, m.round, m.position, m.metro_a_id, m.metro_b_id,
			ma.slug, ma.name, ma.country_code, mb.slug, mb.name, mb.country_code,
			m.score_a, m.score_b, m.active_users_a, m.active_users_b,
			m.winner_metro_id, m.decided_at
		FROM city_war_matches m
		JOIN metro_areas ma ON ma.id = m.metro_a_id
		JOIN metro_areas mb ON mb.id = m.metro_b_id
		WHERE m.war_id = $1
		ORDER BY
			CASE m.round WHEN 'quarterfinal' THEN 0 WHEN 'semifinal' THEN 1 WHEN 'final' THEN 2 ELSE 3 END,
			m.position
	`, warID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WarMatch
	for rows.Next() {
		var m WarMatch
		var winnerID *int64
		var decidedAt *time.Time
		if err := rows.Scan(&m.ID, &m.WarID, &m.Round, &m.Position,
			&m.MetroAID, &m.MetroBID,
			&m.MetroASlug, &m.MetroAName, &m.MetroACountry,
			&m.MetroBSlug, &m.MetroBName, &m.MetroBCountry,
			&m.ScoreA, &m.ScoreB, &m.ActiveUsersA, &m.ActiveUsersB,
			&winnerID, &decidedAt); err != nil {
			return nil, err
		}
		m.WinnerMetroID = winnerID
		m.DecidedAt = decidedAt
		out = append(out, m)
	}
	return out, rows.Err()
}

// errNoRows is exported for the ambassador handler too; keep one place
// that owns it. We re-export the pgx sentinel here for clarity.
var errNoRows = pgxNoRows

// pgxNoRows is just an alias for pgx.ErrNoRows so this file doesn't have
// to import pgx directly.
var pgxNoRows = errors.New("no rows in result set")

// LatestWar returns the war that is currently active, or the most recently
// completed one if none is active.
func (r *Repository) LatestWar(ctx context.Context) (*War, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT slug FROM city_wars
		WHERE status IN ('active', 'scheduled')
		ORDER BY
			CASE status WHEN 'active' THEN 0 ELSE 1 END,
			starts_at DESC
		LIMIT 1
	`)
	var slug string
	if err := row.Scan(&slug); err != nil {
		if errors.Is(err, errNoRows) {
			// No active/scheduled — fall back to the most recent complete.
			err := r.pool.QueryRow(ctx, `SELECT slug FROM city_wars WHERE status = 'complete' ORDER BY ends_at DESC LIMIT 1`).Scan(&slug)
			if err != nil {
				if errors.Is(err, errNoRows) {
					return nil, nil
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return r.GetWar(ctx, slug)
}

// IsWarChampion reports whether `metroID` is the most recent war champion.
// Returns (true, slug) if so, (false, "") otherwise. The war's end date
// plus 7 days is the "spotlight" window (plan §8 #8 "homepage spotlight
// for 1 week").
func (r *Repository) IsWarChampion(ctx context.Context, metroID int64) (bool, string, error) {
	var slug string
	var endsAt time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT slug, ends_at FROM city_wars
		WHERE status = 'complete' AND champion_metro_id = $1
		ORDER BY ends_at DESC LIMIT 1
	`, metroID).Scan(&slug, &endsAt)
	if err != nil {
		if errors.Is(err, errNoRows) {
			return false, "", nil
		}
		return false, "", err
	}
	// Spotlight window: 7 days from war end.
	champion := time.Since(endsAt) <= 7*24*time.Hour
	if !champion {
		return false, "", nil
	}
	return true, slug, nil
}

// SeasonString returns the canonical season string for a reference time
// ("2026-Q3" for Jul–Sep, etc.). Used by the cron to create new wars.
func SeasonString(t time.Time) string {
	q := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%d-Q%d", t.Year(), q)
}

// ── Admin operations ───────────────────────────────────────────────────────

// AdminWar is the full war view for admin UI, including locked status.
type AdminWar struct {
	War
	ScheduledAdvanceAt *time.Time `json:"scheduled_advance_at,omitempty"`
	IsLocked           bool       `json:"is_locked"` // true when status = 'active'
}

// GetWarByID returns a single war by its numeric ID.
func (r *Repository) GetWarByID(ctx context.Context, id int64) (*War, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, season, status, round, starts_at, ends_at,
			champion_metro_id, total_matches, total_active_users, created_at, updated_at
		FROM city_wars WHERE id = $1
	`, id)
	var w War
	if err := row.Scan(&w.ID, &w.Slug, &w.Name, &w.Season, &w.Status, &w.Round,
		&w.StartsAt, &w.EndsAt, &w.ChampionMetroID, &w.TotalMatches,
		&w.TotalActiveUsers, &w.CreatedAt, &w.UpdatedAt); err != nil {
		if errors.Is(err, errNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if w.ChampionMetroID != nil {
		if err := r.pool.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`,
			*w.ChampionMetroID).Scan(&w.ChampionSlug, &w.ChampionName, &w.ChampionCountry); err != nil {
			return nil, err
		}
	}
	matches, err := r.matchesForWar(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	for i := range matches {
		switch matches[i].Round {
		case "quarterfinal":
			w.Quarterfinals = append(w.Quarterfinals, matches[i])
		case "semifinal":
			w.Semifinals = append(w.Semifinals, matches[i])
		case "final":
			m := matches[i]
			w.Final = &m
		}
	}
	return &w, nil
}

// UpdateWar updates a war's name, season, and dates. Only allowed when
// the war is in 'scheduled' status (before activation).
func (r *Repository) UpdateWar(ctx context.Context, id int64, name, season string, startsAt, endsAt time.Time) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE city_wars
		SET name = $2, season = $3, starts_at = $4, ends_at = $5, updated_at = NOW()
		WHERE id = $1 AND status = 'scheduled'
	`, id, name, season, startsAt, endsAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("war not found or no longer editable")
	}
	return nil
}

// ActivateWar transitions a war from 'scheduled' to 'active'. Bracket must
// already exist (call GenerateBracket first).
func (r *Repository) ActivateWar(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE city_wars
		SET status = 'active', updated_at = NOW()
		WHERE id = $1 AND status = 'scheduled'
	`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("war not found or already active")
	}
	return nil
}

// CancelWar transitions a war to 'cancelled'. Cannot cancel active wars.
func (r *Repository) CancelWar(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE city_wars
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status = 'scheduled'
	`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("war not found or cannot be cancelled (may already be active)")
	}
	return nil
}

// EligibleMetro is a metro that can be added to a war bracket.
type EligibleMetro struct {
	ID           int64   `json:"id"`
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	CountryCode  string  `json:"country_code"`
	RankPosition int     `json:"rank_position"`
	ScorePerCapita float64 `json:"score_per_capita"`
	ActiveUsers  int     `json:"active_users"`
}

// ListEligibleMetros returns the top metros eligible for bracket seeding.
// Used by the admin UI to let admins manually select metros.
func (r *Repository) ListEligibleMetros(ctx context.Context, limit int) ([]EligibleMetro, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT cr.metro_area_id, m.slug, m.name, m.country_code,
			cr.rank_position, cr.score_per_capita, cr.active_users
		FROM city_rankings cr
		JOIN metro_areas m ON m.id = cr.metro_area_id
		WHERE cr.ranking_category = 'composite'
			AND cr.active_users >= $1
			AND cr.period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite')
		ORDER BY cr.rank_position ASC
		LIMIT $2
	`, MinActiveUsersForPublic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EligibleMetro
	for rows.Next() {
		var m EligibleMetro
		if err := rows.Scan(&m.ID, &m.Slug, &m.Name, &m.CountryCode,
			&m.RankPosition, &m.ScorePerCapita, &m.ActiveUsers); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// SetQuarterfinalMetros replaces the quarterfinal matches for a war with
// admin-selected metros. Clears any existing quarterfinal matches first.
// Only allowed when war is in 'scheduled' status.
func (r *Repository) SetQuarterfinalMetros(ctx context.Context, warID int64, matchPairs [][2]int64) error {
	if len(matchPairs) != 4 {
		return fmt.Errorf("exactly 4 pairs required")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM city_wars WHERE id = $1`, warID).Scan(&status); err != nil {
		return err
	}
	if status != "scheduled" {
		return fmt.Errorf("can only modify scheduled wars")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM city_war_matches WHERE war_id = $1 AND round = 'quarterfinal'`, warID); err != nil {
		return err
	}

	for i, pair := range matchPairs {
		metroA, metroB := pair[0], pair[1]
		var slugA, nameA, countryA, slugB, nameB, countryB string
		var scoreA, scoreB float64
		if err := tx.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`, metroA).
			Scan(&slugA, &nameA, &countryA); err != nil {
			return fmt.Errorf("metro A %d: %w", metroA, err)
		}
		if err := tx.QueryRow(ctx, `SELECT slug, name, country_code FROM metro_areas WHERE id = $1`, metroB).
			Scan(&slugB, &nameB, &countryB); err != nil {
			return fmt.Errorf("metro B %d: %w", metroB, err)
		}
		if err := tx.QueryRow(ctx, `
			SELECT score_per_capita FROM city_rankings
			WHERE metro_area_id = $1 AND ranking_category = 'composite'
			AND period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite')
		`, metroA).Scan(&scoreA); err != nil {
			scoreA = 0
		}
		if err := tx.QueryRow(ctx, `
			SELECT score_per_capita FROM city_rankings
			WHERE metro_area_id = $1 AND ranking_category = 'composite'
			AND period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite')
		`, metroB).Scan(&scoreB); err != nil {
			scoreB = 0
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO city_war_matches
				(war_id, round, position, metro_a_id, metro_b_id, score_a, score_b, active_users_a, active_users_b)
			VALUES ($1, 'quarterfinal', $2, $3, $4, $5, $6, 0, 0)
		`, warID, i+1, metroA, metroB, scoreA, scoreB); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// OverrideMatchResult allows an admin to manually set the winner and scores
// for a match, with an audit note.
func (r *Repository) OverrideMatchResult(ctx context.Context, matchID int64, scoreA, scoreB float64, activeA, activeB int, winnerID int64, note string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var warID int64
	var round string
	if err := tx.QueryRow(ctx, `SELECT war_id, round FROM city_war_matches WHERE id = $1 FOR UPDATE`, matchID).
		Scan(&warID, &round); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE city_war_matches
		SET winner_metro_id = $1, score_a = $2, score_b = $3,
		    active_users_a = $4, active_users_b = $5, decided_at = NOW()
		WHERE id = $6
	`, winnerID, scoreA, scoreB, activeA, activeB, matchID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_log (event_type, entity_type, entity_id, metadata, created_at)
		VALUES ('city_war_match_override', 'city_war_match', $1,
			$2::jsonb, NOW())
	`, matchID, fmt.Sprintf(`{"note": %q}`, note)); err != nil {
		return err
	}
	if round == "final" {
		if _, err := tx.Exec(ctx, `
			UPDATE city_wars
			SET status = 'complete', round = 'complete', champion_metro_id = $1, updated_at = NOW()
			WHERE id = $2
		`, winnerID, warID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListWarsAdmin returns all wars for admin listing.
func (r *Repository) ListWarsAdmin(ctx context.Context) ([]War, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, season, status, round, starts_at, ends_at,
			champion_metro_id, total_matches, total_active_users, created_at, updated_at
		FROM city_wars
		ORDER BY starts_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []War
	for rows.Next() {
		var w War
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name, &w.Season, &w.Status, &w.Round,
			&w.StartsAt, &w.EndsAt, &w.ChampionMetroID, &w.TotalMatches,
			&w.TotalActiveUsers, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

var _ = strings.HasPrefix
