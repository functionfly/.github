package swarm

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a PostgreSQL test database connection.
// If the database is not available, the test is skipped.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=functionfly sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping test: " + err.Error())
		return nil
	}
	return db
}
