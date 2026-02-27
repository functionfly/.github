package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/functionfly/functionfly/internal/health"
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

	// Initialize health monitor service
	monitor := health.NewMonitor(db.Repository())

	// Start health monitoring
	monitor.Start()
	defer monitor.Stop()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down health monitor...")
}