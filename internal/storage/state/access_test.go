package state

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserHasPermission_NilState(t *testing.T) {
	repo := &StateRepository{}
	allowed, err := repo.UserHasPermission(t.Context(), nil, uuid.New(), uuid.New(), PermRead)
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestUserHasPermission_SameTenantMember(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	state := &State{
		ID:       uuid.New(),
		TenantID: tenantID,
	}

	repo := &StateRepository{}
	allowed, err := repo.UserHasPermission(t.Context(), state, userID, tenantID, PermWrite)
	assert.NoError(t, err)
	assert.True(t, allowed)
}
