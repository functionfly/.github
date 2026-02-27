package execution

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// validateRuntimeForPlan checks if the requested runtime is allowed for the tenant's plan
func validateRuntimeForPlan(tenantPlan string, runtime string) error {
	if !plans.IsValidRuntimeForPlan(tenantPlan, runtime) {
		return fmt.Errorf("runtime '%s' is only available for Enterprise tier. Your current plan is '%s'", runtime, tenantPlan)
	}
	return nil
}

// getTenantPlanFromContext retrieves the tenant's plan from the billing system
func getTenantPlanFromContext(repo storage.Repository, tenantID uuid.UUID) string {
	// Query the billing system for the tenant's active subscription
	subscription, err := repo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		// Log the error but fall back to starter plan for safety
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get tenant subscription, falling back to starter plan")
		return plans.PlanStarter
	}

	// If no subscription found, tenant might be on free/starter plan
	if subscription == nil || subscription.PricingTier == nil {
		logrus.WithField("tenant_id", tenantID).Info("No active subscription found, using starter plan")
		return plans.PlanStarter
	}

	// Return the plan name from the pricing tier
	return subscription.PricingTier.Name
}

// outputsEqual compares two JSON outputs for equality
func outputsEqual(a, b json.RawMessage) bool {
	// Simple JSON comparison - normalize and compare
	var aNormalized, bNormalized interface{}

	if err := json.Unmarshal(a, &aNormalized); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bNormalized); err != nil {
		return false
	}

	// For simple types, use string comparison of marshaled JSON
	aBytes, err := json.Marshal(aNormalized)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(bNormalized)
	if err != nil {
		return false
	}

	return string(aBytes) == string(bBytes)
}

// shouldVerifyReplay determines if a replay should be verified based on various factors
func shouldVerifyReplay(fnVersion *storage.RegistryFunctionVersion, executionCount int, lastVerified *time.Time, recentFailureRate float64) bool {
	// Only verify deterministic functions
	if !fnVersion.Deterministic {
		return false
	}

	// Always verify on first execution
	if executionCount == 1 {
		return true
	}

	now := time.Now()

	// High-risk functions: verify more frequently if they have recent failures
	if recentFailureRate > 0.1 { // More than 10% recent failures
		// Verify every 3rd execution for high-risk functions
		if executionCount%3 == 0 {
			return true
		}
		// Or if not verified in the last 6 hours
		if lastVerified == nil || now.Sub(*lastVerified) > 6*time.Hour {
			return true
		}
	}

	// Medium-risk functions: standard verification schedule
	if executionCount%10 == 0 {
		return true
	}

	// Time-based verification with adaptive intervals
	if lastVerified != nil {
		timeSinceLastVerification := now.Sub(*lastVerified)

		// Base interval: 24 hours for stable functions
		baseInterval := 24 * time.Hour

		// Reduce interval for newer functions (more verification needed initially)
		if executionCount < 100 {
			baseInterval = 12 * time.Hour
		}

		// Increase interval for very stable functions (reduce verification overhead)
		if executionCount > 1000 && recentFailureRate < 0.01 {
			baseInterval = 48 * time.Hour
		}

		if timeSinceLastVerification > baseInterval {
			return true
		}
	} else {
		// Never verified before - verify now
		return true
	}

	// Random sampling for additional coverage (1% of remaining executions)
	// This provides statistical confidence even for stable functions
	return time.Now().UnixNano()%100 == 0
}