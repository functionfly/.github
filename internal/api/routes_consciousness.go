package api

import (
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
	basicConsciousness := middleware.RequireFeature(plans.FeatureConsciousnessBasic)

	// Score endpoint - requires basic consciousness
	c.HandleFunc("/score",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.GetAwarenessScore)),
	).Methods("GET", "OPTIONS")

	// List insights - requires basic consciousness
	c.HandleFunc("/insights",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.ListInsights)),
	).Methods("GET", "OPTIONS")

	// Get single insight - requires basic consciousness
	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.GetInsight)),
	).Methods("GET", "OPTIONS")

	// Dismiss insight - requires basic consciousness
	c.HandleFunc("/insights/{id}/dismiss",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.DismissInsight)),
	).Methods("POST", "OPTIONS")

	// Apply action - requires basic consciousness
	c.HandleFunc("/insights/{id}/apply",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.ApplyAction)),
	).Methods("POST", "OPTIONS")

	// Preferences - requires basic consciousness
	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.GetPreferences)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.UpdatePreferences)),
	).Methods("PUT", "OPTIONS")

	// Run analysis - requires basic consciousness
	c.HandleFunc("/run",
		authMiddleware.RequireAuth(basicConsciousness(consciousnessHandler.RunAnalysis)),
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
	basicFeature := middleware.RequireFeature(plans.FeatureConsciousnessBasic)
	advancedFeature := middleware.RequireFeature(plans.FeatureConsciousnessAdvanced)

	// Basic endpoints (Pro+)
	c.HandleFunc("/score",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.GetAwarenessScore)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/insights",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.ListInsights)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.GetInsight)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.GetPreferences)),
	).Methods("GET", "OPTIONS")

	c.HandleFunc("/preferences",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.UpdatePreferences)),
	).Methods("PUT", "OPTIONS")

	// Advanced endpoints (Enterprise+) - auto-fix, predictive alerts
	c.HandleFunc("/insights/{id}/dismiss",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.DismissInsight)),
	).Methods("POST", "OPTIONS")

	c.HandleFunc("/insights/{id}/apply",
		authMiddleware.RequireAuth(advancedFeature(consciousnessHandler.ApplyAction)),
	).Methods("POST", "OPTIONS")

	c.HandleFunc("/run",
		authMiddleware.RequireAuth(basicFeature(consciousnessHandler.RunAnalysis)),
	).Methods("POST", "OPTIONS")

	// GDPR - available to all authenticated users
	c.HandleFunc("/insights/{id}",
		authMiddleware.RequireAuth(consciousnessHandler.DeleteInsight),
	).Methods("DELETE", "OPTIONS")

	c.HandleFunc("/export",
		authMiddleware.RequireAuth(consciousnessHandler.ExportData),
	).Methods("GET", "OPTIONS")
}
