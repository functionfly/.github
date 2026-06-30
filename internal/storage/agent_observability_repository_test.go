package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AgentObservabilityRepositoryTestSuite struct {
	PostgresTestSuite
}

func TestAgentObservabilityRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AgentObservabilityRepositoryTestSuite))
}

func (s *AgentObservabilityRepositoryTestSuite) SetupTest() {
	s.PostgresTestSuite.SetupTest()
	s.db.DB.Exec("TRUNCATE TABLE agent_observability_runs CASCADE")
	s.db.DB.Exec("TRUNCATE TABLE agent_observability_config CASCADE")
}

func (s *AgentObservabilityRepositoryTestSuite) TearDownTest() {
	s.db.DB.Exec("TRUNCATE TABLE agent_observability_runs CASCADE")
	s.db.DB.Exec("TRUNCATE TABLE agent_observability_config CASCADE")
}

func (s *AgentObservabilityRepositoryTestSuite) TestCreateRun() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	run := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "atlas-run-123",
		AgentID:       "agent-001",
		AgentType:     "flymind",
		Status:        "running",
		Metadata:      JSONMap{"key": "value"},
	}

	err := s.db.GORM.Create(run).Error
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, run.ID)
	assert.NotEqual(s.T(), time.Time{}, run.CreatedAt)
	assert.NotEqual(s.T(), time.Time{}, run.StartedAt)
}

func (s *AgentObservabilityRepositoryTestSuite) TestGetRun() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	run := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "atlas-run-456",
		AgentID:       "agent-002",
		AgentType:     "agent",
		Status:        "running",
	}
	err := s.db.GORM.Create(run).Error
	require.NoError(s.T(), err)

	retrieved := &ObservabilityRun{}
	err = s.db.GORM.Where("id = ? AND tenant_id = ?", run.ID, tenantID).First(retrieved).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), run.ID, retrieved.ID)
	assert.Equal(s.T(), "atlas-run-456", retrieved.AtlasRunID)
}

func (s *AgentObservabilityRepositoryTestSuite) TestListRuns() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	for i := 0; i < 5; i++ {
		run := &ObservabilityRun{
			ID:            uuid.New(),
			TenantID:      tenantID,
			AtlasTenantID: atlasTenantID,
			AtlasRunID:    uuid.New().String(),
			AgentID:       "agent-list",
			AgentType:     "workflow",
			Status:        "running",
		}
		err := s.db.GORM.Create(run).Error
		require.NoError(s.T(), err)
	}

	var runs []*ObservabilityRun
	err := s.db.GORM.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&runs).Error
	require.NoError(s.T(), err)
	assert.Len(s.T(), runs, 5)
}

func (s *AgentObservabilityRepositoryTestSuite) TestUpdateRunStats() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	run := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "atlas-run-stats",
		AgentID:       "agent-stats",
		AgentType:     "flymind",
		Status:        "running",
		EventCount:    0,
	}
	err := s.db.GORM.Create(run).Error
	require.NoError(s.T(), err)

	stats := &RunStats{
		TotalCostUSD:  1.234,
		InputTokens:   100,
		OutputTokens:  50,
		EventCount:    10,
		ErrorCount:    1,
		ToolCallCount: 5,
	}

	err = s.db.GORM.Model(run).Updates(map[string]interface{}{
		"total_cost_usd":      stats.TotalCostUSD,
		"total_input_tokens":  stats.InputTokens,
		"total_output_tokens": stats.OutputTokens,
		"event_count":         stats.EventCount,
		"error_count":         stats.ErrorCount,
		"tool_call_count":     stats.ToolCallCount,
	}).Error
	require.NoError(s.T(), err)

	var updated ObservabilityRun
	err = s.db.GORM.Where("id = ?", run.ID).First(&updated).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1.234, updated.TotalCostUSD)
	assert.Equal(s.T(), 100, updated.TotalInputTokens)
	assert.Equal(s.T(), 10, updated.EventCount)
}

func (s *AgentObservabilityRepositoryTestSuite) TestEndRun() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	run := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "atlas-run-end",
		AgentID:       "agent-end",
		AgentType:     "team",
		Status:        "running",
	}
	err := s.db.GORM.Create(run).Error
	require.NoError(s.T(), err)

	now := time.Now()
	err = s.db.GORM.Model(run).Updates(map[string]interface{}{
		"status":   "completed",
		"ended_at": &now,
	}).Error
	require.NoError(s.T(), err)

	var ended ObservabilityRun
	err = s.db.GORM.Where("id = ?", run.ID).First(&ended).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "completed", ended.Status)
	assert.NotNil(s.T(), ended.EndedAt)
}

func (s *AgentObservabilityRepositoryTestSuite) TestGetConfig() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	config := &ObservabilityConfig{
		TenantID:          tenantID,
		AtlasTenantID:     atlasTenantID,
		SamplingRate:      0.5,
		TraceErrorsOnly:   false,
		SampleHeadPercent: 100,
		SampleTailCount:   10,
		RetentionDays:     90,
		IsActive:          true,
	}
	err := s.db.GORM.Create(config).Error
	require.NoError(s.T(), err)

	var retrieved ObservabilityConfig
	err = s.db.GORM.Where("tenant_id = ?", tenantID).First(&retrieved).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0.5, retrieved.SamplingRate)
	assert.Equal(s.T(), 90, retrieved.RetentionDays)
}

func (s *AgentObservabilityRepositoryTestSuite) TestShouldTrace_FullSampling() {
	ctx := context.Background()
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	config := &ObservabilityConfig{
		TenantID:          tenantID,
		AtlasTenantID:     atlasTenantID,
		SamplingRate:      1.0,
		TraceErrorsOnly:   false,
		SampleHeadPercent: 100,
		SampleTailCount:   10,
		RetentionDays:     90,
		IsActive:          true,
	}
	err := s.db.GORM.Create(config).Error
	require.NoError(s.T(), err)

	repo := NewAgentObservabilityRepository(s.db.GORM)
	shouldTrace, err := repo.ShouldTrace(ctx, tenantID, false)
	require.NoError(s.T(), err)
	assert.True(s.T(), shouldTrace)
}

func (s *AgentObservabilityRepositoryTestSuite) TestShouldTrace_Disabled() {
	ctx := context.Background()
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	config := &ObservabilityConfig{
		TenantID:          tenantID,
		AtlasTenantID:     atlasTenantID,
		SamplingRate:      1.0,
		TraceErrorsOnly:   false,
		SampleHeadPercent: 100,
		SampleTailCount:   10,
		RetentionDays:     90,
		IsActive:          false,
	}
	err := s.db.GORM.Create(config).Error
	require.NoError(s.T(), err)

	repo := NewAgentObservabilityRepository(s.db.GORM)
	shouldTrace, err := repo.ShouldTrace(ctx, tenantID, false)
	require.NoError(s.T(), err)
	assert.False(s.T(), shouldTrace)
}

func (s *AgentObservabilityRepositoryTestSuite) TestShouldTrace_ErrorsOnly() {
	ctx := context.Background()
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	config := &ObservabilityConfig{
		TenantID:          tenantID,
		AtlasTenantID:     atlasTenantID,
		SamplingRate:      0.0,
		TraceErrorsOnly:   true,
		SampleHeadPercent: 100,
		SampleTailCount:   10,
		RetentionDays:     90,
		IsActive:          true,
	}
	err := s.db.GORM.Create(config).Error
	require.NoError(s.T(), err)

	repo := NewAgentObservabilityRepository(s.db.GORM)

	shouldTrace, err := repo.ShouldTrace(ctx, tenantID, false)
	require.NoError(s.T(), err)
	assert.False(s.T(), shouldTrace)

	shouldTrace, err = repo.ShouldTrace(ctx, tenantID, true)
	require.NoError(s.T(), err)
	assert.True(s.T(), shouldTrace)
}

func (s *AgentObservabilityRepositoryTestSuite) TestDeleteOldRuns() {
	tenantID := uuid.New()
	atlasTenantID := DeriveAtlasTenantID(tenantID)

	oldRun := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "old-run",
		AgentID:       "agent-old",
		AgentType:     "flymind",
		Status:        "completed",
		EndedAt:       timePtr(time.Now().AddDate(0, 0, -100)),
	}
	err := s.db.GORM.Create(oldRun).Error
	require.NoError(s.T(), err)

	newRun := &ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    "new-run",
		AgentID:       "agent-new",
		AgentType:     "flymind",
		Status:        "completed",
		EndedAt:       timePtr(time.Now()),
	}
	err = s.db.GORM.Create(newRun).Error
	require.NoError(s.T(), err)

	repo := NewAgentObservabilityRepository(s.db.GORM)
	_, err = repo.DeleteOldRuns(context.Background(), tenantID, 90)
	require.NoError(s.T(), err)

	var remaining []*ObservabilityRun
	err = s.db.GORM.Where("tenant_id = ?", tenantID).Find(&remaining).Error
	require.NoError(s.T(), err)
	assert.Len(s.T(), remaining, 1)
	assert.Equal(s.T(), "new-run", remaining[0].AtlasRunID)
}

func (s *AgentObservabilityRepositoryTestSuite) TestDeriveAtlasTenantID() {
	tenantID := uuid.New()
	atlasID1 := DeriveAtlasTenantID(tenantID)
	atlasID2 := DeriveAtlasTenantID(tenantID)

	assert.Equal(s.T(), atlasID1, atlasID2)
	assert.Len(s.T(), atlasID1, 16)

	differentTenantID := uuid.New()
	atlasID3 := DeriveAtlasTenantID(differentTenantID)

	assert.NotEqual(s.T(), atlasID1, atlasID3)
}

func timePtr(t time.Time) *time.Time {
	return &t
}
