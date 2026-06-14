package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/consciousness"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/gorilla/mux"
)

// registerConsciousnessRoutes registers all Function Consciousness API routes.
// Feature gating is applied based on plan tier:
//   - Basic consciousness (score, insights, preferences) requires FeatureConsciousnessBasic (Pro+)
//   - Run analysis requires FeatureConsciousnessBasic (Pro+)
//   - Advanced features (auto-fix, predictive) require FeatureConsciousnessAdvanced (Enterprise+)
//   - Export/delete (GDPR) is available to all authenticated users
func registerConsciousnessRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	featureMiddleware *middleware.FeatureMiddleware,
	consciousnessHandler *consciousness.Handler,
) {
	// Subrouter for consciousness endpoints
	c := api.PathPrefix("/consciousness").Subrouter()

	// Feature-gated routes - Basic Consciousness (requires Pro+)
	basicConsciousness := featureMiddleware.RequireFeature(plans.FeatureConsciousnessBasic)

	// Score endpoint - requires basic consciousness
	c.HandleFunc("/score",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetAwarenessScore, basicConsciousness)),
	).Methods("GET", "OPTIONS")

	// List insights - requires basic consciousness
	c.HandleFunc("/insights",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.ListInsights, basicConsciousness)),
	).Methods("GET", "OPTIONS")

	// Get single insight - requires basic consciousness
	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetInsight, basicConsciousness)),
	).Methods("GET", "OPTIONS")

	// Dismiss insight - requires basic consciousness
	c.HandleFunc("/insights/{id}/dismiss",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.DismissInsight, basicConsciousness)),
	).Methods("POST", "OPTIONS")

	// Apply action - requires basic consciousness
	c.HandleFunc("/insights/{id}/apply",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.ApplyAction, basicConsciousness)),
	).Methods("POST", "OPTIONS")

	// Preferences - requires basic consciousness
	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetPreferences, basicConsciousness)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.UpdatePreferences, basicConsciousness)),
	).Methods("PUT", "OPTIONS")

	// Run analysis - requires basic consciousness
	c.HandleFunc("/run",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.RunAnalysis, basicConsciousness)),
	).Methods("POST", "OPTIONS")

	// GDPR routes - available to all authenticated users (data belongs to tenant)
	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(consciousnessHandler.DeleteInsight),
	).Methods("DELETE", "OPTIONS")

	c.HandleFunc("/export",
		authMiddleware.RequireAuth(consciousnessHandler.ExportData),
	).Methods("GET", "OPTIONS")
}

// registerConsciousnessRoutesWithCustomConfig registers routes with custom feature checks.
// Use this when you need different feature tiers for different endpoints.
func registerConsciousnessRoutesWithCustomConfig(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	featureMiddleware *middleware.FeatureMiddleware,
	consciousnessHandler *consciousness.Handler,
) {
	c := api.PathPrefix("/consciousness").Subrouter()

	// Advanced feature gating - split by feature tier
	basicFeature := featureMiddleware.RequireFeature(plans.FeatureConsciousnessBasic)
	advancedFeature := featureMiddleware.RequireFeature(plans.FeatureConsciousnessAdvanced)

	// Basic endpoints (Pro+)
	c.HandleFunc("/score",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetAwarenessScore, basicFeature)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/insights",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.ListInsights, basicFeature)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetInsight, basicFeature)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.GetPreferences, basicFeature)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.UpdatePreferences, basicFeature)),
	).Methods("PUT", "OPTIONS")

	// Advanced endpoints (Enterprise+) - auto-fix, predictive alerts
	c.HandleFunc("/insights/{id}/dismiss",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.DismissInsight, basicFeature)),
	).Methods("POST", "OPTIONS")

	c.HandleFunc("/insights/{id}/apply",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.ApplyAction, advancedFeature)),
	).Methods("POST", "OPTIONS")

	c.HandleFunc("/run",
		authMiddleware.RequireAuth(withFeature(consciousnessHandler.RunAnalysis, basicFeature)),
	).Methods("POST", "OPTIONS")

	// GDPR - available to all authenticated users
	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(consciousnessHandler.DeleteInsight),
	).Methods("DELETE", "OPTIONS")

	c.HandleFunc("/export",
		authMiddleware.RequireAuth(consciousnessHandler.ExportData),
	).Methods("GET", "OPTIONS")
}

func withFeature(fn http.HandlerFunc, gate func(http.Handler) http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gate(fn).ServeHTTP(w, r)
	}
}
