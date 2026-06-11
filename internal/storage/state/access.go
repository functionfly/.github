package state

import (
	"context"

	"github.com/google/uuid"
)

// Permission column names for state_permissions checks.
const (
	PermRead    = "can_read"
	PermWrite   = "can_write"
	PermDelete  = "can_delete"
	PermAdmin   = "can_admin"
	PermTrigger = "can_trigger"
)

// UserHasPermission returns true when the user may perform the requested operation.
// Access requires either an explicit permission grant or the resource is public (for read only).
// Tenant membership alone does NOT grant access - explicit permission is always required.
func (r *StateRepository) UserHasPermission(
	ctx context.Context,
	state *State,
	userID uuid.UUID,
	userTenantID uuid.UUID,
	requiredPermission string,
) (bool, error) {
	if state == nil {
		return false, nil
	}

	allowed, err := r.CheckPermission(ctx, state.ID, "user", userID, requiredPermission)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

// GrantCreatorPermissions grants the creating user full control over a state container.
func (r *StateRepository) GrantCreatorPermissions(ctx context.Context, stateID, userID uuid.UUID) error {
	principalID := userID
	_, err := r.GrantPermission(ctx, &StatePermission{
		StateID:       stateID,
		PrincipalType: "user",
		PrincipalID:   &principalID,
		CanRead:       true,
		CanWrite:      true,
		CanDelete:     true,
		CanAdmin:      true,
		CanTrigger:    true,
	})
	return err
}
