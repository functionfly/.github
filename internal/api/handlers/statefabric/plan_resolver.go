package statefabric

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
)

type repoPlanResolver struct {
	repo storage.Repository
}

// RepoPlanResolver adapts storage.Repository to PlanResolver.
func RepoPlanResolver(repo storage.Repository) PlanResolver {
	return &repoPlanResolver{repo: repo}
}

func (r *repoPlanResolver) GetTenantPlan(ctx context.Context, tenantID uuid.UUID) string {
	subscription, err := r.repo.GetSubscriptionByTenantID(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("failed to get tenant subscription for state fabric quota")
		return plans.PlanStarter
	}
	if subscription == nil || subscription.PricingTier == nil {
		return plans.PlanFree
	}
	return subscription.PricingTier.Name
}