package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting FunctionFly Health Monitor")

	// Initialize database connection
	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// TODO: Initialize health monitor service
	// monitor := health.NewMonitor(db)

	// Start health monitoring
	// go monitor.Start()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down health monitor...")
}