package startup

import (
	"context"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

func BenchmarkColdStart(b *testing.B) {
	b.Run("PostgresConnection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			
			db, err := storage.NewPostgresDB()
			elapsed := time.Since(start)
			
			if err == nil && db != nil {
				db.Close()
			}
			
			b.ReportMetric(float64(elapsed.Milliseconds()), "ms")
		}
	})
}

func BenchmarkHealthCheckLatency(b *testing.B) {
	db, err := storage.NewPostgresDB()
	if err != nil {
		b.Skipf("Skipping: cannot connect to database: %v", err)
	}
	defer db.Close()
	
	b.Run("DatabasePing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			_ = db.PingContext(context.Background())
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms")
		}
	})
	
	b.Run("SimpleQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			var count int
			_ = db.QueryRowContext(context.Background(), "SELECT 1").Scan(&count)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms")
		}
	})
}

func BenchmarkRepositoryInitialization(b *testing.B) {
	db, err := storage.NewPostgresDB()
	if err != nil {
		b.Skipf("Skipping: cannot connect to database: %v", err)
	}
	defer db.Close()
	
	b.Run("UserRepository", func(b *testing.B) {
		repo := storage.NewUserRepository(db)
		_ = repo
	})
	
	b.Run("AppRepository", func(b *testing.B) {
		repo := storage.NewAppRepository(db)
		_ = repo
	})
}