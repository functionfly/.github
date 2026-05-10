package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/functionfly/functionfly/internal/autoreadme"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var (
		batchSize     = flag.Int("batch-size", 100, "Number of functions to process per batch")
		force         = flag.Bool("force", false, "Overwrite existing readmes")
		envFile       = flag.String("env", ".env", "Path to .env file")
		projectReadme = flag.Bool("project-readme", false, "Also generate project README")
		repoRoot      = flag.String("repo-root", ".", "Path to repository root")
	)
	flag.Parse()

	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("Warning: could not load %s: %v", *envFile, err)
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "functionfly"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	repo := registry.NewRegistryRepository(db, nil)
	gen := autoreadme.NewService(repo, os.Getenv("API_BASE_URL"))

	ctx := context.Background()

	fmt.Printf("Starting README backfill (batch_size=%d, force=%v)...\n", *batchSize, *force)

	total, err := gen.BackfillAll(ctx, *batchSize, *force)
	if err != nil {
		log.Fatalf("Backfill failed: %v", err)
	}

	fmt.Printf("Backfill complete: %d version(s) updated\n", total)

	if *projectReadme {
		fmt.Printf("\nGenerating project README...\n")
		readme := autoreadme.GenerateProjectReadmeFromDir(*repoRoot)
		readmePath := filepath.Join(*repoRoot, "README.md")
		if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
			log.Printf("Warning: could not write project README: %v", err)
		} else {
			fmt.Printf("Project README generated: %s\n", readmePath)
		}
	}
}
