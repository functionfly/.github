package wasm

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// parseStateFabricPath resolves tenant, fabric, and key from an edge state path.
// defaultTenant and defaultFabric come from the executing function's context.
func parseStateFabricPath(path string, defaultTenant, defaultFabric uuid.UUID) (uuid.UUID, uuid.UUID, string, error) {
	parts := strings.Split(path, "/")

	switch len(parts) {
	case 1:
		return defaultTenant, defaultFabric, parts[0], nil
	case 2:
		fabricID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID: %w", err)
		}
		return defaultTenant, fabricID, parts[1], nil
	case 3:
		tenantID, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid tenant ID: %w", err)
		}
		if tenantID != defaultTenant {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("cross-tenant state access denied")
		}
		fabricID, err := uuid.Parse(parts[1])
		if err != nil {
			return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid fabric ID: %w", err)
		}
		return tenantID, fabricID, parts[2], nil
	default:
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid path format, expected 'tenant/fabric/key', 'fabric/key', or 'key'")
	}
}
