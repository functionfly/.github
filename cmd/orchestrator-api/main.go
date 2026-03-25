package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	skipMigrations := flag.Bool("skip-migrations", false, "Skip database migrations")
	flag.Parse()

	// Prefer PORT from environment (e.g. Fly.io, 12-factor)
	if p := os.Getenv("PORT"); p != "" {
		var portVal int
		if n, err := fmt.Sscanf(p, "%d", &portVal); n == 1 && err == nil && portVal > 0 {
			*port = portVal
		}
	}
	if *port <= 0 {
		*port = 8080
	}

	// Configure logrus for structured JSON logging
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Check if running in development mode
	devMode := os.Getenv("DEVELOPMENT") == "true"
	if devMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("Running in development mode")
	}

	logrus.Info("Starting FunctionFly Orchestrator API")

	// Initialize database connection
	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations if not skipped
	if !*skipMigrations {
		logrus.Info("Running database migrations...")
		if err := storage.RunMigrations(db); err != nil {
			logrus.WithError(err).Fatal("Failed to run migrations")
		}
		logrus.Info("Database migrations completed")
	} else {
		logrus.Info("Skipping database migrations")
	}

	// Create the API server
	server := api.NewServer(db)

	// ListenAndServe registers SIGINT/SIGTERM and runs graceful shutdown internally.
	// Do not also handle signals in main — duplicate Shutdown caused panic (close of closed channel).
	done := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", *port)
		logrus.WithField("addr", addr).Info("Starting HTTP server")
		done <- server.ListenAndServe(addr)
	}()

	if err := <-done; err != nil {
		log.Fatalf("Server error: %v", err)
	}
	logrus.Info("Server stopped")
}
