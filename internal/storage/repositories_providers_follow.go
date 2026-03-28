package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: LLM providers, team invites, follows.

// Provider operations
func (db *PostgresDB) CreateProvider(provider *Provider) error {
	// Encrypt the token before storing
	if provider.Token != "" {
		encryptedToken, err := db.EncryptField(provider.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt provider token: %w", err)
		}
		provider.Token = encryptedToken
	}
	return db.GORM.Create(provider).Error
}

func (db *PostgresDB) GetProviderByID(providerID string) (*Provider, error) {
	var provider Provider
	err := db.GORM.Where("id = ?", providerID).First(&provider).Error
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	if provider.Token != "" {
		decryptedToken, err := db.DecryptField(provider.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt provider token: %w", err)
		}
		provider.Token = decryptedToken
	}

	return &provider, nil
}

func (db *PostgresDB) GetProviderByUserAndType(userID uuid.UUID, providerType string) (*Provider, error) {
	var provider Provider
	err := db.GORM.Where("user_id = ? AND provider = ? AND status = 'active'", userID, providerType).First(&provider).Error
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	if provider.Token != "" {
		decryptedToken, err := db.DecryptField(provider.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt provider token: %w", err)
		}
		provider.Token = decryptedToken
	}

	return &provider, nil
}

func (db *PostgresDB) GetProvidersByUser(userID uuid.UUID) ([]*Provider, error) {
	var providers []*Provider
	err := db.GORM.Where("user_id = ? AND status = 'active'", userID).Find(&providers).Error
	if err != nil {
		return nil, err
	}

	// Decrypt tokens for all providers
	for _, provider := range providers {
		if provider.Token != "" {
			decryptedToken, err := db.DecryptField(provider.Token)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt provider token for %s: %w", provider.ID, err)
			}
			provider.Token = decryptedToken
		}
	}

	return providers, nil
}

func (db *PostgresDB) UpdateProviderStatus(providerID string, status string) error {
	return db.GORM.Model(&Provider{}).Where("id = ?", providerID).Update("status", status).Error
}

func (db *PostgresDB) DeleteProvider(ctx context.Context, providerID string, userID uuid.UUID) error {
	result := db.GORM.WithContext(ctx).
		Where("id = ? AND user_id = ?", providerID, userID).
		Delete(&Provider{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("provider not found or access denied")
	}
	return nil
}

func (db *PostgresDB) ListAllProviders(ctx context.Context) ([]*Provider, error) {
	var providers []*Provider
	err := db.GORM.WithContext(ctx).Find(&providers).Error
	return providers, err
}

func (db *PostgresDB) UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*Provider, error) {
	// Get current provider
	var provider Provider
	if err := db.GORM.WithContext(ctx).Where("id = ?", providerID).First(&provider).Error; err != nil {
		return nil, err
	}

	// Apply updates
	if status, ok := updates["status"].(string); ok {
		provider.Status = status
	}
	if isShared, ok := updates["is_shared"].(bool); ok {
		provider.IsShared = isShared
	}
	if teamID, ok := updates["team_id"].(string); ok {
		provider.TeamID = &teamID
	}

	// Update in database
	if err := db.GORM.WithContext(ctx).Save(&provider).Error; err != nil {
		return nil, err
	}

	return &provider, nil
}

func (db *PostgresDB) ShareProviderWithTeam(providerID string, teamID string) error {
	return db.GORM.Model(&Provider{}).Where("id = ?", providerID).Updates(map[string]interface{}{
		"is_shared": true,
		"team_id":   teamID,
	}).Error
}

// Team invite operations
func (db *PostgresDB) CreateTeamInvite(invite *TeamInvite) error {
	return db.GORM.Create(invite).Error
}

func (db *PostgresDB) GetTeamInviteByToken(token string) (*TeamInvite, error) {
	var invite TeamInvite
	err := db.GORM.Where("token = ? AND status = 'pending' AND expires_at > ?", token, time.Now()).First(&invite).Error
	return &invite, err
}

func (db *PostgresDB) GetTeamInvitesByTeam(teamID uuid.UUID) ([]*TeamInvite, error) {
	var invites []*TeamInvite
	err := db.GORM.Where("team_id = ?", teamID).Find(&invites).Error
	return invites, err
}

func (db *PostgresDB) UpdateTeamInviteStatus(inviteID uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "accepted" {
		now := time.Now()
		updates["accepted_at"] = now
	}
	return db.GORM.Model(&TeamInvite{}).Where("id = ?", inviteID).Updates(updates).Error
}

func (db *PostgresDB) GetTeamByUserID(userID uuid.UUID) (*Team, error) {
	var membership TeamMembership
	err := db.GORM.Where("user_id = ?", userID).First(&membership).Error
	if err != nil {
		return nil, err
	}

	var team Team
	err = db.GORM.Where("id = ?", membership.TeamID).First(&team).Error
	return &team, err
}

func (db *PostgresDB) IsTeamAdmin(userID uuid.UUID, teamID string) (bool, error) {
	var membership TeamMembership
	err := db.GORM.Where("user_id = ? AND team_id = ? AND role = 'admin'", userID, teamID).First(&membership).Error
	return err == nil, err
}

// Follow operations

// User follows
func (db *PostgresDB) FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*UserFollow, error) {
	return db.followRepository.FollowUser(ctx, followerID, followedUserID, reason, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion)
}

func (db *PostgresDB) UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error {
	return db.followRepository.UnfollowUser(ctx, followerID, followedUserID)
}

func (db *PostgresDB) IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error) {
	return db.followRepository.IsFollowingUser(ctx, followerID, followedUserID)
}

func (db *PostgresDB) GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	return db.followRepository.GetUserFollowers(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	return db.followRepository.GetUserFollowing(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.followRepository.GetUserFollowerCount(ctx, userID)
}

func (db *PostgresDB) GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.followRepository.GetUserFollowingCount(ctx, userID)
}

// Function follows
func (db *PostgresDB) FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*FunctionFollow, error) {
	return db.followRepository.FollowFunction(ctx, userID, functionID, reason, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification)
}

func (db *PostgresDB) UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error {
	return db.followRepository.UnfollowFunction(ctx, userID, functionID)
}

func (db *PostgresDB) IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error) {
	return db.followRepository.IsFollowingFunction(ctx, userID, functionID)
}

func (db *PostgresDB) GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	return db.followRepository.GetFunctionFollowers(ctx, functionID, limit, offset)
}

func (db *PostgresDB) GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	return db.followRepository.GetUserFunctionFollows(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	return db.followRepository.GetFunctionFollowerCount(ctx, functionID)
}
