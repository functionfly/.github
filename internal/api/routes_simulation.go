package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/simulation"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerSimulationRoutes wires R-Sim simulation engine endpoints.
// R-Sim provides dry-run workflow simulation, Monte Carlo analysis, failure injection,
// outcome prediction, stress testing, resource collision detection, agent behavior prediction,
// and hallucination risk analysis.
func registerSimulationRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	simHandler *simulation.Handler,
) {
	// Simulation lifecycle — authenticated
	// POST /v1/simulate/workflow — dry-run workflow simulation
	api.HandleFunc("/v1/simulate/workflow", authMiddleware.RequireAuth(simHandler.HandleSimulateWorkflow)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/simulate/workflow/{id}", authMiddleware.RequireAuth(simHandler.HandleGetSimulation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/simulate/workflow/{id}/abort", authMiddleware.RequireAuth(simHandler.HandleAbortSimulation)).Methods("POST", "OPTIONS")

	// Monte Carlo simulation
	api.HandleFunc("/v1/simulate/monte-carlo", authMiddleware.RequireAuth(simHandler.HandleMonteCarloSimulation)).Methods("POST", "OPTIONS")

	// Failure injection
	api.HandleFunc("/v1/simulate/failure-inject", authMiddleware.RequireAuth(simHandler.HandleFailureInjection)).Methods("POST", "OPTIONS")

	// Forecast and prediction
	api.HandleFunc("/v1/forecast/execution", authMiddleware.RequireAuth(simHandler.HandleExecutionForecast)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/forecast/cost", authMiddleware.RequireAuth(simHandler.HandleCostForecast)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/forecast/latency", authMiddleware.RequireAuth(simHandler.HandleLatencyForecast)).Methods("POST", "OPTIONS")

	// Stress test
	api.HandleFunc("/v1/stress-test/start", authMiddleware.RequireAuth(simHandler.HandleStartStressTest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/stress-test/{id}", authMiddleware.RequireAuth(simHandler.HandleGetStressTest)).Methods("GET", "OPTIONS")
	api.HandleFunc("/v1/stress-test/{id}/abort", authMiddleware.RequireAuth(simHandler.HandleAbortStressTest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/v1/stress-test/{id}/results", authMiddleware.RequireAuth(simHandler.HandleGetStressTestResults)).Methods("GET", "OPTIONS")

	// Resource collision detection
	api.HandleFunc("/v1/detect/collisions", authMiddleware.RequireAuth(simHandler.HandleDetectResourceCollisions)).Methods("POST", "OPTIONS")

	// Agent behavior prediction
	api.HandleFunc("/v1/predict/agent-behavior", authMiddleware.RequireAuth(simHandler.HandlePredictAgentBehavior)).Methods("POST", "OPTIONS")

	// Hallucination risk analysis
	api.HandleFunc("/v1/analyze/hallucination-risk", authMiddleware.RequireAuth(simHandler.HandleHallucinationRiskAnalysis)).Methods("POST", "OPTIONS")
}

// suppressUnused silences unused import warnings
var _ http.HandlerFunc