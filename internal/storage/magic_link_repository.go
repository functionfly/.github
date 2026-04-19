package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateMagicLink creates a new magic link token
func (db *PostgresDB) CreateMagicLink(ctx context.Context, email string, token string, userID *uuid.UUID, ipAddress, userAgent, redirectPath string, expiresAt time.Time) (*MagicLink, error) {
	magicLink := &MagicLink{
		Email:        email,
		Token:        token,
		UserID:       userID,
		IPCreated:    ipAddress,
		UserAgent:    userAgent,
		RedirectPath: redirectPath,
		ExpiresAt:    expiresAt,
		Used:         false,
		CreatedAt:    time.Now(),
	}

	if err := db.GORM.WithContext(ctx).Create(magicLink).Error; err != nil {
		return nil, err
	}

	return magicLink, nil
}

// GetMagicLinkByToken retrieves a magic link by its token
func (db *PostgresDB) GetMagicLinkByToken(ctx context.Context, token string) (*MagicLink, error) {
	var magicLink MagicLink
	if err := db.GORM.WithContext(ctx).Where("token = ?", token).First(&magicLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &magicLink, nil
}

// MarkMagicLinkUsed marks a magic link as used
func (db *PostgresDB) MarkMagicLinkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := db.GORM.WithContext(ctx).Model(&MagicLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetRecentMagicLinksByEmail retrieves recent magic links for an email address
func (db *PostgresDB) GetRecentMagicLinksByEmail(ctx context.Context, email string, since time.Time) ([]*MagicLink, error) {
	var magicLinks []*MagicLink
	if err := db.GORM.WithContext(ctx).
		Where("email = ? AND created_at >= ?", email, since).
		Order("created_at DESC").
		Find(&magicLinks).Error; err != nil {
		return nil, err
	}
	return magicLinks, nil
}

// DeleteExpiredMagicLinks deletes all expired magic links
func (db *PostgresDB) DeleteExpiredMagicLinks(ctx context.Context) (int64, error) {
	now := time.Now()
	result := db.GORM.WithContext(ctx).
		Where("expires_at < ? OR (used = ? AND created_at < ?)", now, true, now.Add(-24*time.Hour)).
		Delete(&MagicLink{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
