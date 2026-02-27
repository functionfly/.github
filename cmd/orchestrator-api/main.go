package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api"
	"github.com/functionfly/functionfly/internal/storage"
)

func main() {
	skipMigrations := os.Getenv("SKIP_MIGRATIONS") == "true"
	if skipMigrations {
		log.Println("SKIP_MIGRATIONS=true: migrations will be skipped")
	}

	// Initialize database without prepared statements so migrations can create schema first
	db, err := storage.NewPostgresDBWithOptions(true)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations (skip when SKIP_MIGRATIONS=true, e.g. DB already up-to-date)
	if !skipMigrations {
		log.Println("Running database migrations...")
		if err := storage.RunMigrations(db); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("Skipping migrations (SKIP_MIGRATIONS=true)")
	}

	// Now that schema exists, initialize prepared statements
	stmtCtx, stmtCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stmtCancel()
	if err := db.InitPreparedStatements(stmtCtx); err != nil {
		log.Fatalf("Failed to initialize prepared statements: %v", err)
	}

	// Create API server
	server := api.NewServer(db)

	// Port from env (e.g. .env PORT=8080) so dashboard and scripts stay in sync
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Starting orchestrator API server on %s", addr)

	// Start server
	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
