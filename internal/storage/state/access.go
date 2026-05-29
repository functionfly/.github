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
// Order: explicit user grant → same-tenant membership → public read (handled in CheckPermission).
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

	// Same-tenant members have full access unless explicitly restricted via permission rows.
	if userTenantID != uuid.Nil && userTenantID == state.TenantID {
		return true, nil
	}

	allowed, err := r.CheckPermission(ctx, state.ID, "user", userID, requiredPermission)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	if userTenantID != uuid.Nil && userTenantID == state.TenantID {
		return true, nil
	}

	return false, nil
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
