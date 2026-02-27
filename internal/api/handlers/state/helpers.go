package state

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	staterepo "github.com/functionfly/functionfly/internal/storage/state"
)

// Helper function to convert string to pointer
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Helper for nested value setting
func setNestedValue(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, "/")
	current := data

	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		current = current[part].(map[string]interface{})
	}
	current[parts[len(parts)-1]] = value
}

// Helper for nested value deletion
func deleteNestedValue(data map[string]interface{}, path string) {
	parts := strings.Split(path, "/")
	current := data

	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, ok := current[part]; !ok {
			return
		}
		current = current[part].(map[string]interface{})
	}
	delete(current, parts[len(parts)-1])
}

// Helper for testing nested values
func testNestedValue(data map[string]interface{}, path string, expected interface{}) bool {
	if path == "" || path == "/" {
		return reflect.DeepEqual(data, expected)
	}
	parts := strings.Split(path[1:], "/")
	current := data

	for _, part := range parts {
		if val, ok := current[part]; ok {
			if nested, ok := val.(map[string]interface{}); ok {
				current = nested
			} else {
				return reflect.DeepEqual(val, expected)
			}
		} else {
			return false
		}
	}
	return reflect.DeepEqual(current, expected)
}

// checkPermission checks if the user has the required permission on a state
func (h *Handler) checkPermission(ctx context.Context, stateID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	return h.stateRepo.CheckPermission(ctx, stateID, "user", userID, permission)
}

// requirePermission checks permission and returns HTTP error if denied
func (h *Handler) requirePermission(w http.ResponseWriter, r *http.Request, stateID uuid.UUID, userID uuid.UUID, permission string) bool {
	hasPermission, err := h.checkPermission(r.Context(), stateID, userID, permission)
	if err != nil {
		logrus.Errorf("failed to check permission: %v", err)
		http.Error(w, "permission check failed", http.StatusInternalServerError)
		return false
	}
	if !hasPermission {
		http.Error(w, "permission denied", http.StatusForbidden)
		return false
	}
	return true
}

// checkStateAccess checks if user can access the state (owner or has permission)
// Returns: hasAccess bool, isOwner bool, error
func (h *Handler) checkStateAccess(ctx context.Context, state *staterepo.State, userID uuid.UUID) (bool, bool, error) {
	// Check if user is the owner (via tenant)
	if state.TenantID == uuid.Nil {
		return false, false, nil
	}
	// For now, all tenant users have full access
	// In production, you'd check if user belongs to the tenant
	return true, true, nil
}