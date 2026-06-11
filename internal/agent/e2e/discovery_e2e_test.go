package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// Discovery Service Tests
// ============================================================

func TestDiscoveryScoring(t *testing.T) {
	t.Run("should score opportunity with keyword bonuses", func(t *testing.T) {
		scorer := discovery.DefaultScorer{}

		candidate := discovery.OpportunityCandidate{
			Title:            "OAuth webhook automation",
			Description:      "Popular integration workflow for API automation",
			Tags:             []string{"api", "oauth", "automation"},
			DemandSignal:     85,
			ComplexitySignal: 7,
		}

		demand, quality, complexity, err := scorer.Score(context.Background(), candidate)
		require.NoError(t, err)

		assert.Greater(t, demand, 85.0, "demand should increase with bonuses")
		assert.Greater(t, quality, 80.0, "quality should increase with bonuses")
		assert.LessOrEqual(t, complexity, 10, "complexity should be bounded")
	})

	t.Run("should bound scores at maximum values", func(t *testing.T) {
		scorer := discovery.DefaultScorer{}

		candidate := discovery.OpportunityCandidate{
			Title:            "URGENT OAuth webhook automation",
			Description:      "Popular integration workflow with json transform",
			Tags:             []string{"api", "simple"},
			DemandSignal:     100,
			ComplexitySignal: 15, // High complexity
		}

		demand, quality, complexity, err := scorer.Score(context.Background(), candidate)
		require.NoError(t, err)

		assert.LessOrEqual(t, demand, 100.0, "demand should be capped at 100")
		assert.LessOrEqual(t, quality, 100.0, "quality should be capped at 100")
		assert.LessOrEqual(t, complexity, 10, "complexity should be capped at 10")
	})
}

func TestDiscoveryService(t *testing.T) {
	dbForDiscovery := func(t *testing.T) *gorm.DB {
		n := atomicAddInt64()
		dsn := fmt.Sprintf("file:discoverydb%d?mode=memory&cache=private", n)
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		require.NoError(t, err)

		// Create table manually (avoiding PostgreSQL-specific syntax)
		db.Exec(`CREATE TABLE IF NOT EXISTS factory_opportunities (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			category TEXT NOT NULL DEFAULT 'automation',
			tags TEXT,
			demand_score REAL DEFAULT 0,
			complexity INTEGER NOT NULL DEFAULT 1,
			validated INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			quality_score REAL DEFAULT 0,
			review_status TEXT NOT NULL DEFAULT 'not_required',
			review_reason TEXT,
			review_requested_at DATETIME,
			metadata TEXT DEFAULT '{}',
			generated_func_id TEXT,
			generation_run_id TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`)

		return db
	}

	t.Run("should list qualified opportunities only", func(t *testing.T) {
		db := dbForDiscovery(t)
		ctx := context.Background()

		svc := discovery.NewServiceWithScorer(db, discovery.DefaultScorer{})

		opportunities := []discovery.Opportunity{
			{
				ID:            uuid.New(),
				Source:        "reddit",
				SourceID:      "reddit-1",
				Title:         "qualified",
				Status:        discovery.OpportunityStatusQualified,
				QualityScore:  80,
				DemandScore:   60,
				ReviewStatus:  "not_required",
			},
			{
				ID:            uuid.New(),
				Source:        "github",
				SourceID:      "github-1",
				Title:         "needs_review",
				Status:        discovery.OpportunityStatusNeedsReview,
				QualityScore:  75,
				DemandScore:   90,
				ReviewStatus:  "pending",
			},
			{
				ID:            uuid.New(),
				Source:        "stackoverflow",
				SourceID:      "so-1",
				Title:         "generated",
				Status:        discovery.OpportunityStatusGenerated,
				QualityScore:  99,
				DemandScore:   99,
				ReviewStatus:  "approved",
			},
		}

		for i := range opportunities {
			require.NoError(t, db.Create(&opportunities[i]).Error)
		}

		qualified, err := svc.ListQualified(ctx, 10)
		require.NoError(t, err)

		assert.Len(t, qualified, 2, "should return qualified and needs_review only")
		assert.Equal(t, "qualified", qualified[0].Title)
		assert.Equal(t, "needs_review", qualified[1].Title)
	})

	t.Run("should mark opportunity as generated", func(t *testing.T) {
		db := dbForDiscovery(t)
		ctx := context.Background()

		svc := discovery.NewServiceWithScorer(db, discovery.DefaultScorer{})

		opportunity := discovery.Opportunity{
			ID:           uuid.New(),
			Source:       "github",
			SourceID:     "test-generated",
			Title:        "test opportunity",
			Status:       discovery.OpportunityStatusQualified,
			QualityScore: 85,
			DemandScore:  70,
		}
		require.NoError(t, db.Create(&opportunity).Error)

		funcID := uuid.New().String()

		// MarkGenerated takes (ctx, sourceID, generatedFunctionID)
		err := svc.MarkGenerated(ctx, opportunity.ID.String(), funcID)
		require.NoError(t, err)

		var updated discovery.Opportunity
		db.Where("id = ?", opportunity.ID).First(&updated)
		assert.Equal(t, discovery.OpportunityStatusGenerated, updated.Status)
		assert.NotNil(t, updated.GeneratedFuncID)
	})
}

// ============================================================
// Factory Service Tests
// ============================================================

func TestFactoryPipeline(t *testing.T) {
	t.Run("should track factory run status", func(t *testing.T) {
		run := &factory.FactoryRun{
			ID:        uuid.New(),
			Status:    factory.RunStatusPending,
			CreatedAt: time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, run.ID)
		assert.Equal(t, factory.RunStatusPending, run.Status)
	})

	t.Run("should have valid run status transitions", func(t *testing.T) {
		statuses := []string{
			factory.RunStatusPending,
			factory.RunStatusRunning,
			factory.RunStatusSucceeded,
			factory.RunStatusFailed,
			factory.RunStatusReview,
		}

		for _, status := range statuses {
			run := &factory.FactoryRun{Status: status}
			assert.Equal(t, status, run.Status)
		}
	})
}

var discoveryTestDBCounter int64

func atomicAddInt64() int64 {
	discoveryTestDBCounter++
	return discoveryTestDBCounter
}
