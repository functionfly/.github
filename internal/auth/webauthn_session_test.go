package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a miniredis instance for testing
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

// TestGenerateSecureSessionID tests the cryptographically secure session ID generation
func TestGenerateSecureSessionID(t *testing.T) {
	// Generate multiple session IDs and verify they're unique
	sessionIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sessionID, err := generateSecureSessionID()
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Len(t, sessionID, 32, "session ID should be 32 hex characters (16 bytes)")
		assert.False(t, sessionIDs[sessionID], "session IDs should be unique")
		sessionIDs[sessionID] = true
	}
}

// TestGenerateSecureSessionID_Length tests that session IDs have the correct length
func TestGenerateSecureSessionID_Length(t *testing.T) {
	sessionID, err := generateSecureSessionID()
	require.NoError(t, err)
	assert.Len(t, sessionID, 32)
}

// TestWebAuthnSessionStore_Create tests creating a new WebAuthn session
func TestWebAuthnSessionStore_Create(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	session := &WebAuthnSession{
		Challenge:  "test-challenge",
		UserHandle: "test-user-handle",
		UserID:     "test-user-id",
		Operation:  "registration",
	}

	sessionID, err := store.Create(ctx, session)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.Len(t, sessionID, 32, "session ID should be 32 hex characters from crypto/rand")

	// Verify session was stored
	retrieved, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, session.Challenge, retrieved.Challenge)
	assert.Equal(t, session.UserHandle, retrieved.UserHandle)
	assert.Equal(t, session.UserID, retrieved.UserID)
	assert.Equal(t, session.Operation, retrieved.Operation)
}

// TestWebAuthnSessionStore_Create_Authentication tests creating a WebAuthn session for authentication
func TestWebAuthnSessionStore_Create_Authentication(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	session := &WebAuthnSession{
		Challenge:  "auth-challenge",
		UserHandle: "test-user-handle",
		UserID:     "test-user-id",
		Operation:  "authentication",
	}

	sessionID, err := store.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, "authentication", retrieved.Operation)
}

// TestWebAuthnSessionStore_Get_NotFound tests retrieving a non-existent session
func TestWebAuthnSessionStore_Get_NotFound(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	session, err := store.Get(ctx, "non-existent-id")
	require.NoError(t, err)
	assert.Nil(t, session)
}

// TestWebAuthnSessionStore_Delete tests deleting a WebAuthn session
func TestWebAuthnSessionStore_Delete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	session := &WebAuthnSession{
		Challenge: "test-challenge",
		UserID:    "test-user-id",
		Operation: "registration",
	}

	sessionID, err := store.Create(ctx, session)
	require.NoError(t, err)

	// Verify session exists
	retrieved, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Delete session
	err = store.Delete(ctx, sessionID)
	require.NoError(t, err)

	// Verify session no longer exists
	retrieved, err = store.Get(ctx, sessionID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestWebAuthnSessionStore_DeleteByUserID tests deleting all sessions for a user
func TestWebAuthnSessionStore_DeleteByUserID(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	userID := "user-123"

	// Create multiple sessions for the same user
	for i := 0; i < 3; i++ {
		session := &WebAuthnSession{
			Challenge: "test-challenge",
			UserID:    userID,
			Operation: "registration",
		}
		_, err := store.Create(ctx, session)
		require.NoError(t, err)
	}

	// Create one session for a different user
	otherSession := &WebAuthnSession{
		Challenge: "other-challenge",
		UserID:    "other-user",
		Operation: "registration",
	}
	otherSessionID, err := store.Create(ctx, otherSession)
	require.NoError(t, err)

	// Delete all sessions for user123
	err = store.DeleteByUserID(ctx, userID)
	require.NoError(t, err)

	// Verify the other user's session still exists
	retrieved, err := store.Get(ctx, otherSessionID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "other-user", retrieved.UserID)
}

// TestWebAuthnSessionTTL tests that sessions expire after TTL
func TestWebAuthnSessionTTL(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	ctx := context.Background()

	session := &WebAuthnSession{
		Challenge: "test-challenge",
		UserID:    "test-user-id",
		Operation: "registration",
	}

	sessionID, err := store.Create(ctx, session)
	require.NoError(t, err)

	// Fast forward time in miniredis
	mr.FastForward(2 * time.Minute)

	// Session should be expired
	retrieved, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestNewWebAuthnSessionStore tests the constructor
func TestNewWebAuthnSessionStore(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	store := NewWebAuthnSessionStore(client)
	assert.NotNil(t, store)
	assert.Equal(t, client, store.client)
}
