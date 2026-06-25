package storage

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// FWOSDatabase holds the separate FWOS database connection.
type FWOSDatabase struct {
	*sql.DB
}

// loadFWOSDatabaseConfig loads FWOS-specific database config from environment.
// Falls back to main DB config with "fwos" as database name.
func loadFWOSDatabaseConfig() (*DatabaseConfig, error) {
	// Check for dedicated FWOS database URL first
	if fwosURL := os.Getenv("FWOS_DATABASE_URL"); fwosURL != "" {
		config := &DatabaseConfig{
			ConnectionString: fwosURL,
		}
		return config, nil
	}

	// Otherwise derive from main DB config, overriding the database name
	config, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load base DB config for FWOS: %w", err)
	}

	// Override database name
	fwosDBName := os.Getenv("FWOS_DB_NAME")
	if fwosDBName == "" {
		fwosDBName = "fwos"
	}
	config.Database = fwosDBName
	config.ConnectionString = "" // Clear any connection string to use individual params

	return config, nil
}

// NewFWOSDatabase creates a new connection to the FWOS database.
func NewFWOSDatabase() (*FWOSDatabase, error) {
	config, err := loadFWOSDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load FWOS database config: %w", err)
	}

	connStr := buildConnectionString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open FWOS database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping FWOS database: %w", err)
	}

	// Use smaller pool for FWOS (less traffic than main DB)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	logrus.WithField("database", config.Database).Info("Connected to FWOS database")

	return &FWOSDatabase{DB: db}, nil
}
