package swarm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigningService_SignMessage(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)

	agentID := "agent-1"
	secretKey := "test-secret-key"
	payload := []byte(`{"task":"test"}`)
	nonce := "test-nonce-123"
	sequenceNum := int64(1)

	// Sign message
	signature := svc.SignMessage(agentID, secretKey, payload, nonce, sequenceNum)
	assert.NotEmpty(t, signature)

	// Same inputs should produce same signature
	signature2 := svc.SignMessage(agentID, secretKey, payload, nonce, sequenceNum)
	assert.Equal(t, signature, signature2)

	// Different nonce should produce different signature
	signature3 := svc.SignMessage(agentID, secretKey, payload, "different-nonce", sequenceNum)
	assert.NotEqual(t, signature, signature3)

	// Different sequence should produce different signature
	signature4 := svc.SignMessage(agentID, secretKey, payload, nonce, sequenceNum+1)
	assert.NotEqual(t, signature, signature4)
}

func TestSigningService_VerifySignature(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)

	agentID := "agent-1"
	secretKey := "test-secret-key"
	payload := []byte(`{"task":"test"}`)
	nonce := "test-nonce-123"
	sequenceNum := int64(1)

	// Generate valid signature
	validSig := svc.SignMessage(agentID, secretKey, payload, nonce, sequenceNum)

	// Verify should pass
	assert.True(t, svc.VerifySignature(agentID, validSig, secretKey, payload, nonce, sequenceNum))

	// Wrong key should fail
	assert.False(t, svc.VerifySignature(agentID, validSig, "wrong-key", payload, nonce, sequenceNum))

	// Tampered payload should fail
	tamperedPayload := []byte(`{"task":"tampered"}`)
	assert.False(t, svc.VerifySignature(agentID, validSig, secretKey, tamperedPayload, nonce, sequenceNum))

	// Wrong nonce should fail
	assert.False(t, svc.VerifySignature(agentID, validSig, secretKey, payload, "wrong-nonce", sequenceNum))

	// Wrong sequence should fail
	assert.False(t, svc.VerifySignature(agentID, validSig, secretKey, payload, nonce, sequenceNum+1))
}

func TestSigningService_CheckAndStoreNonce_RedisNil_UsesInMemory(t *testing.T) {
	t.Parallel()

	// nil redis client triggers in-memory fallback
	svc := NewSigningService(nil)
	ctx := context.Background()

	agentID := "agent-1"
	nonce := "unique-nonce-123"

	// First use should be allowed
	allowed, err := svc.CheckAndStoreNonce(ctx, agentID, nonce)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Same nonce again should be rejected (replay protection)
	allowed2, err := svc.CheckAndStoreNonce(ctx, agentID, nonce)
	require.NoError(t, err)
	assert.False(t, allowed2)
}

func TestSigningService_CheckAndStoreNonce_DifferentAgents(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)
	ctx := context.Background()

	nonce := "same-nonce"

	// Agent 1 can use nonce
	allowed1, err := svc.CheckAndStoreNonce(ctx, "agent-1", nonce)
	require.NoError(t, err)
	assert.True(t, allowed1)

	// Agent 2 can also use same nonce (nonces are per-agent)
	allowed2, err := svc.CheckAndStoreNonce(ctx, "agent-2", nonce)
	require.NoError(t, err)
	assert.True(t, allowed2)

	// Agent 1 cannot reuse nonce
	allowed1Again, err := svc.CheckAndStoreNonce(ctx, "agent-1", nonce)
	require.NoError(t, err)
	assert.False(t, allowed1Again)
}

func TestSigningService_ValidateReplay(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)
	ctx := context.Background()

	agentID := "agent-replay-test"
	nonce := "replay-test-nonce"

	// Valid replay
	valid, err := svc.ValidateReplay(ctx, agentID, nonce)
	require.NoError(t, err)
	assert.True(t, valid)

	// Replay should be rejected
	replay, err := svc.ValidateReplay(ctx, agentID, nonce)
	require.NoError(t, err)
	assert.False(t, replay)
}

func TestSigningService_GetLastSequence_RedisNil(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)
	ctx := context.Background()

	// With nil redis, should return 0
	seq, err := svc.GetLastSequence(ctx, "agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), seq)
}

func TestSigningService_IncrementSequence_RedisNil(t *testing.T) {
	t.Parallel()

	svc := NewSigningService(nil)
	ctx := context.Background()

	// With nil redis, should return 0 (no-op)
	seq, err := svc.IncrementSequence(ctx, "agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), seq)
}
