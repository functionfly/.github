package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Start server in a goroutine
	done := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		logrus.WithField("addr", addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(addr); err != nil {
			done <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-done:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case <-quit:
		logrus.Info("Shutting down server...")
		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("Server forced to shutdown")
		}
	}

	logrus.Info("Server stopped")
}
