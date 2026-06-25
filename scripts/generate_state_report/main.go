// Ad-hoc CLI to generate one "State of AI Builders" report. Useful for
// staging verification and for back-filling missed months.
//
// Usage:
//   go run ./scripts/generate_state_report                     # current month
//   go run ./scripts/generate_state_report --month 2026-06     # specific month
//   go run ./scripts/generate_state_report --out /tmp/reports  # custom output dir
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	statereportscheduler "github.com/functionfly/functionfly/internal/jobs/statereport/scheduler"
	"github.com/functionfly/functionfly/internal/storage/cityranking"
	unirankstorage "github.com/functionfly/functionfly/internal/storage/universityranking"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	out := flag.String("out", "web/site/src/content/reports", "output directory for the .md file")
	monthFlag := flag.String("month", "", "target month as YYYY-MM (default: previous month)")
	flag.Parse()

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

	city := cityranking.NewRepository(pool, nil)
	univ := unirankstorage.NewRepository(pool, nil)
	job := statereportscheduler.New(city, univ, *out, nil)

	reference := time.Now().UTC()
	if *monthFlag != "" {
		t, err := time.Parse("2006-01", *monthFlag)
		if err != nil {
			log.Fatalf("invalid --month: %v", err)
		}
		// Use the 15th of the month so the report is unambiguous.
		reference = time.Date(t.Year(), t.Month(), 15, 0, 0, 0, 0, time.UTC)
	}
	path, err := job.RunOnce(ctx, reference)
	if err != nil {
		log.Fatalf("RunOnce: %v", err)
	}
	fmt.Printf("wrote %s\n", path)
}
