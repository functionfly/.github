package discovery

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var discoveryTestDBCounter int64

func TestDefaultScorerScoreAppliesKeywordBonusesAndBounds(t *testing.T) {
	t.Parallel()

	demand, quality, complexity, err := DefaultScorer{}.Score(context.Background(), OpportunityCandidate{
		Title:            "Urgent OAuth webhook automation",
		Description:      "Popular integration workflow with json transform",
		Tags:             []string{"api", "simple"},
		DemandSignal:     95,
		ComplexitySignal: 9,
	})

	require.NoError(t, err)
	assert.Equal(t, 100.0, demand, "demand should be capped at 100 after bonuses")
	assert.Equal(t, 100.0, quality, "quality should be capped at 100 after bonuses")
	assert.Equal(t, 10, complexity, "complexity should be bounded to the maximum score")
}

func TestServiceListQualifiedReturnsQualifiedAndNeedsReviewOnly(t *testing.T) {
	t.Parallel()

	db := newDiscoveryTestDB(t)
	ctx := context.Background()
	// Use raw DDL to avoid PostgreSQL-specific type annotations in the model (uuid, text[], jsonb, decimal)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS factory_opportunities (
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
	)`).Error)

	records := []Opportunity{
		{Source: "reddit", SourceID: "1", Title: "qualified", Status: OpportunityStatusQualified, QualityScore: 80, DemandScore: 60},
		{Source: "reddit", SourceID: "2", Title: "review", Status: OpportunityStatusNeedsReview, QualityScore: 75, DemandScore: 90},
		{Source: "reddit", SourceID: "3", Title: "generated", Status: OpportunityStatusGenerated, QualityScore: 99, DemandScore: 99},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	svc := NewServiceWithScorer(db, DefaultScorer{})
	items, err := svc.ListQualified(ctx, 10)

	require.NoError(t, err)
	require.Len(t, items, 2, "only qualified and needs_review opportunities should be returned")
	assert.Equal(t, "qualified", items[0].Title, "higher quality opportunities should be ordered first")
	assert.Equal(t, "review", items[1].Title, "needs_review opportunities should still be eligible")
}

func newDiscoveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use a unique named in-memory database per test to avoid parallel test interference.
	n := atomic.AddInt64(&discoveryTestDBCounter, 1)
	dsn := fmt.Sprintf("file:discoverydb%d?mode=memory&cache=private", n)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}
