// Command migrate-function-artifacts drains legacy function-artifact bytes
// (wasm_binary, source_code, readme, code) from Postgres into the configured
// object store. Designed to run as either a one-shot drain or as a periodic
// watcher until the cutover window closes.
//
// Usage:
//
//	migrate-function-artifacts --once
//	migrate-function-artifacts --watch --interval=1h
//	migrate-function-artifacts --dry-run --once
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/artifacts"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		envFile  = flag.String("env", ".env", "Path to .env file")
		watch    = flag.Bool("watch", false, "run periodically until interrupted")
		interval = flag.Duration("interval", time.Hour, "poll interval for --watch mode")
		dryRun   = flag.Bool("dry-run", false, "log what would be migrated, write nothing")
		batch    = flag.Int("batch-size", 100, "max rows per batch")
		progress = flag.Duration("progress-log", 30*time.Second, "log cadence in --watch mode")
	)
	flag.Parse()

	if _, err := os.Stat(*envFile); err == nil {
		if err := godotenv.Load(*envFile); err != nil {
			logrus.WithError(err).Warn("could not load .env")
		}
	}
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)

	cfg := artifacts.MigrationConfig{
		BatchSize:   *batch,
		DryRun:      *dryRun,
		ProgressLog: *progress,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, _, err := artifacts.FromEnv(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("artifacts: init store failed")
	}
	if store == nil {
		logrus.Fatal("artifacts: no store configured (set ARTIFACT_STORE=r2 or check R2 env vars)")
	}
	logrus.WithField("backend", store.Backend()).Info("artifacts: store ready")

	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("postgres: connect failed")
	}
	defer func() { _ = db.Close() }()

	m := artifacts.NewMigrator(db.GORM, store, cfg)
	if *watch {
		logrus.WithFields(logrus.Fields{
			"interval": *interval,
			"dry_run":  *dryRun,
		}).Info("watching for legacy artifacts to migrate")
		if err := m.RunWatch(ctx, *interval); err != nil {
			logrus.WithError(err).Error("watch loop exited with error")
			os.Exit(1)
		}
		logrus.Info("watch loop exiting cleanly")
		return
	}

	n, err := m.RunOnce(ctx)
	if err != nil {
		logrus.WithError(err).Error("migration batch failed")
		os.Exit(1)
	}
	logrus.WithField("rows", n).Info("migration batch complete")
}
