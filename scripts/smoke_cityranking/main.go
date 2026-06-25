// Smoke test for the city-ranking recompute cycle. Inserts a handful of fake
// users across 3 metros, runs one recompute cycle in-process, and prints the
// resulting leaderboard. Run with: go run ./scripts/smoke_cityranking
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	crjob "github.com/functionfly/functionfly/internal/jobs/cityranking"
	crstorage "github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	repo := crstorage.NewRepository(pool, logrus.StandardLogger())
	cache := crstorage.NewCache(nil) // no-op cache

	// Drop any leftover user/city rows from prior runs so the leaderboard is
	// deterministic.
	_, _ = pool.Exec(ctx, `UPDATE users SET city_id = NULL`)
	_, _ = pool.Exec(ctx, `DELETE FROM city_rankings`)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'smoke-%@example.com'`)

	// Pick 3 metros: Austin, NYC, SF.
	metros := []string{
		"austin-round-rock-georgetown-tx",
		"new-york-newark-jersey-city-ny-nj-pa",
		"san-francisco-oakland-hayward-ca",
	}
	metroID := map[string]int64{}
	for _, slug := range metros {
		var id int64
		if err := pool.QueryRow(ctx, `SELECT id FROM metro_areas WHERE slug = $1`, slug).Scan(&id); err != nil {
			log.Fatalf("metro %s not found: %v", slug, err)
		}
		metroID[slug] = id
	}

	// Insert 30 fake users across the 3 metros.
	tenantRow := pool.QueryRow(ctx, `SELECT id FROM tenants LIMIT 1`)
	var tenantID string
	if err := tenantRow.Scan(&tenantID); err != nil {
		log.Fatalf("no tenant: %v", err)
	}
	for i := 0; i < 30; i++ {
		slug := metros[i%len(metros)]
		uid := fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1)
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, email, name, password_hash, token_version, profile_number, tenant_id, city_id, city_ranking_opted_out, last_active_at, created_at)
			VALUES ($1, $2, $3, '!smoke-test-no-password', 0, $7, $4,
				(SELECT id FROM cities WHERE metro_area_id = $5 AND is_active LIMIT 1),
				FALSE, NOW() - ($6 || ' days')::interval, NOW() - ($6 || ' days')::interval)
			ON CONFLICT (id) DO UPDATE SET
				city_id = EXCLUDED.city_id,
				city_ranking_opted_out = FALSE,
				last_active_at = EXCLUDED.last_active_at,
				created_at = EXCLUDED.created_at
		`, uid, fmt.Sprintf("smoke-%d@example.com", i), fmt.Sprintf("User %d", i),
			tenantID, metroID[slug], fmt.Sprintf("%d", i%10), 100000+i+1)
		if err != nil {
			log.Fatalf("insert user %d: %v", i, err)
		}
	}
	log.Printf("inserted 30 smoke-test users")

	job := crjob.NewCityRankingRecompute(repo, cache, time.Now)
	// RunOnce blocks until the cycle completes — needed because the script
	// closes the pool on return.
	job.RunOnce(ctx, 4)

	rows, err := repo.ListRankings(ctx, 10, "", crstorage.CategoryComposite)
	if err != nil {
		log.Fatalf("list rankings: %v", err)
	}
	if len(rows) == 0 {
		log.Fatal("no rankings produced — something is wrong")
	}
	log.Printf("got %d ranked rows (privacy threshold = %d active users)",
		len(rows), crstorage.MinActiveUsersForPublic)
	for i, r := range rows {
		fmt.Printf("#%d  %-40s  raw=%.3f  per_capita=%.6f  users=%d  deploys=%d  execs=%d\n",
			i+1, r.MetroName, r.ScoreRaw, r.ScorePerCapita, r.ActiveUsers, r.Deployments, r.Executions30d)
	}
}
