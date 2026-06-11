package api

import (
	dnahandler "github.com/functionfly/functionfly/internal/api/handlers/dna"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerDNARoutes wires all Function DNA endpoints.
func registerDNARoutes(
	_ *Server,
	api *mux.Router,
	protected *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	dnaHandler *dnahandler.Handler,
) {
	// ── Function DNA (protected) ──────────────────────────────────────────────
	dna := protected.PathPrefix("/functions/{id}/dna").Subrouter()
	dna.HandleFunc("", dnaHandler.GetProfile).Methods("GET", "OPTIONS")
	dna.HandleFunc("/mutations", dnaHandler.ListMutations).Methods("GET", "OPTIONS")
	dna.HandleFunc("/mutations/{mutation_id}", dnaHandler.GetMutation).Methods("GET", "OPTIONS")
	dna.HandleFunc("/variants/{mutation_id}/accept", dnaHandler.AcceptVariant).Methods("POST", "OPTIONS")
	dna.HandleFunc("/variants/{mutation_id}/reject", dnaHandler.RejectVariant).Methods("POST", "OPTIONS")
	dna.HandleFunc("/variants/{mutation_id}/rollback", dnaHandler.RollbackVariant).Methods("POST", "OPTIONS")
	dna.HandleFunc("/insights", dnaHandler.GetInsights).Methods("GET", "OPTIONS")
	dna.HandleFunc("/analyze", dnaHandler.TriggerAnalysis).Methods("POST", "OPTIONS")
	dna.HandleFunc("/evolution", dnaHandler.ToggleEvolution).Methods("POST", "OPTIONS")
	dna.HandleFunc("/verify", dnaHandler.VerifyHash).Methods("GET", "OPTIONS")

	// ── Enterprise DNA Insights (protected) ───────────────────────────────────
	protected.HandleFunc("/dna/enterprise/insights", dnaHandler.GetEnterpriseInsights).Methods("GET", "OPTIONS")
}
