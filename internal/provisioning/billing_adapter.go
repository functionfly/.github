package provisioning

import (
	"context"

	"github.com/google/uuid"
)

// ProvisionBundleForBilling is a function adapter that satisfies the billing handler's
// provisionBundleFn signature: func(ctx, tenantID, bundleSlug) (status, componentCount, error).
// This avoids circular imports between provisioning and billing packages.
func ProvisionBundleForBilling(provisioner *BundleProvisioner) func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error) {
	return func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error) {
		result, err := provisioner.ProvisionBundle(ctx, tenantID, bundleSlug)
		if err != nil {
			return "failed", 0, err
		}
		return string(result.Status), len(result.Components), nil
	}
}
