package main

import (
	"log"

	"github.com/functionfly/functionfly/internal/storage"
)

func main() {
	// Initialize database
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := storage.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}