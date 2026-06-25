// Quick CLI to trigger one university recompute cycle. Useful for staging
// verification and for the smoke test to populate the table before
// hitting /v1/university-rankings.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	universityrankingjob "github.com/functionfly/functionfly/internal/jobs/universityranking"
	unirankstorage "github.com/functionfly/functionfly/internal/storage/universityranking"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	repo := unirankstorage.NewRepository(pool, nil)
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() { _ = rdb.Close() }()
	job := universityrankingjob.NewRecompute(repo, unirankstorage.NewCache(rdb), nil)
	job.RunOnce(ctx, 8)
	fmt.Println("university recompute complete")
}
