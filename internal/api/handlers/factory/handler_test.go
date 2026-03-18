package factory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/discovery"
	factorysvc "github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDBCounter int64

type stubRunner struct {
	run *factorysvc.FactoryRun
	err error
}

func (s stubRunner) Run(context.Context) (*factorysvc.FactoryRun, error) { return s.run, s.err }

type stubPublisher struct {
	functions []deployment.PublishedFunction
	total     int64
	err       error
}

func (s stubPublisher) GetPublishedFunctions(context.Context, string, int, int) ([]deployment.PublishedFunction, int64, error) {
	return s.functions, s.total, s.err
}

func TestHandleRunPipelineRequiresAuthenticatedUser(t *testing.T) {
	t.Parallel()

	h := &Handler{config: &factorysvc.Config{AgentID: "factory-agent"}, runner: stubRunner{}}
	req := httptest.NewRequest(http.MethodPost, "/factory/pipeline/run", nil)
	resp := httptest.NewRecorder()

	h.HandleRunPipeline(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.JSONEq(t, `{"error":"authentication required"}`, resp.Body.String())
}

func TestHandleRunPipelineReturnsAcceptedRunAndErrorPayload(t *testing.T) {
	t.Parallel()

	run := &factorysvc.FactoryRun{AgentID: "factory-agent", Status: factorysvc.RunStatusFailed}
	h := &Handler{runner: stubRunner{run: run, err: assert.AnError}, config: &factorysvc.Config{AgentID: "factory-agent"}}
	req := httptest.NewRequest(http.MethodPost, "/factory/pipeline/run", nil)
	req = middleware.SetUserInContext(req, &auth.Claims{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000001")})
	resp := httptest.NewRecorder()

	h.HandleRunPipeline(resp, req)

	assert.Equal(t, http.StatusAccepted, resp.Code)
	assert.Contains(t, resp.Body.String(), `"error":"assert.AnError general error for testing"`, "handler should serialize pipeline errors without changing response status")
	assert.Contains(t, resp.Body.String(), `"status":"failed"`, "run payload should still be returned for partial failure visibility")
}

func TestHandleUpdateConfigAppliesSelectableFields(t *testing.T) {
	t.Parallel()

	config := &factorysvc.Config{AgentID: "factory-agent", DiscoveryBatchSize: 10, MinimumQualityScore: 70, MinimumTestScore: 80, AutoPublish: true, MaxOpportunitiesPerRun: 3, RetryAttempts: 1, RetryBackoff: time.Second, ScheduleEnabled: false, ScheduleCron: "0 0 * * *", ScheduleTimezone: "UTC"}
	h := &Handler{config: config}
	body := bytes.NewBufferString(`{"discovery_batch_size":25,"minimum_quality_score":85,"minimum_test_score":90,"require_all_tests_pass":false,"auto_publish":false,"max_opportunities_per_run":5,"retry_attempts":4,"retry_backoff_ms":250,"schedule_enabled":true,"schedule_cron":"0 12 * * *","schedule_timezone":"America/New_York"}`)
	h.applyConfigUpdate(configUpdateRequest{
		DiscoveryBatchSize:     intPtr(25),
		MinimumQualityScore:    floatPtr(85),
		MinimumTestScore:       floatPtr(90),
		RequireAllTestsPass:    boolPtr(false),
		AutoPublish:            boolPtr(false),
		MaxOpportunitiesPerRun: intPtr(5),
		RetryAttempts:          intPtr(4),
		RetryBackoffMs:         intPtr(250),
		ScheduleEnabled:        boolPtr(true),
		ScheduleCron:           strPtr("0 12 * * *"),
		ScheduleTimezone:       strPtr("America/New_York"),
	})

	assert.Equal(t, 25, config.DiscoveryBatchSize)
	assert.Equal(t, 85.0, config.MinimumQualityScore)
	assert.Equal(t, 90.0, config.MinimumTestScore)
	assert.False(t, config.RequireAllTestsPass)
	assert.False(t, config.AutoPublish)
	assert.Equal(t, 5, config.MaxOpportunitiesPerRun)
	assert.Equal(t, 4, config.RetryAttempts)
	assert.Equal(t, 250*time.Millisecond, config.RetryBackoff)
	assert.True(t, config.ScheduleEnabled)
	assert.Equal(t, "0 12 * * *", config.ScheduleCron)
	assert.Equal(t, "America/New_York", config.ScheduleTimezone)
	_ = body
}

func TestHandleListOpportunitiesFiltersAndPaginates(t *testing.T) {
	t.Parallel()

	db := newHandlerTestDB(t)
	// Use raw DDL to avoid PostgreSQL-specific type annotations in the model (text[], jsonb, decimal)
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
	records := []discovery.Opportunity{
		{Source: "reddit", SourceID: "1", Title: "match", Status: discovery.OpportunityStatusQualified},
		{Source: "reddit", SourceID: "2", Title: "skip-status", Status: discovery.OpportunityStatusRejected},
		{Source: "github", SourceID: "3", Title: "skip-source", Status: discovery.OpportunityStatusQualified},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	h := &Handler{db: db, config: &factorysvc.Config{AgentID: "factory-agent"}}
	req := httptest.NewRequest(http.MethodGet, "/factory/opportunities?status=qualified&source=reddit&limit=1&offset=0", nil)
	resp := httptest.NewRecorder()

	h.HandleListOpportunities(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var payload struct {
		Opportunities []discovery.Opportunity `json:"opportunities"`
		Total         int64                   `json:"total"`
		Limit         int                     `json:"limit"`
		Offset        int                     `json:"offset"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &payload))
	assert.Equal(t, int64(1), payload.Total, "filtered total should reflect only matching records")
	assert.Len(t, payload.Opportunities, 1)
	assert.Equal(t, "match", payload.Opportunities[0].Title)
	assert.Equal(t, 1, payload.Limit)
	assert.Equal(t, 0, payload.Offset)
}

func intPtr(v int) *int { return &v }

func floatPtr(v float64) *float64 { return &v }

func boolPtr(v bool) *bool { return &v }

func strPtr(v string) *string { return &v }

func newHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use a unique named in-memory database per test to avoid parallel test interference.
	n := atomic.AddInt64(&testDBCounter, 1)
	dsn := fmt.Sprintf("file:testdb%d?mode=memory&cache=private", n)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}
