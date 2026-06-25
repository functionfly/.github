// Quick CLI to trigger one city ambassador sync cycle. Useful for staging
// verification and for the smoke test to populate the table before
// hitting /v1/city-rankings/ambassadors.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cityambassadorjob "github.com/functionfly/functionfly/internal/jobs/cityambassador"
	"github.com/functionfly/functionfly/internal/storage/cityranking"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	repo := cityranking.NewRepository(pool, nil)
	job := cityambassadorjob.NewSync(repo, nil)
	res := job.RunOnce(ctx)
	fmt.Printf("ambassador sync: eligible=%d promoted=%d revoked=%d unchanged=%d duration=%s\n",
		res.MetrosEligible, res.Promoted, res.Revoked, res.Unchanged, res.Duration)
}
