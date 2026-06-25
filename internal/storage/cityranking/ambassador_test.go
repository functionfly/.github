package cityranking

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func TestRepository_Ambassador_PromoteAndList(t *testing.T) {
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
	_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (
		SELECT id FROM metro_areas WHERE slug LIKE 'ambtest-%'
	)`)
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL WHERE email LIKE 'ambtest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'ambtest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'ambtest-%')`)
	_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'ambtest-%'`)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (
			SELECT id FROM metro_areas WHERE slug LIKE 'ambtest-%'
		)`)
		_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL WHERE email LIKE 'ambtest-%@example.com'`)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'ambtest-%@example.com'`)
		_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'ambtest-%')`)
		_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'ambtest-%'`)
	}()

	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")

	// Create 1 test metro + city.
	var metroID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO metro_areas (slug, name, country_code, population, latitude, longitude, is_active)
		VALUES ('ambtest-tokyo', 'Ambassador Test Tokyo', 'JP', 1000000, 35.6762, 139.6503, TRUE)
		RETURNING id
	`).Scan(&metroID)
	if err != nil {
		t.Fatalf("create metro: %v", err)
	}
	var cityID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name, latitude, longitude, population, metro_area_id, is_active)
		VALUES ('ambtest-tokyo-city', 'Ambassador Test Tokyo', '13', 'Tokyo', 'JP', 'Japan', 35.6762, 139.6503, 1000000, $1, TRUE)
		RETURNING id
	`, metroID).Scan(&cityID)
	if err != nil {
		t.Fatalf("create city: %v", err)
	}

	// Create 3 users: 2 with activity, 1 without.
	// User A: 0 deployments, 0 executions (lowest score)
	// User B: 5 deployments, 0 executions (highest score)
	// User C: opted out (excluded)
	insertAmbTestUser(t, ctx, pool, tenantID, cityID, "ambtest-a@example.com", 1, false, 0, 0)
	insertAmbTestUser(t, ctx, pool, tenantID, cityID, "ambtest-b@example.com", 2, false, 5, 0)
	insertAmbTestUser(t, ctx, pool, tenantID, cityID, "ambtest-c@example.com", 3, true, 0, 100)

	// Add deployments for user B. Use a timestamped name to avoid clashing
	// with prior runs of the same test (uq_registry_author_name is global).
	fnName := fmt.Sprintf("ambtest-fn-%d", time.Now().UnixNano())
	_, err = pool.Exec(ctx, `
		INSERT INTO registry_functions (owner_user_id, name, author, created_at)
		SELECT id, $1, 'ambtest', NOW() - INTERVAL '5 days'
		FROM users WHERE email = 'ambtest-b@example.com'
	`, fnName)
	if err != nil {
		t.Fatalf("insert functions: %v", err)
	}

	// Now find the top builder.
	top, err := repo.TopBuilderForMetro(ctx, metroID)
	if err != nil {
		t.Fatalf("top builder: %v", err)
	}
	if top == nil {
		t.Fatal("expected a top builder, got nil")
	}
	topEmail := readFirstStringQ(t, ctx, pool, `SELECT email FROM users WHERE id = $1::uuid`, top.UserID)
	if topEmail != "ambtest-b@example.com" {
		t.Errorf("expected top builder to be ambtest-b, got %s", topEmail)
	}

	// Upsert as ambassador.
	if err := repo.UpsertAmbassador(ctx, metroID, top.UserID, "auto"); err != nil {
		t.Fatalf("upsert ambassador: %v", err)
	}

	// Verify we can fetch it back.
	amb, err := repo.GetAmbassadorForMetro(ctx, metroID)
	if err != nil {
		t.Fatalf("get ambassador: %v", err)
	}
	if amb == nil {
		t.Fatal("expected ambassador, got nil")
	}
	if amb.UserID != top.UserID {
		t.Errorf("ambassador user_id mismatch: got %s want %s", amb.UserID, top.UserID)
	}

	// Verify the list includes it.
	entries, err := repo.ListAmbassadors(ctx, "JP", 10)
	if err != nil {
		t.Fatalf("list ambassadors: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.MetroID == metroID {
			found = true
			if e.MetroSlug != "ambtest-tokyo" {
				t.Errorf("metro slug mismatch: %s", e.MetroSlug)
			}
		}
	}
	if !found {
		t.Errorf("ambassador not in list (entries=%d)", len(entries))
	}

	// Promote a different user: should revoke the old and promote the new.
	if err := repo.UpsertAmbassador(ctx, metroID, readFirstStringQ(t, ctx, pool, `SELECT id::text FROM users WHERE email = 'ambtest-a@example.com'`), "manual"); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	amb2, err := repo.GetAmbassadorForMetro(ctx, metroID)
	if err != nil {
		t.Fatalf("get after switch: %v", err)
	}
	if amb2 == nil || amb2.UserID == top.UserID {
		t.Errorf("expected new ambassador, got same as before: %+v", amb2)
	}

	// Revoke.
	if err := repo.RevokeAmbassador(ctx, metroID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	amb3, err := repo.GetAmbassadorForMetro(ctx, metroID)
	if !errors.Is(err, pgx.ErrNoRows) && err != nil {
		t.Fatalf("get after revoke: %v", err)
	}
	if amb3 != nil {
		t.Errorf("expected no ambassador after revoke, got %+v", amb3)
	}
}

// TestRepository_Ambassador_OptedOutExcluded verifies the privacy contract:
// an opted-out user is not even a candidate for ambassador, regardless of
// activity level.
func TestRepository_Ambassador_OptedOutExcluded(t *testing.T) {
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
	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")

	// Hermetic.
	_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'optouttest-%')`)
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL WHERE email LIKE 'optouttest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'optouttest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'optouttest-%')`)
	_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'optouttest-%'`)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'optouttest-%')`)
		_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL WHERE email LIKE 'optouttest-%@example.com'`)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'optouttest-%@example.com'`)
		_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'optouttest-%')`)
		_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'optouttest-%'`)
	}()

	// Create metro + city.
	var metroID int64
	err = pool.QueryRow(ctx, `INSERT INTO metro_areas (slug, name, country_code, population, latitude, longitude, is_active)
		VALUES ('optouttest-paris', 'Opt-Out Test Paris', 'FR', 1000000, 48.8566, 2.3522, TRUE)
		RETURNING id`).Scan(&metroID)
	if err != nil {
		t.Fatalf("create metro: %v", err)
	}
	var cityID int64
	err = pool.QueryRow(ctx, `INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name, latitude, longitude, population, metro_area_id, is_active)
		VALUES ('optouttest-paris-city', 'Opt-Out Test Paris', 'IDF', 'IDF', 'FR', 'France', 48.8566, 2.3522, 1000000, $1, TRUE)
		RETURNING id`, metroID).Scan(&cityID)
	if err != nil {
		t.Fatalf("create city: %v", err)
	}

	// Only opted-out user.
	insertAmbTestUser(t, ctx, pool, tenantID, cityID, "optouttest-only@example.com", 99, true, 100, 0)
	optOutFnName := fmt.Sprintf("ambtest-optout-fn-%d", time.Now().UnixNano())
	_, err = pool.Exec(ctx, `
		INSERT INTO registry_functions (owner_user_id, name, author, created_at)
		SELECT id, $1, 'ambtest', NOW() - INTERVAL '5 days' FROM users WHERE email = 'optouttest-only@example.com'
	`, optOutFnName)
	if err != nil {
		t.Fatalf("insert functions: %v", err)
	}

	top, err := repo.TopBuilderForMetro(ctx, metroID)
	if err != nil {
		t.Fatalf("top builder: %v", err)
	}
	if top != nil {
		t.Errorf("expected nil top builder (only opted-out user), got %+v", top)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func insertAmbTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, cityID int64, email string, profileNumber int64, optedOut bool, deployments, executions int) {
	t.Helper()
	pn := -100000 - profileNumber
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, name, password_hash, token_version, profile_number, city_id, city_ranking_opted_out, last_active_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, '!smoke-test-no-password', 1, $4, $5, $6, $7, $7, $7)
		ON CONFLICT (email) DO UPDATE SET
			city_id = EXCLUDED.city_id,
			city_ranking_opted_out = EXCLUDED.city_ranking_opted_out,
			last_active_at = EXCLUDED.last_active_at
	`, tenantID, email, "AmbTest User", pn, cityID, optedOut, now)
	if err != nil {
		t.Fatalf("insert amb user: %v", err)
	}
}
