package cityambassadorjob

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func TestSync_RunOnce_PromotesTopBuilder(t *testing.T) {
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
	repo := cityranking.NewRepository(pool, logrus.StandardLogger())
	sync := NewSync(repo, nil)

	// Hermetic: only the test metros.
	_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (
		SELECT id FROM metro_areas WHERE slug LIKE 'synctest-%'
	)`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'synctest-%@example.com'`)
	_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'synctest-%')`)
	_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'synctest-%'`)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM city_ambassadors WHERE metro_id IN (
			SELECT id FROM metro_areas WHERE slug LIKE 'synctest-%'
		)`)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'synctest-%@example.com'`)
		_, _ = pool.Exec(ctx, `DELETE FROM cities WHERE metro_area_id IN (SELECT id FROM metro_areas WHERE slug LIKE 'synctest-%')`)
		_, _ = pool.Exec(ctx, `DELETE FROM metro_areas WHERE slug LIKE 'synctest-%'`)
	}()

	// Set up: metro + city + 2 users.
	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	var metroID int64
	err = pool.QueryRow(ctx, `INSERT INTO metro_areas (slug, name, country_code, population, latitude, longitude, is_active)
		VALUES ('synctest-tokyo', 'Sync Test Tokyo', 'JP', 1000000, 35.6762, 139.6503, TRUE)
		RETURNING id`).Scan(&metroID)
	if err != nil {
		t.Fatalf("create metro: %v", err)
	}
	var cityID int64
	err = pool.QueryRow(ctx, `INSERT INTO cities (slug, name, state_code, state_name, country_code, country_name, latitude, longitude, population, metro_area_id, is_active)
		VALUES ('synctest-tokyo-city', 'Sync Test Tokyo', '13', 'Tokyo', 'JP', 'Japan', 35.6762, 139.6503, 1000000, $1, TRUE)
		RETURNING id`, metroID).Scan(&cityID)
	if err != nil {
		t.Fatalf("create city: %v", err)
	}
	insertTestUser(t, ctx, pool, tenantID, cityID, "synctest-a@example.com", 1001, false, 2)
	insertTestUser(t, ctx, pool, tenantID, cityID, "synctest-b@example.com", 1002, false, 5)
	// Add 5 deployments to user B.
	for i := 0; i < 5; i++ {
		_, err = pool.Exec(ctx, `
			INSERT INTO registry_functions (owner_user_id, name, author, created_at)
			SELECT id, $1, 'synctest', NOW() - INTERVAL '5 days' FROM users WHERE email = 'synctest-b@example.com'
		`, fmt.Sprintf("synctest-fn-%d-%d", time.Now().UnixNano(), i))
		if err != nil {
			t.Fatalf("insert fn %d: %v", i, err)
		}
	}
	// Fake a city_ranking row so ListMetrosWithActiveBuilders picks it up.
	_, err = pool.Exec(ctx, `
		INSERT INTO city_rankings (metro_area_id, rank_position, score_raw, score_per_capita,
			active_users, deployments, executions_30d, founder_earnings, new_users_30d,
			period_start, period_end, ranking_category)
		VALUES ($1, 1, 1.0, 1.0, 5, 5, 0, 0, 5, NOW() - INTERVAL '30 days', NOW(), 'composite')
	`, metroID)
	if err != nil {
		t.Fatalf("insert ranking: %v", err)
	}
	// But ListMetrosWithActiveBuilders uses MAX(period_end), so update it.
	_, err = pool.Exec(ctx, `
		UPDATE city_rankings SET period_end = (SELECT MAX(period_end) FROM city_rankings WHERE ranking_category = 'composite') + INTERVAL '1 hour'
		WHERE metro_area_id = $1
	`, metroID)
	if err != nil {
		t.Fatalf("update period_end: %v", err)
	}

	res := sync.RunOnce(ctx)
	if res.MetrosEligible < 1 {
		t.Errorf("expected at least 1 eligible metro, got %d", res.MetrosEligible)
	}
	if res.Promoted < 1 {
		t.Errorf("expected at least 1 promotion, got %d", res.Promoted)
	}

	// Verify the ambassador was created.
	amb, err := repo.GetAmbassadorForMetro(ctx, metroID)
	if err != nil {
		t.Fatalf("get ambassador: %v", err)
	}
	if amb == nil {
		t.Fatal("expected ambassador, got nil")
	}
	topEmail := readFirstString(t, ctx, pool, `SELECT email FROM users WHERE id = $1::uuid`, amb.UserID)
	if topEmail != "synctest-b@example.com" {
		t.Errorf("expected synctest-b, got %s", topEmail)
	}
	if amb.Source != "auto" {
		t.Errorf("expected source=auto, got %s", amb.Source)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func readFirstString(t *testing.T, ctx context.Context, pool *pgxpool.Pool, q string, args ...any) string {
	t.Helper()
	var s string
	if err := pool.QueryRow(ctx, q, args...).Scan(&s); err != nil {
		t.Fatalf("readFirstString: %v", err)
	}
	return s
}

func insertTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, cityID int64, email string, profileNumber int64, optedOut bool, deployments int) {
	t.Helper()
	pn := -200000 - profileNumber
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, name, password_hash, token_version, profile_number, city_id, city_ranking_opted_out, last_active_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, 'SyncTest', '!no-pw', 1, $3, $4, $5, $6, $6, $6)
		ON CONFLICT (email) DO UPDATE SET
			city_id = EXCLUDED.city_id,
			city_ranking_opted_out = EXCLUDED.city_ranking_opted_out
	`, tenantID, email, pn, cityID, optedOut, now)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
}
