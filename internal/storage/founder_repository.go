package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const maxFounders = 10000

func (db *PostgresDB) GetFounderCount(ctx context.Context) (int, error) {
	var count int64
	if err := db.GORM.WithContext(ctx).Model(&User{}).Where("is_founder = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get founder count: %w", err)
	}
	return int(count), nil
}

func (db *PostgresDB) GetUserFounderStatus(ctx context.Context, userID uuid.UUID) (bool, *int, error) {
	var user struct {
		IsFounder     bool
		FounderNumber sql.NullInt64
	}
	if err := db.GORM.WithContext(ctx).Table("users").
		Select("is_founder, founder_number").
		Where("id = ?", userID).
		Scan(&user).Error; err != nil {
		return false, nil, fmt.Errorf("failed to get user founder status: %w", err)
	}
	if !user.IsFounder {
		return false, nil, nil
	}
	if user.FounderNumber.Valid {
		fn := int(user.FounderNumber.Int64)
		return true, &fn, nil
	}
	return true, nil, nil
}

func (db *PostgresDB) AssignFounderStatus(ctx context.Context, userID uuid.UUID) (int, error) {
	lockID := int64(88374401)

	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", lockID); err != nil {
		return 0, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	count, err := db.GetFounderCount(ctx)
	if err != nil {
		return 0, err
	}
	if count >= maxFounders {
		return 0, nil
	}

	var founderNumber int
	if err := db.GORM.WithContext(ctx).Exec(`
		UPDATE users
		SET is_founder = TRUE,
			founder_number = (
				SELECT COALESCE(MAX(founder_number), 0) + 1
				FROM users
				WHERE founder_number IS NOT NULL
			),
			updated_at = NOW()
		WHERE id = ? AND is_founder = FALSE
	`, userID).Error; err != nil {
		return 0, fmt.Errorf("failed to assign founder status: %w", err)
	}

	if err := db.GORM.WithContext(ctx).Table("users").
		Select("founder_number").
		Where("id = ?", userID).
		Scan(&founderNumber).Error; err != nil {
		return 0, fmt.Errorf("failed to get assigned founder number: %w", err)
	}

	return founderNumber, nil
}

func (db *PostgresDB) GetFounderVote(ctx context.Context, voteID uuid.UUID) (*FounderVote, error) {
	var vote FounderVote
	if err := db.GORM.WithContext(ctx).Where("id = ?", voteID).First(&vote).Error; err != nil {
		return nil, fmt.Errorf("failed to get founder vote: %w", err)
	}
	return &vote, nil
}

func (db *PostgresDB) ListActiveFounderVotes(ctx context.Context) ([]*FounderVote, error) {
	var votes []*FounderVote
	if err := db.GORM.WithContext(ctx).
		Where("status = ?", "active").
		Order("created_at DESC").
		Find(&votes).Error; err != nil {
		return nil, fmt.Errorf("failed to list active founder votes: %w", err)
	}
	return votes, nil
}

func (db *PostgresDB) ListFounderVotes(ctx context.Context) ([]*FounderVote, error) {
	var votes []*FounderVote
	if err := db.GORM.WithContext(ctx).Order("created_at DESC").Find(&votes).Error; err != nil {
		return nil, fmt.Errorf("failed to list founder votes: %w", err)
	}
	return votes, nil
}

func (db *PostgresDB) CreateFounderVote(ctx context.Context, vote *FounderVote) error {
	if err := db.GORM.WithContext(ctx).Create(vote).Error; err != nil {
		return fmt.Errorf("failed to create founder vote: %w", err)
	}
	return nil
}

func (db *PostgresDB) UpdateFounderVote(ctx context.Context, voteID uuid.UUID, updates map[string]interface{}) error {
	if err := db.GORM.WithContext(ctx).Model(&FounderVote{}).Where("id = ?", voteID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update founder vote: %w", err)
	}
	return nil
}

func (db *PostgresDB) GetFounderVoteResponse(ctx context.Context, voteID, userID uuid.UUID) (*FounderVoteResponse, error) {
	var response FounderVoteResponse
	if err := db.GORM.WithContext(ctx).
		Where("vote_id = ? AND user_id = ?", voteID, userID).
		First(&response).Error; err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get founder vote response: %w", err)
	}
	return &response, nil
}

func (db *PostgresDB) CastFounderVote(ctx context.Context, voteID, userID uuid.UUID, optionID string) error {
	response := &FounderVoteResponse{
		ID:        uuid.New(),
		VoteID:    voteID,
		UserID:    userID,
		OptionID:  optionID,
		CreatedAt: time.Now(),
	}
	if err := db.GORM.WithContext(ctx).Create(response).Error; err != nil {
		return fmt.Errorf("failed to cast founder vote: %w", err)
	}
	return nil
}

func (db *PostgresDB) GetFounderVoteResults(ctx context.Context, voteID uuid.UUID) (map[string]int, int, error) {
	var responses []struct {
		OptionID string `gorm:"column:option_id"`
		Count    int    `gorm:"column:cnt"`
	}
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT option_id, COUNT(*)::int AS cnt
		FROM founder_vote_responses
		WHERE vote_id = ?
		GROUP BY option_id
	`, voteID).Scan(&responses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get vote results: %w", err)
	}

	results := make(map[string]int)
	var total int
	for _, r := range responses {
		results[r.OptionID] = r.Count
		total += r.Count
	}
	return results, total, nil
}

func (db *PostgresDB) GetFounderEarlyAccessFeatures(ctx context.Context) ([]*FounderEarlyAccessFeature, error) {
	var features []*FounderEarlyAccessFeature
	if err := db.GORM.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&features).Error; err != nil {
		return nil, fmt.Errorf("failed to get early access features: %w", err)
	}
	return features, nil
}

func (db *PostgresDB) GetFounderEarlyAccessFeatureBySlug(ctx context.Context, slug string) (*FounderEarlyAccessFeature, error) {
	var feature FounderEarlyAccessFeature
	if err := db.GORM.WithContext(ctx).Where("slug = ?", slug).First(&feature).Error; err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get early access feature: %w", err)
	}
	return &feature, nil
}

func (db *PostgresDB) GetUserFounderEarlyAccess(ctx context.Context, userID uuid.UUID) ([]*FounderEarlyAccess, error) {
	var access []*FounderEarlyAccess
	if err := db.GORM.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("accessed_at DESC").
		Find(&access).Error; err != nil {
		return nil, fmt.Errorf("failed to get user early access: %w", err)
	}
	return access, nil
}

func (db *PostgresDB) ClaimFounderEarlyAccess(ctx context.Context, userID uuid.UUID, feature *FounderEarlyAccessFeature) error {
	access := &FounderEarlyAccess{
		ID:          uuid.New(),
		UserID:      userID,
		FeatureSlug: feature.Slug,
		FeatureName: feature.Name,
		AccessedAt:  time.Now(),
	}
	if err := db.GORM.WithContext(ctx).Create(access).Error; err != nil {
		return fmt.Errorf("failed to claim early access: %w", err)
	}
	return nil
}

func (db *PostgresDB) HasUserClaimedEarlyAccess(ctx context.Context, userID uuid.UUID, slug string) (bool, error) {
	var count int64
	if err := db.GORM.WithContext(ctx).Model(&FounderEarlyAccess{}).
		Where("user_id = ? AND feature_slug = ?", userID, slug).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check early access: %w", err)
	}
	return count > 0, nil
}

func (db *PostgresDB) GetFoundersLeaderboard(ctx context.Context, limit int) ([]*User, int, error) {
	var users []*User
	if err := db.GORM.WithContext(ctx).
		Where("is_founder = ?", true).
		Order("founder_number ASC").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get founders leaderboard: %w", err)
	}

	var total int64
	if err := db.GORM.WithContext(ctx).Model(&User{}).
		Where("is_founder = ?", true).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count founders: %w", err)
	}

	return users, int(total), nil
}

const (
	FounderTierStandard = "founder"
	FounderTierPro      = "founder_pro"
	FounderTierElite    = "founder_elite"
)

type FounderTier struct {
	UserID           uuid.UUID `gorm:"column:user_id;primaryKey"`
	Tier             string    `gorm:"column:tier"`
	ReferralCount    int       `gorm:"column:referral_count"`
	TotalEarningsCents int64   `gorm:"column:total_earnings_cents"`
	Rank             *int      `gorm:"column:rank"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (db *PostgresDB) GetFounderTier(ctx context.Context, userID uuid.UUID) (string, error) {
	var tier FounderTier
	if err := db.GORM.WithContext(ctx).Where("user_id = ?", userID).First(&tier).Error; err != nil {
		if err == sql.ErrNoRows {
			return FounderTierStandard, nil
		}
		return "", fmt.Errorf("failed to get founder tier: %w", err)
	}
	return tier.Tier, nil
}

func (db *PostgresDB) UpsertFounderTier(ctx context.Context, userID uuid.UUID, tier string, referralCount int, totalEarningsCents int64) error {
	tierRecord := FounderTier{
		UserID:           userID,
		Tier:             tier,
		ReferralCount:    referralCount,
		TotalEarningsCents: totalEarningsCents,
		UpdatedAt:        time.Now(),
	}

	err := db.GORM.WithContext(ctx).Save(&tierRecord).Error
	if err != nil {
		return fmt.Errorf("failed to upsert founder tier: %w", err)
	}
	return nil
}

func (db *PostgresDB) GetFounderRank(ctx context.Context, userID uuid.UUID) (int, error) {
	var tier FounderTier
	if err := db.GORM.WithContext(ctx).Where("user_id = ?", userID).First(&tier).Error; err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get founder rank: %w", err)
	}
	if tier.Rank == nil {
		return 0, nil
	}
	return *tier.Rank, nil
}

func (db *PostgresDB) CalculateAndUpdateFounderTiers(ctx context.Context) error {
	type ReferralStats struct {
		PublisherID uuid.UUID
		Count       int
	}

	var stats []ReferralStats
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT ac.publisher_id as publisher_id, COUNT(ar.id) as count
		FROM affiliate_codes ac
		JOIN affiliate_referrals ar ON ar.affiliate_code_id = ac.id
		WHERE ar.status IN ('converted', 'qualified')
		GROUP BY ac.publisher_id
	`).Scan(&stats).Error; err != nil {
		return fmt.Errorf("failed to calculate referral stats: %w", err)
	}

	for _, stat := range stats {
		var tier string
		if stat.Count >= 10 {
			tier = FounderTierElite
		} else if stat.Count >= 3 {
			tier = FounderTierPro
		} else {
			tier = FounderTierStandard
		}

		if err := db.UpsertFounderTier(ctx, stat.PublisherID, tier, stat.Count, 0); err != nil {
			logrus.WithError(err).WithField("user_id", stat.PublisherID).Warn("failed to update founder tier")
		}
	}

	return nil
}
