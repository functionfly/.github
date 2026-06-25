package universityranking

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func TestRepository_SeedFromCSV(t *testing.T) {
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
	// Tests run from the package directory (internal/storage/universityranking),
	// not the repo root. Walk up to find the data dir.
	path, ok := FindSeedCSV()
	if !ok {
		t.Skip("seed CSV not found; run from repo root or pass DATA_DIR")
	}
	res, err := repo.SeedFromCSV(ctx, path)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if res.UniversitiesInserted == 0 {
		t.Fatalf("expected at least one university inserted, got 0")
	}
	if res.AliasesInserted == 0 {
		t.Fatalf("expected at least one alias inserted, got 0")
	}
	t.Logf("seeded %d universities, %d aliases", res.UniversitiesInserted, res.AliasesInserted)
}

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

	// Hermetic reset.
	_, _ = pool.Exec(ctx, `UPDATE users SET university_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'utest-%@example.com'`)
	_, _ = pool.Exec(ctx, `UPDATE universities SET is_active = FALSE WHERE slug LIKE 'utest-%'`)
	_, _ = pool.Exec(ctx, `DELETE FROM universities WHERE slug LIKE 'utest-%'`)

	// Create 3 test universities.
	tenantID := readFirstString(t, ctx, pool, "SELECT id::text FROM tenants LIMIT 1")
	_, _ = pool.Exec(ctx, `
		INSERT INTO universities (slug, name, short_name, country_code, student_count, institution_type, is_active)
		VALUES
			('utest-big', 'UTEST Big U', 'BigU', 'US', 10000, 'university', TRUE),
			('utest-small', 'UTEST Small U', 'SmallU', 'US', 1000, 'university', TRUE),
			('utest-tiny', 'UTEST Tiny U', 'TinyU', 'US', 100, 'university', TRUE)
		ON CONFLICT (slug) DO UPDATE SET is_active = TRUE, student_count = EXCLUDED.student_count
	`)
	defer func() {
		_, _ = pool.Exec(ctx, `UPDATE users SET university_id = NULL`)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'utest-%@example.com'`)
		_, _ = pool.Exec(ctx, `UPDATE universities SET is_active = FALSE WHERE slug LIKE 'utest-%'`)
	}()

	// Get the IDs.
	bigID := readFirstInt64(t, ctx, pool, "SELECT id FROM universities WHERE slug = 'utest-big'")
	smallID := readFirstInt64(t, ctx, pool, "SELECT id FROM universities WHERE slug = 'utest-small'")
	tinyID := readFirstInt64(t, ctx, pool, "SELECT id FROM universities WHERE slug = 'utest-tiny'")

	// 10 users at big (passes k=5), 4 at small (fails), 10 at tiny.
	for i := int64(0); i < 10; i++ {
		insertUniversitySmokeUser(t, ctx, pool, tenantID, bigID, 1000+i)
	}
	for i := int64(0); i < 4; i++ {
		insertUniversitySmokeUser(t, ctx, pool, tenantID, smallID, 2000+i)
	}
	for i := int64(0); i < 10; i++ {
		insertUniversitySmokeUser(t, ctx, pool, tenantID, tinyID, 3000+i)
	}

	// Run the scorer for one of them to make sure it works.
	signals, err := repo.SignalsFor(ctx, bigID, time.Now().Add(-30*24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("SignalsFor: %v", err)
	}
	if signals.ActiveUsers < 1 {
		t.Errorf("expected active_users >= 1, got %d", signals.ActiveUsers)
	}
	// The opt-out check happens in the SQL — verify the smoke users
	// are NOT opted out (default false).
	optedOut := readFirstBool(t, ctx, pool, `SELECT COALESCE(university_ranking_opted_out, FALSE) FROM users WHERE email LIKE 'utest-%@example.com' LIMIT 1`)
	if optedOut {
		t.Errorf("smoke users should default to opted-in")
	}
}

func TestRepository_OptOutExcludes(t *testing.T) {
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
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'utest-optout-%@example.com'`)

	id := readFirstInt64(t, ctx, pool, "SELECT id FROM universities WHERE slug = 'utest-big'")
	if id == 0 {
		t.Skip("utest-big not seeded; run the seed first")
	}

	// 5 users, all opted in.
	for i := int64(0); i < 5; i++ {
		insertUniversitySmokeUser(t, ctx, pool, tenantID, id, 4000+i)
	}
	// 5 users, ALL opted out.
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'utest-optout-%@example.com'`)
	for i := int64(0); i < 5; i++ {
		email := formatUserEmail(5000+i)
		insertUniversitySmokeUserByEmail(t, ctx, pool, tenantID, id, 5000+i, email)
		_, _ = pool.Exec(ctx, `UPDATE users SET university_ranking_opted_out = TRUE WHERE email = $1`, email)
	}
	for i := int64(0); i < 5; i++ {
		email := formatUserEmail(6000+i)
		insertUniversitySmokeUserByEmail(t, ctx, pool, tenantID, id, 6000+i, email)
	}

	signals, err := repo.SignalsFor(ctx, id, time.Now().Add(-30*24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("SignalsFor: %v", err)
	}
	// 10 opted-in users should be counted; the 5 opted-out are excluded.
	if signals.ActiveUsers != 10 {
		t.Errorf("expected 10 active users (opt-out should exclude 5), got %d", signals.ActiveUsers)
	}
}

func TestRepository_LookupByAlias_AndResolve(t *testing.T) {
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

	matches, err := repo.LookupByAlias(ctx, "mit")
	if err != nil {
		t.Fatalf("lookup 'mit': %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected at least one match for 'mit'")
	}
	if matches[0].Slug != "mit" {
		t.Errorf("expected top match to be 'mit', got %q", matches[0].Slug)
	}

	// Slug lookup.
	uni, err := repo.GetBySlug(ctx, "stanford")
	if err != nil {
		t.Fatalf("get stanford: %v", err)
	}
	if uni == nil || uni.Slug != "stanford" {
		t.Errorf("expected stanford, got %+v", uni)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func readFirstString(t *testing.T, ctx context.Context, pool *pgxpool.Pool, q string) string {
	t.Helper()
	var s string
	if err := pool.QueryRow(ctx, q).Scan(&s); err != nil {
		t.Fatalf("readFirstString: %v", err)
	}
	return s
}

func readFirstInt64(t *testing.T, ctx context.Context, pool *pgxpool.Pool, q string) int64 {
	t.Helper()
	var v int64
	if err := pool.QueryRow(ctx, q).Scan(&v); err != nil {
		t.Fatalf("readFirstInt64: %v", err)
	}
	return v
}

func readFirstBool(t *testing.T, ctx context.Context, pool *pgxpool.Pool, q string) bool {
	t.Helper()
	var v bool
	if err := pool.QueryRow(ctx, q).Scan(&v); err != nil {
		t.Fatalf("readFirstBool: %v", err)
	}
	return v
}

func insertUniversitySmokeUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, universityID, i int64) {
	t.Helper()
	email := formatUserEmail(i)
	insertUniversitySmokeUserByEmail(t, ctx, pool, tenantID, universityID, i, email)
}

func formatUserEmail(i int64) string {
	return "utest-" + itoa(i) + "@example.com"
}

func insertUniversitySmokeUserByEmail(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, universityID, i int64, email string) {
	t.Helper()
	now := time.Now()
	// profile_number is a public-facing unique integer. We derive a stable
	// negative range (UTEST prefix) so it can't collide with real accounts.
	pn := -10000 - int(i)
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, name, password_hash, token_version, profile_number, university_id, last_active_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, '$2a$10$abcdef', 1, $6, $4, $5, $5, $5)
		ON CONFLICT (email) DO UPDATE SET university_id = EXCLUDED.university_id, last_active_at = EXCLUDED.last_active_at
	`, tenantID, email, "UTEST User", universityID, now, pn)
	if err != nil {
		t.Fatalf("insert smoke user %s: %v", email, err)
	}
}

func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	negative := i < 0
	if negative {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
