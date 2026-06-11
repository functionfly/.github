package statefabric

import (
	"context"

	"github.com/google/uuid"
)

// UserHasFabricPermission checks whether a user may access a fabric-backed state container.
func (r *Repository) UserHasFabricPermission(
	ctx context.Context,
	tenantID, fabricID, userID uuid.UUID,
	requiredPermission string,
) (bool, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return false, err
	}
	if state.TenantID != tenantID {
		return false, nil
	}
	return r.stateRepo.UserHasPermission(ctx, state, userID, tenantID, requiredPermission)
}

// GrantFabricCreatorPermissions grants the creating user full control over a fabric.
func (r *Repository) GrantFabricCreatorPermissions(ctx context.Context, fabricID, userID uuid.UUID) error {
	return r.stateRepo.GrantCreatorPermissions(ctx, fabricID, userID)
}

// CountFabricsByTenant returns the number of state containers for a tenant (fabric quota).
func (r *Repository) CountFabricsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	_, total, err := r.ListFabrics(ctx, ListOptions{TenantID: tenantID, Limit: 1, Offset: 0})
	return int(total), err
}

// IsPipelineExecutionConfigured reports whether pipeline steps can invoke functions.
func (r *Repository) IsPipelineExecutionConfigured() bool {
	return r.baseURL != "" && r.httpClient != nil
}

// PipelineExecutionStatus returns configuration details for health checks.
func (r *Repository) PipelineExecutionStatus() map[string]interface{} {
	return map[string]interface{}{
		"configured":       r.IsPipelineExecutionConfigured(),
		"base_url_set":     r.baseURL != "",
		"http_client_set":  r.httpClient != nil,
		"api_key_set":      r.apiKey != "",
	}
}

// HealthCheck verifies connectivity to R2 storage and pipeline execution service.
func (r *Repository) HealthCheck(ctx context.Context) map[string]interface{} {
	result := map[string]interface{}{
		"r2_storage":      map[string]interface{}{"available": false},
		"pipeline_exec":    r.PipelineExecutionStatus(),
	}

	if r.r2Backend != nil {
		if err := r.r2Backend.HealthCheck(ctx); err != nil {
			result["r2_storage"] = map[string]interface{}{
				"available": false,
				"error":     err.Error(),
			}
		} else {
			result["r2_storage"] = map[string]interface{}{
				"available": true,
			}
		}
	} else {
		result["r2_storage"] = map[string]interface{}{
			"available": false,
			"error":     "R2 backend not configured",
		}
	}

	return result
}
