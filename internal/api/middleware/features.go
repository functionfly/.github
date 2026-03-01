package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	// ContextKeyTenantPlan is the context key for tenant plan
	ContextKeyTenantPlan = "tenant_plan"
	// ContextKeyFeatureChecker is the context key for feature checker
	ContextKeyFeatureChecker = "feature_checker"
)

// FeatureMiddleware provides middleware for feature gating
type FeatureMiddleware struct{}

// NewFeatureMiddleware creates a new feature middleware
func NewFeatureMiddleware() *FeatureMiddleware {
	return &FeatureMiddleware{}
}

// RequireFeature returns a handler that checks for a feature before proceeding
func (fm *FeatureMiddleware) RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plan := GetTenantPlan(r)
			if plan == "" {
				http.Error(w, "Tenant plan not found", http.StatusInternalServerError)
				return
			}

			checker := plans.NewFeatureChecker(plan)
			if err := checker.ValidateFeatureAccess(feature); err != nil {
				logrus.Warnf("Feature access denied: %v for plan %s", err, plan)
				fm.writeFeatureError(w, r, feature, plan)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyFeature returns a handler that checks if the tenant has any of the features
func (fm *FeatureMiddleware) RequireAnyFeature(features ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plan := GetTenantPlan(r)
			if plan == "" {
				http.Error(w, "Tenant plan not found", http.StatusInternalServerError)
				return
			}

			checker := plans.NewFeatureChecker(plan)
			hasFeature := checker.HasAnyFeature(features...)
			if !hasFeature {
				logrus.Warnf("Feature access denied: none of %v for plan %s", features, plan)
				fm.writeFeatureError(w, r, features[0], plan)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllFeatures returns a handler that checks if the tenant has all of the features
func (fm *FeatureMiddleware) RequireAllFeatures(features ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plan := GetTenantPlan(r)
			if plan == "" {
				http.Error(w, "Tenant plan not found", http.StatusInternalServerError)
				return
			}

			checker := plans.NewFeatureChecker(plan)
			if err := checker.ValidateFeaturesAccess(features...); err != nil {
				logrus.Warnf("Feature access denied: %v for plan %s", err, plan)
				fm.writeFeatureError(w, r, features[0], plan)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithFeatureChecker adds a feature checker to the request context
func (fm *FeatureMiddleware) WithFeatureChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plan := GetTenantPlan(r)
		if plan != "" {
			checker := plans.NewFeatureChecker(plan)
			ctx := context.WithValue(r.Context(), ContextKeyFeatureChecker, checker)
			ctx = context.WithValue(ctx, ContextKeyTenantPlan, plan)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetFeatureChecker retrieves the feature checker from the request context
func GetFeatureChecker(r *http.Request) *plans.FeatureChecker {
	if checker, ok := r.Context().Value(ContextKeyFeatureChecker).(*plans.FeatureChecker); ok {
		return checker
	}
	return nil
}

// GetTenantPlan retrieves the tenant plan from the request context
func GetTenantPlan(r *http.Request) string {
	if plan, ok := r.Context().Value(ContextKeyTenantPlan).(string); ok {
		return plan
	}
	return ""
}

// writeFeatureError writes a feature access error response
func (fm *FeatureMiddleware) writeFeatureError(w http.ResponseWriter, r *http.Request, feature, plan string) {
	def, _ := plans.GetFeatureDefinition(feature)
	requiredPlan := plans.GetRequiredPlan(feature)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "FEATURE_NOT_AVAILABLE",
			"message":     fmt.Sprintf("Feature '%s' requires a higher tier", def.Name),
			"feature":     feature,
			"current_plan": plan,
			"required_plan": requiredPlan,
		},
	}

	if plan == plans.PlanStarter || plan == plans.PlanAgentStarter {
		errorResp["upgrade_options"] = []string{plans.PlanPro, plans.PlanEnterprise}
	}

	json.NewEncoder(w).Encode(errorResp)
}

// GetFeatureFromRequest extracts feature from request path or query
func GetFeatureFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	if feature, ok := vars["feature"]; ok {
		return feature
	}
	return r.URL.Query().Get("feature")
}
