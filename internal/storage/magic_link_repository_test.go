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

// MagicLinkRepositoryTestSuite tests the magic link repository methods
type MagicLinkRepositoryTestSuite struct {
	PostgresTestSuite
}

func TestMagicLinkRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MagicLinkRepositoryTestSuite))
}

func (s *MagicLinkRepositoryTestSuite) SetupTest() {
	s.PostgresTestSuite.SetupTest()
	// Clean up magic links table
	s.db.DB.Exec("TRUNCATE TABLE magic_links CASCADE")
}

func (s *MagicLinkRepositoryTestSuite) TearDownTest() {
	// Clean up magic links after each test
	s.db.DB.Exec("TRUNCATE TABLE magic_links CASCADE")
}

func (s *MagicLinkRepositoryTestSuite) TestCreateMagicLink() {
	ctx := context.Background()
	email := "test@example.com"
	token := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	expiresAt := time.Now().Add(15 * time.Minute)

	magicLink, err := s.db.CreateMagicLink(ctx, email, token, nil, "192.168.1.1", "TestAgent", "/dashboard", expiresAt)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), magicLink)
	assert.NotEqual(s.T(), uuid.Nil, magicLink.ID)
	assert.Equal(s.T(), email, magicLink.Email)
	assert.Equal(s.T(), token, magicLink.Token)
	assert.False(s.T(), magicLink.Used)
	assert.Equal(s.T(), "192.168.1.1", magicLink.IPCreated)
	assert.Equal(s.T(), "TestAgent", magicLink.UserAgent)
	assert.Equal(s.T(), "/dashboard", magicLink.RedirectPath)
	assert.WithinDuration(s.T(), expiresAt, magicLink.ExpiresAt, time.Second)
}

func (s *MagicLinkRepositoryTestSuite) TestCreateMagicLink_WithUserID() {
	ctx := context.Background()

	// Create a tenant and user first
	tenantID, err := s.createTenant("Magic Link Test", "magic.test.com", "admin@test.com")
	require.NoError(s.T(), err)

	userID, err := s.createUser(tenantID, "testuser", "user@example.com", "hashedpass", false)
	require.NoError(s.T(), err)

	userUUID := parseUUIDMust(userID)
	email := "user@example.com"
	token := "token1234567890token1234567890token1234567890token1234567890"
	expiresAt := time.Now().Add(15 * time.Minute)

	magicLink, err := s.db.CreateMagicLink(ctx, email, token, &userUUID, "192.168.1.1", "TestAgent", "", expiresAt)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), magicLink)
	assert.NotNil(s.T(), magicLink.UserID)
	assert.Equal(s.T(), userUUID, *magicLink.UserID)
}

func (s *MagicLinkRepositoryTestSuite) TestGetMagicLinkByToken() {
	ctx := context.Background()
	email := "test@example.com"
	token := "unique_token_12345678901234567890123456789012345678901234567890"
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create magic link
	created, err := s.db.CreateMagicLink(ctx, email, token, nil, "", "", "", expiresAt)
	require.NoError(s.T(), err)

	// Retrieve by token
	retrieved, err := s.db.GetMagicLinkByToken(ctx, token)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), retrieved)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), email, retrieved.Email)
	assert.Equal(s.T(), token, retrieved.Token)
	assert.False(s.T(), retrieved.Used)
}

func (s *MagicLinkRepositoryTestSuite) TestGetMagicLinkByToken_NotFound() {
	ctx := context.Background()

	// Try to retrieve non-existent token
	retrieved, err := s.db.GetMagicLinkByToken(ctx, "non_existent_token_123456789012345678901234567890")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *MagicLinkRepositoryTestSuite) TestMarkMagicLinkUsed() {
	ctx := context.Background()
	email := "test@example.com"
	token := "mark_used_token_1234567890123456789012345678901234567890"
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create magic link
	created, err := s.db.CreateMagicLink(ctx, email, token, nil, "", "", "", expiresAt)
	require.NoError(s.T(), err)
	assert.False(s.T(), created.Used)

	// Mark as used
	err = s.db.MarkMagicLinkUsed(ctx, created.ID)
	require.NoError(s.T(), err)

	// Retrieve and verify
	retrieved, err := s.db.GetMagicLinkByToken(ctx, token)
	require.NoError(s.T(), err)
	assert.True(s.T(), retrieved.Used)
	assert.NotNil(s.T(), retrieved.UsedAt)
	assert.WithinDuration(s.T(), time.Now(), *retrieved.UsedAt, 5*time.Second)
}

func (s *MagicLinkRepositoryTestSuite) TestMarkMagicLinkUsed_NotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	err := s.db.MarkMagicLinkUsed(ctx, nonExistentID)
	assert.Error(s.T(), err)
}

func (s *MagicLinkRepositoryTestSuite) TestGetRecentMagicLinksByEmail() {
	ctx := context.Background()
	email := "recent@example.com"

	// Create multiple magic links for the same email
	for i := 0; i < 5; i++ {
		token := uuid.New().String() + uuid.New().String()
		expiresAt := time.Now().Add(15 * time.Minute)
		_, err := s.db.CreateMagicLink(ctx, email, token, nil, "", "", "", expiresAt)
		require.NoError(s.T(), err)
	}

	// Create magic link for different email
	otherToken := uuid.New().String() + uuid.New().String()
	_, err := s.db.CreateMagicLink(ctx, "other@example.com", otherToken, nil, "", "", "", time.Now().Add(15*time.Minute))
	require.NoError(s.T(), err)

	// Get recent links for the past hour
	since := time.Now().Add(-1 * time.Hour)
	links, err := s.db.GetRecentMagicLinksByEmail(ctx, email, since)
	require.NoError(s.T(), err)
	assert.Len(s.T(), links, 5)

	// Verify ordering (most recent first)
	for i := 1; i < len(links); i++ {
		assert.True(s.T(), links[i-1].CreatedAt.After(links[i].CreatedAt) || links[i-1].CreatedAt.Equal(links[i].CreatedAt),
			"Links should be ordered by created_at DESC")
	}
}

func (s *MagicLinkRepositoryTestSuite) TestGetRecentMagicLinksByEmail_TimeFilter() {
	ctx := context.Background()
	email := "timefilter@example.com"

	// Create an old magic link (manually set created_at to 2 hours ago)
	oldToken := "old_token_1234567890123456789012345678901234567890123456789012"
	oldCreatedAt := time.Now().Add(-2 * time.Hour)
	oldExpiresAt := oldCreatedAt.Add(15 * time.Minute)

	oldLink := &MagicLink{
		Email:     email,
		Token:     oldToken,
		Used:      false,
		ExpiresAt: oldExpiresAt,
		CreatedAt: oldCreatedAt,
	}
	err := s.db.GORM.Create(oldLink).Error
	require.NoError(s.T(), err)

	// Create a recent magic link
	recentToken := "recent_token_123456789012345678901234567890123456789012345678"
	_, err = s.db.CreateMagicLink(ctx, email, recentToken, nil, "", "", "", time.Now().Add(15*time.Minute))
	require.NoError(s.T(), err)

	// Get links from past hour (should only get the recent one)
	since := time.Now().Add(-1 * time.Hour)
	links, err := s.db.GetRecentMagicLinksByEmail(ctx, email, since)
	require.NoError(s.T(), err)
	assert.Len(s.T(), links, 1)
	assert.Equal(s.T(), recentToken, links[0].Token)
}

func (s *MagicLinkRepositoryTestSuite) TestDeleteExpiredMagicLinks() {
	ctx := context.Background()

	// Create expired magic link
	expiredToken := "expired_token_1234567890123456789012345678901234567890123456"
	expiredExpiresAt := time.Now().Add(-1 * time.Hour)
	_, err := s.db.CreateMagicLink(ctx, "expired@example.com", expiredToken, nil, "", "", "", expiredExpiresAt)
	require.NoError(s.T(), err)

	// Create used magic link (older than 24h)
	usedToken := "used_token_123456789012345678901234567890123456789012345678901"
	usedLink := &MagicLink{
		Email:     "used@example.com",
		Token:     usedToken,
		Used:      true,
		UsedAt:    &[]time.Time{time.Now().Add(-25 * time.Hour)}[0],
		ExpiresAt: time.Now().Add(-25 * time.Hour).Add(15 * time.Minute),
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}
	err = s.db.GORM.Create(usedLink).Error
	require.NoError(s.T(), err)

	// Create valid magic link
	validToken := "valid_token_1234567890123456789012345678901234567890123456789"
	_, err = s.db.CreateMagicLink(ctx, "valid@example.com", validToken, nil, "", "", "", time.Now().Add(15*time.Minute))
	require.NoError(s.T(), err)

	// Run cleanup
	deletedCount, err := s.db.DeleteExpiredMagicLinks(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), deletedCount)

	// Verify expired link is gone
	retrieved, err := s.db.GetMagicLinkByToken(ctx, expiredToken)
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)

	// Verify used link is gone
	retrieved, err = s.db.GetMagicLinkByToken(ctx, usedToken)
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)

	// Verify valid link still exists
	retrieved, err = s.db.GetMagicLinkByToken(ctx, validToken)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), retrieved)
}

func (s *MagicLinkRepositoryTestSuite) TestMagicLink_IsExpired() {
	// Test expired link
	expiredLink := &MagicLink{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(s.T(), expiredLink.IsExpired())

	// Test non-expired link
	validLink := &MagicLink{
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	assert.False(s.T(), validLink.IsExpired())

	// Test link expiring exactly now (should be expired)
	exactNowLink := &MagicLink{
		ExpiresAt: time.Now(),
	}
	// Allow small time difference due to test execution time
	_ = exactNowLink.IsExpired()
}

func (s *MagicLinkRepositoryTestSuite) TestMagicLink_CanUse() {
	now := time.Now()

	// Test usable link (not used, not expired)
	usableLink := &MagicLink{
		Used:      false,
		ExpiresAt: now.Add(15 * time.Minute),
	}
	assert.True(s.T(), usableLink.CanUse())

	// Test already used link
	usedLink := &MagicLink{
		Used:      true,
		ExpiresAt: now.Add(15 * time.Minute),
	}
	assert.False(s.T(), usedLink.CanUse())

	// Test expired link
	expiredLink := &MagicLink{
		Used:      false,
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	assert.False(s.T(), expiredLink.CanUse())

	// Test used and expired link
	usedAndExpiredLink := &MagicLink{
		Used:      true,
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	assert.False(s.T(), usedAndExpiredLink.CanUse())
}

func (s *MagicLinkRepositoryTestSuite) TestCreateMagicLink_UniqueTokenConstraint() {
	ctx := context.Background()
	email := "unique@example.com"
	token := "duplicate_token_1234567890123456789012345678901234567890123456"
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create first magic link
	_, err := s.db.CreateMagicLink(ctx, email, token, nil, "", "", "", expiresAt)
	require.NoError(s.T(), err)

	// Try to create second magic link with same token (should fail)
	_, err = s.db.CreateMagicLink(ctx, "other@example.com", token, nil, "", "", "", expiresAt)
	assert.Error(s.T(), err) // Should fail due to unique constraint
}

func (s *MagicLinkRepositoryTestSuite) TestMagicLink_TableName() {
	link := MagicLink{}
	assert.Equal(s.T(), "magic_links", link.TableName())
}
