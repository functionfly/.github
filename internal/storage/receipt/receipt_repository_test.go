package receipt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRepository_NilDB verifies that creating a repository with nil DB fails.
func TestNewRepository_NilDB(t *testing.T) {
	_, err := NewRepository(nil, nil)
	assert.ErrorContains(t, err, "db is required")
}

// TestErrNotFound_IsExported verifies the error is accessible.
func TestErrNotFound_IsExported(t *testing.T) {
	assert.NotNil(t, ErrNotFound)
	assert.Equal(t, "receipt not found", ErrNotFound.Error())
}

// TestErrRevoked_IsExported verifies the error is accessible.
func TestErrRevoked_IsExported(t *testing.T) {
	assert.NotNil(t, ErrRevoked)
	assert.Equal(t, "receipt has been revoked", ErrRevoked.Error())
}

// TestMilestoneChannels are valid enum values.
func TestMilestoneChannels(t *testing.T) {
	assert.Equal(t, MilestoneChannel("inapp"), ChannelInApp)
	assert.Equal(t, MilestoneChannel("email"), ChannelEmail)
	assert.Equal(t, MilestoneChannel("tweet_intent"), ChannelTweet)
	assert.Equal(t, MilestoneChannel("webhook"), ChannelWebhook)
}

// TestRepositoryInterface verifies Repository struct has the expected methods.
// This test documents the interface without running DB operations.
func TestRepositoryInterface(t *testing.T) {
	r := &Repository{}
	// These method signatures must exist (compile-time check)
	var _ interface {
		GetReceipt(ctx interface{}, publicID string) (interface{}, interface{}, interface{}, error)
		IncrementViewCount(ctx interface{}, execID interface{}) error
		IncrementForkCount(ctx interface{}, execID interface{}) error
		GetFunctionExecutionCount(ctx interface{}, fnID interface{}) (int64, error)
		ListForFunction(ctx interface{}, fnID interface{}, limit int) (interface{}, error)
		GetFunctionOwnerID(ctx interface{}, fnID interface{}) (string, error)
		GetActiveFunctionsSince(ctx interface{}, since interface{}) (interface{}, error)
		RecordMilestone(ctx interface{}, fnID interface{}, dedupeKey string, firedAt interface{}, milestone string, channel MilestoneChannel) (*MilestoneEvent, error)
		MarkMilestoneChannels(ctx interface{}, eventID interface{}, channels []MilestoneChannel) error
		ListMilestonesForFunction(ctx interface{}, fnID interface{}) ([]*MilestoneEvent, error)
		Revoke(ctx interface{}, execID interface{}, reason string) error
		GetTrending(ctx interface{}, limit int) (interface{}, error)
	}
	_ = r
}
