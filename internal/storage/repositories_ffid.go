package storage

import (
	"context"

	"github.com/google/uuid"
)

// FFID Identity System operations

func (db *PostgresDB) GenerateFFID(ctx context.Context, tenantID uuid.UUID) (string, error) {
	return db.ffidRepository.GenerateFFID(ctx, tenantID)
}

func (db *PostgresDB) GetIdentityCard(ctx context.Context, employeeID uuid.UUID) (*IdentityCard, error) {
	return db.ffidRepository.GetIdentityCard(ctx, employeeID)
}

func (db *PostgresDB) CreateAchievementDefinition(ctx context.Context, ach *AchievementDefinition) (*AchievementDefinition, error) {
	return db.ffidRepository.CreateAchievementDefinition(ctx, ach)
}

func (db *PostgresDB) ListAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*AchievementDefinition, error) {
	return db.ffidRepository.ListAchievementDefinitions(ctx, tenantID)
}

func (db *PostgresDB) CheckAndAwardAchievements(ctx context.Context, employeeID uuid.UUID) error {
	return db.ffidRepository.CheckAndAwardAchievements(ctx, employeeID)
}

func (db *PostgresDB) GetAchievementProgress(ctx context.Context, employeeID uuid.UUID) ([]*AchievementProgress, error) {
	return db.ffidRepository.GetAchievementProgress(ctx, employeeID)
}

func (db *PostgresDB) CreateCareerTimelineEvent(ctx context.Context, ev *CareerTimelineEvent) (*CareerTimelineEvent, error) {
	return db.ffidRepository.CreateCareerTimelineEvent(ctx, ev)
}

func (db *PostgresDB) GetCareerTimeline(ctx context.Context, employeeID uuid.UUID) ([]*CareerTimelineEvent, error) {
	return db.ffidRepository.GetCareerTimeline(ctx, employeeID)
}

func (db *PostgresDB) RecordReputationSnapshot(ctx context.Context, employeeID, tenantID uuid.UUID, category string, score float64) error {
	return db.ffidRepository.RecordReputationSnapshot(ctx, employeeID, tenantID, category, score)
}

func (db *PostgresDB) GetReputationHistory(ctx context.Context, employeeID uuid.UUID, category string) ([]*ReputationHistory, error) {
	return db.ffidRepository.GetReputationHistory(ctx, employeeID, category)
}

func (db *PostgresDB) UpdateClearanceLevel(ctx context.Context, employeeID uuid.UUID, level int) error {
	return db.ffidRepository.UpdateClearanceLevel(ctx, employeeID, level)
}

func (db *PostgresDB) UpdateIdentitySignature(ctx context.Context, employeeID uuid.UUID, signature string) error {
	return db.ffidRepository.UpdateIdentitySignature(ctx, employeeID, signature)
}

func (db *PostgresDB) SeedAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) error {
	return db.ffidRepository.SeedAchievementDefinitions(ctx, tenantID)
}
